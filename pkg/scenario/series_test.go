package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScenarioSeriesPathsSortsDirectoryCopies(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"Clean v159.2.aoe2scenario", "Clean v159.1.aoe2scenario"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := scenarioSeriesPaths([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(paths[0]) != "Clean v159.1.aoe2scenario" || filepath.Base(paths[1]) != "Clean v159.2.aoe2scenario" {
		t.Fatalf("paths = %v", paths)
	}
}
