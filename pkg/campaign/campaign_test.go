package campaign

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSamePartyCampaign(t *testing.T) {
	dir := t.TempDir()
	spec := &Spec{
		SchemaVersion:  1,
		Name:           "Test Campaign",
		WriterScenario: "Writer Scenario",
		Players:        2,
		Fields: []Field{
			{Name: "level", Type: "int", PerPlayer: true, Default: 1},
			{Name: "chapter", Type: "string", Default: "start"},
		},
	}
	report, err := Generate(spec, GenerateOptions{OutDir: dir, Prefix: "Camp"})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK || len(report.Files) != 4 {
		t.Fatalf("report = %+v", report)
	}
	reader, err := os.ReadFile(filepath.Join(dir, "campaign_reader.xs"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(reader)
	for _, want := range []string{
		`xsOpenFile("Writer Scenario")`,
		"Camp_p1_level = xsReadInt();",
		"Camp_p2_level = xsReadInt();",
		`xsChatData("A2KCAMPAIGN READ_OK")`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("reader missing %q:\n%s", want, text)
		}
	}
}

func TestGenerateRejectsBadFieldName(t *testing.T) {
	_, err := Generate(&Spec{
		Name:           "Bad",
		WriterScenario: "Writer",
		Players:        1,
		Fields:         []Field{{Name: "not valid", Type: "int"}},
	}, GenerateOptions{OutDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error")
	}
}
