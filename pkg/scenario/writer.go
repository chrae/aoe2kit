package scenario

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"aoe2kit/pkg/aoe2"
)

type Recipe struct {
	Scenario         *ScenarioRecipe         `json:"scenario,omitempty"`
	XS               *XSRecipe               `json:"xs,omitempty"`
	Victory          *VictoryRecipe          `json:"victory,omitempty"`
	Players          []PlayerRecipe          `json:"players,omitempty"`
	Diplomacy        []DiplomacyRecipe       `json:"diplomacy,omitempty"`
	DiplomacyOptions *DiplomacyOptionsRecipe `json:"diplomacy_options,omitempty"`
	Resources        []ResourceRecipe        `json:"resources,omitempty"`
	Strings          []StringRecipe          `json:"strings,omitempty"`
	Variables        []VariableRecipe        `json:"variables,omitempty"`
	Triggers         []TriggerRecipe         `json:"triggers"`
	Units            []UnitRecipe            `json:"units,omitempty"`
	Map              []MapRecipe             `json:"map,omitempty"`
}

type ScenarioRecipe struct {
	PlayerCount         *int `json:"player_count,omitempty"`
	TimestampOfLastSave *int `json:"timestamp_of_last_save,omitempty"`
}

const (
	MinScenarioMapSize = 80
	MaxScenarioMapSize = 480
)

type ScenarioMapPreset struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	StringID  int    `json:"string_id"`
	StringKey string `json:"string_key"`
}

// scenarioMapPresets is grounded in the shipped DE string table keys
// MAPSIZE_* at 25080..25480. The low three digits encode the editor map edge.
var scenarioMapPresets = []ScenarioMapPreset{
	{Name: "Mini", Size: 80, StringID: 25080, StringKey: "MAPSIZE_MINI"},
	{Name: "Tiny", Size: 120, StringID: 25120, StringKey: "MAPSIZE_TINY"},
	{Name: "Small", Size: 144, StringID: 25144, StringKey: "MAPSIZE_SMALL"},
	{Name: "Medium", Size: 168, StringID: 25168, StringKey: "MAPSIZE_MEDIUM"},
	{Name: "Normal", Size: 200, StringID: 25200, StringKey: "MAPSIZE_NORMAL"},
	{Name: "Large", Size: 220, StringID: 25220, StringKey: "MAPSIZE_LARGE"},
	{Name: "Huge", Size: 240, StringID: 25240, StringKey: "MAPSIZE_HUGE"},
	{Name: "Giant", Size: 252, StringID: 25252, StringKey: "MAPSIZE_GIANT"},
	{Name: "Massive", Size: 276, StringID: 25276, StringKey: "MAPSIZE_MASSIVE"},
	{Name: "Enormous", Size: 300, StringID: 25300, StringKey: "MAPSIZE_ENORMOUS"},
	{Name: "Colossal", Size: 320, StringID: 25320, StringKey: "MAPSIZE_COLOSSAL"},
	{Name: "Incredible", Size: 360, StringID: 25360, StringKey: "MAPSIZE_INCREDIBLE"},
	{Name: "Monstrous", Size: 400, StringID: 25400, StringKey: "MAPSIZE_MONSTROUS"},
	{Name: "Ludicrous", Size: 480, StringID: 25480, StringKey: "MAPSIZE_LUDICROUS"},
}

type XSRecipe struct {
	Name                string `json:"name,omitempty"`
	Content             string `json:"content,omitempty"`
	ContentFile         string `json:"content_file,omitempty"`
	Mode                string `json:"mode,omitempty"`
	CarrierTitle        string `json:"carrier_title,omitempty"`
	CarrierTriggerName  string `json:"carrier_trigger_name,omitempty"`
	CarrierTriggerIndex *int   `json:"carrier_trigger_index,omitempty"`
	ReplaceCarrier      *bool  `json:"replace_carrier,omitempty"`
	ClearAttachment     *bool  `json:"clear_attachment,omitempty"`
}

type VictoryRecipe struct {
	ConquestRequired               *int `json:"conquest_required,omitempty"`
	Ruins                          *int `json:"ruins,omitempty"`
	ArtifactsRequired              *int `json:"artifacts_required,omitempty"`
	Discovery                      *int `json:"discovery,omitempty"`
	ExploredPercentOfMapRequired   *int `json:"explored_percent_of_map_required,omitempty"`
	GoldRequired                   *int `json:"gold_required,omitempty"`
	AllCustomConditionsRequired    *int `json:"all_custom_conditions_required,omitempty"`
	Mode                           *int `json:"mode,omitempty"`
	RequiredScoreForScoreVictory   *int `json:"required_score_for_score_victory,omitempty"`
	TimeForTimedGameIn10thsOfAYear *int `json:"time_for_timed_game_in_10ths_of_a_year,omitempty"`
}

type PlayerRecipe struct {
	Player           int     `json:"player"`
	Active           *bool   `json:"active,omitempty"`
	Human            *bool   `json:"human,omitempty"`
	TribeName        *string `json:"tribe_name,omitempty"`
	Civilization     *string `json:"civilization,omitempty"`
	LockCivilization *bool   `json:"lock_civilization,omitempty"`
	LockPersonality  *bool   `json:"lock_personality,omitempty"`
	AIName           *string `json:"ai_name,omitempty"`
	AIType           *int    `json:"ai_type,omitempty"`
}

type DiplomacyRecipe struct {
	From   int `json:"from"`
	To     int `json:"to"`
	Stance int `json:"stance"`
}

type DiplomacyOptionsRecipe struct {
	LockTeams               *bool                 `json:"lock_teams,omitempty"`
	AllowPlayersChooseTeams *bool                 `json:"allow_players_choose_teams,omitempty"`
	RandomStartPoints       *bool                 `json:"random_start_points,omitempty"`
	MaxNumberOfTeams        *int                  `json:"max_number_of_teams,omitempty"`
	AlliedVictory           []AlliedVictoryRecipe `json:"allied_victory,omitempty"`
}

type AlliedVictoryRecipe struct {
	Player  int  `json:"player"`
	Enabled bool `json:"enabled"`
}

type ResourceRecipe struct {
	Player     int  `json:"player"`
	Gold       *int `json:"gold,omitempty"`
	Wood       *int `json:"wood,omitempty"`
	Food       *int `json:"food,omitempty"`
	Stone      *int `json:"stone,omitempty"`
	TradeGoods *int `json:"trade_goods,omitempty"`
}

type VariableRecipe struct {
	Op         string  `json:"op"`
	ID         *int    `json:"id,omitempty"`
	Name       string  `json:"name,omitempty"`
	TargetID   *int    `json:"target_id,omitempty"`
	TargetName string  `json:"target_name,omitempty"`
	SetName    *string `json:"set_name,omitempty"`
}

type StringRecipe struct {
	Op       string  `json:"op"`
	ID       *int    `json:"id,omitempty"`
	Text     string  `json:"text,omitempty"`
	SetText  *string `json:"set_text,omitempty"`
	OldText  string  `json:"old_text,omitempty"`
	Required *bool   `json:"required,omitempty"`
}

type MapRecipe struct {
	Op        string `json:"op"`
	X1        int    `json:"x1"`
	Y1        int    `json:"y1"`
	X2        int    `json:"x2"`
	Y2        int    `json:"y2"`
	TargetX   *int   `json:"target_x,omitempty"`
	TargetY   *int   `json:"target_y,omitempty"`
	Radius    *int   `json:"radius,omitempty"`
	Thickness *int   `json:"thickness,omitempty"`
	TerrainID *int   `json:"terrain_id,omitempty"`
	Elevation *int   `json:"elevation,omitempty"`
	Layer     *int   `json:"layer,omitempty"`
}

type UnitRecipe struct {
	Op                    string   `json:"op"`
	Player                int      `json:"player"`
	UnitConst             int      `json:"unit_const"`
	X                     *float64 `json:"x,omitempty"`
	Y                     *float64 `json:"y,omitempty"`
	Z                     *float64 `json:"z,omitempty"`
	ReferenceID           *int     `json:"reference_id,omitempty"`
	TargetPlayer          *int     `json:"target_player,omitempty"`
	TargetIndex           *int     `json:"target_index,omitempty"`
	TargetCaption         string   `json:"target_caption,omitempty"`
	TargetUnitConst       *int     `json:"target_unit_const,omitempty"`
	TargetAreaX1          *float64 `json:"target_area_x1,omitempty"`
	TargetAreaY1          *float64 `json:"target_area_y1,omitempty"`
	TargetAreaX2          *float64 `json:"target_area_x2,omitempty"`
	TargetAreaY2          *float64 `json:"target_area_y2,omitempty"`
	TargetX               *float64 `json:"target_x,omitempty"`
	TargetY               *float64 `json:"target_y,omitempty"`
	OffsetX               *float64 `json:"offset_x,omitempty"`
	OffsetY               *float64 `json:"offset_y,omitempty"`
	ReferenceIDBase       *int     `json:"reference_id_base,omitempty"`
	CaptionSuffix         string   `json:"caption_suffix,omitempty"`
	SetPlayer             *int     `json:"set_player,omitempty"`
	Status                *int     `json:"status,omitempty"`
	Rotation              *float64 `json:"rotation,omitempty"`
	InitialAnimationFrame *int     `json:"initial_animation_frame,omitempty"`
	GarrisonedInID        *int     `json:"garrisoned_in_id,omitempty"`
	CaptionStringID       *int     `json:"caption_string_id,omitempty"`
	CaptionString         string   `json:"caption_string,omitempty"`
}

type TriggerRecipe struct {
	Op                        string            `json:"op"`
	Name                      string            `json:"name"`
	Message                   string            `json:"message"`
	Description               string            `json:"description,omitempty"`
	ShortDescription          string            `json:"short_description,omitempty"`
	TargetIndex               *int              `json:"target_index,omitempty"`
	TargetIndexes             []int             `json:"target_indexes,omitempty"`
	TargetName                string            `json:"target_name,omitempty"`
	TargetPrefix              string            `json:"target_prefix,omitempty"`
	SetName                   *string           `json:"set_name,omitempty"`
	DescriptionStringID       *int              `json:"description_string_table_id,omitempty"`
	ShortDescriptionStringID  *int              `json:"short_description_string_table_id,omitempty"`
	DisplayAsObjective        *bool             `json:"display_as_objective,omitempty"`
	DisplayOnScreen           *bool             `json:"display_on_screen,omitempty"`
	MakeHeader                *bool             `json:"make_header,omitempty"`
	MuteObjectives            *bool             `json:"mute_objectives,omitempty"`
	ExecuteOnLoad             *bool             `json:"execute_on_load,omitempty"`
	ObjectiveDescriptionOrder *int              `json:"objective_description_order,omitempty"`
	Enabled                   *bool             `json:"enabled,omitempty"`
	Looping                   *bool             `json:"looping,omitempty"`
	RemoveEffects             []int             `json:"remove_effects,omitempty"`
	RemoveConditions          []int             `json:"remove_conditions,omitempty"`
	ClearEffects              *bool             `json:"clear_effects,omitempty"`
	ClearConditions           *bool             `json:"clear_conditions,omitempty"`
	ReplaceEffects            []EffectRecipe    `json:"replace_effects,omitempty"`
	ReplaceConditions         []ConditionRecipe `json:"replace_conditions,omitempty"`
	Effects                   []EffectRecipe    `json:"effects,omitempty"`
	Conditions                []ConditionRecipe `json:"conditions,omitempty"`
	DisplayTime               *int              `json:"display_time,omitempty"`
	InstructionPanelPosition  *int              `json:"instruction_panel_position,omitempty"`
	SourcePlayer              *int              `json:"source_player,omitempty"`
	PlaySound                 *int              `json:"play_sound,omitempty"`
	UseTagColorForIcon        *int              `json:"use_tag_color_for_icon,omitempty"`
	ObjectListUnitID          *int              `json:"object_list_unit_id,omitempty"`
	LocationX                 *int              `json:"location_x,omitempty"`
	LocationY                 *int              `json:"location_y,omitempty"`
	ItemID                    *int              `json:"item_id,omitempty"`
	Facet                     *int              `json:"facet,omitempty"`
	DisableSound              *int              `json:"disable_sound,omitempty"`
}

type EffectRecipe struct {
	Op                         string   `json:"op"`
	Message                    string   `json:"message,omitempty"`
	ObjectListUnitID           *int     `json:"object_list_unit_id,omitempty"`
	ObjectListUnitID2          *int     `json:"object_list_unit_id_2,omitempty"`
	SourcePlayer               *int     `json:"source_player,omitempty"`
	TargetPlayer               *int     `json:"target_player,omitempty"`
	TriggerID                  *int     `json:"trigger_id,omitempty"`
	Technology                 *int     `json:"technology,omitempty"`
	Diplomacy                  *int     `json:"diplomacy,omitempty"`
	LocationX                  *int     `json:"location_x,omitempty"`
	LocationY                  *int     `json:"location_y,omitempty"`
	LocationObjectReference    *int     `json:"location_object_reference,omitempty"`
	AreaX1                     *int     `json:"area_x1,omitempty"`
	AreaY1                     *int     `json:"area_y1,omitempty"`
	AreaX2                     *int     `json:"area_x2,omitempty"`
	AreaY2                     *int     `json:"area_y2,omitempty"`
	ObjectGroup                *int     `json:"object_group,omitempty"`
	ObjectType                 *int     `json:"object_type,omitempty"`
	ActionType                 *int     `json:"action_type,omitempty"`
	AttackStance               *int     `json:"attack_stance,omitempty"`
	MaxUnitsAffected           *int     `json:"max_units_affected,omitempty"`
	IssueGroupCommand          *int     `json:"issue_group_command,omitempty"`
	QueueAction                *int     `json:"queue_action,omitempty"`
	DisableGarrisonUnloadSound *int     `json:"disable_garrison_unload_sound,omitempty"`
	DisplayTime                *int     `json:"display_time,omitempty"`
	InstructionPanelPosition   *int     `json:"instruction_panel_position,omitempty"`
	PlaySound                  *int     `json:"play_sound,omitempty"`
	SoundName                  string   `json:"sound_name,omitempty"`
	UseTagColorForIcon         *int     `json:"use_tag_color_for_icon,omitempty"`
	ItemID                     *int     `json:"item_id,omitempty"`
	Quantity                   *int     `json:"quantity,omitempty"`
	QuantityFloat              *float64 `json:"quantity_float,omitempty"`
	Operation                  *int     `json:"operation,omitempty"`
	TributeList                *int     `json:"tribute_list,omitempty"`
	Resource                   *int     `json:"resource,omitempty"`
	Resource1                  *int     `json:"resource_1,omitempty"`
	Resource1Quantity          *int     `json:"resource_1_quantity,omitempty"`
	Resource2                  *int     `json:"resource_2,omitempty"`
	Resource2Quantity          *int     `json:"resource_2_quantity,omitempty"`
	Resource3                  *int     `json:"resource_3,omitempty"`
	Resource3Quantity          *int     `json:"resource_3_quantity,omitempty"`
	StringID                   *int     `json:"string_id,omitempty"`
	ForceResearchTechnology    *int     `json:"force_research_technology,omitempty"`
	VisibilityState            *int     `json:"visibility_state,omitempty"`
	Scroll                     *int     `json:"scroll,omitempty"`
	FlashObject                *int     `json:"flash_object,omitempty"`
	ObjectAttributes           *int     `json:"object_attributes,omitempty"`
	ObjectState                *int     `json:"object_state,omitempty"`
	Facet                      *int     `json:"facet,omitempty"`
	DisableSound               *int     `json:"disable_sound,omitempty"`
	Enabled                    *int     `json:"enabled,omitempty"`
	ButtonLocation             *int     `json:"button_location,omitempty"`
	Hotkey                     *int     `json:"hotkey,omitempty"`
	TrainTime                  *int     `json:"train_time,omitempty"`
	LocalTechnology            *int     `json:"local_technology,omitempty"`
	PlayerColor                *int     `json:"player_color,omitempty"`
	GlobalSound                *int     `json:"global_sound,omitempty"`
	MutualDiplomacy            *int     `json:"mutual_diplomacy,omitempty"`
	ObjectFilter               *int     `json:"object_filter,omitempty"`
	TimeUnit                   *int     `json:"time_unit,omitempty"`
	TimerID                    *int     `json:"timer_id,omitempty"`
	ResetTimer                 *int     `json:"reset_timer,omitempty"`
	Variable                   *int     `json:"variable,omitempty"`
	Variable2                  *int     `json:"variable2,omitempty"`
	XSFunction                 string   `json:"xs_function,omitempty"`
	SelectedObjectIDs          []int    `json:"selected_object_ids,omitempty"`
}

type ConditionRecipe struct {
	Op                             string `json:"op"`
	Timer                          *int   `json:"timer,omitempty"`
	UnitObject                     *int   `json:"unit_object,omitempty"`
	Quantity                       *int   `json:"quantity,omitempty"`
	Attribute                      *int   `json:"attribute,omitempty"`
	ObjectList                     *int   `json:"object_list,omitempty"`
	SourcePlayer                   *int   `json:"source_player,omitempty"`
	Technology                     *int   `json:"technology,omitempty"`
	AreaX1                         *int   `json:"area_x1,omitempty"`
	AreaY1                         *int   `json:"area_y1,omitempty"`
	AreaX2                         *int   `json:"area_x2,omitempty"`
	AreaY2                         *int   `json:"area_y2,omitempty"`
	ObjectGroup                    *int   `json:"object_group,omitempty"`
	ObjectType                     *int   `json:"object_type,omitempty"`
	ObjectState                    *int   `json:"object_state,omitempty"`
	Variable                       *int   `json:"variable,omitempty"`
	Comparison                     *int   `json:"comparison,omitempty"`
	TargetPlayer                   *int   `json:"target_player,omitempty"`
	IncludeChangeableWeaponObjects *int   `json:"include_changeable_weapon_objects,omitempty"`
	Inverted                       *int   `json:"inverted,omitempty"`
}

type Plan struct {
	TriggerCountBefore int            `json:"trigger_count_before"`
	TriggerCountAfter  int            `json:"trigger_count_after"`
	UnitCountBefore    int            `json:"unit_count_before,omitempty"`
	UnitCountAfter     int            `json:"unit_count_after,omitempty"`
	MapTilesChanged    int            `json:"map_tiles_changed,omitempty"`
	Operations         []PlanOp       `json:"operations"`
	Warnings           []string       `json:"warnings,omitempty"`
	Recipe             map[string]any `json:"recipe,omitempty"`
}

type PlanOp struct {
	Op      string `json:"op"`
	Name    string `json:"name,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
	Message string `json:"message,omitempty"`
}

type PatchReport struct {
	Input               string                 `json:"input"`
	Output              string                 `json:"output"`
	TriggerCountBefore  int                    `json:"trigger_count_before"`
	TriggerCountAfter   int                    `json:"trigger_count_after"`
	UnitCountBefore     int                    `json:"unit_count_before,omitempty"`
	UnitCountAfter      int                    `json:"unit_count_after,omitempty"`
	MapTilesChanged     int                    `json:"map_tiles_changed,omitempty"`
	TimestampOfLastSave int                    `json:"timestamp_of_last_save,omitempty"`
	RebuildOK           bool                   `json:"rebuild_ok"`
	InvariantOK         bool                   `json:"invariant_ok"`
	Verification        aoe2.VerificationClaim `json:"verification"`
}

type SmokeRecipeOptions struct {
	Player       int
	PlayerSet    bool
	UnitConst    int
	UnitConstSet bool
	X            int
	XSet         bool
	Y            int
	YSet         bool
	TerrainID    int
	TerrainIDSet bool
	Elevation    int
	ElevationSet bool
	Layer        int
	LayerSet     bool
}

func LoadRecipe(path string) (Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Recipe{}, err
	}
	var recipe Recipe
	if err := json.Unmarshal(data, &recipe); err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

func PlanRecipeFile(input string, recipe Recipe) (Plan, error) {
	file, err := Open(input)
	if err != nil {
		return Plan{}, err
	}
	return file.Plan(recipe)
}

func PatchRecipeFile(input, output string, recipe Recipe) (PatchReport, error) {
	file, err := Open(input)
	if err != nil {
		return PatchReport{}, err
	}
	if !SupportsWriteVersion(file.Version) {
		return PatchReport{}, fmt.Errorf("scenario version %q is read-only in AoE2Kit; writing currently targets DE 1.57 and 1.58 only", file.Version)
	}
	before := 0
	if file.Triggers != nil {
		before = file.Triggers.Count
	}
	unitsBefore := 0
	if file.Units != nil {
		unitsBefore = file.Units.Total
	}
	if err := file.ApplyRecipe(recipe); err != nil {
		return PatchReport{}, err
	}
	timestamp := recipeTimestamp(recipe)
	if err := file.SetScenario(ScenarioRecipe{TimestampOfLastSave: &timestamp}); err != nil {
		return PatchReport{}, err
	}
	mapTilesChanged := plannedMapTiles(recipe)
	if err := file.Write(output); err != nil {
		return PatchReport{}, err
	}
	verified, err := Open(output)
	if err != nil {
		return PatchReport{}, err
	}
	if err := verified.VerifyRebuild(); err != nil {
		return PatchReport{}, err
	}
	after := 0
	invariantOK := false
	if verified.Triggers != nil {
		after = verified.Triggers.Count
		invariantOK = verified.Triggers.InvariantOK
	}
	unitsAfter := 0
	if verified.Units != nil {
		unitsAfter = verified.Units.Total
	}
	if !invariantOK {
		return PatchReport{}, fmt.Errorf("patched scenario trigger invariants failed: %s", verified.Triggers.InvariantNote)
	}
	return PatchReport{
		Input:               input,
		Output:              output,
		TriggerCountBefore:  before,
		TriggerCountAfter:   after,
		UnitCountBefore:     unitsBefore,
		UnitCountAfter:      unitsAfter,
		MapTilesChanged:     mapTilesChanged,
		TimestampOfLastSave: timestamp,
		RebuildOK:           true,
		InvariantOK:         invariantOK,
		Verification:        scenarioWriteVerification(),
	}, nil
}

func recipeTimestamp(recipe Recipe) int {
	if recipe.Scenario != nil && recipe.Scenario.TimestampOfLastSave != nil {
		return *recipe.Scenario.TimestampOfLastSave
	}
	return int(time.Now().Unix())
}

func (f *File) writeSpec() (*Spec, error) {
	if f == nil {
		return LoadCurrentDESpec()
	}
	return LoadDESpecForVersion(f.Version)
}

func SmokeRecipe(opts SmokeRecipeOptions) Recipe {
	if !opts.PlayerSet {
		opts.Player = 1
	}
	if !opts.UnitConstSet {
		opts.UnitConst = 83
	}
	if !opts.XSet {
		opts.X = 10
	}
	if !opts.YSet {
		opts.Y = 10
	}
	if !opts.TerrainIDSet {
		opts.TerrainID = 47
	}
	if !opts.LayerSet {
		opts.Layer = -1
	}
	enabled := false
	sourcePlayer := opts.Player
	displayTime := 8
	createX := opts.X + 1
	createY := opts.Y + 1
	unitX := float64(opts.X) + 1.5
	unitY := float64(opts.Y) + 1.5
	unitZ := float64(0)
	rotation := float64(0)
	status := 2
	timer := 5
	actionType := 1
	triggerZero := 0
	terrain := opts.TerrainID
	elevation := opts.Elevation
	layer := opts.Layer
	return Recipe{
		Triggers: []TriggerRecipe{
			{
				Op:      "add_trigger",
				Name:    "AOE2KIT WRITE SMOKE - display/chat/timer",
				Enabled: &enabled,
				Conditions: []ConditionRecipe{
					{Op: "timer", Timer: &timer},
				},
				Effects: []EffectRecipe{
					{Op: "display_instructions", SourcePlayer: &sourcePlayer, Message: "AOE2KIT WRITE SMOKE: display-instructions written by Go.", DisplayTime: &displayTime},
					{Op: "send_chat", SourcePlayer: &sourcePlayer, Message: "AOE2KIT WRITE SMOKE: send-chat written by Go."},
					{Op: "display_timer", SourcePlayer: &sourcePlayer, Message: "AOE2KIT WRITE SMOKE TIMER", DisplayTime: &displayTime},
				},
			},
			{
				Op:      "add_trigger",
				Name:    "AOE2KIT WRITE SMOKE - create/task/kill/remove",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "create_object", SourcePlayer: &sourcePlayer, ObjectListUnitID: &opts.UnitConst, LocationX: &createX, LocationY: &createY},
					{Op: "task_object", SourcePlayer: &sourcePlayer, ActionType: &actionType, LocationX: &createX, LocationY: &createY},
					{Op: "kill_object", SourcePlayer: &sourcePlayer, ObjectListUnitID: &opts.UnitConst, AreaX1: &opts.X, AreaY1: &opts.Y, AreaX2: &createX, AreaY2: &createY},
					{Op: "remove_object", SourcePlayer: &sourcePlayer, ObjectListUnitID: &opts.UnitConst, AreaX1: &opts.X, AreaY1: &opts.Y, AreaX2: &createX, AreaY2: &createY},
				},
			},
			{
				Op:      "add_trigger",
				Name:    "AOE2KIT WRITE SMOKE - trigger control",
				Enabled: &enabled,
				Effects: []EffectRecipe{
					{Op: "activate_trigger", TriggerID: &triggerZero},
					{Op: "deactivate_trigger", TriggerID: &triggerZero},
				},
			},
		},
		Units: []UnitRecipe{
			{
				Op:        "add_unit",
				Player:    opts.Player,
				UnitConst: opts.UnitConst,
				X:         &unitX,
				Y:         &unitY,
				Z:         &unitZ,
				Status:    &status,
				Rotation:  &rotation,
			},
		},
		Map: []MapRecipe{
			{
				Op:        "set_terrain_rect",
				X1:        opts.X,
				Y1:        opts.Y,
				X2:        opts.X + 2,
				Y2:        opts.Y + 2,
				TerrainID: &terrain,
				Elevation: &elevation,
				Layer:     &layer,
			},
		},
	}
}

func PatchSmokeRecipeFile(input, output string, opts SmokeRecipeOptions) (PatchReport, error) {
	return PatchRecipeFile(input, output, SmokeRecipe(opts))
}

func scenarioWriteVerification() aoe2.VerificationClaim {
	return aoe2.VerificationClaim{
		Label:             "structure_verified_not_engine_verified",
		StructureVerified: true,
		EngineVerified:    false,
		Note:              "scenario writer reparsed the patched scenario, verified body rebuild and trigger invariants, and checked declared counts; DE editor/game load remains the external oracle.",
	}
}

func (f *File) Plan(recipe Recipe) (Plan, error) {
	if f.Triggers == nil {
		return Plan{}, fmt.Errorf("scenario has no trigger info")
	}
	plan := Plan{TriggerCountBefore: f.Triggers.Count, TriggerCountAfter: f.Triggers.Count}
	if f.Units != nil {
		plan.UnitCountBefore = f.Units.Total
		plan.UnitCountAfter = f.Units.Total
	}
	for _, trigger := range recipe.Triggers {
		enabled := false
		if trigger.Enabled != nil {
			enabled = *trigger.Enabled
		}
		switch trigger.Op {
		case "add_display_instructions", "add_create_object", "add_trigger", "copy_trigger", "edit_trigger", "tombstone_trigger", "remove_trigger", "remove_triggers", "remove_system_prefix", "clear_triggers":
			if trigger.Op == "edit_trigger" || trigger.Op == "tombstone_trigger" {
				if _, err := f.findTrigger(trigger); err != nil {
					return Plan{}, err
				}
			}
			if trigger.Op == "copy_trigger" {
				if _, err := f.findTrigger(trigger); err != nil {
					return Plan{}, err
				}
			}
			if trigger.Op == "remove_trigger" {
				index, err := f.findTriggerIndex(trigger)
				if err != nil {
					return Plan{}, err
				}
				deletePlan, err := f.DeletePlan(DeletePlanRequest{Kind: "trigger", ID: index})
				if err != nil {
					return Plan{}, err
				}
				if !deletePlan.CanDelete {
					return Plan{}, fmt.Errorf("remove_trigger target_index %d blocked by %d reference(s)", index, deletePlan.ReferenceSummary.Total)
				}
			}
			if trigger.Op == "remove_triggers" {
				count, err := f.validateRemoveTriggers(trigger.TargetIndexes)
				if err != nil {
					return Plan{}, err
				}
				plan.TriggerCountAfter -= count
			}
			if trigger.Op == "remove_system_prefix" {
				matches, externalRefs, err := f.validateRemoveSystemPrefix(trigger.TargetPrefix)
				if err != nil {
					return Plan{}, err
				}
				if externalRefs.Summary.Total > 0 {
					return Plan{}, fmt.Errorf("remove_system_prefix %q blocked by %d external reference(s)", trigger.TargetPrefix, externalRefs.Summary.Total)
				}
				plan.TriggerCountAfter -= len(matches.TriggerIndexes)
				plan.UnitCountAfter -= len(matches.UnitReferenceIDs)
			}
			plan.Operations = append(plan.Operations, PlanOp{
				Op:      trigger.Op,
				Name:    trigger.Name,
				Enabled: &enabled,
				Message: trigger.Message,
			})
			switch trigger.Op {
			case "add_display_instructions", "add_create_object", "add_trigger", "copy_trigger":
				plan.TriggerCountAfter++
			case "remove_trigger":
				plan.TriggerCountAfter--
			case "remove_triggers", "remove_system_prefix":
			case "clear_triggers":
				plan.TriggerCountAfter = 0
			}
		default:
			return Plan{}, fmt.Errorf("unsupported trigger op %q", trigger.Op)
		}
	}
	if recipe.XS != nil {
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_xs", Name: recipe.XS.Name})
	}
	if recipe.Scenario != nil && recipe.Scenario.PlayerCount != nil {
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_scenario_player_count", Name: fmt.Sprintf("%d", *recipe.Scenario.PlayerCount)})
	}
	if recipe.Victory != nil {
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_global_victory", Name: recipe.Victory.summary()})
	}
	for _, player := range recipe.Players {
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_player", Name: fmt.Sprintf("P%d", player.Player)})
	}
	for _, diplomacy := range recipe.Diplomacy {
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_diplomacy", Name: fmt.Sprintf("P%d->P%d=%d", diplomacy.From, diplomacy.To, diplomacy.Stance)})
	}
	if recipe.DiplomacyOptions != nil {
		if err := f.validateDiplomacyOptions(*recipe.DiplomacyOptions); err != nil {
			return Plan{}, err
		}
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_diplomacy_options", Name: recipe.DiplomacyOptions.summary()})
	}
	for _, resource := range recipe.Resources {
		plan.Operations = append(plan.Operations, PlanOp{Op: "set_resources", Name: fmt.Sprintf("P%d", resource.Player)})
	}
	for _, stringRecipe := range recipe.Strings {
		switch stringRecipe.Op {
		case "add_string":
			if _, err := f.planStringSlot(stringRecipe, true); err != nil {
				return Plan{}, err
			}
		case "set_string":
			if _, err := f.planStringSlot(stringRecipe, false); err != nil {
				return Plan{}, err
			}
			if stringRecipe.SetText == nil && stringRecipe.Text == "" {
				return Plan{}, fmt.Errorf("set_string requires set_text or text")
			}
		case "tombstone_string", "clear_string":
			id, err := f.planStringSlot(stringRecipe, false)
			if err != nil {
				return Plan{}, err
			}
			refs, err := f.References(ReferenceOptions{Kind: "string", TargetID: &id})
			if err != nil {
				return Plan{}, err
			}
			if refs.Summary.Total > 0 {
				return Plan{}, fmt.Errorf("refuses to %s string %d because it is referenced %d time(s)", strings.TrimSuffix(stringRecipe.Op, "_string"), id, refs.Summary.Total)
			}
		default:
			return Plan{}, fmt.Errorf("unsupported string op %q", stringRecipe.Op)
		}
		plan.Operations = append(plan.Operations, PlanOp{Op: stringRecipe.Op, Name: stringRecipe.summary()})
	}
	for _, variable := range recipe.Variables {
		switch variable.Op {
		case "add_variable":
			if variable.Name == "" {
				return Plan{}, fmt.Errorf("add_variable requires name")
			}
		case "edit_variable":
			if _, err := f.findVariable(variable); err != nil {
				return Plan{}, err
			}
			if variable.SetName == nil && variable.Name == "" {
				return Plan{}, fmt.Errorf("edit_variable requires set_name or name")
			}
		case "tombstone_variable", "remove_variable":
			target, err := f.findVariable(variable)
			if err != nil {
				return Plan{}, err
			}
			id, _ := target.intValue("variable_id")
			refs, err := f.References(ReferenceOptions{Kind: "variable", TargetID: &id})
			if err != nil {
				return Plan{}, err
			}
			if refs.Summary.Total > 0 {
				return Plan{}, fmt.Errorf("refuses to tombstone variable %d because it is referenced %d time(s)", id, refs.Summary.Total)
			}
		default:
			return Plan{}, fmt.Errorf("unsupported variable op %q", variable.Op)
		}
		plan.Operations = append(plan.Operations, PlanOp{Op: variable.Op, Name: variable.summary()})
	}
	for _, unit := range recipe.Units {
		switch unit.Op {
		case "add_unit", "edit_unit", "remove_unit", "remove_units_in_area", "remove_units_for_player", "copy_units_in_area", "move_units_in_area", "edit_units_in_area":
			if unit.Op == "edit_unit" {
				if _, _, _, err := f.findUnit(unit); err != nil {
					return Plan{}, err
				}
			}
			if unit.Op == "remove_unit" {
				target, _, _, err := f.findUnit(unit)
				if err != nil {
					return Plan{}, err
				}
				if referenceID, ok := target.intValue("reference_id"); ok && referenceID >= 0 {
					if err := f.ensureUnitNotReferenced(referenceID); err != nil {
						return Plan{}, err
					}
				}
			}
			matches := 0
			if unit.Op == "remove_units_in_area" {
				targets, err := f.findUnitsInArea(unit)
				if err != nil {
					return Plan{}, err
				}
				for _, target := range targets {
					if referenceID, ok := target.Unit.intValue("reference_id"); ok && referenceID >= 0 {
						if err := f.ensureUnitNotReferenced(referenceID); err != nil {
							return Plan{}, err
						}
					}
				}
				matches = len(targets)
			}
			if unit.Op == "remove_units_for_player" {
				targets, err := f.findUnitsByPlayer(unit.TargetPlayer)
				if err != nil {
					return Plan{}, err
				}
				for _, target := range targets {
					if referenceID, ok := target.Unit.intValue("reference_id"); ok && referenceID >= 0 {
						if err := f.ensureUnitNotReferenced(referenceID); err != nil {
							return Plan{}, err
						}
					}
				}
				matches = len(targets)
			}
			if unit.Op == "copy_units_in_area" {
				targets, err := f.findUnitsInArea(unit)
				if err != nil {
					return Plan{}, err
				}
				if err := f.validateCopiedUnitReferenceIDs(targets, unit); err != nil {
					return Plan{}, err
				}
				matches = len(targets)
			}
			if unit.Op == "move_units_in_area" {
				targets, err := f.findUnitsInArea(unit)
				if err != nil {
					return Plan{}, err
				}
				if err := f.validateMoveUnitsInArea(targets, unit); err != nil {
					return Plan{}, err
				}
				matches = len(targets)
			}
			if unit.Op == "edit_units_in_area" {
				targets, err := f.findUnitsInArea(unit)
				if err != nil {
					return Plan{}, err
				}
				if err := f.validateEditUnitsInArea(unit); err != nil {
					return Plan{}, err
				}
				matches = len(targets)
			}
			plan.Operations = append(plan.Operations, PlanOp{Op: unit.Op})
			switch unit.Op {
			case "add_unit":
				plan.UnitCountAfter++
			case "remove_unit":
				plan.UnitCountAfter--
			case "remove_units_in_area":
				plan.UnitCountAfter -= matches
			case "remove_units_for_player":
				plan.UnitCountAfter -= matches
			case "copy_units_in_area":
				plan.UnitCountAfter += matches
			case "move_units_in_area":
				_ = matches
			case "edit_units_in_area":
				_ = matches
			}
		default:
			return Plan{}, fmt.Errorf("unsupported unit op %q", unit.Op)
		}
	}
	for _, mapPatch := range recipe.Map {
		switch mapPatch.Op {
		case "set_terrain_rect", "set_terrain_circle", "set_terrain_line", "set_terrain_border", "copy_terrain_area":
			changed, err := f.countMapTiles(mapPatch)
			if err != nil {
				return Plan{}, err
			}
			plan.Operations = append(plan.Operations, PlanOp{Op: mapPatch.Op})
			plan.MapTilesChanged += changed
		default:
			return Plan{}, fmt.Errorf("unsupported map op %q", mapPatch.Op)
		}
	}
	return plan, nil
}

func (f *File) ApplyRecipe(recipe Recipe) error {
	if recipe.Scenario != nil {
		if err := f.SetScenario(*recipe.Scenario); err != nil {
			return err
		}
	}
	if recipe.Victory != nil {
		if err := f.SetGlobalVictory(*recipe.Victory); err != nil {
			return err
		}
	}
	for _, player := range recipe.Players {
		if err := f.SetPlayer(player); err != nil {
			return err
		}
	}
	for _, diplomacy := range recipe.Diplomacy {
		if err := f.SetDiplomacy(diplomacy); err != nil {
			return err
		}
	}
	if recipe.DiplomacyOptions != nil {
		if err := f.SetDiplomacyOptions(*recipe.DiplomacyOptions); err != nil {
			return err
		}
	}
	for _, resource := range recipe.Resources {
		if err := f.SetResources(resource); err != nil {
			return err
		}
	}
	for _, stringRecipe := range recipe.Strings {
		switch stringRecipe.Op {
		case "add_string":
			if _, err := f.AddString(stringRecipe); err != nil {
				return err
			}
		case "set_string":
			if err := f.SetString(stringRecipe); err != nil {
				return err
			}
		case "tombstone_string":
			if err := f.TombstoneString(stringRecipe); err != nil {
				return err
			}
		case "clear_string":
			if err := f.ClearString(stringRecipe); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported string op %q", stringRecipe.Op)
		}
	}
	for _, variable := range recipe.Variables {
		switch variable.Op {
		case "add_variable":
			if err := f.AddVariable(variable); err != nil {
				return err
			}
		case "edit_variable":
			if err := f.EditVariable(variable); err != nil {
				return err
			}
		case "tombstone_variable":
			if err := f.TombstoneVariable(variable); err != nil {
				return err
			}
		case "remove_variable":
			if err := f.RemoveVariable(variable); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported variable op %q", variable.Op)
		}
	}
	for _, trigger := range recipe.Triggers {
		switch trigger.Op {
		case "add_display_instructions":
			if err := f.AddDisplayInstructionsTrigger(trigger); err != nil {
				return err
			}
		case "add_create_object":
			if err := f.AddCreateObjectTrigger(trigger); err != nil {
				return err
			}
		case "add_trigger":
			if err := f.AddTrigger(trigger); err != nil {
				return err
			}
		case "copy_trigger":
			if err := f.CopyTrigger(trigger); err != nil {
				return err
			}
		case "edit_trigger":
			if err := f.EditTrigger(trigger); err != nil {
				return err
			}
		case "tombstone_trigger":
			if err := f.TombstoneTrigger(trigger); err != nil {
				return err
			}
		case "remove_trigger":
			if err := f.RemoveTrigger(trigger); err != nil {
				return err
			}
		case "remove_triggers":
			if err := f.RemoveTriggers(trigger); err != nil {
				return err
			}
		case "remove_system_prefix":
			if err := f.RemoveSystemPrefix(trigger.TargetPrefix); err != nil {
				return err
			}
		case "clear_triggers":
			if err := f.ClearTriggers(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported trigger op %q", trigger.Op)
		}
	}
	if recipe.XS != nil {
		if err := f.SetXS(*recipe.XS); err != nil {
			return err
		}
	}
	var pendingAddUnits []UnitRecipe
	flushAddUnits := func() error {
		if len(pendingAddUnits) == 0 {
			return nil
		}
		if err := f.AddUnits(pendingAddUnits); err != nil {
			return err
		}
		pendingAddUnits = nil
		return nil
	}
	for _, unit := range recipe.Units {
		switch unit.Op {
		case "add_unit":
			pendingAddUnits = append(pendingAddUnits, unit)
		case "edit_unit":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if err := f.EditUnit(unit); err != nil {
				return err
			}
		case "remove_unit":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if err := f.RemoveUnit(unit); err != nil {
				return err
			}
		case "remove_units_in_area":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if _, err := f.RemoveUnitsInArea(unit); err != nil {
				return err
			}
		case "remove_units_for_player":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if _, err := f.RemoveUnitsForPlayer(unit); err != nil {
				return err
			}
		case "copy_units_in_area":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if _, err := f.CopyUnitsInArea(unit); err != nil {
				return err
			}
		case "move_units_in_area":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if _, err := f.MoveUnitsInArea(unit); err != nil {
				return err
			}
		case "edit_units_in_area":
			if err := flushAddUnits(); err != nil {
				return err
			}
			if _, err := f.EditUnitsInArea(unit); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported unit op %q", unit.Op)
		}
	}
	if err := flushAddUnits(); err != nil {
		return err
	}
	for _, mapPatch := range recipe.Map {
		switch mapPatch.Op {
		case "set_terrain_rect", "set_terrain_circle", "set_terrain_line", "set_terrain_border":
			if err := f.SetTerrain(mapPatch); err != nil {
				return err
			}
		case "copy_terrain_area":
			if err := f.CopyTerrainArea(mapPatch); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported map op %q", mapPatch.Op)
		}
	}
	return nil
}

func (f *File) SetScenario(recipe ScenarioRecipe) error {
	changed := false
	if recipe.TimestampOfLastSave != nil {
		if f.headerRoot == nil {
			return fmt.Errorf("missing FileHeader section")
		}
		timestamp := *recipe.TimestampOfLastSave
		if timestamp < 0 {
			return fmt.Errorf("scenario timestamp_of_last_save %d must be non-negative", timestamp)
		}
		if err := setIntField(f.headerRoot, "timestamp_of_last_save", "u32", timestamp); err != nil {
			return err
		}
		f.header = f.headerRoot.raw()
		f.HeaderBytes = len(f.header)
		changed = true
	}
	if recipe.PlayerCount != nil {
		if f.headerRoot == nil {
			return fmt.Errorf("missing FileHeader section")
		}
		playerCount := *recipe.PlayerCount
		if playerCount < 1 || playerCount > 8 {
			return fmt.Errorf("scenario player_count %d out of range 1..8", playerCount)
		}
		if err := setIntField(f.headerRoot, "player_count", "u32", playerCount); err != nil {
			return err
		}
		if units := f.root.section("Units"); units != nil {
			if err := setIntField(units, "number_of_players", "u32", len(units.list("players_units"))); err != nil {
				return err
			}
		}
		f.header = f.headerRoot.raw()
		f.HeaderBytes = len(f.header)
		f.PlayerCount = playerCount
		changed = true
	}
	if !changed {
		return fmt.Errorf("scenario recipe has no fields to set")
	}
	return nil
}

func (r VictoryRecipe) summary() string {
	var parts []string
	fields := []struct {
		name  string
		value *int
	}{
		{"conquest_required", r.ConquestRequired},
		{"ruins", r.Ruins},
		{"artifacts_required", r.ArtifactsRequired},
		{"discovery", r.Discovery},
		{"explored_percent_of_map_required", r.ExploredPercentOfMapRequired},
		{"gold_required", r.GoldRequired},
		{"all_custom_conditions_required", r.AllCustomConditionsRequired},
		{"mode", r.Mode},
		{"required_score_for_score_victory", r.RequiredScoreForScoreVictory},
		{"time_for_timed_game_in_10ths_of_a_year", r.TimeForTimedGameIn10thsOfAYear},
	}
	for _, field := range fields {
		if field.value != nil {
			parts = append(parts, fmt.Sprintf("%s=%d", field.name, *field.value))
		}
	}
	if len(parts) == 0 {
		return "no fields"
	}
	return strings.Join(parts, ",")
}

func (f *File) SetGlobalVictory(recipe VictoryRecipe) error {
	victory := f.root.section("GlobalVictory")
	if victory == nil {
		return fmt.Errorf("missing GlobalVictory section")
	}
	fields := []struct {
		name  string
		value *int
	}{
		{"conquest_required", recipe.ConquestRequired},
		{"ruins", recipe.Ruins},
		{"artifacts_required", recipe.ArtifactsRequired},
		{"discovery", recipe.Discovery},
		{"explored_percent_of_map_required", recipe.ExploredPercentOfMapRequired},
		{"gold_required", recipe.GoldRequired},
		{"all_custom_conditions_required", recipe.AllCustomConditionsRequired},
		{"mode", recipe.Mode},
		{"required_score_for_score_victory", recipe.RequiredScoreForScoreVictory},
		{"time_for_timed_game_in_10ths_of_a_year", recipe.TimeForTimedGameIn10thsOfAYear},
	}
	changed := false
	for _, field := range fields {
		if field.value != nil {
			if err := setIntField(victory, field.name, "u32", *field.value); err != nil {
				return err
			}
			changed = true
		}
	}
	if !changed {
		return fmt.Errorf("victory recipe has no fields to set")
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (f *File) SetXS(recipe XSRecipe) error {
	content, err := xsRecipeContent(recipe)
	if err != nil {
		return err
	}
	mode := normalizeXSMode(recipe.Mode)
	switch mode {
	case "attachment":
		if recipe.Name == "" {
			return fmt.Errorf("xs attachment mode requires name")
		}
		return f.setXSAttachment(recipe.Name, content)
	case "carrier":
		if content == "" {
			return fmt.Errorf("xs carrier mode requires content or content_file")
		}
		clearAttachment := true
		if recipe.ClearAttachment != nil {
			clearAttachment = *recipe.ClearAttachment
		}
		if clearAttachment {
			if err := f.clearXSAttachment(); err != nil {
				return err
			}
		}
		return f.SetXSCarrier(recipe, content)
	case "inline_runtime":
		if content == "" {
			return fmt.Errorf("xs inline_runtime mode requires content or content_file")
		}
		return f.SetXSInlineRuntime(recipe, content)
	case "attachment_and_carrier":
		if recipe.Name == "" {
			return fmt.Errorf("xs attachment_and_carrier mode requires name")
		}
		if content == "" {
			return fmt.Errorf("xs attachment_and_carrier mode requires content or content_file")
		}
		if err := f.setXSAttachment(recipe.Name, content); err != nil {
			return err
		}
		return f.SetXSCarrier(recipe, content)
	default:
		return fmt.Errorf("unsupported xs mode %q", recipe.Mode)
	}
}

func xsRecipeContent(recipe XSRecipe) (string, error) {
	if recipe.Content != "" && recipe.ContentFile != "" {
		return "", fmt.Errorf("xs supports content or content_file, not both")
	}
	if recipe.ContentFile == "" {
		return recipe.Content, nil
	}
	data, err := os.ReadFile(recipe.ContentFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func normalizeXSMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "attachment", "file", "files":
		return "attachment"
	case "carrier", "embedded_carrier", "parser_carrier", "parser-style", "parser_style":
		return "carrier"
	case "inline", "inline_runtime", "runtime_inline", "runtime_carrier", "spiral", "spiral_style":
		return "inline_runtime"
	case "both", "attachment_and_carrier", "file_and_carrier", "files_and_carrier":
		return "attachment_and_carrier"
	default:
		return strings.ToLower(strings.TrimSpace(mode))
	}
}

func (f *File) setXSAttachment(name, content string) error {
	name = normalizeXSFileName(name)
	mapSection := f.root.section("Map")
	if mapSection == nil {
		return fmt.Errorf("missing Map section")
	}
	if err := setStringField(mapSection, "script_name", "str16", name); err != nil {
		return err
	}
	files := f.root.section("Files")
	if files == nil {
		return fmt.Errorf("missing Files section")
	}
	if err := setStringField(files, "script_file_path", "str16", name); err != nil {
		return err
	}
	if err := setStringField(files, "script_file_content", "str32", content); err != nil {
		return err
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (f *File) clearXSAttachment() error {
	mapSection := f.root.section("Map")
	if mapSection == nil {
		return fmt.Errorf("missing Map section")
	}
	if err := setZeroLengthStringField(mapSection, "script_name", "str16"); err != nil {
		return err
	}
	files := f.root.section("Files")
	if files == nil {
		return fmt.Errorf("missing Files section")
	}
	if err := setZeroLengthStringField(files, "script_file_path", "str16"); err != nil {
		return err
	}
	if err := setZeroLengthStringField(files, "script_file_content", "str32"); err != nil {
		return err
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (f *File) SetXSCarrier(recipe XSRecipe, content string) error {
	return f.setXSCarrier(recipe, content, false, false)
}

func (f *File) SetXSInlineRuntime(recipe XSRecipe, content string) error {
	if err := f.clearXSAttachment(); err != nil {
		return err
	}
	if strings.TrimSpace(recipe.CarrierTitle) == "" {
		recipe.CarrierTitle = "XS string"
	}
	if strings.TrimSpace(recipe.CarrierTriggerName) == "" {
		recipe.CarrierTriggerName = "XS SCRIPT"
	}
	if recipe.CarrierTriggerIndex == nil {
		index := 0
		recipe.CarrierTriggerIndex = &index
	}
	return f.setXSCarrier(recipe, content, true, true)
}

func (f *File) setXSCarrier(recipe XSRecipe, content string, enabled bool, preferInsert bool) error {
	replace := true
	if recipe.ReplaceCarrier != nil {
		replace = *recipe.ReplaceCarrier
	}
	triggerName := strings.TrimSpace(recipe.CarrierTriggerName)
	if triggerName == "" {
		triggerName = "XS SCRIPT"
	}
	title := strings.TrimSpace(recipe.CarrierTitle)
	if title == "" {
		if recipe.ContentFile != "" {
			title = filepath.Base(recipe.ContentFile)
		} else if recipe.Name != "" {
			title = normalizeXSFileName(recipe.Name)
		} else {
			title = "XS string"
		}
	}
	message := xsCarrierMessage(title, content)
	triggerEnabled := enabled
	notLooping := false
	sourcePlayer := 1
	effect := EffectRecipe{Op: "script_call", SourcePlayer: &sourcePlayer, Message: message}
	if recipe.CarrierTriggerIndex != nil {
		index := *recipe.CarrierTriggerIndex
		if preferInsert {
			triggers := f.root.section("Triggers")
			if triggers == nil {
				return fmt.Errorf("missing Triggers section")
			}
			triggerNodes := triggers.list("trigger_data")
			if index < 0 || index > len(triggerNodes) {
				return fmt.Errorf("carrier_trigger_index %d out of insert range 0..%d", index, len(triggerNodes))
			}
			if index == len(triggerNodes) || !triggerHasXSCarrier(triggerNodes[index]) {
				insert := TriggerRecipe{
					Op:      "add_trigger",
					Name:    triggerName,
					Enabled: &triggerEnabled,
					Looping: &notLooping,
					Effects: []EffectRecipe{effect},
				}
				raw, err := f.buildTriggerRaw(insert)
				if err != nil {
					return err
				}
				return f.insertRawTrigger(index, raw)
			}
		}
		clearConditions := true
		edit := TriggerRecipe{
			Op:              "edit_trigger",
			TargetIndex:     recipe.CarrierTriggerIndex,
			SetName:         &triggerName,
			Enabled:         &triggerEnabled,
			Looping:         &notLooping,
			ReplaceEffects:  []EffectRecipe{effect},
			ClearConditions: &clearConditions,
		}
		return f.EditTrigger(edit)
	}
	if replace {
		carriers := f.xsEmbeddedCarriers()
		if len(carriers) > 0 {
			clearConditions := true
			edit := TriggerRecipe{
				Op:                "edit_trigger",
				TargetIndex:       &carriers[0].TriggerIndex,
				SetName:           &triggerName,
				Enabled:           &triggerEnabled,
				Looping:           &notLooping,
				ReplaceEffects:    []EffectRecipe{effect},
				ClearConditions:   &clearConditions,
				ClearEffects:      nil,
				ReplaceConditions: nil,
			}
			return f.EditTrigger(edit)
		}
	}
	return f.AddTrigger(TriggerRecipe{
		Op:      "add_trigger",
		Name:    triggerName,
		Enabled: &triggerEnabled,
		Looping: &notLooping,
		Effects: []EffectRecipe{effect},
	})
}

func xsCarrierMessage(title, content string) string {
	return fmt.Sprintf("// ------------------------- %s -------------------------\n%s\n\n", title, strings.TrimRight(content, "\r\n"))
}

func normalizeXSFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || strings.HasSuffix(strings.ToLower(name), ".xs") {
		return name
	}
	return name + ".xs"
}

func (f *File) SetPlayer(recipe PlayerRecipe) error {
	dataHeader := f.root.section("DataHeader")
	if dataHeader == nil {
		return fmt.Errorf("missing DataHeader section")
	}
	playerData := dataHeader.list("player_data_1")
	if recipe.Player < 0 || recipe.Player >= len(playerData) {
		return fmt.Errorf("player %d out of range 0..%d", recipe.Player, len(playerData)-1)
	}
	node := playerData[recipe.Player]
	if recipe.Active != nil {
		if err := setIntField(node, "active", "u32", boolInt(*recipe.Active)); err != nil {
			return err
		}
	}
	if recipe.Human != nil {
		if err := setIntField(node, "human", "u32", boolInt(*recipe.Human)); err != nil {
			return err
		}
	}
	if recipe.Civilization != nil {
		if err := setStringField(node, "civilization", "str16", *recipe.Civilization); err != nil {
			return err
		}
	}
	if recipe.TribeName != nil {
		tribeNames := dataHeader.stringList("tribe_names")
		if recipe.Player >= len(tribeNames) {
			return fmt.Errorf("tribe_names player %d out of range 0..%d", recipe.Player, len(tribeNames)-1)
		}
		tribeNames[recipe.Player] = *recipe.TribeName
		if err := setStringListField(dataHeader, "tribe_names", "c256", tribeNames); err != nil {
			return err
		}
	}
	if recipe.LockCivilization != nil {
		locks := dataHeader.intList("per_player_lock_civilization")
		if recipe.Player >= len(locks) {
			return fmt.Errorf("per_player_lock_civilization player %d out of range 0..%d", recipe.Player, len(locks)-1)
		}
		locks[recipe.Player] = boolInt(*recipe.LockCivilization)
		if err := setIntListField(dataHeader, "per_player_lock_civilization", "u32", locks); err != nil {
			return err
		}
	}
	if recipe.LockPersonality != nil {
		locks := dataHeader.intList("per_player_lock_personality")
		if recipe.Player >= len(locks) {
			return fmt.Errorf("per_player_lock_personality player %d out of range 0..%d", recipe.Player, len(locks)-1)
		}
		locks[recipe.Player] = boolInt(*recipe.LockPersonality)
		if err := setIntListField(dataHeader, "per_player_lock_personality", "u32", locks); err != nil {
			return err
		}
	}
	if recipe.AIName != nil || recipe.AIType != nil {
		playerDataTwo := f.root.section("PlayerDataTwo")
		if playerDataTwo == nil {
			return fmt.Errorf("missing PlayerDataTwo section")
		}
		if recipe.AIName != nil {
			names := playerDataTwo.stringList("ai_names")
			if recipe.Player >= len(names) {
				return fmt.Errorf("ai_names player %d out of range 0..%d", recipe.Player, len(names)-1)
			}
			names[recipe.Player] = *recipe.AIName
			if err := setStringListField(playerDataTwo, "ai_names", "str16", names); err != nil {
				return err
			}
		}
		if recipe.AIType != nil {
			types := playerDataTwo.intList("ai_type")
			if recipe.Player >= len(types) {
				return fmt.Errorf("ai_type player %d out of range 0..%d", recipe.Player, len(types)-1)
			}
			types[recipe.Player] = *recipe.AIType
			if err := setIntListField(playerDataTwo, "ai_type", "u8", types); err != nil {
				return err
			}
		}
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	f.Players = f.root.playerInfo()
	return nil
}

func (f *File) SetDiplomacy(recipe DiplomacyRecipe) error {
	diplomacy := f.root.section("Diplomacy")
	if diplomacy == nil {
		return fmt.Errorf("missing Diplomacy section")
	}
	rows := diplomacy.list("per_player_diplomacy")
	if recipe.From < 0 || recipe.From >= len(rows) {
		return fmt.Errorf("diplomacy from-player %d out of range 0..%d", recipe.From, len(rows)-1)
	}
	row := rows[recipe.From]
	stances := row.uint32List("stance_with_each_player")
	if recipe.To < 0 || recipe.To >= len(stances) {
		return fmt.Errorf("diplomacy to-player %d out of range 0..%d", recipe.To, len(stances)-1)
	}
	stances[recipe.To] = uint32(recipe.Stance)
	if err := setUint32ListField(row, "stance_with_each_player", stances); err != nil {
		return err
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (r DiplomacyOptionsRecipe) summary() string {
	var parts []string
	if r.LockTeams != nil {
		parts = append(parts, fmt.Sprintf("lock_teams=%t", *r.LockTeams))
	}
	if r.AllowPlayersChooseTeams != nil {
		parts = append(parts, fmt.Sprintf("allow_players_choose_teams=%t", *r.AllowPlayersChooseTeams))
	}
	if r.RandomStartPoints != nil {
		parts = append(parts, fmt.Sprintf("random_start_points=%t", *r.RandomStartPoints))
	}
	if r.MaxNumberOfTeams != nil {
		parts = append(parts, fmt.Sprintf("max_number_of_teams=%d", *r.MaxNumberOfTeams))
	}
	for _, allied := range r.AlliedVictory {
		parts = append(parts, fmt.Sprintf("P%d allied_victory=%t", allied.Player, allied.Enabled))
	}
	if len(parts) == 0 {
		return "no fields"
	}
	return strings.Join(parts, ",")
}

func (f *File) validateDiplomacyOptions(recipe DiplomacyOptionsRecipe) error {
	if recipe.LockTeams == nil && recipe.AllowPlayersChooseTeams == nil && recipe.RandomStartPoints == nil &&
		recipe.MaxNumberOfTeams == nil && len(recipe.AlliedVictory) == 0 {
		return fmt.Errorf("diplomacy_options has no fields to set")
	}
	if recipe.MaxNumberOfTeams != nil && (*recipe.MaxNumberOfTeams < 0 || *recipe.MaxNumberOfTeams > 16) {
		return fmt.Errorf("max_number_of_teams %d out of range 0..16", *recipe.MaxNumberOfTeams)
	}
	for _, allied := range recipe.AlliedVictory {
		if allied.Player < 0 || allied.Player > 15 {
			return fmt.Errorf("allied_victory player %d out of range 0..15", allied.Player)
		}
	}
	return nil
}

func (f *File) SetDiplomacyOptions(recipe DiplomacyOptionsRecipe) error {
	if err := f.validateDiplomacyOptions(recipe); err != nil {
		return err
	}
	diplomacy := f.root.section("Diplomacy")
	if diplomacy == nil {
		return fmt.Errorf("missing Diplomacy section")
	}
	if recipe.LockTeams != nil {
		if err := setBoolByteField(diplomacy, "lock_teams", *recipe.LockTeams); err != nil {
			return err
		}
	}
	if recipe.AllowPlayersChooseTeams != nil {
		if err := setBoolByteField(diplomacy, "allow_players_choose_teams", *recipe.AllowPlayersChooseTeams); err != nil {
			return err
		}
	}
	if recipe.RandomStartPoints != nil {
		if err := setBoolByteField(diplomacy, "random_start_points", *recipe.RandomStartPoints); err != nil {
			return err
		}
	}
	if recipe.MaxNumberOfTeams != nil {
		if err := setIntField(diplomacy, "max_number_of_teams", "u8", *recipe.MaxNumberOfTeams); err != nil {
			return err
		}
	}
	if len(recipe.AlliedVictory) > 0 {
		values := diplomacy.uint32List("per_player_allied_victory")
		if len(values) == 0 {
			return fmt.Errorf("missing per_player_allied_victory")
		}
		for _, allied := range recipe.AlliedVictory {
			if allied.Player < 0 || allied.Player >= len(values) {
				return fmt.Errorf("allied_victory player %d out of range 0..%d", allied.Player, len(values)-1)
			}
			if allied.Enabled {
				values[allied.Player] = 1
			} else {
				values[allied.Player] = 0
			}
		}
		if err := setUint32ListField(diplomacy, "per_player_allied_victory", values); err != nil {
			return err
		}
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (f *File) SetResources(recipe ResourceRecipe) error {
	playerDataTwo := f.root.section("PlayerDataTwo")
	if playerDataTwo == nil {
		return fmt.Errorf("missing PlayerDataTwo section")
	}
	resources := playerDataTwo.list("resources")
	if recipe.Player < 0 || recipe.Player >= len(resources) {
		return fmt.Errorf("resource player %d out of range 0..%d", recipe.Player, len(resources)-1)
	}
	node := resources[recipe.Player]
	fields := []struct {
		name  string
		value *int
	}{
		{"gold", recipe.Gold},
		{"wood", recipe.Wood},
		{"food", recipe.Food},
		{"stone", recipe.Stone},
		{"trade_goods", recipe.TradeGoods},
	}
	for _, field := range fields {
		if field.value != nil {
			if err := setIntField(node, field.name, "s32", *field.value); err != nil {
				return err
			}
		}
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (r StringRecipe) summary() string {
	switch r.Op {
	case "add_string":
		if r.ID != nil {
			return fmt.Sprintf("%d:%s", *r.ID, r.Text)
		}
		return r.Text
	case "set_string", "tombstone_string", "clear_string":
		if r.ID != nil {
			if r.SetText != nil {
				return fmt.Sprintf("%d->%s", *r.ID, *r.SetText)
			}
			if r.Text != "" {
				return fmt.Sprintf("%d->%s", *r.ID, r.Text)
			}
			return fmt.Sprintf("%d", *r.ID)
		}
		if r.OldText != "" {
			return r.OldText
		}
	}
	return r.Op
}

func (f *File) AddString(recipe StringRecipe) (int, error) {
	text := recipe.Text
	if recipe.SetText != nil {
		text = *recipe.SetText
	}
	if text == "" {
		return 0, fmt.Errorf("add_string requires text or set_text")
	}
	id, err := f.planStringSlot(recipe, true)
	if err != nil {
		return 0, err
	}
	values, err := f.scenarioStringValues()
	if err != nil {
		return 0, err
	}
	if values[id] != "" {
		return 0, fmt.Errorf("string id %d is already occupied", id)
	}
	values[id] = text
	if err := f.setScenarioStringValues(values); err != nil {
		return 0, err
	}
	return id, nil
}

func (f *File) SetString(recipe StringRecipe) error {
	id, err := f.planStringSlot(recipe, false)
	if err != nil {
		return err
	}
	text := recipe.Text
	if recipe.SetText != nil {
		text = *recipe.SetText
	}
	if text == "" {
		return fmt.Errorf("set_string requires text or set_text")
	}
	values, err := f.scenarioStringValues()
	if err != nil {
		return err
	}
	if recipe.OldText != "" && values[id] != recipe.OldText {
		return fmt.Errorf("string id %d text %q does not match required old_text %q", id, values[id], recipe.OldText)
	}
	values[id] = text
	return f.setScenarioStringValues(values)
}

func (f *File) TombstoneString(recipe StringRecipe) error {
	id, err := f.planStringSlot(recipe, false)
	if err != nil {
		return err
	}
	refs, err := f.References(ReferenceOptions{Kind: "string", TargetID: &id})
	if err != nil {
		return err
	}
	if refs.Summary.Total > 0 {
		return fmt.Errorf("refuses to tombstone string %d because it is referenced %d time(s)", id, refs.Summary.Total)
	}
	values, err := f.scenarioStringValues()
	if err != nil {
		return err
	}
	if recipe.OldText != "" && values[id] != recipe.OldText {
		return fmt.Errorf("string id %d text %q does not match required old_text %q", id, values[id], recipe.OldText)
	}
	tombstone := fmt.Sprintf("_DeletedString%d", id)
	if recipe.SetText != nil {
		tombstone = *recipe.SetText
	}
	values[id] = tombstone
	return f.setScenarioStringValues(values)
}

func (f *File) ClearString(recipe StringRecipe) error {
	id, err := f.planStringSlot(recipe, false)
	if err != nil {
		return err
	}
	refs, err := f.References(ReferenceOptions{Kind: "string", TargetID: &id})
	if err != nil {
		return err
	}
	if refs.Summary.Total > 0 {
		return fmt.Errorf("refuses to clear string %d because it is referenced %d time(s)", id, refs.Summary.Total)
	}
	values, err := f.scenarioStringValues()
	if err != nil {
		return err
	}
	if recipe.OldText != "" && values[id] != recipe.OldText {
		return fmt.Errorf("string id %d text %q does not match required old_text %q", id, values[id], recipe.OldText)
	}
	values[id] = ""
	return f.setScenarioStringValues(values)
}

func (f *File) planStringSlot(recipe StringRecipe, allowEmptySearch bool) (int, error) {
	values, err := f.scenarioStringValues()
	if err != nil {
		return 0, err
	}
	if recipe.ID != nil {
		id := *recipe.ID
		if id < 0 || id >= len(values) {
			return 0, fmt.Errorf("string id %d out of range 0..%d", id, len(values)-1)
		}
		return id, nil
	}
	if recipe.OldText != "" {
		found := -1
		for i, value := range values {
			if value == recipe.OldText {
				if found >= 0 {
					return 0, fmt.Errorf("old_text %q matches multiple string ids; supply id", recipe.OldText)
				}
				found = i
			}
		}
		if found >= 0 {
			return found, nil
		}
		if isTrue(recipe.Required) {
			return 0, fmt.Errorf("old_text %q not found", recipe.OldText)
		}
	}
	if allowEmptySearch {
		for i, value := range values {
			if value == "" {
				return i, nil
			}
		}
		return 0, fmt.Errorf("no empty scenario string slots available")
	}
	return 0, fmt.Errorf("%s requires id or old_text", recipe.Op)
}

func (f *File) scenarioStringValues() ([]string, error) {
	playerDataTwo := f.root.section("PlayerDataTwo")
	if playerDataTwo == nil {
		return nil, fmt.Errorf("missing PlayerDataTwo section")
	}
	values := playerDataTwo.stringList("strings")
	if len(values) == 0 {
		return nil, fmt.Errorf("missing PlayerDataTwo strings")
	}
	return values, nil
}

func (f *File) setScenarioStringValues(values []string) error {
	playerDataTwo := f.root.section("PlayerDataTwo")
	if playerDataTwo == nil {
		return fmt.Errorf("missing PlayerDataTwo section")
	}
	if err := setStringListField(playerDataTwo, "strings", "str16", values); err != nil {
		return err
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	return nil
}

func (r VariableRecipe) summary() string {
	switch r.Op {
	case "add_variable":
		if r.ID != nil {
			return fmt.Sprintf("%d:%s", *r.ID, r.Name)
		}
		return r.Name
	case "edit_variable", "tombstone_variable", "remove_variable":
		if r.TargetID != nil {
			if r.SetName != nil {
				return fmt.Sprintf("%d->%s", *r.TargetID, *r.SetName)
			}
			return fmt.Sprintf("%d", *r.TargetID)
		}
		if r.TargetName != "" {
			if r.SetName != nil {
				return fmt.Sprintf("%s->%s", r.TargetName, *r.SetName)
			}
			return r.TargetName
		}
	}
	return r.Op
}

func (f *File) AddVariable(recipe VariableRecipe) error {
	if recipe.Name == "" {
		return fmt.Errorf("add_variable requires name")
	}
	id := f.nextVariableID()
	if recipe.ID != nil {
		id = *recipe.ID
		if id < 0 {
			return fmt.Errorf("variable id %d must be non-negative", id)
		}
		if _, err := f.findVariableByID(id); err == nil {
			return fmt.Errorf("variable id %d already exists", id)
		}
	}
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	field := triggers.field("variable_data")
	if field == nil {
		return fmt.Errorf("missing variable_data")
	}
	variableSpec, err := f.variableStructSpec()
	if err != nil {
		return err
	}
	raw, err := buildStructRaw(variableSpec, map[string]any{
		"variable_id":   id,
		"variable_name": recipe.Name,
	})
	if err != nil {
		return err
	}
	parser := parser{data: raw, sections: map[string]*parsedSection{}}
	node, err := parser.parseNode("VariableStruct", variableSpec, nil)
	if err != nil {
		return err
	}
	if parser.off != len(raw) {
		return fmt.Errorf("new variable parse stopped at %d of %d", parser.off, len(raw))
	}
	field.Elements = append(field.Elements, node)
	if err := setIntField(triggers, "number_of_variables", "u32", len(field.Elements)); err != nil {
		return err
	}
	return f.refreshTriggers()
}

func (f *File) EditVariable(recipe VariableRecipe) error {
	node, err := f.findVariable(recipe)
	if err != nil {
		return err
	}
	name := recipe.Name
	if recipe.SetName != nil {
		name = *recipe.SetName
	}
	if name == "" {
		return fmt.Errorf("edit_variable requires set_name or name")
	}
	if err := setStringField(node, "variable_name", "str32", name); err != nil {
		return err
	}
	return f.refreshTriggers()
}

func (f *File) TombstoneVariable(recipe VariableRecipe) error {
	node, err := f.findVariable(recipe)
	if err != nil {
		return err
	}
	id, _ := node.intValue("variable_id")
	refs, err := f.References(ReferenceOptions{Kind: "variable", TargetID: &id})
	if err != nil {
		return err
	}
	if refs.Summary.Total > 0 {
		return fmt.Errorf("refuses to tombstone variable %d because it is referenced %d time(s)", id, refs.Summary.Total)
	}
	name := fmt.Sprintf("_DeletedVariable%d", id)
	if recipe.SetName != nil {
		name = *recipe.SetName
	}
	if name == "" {
		return fmt.Errorf("tombstone_variable set_name cannot be empty")
	}
	if err := setStringField(node, "variable_name", "str32", name); err != nil {
		return err
	}
	return f.refreshTriggers()
}

func (f *File) RemoveVariable(recipe VariableRecipe) error {
	node, index, err := f.findVariableWithIndex(recipe)
	if err != nil {
		return err
	}
	id, _ := node.intValue("variable_id")
	refs, err := f.References(ReferenceOptions{Kind: "variable", TargetID: &id})
	if err != nil {
		return err
	}
	if refs.Summary.Total > 0 {
		return fmt.Errorf("refuses to remove variable %d because it is referenced %d time(s)", id, refs.Summary.Total)
	}
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	field := triggers.field("variable_data")
	if field == nil {
		return fmt.Errorf("missing variable_data")
	}
	field.Elements = append(field.Elements[:index], field.Elements[index+1:]...)
	if err := setIntField(triggers, "number_of_variables", "u32", len(field.Elements)); err != nil {
		return err
	}
	return f.refreshTriggers()
}

func (f *File) AddUnit(recipe UnitRecipe) error {
	return f.AddUnits([]UnitRecipe{recipe})
}

func (f *File) AddUnits(recipes []UnitRecipe) error {
	if len(recipes) == 0 {
		return nil
	}
	unitsSection := f.root.section("Units")
	if unitsSection == nil {
		return fmt.Errorf("missing Units section")
	}
	playerSections := unitsSection.list("players_units")
	spec, err := f.writeSpec()
	if err != nil {
		return err
	}
	unitSpec, err := unitStructSpec(spec)
	if err != nil {
		return err
	}
	nextReferenceID := f.nextUnitReferenceID()
	for _, recipe := range recipes {
		if recipe.Op != "add_unit" {
			return fmt.Errorf("unsupported unit op %q", recipe.Op)
		}
		if recipe.Player < 0 || recipe.Player >= len(playerSections) {
			return fmt.Errorf("unit player %d out of range 0..%d", recipe.Player, len(playerSections)-1)
		}
		fallbackReferenceID := nextReferenceID
		if recipe.ReferenceID != nil && *recipe.ReferenceID >= nextReferenceID {
			nextReferenceID = *recipe.ReferenceID + 1
		} else {
			nextReferenceID++
		}
		raw, err := buildUnitFromRecipe(spec, recipe, fallbackReferenceID)
		if err != nil {
			return err
		}
		parser := parser{data: raw, sections: map[string]*parsedSection{}}
		unitNode, err := parser.parseNode("UnitStruct", unitSpec, nil)
		if err != nil {
			return err
		}
		if parser.off != len(raw) {
			return fmt.Errorf("new unit parse stopped at %d of %d", parser.off, len(raw))
		}
		playerSection := playerSections[recipe.Player]
		unitField := playerSection.field("units")
		if unitField == nil {
			return fmt.Errorf("missing units field for player %d", recipe.Player)
		}
		unitField.Elements = append(unitField.Elements, unitNode)
		if err := setIntField(playerSection, "unit_count", "u32", len(unitField.Elements)); err != nil {
			return err
		}
	}
	return f.refreshUnits()
}

func (f *File) EditUnit(recipe UnitRecipe) error {
	unitNode, playerSection, unitIndex, err := f.findUnit(recipe)
	if err != nil {
		return err
	}
	if recipe.X != nil {
		if err := setFloatField(unitNode, "x", "f32", *recipe.X); err != nil {
			return err
		}
	}
	if recipe.Y != nil {
		if err := setFloatField(unitNode, "y", "f32", *recipe.Y); err != nil {
			return err
		}
	}
	if recipe.Z != nil {
		if err := setFloatField(unitNode, "z", "f32", *recipe.Z); err != nil {
			return err
		}
	}
	if recipe.UnitConst > 0 {
		if err := setIntField(unitNode, "unit_const", "u16", recipe.UnitConst); err != nil {
			return err
		}
	}
	if recipe.Status != nil {
		if err := setIntField(unitNode, "status", "u8", *recipe.Status); err != nil {
			return err
		}
	}
	if recipe.Rotation != nil {
		if err := setFloatField(unitNode, "rotation", "f32", *recipe.Rotation); err != nil {
			return err
		}
	}
	if recipe.InitialAnimationFrame != nil {
		if err := setIntField(unitNode, "initial_animation_frame", "u16", *recipe.InitialAnimationFrame); err != nil {
			return err
		}
	}
	if recipe.GarrisonedInID != nil {
		if err := setIntField(unitNode, "garrisoned_in_id", "s32", *recipe.GarrisonedInID); err != nil {
			return err
		}
	}
	if recipe.CaptionStringID != nil {
		if err := setIntField(unitNode, "caption_string_id", "s32", *recipe.CaptionStringID); err != nil {
			return err
		}
	}
	if recipe.CaptionString != "" {
		if err := setStringField(unitNode, "caption_string", "str32", recipe.CaptionString); err != nil {
			return err
		}
	}
	if recipe.SetPlayer != nil {
		if err := f.moveUnitToPlayer(unitNode, playerSection, unitIndex, *recipe.SetPlayer); err != nil {
			return err
		}
	}
	return f.refreshUnits()
}

func (f *File) RemoveUnit(recipe UnitRecipe) error {
	unit, playerSection, unitIndex, err := f.findUnit(recipe)
	if err != nil {
		return err
	}
	if referenceID, ok := unit.intValue("reference_id"); ok && referenceID >= 0 {
		if err := f.ensureUnitNotReferenced(referenceID); err != nil {
			return err
		}
	}
	unitField := playerSection.field("units")
	if unitField == nil {
		return fmt.Errorf("missing units field")
	}
	unitField.Elements = append(unitField.Elements[:unitIndex], unitField.Elements[unitIndex+1:]...)
	if err := setIntField(playerSection, "unit_count", "u32", len(unitField.Elements)); err != nil {
		return err
	}
	return f.refreshUnits()
}

type unitAreaMatch struct {
	Unit          *parsedNode
	PlayerSection *parsedNode
	Player        int
	Index         int
}

func (f *File) RemoveUnitsInArea(recipe UnitRecipe) (int, error) {
	targets, err := f.findUnitsInArea(recipe)
	if err != nil {
		return 0, err
	}
	for _, target := range targets {
		if referenceID, ok := target.Unit.intValue("reference_id"); ok && referenceID >= 0 {
			if err := f.ensureUnitNotReferenced(referenceID); err != nil {
				return 0, err
			}
		}
	}
	byPlayer := map[int][]unitAreaMatch{}
	for _, target := range targets {
		byPlayer[target.Player] = append(byPlayer[target.Player], target)
	}
	for _, matches := range byPlayer {
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Index > matches[j].Index
		})
		playerSection := matches[0].PlayerSection
		unitField := playerSection.field("units")
		if unitField == nil {
			return 0, fmt.Errorf("missing units field for player %d", matches[0].Player)
		}
		for _, target := range matches {
			if target.Index < 0 || target.Index >= len(unitField.Elements) || unitField.Elements[target.Index] != target.Unit {
				return 0, fmt.Errorf("unit index changed before area removal for player %d index %d", target.Player, target.Index)
			}
			unitField.Elements = append(unitField.Elements[:target.Index], unitField.Elements[target.Index+1:]...)
		}
		if err := setIntField(playerSection, "unit_count", "u32", len(unitField.Elements)); err != nil {
			return 0, err
		}
	}
	if err := f.refreshUnits(); err != nil {
		return 0, err
	}
	return len(targets), nil
}

func (f *File) RemoveUnitsForPlayer(recipe UnitRecipe) (int, error) {
	targets, err := f.findUnitsByPlayer(recipe.TargetPlayer)
	if err != nil {
		return 0, err
	}
	for _, target := range targets {
		if referenceID, ok := target.Unit.intValue("reference_id"); ok && referenceID >= 0 {
			if err := f.ensureUnitNotReferenced(referenceID); err != nil {
				return 0, err
			}
		}
	}
	playerSection := targets[0].PlayerSection
	unitField := playerSection.field("units")
	if unitField == nil {
		return 0, fmt.Errorf("missing units field for player %d", targets[0].Player)
	}
	unitField.Elements = nil
	if err := setIntField(playerSection, "unit_count", "u32", 0); err != nil {
		return 0, err
	}
	if err := f.refreshUnits(); err != nil {
		return 0, err
	}
	return len(targets), nil
}

func (f *File) CopyUnitsInArea(recipe UnitRecipe) (int, error) {
	targets, err := f.findUnitsInArea(recipe)
	if err != nil {
		return 0, err
	}
	if err := f.validateCopiedUnitReferenceIDs(targets, recipe); err != nil {
		return 0, err
	}
	spec, err := f.writeSpec()
	if err != nil {
		return 0, err
	}
	refMap := copiedUnitReferenceMap(targets, copiedUnitReferenceBase(f, recipe))
	for _, target := range targets {
		add, err := copiedUnitRecipe(target, recipe, refMap)
		if err != nil {
			return 0, err
		}
		raw, err := buildUnitFromRecipe(spec, add, *add.ReferenceID)
		if err != nil {
			return 0, err
		}
		if err := f.addRawUnit(raw, add.Player); err != nil {
			return 0, err
		}
	}
	return len(targets), nil
}

func (f *File) MoveUnitsInArea(recipe UnitRecipe) (int, error) {
	targets, err := f.findUnitsInArea(recipe)
	if err != nil {
		return 0, err
	}
	if err := f.validateMoveUnitsInArea(targets, recipe); err != nil {
		return 0, err
	}
	offsetX, offsetY, err := unitAreaOffset(recipe)
	if err != nil {
		return 0, err
	}
	for _, target := range targets {
		x, _ := target.Unit.floatValue("x")
		y, _ := target.Unit.floatValue("y")
		x += offsetX
		y += offsetY
		if err := setFloatField(target.Unit, "x", "f32", x); err != nil {
			return 0, err
		}
		if err := setFloatField(target.Unit, "y", "f32", y); err != nil {
			return 0, err
		}
	}
	if err := f.refreshUnits(); err != nil {
		return 0, err
	}
	return len(targets), nil
}

func (f *File) EditUnitsInArea(recipe UnitRecipe) (int, error) {
	targets, err := f.findUnitsInArea(recipe)
	if err != nil {
		return 0, err
	}
	if err := f.validateEditUnitsInArea(recipe); err != nil {
		return 0, err
	}
	for _, target := range targets {
		if err := applyUnitAreaFieldEdits(target.Unit, recipe); err != nil {
			return 0, err
		}
	}
	if recipe.SetPlayer != nil {
		byPlayer := map[int][]unitAreaMatch{}
		for _, target := range targets {
			byPlayer[target.Player] = append(byPlayer[target.Player], target)
		}
		for _, matches := range byPlayer {
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].Index > matches[j].Index
			})
			for _, target := range matches {
				if err := f.moveUnitToPlayer(target.Unit, target.PlayerSection, target.Index, *recipe.SetPlayer); err != nil {
					return 0, err
				}
			}
		}
	}
	if err := f.refreshUnits(); err != nil {
		return 0, err
	}
	return len(targets), nil
}

func (f *File) SetTerrainRect(recipe MapRecipe) error {
	recipe.Op = "set_terrain_rect"
	return f.SetTerrain(recipe)
}

func ValidateScenarioMapSize(width, height int) error {
	if width != height {
		return fmt.Errorf("scenario map size %dx%d is non-square; kit scen blank only emits canonical square DE map-size presets", width, height)
	}
	if _, ok := ScenarioMapPresetBySize(width); !ok {
		return fmt.Errorf("scenario map size %d is not a canonical DE preset; valid sizes: %s", width, ScenarioMapPresetSizesString())
	}
	return nil
}

func ScenarioMapPresets() []ScenarioMapPreset {
	return append([]ScenarioMapPreset(nil), scenarioMapPresets...)
}

func ScenarioMapPresetBySize(size int) (ScenarioMapPreset, bool) {
	for _, preset := range scenarioMapPresets {
		if preset.Size == size {
			return preset, true
		}
	}
	return ScenarioMapPreset{}, false
}

func ScenarioMapPresetSizesString() string {
	parts := make([]string, 0, len(scenarioMapPresets))
	for _, preset := range scenarioMapPresets {
		parts = append(parts, fmt.Sprintf("%d=%s", preset.Size, preset.Name))
	}
	return strings.Join(parts, ", ")
}

func (f *File) ResizeMap(width, height int) error {
	if err := ValidateScenarioMapSize(width, height); err != nil {
		return err
	}
	mapSection := f.root.section("Map")
	if mapSection == nil {
		return fmt.Errorf("missing Map section")
	}
	oldTiles, oldWidth, oldHeight, err := f.mapTiles()
	if err != nil {
		return err
	}
	if len(oldTiles) == 0 {
		return fmt.Errorf("cannot resize map with no terrain tile template")
	}
	terrainField := mapSection.field("terrain_data")
	if terrainField == nil {
		return fmt.Errorf("missing terrain_data")
	}
	baseTile := oldTiles[0]
	newTiles := make([]*parsedNode, 0, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			src := baseTile
			if x < oldWidth && y < oldHeight {
				src = oldTiles[y*oldWidth+x]
			}
			newTiles = append(newTiles, cloneParsedNode(src))
		}
	}
	if err := setIntField(mapSection, "map_width", "s32", width); err != nil {
		return err
	}
	if err := setIntField(mapSection, "map_height", "s32", height); err != nil {
		return err
	}
	terrainField.Raw = nil
	terrainField.Elements = newTiles
	terrainField.Value = terrainField.Elements
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if mapInfo, err := f.root.mapInfo(); err == nil {
		f.Map = mapInfo
	} else {
		return err
	}
	return nil
}

func (f *File) SetTerrain(recipe MapRecipe) error {
	tiles, width, height, err := f.mapTiles()
	if err != nil {
		return err
	}
	points, err := mapPatchPoints(recipe, width, height)
	if err != nil {
		return err
	}
	if recipe.TerrainID == nil && recipe.Elevation == nil && recipe.Layer == nil {
		return fmt.Errorf("%s requires terrain_id, elevation, or layer", recipe.Op)
	}
	for _, point := range points {
		tile := tiles[point.Y*width+point.X]
		if recipe.TerrainID != nil {
			if err := setIntField(tile, "terrain_id", "u8", *recipe.TerrainID); err != nil {
				return err
			}
		}
		if recipe.Elevation != nil {
			if err := setIntField(tile, "elevation", "u8", *recipe.Elevation); err != nil {
				return err
			}
		}
		if recipe.Layer != nil {
			if err := setIntField(tile, "layer", "s16", *recipe.Layer); err != nil {
				return err
			}
		}
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if mapInfo, err := f.root.mapInfo(); err == nil {
		f.Map = mapInfo
	} else {
		return err
	}
	return nil
}

func (f *File) CopyTerrainArea(recipe MapRecipe) error {
	tiles, width, height, err := f.mapTiles()
	if err != nil {
		return err
	}
	if recipe.TargetX == nil || recipe.TargetY == nil {
		return fmt.Errorf("copy_terrain_area requires target_x and target_y")
	}
	if err := validateMapRect(recipe, width, height); err != nil {
		return err
	}
	srcWidth := recipe.X2 - recipe.X1 + 1
	srcHeight := recipe.Y2 - recipe.Y1 + 1
	target := MapRecipe{
		Op: "copy_terrain_area target",
		X1: *recipe.TargetX,
		Y1: *recipe.TargetY,
		X2: *recipe.TargetX + srcWidth - 1,
		Y2: *recipe.TargetY + srcHeight - 1,
	}
	if err := validateMapRect(target, width, height); err != nil {
		return err
	}
	snapshots := make([]terrainTileSnapshot, 0, srcWidth*srcHeight)
	for dy := 0; dy < srcHeight; dy++ {
		for dx := 0; dx < srcWidth; dx++ {
			src := tiles[(recipe.Y1+dy)*width+recipe.X1+dx]
			snapshot, err := snapshotTerrainTile(src)
			if err != nil {
				return err
			}
			snapshots = append(snapshots, snapshot)
		}
	}
	for dy := 0; dy < srcHeight; dy++ {
		for dx := 0; dx < srcWidth; dx++ {
			dst := tiles[(*recipe.TargetY+dy)*width+*recipe.TargetX+dx]
			if err := applyTerrainTileSnapshot(dst, snapshots[dy*srcWidth+dx]); err != nil {
				return err
			}
		}
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if mapInfo, err := f.root.mapInfo(); err == nil {
		f.Map = mapInfo
	} else {
		return err
	}
	return nil
}

func (f *File) AddTrigger(recipe TriggerRecipe) error {
	raw, err := f.buildTriggerRaw(recipe)
	if err != nil {
		return err
	}
	return f.addRawTrigger(raw)
}

func (f *File) buildTriggerRaw(recipe TriggerRecipe) ([]byte, error) {
	spec, err := f.writeSpec()
	if err != nil {
		return nil, err
	}
	raw, err := buildTriggerFromRecipe(spec, recipe)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (f *File) CopyTrigger(recipe TriggerRecipe) error {
	source, err := f.findTrigger(recipe)
	if err != nil {
		return err
	}
	sourceName, _ := source.stringValue("trigger_name")
	raw := source.raw()
	if err := f.addRawTrigger(raw); err != nil {
		return err
	}
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	newIndex := len(triggers.list("trigger_data")) - 1
	edit := recipe
	edit.Op = "edit_trigger"
	edit.TargetIndex = &newIndex
	edit.TargetName = ""
	if edit.SetName == nil {
		name := edit.Name
		if name == "" {
			name = strings.TrimSpace(sourceName + " copy")
			if name == "" {
				name = fmt.Sprintf("Trigger %d copy", newIndex)
			}
		}
		edit.SetName = &name
	}
	return f.EditTrigger(edit)
}

func (f *File) EditTrigger(recipe TriggerRecipe) error {
	spec, err := f.writeSpec()
	if err != nil {
		return err
	}
	trigger, err := f.findTrigger(recipe)
	if err != nil {
		return err
	}
	if recipe.Enabled != nil {
		if err := setIntField(trigger, "enabled", "u32", boolInt(*recipe.Enabled)); err != nil {
			return err
		}
	}
	if recipe.Looping != nil {
		if err := setIntField(trigger, "looping", "s8", boolInt(*recipe.Looping)); err != nil {
			return err
		}
	}
	if recipe.SetName != nil {
		if err := setStringField(trigger, "trigger_name", "str32", *recipe.SetName); err != nil {
			return err
		}
	}
	if recipe.DescriptionStringID != nil {
		if err := setIntField(trigger, "description_string_table_id", "s32", *recipe.DescriptionStringID); err != nil {
			return err
		}
	}
	if recipe.ShortDescriptionStringID != nil {
		if err := setIntField(trigger, "short_description_string_table_id", "s32", *recipe.ShortDescriptionStringID); err != nil {
			return err
		}
	}
	if err := editTriggerChildLists(spec, trigger, recipe); err != nil {
		return err
	}
	if len(recipe.Effects) > 0 {
		if err := appendEffectsToTrigger(spec, trigger, recipe.Effects); err != nil {
			return err
		}
	}
	if len(recipe.Conditions) > 0 {
		if err := appendConditionsToTrigger(spec, trigger, recipe.Conditions); err != nil {
			return err
		}
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if triggers, err := f.root.triggerInfo(); err == nil {
		f.Triggers = triggers
	} else {
		return err
	}
	return nil
}

func (f *File) TombstoneTrigger(recipe TriggerRecipe) error {
	target, err := f.findTrigger(recipe)
	if err != nil {
		return err
	}
	index, err := f.findTriggerIndex(recipe)
	if err != nil {
		return err
	}
	disabled := false
	notLooping := false
	clear := true
	name := recipe.SetName
	if name == nil {
		currentName, _ := target.stringValue("trigger_name")
		tombstoneName := fmt.Sprintf("TOMBSTONED trigger %d", index)
		if strings.TrimSpace(currentName) != "" {
			tombstoneName = "TOMBSTONED: " + currentName
		}
		name = &tombstoneName
	}
	edit := TriggerRecipe{
		Op:              "edit_trigger",
		TargetIndex:     &index,
		Enabled:         &disabled,
		Looping:         &notLooping,
		SetName:         name,
		ClearEffects:    &clear,
		ClearConditions: &clear,
	}
	return f.EditTrigger(edit)
}

func (f *File) AddCreateObjectTrigger(recipe TriggerRecipe) error {
	if recipe.ObjectListUnitID == nil {
		return fmt.Errorf("add_create_object requires object_list_unit_id")
	}
	if recipe.LocationX == nil {
		return fmt.Errorf("add_create_object requires location_x")
	}
	if recipe.LocationY == nil {
		return fmt.Errorf("add_create_object requires location_y")
	}
	spec, err := f.writeSpec()
	if err != nil {
		return err
	}
	raw, err := buildCreateObjectTrigger(spec, recipe)
	if err != nil {
		return err
	}
	return f.addRawTrigger(raw)
}

func (f *File) AddDisplayInstructionsTrigger(recipe TriggerRecipe) error {
	spec, err := f.writeSpec()
	if err != nil {
		return err
	}
	raw, err := buildDisplayInstructionTrigger(spec, recipe)
	if err != nil {
		return err
	}
	return f.addRawTrigger(raw)
}

func (f *File) addRawTrigger(raw []byte) error {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	return f.insertRawTrigger(len(triggers.list("trigger_data")), raw)
}

func (f *File) insertRawTrigger(index int, raw []byte) error {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	spec, err := f.writeSpec()
	if err != nil {
		return err
	}
	triggerSpec, _, err := triggerEffectSpecs(spec)
	if err != nil {
		return err
	}
	parser := parser{data: raw, sections: map[string]*parsedSection{}}
	triggerNode, err := parser.parseNode("TriggerStruct", triggerSpec, nil)
	if err != nil {
		return err
	}
	if parser.off != len(raw) {
		return fmt.Errorf("new trigger parse stopped at %d of %d", parser.off, len(raw))
	}
	count, _ := triggers.intValue("number_of_triggers")
	triggerData := triggers.field("trigger_data")
	if triggerData == nil {
		return fmt.Errorf("missing trigger_data")
	}
	if index < 0 || index > len(triggerData.Elements) {
		return fmt.Errorf("trigger insert index %d out of range 0..%d", index, len(triggerData.Elements))
	}
	if index < len(triggerData.Elements) {
		if err := rewriteTriggerReferencesAfterInsert(triggerData.Elements, index); err != nil {
			return err
		}
	}
	triggerData.Elements = append(triggerData.Elements, nil)
	copy(triggerData.Elements[index+1:], triggerData.Elements[index:])
	triggerNode.Start = triggerData.Start
	triggerNode.End = triggerData.Start + len(raw)
	triggerData.Elements[index] = triggerNode
	count = len(triggerData.Elements)
	if err := setIntField(triggers, "number_of_triggers", "s32", count); err != nil {
		return err
	}
	if err := setUint32ListField(triggers, "trigger_display_order_array", indexUint32List(count)); err != nil {
		return err
	}
	if options := f.root.section("Options"); options != nil {
		if err := setIntField(options, "number_of_triggers", "u32", count); err != nil {
			return err
		}
	}
	if f.headerRoot != nil {
		if err := setIntField(f.headerRoot, "trigger_count", "u32", count); err != nil {
			return err
		}
		f.header = f.headerRoot.raw()
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if triggers, err := f.root.triggerInfo(); err == nil {
		f.Triggers = triggers
	} else {
		return err
	}
	return nil
}

func (f *File) RemoveTrigger(recipe TriggerRecipe) error {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	triggerData := triggers.field("trigger_data")
	if triggerData == nil {
		return fmt.Errorf("missing trigger_data")
	}
	index, err := f.findTriggerIndex(recipe)
	if err != nil {
		return err
	}
	if index < 0 || index >= len(triggerData.Elements) {
		return fmt.Errorf("trigger index %d out of range 0..%d", index, len(triggerData.Elements)-1)
	}
	if err := rewriteTriggerReferencesAfterRemove(triggerData.Elements, index); err != nil {
		return err
	}
	triggerData.Elements = append(triggerData.Elements[:index], triggerData.Elements[index+1:]...)
	count := len(triggerData.Elements)
	if err := setIntField(triggers, "number_of_triggers", "s32", count); err != nil {
		return err
	}
	if err := setUint32ListField(triggers, "trigger_display_order_array", indexUint32List(count)); err != nil {
		return err
	}
	if options := f.root.section("Options"); options != nil {
		if err := setIntField(options, "number_of_triggers", "u32", count); err != nil {
			return err
		}
	}
	if f.headerRoot != nil {
		if err := setIntField(f.headerRoot, "trigger_count", "u32", count); err != nil {
			return err
		}
		f.header = f.headerRoot.raw()
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if triggers, err := f.root.triggerInfo(); err == nil {
		f.Triggers = triggers
	} else {
		return err
	}
	return nil
}

func (f *File) RemoveTriggers(recipe TriggerRecipe) error {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	triggerData := triggers.field("trigger_data")
	if triggerData == nil {
		return fmt.Errorf("missing trigger_data")
	}
	indexes, err := normalizeTriggerIndexes(recipe.TargetIndexes, len(triggerData.Elements))
	if err != nil {
		return err
	}
	if err := rewriteTriggerReferencesAfterBatchRemove(triggerData.Elements, indexes); err != nil {
		return err
	}
	for _, index := range descendingInts(indexes) {
		triggerData.Elements = append(triggerData.Elements[:index], triggerData.Elements[index+1:]...)
	}
	count := len(triggerData.Elements)
	if err := setIntField(triggers, "number_of_triggers", "s32", count); err != nil {
		return err
	}
	if err := setUint32ListField(triggers, "trigger_display_order_array", indexUint32List(count)); err != nil {
		return err
	}
	if options := f.root.section("Options"); options != nil {
		if err := setIntField(options, "number_of_triggers", "u32", count); err != nil {
			return err
		}
	}
	if f.headerRoot != nil {
		if err := setIntField(f.headerRoot, "trigger_count", "u32", count); err != nil {
			return err
		}
		f.header = f.headerRoot.raw()
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if triggers, err := f.root.triggerInfo(); err == nil {
		f.Triggers = triggers
	} else {
		return err
	}
	return nil
}

func (f *File) validateRemoveSystemPrefix(prefix string) (systemPrefixMatches, ReferenceReport, error) {
	matches, err := f.matchSystemPrefix(prefix)
	if err != nil {
		return systemPrefixMatches{}, ReferenceReport{}, err
	}
	externalRefs, _, err := f.systemPrefixReferences(matches)
	if err != nil {
		return systemPrefixMatches{}, ReferenceReport{}, err
	}
	return matches, externalRefs, nil
}

func (f *File) RemoveSystemPrefix(prefix string) error {
	matches, externalRefs, err := f.validateRemoveSystemPrefix(prefix)
	if err != nil {
		return err
	}
	if externalRefs.Summary.Total > 0 {
		return fmt.Errorf("refuses to remove system prefix %q because %d external reference(s) point into the matched bundle", prefix, externalRefs.Summary.Total)
	}
	if len(matches.TriggerIndexes) > 0 {
		if err := f.RemoveTriggers(TriggerRecipe{Op: "remove_triggers", TargetIndexes: matches.TriggerIndexes}); err != nil {
			return err
		}
	}
	if len(matches.UnitReferenceIDs) > 0 {
		if err := f.removeUnitsByReferenceIDs(matches.UnitReferenceIDs); err != nil {
			return err
		}
	}
	for _, id := range matches.VariableIDs {
		targetID := id
		if err := f.RemoveVariable(VariableRecipe{Op: "remove_variable", TargetID: &targetID}); err != nil {
			return err
		}
	}
	for _, id := range matches.StringIDs {
		targetID := id
		if err := f.ClearString(StringRecipe{Op: "clear_string", ID: &targetID}); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) removeUnitsByReferenceIDs(referenceIDs []int) error {
	targetSet := intSet(referenceIDs)
	if len(targetSet) == 0 {
		return nil
	}
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return err
	}
	byPlayer := map[int][]unitAreaMatch{}
	found := map[int]bool{}
	for player, playerSection := range playerSections {
		unitField := playerSection.field("units")
		if unitField == nil {
			continue
		}
		for index, unit := range unitField.Elements {
			referenceID, ok := unit.intValue("reference_id")
			if !ok || !targetSet[referenceID] {
				continue
			}
			found[referenceID] = true
			byPlayer[player] = append(byPlayer[player], unitAreaMatch{Unit: unit, PlayerSection: playerSection, Player: player, Index: index})
		}
	}
	for _, id := range referenceIDs {
		if !found[id] {
			return fmt.Errorf("system-prefix unit reference_id %d was not found at mutation time", id)
		}
	}
	for _, matches := range byPlayer {
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Index > matches[j].Index
		})
		playerSection := matches[0].PlayerSection
		unitField := playerSection.field("units")
		if unitField == nil {
			return fmt.Errorf("missing units field for player %d", matches[0].Player)
		}
		for _, target := range matches {
			if target.Index < 0 || target.Index >= len(unitField.Elements) || unitField.Elements[target.Index] != target.Unit {
				return fmt.Errorf("unit index changed before system-prefix removal for player %d index %d", target.Player, target.Index)
			}
			unitField.Elements = append(unitField.Elements[:target.Index], unitField.Elements[target.Index+1:]...)
		}
		if err := setIntField(playerSection, "unit_count", "u32", len(unitField.Elements)); err != nil {
			return err
		}
	}
	return f.refreshUnits()
}

func (f *File) validateRemoveTriggers(targetIndexes []int) (int, error) {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return 0, fmt.Errorf("missing Triggers section")
	}
	triggerData := triggers.field("trigger_data")
	if triggerData == nil {
		return 0, fmt.Errorf("missing trigger_data")
	}
	indexes, err := normalizeTriggerIndexes(targetIndexes, len(triggerData.Elements))
	if err != nil {
		return 0, err
	}
	if err := ensureNoExternalTriggerReferencesToRemoved(triggerData.Elements, indexes); err != nil {
		return 0, err
	}
	return len(indexes), nil
}

func normalizeTriggerIndexes(indexes []int, triggerCount int) ([]int, error) {
	if len(indexes) == 0 {
		return nil, fmt.Errorf("remove_triggers requires target_indexes")
	}
	out := append([]int(nil), indexes...)
	sort.Ints(out)
	write := 0
	for _, index := range out {
		if index < 0 || index >= triggerCount {
			return nil, fmt.Errorf("trigger index %d out of range 0..%d", index, triggerCount-1)
		}
		if write > 0 && out[write-1] == index {
			continue
		}
		out[write] = index
		write++
	}
	return out[:write], nil
}

func rewriteTriggerReferencesAfterRemove(triggerNodes []*parsedNode, removedIndex int) error {
	for triggerIndex, trigger := range triggerNodes {
		if triggerIndex == removedIndex {
			continue
		}
		for effectIndex, effect := range trigger.list("effect_data") {
			ref, ok := effect.intValue("trigger_id")
			if !ok || ref < 0 {
				continue
			}
			if ref == removedIndex {
				name, _ := trigger.stringValue("trigger_name")
				return fmt.Errorf("refuses to remove trigger %d because trigger %d %q effect %d references it", removedIndex, triggerIndex, name, effectIndex)
			}
		}
	}
	for triggerIndex, trigger := range triggerNodes {
		if triggerIndex == removedIndex {
			continue
		}
		for effectIndex, effect := range trigger.list("effect_data") {
			ref, ok := effect.intValue("trigger_id")
			if !ok || ref <= removedIndex {
				continue
			}
			if err := setIntField(effect, "trigger_id", "s32", ref-1); err != nil {
				return fmt.Errorf("rewrite trigger %d effect %d trigger_id: %w", triggerIndex, effectIndex, err)
			}
		}
	}
	return nil
}

func rewriteTriggerReferencesAfterInsert(triggerNodes []*parsedNode, insertedIndex int) error {
	for triggerIndex, trigger := range triggerNodes {
		for effectIndex, effect := range trigger.list("effect_data") {
			ref, ok := effect.intValue("trigger_id")
			if !ok || ref < insertedIndex {
				continue
			}
			if err := setIntField(effect, "trigger_id", "s32", ref+1); err != nil {
				return fmt.Errorf("rewrite trigger %d effect %d trigger_id after insert: %w", triggerIndex, effectIndex, err)
			}
		}
		for conditionIndex, condition := range trigger.list("condition_data") {
			ref, ok := condition.intValue("trigger_id")
			if !ok || ref < insertedIndex {
				continue
			}
			if err := setIntField(condition, "trigger_id", "s32", ref+1); err != nil {
				return fmt.Errorf("rewrite trigger %d condition %d trigger_id after insert: %w", triggerIndex, conditionIndex, err)
			}
		}
	}
	return nil
}

func rewriteTriggerReferencesAfterBatchRemove(triggerNodes []*parsedNode, removedIndexes []int) error {
	indexes, err := normalizeTriggerIndexes(removedIndexes, len(triggerNodes))
	if err != nil {
		return err
	}
	if err := ensureNoExternalTriggerReferencesToRemoved(triggerNodes, indexes); err != nil {
		return err
	}
	removed := intSet(indexes)
	for triggerIndex, trigger := range triggerNodes {
		if removed[triggerIndex] {
			continue
		}
		for effectIndex, effect := range trigger.list("effect_data") {
			ref, ok := effect.intValue("trigger_id")
			if !ok || ref < 0 {
				continue
			}
			shift := countIntsLessThan(indexes, ref)
			if shift == 0 {
				continue
			}
			if err := setIntField(effect, "trigger_id", "s32", ref-shift); err != nil {
				return fmt.Errorf("rewrite trigger %d effect %d trigger_id: %w", triggerIndex, effectIndex, err)
			}
		}
	}
	return nil
}

func triggerHasXSCarrier(trigger *parsedNode) bool {
	if trigger == nil {
		return false
	}
	for _, effect := range trigger.list("effect_data") {
		effectType, _ := effect.intValue("effect_type")
		if effectType != effectTypeID("script_call") {
			continue
		}
		message, _ := effect.stringValue("message")
		if isXSCarrierMessage(message) {
			return true
		}
	}
	return false
}

func ensureNoExternalTriggerReferencesToRemoved(triggerNodes []*parsedNode, removedIndexes []int) error {
	removed := intSet(removedIndexes)
	for triggerIndex, trigger := range triggerNodes {
		if removed[triggerIndex] {
			continue
		}
		for effectIndex, effect := range trigger.list("effect_data") {
			ref, ok := effect.intValue("trigger_id")
			if !ok || ref < 0 || !removed[ref] {
				continue
			}
			name, _ := trigger.stringValue("trigger_name")
			return fmt.Errorf("refuses to remove trigger batch because trigger %d %q effect %d references removed trigger %d", triggerIndex, name, effectIndex, ref)
		}
	}
	return nil
}

func intSet(values []int) map[int]bool {
	out := make(map[int]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}

func countIntsLessThan(values []int, target int) int {
	count := 0
	for _, value := range values {
		if value >= target {
			break
		}
		count++
	}
	return count
}

func (f *File) ensureUnitNotReferenced(referenceID int) error {
	units := f.root.section("Units")
	if units != nil {
		for player, playerSection := range units.list("players_units") {
			for unitIndex, unit := range playerSection.list("units") {
				ref, ok := unit.intValue("garrisoned_in_id")
				if ok && ref == referenceID {
					sourceRef, _ := unit.intValue("reference_id")
					return fmt.Errorf("refuses to remove unit reference_id %d because player %d unit %d reference_id %d field garrisoned_in_id references it", referenceID, player, unitIndex, sourceRef)
				}
			}
		}
	}
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return nil
	}
	triggerData := triggers.field("trigger_data")
	if triggerData == nil {
		return nil
	}
	for triggerIndex, trigger := range triggerData.Elements {
		triggerName, _ := trigger.stringValue("trigger_name")
		for conditionIndex, condition := range trigger.list("condition_data") {
			for _, field := range []string{"unit_object", "next_object"} {
				ref, ok := condition.intValue(field)
				if ok && ref == referenceID {
					return fmt.Errorf("refuses to remove unit reference_id %d because trigger %d %q condition %d field %s references it", referenceID, triggerIndex, triggerName, conditionIndex, field)
				}
			}
		}
		for effectIndex, effect := range trigger.list("effect_data") {
			for _, field := range []string{"legacy_location_object_reference", "location_object_reference"} {
				ref, ok := effect.intValue(field)
				if ok && ref == referenceID {
					return fmt.Errorf("refuses to remove unit reference_id %d because trigger %d %q effect %d field %s references it", referenceID, triggerIndex, triggerName, effectIndex, field)
				}
			}
			for selectedIndex, ref := range effect.intList("selected_object_ids") {
				if ref == referenceID {
					return fmt.Errorf("refuses to remove unit reference_id %d because trigger %d %q effect %d selected_object_ids[%d] references it", referenceID, triggerIndex, triggerName, effectIndex, selectedIndex)
				}
			}
		}
	}
	return nil
}

func (f *File) ClearTriggers() error {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return fmt.Errorf("missing Triggers section")
	}
	triggerData := triggers.field("trigger_data")
	if triggerData == nil {
		return fmt.Errorf("missing trigger_data")
	}
	triggerData.Elements = nil
	if err := setIntField(triggers, "number_of_triggers", "s32", 0); err != nil {
		return err
	}
	if err := setUint32ListField(triggers, "trigger_display_order_array", nil); err != nil {
		return err
	}
	if options := f.root.section("Options"); options != nil {
		if err := setIntField(options, "number_of_triggers", "u32", 0); err != nil {
			return err
		}
	}
	if f.headerRoot != nil {
		if err := setIntField(f.headerRoot, "trigger_count", "u32", 0); err != nil {
			return err
		}
		f.header = f.headerRoot.raw()
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if triggers, err := f.root.triggerInfo(); err == nil {
		f.Triggers = triggers
	} else {
		return err
	}
	return nil
}

func (f *File) addRawUnit(raw []byte, player int) error {
	unitsSection := f.root.section("Units")
	if unitsSection == nil {
		return fmt.Errorf("missing Units section")
	}
	playerSections := unitsSection.list("players_units")
	if player < 0 || player >= len(playerSections) {
		return fmt.Errorf("unit player %d out of range 0..%d", player, len(playerSections)-1)
	}
	spec, err := f.writeSpec()
	if err != nil {
		return err
	}
	unitSpec, err := unitStructSpec(spec)
	if err != nil {
		return err
	}
	parser := parser{data: raw, sections: map[string]*parsedSection{}}
	unitNode, err := parser.parseNode("UnitStruct", unitSpec, nil)
	if err != nil {
		return err
	}
	if parser.off != len(raw) {
		return fmt.Errorf("new unit parse stopped at %d of %d", parser.off, len(raw))
	}
	playerSection := playerSections[player]
	unitField := playerSection.field("units")
	if unitField == nil {
		return fmt.Errorf("missing units field for player %d", player)
	}
	unitField.Elements = append(unitField.Elements, unitNode)
	if err := setIntField(playerSection, "unit_count", "u32", len(unitField.Elements)); err != nil {
		return err
	}
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	if units, err := f.root.unitInfo(); err == nil {
		f.Units = units
	} else {
		return err
	}
	return nil
}

func (f *File) findUnitByReferenceID(referenceID int) (*parsedNode, *parsedNode, int, error) {
	unitsSection := f.root.section("Units")
	if unitsSection == nil {
		return nil, nil, 0, fmt.Errorf("missing Units section")
	}
	playerSections := unitsSection.list("players_units")
	for _, playerSection := range playerSections {
		unitField := playerSection.field("units")
		if unitField == nil {
			continue
		}
		for i, unit := range unitField.Elements {
			refID, _ := unit.intValue("reference_id")
			if refID == referenceID {
				return unit, playerSection, i, nil
			}
		}
	}
	return nil, nil, 0, fmt.Errorf("unit reference_id %d not found", referenceID)
}

func (f *File) findUnit(recipe UnitRecipe) (*parsedNode, *parsedNode, int, error) {
	if recipe.ReferenceID != nil {
		return f.findUnitByReferenceID(*recipe.ReferenceID)
	}
	if recipe.TargetIndex != nil {
		if recipe.TargetPlayer == nil {
			return nil, nil, 0, fmt.Errorf("%s target_index requires target_player", recipe.Op)
		}
		return f.findUnitByPlayerIndex(*recipe.TargetPlayer, *recipe.TargetIndex)
	}
	if recipe.TargetCaption != "" {
		return f.findUnitByCaption(recipe.TargetCaption, recipe.TargetPlayer)
	}
	return nil, nil, 0, fmt.Errorf("%s requires reference_id, target_index+target_player, or unique target_caption", recipe.Op)
}

func (f *File) findUnitsInArea(recipe UnitRecipe) ([]unitAreaMatch, error) {
	if recipe.TargetAreaX1 == nil || recipe.TargetAreaY1 == nil || recipe.TargetAreaX2 == nil || recipe.TargetAreaY2 == nil {
		return nil, fmt.Errorf("%s requires target_area_x1, target_area_y1, target_area_x2, and target_area_y2", unitAreaOpName(recipe))
	}
	x1, x2 := *recipe.TargetAreaX1, *recipe.TargetAreaX2
	y1, y2 := *recipe.TargetAreaY1, *recipe.TargetAreaY2
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return nil, err
	}
	var matches []unitAreaMatch
	for player, playerSection := range playerSections {
		if recipe.TargetPlayer != nil && player != *recipe.TargetPlayer {
			continue
		}
		unitField := playerSection.field("units")
		if unitField == nil {
			continue
		}
		for index, unit := range unitField.Elements {
			if recipe.TargetUnitConst != nil {
				unitConst, ok := unit.intValue("unit_const")
				if !ok || unitConst != *recipe.TargetUnitConst {
					continue
				}
			}
			x, okX := unit.floatValue("x")
			y, okY := unit.floatValue("y")
			if !okX || !okY {
				continue
			}
			if x < x1 || x > x2 || y < y1 || y > y2 {
				continue
			}
			matches = append(matches, unitAreaMatch{Unit: unit, PlayerSection: playerSection, Player: player, Index: index})
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("%s matched no units", unitAreaOpName(recipe))
	}
	return matches, nil
}

func unitAreaOpName(recipe UnitRecipe) string {
	if recipe.Op != "" {
		return recipe.Op
	}
	return "unit area operation"
}

func (f *File) validateMoveUnitsInArea(targets []unitAreaMatch, recipe UnitRecipe) error {
	if recipe.OffsetX == nil && recipe.OffsetY == nil && recipe.TargetX == nil && recipe.TargetY == nil {
		return fmt.Errorf("move_units_in_area requires offset_x/offset_y or target_x/target_y")
	}
	offsetX, offsetY, err := unitAreaOffset(recipe)
	if err != nil {
		return err
	}
	_, width, height, err := f.mapTiles()
	if err != nil {
		return err
	}
	for _, target := range targets {
		x, okX := target.Unit.floatValue("x")
		y, okY := target.Unit.floatValue("y")
		if !okX || !okY {
			return fmt.Errorf("move_units_in_area source player %d unit %d has no x/y", target.Player, target.Index)
		}
		x += offsetX
		y += offsetY
		if x < 0 || y < 0 || x > float64(width) || y > float64(height) {
			return fmt.Errorf("move_units_in_area would move player %d unit %d to %.2f,%.2f outside map %dx%d", target.Player, target.Index, x, y, width, height)
		}
	}
	return nil
}

func (f *File) validateEditUnitsInArea(recipe UnitRecipe) error {
	if recipe.X != nil || recipe.Y != nil || recipe.Z != nil || recipe.OffsetX != nil || recipe.OffsetY != nil || recipe.TargetX != nil || recipe.TargetY != nil {
		return fmt.Errorf("edit_units_in_area does not edit positions; use edit_unit or move_units_in_area")
	}
	if recipe.UnitConst <= 0 &&
		recipe.Status == nil &&
		recipe.Rotation == nil &&
		recipe.InitialAnimationFrame == nil &&
		recipe.GarrisonedInID == nil &&
		recipe.CaptionStringID == nil &&
		recipe.CaptionString == "" &&
		recipe.SetPlayer == nil {
		return fmt.Errorf("edit_units_in_area requires unit_const, status, rotation, initial_animation_frame, garrisoned_in_id, caption_string_id, caption_string, or set_player")
	}
	if recipe.SetPlayer != nil {
		playerSections, err := f.unitPlayerSections()
		if err != nil {
			return err
		}
		if *recipe.SetPlayer < 0 || *recipe.SetPlayer >= len(playerSections) {
			return fmt.Errorf("set_player %d out of range 0..%d", *recipe.SetPlayer, len(playerSections)-1)
		}
	}
	return nil
}

func unitAreaOffset(recipe UnitRecipe) (float64, float64, error) {
	hasOffset := recipe.OffsetX != nil || recipe.OffsetY != nil
	hasTarget := recipe.TargetX != nil || recipe.TargetY != nil
	if hasOffset && hasTarget {
		return 0, 0, fmt.Errorf("%s cannot combine offset_x/offset_y with target_x/target_y", unitAreaOpName(recipe))
	}
	if hasTarget {
		if recipe.TargetX == nil || recipe.TargetY == nil {
			return 0, 0, fmt.Errorf("%s target placement requires both target_x and target_y", unitAreaOpName(recipe))
		}
		if recipe.TargetAreaX1 == nil || recipe.TargetAreaY1 == nil || recipe.TargetAreaX2 == nil || recipe.TargetAreaY2 == nil {
			return 0, 0, fmt.Errorf("%s target placement requires source target_area bounds", unitAreaOpName(recipe))
		}
		x1 := *recipe.TargetAreaX1
		y1 := *recipe.TargetAreaY1
		if *recipe.TargetAreaX2 < x1 {
			x1 = *recipe.TargetAreaX2
		}
		if *recipe.TargetAreaY2 < y1 {
			y1 = *recipe.TargetAreaY2
		}
		return *recipe.TargetX - x1, *recipe.TargetY - y1, nil
	}
	offsetX := 0.0
	offsetY := 0.0
	if recipe.OffsetX != nil {
		offsetX = *recipe.OffsetX
	}
	if recipe.OffsetY != nil {
		offsetY = *recipe.OffsetY
	}
	return offsetX, offsetY, nil
}

func applyUnitAreaFieldEdits(unitNode *parsedNode, recipe UnitRecipe) error {
	if recipe.UnitConst > 0 {
		if err := setIntField(unitNode, "unit_const", "u16", recipe.UnitConst); err != nil {
			return err
		}
	}
	if recipe.Status != nil {
		if err := setIntField(unitNode, "status", "u8", *recipe.Status); err != nil {
			return err
		}
	}
	if recipe.Rotation != nil {
		if err := setFloatField(unitNode, "rotation", "f32", *recipe.Rotation); err != nil {
			return err
		}
	}
	if recipe.InitialAnimationFrame != nil {
		if err := setIntField(unitNode, "initial_animation_frame", "u16", *recipe.InitialAnimationFrame); err != nil {
			return err
		}
	}
	if recipe.GarrisonedInID != nil {
		if err := setIntField(unitNode, "garrisoned_in_id", "s32", *recipe.GarrisonedInID); err != nil {
			return err
		}
	}
	if recipe.CaptionStringID != nil {
		if err := setIntField(unitNode, "caption_string_id", "s32", *recipe.CaptionStringID); err != nil {
			return err
		}
	}
	if recipe.CaptionString != "" {
		if err := setStringField(unitNode, "caption_string", "str32", recipe.CaptionString); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) validateCopiedUnitReferenceIDs(targets []unitAreaMatch, recipe UnitRecipe) error {
	if _, _, err := unitAreaOffset(recipe); err != nil {
		return err
	}
	base := copiedUnitReferenceBase(f, recipe)
	if base <= 0 {
		return fmt.Errorf("copy_units_in_area reference_id_base %d must be positive", base)
	}
	seen := map[int]bool{}
	for i := range targets {
		id := base + i
		if seen[id] {
			return fmt.Errorf("copy_units_in_area duplicate generated reference_id %d", id)
		}
		seen[id] = true
		if _, _, _, err := f.findUnitByReferenceID(id); err == nil {
			return fmt.Errorf("copy_units_in_area generated reference_id %d already exists", id)
		}
	}
	if recipe.SetPlayer != nil {
		playerSections, err := f.unitPlayerSections()
		if err != nil {
			return err
		}
		if *recipe.SetPlayer < 0 || *recipe.SetPlayer >= len(playerSections) {
			return fmt.Errorf("set_player %d out of range 0..%d", *recipe.SetPlayer, len(playerSections)-1)
		}
	}
	return nil
}

func copiedUnitReferenceBase(f *File, recipe UnitRecipe) int {
	if recipe.ReferenceIDBase != nil {
		return *recipe.ReferenceIDBase
	}
	return f.nextUnitReferenceID()
}

func copiedUnitReferenceMap(targets []unitAreaMatch, base int) map[int]int {
	refMap := map[int]int{}
	for i, target := range targets {
		if oldRef, ok := target.Unit.intValue("reference_id"); ok && oldRef >= 0 {
			refMap[oldRef] = base + i
		}
	}
	return refMap
}

func copiedUnitRecipe(target unitAreaMatch, recipe UnitRecipe, refMap map[int]int) (UnitRecipe, error) {
	unitConst, ok := target.Unit.intValue("unit_const")
	if !ok || unitConst <= 0 {
		return UnitRecipe{}, fmt.Errorf("copy_units_in_area source player %d unit %d has invalid unit_const", target.Player, target.Index)
	}
	x, okX := target.Unit.floatValue("x")
	y, okY := target.Unit.floatValue("y")
	z, _ := target.Unit.floatValue("z")
	if !okX || !okY {
		return UnitRecipe{}, fmt.Errorf("copy_units_in_area source player %d unit %d has no x/y", target.Player, target.Index)
	}
	offsetX, offsetY, err := unitAreaOffset(recipe)
	if err != nil {
		return UnitRecipe{}, err
	}
	x += offsetX
	y += offsetY
	player := target.Player
	if recipe.SetPlayer != nil {
		player = *recipe.SetPlayer
	}
	oldRef, ok := target.Unit.intValue("reference_id")
	if !ok || oldRef < 0 {
		return UnitRecipe{}, fmt.Errorf("copy_units_in_area source player %d unit %d has no non-negative reference_id", target.Player, target.Index)
	}
	newRef := refMap[oldRef]
	status, _ := target.Unit.intValue("status")
	rotation, _ := target.Unit.floatValue("rotation")
	initialFrame, _ := target.Unit.intValue("initial_animation_frame")
	garrisonedInID, ok := target.Unit.intValue("garrisoned_in_id")
	if !ok {
		garrisonedInID = -1
	}
	if mapped, ok := refMap[garrisonedInID]; ok {
		garrisonedInID = mapped
	} else if garrisonedInID >= 0 {
		garrisonedInID = -1
	}
	captionStringID, ok := target.Unit.intValue("caption_string_id")
	if !ok {
		captionStringID = -1
	}
	captionString, _ := target.Unit.stringValue("caption_string")
	if recipe.CaptionSuffix != "" {
		captionString += recipe.CaptionSuffix
	}
	return UnitRecipe{
		Op:                    "add_unit",
		Player:                player,
		UnitConst:             unitConst,
		X:                     &x,
		Y:                     &y,
		Z:                     &z,
		ReferenceID:           &newRef,
		Status:                &status,
		Rotation:              &rotation,
		InitialAnimationFrame: &initialFrame,
		GarrisonedInID:        &garrisonedInID,
		CaptionStringID:       &captionStringID,
		CaptionString:         captionString,
	}, nil
}

func (f *File) findUnitByPlayerIndex(player, index int) (*parsedNode, *parsedNode, int, error) {
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return nil, nil, 0, err
	}
	if player < 0 || player >= len(playerSections) {
		return nil, nil, 0, fmt.Errorf("unit target_player %d out of range 0..%d", player, len(playerSections)-1)
	}
	playerSection := playerSections[player]
	unitField := playerSection.field("units")
	if unitField == nil {
		return nil, nil, 0, fmt.Errorf("missing units field for player %d", player)
	}
	if index < 0 || index >= len(unitField.Elements) {
		return nil, nil, 0, fmt.Errorf("unit target_index %d out of range for player %d units 0..%d", index, player, len(unitField.Elements)-1)
	}
	return unitField.Elements[index], playerSection, index, nil
}

func (f *File) findUnitByCaption(caption string, targetPlayer *int) (*parsedNode, *parsedNode, int, error) {
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return nil, nil, 0, err
	}
	var foundUnit *parsedNode
	var foundSection *parsedNode
	foundIndex := -1
	matches := 0
	for player, playerSection := range playerSections {
		if targetPlayer != nil && player != *targetPlayer {
			continue
		}
		unitField := playerSection.field("units")
		if unitField == nil {
			continue
		}
		for i, unit := range unitField.Elements {
			got, _ := unit.stringValue("caption_string")
			if got != caption {
				continue
			}
			matches++
			foundUnit = unit
			foundSection = playerSection
			foundIndex = i
		}
	}
	if matches == 0 {
		if targetPlayer != nil {
			return nil, nil, 0, fmt.Errorf("unit target_caption %q not found for player %d", caption, *targetPlayer)
		}
		return nil, nil, 0, fmt.Errorf("unit target_caption %q not found", caption)
	}
	if matches > 1 {
		return nil, nil, 0, fmt.Errorf("unit target_caption %q matched %d units; use reference_id or target_player+target_index", caption, matches)
	}
	return foundUnit, foundSection, foundIndex, nil
}

func (f *File) findUnitsByCaptionPrefix(prefix string, targetPlayer *int) ([]unitAreaMatch, error) {
	return f.findUnitsByCaptionText(prefix, targetPlayer, true)
}

func (f *File) findUnitsByCaptionText(text string, targetPlayer *int, prefixOnly bool) ([]unitAreaMatch, error) {
	if strings.TrimSpace(text) == "" {
		if prefixOnly {
			return nil, fmt.Errorf("unit caption prefix must be non-empty")
		}
		return nil, fmt.Errorf("unit caption contains marker must be non-empty")
	}
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return nil, err
	}
	var matches []unitAreaMatch
	for player, playerSection := range playerSections {
		if targetPlayer != nil && player != *targetPlayer {
			continue
		}
		unitField := playerSection.field("units")
		if unitField == nil {
			continue
		}
		for index, unit := range unitField.Elements {
			caption, _ := unit.stringValue("caption_string")
			matched := strings.Contains(caption, text)
			if prefixOnly {
				matched = strings.HasPrefix(caption, text)
			}
			if matched {
				matches = append(matches, unitAreaMatch{Unit: unit, PlayerSection: playerSection, Player: player, Index: index})
			}
		}
	}
	if len(matches) == 0 {
		if targetPlayer != nil {
			if prefixOnly {
				return nil, fmt.Errorf("unit caption prefix %q matched no units for player %d", text, *targetPlayer)
			}
			return nil, fmt.Errorf("unit caption contains marker %q matched no units for player %d", text, *targetPlayer)
		}
		if prefixOnly {
			return nil, fmt.Errorf("unit caption prefix %q matched no units", text)
		}
		return nil, fmt.Errorf("unit caption contains marker %q matched no units", text)
	}
	return matches, nil
}

func (f *File) findUnitsByUnitConst(unitConst int, targetPlayer *int) ([]unitAreaMatch, error) {
	if unitConst < 0 {
		return nil, fmt.Errorf("unit const must be non-negative")
	}
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return nil, err
	}
	var matches []unitAreaMatch
	for player, playerSection := range playerSections {
		if targetPlayer != nil && player != *targetPlayer {
			continue
		}
		unitField := playerSection.field("units")
		if unitField == nil {
			continue
		}
		for index, unit := range unitField.Elements {
			got, ok := unit.intValue("unit_const")
			if !ok || got != unitConst {
				continue
			}
			matches = append(matches, unitAreaMatch{Unit: unit, PlayerSection: playerSection, Player: player, Index: index})
		}
	}
	if len(matches) == 0 {
		if targetPlayer != nil {
			return nil, fmt.Errorf("unit const %d matched no units for player %d", unitConst, *targetPlayer)
		}
		return nil, fmt.Errorf("unit const %d matched no units", unitConst)
	}
	return matches, nil
}

func (f *File) findUnitsByPlayer(targetPlayer *int) ([]unitAreaMatch, error) {
	if targetPlayer == nil {
		return nil, fmt.Errorf("unit player selector requires target_player")
	}
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return nil, err
	}
	player := *targetPlayer
	if player < 0 || player >= len(playerSections) {
		return nil, fmt.Errorf("unit target_player %d out of range 0..%d", player, len(playerSections)-1)
	}
	playerSection := playerSections[player]
	unitField := playerSection.field("units")
	if unitField == nil {
		return nil, fmt.Errorf("missing units field for player %d", player)
	}
	matches := make([]unitAreaMatch, 0, len(unitField.Elements))
	for index, unit := range unitField.Elements {
		matches = append(matches, unitAreaMatch{Unit: unit, PlayerSection: playerSection, Player: player, Index: index})
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("player %d has no placed units", player)
	}
	return matches, nil
}

func (f *File) unitPlayerSections() ([]*parsedNode, error) {
	unitsSection := f.root.section("Units")
	if unitsSection == nil {
		return nil, fmt.Errorf("missing Units section")
	}
	playerSections := unitsSection.list("players_units")
	if len(playerSections) == 0 {
		return nil, fmt.Errorf("missing players_units")
	}
	return playerSections, nil
}

func (f *File) moveUnitToPlayer(unitNode, fromSection *parsedNode, fromIndex, toPlayer int) error {
	playerSections, err := f.unitPlayerSections()
	if err != nil {
		return err
	}
	if toPlayer < 0 || toPlayer >= len(playerSections) {
		return fmt.Errorf("set_player %d out of range 0..%d", toPlayer, len(playerSections)-1)
	}
	toSection := playerSections[toPlayer]
	if fromSection == toSection {
		return nil
	}
	fromField := fromSection.field("units")
	if fromField == nil {
		return fmt.Errorf("missing source units field")
	}
	toField := toSection.field("units")
	if toField == nil {
		return fmt.Errorf("missing destination units field for player %d", toPlayer)
	}
	if fromIndex < 0 || fromIndex >= len(fromField.Elements) || fromField.Elements[fromIndex] != unitNode {
		return fmt.Errorf("source unit index changed before set_player")
	}
	fromField.Elements = append(fromField.Elements[:fromIndex], fromField.Elements[fromIndex+1:]...)
	toField.Elements = append(toField.Elements, unitNode)
	if err := setIntField(fromSection, "unit_count", "u32", len(fromField.Elements)); err != nil {
		return err
	}
	if err := setIntField(toSection, "unit_count", "u32", len(toField.Elements)); err != nil {
		return err
	}
	return nil
}

func (f *File) refreshUnits() error {
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	units, err := f.root.unitInfo()
	if err != nil {
		return err
	}
	f.Units = units
	return nil
}

func (f *File) nextUnitReferenceID() int {
	maxID := 0
	if f.Units == nil {
		return 1
	}
	for _, section := range f.Units.Sections {
		for _, unit := range section.Units {
			if unit.ReferenceID > maxID {
				maxID = unit.ReferenceID
			}
		}
	}
	return maxID + 1
}

func (f *File) mapTiles() ([]*parsedNode, int, int, error) {
	mapSection := f.root.section("Map")
	if mapSection == nil {
		return nil, 0, 0, fmt.Errorf("missing Map section")
	}
	width, _ := mapSection.intValue("map_width")
	height, _ := mapSection.intValue("map_height")
	tiles := mapSection.list("terrain_data")
	if width <= 0 || height <= 0 {
		return nil, 0, 0, fmt.Errorf("invalid map size %dx%d", width, height)
	}
	if len(tiles) != width*height {
		return nil, 0, 0, fmt.Errorf("terrain_data has %d tiles, want %d", len(tiles), width*height)
	}
	return tiles, width, height, nil
}

type terrainTileSnapshot map[string]terrainFieldSnapshot

type terrainFieldSnapshot struct {
	Raw   []byte
	Value any
}

func snapshotTerrainTile(src *parsedNode) (terrainTileSnapshot, error) {
	snapshot := terrainTileSnapshot{}
	for _, name := range []string{"terrain_id", "elevation", "layer", "unused"} {
		srcField := src.field(name)
		if srcField == nil {
			return nil, fmt.Errorf("terrain tile missing %s field", name)
		}
		snapshot[name] = terrainFieldSnapshot{
			Raw:   append([]byte(nil), srcField.Raw...),
			Value: srcField.Value,
		}
	}
	return snapshot, nil
}

func applyTerrainTileSnapshot(dst *parsedNode, snapshot terrainTileSnapshot) error {
	for name, srcField := range snapshot {
		dstField := dst.field(name)
		if dstField == nil {
			return fmt.Errorf("terrain tile missing %s field", name)
		}
		dstField.Raw = append([]byte(nil), srcField.Raw...)
		dstField.Value = srcField.Value
	}
	return nil
}

func cloneParsedNode(src *parsedNode) *parsedNode {
	if src == nil {
		return nil
	}
	dst := &parsedNode{
		Name:  src.Name,
		Start: src.Start,
		End:   src.End,
		Raw:   append([]byte(nil), src.Raw...),
		Value: src.Value,
	}
	if len(src.Fields) > 0 {
		dst.Fields = make([]*parsedNode, 0, len(src.Fields))
		for _, field := range src.Fields {
			dst.Fields = append(dst.Fields, cloneParsedNode(field))
		}
	}
	if len(src.Elements) > 0 {
		dst.Elements = make([]*parsedNode, 0, len(src.Elements))
		for _, elem := range src.Elements {
			dst.Elements = append(dst.Elements, cloneParsedNode(elem))
		}
		dst.Value = dst.Elements
	}
	return dst
}

func (f *File) countMapTiles(recipe MapRecipe) (int, error) {
	_, width, height, err := f.mapTiles()
	if err != nil {
		return 0, err
	}
	points, err := mapPatchPoints(recipe, width, height)
	if err != nil {
		return 0, err
	}
	return len(points), nil
}

func validateMapRect(recipe MapRecipe, width, height int) error {
	if recipe.X1 > recipe.X2 || recipe.Y1 > recipe.Y2 {
		return fmt.Errorf("bad rectangle (%d,%d)-(%d,%d)", recipe.X1, recipe.Y1, recipe.X2, recipe.Y2)
	}
	if recipe.X1 < 0 || recipe.Y1 < 0 || recipe.X2 >= width || recipe.Y2 >= height {
		return fmt.Errorf("rectangle (%d,%d)-(%d,%d) outside map %dx%d", recipe.X1, recipe.Y1, recipe.X2, recipe.Y2, width, height)
	}
	return nil
}

func plannedMapTiles(recipe Recipe) int {
	total := 0
	for _, mapPatch := range recipe.Map {
		switch mapPatch.Op {
		case "set_terrain_rect":
			if mapPatch.X1 <= mapPatch.X2 && mapPatch.Y1 <= mapPatch.Y2 {
				total += (mapPatch.X2 - mapPatch.X1 + 1) * (mapPatch.Y2 - mapPatch.Y1 + 1)
			}
		case "set_terrain_circle":
			if mapPatch.Radius != nil && *mapPatch.Radius >= 0 {
				r := *mapPatch.Radius
				for dy := -r; dy <= r; dy++ {
					for dx := -r; dx <= r; dx++ {
						if dx*dx+dy*dy <= r*r {
							total++
						}
					}
				}
			}
		case "set_terrain_line":
			total += linePointEstimate(mapPatch.X1, mapPatch.Y1, mapPatch.X2, mapPatch.Y2)
		case "set_terrain_border":
			if mapPatch.X1 <= mapPatch.X2 && mapPatch.Y1 <= mapPatch.Y2 {
				thickness := 1
				if mapPatch.Thickness != nil {
					thickness = *mapPatch.Thickness
				}
				if thickness > 0 {
					width := mapPatch.X2 - mapPatch.X1 + 1
					height := mapPatch.Y2 - mapPatch.Y1 + 1
					innerWidth := width - 2*thickness
					if innerWidth < 0 {
						innerWidth = 0
					}
					innerHeight := height - 2*thickness
					if innerHeight < 0 {
						innerHeight = 0
					}
					total += width*height - innerWidth*innerHeight
				}
			}
		case "copy_terrain_area":
			if mapPatch.X1 <= mapPatch.X2 && mapPatch.Y1 <= mapPatch.Y2 {
				total += (mapPatch.X2 - mapPatch.X1 + 1) * (mapPatch.Y2 - mapPatch.Y1 + 1)
			}
		}
	}
	return total
}

type mapPoint struct {
	X int
	Y int
}

func mapPatchPoints(recipe MapRecipe, width, height int) ([]mapPoint, error) {
	switch recipe.Op {
	case "set_terrain_rect", "":
		if recipe.Op == "" {
			recipe.Op = "set_terrain_rect"
		}
		if err := validateMapRect(recipe, width, height); err != nil {
			return nil, err
		}
		var points []mapPoint
		for y := recipe.Y1; y <= recipe.Y2; y++ {
			for x := recipe.X1; x <= recipe.X2; x++ {
				points = append(points, mapPoint{X: x, Y: y})
			}
		}
		return points, nil
	case "set_terrain_circle":
		if recipe.Radius == nil {
			return nil, fmt.Errorf("set_terrain_circle requires radius")
		}
		if *recipe.Radius < 0 {
			return nil, fmt.Errorf("set_terrain_circle radius must be >= 0")
		}
		r := *recipe.Radius
		var points []mapPoint
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if dx*dx+dy*dy <= r*r {
					points = append(points, mapPoint{X: recipe.X1 + dx, Y: recipe.Y1 + dy})
				}
			}
		}
		return validateMapPoints(recipe.Op, uniqueSortedMapPoints(points), width, height)
	case "set_terrain_line":
		points := bresenhamLine(recipe.X1, recipe.Y1, recipe.X2, recipe.Y2)
		return validateMapPoints(recipe.Op, points, width, height)
	case "set_terrain_border":
		if err := validateMapRect(recipe, width, height); err != nil {
			return nil, err
		}
		thickness := 1
		if recipe.Thickness != nil {
			thickness = *recipe.Thickness
		}
		if thickness <= 0 {
			return nil, fmt.Errorf("set_terrain_border thickness must be > 0")
		}
		var points []mapPoint
		for y := recipe.Y1; y <= recipe.Y2; y++ {
			for x := recipe.X1; x <= recipe.X2; x++ {
				left := x - recipe.X1
				right := recipe.X2 - x
				top := y - recipe.Y1
				bottom := recipe.Y2 - y
				if left < thickness || right < thickness || top < thickness || bottom < thickness {
					points = append(points, mapPoint{X: x, Y: y})
				}
			}
		}
		return points, nil
	case "copy_terrain_area":
		if recipe.TargetX == nil || recipe.TargetY == nil {
			return nil, fmt.Errorf("copy_terrain_area requires target_x and target_y")
		}
		if err := validateMapRect(recipe, width, height); err != nil {
			return nil, err
		}
		srcWidth := recipe.X2 - recipe.X1 + 1
		srcHeight := recipe.Y2 - recipe.Y1 + 1
		target := MapRecipe{
			Op: "copy_terrain_area target",
			X1: *recipe.TargetX,
			Y1: *recipe.TargetY,
			X2: *recipe.TargetX + srcWidth - 1,
			Y2: *recipe.TargetY + srcHeight - 1,
		}
		if err := validateMapRect(target, width, height); err != nil {
			return nil, err
		}
		var points []mapPoint
		for y := recipe.Y1; y <= recipe.Y2; y++ {
			for x := recipe.X1; x <= recipe.X2; x++ {
				points = append(points, mapPoint{X: x, Y: y})
			}
		}
		return points, nil
	default:
		return nil, fmt.Errorf("unsupported map op %q", recipe.Op)
	}
}

func validateMapPoints(op string, points []mapPoint, width, height int) ([]mapPoint, error) {
	if len(points) == 0 {
		return nil, fmt.Errorf("%s produced no tiles", op)
	}
	for _, point := range points {
		if point.X < 0 || point.Y < 0 || point.X >= width || point.Y >= height {
			return nil, fmt.Errorf("%s tile (%d,%d) outside map %dx%d", op, point.X, point.Y, width, height)
		}
	}
	return points, nil
}

func bresenhamLine(x1, y1, x2, y2 int) []mapPoint {
	var points []mapPoint
	dx := absInt(x2 - x1)
	dy := -absInt(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx + dy
	for {
		points = append(points, mapPoint{X: x1, Y: y1})
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
	return uniqueSortedMapPoints(points)
}

func uniqueSortedMapPoints(points []mapPoint) []mapPoint {
	seen := map[mapPoint]bool{}
	out := make([]mapPoint, 0, len(points))
	for _, point := range points {
		if seen[point] {
			continue
		}
		seen[point] = true
		out = append(out, point)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].X < out[j].X
	})
	return out
}

func linePointEstimate(x1, y1, x2, y2 int) int {
	dx := absInt(x2 - x1)
	dy := absInt(y2 - y1)
	if dx > dy {
		return dx + 1
	}
	return dy + 1
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (f *File) variableStructSpec() (SectionSpec, error) {
	spec, err := f.writeSpec()
	if err != nil {
		return SectionSpec{}, err
	}
	triggers, ok := spec.section("Triggers")
	if !ok {
		return SectionSpec{}, fmt.Errorf("missing Triggers spec")
	}
	variableSpec, ok := triggers.Structs["VariableStruct"]
	if !ok {
		return SectionSpec{}, fmt.Errorf("missing VariableStruct spec")
	}
	return variableSpec, nil
}

func (f *File) refreshTriggers() error {
	f.body = f.root.raw()
	f.InflatedBytes = len(f.body)
	triggers, err := f.root.triggerInfo()
	if err != nil {
		return err
	}
	f.Triggers = triggers
	return nil
}

func (f *File) nextVariableID() int {
	maxID := -1
	if triggers := f.root.section("Triggers"); triggers != nil {
		for _, variable := range triggers.list("variable_data") {
			id, ok := variable.intValue("variable_id")
			if ok && id > maxID {
				maxID = id
			}
		}
	}
	return maxID + 1
}

func (f *File) findVariable(recipe VariableRecipe) (*parsedNode, error) {
	node, _, err := f.findVariableWithIndex(recipe)
	return node, err
}

func (f *File) findVariableWithIndex(recipe VariableRecipe) (*parsedNode, int, error) {
	if recipe.TargetID != nil {
		return f.findVariableByIDWithIndex(*recipe.TargetID)
	}
	if recipe.ID != nil {
		return f.findVariableByIDWithIndex(*recipe.ID)
	}
	if recipe.TargetName != "" {
		return f.findVariableByNameWithIndex(recipe.TargetName)
	}
	if recipe.Name != "" && recipe.Op != "add_variable" {
		return f.findVariableByNameWithIndex(recipe.Name)
	}
	return nil, -1, fmt.Errorf("%s requires target_id, id, target_name, or name", recipe.Op)
}

func (f *File) findVariableByID(id int) (*parsedNode, error) {
	node, _, err := f.findVariableByIDWithIndex(id)
	return node, err
}

func (f *File) findVariableByIDWithIndex(id int) (*parsedNode, int, error) {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return nil, -1, fmt.Errorf("missing Triggers section")
	}
	for index, variable := range triggers.list("variable_data") {
		variableID, ok := variable.intValue("variable_id")
		if ok && variableID == id {
			return variable, index, nil
		}
	}
	return nil, -1, fmt.Errorf("variable id %d not found", id)
}

func (f *File) findVariableByName(name string) (*parsedNode, error) {
	node, _, err := f.findVariableByNameWithIndex(name)
	return node, err
}

func (f *File) findVariableByNameWithIndex(name string) (*parsedNode, int, error) {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return nil, -1, fmt.Errorf("missing Triggers section")
	}
	for index, variable := range triggers.list("variable_data") {
		variableName, _ := variable.stringValue("variable_name")
		if variableName == name {
			return variable, index, nil
		}
	}
	return nil, -1, fmt.Errorf("variable name %q not found", name)
}

func (f *File) findTrigger(recipe TriggerRecipe) (*parsedNode, error) {
	index, err := f.findTriggerIndex(recipe)
	if err != nil {
		return nil, err
	}
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return nil, fmt.Errorf("missing Triggers section")
	}
	triggerNodes := triggers.list("trigger_data")
	return triggerNodes[index], nil
}

func (f *File) findTriggerIndex(recipe TriggerRecipe) (int, error) {
	triggers := f.root.section("Triggers")
	if triggers == nil {
		return 0, fmt.Errorf("missing Triggers section")
	}
	triggerNodes := triggers.list("trigger_data")
	if recipe.TargetIndex != nil {
		if *recipe.TargetIndex < 0 || *recipe.TargetIndex >= len(triggerNodes) {
			return 0, fmt.Errorf("target_index %d out of range 0..%d", *recipe.TargetIndex, len(triggerNodes)-1)
		}
		return *recipe.TargetIndex, nil
	}
	if recipe.TargetName != "" {
		for i, trigger := range triggerNodes {
			name, _ := trigger.stringValue("trigger_name")
			if name == recipe.TargetName {
				return i, nil
			}
		}
		return 0, fmt.Errorf("target_name %q not found", recipe.TargetName)
	}
	return 0, fmt.Errorf("%s requires target_index or target_name", recipe.Op)
}

func appendEffectsToTrigger(spec *Spec, trigger *parsedNode, recipes []EffectRecipe) error {
	_, effectSpec, _, err := triggerChildSpecs(spec)
	if err != nil {
		return err
	}
	field := trigger.field("effect_data")
	if field == nil {
		return fmt.Errorf("missing effect_data")
	}
	for _, recipe := range recipes {
		overrides, err := effectOverrides(recipe)
		if err != nil {
			return err
		}
		raw, err := buildStructRaw(effectSpec, overrides)
		if err != nil {
			return err
		}
		field.Elements = append(field.Elements, &parsedNode{
			Name: "EffectStruct",
			Raw:  raw,
		})
	}
	if err := setIntField(trigger, "number_of_effects", "s32", len(field.Elements)); err != nil {
		return err
	}
	return setIntListField(trigger, "effect_display_order_array", "s32", indexIntList(len(field.Elements)))
}

func appendConditionsToTrigger(spec *Spec, trigger *parsedNode, recipes []ConditionRecipe) error {
	_, _, conditionSpec, err := triggerChildSpecs(spec)
	if err != nil {
		return err
	}
	field := trigger.field("condition_data")
	if field == nil {
		return fmt.Errorf("missing condition_data")
	}
	for _, recipe := range recipes {
		overrides, err := conditionOverrides(recipe)
		if err != nil {
			return err
		}
		raw, err := buildStructRaw(conditionSpec, overrides)
		if err != nil {
			return err
		}
		field.Elements = append(field.Elements, &parsedNode{
			Name: "ConditionStruct",
			Raw:  raw,
		})
	}
	if err := setIntField(trigger, "number_of_conditions", "s32", len(field.Elements)); err != nil {
		return err
	}
	return setIntListField(trigger, "condition_display_order_array", "s32", indexIntList(len(field.Elements)))
}

func editTriggerChildLists(spec *Spec, trigger *parsedNode, recipe TriggerRecipe) error {
	if isTrue(recipe.ClearEffects) && len(recipe.RemoveEffects) > 0 {
		return fmt.Errorf("edit_trigger cannot combine clear_effects with remove_effects")
	}
	if isTrue(recipe.ClearConditions) && len(recipe.RemoveConditions) > 0 {
		return fmt.Errorf("edit_trigger cannot combine clear_conditions with remove_conditions")
	}
	if recipe.ReplaceEffects != nil && (isTrue(recipe.ClearEffects) || len(recipe.RemoveEffects) > 0 || len(recipe.Effects) > 0) {
		return fmt.Errorf("edit_trigger replace_effects cannot be combined with clear_effects, remove_effects, or effects")
	}
	if recipe.ReplaceConditions != nil && (isTrue(recipe.ClearConditions) || len(recipe.RemoveConditions) > 0 || len(recipe.Conditions) > 0) {
		return fmt.Errorf("edit_trigger replace_conditions cannot be combined with clear_conditions, remove_conditions, or conditions")
	}
	if recipe.ReplaceEffects != nil {
		effects, err := buildEffectNodes(spec, recipe.ReplaceEffects)
		if err != nil {
			return err
		}
		if err := replaceTriggerChildList(trigger, "effect_data", "number_of_effects", "effect_display_order_array", effects); err != nil {
			return err
		}
	} else if isTrue(recipe.ClearEffects) {
		if err := replaceTriggerChildList(trigger, "effect_data", "number_of_effects", "effect_display_order_array", nil); err != nil {
			return err
		}
	} else if len(recipe.RemoveEffects) > 0 {
		if err := removeTriggerChildren(trigger, "effect_data", "number_of_effects", "effect_display_order_array", recipe.RemoveEffects); err != nil {
			return err
		}
	}
	if recipe.ReplaceConditions != nil {
		conditions, err := buildConditionNodes(spec, recipe.ReplaceConditions)
		if err != nil {
			return err
		}
		if err := replaceTriggerChildList(trigger, "condition_data", "number_of_conditions", "condition_display_order_array", conditions); err != nil {
			return err
		}
	} else if isTrue(recipe.ClearConditions) {
		if err := replaceTriggerChildList(trigger, "condition_data", "number_of_conditions", "condition_display_order_array", nil); err != nil {
			return err
		}
	} else if len(recipe.RemoveConditions) > 0 {
		if err := removeTriggerChildren(trigger, "condition_data", "number_of_conditions", "condition_display_order_array", recipe.RemoveConditions); err != nil {
			return err
		}
	}
	return nil
}

func buildEffectNodes(spec *Spec, recipes []EffectRecipe) ([]*parsedNode, error) {
	_, effectSpec, _, err := triggerChildSpecs(spec)
	if err != nil {
		return nil, err
	}
	nodes := make([]*parsedNode, 0, len(recipes))
	for _, recipe := range recipes {
		overrides, err := effectOverrides(recipe)
		if err != nil {
			return nil, err
		}
		raw, err := buildStructRaw(effectSpec, overrides)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &parsedNode{Name: "EffectStruct", Raw: raw})
	}
	return nodes, nil
}

func buildConditionNodes(spec *Spec, recipes []ConditionRecipe) ([]*parsedNode, error) {
	_, _, conditionSpec, err := triggerChildSpecs(spec)
	if err != nil {
		return nil, err
	}
	nodes := make([]*parsedNode, 0, len(recipes))
	for _, recipe := range recipes {
		overrides, err := conditionOverrides(recipe)
		if err != nil {
			return nil, err
		}
		raw, err := buildStructRaw(conditionSpec, overrides)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &parsedNode{Name: "ConditionStruct", Raw: raw})
	}
	return nodes, nil
}

func removeTriggerChildren(trigger *parsedNode, dataField, countField, orderField string, indices []int) error {
	field := trigger.field(dataField)
	if field == nil {
		return fmt.Errorf("missing %s", dataField)
	}
	remove, err := normalizedRemovalIndices(indices, len(field.Elements), dataField)
	if err != nil {
		return err
	}
	next := make([]*parsedNode, 0, len(field.Elements)-len(remove))
	for i, node := range field.Elements {
		if !remove[i] {
			next = append(next, node)
		}
	}
	return replaceTriggerChildList(trigger, dataField, countField, orderField, next)
}

func replaceTriggerChildList(trigger *parsedNode, dataField, countField, orderField string, elements []*parsedNode) error {
	field := trigger.field(dataField)
	if field == nil {
		return fmt.Errorf("missing %s", dataField)
	}
	field.Elements = elements
	if err := setIntField(trigger, countField, "s32", len(field.Elements)); err != nil {
		return err
	}
	return setIntListField(trigger, orderField, "s32", indexIntList(len(field.Elements)))
}

func normalizedRemovalIndices(indices []int, count int, label string) (map[int]bool, error) {
	remove := map[int]bool{}
	for _, index := range indices {
		if index < 0 || index >= count {
			return nil, fmt.Errorf("%s removal index %d out of range 0..%d", label, index, count-1)
		}
		if remove[index] {
			return nil, fmt.Errorf("%s removal index %d repeated", label, index)
		}
		remove[index] = true
	}
	return remove, nil
}

func isTrue(value *bool) bool {
	return value != nil && *value
}

func (f *File) Write(path string) error {
	if !SupportsWriteVersion(f.Version) {
		return fmt.Errorf("scenario version %q is read-only in AoE2Kit; writing currently targets DE 1.57 and 1.58 only", f.Version)
	}
	body, err := f.RebuildBody()
	if err != nil {
		return err
	}
	compressed, err := DeflateRaw(body)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	out.Write(f.header)
	out.Write(compressed)
	return os.WriteFile(path, out.Bytes(), 0644)
}

func buildUnitFromRecipe(spec *Spec, recipe UnitRecipe, fallbackReferenceID int) ([]byte, error) {
	if recipe.Op != "add_unit" {
		return nil, fmt.Errorf("unsupported unit op %q", recipe.Op)
	}
	if recipe.UnitConst <= 0 {
		return nil, fmt.Errorf("add_unit requires positive unit_const")
	}
	unitSpec, err := unitStructSpec(spec)
	if err != nil {
		return nil, err
	}
	return buildStructRaw(unitSpec, map[string]any{
		"x":                       floatPtrValue(recipe.X, 0.5),
		"y":                       floatPtrValue(recipe.Y, 0.5),
		"z":                       floatPtrValue(recipe.Z, 0),
		"reference_id":            intValue(recipe.ReferenceID, fallbackReferenceID),
		"unit_const":              recipe.UnitConst,
		"status":                  intValue(recipe.Status, 2),
		"rotation":                floatPtrValue(recipe.Rotation, 0),
		"initial_animation_frame": intValue(recipe.InitialAnimationFrame, 0),
		"garrisoned_in_id":        intValue(recipe.GarrisonedInID, -1),
		"caption_string_id":       intValue(recipe.CaptionStringID, -1),
		"caption_string":          recipe.CaptionString,
	})
}

func buildDisplayInstructionTrigger(spec *Spec, recipe TriggerRecipe) ([]byte, error) {
	return buildSingleEffectTrigger(spec, recipe, map[string]any{
		"effect_type":                      20,
		"source_player":                    intValue(recipe.SourcePlayer, 1),
		"display_time":                     intValue(recipe.DisplayTime, 10),
		"instruction_panel_position":       intValue(recipe.InstructionPanelPosition, 0),
		"play_sound":                       intValue(recipe.PlaySound, 0),
		"message":                          recipe.Message,
		"use_tag_color_for_icon":           intValue(recipe.UseTagColorForIcon, 0),
		"number_of_units_selected":         -1,
		"selected_object_ids":              []any{},
		"message_option1":                  "",
		"message_option2":                  "",
		"sound_name":                       "",
		"object_list_unit_id":              -1,
		"string_id":                        -1,
		"legacy_location_object_reference": -1,
	})
}

func buildCreateObjectTrigger(spec *Spec, recipe TriggerRecipe) ([]byte, error) {
	if recipe.ObjectListUnitID == nil {
		return nil, fmt.Errorf("add_create_object requires object_list_unit_id")
	}
	if recipe.LocationX == nil {
		return nil, fmt.Errorf("add_create_object requires location_x")
	}
	if recipe.LocationY == nil {
		return nil, fmt.Errorf("add_create_object requires location_y")
	}
	return buildSingleEffectTrigger(spec, recipe, map[string]any{
		"effect_type":              11,
		"object_list_unit_id":      *recipe.ObjectListUnitID,
		"source_player":            intValue(recipe.SourcePlayer, 1),
		"location_x":               *recipe.LocationX,
		"location_y":               *recipe.LocationY,
		"item_id":                  intValue(recipe.ItemID, -1),
		"facet":                    intValue(recipe.Facet, -1),
		"disable_sound":            intValue(recipe.DisableSound, 0),
		"number_of_units_selected": -1,
		"selected_object_ids":      []any{},
	})
}

func buildSingleEffectTrigger(spec *Spec, recipe TriggerRecipe, effectOverrides map[string]any) ([]byte, error) {
	return buildTriggerRaw(spec, recipe, []map[string]any{effectOverrides}, nil)
}

func buildTriggerFromRecipe(spec *Spec, recipe TriggerRecipe) ([]byte, error) {
	var effects []map[string]any
	for _, effect := range recipe.Effects {
		overrides, err := effectOverrides(effect)
		if err != nil {
			return nil, err
		}
		effects = append(effects, overrides)
	}
	var conditions []map[string]any
	for _, condition := range recipe.Conditions {
		overrides, err := conditionOverrides(condition)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, overrides)
	}
	return buildTriggerRaw(spec, recipe, effects, conditions)
}

func buildTriggerRaw(spec *Spec, recipe TriggerRecipe, effectOverrides []map[string]any, conditionOverrides []map[string]any) ([]byte, error) {
	triggerSpec, effectSpec, conditionSpec, err := triggerChildSpecs(spec)
	if err != nil {
		return nil, err
	}
	enabled := 0
	if recipe.Enabled != nil && *recipe.Enabled {
		enabled = 1
	}
	looping := 0
	if recipe.Looping != nil && *recipe.Looping {
		looping = 1
	}
	executeOnLoad := 0
	if recipe.ExecuteOnLoad != nil && *recipe.ExecuteOnLoad {
		executeOnLoad = 1
	}
	displayAsObjective := 0
	if recipe.DisplayAsObjective != nil && *recipe.DisplayAsObjective {
		displayAsObjective = 1
	}
	displayOnScreen := 0
	if recipe.DisplayOnScreen != nil && *recipe.DisplayOnScreen {
		displayOnScreen = 1
	}
	makeHeader := 0
	if recipe.MakeHeader != nil && *recipe.MakeHeader {
		makeHeader = 1
	}
	muteObjectives := 0
	if recipe.MuteObjectives != nil && *recipe.MuteObjectives {
		muteObjectives = 1
	}
	name := recipe.Name
	if name == "" {
		name = "AoE2Kit Trigger"
	}
	effects := make([][]byte, 0, len(effectOverrides))
	for _, overrides := range effectOverrides {
		effect, err := buildStructRaw(effectSpec, overrides)
		if err != nil {
			return nil, err
		}
		effects = append(effects, effect)
	}
	conditions := make([][]byte, 0, len(conditionOverrides))
	for _, overrides := range conditionOverrides {
		condition, err := buildStructRaw(conditionSpec, overrides)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, condition)
	}
	return buildStructRaw(triggerSpec, map[string]any{
		"enabled":                           enabled,
		"looping":                           looping,
		"execute_on_load":                   executeOnLoad,
		"description_string_table_id":       intValue(recipe.DescriptionStringID, 0),
		"display_as_objective":              displayAsObjective,
		"objective_description_order":       intValue(recipe.ObjectiveDescriptionOrder, 0),
		"make_header":                       makeHeader,
		"short_description_string_table_id": intValue(recipe.ShortDescriptionStringID, 0),
		"display_on_screen":                 displayOnScreen,
		"mute_objectives":                   muteObjectives,
		"trigger_description":               recipe.Description,
		"trigger_name":                      name,
		"short_description":                 recipe.ShortDescription,
		"number_of_effects":                 len(effects),
		"effect_data":                       effects,
		"effect_display_order_array":        indexAnyList(len(effects)),
		"number_of_conditions":              len(conditions),
		"condition_data":                    conditions,
		"condition_display_order_array":     indexAnyList(len(conditions)),
	})
}

func intValue(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func floatPtrValue(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}

func effectOverrides(effect EffectRecipe) (map[string]any, error) {
	switch effect.Op {
	case "change_diplomacy":
		if effect.TargetPlayer == nil {
			return nil, fmt.Errorf("change_diplomacy requires target_player")
		}
		if effect.Diplomacy == nil {
			return nil, fmt.Errorf("change_diplomacy requires diplomacy")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_diplomacy"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"target_player":            *effect.TargetPlayer,
			"diplomacy":                *effect.Diplomacy,
			"mutual_diplomacy":         intValue(effect.MutualDiplomacy, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "display_instructions":
		return map[string]any{
			"effect_type":                      effectTypeID("display_instructions"),
			"source_player":                    intValue(effect.SourcePlayer, 1),
			"display_time":                     intValue(effect.DisplayTime, 10),
			"instruction_panel_position":       intValue(effect.InstructionPanelPosition, 0),
			"play_sound":                       intValue(effect.PlaySound, 0),
			"message":                          effect.Message,
			"use_tag_color_for_icon":           intValue(effect.UseTagColorForIcon, 0),
			"number_of_units_selected":         -1,
			"selected_object_ids":              []any{},
			"message_option1":                  "",
			"message_option2":                  "",
			"sound_name":                       "",
			"object_list_unit_id":              -1,
			"string_id":                        -1,
			"legacy_location_object_reference": -1,
		}, nil
	case "display_timer":
		return map[string]any{
			"effect_type":                      effectTypeID("display_timer"),
			"source_player":                    intValue(effect.SourcePlayer, 1),
			"display_time":                     intValue(effect.DisplayTime, 999999),
			"time_unit":                        intValue(effect.TimeUnit, 2),
			"timer":                            intValue(effect.TimerID, 0),
			"reset_timer":                      intValue(effect.ResetTimer, 0),
			"message":                          effect.Message,
			"number_of_units_selected":         -1,
			"selected_object_ids":              []any{},
			"message_option1":                  "",
			"message_option2":                  "",
			"sound_name":                       "",
			"object_list_unit_id":              -1,
			"string_id":                        -1,
			"legacy_location_object_reference": -1,
		}, nil
	case "send_chat":
		return map[string]any{
			"effect_type":                      effectTypeID("send_chat"),
			"source_player":                    intValue(effect.SourcePlayer, 1),
			"message":                          effect.Message,
			"number_of_units_selected":         -1,
			"selected_object_ids":              []any{},
			"message_option1":                  "",
			"message_option2":                  "",
			"sound_name":                       "",
			"object_list_unit_id":              -1,
			"string_id":                        -1,
			"legacy_location_object_reference": -1,
		}, nil
	case "play_sound":
		if effect.SoundName == "" {
			return nil, fmt.Errorf("play_sound requires sound_name")
		}
		return map[string]any{
			"effect_type":              effectTypeID("play_sound"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"sound_name":               effect.SoundName,
			"play_sound":               intValue(effect.PlaySound, -1),
			"global_sound":             intValue(effect.GlobalSound, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
		}, nil
	case "research_technology":
		if effect.Technology == nil {
			return nil, fmt.Errorf("research_technology requires technology")
		}
		return map[string]any{
			"effect_type":                      effectTypeID("research_technology"),
			"source_player":                    intValue(effect.SourcePlayer, 1),
			"technology":                       *effect.Technology,
			"force_research_technology":        intValue(effect.ForceResearchTechnology, 0),
			"number_of_units_selected":         -1,
			"selected_object_ids":              []any{},
			"message":                          "",
			"sound_name":                       "",
			"legacy_location_object_reference": -1,
		}, nil
	case "create_object":
		if effect.ObjectListUnitID == nil {
			return nil, fmt.Errorf("create_object requires object_list_unit_id")
		}
		if effect.LocationX == nil {
			return nil, fmt.Errorf("create_object requires location_x")
		}
		if effect.LocationY == nil {
			return nil, fmt.Errorf("create_object requires location_y")
		}
		return map[string]any{
			"effect_type":              effectTypeID("create_object"),
			"object_list_unit_id":      *effect.ObjectListUnitID,
			"source_player":            intValue(effect.SourcePlayer, 1),
			"location_x":               *effect.LocationX,
			"location_y":               *effect.LocationY,
			"item_id":                  intValue(effect.ItemID, -1),
			"facet":                    intValue(effect.Facet, -1),
			"disable_sound":            intValue(effect.DisableSound, 0),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
		}, nil
	case "place_foundation", "build_object":
		if effect.ObjectListUnitID == nil {
			return nil, fmt.Errorf("%s requires object_list_unit_id", effect.Op)
		}
		if effect.LocationX == nil {
			return nil, fmt.Errorf("%s requires location_x", effect.Op)
		}
		if effect.LocationY == nil {
			return nil, fmt.Errorf("%s requires location_y", effect.Op)
		}
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"object_list_unit_id":      *effect.ObjectListUnitID,
			"source_player":            intValue(effect.SourcePlayer, 1),
			"location_x":               *effect.LocationX,
			"location_y":               *effect.LocationY,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
		}, nil
	case "kill_object":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("kill_object"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "remove_object":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("remove_object"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"object_state":             intValue(effect.ObjectState, -1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "task_object":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":                   effectTypeID("task_object"),
			"object_list_unit_id":           intValue(effect.ObjectListUnitID, -1),
			"source_player":                 intValue(effect.SourcePlayer, 1),
			"location_x":                    intValue(effect.LocationX, -1),
			"location_y":                    intValue(effect.LocationY, -1),
			"location_object_reference":     intValue(effect.LocationObjectReference, -1),
			"area_x1":                       intValue(effect.AreaX1, -1),
			"area_y1":                       intValue(effect.AreaY1, -1),
			"area_x2":                       intValue(effect.AreaX2, -1),
			"area_y2":                       intValue(effect.AreaY2, -1),
			"object_group":                  intValue(effect.ObjectGroup, -1),
			"object_type":                   intValue(effect.ObjectType, -1),
			"action_type":                   intValue(effect.ActionType, -1),
			"disable_garrison_unload_sound": intValue(effect.DisableGarrisonUnloadSound, -1),
			"max_units_affected":            intValue(effect.MaxUnitsAffected, -1),
			"issue_group_command":           intValue(effect.IssueGroupCommand, -1),
			"queue_action":                  intValue(effect.QueueAction, -1),
			"number_of_units_selected":      selectedCount(selected),
			"selected_object_ids":           selected,
		}, nil
	case "change_view":
		if effect.LocationX == nil {
			return nil, fmt.Errorf("change_view requires location_x")
		}
		if effect.LocationY == nil {
			return nil, fmt.Errorf("change_view requires location_y")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_view"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"location_x":               *effect.LocationX,
			"location_y":               *effect.LocationY,
			"scroll":                   intValue(effect.Scroll, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_player_name", "change_civilization_name":
		if effect.Message == "" {
			return nil, fmt.Errorf("%s requires message", effect.Op)
		}
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"message":                  effect.Message,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"sound_name":               "",
		}, nil
	case "change_player_color":
		if effect.PlayerColor == nil {
			return nil, fmt.Errorf("change_player_color requires player_color")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_player_color"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"player_color":             *effect.PlayerColor,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_ownership":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("change_ownership"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"target_player":            intValue(effect.TargetPlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"flash_object":             intValue(effect.FlashObject, 0),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "change_object_name":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("change_object_name"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"string_id":                intValue(effect.StringID, -1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"message":                  effect.Message,
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "change_object_description":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("change_object_description"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"string_id":                intValue(effect.StringID, -1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"message":                  effect.Message,
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "change_object_hp":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("change_object_hp"),
			"quantity":                 intValue(effect.Quantity, 0),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"operation":                intValue(effect.Operation, 1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "damage_object", "heal_object", "change_object_attack", "change_object_armor", "change_object_range", "change_object_speed", "change_object_caption":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"quantity":                 intValue(effect.Quantity, 0),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"string_id":                intValue(effect.StringID, -1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"operation":                intValue(effect.Operation, 1),
			"message":                  effect.Message,
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "freeze_unit", "stop_unit", "disable_unit_targeting", "enable_unit_targeting", "disable_object_selection", "enable_object_selection", "enable_object_deletion", "disable_object_deletion", "disable_unit_attackable", "enable_unit_attackable":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"object_filter":            intValue(effect.ObjectFilter, -1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "teleport_object":
		if effect.LocationX == nil {
			return nil, fmt.Errorf("teleport_object requires location_x")
		}
		if effect.LocationY == nil {
			return nil, fmt.Errorf("teleport_object requires location_y")
		}
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("teleport_object"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"location_x":               *effect.LocationX,
			"location_y":               *effect.LocationY,
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "change_object_stance":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("change_object_stance"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"attack_stance":            intValue(effect.AttackStance, -1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
		}, nil
	case "set_player_visibility", "set_visibility", "reveal_map":
		return map[string]any{
			"effect_type":              effectTypeID("set_player_visibility"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"target_player":            intValue(effect.TargetPlayer, 1),
			"visibility_state":         intValue(effect.VisibilityState, 0),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "enable_disable_object":
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("enable_disable_object"),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"enabled":                  intValue(effect.Enabled, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"object_group":             intValue(effect.ObjectGroup, -1),
			"object_type":              intValue(effect.ObjectType, -1),
			"max_units_affected":       intValue(effect.MaxUnitsAffected, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "enable_disable_technology":
		if effect.Technology == nil {
			return nil, fmt.Errorf("enable_disable_technology requires technology")
		}
		return map[string]any{
			"effect_type":              effectTypeID("enable_disable_technology"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"technology":               *effect.Technology,
			"enabled":                  intValue(effect.Enabled, 1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_train_location", "add_train_location":
		if effect.ObjectListUnitID == nil {
			return nil, fmt.Errorf("%s requires object_list_unit_id", effect.Op)
		}
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"object_list_unit_id":      *effect.ObjectListUnitID,
			"object_list_unit_id_2":    intValue(effect.ObjectListUnitID2, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"button_location":          intValue(effect.ButtonLocation, -1),
			"hotkey":                   intValue(effect.Hotkey, -1),
			"train_time":               intValue(effect.TrainTime, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_technology_location":
		if effect.Technology == nil {
			return nil, fmt.Errorf("change_technology_location requires technology")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_technology_location"),
			"technology":               *effect.Technology,
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"button_location":          intValue(effect.ButtonLocation, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "modify_attribute":
		if effect.ObjectAttributes == nil {
			return nil, fmt.Errorf("modify_attribute requires object_attributes")
		}
		out := map[string]any{
			"effect_type":              effectTypeID("modify_attribute"),
			"quantity":                 intValue(effect.Quantity, -1),
			"object_list_unit_id":      intValue(effect.ObjectListUnitID, -1),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"item_id":                  intValue(effect.ItemID, -1),
			"operation":                intValue(effect.Operation, 1),
			"object_attributes":        *effect.ObjectAttributes,
			"message":                  effect.Message,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
		}
		if effect.QuantityFloat != nil {
			out["quantity_float"] = *effect.QuantityFloat
		}
		return out, nil
	case "modify_resource":
		resource := intValue(effect.Resource, intValue(effect.TributeList, 0))
		return map[string]any{
			"effect_type":              effectTypeID("modify_resource"),
			"quantity":                 intValue(effect.Quantity, 0),
			"tribute_list":             resource,
			"source_player":            intValue(effect.SourcePlayer, 1),
			"item_id":                  intValue(effect.ItemID, -1),
			"operation":                intValue(effect.Operation, 1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_object_cost":
		if effect.ObjectListUnitID == nil {
			return nil, fmt.Errorf("change_object_cost requires object_list_unit_id")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_object_cost"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"object_list_unit_id":      *effect.ObjectListUnitID,
			"resource_1":               intValue(effect.Resource1, -1),
			"resource_1_quantity":      intValue(effect.Resource1Quantity, -1),
			"resource_2":               intValue(effect.Resource2, -1),
			"resource_2_quantity":      intValue(effect.Resource2Quantity, -1),
			"resource_3":               intValue(effect.Resource3, -1),
			"resource_3_quantity":      intValue(effect.Resource3Quantity, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_technology_cost":
		if effect.Technology == nil {
			return nil, fmt.Errorf("change_technology_cost requires technology")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_technology_cost"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"technology":               *effect.Technology,
			"resource_1":               intValue(effect.Resource1, -1),
			"resource_1_quantity":      intValue(effect.Resource1Quantity, -1),
			"resource_2":               intValue(effect.Resource2, -1),
			"resource_2_quantity":      intValue(effect.Resource2Quantity, -1),
			"resource_3":               intValue(effect.Resource3, -1),
			"resource_3_quantity":      intValue(effect.Resource3Quantity, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_technology_research_time":
		if effect.Technology == nil {
			return nil, fmt.Errorf("change_technology_research_time requires technology")
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_technology_research_time"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"technology":               *effect.Technology,
			"quantity":                 intValue(effect.Quantity, intValue(effect.TrainTime, -1)),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "change_technology_name", "change_technology_description":
		if effect.Technology == nil {
			return nil, fmt.Errorf("%s requires technology", effect.Op)
		}
		if effect.Message == "" {
			return nil, fmt.Errorf("%s requires message", effect.Op)
		}
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"technology":               *effect.Technology,
			"message":                  effect.Message,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"sound_name":               "",
		}, nil
	case "change_technology_icon", "change_technology_hotkey":
		if effect.Technology == nil {
			return nil, fmt.Errorf("%s requires technology", effect.Op)
		}
		value := intValue(effect.Quantity, -1)
		if effect.Op == "change_technology_hotkey" {
			value = intValue(effect.Hotkey, value)
		}
		return map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"technology":               *effect.Technology,
			"quantity":                 value,
			"hotkey":                   intValue(effect.Hotkey, -1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "train_unit":
		if effect.ObjectListUnitID == nil {
			return nil, fmt.Errorf("train_unit requires object_list_unit_id")
		}
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		return map[string]any{
			"effect_type":              effectTypeID("train_unit"),
			"object_list_unit_id":      *effect.ObjectListUnitID,
			"source_player":            intValue(effect.SourcePlayer, 1),
			"area_x1":                  intValue(effect.AreaX1, -1),
			"area_y1":                  intValue(effect.AreaY1, -1),
			"area_x2":                  intValue(effect.AreaX2, -1),
			"area_y2":                  intValue(effect.AreaY2, -1),
			"location_x":               intValue(effect.LocationX, -1),
			"location_y":               intValue(effect.LocationY, -1),
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "initiate_research", "research_local_technology":
		if effect.Technology == nil {
			return nil, fmt.Errorf("%s requires technology", effect.Op)
		}
		selected := selectedObjectAnyList(effect.SelectedObjectIDs)
		out := map[string]any{
			"effect_type":              effectTypeID(effect.Op),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"technology":               *effect.Technology,
			"number_of_units_selected": selectedCount(selected),
			"selected_object_ids":      selected,
			"message":                  "",
			"sound_name":               "",
		}
		if effect.LocalTechnology != nil {
			out["local_technology"] = *effect.LocalTechnology
		}
		return out, nil
	case "script_call":
		call := effect.Message
		if call == "" {
			call = effect.XSFunction
		}
		if call == "" {
			return nil, fmt.Errorf("script_call requires message or xs_function")
		}
		return map[string]any{
			"effect_type":              effectTypeID("script_call"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  call,
			"sound_name":               "",
		}, nil
	case "change_variable", "modify_variable":
		if effect.Variable == nil {
			return nil, fmt.Errorf("%s requires variable", effect.Op)
		}
		return map[string]any{
			"effect_type":              effectTypeID("change_variable"),
			"quantity":                 intValue(effect.Quantity, 0),
			"operation":                intValue(effect.Operation, 1),
			"variable":                 *effect.Variable,
			"message":                  effect.Message,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"sound_name":               "",
		}, nil
	case "modify_variable_by_variable":
		if effect.Variable == nil {
			return nil, fmt.Errorf("modify_variable_by_variable requires variable")
		}
		if effect.Variable2 == nil {
			return nil, fmt.Errorf("modify_variable_by_variable requires variable2")
		}
		return map[string]any{
			"effect_type":              effectTypeID("modify_variable_by_variable"),
			"operation":                intValue(effect.Operation, 1),
			"variable":                 *effect.Variable,
			"variable2":                *effect.Variable2,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "clear_timer":
		if effect.TimerID == nil {
			return nil, fmt.Errorf("clear_timer requires timer_id")
		}
		return map[string]any{
			"effect_type":              effectTypeID("clear_timer"),
			"timer":                    *effect.TimerID,
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
			"message":                  "",
			"sound_name":               "",
		}, nil
	case "activate_trigger":
		if effect.TriggerID == nil {
			return nil, fmt.Errorf("activate_trigger requires trigger_id")
		}
		return map[string]any{
			"effect_type": effectTypeID("activate_trigger"),
			"trigger_id":  *effect.TriggerID,
		}, nil
	case "deactivate_trigger":
		if effect.TriggerID == nil {
			return nil, fmt.Errorf("deactivate_trigger requires trigger_id")
		}
		return map[string]any{
			"effect_type": effectTypeID("deactivate_trigger"),
			"trigger_id":  *effect.TriggerID,
		}, nil
	case "declare_victory":
		return map[string]any{
			"effect_type":              effectTypeID("declare_victory"),
			"source_player":            intValue(effect.SourcePlayer, 1),
			"enabled":                  intValue(effect.Enabled, 1),
			"number_of_units_selected": -1,
			"selected_object_ids":      []any{},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported effect op %q", effect.Op)
	}
}

func effectTypeID(op string) int {
	id, ok := EffectTypeForOp(op)
	if !ok {
		return -1
	}
	return id
}

func conditionOverrides(condition ConditionRecipe) (map[string]any, error) {
	switch condition.Op {
	case "timer":
		if condition.Timer == nil {
			return nil, fmt.Errorf("timer condition requires timer")
		}
		return map[string]any{
			"condition_type": 10,
			"timer":          *condition.Timer,
			"inverted":       intValue(condition.Inverted, 0),
		}, nil
	case "object_selected":
		if condition.UnitObject == nil {
			return nil, fmt.Errorf("object_selected condition requires unit_object")
		}
		return map[string]any{
			"condition_type": 11,
			"unit_object":    *condition.UnitObject,
			"inverted":       intValue(condition.Inverted, 0),
		}, nil
	case "object_in_area", "objects_in_area":
		if condition.AreaX1 == nil || condition.AreaY1 == nil || condition.AreaX2 == nil || condition.AreaY2 == nil {
			return nil, fmt.Errorf("object_in_area requires area_x1, area_y1, area_x2, and area_y2")
		}
		return map[string]any{
			"condition_type":                    5,
			"quantity":                          intValue(condition.Quantity, 1),
			"object_list":                       intValue(condition.ObjectList, -1),
			"source_player":                     intValue(condition.SourcePlayer, 1),
			"area_x1":                           *condition.AreaX1,
			"area_y1":                           *condition.AreaY1,
			"area_x2":                           *condition.AreaX2,
			"area_y2":                           *condition.AreaY2,
			"object_group":                      intValue(condition.ObjectGroup, -1),
			"object_type":                       intValue(condition.ObjectType, -1),
			"object_state":                      intValue(condition.ObjectState, -1),
			"include_changeable_weapon_objects": intValue(condition.IncludeChangeableWeaponObjects, -1),
			"inverted":                          intValue(condition.Inverted, 0),
		}, nil
	case "own_objects", "own_fewer_objects":
		conditionType := 3
		if condition.Op == "own_fewer_objects" {
			conditionType = 4
		}
		return map[string]any{
			"condition_type":                    conditionType,
			"quantity":                          intValue(condition.Quantity, 1),
			"object_list":                       intValue(condition.ObjectList, -1),
			"source_player":                     intValue(condition.SourcePlayer, 1),
			"area_x1":                           intValue(condition.AreaX1, -1),
			"area_y1":                           intValue(condition.AreaY1, -1),
			"area_x2":                           intValue(condition.AreaX2, -1),
			"area_y2":                           intValue(condition.AreaY2, -1),
			"object_group":                      intValue(condition.ObjectGroup, -1),
			"object_type":                       intValue(condition.ObjectType, -1),
			"object_state":                      intValue(condition.ObjectState, -1),
			"include_changeable_weapon_objects": intValue(condition.IncludeChangeableWeaponObjects, -1),
			"inverted":                          intValue(condition.Inverted, 0),
		}, nil
	case "accumulate_attribute":
		if condition.Attribute == nil {
			return nil, fmt.Errorf("accumulate_attribute requires attribute")
		}
		return map[string]any{
			"condition_type": 8,
			"quantity":       intValue(condition.Quantity, 0),
			"attribute":      *condition.Attribute,
			"source_player":  intValue(condition.SourcePlayer, 1),
			"inverted":       intValue(condition.Inverted, 0),
		}, nil
	case "object_visible":
		if condition.UnitObject == nil {
			return nil, fmt.Errorf("object_visible condition requires unit_object")
		}
		return map[string]any{
			"condition_type": 15,
			"unit_object":    *condition.UnitObject,
			"inverted":       intValue(condition.Inverted, 0),
		}, nil
	case "variable_value":
		if condition.Variable == nil {
			return nil, fmt.Errorf("variable_value requires variable")
		}
		if condition.Comparison == nil {
			return nil, fmt.Errorf("variable_value requires comparison")
		}
		return map[string]any{
			"condition_type": 22,
			"quantity":       intValue(condition.Quantity, 0),
			"variable":       *condition.Variable,
			"comparison":     *condition.Comparison,
			"inverted":       intValue(condition.Inverted, 0),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported condition op %q", condition.Op)
	}
}

func selectedObjectAnyList(ids []int) []any {
	out := make([]any, 0, len(ids))
	for _, id := range ids {
		out = append(out, id)
	}
	return out
}

func selectedCount(ids []any) int {
	if len(ids) == 0 {
		return -1
	}
	return len(ids)
}

func triggerEffectSpecs(spec *Spec) (SectionSpec, SectionSpec, error) {
	triggerSpec, effectSpec, _, err := triggerChildSpecs(spec)
	return triggerSpec, effectSpec, err
}

func unitStructSpec(spec *Spec) (SectionSpec, error) {
	units, ok := spec.section("Units")
	if !ok {
		return SectionSpec{}, fmt.Errorf("missing Units section")
	}
	playerUnits, ok := units.Structs["PlayerUnitsStruct"]
	if !ok {
		return SectionSpec{}, fmt.Errorf("missing PlayerUnitsStruct")
	}
	unitSpec, ok := playerUnits.Structs["UnitStruct"]
	if !ok {
		return SectionSpec{}, fmt.Errorf("missing UnitStruct")
	}
	return unitSpec, nil
}

func triggerChildSpecs(spec *Spec) (SectionSpec, SectionSpec, SectionSpec, error) {
	triggers, ok := spec.section("Triggers")
	if !ok {
		return SectionSpec{}, SectionSpec{}, SectionSpec{}, fmt.Errorf("missing Triggers section")
	}
	triggerSpec, ok := triggers.Structs["TriggerStruct"]
	if !ok {
		return SectionSpec{}, SectionSpec{}, SectionSpec{}, fmt.Errorf("missing TriggerStruct")
	}
	effectSpec, ok := triggerSpec.Structs["EffectStruct"]
	if !ok {
		return SectionSpec{}, SectionSpec{}, SectionSpec{}, fmt.Errorf("missing EffectStruct")
	}
	conditionSpec, ok := triggerSpec.Structs["ConditionStruct"]
	if !ok {
		return SectionSpec{}, SectionSpec{}, SectionSpec{}, fmt.Errorf("missing ConditionStruct")
	}
	return triggerSpec, effectSpec, conditionSpec, nil
}

func indexAnyList(count int) []any {
	out := make([]any, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, i)
	}
	return out
}

func indexUint32List(count int) []uint32 {
	out := make([]uint32, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, uint32(i))
	}
	return out
}

func buildStructRaw(spec SectionSpec, overrides map[string]any) ([]byte, error) {
	var out bytes.Buffer
	for _, retriever := range spec.Retrievers {
		value, ok := overrides[retriever.Name]
		if !ok {
			value = retrieverDefault(retriever)
		}
		if stringsHasStructPrefix(retriever.Type) {
			chunks, _ := value.([][]byte)
			for _, chunk := range chunks {
				out.Write(chunk)
			}
			continue
		}
		chunk, err := encodeRetrieverValue(retriever, value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", retriever.Name, err)
		}
		out.Write(chunk)
	}
	return out.Bytes(), nil
}

func encodeRetrieverValue(retriever RetrieverSpec, value any) ([]byte, error) {
	var out bytes.Buffer
	if values, ok := value.([]any); ok {
		for _, item := range values {
			chunk, err := encodePrimitive(retriever.Type, item)
			if err != nil {
				return nil, err
			}
			out.Write(chunk)
		}
		return out.Bytes(), nil
	}
	return encodePrimitive(retriever.Type, value)
}

func retrieverDefault(retriever RetrieverSpec) any {
	if len(retriever.Default) == 0 {
		return nil
	}
	var value any
	if err := json.Unmarshal(retriever.Default, &value); err != nil {
		return nil
	}
	return value
}

func setIntField(node *parsedNode, name, kind string, value int) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing %s", name)
	}
	raw, err := encodePrimitive(kind, value)
	if err != nil {
		return err
	}
	field.Raw = raw
	switch kind[0] {
	case 'u':
		field.Value = uint32(value)
	default:
		field.Value = int32(value)
	}
	return nil
}

func setBoolByteField(node *parsedNode, name string, value bool) error {
	intValue := 0
	if value {
		intValue = 1
	}
	return setIntField(node, name, "u8", intValue)
}

func setFloatField(node *parsedNode, name, kind string, value float64) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing field %q", name)
	}
	if kind != "f32" {
		return fmt.Errorf("unsupported float field kind %q", kind)
	}
	var raw [4]byte
	binary.LittleEndian.PutUint32(raw[:], math.Float32bits(float32(value)))
	field.Raw = raw[:]
	field.Value = float32(value)
	return nil
}

func setUint32ListField(node *parsedNode, name string, values []uint32) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing %s", name)
	}
	var raw []byte
	list := make([]any, 0, len(values))
	for _, value := range values {
		chunk, err := encodePrimitive("u32", value)
		if err != nil {
			return err
		}
		raw = append(raw, chunk...)
		list = append(list, value)
	}
	field.Raw = raw
	field.Value = list
	return nil
}

func setIntListField(node *parsedNode, name, kind string, values []int) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing %s", name)
	}
	var raw []byte
	list := make([]any, 0, len(values))
	for _, value := range values {
		chunk, err := encodePrimitive(kind, value)
		if err != nil {
			return err
		}
		raw = append(raw, chunk...)
		list = append(list, int32(value))
	}
	field.Raw = raw
	field.Value = list
	return nil
}

func setStringField(node *parsedNode, name, kind, value string) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing %s", name)
	}
	raw, err := encodePrimitive(kind, value)
	if err != nil {
		return err
	}
	field.Raw = raw
	field.Value = value
	return nil
}

func setZeroLengthStringField(node *parsedNode, name, kind string) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing %s", name)
	}
	typ, size, err := primitiveType(kind)
	if err != nil {
		return err
	}
	if typ != "str" {
		return fmt.Errorf("%s is %s, want length-prefixed string", name, kind)
	}
	field.Raw = make([]byte, size)
	field.Value = ""
	return nil
}

func setStringListField(node *parsedNode, name, kind string, values []string) error {
	field := node.field(name)
	if field == nil {
		return fmt.Errorf("missing %s", name)
	}
	var raw []byte
	list := make([]any, 0, len(values))
	for _, value := range values {
		chunk, err := encodePrimitive(kind, value)
		if err != nil {
			return err
		}
		raw = append(raw, chunk...)
		list = append(list, value)
	}
	field.Raw = raw
	field.Value = list
	return nil
}

func indexIntList(count int) []int {
	out := make([]int, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, i)
	}
	return out
}

func encodePrimitive(kind string, value any) ([]byte, error) {
	typ, size, err := primitiveType(kind)
	if err != nil {
		return nil, err
	}
	if typ == "str" {
		text, _ := value.(string)
		text = trimAtNUL(text)
		payload := append([]byte(text), 0)
		var out bytes.Buffer
		if err := writeSignedInt(&out, size, len(payload)); err != nil {
			return nil, err
		}
		out.Write(payload)
		return out.Bytes(), nil
	}
	if typ == "c" {
		out := make([]byte, size)
		copy(out, []byte(trimAtNUL(fmt.Sprint(value))))
		return out, nil
	}
	if typ == "data" {
		out := make([]byte, size)
		switch v := value.(type) {
		case string:
			if decoded, err := hex.DecodeString(v); err == nil {
				copy(out, decoded)
			} else {
				copy(out, []byte(v))
			}
		case []byte:
			copy(out, v)
		}
		return out, nil
	}
	out := make([]byte, size)
	switch typ {
	case "s":
		writeSignedBytes(out, int64(asInt(value)))
	case "u":
		writeUnsignedBytes(out, uint64(asInt(value)))
	case "f":
		switch size {
		case 4:
			binary.LittleEndian.PutUint32(out, math.Float32bits(float32(asFloat(value))))
		case 8:
			binary.LittleEndian.PutUint64(out, math.Float64bits(asFloat(value)))
		default:
			return nil, fmt.Errorf("unsupported float size %d", size)
		}
	default:
		return nil, fmt.Errorf("unsupported primitive type %q", typ)
	}
	return out, nil
}

func writeSignedInt(out *bytes.Buffer, size int, value int) error {
	buf := make([]byte, size)
	writeSignedBytes(buf, int64(value))
	_, err := out.Write(buf)
	return err
}

func writeSignedBytes(out []byte, value int64) {
	switch len(out) {
	case 1:
		out[0] = byte(int8(value))
	case 2:
		binary.LittleEndian.PutUint16(out, uint16(int16(value)))
	case 4:
		binary.LittleEndian.PutUint32(out, uint32(int32(value)))
	case 8:
		binary.LittleEndian.PutUint64(out, uint64(value))
	}
}

func writeUnsignedBytes(out []byte, value uint64) {
	switch len(out) {
	case 1:
		out[0] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(out, uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(out, uint32(value))
	case 8:
		binary.LittleEndian.PutUint64(out, value)
	}
}

func stringsHasStructPrefix(value string) bool {
	return len(value) >= len("struct:") && value[:len("struct:")] == "struct:"
}
