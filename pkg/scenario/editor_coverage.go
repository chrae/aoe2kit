package scenario

import "sort"

// EditorCoverageReport separates editor observations from byte evidence. The
// editor field tables describe controls; they do not, by themselves, prove a
// control's byte mapping.
type EditorCoverageReport struct {
	Path         string               `json:"path,omitempty"`
	Version      string               `json:"version"`
	Verification string               `json:"verification"`
	Effects      []EditorTypeCoverage `json:"effects"`
	Conditions   []EditorTypeCoverage `json:"conditions"`
}

type EditorTypeCoverage struct {
	Type              int                   `json:"type"`
	Name              string                `json:"name"`
	Occurrences       int                   `json:"occurrences"`
	ByteEvidence      string                `json:"byte_evidence"`
	EditorFields      []EditorFieldCoverage `json:"editor_fields"`
	KnownParserFields []string              `json:"known_parser_fields,omitempty"`
}

type EditorFieldCoverage struct {
	Name       string `json:"name"`
	RevealedBy string `json:"revealed_by,omitempty"`
	Status     string `json:"status"`
	Evidence   string `json:"evidence"`
}

// EditorCoverageFile reports the editor schemas against one scenario. A field
// is not called byte-mapped merely because its record parsed: only the type's
// structural presence is established here. Calibration evidence that binds a
// particular control to a byte belongs in the field decoder or a diff test.
func EditorCoverageFile(path string) (EditorCoverageReport, error) {
	f, err := Open(path)
	if err != nil {
		return EditorCoverageReport{}, err
	}
	report := EditorCoverageReport{
		Path:         path,
		Version:      f.Version,
		Verification: "editor_schema_observed_structural_bytes_verified_field_mappings_not_individually_verified",
	}

	effects := f.EffectsWithOptions(EffectsOptions{IncludeRawFields: true})
	effectCounts := make(map[int]int)
	effectFields := make(map[int]map[string]bool)
	for _, bucket := range effects.Types {
		effectCounts[bucket.Type] = bucket.Count
		for _, sample := range bucket.Sample {
			for field := range sample.KnownFields {
				if effectFields[bucket.Type] == nil {
					effectFields[bucket.Type] = map[string]bool{}
				}
				effectFields[bucket.Type][field] = true
			}
		}
	}
	for _, schema := range EffectEditorSchemas() {
		report.Effects = append(report.Effects, editorTypeCoverage(schema.Type, schema.Name, schema.VisibleFields, schema.RevealedBy, effectCounts[schema.Type], effectFields[schema.Type]))
	}

	conditionDumps, err := f.DumpTriggersWithConditions()
	if err != nil {
		return EditorCoverageReport{}, err
	}
	conditionCounts := make(map[int]int)
	conditionFields := make(map[int]map[string]bool)
	for _, trigger := range conditionDumps {
		for _, condition := range trigger.Conditions {
			typeID, ok := condition["type"].(int)
			if !ok {
				continue
			}
			conditionCounts[typeID]++
			for field := range condition {
				if field == "type" || field == "type_name" || field == "editor_fields" || field == "editor_fields_revealed_by" {
					continue
				}
				if conditionFields[typeID] == nil {
					conditionFields[typeID] = map[string]bool{}
				}
				conditionFields[typeID][field] = true
			}
		}
	}
	for _, schema := range ConditionEditorSchemas() {
		report.Conditions = append(report.Conditions, editorTypeCoverage(schema.Type, schema.Name, schema.VisibleFields, schema.RevealedBy, conditionCounts[schema.Type], conditionFields[schema.Type]))
	}
	return report, nil
}

func editorTypeCoverage(typeID int, name string, fields []string, revealedBy map[string]string, occurrences int, known map[string]bool) EditorTypeCoverage {
	byteEvidence := "type_not_observed_in_input"
	if occurrences > 0 {
		byteEvidence = "type_record_structurally_parsed"
	}
	row := EditorTypeCoverage{Type: typeID, Name: name, Occurrences: occurrences, ByteEvidence: byteEvidence}
	for _, field := range fields {
		row.EditorFields = append(row.EditorFields, EditorFieldCoverage{
			Name: field, RevealedBy: revealedBy[field], Status: "editor_visible_unmapped", Evidence: "owned_editor_ui_scan",
		})
	}
	for field := range known {
		row.KnownParserFields = append(row.KnownParserFields, field)
	}
	sort.Strings(row.KnownParserFields)
	return row
}
