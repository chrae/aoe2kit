package aifile

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"aoe2kit/pkg/scenario"
	_ "modernc.org/sqlite"
)

type LintOptions struct {
	DATDatabase       string
	ExpectedDATSHA256 string
	Scenario          string
}

type DATCatalogue struct {
	Path                  string `json:"path"`
	SourceSHA256          string `json:"source_sha256"`
	Version               string `json:"version"`
	ExpectedSHA256        string `json:"expected_sha256,omitempty"`
	ExpectedSourceMatches *bool  `json:"expected_source_matches,omitempty"`
	sourceMatches         bool
}

type DATUnitMatch struct {
	UnitIndex    int    `json:"unit_index"`
	InternalName string `json:"internal_name"`
	StringID     int    `json:"string_id"`
	Civs         []int  `json:"civs"`
}

type ScenarioUnitRef struct {
	Trigger           int                     `json:"trigger"`
	TriggerName       string                  `json:"trigger_name"`
	Enabled           uint32                  `json:"enabled"`
	Looping           int8                    `json:"looping"`
	Kind              string                  `json:"kind"`
	Index             int                     `json:"index"`
	TypeName          string                  `json:"type_name"`
	Effect            *scenario.EffectSummary `json:"effect,omitempty"`
	Condition         map[string]any          `json:"condition,omitempty"`
	TriggerConditions []map[string]any        `json:"trigger_conditions,omitempty"`
	Confidence        string                  `json:"confidence"`
}

type ScenarioCatalogue struct {
	Path             string `json:"path"`
	SHA256           string `json:"sha256"`
	Confidence       string `json:"confidence"`
	PlacementsStatus string `json:"placements_status"`
}

type ScenarioPlacement struct {
	PlayerSection int     `json:"player_section"`
	UnitIndex     int     `json:"unit_index"`
	ReferenceID   int     `json:"reference_id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             float64 `json:"z"`
	Status        int     `json:"status"`
	Confidence    string  `json:"confidence"`
}

type UnitResolution struct {
	File                   string              `json:"file"`
	Line                   int                 `json:"line"`
	Column                 int                 `json:"column"`
	Command                string              `json:"command"`
	Operand                string              `json:"operand"`
	UnitID                 *int                `json:"unit_id,omitempty"`
	Status                 string              `json:"status"`
	Source                 string              `json:"source,omitempty"`
	Confidence             string              `json:"confidence"`
	DATMatches             []DATUnitMatch      `json:"dat_matches,omitempty"`
	ScenarioReferences     []ScenarioUnitRef   `json:"scenario_references,omitempty"`
	ScenarioPlacements     []ScenarioPlacement `json:"scenario_placements,omitempty"`
	ScenarioPlacementCount *int                `json:"scenario_placement_count,omitempty"`
}

type unitResolver struct {
	db                  *sql.DB
	meta                *DATCatalogue
	matches             map[int][]DATUnitMatch
	scenario            map[int][]ScenarioUnitRef
	placements          map[int][]ScenarioPlacement
	placementsAvailable bool
}

func openUnitResolver(opts LintOptions, report *LintReport) (*unitResolver, error) {
	r := &unitResolver{matches: map[int][]DATUnitMatch{}, scenario: map[int][]ScenarioUnitRef{}}
	if opts.DATDatabase == "" && opts.ExpectedDATSHA256 != "" {
		return nil, fmt.Errorf("expected DAT fingerprint requires --dat-db")
	}
	if opts.DATDatabase != "" {
		absolute, err := filepath.Abs(opts.DATDatabase)
		if err != nil {
			return nil, err
		}
		uri := url.URL{Scheme: "file", Path: absolute}
		db, err := sql.Open("sqlite", uri.String()+"?mode=ro")
		if err != nil {
			return nil, err
		}
		r.db = db
		db.SetMaxOpenConns(1)
		r.meta = &DATCatalogue{Path: opts.DATDatabase, ExpectedSHA256: strings.ToLower(opts.ExpectedDATSHA256), sourceMatches: true}
		for key, dest := range map[string]*string{"source_sha256": &r.meta.SourceSHA256, "dat_version": &r.meta.Version} {
			if err := db.QueryRow("SELECT value FROM metadata WHERE key=?", key).Scan(dest); err != nil {
				db.Close()
				return nil, fmt.Errorf("DAT SQLite export metadata %s: %w; select an export built with kit dat sql", key, err)
			}
		}
		for _, hash := range []string{r.meta.SourceSHA256, r.meta.ExpectedSHA256} {
			if hash == "" {
				continue
			}
			decoded, err := hex.DecodeString(hash)
			if err != nil || len(decoded) != 32 {
				db.Close()
				return nil, fmt.Errorf("DAT fingerprint must be a 64-digit SHA-256")
			}
		}
		if len(r.meta.SourceSHA256) != 64 {
			db.Close()
			return nil, fmt.Errorf("DAT export lacks a source fingerprint; rebuild with kit dat sql")
		}
		r.meta.SourceSHA256 = strings.ToLower(r.meta.SourceSHA256)
		if r.meta.ExpectedSHA256 != "" {
			match := r.meta.ExpectedSHA256 == r.meta.SourceSHA256
			r.meta.ExpectedSourceMatches = &match
		}
		if r.meta.ExpectedSHA256 != "" && r.meta.ExpectedSHA256 != r.meta.SourceSHA256 {
			r.meta.sourceMatches = false
			report.addIssue("warning", "dat_source_mismatch", opts.DATDatabase, "selected export differs from the expected DAT fingerprint; unit identities remain unresolved")
		}
		report.DATCatalogue = r.meta
	}
	if opts.Scenario != "" {
		data, err := os.ReadFile(opts.Scenario)
		if err != nil {
			r.close()
			return nil, err
		}
		f, err := scenario.Parse(data)
		if err != nil {
			r.close()
			return nil, fmt.Errorf("AI scenario context: %w", err)
		}
		dumps, err := f.DumpTriggersWithConditions()
		if err != nil {
			r.close()
			return nil, err
		}
		r.scenario = scenarioUnitReferences(dumps)
		report.ScenarioCatalogue = &ScenarioCatalogue{Path: opts.Scenario, SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), Confidence: "scenario_structure"}
		r.placements = scenarioUnitPlacements(f.Units)
		r.placementsAvailable = f.Units != nil
		report.ScenarioCatalogue.PlacementsStatus = "unavailable"
		if r.placementsAvailable {
			report.ScenarioCatalogue.PlacementsStatus = "indexed"
		}
	}
	return r, nil
}

func (r *unitResolver) close() {
	if r != nil && r.db != nil {
		r.db.Close()
	}
}

func scenarioUnitPlacements(info *scenario.UnitInfo) map[int][]ScenarioPlacement {
	if info == nil {
		return nil
	}
	placements := map[int][]ScenarioPlacement{}
	for _, section := range info.Sections {
		for _, unit := range section.Units {
			placements[unit.UnitConst] = append(placements[unit.UnitConst], ScenarioPlacement{
				PlayerSection: section.Player, UnitIndex: unit.Index, ReferenceID: unit.ReferenceID,
				X: unit.X, Y: unit.Y, Z: unit.Z, Status: unit.Status, Confidence: "scenario_structure",
			})
		}
	}
	return placements
}

func scenarioUnitReferences(triggers []scenario.TriggerConditionDump) map[int][]ScenarioUnitRef {
	refs := map[int][]ScenarioUnitRef{}
	for _, trigger := range triggers {
		base := ScenarioUnitRef{Trigger: trigger.Index, TriggerName: trigger.Name, Enabled: trigger.Enabled, Looping: trigger.Looping, TriggerConditions: trigger.Conditions, Confidence: "scenario_structure"}
		for i, effect := range trigger.Effects {
			_, hasUnit := effect.KnownFields["unit_const"]
			if !hasUnit || effect.UnitConst < 0 {
				continue
			}
			ref := base
			ref.Kind = "effect"
			ref.Index = i
			ref.TypeName = effect.TypeName
			decoded := effect
			decoded.TriggerIndex = trigger.Index
			decoded.TriggerName = trigger.Name
			decoded.EffectIndex = i
			ref.Effect = &decoded
			refs[effect.UnitConst] = append(refs[effect.UnitConst], ref)
		}
		for i, condition := range trigger.Conditions {
			raw, found := condition["object_list"]
			if !found {
				continue
			}
			id, err := strconv.Atoi(fmt.Sprint(raw))
			if err != nil || id < 0 {
				continue
			}
			ref := base
			ref.Kind = "condition"
			ref.Index = i
			ref.TypeName = fmt.Sprint(condition["type_name"])
			ref.Condition = condition
			refs[id] = append(refs[id], ref)
		}
	}
	return refs
}

func (r *unitResolver) lookup(id int) ([]DATUnitMatch, error) {
	if matches, found := r.matches[id]; found {
		return matches, nil
	}
	rows, err := r.db.Query("SELECT unit_index,name,string_id,civ_index FROM units WHERE unit_id=? ORDER BY unit_index,name,string_id,civ_index", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var matches []DATUnitMatch
	for rows.Next() {
		var index, stringID, civ int
		var name string
		if err := rows.Scan(&index, &name, &stringID, &civ); err != nil {
			return nil, err
		}
		if len(matches) == 0 || matches[len(matches)-1].UnitIndex != index || matches[len(matches)-1].InternalName != name || matches[len(matches)-1].StringID != stringID {
			matches = append(matches, DATUnitMatch{UnitIndex: index, InternalName: name, StringID: stringID})
		}
		last := &matches[len(matches)-1]
		last.Civs = append(last.Civs, civ)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	r.matches[id] = matches
	return matches, nil
}

func (c *perChecker) resolveUnit(arg, op *perNode) {
	r := c.report.unitResolver
	result := UnitResolution{File: c.file, Line: arg.line, Column: arg.column, Command: c.command, Operand: arg.text, Status: "unresolved_operand", Confidence: "unverified"}
	defer func() { c.report.Units = append(c.report.Units, result) }()
	value, ok := c.resolve(arg)
	if op.text != "c:" || !ok || value.quoted {
		return
	}
	id, err := strconv.Atoi(value.text)
	if err != nil {
		return
	}
	result.UnitID = &id
	result.ScenarioReferences = r.scenario[id]
	if r.placementsAvailable {
		result.ScenarioPlacements = r.placements[id]
		count := len(result.ScenarioPlacements)
		result.ScenarioPlacementCount = &count
	}
	if r.db == nil {
		result.Status = "no_dat_catalogue_selected"
		return
	}
	result.Source = r.meta.Path
	if !r.meta.sourceMatches {
		result.Status = "dat_source_mismatch"
		return
	}
	matches, err := r.lookup(id)
	if err != nil {
		result.Status = "catalogue_query_failed"
		c.report.perIssue("warning", "unit_catalogue_query_failed", c.file, arg, err.Error(), r.meta.Path, "unverified")
		return
	}
	result.DATMatches = matches
	if len(matches) == 0 {
		result.Status = "not_in_selected_dat"
		c.report.perIssue("warning", "unit_id_unresolved", c.file, arg, fmt.Sprintf("unit id %d has no match in the selected DAT; it may be a class/line/scenario selector or from another DAT, so it is not rejected", id), r.meta.Path, "unverified")
		return
	}
	result.Status = "dat_match"
	result.Confidence = "dat_projection"
	if len(matches) > 1 {
		identityVariants := false
		for _, match := range matches[1:] {
			if match.UnitIndex != matches[0].UnitIndex || match.InternalName != matches[0].InternalName {
				identityVariants = true
				break
			}
		}
		if !identityVariants {
			result.Status = "dat_match_civ_variants"
			return
		}
		result.Status = "ambiguous_dat_match"
		c.report.perIssue("warning", "unit_id_ambiguous", c.file, arg, fmt.Sprintf("unit id %d has multiple civ-scoped names/records; no canonical display name is inferred", id), r.meta.Path, "dat_projection")
	}
}
