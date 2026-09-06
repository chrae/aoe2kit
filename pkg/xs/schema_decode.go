package xs

import (
	"fmt"
	"os"
	"strings"
)

type DataSchemaDecodeOptions struct {
	Schema string
}

type DataSchemaDecodeReport struct {
	Path         string                  `json:"path,omitempty"`
	OK           bool                    `json:"ok"`
	Schema       string                  `json:"schema"`
	Verification VerificationClaim       `json:"verification"`
	Header       map[string]any          `json:"header,omitempty"`
	Rows         []DataSchemaRow         `json:"rows,omitempty"`
	Footer       map[string]any          `json:"footer,omitempty"`
	Summary      DataSchemaDecodeSummary `json:"summary"`
	Warnings     []string                `json:"warnings,omitempty"`
	Errors       []string                `json:"errors,omitempty"`
}

type DataSchemaDecodeSummary struct {
	SizeBytes     int  `json:"size_bytes"`
	Rows          int  `json:"rows"`
	CompleteRows  int  `json:"complete_rows"`
	HasFooter     bool `json:"has_footer"`
	RemainingFrom int  `json:"remaining_from,omitempty"`
}

type DataSchemaRow struct {
	Index   int            `json:"index"`
	Offset  int            `json:"offset"`
	PhaseID int            `json:"phase_id,omitempty"`
	TimeS   int            `json:"time_s,omitempty"`
	Fields  map[string]any `json:"fields"`
}

func DecodeDataFileSchema(path string, opts DataSchemaDecodeOptions) (DataSchemaDecodeReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DataSchemaDecodeReport{}, err
	}
	report := DecodeDataBytesSchema(data, opts)
	report.Path = path
	return report, nil
}

func DecodeDataBytesSchema(data []byte, opts DataSchemaDecodeOptions) DataSchemaDecodeReport {
	schema := normalizeDataSchema(opts.Schema)
	report := DataSchemaDecodeReport{
		Schema: schema,
		Verification: StructureVerifiedClaim(
			"XS sidecar schema decode uses fixture-specific write order. It verifies byte structure and complete rows; engine meaning still depends on the scenario that produced the sidecar.",
		),
		Summary: DataSchemaDecodeSummary{SizeBytes: len(data)},
	}
	switch schema {
	case "rtv12-executioner":
		decodeRTV12Executioner(data, &report)
	case "rtv13-directed-attribution":
		decodeRTV13DirectedAttribution(data, &report)
	case "rtv14-castle-kill-calibration":
		decodeRTV14CastleKillCalibration(data, &report)
	case "rtv141-castle-kill-calibration":
		decodeRTV141CastleKillCalibration(data, &report)
	case "rtv15-castle-kill-ground-truth":
		decodeRTV15CastleKillGroundTruth(data, &report)
	case "rtv16-semantic-promotion":
		decodeRTV16SemanticPromotion(data, &report)
	case "rtv17-packed-test":
		decodeRTV17PackedTest(data, &report)
	case "a2ksem2-dat-command-semantics":
		decodeA2KSEM2DATCommandSemantics(data, &report)
	default:
		report.Errors = append(report.Errors, fmt.Sprintf("unsupported xsdat schema %q", opts.Schema))
	}
	report.Summary.Rows = len(report.Rows)
	report.Summary.CompleteRows = len(report.Rows)
	report.OK = len(report.Errors) == 0
	return report
}

func normalizeDataSchema(schema string) string {
	schema = strings.ToLower(strings.TrimSpace(schema))
	switch schema {
	case "rtv12", "v12", "v12-executioner", "rtv12-executioner", "executioner":
		return "rtv12-executioner"
	case "rtv13", "v13", "rtv13-directed", "rtv13-directed-attribution", "directed-attribution":
		return "rtv13-directed-attribution"
	case "rtv14", "v14", "rtv14-castle", "rtv14-castle-kill", "rtv14-castle-kill-calibration", "castle-kill-calibration":
		return "rtv14-castle-kill-calibration"
	case "rtv141", "v141", "v14.1", "rtv14.1", "rtv141-castle", "rtv141-castle-kill", "rtv141-castle-kill-calibration",
		"rtv142", "v142", "v14.2", "rtv14.2", "rtv142-castle", "rtv142-castle-kill", "rtv142-castle-kill-calibration":
		return "rtv141-castle-kill-calibration"
	case "rtv15", "v15", "rtv15-castle", "rtv15-castle-kill", "rtv15-castle-kill-ground-truth", "castle-kill-ground-truth":
		return "rtv15-castle-kill-ground-truth"
	case "rtv16", "v16", "rtv16-semantic", "rtv16-semantic-promotion", "semantic-promotion":
		return "rtv16-semantic-promotion"
	case "rtv17", "v17", "rtv17-packed", "rtv17-packed-test", "packed-test":
		return "rtv17-packed-test"
	case "a2ksem2", "datsem2", "dat-semantics-v2", "a2ksem2-dat-command-semantics":
		return "a2ksem2-dat-command-semantics"
	default:
		return schema
	}
}

var rtv12HeaderFields = []schemaField{
	{name: "magic", typ: "string"},
	{name: "version", typ: "int"},
	{name: "fixture", typ: "string"},
}

var rtv12RowFields = []schemaField{
	{name: "tag", typ: "string"},
	{name: "phase_id", typ: "int"},
	{name: "expected_kind", typ: "int"},
	{name: "family", typ: "int"},
	{name: "outcome", typ: "int"},
	{name: "xs_time", typ: "int"},
	{name: "p1_resource_sum_carrier_food_attr0", typ: "float"},
	{name: "p1_wood_family_attr1", typ: "float"},
	{name: "p1_gold_outcome_attr3", typ: "float"},
	{name: "p1_stone_kind_attr2", typ: "float"},
	{name: "p1_kills_attr20", typ: "float"},
	{name: "p1_razings_attr43", typ: "float"},
	{name: "p1_killed_by_others_attr154", typ: "float"},
	{name: "p1_razed_by_others_attr155", typ: "float"},
	{name: "p1_kill_value_attr170", typ: "float"},
	{name: "p1_raze_value_attr172", typ: "float"},
	{name: "p2_kills_attr20", typ: "float"},
	{name: "p2_razings_attr43", typ: "float"},
	{name: "p2_killed_by_others_attr154", typ: "float"},
	{name: "p2_razed_by_others_attr155", typ: "float"},
	{name: "p2_kill_value_attr170", typ: "float"},
	{name: "p2_raze_value_attr172", typ: "float"},
	{name: "p3_kills_attr20", typ: "float"},
	{name: "p3_razings_attr43", typ: "float"},
	{name: "p3_killed_by_others_attr154", typ: "float"},
	{name: "p3_razed_by_others_attr155", typ: "float"},
	{name: "p3_kill_value_attr170", typ: "float"},
	{name: "p3_raze_value_attr172", typ: "float"},
	{name: "p1_gaia_kills_attr300", typ: "float"},
	{name: "p1_player1_kills_attr301", typ: "float"},
	{name: "p1_player2_kills_attr302", typ: "float"},
	{name: "p1_player3_kills_attr303", typ: "float"},
	{name: "p1_kills_by_gaia_attr325", typ: "float"},
	{name: "p1_kills_by_player1_attr326", typ: "float"},
	{name: "p1_kills_by_player2_attr327", typ: "float"},
	{name: "p1_kills_by_player3_attr328", typ: "float"},
	{name: "p1_gaia_razings_attr350", typ: "float"},
	{name: "p1_player1_razings_attr351", typ: "float"},
	{name: "p1_player2_razings_attr352", typ: "float"},
	{name: "p1_player3_razings_attr353", typ: "float"},
	{name: "p1_razings_by_gaia_attr375", typ: "float"},
	{name: "p1_razings_by_player1_attr376", typ: "float"},
	{name: "p1_razings_by_player2_attr377", typ: "float"},
	{name: "p1_razings_by_player3_attr378", typ: "float"},
	{name: "p1_cobra_car_748_count", typ: "int"},
	{name: "p2_militia_74_count", typ: "int"},
	{name: "p2_barracks_12_count", typ: "int"},
	{name: "p3_militia_74_count", typ: "int"},
	{name: "p3_barracks_12_count", typ: "int"},
	{name: "gaia_militia_74_count", typ: "int"},
	{name: "gaia_barracks_12_count", typ: "int"},
	{name: "p1_archer_4_count", typ: "int"},
}

var rtv12FooterFields = []schemaField{
	{name: "tag", typ: "string"},
	{name: "rows_written", typ: "int"},
	{name: "append_tag", typ: "string"},
	{name: "append_probe", typ: "int"},
}

var rtv13HeaderFields = rtv12HeaderFields

var rtv13RowFields = buildRTV13RowFields()

var rtv13FooterFields = rtv12FooterFields

var rtv14HeaderFields = rtv12HeaderFields

var rtv14RowFields = buildRTV14RowFields()

var rtv14FooterFields = rtv12FooterFields

var rtv141HeaderFields = rtv12HeaderFields

var rtv141RowFields = buildRTV141RowFields()

var rtv141FooterFields = rtv12FooterFields

var rtv15HeaderFields = rtv12HeaderFields

var rtv15RowFields = buildRTV15RowFields()

var rtv15FooterFields = rtv12FooterFields

var rtv16HeaderFields = rtv12HeaderFields

var rtv16RowFields = buildRTV16RowFields()

var rtv16FooterFields = rtv12FooterFields

var rtv17HeaderFields = rtv12HeaderFields

var rtv17RowFields = buildRTV17RowFields()

var rtv17FooterFields = rtv12FooterFields

var a2ksem2HeaderFields = []schemaField{
	{name: "magic", typ: "string"},
	{name: "version", typ: "int"},
	{name: "fixture", typ: "string"},
}

var a2ksem2RowFields = []schemaField{
	{name: "tag", typ: "string"},
	{name: "phase_id", typ: "int"},
	{name: "lane_id", typ: "int"},
	{name: "xs_time", typ: "int"},
	{name: "p1_food_attr0", typ: "float"},
	{name: "p1_wood_attr1", typ: "float"},
	{name: "p1_stone_attr2", typ: "float"},
	{name: "p1_gold_attr3", typ: "float"},
	{name: "p1_population_attr11", typ: "float"},
	{name: "p1_research_count_attr21", typ: "float"},
	{name: "p1_militia_74_count", typ: "int"},
	{name: "p1_man_at_arms_75_count", typ: "int"},
	{name: "p1_villager_83_count", typ: "int"},
	{name: "p1_scout_448_count", typ: "int"},
	{name: "p1_town_center_109_count", typ: "int"},
	{name: "p1_barracks_12_count", typ: "int"},
	{name: "p1_stable_101_count", typ: "int"},
}

var a2ksem2FooterFields = []schemaField{
	{name: "tag", typ: "string"},
	{name: "rows_written", typ: "int"},
}

type schemaField struct {
	name string
	typ  string
}

func decodeRTV12Executioner(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv12HeaderFields, rtv12RowFields, rtv12FooterFields)
}

func decodeRTV13DirectedAttribution(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv13HeaderFields, rtv13RowFields, rtv13FooterFields)
}

func decodeRTV14CastleKillCalibration(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv14HeaderFields, rtv14RowFields, rtv14FooterFields)
}

func decodeRTV141CastleKillCalibration(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv141HeaderFields, rtv141RowFields, rtv141FooterFields)
}

func decodeRTV15CastleKillGroundTruth(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv15HeaderFields, rtv15RowFields, rtv15FooterFields)
}

func decodeRTV16SemanticPromotion(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv16HeaderFields, rtv16RowFields, rtv16FooterFields)
}

func decodeRTV17PackedTest(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, rtv17HeaderFields, rtv17RowFields, rtv17FooterFields)
}

func decodeA2KSEM2DATCommandSemantics(data []byte, report *DataSchemaDecodeReport) {
	decodePhaseRows(data, report, a2ksem2HeaderFields, a2ksem2RowFields, a2ksem2FooterFields)
}

func decodePhaseRows(data []byte, report *DataSchemaDecodeReport, headerFields, rowFields, footerFields []schemaField) {
	offset := 0
	header, next, err := readSchemaFields(data, offset, 0, headerFields)
	if err != nil {
		report.Errors = append(report.Errors, err.Error())
		return
	}
	report.Header = header
	offset = next
	indexBase := len(headerFields)

	for offset < len(data) {
		tagValue, _, err := readTypedValue(data, offset, indexBase, "string", false)
		if err != nil {
			report.Errors = append(report.Errors, err.Error())
			report.Summary.RemainingFrom = offset
			return
		}
		switch tagValue.String {
		case "phase":
			rowOffset := offset
			fields, next, err := readSchemaFields(data, offset, indexBase, rowFields)
			if err != nil {
				report.Errors = append(report.Errors, err.Error())
				report.Summary.RemainingFrom = offset
				return
			}
			row := DataSchemaRow{
				Index:  len(report.Rows),
				Offset: rowOffset,
				Fields: fields,
			}
			row.PhaseID = intFromAny(fields["phase_id"])
			row.TimeS = intFromAny(fields["xs_time"])
			report.Rows = append(report.Rows, row)
			offset = next
			indexBase += len(rowFields)
		case "end":
			footer, next, err := readSchemaFields(data, offset, indexBase, footerFields)
			if err != nil {
				report.Errors = append(report.Errors, err.Error())
				report.Summary.RemainingFrom = offset
				return
			}
			report.Footer = footer
			report.Summary.HasFooter = true
			offset = next
			indexBase += len(footerFields)
			if offset < len(data) {
				report.Errors = append(report.Errors, fmt.Sprintf("trailing bytes after footer at offset %d", offset))
				report.Summary.RemainingFrom = offset
			}
			return
		default:
			report.Errors = append(report.Errors, fmt.Sprintf("unexpected row tag %q at offset %d", tagValue.String, offset))
			report.Summary.RemainingFrom = offset
			return
		}
	}

	report.Warnings = append(report.Warnings, "sidecar has no end/footer row; decoded complete phase rows only")
}

func buildRTV13RowFields() []schemaField {
	fields := []schemaField{
		{name: "tag", typ: "string"},
		{name: "phase_id", typ: "int"},
		{name: "case_player", typ: "int"},
		{name: "case_kind", typ: "int"},
		{name: "xs_time", typ: "int"},
		{name: "p1_resource_sum_carrier_food_attr0", typ: "float"},
		{name: "p1_wood_case_player_attr1", typ: "float"},
		{name: "p1_gold_case_kind_attr3", typ: "float"},
		{name: "p1_stone_row_index_attr2", typ: "float"},
	}
	for player := 1; player <= 8; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_kills_attr20", typ: "float"},
			schemaField{name: prefix + "_razings_attr43", typ: "float"},
			schemaField{name: prefix + "_killed_by_others_attr154", typ: "float"},
			schemaField{name: prefix + "_razed_by_others_attr155", typ: "float"},
			schemaField{name: prefix + "_kill_value_attr170", typ: "float"},
			schemaField{name: prefix + "_raze_value_attr172", typ: "float"},
		)
	}
	fields = append(fields,
		schemaField{name: "p1_gaia_kills_attr300", typ: "float"},
		schemaField{name: "p1_player1_kills_attr301", typ: "float"},
		schemaField{name: "p1_player2_kills_attr302", typ: "float"},
		schemaField{name: "p1_player3_kills_attr303", typ: "float"},
		schemaField{name: "p1_player4_kills_attr304", typ: "float"},
		schemaField{name: "p1_player5_kills_attr305", typ: "float"},
		schemaField{name: "p1_player6_kills_attr306", typ: "float"},
		schemaField{name: "p1_player7_kills_attr307", typ: "float"},
		schemaField{name: "p1_player8_kills_attr308", typ: "float"},

		schemaField{name: "p1_gaia_razings_attr350", typ: "float"},
		schemaField{name: "p1_player1_razings_attr351", typ: "float"},
		schemaField{name: "p1_player2_razings_attr352", typ: "float"},
		schemaField{name: "p1_player3_razings_attr353", typ: "float"},
		schemaField{name: "p1_player4_razings_attr354", typ: "float"},
		schemaField{name: "p1_player5_razings_attr355", typ: "float"},
		schemaField{name: "p1_player6_razings_attr356", typ: "float"},
		schemaField{name: "p1_player7_razings_attr357", typ: "float"},
		schemaField{name: "p1_player8_razings_attr358", typ: "float"},
	)
	for player := 2; player <= 8; player++ {
		fields = append(fields, schemaField{name: fmt.Sprintf("p%d_kills_by_player1_attr326", player), typ: "float"})
	}
	for player := 2; player <= 8; player++ {
		fields = append(fields, schemaField{name: fmt.Sprintf("p%d_razings_by_player1_attr376", player), typ: "float"})
	}
	fields = append(fields,
		schemaField{name: "p1_militia_74_count", typ: "int"},
		schemaField{name: "p1_barracks_12_count", typ: "int"},
	)
	for player := 2; player <= 8; player++ {
		fields = append(fields,
			schemaField{name: fmt.Sprintf("p%d_militia_74_count", player), typ: "int"},
			schemaField{name: fmt.Sprintf("p%d_barracks_12_count", player), typ: "int"},
		)
	}
	return fields
}

func buildRTV14RowFields() []schemaField {
	fields := []schemaField{
		{name: "tag", typ: "string"},
		{name: "phase_id", typ: "int"},
		{name: "case_player", typ: "int"},
		{name: "case_unit", typ: "int"},
		{name: "spawn_count", typ: "int"},
		{name: "expected_p1_kills", typ: "int"},
		{name: "xs_time", typ: "int"},
		{name: "p1_resource_sum_carrier_food_attr0", typ: "float"},
		{name: "p1_wood_case_player_attr1", typ: "float"},
		{name: "p1_gold_case_unit_attr3", typ: "float"},
		{name: "p1_stone_expected_p1_kills_attr2", typ: "float"},
	}
	for player := 1; player <= 8; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_kills_attr20", typ: "float"},
			schemaField{name: prefix + "_razings_attr43", typ: "float"},
			schemaField{name: prefix + "_killed_by_others_attr154", typ: "float"},
			schemaField{name: prefix + "_razed_by_others_attr155", typ: "float"},
			schemaField{name: prefix + "_kill_value_attr170", typ: "float"},
			schemaField{name: prefix + "_raze_value_attr172", typ: "float"},
		)
	}
	fields = append(fields,
		schemaField{name: "p1_gaia_kills_attr300", typ: "float"},
		schemaField{name: "p1_player1_kills_attr301", typ: "float"},
		schemaField{name: "p1_player2_kills_attr302", typ: "float"},
		schemaField{name: "p1_player3_kills_attr303", typ: "float"},
		schemaField{name: "p1_player4_kills_attr304", typ: "float"},
		schemaField{name: "p1_player5_kills_attr305", typ: "float"},
		schemaField{name: "p1_player6_kills_attr306", typ: "float"},
		schemaField{name: "p1_player7_kills_attr307", typ: "float"},
		schemaField{name: "p1_player8_kills_attr308", typ: "float"},

		schemaField{name: "p1_gaia_razings_attr350", typ: "float"},
		schemaField{name: "p1_player1_razings_attr351", typ: "float"},
		schemaField{name: "p1_player2_razings_attr352", typ: "float"},
		schemaField{name: "p1_player3_razings_attr353", typ: "float"},
		schemaField{name: "p1_player4_razings_attr354", typ: "float"},
		schemaField{name: "p1_player5_razings_attr355", typ: "float"},
		schemaField{name: "p1_player6_razings_attr356", typ: "float"},
		schemaField{name: "p1_player7_razings_attr357", typ: "float"},
		schemaField{name: "p1_player8_razings_attr358", typ: "float"},
	)
	for player := 2; player <= 8; player++ {
		fields = append(fields, schemaField{name: fmt.Sprintf("p%d_kills_by_player1_attr326", player), typ: "float"})
	}
	for player := 2; player <= 8; player++ {
		fields = append(fields, schemaField{name: fmt.Sprintf("p%d_razings_by_player1_attr376", player), typ: "float"})
	}
	fields = append(fields, schemaField{name: "p1_castle_82_count", typ: "int"})
	for player := 2; player <= 8; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_militia_74_count", typ: "int"},
			schemaField{name: prefix + "_archer_4_count", typ: "int"},
			schemaField{name: prefix + "_spearman_93_count", typ: "int"},
			schemaField{name: prefix + "_villager_83_count", typ: "int"},
			schemaField{name: prefix + "_scout_448_count", typ: "int"},
			schemaField{name: prefix + "_knight_38_count", typ: "int"},
		)
	}
	return fields
}

func buildRTV141RowFields() []schemaField {
	fields := []schemaField{
		{name: "tag", typ: "string"},
		{name: "phase_id", typ: "int"},
		{name: "case_player", typ: "int"},
		{name: "case_unit", typ: "int"},
		{name: "spawn_count", typ: "int"},
		{name: "expected_p1_kills", typ: "int"},
		{name: "xs_time", typ: "int"},
		{name: "p1_resource_sum_carrier_food_attr0", typ: "float"},
		{name: "p1_wood_case_player_attr1", typ: "float"},
		{name: "p1_gold_case_unit_attr3", typ: "float"},
		{name: "p1_stone_expected_p1_kills_attr2", typ: "float"},
	}
	for player := 1; player <= 3; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_kills_attr20", typ: "float"},
			schemaField{name: prefix + "_razings_attr43", typ: "float"},
			schemaField{name: prefix + "_killed_by_others_attr154", typ: "float"},
			schemaField{name: prefix + "_razed_by_others_attr155", typ: "float"},
			schemaField{name: prefix + "_kill_value_attr170", typ: "float"},
			schemaField{name: prefix + "_raze_value_attr172", typ: "float"},
		)
	}
	fields = append(fields,
		schemaField{name: "p1_gaia_kills_attr300", typ: "float"},
		schemaField{name: "p1_player1_kills_attr301", typ: "float"},
		schemaField{name: "p1_player2_kills_attr302", typ: "float"},
		schemaField{name: "p1_player3_kills_attr303", typ: "float"},
		schemaField{name: "p1_gaia_razings_attr350", typ: "float"},
		schemaField{name: "p1_player1_razings_attr351", typ: "float"},
		schemaField{name: "p1_player2_razings_attr352", typ: "float"},
		schemaField{name: "p1_player3_razings_attr353", typ: "float"},
		schemaField{name: "p2_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p3_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p2_razings_by_player1_attr376", typ: "float"},
		schemaField{name: "p3_razings_by_player1_attr376", typ: "float"},
		schemaField{name: "p1_castle_82_count", typ: "int"},
	)
	for player := 2; player <= 3; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_militia_74_count", typ: "int"},
			schemaField{name: prefix + "_archer_4_count", typ: "int"},
			schemaField{name: prefix + "_spearman_93_count", typ: "int"},
			schemaField{name: prefix + "_villager_83_count", typ: "int"},
			schemaField{name: prefix + "_scout_448_count", typ: "int"},
			schemaField{name: prefix + "_knight_38_count", typ: "int"},
		)
	}
	return fields
}

func buildRTV15RowFields() []schemaField {
	fields := []schemaField{
		{name: "tag", typ: "string"},
		{name: "phase_id", typ: "int"},
		{name: "case_player", typ: "int"},
		{name: "case_unit", typ: "int"},
		{name: "spawn_count", typ: "int"},
		{name: "expected_p1_kills", typ: "int"},
		{name: "xs_time", typ: "int"},
		{name: "p1_resource_sum_carrier_food_attr0", typ: "float"},
		{name: "p1_wood_case_player_attr1", typ: "float"},
		{name: "p1_gold_case_unit_attr3", typ: "float"},
		{name: "p1_stone_expected_p1_kills_attr2", typ: "float"},
	}
	for player := 1; player <= 4; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_kills_attr20", typ: "float"},
			schemaField{name: prefix + "_razings_attr43", typ: "float"},
			schemaField{name: prefix + "_killed_by_others_attr154", typ: "float"},
			schemaField{name: prefix + "_razed_by_others_attr155", typ: "float"},
			schemaField{name: prefix + "_kill_value_attr170", typ: "float"},
			schemaField{name: prefix + "_raze_value_attr172", typ: "float"},
		)
	}
	fields = append(fields,
		schemaField{name: "p1_gaia_kills_attr300", typ: "float"},
		schemaField{name: "p1_player1_kills_attr301", typ: "float"},
		schemaField{name: "p1_player2_kills_attr302", typ: "float"},
		schemaField{name: "p1_player3_kills_attr303", typ: "float"},
		schemaField{name: "p1_player4_kills_attr304", typ: "float"},
		schemaField{name: "p2_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p3_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p4_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p1_castle_82_count", typ: "int"},
	)
	for player := 2; player <= 4; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_barracks_12_count", typ: "int"},
			schemaField{name: prefix + "_militia_74_count", typ: "int"},
			schemaField{name: prefix + "_archer_4_count", typ: "int"},
			schemaField{name: prefix + "_spearman_93_count", typ: "int"},
			schemaField{name: prefix + "_villager_83_count", typ: "int"},
			schemaField{name: prefix + "_scout_448_count", typ: "int"},
			schemaField{name: prefix + "_knight_38_count", typ: "int"},
		)
	}
	return fields
}

func buildRTV16RowFields() []schemaField {
	fields := []schemaField{
		{name: "tag", typ: "string"},
		{name: "phase_id", typ: "int"},
		{name: "case_player", typ: "int"},
		{name: "case_kind", typ: "int"},
		{name: "expected_p1_kills", typ: "int"},
		{name: "expected_p1_razes", typ: "int"},
		{name: "xs_time", typ: "int"},
		{name: "p1_resource_sum_carrier_food_attr0", typ: "float"},
		{name: "p1_wood_case_player_attr1", typ: "float"},
		{name: "p1_gold_case_kind_attr3", typ: "float"},
		{name: "p1_stone_expected_p1_kills_attr2", typ: "float"},
	}
	for player := 1; player <= 4; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_current_age_attr6", typ: "float"},
			schemaField{name: prefix + "_population_attr11", typ: "float"},
			schemaField{name: prefix + "_discovery_attr13", typ: "float"},
			schemaField{name: prefix + "_exploration_attr22", typ: "float"},
			schemaField{name: prefix + "_kills_attr20", typ: "float"},
			schemaField{name: prefix + "_razings_attr43", typ: "float"},
			schemaField{name: prefix + "_value_killed_by_others_attr152", typ: "float"},
			schemaField{name: prefix + "_value_razed_by_others_attr153", typ: "float"},
			schemaField{name: prefix + "_killed_by_others_attr154", typ: "float"},
			schemaField{name: prefix + "_razed_by_others_attr155", typ: "float"},
			schemaField{name: prefix + "_value_current_units_attr164", typ: "float"},
			schemaField{name: prefix + "_value_current_buildings_attr165", typ: "float"},
			schemaField{name: prefix + "_food_total_attr166", typ: "float"},
			schemaField{name: prefix + "_wood_total_attr167", typ: "float"},
			schemaField{name: prefix + "_stone_total_attr168", typ: "float"},
			schemaField{name: prefix + "_gold_total_attr169", typ: "float"},
			schemaField{name: prefix + "_kill_value_attr170", typ: "float"},
			schemaField{name: prefix + "_raze_value_attr172", typ: "float"},
			schemaField{name: prefix + "_tribute_score_attr175", typ: "float"},
			schemaField{name: prefix + "_food_score_attr185", typ: "float"},
			schemaField{name: prefix + "_wood_score_attr186", typ: "float"},
			schemaField{name: prefix + "_stone_score_attr187", typ: "float"},
			schemaField{name: prefix + "_gold_score_attr188", typ: "float"},
			schemaField{name: prefix + "_map_reveal_attr203", typ: "float"},
			schemaField{name: prefix + "_unit_reveal_attr204", typ: "float"},
			schemaField{name: prefix + "_temporary_map_reveal_attr209", typ: "float"},
		)
	}
	fields = append(fields,
		schemaField{name: "p1_gaia_kills_attr300", typ: "float"},
		schemaField{name: "p1_player1_kills_attr301", typ: "float"},
		schemaField{name: "p1_player2_kills_attr302", typ: "float"},
		schemaField{name: "p1_player3_kills_attr303", typ: "float"},
		schemaField{name: "p1_player4_kills_attr304", typ: "float"},
		schemaField{name: "p1_gaia_razings_attr350", typ: "float"},
		schemaField{name: "p1_player1_razings_attr351", typ: "float"},
		schemaField{name: "p1_player2_razings_attr352", typ: "float"},
		schemaField{name: "p1_player3_razings_attr353", typ: "float"},
		schemaField{name: "p1_player4_razings_attr354", typ: "float"},
		schemaField{name: "p1_gaia_kill_value_attr400", typ: "float"},
		schemaField{name: "p1_player1_kill_value_attr401", typ: "float"},
		schemaField{name: "p1_player2_kill_value_attr402", typ: "float"},
		schemaField{name: "p1_player3_kill_value_attr403", typ: "float"},
		schemaField{name: "p1_player4_kill_value_attr404", typ: "float"},
		schemaField{name: "p1_gaia_raze_value_attr425", typ: "float"},
		schemaField{name: "p1_player1_raze_value_attr426", typ: "float"},
		schemaField{name: "p1_player2_raze_value_attr427", typ: "float"},
		schemaField{name: "p1_player3_raze_value_attr428", typ: "float"},
		schemaField{name: "p1_player4_raze_value_attr429", typ: "float"},
		schemaField{name: "p2_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p3_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p4_kills_by_player1_attr326", typ: "float"},
		schemaField{name: "p2_razings_by_player1_attr376", typ: "float"},
		schemaField{name: "p3_razings_by_player1_attr376", typ: "float"},
		schemaField{name: "p4_razings_by_player1_attr376", typ: "float"},
		schemaField{name: "p1_castle_82_count", typ: "int"},
		schemaField{name: "p1_cobra_748_count", typ: "int"},
	)
	for player := 2; player <= 4; player++ {
		prefix := fmt.Sprintf("p%d", player)
		fields = append(fields,
			schemaField{name: prefix + "_barracks_12_count", typ: "int"},
			schemaField{name: prefix + "_militia_74_count", typ: "int"},
			schemaField{name: prefix + "_archer_4_count", typ: "int"},
			schemaField{name: prefix + "_spearman_93_count", typ: "int"},
		)
	}
	fields = append(fields, schemaField{name: "gaia_militia_74_count", typ: "int"})
	return fields
}

func buildRTV17RowFields() []schemaField {
	fields := append([]schemaField{}, rtv16RowFields...)
	fields = append(fields,
		schemaField{name: "expected_p1_deaths", typ: "int"},
		schemaField{name: "expected_p2_kills", typ: "int"},
		schemaField{name: "p2_player1_kills_attr301", typ: "float"},
		schemaField{name: "p2_player1_razings_attr351", typ: "float"},
		schemaField{name: "p2_player1_kill_value_attr401", typ: "float"},
		schemaField{name: "p2_player1_raze_value_attr426", typ: "float"},
		schemaField{name: "p1_kills_by_player2_attr327", typ: "float"},
		schemaField{name: "p1_razings_by_player2_attr377", typ: "float"},
		schemaField{name: "p1_militia_74_count", typ: "int"},
		schemaField{name: "p1_archer_4_count", typ: "int"},
		schemaField{name: "p1_spearman_93_count", typ: "int"},
		schemaField{name: "p1_villager_83_count", typ: "int"},
		schemaField{name: "p1_scout_448_count", typ: "int"},
		schemaField{name: "p1_knight_38_count", typ: "int"},
		schemaField{name: "p2_castle_82_count", typ: "int"},
	)
	return fields
}

func readSchemaFields(data []byte, offset int, indexBase int, fields []schemaField) (map[string]any, int, error) {
	out := make(map[string]any, len(fields))
	for i, field := range fields {
		value, next, err := readTypedValue(data, offset, indexBase+i, field.typ, false)
		if err != nil {
			return nil, offset, err
		}
		out[field.name] = scalarDataValue(value)
		offset = next
	}
	return out, offset, nil
}

func scalarDataValue(value DataValue) any {
	switch value.Type {
	case "string":
		return value.String
	case "int":
		if value.Int != nil {
			return int(*value.Int)
		}
	case "uint":
		if value.UInt != nil {
			return *value.UInt
		}
	case "float":
		if value.Float != nil {
			return *value.Float
		}
	}
	return nil
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case uint32:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}
