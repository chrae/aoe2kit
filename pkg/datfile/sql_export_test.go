package datfile

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestAttributeMapUsesCanonicalDecodedFields(t *testing.T) {
	tests := map[int]string{
		0:   "hit_points",
		66:  "blood_unit_id",
		71:  "standing_graphic_1",
		145: "trailing_unit",
		185: "fly_mode",
		186: "can_be_gathered",
		187: "hill_mode",
	}
	if field, confidence := UGCAttributeDATField(125); field != "" || confidence != "unmapped" {
		t.Fatalf("attribute 125 = %q/%q, want unmapped until its byte location is proven", field, confidence)
	}
	if field, _ := UGCAttributeDATField(145); field == "" {
		t.Fatal("attribute 145 lost its proven trailing-unit mapping")
	}
	for id, want := range tests {
		got, confidence := UGCAttributeDATField(id)
		if got != want || confidence != "layout_correlated" {
			t.Errorf("attribute %d = %q/%q, want %q/layout_correlated", id, got, confidence, want)
		}
	}
}

// The output must already have been built from the opt-in corpus DAT. Keeping
// verification separate lets the same persistent database serve later lookups.
func TestPersistentSQLiteExport(t *testing.T) {
	path := os.Getenv("AOE2KIT_TEST_SQLITE")
	if path == "" {
		t.Skip("set AOE2KIT_TEST_SQLITE to validate a persistent DAT export")
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("SQLite export unavailable: %v", err)
	}
	db, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var integrity string
	if err := db.QueryRow("PRAGMA quick_check").Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("SQLite integrity = %q, err=%v", integrity, err)
	}
	var units, costs, locations, graphics, creatables int
	for query, dest := range map[string]*int{
		"SELECT count(*) FROM units":                                        &units,
		"SELECT count(*) FROM unit_costs":                                   &costs,
		"SELECT count(*) FROM unit_train_locations":                         &locations,
		"SELECT count(*) FROM unit_damage_graphics":                         &graphics,
		"SELECT count(*) FROM unit_fields WHERE field_name='charge_target'": &creatables,
	} {
		if err := db.QueryRow(query).Scan(dest); err != nil {
			t.Fatal(err)
		}
	}
	if units == 0 || creatables == 0 || costs != 3*creatables || locations == 0 || graphics == 0 {
		t.Fatalf("units=%d creatables=%d costs=%d locations=%d graphics=%d", units, creatables, costs, locations, graphics)
	}
	coverage := exportAttributeCoverage()
	var exported, unexported int
	if err := db.QueryRow("SELECT sum(exported),sum(CASE WHEN exported=0 AND dat_field_name<>'' THEN 1 ELSE 0 END) FROM attribute_map").Scan(&exported, &unexported); err != nil {
		t.Fatal(err)
	}
	if exported != coverage.Exported || unexported != len(coverage.Unexported) {
		t.Fatalf("database coverage %d/%d differs from decoder %d/%d", exported, unexported, coverage.Exported, len(coverage.Unexported))
	}
	var sourceHash string
	if err := db.QueryRow("SELECT value FROM metadata WHERE key='source_sha256'").Scan(&sourceHash); err != nil || len(sourceHash) != 64 {
		t.Fatalf("source fingerprint=%q, err=%v", sourceHash, err)
	}
	if datPath := os.Getenv("AOE2KIT_TEST_DAT"); datPath != "" {
		if _, err := os.Stat(datPath); err != nil {
			t.Skipf("DAT unavailable for projection comparison: %v", err)
		}
		compressed, err := os.ReadFile(datPath)
		if err != nil {
			t.Fatal(err)
		}
		if hash := fmt.Sprintf("%x", sha256.Sum256(compressed)); hash != sourceHash {
			t.Fatalf("export source fingerprint %s differs from DAT %s", sourceHash, hash)
		}
		_, fields, err := DecodeUnitFile(datPath, 0, 109)
		if err != nil {
			t.Fatal(err)
		}
		for field, value := range fields {
			if !isScalarSQLValue(value) {
				continue
			}
			var stored string
			if err := db.QueryRow("SELECT field_value FROM unit_fields WHERE civ_index=0 AND unit_index=109 AND field_name=?", field).Scan(&stored); err != nil {
				t.Errorf("exported %s: %v", field, err)
			} else if stored != fmt.Sprint(value) {
				t.Errorf("exported %s=%s, decoded=%v", field, stored, value)
			}
		}
	}
	t.Logf("persistent export: units=%d costs=%d train_locations=%d damage_graphics=%d coverage=%d exported, %d gaps", units, costs, locations, graphics, exported, unexported)
}

func TestAttributeExportCoverageMarksOpaqueFields(t *testing.T) {
	coverage := exportAttributeCoverage()
	if coverage.Mapped == 0 || coverage.Exported == 0 || len(coverage.Unexported) == 0 {
		t.Fatalf("coverage = %+v, want mapped/exported/opaque values", coverage)
	}
	fields := exportedDATFields()
	for _, field := range []string{"trailing_unit", "trailing_options", "trailing_spacing", "blast_attack_level", "area_effect_specials", "blast_damage", "fly_mode"} {
		if !exportFieldAvailable(fields, field) {
			t.Errorf("newly decoded field %q was not marked exported", field)
		}
	}
	if !exportFieldAvailable(fields, "death_spawn") {
		t.Error("decoded death_spawn was not marked exported")
	}
	for _, field := range []string{"secondary_projectile_unit", "charge_target"} {
		if !exportFieldAvailable(fields, field) {
			t.Errorf("decoded creatable field %q was not marked exported", field)
		}
	}
	if exportFieldAvailable(fields, "unit_level") {
		t.Error("opaque unit_level was incorrectly marked exported")
	}
	for _, field := range []string{"spawning_graphic", "upgrade_graphic", "stack_unit", "head_unit", "transform_unit", "pile_unit", "annex_unit_1_offset_x", "annex_unit_4_offset_y"} {
		if !exportFieldAvailable(fields, field) {
			t.Errorf("decoded building field %q was not marked exported", field)
		}
	}
	if !exportFieldAvailable(fields, "attacks.value") || !exportFieldAvailable(fields, "armours.value") {
		t.Fatal("child weapon fields were not recognized as exported")
	}
}
