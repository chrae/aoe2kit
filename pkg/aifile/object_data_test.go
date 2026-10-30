package aifile

import (
	"strings"
	"testing"
)

func TestObjectDataRelevantSelectorsHaveContextNotRuntimeClaims(t *testing.T) {
	r, err := LintPath("../../testdata/ai/object-data.per")
	if err != nil {
		t.Fatal(err)
	}
	if !r.OK || len(r.Issues) != 0 || r.Coverage.UncheckedForms != 0 || r.Verification.EngineVerified {
		t.Fatalf("report=%+v", r)
	}
	var constants map[string]ReferenceConstant
	for _, check := range r.Checks {
		if check.Command == "up-get-object-data" {
			constants = check.Parameters[0].Constants
			break
		}
	}
	if len(constants) != 22 {
		t.Fatalf("selector count=%d", len(constants))
	}
	for _, tt := range []struct {
		name  string
		value int64
	}{
		{"object-data-hitpoints", 10}, {"object-data-maxhp", 11}, {"object-data-range", 12},
		{"object-data-action", 5}, {"object-data-order", 6}, {"object-data-target", 7},
		{"object-data-target-id", 23}, {"object-data-distance", 44}, {"object-data-precise-distance", 45},
		{"object-data-reload-time", 54}, {"object-data-next-attack", 55}, {"object-data-min-range", 62},
		{"object-data-frame-delay", 78}, {"object-data-attack-delay", 90},
	} {
		entry, found := constants[tt.name]
		if !found || entry.Value != tt.value || entry.Source != objectDataSource || entry.ReferenceFamily == "" || entry.Meaning == "" || entry.Units == "" || entry.Context == "" || entry.RuntimeAvailability != "unverified" {
			t.Errorf("%s metadata=%+v", tt.name, entry)
		}
	}
	if !strings.Contains(constants["object-data-action"].Context, "formation") || !strings.Contains(constants["object-data-target"].Context, "UP -1 versus DE -2") || constants["object-data-target"].Units == constants["object-data-target-id"].Units {
		t.Fatal("target/formation contexts collapsed")
	}
	if !strings.Contains(constants["object-data-reload-time"].Context, "not the remaining") || !strings.Contains(constants["object-data-next-attack"].Context, "not a confirmed damage") || constants["object-data-frame-delay"].Units == constants["object-data-reload-time"].Units {
		t.Fatal("interval/countdown/frame timing collapsed")
	}
	delay := constants["object-data-attack-delay"]
	if !delay.DEOnly || !delay.NeedsDefinition || delay.ReferenceFamily != "DE" {
		t.Fatalf("DE extension=%+v", delay)
	}
}

func TestObjectDataDEExtensionRemainsExplicitAndOtherSelectorsPartial(t *testing.T) {
	r := lintPERText(t, "(defrule (true) => (up-get-object-data object-data-attack-delay 100))")
	issueWithCode(t, r, "unresolved_constant")
	if !r.OK {
		t.Fatalf("unverified builtin rejected: %+v", r.Issues)
	}
	r = lintPERText(t, "(defconst object-data-attack-delay 90) (defrule (true) => (up-get-object-data object-data-attack-delay 100))")
	if !r.OK || len(r.Issues) != 0 {
		t.Fatalf("explicit extension=%+v", r.Issues)
	}
	r = lintPERText(t, "(defrule (true) => (up-get-object-data 999 100))")
	issueWithCode(t, r, "enum_value_not_catalogued")
}
