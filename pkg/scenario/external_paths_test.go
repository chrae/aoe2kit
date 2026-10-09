package scenario

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"aoe2kit/pkg/enginefacts"
)

func requireFixture(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			t.Skipf("private editor fixture unavailable: %s", path)
		}
		t.Fatalf("stat fixture %s: %v", path, err)
	}
}

func requireFixtureDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("private fixture directory unavailable: %s", path)
		}
		t.Fatalf("stat fixture directory %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("fixture path is not a directory: %s", path)
	}
}

func requireEngineFacts(t *testing.T) {
	t.Helper()
	if _, err := enginefacts.LoadDefault(); err != nil {
		t.Skipf("private engine-facts ledger unavailable: %v", err)
	}
}

func scenarioProjectPath(parts ...string) string {
	_, source, _, _ := runtime.Caller(0)
	repo := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	args := append([]string{repo, "..", ".."}, parts...)
	return filepath.Join(args...)
}

func scenarioAoe2DEPath(parts ...string) string {
	_, source, _, _ := runtime.Caller(0)
	repo := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	args := append([]string{repo, ".."}, parts...)
	return filepath.Join(args...)
}
