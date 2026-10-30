package aifile

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aoe2kit/pkg/scenario"
)

func unitDatabaseFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "units # catalogue.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE metadata(key TEXT PRIMARY KEY,value TEXT);
CREATE TABLE units(civ_index INTEGER,unit_index INTEGER,unit_id INTEGER,name TEXT,string_id INTEGER);
INSERT INTO units VALUES(0,569,569,'HEROC',123),(1,569,569,'HEROC',123),(2,569,569,'HEROC',789),(0,83,83,'WORKER',456),(1,83,83,'WORKER_VARIANT',456);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO metadata VALUES('source_sha256',?),('dat_version','VER 8.9')", strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestScenarioPlacementsRemainSeparateFromTriggers(t *testing.T) {
	info := &scenario.UnitInfo{Sections: []scenario.PlayerUnitsInfo{
		{Player: 0, Units: []scenario.UnitSummary{{Index: 2, ReferenceID: 10, UnitConst: 432, X: 1, Y: 2, Z: 3, Status: 4}}},
		{Player: 1, Units: []scenario.UnitSummary{{Index: 7, ReferenceID: 20, UnitConst: 557, X: 5, Y: 6, Z: 7, Status: 8}}},
		{Player: 2, Units: []scenario.UnitSummary{{Index: 9, ReferenceID: 30, UnitConst: 557, X: 9, Y: 10, Z: 11, Status: 12}}},
	}}
	r := &unitResolver{placements: scenarioUnitPlacements(info), placementsAvailable: true, scenario: map[int][]ScenarioUnitRef{
		569: {{Trigger: 1, TriggerName: "Cavalry replenishment", Kind: "effect", Index: 0, TypeName: "create_object"}},
	}}
	report := &LintReport{unitResolver: r}
	checkPER("Probe.per", "(defrule (true) => (up-find-local c: 432 c: 10) (up-find-local c: 557 c: 10) (up-find-local c: 569 c: 10) (up-find-local g: 100 c: 10))", report)
	if len(report.Units) != 4 {
		t.Fatalf("units=%+v", report.Units)
	}
	first, second, triggerOnly, dynamic := report.Units[0], report.Units[1], report.Units[2], report.Units[3]
	if first.ScenarioPlacementCount == nil || *first.ScenarioPlacementCount != 1 || len(first.ScenarioReferences) != 0 {
		t.Fatalf("placed-only evidence=%+v", first)
	}
	p := first.ScenarioPlacements[0]
	if p.PlayerSection != 0 || p.UnitIndex != 2 || p.ReferenceID != 10 || p.X != 1 || p.Y != 2 || p.Z != 3 || p.Status != 4 || p.Confidence != "scenario_structure" {
		t.Fatalf("placement changed=%+v", p)
	}
	if second.ScenarioPlacementCount == nil || *second.ScenarioPlacementCount != 2 || second.ScenarioPlacements[1].PlayerSection != 2 {
		t.Fatalf("player grouping=%+v", second)
	}
	if triggerOnly.ScenarioPlacementCount == nil || *triggerOnly.ScenarioPlacementCount != 0 || len(triggerOnly.ScenarioReferences) != 1 {
		t.Fatalf("trigger-only evidence=%+v", triggerOnly)
	}
	data, err := json.Marshal(triggerOnly)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"scenario_placement_count":0`) {
		t.Fatalf("zero placement count omitted: %s", data)
	}
	if dynamic.ScenarioPlacementCount != nil {
		t.Fatal("dynamic selector claimed placement absence")
	}
	r.placementsAvailable = false
	report.Units = nil
	checkPER("Probe.per", "(defrule (true) => (up-find-local c: 569 c: 10))", report)
	if report.Units[0].ScenarioPlacementCount != nil {
		t.Fatal("unavailable placement section claimed zero matches")
	}
	if scenarioUnitPlacements(nil) != nil {
		t.Fatal("missing unit section was indexed as present")
	}
}

func TestScenarioReferencesPreserveOperandsAndSiblingConditions(t *testing.T) {
	conditions := []map[string]any{
		{"type_name": "own_fewer_objects", "object_list": 569, "source_player": 1, "quantity": 4},
		{"type_name": "timer", "timer": 7},
	}
	refs := scenarioUnitReferences([]scenario.TriggerConditionDump{{Index: 3, Name: "Configured replenishment", Enabled: 1, Looping: 1, Conditions: conditions,
		Effects: []scenario.EffectSummary{
			{UnitConst: 569, TypeName: "create_object", SourcePlayer: 1, Location: []int{10, 20}, KnownFields: map[string]any{"unit_const": 569, "source_player": 1, "target_player": 0, "location_x": 10, "location_y": 20, "quantity": 0}},
			{UnitConst: 83, TypeName: "create_object", KnownFields: map[string]any{"unit_const": 83, "source_player": 2}},
		},
	}})
	if len(refs[569]) != 2 {
		t.Fatalf("refs=%+v", refs)
	}
	effect, condition := refs[569][0], refs[569][1]
	if effect.Trigger != 3 || effect.Index != 0 || effect.Confidence != "scenario_structure" || effect.Effect == nil || effect.Effect.TriggerIndex != 3 || effect.Effect.SourcePlayer != 1 || effect.Effect.Location[1] != 20 || effect.Effect.KnownFields["quantity"] != 0 {
		t.Fatalf("effect evidence=%+v", effect)
	}
	if condition.Condition["quantity"] != 4 || condition.Condition["source_player"] != 1 || condition.Index != 0 || condition.Effect != nil {
		t.Fatalf("condition evidence=%+v", condition)
	}
	if len(effect.TriggerConditions) != 2 || effect.TriggerConditions[1]["timer"] != 7 || condition.TriggerConditions[1]["type_name"] != "timer" {
		t.Fatalf("missing sibling condition: %+v", effect.TriggerConditions)
	}
	if refs[83][0].Effect == effect.Effect || refs[83][0].Effect.UnitConst != 83 {
		t.Fatal("effect records alias different operands")
	}
	data, err := json.Marshal(effect)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"target_player":0`) || !strings.Contains(string(data), `"quantity":0`) {
		t.Fatalf("decoded zeros dropped: %s", data)
	}
}

func TestUnitIdentityAndScenarioReferencesStaySeparate(t *testing.T) {
	path := unitDatabaseFixture(t)
	report := &LintReport{}
	r, err := openUnitResolver(LintOptions{DATDatabase: path, ExpectedDATSHA256: strings.Repeat("a", 64)}, report)
	if err != nil {
		t.Fatal(err)
	}
	defer r.close()
	report.unitResolver = r
	r.scenario = scenarioUnitReferences([]scenario.TriggerConditionDump{{Index: 1, Name: "P1 Cavalry role", Enabled: 1, Looping: 1,
		Effects:    []scenario.EffectSummary{{UnitConst: 569, TypeName: "create_object", KnownFields: map[string]any{"unit_const": 569}}, {TypeName: "task_object"}},
		Conditions: []map[string]any{{"object_list": 569, "type_name": "own_fewer_objects"}},
	}})
	checkPER("Probe.per", "(defconst unit 569) (defrule (true) => (up-find-local c: unit c: 240))", report)
	if len(report.Units) != 1 {
		t.Fatalf("units=%+v", report.Units)
	}
	u := report.Units[0]
	if u.Status != "dat_match_civ_variants" || u.Confidence != "dat_projection" || len(u.DATMatches) != 2 || u.DATMatches[0].InternalName != "HEROC" || len(u.DATMatches[0].Civs) != 2 {
		t.Fatalf("DAT identity=%+v", u)
	}
	if len(u.ScenarioReferences) != 2 || u.ScenarioReferences[0].TriggerName != "P1 Cavalry role" || u.ScenarioReferences[0].Trigger != 1 || u.ScenarioReferences[1].Kind != "condition" {
		t.Fatalf("scenario references=%+v", u.ScenarioReferences)
	}
	if len(r.scenario[0]) != 0 {
		t.Fatal("effect with no unit field created a spurious unit zero reference")
	}
	if report.DATCatalogue.SourceSHA256 != strings.Repeat("a", 64) || report.DATCatalogue.ExpectedSourceMatches == nil || !*report.DATCatalogue.ExpectedSourceMatches {
		t.Fatalf("provenance=%+v", report.DATCatalogue)
	}
	if _, err := r.db.Exec("DELETE FROM units"); err == nil {
		t.Fatal("catalogue connection permits writes")
	}
}

func TestUnitResolutionMissingAmbiguousAndDynamic(t *testing.T) {
	r := &LintReport{}
	resolver, err := openUnitResolver(LintOptions{DATDatabase: unitDatabaseFixture(t)}, r)
	if err != nil {
		t.Fatal(err)
	}
	defer resolver.close()
	r.unitResolver = resolver
	checkPER("Probe.per", "(defrule (true) => (up-find-local c: 83 c: 10) (up-find-local c: 999999 c: 10) (up-find-local g: 100 c: 10))", r)
	if len(r.Units) != 3 || r.Units[0].Status != "ambiguous_dat_match" || r.Units[1].Status != "not_in_selected_dat" || r.Units[2].Status != "unresolved_operand" {
		t.Fatalf("resolutions=%+v", r.Units)
	}
	issueWithCode(t, r, "unit_id_ambiguous")
	issueWithCode(t, r, "unit_id_unresolved")
	for _, issue := range r.Issues {
		if issue.Severity == "error" {
			t.Fatalf("uncertain unit rejected: %+v", issue)
		}
	}
	if r.DATCatalogue.ExpectedSourceMatches != nil {
		t.Fatal("freshness claimed without an expected fingerprint")
	}
}

func TestUnitCatalogueFingerprintMismatchAndMissingFile(t *testing.T) {
	report := &LintReport{}
	r, err := openUnitResolver(LintOptions{DATDatabase: unitDatabaseFixture(t), ExpectedDATSHA256: strings.Repeat("b", 64)}, report)
	if err != nil {
		t.Fatal(err)
	}
	defer r.close()
	report.unitResolver = r
	checkPER("Probe.per", "(defrule (true) => (up-find-local c: 569 c: 10))", report)
	issueWithCode(t, report, "dat_source_mismatch")
	if len(report.Units) != 1 || report.Units[0].Status != "dat_source_mismatch" || len(report.Units[0].DATMatches) != 0 {
		t.Fatalf("stale source resolved: %+v", report.Units)
	}
	path := filepath.Join(t.TempDir(), "missing.sqlite")
	if _, err := openUnitResolver(LintOptions{DATDatabase: path}, &LintReport{}); err == nil {
		t.Fatal("missing database accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("read-only open created missing database: %v", err)
	}
	if _, err := openUnitResolver(LintOptions{ExpectedDATSHA256: strings.Repeat("a", 64)}, &LintReport{}); err == nil {
		t.Fatal("fingerprint accepted without database")
	}
}

func TestLoaderPairPresenceDoesNotClaimRuntime(t *testing.T) {
	r, err := LintPath("../../testdata/ai/passive")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Pairs) != 1 || r.Pairs[0].Status != "present" || len(r.Pairs[0].Scripts) != 1 || r.Pairs[0].RuntimeStatus != "unverified" {
		t.Fatalf("pairs=%+v", r.Pairs)
	}
	issueWithCode(t, r, "ai_runtime_loading_unverified")
	for _, issue := range r.Issues {
		if issue.Code == "ai_pair_unresolved" {
			t.Fatal("present pair marked unresolved")
		}
	}
	path := filepath.Join(t.TempDir(), "Alone.ai")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	r, err = LintPath(path)
	if err != nil {
		t.Fatal(err)
	}
	issueWithCode(t, r, "ai_pair_unresolved")
}
