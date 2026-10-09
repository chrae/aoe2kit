package scenario

import "testing"

func TestEffectEditorSchemasCoverObservedTypes(t *testing.T) {
	rows := EffectEditorSchemas()
	if got, want := len(rows), 109; got != want {
		t.Fatalf("schema rows = %d, want %d", got, want)
	}
	for _, effectType := range []int{1, 32, 50, 69, 109} {
		schema, ok := EffectEditorSchemaForType(effectType)
		if !ok {
			t.Fatalf("schema %d missing", effectType)
		}
		if schema.Type != effectType || schema.Name == "" {
			t.Fatalf("schema %d = %#v", effectType, schema)
		}
	}
	for _, effectType := range []int{91, 92, 93, 94, 95, 96, 97} {
		schema, ok := EffectEditorSchemaForType(effectType)
		if !ok || schema.Offered {
			t.Fatalf("schema %d should be present but not offered: %#v, %t", effectType, schema, ok)
		}
	}
	got, _ := EffectEditorSchemaForType(32)
	if len(got.VisibleFields) == 0 || got.VisibleFields[0] != "Source Player" {
		t.Fatalf("Change Object Range fields = %#v", got.VisibleFields)
	}
	for _, field := range []string{"X1", "Y1", "X2", "Y2", "Max Units Affected"} {
		if got.RevealedBy[field] != "Set Area" {
			t.Fatalf("Change Object Range revealed_by[%q] = %q, want Set Area", field, got.RevealedBy[field])
		}
	}
}

func TestSummarizeEffectIncludesObservedPerTypeFields(t *testing.T) {
	effect := &parsedNode{Fields: []*parsedNode{
		{Name: "effect_type", Value: int32(32)},
		{Name: "max_units_affected", Value: int32(9)},
		{Name: "object_group", Value: int32(36)},
		{Name: "object_type", Value: int32(2)},
	}}
	summary := summarizeEffect(effect, EffectsOptions{})
	for key, want := range map[string]any{
		"max_units_affected": 9,
		"object_group":       36,
		"object_type":        2,
	} {
		if got := summary.KnownFields[key]; got != want {
			t.Fatalf("range known_fields[%s] = %#v, want %#v", key, got, want)
		}
	}
	if len(summary.EditorFields) == 0 {
		t.Fatal("range editor_fields is empty")
	}
	if summary.EditorFieldsRevealedBy["X1"] != "Set Area" {
		t.Fatalf("range revealed_by = %#v", summary.EditorFieldsRevealedBy)
	}

	ai := &parsedNode{Fields: []*parsedNode{
		{Name: "effect_type", Value: int32(50)},
		{Name: "ai_signal_value", Value: int32(7)},
	}}
	aiSummary := summarizeEffect(ai, EffectsOptions{})
	if got := aiSummary.KnownFields["ai_signal_value"]; got != 7 {
		t.Fatalf("AI signal known_fields = %#v, want 7", got)
	}
}
