package aifile

import (
	"strings"
	"testing"
)

func TestCCAddResourceSignatureAndLimits(t *testing.T) {
	for _, body := range []string{"(cc-add-resource food -32768)", "(cc-add-resource wood 32767)", "(cc-add-resource stone 0)", "(defconst coin gold) (defrule (true) => (cc-add-resource coin -1))"} {
		text := "(defrule (true) => " + body + ")"
		if strings.HasPrefix(body, "(defconst") {
			text = body
		}
		r := lintPERText(t, text)
		if !r.OK || len(r.Issues) != 0 {
			t.Fatalf("%s: %+v", body, r.Issues)
		}
	}
	for _, tt := range []struct{ body, code string }{
		{"(cc-add-resource stone)", "wrong_arity"},
		{"(cc-add-resource stone 1 2)", "wrong_arity"},
		{"(cc-add-resource stone \"1\")", "argument_kind"},
		{"(cc-add-resource 4 1)", "reference_enum"},
		{"(cc-add-resource gold -32769)", "reference_range"},
		{"(cc-add-resource gold 32768)", "reference_range"},
	} {
		r := lintPERText(t, "(defrule (true) => "+tt.body+")")
		found := false
		for _, issue := range r.Issues {
			if issue.Code == tt.code {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: missing %s: %+v", tt.body, tt.code, r.Issues)
		}
	}
	r := lintPERText(t, "(defrule (true) => (set-signal 1) (fe-set-signal c: 1 c: 1) (cc-add-resource gold -1))")
	for _, c := range r.Checks {
		if c.Confidence != "reference_backed" {
			t.Fatal(c)
		}
		if c.Command == "set-signal" && (!strings.Contains(c.Notes, "101.103.57257.0") || c.RuntimeSemantics != "unverified") {
			t.Fatal(c)
		}
		if c.Command == "fe-set-signal" && strings.Contains(c.Notes, "canary") {
			t.Fatal("evidence leaked to MP")
		}
	}
}
