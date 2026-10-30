package scenario

import (
	"path/filepath"
	"testing"
)

func TestAISignalRecipeRoundTrip(t *testing.T) {
	input := filepath.Join(t.TempDir(), "seed.aoe2scenario")
	output := filepath.Join(t.TempDir(), "signals.aoe2scenario")
	if _, err := WriteBlankScenarioFile(input, BlankOptions{PlayerCount: 2, HumanSlots: 1, NoStarters: true, ClearTriggers: true}); err != nil {
		t.Fatal(err)
	}
	recipe := Recipe{}
	for _, op := range []string{"ai_signal", "multiplayer_ai_signal", "ai_signal_multiplayer"} {
		recipe.Triggers = append(recipe.Triggers, TriggerRecipe{Op: "add_trigger", Name: op, Conditions: []ConditionRecipe{{Op: op, AISignal: intPtr(0), Inverted: intPtr(1)}}, Effects: []EffectRecipe{{Op: "script_call", Message: "Probe_Write();"}, {Op: "acknowledge_ai_signal", AISignalValue: intPtr(0)}, {Op: "acknowledge_multiplayer_ai_signal", AISignalValue: intPtr(7)}}})
	}
	report, err := PatchRecipeFile(input, output, recipe)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) != 3 {
		t.Fatalf("warnings = %v", report.Warnings)
	}
	file, err := Open(output)
	if err != nil {
		t.Fatal(err)
	}
	dumps, err := file.DumpTriggersWithConditions()
	if err != nil {
		t.Fatal(err)
	}
	if len(dumps) != 3 {
		t.Fatalf("triggers = %d", len(dumps))
	}
	for i, dump := range dumps {
		want := 30
		if i == 0 {
			want = 12
		}
		cond := dump.Conditions[0]
		wantName := "multiplayer_ai_signal"
		if i == 0 {
			wantName = "ai_signal"
		}
		if cond["type"] != want || cond["type_name"] != wantName || cond["ai_signal"] != 0 || cond["inverted"] != 1 {
			t.Fatalf("condition = %#v", cond)
		}
		effects := file.root.section("Triggers").list("trigger_data")[i].list("effect_data")
		if got, _ := effects[0].stringValue("message"); got != "Probe_Write();" {
			t.Fatalf("script call = %q", got)
		}
		for j, wantType := range []int{50, 69} {
			if got, _ := effects[j+1].intValue("effect_type"); got != wantType {
				t.Fatalf("effect type = %d", got)
			}
			wantSignal := 0
			if j == 1 {
				wantSignal = 7
			}
			if got, _ := effects[j+1].intValue("ai_signal_value"); got != wantSignal {
				t.Fatalf("signal = %d", got)
			}
		}
	}
}

func TestAISignalRecipeRequiredFields(t *testing.T) {
	for _, op := range []string{"ai_signal", "multiplayer_ai_signal"} {
		if _, err := conditionOverrides(ConditionRecipe{Op: op}); err == nil {
			t.Fatal("missing signal accepted")
		}
		if _, err := conditionOverrides(ConditionRecipe{Op: op, AISignal: intPtr(1), SourcePlayer: intPtr(1)}); err == nil {
			t.Fatal("player selector accepted")
		}
	}
	for _, op := range []string{"acknowledge_ai_signal", "acknowledge_multiplayer_ai_signal"} {
		if _, err := effectOverrides(EffectRecipe{Op: op}); err == nil {
			t.Fatal("missing acknowledgement signal accepted")
		}
	}
}
