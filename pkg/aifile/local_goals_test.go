package aifile

import "testing"

func TestLocalGoals(t *testing.T) {
	r, err := LintPath("../../testdata/ai/local-goals.per")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || len(r.Issues) != 0 || r.Coverage.UncheckedForms != 0 {
		t.Fatalf("%+v", r)
	}
	for _, c := range r.Checks {
		if c.Command == "goal" || c.Command == "set-goal" {
			if c.Confidence != "reference_backed" || c.ReferenceFamily != "AoC" || c.Source == "" || len(c.Parameters) != 2 || c.Notes == "" {
				t.Fatal(c)
			}
		}
	}
	for _, tt := range []struct{ fact, action, code string }{
		{"(goal 0 1)", "(do-nothing)", "reference_range"},
		{"(true)", "(set-goal 16001 1)", "reference_range"},
		{"(goal 1 == 0)", "(do-nothing)", "wrong_arity"},
		{"(true)", "(set-goal 1)", "wrong_arity"},
		{"(goal 1 \"zero\")", "(do-nothing)", "argument_kind"},
		{"(set-goal 1 0)", "(goal 1 0)", "invalid_context"},
		{"(goal phase 0)", "(set-goal phase 1)", "unresolved_constant"},
	} {
		r := lintPERText(t, "(defrule "+tt.fact+" => "+tt.action+")")
		issueWithCode(t, r, tt.code)
	}
}
