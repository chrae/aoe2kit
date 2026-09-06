package scenario

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"

	"aoe2kit/pkg/datfile"
	"aoe2kit/pkg/testfixtures"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLoadCurrentDESpec(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	if len(spec.Sections) == 0 {
		t.Fatal("expected sections")
	}
	if spec.Sections[0].Name != "FileHeader" {
		t.Fatalf("first section = %q, want FileHeader", spec.Sections[0].Name)
	}
	if _, ok := spec.section("Triggers"); !ok {
		t.Fatal("missing Triggers section")
	}
}

func TestScenarioVersionSupport(t *testing.T) {
	for _, version := range []string{"1.55", "1.56", "1.57", "1.58"} {
		if !SupportsReadVersion(version) {
			t.Fatalf("SupportsReadVersion(%q) = false", version)
		}
	}
	if SupportsReadVersion("1.54") {
		t.Fatal("SupportsReadVersion(1.54) = true")
	}
	if !SupportsWriteVersion("1.58") {
		t.Fatal("SupportsWriteVersion(1.58) = false")
	}
	if !SupportsWriteVersion("1.57") {
		t.Fatal("SupportsWriteVersion(1.57) = false")
	}
}

func TestBlankScenarioSeedIntegrity(t *testing.T) {
	data, err := BlankScenarioSeedBytes()
	if err != nil {
		t.Fatalf("BlankScenarioSeedBytes: %v", err)
	}
	if len(data) != 1029 {
		t.Fatalf("blank seed length = %d, want 1029", len(data))
	}
	file, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse blank seed: %v", err)
	}
	if file.Version != "1.58" {
		t.Fatalf("blank seed version = %q, want 1.58", file.Version)
	}
	if file.Triggers == nil || file.Triggers.Count != 3 {
		t.Fatalf("blank seed trigger count = %#v, want 3 editor seed triggers", file.Triggers)
	}
}

func TestWriteBlankScenarioFileClearsSeedAndAddsOutpostDummies(t *testing.T) {
	out := filepath.Join(t.TempDir(), "BlankRoot.aoe2scenario")
	report, err := WriteBlankScenarioFile(out, BlankOptions{
		PlayerCount:   4,
		HumanSlots:    1,
		Timestamp:     1800000000,
		DummyStarters: true,
	})
	if err != nil {
		t.Fatalf("WriteBlankScenarioFile: %v", err)
	}
	if report.TriggerCountBefore != 3 || report.TriggerCountAfter != 0 {
		t.Fatalf("trigger counts = %d -> %d, want 3 -> 0", report.TriggerCountBefore, report.TriggerCountAfter)
	}
	if report.UnitCountBefore != 0 || report.UnitCountAfter != 4 {
		t.Fatalf("unit counts = %d -> %d, want 0 -> 4", report.UnitCountBefore, report.UnitCountAfter)
	}
	if report.DummyUnit != defaultBlankDummyUnit {
		t.Fatalf("dummy unit = %d, want %d", report.DummyUnit, defaultBlankDummyUnit)
	}
	file, err := Open(out)
	if err != nil {
		t.Fatalf("Open blank output: %v", err)
	}
	if file.PlayerCount != 4 {
		t.Fatalf("player_count = %d, want 4", file.PlayerCount)
	}
	if file.Triggers == nil || file.Triggers.Count != 0 || !file.Triggers.InvariantOK {
		t.Fatalf("trigger invariant/count = %#v, want zero invariant-ok triggers", file.Triggers)
	}
	if file.Units == nil || file.Units.Total != 4 {
		t.Fatalf("unit total = %#v, want 4", file.Units)
	}
	for player := 1; player <= 4; player++ {
		section := file.Units.Sections[player]
		if section.Count != 1 {
			t.Fatalf("P%d unit count = %d, want 1", player, section.Count)
		}
		unit := section.Units[0]
		if unit.UnitConst != defaultBlankDummyUnit {
			t.Fatalf("P%d dummy unit = %d, want %d", player, unit.UnitConst, defaultBlankDummyUnit)
		}
		if unit.CaptionString != fmt.Sprintf("A2K_BLANK_DUMMY_P%d", player) {
			t.Fatalf("P%d caption = %q", player, unit.CaptionString)
		}
	}
	settings := file.Settings()
	if settings.Players[1].Active != true || settings.Players[1].Human != true {
		t.Fatalf("P1 settings = %+v, want active human", settings.Players[1])
	}
	if settings.Players[0].Active != false || settings.Players[0].Human != false {
		t.Fatalf("P0 settings = %+v, want inactive non-human Gaia for clean blank fixture", settings.Players[0])
	}
	if settings.Players[2].Active != true || settings.Players[2].Human != false {
		t.Fatalf("P2 settings = %+v, want active computer", settings.Players[2])
	}
	if settings.Players[5].Active != false {
		t.Fatalf("P5 active = true, want inactive")
	}
	if settings.Players[9].Active != false || settings.Players[9].Human != false {
		t.Fatalf("P9 settings = %+v, want inactive computer", settings.Players[9])
	}
	checkIntField(t, file.root.section("Units"), "number_of_players", 5)
	checkIntField(t, file.headerRoot, "timestamp_of_last_save", 1800000000)
}

func TestMaxInflatedScenarioBytes(t *testing.T) {
	t.Setenv("AOE2KIT_ALLOW_HUGE_SCENARIO", "")
	t.Setenv("AOE2KIT_MAX_SCENARIO_MB", "")
	if got := maxInflatedScenarioBytes(); got != defaultMaxInflatedScenarioBytes {
		t.Fatalf("default max = %d, want %d", got, defaultMaxInflatedScenarioBytes)
	}
	t.Setenv("AOE2KIT_MAX_SCENARIO_MB", "2")
	if got := maxInflatedScenarioBytes(); got != 2*1024*1024 {
		t.Fatalf("env max = %d, want 2 MiB", got)
	}
	t.Setenv("AOE2KIT_ALLOW_HUGE_SCENARIO", "1")
	if got := maxInflatedScenarioBytes(); got != 0 {
		t.Fatalf("allow huge max = %d, want 0", got)
	}
}

func TestInflateRawLimitedStopsAboveLimit(t *testing.T) {
	compressed, err := DeflateRaw([]byte(strings.Repeat("x", 128)))
	if err != nil {
		t.Fatalf("DeflateRaw: %v", err)
	}
	if _, err := inflateRawLimited(compressed, 64); err == nil || !strings.Contains(err.Error(), "safety limit") {
		t.Fatalf("inflateRawLimited error = %v, want safety limit", err)
	}
	body, err := inflateRawLimited(compressed, 128)
	if err != nil {
		t.Fatalf("inflateRawLimited exact limit: %v", err)
	}
	if len(body) != 128 {
		t.Fatalf("body len = %d, want 128", len(body))
	}
}

func TestDE155SpecUsesLegacyPlayerDataShape(t *testing.T) {
	spec, err := LoadDESpecForVersion("1.55")
	if err != nil {
		t.Fatalf("LoadDESpecForVersion: %v", err)
	}
	dataHeader, ok := spec.section("DataHeader")
	if !ok {
		t.Fatal("missing DataHeader")
	}
	playerData := dataHeader.Structs["PlayerDataOneStruct"]
	var fields []string
	for _, retriever := range playerData.Retrievers {
		fields = append(fields, retriever.Name+":"+retriever.Type)
	}
	got := strings.Join(fields, ",")
	want := "active:u32,human:u32,civilization:u32,architecture_set:u32,cty_mode:u32"
	if got != want {
		t.Fatalf("PlayerDataOneStruct = %s, want %s", got, want)
	}
}

func TestOlderDEEffectSpecsRemoveNewerFields(t *testing.T) {
	tests := []struct {
		version string
		absent  []string
	}{
		{
			version: "1.55",
			absent: []string{
				"issue_group_command",
				"queue_action",
				"mutual_diplomacy",
				"building_list",
				"wall_x1",
				"wall_y1",
				"wall_x2",
				"wall_y2",
				"object_filter",
				"use_tag_color_for_icon",
			},
		},
		{
			version: "1.56",
			absent: []string{
				"mutual_diplomacy",
				"building_list",
				"wall_x1",
				"wall_y1",
				"wall_x2",
				"wall_y2",
				"object_filter",
				"use_tag_color_for_icon",
			},
		},
		{
			version: "1.57",
			absent: []string{
				"object_filter",
				"use_tag_color_for_icon",
			},
		},
	}
	for _, test := range tests {
		spec, err := LoadDESpecForVersion(test.version)
		if err != nil {
			t.Fatalf("LoadDESpecForVersion(%s): %v", test.version, err)
		}
		triggers, ok := spec.section("Triggers")
		if !ok {
			t.Fatal("missing Triggers section")
		}
		effects := triggers.Structs["TriggerStruct"].Structs["EffectStruct"]
		fields := map[string]bool{}
		for _, retriever := range effects.Retrievers {
			fields[retriever.Name] = true
		}
		for _, name := range test.absent {
			if fields[name] {
				t.Fatalf("version %s still has EffectStruct.%s", test.version, name)
			}
		}
		if !fields["message"] {
			t.Fatalf("version %s missing EffectStruct.message", test.version)
		}
	}
}

func TestWriteRejectsReadOnlyScenarioVersions(t *testing.T) {
	out := filepath.Join(t.TempDir(), "readonly.aoe2scenario")
	err := (&File{Version: "1.56"}).Write(out)
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("Write error = %v, want read-only version error", err)
	}
}

func TestIntListNormalizesScalarPrimitive(t *testing.T) {
	node := &parsedNode{
		Fields: []*parsedNode{
			{Name: "selected_object_ids", Value: int32(12345)},
		},
	}
	got := node.intList("selected_object_ids")
	if len(got) != 1 || got[0] != 12345 {
		t.Fatalf("intList = %#v, want [12345]", got)
	}
}

func TestValidateOptionalOrder(t *testing.T) {
	if ok, note := validateOptionalOrder(nil, 3, "effect", 4); !ok || note != "" {
		t.Fatalf("empty optional order = %v %q, want ok", ok, note)
	}
	if ok, note := validateOptionalOrder([]int{2, 0, 1}, 3, "effect", 4); !ok || note != "" {
		t.Fatalf("valid optional order = %v %q, want ok", ok, note)
	}
	if ok, _ := validateOptionalOrder([]int{0, 0}, 2, "effect", 4); ok {
		t.Fatal("duplicate optional order accepted")
	}
	if ok, _ := validateOptionalOrder([]int{0, 2}, 2, "effect", 4); ok {
		t.Fatal("out-of-range optional order accepted")
	}
}

func TestBuildDisplayInstructionTriggerRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	raw, err := buildDisplayInstructionTrigger(spec, TriggerRecipe{
		Name:    "Smoke",
		Message: "hello",
	})
	if err != nil {
		t.Fatalf("buildDisplayInstructionTrigger: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	if p.off != len(raw) {
		t.Fatalf("parse consumed %d bytes, want %d", p.off, len(raw))
	}
	name, _ := node.stringValue("trigger_name")
	if name != "Smoke" {
		t.Fatalf("trigger_name = %q, want Smoke", name)
	}
	effects := node.list("effect_data")
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	effectType, _ := effects[0].intValue("effect_type")
	if effectType != 20 {
		t.Fatalf("effect_type = %d, want 20", effectType)
	}
	message, _ := effects[0].stringValue("message")
	if message != "hello" {
		t.Fatalf("message = %q, want hello", message)
	}
	assertLengthPrefixedStringRaw(t, node.field("trigger_name").Raw, "Smoke")
	assertLengthPrefixedStringRaw(t, effects[0].field("message").Raw, "hello")
}

func TestBuildCreateObjectTriggerRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	unitID := 83
	sourcePlayer := 0
	x := 12
	y := 34
	raw, err := buildCreateObjectTrigger(spec, TriggerRecipe{
		Name:             "Create Smoke",
		ObjectListUnitID: &unitID,
		SourcePlayer:     &sourcePlayer,
		LocationX:        &x,
		LocationY:        &y,
	})
	if err != nil {
		t.Fatalf("buildCreateObjectTrigger: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	effects := node.list("effect_data")
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	effect := effects[0]
	checkIntField(t, effect, "effect_type", 11)
	checkIntField(t, effect, "object_list_unit_id", unitID)
	checkIntField(t, effect, "source_player", sourcePlayer)
	checkIntField(t, effect, "location_x", x)
	checkIntField(t, effect, "location_y", y)
}

func checkIntField(t *testing.T, node *parsedNode, name string, want int) {
	t.Helper()
	got, ok := node.intValue(name)
	if !ok {
		t.Fatalf("%s missing", name)
	}
	if got != want {
		t.Fatalf("%s = %d, want %d", name, got, want)
	}
}

func checkFloatField(t *testing.T, node *parsedNode, name string, want float64) {
	t.Helper()
	got, ok := node.floatValue(name)
	if !ok {
		t.Fatalf("%s missing", name)
	}
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("%s = %.4f, want %.4f", name, got, want)
	}
}

func assertLengthPrefixedStringRaw(t *testing.T, raw []byte, want string) {
	t.Helper()
	wantPayload := want + "\x00"
	for _, prefixSize := range []int{2, 4} {
		if len(raw) < prefixSize {
			continue
		}
		if signedInt(raw[:prefixSize]) == len([]byte(wantPayload)) && string(raw[prefixSize:]) == wantPayload {
			return
		}
	}
	t.Fatalf("string raw = % x, want length=%d payload=%q", raw, len([]byte(wantPayload)), wantPayload)
}

func TestBuildGeneralTriggerWithConditionRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	timer := 5
	triggerID := 7
	enabled := true
	looping := true
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:      "add_trigger",
		Name:    "Timer Activate",
		Enabled: &enabled,
		Looping: &looping,
		Conditions: []ConditionRecipe{
			{Op: "timer", Timer: &timer},
		},
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &triggerID},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	checkIntField(t, node, "enabled", 1)
	checkIntField(t, node, "looping", 1)
	conditions := node.list("condition_data")
	if len(conditions) != 1 {
		t.Fatalf("conditions = %d, want 1", len(conditions))
	}
	checkIntField(t, conditions[0], "condition_type", 10)
	checkIntField(t, conditions[0], "timer", timer)
	effects := node.list("effect_data")
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 8)
	checkIntField(t, effects[0], "trigger_id", triggerID)
}

func TestBuildOwnershipConditionsRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	quantity := 3
	unitConst := 600
	player := 2
	areaX1 := 10
	areaY1 := 11
	areaX2 := 12
	areaY2 := 13
	objectState := 2
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Ownership Conditions",
		Conditions: []ConditionRecipe{
			{Op: "own_objects", Quantity: &quantity, ObjectList: &unitConst, SourcePlayer: &player},
			{Op: "own_fewer_objects", Quantity: &quantity, ObjectList: &unitConst, SourcePlayer: &player, AreaX1: &areaX1, AreaY1: &areaY1, AreaX2: &areaX2, AreaY2: &areaY2, ObjectState: &objectState},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	conditions := node.list("condition_data")
	if len(conditions) != 2 {
		t.Fatalf("conditions = %d, want 2", len(conditions))
	}
	checkIntField(t, conditions[0], "condition_type", 3)
	checkIntField(t, conditions[0], "quantity", quantity)
	checkIntField(t, conditions[0], "object_list", unitConst)
	checkIntField(t, conditions[0], "source_player", player)
	checkIntField(t, conditions[1], "condition_type", 4)
	checkIntField(t, conditions[1], "quantity", quantity)
	checkIntField(t, conditions[1], "object_list", unitConst)
	checkIntField(t, conditions[1], "source_player", player)
	checkIntField(t, conditions[1], "area_x1", areaX1)
	checkIntField(t, conditions[1], "area_y1", areaY1)
	checkIntField(t, conditions[1], "area_x2", areaX2)
	checkIntField(t, conditions[1], "area_y2", areaY2)
	checkIntField(t, conditions[1], "object_state", objectState)
}

func TestBuildDeclareVictoryTriggerRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	timer := 120
	player := 1
	enabled := 1
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Declare Victory",
		Conditions: []ConditionRecipe{
			{Op: "timer", Timer: &timer},
		},
		Effects: []EffectRecipe{
			{Op: "declare_victory", SourcePlayer: &player, Enabled: &enabled},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	conditions := node.list("condition_data")
	if len(conditions) != 1 {
		t.Fatalf("conditions = %d, want 1", len(conditions))
	}
	checkIntField(t, conditions[0], "condition_type", 10)
	checkIntField(t, conditions[0], "timer", timer)
	effects := node.list("effect_data")
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 13)
	checkIntField(t, effects[0], "source_player", player)
	checkIntField(t, effects[0], "enabled", enabled)
}

func TestBuildChatAndTimerEffectTriggersRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	sourcePlayer := 2
	displayTime := 45
	timeUnit := 1
	timerID := 7
	resetTimer := 1
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Chat and Timer",
		Effects: []EffectRecipe{
			{Op: "send_chat", SourcePlayer: &sourcePlayer, Message: "<GREEN>Hello"},
			{
				Op:           "display_timer",
				SourcePlayer: &sourcePlayer,
				DisplayTime:  &displayTime,
				TimeUnit:     &timeUnit,
				TimerID:      &timerID,
				ResetTimer:   &resetTimer,
				Message:      "Timer label",
			},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	effects := node.list("effect_data")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 3)
	checkIntField(t, effects[0], "source_player", sourcePlayer)
	message, _ := effects[0].stringValue("message")
	if message != "<GREEN>Hello" {
		t.Fatalf("chat message = %q, want <GREEN>Hello", message)
	}
	assertLengthPrefixedStringRaw(t, effects[0].field("message").Raw, "<GREEN>Hello")
	checkIntField(t, effects[1], "effect_type", 37)
	checkIntField(t, effects[1], "source_player", sourcePlayer)
	checkIntField(t, effects[1], "display_time", displayTime)
	checkIntField(t, effects[1], "time_unit", timeUnit)
	checkIntField(t, effects[1], "timer", timerID)
	checkIntField(t, effects[1], "reset_timer", resetTimer)
	message, _ = effects[1].stringValue("message")
	if message != "Timer label" {
		t.Fatalf("timer message = %q, want Timer label", message)
	}
	assertLengthPrefixedStringRaw(t, node.field("trigger_name").Raw, "Chat and Timer")
	assertLengthPrefixedStringRaw(t, effects[1].field("message").Raw, "Timer label")
	timerSummary := summarizeEffect(effects[1], EffectsOptions{})
	if timerSummary.TypeName != "display_timer" {
		t.Fatalf("timer type_name = %q, want display_timer", timerSummary.TypeName)
	}
	if len(timerSummary.RawFields) != 0 {
		t.Fatalf("default timer raw_fields len = %d, want 0", len(timerSummary.RawFields))
	}
	for key, want := range map[string]any{
		"source_player": sourcePlayer,
		"display_time":  displayTime,
		"time_unit":     timeUnit,
		"timer_id":      timerID,
		"reset_timer":   resetTimer,
		"message":       "Timer label",
	} {
		if got := timerSummary.KnownFields[key]; got != want {
			t.Fatalf("timer known_fields[%s] = %#v, want %#v", key, got, want)
		}
	}
	rawSummary := summarizeEffect(effects[1], EffectsOptions{IncludeRawFields: true})
	if len(rawSummary.RawFields) == 0 {
		t.Fatal("raw timer summary has empty raw_fields")
	}
}

func TestBuildVariableAndClearTimerEffectTriggersRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	variable := 9
	variable2 := 10
	quantity := 77
	operation := 1
	timerID := 3
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Variable and Timer Effects",
		Effects: []EffectRecipe{
			{Op: "modify_variable", Variable: &variable, Quantity: &quantity, Operation: &operation, Message: "set var"},
			{Op: "modify_variable_by_variable", Variable: &variable, Variable2: &variable2, Operation: &operation},
			{Op: "clear_timer", TimerID: &timerID},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	effects := node.list("effect_data")
	if len(effects) != 3 {
		t.Fatalf("effects = %d, want 3", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 56)
	checkIntField(t, effects[0], "variable", variable)
	checkIntField(t, effects[0], "quantity", quantity)
	checkIntField(t, effects[1], "effect_type", 100)
	checkIntField(t, effects[1], "variable", variable)
	checkIntField(t, effects[1], "variable2", variable2)
	checkIntField(t, effects[2], "effect_type", 57)
	checkIntField(t, effects[2], "timer", timerID)
}

func TestReadPrimitiveTruncatesStringsAtFirstNUL(t *testing.T) {
	payload := []byte("Visible\x00hidden")
	raw := make([]byte, 4+len(payload))
	writeSignedBytes(raw[:4], int64(len(payload)))
	copy(raw[4:], payload)
	p := parser{data: raw}
	got, gotRaw, err := p.readPrimitive("str32")
	if err != nil {
		t.Fatalf("read str32: %v", err)
	}
	if got != "Visible" {
		t.Fatalf("str32 = %q, want Visible", got)
	}
	if len(gotRaw) != len(raw) {
		t.Fatalf("raw len = %d, want %d", len(gotRaw), len(raw))
	}

	p = parser{data: []byte("Fixed\x00XX")}
	got, _, err = p.readPrimitive("c8")
	if err != nil {
		t.Fatalf("read c8: %v", err)
	}
	if got != "Fixed" {
		t.Fatalf("c8 = %q, want Fixed", got)
	}
}

func TestEncodePrimitiveStringsUseInclusiveNULTerminator(t *testing.T) {
	raw, err := encodePrimitive("str32", "SPAWN ROSTER BOARD")
	if err != nil {
		t.Fatalf("encodePrimitive: %v", err)
	}
	assertLengthPrefixedStringRaw(t, raw, "SPAWN ROSTER BOARD")

	raw, err = encodePrimitive("str16", "Clean\x00garbage")
	if err != nil {
		t.Fatalf("encodePrimitive embedded NUL: %v", err)
	}
	assertLengthPrefixedStringRaw(t, raw, "Clean")
}

func TestBuildDisplayInstructionTriggerPreservesUTF8MarkupRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	message := "<AQUA>~ WELCOME \u2605 ~<PURPLE>\u00c6ther caf\u00e9"
	raw, err := buildDisplayInstructionTrigger(spec, TriggerRecipe{
		Name:    "UTF8 Markup",
		Message: message,
	})
	if err != nil {
		t.Fatalf("buildDisplayInstructionTrigger: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	if p.off != len(raw) {
		t.Fatalf("parse consumed %d bytes, want %d", p.off, len(raw))
	}
	assertLengthPrefixedStringRaw(t, node.field("trigger_name").Raw, "UTF8 Markup")
	effects := node.list("effect_data")
	if len(effects) != 1 {
		t.Fatalf("effect count = %d, want 1", len(effects))
	}
	got, _ := effects[0].stringValue("message")
	if got != message {
		t.Fatalf("message = %q, want %q", got, message)
	}
	assertLengthPrefixedStringRaw(t, effects[0].field("message").Raw, message)
}

func TestLintCorruptStringsFlagsInvalidUTF8AndEmbeddedNUL(t *testing.T) {
	raw := make([]byte, 4+len("Visible\x00hidden"))
	writeSignedBytes(raw[:4], int64(len("Visible\x00hidden")))
	copy(raw[4:], "Visible\x00hidden")
	report := LintReport{}
	lintCorruptStrings(&report, &parsedNode{
		Name:  "effect.message",
		Start: 100,
		Raw:   raw,
		Value: "Visible",
	}, &parsedNode{
		Name:  "bad.utf8",
		Start: 200,
		Raw:   []byte{0xff, 0xfe},
		Value: string([]byte{0xff, 0xfe}),
	})
	if !hasIssueCode(report.Issues, "corrupt_string_embedded_nul") {
		t.Fatalf("missing embedded NUL issue: %+v", report.Issues)
	}
	if !hasIssueCode(report.Issues, "corrupt_string_invalid_utf8") {
		t.Fatalf("missing invalid UTF-8 issue: %+v", report.Issues)
	}
}

func TestBuildGeneralTriggerPrimitiveSetRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	unitObject := 1234
	triggerID := 3
	actionType := 1
	selectedA := 44
	selectedB := 45
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Primitive Set",
		Conditions: []ConditionRecipe{
			{Op: "object_selected", UnitObject: &unitObject},
		},
		Effects: []EffectRecipe{
			{Op: "kill_object", SelectedObjectIDs: []int{selectedA, selectedB}},
			{Op: "task_object", ActionType: &actionType, SelectedObjectIDs: []int{selectedA}},
			{Op: "deactivate_trigger", TriggerID: &triggerID},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	conditions := node.list("condition_data")
	if len(conditions) != 1 {
		t.Fatalf("conditions = %d, want 1", len(conditions))
	}
	checkIntField(t, conditions[0], "condition_type", 11)
	checkIntField(t, conditions[0], "unit_object", unitObject)
	effects := node.list("effect_data")
	if len(effects) != 3 {
		t.Fatalf("effects = %d, want 3", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 14)
	checkIntField(t, effects[0], "number_of_units_selected", 2)
	checkIntField(t, effects[1], "effect_type", 12)
	checkIntField(t, effects[1], "action_type", actionType)
	checkIntField(t, effects[1], "number_of_units_selected", 1)
	checkIntField(t, effects[2], "effect_type", 9)
	checkIntField(t, effects[2], "trigger_id", triggerID)
}

func TestSmokeRecipeShape(t *testing.T) {
	recipe := SmokeRecipe(SmokeRecipeOptions{
		Player:       2,
		PlayerSet:    true,
		UnitConst:    74,
		UnitConstSet: true,
		X:            20,
		XSet:         true,
		Y:            30,
		YSet:         true,
		TerrainID:    5,
		TerrainIDSet: true,
		Elevation:    1,
		ElevationSet: true,
		Layer:        -1,
		LayerSet:     true,
	})
	if len(recipe.Triggers) != 3 {
		t.Fatalf("smoke triggers = %d, want 3", len(recipe.Triggers))
	}
	for i, trigger := range recipe.Triggers {
		if trigger.Enabled == nil || *trigger.Enabled {
			t.Fatalf("smoke trigger %d enabled = %v, want disabled", i, trigger.Enabled)
		}
	}
	if len(recipe.Units) != 1 || recipe.Units[0].Player != 2 || recipe.Units[0].UnitConst != 74 {
		t.Fatalf("smoke units = %+v", recipe.Units)
	}
	if recipe.Units[0].X == nil || *recipe.Units[0].X != 21.5 || recipe.Units[0].Y == nil || *recipe.Units[0].Y != 31.5 {
		t.Fatalf("smoke unit location = x %v y %v", recipe.Units[0].X, recipe.Units[0].Y)
	}
	if len(recipe.Map) != 1 || recipe.Map[0].X1 != 20 || recipe.Map[0].Y1 != 30 || recipe.Map[0].X2 != 22 || recipe.Map[0].Y2 != 32 {
		t.Fatalf("smoke map = %+v", recipe.Map)
	}
	if recipe.Map[0].TerrainID == nil || *recipe.Map[0].TerrainID != 5 || recipe.Map[0].Elevation == nil || *recipe.Map[0].Elevation != 1 || recipe.Map[0].Layer == nil || *recipe.Map[0].Layer != -1 {
		t.Fatalf("smoke map fields = %+v", recipe.Map[0])
	}
}

func TestPlanOperationEnabledIsOnlySerializedWhenApplicable(t *testing.T) {
	disabled := false
	plan := Plan{Operations: []PlanOp{
		{Op: "add_trigger", Name: "disabled trigger", Enabled: &disabled},
		{Op: "set_resources", Name: "P1"},
	}}
	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"op":"add_trigger"`) || !strings.Contains(text, `"enabled":false`) {
		t.Fatalf("trigger enabled field was not preserved: %s", text)
	}
	if strings.Contains(text, `"op":"set_resources","name":"P1","enabled"`) {
		t.Fatalf("non-trigger operation serialized misleading enabled field: %s", text)
	}
}

func hasIssueCode(issues []LintIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func TestBuildAuthoringFrontierTriggerRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	tech := 408
	vis := 0
	unitConst := 83
	player := 1
	targetPlayer := 2
	x := 20
	y := 21
	attr := 0
	resource := 3
	quantity := 50
	operation := 1
	unitObject := 12345
	areaX1 := 10
	areaY1 := 11
	areaX2 := 12
	areaY2 := 13
	variable := 7
	comparison := 4
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Authoring Frontier",
		Conditions: []ConditionRecipe{
			{Op: "object_in_area", SourcePlayer: &player, ObjectList: &unitConst, AreaX1: &areaX1, AreaY1: &areaY1, AreaX2: &areaX2, AreaY2: &areaY2},
			{Op: "accumulate_attribute", SourcePlayer: &player, Attribute: &resource, Quantity: &quantity},
			{Op: "object_visible", UnitObject: &unitObject},
			{Op: "variable_value", Variable: &variable, Comparison: &comparison, Quantity: &quantity},
		},
		Effects: []EffectRecipe{
			{Op: "research_technology", SourcePlayer: &player, Technology: &tech},
			{Op: "reveal_map", SourcePlayer: &player, TargetPlayer: &targetPlayer, VisibilityState: &vis},
			{Op: "modify_attribute", SourcePlayer: &player, ObjectListUnitID: &unitConst, ObjectAttributes: &attr, Quantity: &quantity, Operation: &operation},
			{Op: "modify_resource", SourcePlayer: &player, Resource: &resource, Quantity: &quantity, Operation: &operation},
			{Op: "script_call", SourcePlayer: &player, Message: "main();"},
			{Op: "change_ownership", SourcePlayer: &player, TargetPlayer: &targetPlayer, SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_name", SourcePlayer: &player, Message: "Renamed", SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_description", SourcePlayer: &player, Message: "Described", SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_hp", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "teleport_object", SourcePlayer: &player, LocationX: &x, LocationY: &y, SelectedObjectIDs: []int{unitObject}},
			{Op: "remove_object", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	conditions := node.list("condition_data")
	if len(conditions) != 4 {
		t.Fatalf("conditions = %d, want 4", len(conditions))
	}
	checkIntField(t, conditions[0], "condition_type", 5)
	checkIntField(t, conditions[0], "object_list", unitConst)
	checkIntField(t, conditions[1], "condition_type", 8)
	checkIntField(t, conditions[1], "attribute", resource)
	checkIntField(t, conditions[2], "condition_type", 15)
	checkIntField(t, conditions[2], "unit_object", unitObject)
	checkIntField(t, conditions[3], "condition_type", 22)
	checkIntField(t, conditions[3], "variable", variable)
	checkIntField(t, conditions[3], "comparison", comparison)
	effects := node.list("effect_data")
	if len(effects) != 11 {
		t.Fatalf("effects = %d, want 11", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 2)
	checkIntField(t, effects[0], "technology", tech)
	checkIntField(t, effects[1], "effect_type", 41)
	checkIntField(t, effects[1], "visibility_state", vis)
	checkIntField(t, effects[2], "effect_type", 51)
	checkIntField(t, effects[2], "object_attributes", attr)
	checkIntField(t, effects[3], "effect_type", 52)
	checkIntField(t, effects[3], "tribute_list", resource)
	checkIntField(t, effects[4], "effect_type", 55)
	if got, _ := effects[4].stringValue("message"); got != "main();" {
		t.Fatalf("effects[4].message = %q, want main();", got)
	}
	checkIntField(t, effects[5], "effect_type", 18)
	checkIntField(t, effects[5], "target_player", targetPlayer)
	checkIntField(t, effects[6], "effect_type", 26)
	checkIntField(t, effects[7], "effect_type", 44)
	if got, _ := effects[7].stringValue("message"); got != "Described" {
		t.Fatalf("effects[7].message = %q, want Described", got)
	}
	checkIntField(t, effects[8], "effect_type", 27)
	checkIntField(t, effects[8], "quantity", quantity)
	checkIntField(t, effects[9], "effect_type", 35)
	checkIntField(t, effects[9], "location_x", x)
	checkIntField(t, effects[10], "effect_type", 15)
	checkIntField(t, effects[10], "number_of_units_selected", 1)
	summary := summarizeEffect(effects[10], EffectsOptions{})
	if len(summary.SelectedObjectIDs) != 1 || summary.SelectedObjectIDs[0] != unitObject {
		t.Fatalf("selected_object_ids summary = %#v, want [%d]", summary.SelectedObjectIDs, unitObject)
	}
}

func TestBuildExpandedAuthoringEffectsRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	player := 1
	targetPlayer := 2
	diplomacy := 3
	mutualDiplomacy := 1
	unitConst := 83
	tech := 408
	x := 20
	y := 21
	scroll := 1
	enabled := 0
	quantity := 25
	operation := 1
	unitObject := 12345
	localTech := 1
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Expanded Authoring Effects",
		Effects: []EffectRecipe{
			{Op: "change_diplomacy", SourcePlayer: &player, TargetPlayer: &targetPlayer, Diplomacy: &diplomacy, MutualDiplomacy: &mutualDiplomacy},
			{Op: "play_sound", SourcePlayer: &player, SoundName: "ui\\default"},
			{Op: "change_view", SourcePlayer: &player, LocationX: &x, LocationY: &y, Scroll: &scroll},
			{Op: "place_foundation", SourcePlayer: &player, ObjectListUnitID: &unitConst, LocationX: &x, LocationY: &y},
			{Op: "build_object", SourcePlayer: &player, ObjectListUnitID: &unitConst, LocationX: &x, LocationY: &y},
			{Op: "damage_object", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "heal_object", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_attack", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_armor", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_range", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_speed", SourcePlayer: &player, Quantity: &quantity, Operation: &operation, SelectedObjectIDs: []int{unitObject}},
			{Op: "change_object_caption", SourcePlayer: &player, Message: "Caption", SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_disable_object", SourcePlayer: &player, Enabled: &enabled, SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_disable_technology", SourcePlayer: &player, Technology: &tech, Enabled: &enabled},
			{Op: "train_unit", SourcePlayer: &player, ObjectListUnitID: &unitConst, SelectedObjectIDs: []int{unitObject}},
			{Op: "initiate_research", SourcePlayer: &player, Technology: &tech, SelectedObjectIDs: []int{unitObject}},
			{Op: "research_local_technology", SourcePlayer: &player, Technology: &tech, LocalTechnology: &localTech, SelectedObjectIDs: []int{unitObject}},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	effects := node.list("effect_data")
	wantTypes := []int{1, 4, 16, 25, 108, 24, 34, 28, 31, 32, 33, 88, 38, 39, 75, 76, 103}
	if len(effects) != len(wantTypes) {
		t.Fatalf("effects = %d, want %d", len(effects), len(wantTypes))
	}
	for i, want := range wantTypes {
		checkIntField(t, effects[i], "effect_type", want)
	}
	checkIntField(t, effects[0], "diplomacy", diplomacy)
	checkIntField(t, effects[0], "mutual_diplomacy", mutualDiplomacy)
	if got, _ := effects[1].stringValue("sound_name"); got != "ui\\default" {
		t.Fatalf("sound_name = %q, want ui\\default", got)
	}
	checkIntField(t, effects[2], "scroll", scroll)
	checkIntField(t, effects[3], "object_list_unit_id", unitConst)
	checkIntField(t, effects[4], "object_list_unit_id", unitConst)
	checkIntField(t, effects[5], "quantity", quantity)
	checkIntField(t, effects[11], "number_of_units_selected", 1)
	if got, _ := effects[11].stringValue("message"); got != "Caption" {
		t.Fatalf("caption message = %q, want Caption", got)
	}
	checkIntField(t, effects[12], "enabled", enabled)
	checkIntField(t, effects[13], "technology", tech)
	checkIntField(t, effects[13], "enabled", enabled)
	checkIntField(t, effects[14], "object_list_unit_id", unitConst)
	checkIntField(t, effects[15], "technology", tech)
	checkIntField(t, effects[16], "local_technology", localTech)
}

func TestBuildShopTechAndControlEffectsRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	player := 1
	unitConst := 83
	buildingConst := 12
	tech := 408
	button := 3
	hotkey := 16121
	trainTime := 7
	enabled := 0
	color := 4
	unitObject := 12345
	food := 0
	foodCost := 11
	gold := 3
	goldCost := 22
	icon := 77
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Shop Tech And Control Effects",
		Effects: []EffectRecipe{
			{Op: "change_train_location", SourcePlayer: &player, ObjectListUnitID: &unitConst, ObjectListUnitID2: &buildingConst, ButtonLocation: &button, Hotkey: &hotkey, TrainTime: &trainTime},
			{Op: "add_train_location", SourcePlayer: &player, ObjectListUnitID: &unitConst, ObjectListUnitID2: &buildingConst, ButtonLocation: &button},
			{Op: "change_technology_location", SourcePlayer: &player, Technology: &tech, ObjectListUnitID: &buildingConst, ButtonLocation: &button},
			{Op: "change_technology_cost", SourcePlayer: &player, Technology: &tech, Resource1: &food, Resource1Quantity: &foodCost, Resource2: &gold, Resource2Quantity: &goldCost},
			{Op: "change_technology_research_time", SourcePlayer: &player, Technology: &tech, TrainTime: &trainTime},
			{Op: "change_technology_name", SourcePlayer: &player, Technology: &tech, Message: "Tech Name"},
			{Op: "change_technology_description", SourcePlayer: &player, Technology: &tech, Message: "Tech Description"},
			{Op: "change_technology_icon", SourcePlayer: &player, Technology: &tech, Quantity: &icon},
			{Op: "change_technology_hotkey", SourcePlayer: &player, Technology: &tech, Hotkey: &hotkey},
			{Op: "change_player_name", SourcePlayer: &player, Message: "Player Name"},
			{Op: "change_civilization_name", SourcePlayer: &player, Message: "Civilization Name"},
			{Op: "change_player_color", SourcePlayer: &player, PlayerColor: &color},
			{Op: "freeze_unit", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "stop_unit", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "disable_unit_targeting", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_unit_targeting", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "disable_object_selection", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_object_selection", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_object_deletion", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "disable_object_deletion", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "disable_unit_attackable", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_unit_attackable", SourcePlayer: &player, SelectedObjectIDs: []int{unitObject}},
			{Op: "enable_disable_technology", SourcePlayer: &player, Technology: &tech, Enabled: &enabled},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	effects := node.list("effect_data")
	wantTypes := []int{46, 102, 47, 63, 64, 65, 66, 84, 85, 45, 48, 89, 22, 29, 61, 62, 70, 71, 73, 74, 98, 99, 39}
	if len(effects) != len(wantTypes) {
		t.Fatalf("effects = %d, want %d", len(effects), len(wantTypes))
	}
	for i, want := range wantTypes {
		checkIntField(t, effects[i], "effect_type", want)
	}
	checkIntField(t, effects[0], "object_list_unit_id", unitConst)
	checkIntField(t, effects[0], "object_list_unit_id_2", buildingConst)
	checkIntField(t, effects[0], "button_location", button)
	checkIntField(t, effects[0], "hotkey", hotkey)
	checkIntField(t, effects[0], "train_time", trainTime)
	checkIntField(t, effects[2], "technology", tech)
	checkIntField(t, effects[2], "button_location", button)
	checkIntField(t, effects[3], "resource_1", food)
	checkIntField(t, effects[3], "resource_1_quantity", foodCost)
	checkIntField(t, effects[3], "resource_2", gold)
	checkIntField(t, effects[3], "resource_2_quantity", goldCost)
	checkIntField(t, effects[4], "quantity", trainTime)
	if got, _ := effects[5].stringValue("message"); got != "Tech Name" {
		t.Fatalf("tech name message = %q, want Tech Name", got)
	}
	checkIntField(t, effects[7], "quantity", icon)
	checkIntField(t, effects[8], "hotkey", hotkey)
	if got, _ := effects[9].stringValue("message"); got != "Player Name" {
		t.Fatalf("player name message = %q, want Player Name", got)
	}
	checkIntField(t, effects[11], "player_color", color)
	checkIntField(t, effects[12], "number_of_units_selected", 1)
	checkIntField(t, effects[22], "enabled", enabled)
}

func TestPatchRecipeScenarioSettingsAndXS(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "xs_probe.aoe2scenario")
	playerCount := 4
	timestamp := 1788336000
	active := true
	human := false
	food := 888888
	recipe := Recipe{
		Scenario: &ScenarioRecipe{PlayerCount: &playerCount, TimestampOfLastSave: &timestamp},
		XS:       &XSRecipe{Name: "TestProbe", Content: "rule TestProbe active minInterval 5 { xsChatData(\"SDSTELEM\"); }\n"},
		Players: []PlayerRecipe{
			{Player: 2, Active: &active, Human: &human},
		},
		Diplomacy: []DiplomacyRecipe{
			{From: 1, To: 2, Stance: 3},
		},
		Resources: []ResourceRecipe{
			{Player: 1, Food: &food},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	mapSection := scen.root.section("Map")
	if got, _ := mapSection.stringValue("script_name"); got != "TestProbe.xs" {
		t.Fatalf("script_name = %q, want TestProbe.xs", got)
	}
	files := scen.root.section("Files")
	if got, _ := files.stringValue("script_file_path"); got != "TestProbe.xs" {
		t.Fatalf("script_file_path = %q, want TestProbe.xs", got)
	}
	if got, _ := files.stringValue("script_file_content"); got == "" {
		t.Fatal("script_file_content is empty")
	}
	checkIntField(t, scen.headerRoot, "player_count", playerCount)
	checkIntField(t, scen.headerRoot, "timestamp_of_last_save", timestamp)
	players := scen.root.section("DataHeader").list("player_data_1")
	checkIntField(t, players[2], "active", 1)
	checkIntField(t, players[2], "human", 0)
	resources := scen.root.section("PlayerDataTwo").list("resources")
	checkIntField(t, resources[1], "food", food)
	rows := scen.root.section("Diplomacy").list("per_player_diplomacy")
	stances := rows[1].uint32List("stance_with_each_player")
	if stances[2] != 3 {
		t.Fatalf("P1->P2 stance = %d, want 3", stances[2])
	}
}

func TestPatchRecipeFileRefreshesTimestampWhenRecipeDoesNotSpecifyOne(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "old.aoe2scenario")
	if _, err := WriteBlankScenarioFile(input, BlankOptions{PlayerCount: 2, HumanSlots: 1, Timestamp: 1, DummyStarters: true}); err != nil {
		t.Fatalf("WriteBlankScenarioFile: %v", err)
	}
	out := filepath.Join(dir, "patched.aoe2scenario")
	before := int(time.Now().Unix()) - 1
	playerCount := 3
	report, err := PatchRecipeFile(input, out, Recipe{Scenario: &ScenarioRecipe{PlayerCount: &playerCount}})
	if err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	if report.TimestampOfLastSave < before {
		t.Fatalf("report timestamp = %d, want >= %d", report.TimestampOfLastSave, before)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	got, ok := scen.headerRoot.intValue("timestamp_of_last_save")
	if !ok {
		t.Fatal("timestamp_of_last_save missing")
	}
	if got != report.TimestampOfLastSave {
		t.Fatalf("written timestamp = %d, report = %d", got, report.TimestampOfLastSave)
	}
	if got < before {
		t.Fatalf("written timestamp = %d, want >= %d", got, before)
	}
}

func TestPatchRecipeXSCarrierMode(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "xs_carrier.aoe2scenario")
	enabled := true
	recipe := Recipe{
		XS: &XSRecipe{
			Mode:                "carrier",
			CarrierTitle:        "XS string",
			CarrierTriggerIndex: intPtr(0),
			Content:             "const int PROBE = 22;\nvoid BootProbe() { xsChatData(\"A2K\"); }\n",
		},
		Triggers: []TriggerRecipe{{
			Op:      "add_trigger",
			Name:    "Boot Probe",
			Enabled: &enabled,
			Effects: []EffectRecipe{{
				Op:      "script_call",
				Message: "BootProbe();",
			}},
		}},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	mapSection := scen.root.section("Map")
	if got, _ := mapSection.stringValue("script_name"); got != "" {
		t.Fatalf("script_name = %q, want empty carrier mode", got)
	}
	files := scen.root.section("Files")
	if got, _ := files.stringValue("script_file_path"); got != "" {
		t.Fatalf("script_file_path = %q, want empty carrier mode", got)
	}
	if got, _ := files.stringValue("script_file_content"); got != "" {
		t.Fatalf("script_file_content = %q, want empty carrier mode", got)
	}
	report, err := scen.XSCensus(XSCensusOptions{})
	if err != nil {
		t.Fatalf("XSCensus: %v", err)
	}
	if report.Counts.EmbeddedCarriers != 1 || report.Counts.ScriptCalls != 1 || report.Counts.Functions != 1 {
		t.Fatalf("counts = %+v", report.Counts)
	}
	if len(report.EmbeddedCarriers) != 1 || report.EmbeddedCarriers[0].Title != "XS string" {
		t.Fatalf("carriers = %+v", report.EmbeddedCarriers)
	}
	if report.EmbeddedCarriers[0].TriggerIndex != 0 {
		t.Fatalf("carrier trigger index = %d, want 0", report.EmbeddedCarriers[0].TriggerIndex)
	}
	if report.EmbeddedCarriers[0].Content != "const int PROBE = 22;\nvoid BootProbe() { xsChatData(\"A2K\"); }\n" {
		t.Fatalf("carrier content = %q", report.EmbeddedCarriers[0].Content)
	}
	if len(report.ScriptCalls) != 1 || report.ScriptCalls[0].Message != "BootProbe();" {
		t.Fatalf("script calls = %+v", report.ScriptCalls)
	}
}

func TestXSCensusClassifiesParserStyleCarrier(t *testing.T) {
	carrier := xsCarrierMessage("XS string", "const int PROBE = 22;\nvoid BootProbe() {}\n")
	scen := scenarioWithXSFields("", "", "")
	scen.Version = "1.58"
	scen.Triggers = &TriggerInfo{Triggers: []TriggerSummary{{
		Index:   0,
		Name:    "XS SCRIPT",
		Enabled: 0,
		Looping: 0,
		EffectData: []EffectSummary{{
			EffectIndex: 0,
			Type:        55,
			TypeName:    "script_call",
			Text:        carrier,
		}},
	}, {
		Index: 1,
		Name:  "Boot",
		EffectData: []EffectSummary{{
			EffectIndex: 0,
			Type:        55,
			TypeName:    "script_call",
			Text:        "BootProbe();",
		}},
	}}}
	report, err := scen.XSCensus(XSCensusOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if report.Counts.EmbeddedCarriers != 1 || report.Counts.ScriptCalls != 1 || report.Counts.Functions != 1 {
		t.Fatalf("counts = %+v", report.Counts)
	}
	if report.EmbeddedCarriers[0].SourceBytes != len([]byte(carrier)) || report.EmbeddedCarriers[0].SourceSHA256 == "" {
		t.Fatalf("carrier summary = %+v", report.EmbeddedCarriers[0])
	}
	if report.EmbeddedCarriers[0].Content != "const int PROBE = 22;\nvoid BootProbe() {}\n" || report.EmbeddedCarriers[0].ContentSHA256 == "" {
		t.Fatalf("carrier content summary = %+v", report.EmbeddedCarriers[0])
	}
	if report.ScriptCalls[0].Message != "BootProbe();" {
		t.Fatalf("ordinary call not preserved: %+v", report.ScriptCalls)
	}
}

func TestTerrainFileAggregatesAndFilteredTiles(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	report, err := TerrainFile(input, TerrainOptions{})
	if err != nil {
		t.Fatalf("TerrainFile: %v", err)
	}
	if report.Width <= 0 || report.Height <= 0 || report.TileCount != report.Width*report.Height {
		t.Fatalf("bad terrain dimensions: %+v", report)
	}
	if len(report.Tiles) != 0 {
		t.Fatalf("unfiltered terrain report returned %d tiles without --tiles", len(report.Tiles))
	}
	if len(report.Aggregates) == 0 {
		t.Fatal("terrain aggregates are empty")
	}
	id := report.Aggregates[0].TerrainID
	filtered, err := TerrainFile(input, TerrainOptions{ID: &id})
	if err != nil {
		t.Fatalf("TerrainFile filtered: %v", err)
	}
	if filtered.Returned != filtered.Aggregates[0].Count || len(filtered.Tiles) != filtered.Returned {
		t.Fatalf("filtered returned/count mismatch: returned=%d tiles=%d aggregate=%d", filtered.Returned, len(filtered.Tiles), filtered.Aggregates[0].Count)
	}
}

func TestScenarioUnitsReadRotationFromFixture(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Mandala Foundry.aoe2scenario")
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open Mandala: %v", err)
	}
	found := false
	for _, section := range file.Units.Sections {
		for _, unit := range section.Units {
			if unit.UnitConst == 1777 && unit.Rotation == 5 {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected Mandala fixture to contain IndianStatues with rotation index 5")
	}
}

func TestPaletteUsageReportsUnusedAddressableIndices(t *testing.T) {
	file := &File{
		Units: &UnitInfo{
			Sections: []PlayerUnitsInfo{
				{
					Player: 0,
					Units: []UnitSummary{
						{UnitConst: 1777, Rotation: 0},
						{UnitConst: 1777, Rotation: 5},
						{UnitConst: 1777, Rotation: 11},
					},
				},
			},
		},
	}
	variantCount := 16
	report := file.PaletteUsage(datfile.PaletteReport{Rows: []datfile.PaletteRow{
		{
			UnitID:           1777,
			UnitName:         "INDIANSTATUES",
			StandingGraphic1: 12530,
			GraphicName:      "IndianStatues",
			AngleCount:       16,
			FrameCount:       1,
			VariantCount:     &variantCount,
			Classification:   "multi_variant",
			Confidence:       "engine_measured_one_case",
		},
	}})
	if len(report.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got, want := row.UsedIndices, []int{0, 5, 11}; !sameInts(got, want) {
		t.Fatalf("used indices = %v, want %v", got, want)
	}
	if len(row.UnusedIndices) != 13 || row.UnusedIndices[0] != 1 || row.UnusedIndices[len(row.UnusedIndices)-1] != 15 {
		t.Fatalf("unused indices = %v, want 13 missing including 1..15", row.UnusedIndices)
	}
	if report.Summary.MultiVariantTypeCount != 1 || report.Summary.ArtworksAvailable != 16 || report.Summary.ArtworksUsed != 3 {
		t.Fatalf("summary = %+v, want 1 type / 16 available / 3 used", report.Summary)
	}
	if len(report.Summary.BiggestUntapped) != 1 || report.Summary.BiggestUntapped[0].UnusedVariantCount != 13 {
		t.Fatalf("biggest untapped = %+v, want one row with 13 unused", report.Summary.BiggestUntapped)
	}
}

func TestPaletteUsageAnimatedSharedPhaseIsNeutral(t *testing.T) {
	file := &File{
		Units: &UnitInfo{
			Sections: []PlayerUnitsInfo{
				{
					Player: 0,
					Units: []UnitSummary{
						{UnitConst: 455, Rotation: 7},
						{UnitConst: 455, Rotation: 7},
					},
				},
			},
		},
	}
	report := file.PaletteUsage(datfile.PaletteReport{Rows: []datfile.PaletteRow{
		{
			UnitID:         455,
			UnitName:       "FISH1",
			GraphicName:    "Fish Dorado",
			AngleCount:     15,
			FrameCount:     1,
			SequenceType:   6,
			Classification: "animated",
			Confidence:     "author_confirmed",
		},
	}})
	if len(report.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if row.AvailableVariantCount != nil || len(row.UsedIndices) != 0 || len(row.UnusedIndices) != 0 {
		t.Fatalf("animated row should not be treated as variant artwork: %+v", row)
	}
	if !strings.Contains(row.Note, "phase distribution: 2 placements across 1 distinct raw rotation value") {
		t.Fatalf("note = %q, want neutral phase distribution", row.Note)
	}
	if strings.Contains(row.Note, "varying rotation") || strings.Contains(row.Note, "under-use") {
		t.Fatalf("note = %q, should not recommend variation or imply under-use", row.Note)
	}
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPatchRecipeGlobalVictory(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "global_victory.aoe2scenario")
	conquest := 0
	allCustom := 1
	score := 0
	timer := 0
	recipe := Recipe{
		Victory: &VictoryRecipe{
			ConquestRequired:               &conquest,
			AllCustomConditionsRequired:    &allCustom,
			RequiredScoreForScoreVictory:   &score,
			TimeForTimedGameIn10thsOfAYear: &timer,
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	victory := scen.root.section("GlobalVictory")
	if victory == nil {
		t.Fatal("missing GlobalVictory")
	}
	checkIntField(t, victory, "conquest_required", conquest)
	checkIntField(t, victory, "all_custom_conditions_required", allCustom)
	checkIntField(t, victory, "required_score_for_score_victory", score)
	checkIntField(t, victory, "time_for_timed_game_in_10ths_of_a_year", timer)
}

func TestPatchRecipeClearTriggers(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "clear_triggers.aoe2scenario")
	timer := 3
	enabled := true
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Only New Trigger",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timer},
				},
				Effects: []EffectRecipe{
					{Op: "display_instructions", Message: "new only"},
				},
			},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	if scen.Triggers.Count != 1 {
		t.Fatalf("trigger count = %d, want 1", scen.Triggers.Count)
	}
	if len(scen.Triggers.DisplayOrder) != 1 || scen.Triggers.DisplayOrder[0] != 0 {
		t.Fatalf("display order = %#v, want [0]", scen.Triggers.DisplayOrder)
	}
	triggers := scen.root.section("Triggers")
	nodes := triggers.list("trigger_data")
	if len(nodes) != 1 {
		t.Fatalf("trigger nodes = %d, want 1", len(nodes))
	}
	if got, _ := nodes[0].stringValue("trigger_name"); got != "Only New Trigger" {
		t.Fatalf("trigger name = %q, want Only New Trigger", got)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
}

func TestRemoveTriggerRewritesSurvivingReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "remove_rewrite_refs.aoe2scenario")
	enabled := false
	targetTwo := 2
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Refers To T2",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &targetTwo},
				},
			},
			{Op: "add_trigger", Name: "Remove Me", Enabled: &enabled},
			{Op: "add_trigger", Name: "Survives", Enabled: &enabled},
			{Op: "remove_trigger", TargetName: "Remove Me"},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	if scen.Triggers.Count != 2 {
		t.Fatalf("trigger count = %d, want 2", scen.Triggers.Count)
	}
	triggers := scen.root.section("Triggers").list("trigger_data")
	effects := triggers[0].list("effect_data")
	checkIntField(t, effects[0], "trigger_id", 1)
	name, _ := triggers[1].stringValue("trigger_name")
	if name != "Survives" {
		t.Fatalf("trigger 1 name = %q, want Survives", name)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
}

func TestRemoveTriggersAllowsInternalRefsAndRewritesSurvivors(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "remove_batch_refs.aoe2scenario")
	enabled := false
	internalTarget := 2
	survivorTarget := 4
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Keeper",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &survivorTarget},
				},
			},
			{
				Op:      "add_trigger",
				Name:    "Batch Source",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &internalTarget},
				},
			},
			{Op: "add_trigger", Name: "Batch Target", Enabled: &enabled},
			{Op: "add_trigger", Name: "Batch Spare", Enabled: &enabled},
			{Op: "add_trigger", Name: "Survives", Enabled: &enabled},
			{Op: "remove_triggers", TargetIndexes: []int{3, 1, 2}},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	if scen.Triggers.Count != 2 {
		t.Fatalf("trigger count = %d, want 2", scen.Triggers.Count)
	}
	triggers := scen.root.section("Triggers").list("trigger_data")
	effects := triggers[0].list("effect_data")
	checkIntField(t, effects[0], "trigger_id", 1)
	name, _ := triggers[1].stringValue("trigger_name")
	if name != "Survives" {
		t.Fatalf("trigger 1 name = %q, want Survives", name)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
}

func TestRemoveTriggersRefusesExternalReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "remove_batch_external_ref.aoe2scenario")
	enabled := false
	targetOne := 1
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "External Ref",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &targetOne},
				},
			},
			{Op: "add_trigger", Name: "Batch Target", Enabled: &enabled},
			{Op: "remove_triggers", TargetIndexes: []int{1}},
		},
	}
	err := func() error {
		_, err := PatchRecipeFile(input, out, recipe)
		return err
	}()
	if err == nil || !strings.Contains(err.Error(), "references removed trigger 1") {
		t.Fatalf("remove_triggers external ref error = %v, want refusal", err)
	}
}

func TestRemoveTriggerRefusesReferencedTarget(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "remove_referenced_target.aoe2scenario")
	enabled := false
	targetOne := 1
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Refers To Removed",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &targetOne},
				},
			},
			{Op: "add_trigger", Name: "Referenced Target", Enabled: &enabled},
			{Op: "remove_trigger", TargetName: "Referenced Target"},
		},
	}
	err := func() error {
		_, err := PatchRecipeFile(input, out, recipe)
		return err
	}()
	if err == nil || !strings.Contains(err.Error(), "refuses to remove trigger 1") {
		t.Fatalf("remove referenced trigger error = %v, want refuse message", err)
	}
}

func TestTombstoneTriggerPreservesReferencedIndex(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "tombstone_referenced_target.aoe2scenario")
	enabled := true
	targetOne := 1
	timer := 5
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Activator",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &targetOne},
				},
			},
			{
				Op:      "add_trigger",
				Name:    "Referenced Behavior",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timer},
				},
				Effects: []EffectRecipe{
					{Op: "send_chat", SourcePlayer: &targetOne, Message: "should be cleared"},
				},
			},
			{Op: "tombstone_trigger", TargetName: "Referenced Behavior"},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	if scen.Triggers.Count != 2 {
		t.Fatalf("trigger count = %d, want 2", scen.Triggers.Count)
	}
	triggers := scen.root.section("Triggers").list("trigger_data")
	effects := triggers[0].list("effect_data")
	checkIntField(t, effects[0], "trigger_id", 1)
	checkIntField(t, triggers[1], "enabled", 0)
	checkIntField(t, triggers[1], "number_of_effects", 0)
	checkIntField(t, triggers[1], "number_of_conditions", 0)
	name, _ := triggers[1].stringValue("trigger_name")
	if !strings.Contains(name, "TOMBSTONED") {
		t.Fatalf("tombstone trigger name = %q, want marker", name)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
}

func TestPatchRecipeEditTriggerRemoveChildren(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "edit_trigger_children.aoe2scenario")
	enabled := false
	timerA := 3
	timerB := 7
	triggerZero := 0
	removeEffects := []int{0}
	removeConditions := []int{1}
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Child Surgery",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timerA},
					{Op: "timer", Timer: &timerB},
				},
				Effects: []EffectRecipe{
					{Op: "display_instructions", Message: "remove me"},
					{Op: "send_chat", Message: "keep me"},
				},
			},
			{
				Op:               "edit_trigger",
				TargetName:       "Child Surgery",
				RemoveEffects:    removeEffects,
				RemoveConditions: removeConditions,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &triggerZero},
				},
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timerB},
				},
			},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
	trigger := scen.root.section("Triggers").list("trigger_data")[0]
	effects := trigger.list("effect_data")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 3)
	checkIntField(t, effects[1], "effect_type", 8)
	if got := trigger.intList("effect_display_order_array"); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("effect order = %#v, want [0 1]", got)
	}
	conditions := trigger.list("condition_data")
	if len(conditions) != 2 {
		t.Fatalf("conditions = %d, want 2", len(conditions))
	}
	checkIntField(t, conditions[0], "timer", timerA)
	checkIntField(t, conditions[1], "timer", timerB)
	if got := trigger.intList("condition_display_order_array"); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("condition order = %#v, want [0 1]", got)
	}
}

func TestPatchRecipeCopyTriggerCanRenameAndEditClone(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "copy_trigger.aoe2scenario")
	enabled := false
	timerA := 3
	timerB := 7
	removeEffects := []int{0}
	newName := "Copied Child Surgery"
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Child Surgery",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timerA},
				},
				Effects: []EffectRecipe{
					{Op: "display_instructions", Message: "remove me"},
					{Op: "send_chat", Message: "keep me"},
				},
			},
			{
				Op:            "copy_trigger",
				TargetName:    "Child Surgery",
				SetName:       &newName,
				RemoveEffects: removeEffects,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timerB},
				},
			},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
	triggers := scen.root.section("Triggers").list("trigger_data")
	if len(triggers) != 2 {
		t.Fatalf("triggers = %d, want 2", len(triggers))
	}
	clone := triggers[1]
	if got, _ := clone.stringValue("trigger_name"); got != newName {
		t.Fatalf("clone trigger_name = %q, want %q", got, newName)
	}
	checkIntField(t, clone, "enabled", 0)
	if got := len(clone.list("effect_data")); got != 1 {
		t.Fatalf("clone effects = %d, want 1", got)
	}
	checkIntField(t, clone.list("effect_data")[0], "effect_type", 3)
	if got := len(clone.list("condition_data")); got != 2 {
		t.Fatalf("clone conditions = %d, want 2", got)
	}
	checkIntField(t, clone.list("condition_data")[0], "timer", timerA)
	checkIntField(t, clone.list("condition_data")[1], "timer", timerB)
	if got := clone.intList("effect_display_order_array"); len(got) != 1 || got[0] != 0 {
		t.Fatalf("clone effect order = %#v, want [0]", got)
	}
	if got := clone.intList("condition_display_order_array"); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("clone condition order = %#v, want [0 1]", got)
	}
}

func TestPatchRecipeEditTriggerClearChildren(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "clear_trigger_children.aoe2scenario")
	enabled := false
	clear := true
	timer := 3
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Clear Child Lists",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timer},
				},
				Effects: []EffectRecipe{
					{Op: "display_instructions", Message: "clear me"},
				},
			},
			{
				Op:              "edit_trigger",
				TargetName:      "Clear Child Lists",
				ClearEffects:    &clear,
				ClearConditions: &clear,
			},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	trigger := scen.root.section("Triggers").list("trigger_data")[0]
	checkIntField(t, trigger, "number_of_effects", 0)
	checkIntField(t, trigger, "number_of_conditions", 0)
	if got := trigger.intList("effect_display_order_array"); len(got) != 0 {
		t.Fatalf("effect order = %#v, want empty", got)
	}
	if got := trigger.intList("condition_display_order_array"); len(got) != 0 {
		t.Fatalf("condition order = %#v, want empty", got)
	}
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
}

func TestPatchRecipeEditTriggerReplaceChildren(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "replace_trigger_children.aoe2scenario")
	enabled := false
	timer := 3
	newTimer := 11
	triggerZero := 0
	recipe := Recipe{
		Triggers: []TriggerRecipe{
			{Op: "clear_triggers"},
			{
				Op:      "add_trigger",
				Name:    "Replace Child Lists",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timer},
				},
				Effects: []EffectRecipe{
					{Op: "display_instructions", Message: "replace me"},
				},
			},
			{
				Op:         "edit_trigger",
				TargetName: "Replace Child Lists",
				ReplaceEffects: []EffectRecipe{
					{Op: "send_chat", Message: "replacement chat"},
					{Op: "activate_trigger", TriggerID: &triggerZero},
				},
				ReplaceConditions: []ConditionRecipe{
					{Op: "timer", Timer: &newTimer},
				},
			},
		},
	}
	if _, err := PatchRecipeFile(input, out, recipe); err != nil {
		t.Fatalf("PatchRecipeFile: %v", err)
	}
	scen, err := Open(out)
	if err != nil {
		t.Fatalf("Open patched: %v", err)
	}
	trigger := scen.root.section("Triggers").list("trigger_data")[0]
	effects := trigger.list("effect_data")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 3)
	checkIntField(t, effects[1], "effect_type", 8)
	conditions := trigger.list("condition_data")
	if len(conditions) != 1 {
		t.Fatalf("conditions = %d, want 1", len(conditions))
	}
	checkIntField(t, conditions[0], "timer", newTimer)
	if !scen.Triggers.InvariantOK {
		t.Fatalf("trigger invariant failed: %s", scen.Triggers.InvariantNote)
	}
}

func TestEditTriggerReplaceChildrenRejectsMixedModes(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	enabled := false
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:      "add_trigger",
		Name:    "Mixed Modes",
		Enabled: &enabled,
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "old"},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	if err := editTriggerChildLists(spec, node, TriggerRecipe{
		ReplaceEffects: []EffectRecipe{{Op: "send_chat", Message: "new"}},
		RemoveEffects:  []int{0},
	}); err == nil || !strings.Contains(err.Error(), "replace_effects cannot be combined") {
		t.Fatalf("mixed replace/remove error = %v, want replace_effects guard", err)
	}
}

func TestLintTriggerRuntimeHazardsFlagsLoopingDisplay(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	enabled := true
	looping := true
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:      "add_trigger",
		Name:    "Looping Display",
		Enabled: &enabled,
		Looping: &looping,
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "unsafe"},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	report := LintReport{}
	lintTriggerRuntimeHazards(&report, []*parsedNode{node}, LintOptions{})
	if len(report.Issues) != 1 {
		t.Fatalf("issues = %d, want 1: %+v", len(report.Issues), report.Issues)
	}
	if report.Issues[0].Severity != "error" || report.Issues[0].Code != "looping_display_effect" {
		t.Fatalf("issue = %+v, want looping_display_effect error", report.Issues[0])
	}
	if report.Issues[0].FactID != "scenario.looping_display_effect_spams" || report.Issues[0].FactTier != "engine_verified" {
		t.Fatalf("issue fact = %+v, want engine fact citation", report.Issues[0])
	}
}

func TestLintTriggerTextMarkupCitesEngineFacts(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	enabled := true
	looping := false
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:      "add_trigger",
		Name:    "Bad Color Markup",
		Enabled: &enabled,
		Looping: &looping,
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "<GREEN>first <RED>second"},
			{Op: "send_chat", Message: "<WHITE>bad"},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	report := LintReport{}
	lintTriggerRuntimeHazards(&report, []*parsedNode{node}, LintOptions{})
	if !hasIssueCode(report.Issues, "trigger_color_tag_not_leading_singleton") {
		t.Fatalf("missing multiple-color-tag issue: %+v", report.Issues)
	}
	if !hasIssueCode(report.Issues, "unsupported_white_color_tag") {
		t.Fatalf("missing white-tag issue: %+v", report.Issues)
	}
	for _, issue := range report.Issues {
		if issue.FactID == "" || issue.FactTier != "engine_verified" {
			t.Fatalf("issue missing engine fact citation: %+v", issue)
		}
	}
}

func TestLintTriggerRuntimeHazardsAllowsLoopingNonDisplay(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	enabled := true
	looping := true
	unitID := 83
	x := 14
	y := 14
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:      "add_trigger",
		Name:    "Looping Create",
		Enabled: &enabled,
		Looping: &looping,
		Effects: []EffectRecipe{
			{Op: "create_object", ObjectListUnitID: &unitID, SourcePlayer: intPtr(1), LocationX: &x, LocationY: &y},
		},
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	report := LintReport{}
	lintTriggerRuntimeHazards(&report, []*parsedNode{node}, LintOptions{})
	if len(report.Issues) != 0 {
		t.Fatalf("issues = %+v, want none", report.Issues)
	}
}

func TestLintActiveComputerWithoutAuthoredUnits(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay Transparency Diagnostic v14.1 Castle Kill Calibration.aoe2scenario")
	report, err := LintFile(input)
	if err != nil {
		t.Fatalf("LintFile: %v", err)
	}
	if !hasIssueCode(report.Issues, "active_computer_without_authored_units") {
		t.Fatalf("missing active computer starter-injection warning: %+v", report.Issues)
	}
	if !report.OK {
		t.Fatalf("report should remain OK for warning-only starter-injection lint: %+v", report.Issues)
	}
}

func intPtr(value int) *int {
	return &value
}

func TestDeployCheckCatchesXSNameFailures(t *testing.T) {
	dir := t.TempDir()
	xsDir := filepath.Join(dir, "resources", "_common", "xs")
	if err := os.MkdirAll(xsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "Probe.xs"), []byte("include \"constants.xs\";\nvoid Boot() { xsSetPlayerAttribute(1, BAD_CONST, 1); }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "constants.xs"), []byte("const int BAD_CONST = 22;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scen := scenarioWithXSFields("Probe", "Probe.xs", "embedded")
	report, err := scen.DeployCheck(dir)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK {
		t.Fatalf("report unexpectedly OK: %+v", report)
	}
	if !hasDeployFinding(report, "does not match") || !hasDeployFinding(report, "has no .xs extension") || !hasDeployFinding(report, "does not resolve") {
		t.Fatalf("missing expected findings: %+v", report.Findings)
	}
}

func TestDeployCheckCatchesCrossFileNonExtern(t *testing.T) {
	dir := t.TempDir()
	xsDir := filepath.Join(dir, "resources", "_common", "xs")
	if err := os.MkdirAll(xsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "Probe.xs"), []byte("include \"constants.xs\";\nvoid Boot() { xsSetPlayerAttribute(1, BAD_CONST, 1); }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "constants.xs"), []byte("const int BAD_CONST = 22;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scen := scenarioWithXSFields("Probe.xs", "Probe.xs", "embedded")
	report, err := scen.DeployCheck(dir)
	if err != nil {
		t.Fatal(err)
	}
	if report.OK {
		t.Fatalf("report unexpectedly OK: %+v", report)
	}
	if !hasDeployFinding(report, "declared without extern") {
		t.Fatalf("missing expected finding: %+v", report.Findings)
	}
	foundFact := false
	for _, finding := range report.Findings {
		if finding.FactID == "xs.cross_file_symbols_require_extern" && finding.FactTier == "engine_verified" {
			foundFact = true
		}
	}
	if !foundFact {
		t.Fatalf("missing extern fact citation: %+v", report.Findings)
	}
}

func TestDeployCheckPassesResolvedExternXS(t *testing.T) {
	dir := t.TempDir()
	xsDir := filepath.Join(dir, "resources", "_common", "xs")
	if err := os.MkdirAll(xsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "Probe.xs"), []byte("include \"constants.xs\";\nvoid Boot() { xsSetPlayerAttribute(1, GOOD_CONST, 1); }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "constants.xs"), []byte("extern const int GOOD_CONST = 22;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scen := scenarioWithXSFields("Probe.xs", "Probe.xs", "embedded")
	report, err := scen.DeployCheck(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK || report.XS.ResolvedRelativePath != "Probe.xs" || report.Analysis == nil {
		t.Fatalf("report = %+v", report)
	}
}

func TestXSCensusInventoriesEmbeddedXS(t *testing.T) {
	scen := scenarioWithXSFields("Entry.xs", "Entry.xs", "include \"missing.xs\";\nconst int LOCAL = 1;\nvoid Boot() {}\n")
	scen.Version = "1.58"
	scen.Triggers = &TriggerInfo{Triggers: []TriggerSummary{{
		Index: 7,
		Name:  "Call Boot",
		EffectData: []EffectSummary{{
			EffectIndex: 3,
			Type:        55,
			TypeName:    "script_call",
			Text:        "Boot(); Other(); Boot();",
		}},
	}}}
	report, err := scen.XSCensus(XSCensusOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK {
		t.Fatalf("report unexpectedly failed: %+v", report)
	}
	if report.Counts.ScriptCalls != 1 || report.Counts.Functions != 1 || report.Counts.UniqueCalledFunctions != 2 {
		t.Fatalf("counts = %+v", report.Counts)
	}
	if len(report.ScriptCalls) != 1 || report.ScriptCalls[0].TriggerIndex != 7 || len(report.ScriptCalls[0].Calls) != 2 {
		t.Fatalf("script calls = %+v", report.ScriptCalls)
	}
	if !hasDeployFinding(xsCensusFindings(report), "embedded XS include cannot be resolved") {
		t.Fatalf("missing embedded include finding: %+v", report.Findings)
	}
}

func TestXSCensusUsesDeployTree(t *testing.T) {
	dir := t.TempDir()
	xsDir := filepath.Join(dir, "resources", "_common", "xs")
	if err := os.MkdirAll(xsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "Entry.xs"), []byte("include \"shared.xs\";\nvoid Boot() { xsSetPlayerAttribute(1, SHARED, 1); }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xsDir, "shared.xs"), []byte("extern const int SHARED = 22;\nvoid Helper() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scen := scenarioWithXSFields("Entry.xs", "Entry.xs", "embedded")
	report, err := scen.XSCensus(XSCensusOptions{DeployTree: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK || report.Counts.ResolvedIncludes != 1 || report.Counts.Functions != 2 || report.XS.ResolvedRelativePath != "Entry.xs" {
		t.Fatalf("report = %+v", report)
	}
}

func scenarioWithXSFields(scriptName, scriptPath, content string) *File {
	return &File{root: &parsedRoot{Sections: []*parsedSection{
		{
			Name: "Map",
			Fields: []*parsedNode{
				{Name: "script_name", Value: scriptName},
			},
		},
		{
			Name: "Files",
			Fields: []*parsedNode{
				{Name: "script_file_path", Value: scriptPath},
				{Name: "script_file_content", Value: content},
			},
		},
	}}}
}

func xsCensusFindings(report XSCensusReport) DeployCheckReport {
	return DeployCheckReport{Findings: report.Findings}
}

func hasDeployFinding(report DeployCheckReport, needle string) bool {
	for _, finding := range report.Findings {
		if strings.Contains(finding.What, needle) {
			return true
		}
	}
	return false
}

func TestMapPatchPointsCircleAndLine(t *testing.T) {
	radius := 1
	circle, err := mapPatchPoints(MapRecipe{Op: "set_terrain_circle", X1: 10, Y1: 10, Radius: &radius}, 20, 20)
	if err != nil {
		t.Fatalf("circle points: %v", err)
	}
	if len(circle) != 5 {
		t.Fatalf("circle points len=%d want 5: %+v", len(circle), circle)
	}
	line, err := mapPatchPoints(MapRecipe{Op: "set_terrain_line", X1: 1, Y1: 1, X2: 3, Y2: 3}, 20, 20)
	if err != nil {
		t.Fatalf("line points: %v", err)
	}
	if len(line) != 3 || line[0] != (mapPoint{X: 1, Y: 1}) || line[2] != (mapPoint{X: 3, Y: 3}) {
		t.Fatalf("line points = %+v", line)
	}
}

func TestMapPatchPointsBorder(t *testing.T) {
	thickness := 1
	points, err := mapPatchPoints(MapRecipe{Op: "set_terrain_border", X1: 2, Y1: 3, X2: 5, Y2: 6, Thickness: &thickness}, 10, 10)
	if err != nil {
		t.Fatalf("border points: %v", err)
	}
	if len(points) != 12 {
		t.Fatalf("border points len=%d want 12: %+v", len(points), points)
	}
	if !hasMapPoint(points, mapPoint{X: 2, Y: 3}) || !hasMapPoint(points, mapPoint{X: 5, Y: 6}) {
		t.Fatalf("border missing corner points: %+v", points)
	}
	if hasMapPoint(points, mapPoint{X: 3, Y: 4}) || hasMapPoint(points, mapPoint{X: 4, Y: 5}) {
		t.Fatalf("border included inner points: %+v", points)
	}

	thick := 2
	thickPoints, err := mapPatchPoints(MapRecipe{Op: "set_terrain_border", X1: 0, Y1: 0, X2: 4, Y2: 4, Thickness: &thick}, 10, 10)
	if err != nil {
		t.Fatalf("thick border points: %v", err)
	}
	if len(thickPoints) != 24 {
		t.Fatalf("thick border points len=%d want 24: %+v", len(thickPoints), thickPoints)
	}
	if hasMapPoint(thickPoints, mapPoint{X: 2, Y: 2}) {
		t.Fatalf("thick border included center: %+v", thickPoints)
	}

	zero := 0
	if _, err := mapPatchPoints(MapRecipe{Op: "set_terrain_border", X1: 0, Y1: 0, X2: 4, Y2: 4, Thickness: &zero}, 10, 10); err == nil {
		t.Fatalf("zero-thickness border succeeded")
	}
}

func TestMapPatchPointsCopyTerrainArea(t *testing.T) {
	targetX := 6
	targetY := 7
	points, err := mapPatchPoints(MapRecipe{Op: "copy_terrain_area", X1: 1, Y1: 2, X2: 3, Y2: 4, TargetX: &targetX, TargetY: &targetY}, 10, 10)
	if err != nil {
		t.Fatalf("copy terrain points: %v", err)
	}
	if len(points) != 9 {
		t.Fatalf("copy terrain points len=%d want 9: %+v", len(points), points)
	}
	if points[0] != (mapPoint{X: 1, Y: 2}) || points[len(points)-1] != (mapPoint{X: 3, Y: 4}) {
		t.Fatalf("copy terrain source points = %+v", points)
	}
	if _, err := mapPatchPoints(MapRecipe{Op: "copy_terrain_area", X1: 1, Y1: 2, X2: 3, Y2: 4}, 10, 10); err == nil {
		t.Fatalf("copy terrain without target succeeded")
	}
	targetX = 8
	targetY = 8
	if _, err := mapPatchPoints(MapRecipe{Op: "copy_terrain_area", X1: 1, Y1: 2, X2: 3, Y2: 4, TargetX: &targetX, TargetY: &targetY}, 10, 10); err == nil {
		t.Fatalf("copy terrain accepted out-of-bounds destination")
	}
}

func hasMapPoint(points []mapPoint, needle mapPoint) bool {
	for _, point := range points {
		if point == needle {
			return true
		}
	}
	return false
}

func TestCopyTerrainArea(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	terrain := 222
	elevation := 7
	layer := 3
	targetX := 8
	targetY := 9
	recipe := Recipe{Map: []MapRecipe{
		{Op: "set_terrain_rect", X1: 2, Y1: 3, X2: 3, Y2: 5, TerrainID: &terrain, Elevation: &elevation, Layer: &layer},
		{Op: "copy_terrain_area", X1: 2, Y1: 3, X2: 3, Y2: 5, TargetX: &targetX, TargetY: &targetY},
	}}
	plan, err := file.Plan(recipe)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.MapTilesChanged != 12 {
		t.Fatalf("plan map tiles changed=%d want 12", plan.MapTilesChanged)
	}
	if err := file.ApplyRecipe(recipe); err != nil {
		t.Fatalf("ApplyRecipe: %v", err)
	}
	tiles, width, _, err := file.mapTiles()
	if err != nil {
		t.Fatalf("mapTiles: %v", err)
	}
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 2; dx++ {
			src := tiles[(3+dy)*width+2+dx]
			dst := tiles[(targetY+dy)*width+targetX+dx]
			for _, field := range []string{"terrain_id", "elevation", "layer"} {
				want, _ := src.intValue(field)
				got, _ := dst.intValue(field)
				if got != want {
					t.Fatalf("copied %s at delta (%d,%d) = %d want %d", field, dx, dy, got, want)
				}
			}
		}
	}
}

func TestCopyTerrainAreaOverlappingUsesSourceSnapshot(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	targetX := 1
	targetY := 0
	values := []int{201, 202, 203}
	recipe := Recipe{
		Map: []MapRecipe{
			{Op: "set_terrain_rect", X1: 0, Y1: 0, X2: 0, Y2: 0, TerrainID: &values[0]},
			{Op: "set_terrain_rect", X1: 1, Y1: 0, X2: 1, Y2: 0, TerrainID: &values[1]},
			{Op: "set_terrain_rect", X1: 2, Y1: 0, X2: 2, Y2: 0, TerrainID: &values[2]},
			{Op: "copy_terrain_area", X1: 0, Y1: 0, X2: 2, Y2: 0, TargetX: &targetX, TargetY: &targetY},
		},
	}
	if err := file.ApplyRecipe(recipe); err != nil {
		t.Fatalf("ApplyRecipe: %v", err)
	}
	tiles, width, _, err := file.mapTiles()
	if err != nil {
		t.Fatalf("mapTiles: %v", err)
	}
	want := []int{201, 201, 202, 203}
	for x, wantID := range want {
		got, _ := tiles[x].intValue("terrain_id")
		if got != wantID {
			t.Fatalf("terrain at x=%d = %d want %d; width=%d", x, got, wantID, width)
		}
	}
}

func TestEditExistingTriggerRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	raw, err := buildTriggerFromRecipe(spec, TriggerRecipe{
		Op:   "add_trigger",
		Name: "Original",
	})
	if err != nil {
		t.Fatalf("buildTriggerFromRecipe: %v", err)
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		t.Fatalf("triggerEffectSpecs: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := p.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse trigger raw: %v", err)
	}
	newName := "Edited"
	if err := setStringField(node, "trigger_name", "str32", newName); err != nil {
		t.Fatalf("setStringField: %v", err)
	}
	if err := setIntField(node, "enabled", "u32", 1); err != nil {
		t.Fatalf("setIntField enabled: %v", err)
	}
	triggerID := 2
	if err := appendEffectsToTrigger(spec, node, []EffectRecipe{{Op: "deactivate_trigger", TriggerID: &triggerID}}); err != nil {
		t.Fatalf("appendEffectsToTrigger: %v", err)
	}
	timer := 9
	if err := appendConditionsToTrigger(spec, node, []ConditionRecipe{{Op: "timer", Timer: &timer}}); err != nil {
		t.Fatalf("appendConditionsToTrigger: %v", err)
	}
	roundTrip := parser{data: node.raw(), sections: map[string]*parsedSection{}}
	edited, err := roundTrip.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		t.Fatalf("parse edited trigger raw: %v", err)
	}
	name, _ := edited.stringValue("trigger_name")
	if name != newName {
		t.Fatalf("trigger_name = %q, want %q", name, newName)
	}
	checkIntField(t, edited, "enabled", 1)
	effects := edited.list("effect_data")
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	checkIntField(t, effects[0], "effect_type", 9)
	checkIntField(t, effects[0], "trigger_id", triggerID)
	conditions := edited.list("condition_data")
	if len(conditions) != 1 {
		t.Fatalf("conditions = %d, want 1", len(conditions))
	}
	checkIntField(t, conditions[0], "condition_type", 10)
	checkIntField(t, conditions[0], "timer", timer)
}

func TestBuildUnitFromRecipeRaw(t *testing.T) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		t.Fatalf("LoadCurrentDESpec: %v", err)
	}
	x := 10.5
	y := 12.5
	refID := 999
	caption := "Click This Marker"
	raw, err := buildUnitFromRecipe(spec, UnitRecipe{
		Op:            "add_unit",
		Player:        1,
		UnitConst:     83,
		X:             &x,
		Y:             &y,
		ReferenceID:   &refID,
		CaptionString: caption,
	}, 1)
	if err != nil {
		t.Fatalf("buildUnitFromRecipe: %v", err)
	}
	unitSpec, err := unitStructSpec(spec)
	if err != nil {
		t.Fatalf("unitStructSpec: %v", err)
	}
	p := parser{data: raw, sections: map[string]*parsedSection{}}
	unit, err := p.parseNode("UnitStruct", unitSpec, nil)
	if err != nil {
		t.Fatalf("parse unit raw: %v", err)
	}
	checkIntField(t, unit, "reference_id", refID)
	checkIntField(t, unit, "unit_const", 83)
	checkIntField(t, unit, "status", 2)
	gotX, _ := unit.floatValue("x")
	gotY, _ := unit.floatValue("y")
	if gotX != x || gotY != y {
		t.Fatalf("coords = %v,%v want %v,%v", gotX, gotY, x, y)
	}
	gotCaption, _ := unit.stringValue("caption_string")
	if gotCaption != caption {
		t.Fatalf("caption_string = %q, want %q", gotCaption, caption)
	}
	assertLengthPrefixedStringRaw(t, unit.field("caption_string").Raw, caption)
}

func TestEditAndRemoveUnitSelectors(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	player := 1
	section := file.root.section("Units").list("players_units")[player]
	unitField := section.field("units")
	index := len(unitField.Elements)
	x := 12.5
	y := 13.5
	caption := "selector-index-target"
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: 83, X: &x, Y: &y, CaptionString: caption}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	newX := 44.5
	newCaption := "selector-index-edited"
	if err := file.EditUnit(UnitRecipe{Op: "edit_unit", TargetPlayer: &player, TargetIndex: &index, X: &newX, CaptionString: newCaption}); err != nil {
		t.Fatalf("EditUnit by target_index: %v", err)
	}
	unit, _, _, err := file.findUnitByCaption(newCaption, &player)
	if err != nil {
		t.Fatalf("find edited unit: %v", err)
	}
	gotX, _ := unit.floatValue("x")
	if gotX != newX {
		t.Fatalf("edited x = %v, want %v", gotX, newX)
	}
	destPlayer := 2
	if err := file.EditUnit(UnitRecipe{Op: "edit_unit", TargetCaption: newCaption, TargetPlayer: &player, SetPlayer: &destPlayer}); err != nil {
		t.Fatalf("EditUnit set_player by caption: %v", err)
	}
	if _, _, _, err := file.findUnitByCaption(newCaption, &destPlayer); err != nil {
		t.Fatalf("moved unit not found on destination player: %v", err)
	}
	if err := file.RemoveUnit(UnitRecipe{Op: "remove_unit", TargetCaption: newCaption, TargetPlayer: &destPlayer}); err != nil {
		t.Fatalf("RemoveUnit by caption: %v", err)
	}
	if _, _, _, err := file.findUnitByCaption(newCaption, &destPlayer); err == nil {
		t.Fatal("removed unit still found by caption")
	}
}

func TestUnitCaptionSelectorRejectsAmbiguousMatches(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	caption := "selector-ambiguous"
	x := 10.5
	y := 10.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, CaptionString: caption}); err != nil {
		t.Fatalf("AddUnit P1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 2, UnitConst: 83, X: &x, Y: &y, CaptionString: caption}); err != nil {
		t.Fatalf("AddUnit P2: %v", err)
	}
	if err := file.EditUnit(UnitRecipe{Op: "edit_unit", TargetCaption: caption, X: &x}); err == nil || !strings.Contains(err.Error(), "matched 2 units") {
		t.Fatalf("ambiguous caption error = %v, want matched 2 units", err)
	}
	targetPlayer := 2
	if err := file.EditUnit(UnitRecipe{Op: "edit_unit", TargetCaption: caption, TargetPlayer: &targetPlayer, X: &x}); err != nil {
		t.Fatalf("player-scoped caption edit: %v", err)
	}
}

func TestRemoveUnitRejectsSelectedObjectReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	refID := 990501
	x := 15.5
	y := 15.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "references selected unit",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{refID}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	err = file.RemoveUnit(UnitRecipe{Op: "remove_unit", ReferenceID: &refID})
	if err == nil || !strings.Contains(err.Error(), "selected_object_ids[0] references it") {
		t.Fatalf("RemoveUnit error = %v, want selected_object_ids reference refusal", err)
	}
}

func TestRemoveUnitRejectsConditionAndLocationReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	refID := 990502
	x := 16.5
	y := 16.5
	timer := 1
	locationX := 20
	locationY := 20
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "references location unit",
		Conditions: []ConditionRecipe{
			{Op: "object_selected", UnitObject: &refID},
			{Op: "timer", Timer: &timer},
		},
		Effects: []EffectRecipe{
			{Op: "task_object", SourcePlayer: intPtr(1), LocationX: &locationX, LocationY: &locationY, LocationObjectReference: &refID},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	err = file.RemoveUnit(UnitRecipe{Op: "remove_unit", ReferenceID: &refID})
	if err == nil || !strings.Contains(err.Error(), "condition 0 field unit_object references it") {
		t.Fatalf("RemoveUnit error = %v, want unit_object reference refusal", err)
	}

	file, err = Open(input)
	if err != nil {
		t.Fatalf("Open second file: %v", err)
	}
	refID = 990503
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit location target: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "references location unit only",
		Effects: []EffectRecipe{
			{Op: "task_object", SourcePlayer: intPtr(1), LocationX: &locationX, LocationY: &locationY, LocationObjectReference: &refID},
		},
	}); err != nil {
		t.Fatalf("AddTrigger location-only: %v", err)
	}
	err = file.RemoveUnit(UnitRecipe{Op: "remove_unit", ReferenceID: &refID})
	if err == nil || !strings.Contains(err.Error(), "effect 0 field location_object_reference references it") {
		t.Fatalf("RemoveUnit location error = %v, want location_object_reference refusal", err)
	}
}

func TestRemoveUnitsInArea(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := file.Units.Total
	player := 1
	unitConst := 83
	x1, y1 := 30.5, 30.5
	x2, y2 := 31.5, 31.5
	x3, y3 := 45.5, 45.5
	ref1 := 991001
	ref2 := 991002
	ref3 := 991003
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x1, Y: &y1, ReferenceID: &ref1}); err != nil {
		t.Fatalf("AddUnit ref1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x2, Y: &y2, ReferenceID: &ref2}); err != nil {
		t.Fatalf("AddUnit ref2: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x3, Y: &y3, ReferenceID: &ref3}); err != nil {
		t.Fatalf("AddUnit ref3: %v", err)
	}
	ax1, ay1, ax2, ay2 := 30.0, 30.0, 32.0, 32.0
	plan, err := file.Plan(Recipe{Units: []UnitRecipe{{
		Op:              "remove_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
	}}})
	if err != nil {
		t.Fatalf("Plan remove_units_in_area: %v", err)
	}
	if plan.UnitCountAfter != before+1 {
		t.Fatalf("plan unit_count_after = %d, want %d", plan.UnitCountAfter, before+1)
	}
	removed, err := file.RemoveUnitsInArea(UnitRecipe{
		Op:              "remove_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
	})
	if err != nil {
		t.Fatalf("RemoveUnitsInArea: %v", err)
	}
	if removed != 2 {
		t.Fatalf("removed = %d, want 2", removed)
	}
	if file.Units.Total != before+1 {
		t.Fatalf("unit total = %d, want %d", file.Units.Total, before+1)
	}
	if _, _, _, err := file.findUnitByReferenceID(ref1); err == nil {
		t.Fatal("ref1 survived area removal")
	}
	if _, _, _, err := file.findUnitByReferenceID(ref2); err == nil {
		t.Fatal("ref2 survived area removal")
	}
	if _, _, _, err := file.findUnitByReferenceID(ref3); err != nil {
		t.Fatalf("outside unit missing after area removal: %v", err)
	}
}

func TestRemoveUnitsInAreaRejectsReferencesBeforeMutation(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := file.Units.Total
	player := 1
	unitConst := 83
	x1, y1 := 34.5, 34.5
	x2, y2 := 35.5, 35.5
	ref1 := 991101
	ref2 := 991102
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x1, Y: &y1, ReferenceID: &ref1}); err != nil {
		t.Fatalf("AddUnit ref1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x2, Y: &y2, ReferenceID: &ref2}); err != nil {
		t.Fatalf("AddUnit ref2: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "area delete blocker",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{ref2}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	ax1, ay1, ax2, ay2 := 34.0, 34.0, 36.0, 36.0
	recipe := UnitRecipe{
		Op:              "remove_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
	}
	if _, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}}); err == nil || !strings.Contains(err.Error(), "selected_object_ids[0] references it") {
		t.Fatalf("Plan remove_units_in_area error = %v, want selected-object reference refusal", err)
	}
	if _, err := file.RemoveUnitsInArea(recipe); err == nil || !strings.Contains(err.Error(), "selected_object_ids[0] references it") {
		t.Fatalf("RemoveUnitsInArea error = %v, want selected-object reference refusal", err)
	}
	if file.Units.Total != before+2 {
		t.Fatalf("unit total changed after rejected area removal: %d, want %d", file.Units.Total, before+2)
	}
	if _, _, _, err := file.findUnitByReferenceID(ref1); err != nil {
		t.Fatalf("unreferenced matched unit was removed before blocker: %v", err)
	}
	if _, _, _, err := file.findUnitByReferenceID(ref2); err != nil {
		t.Fatalf("referenced matched unit missing after rejected removal: %v", err)
	}
}

func TestMoveUnitsInArea(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := file.Units.Total
	player := 1
	unitConst := 83
	x1, y1 := 20.5, 20.5
	x2, y2 := 21.5, 21.5
	x3, y3 := 30.5, 30.5
	ref1 := 991401
	ref2 := 991402
	ref3 := 991403
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x1, Y: &y1, ReferenceID: &ref1}); err != nil {
		t.Fatalf("AddUnit ref1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x2, Y: &y2, ReferenceID: &ref2}); err != nil {
		t.Fatalf("AddUnit ref2: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x3, Y: &y3, ReferenceID: &ref3}); err != nil {
		t.Fatalf("AddUnit ref3: %v", err)
	}
	ax1, ay1, ax2, ay2 := 20.0, 20.0, 22.0, 22.0
	offsetX, offsetY := 4.0, 5.0
	recipe := UnitRecipe{
		Op:              "move_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
		OffsetX:         &offsetX,
		OffsetY:         &offsetY,
	}
	plan, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}})
	if err != nil {
		t.Fatalf("Plan move_units_in_area: %v", err)
	}
	if plan.UnitCountAfter != before+3 {
		t.Fatalf("plan unit_count_after = %d, want %d", plan.UnitCountAfter, before+3)
	}
	moved, err := file.MoveUnitsInArea(recipe)
	if err != nil {
		t.Fatalf("MoveUnitsInArea: %v", err)
	}
	if moved != 2 {
		t.Fatalf("moved = %d, want 2", moved)
	}
	if file.Units.Total != before+3 {
		t.Fatalf("unit total = %d, want %d", file.Units.Total, before+3)
	}
	unit1, _, _, err := file.findUnitByReferenceID(ref1)
	if err != nil {
		t.Fatalf("moved ref1 missing: %v", err)
	}
	checkFloatField(t, unit1, "x", x1+offsetX)
	checkFloatField(t, unit1, "y", y1+offsetY)
	unit2, _, _, err := file.findUnitByReferenceID(ref2)
	if err != nil {
		t.Fatalf("moved ref2 missing: %v", err)
	}
	checkFloatField(t, unit2, "x", x2+offsetX)
	checkFloatField(t, unit2, "y", y2+offsetY)
	unit3, _, _, err := file.findUnitByReferenceID(ref3)
	if err != nil {
		t.Fatalf("outside ref3 missing: %v", err)
	}
	checkFloatField(t, unit3, "x", x3)
	checkFloatField(t, unit3, "y", y3)
}

func TestMoveUnitsInAreaTargetAnchor(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	player := 1
	unitConst := 83
	x, y := 10.5, 11.5
	refID := 991801
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	ax1, ay1, ax2, ay2 := 10.0, 11.0, 12.0, 13.0
	targetX, targetY := 30.0, 31.0
	recipe := UnitRecipe{
		Op:              "move_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
		TargetX:         &targetX,
		TargetY:         &targetY,
	}
	if _, err := file.MoveUnitsInArea(recipe); err != nil {
		t.Fatalf("MoveUnitsInArea target anchor: %v", err)
	}
	unit, _, _, err := file.findUnitByReferenceID(refID)
	if err != nil {
		t.Fatalf("unit missing after target move: %v", err)
	}
	checkFloatField(t, unit, "x", 30.5)
	checkFloatField(t, unit, "y", 31.5)
}

func TestMoveUnitsInAreaRejectsMissingOffsetAndOutOfBounds(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	player := 1
	unitConst := 83
	x, y := 1.5, 1.5
	refID := 991501
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	ax1, ay1, ax2, ay2 := 1.0, 1.0, 2.0, 2.0
	recipe := UnitRecipe{
		Op:              "move_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
	}
	if _, err := file.MoveUnitsInArea(recipe); err == nil || !strings.Contains(err.Error(), "requires offset_x/offset_y or target_x/target_y") {
		t.Fatalf("MoveUnitsInArea missing offset error = %v", err)
	}
	offsetX := -3.0
	recipe.OffsetX = &offsetX
	if _, err := file.MoveUnitsInArea(recipe); err == nil || !strings.Contains(err.Error(), "outside map") {
		t.Fatalf("MoveUnitsInArea out-of-bounds error = %v", err)
	}
	targetX := 10.0
	targetY := 10.0
	recipe.TargetX = &targetX
	recipe.TargetY = &targetY
	if _, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}}); err == nil || !strings.Contains(err.Error(), "cannot combine offset_x/offset_y with target_x/target_y") {
		t.Fatalf("Plan move_units_in_area mixed target/offset error = %v", err)
	}
	unit, _, _, err := file.findUnitByReferenceID(refID)
	if err != nil {
		t.Fatalf("unit missing after rejected move: %v", err)
	}
	checkFloatField(t, unit, "x", x)
	checkFloatField(t, unit, "y", y)
}

func TestEditUnitsInArea(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := file.Units.Total
	player := 1
	targetPlayer := 2
	unitConst := 83
	newUnitConst := 74
	status := 4
	rotation := 1.25
	caption := "area edited"
	x1, y1 := 23.5, 23.5
	x2, y2 := 24.5, 24.5
	x3, y3 := 32.5, 32.5
	ref1 := 991601
	ref2 := 991602
	ref3 := 991603
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x1, Y: &y1, ReferenceID: &ref1}); err != nil {
		t.Fatalf("AddUnit ref1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x2, Y: &y2, ReferenceID: &ref2}); err != nil {
		t.Fatalf("AddUnit ref2: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x3, Y: &y3, ReferenceID: &ref3}); err != nil {
		t.Fatalf("AddUnit ref3: %v", err)
	}
	ax1, ay1, ax2, ay2 := 23.0, 23.0, 25.0, 25.0
	recipe := UnitRecipe{
		Op:              "edit_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
		UnitConst:       newUnitConst,
		Status:          &status,
		Rotation:        &rotation,
		CaptionString:   caption,
		SetPlayer:       &targetPlayer,
	}
	plan, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}})
	if err != nil {
		t.Fatalf("Plan edit_units_in_area: %v", err)
	}
	if plan.UnitCountAfter != before+3 {
		t.Fatalf("plan unit_count_after = %d, want %d", plan.UnitCountAfter, before+3)
	}
	edited, err := file.EditUnitsInArea(recipe)
	if err != nil {
		t.Fatalf("EditUnitsInArea: %v", err)
	}
	if edited != 2 {
		t.Fatalf("edited = %d, want 2", edited)
	}
	if file.Units.Total != before+3 {
		t.Fatalf("unit total = %d, want %d", file.Units.Total, before+3)
	}
	playerSections := file.root.section("Units").list("players_units")
	for _, refID := range []int{ref1, ref2} {
		unit, section, _, err := file.findUnitByReferenceID(refID)
		if err != nil {
			t.Fatalf("edited unit %d missing: %v", refID, err)
		}
		if section != playerSections[targetPlayer] {
			t.Fatalf("edited unit %d not moved to player %d", refID, targetPlayer)
		}
		checkIntField(t, unit, "unit_const", newUnitConst)
		checkIntField(t, unit, "status", status)
		checkFloatField(t, unit, "rotation", rotation)
		if got, _ := unit.stringValue("caption_string"); got != caption {
			t.Fatalf("caption_string = %q, want %q", got, caption)
		}
	}
	outside, section, _, err := file.findUnitByReferenceID(ref3)
	if err != nil {
		t.Fatalf("outside unit missing: %v", err)
	}
	if section != playerSections[player] {
		t.Fatalf("outside unit moved unexpectedly")
	}
	checkIntField(t, outside, "unit_const", unitConst)
}

func TestEditUnitsInAreaRejectsNoFieldsAndPositionFields(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	player := 1
	unitConst := 83
	x, y := 26.5, 26.5
	refID := 991701
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	ax1, ay1, ax2, ay2 := 26.0, 26.0, 27.0, 27.0
	recipe := UnitRecipe{
		Op:              "edit_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
	}
	if _, err := file.EditUnitsInArea(recipe); err == nil || !strings.Contains(err.Error(), "requires unit_const") {
		t.Fatalf("EditUnitsInArea no-fields error = %v", err)
	}
	recipe.X = &x
	if _, err := file.EditUnitsInArea(recipe); err == nil || !strings.Contains(err.Error(), "does not edit positions") {
		t.Fatalf("EditUnitsInArea position error = %v", err)
	}
	unit, _, _, err := file.findUnitByReferenceID(refID)
	if err != nil {
		t.Fatalf("unit missing after rejected edit: %v", err)
	}
	checkFloatField(t, unit, "x", x)
	checkFloatField(t, unit, "y", y)
	checkIntField(t, unit, "unit_const", unitConst)
}

func TestCopyUnitsInArea(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := file.Units.Total
	player := 1
	targetPlayer := 2
	unitConst := 83
	x1, y1 := 40.5, 40.5
	x2, y2 := 41.5, 41.5
	ref1 := 991201
	ref2 := 991202
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x1, Y: &y1, ReferenceID: &ref1, CaptionString: "CopyMeA"}); err != nil {
		t.Fatalf("AddUnit ref1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x2, Y: &y2, ReferenceID: &ref2, GarrisonedInID: &ref1, CaptionString: "CopyMeB"}); err != nil {
		t.Fatalf("AddUnit ref2: %v", err)
	}
	ax1, ay1, ax2, ay2 := 40.0, 40.0, 42.0, 42.0
	offsetX, offsetY := 10.0, 5.0
	base := 992200
	recipe := UnitRecipe{
		Op:              "copy_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
		OffsetX:         &offsetX,
		OffsetY:         &offsetY,
		ReferenceIDBase: &base,
		SetPlayer:       &targetPlayer,
		CaptionSuffix:   " copied",
	}
	plan, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}})
	if err != nil {
		t.Fatalf("Plan copy_units_in_area: %v", err)
	}
	if plan.UnitCountAfter != before+4 {
		t.Fatalf("plan unit_count_after = %d, want %d", plan.UnitCountAfter, before+4)
	}
	copied, err := file.CopyUnitsInArea(recipe)
	if err != nil {
		t.Fatalf("CopyUnitsInArea: %v", err)
	}
	if copied != 2 {
		t.Fatalf("copied = %d, want 2", copied)
	}
	if file.Units.Total != before+4 {
		t.Fatalf("unit total = %d, want %d", file.Units.Total, before+4)
	}
	copyA, _, _, err := file.findUnitByReferenceID(base)
	if err != nil {
		t.Fatalf("copy A missing: %v", err)
	}
	copyB, _, _, err := file.findUnitByReferenceID(base + 1)
	if err != nil {
		t.Fatalf("copy B missing: %v", err)
	}
	gotX, _ := copyA.floatValue("x")
	gotY, _ := copyA.floatValue("y")
	if gotX != x1+offsetX || gotY != y1+offsetY {
		t.Fatalf("copy A position = %.1f,%.1f want %.1f,%.1f", gotX, gotY, x1+offsetX, y1+offsetY)
	}
	caption, _ := copyB.stringValue("caption_string")
	if caption != "CopyMeB copied" {
		t.Fatalf("copy B caption = %q", caption)
	}
	garrisonedIn, _ := copyB.intValue("garrisoned_in_id")
	if garrisonedIn != base {
		t.Fatalf("copy B garrisoned_in_id = %d, want copied container %d", garrisonedIn, base)
	}
	if _, _, _, err := file.findUnitByReferenceID(ref1); err != nil {
		t.Fatalf("original ref1 missing after copy: %v", err)
	}
	if _, _, _, err := file.findUnitByReferenceID(ref2); err != nil {
		t.Fatalf("original ref2 missing after copy: %v", err)
	}
}

func TestCopyUnitsInAreaTargetAnchor(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	player := 1
	unitConst := 83
	x, y := 12.5, 13.5
	refID := 991901
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	ax1, ay1, ax2, ay2 := 12.0, 13.0, 13.0, 14.0
	targetX, targetY := 30.0, 31.0
	base := 992900
	recipe := UnitRecipe{
		Op:              "copy_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
		TargetX:         &targetX,
		TargetY:         &targetY,
		ReferenceIDBase: &base,
	}
	if _, err := file.CopyUnitsInArea(recipe); err != nil {
		t.Fatalf("CopyUnitsInArea target anchor: %v", err)
	}
	copy, _, _, err := file.findUnitByReferenceID(base)
	if err != nil {
		t.Fatalf("copy missing: %v", err)
	}
	checkFloatField(t, copy, "x", 30.5)
	checkFloatField(t, copy, "y", 31.5)
	offsetX := 1.0
	recipe.OffsetX = &offsetX
	if _, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}}); err == nil || !strings.Contains(err.Error(), "cannot combine offset_x/offset_y with target_x/target_y") {
		t.Fatalf("Plan copy_units_in_area mixed target/offset error = %v", err)
	}
}

func TestCopyUnitsInAreaRejectsDuplicateReferenceBase(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	player := 1
	unitConst := 83
	x, y := 46.5, 46.5
	ref := 991301
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: player, UnitConst: unitConst, X: &x, Y: &y, ReferenceID: &ref}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	ax1, ay1, ax2, ay2 := 46.0, 46.0, 47.0, 47.0
	recipe := UnitRecipe{
		Op:              "copy_units_in_area",
		TargetPlayer:    &player,
		TargetUnitConst: &unitConst,
		TargetAreaX1:    &ax1,
		TargetAreaY1:    &ay1,
		TargetAreaX2:    &ax2,
		TargetAreaY2:    &ay2,
		ReferenceIDBase: &ref,
	}
	if _, err := file.Plan(Recipe{Units: []UnitRecipe{recipe}}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Plan duplicate copy ref error = %v, want already exists", err)
	}
	if _, err := file.CopyUnitsInArea(recipe); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("CopyUnitsInArea duplicate ref error = %v, want already exists", err)
	}
}

func TestScenarioReferencesReportsKnownDependencyClasses(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	unitRef := 990601
	x := 17.5
	y := 17.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &unitRef}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	targetTrigger := 0
	variable := 0
	stringID := 0
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "references many things",
		Conditions: []ConditionRecipe{
			{Op: "object_selected", UnitObject: &unitRef},
			{Op: "variable_value", Variable: &variable, Comparison: intPtr(0)},
		},
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &targetTrigger},
			{Op: "display_instructions", Message: "uses inline text"},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	trigger := file.root.section("Triggers").list("trigger_data")
	added := trigger[len(trigger)-1]
	if err := setIntField(added.list("effect_data")[1], "string_id", "s32", stringID); err != nil {
		t.Fatalf("set string_id: %v", err)
	}
	report, err := file.References(ReferenceOptions{})
	if err != nil {
		t.Fatalf("References: %v", err)
	}
	if report.Summary.ByKind["unit"] == 0 || report.Summary.ByKind["trigger"] == 0 || report.Summary.ByKind["variable"] == 0 || report.Summary.ByKind["string"] == 0 {
		t.Fatalf("reference summary missing kinds: %+v", report.Summary.ByKind)
	}
	unitOnly, err := file.References(ReferenceOptions{Kind: "unit", TargetID: &unitRef})
	if err != nil {
		t.Fatalf("References filtered: %v", err)
	}
	if len(unitOnly.References) != 1 || unitOnly.References[0].Field != "unit_object" {
		t.Fatalf("filtered unit refs = %+v, want unit_object only", unitOnly.References)
	}
}

func TestRemoveUnitRejectsGarrisonReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	containerRef := 990701
	garrisonRef := 990702
	x := 18.5
	y := 18.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 82, X: &x, Y: &y, ReferenceID: &containerRef}); err != nil {
		t.Fatalf("AddUnit container: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &garrisonRef, GarrisonedInID: &containerRef}); err != nil {
		t.Fatalf("AddUnit garrisoned: %v", err)
	}
	err = file.RemoveUnit(UnitRecipe{Op: "remove_unit", ReferenceID: &containerRef})
	if err == nil || !strings.Contains(err.Error(), "garrisoned_in_id references it") {
		t.Fatalf("RemoveUnit error = %v, want garrisoned_in_id refusal", err)
	}
}

func TestScenarioDeletePlanUnitBlockedAndClear(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	blockedRef := 990801
	clearRef := 990802
	blockedCaption := "delete-plan-caption-blocked"
	clearCaption := "delete-plan-caption-clear"
	x := 19.5
	y := 19.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &blockedRef, CaptionString: blockedCaption}); err != nil {
		t.Fatalf("AddUnit blocked: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &clearRef, CaptionString: clearCaption}); err != nil {
		t.Fatalf("AddUnit clear: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks unit delete",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{blockedRef}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	blocked, err := file.DeletePlan(DeletePlanRequest{Kind: "unit", ID: blockedRef})
	if err != nil {
		t.Fatalf("DeletePlan blocked: %v", err)
	}
	if blocked.CanDelete || len(blocked.BlockingRefs) != 1 || !strings.Contains(blocked.Strategy, "blocked") {
		t.Fatalf("blocked delete plan = %+v, want one blocking ref", blocked)
	}
	if blocked.CleanupRecipe == nil || blocked.CleanupCommand == "" {
		t.Fatalf("blocked delete plan = %+v, want unit cleanup recipe/command", blocked)
	}
	blockedByCaption, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-caption", TargetCaption: blockedCaption, Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan blocked by caption: %v", err)
	}
	if blockedByCaption.CanDelete || len(blockedByCaption.BlockingRefs) != 1 || blockedByCaption.CleanupRecipe == nil || !strings.Contains(blockedByCaption.CleanupCommand, fmt.Sprintf("unit %d", blockedRef)) {
		t.Fatalf("blocked caption delete plan = %+v, want reference-id cleanup", blockedByCaption)
	}
	_, disconnectRecipe, err := file.UnitDisconnectRecipe(blockedRef)
	if err != nil {
		t.Fatalf("UnitDisconnectRecipe: %v", err)
	}
	if len(disconnectRecipe.Triggers) != 1 {
		t.Fatalf("unit disconnect trigger edits = %+v, want one edit", disconnectRecipe.Triggers)
	}
	clear, err := file.DeletePlan(DeletePlanRequest{Kind: "unit", ID: clearRef})
	if err != nil {
		t.Fatalf("DeletePlan clear: %v", err)
	}
	if !clear.CanDelete || clear.SuggestedRecipe == nil {
		t.Fatalf("clear delete plan = %+v, want suggested recipe", clear)
	}
	clearByCaption, err := file.DeletePlan(DeletePlanRequest{Kind: "unit_caption", TargetCaption: clearCaption, Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan clear by caption: %v", err)
	}
	if !clearByCaption.CanDelete || clearByCaption.SuggestedRecipe == nil || !strings.Contains(clearByCaption.Strategy, "caption selector") {
		t.Fatalf("clear caption delete plan = %+v, want caption recipe", clearByCaption)
	}
	captionData, err := json.Marshal(clearByCaption.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal caption delete recipe: %v", err)
	}
	if !strings.Contains(string(captionData), "target_caption") || !strings.Contains(string(captionData), clearCaption) {
		t.Fatalf("caption suggested recipe = %s, want target_caption", captionData)
	}
	prefixBlockedRef := 990805
	prefixClearRef1 := 990806
	prefixClearRef2 := 990807
	prefixX1, prefixY1 := 24.5, 24.5
	prefixX2, prefixY2 := 25.5, 25.5
	prefixBlockedCaption := "delete-plan-prefix-blocked alpha"
	prefixClearCaption := "delete-plan-prefix-clear "
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &prefixX1, Y: &prefixY1, ReferenceID: &prefixBlockedRef, CaptionString: prefixBlockedCaption}); err != nil {
		t.Fatalf("AddUnit prefix blocked: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &prefixX1, Y: &prefixY1, ReferenceID: &prefixClearRef1, CaptionString: prefixClearCaption + "alpha"}); err != nil {
		t.Fatalf("AddUnit prefix clear 1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &prefixX2, Y: &prefixY2, ReferenceID: &prefixClearRef2, CaptionString: prefixClearCaption + "beta"}); err != nil {
		t.Fatalf("AddUnit prefix clear 2: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks caption prefix unit delete",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{prefixBlockedRef}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger prefix blocker: %v", err)
	}
	prefixBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-caption-prefix", TargetPrefix: "delete-plan-prefix-blocked", Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan blocked unit-caption-prefix: %v", err)
	}
	if prefixBlocked.CanDelete || len(prefixBlocked.ResolvedIDs) != 1 || prefixBlocked.ResolvedIDs[0] != prefixBlockedRef ||
		len(prefixBlocked.BlockingRefs) != 1 || prefixBlocked.CleanupRecipe == nil ||
		!strings.Contains(prefixBlocked.CleanupCommand, "unit-caption-prefix") {
		t.Fatalf("blocked unit-caption-prefix plan = %+v, want resolved id cleanup", prefixBlocked)
	}
	resolvedPrefixRefs, prefixRefs, prefixRecipe, err := file.UnitCaptionPrefixDisconnectRecipe("delete-plan-prefix-blocked", intPtr(1))
	if err != nil {
		t.Fatalf("UnitCaptionPrefixDisconnectRecipe: %v", err)
	}
	if len(resolvedPrefixRefs) != 1 || resolvedPrefixRefs[0] != prefixBlockedRef ||
		prefixRefs.Summary.Total != 1 || len(prefixRecipe.Triggers) != 1 {
		t.Fatalf("unit-caption-prefix disconnect resolved=%v refs=%+v recipe=%+v, want one trigger cleanup", resolvedPrefixRefs, prefixRefs.Summary, prefixRecipe)
	}
	prefixClear, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-caption-prefix", TargetPrefix: prefixClearCaption, Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan clear unit-caption-prefix: %v", err)
	}
	if !prefixClear.CanDelete || prefixClear.SuggestedRecipe == nil || len(prefixClear.ResolvedIDs) != 2 ||
		prefixClear.ResolvedIDs[0] != prefixClearRef1 || prefixClear.ResolvedIDs[1] != prefixClearRef2 {
		t.Fatalf("clear unit-caption-prefix plan = %+v, want two resolved reference-id removes", prefixClear)
	}
	prefixData, err := json.Marshal(prefixClear.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal prefix delete recipe: %v", err)
	}
	if !strings.Contains(string(prefixData), `"reference_id":`+strconv.Itoa(prefixClearRef1)) ||
		!strings.Contains(string(prefixData), `"reference_id":`+strconv.Itoa(prefixClearRef2)) {
		t.Fatalf("unit-caption-prefix suggested recipe = %s, want reference_id removes", prefixData)
	}
	containsBlockedRef := 990811
	containsClearRef1 := 990812
	containsClearRef2 := 990813
	containsX1, containsY1 := 26.5, 24.5
	containsX2, containsY2 := 26.5, 25.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &containsX1, Y: &containsY1, ReferenceID: &containsBlockedRef, CaptionString: "blocked middle CAPMARK alpha"}); err != nil {
		t.Fatalf("AddUnit contains blocked: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &containsX1, Y: &containsY1, ReferenceID: &containsClearRef1, CaptionString: "clear middle CAPMARK alpha"}); err != nil {
		t.Fatalf("AddUnit contains clear 1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &containsX2, Y: &containsY2, ReferenceID: &containsClearRef2, CaptionString: "clear middle CAPMARK beta"}); err != nil {
		t.Fatalf("AddUnit contains clear 2: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks caption contains unit delete",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{containsBlockedRef}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger contains blocker: %v", err)
	}
	containsBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-caption-contains", TargetText: "blocked middle CAPMARK", Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan blocked unit-caption-contains: %v", err)
	}
	if containsBlocked.CanDelete || len(containsBlocked.ResolvedIDs) != 1 || containsBlocked.ResolvedIDs[0] != containsBlockedRef ||
		len(containsBlocked.BlockingRefs) != 1 || containsBlocked.CleanupRecipe == nil ||
		!strings.Contains(containsBlocked.CleanupCommand, "unit-caption-contains") {
		t.Fatalf("blocked unit-caption-contains plan = %+v, want resolved id cleanup", containsBlocked)
	}
	resolvedContainsRefs, containsRefs, containsRecipe, err := file.UnitCaptionContainsDisconnectRecipe("blocked middle CAPMARK", intPtr(1))
	if err != nil {
		t.Fatalf("UnitCaptionContainsDisconnectRecipe: %v", err)
	}
	if len(resolvedContainsRefs) != 1 || resolvedContainsRefs[0] != containsBlockedRef ||
		containsRefs.Summary.Total != 1 || len(containsRecipe.Triggers) != 1 {
		t.Fatalf("unit-caption-contains disconnect resolved=%v refs=%+v recipe=%+v, want one trigger cleanup", resolvedContainsRefs, containsRefs.Summary, containsRecipe)
	}
	containsClear, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-caption-contains", TargetText: "clear middle CAPMARK", Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan clear unit-caption-contains: %v", err)
	}
	if !containsClear.CanDelete || containsClear.SuggestedRecipe == nil || len(containsClear.ResolvedIDs) != 2 ||
		containsClear.ResolvedIDs[0] != containsClearRef1 || containsClear.ResolvedIDs[1] != containsClearRef2 {
		t.Fatalf("clear unit-caption-contains plan = %+v, want two resolved reference-id removes", containsClear)
	}
	containsData, err := json.Marshal(containsClear.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal contains delete recipe: %v", err)
	}
	if !strings.Contains(string(containsData), `"reference_id":`+strconv.Itoa(containsClearRef1)) ||
		!strings.Contains(string(containsData), `"reference_id":`+strconv.Itoa(containsClearRef2)) {
		t.Fatalf("unit-caption-contains suggested recipe = %s, want reference_id removes", containsData)
	}
	typeBlockedRef := 990808
	typeClearRef1 := 990809
	typeClearRef2 := 990810
	typeBlockedUnitConst := 997
	typeClearUnitConst := 998
	typeX1, typeY1 := 27.5, 27.5
	typeX2, typeY2 := 28.5, 28.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: typeBlockedUnitConst, X: &typeX1, Y: &typeY1, ReferenceID: &typeBlockedRef}); err != nil {
		t.Fatalf("AddUnit type blocked: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: typeClearUnitConst, X: &typeX1, Y: &typeY1, ReferenceID: &typeClearRef1}); err != nil {
		t.Fatalf("AddUnit type clear 1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: typeClearUnitConst, X: &typeX2, Y: &typeY2, ReferenceID: &typeClearRef2}); err != nil {
		t.Fatalf("AddUnit type clear 2: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks unit-type delete",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{typeBlockedRef}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger type blocker: %v", err)
	}
	typeBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-type", UnitConst: &typeBlockedUnitConst, Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan blocked unit-type: %v", err)
	}
	if typeBlocked.CanDelete || len(typeBlocked.ResolvedIDs) != 1 || typeBlocked.ResolvedIDs[0] != typeBlockedRef ||
		len(typeBlocked.BlockingRefs) != 1 || typeBlocked.CleanupRecipe == nil ||
		!strings.Contains(typeBlocked.CleanupCommand, "unit-type") {
		t.Fatalf("blocked unit-type plan = %+v, want resolved id cleanup", typeBlocked)
	}
	resolvedTypeRefs, typeRefs, typeRecipe, err := file.UnitTypeDisconnectRecipe(typeBlockedUnitConst, intPtr(1))
	if err != nil {
		t.Fatalf("UnitTypeDisconnectRecipe: %v", err)
	}
	if len(resolvedTypeRefs) != 1 || resolvedTypeRefs[0] != typeBlockedRef ||
		typeRefs.Summary.Total != 1 || len(typeRecipe.Triggers) != 1 {
		t.Fatalf("unit-type disconnect resolved=%v refs=%+v recipe=%+v, want one trigger cleanup", resolvedTypeRefs, typeRefs.Summary, typeRecipe)
	}
	typeClear, err := file.DeletePlan(DeletePlanRequest{Kind: "unit-type", UnitConst: &typeClearUnitConst, Player: intPtr(1)})
	if err != nil {
		t.Fatalf("DeletePlan clear unit-type: %v", err)
	}
	if !typeClear.CanDelete || typeClear.SuggestedRecipe == nil || len(typeClear.ResolvedIDs) != 2 ||
		typeClear.ResolvedIDs[0] != typeClearRef1 || typeClear.ResolvedIDs[1] != typeClearRef2 {
		t.Fatalf("clear unit-type plan = %+v, want two resolved reference-id removes", typeClear)
	}
	typeData, err := json.Marshal(typeClear.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal unit-type delete recipe: %v", err)
	}
	if !strings.Contains(string(typeData), `"reference_id":`+strconv.Itoa(typeClearRef1)) ||
		!strings.Contains(string(typeData), `"reference_id":`+strconv.Itoa(typeClearRef2)) {
		t.Fatalf("unit-type suggested recipe = %s, want reference_id removes", typeData)
	}
	playerSections, err := file.unitPlayerSections()
	if err != nil {
		t.Fatalf("unitPlayerSections: %v", err)
	}
	var emptyPlayers []int
	for player, section := range playerSections {
		if player == 0 {
			continue
		}
		if len(section.list("units")) == 0 {
			emptyPlayers = append(emptyPlayers, player)
		}
	}
	if len(emptyPlayers) < 2 {
		t.Skipf("need two empty player slots for units-player test, found %v", emptyPlayers)
	}
	playerBlocked := emptyPlayers[0]
	playerClear := emptyPlayers[1]
	playerBlockedRef := 990831
	playerClearRef1 := 990832
	playerClearRef2 := 990833
	playerUnitConst := 999
	playerX1, playerY1 := 30.5, 30.5
	playerX2, playerY2 := 31.5, 31.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: playerBlocked, UnitConst: playerUnitConst, X: &playerX1, Y: &playerY1, ReferenceID: &playerBlockedRef}); err != nil {
		t.Fatalf("AddUnit player blocked: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: playerClear, UnitConst: playerUnitConst, X: &playerX1, Y: &playerY1, ReferenceID: &playerClearRef1}); err != nil {
		t.Fatalf("AddUnit player clear 1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: playerClear, UnitConst: playerUnitConst, X: &playerX2, Y: &playerY2, ReferenceID: &playerClearRef2}); err != nil {
		t.Fatalf("AddUnit player clear 2: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks units-player delete",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{playerBlockedRef}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger units-player blocker: %v", err)
	}
	playerBlockedPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "units-player", Player: &playerBlocked})
	if err != nil {
		t.Fatalf("DeletePlan blocked units-player: %v", err)
	}
	if playerBlockedPlan.CanDelete || len(playerBlockedPlan.ResolvedIDs) != 1 || playerBlockedPlan.ResolvedIDs[0] != playerBlockedRef ||
		len(playerBlockedPlan.BlockingRefs) != 1 || playerBlockedPlan.CleanupRecipe == nil ||
		!strings.Contains(playerBlockedPlan.CleanupCommand, "units-player") {
		t.Fatalf("blocked units-player plan = %+v, want resolved id cleanup", playerBlockedPlan)
	}
	resolvedPlayerRefs, playerRefs, playerRecipe, err := file.UnitsPlayerDisconnectRecipe(playerBlocked)
	if err != nil {
		t.Fatalf("UnitsPlayerDisconnectRecipe: %v", err)
	}
	if len(resolvedPlayerRefs) != 1 || resolvedPlayerRefs[0] != playerBlockedRef ||
		playerRefs.Summary.Total != 1 || len(playerRecipe.Triggers) != 1 {
		t.Fatalf("units-player disconnect resolved=%v refs=%+v recipe=%+v, want one trigger cleanup", resolvedPlayerRefs, playerRefs.Summary, playerRecipe)
	}
	playerClearPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "units-player", Player: &playerClear})
	if err != nil {
		t.Fatalf("DeletePlan clear units-player: %v", err)
	}
	if !playerClearPlan.CanDelete || playerClearPlan.SuggestedRecipe == nil || len(playerClearPlan.ResolvedIDs) != 2 ||
		playerClearPlan.ResolvedIDs[0] != playerClearRef1 || playerClearPlan.ResolvedIDs[1] != playerClearRef2 {
		t.Fatalf("clear units-player plan = %+v, want two resolved reference ids", playerClearPlan)
	}
	playerData, err := json.Marshal(playerClearPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal units-player delete recipe: %v", err)
	}
	if !strings.Contains(string(playerData), `"remove_units_for_player"`) ||
		!strings.Contains(string(playerData), `"target_player":`+strconv.Itoa(playerClear)) {
		t.Fatalf("units-player suggested recipe = %s, want remove_units_for_player", playerData)
	}
	removed, err := file.RemoveUnitsForPlayer(UnitRecipe{Op: "remove_units_for_player", TargetPlayer: &playerClear})
	if err != nil {
		t.Fatalf("RemoveUnitsForPlayer: %v", err)
	}
	if removed != 2 {
		t.Fatalf("RemoveUnitsForPlayer removed %d, want 2", removed)
	}
	if matches, err := file.findUnitsByPlayer(&playerClear); err == nil {
		t.Fatalf("player %d still has units after RemoveUnitsForPlayer: %+v", playerClear, matches)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "unit", ID: 999999}); err == nil {
		t.Fatal("DeletePlan accepted nonexistent unit")
	}
	areaBlockedX1, areaBlockedY1, areaBlockedX2, areaBlockedY2 := 19.0, 19.0, 20.0, 20.0
	areaBlocked, err := file.DeletePlan(DeletePlanRequest{
		Kind:      "units-area",
		AreaX1:    &areaBlockedX1,
		AreaY1:    &areaBlockedY1,
		AreaX2:    &areaBlockedX2,
		AreaY2:    &areaBlockedY2,
		Player:    intPtr(1),
		UnitConst: intPtr(83),
	})
	if err != nil {
		t.Fatalf("DeletePlan blocked area: %v", err)
	}
	if areaBlocked.CanDelete || len(areaBlocked.BlockingRefs) != 1 || !strings.Contains(areaBlocked.Strategy, "area selector matched 2 placed unit") {
		t.Fatalf("blocked area delete plan = %+v, want one blocking ref across two matches", areaBlocked)
	}
	if areaBlocked.CleanupRecipe == nil || areaBlocked.CleanupCommand == "" || !strings.Contains(areaBlocked.CleanupCommand, "units-area") {
		t.Fatalf("blocked area cleanup command=%q recipe=%+v, want units-area cleanup", areaBlocked.CleanupCommand, areaBlocked.CleanupRecipe)
	}
	areaRef1 := 990803
	areaRef2 := 990804
	areaX1, areaY1 := 21.5, 21.5
	areaX2, areaY2 := 22.5, 22.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &areaX1, Y: &areaY1, ReferenceID: &areaRef1}); err != nil {
		t.Fatalf("AddUnit areaRef1: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &areaX2, Y: &areaY2, ReferenceID: &areaRef2}); err != nil {
		t.Fatalf("AddUnit areaRef2: %v", err)
	}
	areaClearX1, areaClearY1, areaClearX2, areaClearY2 := 21.0, 21.0, 23.0, 23.0
	areaClear, err := file.DeletePlan(DeletePlanRequest{
		Kind:      "units_area",
		AreaX1:    &areaClearX1,
		AreaY1:    &areaClearY1,
		AreaX2:    &areaClearX2,
		AreaY2:    &areaClearY2,
		Player:    intPtr(1),
		UnitConst: intPtr(83),
	})
	if err != nil {
		t.Fatalf("DeletePlan clear area: %v", err)
	}
	if !areaClear.CanDelete || areaClear.SuggestedRecipe == nil || !strings.Contains(areaClear.Strategy, "remove 2 placed unit") {
		t.Fatalf("clear area delete plan = %+v, want area recipe", areaClear)
	}
	data, err := json.Marshal(areaClear.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal area delete recipe: %v", err)
	}
	if !strings.Contains(string(data), "remove_units_in_area") || !strings.Contains(string(data), "target_area_x1") {
		t.Fatalf("area suggested recipe = %s, want remove_units_in_area bounds", data)
	}
}

func TestScenarioUnitDisconnectClearsTriggerAndGarrisonReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	containerRef := 990811
	garrisonRef := 990812
	containerCaption := "unit-disconnect-caption-container"
	x := 19.5
	y := 19.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 82, X: &x, Y: &y, ReferenceID: &containerRef, CaptionString: containerCaption}); err != nil {
		t.Fatalf("AddUnit container: %v", err)
	}
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &garrisonRef, GarrisonedInID: &containerRef}); err != nil {
		t.Fatalf("AddUnit garrisoned: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "unit disconnect source",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{containerRef}},
		},
		Conditions: []ConditionRecipe{
			{Op: "object_selected", UnitObject: &containerRef, SourcePlayer: intPtr(1)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger blocker: %v", err)
	}
	refs, recipe, err := file.UnitDisconnectRecipe(containerRef)
	if err != nil {
		t.Fatalf("UnitDisconnectRecipe: %v", err)
	}
	if refs.Summary.Total != 3 || len(recipe.Triggers) != 1 || len(recipe.Units) != 1 {
		t.Fatalf("disconnect refs=%+v recipe=%+v, want trigger edit plus garrison clear", refs.Summary, recipe)
	}
	resolvedRef, captionRefs, captionRecipe, err := file.UnitCaptionDisconnectRecipe(containerCaption, intPtr(1))
	if err != nil {
		t.Fatalf("UnitCaptionDisconnectRecipe: %v", err)
	}
	if resolvedRef != containerRef || captionRefs.Summary.Total != refs.Summary.Total || len(captionRecipe.Triggers) != 1 || len(captionRecipe.Units) != 1 {
		t.Fatalf("caption disconnect ref=%d refs=%+v recipe=%+v, want same unit cleanup", resolvedRef, captionRefs.Summary, captionRecipe)
	}
	if got := recipe.Triggers[0].RemoveEffects; len(got) != 1 || got[0] != 0 {
		t.Fatalf("remove effects = %v, want [0]", got)
	}
	if got := recipe.Triggers[0].RemoveConditions; len(got) != 1 || got[0] != 0 {
		t.Fatalf("remove conditions = %v, want [0]", got)
	}
	if recipe.Units[0].ReferenceID == nil || *recipe.Units[0].ReferenceID != garrisonRef || recipe.Units[0].GarrisonedInID == nil || *recipe.Units[0].GarrisonedInID != -1 {
		t.Fatalf("unit cleanup = %+v, want garrison clear by source reference id", recipe.Units[0])
	}
	if err := file.EditTrigger(recipe.Triggers[0]); err != nil {
		t.Fatalf("EditTrigger cleanup: %v", err)
	}
	if err := file.EditUnit(recipe.Units[0]); err != nil {
		t.Fatalf("EditUnit cleanup: %v", err)
	}
	clear, err := file.DeletePlan(DeletePlanRequest{Kind: "unit", ID: containerRef})
	if err != nil {
		t.Fatalf("DeletePlan after cleanup: %v", err)
	}
	if !clear.CanDelete || clear.ReferenceSummary.Total != 0 {
		t.Fatalf("delete plan after cleanup = %+v, want clean physical delete", clear)
	}
}

func TestScenarioDeletePlanTriggerAndUnsupportedKinds(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	targetTrigger := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "delete-plan trigger target"}); err != nil {
		t.Fatalf("AddTrigger target: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks trigger delete",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &targetTrigger},
		},
	}); err != nil {
		t.Fatalf("AddTrigger blocker: %v", err)
	}
	blocked, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger", ID: targetTrigger})
	if err != nil {
		t.Fatalf("DeletePlan trigger blocked: %v", err)
	}
	if blocked.CanDelete || !blocked.CanTombstone || len(blocked.BlockingRefs) != 1 || blocked.SuggestedRecipe == nil || blocked.CleanupRecipe == nil || blocked.CleanupCommand == "" {
		t.Fatalf("trigger blocked plan = %+v, want physical block plus tombstone and cleanup recipes", blocked)
	}
	data, err := json.Marshal(blocked.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal tombstone recipe: %v", err)
	}
	if !strings.Contains(string(data), "tombstone_trigger") {
		t.Fatalf("suggested recipe = %s, want tombstone_trigger", data)
	}
	cleanupData, err := json.Marshal(blocked.CleanupRecipe)
	if err != nil {
		t.Fatalf("marshal cleanup recipe: %v", err)
	}
	if !strings.Contains(string(cleanupData), "edit_trigger") || !strings.Contains(string(cleanupData), "remove_effects") || !strings.Contains(blocked.CleanupCommand, "scen disconnect") {
		t.Fatalf("cleanup command=%q recipe=%s, want trigger disconnect cleanup", blocked.CleanupCommand, cleanupData)
	}
	blockedByName, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-name", TargetName: "delete-plan trigger target"})
	if err != nil {
		t.Fatalf("DeletePlan trigger-name blocked: %v", err)
	}
	if blockedByName.CanDelete || !blockedByName.CanTombstone || blockedByName.CleanupRecipe == nil || !strings.Contains(blockedByName.Strategy, fmt.Sprintf("index %d", targetTrigger)) {
		t.Fatalf("trigger-name blocked plan = %+v, want resolved-index tombstone", blockedByName)
	}
	resolvedTrigger, nameRefs, nameRecipe, err := file.TriggerNameDisconnectRecipe("delete-plan trigger target")
	if err != nil {
		t.Fatalf("TriggerNameDisconnectRecipe: %v", err)
	}
	if resolvedTrigger != targetTrigger || nameRefs.Summary.Total != blocked.ReferenceSummary.Total || len(nameRecipe.Triggers) != 1 {
		t.Fatalf("trigger-name disconnect resolved=%d refs=%+v recipe=%+v, want same trigger cleanup", resolvedTrigger, nameRefs.Summary, nameRecipe)
	}
	clear, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger", ID: targetTrigger + 1})
	if err != nil {
		t.Fatalf("DeletePlan trigger clear: %v", err)
	}
	if !clear.CanDelete || clear.SuggestedRecipe == nil {
		t.Fatalf("trigger clear plan = %+v, want recipe", clear)
	}
	clearByName, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger_name", TargetName: "blocks trigger delete"})
	if err != nil {
		t.Fatalf("DeletePlan trigger-name clear: %v", err)
	}
	if !clearByName.CanDelete || clearByName.SuggestedRecipe == nil || !strings.Contains(clearByName.Strategy, "unique name selector") {
		t.Fatalf("trigger-name clear plan = %+v, want name recipe", clearByName)
	}
	scriptCallTrigger := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "A2K Script Call Delete",
		Effects: []EffectRecipe{{
			Op:      "script_call",
			Message: "A2K_DeleteProbe();",
		}},
	}); err != nil {
		t.Fatalf("AddTrigger script_call: %v", err)
	}
	scriptCallPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger", ID: scriptCallTrigger})
	if err != nil {
		t.Fatalf("DeletePlan script_call trigger: %v", err)
	}
	if !scriptCallPlan.CanDelete || !strings.Contains(strings.Join(scriptCallPlan.StructuralCaveats, "\n"), "script_call") {
		t.Fatalf("script_call trigger plan = %+v, want deletable plan with semantic caveat", scriptCallPlan)
	}
	duplicateName := "delete-plan duplicate trigger name"
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: duplicateName}); err != nil {
		t.Fatalf("AddTrigger duplicate A: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: duplicateName}); err != nil {
		t.Fatalf("AddTrigger duplicate B: %v", err)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-name", TargetName: duplicateName}); err == nil {
		t.Fatal("DeletePlan trigger-name accepted duplicate trigger names")
	}
	prefixStart := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "A2K Prefix Delete: alpha"}); err != nil {
		t.Fatalf("AddTrigger prefix A: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "A2K Prefix Delete: beta"}); err != nil {
		t.Fatalf("AddTrigger prefix B: %v", err)
	}
	prefixPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-prefix", TargetPrefix: "A2K Prefix Delete:"})
	if err != nil {
		t.Fatalf("DeletePlan trigger-prefix clear: %v", err)
	}
	if !prefixPlan.CanDelete || prefixPlan.SuggestedRecipe == nil || len(prefixPlan.ResolvedIDs) != 2 ||
		prefixPlan.ResolvedIDs[0] != prefixStart || prefixPlan.ResolvedIDs[1] != prefixStart+1 {
		t.Fatalf("trigger-prefix clear plan = %+v, want two resolved physical removes", prefixPlan)
	}
	data, err = json.Marshal(prefixPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal trigger-prefix remove recipe: %v", err)
	}
	if !strings.Contains(string(data), `"remove_triggers"`) ||
		!strings.Contains(string(data), strconv.Itoa(prefixStart)) ||
		!strings.Contains(string(data), strconv.Itoa(prefixStart+1)) {
		t.Fatalf("trigger-prefix remove recipe = %s, want remove_triggers target indexes", data)
	}
	internalPrefixStart := len(file.root.section("Triggers").list("trigger_data"))
	internalPrefixTarget := internalPrefixStart + 1
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "A2K Prefix Internal Delete: source",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &internalPrefixTarget},
		},
	}); err != nil {
		t.Fatalf("AddTrigger internal prefix source: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "A2K Prefix Internal Delete: target"}); err != nil {
		t.Fatalf("AddTrigger internal prefix target: %v", err)
	}
	internalPrefixPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-prefix", TargetPrefix: "A2K Prefix Internal Delete:"})
	if err != nil {
		t.Fatalf("DeletePlan trigger-prefix internal refs: %v", err)
	}
	if !internalPrefixPlan.CanDelete || internalPrefixPlan.ReferenceSummary.Total != 0 ||
		len(internalPrefixPlan.ResolvedIDs) != 2 ||
		!strings.Contains(strings.Join(internalPrefixPlan.StructuralCaveats, "\n"), "internal trigger-control reference") {
		t.Fatalf("trigger-prefix internal plan = %+v, want physical delete with internal-ref caveat only", internalPrefixPlan)
	}
	referencedPrefixTarget := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "A2K Prefix Tombstone: target"}); err != nil {
		t.Fatalf("AddTrigger referenced prefix target: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks trigger-prefix delete",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &referencedPrefixTarget},
		},
	}); err != nil {
		t.Fatalf("AddTrigger prefix blocker: %v", err)
	}
	prefixBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-prefix", TargetPrefix: "A2K Prefix Tombstone:"})
	if err != nil {
		t.Fatalf("DeletePlan trigger-prefix blocked: %v", err)
	}
	if prefixBlocked.CanDelete || !prefixBlocked.CanTombstone || len(prefixBlocked.ResolvedIDs) != 1 ||
		prefixBlocked.CleanupRecipe == nil || prefixBlocked.CleanupCommand == "" ||
		!strings.Contains(prefixBlocked.Strategy, "semantic tombstones") {
		t.Fatalf("trigger-prefix blocked plan = %+v, want tombstone plus cleanup", prefixBlocked)
	}
	resolvedPrefix, prefixRefs, prefixCleanup, err := file.TriggerPrefixDisconnectRecipe("A2K Prefix Tombstone:")
	if err != nil {
		t.Fatalf("TriggerPrefixDisconnectRecipe: %v", err)
	}
	if len(resolvedPrefix) != 1 || resolvedPrefix[0] != referencedPrefixTarget ||
		prefixRefs.Summary.Total != 1 || len(prefixCleanup.Triggers) != 1 {
		t.Fatalf("trigger-prefix disconnect resolved=%v refs=%+v cleanup=%+v, want one trigger cleanup", resolvedPrefix, prefixRefs.Summary, prefixCleanup)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger", ID: targetTrigger + 20}); err == nil {
		t.Fatal("DeletePlan accepted nonexistent trigger")
	}
	containsA := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "contains delete target by name MARKER",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "no marker here"},
		},
	}); err != nil {
		t.Fatalf("AddTrigger contains A: %v", err)
	}
	containsB := containsA + 1
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "contains delete target by effect",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "MARKER in effect text"},
			{Op: "activate_trigger", TriggerID: &containsA},
		},
	}); err != nil {
		t.Fatalf("AddTrigger contains B: %v", err)
	}
	containsPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-contains", TargetText: "MARKER"})
	if err != nil {
		t.Fatalf("DeletePlan trigger-contains: %v", err)
	}
	if !containsPlan.CanDelete || containsPlan.ReferenceSummary.Total != 0 || len(containsPlan.ResolvedIDs) != 2 {
		t.Fatalf("trigger-contains plan = %+v, want clean two-trigger batch", containsPlan)
	}
	var containsRecipe Recipe
	containsData, err := json.Marshal(containsPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal trigger-contains recipe: %v", err)
	}
	if err := json.Unmarshal(containsData, &containsRecipe); err != nil {
		t.Fatalf("unmarshal trigger-contains recipe: %v", err)
	}
	if err := file.ApplyRecipe(containsRecipe); err != nil {
		t.Fatalf("ApplyRecipe trigger-contains: %v", err)
	}
	if _, err := file.findUniqueTriggerIndexByName("contains delete target by name MARKER"); err == nil {
		t.Fatal("trigger-contains left name-matched trigger behind")
	}
	if _, err := file.findUniqueTriggerIndexByName("contains delete target by effect"); err == nil {
		t.Fatalf("trigger-contains left effect-matched trigger behind at original index %d", containsB)
	}
	blockedContains := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "blocked contains MARKER2"}); err != nil {
		t.Fatalf("AddTrigger blocked contains: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "outside contains blocker",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &blockedContains},
		},
	}); err != nil {
		t.Fatalf("AddTrigger outside contains blocker: %v", err)
	}
	blockedContainsPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "trigger-contains", TargetText: "MARKER2"})
	if err != nil {
		t.Fatalf("DeletePlan blocked trigger-contains: %v", err)
	}
	if blockedContainsPlan.CanDelete || !blockedContainsPlan.CanTombstone || blockedContainsPlan.ReferenceSummary.Total != 1 || blockedContainsPlan.CleanupRecipe == nil {
		t.Fatalf("blocked trigger-contains plan = %+v, want tombstone plus cleanup", blockedContainsPlan)
	}
	variableID := 9903
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &variableID, Name: "DeletePlanVar"}); err != nil {
		t.Fatalf("AddVariable: %v", err)
	}
	variable, err := file.DeletePlan(DeletePlanRequest{Kind: "variable", ID: variableID})
	if err != nil {
		t.Fatalf("DeletePlan variable: %v", err)
	}
	if !variable.CanDelete || variable.SuggestedRecipe == nil || !strings.Contains(variable.Strategy, "remove unreferenced variable record") {
		t.Fatalf("variable plan = %+v, want remove recipe", variable)
	}
	variableByName, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-name", TargetName: "DeletePlanVar"})
	if err != nil {
		t.Fatalf("DeletePlan variable-name: %v", err)
	}
	if !variableByName.CanDelete || variableByName.SuggestedRecipe == nil || !strings.Contains(variableByName.Strategy, fmt.Sprintf("id %d", variableID)) {
		t.Fatalf("variable-name plan = %+v, want resolved variable recipe", variableByName)
	}
	prefixVarA := 99041
	prefixVarB := 99042
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &prefixVarA, Name: "A2K Prefix Var: alpha"}); err != nil {
		t.Fatalf("AddVariable prefix A: %v", err)
	}
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &prefixVarB, Name: "A2K Prefix Var: beta"}); err != nil {
		t.Fatalf("AddVariable prefix B: %v", err)
	}
	variablePrefix, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-prefix", TargetPrefix: "A2K Prefix Var:"})
	if err != nil {
		t.Fatalf("DeletePlan variable-prefix: %v", err)
	}
	if !variablePrefix.CanDelete || variablePrefix.SuggestedRecipe == nil || len(variablePrefix.ResolvedIDs) != 2 ||
		variablePrefix.ResolvedIDs[0] != prefixVarA || variablePrefix.ResolvedIDs[1] != prefixVarB {
		t.Fatalf("variable-prefix plan = %+v, want two resolved variable deletes", variablePrefix)
	}
	data, err = json.Marshal(variablePrefix.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal variable-prefix remove recipe: %v", err)
	}
	if !strings.Contains(string(data), `"target_id":`+strconv.Itoa(prefixVarA)) ||
		!strings.Contains(string(data), `"target_id":`+strconv.Itoa(prefixVarB)) {
		t.Fatalf("variable-prefix remove recipe = %s, want target ids", data)
	}
	referencedPrefixVar := 99043
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &referencedPrefixVar, Name: "A2K Referenced Prefix Var: target"}); err != nil {
		t.Fatalf("AddVariable referenced prefix: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks variable-prefix delete",
		Conditions: []ConditionRecipe{
			{Op: "variable_value", Variable: &referencedPrefixVar, Comparison: intPtr(0), Quantity: intPtr(1)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger variable prefix blocker: %v", err)
	}
	variablePrefixBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-prefix", TargetPrefix: "A2K Referenced Prefix Var:"})
	if err != nil {
		t.Fatalf("DeletePlan blocked variable-prefix: %v", err)
	}
	if variablePrefixBlocked.CanDelete || variablePrefixBlocked.CleanupRecipe == nil || variablePrefixBlocked.CleanupCommand == "" ||
		len(variablePrefixBlocked.ResolvedIDs) != 1 || variablePrefixBlocked.ResolvedIDs[0] != referencedPrefixVar {
		t.Fatalf("blocked variable-prefix plan = %+v, want blocked cleanup by resolved id", variablePrefixBlocked)
	}
	resolvedVarPrefix, varPrefixRefs, varPrefixCleanup, err := file.VariablePrefixDisconnectRecipe("A2K Referenced Prefix Var:")
	if err != nil {
		t.Fatalf("VariablePrefixDisconnectRecipe: %v", err)
	}
	if len(resolvedVarPrefix) != 1 || resolvedVarPrefix[0] != referencedPrefixVar ||
		varPrefixRefs.Summary.Total != 1 || len(varPrefixCleanup.Triggers) != 1 {
		t.Fatalf("variable-prefix disconnect resolved=%v refs=%+v cleanup=%+v, want one trigger cleanup", resolvedVarPrefix, varPrefixRefs.Summary, varPrefixCleanup)
	}
	containsVarA := 99044
	containsVarB := 99045
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &containsVarA, Name: "alpha A2K Contains Var target"}); err != nil {
		t.Fatalf("AddVariable contains A: %v", err)
	}
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &containsVarB, Name: "beta A2K Contains Var target"}); err != nil {
		t.Fatalf("AddVariable contains B: %v", err)
	}
	variableContains, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-contains", TargetText: "A2K Contains Var"})
	if err != nil {
		t.Fatalf("DeletePlan variable-contains: %v", err)
	}
	if !variableContains.CanDelete || variableContains.SuggestedRecipe == nil || len(variableContains.ResolvedIDs) != 2 ||
		variableContains.ResolvedIDs[0] != containsVarA || variableContains.ResolvedIDs[1] != containsVarB {
		t.Fatalf("variable-contains plan = %+v, want two resolved variable deletes", variableContains)
	}
	data, err = json.Marshal(variableContains.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal variable-contains remove recipe: %v", err)
	}
	if !strings.Contains(string(data), `"target_id":`+strconv.Itoa(containsVarA)) ||
		!strings.Contains(string(data), `"target_id":`+strconv.Itoa(containsVarB)) {
		t.Fatalf("variable-contains remove recipe = %s, want target ids", data)
	}
	referencedContainsVar := 99046
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &referencedContainsVar, Name: "prefix A2K Referenced Contains Var suffix"}); err != nil {
		t.Fatalf("AddVariable referenced contains: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "blocks variable-contains delete",
		Conditions: []ConditionRecipe{
			{Op: "variable_value", Variable: &referencedContainsVar, Comparison: intPtr(0), Quantity: intPtr(1)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger variable contains blocker: %v", err)
	}
	variableContainsBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-contains", TargetText: "A2K Referenced Contains Var"})
	if err != nil {
		t.Fatalf("DeletePlan blocked variable-contains: %v", err)
	}
	if variableContainsBlocked.CanDelete || variableContainsBlocked.CleanupRecipe == nil || variableContainsBlocked.CleanupCommand == "" ||
		len(variableContainsBlocked.ResolvedIDs) != 1 || variableContainsBlocked.ResolvedIDs[0] != referencedContainsVar {
		t.Fatalf("blocked variable-contains plan = %+v, want blocked cleanup by resolved id", variableContainsBlocked)
	}
	resolvedVarContains, varContainsRefs, varContainsCleanup, err := file.VariableContainsDisconnectRecipe("A2K Referenced Contains Var")
	if err != nil {
		t.Fatalf("VariableContainsDisconnectRecipe: %v", err)
	}
	if len(resolvedVarContains) != 1 || resolvedVarContains[0] != referencedContainsVar ||
		varContainsRefs.Summary.Total != 1 || len(varContainsCleanup.Triggers) != 1 {
		t.Fatalf("variable-contains disconnect resolved=%v refs=%+v cleanup=%+v, want one trigger cleanup", resolvedVarContains, varContainsRefs.Summary, varContainsCleanup)
	}
	duplicateVariableName := "DeletePlanDuplicateVar"
	dupA := 99031
	dupB := 99032
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &dupA, Name: duplicateVariableName}); err != nil {
		t.Fatalf("AddVariable duplicate A: %v", err)
	}
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &dupB, Name: duplicateVariableName}); err != nil {
		t.Fatalf("AddVariable duplicate B: %v", err)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-name", TargetName: duplicateVariableName}); err == nil {
		t.Fatal("DeletePlan variable-name accepted duplicate variable names")
	}
	stringID, err := file.AddString(StringRecipe{Op: "add_string", Text: "DeletePlanString"})
	if err != nil {
		t.Fatalf("AddString: %v", err)
	}
	stringPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "string", ID: stringID})
	if err != nil {
		t.Fatalf("DeletePlan string: %v", err)
	}
	if !stringPlan.CanDelete || stringPlan.SuggestedRecipe == nil || !strings.Contains(stringPlan.Strategy, "clear fixed string-table slot") {
		t.Fatalf("string plan = %+v, want clear recipe", stringPlan)
	}
	stringByText, err := file.DeletePlan(DeletePlanRequest{Kind: "string-text", TargetText: "DeletePlanString"})
	if err != nil {
		t.Fatalf("DeletePlan string-text: %v", err)
	}
	if !stringByText.CanDelete || stringByText.SuggestedRecipe == nil || !strings.Contains(stringByText.Strategy, fmt.Sprintf("slot %d", stringID)) {
		t.Fatalf("string-text plan = %+v, want clear recipe resolved to string id", stringByText)
	}
	data, err = json.Marshal(stringByText.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal string-text clear recipe: %v", err)
	}
	if !strings.Contains(string(data), "clear_string") || !strings.Contains(string(data), "DeletePlanString") {
		t.Fatalf("string-text suggested recipe = %s, want old_text clear recipe", data)
	}
	if _, err := file.AddString(StringRecipe{Op: "add_string", Text: "DeletePlanString"}); err != nil {
		t.Fatalf("AddString duplicate text: %v", err)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "string-text", TargetText: "DeletePlanString"}); err == nil {
		t.Fatal("DeletePlan string-text accepted duplicate exact text")
	}
	uniqueReferencedText := "DeletePlanStringRefsByText"
	referencedStringID, err := file.AddString(StringRecipe{Op: "add_string", Text: uniqueReferencedText})
	if err != nil {
		t.Fatalf("AddString referenced text: %v", err)
	}
	refID := 997701
	x := 10.5
	y := 11.5
	z := 0.0
	status := 2
	if err := file.AddUnit(UnitRecipe{
		Op:              "add_unit",
		Player:          1,
		UnitConst:       600,
		X:               &x,
		Y:               &y,
		Z:               &z,
		Status:          &status,
		ReferenceID:     &refID,
		CaptionStringID: &referencedStringID,
	}); err != nil {
		t.Fatalf("AddUnit caption ref: %v", err)
	}
	referencedString, err := file.DeletePlan(DeletePlanRequest{Kind: "string", ID: referencedStringID})
	if err != nil {
		t.Fatalf("DeletePlan referenced string: %v", err)
	}
	if referencedString.CanDelete || !referencedString.CanTombstone || len(referencedString.BlockingRefs) != 1 || referencedString.SuggestedRecipe == nil {
		t.Fatalf("referenced string plan = %+v, want physical block plus tombstone recipe", referencedString)
	}
	if referencedString.CleanupRecipe == nil || referencedString.CleanupCommand == "" {
		t.Fatalf("referenced string plan = %+v, want cleanup recipe/command", referencedString)
	}
	data, err = json.Marshal(referencedString.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal string tombstone recipe: %v", err)
	}
	if !strings.Contains(string(data), "TOMBSTONED string") {
		t.Fatalf("referenced string suggested recipe = %s, want tombstone text", data)
	}
	referencedStringByText, err := file.DeletePlan(DeletePlanRequest{Kind: "string-text", TargetText: uniqueReferencedText})
	if err != nil {
		t.Fatalf("DeletePlan referenced string-text: %v", err)
	}
	if referencedStringByText.CanDelete || !referencedStringByText.CanTombstone ||
		referencedStringByText.CleanupRecipe == nil || referencedStringByText.CleanupCommand == "" ||
		!strings.Contains(referencedStringByText.Strategy, fmt.Sprintf("id %d", referencedStringID)) {
		t.Fatalf("referenced string-text plan = %+v, want resolved tombstone plus cleanup", referencedStringByText)
	}
	prefixStringA, err := file.AddString(StringRecipe{Op: "add_string", Text: "A2K Prefix String: alpha"})
	if err != nil {
		t.Fatalf("AddString prefix A: %v", err)
	}
	prefixStringB, err := file.AddString(StringRecipe{Op: "add_string", Text: "A2K Prefix String: beta"})
	if err != nil {
		t.Fatalf("AddString prefix B: %v", err)
	}
	stringPrefixPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "string-prefix", TargetPrefix: "A2K Prefix String:"})
	if err != nil {
		t.Fatalf("DeletePlan string-prefix: %v", err)
	}
	if !stringPrefixPlan.CanDelete || stringPrefixPlan.SuggestedRecipe == nil || len(stringPrefixPlan.ResolvedIDs) != 2 ||
		stringPrefixPlan.ResolvedIDs[0] != prefixStringA || stringPrefixPlan.ResolvedIDs[1] != prefixStringB {
		t.Fatalf("string-prefix plan = %+v, want two resolved clear_string recipes", stringPrefixPlan)
	}
	data, err = json.Marshal(stringPrefixPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal string-prefix clear recipe: %v", err)
	}
	if !strings.Contains(string(data), `"id":`+strconv.Itoa(prefixStringA)) ||
		!strings.Contains(string(data), `"id":`+strconv.Itoa(prefixStringB)) {
		t.Fatalf("string-prefix clear recipe = %s, want target ids", data)
	}
	referencedPrefixString, err := file.AddString(StringRecipe{Op: "add_string", Text: "A2K Referenced Prefix String: target"})
	if err != nil {
		t.Fatalf("AddString referenced prefix: %v", err)
	}
	refID2 := 997702
	if err := file.AddUnit(UnitRecipe{
		Op:              "add_unit",
		Player:          1,
		UnitConst:       600,
		X:               &x,
		Y:               &y,
		Z:               &z,
		Status:          &status,
		ReferenceID:     &refID2,
		CaptionStringID: &referencedPrefixString,
	}); err != nil {
		t.Fatalf("AddUnit prefix string caption ref: %v", err)
	}
	stringPrefixBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "string-prefix", TargetPrefix: "A2K Referenced Prefix String:"})
	if err != nil {
		t.Fatalf("DeletePlan referenced string-prefix: %v", err)
	}
	if stringPrefixBlocked.CanDelete || !stringPrefixBlocked.CanTombstone || stringPrefixBlocked.CleanupRecipe == nil ||
		stringPrefixBlocked.CleanupCommand == "" || len(stringPrefixBlocked.ResolvedIDs) != 1 ||
		stringPrefixBlocked.ResolvedIDs[0] != referencedPrefixString {
		t.Fatalf("referenced string-prefix plan = %+v, want tombstone plus cleanup by resolved id", stringPrefixBlocked)
	}
	resolvedStringPrefix, stringPrefixRefs, stringPrefixCleanup, err := file.StringPrefixDisconnectRecipe("A2K Referenced Prefix String:")
	if err != nil {
		t.Fatalf("StringPrefixDisconnectRecipe: %v", err)
	}
	if len(resolvedStringPrefix) != 1 || resolvedStringPrefix[0] != referencedPrefixString ||
		stringPrefixRefs.Summary.Total != 1 || len(stringPrefixCleanup.Units) != 1 {
		t.Fatalf("string-prefix disconnect resolved=%v refs=%+v cleanup=%+v, want one unit cleanup", resolvedStringPrefix, stringPrefixRefs.Summary, stringPrefixCleanup)
	}
	containsStringA, err := file.AddString(StringRecipe{Op: "add_string", Text: "alpha A2K Contains String target"})
	if err != nil {
		t.Fatalf("AddString contains A: %v", err)
	}
	containsStringB, err := file.AddString(StringRecipe{Op: "add_string", Text: "beta A2K Contains String target"})
	if err != nil {
		t.Fatalf("AddString contains B: %v", err)
	}
	stringContainsPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "string-contains", TargetText: "A2K Contains String"})
	if err != nil {
		t.Fatalf("DeletePlan string-contains: %v", err)
	}
	if !stringContainsPlan.CanDelete || stringContainsPlan.SuggestedRecipe == nil || len(stringContainsPlan.ResolvedIDs) != 2 ||
		stringContainsPlan.ResolvedIDs[0] != containsStringA || stringContainsPlan.ResolvedIDs[1] != containsStringB {
		t.Fatalf("string-contains plan = %+v, want two resolved clear_string recipes", stringContainsPlan)
	}
	data, err = json.Marshal(stringContainsPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal string-contains clear recipe: %v", err)
	}
	if !strings.Contains(string(data), `"id":`+strconv.Itoa(containsStringA)) ||
		!strings.Contains(string(data), `"id":`+strconv.Itoa(containsStringB)) {
		t.Fatalf("string-contains clear recipe = %s, want target ids", data)
	}
	referencedContainsString, err := file.AddString(StringRecipe{Op: "add_string", Text: "prefix A2K Referenced Contains String suffix"})
	if err != nil {
		t.Fatalf("AddString referenced contains: %v", err)
	}
	refID3 := 997703
	if err := file.AddUnit(UnitRecipe{
		Op:              "add_unit",
		Player:          1,
		UnitConst:       600,
		X:               &x,
		Y:               &y,
		Z:               &z,
		Status:          &status,
		ReferenceID:     &refID3,
		CaptionStringID: &referencedContainsString,
	}); err != nil {
		t.Fatalf("AddUnit contains string caption ref: %v", err)
	}
	stringContainsBlocked, err := file.DeletePlan(DeletePlanRequest{Kind: "string-contains", TargetText: "A2K Referenced Contains String"})
	if err != nil {
		t.Fatalf("DeletePlan referenced string-contains: %v", err)
	}
	if stringContainsBlocked.CanDelete || !stringContainsBlocked.CanTombstone || stringContainsBlocked.CleanupRecipe == nil ||
		stringContainsBlocked.CleanupCommand == "" || len(stringContainsBlocked.ResolvedIDs) != 1 ||
		stringContainsBlocked.ResolvedIDs[0] != referencedContainsString {
		t.Fatalf("referenced string-contains plan = %+v, want tombstone plus cleanup by resolved id", stringContainsBlocked)
	}
	resolvedStringContains, stringContainsRefs, stringContainsCleanup, err := file.StringContainsDisconnectRecipe("A2K Referenced Contains String")
	if err != nil {
		t.Fatalf("StringContainsDisconnectRecipe: %v", err)
	}
	if len(resolvedStringContains) != 1 || resolvedStringContains[0] != referencedContainsString ||
		stringContainsRefs.Summary.Total != 1 || len(stringContainsCleanup.Units) != 1 {
		t.Fatalf("string-contains disconnect resolved=%v refs=%+v cleanup=%+v, want one unit cleanup", resolvedStringContains, stringContainsRefs.Summary, stringContainsCleanup)
	}
}

func TestScenarioStringDisconnectClearsDirectReferences(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	stringID, err := file.AddString(StringRecipe{Op: "add_string", Text: "DeletePlanStringRefs"})
	if err != nil {
		t.Fatalf("AddString: %v", err)
	}
	refID := 990805
	x := 10.5
	y := 11.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 600, X: &x, Y: &y, ReferenceID: &refID, CaptionStringID: &stringID}); err != nil {
		t.Fatalf("AddUnit caption ref: %v", err)
	}
	triggerIndex := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "string disconnect source",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "uses string id"},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	trigger := file.root.section("Triggers").list("trigger_data")[triggerIndex]
	if err := setIntField(trigger, "description_string_table_id", "s32", stringID); err != nil {
		t.Fatalf("set description_string_table_id: %v", err)
	}
	if err := setIntField(trigger, "short_description_string_table_id", "s32", stringID); err != nil {
		t.Fatalf("set short_description_string_table_id: %v", err)
	}
	if err := setIntField(trigger.list("effect_data")[0], "string_id", "s32", stringID); err != nil {
		t.Fatalf("set effect string_id: %v", err)
	}
	refs, cleanup, err := file.StringDisconnectRecipe(stringID)
	if err != nil {
		t.Fatalf("StringDisconnectRecipe: %v", err)
	}
	if refs.Summary.Total != 4 || len(cleanup.Triggers) != 1 || len(cleanup.Units) != 1 {
		t.Fatalf("string disconnect refs=%+v cleanup=%+v, want one trigger edit plus one unit edit", refs.Summary, cleanup)
	}
	resolvedStringID, textRefs, textCleanup, err := file.StringTextDisconnectRecipe("DeletePlanStringRefs")
	if err != nil {
		t.Fatalf("StringTextDisconnectRecipe: %v", err)
	}
	if resolvedStringID != stringID || textRefs.Summary.Total != refs.Summary.Total ||
		len(textCleanup.Triggers) != len(cleanup.Triggers) || len(textCleanup.Units) != len(cleanup.Units) {
		t.Fatalf("string-text disconnect resolved=%d refs=%+v cleanup=%+v, want same string cleanup", resolvedStringID, textRefs.Summary, textCleanup)
	}
	triggerCleanup := cleanup.Triggers[0]
	if triggerCleanup.DescriptionStringID == nil || *triggerCleanup.DescriptionStringID != -1 ||
		triggerCleanup.ShortDescriptionStringID == nil || *triggerCleanup.ShortDescriptionStringID != -1 {
		t.Fatalf("trigger cleanup = %+v, want both description ids cleared", triggerCleanup)
	}
	if len(triggerCleanup.RemoveEffects) != 1 || triggerCleanup.RemoveEffects[0] != 0 {
		t.Fatalf("remove effects = %v, want [0]", triggerCleanup.RemoveEffects)
	}
	if cleanup.Units[0].ReferenceID == nil || *cleanup.Units[0].ReferenceID != refID ||
		cleanup.Units[0].CaptionStringID == nil || *cleanup.Units[0].CaptionStringID != -1 {
		t.Fatalf("unit cleanup = %+v, want caption string clear", cleanup.Units[0])
	}
	if err := file.EditTrigger(triggerCleanup); err != nil {
		t.Fatalf("EditTrigger string cleanup: %v", err)
	}
	if err := file.EditUnit(cleanup.Units[0]); err != nil {
		t.Fatalf("EditUnit string cleanup: %v", err)
	}
	clearPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "string", ID: stringID})
	if err != nil {
		t.Fatalf("DeletePlan string after cleanup: %v", err)
	}
	if !clearPlan.CanDelete || clearPlan.ReferenceSummary.Total != 0 {
		t.Fatalf("string plan after cleanup = %+v, want clean clear", clearPlan)
	}
}

func TestAddEditTombstoneAndRemoveVariableRecipes(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := file.Triggers.Variables
	id := 9901
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &id, Name: "DexTemp"}); err != nil {
		t.Fatalf("AddVariable: %v", err)
	}
	if file.Triggers.Variables != before+1 {
		t.Fatalf("variable count = %d, want %d", file.Triggers.Variables, before+1)
	}
	setName := "DexRenamed"
	if err := file.EditVariable(VariableRecipe{Op: "edit_variable", TargetID: &id, SetName: &setName}); err != nil {
		t.Fatalf("EditVariable: %v", err)
	}
	variable, err := file.findVariableByID(id)
	if err != nil {
		t.Fatalf("findVariableByID: %v", err)
	}
	name, _ := variable.stringValue("variable_name")
	if name != setName {
		t.Fatalf("variable name = %q, want %q", name, setName)
	}
	if err := file.TombstoneVariable(VariableRecipe{Op: "tombstone_variable", TargetName: setName}); err != nil {
		t.Fatalf("TombstoneVariable: %v", err)
	}
	variable, err = file.findVariableByID(id)
	if err != nil {
		t.Fatalf("find tombstoned variable: %v", err)
	}
	name, _ = variable.stringValue("variable_name")
	if name != "_DeletedVariable9901" {
		t.Fatalf("tombstone name = %q, want _DeletedVariable9901", name)
	}
	if file.Triggers.Variables != before+1 {
		t.Fatalf("tombstone changed variable count to %d", file.Triggers.Variables)
	}
	id2 := 9904
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &id2, Name: "DexRemoveMe"}); err != nil {
		t.Fatalf("AddVariable remove target: %v", err)
	}
	if err := file.RemoveVariable(VariableRecipe{Op: "remove_variable", TargetID: &id2}); err != nil {
		t.Fatalf("RemoveVariable: %v", err)
	}
	if _, err := file.findVariableByID(id2); err == nil {
		t.Fatalf("removed variable id %d is still present", id2)
	}
	if file.Triggers.Variables != before+1 {
		t.Fatalf("remove changed variable count to %d, want %d", file.Triggers.Variables, before+1)
	}
	out := filepath.Join(t.TempDir(), "removed_variable.aoe2scenario")
	if err := file.Write(out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	reopened, err := Open(out)
	if err != nil {
		t.Fatalf("Open written scenario: %v", err)
	}
	if err := reopened.VerifyRebuild(); err != nil {
		t.Fatalf("VerifyRebuild: %v", err)
	}
	if _, err := reopened.findVariableByID(id); err != nil {
		t.Fatalf("surviving tombstoned variable id %d missing after reopen: %v", id, err)
	}
	if _, err := reopened.findVariableByID(id2); err == nil {
		t.Fatalf("removed variable id %d returned after reopen", id2)
	}
}

func TestPlanAndPatchRejectReferencedVariableTombstone(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	id := 9902
	name := "DexReferenced"
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &id, Name: name}); err != nil {
		t.Fatalf("AddVariable: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "variable blocker",
		Conditions: []ConditionRecipe{
			{Op: "variable_value", Variable: &id, Comparison: intPtr(0), Quantity: intPtr(7)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	recipe := Recipe{Variables: []VariableRecipe{{Op: "tombstone_variable", TargetID: &id}}}
	if _, err := file.Plan(recipe); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("Plan tombstone error = %v, want referenced refusal", err)
	}
	if err := file.TombstoneVariable(VariableRecipe{Op: "tombstone_variable", TargetID: &id}); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("TombstoneVariable error = %v, want referenced refusal", err)
	}
	removeRecipe := Recipe{Variables: []VariableRecipe{{Op: "remove_variable", TargetID: &id}}}
	if _, err := file.Plan(removeRecipe); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("Plan remove error = %v, want referenced refusal", err)
	}
	if err := file.RemoveVariable(VariableRecipe{Op: "remove_variable", TargetID: &id}); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("RemoveVariable error = %v, want referenced refusal", err)
	}
	plan, err := file.DeletePlan(DeletePlanRequest{Kind: "variable", ID: id})
	if err != nil {
		t.Fatalf("DeletePlan referenced variable: %v", err)
	}
	if plan.CanDelete || plan.CleanupRecipe == nil || plan.CleanupCommand == "" {
		t.Fatalf("referenced variable plan = %+v, want cleanup recipe/command and blocked delete", plan)
	}
	planByName, err := file.DeletePlan(DeletePlanRequest{Kind: "variable-name", TargetName: name})
	if err != nil {
		t.Fatalf("DeletePlan referenced variable-name: %v", err)
	}
	if planByName.CanDelete || planByName.CleanupRecipe == nil || !strings.Contains(planByName.CleanupCommand, fmt.Sprintf("variable %d", id)) {
		t.Fatalf("referenced variable-name plan = %+v, want cleanup by resolved id", planByName)
	}
	refs, cleanup, err := file.VariableDisconnectRecipe(id)
	if err != nil {
		t.Fatalf("VariableDisconnectRecipe: %v", err)
	}
	resolvedID, nameRefs, nameCleanup, err := file.VariableNameDisconnectRecipe(name)
	if err != nil {
		t.Fatalf("VariableNameDisconnectRecipe: %v", err)
	}
	if resolvedID != id || nameRefs.Summary.Total != refs.Summary.Total || len(nameCleanup.Triggers) != 1 {
		t.Fatalf("variable-name disconnect id=%d refs=%+v cleanup=%+v, want same variable cleanup", resolvedID, nameRefs.Summary, nameCleanup)
	}
	if refs.Summary.Total != 1 || len(cleanup.Triggers) != 1 || len(cleanup.Triggers[0].RemoveConditions) != 1 {
		t.Fatalf("variable disconnect refs=%+v cleanup=%+v, want one removed condition", refs.Summary, cleanup)
	}
	if err := file.EditTrigger(cleanup.Triggers[0]); err != nil {
		t.Fatalf("EditTrigger variable cleanup: %v", err)
	}
	cleanPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "variable", ID: id})
	if err != nil {
		t.Fatalf("DeletePlan after variable cleanup: %v", err)
	}
	if !cleanPlan.CanDelete || cleanPlan.ReferenceSummary.Total != 0 {
		t.Fatalf("variable plan after cleanup = %+v, want clean physical delete", cleanPlan)
	}
}

func TestScenarioDeletePlanEffectAndConditionRows(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	triggerIndex := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "child delete target",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "keep"},
			{Op: "display_instructions", Message: "delete"},
		},
		Conditions: []ConditionRecipe{
			{Op: "timer", Timer: intPtr(3)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	effectIndex := 1
	effectPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "effect", TriggerID: &triggerIndex, ChildIndex: &effectIndex})
	if err != nil {
		t.Fatalf("DeletePlan effect: %v", err)
	}
	if !effectPlan.CanDelete || effectPlan.SuggestedRecipe == nil || !strings.Contains(effectPlan.Strategy, "effect row 1") {
		t.Fatalf("effect plan = %+v, want child-row removal recipe", effectPlan)
	}
	conditionIndex := 0
	conditionPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "condition", TriggerID: &triggerIndex, ChildIndex: &conditionIndex})
	if err != nil {
		t.Fatalf("DeletePlan condition: %v", err)
	}
	if !conditionPlan.CanDelete || conditionPlan.SuggestedRecipe == nil || !strings.Contains(conditionPlan.Strategy, "condition row 0") {
		t.Fatalf("condition plan = %+v, want child-row removal recipe", conditionPlan)
	}
	if err := file.EditTrigger(TriggerRecipe{Op: "edit_trigger", TargetIndex: &triggerIndex, RemoveEffects: []int{effectIndex}, RemoveConditions: []int{conditionIndex}}); err != nil {
		t.Fatalf("EditTrigger remove children: %v", err)
	}
	trigger, err := file.triggerByIndex(triggerIndex)
	if err != nil {
		t.Fatalf("triggerByIndex: %v", err)
	}
	if got := len(trigger.list("effect_data")); got != 1 {
		t.Fatalf("effect count = %d, want 1", got)
	}
	if got := len(trigger.list("condition_data")); got != 0 {
		t.Fatalf("condition count = %d, want 0", got)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "effect", TriggerID: &triggerIndex, ChildIndex: &effectIndex}); err == nil {
		t.Fatal("DeletePlan accepted removed effect index")
	}

	typeTriggerA := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "type delete target A",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "delete display A1"},
			{Op: "send_chat", Message: "keep chat A"},
			{Op: "display_instructions", Message: "delete display A2"},
		},
		Conditions: []ConditionRecipe{
			{Op: "timer", Timer: intPtr(4)},
			{Op: "object_selected", UnitObject: intPtr(12345)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger type A: %v", err)
	}
	typeTriggerB := typeTriggerA + 1
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "type delete target B",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "delete display B"},
		},
		Conditions: []ConditionRecipe{
			{Op: "timer", Timer: intPtr(5)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger type B: %v", err)
	}
	effectTypePlan, err := file.DeletePlan(DeletePlanRequest{Kind: "effect-type", ID: 20, TargetName: "display_instructions", TargetPrefix: "type delete target"})
	if err != nil {
		t.Fatalf("DeletePlan effect-type: %v", err)
	}
	if !effectTypePlan.CanDelete || len(effectTypePlan.ResolvedIDs) != 6 || !strings.Contains(effectTypePlan.Strategy, "3 effect row(s)") {
		t.Fatalf("effect-type plan = %+v, want three matched effect rows", effectTypePlan)
	}
	var effectTypeRecipe Recipe
	effectTypeData, err := json.Marshal(effectTypePlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal effect-type recipe: %v", err)
	}
	if err := json.Unmarshal(effectTypeData, &effectTypeRecipe); err != nil {
		t.Fatalf("unmarshal effect-type recipe: %v", err)
	}
	if err := file.ApplyRecipe(effectTypeRecipe); err != nil {
		t.Fatalf("ApplyRecipe effect-type: %v", err)
	}
	triggerA, err := file.triggerByIndex(typeTriggerA)
	if err != nil {
		t.Fatalf("triggerByIndex A: %v", err)
	}
	triggerB, err := file.triggerByIndex(typeTriggerB)
	if err != nil {
		t.Fatalf("triggerByIndex B: %v", err)
	}
	if got := len(triggerA.list("effect_data")); got != 1 {
		t.Fatalf("type trigger A effect count = %d, want 1", got)
	}
	effectType, ok := triggerA.list("effect_data")[0].intValue("effect_type")
	if !ok {
		t.Fatal("type trigger A remaining effect missing effect_type")
	}
	if got := EffectTypeName(effectType); got != "send_chat" {
		t.Fatalf("type trigger A remaining effect = %q, want send_chat", got)
	}
	if got := len(triggerB.list("effect_data")); got != 0 {
		t.Fatalf("type trigger B effect count = %d, want 0", got)
	}
	if got := triggerA.intList("effect_display_order_array"); len(got) != 1 || got[0] != 0 {
		t.Fatalf("type trigger A effect order = %#v, want [0]", got)
	}

	conditionTypePlan, err := file.DeletePlan(DeletePlanRequest{Kind: "condition-type", ID: 10, TargetName: "timer", TriggerID: &typeTriggerA})
	if err != nil {
		t.Fatalf("DeletePlan condition-type: %v", err)
	}
	if !conditionTypePlan.CanDelete || len(conditionTypePlan.ResolvedIDs) != 2 || !strings.Contains(conditionTypePlan.Strategy, "1 condition row(s)") {
		t.Fatalf("condition-type plan = %+v, want one scoped condition row", conditionTypePlan)
	}
	var conditionTypeRecipe Recipe
	conditionTypeData, err := json.Marshal(conditionTypePlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal condition-type recipe: %v", err)
	}
	if err := json.Unmarshal(conditionTypeData, &conditionTypeRecipe); err != nil {
		t.Fatalf("unmarshal condition-type recipe: %v", err)
	}
	if err := file.ApplyRecipe(conditionTypeRecipe); err != nil {
		t.Fatalf("ApplyRecipe condition-type: %v", err)
	}
	if got := len(triggerA.list("condition_data")); got != 1 {
		t.Fatalf("type trigger A condition count = %d, want 1", got)
	}
	conditionType, ok := triggerA.list("condition_data")[0].intValue("condition_type")
	if !ok {
		t.Fatal("type trigger A remaining condition missing condition_type")
	}
	if got := ConditionTypeName(conditionType); got != "object_selected" {
		t.Fatalf("type trigger A remaining condition = %q, want object_selected", got)
	}
	if got := len(triggerB.list("condition_data")); got != 1 {
		t.Fatalf("type trigger B condition count = %d, want 1", got)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "effect-type", ID: 20, TriggerID: &typeTriggerA}); err == nil {
		t.Fatal("DeletePlan effect-type accepted scope with no remaining matching rows")
	}

	textTrigger := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "text prefix delete target",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "A2KTXT delete display"},
			{Op: "send_chat", Message: "keep chat"},
			{Op: "script_call", Message: "A2KTXT_delete_script();"},
		},
	}); err != nil {
		t.Fatalf("AddTrigger text prefix: %v", err)
	}
	textPrefixPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "effect-text-prefix", TargetText: "A2KTXT", TriggerID: &textTrigger})
	if err != nil {
		t.Fatalf("DeletePlan effect-text-prefix: %v", err)
	}
	if !textPrefixPlan.CanDelete || len(textPrefixPlan.ResolvedIDs) != 4 || !strings.Contains(textPrefixPlan.Strategy, "2 effect row(s)") {
		t.Fatalf("effect-text-prefix plan = %+v, want two scoped effect rows", textPrefixPlan)
	}
	var textPrefixRecipe Recipe
	textPrefixData, err := json.Marshal(textPrefixPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal effect-text-prefix recipe: %v", err)
	}
	if err := json.Unmarshal(textPrefixData, &textPrefixRecipe); err != nil {
		t.Fatalf("unmarshal effect-text-prefix recipe: %v", err)
	}
	if err := file.ApplyRecipe(textPrefixRecipe); err != nil {
		t.Fatalf("ApplyRecipe effect-text-prefix: %v", err)
	}
	textNode, err := file.triggerByIndex(textTrigger)
	if err != nil {
		t.Fatalf("triggerByIndex text prefix: %v", err)
	}
	if got := len(textNode.list("effect_data")); got != 1 {
		t.Fatalf("text-prefix trigger effect count = %d, want 1", got)
	}
	message, _ := textNode.list("effect_data")[0].stringValue("message")
	if message != "keep chat" {
		t.Fatalf("text-prefix remaining message = %q, want keep chat", message)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "effect-text-prefix", TargetText: "A2KTXT", TriggerID: &textTrigger}); err == nil {
		t.Fatal("DeletePlan effect-text-prefix accepted scope with no matching rows")
	}
	containsTextTrigger := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "text contains delete target",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "keep before"},
			{Op: "display_instructions", Message: "display with MIDMARK in middle"},
			{Op: "send_chat", Message: "chat with MIDMARK suffix"},
			{Op: "script_call", Message: "keep_after();"},
		},
	}); err != nil {
		t.Fatalf("AddTrigger text contains: %v", err)
	}
	textContainsPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "effect-text-contains", TargetText: "MIDMARK", TriggerID: &containsTextTrigger})
	if err != nil {
		t.Fatalf("DeletePlan effect-text-contains: %v", err)
	}
	if !textContainsPlan.CanDelete || len(textContainsPlan.ResolvedIDs) != 4 || !strings.Contains(textContainsPlan.Strategy, "2 effect row(s)") {
		t.Fatalf("effect-text-contains plan = %+v, want two scoped effect rows", textContainsPlan)
	}
	var textContainsRecipe Recipe
	textContainsData, err := json.Marshal(textContainsPlan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal effect-text-contains recipe: %v", err)
	}
	if err := json.Unmarshal(textContainsData, &textContainsRecipe); err != nil {
		t.Fatalf("unmarshal effect-text-contains recipe: %v", err)
	}
	if err := file.ApplyRecipe(textContainsRecipe); err != nil {
		t.Fatalf("ApplyRecipe effect-text-contains: %v", err)
	}
	containsTextNode, err := file.triggerByIndex(containsTextTrigger)
	if err != nil {
		t.Fatalf("triggerByIndex text contains: %v", err)
	}
	remainingMessages := []string{}
	for _, effect := range containsTextNode.list("effect_data") {
		message, _ := effect.stringValue("message")
		remainingMessages = append(remainingMessages, message)
	}
	if !reflect.DeepEqual(remainingMessages, []string{"keep before", "keep_after();"}) {
		t.Fatalf("text-contains remaining messages = %#v, want keep rows only", remainingMessages)
	}
	if _, err := file.DeletePlan(DeletePlanRequest{Kind: "effect-text-contains", TargetText: "MIDMARK", TriggerID: &containsTextTrigger}); err == nil {
		t.Fatal("DeletePlan effect-text-contains accepted scope with no matching rows")
	}
}

func TestScenarioDeletePlanSystemPrefix(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	prefix := "A2K System Clean:"
	stringID, err := file.AddString(StringRecipe{Op: "add_string", Text: prefix + " string"})
	if err != nil {
		t.Fatalf("AddString: %v", err)
	}
	variableID := 880101
	if err := file.AddVariable(VariableRecipe{Op: "add_variable", ID: &variableID, Name: prefix + " variable"}); err != nil {
		t.Fatalf("AddVariable: %v", err)
	}
	unitX, unitY := 24.0, 24.0
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 600, X: &unitX, Y: &unitY, CaptionString: prefix + " unit"}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	triggerB := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: prefix + " B"}); err != nil {
		t.Fatalf("AddTrigger B: %v", err)
	}
	triggerA := triggerB + 1
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: prefix + " A",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &triggerB},
			{Op: "change_variable", Variable: &variableID, Quantity: intPtr(1)},
		},
	}); err != nil {
		t.Fatalf("AddTrigger A: %v", err)
	}
	plan, err := file.DeletePlan(DeletePlanRequest{Kind: "system-prefix", TargetPrefix: prefix})
	if err != nil {
		t.Fatalf("DeletePlan system-prefix: %v", err)
	}
	if !plan.CanDelete || plan.ReferenceSummary.Total != 0 || plan.SuggestedRecipe == nil {
		t.Fatalf("system-prefix plan = %+v, want clean delete", plan)
	}
	if len(plan.ResolvedSets["triggers"]) != 2 || len(plan.ResolvedSets["variables"]) != 1 || len(plan.ResolvedSets["strings"]) != 1 || len(plan.ResolvedSets["units"]) != 1 {
		t.Fatalf("system-prefix resolved sets = %#v, want trigger/variable/string/unit matches", plan.ResolvedSets)
	}
	var recipe Recipe
	data, err := json.Marshal(plan.SuggestedRecipe)
	if err != nil {
		t.Fatalf("marshal system-prefix recipe: %v", err)
	}
	if err := json.Unmarshal(data, &recipe); err != nil {
		t.Fatalf("unmarshal system-prefix recipe: %v", err)
	}
	if err := file.ApplyRecipe(recipe); err != nil {
		t.Fatalf("ApplyRecipe system-prefix: %v", err)
	}
	if _, err := file.findVariableByID(variableID); err == nil {
		t.Fatal("system-prefix left matched variable behind")
	}
	values, err := file.scenarioStringValues()
	if err != nil {
		t.Fatalf("scenarioStringValues: %v", err)
	}
	if values[stringID] != "" {
		t.Fatalf("system-prefix string id %d = %q, want cleared", stringID, values[stringID])
	}
	if _, err := file.findUniqueTriggerIndexByName(prefix + " A"); err == nil {
		t.Fatalf("system-prefix left trigger A behind at original index %d", triggerA)
	}
	if _, err := file.findUniqueTriggerIndexByName(prefix + " B"); err == nil {
		t.Fatalf("system-prefix left trigger B behind")
	}
	if _, err := file.findUnitsByCaptionPrefix(prefix, nil); err == nil {
		t.Fatal("system-prefix left captioned unit behind")
	}

	blockedPrefix := "A2K System Blocked:"
	blockedTrigger := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: blockedPrefix + " target"}); err != nil {
		t.Fatalf("AddTrigger blocked target: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "outside system blocker",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &blockedTrigger},
		},
	}); err != nil {
		t.Fatalf("AddTrigger external blocker: %v", err)
	}
	blockedPlan, err := file.DeletePlan(DeletePlanRequest{Kind: "system-prefix", TargetPrefix: blockedPrefix})
	if err != nil {
		t.Fatalf("DeletePlan blocked system-prefix: %v", err)
	}
	if blockedPlan.CanDelete || blockedPlan.ReferenceSummary.Total != 1 || len(blockedPlan.BlockingRefs) != 1 {
		t.Fatalf("blocked system-prefix plan = %+v, want one external blocker", blockedPlan)
	}
}

func TestAddSetTombstoneAndClearStringRecipes(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	id, err := file.AddString(StringRecipe{Op: "add_string", Text: "DexStringSmoke"})
	if err != nil {
		t.Fatalf("AddString: %v", err)
	}
	values, err := file.scenarioStringValues()
	if err != nil {
		t.Fatalf("scenarioStringValues: %v", err)
	}
	if values[id] != "DexStringSmoke" {
		t.Fatalf("string %d = %q, want DexStringSmoke", id, values[id])
	}
	oldText := "DexStringSmoke"
	setText := "DexStringRenamed"
	if err := file.SetString(StringRecipe{Op: "set_string", ID: &id, OldText: oldText, SetText: &setText}); err != nil {
		t.Fatalf("SetString: %v", err)
	}
	values, _ = file.scenarioStringValues()
	if values[id] != setText {
		t.Fatalf("string %d after set = %q, want %q", id, values[id], setText)
	}
	if err := file.TombstoneString(StringRecipe{Op: "tombstone_string", ID: &id, OldText: setText}); err != nil {
		t.Fatalf("TombstoneString: %v", err)
	}
	values, _ = file.scenarioStringValues()
	if values[id] != fmt.Sprintf("_DeletedString%d", id) {
		t.Fatalf("string %d after tombstone = %q", id, values[id])
	}
	clearText := fmt.Sprintf("_DeletedString%d", id)
	if err := file.ClearString(StringRecipe{Op: "clear_string", OldText: clearText}); err != nil {
		t.Fatalf("ClearString: %v", err)
	}
	values, _ = file.scenarioStringValues()
	if values[id] != "" {
		t.Fatalf("string %d after clear = %q, want empty", id, values[id])
	}
	reused, err := file.AddString(StringRecipe{Op: "add_string", Text: "DexStringReuse"})
	if err != nil {
		t.Fatalf("AddString after clear: %v", err)
	}
	if reused != id {
		t.Fatalf("reused string id = %d, want cleared id %d", reused, id)
	}
}

func TestPlanAndPatchRejectReferencedStringTombstone(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	id := 7
	if err := file.SetString(StringRecipe{Op: "set_string", ID: &id, Text: "DexReferencedString"}); err != nil {
		t.Fatalf("SetString: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "string blocker",
		Effects: []EffectRecipe{
			{Op: "display_instructions", Message: "uses string id"},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	trigger := file.root.section("Triggers").list("trigger_data")
	added := trigger[len(trigger)-1]
	if err := setIntField(added.list("effect_data")[0], "string_id", "s32", id); err != nil {
		t.Fatalf("set string_id: %v", err)
	}
	recipe := Recipe{Strings: []StringRecipe{{Op: "tombstone_string", ID: &id}}}
	if _, err := file.Plan(recipe); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("Plan tombstone string error = %v, want referenced refusal", err)
	}
	if err := file.TombstoneString(StringRecipe{Op: "tombstone_string", ID: &id}); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("TombstoneString error = %v, want referenced refusal", err)
	}
	clearRecipe := Recipe{Strings: []StringRecipe{{Op: "clear_string", ID: &id}}}
	if _, err := file.Plan(clearRecipe); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("Plan clear string error = %v, want referenced refusal", err)
	}
	if err := file.ClearString(StringRecipe{Op: "clear_string", ID: &id}); err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("ClearString error = %v, want referenced refusal", err)
	}
}

func TestScenarioSettingsReport(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	report := file.Settings()
	if report.PlayerCount == 0 || len(report.Players) == 0 {
		t.Fatalf("empty settings players: %+v", report)
	}
	if len(report.Resources) == 0 || report.Players[0].Resources.Player != 0 {
		t.Fatalf("settings resources not wired into players: %+v", report.Players[0])
	}
	if len(report.Diplomacy.Matrix) == 0 || len(report.Players[0].Diplomacy) == 0 {
		t.Fatalf("settings diplomacy missing: %+v", report.Diplomacy)
	}
	if _, ok := report.Victory["conquest_required"]; !ok {
		t.Fatalf("settings victory missing conquest_required: %+v", report.Victory)
	}
	if report.Verification != "structure_verified_not_engine_verified" {
		t.Fatalf("verification = %q", report.Verification)
	}
}

func TestSetDiplomacyOptionsRecipe(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	lockTeams := true
	allowChoose := false
	randomStarts := true
	maxTeams := 2
	recipe := DiplomacyOptionsRecipe{
		LockTeams:               &lockTeams,
		AllowPlayersChooseTeams: &allowChoose,
		RandomStartPoints:       &randomStarts,
		MaxNumberOfTeams:        &maxTeams,
		AlliedVictory: []AlliedVictoryRecipe{
			{Player: 1, Enabled: true},
			{Player: 2, Enabled: false},
		},
	}
	if _, err := file.Plan(Recipe{DiplomacyOptions: &recipe}); err != nil {
		t.Fatalf("Plan diplomacy options: %v", err)
	}
	if err := file.SetDiplomacyOptions(recipe); err != nil {
		t.Fatalf("SetDiplomacyOptions: %v", err)
	}
	settings := file.Settings()
	if !settings.Diplomacy.LockTeams || settings.Diplomacy.AllowPlayersChooseTeams || !settings.Diplomacy.RandomStartPoints || settings.Diplomacy.MaxNumberOfTeams != maxTeams {
		t.Fatalf("diplomacy options readback = %+v", settings.Diplomacy)
	}
	if !settings.Diplomacy.AlliedVictory[1] || settings.Diplomacy.AlliedVictory[2] {
		t.Fatalf("allied victory readback = %+v", settings.Diplomacy.AlliedVictory[:4])
	}
	badMax := 17
	if err := file.SetDiplomacyOptions(DiplomacyOptionsRecipe{MaxNumberOfTeams: &badMax}); err == nil {
		t.Fatal("SetDiplomacyOptions accepted max_number_of_teams above 16")
	}
}

func TestSetPlayerRecipeExtendedFields(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	active := true
	human := false
	tribe := "Dex Tribe"
	civ := "MAYANS"
	lockCiv := true
	lockPersonality := true
	aiName := "PlaygroundHelper"
	aiType := 2
	recipe := PlayerRecipe{
		Player:           2,
		Active:           &active,
		Human:            &human,
		TribeName:        &tribe,
		Civilization:     &civ,
		LockCivilization: &lockCiv,
		LockPersonality:  &lockPersonality,
		AIName:           &aiName,
		AIType:           &aiType,
	}
	if _, err := file.Plan(Recipe{Players: []PlayerRecipe{recipe}}); err != nil {
		t.Fatalf("Plan player: %v", err)
	}
	if err := file.SetPlayer(recipe); err != nil {
		t.Fatalf("SetPlayer: %v", err)
	}
	settings := file.Settings()
	player := settings.Players[2]
	if !player.Active || player.Human || player.TribeName != tribe || player.Civilization != civ ||
		!player.LockCivilization || !player.LockPersonality || player.AIName != aiName || player.AIType != aiType {
		t.Fatalf("player settings readback = %+v", player)
	}
	out := filepath.Join(t.TempDir(), "player_fields.aoe2scenario")
	if err := file.Write(out); err != nil {
		t.Fatalf("Write: %v", err)
	}
	reopened, err := Open(out)
	if err != nil {
		t.Fatalf("Open written player fields: %v", err)
	}
	reopenedPlayer := reopened.Settings().Players[2]
	if reopenedPlayer.TribeName != tribe || reopenedPlayer.Civilization != civ || !reopenedPlayer.LockCivilization ||
		!reopenedPlayer.LockPersonality || reopenedPlayer.AIName != aiName || reopenedPlayer.AIType != aiType {
		t.Fatalf("reopened player settings = %+v", reopenedPlayer)
	}
}

func TestParserRejectsImpossibleRepeatsBeforeAllocation(t *testing.T) {
	p := parser{data: make([]byte, 4)}
	_, err := p.parseNode("Huge", SectionSpec{Retrievers: []RetrieverSpec{
		{Name: "huge", Type: "u32", Repeat: maxScenarioFieldRepeat + 1},
	}}, nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds parser safety limit") {
		t.Fatalf("huge repeat error = %v, want safety limit", err)
	}

	p = parser{data: make([]byte, 4)}
	_, err = p.parseNode("Overrun", SectionSpec{Retrievers: []RetrieverSpec{
		{Name: "too_many", Type: "u32", Repeat: 2},
	}}, nil)
	if err == nil || !strings.Contains(err.Error(), "exceeds remaining bytes") {
		t.Fatalf("overrun repeat error = %v, want remaining-bytes guard", err)
	}
}

func TestPlanRecipeValidatesRemoveUnitSafety(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	refID := 990901
	x := 20.5
	y := 20.5
	if err := file.AddUnit(UnitRecipe{Op: "add_unit", Player: 1, UnitConst: 83, X: &x, Y: &y, ReferenceID: &refID}); err != nil {
		t.Fatalf("AddUnit: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "plan blocker",
		Effects: []EffectRecipe{
			{Op: "kill_object", SourcePlayer: intPtr(1), SelectedObjectIDs: []int{refID}},
		},
	}); err != nil {
		t.Fatalf("AddTrigger: %v", err)
	}
	_, err = file.Plan(Recipe{Units: []UnitRecipe{{Op: "remove_unit", ReferenceID: &refID}}})
	if err == nil || !strings.Contains(err.Error(), "selected_object_ids[0] references it") {
		t.Fatalf("Plan remove_unit error = %v, want selected-object reference refusal", err)
	}
	missing := 999999
	_, err = file.Plan(Recipe{Units: []UnitRecipe{{Op: "remove_unit", ReferenceID: &missing}}})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Plan missing remove_unit error = %v, want not found", err)
	}
}

func TestPlanRecipeValidatesRemoveTriggerSafety(t *testing.T) {
	input := testfixtures.Path(t, "save-analysis/diagnostics/Replay_Transparency_Diagnostic_v2b_flagged.aoe2scenario")
	if _, err := os.Stat(input); err != nil {
		t.Skipf("diagnostic fixture unavailable: %v", err)
	}
	file, err := Open(input)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	targetTrigger := len(file.root.section("Triggers").list("trigger_data"))
	if err := file.AddTrigger(TriggerRecipe{Op: "add_trigger", Name: "plan trigger target"}); err != nil {
		t.Fatalf("AddTrigger target: %v", err)
	}
	if err := file.AddTrigger(TriggerRecipe{
		Op:   "add_trigger",
		Name: "plan trigger blocker",
		Effects: []EffectRecipe{
			{Op: "activate_trigger", TriggerID: &targetTrigger},
		},
	}); err != nil {
		t.Fatalf("AddTrigger blocker: %v", err)
	}
	_, err = file.Plan(Recipe{Triggers: []TriggerRecipe{{Op: "remove_trigger", TargetIndex: &targetTrigger}}})
	if err == nil || !strings.Contains(err.Error(), "blocked by 1 reference") {
		t.Fatalf("Plan remove_trigger error = %v, want blocker", err)
	}
}

func TestValidateMapRect(t *testing.T) {
	ok := MapRecipe{Op: "set_terrain_rect", X1: 1, Y1: 2, X2: 3, Y2: 4}
	if err := validateMapRect(ok, 10, 10); err != nil {
		t.Fatalf("validateMapRect valid: %v", err)
	}
	reversed := MapRecipe{Op: "set_terrain_rect", X1: 3, Y1: 2, X2: 1, Y2: 4}
	if err := validateMapRect(reversed, 10, 10); err == nil {
		t.Fatal("validateMapRect accepted reversed rectangle")
	}
	outside := MapRecipe{Op: "set_terrain_rect", X1: 8, Y1: 8, X2: 10, Y2: 9}
	if err := validateMapRect(outside, 10, 10); err == nil {
		t.Fatal("validateMapRect accepted out-of-bounds rectangle")
	}
}
