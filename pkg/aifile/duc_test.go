package aifile

import (
	"strings"
	"testing"
)

func TestDUCMicroFixtureAndReferenceMetadata(t *testing.T) {
	r, err := LintPath("../../testdata/ai/duc-micro.per")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || len(r.Issues) != 0 || r.Coverage.UncheckedForms != 0 || r.Catalogue.Version != catalogueVersion {
		t.Fatalf("report=%+v", r)
	}
	byName := map[string]CommandCheck{}
	for _, check := range r.Checks {
		byName[check.Command] = check
	}
	for _, name := range []string{"up-find-local", "up-find-remote", "up-filter-distance", "up-filter-status", "up-clean-search", "up-get-search-state", "up-get-object-data", "up-set-target-object", "up-set-target-point", "up-target-objects", "up-target-point"} {
		check, found := byName[name]
		if !found || check.Source == "" || check.ReferenceFamily != "UP" || check.Confidence != "reference_backed" {
			t.Errorf("missing signature/provenance for %s: %+v", name, check)
		}
	}
	remote := byName["up-find-remote"].Parameters
	if len(remote) != 4 || remote[3].Range == nil || *remote[3].Range != ([2]int64{0, 40}) {
		t.Fatalf("remote bounds=%+v", remote)
	}
	point := byName["up-target-point"]
	if len(point.Parameters) != 4 || !hasValue(point.Parameters[0].Exceptions, 0) || !point.Parameters[1].Constants["action-attack-move"].DEOnly || !point.Parameters[1].Constants["action-attack-move"].NeedsDefinition {
		t.Fatalf("DE metadata=%+v", point)
	}
	if !strings.Contains(byName["up-set-target-point"].Notes, "uncertain") || len(byName["up-set-target-point"].Parameters) != 1 || byName["up-set-target-point"].Parameters[0].Range != nil {
		t.Fatal("conflicting upper bounds were silently hardened")
	}
	if r.Verification.EngineVerified {
		t.Fatal("source fixture promoted to engine evidence")
	}
}

func TestDUCDiagnosticsAndDynamicOperands(t *testing.T) {
	for _, tt := range []struct {
		name, form, code string
		ok               bool
	}{
		{"remote cap", "(up-find-remote c: 83 c: 41)", "reference_range", true},
		{"filter arity", "(up-filter-status c: status-ready)", "wrong_arity", false},
		{"distance type", "(up-filter-distance c: \"near\" c: 10)", "argument_kind", false},
		{"source enum", "(up-clean-search 3 -1 search-order-asc)", "reference_enum", true},
		{"partial object data", "(up-get-object-data 999 100)", "enum_value_not_catalogued", true},
		{"wrong output type", "(up-get-object-data object-data-hitpoints \"goal\")", "argument_kind", false},
		{"point gap", "(up-target-point 10 action-move -1 -1)", "reference_range", true},
		{"object action subset", "(up-target-objects 0 action-unload -1 -1)", "reference_enum", true},
		{"formation enum", "(up-target-objects 0 action-patrol 3 -1)", "reference_enum", true},
		{"DE definition missing", "(up-target-point 0 action-attack-move -1 -1)", "unresolved_constant", true},
		{"unresolved overload", "(up-target-point 0 unknown-action c: 83)", "unresolved_constant", true},
		{"noncombat overload", "(up-target-point 0 action-train c: 83)", "catalogue_uncertainty", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := lintPERText(t, "(defrule (true) => "+tt.form+")")
			issueWithCode(t, r, tt.code)
			if r.OK != tt.ok {
				t.Fatalf("OK=%v issues=%+v", r.OK, r.Issues)
			}
		})
	}
	for _, form := range []string{
		"(up-filter-distance g: 100 s: 101)",
		"(up-filter-status g: 999 c: 0)",
		"(up-find-remote c: 83 g: 100)",
		"(up-clean-search 1 -1 0)",
		"(up-target-point 0 1 -1 -1)",
	} {
		r := lintPERText(t, "(defrule (true) => "+form+")")
		if !r.OK || len(r.Issues) != 0 {
			t.Errorf("dynamic/sentinel rejected: %s %+v", form, r.Issues)
		}
	}
	r := lintPERText(t, "(defconst action-attack-move 19) (defconst action-patrol 2) (defconst patrol action-patrol) (defrule (true) => (up-target-point 0 action-attack-move -1 -1) (up-target-objects 0 patrol -1 -1))")
	if !r.OK || len(r.Issues) != 0 {
		t.Fatalf("explicit DE definition/alias=%+v", r.Issues)
	}
	r = lintPERText(t, "(defconst search-local 1) (defconst search-local 2) (defrule (true) => (up-clean-search search-local -1 search-order-none))")
	issueWithCode(t, r, "duplicate_constant")
	issueWithCode(t, r, "unresolved_constant")
}
