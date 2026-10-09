package scenario

import "testing"

func TestCoverageVarianceClassifiesDistinctValues(t *testing.T) {
	for _, test := range []struct {
		name     string
		distinct int
		want     string
	}{
		{name: "empty", distinct: 0, want: "constant"},
		{name: "constant", distinct: 1, want: "constant"},
		{name: "variable", distinct: 2, want: "variable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := coverageVariance(test.distinct); got != test.want {
				t.Fatalf("coverageVariance(%d) = %q, want %q", test.distinct, got, test.want)
			}
		})
	}
}

func TestCanonicalCoveragePathNamesCalibratedOpaqueRecords(t *testing.T) {
	if got := canonicalCoveragePath("Units.player_data_3[2].unknown_structure_3[0]"); got != "Units.player_data_3[].custom_victory_condition_record_bytes[]" {
		t.Fatalf("canonical custom-victory path = %q", got)
	}
	if got := canonicalCoveragePath("PlayerDataTwo.ai_files[4].unknown"); got != "PlayerDataTwo.ai_files[].opaque_ai_file_prefix" {
		t.Fatalf("canonical AI prefix path = %q", got)
	}
}

func TestCanonicalCoveragePathNamesReservedTriggerBlock(t *testing.T) {
	if got := canonicalCoveragePath("Triggers.unknown_bytes"); got != "Triggers.reserved_trigger_block" {
		t.Fatalf("canonical reserved trigger path = %q", got)
	}
	if got := canonicalCoveragePath("Triggers.unknown_bytes2"); got != "Triggers.reserved_trigger_tail" {
		t.Fatalf("canonical reserved trigger tail = %q", got)
	}
}
