package aifile

import (
	"strings"
	"testing"
)

func TestCoordinationShapesRetainMechanismAndUncertainty(t *testing.T) {
	r, err := LintPath("../../testdata/ai/coordination.per")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || len(r.Issues) != 0 || r.Coverage.UncheckedForms != 0 || r.Verification.EngineVerified {
		t.Fatalf("report=%+v", r)
	}
	checks := map[string]CommandCheck{}
	for _, check := range r.Checks {
		checks[check.Command] = check
	}
	for _, tt := range []struct {
		name, family, mechanism string
		arity                   int
	}{
		{"shared-goal", "AoC", "global_shared_goals", 2},
		{"set-shared-goal", "AoC", "global_shared_goals", 2},
		{"up-get-shared-goal", "UP", "global_shared_goals", 3},
		{"up-set-shared-goal", "UP", "global_shared_goals", 4},
		{"set-signal", "AoC", "legacy_scenario_signal", 1},
		{"up-get-signal", "UP", "legacy_scenario_signal", 3},
		{"up-set-signal", "UP", "legacy_scenario_signal", 4},
		{"fe-set-signal", "DE", "multiplayer_scenario_signal", 4},
		{"chat-to-allies", "AoC", "ally_chat", 1},
		{"up-chat-data-to-all", "UP", "all_player_chat", 3},
		{"up-chat-data-to-player", "UP", "recipient_chat", 4},
	} {
		c, found := checks[tt.name]
		if !found || len(c.Parameters) != tt.arity || c.ReferenceFamily != tt.family || c.Mechanism != tt.mechanism || c.RuntimeSemantics != "unverified" || c.Source == "" || c.Notes == "" {
			t.Errorf("%s check=%+v", tt.name, c)
		}
	}
	if !strings.Contains(checks["shared-goal"].Notes, "enemies") || !strings.Contains(checks["up-get-shared-goal"].Notes, "example uses 0") || !strings.Contains(checks["up-get-signal"].Notes, "not established") {
		t.Fatal("cross-channel or shared-state caveat lost")
	}
	if checks["up-chat-data-to-player"].Parameters[0].Constants["every-ally"].Value != -201 {
		t.Fatal("formatted ally recipient differs from reference")
	}
}

func TestCoordinationArityKindsAndRanges(t *testing.T) {
	for _, tt := range []struct {
		name, condition, action, code string
		ok                            bool
	}{
		{"shared equality", "(shared-goal 1 == 0)", "(disable-self)", "wrong_arity", false},
		{"shared lower bound", "(shared-goal 0 0)", "(disable-self)", "reference_range", true},
		{"shared upper bound", "(true)", "(set-shared-goal 257 1)", "reference_range", true},
		{"getter output", "(true)", "(up-get-shared-goal c: 1 \"goal\")", "argument_kind", false},
		{"shared example conflict", "(true)", "(up-get-shared-goal c: 0 100)", "reference_range", true},
		{"signal upper bound", "(true)", "(set-signal 256)", "reference_range", true},
		{"signal arity", "(true)", "(up-set-signal c: 1)", "wrong_arity", false},
		{"multiplayer operator", "(true)", "(fe-set-signal x: 1 c: 1)", "argument_kind", false},
		{"chat text", "(true)", "(chat-to-allies 42)", "argument_kind", false},
		{"formatted arity", "(true)", "(up-chat-data-to-player every-ally \"x\" c:)", "wrong_arity", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := lintPERText(t, "(defrule "+tt.condition+" => "+tt.action+")")
			issueWithCode(t, r, tt.code)
			if r.OK != tt.ok {
				t.Fatalf("OK=%v issues=%+v", r.OK, r.Issues)
			}
		})
	}
	r := lintPERText(t, "(defrule (true) => (set-signal 0) (up-get-shared-goal g: 1000 16000) (up-set-shared-goal s: 900 g: 100) (up-set-signal g: 999 s: 500) (fe-set-signal g: 999 c: 0))")
	if !r.OK || len(r.Issues) != 0 {
		t.Fatalf("signal zero/dynamic address mistaken for literal: %+v", r.Issues)
	}
}
