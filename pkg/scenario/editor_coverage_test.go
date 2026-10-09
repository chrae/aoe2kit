package scenario

import (
	"path/filepath"
	"testing"
)

func TestEditorCoverageReportsSchemaAndEvidenceBoundary(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "editor-refs", "Create scenario, do nothing, click save.aoe2scenario")
	requireFixture(t, path)
	report, err := EditorCoverageFile(path)
	if err != nil {
		t.Fatalf("EditorCoverageFile: %v", err)
	}
	if report.Verification == "" || len(report.Effects) == 0 || len(report.Conditions) == 0 {
		t.Fatalf("incomplete editor coverage report: %+v", report)
	}
	rangeSchema, ok := findEditorTypeCoverage(report.Effects, 32)
	if !ok {
		t.Fatal("missing Change Object Range schema")
	}
	if len(rangeSchema.EditorFields) == 0 {
		t.Fatal("Change Object Range has no editor fields")
	}
	for _, field := range rangeSchema.EditorFields {
		if field.Name == "X1" && field.RevealedBy != "Set Area" {
			t.Fatalf("X1 reveal control = %q, want Set Area", field.RevealedBy)
		}
		if field.Status != "editor_visible_unmapped" {
			t.Fatalf("field status = %q, want explicit unmapped status", field.Status)
		}
	}
	if _, ok := findEditorTypeCoverage(report.Conditions, 20); !ok {
		t.Fatal("missing Chance condition schema")
	}
}

func findEditorTypeCoverage(rows []EditorTypeCoverage, typeID int) (EditorTypeCoverage, bool) {
	for _, row := range rows {
		if row.Type == typeID {
			return row, true
		}
	}
	return EditorTypeCoverage{}, false
}
