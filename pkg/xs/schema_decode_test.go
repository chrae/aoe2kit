package xs

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestDecodeRTV12ExecutionerPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV12_EXECUTIONER_LEDGER")
	appendInt(12)
	appendString("manual_combat_attr_matrix_v12")
	appendString("phase")
	appendInt(20)
	appendInt(20)
	appendInt(1)
	appendInt(1)
	appendInt(20)
	for i := 0; i < 38; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 8; i++ {
		appendInt(i)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "rtv12"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if report.Summary.HasFooter {
		t.Fatalf("partial sidecar unexpectedly has footer")
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	if got := report.Rows[0].PhaseID; got != 20 {
		t.Fatalf("phase=%d, want 20", got)
	}
	if got := report.Rows[0].Fields["p1_kills_attr20"]; got != float32(4) {
		t.Fatalf("p1_kills_attr20=%v, want 4", got)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected no-footer warning")
	}
}

func TestDecodeRTV13DirectedAttributionPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV13_DIRECTED_ATTRIBUTION")
	appendInt(13)
	appendString("directed_kill_raze_matrix_v13")
	appendString("phase")
	appendInt(101)
	appendInt(4)
	appendInt(1)
	appendInt(205)
	for i := 0; i < 4; i++ {
		appendFloat(float32(884101 + i))
	}
	for i := 0; i < 48; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 18; i++ {
		appendFloat(float32(100 + i))
	}
	for i := 0; i < 14; i++ {
		appendFloat(float32(200 + i))
	}
	for i := 0; i < 16; i++ {
		appendInt(300 + i)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "rtv13"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got := row.PhaseID; got != 101 {
		t.Fatalf("phase=%d, want 101", got)
	}
	if got := row.TimeS; got != 205 {
		t.Fatalf("time=%d, want 205", got)
	}
	if got := row.Fields["case_player"]; got != 4 {
		t.Fatalf("case_player=%v, want 4", got)
	}
	if got := row.Fields["p1_kills_attr20"]; got != float32(0) {
		t.Fatalf("p1_kills_attr20=%v, want 0", got)
	}
	if got := row.Fields["p1_player8_razings_attr358"]; got != float32(117) {
		t.Fatalf("p1_player8_razings_attr358=%v, want 117", got)
	}
	if got := row.Fields["p8_razings_by_player1_attr376"]; got != float32(213) {
		t.Fatalf("p8_razings_by_player1_attr376=%v, want 213", got)
	}
	if got := row.Fields["p8_barracks_12_count"]; got != 315 {
		t.Fatalf("p8_barracks_12_count=%v, want 315", got)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected no-footer warning")
	}
}

func TestDecodeRTV14CastleKillCalibrationPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV14_CASTLE_KILL_CALIBRATION")
	appendInt(14)
	appendString("castle_combat_death_matrix_v14")
	appendString("phase")
	appendInt(73)
	appendInt(4)
	appendInt(93)
	appendInt(2)
	appendInt(7)
	appendInt(85)
	for i := 0; i < 4; i++ {
		appendFloat(float32(885073 + i))
	}
	for i := 0; i < 48; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 18; i++ {
		appendFloat(float32(100 + i))
	}
	for i := 0; i < 14; i++ {
		appendFloat(float32(200 + i))
	}
	appendInt(1)
	for i := 0; i < 42; i++ {
		appendInt(300 + i)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "rtv14"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got := row.PhaseID; got != 73 {
		t.Fatalf("phase=%d, want 73", got)
	}
	if got := row.TimeS; got != 85 {
		t.Fatalf("time=%d, want 85", got)
	}
	if got := row.Fields["case_unit"]; got != 93 {
		t.Fatalf("case_unit=%v, want 93", got)
	}
	if got := row.Fields["expected_p1_kills"]; got != 7 {
		t.Fatalf("expected_p1_kills=%v, want 7", got)
	}
	if got := row.Fields["p1_player8_razings_attr358"]; got != float32(117) {
		t.Fatalf("p1_player8_razings_attr358=%v, want 117", got)
	}
	if got := row.Fields["p8_razings_by_player1_attr376"]; got != float32(213) {
		t.Fatalf("p8_razings_by_player1_attr376=%v, want 213", got)
	}
	if got := row.Fields["p1_castle_82_count"]; got != 1 {
		t.Fatalf("p1_castle_82_count=%v, want 1", got)
	}
	if got := row.Fields["p8_knight_38_count"]; got != 341 {
		t.Fatalf("p8_knight_38_count=%v, want 341", got)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected no-footer warning")
	}
}

func TestDecodeRTV141CastleKillCalibrationPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV141_CASTLE_KILL_CALIBRATION")
	appendInt(141)
	appendString("castle_combat_death_matrix_v141")
	appendString("phase")
	appendInt(93)
	appendInt(2)
	appendInt(83)
	appendInt(1)
	appendInt(8)
	appendInt(105)
	for i := 0; i < 4; i++ {
		appendFloat(float32(885193 + i))
	}
	for i := 0; i < 18; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 12; i++ {
		appendFloat(float32(100 + i))
	}
	appendInt(1)
	for i := 0; i < 12; i++ {
		appendInt(300 + i)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "rtv142"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if report.Schema != "rtv141-castle-kill-calibration" {
		t.Fatalf("schema=%q, want rtv141-castle-kill-calibration", report.Schema)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got := row.PhaseID; got != 93 {
		t.Fatalf("phase=%d, want 93", got)
	}
	if got := row.Fields["case_unit"]; got != 83 {
		t.Fatalf("case_unit=%v, want 83", got)
	}
	if got := row.Fields["p1_player3_razings_attr353"]; got != float32(107) {
		t.Fatalf("p1_player3_razings_attr353=%v, want 107", got)
	}
	if got := row.Fields["p3_razings_by_player1_attr376"]; got != float32(111) {
		t.Fatalf("p3_razings_by_player1_attr376=%v, want 111", got)
	}
	if got := row.Fields["p1_castle_82_count"]; got != 1 {
		t.Fatalf("p1_castle_82_count=%v, want 1", got)
	}
	if got := row.Fields["p3_knight_38_count"]; got != 311 {
		t.Fatalf("p3_knight_38_count=%v, want 311", got)
	}
}

func TestDecodeRTV15CastleKillGroundTruthPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV15_CASTLE_KILL_GROUND_TRUTH")
	appendInt(15)
	appendString("autonomous_castle_kill_matrix_v15")
	appendString("phase")
	appendInt(133)
	appendInt(234)
	appendInt(74)
	appendInt(3)
	appendInt(12)
	appendInt(150)
	for i := 0; i < 4; i++ {
		appendFloat(float32(886133 + i))
	}
	for i := 0; i < 24; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 5; i++ {
		appendFloat(float32(100 + i))
	}
	for i := 0; i < 3; i++ {
		appendFloat(float32(200 + i))
	}
	appendInt(1)
	for i := 0; i < 21; i++ {
		appendInt(300 + i)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "v15"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if report.Schema != "rtv15-castle-kill-ground-truth" {
		t.Fatalf("schema=%q, want rtv15-castle-kill-ground-truth", report.Schema)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got := row.PhaseID; got != 133 {
		t.Fatalf("phase=%d, want 133", got)
	}
	if got := row.TimeS; got != 150 {
		t.Fatalf("time=%d, want 150", got)
	}
	if got := row.Fields["case_player"]; got != 234 {
		t.Fatalf("case_player=%v, want 234", got)
	}
	if got := row.Fields["case_unit"]; got != 74 {
		t.Fatalf("case_unit=%v, want 74", got)
	}
	if got := row.Fields["spawn_count"]; got != 3 {
		t.Fatalf("spawn_count=%v, want 3", got)
	}
	if got := row.Fields["expected_p1_kills"]; got != 12 {
		t.Fatalf("expected_p1_kills=%v, want 12", got)
	}
	if got := row.Fields["p4_raze_value_attr172"]; got != float32(23) {
		t.Fatalf("p4_raze_value_attr172=%v, want 23", got)
	}
	if got := row.Fields["p1_player4_kills_attr304"]; got != float32(104) {
		t.Fatalf("p1_player4_kills_attr304=%v, want 104", got)
	}
	if got := row.Fields["p4_kills_by_player1_attr326"]; got != float32(202) {
		t.Fatalf("p4_kills_by_player1_attr326=%v, want 202", got)
	}
	if got := row.Fields["p1_castle_82_count"]; got != 1 {
		t.Fatalf("p1_castle_82_count=%v, want 1", got)
	}
	if got := row.Fields["p4_knight_38_count"]; got != 320 {
		t.Fatalf("p4_knight_38_count=%v, want 320", got)
	}
}

func TestDecodeRTV16SemanticPromotionPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV16_SEMANTIC_PROMOTION")
	appendInt(16)
	appendString("semantic_promotion_kill_raze_score_fog_v16")
	appendString("phase")
	appendInt(900)
	appendInt(9)
	appendInt(9)
	appendInt(4)
	appendInt(1)
	appendInt(245)
	for i := 0; i < 4; i++ {
		appendFloat(float32(887900 + i))
	}
	for i := 0; i < 104; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 20; i++ {
		appendFloat(float32(200 + i))
	}
	for i := 0; i < 6; i++ {
		appendFloat(float32(300 + i))
	}
	appendInt(1)
	appendInt(1)
	for _, v := range []int{10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33, 40} {
		appendInt(v)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "v16"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if report.Schema != "rtv16-semantic-promotion" {
		t.Fatalf("schema=%q, want rtv16-semantic-promotion", report.Schema)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got := row.PhaseID; got != 900 {
		t.Fatalf("phase=%d, want 900", got)
	}
	if got := row.TimeS; got != 245 {
		t.Fatalf("time=%d, want 245", got)
	}
	if got := row.Fields["expected_p1_kills"]; got != 4 {
		t.Fatalf("expected_p1_kills=%v, want 4", got)
	}
	if got := row.Fields["expected_p1_razes"]; got != 1 {
		t.Fatalf("expected_p1_razes=%v, want 1", got)
	}
	if got := row.Fields["p1_kills_attr20"]; got != float32(4) {
		t.Fatalf("p1_kills_attr20=%v, want 4", got)
	}
	if got := row.Fields["p4_gold_score_attr188"]; got != float32(100) {
		t.Fatalf("p4_gold_score_attr188=%v, want 100", got)
	}
	if got := row.Fields["p1_player4_razings_attr354"]; got != float32(209) {
		t.Fatalf("p1_player4_razings_attr354=%v, want 209", got)
	}
	if got := row.Fields["p1_player3_kill_value_attr403"]; got != float32(213) {
		t.Fatalf("p1_player3_kill_value_attr403=%v, want 213", got)
	}
	if got := row.Fields["p4_razings_by_player1_attr376"]; got != float32(305) {
		t.Fatalf("p4_razings_by_player1_attr376=%v, want 305", got)
	}
	if got := row.Fields["p1_castle_82_count"]; got != 1 {
		t.Fatalf("p1_castle_82_count=%v, want 1", got)
	}
	if got := row.Fields["p4_spearman_93_count"]; got != 33 {
		t.Fatalf("p4_spearman_93_count=%v, want 33", got)
	}
	if got := row.Fields["gaia_militia_74_count"]; got != 40 {
		t.Fatalf("gaia_militia_74_count=%v, want 40", got)
	}
}

func TestDecodeRTV17PackedTestPartialSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_RTV17_PACKED_TEST")
	appendInt(17)
	appendString("packed_death_kill_raze_resource_visibility_v17")
	appendString("phase")
	appendInt(900)
	appendInt(9)
	appendInt(900)
	appendInt(1)
	appendInt(1)
	appendInt(245)
	for i := 0; i < 4; i++ {
		appendFloat(float32(888900 + i))
	}
	for i := 0; i < 104; i++ {
		appendFloat(float32(i))
	}
	for i := 0; i < 20; i++ {
		appendFloat(float32(200 + i))
	}
	for i := 0; i < 6; i++ {
		appendFloat(float32(300 + i))
	}
	appendInt(1)
	appendInt(1)
	for _, v := range []int{10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33, 40} {
		appendInt(v)
	}
	appendInt(4)
	appendInt(4)
	for i := 0; i < 6; i++ {
		appendFloat(float32(500 + i))
	}
	for _, v := range []int{70, 71, 72, 73, 74, 75, 82} {
		appendInt(v)
	}

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "v17"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if report.Schema != "rtv17-packed-test" {
		t.Fatalf("schema=%q, want rtv17-packed-test", report.Schema)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if got := row.Fields["expected_p1_deaths"]; got != 4 {
		t.Fatalf("expected_p1_deaths=%v, want 4", got)
	}
	if got := row.Fields["p2_player1_kills_attr301"]; got != float32(500) {
		t.Fatalf("p2_player1_kills_attr301=%v, want 500", got)
	}
	if got := row.Fields["p1_kills_by_player2_attr327"]; got != float32(504) {
		t.Fatalf("p1_kills_by_player2_attr327=%v, want 504", got)
	}
	if got := row.Fields["p1_knight_38_count"]; got != 75 {
		t.Fatalf("p1_knight_38_count=%v, want 75", got)
	}
	if got := row.Fields["p2_castle_82_count"]; got != 82 {
		t.Fatalf("p2_castle_82_count=%v, want 82", got)
	}
}

func TestDecodeA2KSEM2DATCommandSemanticsSidecar(t *testing.T) {
	data := []byte{}
	appendString := func(s string) {
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, []byte(s)...)
	}
	appendInt := func(v int) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(v))
		data = append(data, buf[:]...)
	}
	appendFloat := func(v float32) {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
		data = append(data, buf[:]...)
	}

	appendString("A2K_DAT_COMMAND_SEMANTICS")
	appendInt(2)
	appendString("run1_resource_object_count_contract")
	appendString("phase")
	appendInt(15)
	appendInt(1)
	appendInt(15)
	for _, v := range []float32{1111, 0, 0, 222, 2, 3} {
		appendFloat(v)
	}
	for _, v := range []int{1, 0, 1, 2, 1, 1, 1} {
		appendInt(v)
	}
	appendString("end")
	appendInt(1)

	report := DecodeDataBytesSchema(data, DataSchemaDecodeOptions{Schema: "datsem2"})
	if !report.OK {
		t.Fatalf("DecodeDataBytesSchema OK=false errors=%v", report.Errors)
	}
	if report.Schema != "a2ksem2-dat-command-semantics" {
		t.Fatalf("schema=%q, want a2ksem2-dat-command-semantics", report.Schema)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(report.Rows))
	}
	row := report.Rows[0]
	if row.PhaseID != 15 || row.TimeS != 15 {
		t.Fatalf("row phase/time = %d/%d, want 15/15", row.PhaseID, row.TimeS)
	}
	if got := row.Fields["p1_food_attr0"]; got != float32(1111) {
		t.Fatalf("p1_food_attr0=%v, want 1111", got)
	}
	if got := row.Fields["p1_scout_448_count"]; got != 2 {
		t.Fatalf("p1_scout_448_count=%v, want 2", got)
	}
	if !report.Summary.HasFooter {
		t.Fatalf("missing footer")
	}
	if got := report.Footer["rows_written"]; got != 1 {
		t.Fatalf("rows_written=%v, want 1", got)
	}
}
