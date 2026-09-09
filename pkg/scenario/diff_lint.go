package scenario

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"aoe2kit/pkg/aoe2"
	"aoe2kit/pkg/enginefacts"
)

type DiffReport struct {
	Before  string       `json:"before"`
	After   string       `json:"after"`
	Same    bool         `json:"same"`
	Changes []DiffChange `json:"changes"`
}

type DiffChange struct {
	Kind   string `json:"kind"`
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type LintReport struct {
	Path         string                 `json:"path,omitempty"`
	OK           bool                   `json:"ok"`
	Verification aoe2.VerificationClaim `json:"verification"`
	Issues       []LintIssue            `json:"issues,omitempty"`
	Summary      LintSummary            `json:"summary"`
}

type LintIssue struct {
	Severity     string `json:"severity"`
	Code         string `json:"code"`
	Message      string `json:"message"`
	FactID       string `json:"fact_id,omitempty"`
	FactTier     string `json:"fact_tier,omitempty"`
	VerifiedDate string `json:"verified_date,omitempty"`
	FixtureRef   string `json:"fixture_ref,omitempty"`
}

type LintSummary struct {
	Triggers int `json:"triggers"`
	Units    int `json:"units"`
	Players  int `json:"players"`
	AIFiles  int `json:"ai_files"`
}

type LintOptions struct {
	IncludeProvisional bool
}

func DiffFiles(beforePath, afterPath string) (DiffReport, error) {
	before, err := Open(beforePath)
	if err != nil {
		return DiffReport{}, fmt.Errorf("open before: %w", err)
	}
	after, err := Open(afterPath)
	if err != nil {
		return DiffReport{}, fmt.Errorf("open after: %w", err)
	}
	return Diff(before, after), nil
}

func Diff(before, after *File) DiffReport {
	report := DiffReport{Before: before.Path, After: after.Path}
	addChange := func(kind, field string, b, a any, detail string) {
		report.Changes = append(report.Changes, DiffChange{Kind: kind, Field: field, Before: b, After: a, Detail: detail})
	}
	if before.Version != after.Version {
		addChange("scenario", "version", before.Version, after.Version, "")
	}
	if before.InflatedBytes != after.InflatedBytes {
		addChange("scenario", "inflated_body_bytes", before.InflatedBytes, after.InflatedBytes, "")
	}
	if before.Map != nil && after.Map != nil {
		if before.Map.Width != after.Map.Width || before.Map.Height != after.Map.Height {
			addChange("map", "dimensions", fmt.Sprintf("%dx%d", before.Map.Width, before.Map.Height), fmt.Sprintf("%dx%d", after.Map.Width, after.Map.Height), "")
		}
		for _, terrainID := range sortedTerrainIDs(before.Map.TerrainCounts, after.Map.TerrainCounts) {
			b := before.Map.TerrainCounts[terrainID]
			a := after.Map.TerrainCounts[terrainID]
			if b != a {
				addChange("map", fmt.Sprintf("terrain_%d_count", terrainID), b, a, "")
			}
		}
	}
	if before.Triggers != nil && after.Triggers != nil {
		if before.Triggers.Count != after.Triggers.Count {
			addChange("triggers", "count", before.Triggers.Count, after.Triggers.Count, "")
		}
		if before.Triggers.GraphSHA256 != after.Triggers.GraphSHA256 {
			addChange("triggers", "graph_sha256", before.Triggers.GraphSHA256, after.Triggers.GraphSHA256, "")
		}
		diffTriggerSummaries(before.Triggers.Triggers, after.Triggers.Triggers, addChange)
	}
	if before.Units != nil && after.Units != nil {
		if before.Units.NumberOfUnitSections != after.Units.NumberOfUnitSections {
			addChange("units", "number_of_unit_sections", before.Units.NumberOfUnitSections, after.Units.NumberOfUnitSections, "")
		}
		if before.Units.NumberOfPlayers != after.Units.NumberOfPlayers {
			addChange("units", "number_of_players", before.Units.NumberOfPlayers, after.Units.NumberOfPlayers, "")
		}
		if before.Units.Total != after.Units.Total {
			addChange("units", "total", before.Units.Total, after.Units.Total, "")
		}
		beforeCounts := unitCountsByPlayer(before.Units)
		afterCounts := unitCountsByPlayer(after.Units)
		for _, player := range sortedIntKeys(beforeCounts, afterCounts) {
			if beforeCounts[player] != afterCounts[player] {
				addChange("units", fmt.Sprintf("player_%d_count", player), beforeCounts[player], afterCounts[player], "")
			}
		}
	}
	diffPlayers(before.Players, after.Players, addChange)
	diffAI(before.AI, after.AI, addChange)
	report.Same = len(report.Changes) == 0
	return report
}

func LintFile(path string) (LintReport, error) {
	return LintFileWithOptions(path, LintOptions{})
}

func LintFileWithOptions(path string, opts LintOptions) (LintReport, error) {
	file, err := Open(path)
	if err != nil {
		return LintReport{}, err
	}
	return file.LintWithOptions(opts), nil
}

func (f *File) Lint() LintReport {
	return f.LintWithOptions(LintOptions{})
}

func (f *File) LintWithOptions(opts LintOptions) LintReport {
	report := LintReport{
		Path: f.Path,
		Summary: LintSummary{
			Players: len(f.Players),
			AIFiles: len(f.AI),
		},
	}
	if f.Triggers != nil {
		report.Summary.Triggers = f.Triggers.Count
		if !f.Triggers.InvariantOK {
			report.addIssue("error", "trigger_invariant", f.Triggers.InvariantNote)
		}
		seenNames := map[string]int{}
		for _, trigger := range f.Triggers.Triggers {
			if trigger.Name != "" {
				seenNames[trigger.Name]++
			}
			if trigger.Enabled != 0 && trigger.Effects == 0 && trigger.Conditions == 0 {
				report.addIssue("warning", "empty_enabled_trigger", fmt.Sprintf("trigger %d %q is enabled with no effects or conditions", trigger.Index, trigger.Name))
			}
		}
		if f.root != nil {
			if section := f.root.section("Triggers"); section != nil {
				lintTriggerRuntimeHazards(&report, section.list("trigger_data"), opts)
			}
			lintTriggerPlayerCoverage(&report, f, opts)
		}
		lintCorruptStrings(&report, f.headerRoot)
		if f.root != nil {
			for _, section := range f.root.Sections {
				lintCorruptStrings(&report, section)
			}
		}
		for name, count := range seenNames {
			if count > 1 {
				report.addIssue("warning", "duplicate_trigger_name", fmt.Sprintf("%q appears %d times", name, count))
			}
		}
	} else {
		report.addIssue("error", "missing_triggers", "scenario has no parsed trigger section")
	}
	if f.Map != nil {
		if f.Map.Width*f.Map.Height != f.Map.TileCount {
			report.addIssue("error", "map_tile_count", fmt.Sprintf("map dimensions %dx%d imply %d tiles but parsed %d", f.Map.Width, f.Map.Height, f.Map.Width*f.Map.Height, f.Map.TileCount))
		}
		if err := ValidateScenarioMapSize(f.Map.Width, f.Map.Height); err != nil {
			report.addIssueFact("error", "map_size_preset", err.Error(), "scenario.map_size_editor_presets_are_discrete", opts)
		}
	} else {
		report.addIssue("error", "missing_map", "scenario has no parsed map section")
	}
	if f.Units != nil {
		report.Summary.Units = f.Units.Total
		if f.Units.NumberOfUnitSections != len(f.Units.Sections) {
			report.addIssue("error", "unit_section_count_mismatch", fmt.Sprintf("Units number_of_unit_sections=%d but parsed %d player unit sections", f.Units.NumberOfUnitSections, len(f.Units.Sections)))
		}
		if f.Units.NumberOfPlayers != len(f.Units.Sections) {
			report.addIssue("error", "unit_owner_count_mismatch", fmt.Sprintf("Units number_of_players=%d but parsed %d player unit sections; current DE editor-authored scenarios use all Gaia+8 unit-owner sections", f.Units.NumberOfPlayers, len(f.Units.Sections)))
		}
		seenRefs := map[int]string{}
		unitCountsByPlayer := map[int]int{}
		for _, player := range f.Units.Sections {
			unitCountsByPlayer[player.Player] = len(player.Units)
			if player.Count != len(player.Units) {
				report.addIssue("error", "unit_count_mismatch", fmt.Sprintf("player %d unit_count=%d but parsed %d units", player.Player, player.Count, len(player.Units)))
			}
			for _, unit := range player.Units {
				if unit.ReferenceID < 0 {
					report.addIssue("warning", "negative_unit_reference", fmt.Sprintf("player %d unit %d has reference_id %d", player.Player, unit.Index, unit.ReferenceID))
				}
				if unit.ReferenceID == 0 {
					continue
				}
				where := fmt.Sprintf("P%d[%d]", player.Player, unit.Index)
				if prev, ok := seenRefs[unit.ReferenceID]; ok {
					report.addIssue("error", "duplicate_unit_reference", fmt.Sprintf("reference_id %d appears at %s and %s", unit.ReferenceID, prev, where))
				} else {
					seenRefs[unit.ReferenceID] = where
				}
			}
		}
		lintActiveComputerStarterInjection(&report, f.Players, unitCountsByPlayer, opts)
	} else {
		report.addIssue("error", "missing_units", "scenario has no parsed unit section")
	}
	for _, player := range f.Players {
		if player.Active && !player.Human && player.AIName == "" && player.AIType != 0 {
			report.addIssue("warning", "active_ai_without_name", fmt.Sprintf("P%d is active non-human with ai_type=%d and no AI name", player.Player, player.AIType))
		}
	}
	for _, ai := range f.AI {
		if ai.Name == "" {
			report.addIssue("warning", "embedded_ai_missing_name", fmt.Sprintf("embedded AI index %d has no name", ai.Index))
		}
		if ai.Name != "" && ai.ContentBytes == 0 {
			report.addIssue("warning", "embedded_ai_empty_content", fmt.Sprintf("embedded AI %q has no content", ai.Name))
		}
	}
	report.OK = !hasLintSeverity(report.Issues, "error")
	report.Verification = aoe2.StructureVerification(report.OK)
	return report
}

func lintTriggerPlayerCoverage(report *LintReport, file *File, _ LintOptions) {
	if file == nil {
		return
	}
	var players []int
	for _, player := range file.Players {
		if player.Player > 0 && player.Active {
			players = append(players, player.Player)
		}
	}
	if len(players) < 2 {
		return
	}
	rows := collectTriggerRows(file)
	type familyState struct {
		players map[int]bool
		samples []TriggerBrief
	}
	families := map[string]*familyState{}
	for _, row := range rows {
		key, namedPlayer := playerFamilyKey(row.Name)
		if key == "" {
			continue
		}
		touched := rowTouchedPlayers(row)
		if namedPlayer > 0 {
			touched[namedPlayer] = true
		}
		if len(touched) == 0 {
			continue
		}
		state := families[key]
		if state == nil {
			state = &familyState{players: map[int]bool{}}
			families[key] = state
		}
		for player := range touched {
			if containsInt(players, player) {
				state.players[player] = true
			}
		}
		if len(state.samples) < 3 {
			state.samples = append(state.samples, TriggerBrief{Index: row.Index, Name: row.Name, Enabled: row.Enabled, Looping: row.Looping})
		}
	}
	for family, state := range families {
		if len(state.players) == 0 || len(state.players) == len(players) {
			continue
		}
		var missing []int
		for _, player := range players {
			if !state.players[player] {
				missing = append(missing, player)
			}
		}
		report.addIssue("warning", "trigger_player_family_asymmetry", fmt.Sprintf("trigger family %q touches players %v but is missing %v; this can indicate a per-slot refresh/transition bug", family, sortedBoolKeys(state.players), missing))
	}
}

func lintActiveComputerStarterInjection(report *LintReport, players []PlayerInfo, unitCountsByPlayer map[int]int, opts LintOptions) {
	for _, player := range players {
		if player.Player <= 0 || !player.Active || player.Human {
			continue
		}
		if unitCountsByPlayer[player.Player] == 0 {
			report.addIssueFact("warning", "active_computer_without_authored_units", fmt.Sprintf("P%d is active non-human with no authored units; DE scenario-mode runs may inject starter TC/villager/scout state and contaminate replay diagnostics", player.Player), "scenario.active_computer_starter_injection", opts)
		}
	}
}

func lintTriggerRuntimeHazards(report *LintReport, triggerNodes []*parsedNode, opts LintOptions) {
	for i, trigger := range triggerNodes {
		looping, _ := trigger.intValue("looping")
		name, _ := trigger.stringValue("trigger_name")
		for j, effect := range trigger.list("effect_data") {
			effectType, _ := effect.intValue("effect_type")
			if looping != 0 && effectType == 20 {
				report.addIssueFact("error", "looping_display_effect", fmt.Sprintf("trigger %d %q is looping and effect %d is display-instructions; timer conditions become permanently true after elapsed and can spam text every tick", i, name, j), "scenario.looping_display_effect_spams", opts)
			}
			if text, ok := effect.stringValue("message"); ok && text != "" {
				lintTriggerTextMarkup(report, i, name, j, effectType, text, opts)
			}
		}
	}
}

func lintTriggerTextMarkup(report *LintReport, triggerIndex int, triggerName string, effectIndex int, effectType int, text string, opts LintOptions) {
	if !isTriggerTextChannel(effectType) {
		return
	}
	if strings.Contains(strings.ToUpper(text), "<WHITE>") {
		report.addIssueFact("warning", "unsupported_white_color_tag", fmt.Sprintf("trigger %d %q effect %d contains <WHITE>; DE trigger text color markup does not support that tag", triggerIndex, triggerName, effectIndex), "text.white_tag_not_supported", opts)
	}
	colorTags := triggerColorTagPositions(text)
	if len(colorTags) == 0 {
		return
	}
	if len(colorTags) > 1 || colorTags[0] != 0 {
		report.addIssueFact("warning", "trigger_color_tag_not_leading_singleton", fmt.Sprintf("trigger %d %q effect %d has %d color tags and first tag offset %d; DE trigger text applies only one leading color tag", triggerIndex, triggerName, effectIndex, len(colorTags), colorTags[0]), "text.trigger_color_first_tag_only", opts)
	}
}

func isTriggerTextChannel(effectType int) bool {
	switch effectType {
	case 3, 20, 37, 44, 55, 65, 66, 88:
		return true
	default:
		return false
	}
}

func triggerColorTagPositions(text string) []int {
	upper := strings.ToUpper(text)
	var positions []int
	for _, tag := range []string{"<RED>", "<GREEN>", "<YELLOW>", "<BLUE>", "<AQUA>", "<PURPLE>", "<ORANGE>", "<GREY>", "<GRAY>", "<WHITE>"} {
		start := 0
		for {
			idx := strings.Index(upper[start:], tag)
			if idx < 0 {
				break
			}
			positions = append(positions, start+idx)
			start += idx + len(tag)
		}
	}
	sort.Ints(positions)
	return positions
}

func lintCorruptStrings(report *LintReport, roots ...*parsedNode) {
	for _, root := range roots {
		lintCorruptStringNode(report, root, rootName(root))
	}
}

func rootName(node *parsedNode) string {
	if node == nil || node.Name == "" {
		return "root"
	}
	return node.Name
}

func lintCorruptStringNode(report *LintReport, node *parsedNode, path string) {
	if node == nil {
		return
	}
	if text, ok := node.Value.(string); ok {
		lintStringValue(report, path, node, text)
	}
	for _, field := range node.Fields {
		childPath := field.Name
		if path != "" {
			childPath = path + "." + field.Name
		}
		lintCorruptStringNode(report, field, childPath)
	}
	for i, elem := range node.Elements {
		childPath := fmt.Sprintf("%s[%d]", path, i)
		if elem.Name != "" {
			childPath = fmt.Sprintf("%s.%s", childPath, elem.Name)
		}
		lintCorruptStringNode(report, elem, childPath)
	}
}

func lintStringValue(report *LintReport, path string, node *parsedNode, text string) {
	if !utf8.ValidString(text) {
		report.addIssue("error", "corrupt_string_invalid_utf8", fmt.Sprintf("%s contains invalid UTF-8 at scenario byte offset %d", path, node.Start))
	}
	for offset, r := range text {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			report.addIssue("error", "corrupt_string_control_byte", fmt.Sprintf("%s contains control byte 0x%02x at string offset %d scenario byte offset %d", path, r, offset, node.Start+stringPayloadPrefixLen(node)+offset))
			break
		}
	}
	payload, payloadOffset, ok := stringPayloadBytes(node)
	if ok {
		for i, b := range payload {
			if b == 0 && i < len(payload)-1 {
				report.addIssue("error", "corrupt_string_embedded_nul", fmt.Sprintf("%s contains NUL before declared string end at payload offset %d scenario byte offset %d", path, i, node.Start+payloadOffset+i))
				break
			}
		}
	}
}

func stringPayloadPrefixLen(node *parsedNode) int {
	_, prefix, ok := stringPayloadBytes(node)
	if !ok {
		return 0
	}
	return prefix
}

func stringPayloadBytes(node *parsedNode) ([]byte, int, bool) {
	raw := node.Raw
	for _, prefixSize := range []int{2, 4} {
		if len(raw) < prefixSize {
			continue
		}
		length := signedInt(raw[:prefixSize])
		if length >= 0 && prefixSize+length == len(raw) {
			return raw[prefixSize:], prefixSize, true
		}
	}
	return nil, 0, false
}

func (r *LintReport) addIssue(severity, code, message string) {
	r.Issues = append(r.Issues, LintIssue{Severity: severity, Code: code, Message: message})
}

func (r *LintReport) addIssueFact(severity, code, message, factID string, opts LintOptions) {
	issue := LintIssue{Severity: severity, Code: code, Message: message}
	citation := enginefacts.CitationFor(factID, opts.IncludeProvisional)
	issue.FactID = citation.FactID
	issue.FactTier = citation.Tier
	issue.VerifiedDate = citation.VerifiedDate
	issue.FixtureRef = citation.FixtureRef
	r.Issues = append(r.Issues, issue)
}

func hasLintSeverity(issues []LintIssue, severity string) bool {
	for _, issue := range issues {
		if issue.Severity == severity {
			return true
		}
	}
	return false
}

func sortedTerrainIDs(a, b map[int]int) []int {
	seen := map[int]bool{}
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	out := make([]int, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func sortedIntKeys(a, b map[int]int) []int {
	seen := map[int]bool{}
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	out := make([]int, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func unitCountsByPlayer(info *UnitInfo) map[int]int {
	out := map[int]int{}
	for _, player := range info.Sections {
		out[player.Player] = len(player.Units)
	}
	return out
}

func diffTriggerSummaries(before, after []TriggerSummary, add func(kind, field string, b, a any, detail string)) {
	limit := len(before)
	if len(after) < limit {
		limit = len(after)
	}
	for i := 0; i < limit; i++ {
		prefix := fmt.Sprintf("trigger_%d", i)
		if before[i].Name != after[i].Name {
			add("triggers", prefix+".name", before[i].Name, after[i].Name, "")
		}
		if before[i].Enabled != after[i].Enabled {
			add("triggers", prefix+".enabled", before[i].Enabled, after[i].Enabled, before[i].Name)
		}
		if before[i].Looping != after[i].Looping {
			add("triggers", prefix+".looping", before[i].Looping, after[i].Looping, before[i].Name)
		}
		if before[i].Effects != after[i].Effects {
			add("triggers", prefix+".effects", before[i].Effects, after[i].Effects, before[i].Name)
		}
		if before[i].Conditions != after[i].Conditions {
			add("triggers", prefix+".conditions", before[i].Conditions, after[i].Conditions, before[i].Name)
		}
	}
	for i := limit; i < len(before); i++ {
		add("triggers", fmt.Sprintf("trigger_%d", i), before[i].Name, nil, "removed")
	}
	for i := limit; i < len(after); i++ {
		add("triggers", fmt.Sprintf("trigger_%d", i), nil, after[i].Name, "added")
	}
}

func diffPlayers(before, after []PlayerInfo, add func(kind, field string, b, a any, detail string)) {
	limit := len(before)
	if len(after) < limit {
		limit = len(after)
	}
	for i := 0; i < limit; i++ {
		if before[i].Active != after[i].Active {
			add("players", fmt.Sprintf("player_%d.active", i), before[i].Active, after[i].Active, "")
		}
		if before[i].Human != after[i].Human {
			add("players", fmt.Sprintf("player_%d.human", i), before[i].Human, after[i].Human, "")
		}
		if before[i].TribeName != after[i].TribeName {
			add("players", fmt.Sprintf("player_%d.tribe_name", i), before[i].TribeName, after[i].TribeName, "")
		}
		if before[i].AIName != after[i].AIName {
			add("players", fmt.Sprintf("player_%d.ai_name", i), before[i].AIName, after[i].AIName, "")
		}
	}
	if len(before) != len(after) {
		add("players", "count", len(before), len(after), "")
	}
}

func diffAI(before, after []AIFileInfo, add func(kind, field string, b, a any, detail string)) {
	beforeByName := map[string]AIFileInfo{}
	afterByName := map[string]AIFileInfo{}
	for _, ai := range before {
		beforeByName[ai.Name] = ai
	}
	for _, ai := range after {
		afterByName[ai.Name] = ai
	}
	names := map[string]bool{}
	for name := range beforeByName {
		names[name] = true
	}
	for name := range afterByName {
		names[name] = true
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	for _, name := range sorted {
		b, bok := beforeByName[name]
		a, aok := afterByName[name]
		switch {
		case !bok:
			add("ai", name, nil, a.ContentSHA256, "added")
		case !aok:
			add("ai", name, b.ContentSHA256, nil, "removed")
		case b.ContentSHA256 != a.ContentSHA256 || b.ContentBytes != a.ContentBytes:
			add("ai", name+".content", b.ContentSHA256, a.ContentSHA256, "")
		}
	}
}
