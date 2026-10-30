package aifile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lintPERText(t *testing.T, text string) *LintReport {
	t.Helper()
	path := filepath.Join(t.TempDir(), "Probe.per")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := LintPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func issueWithCode(t *testing.T, r *LintReport, code string) LintIssue {
	t.Helper()
	for _, issue := range r.Issues {
		if issue.Code == code {
			return issue
		}
	}
	t.Fatalf("missing %s: %+v", code, r.Issues)
	return LintIssue{}
}

func TestPassiveProbeStaticCoverage(t *testing.T) {
	r, err := LintPath("../../testdata/ai/passive")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || r.Summary.Files != 2 || r.Coverage.CheckedForms == 0 || r.Coverage.UncheckedForms != 0 {
		t.Fatalf("report=%+v", r)
	}
	if r.Catalogue.Revision != catalogueRevision || r.Verification.EngineVerified {
		t.Fatalf("unsupported verification claim: %+v", r)
	}
	if len(r.Checks) == 0 || r.Checks[0].Source == "" || r.Checks[0].Confidence != "reference_backed" {
		t.Fatalf("missing provenance: %+v", r.Checks)
	}
	issueWithCode(t, r, "xs_runtime_resolution_unverified")
	found := false
	for _, link := range r.Links {
		if link.Kind == "xs_call" && link.Name == "probe_tick" && link.Status == "declaration_found_runtime_unverified" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing XS source linkage: %+v", r.Links)
	}
}

func TestPERUnknownCommandHasLocationAndProvenance(t *testing.T) {
	r, err := LintPath("../../testdata/ai/typo.per")
	if err != nil {
		t.Fatal(err)
	}
	issue := issueWithCode(t, r, "unknown_command")
	if r.OK || issue.Line != 4 || issue.Column != 3 || !strings.Contains(issue.Message, "up-find-local") || issue.Confidence != "inferred_typo" || issue.Source == "" {
		t.Fatalf("report=%+v issue=%+v", r, issue)
	}
}

func TestPERArityKindsRangesAndConstants(t *testing.T) {
	for _, tt := range []struct {
		name, form, code string
		ok               bool
	}{
		{"arity", "(up-find-local c: 83)", "wrong_arity", false},
		{"operator", "(up-find-local x: 83 c: 10)", "argument_kind", false},
		{"quoted integer", "(up-find-local c: \"83\" c: 10)", "argument_kind", false},
		{"missing constant", "(up-get-search-state missing-goal)", "unresolved_constant", true},
		{"reference range", "(up-find-local c: 83 c: 255)", "reference_range", true},
		{"partial catalogue", "(up-get-player-fact 1 0 100)", "command_not_catalogued", true},
		{"action as fact", "(true)", "invalid_context", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := lintPERText(t, "(defrule (true) => "+tt.form+")")
			issueWithCode(t, r, tt.code)
			if r.OK != tt.ok {
				t.Fatalf("OK=%v, issues=%+v", r.OK, r.Issues)
			}
		})
	}
	r := lintPERText(t, "(defconst root 100) (defconst alias root) (defrule (true) => (up-get-search-state alias) (up-find-local c: 83 g: alias))")
	if !r.OK || len(r.Issues) != 0 {
		t.Fatalf("alias resolution: %+v", r.Issues)
	}
	r = lintPERText(t, "(defrule (nor (true) (false)) => (disable-self))")
	if !r.OK || len(r.Issues) != 0 {
		t.Fatalf("valid logical operator rejected: %+v", r.Issues)
	}
}

func TestPERXSCallLinkageIsConservative(t *testing.T) {
	for _, tt := range []struct {
		name, source, code string
		ok                 bool
	}{
		{"argument function", "void tick(int value=0) {}", "xs_function_requires_arguments", false},
		{"missing function", "// void tick() {}", "xs_function_unresolved", true},
		{"ambiguous function", "void tick() {} void tick() {}", "xs_function_unresolved", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			aiDir := filepath.Join(root, "resources", "_common", "ai")
			xsDir := filepath.Join(root, "resources", "_common", "xs")
			for _, dir := range []string{aiDir, xsDir} {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(aiDir, "Probe.per"), []byte("(include \"Probe.xs\") (defrule (true) => (xs-script-call \"tick\"))"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(xsDir, "Probe.xs"), []byte(tt.source), 0600); err != nil {
				t.Fatal(err)
			}
			r, err := LintPath(root)
			if err != nil {
				t.Fatal(err)
			}
			issueWithCode(t, r, tt.code)
			if r.OK != tt.ok || r.Verification.EngineVerified {
				t.Fatalf("report=%+v", r)
			}
		})
	}
}

func TestPERStringsCommentsAndMalformedForms(t *testing.T) {
	r := lintPERText(t, "; unmatched comment )\n(defrule (true) => (up-chat-data-to-all \"semi; paren( quote\\\"\" c: 1))")
	if !r.OK {
		t.Fatalf("string/comment parsing: %+v", r.Issues)
	}
	for _, tt := range []struct{ text, code string }{
		{"(defrule (true))", "invalid_rule"},
		{"(defrule (true) =>)", "invalid_rule"},
		{"(defrule (true) => (disable-self)", "unbalanced_parentheses"},
		{"(defconst x \"not closed)", "unterminated_string"},
		{strings.Repeat("(", 257), "nesting_limit"},
	} {
		issueWithCode(t, lintPERText(t, tt.text), tt.code)
	}
}

func TestPERConditionalAndLoadedSymbolsRemainUnverified(t *testing.T) {
	r := lintPERText(t, "#load-if-defined SOME-CIV\n(defconst t 97)\n#else\n(defconst t 98)\n#end-if\n(load \"constants\")\n(defrule (true) => (disable-timer t))")
	for _, code := range []string{"preprocessor_not_evaluated", "duplicate_constant", "load_not_expanded", "unresolved_constant"} {
		issueWithCode(t, r, code)
	}
	if !r.OK {
		t.Fatalf("conditional uncertainty became rejection: %+v", r.Issues)
	}
}
