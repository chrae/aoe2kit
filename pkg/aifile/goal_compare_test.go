package aifile

import "testing"

func TestGoalComparisonAndStrategicNumber(t *testing.T) {
	for _, op := range []string{">", ">=", "<", "<=", "==", "!=", "c:>", "g:<=", "s:!="} {
		r := lintPERText(t, "(defrule (up-compare-goal 112 "+op+" 1) => (set-strategic-number 511 -2147483648))")
		if !r.OK || len(r.Issues) != 0 {
			t.Fatalf("%s: %+v", op, r.Issues)
		}
	}
	for _, tt := range []struct{ fact, action, code string }{
		{"(up-compare-goal 112 x:> 1)", "(do-nothing)", "argument_kind"},
		{"(up-compare-goal 0 > 1)", "(do-nothing)", "reference_range"},
		{"(up-compare-goal 112 s:> 512)", "(do-nothing)", "reference_range"},
		{"(up-compare-goal 112 g:> 0)", "(do-nothing)", "reference_range"},
		{"(up-compare-goal 112 >)", "(do-nothing)", "wrong_arity"},
		{"(true)", "(set-strategic-number 512 1)", "reference_range"},
		{"(true)", "(set-strategic-number 1 \"x\")", "argument_kind"},
		{"(true)", "(set-strategic-number 1)", "wrong_arity"},
	} {
		r := lintPERText(t, "(defrule "+tt.fact+" => "+tt.action+")")
		issueWithCode(t, r, tt.code)
	}
	r := lintPERText(t, "(defrule (up-compare-goal 112 > 0) => (set-strategic-number sn-focus-player-number my-player-number))")
	issueWithCode(t, r, "unresolved_constant")
	for _, c := range r.Checks {
		if c.Command == "up-compare-goal" || c.Command == "set-strategic-number" {
			if c.Source == "" || c.Confidence != "reference_backed" {
				t.Fatal(c)
			}
		}
	}
}
