package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPortableIdentityLeaksInText(t *testing.T) {
	name := "Ry" + "an"
	leaks := portableIdentityLeaksInText("data/engine_facts.json", []byte("safe B"+name+"\nleak "+name+" here\n"+name+"ic safe\n"))
	if len(leaks) != 1 {
		t.Fatalf("identity leaks = %+v, want one exact target-name word", leaks)
	}
	if leaks[0] != "data/engine_facts.json:2: "+name {
		t.Fatalf("leak detail = %q", leaks[0])
	}
}

func TestPortableIdentityLeaksUsesHandoffBoundary(t *testing.T) {
	root := t.TempDir()
	name := "Ry" + "an"
	mustWritePortable(t, filepath.Join(root, "README.md"), "portable root\n")
	mustWritePortable(t, filepath.Join(root, "data", "engine_facts.json"), `{"statement":"`+name+` leaked here"}`)
	mustWritePortable(t, filepath.Join(root, "DEX_TASK_internal.md"), name+" is allowed in task notes\n")
	mustWritePortable(t, filepath.Join(root, "docs", "internal.md"), name+" is allowed in reduced-profile docs exclusion\n")
	mustWritePortable(t, filepath.Join(root, "LICENSE"), "Copyright "+name+"\n")

	leaks, err := portableIdentityLeaks(root)
	if err != nil {
		t.Fatalf("portableIdentityLeaks failed: %v", err)
	}
	if len(leaks) != 1 {
		t.Fatalf("identity leaks = %+v, want only data leak", leaks)
	}
	if leaks[0] != "data/engine_facts.json:1: "+name {
		t.Fatalf("leak detail = %q", leaks[0])
	}
}

func TestPortableIdentityLeaksUsesSelectedPublicBoundary(t *testing.T) {
	root := t.TempDir()
	name := "Ry" + "an"
	mustWritePortable(t, filepath.Join(root, "README.md"), "portable root\n")
	mustWritePortable(t, filepath.Join(root, "NOTICE.md"), name+" in public notice\n")
	mustWritePortable(t, filepath.Join(root, "docs", "REPLAY_TRANSPARENCY_INTERNAL.md"), name+" excluded diagnostic\n")

	leaks, err := portableIdentityLeaksForProfile(root, ProfilePublic)
	if err != nil {
		t.Fatalf("portableIdentityLeaksForProfile failed: %v", err)
	}
	if len(leaks) != 1 || leaks[0] != "NOTICE.md:1: "+name {
		t.Fatalf("public identity leaks = %+v, want only NOTICE leak", leaks)
	}
}

func TestPortableLocalPathLeaksUseHandoffBoundary(t *testing.T) {
	root := t.TempDir()
	localPath := "C:" + "\\Users\\" + "Example\\fixture.aoe2record"
	mustWritePortable(t, filepath.Join(root, "README.md"), "portable root\n")
	mustWritePortable(t, filepath.Join(root, "data", "engine_facts.json"), `{"statement":"`+localPath+`"}`)
	mustWritePortable(t, filepath.Join(root, "docs", "internal.md"), localPath+"\n")

	leaks, err := portableLocalPathLeaks(root)
	if err != nil {
		t.Fatalf("portableLocalPathLeaks failed: %v", err)
	}
	if len(leaks) != 1 {
		t.Fatalf("local path leaks = %+v, want only shipped data leak", leaks)
	}
	if leaks[0] != "data/engine_facts.json: C:\\Users\\" {
		t.Fatalf("leak detail = %q", leaks[0])
	}

	mustWritePortable(t, filepath.Join(root, "data", "engine_facts.json"), `{"statement":"clean"}`)
	leaks, err = portableLocalPathLeaks(root)
	if err != nil {
		t.Fatalf("portableLocalPathLeaks after clean shipped file failed: %v", err)
	}
	if len(leaks) != 0 {
		t.Fatalf("local path leaks = %+v, want non-shipped doc ignored", leaks)
	}
	excludedCount, err := portableExcludedLocalPathReferenceCount(root)
	if err != nil {
		t.Fatalf("portableExcludedLocalPathReferenceCount failed: %v", err)
	}
	if excludedCount != 1 {
		t.Fatalf("excluded local path reference count = %d, want 1", excludedCount)
	}
}

func mustWritePortable(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
