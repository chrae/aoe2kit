package scenario

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"aoe2kit/pkg/triggergraph"
)

const defaultMaxInflatedScenarioBytes = 32 * 1024 * 1024
const maxScenarioFieldRepeat = 1_000_000

type File struct {
	Path            string        `json:"path,omitempty"`
	Version         string        `json:"version"`
	PlayerCount     int           `json:"player_count"`
	HeaderBytes     int           `json:"header_bytes"`
	CompressedBytes int           `json:"compressed_body_bytes"`
	InflatedBytes   int           `json:"inflated_body_bytes"`
	Sections        []SectionInfo `json:"sections,omitempty"`
	Triggers        *TriggerInfo  `json:"triggers,omitempty"`
	Units           *UnitInfo     `json:"units,omitempty"`
	Map             *MapInfo      `json:"map,omitempty"`
	Players         []PlayerInfo  `json:"players,omitempty"`
	AI              []AIFileInfo  `json:"ai,omitempty"`

	header     []byte
	body       []byte
	headerRoot *parsedNode
	root       *parsedRoot
}

type ParseOptions struct {
	MaxInflatedBytes int
}

type SectionInfo struct {
	Name  string `json:"name"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type TriggerInfo struct {
	Version       float64          `json:"version"`
	Count         int              `json:"count"`
	GraphSHA256   string           `json:"graph_sha256,omitempty"`
	DisplayOrder  []uint32         `json:"display_order,omitempty"`
	Variables     int              `json:"variables"`
	RecordStart   int              `json:"record_start"`
	RecordEnd     int              `json:"record_end"`
	Triggers      []TriggerSummary `json:"triggers,omitempty"`
	InvariantOK   bool             `json:"invariant_ok"`
	InvariantNote string           `json:"invariant_note,omitempty"`
}

type TriggerSummary struct {
	Index       int             `json:"index"`
	Name        string          `json:"name"`
	Enabled     uint32          `json:"enabled"`
	Looping     int8            `json:"looping"`
	Effects     int             `json:"effects"`
	EffectData  []EffectSummary `json:"effect_data,omitempty"`
	Conditions  int             `json:"conditions"`
	RecordStart int             `json:"record_start"`
	RecordEnd   int             `json:"record_end"`
}

type UnitInfo struct {
	Sections []PlayerUnitsInfo `json:"sections"`
	Total    int               `json:"total"`
}

type PlayerUnitsInfo struct {
	Player int           `json:"player"`
	Count  int           `json:"count"`
	Units  []UnitSummary `json:"units,omitempty"`
}

type UnitSummary struct {
	Index           int     `json:"index"`
	ReferenceID     int     `json:"reference_id"`
	UnitConst       int     `json:"unit_const"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	Z               float64 `json:"z"`
	Rotation        float64 `json:"rotation"`
	Status          int     `json:"status"`
	CaptionStringID int     `json:"caption_string_id,omitempty"`
	CaptionString   string  `json:"caption_string,omitempty"`
}

type MapInfo struct {
	Width         int         `json:"width"`
	Height        int         `json:"height"`
	TileCount     int         `json:"tile_count"`
	TerrainCounts map[int]int `json:"terrain_counts"`
}

type PlayerInfo struct {
	Player           int    `json:"player"`
	Active           bool   `json:"active"`
	Human            bool   `json:"human"`
	TribeName        string `json:"tribe_name,omitempty"`
	Civilization     string `json:"civilization,omitempty"`
	LockCivilization bool   `json:"lock_civilization"`
	LockPersonality  bool   `json:"lock_personality"`
	AIName           string `json:"ai_name,omitempty"`
	AIType           int    `json:"ai_type,omitempty"`
}

type AIFileInfo struct {
	Index         int    `json:"index"`
	Name          string `json:"name,omitempty"`
	ContentBytes  int    `json:"content_bytes,omitempty"`
	ContentSHA256 string `json:"content_sha256,omitempty"`
}

type DescribeOptions struct {
	Sections []string
	Full     bool
}

type Description struct {
	Path            string        `json:"path,omitempty"`
	Version         string        `json:"version"`
	PlayerCount     int           `json:"player_count"`
	HeaderBytes     int           `json:"header_bytes"`
	CompressedBytes int           `json:"compressed_body_bytes"`
	InflatedBytes   int           `json:"inflated_body_bytes"`
	SectionIndex    []SectionInfo `json:"section_index"`
	Players         []PlayerInfo  `json:"players,omitempty"`
	Triggers        *TriggerInfo  `json:"triggers,omitempty"`
	Units           *UnitInfo     `json:"units,omitempty"`
	Map             *MapInfo      `json:"map,omitempty"`
	AI              []AIFileInfo  `json:"ai,omitempty"`
	Sections        []SectionDump `json:"sections,omitempty"`
}

type SectionDump struct {
	Name   string     `json:"name"`
	Start  int        `json:"start"`
	End    int        `json:"end"`
	Fields []NodeDump `json:"fields,omitempty"`
}

type NodeDump struct {
	Name     string     `json:"name"`
	Start    int        `json:"start"`
	End      int        `json:"end"`
	Value    any        `json:"value,omitempty"`
	Fields   []NodeDump `json:"fields,omitempty"`
	Elements []NodeDump `json:"elements,omitempty"`
}

func Open(path string) (*File, error) {
	return OpenWithOptions(path, ParseOptions{})
}

func OpenWithOptions(path string, opts ParseOptions) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	file, err := ParseWithOptions(data, opts)
	if err != nil {
		return nil, err
	}
	file.Path = path
	return file, nil
}

func Parse(data []byte) (*File, error) {
	return ParseWithOptions(data, ParseOptions{})
}

func ParseWithOptions(data []byte, opts ParseOptions) (*File, error) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		return nil, err
	}
	if len(data) < 8 {
		return nil, errors.New("scenario too short")
	}
	fileHeaderSpec, ok := spec.section("FileHeader")
	if !ok {
		return nil, errors.New("embedded scenario spec has no FileHeader")
	}
	headerParser := parser{
		data:     data,
		sections: map[string]*parsedSection{},
	}
	headerNode, err := headerParser.parseNode("FileHeader", fileHeaderSpec, nil)
	if err != nil {
		return nil, fmt.Errorf("parse FileHeader: %w", err)
	}
	version, _ := headerNode.stringValue("version")
	playerCount, _ := headerNode.intValue("player_count")
	if !SupportsReadVersion(version) {
		return nil, fmt.Errorf("unsupported scenario version %q; AoE2Kit read support currently covers DE 1.55 through 1.58", version)
	}
	bodySpec, err := LoadDESpecForVersion(version)
	if err != nil {
		return nil, err
	}
	headerLen := headerParser.off
	header := append([]byte(nil), data[:headerLen]...)
	maxScenarioBytes := opts.MaxInflatedBytes
	if maxScenarioBytes == 0 {
		maxScenarioBytes = maxInflatedScenarioBytes()
	}
	body, err := inflateRawLimited(data[headerLen:], maxScenarioBytes)
	if err != nil {
		return nil, fmt.Errorf("inflate scenario body: %w", err)
	}
	parser := parser{
		data:     body,
		sections: map[string]*parsedSection{},
	}
	root, err := parser.parse(bodySpec)
	if err != nil {
		return nil, err
	}
	if parser.off != len(body) {
		return nil, fmt.Errorf("scenario body parse stopped at %d of %d", parser.off, len(body))
	}
	file := &File{
		Version:         version,
		PlayerCount:     playerCount,
		HeaderBytes:     len(header),
		CompressedBytes: len(data) - headerLen,
		InflatedBytes:   len(body),
		Sections:        root.sectionInfos(),
		header:          header,
		body:            body,
		headerRoot:      headerNode,
		root:            root,
	}
	if triggers, err := root.triggerInfo(); err == nil {
		file.Triggers = triggers
	} else {
		return nil, err
	}
	if units, err := root.unitInfo(); err == nil {
		file.Units = units
	} else {
		return nil, err
	}
	if mapInfo, err := root.mapInfo(); err == nil {
		file.Map = mapInfo
	} else {
		return nil, err
	}
	file.Players = root.playerInfo()
	file.AI = root.aiInfo()
	return file, nil
}

func maxInflatedScenarioBytes() int {
	if os.Getenv("AOE2KIT_ALLOW_HUGE_SCENARIO") == "1" {
		return 0
	}
	raw := strings.TrimSpace(os.Getenv("AOE2KIT_MAX_SCENARIO_MB"))
	if raw == "" {
		return defaultMaxInflatedScenarioBytes
	}
	mb, err := strconv.Atoi(raw)
	if err != nil || mb < 0 {
		return defaultMaxInflatedScenarioBytes
	}
	return mb * 1024 * 1024
}

func SupportsReadVersion(version string) bool {
	switch version {
	case "1.55", "1.56", "1.57", "1.58":
		return true
	default:
		return false
	}
}

func SupportsWriteVersion(version string) bool {
	switch version {
	case "1.57", "1.58":
		return true
	default:
		return false
	}
}

func (f *File) Describe(opts DescribeOptions) Description {
	desc := Description{
		Path:            f.Path,
		Version:         f.Version,
		PlayerCount:     f.PlayerCount,
		HeaderBytes:     f.HeaderBytes,
		CompressedBytes: f.CompressedBytes,
		InflatedBytes:   f.InflatedBytes,
		SectionIndex:    f.Sections,
		Players:         f.Players,
		Triggers:        f.Triggers,
		Units:           f.Units,
		Map:             f.Map,
		AI:              f.AI,
	}
	if f.root == nil {
		return desc
	}
	selected := sectionSelection(opts.Sections)
	if opts.Full || len(selected) > 0 {
		for _, section := range f.root.Sections {
			if len(selected) > 0 && !selected[section.Name] {
				continue
			}
			desc.Sections = append(desc.Sections, SectionDump{
				Name:   section.Name,
				Start:  section.Start,
				End:    section.End,
				Fields: dumpNodes(section.Fields),
			})
		}
	}
	return desc
}

func (f *File) RebuildBody() ([]byte, error) {
	if f.root == nil {
		return nil, errors.New("scenario has no parsed root")
	}
	return f.root.raw(), nil
}

func (f *File) VerifyRebuild() error {
	rebuilt, err := f.RebuildBody()
	if err != nil {
		return err
	}
	if !bytes.Equal(rebuilt, f.body) {
		return fmt.Errorf("rebuilt body differs: got %d bytes want %d", len(rebuilt), len(f.body))
	}
	compressed, err := DeflateRaw(rebuilt)
	if err != nil {
		return err
	}
	roundTrip, err := InflateRaw(compressed)
	if err != nil {
		return err
	}
	if !bytes.Equal(roundTrip, rebuilt) {
		return errors.New("rebuilt body changed after deflate/inflate")
	}
	return nil
}

func InflateRaw(compressed []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(compressed))
	defer r.Close()
	return io.ReadAll(r)
}

func inflateRawLimited(compressed []byte, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		return InflateRaw(compressed)
	}
	r := flate.NewReader(bytes.NewReader(compressed))
	defer r.Close()
	body, err := io.ReadAll(io.LimitReader(r, int64(maxBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBytes {
		return nil, fmt.Errorf("inflated scenario body is above safety limit %d bytes; set AOE2KIT_MAX_SCENARIO_MB higher or AOE2KIT_ALLOW_HUGE_SCENARIO=1 for an intentional full parse", maxBytes)
	}
	return body, nil
}

func DeflateRaw(payload []byte) ([]byte, error) {
	var out bytes.Buffer
	w, err := flate.NewWriter(&out, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(payload); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

type parsedRoot struct {
	Sections []*parsedSection
}

func (r *parsedRoot) raw() []byte {
	var out []byte
	for _, section := range r.Sections {
		out = append(out, section.raw()...)
	}
	return out
}

func (r *parsedRoot) sectionInfos() []SectionInfo {
	out := make([]SectionInfo, 0, len(r.Sections))
	for _, section := range r.Sections {
		out = append(out, SectionInfo{Name: section.Name, Start: section.Start, End: section.End})
	}
	return out
}

func (r *parsedRoot) section(name string) *parsedSection {
	for _, section := range r.Sections {
		if section.Name == name {
			return section
		}
	}
	return nil
}

func (r *parsedRoot) triggerInfo() (*TriggerInfo, error) {
	section := r.section("Triggers")
	if section == nil {
		return nil, errors.New("missing Triggers section")
	}
	version, _ := section.float64("trigger_version")
	count, _ := section.intValue("number_of_triggers")
	display := section.uint32List("trigger_display_order_array")
	variableCount, _ := section.intValue("number_of_variables")
	triggerNodes := section.list("trigger_data")
	graph, graphErr := triggerGraphFromTypedNodes(triggerNodes, triggerDataStart(section), triggerDataEnd(section))
	info := &TriggerInfo{
		Version:      version,
		Count:        count,
		DisplayOrder: display,
		Variables:    variableCount,
		RecordStart:  section.Start,
		RecordEnd:    section.End,
	}
	if graphErr == nil {
		info.GraphSHA256 = graph.SHA256
	}
	for i, trigger := range triggerNodes {
		name, _ := trigger.stringValue("trigger_name")
		enabled, _ := trigger.uint32Value("enabled")
		looping, _ := trigger.int8Value("looping")
		effects := trigger.list("effect_data")
		conditions := trigger.list("condition_data")
		effectData := make([]EffectSummary, 0, len(effects))
		for j, effect := range effects {
			summary := summarizeEffect(effect, EffectsOptions{})
			summary.TriggerIndex = i
			summary.TriggerName = name
			summary.EffectIndex = j
			effectData = append(effectData, summary)
		}
		info.Triggers = append(info.Triggers, TriggerSummary{
			Index:       i,
			Name:        name,
			Enabled:     enabled,
			Looping:     looping,
			Effects:     len(effects),
			EffectData:  effectData,
			Conditions:  len(conditions),
			RecordStart: trigger.Start,
			RecordEnd:   trigger.End,
		})
	}
	info.InvariantOK, info.InvariantNote = validateTriggerInfo(info, triggerNodes)
	if graphErr != nil {
		info.InvariantOK = false
		if info.InvariantNote != "" {
			info.InvariantNote += "; "
		}
		info.InvariantNote += "trigger graph fingerprint failed: " + graphErr.Error()
	}
	return info, nil
}

func triggerDataStart(section *parsedSection) int {
	if field := section.field("trigger_data"); field != nil {
		return field.Start
	}
	return section.Start
}

func triggerDataEnd(section *parsedSection) int {
	if field := section.field("trigger_data"); field != nil {
		return field.End
	}
	return section.End
}

func triggerGraphFromTypedNodes(triggerNodes []*parsedNode, start, end int) (*triggergraph.Graph, error) {
	triggers := make([]map[string]any, 0, len(triggerNodes))
	for _, trigger := range triggerNodes {
		canonical, err := canonicalTrigger(trigger)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, canonical)
	}
	return triggergraph.FromTriggers(triggers, start, end, nil)
}

func (f *File) TriggerGraph() (*triggergraph.Graph, error) {
	if f.root == nil {
		return nil, errors.New("scenario body not parsed")
	}
	section := f.root.section("Triggers")
	if section == nil {
		return nil, errors.New("missing Triggers section")
	}
	return triggerGraphFromTypedNodes(section.list("trigger_data"), triggerDataStart(section), triggerDataEnd(section))
}

func canonicalTrigger(trigger *parsedNode) (map[string]any, error) {
	effects := trigger.list("effect_data")
	conditions := trigger.list("condition_data")
	canonicalEffects := make([]map[string]any, 0, len(effects))
	for _, effect := range effects {
		canonical, err := canonicalEffect(effect)
		if err != nil {
			return nil, err
		}
		canonicalEffects = append(canonicalEffects, canonical)
	}
	canonicalConditions := make([]map[string]any, 0, len(conditions))
	for _, condition := range conditions {
		canonical, err := canonicalCondition(condition)
		if err != nil {
			return nil, err
		}
		canonicalConditions = append(canonicalConditions, canonical)
	}
	enabled, _ := trigger.uint32Value("enabled")
	looping, _ := trigger.intValue("looping")
	executeOnLoad, _ := trigger.intValue("execute_on_load")
	descriptionSTID, _ := trigger.intValue("description_string_table_id")
	displayAsObjective, _ := trigger.intValue("display_as_objective")
	descriptionOrder, _ := trigger.uint32Value("objective_description_order")
	makeHeader, _ := trigger.intValue("make_header")
	shortDescriptionSTID, _ := trigger.intValue("short_description_string_table_id")
	displayOnScreen, _ := trigger.intValue("display_on_screen")
	muteObjectives, _ := trigger.intValue("mute_objectives")
	description, _ := trigger.stringValue("trigger_description")
	name, _ := trigger.stringValue("trigger_name")
	shortDescription, _ := trigger.stringValue("short_description")
	return map[string]any{
		"conditions":             canonicalConditions,
		"condition_order":        trigger.intList("condition_display_order_array"),
		"description":            description,
		"description_order":      int(descriptionOrder),
		"description_stid":       descriptionSTID,
		"display_as_objective":   displayAsObjective,
		"display_on_screen":      displayOnScreen,
		"effect_order":           trigger.intList("effect_display_order_array"),
		"effects":                canonicalEffects,
		"enabled":                int(enabled),
		"execute_on_load":        executeOnLoad,
		"looping":                looping,
		"make_header":            makeHeader,
		"mute_objectives":        muteObjectives,
		"name":                   name,
		"short_description":      shortDescription,
		"short_description_stid": shortDescriptionSTID,
	}, nil
}

func canonicalEffect(effect *parsedNode) (map[string]any, error) {
	raw := effect.raw()
	if len(raw) < 84*4 {
		return nil, fmt.Errorf("effect at %d has %d bytes, need at least %d", effect.Start, len(raw), 84*4)
	}
	values := make([]int, 84)
	for i := range values {
		values[i] = int(int32(binary.LittleEndian.Uint32(raw[i*4:])))
	}
	if values[1] != 83 && values[1] != 81 {
		return nil, fmt.Errorf("bad effect static value at %d: %d", effect.Start, values[1])
	}
	message, _ := effect.stringValue("message")
	sum := sha256.Sum256(raw)
	return map[string]any{
		"fields_prefix": values,
		"message":       message,
		"raw_sha256":    hex.EncodeToString(sum[:]),
		"type":          values[0],
	}, nil
}

func canonicalCondition(condition *parsedNode) (map[string]any, error) {
	raw := condition.raw()
	if len(raw) < 35*4 {
		return nil, fmt.Errorf("condition at %d has %d bytes, need at least %d", condition.Start, len(raw), 35*4)
	}
	values := make([]int, 35)
	for i := range values {
		values[i] = int(int32(binary.LittleEndian.Uint32(raw[i*4:])))
	}
	if values[1] != 33 {
		return nil, fmt.Errorf("bad condition static value at %d: %d", condition.Start, values[1])
	}
	xsFunction, _ := condition.stringValue("xs_function")
	return map[string]any{
		"fields":      values,
		"type":        values[0],
		"xs_function": xsFunction,
	}, nil
}

func (r *parsedRoot) unitInfo() (*UnitInfo, error) {
	section := r.section("Units")
	if section == nil {
		return nil, errors.New("missing Units section")
	}
	playerSections := section.list("players_units")
	info := &UnitInfo{}
	for i, playerSection := range playerSections {
		count, _ := playerSection.intValue("unit_count")
		unitNodes := playerSection.list("units")
		playerInfo := PlayerUnitsInfo{Player: i, Count: count}
		for j, unit := range unitNodes {
			refID, _ := unit.intValue("reference_id")
			unitConst, _ := unit.intValue("unit_const")
			status, _ := unit.intValue("status")
			captionStringID, _ := unit.intValue("caption_string_id")
			captionString, _ := unit.stringValue("caption_string")
			x, _ := unit.floatValue("x")
			y, _ := unit.floatValue("y")
			z, _ := unit.floatValue("z")
			rotation, _ := unit.floatValue("rotation")
			playerInfo.Units = append(playerInfo.Units, UnitSummary{
				Index:           j,
				ReferenceID:     refID,
				UnitConst:       unitConst,
				X:               x,
				Y:               y,
				Z:               z,
				Rotation:        rotation,
				Status:          status,
				CaptionStringID: captionStringID,
				CaptionString:   captionString,
			})
		}
		info.Sections = append(info.Sections, playerInfo)
		info.Total += len(unitNodes)
	}
	return info, nil
}

func (r *parsedRoot) mapInfo() (*MapInfo, error) {
	section := r.section("Map")
	if section == nil {
		return nil, errors.New("missing Map section")
	}
	width, _ := section.intValue("map_width")
	height, _ := section.intValue("map_height")
	tiles := section.list("terrain_data")
	info := &MapInfo{
		Width:         width,
		Height:        height,
		TileCount:     len(tiles),
		TerrainCounts: map[int]int{},
	}
	for _, tile := range tiles {
		terrainID, _ := tile.intValue("terrain_id")
		info.TerrainCounts[terrainID]++
	}
	return info, nil
}

func (r *parsedRoot) playerInfo() []PlayerInfo {
	dataHeader := r.section("DataHeader")
	playerDataTwo := r.section("PlayerDataTwo")
	if dataHeader == nil {
		return nil
	}
	tribeNames := dataHeader.stringList("tribe_names")
	lockCiv := dataHeader.intList("per_player_lock_civilization")
	lockPersonality := dataHeader.intList("per_player_lock_personality")
	playerData := dataHeader.list("player_data_1")
	var aiNames []string
	var aiTypes []int
	if playerDataTwo != nil {
		aiNames = playerDataTwo.stringList("ai_names")
		aiTypes = playerDataTwo.intList("ai_type")
	}
	out := make([]PlayerInfo, 0, len(playerData))
	for i, player := range playerData {
		active, _ := player.intValue("active")
		human, _ := player.intValue("human")
		civ, _ := player.stringValue("civilization")
		info := PlayerInfo{
			Player:           i,
			Active:           active != 0,
			Human:            human != 0,
			TribeName:        stringAt(tribeNames, i),
			Civilization:     civ,
			LockCivilization: intAt(lockCiv, i) != 0,
			LockPersonality:  intAt(lockPersonality, i) != 0,
			AIName:           stringAt(aiNames, i),
			AIType:           intAt(aiTypes, i),
		}
		out = append(out, info)
	}
	return out
}

func (r *parsedRoot) aiInfo() []AIFileInfo {
	files := r.section("Files")
	if files == nil {
		return nil
	}
	aiFiles := files.list("ai_files")
	out := make([]AIFileInfo, 0, len(aiFiles))
	for i, ai := range aiFiles {
		name, _ := ai.stringValue("ai_file_name")
		content, _ := ai.stringValue("ai_file")
		sum := sha256.Sum256([]byte(content))
		info := AIFileInfo{Index: i, Name: name, ContentBytes: len(content)}
		if content != "" {
			info.ContentSHA256 = hex.EncodeToString(sum[:])
		}
		out = append(out, info)
	}
	return out
}

func validateTriggerInfo(info *TriggerInfo, triggers []*parsedNode) (bool, string) {
	if info.Count != len(triggers) {
		return false, fmt.Sprintf("number_of_triggers=%d but trigger_data has %d", info.Count, len(triggers))
	}
	if info.Count != len(info.DisplayOrder) {
		return false, fmt.Sprintf("number_of_triggers=%d but display_order has %d", info.Count, len(info.DisplayOrder))
	}
	seen := map[uint32]bool{}
	for _, id := range info.DisplayOrder {
		if int(id) >= info.Count {
			return false, fmt.Sprintf("display_order contains out-of-range trigger id %d", id)
		}
		if seen[id] {
			return false, fmt.Sprintf("display_order contains duplicate trigger id %d", id)
		}
		seen[id] = true
	}
	for i, trigger := range triggers {
		effectCount, _ := trigger.intValue("number_of_effects")
		effects := trigger.list("effect_data")
		if effectCount != len(effects) {
			return false, fmt.Sprintf("trigger %d number_of_effects=%d but effect_data has %d", i, effectCount, len(effects))
		}
		effectOrder := trigger.intList("effect_display_order_array")
		if ok, note := validateOptionalOrder(effectOrder, len(effects), "effect", i); !ok {
			return false, note
		}
		conditionCount, _ := trigger.intValue("number_of_conditions")
		conditions := trigger.list("condition_data")
		if conditionCount != len(conditions) {
			return false, fmt.Sprintf("trigger %d number_of_conditions=%d but condition_data has %d", i, conditionCount, len(conditions))
		}
		conditionOrder := trigger.intList("condition_display_order_array")
		if ok, note := validateOptionalOrder(conditionOrder, len(conditions), "condition", i); !ok {
			return false, note
		}
		for j, effect := range effects {
			selectedCount, _ := effect.intValue("number_of_units_selected")
			selected := effect.intList("selected_object_ids")
			if selectedCount == -1 && len(selected) == 0 {
				continue
			}
			if selectedCount != len(selected) {
				return false, fmt.Sprintf("trigger %d effect %d selected count=%d but ids has %d", i, j, selectedCount, len(selected))
			}
		}
	}
	return true, ""
}

func validateOptionalOrder(order []int, count int, label string, triggerIndex int) (bool, string) {
	if len(order) == 0 {
		return true, ""
	}
	if len(order) != count {
		return false, fmt.Sprintf("trigger %d %s display order has %d entries for %d %ss", triggerIndex, label, len(order), count, label)
	}
	seen := map[int]bool{}
	for _, id := range order {
		if id < 0 || id >= count {
			return false, fmt.Sprintf("trigger %d %s display order contains out-of-range id %d", triggerIndex, label, id)
		}
		if seen[id] {
			return false, fmt.Sprintf("trigger %d %s display order contains duplicate id %d", triggerIndex, label, id)
		}
		seen[id] = true
	}
	return true, ""
}

type parsedSection = parsedNode

type parsedNode struct {
	Name     string
	Start    int
	End      int
	Fields   []*parsedNode
	Elements []*parsedNode
	Raw      []byte
	Value    any
}

func (n *parsedNode) raw() []byte {
	if len(n.Raw) != 0 {
		return append([]byte(nil), n.Raw...)
	}
	var out []byte
	for _, field := range n.Fields {
		out = append(out, field.raw()...)
	}
	for _, elem := range n.Elements {
		out = append(out, elem.raw()...)
	}
	return out
}

func (n *parsedNode) field(name string) *parsedNode {
	for _, field := range n.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (n *parsedNode) list(name string) []*parsedNode {
	field := n.field(name)
	if field == nil {
		return nil
	}
	return field.Elements
}

func (n *parsedNode) stringValue(name string) (string, bool) {
	field := n.field(name)
	if field == nil {
		return "", false
	}
	value, ok := field.Value.(string)
	return value, ok
}

func (n *parsedNode) stringList(name string) []string {
	field := n.field(name)
	if field == nil {
		return nil
	}
	values, ok := field.Value.([]any)
	if !ok {
		if one, ok := field.Value.(string); ok {
			return []string{one}
		}
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if v, ok := value.(string); ok {
			out = append(out, v)
		}
	}
	return out
}

func (n *parsedNode) float64(name string) (float64, bool) {
	field := n.field(name)
	if field == nil {
		return 0, false
	}
	value, ok := field.Value.(float64)
	return value, ok
}

func (n *parsedNode) floatValue(name string) (float64, bool) {
	field := n.field(name)
	if field == nil {
		return 0, false
	}
	switch value := field.Value.(type) {
	case float32:
		return float64(value), true
	case float64:
		return value, true
	}
	return 0, false
}

func (n *parsedNode) uint32Value(name string) (uint32, bool) {
	field := n.field(name)
	if field == nil {
		return 0, false
	}
	value, ok := field.Value.(uint32)
	return value, ok
}

func (n *parsedNode) int8Value(name string) (int8, bool) {
	field := n.field(name)
	if field == nil {
		return 0, false
	}
	value, ok := field.Value.(int8)
	return value, ok
}

func (n *parsedNode) intValue(name string) (int, bool) {
	field := n.field(name)
	if field == nil {
		return 0, false
	}
	switch value := field.Value.(type) {
	case int:
		return value, true
	case int8:
		return int(value), true
	case int16:
		return int(value), true
	case int32:
		return int(value), true
	case uint8:
		return int(value), true
	case uint16:
		return int(value), true
	case uint32:
		return int(value), true
	}
	return 0, false
}

func (n *parsedNode) intList(name string) []int {
	field := n.field(name)
	if field == nil {
		return nil
	}
	values, ok := field.Value.([]any)
	if !ok {
		if one, ok := numericListItem(field.Value); ok {
			return []int{one}
		}
		return nil
	}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if v, ok := numericListItem(value); ok {
			out = append(out, v)
		}
	}
	return out
}

func numericListItem(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	}
	return 0, false
}

func sectionSelection(sections []string) map[string]bool {
	if len(sections) == 0 {
		return nil
	}
	selected := map[string]bool{}
	for _, section := range sections {
		for _, name := range strings.Split(section, ",") {
			name = strings.TrimSpace(strings.ToLower(name))
			if name == "" || name == "all" {
				return nil
			}
			for _, canonical := range sectionAliases(name) {
				selected[canonical] = true
			}
		}
	}
	return selected
}

func sectionAliases(name string) []string {
	switch name {
	case "triggers", "trigger":
		return []string{"Triggers"}
	case "units", "unit":
		return []string{"Units"}
	case "map", "terrain":
		return []string{"Map"}
	case "players", "player", "resources":
		return []string{"DataHeader", "PlayerDataTwo", "Diplomacy"}
	case "victory":
		return []string{"GlobalVictory"}
	case "messages", "message":
		return []string{"Messages", "Cinematics"}
	case "disables", "disabled", "options":
		return []string{"Options"}
	case "ai", "files":
		return []string{"Files", "PlayerDataTwo"}
	default:
		return []string{name}
	}
}

func dumpNodes(nodes []*parsedNode) []NodeDump {
	out := make([]NodeDump, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, dumpNode(node))
	}
	return out
}

func dumpNode(node *parsedNode) NodeDump {
	out := NodeDump{
		Name:  node.Name,
		Start: node.Start,
		End:   node.End,
	}
	if node.Value != nil && len(node.Elements) == 0 {
		out.Value = dumpValue(node.Value)
	}
	if len(node.Fields) > 0 {
		out.Fields = dumpNodes(node.Fields)
	}
	if len(node.Elements) > 0 {
		out.Elements = dumpNodes(node.Elements)
	}
	return out
}

func dumpValue(value any) any {
	switch v := value.(type) {
	case []byte:
		return hex.EncodeToString(v)
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, dumpValue(item))
		}
		return out
	case float32:
		if math.IsNaN(float64(v)) {
			return "NaN"
		}
		if math.IsInf(float64(v), 1) {
			return "+Inf"
		}
		if math.IsInf(float64(v), -1) {
			return "-Inf"
		}
		return v
	case float64:
		if math.IsNaN(v) {
			return "NaN"
		}
		if math.IsInf(v, 1) {
			return "+Inf"
		}
		if math.IsInf(v, -1) {
			return "-Inf"
		}
		return v
	default:
		return v
	}
}

func stringAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}

func intAt(values []int, index int) int {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}

func (n *parsedNode) uint32List(name string) []uint32 {
	field := n.field(name)
	if field == nil {
		return nil
	}
	values, ok := field.Value.([]any)
	if !ok {
		if one, ok := field.Value.(uint32); ok {
			return []uint32{one}
		}
		return nil
	}
	out := make([]uint32, 0, len(values))
	for _, value := range values {
		if v, ok := value.(uint32); ok {
			out = append(out, v)
		}
	}
	return out
}

type parser struct {
	data     []byte
	off      int
	sections map[string]*parsedSection
}

func (p *parser) parse(spec *Spec) (*parsedRoot, error) {
	root := &parsedRoot{}
	for _, sectionSpec := range spec.Sections {
		if sectionSpec.Name == "FileHeader" {
			continue
		}
		section, err := p.parseNode(sectionSpec.Name, sectionSpec.Spec, nil)
		if err != nil {
			return nil, fmt.Errorf("section %s: %w", sectionSpec.Name, err)
		}
		root.Sections = append(root.Sections, section)
		p.sections[section.Name] = section
	}
	return root, nil
}

func (p *parser) parseNode(name string, spec SectionSpec, parent *parsedNode) (*parsedNode, error) {
	node := &parsedNode{Name: name, Start: p.off}
	for _, retriever := range spec.Retrievers {
		field, err := p.parseRetriever(retriever, spec, node)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", retriever.Name, err)
		}
		node.Fields = append(node.Fields, field)
	}
	node.End = p.off
	return node, nil
}

func (p *parser) parseRetriever(retriever RetrieverSpec, sectionSpec SectionSpec, node *parsedNode) (*parsedNode, error) {
	repeat, err := p.constructRepeat(retriever, node)
	if err != nil {
		return nil, err
	}
	if retriever.Name == "__END_OF_FILE_MARK__" && p.off == len(p.data) {
		repeat = 0
	}
	if repeat < 0 {
		repeat = 0
	}
	if err := p.validateRetrieverRepeat(retriever, repeat); err != nil {
		return nil, err
	}
	field := &parsedNode{Name: retriever.Name, Start: p.off}
	if strings.HasPrefix(retriever.Type, "struct:") {
		structName := strings.TrimPrefix(retriever.Type, "struct:")
		structSpec, ok := sectionSpec.Structs[structName]
		if !ok {
			return nil, fmt.Errorf("missing struct spec %s", structName)
		}
		for i := 0; i < repeat; i++ {
			elem, err := p.parseNode(structName, structSpec, field)
			if err != nil {
				return nil, fmt.Errorf("%s[%d]: %w", structName, i, err)
			}
			field.Elements = append(field.Elements, elem)
		}
		field.Value = field.Elements
		field.End = p.off
		return field, nil
	}
	values := make([]any, 0, repeat)
	var raw []byte
	for i := 0; i < repeat; i++ {
		value, chunk, err := p.readPrimitive(retriever.Type)
		if err != nil {
			return nil, fmt.Errorf("item %d: %w", i, err)
		}
		values = append(values, value)
		raw = append(raw, chunk...)
	}
	field.Raw = raw
	field.End = p.off
	if retriever.IsList != nil && !*retriever.IsList && len(values) > 0 {
		field.Value = values[0]
	} else if repeat == 1 && retriever.IsList == nil {
		field.Value = values[0]
	} else {
		field.Value = values
	}
	return field, nil
}

func (p *parser) validateRetrieverRepeat(retriever RetrieverSpec, repeat int) error {
	if repeat > maxScenarioFieldRepeat {
		return fmt.Errorf("%s repeat %d exceeds parser safety limit %d", retriever.Name, repeat, maxScenarioFieldRepeat)
	}
	if repeat == 0 || strings.HasPrefix(retriever.Type, "struct:") {
		return nil
	}
	typ, size, err := primitiveType(retriever.Type)
	if err != nil {
		return err
	}
	if typ == "str" {
		return nil
	}
	remaining := len(p.data) - p.off
	if size > 0 && repeat > remaining/size {
		return fmt.Errorf("%s repeat %d of %s exceeds remaining bytes %d", retriever.Name, repeat, retriever.Type, remaining)
	}
	return nil
}

func (p *parser) constructRepeat(retriever RetrieverSpec, node *parsedNode) (int, error) {
	repeat := retriever.Repeat
	for _, dep := range retriever.Dependencies["on_construct"] {
		switch dep.Action {
		case "REFRESH_SELF":
			value, err := p.evalRefreshRepeat(retriever, node)
			if err != nil {
				return 0, err
			}
			repeat = value
		case "SET_REPEAT":
			value, err := p.evalDependency(dep, node)
			if err != nil {
				return 0, err
			}
			repeat = value
		}
	}
	return repeat, nil
}

func (p *parser) evalRefreshRepeat(retriever RetrieverSpec, node *parsedNode) (int, error) {
	for _, dep := range retriever.Dependencies["on_refresh"] {
		if dep.Action != "SET_REPEAT" {
			continue
		}
		return p.evalDependency(dep, node)
	}
	return retriever.Repeat, nil
}

func (p *parser) evalDependency(dep Dependency, node *parsedNode) (int, error) {
	values := map[string]any{}
	for _, target := range dep.Target {
		value, err := p.targetValue(target, node)
		if err != nil {
			return 0, err
		}
		values[target.Name] = value
	}
	if dep.Eval == "" {
		if len(dep.Target) != 1 {
			return 0, fmt.Errorf("dependency without eval has %d targets", len(dep.Target))
		}
		return numericRepeat(values[dep.Target[0].Name])
	}
	return evalInt(dep.Eval, values)
}

func (p *parser) targetValue(target Target, node *parsedNode) (any, error) {
	var owner *parsedNode
	if target.Section == "self" {
		owner = node
	} else {
		owner = p.sections[target.Section]
	}
	if owner == nil {
		return nil, fmt.Errorf("unknown dependency target section %s", target.Section)
	}
	field := owner.field(target.Name)
	if field == nil {
		return nil, fmt.Errorf("unknown dependency target %s:%s", target.Section, target.Name)
	}
	return field.Value, nil
}

func evalInt(expr string, values map[string]any) (int, error) {
	switch expr {
	case "1 if bitmap_width * bitmap_height else 0":
		return boolInt(asInt(values["bitmap_width"])*asInt(values["bitmap_height"]) != 0), nil
	case "width * height":
		return asInt(values["width"]) * asInt(values["height"]), nil
	case "map_width * map_height":
		return asInt(values["map_width"]) * asInt(values["map_height"]), nil
	case "7 if victory_version >= 2 else 0":
		return boolChoice(asFloat(values["victory_version"]) >= 2, 7, 0), nil
	case "1 if round(victory_version, 2) >= 2 else 0":
		return boolInt(math.Round(asFloat(values["victory_version"])*100)/100 >= 2), nil
	case "1 if trigger_version >= 3.5 else 0":
		return boolInt(asFloat(values["trigger_version"]) >= 3.5), nil
	case "1 if trigger_version >= 4.5 else 0":
		return boolInt(asFloat(values["trigger_version"]) >= 4.5), nil
	case "0 if number_of_ai_files == 0 else 1":
		return boolInt(asInt(values["number_of_ai_files"]) != 0), nil
	case "number_of_ai_files if number_of_ai_files != [] else 0":
		return asInt(values["number_of_ai_files"]), nil
	case "0 if type(unknown_5) is list else unknown_5":
		if _, ok := values["unknown_5"].([]any); ok {
			return 0, nil
		}
		return asInt(values["unknown_5"]), nil
	}
	if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			return asInt(values[strings.TrimSpace(parts[0])]) * asInt(values[strings.TrimSpace(parts[1])]), nil
		}
	}
	if strings.HasPrefix(expr, "len(") && strings.HasSuffix(expr, ")") {
		name := strings.TrimSuffix(strings.TrimPrefix(expr, "len("), ")")
		return valueLen(values[name]), nil
	}
	if strings.Contains(expr, "[") && strings.HasSuffix(expr, "]") {
		base := expr[:strings.Index(expr, "[")]
		idxText := strings.TrimSuffix(expr[strings.Index(expr, "[")+1:], "]")
		var idx int
		if _, err := fmt.Sscanf(idxText, "%d", &idx); err == nil {
			return listIntAt(values[base], idx), nil
		}
	}
	if value, ok := values[expr]; ok {
		return numericRepeat(value)
	}
	return 0, fmt.Errorf("unsupported dependency eval %q", expr)
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func boolChoice(v bool, yes, no int) int {
	if v {
		return yes
	}
	return no
}

func numericRepeat(value any) (int, error) {
	return asInt(value), nil
}

func asInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case float64:
		return int(v)
	case []any:
		if len(v) == 0 {
			return 0
		}
		return asInt(v[0])
	}
	return 0
}

func asFloat(value any) float64 {
	switch v := value.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case []any:
		if len(v) == 0 {
			return 0
		}
		return asFloat(v[0])
	}
	return float64(asInt(value))
}

func valueLen(value any) int {
	switch v := value.(type) {
	case []any:
		return len(v)
	case []*parsedNode:
		return len(v)
	case string:
		return len(v)
	}
	return 0
}

func listIntAt(value any, idx int) int {
	switch v := value.(type) {
	case []any:
		if idx >= 0 && idx < len(v) {
			return asInt(v[idx])
		}
	}
	return 0
}

func (p *parser) readPrimitive(kind string) (any, []byte, error) {
	typ, size, err := primitiveType(kind)
	if err != nil {
		return nil, nil, err
	}
	if typ == "str" {
		prefix, err := p.bytes(size)
		if err != nil {
			return nil, nil, err
		}
		n := signedInt(prefix)
		if n < 0 {
			return nil, nil, fmt.Errorf("negative string length %d", n)
		}
		data, err := p.bytes(n)
		if err != nil {
			return nil, nil, err
		}
		raw := append(append([]byte(nil), prefix...), data...)
		return trimAtNUL(string(data)), raw, nil
	}
	raw, err := p.bytes(size)
	if err != nil {
		return nil, nil, err
	}
	switch typ {
	case "data":
		return append([]byte(nil), raw...), raw, nil
	case "c":
		return trimAtNUL(string(raw)), raw, nil
	case "s":
		switch size {
		case 1:
			return int8(raw[0]), raw, nil
		case 2:
			return int16(binary.LittleEndian.Uint16(raw)), raw, nil
		case 4:
			return int32(binary.LittleEndian.Uint32(raw)), raw, nil
		}
	case "u":
		switch size {
		case 1:
			return uint8(raw[0]), raw, nil
		case 2:
			return binary.LittleEndian.Uint16(raw), raw, nil
		case 4:
			return binary.LittleEndian.Uint32(raw), raw, nil
		}
	case "f":
		switch size {
		case 4:
			return math.Float32frombits(binary.LittleEndian.Uint32(raw)), raw, nil
		case 8:
			return math.Float64frombits(binary.LittleEndian.Uint64(raw)), raw, nil
		}
	}
	return nil, nil, fmt.Errorf("unsupported primitive %s", kind)
}

func primitiveType(kind string) (string, int, error) {
	if kind == "" {
		return "data", 0, nil
	}
	i := 0
	for i < len(kind) && (kind[i] < '0' || kind[i] > '9') {
		i++
	}
	prefix := kind[:i]
	num := kind[i:]
	if prefix == "" {
		prefix = "data"
	}
	var bits int
	if _, err := fmt.Sscanf(num, "%d", &bits); err != nil {
		return "", 0, fmt.Errorf("bad primitive kind %q", kind)
	}
	if prefix == "c" || prefix == "data" {
		return prefix, bits, nil
	}
	return prefix, bits / 8, nil
}

func signedInt(raw []byte) int {
	switch len(raw) {
	case 2:
		return int(int16(binary.LittleEndian.Uint16(raw)))
	case 4:
		return int(int32(binary.LittleEndian.Uint32(raw)))
	}
	return 0
}

func trimAtNUL(value string) string {
	if idx := strings.IndexByte(value, 0); idx >= 0 {
		return value[:idx]
	}
	return value
}

func (p *parser) bytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("negative byte count %d", n)
	}
	if p.off+n > len(p.data) {
		return nil, fmt.Errorf("need %d bytes at %d, only %d remain", n, p.off, len(p.data)-p.off)
	}
	out := p.data[p.off : p.off+n]
	p.off += n
	return out, nil
}
