package scenario

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestScenarioDiffRawFallbackCalibrationPairs(t *testing.T) {
	dir := scenarioAoe2DEPath("runs", "cal159", "round3")
	openPair := func(before, after int) DiffReport {
		t.Helper()
		beforePath := filepath.Join(dir, "Clean v159."+strconv.Itoa(before)+".aoe2scenario")
		afterPath := filepath.Join(dir, "Clean v159."+strconv.Itoa(after)+".aoe2scenario")
		for _, path := range []string{beforePath, afterPath} {
			if _, err := os.Stat(path); err != nil {
				t.Skipf("calibration fixture unavailable: %v", err)
			}
		}
		report, err := DiffFiles(beforePath, afterPath)
		if err != nil {
			t.Fatalf("DiffFiles %d -> %d: %v", before, after, err)
		}
		return report
	}

	phantom := openPair(15, 16)
	if phantom.Same || !rawFieldHasPath(phantom, "Units.player_data_3[1].editor_camera_x") {
		t.Fatalf("phantom diff = same=%t raw=%+v", phantom.Same, phantom.RawFields)
	}
	saveCounter := openPair(17, 18)
	if saveCounter.Same || !rawFieldHasPath(saveCounter, "DataHeader.next_unit_id_to_place") {
		t.Fatalf("save-counter diff = same=%t raw=%+v", saveCounter.Same, saveCounter.RawFields)
	}

	rangeDiff := openPair(52, 53)
	if !diffHasField(rangeDiff, "trigger_0.effect_data[0].max_units_affected") {
		t.Fatalf("range diff missing decoded field: %+v", rangeDiff.Changes)
	}
}

func rawFieldHasPath(report DiffReport, want string) bool {
	for _, field := range report.RawFields {
		if field.Path == want {
			return true
		}
	}
	return false
}
