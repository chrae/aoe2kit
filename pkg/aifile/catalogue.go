package aifile

import (
	"sort"
	"strings"
)

const catalogueRevision = "16b6b78a6d9eb0b2d23e48ffd0b261212cf3a8bd"
const catalogueVersion = "de-local-goals-v8"
const catalogueSource = "https://github.com/airef/airef.github.io/blob/" + catalogueRevision + "/js/commands.js"

type CatalogueInfo struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
	Source   string `json:"source"`
	Scope    string `json:"scope"`
}

type LintCoverage struct {
	CheckedForms   int `json:"checked_forms"`
	UncheckedForms int `json:"unchecked_forms"`
}

type CommandCheck struct {
	File             string           `json:"file"`
	Line             int              `json:"line"`
	Column           int              `json:"column"`
	Command          string           `json:"command"`
	Signature        string           `json:"signature"`
	Source           string           `json:"source"`
	Confidence       string           `json:"confidence"`
	ReferenceFamily  string           `json:"reference_family"`
	Parameters       []ParameterCheck `json:"parameters,omitempty"`
	Notes            string           `json:"notes,omitempty"`
	Mechanism        string           `json:"mechanism,omitempty"`
	RuntimeSemantics string           `json:"runtime_semantics,omitempty"`
}

type ReferenceConstant struct {
	Value               int64  `json:"value"`
	DEOnly              bool   `json:"de_only,omitempty"`
	NeedsDefinition     bool   `json:"needs_definition,omitempty"`
	ReferenceFamily     string `json:"reference_family,omitempty"`
	Source              string `json:"source,omitempty"`
	Meaning             string `json:"meaning,omitempty"`
	Units               string `json:"units,omitempty"`
	Context             string `json:"context,omitempty"`
	RuntimeAvailability string `json:"runtime_availability,omitempty"`
}

type ParameterCheck struct {
	Kind       string                       `json:"kind"`
	Range      *[2]int64                    `json:"range,omitempty"`
	Allowed    []int64                      `json:"allowed,omitempty"`
	Exceptions []int64                      `json:"exceptions,omitempty"`
	Constants  map[string]ReferenceConstant `json:"constants,omitempty"`
}

type commandSpec struct {
	context    string
	params     []string
	bounds     map[int][2]int64
	allowed    map[int][]int64
	exceptions map[int][]int64
	notes      string
	mechanism  string
}

// This deliberately small catalogue covers census probes and DUC micro. Signatures
// and ranges were checked against the pinned reference, not a running engine.
// Unknown commands stay outside coverage rather than becoming invented APIs.
var perCatalogue = map[string]commandSpec{
	"goal":                   {context: "fact", params: []string{"goal-id", "goal-value"}, bounds: map[int][2]int64{0: {1, 16000}, 1: {-2147483648, 2147483647}}, notes: "Local AI goal equality test with two constant operands; no comparison operator. DE goal limits; not shared-goal storage. Runtime initialization and phase sequencing are not verified."},
	"set-goal":               {context: "action", params: []string{"goal-id", "goal-value"}, bounds: map[int][2]int64{0: {1, 16000}, 1: {-2147483648, 2147483647}}, notes: "Sets a local AI goal using two constant operands. DE goal limits; independent of shared goals and strategic numbers. Static signature only, no runtime phase-order guarantee."},
	"set-strategic-number":   {context: "action", params: []string{"sn-id", "sn-value"}, bounds: map[int][2]int64{0: {0, 511}, 1: {-2147483648, 2147483647}}, notes: "DE signed 32-bit values; individual SNs may have narrower usable ranges. Symbolic SN/player names are only partially resolved; no unused-slot or runtime behavior guarantee."},
	"up-compare-goal":        {context: "fact", params: []string{"goal-id", "compare-op", "comparison-value"}, bounds: map[int][2]int64{0: {1, 16000}, 2: {-2147483648, 2147483647}}, notes: "Operators >, >=, <, <=, ==, != accept optional c: or required g:/s: prefix for indirect values. Dynamic operands are goal/SN addresses. Runtime evaluation unverified."},
	"true":                   {context: "fact"},
	"false":                  {context: "fact"},
	"disable-self":           {context: "action"},
	"do-nothing":             {context: "action"},
	"up-full-reset-search":   {context: "action"},
	"enable-timer":           {context: "action", params: []string{"timer-id", "seconds"}, bounds: map[int][2]int64{0: {1, 100}}},
	"disable-timer":          {context: "action", params: []string{"timer-id"}, bounds: map[int][2]int64{0: {1, 100}}},
	"timer-triggered":        {context: "fact", params: []string{"timer-id"}, bounds: map[int][2]int64{0: {1, 100}}},
	"up-find-local":          {context: "fact action", params: []string{"type-op", "unit-id", "type-op", "result-count"}, bounds: map[int][2]int64{3: {0, 240}}},
	"up-get-search-state":    {context: "action", params: []string{"output-goal-id"}, bounds: map[int][2]int64{0: {41, 15996}}, notes: "Upper boundary follows the command page (15996); generic OutputGoalId documentation gives 15997. Boundary remains engine-unverified."},
	"up-chat-data-to-all":    {context: "action", params: []string{"text", "type-op", "integer"}, mechanism: "all_player_chat", notes: "Formatted chat to all players, not an ally-only state channel. Delivery, ordering and machine consumption are unverified."},
	"xs-script-call":         {context: "fact action", params: []string{"text"}},
	"include":                {context: "top", params: []string{"text"}},
	"load":                   {context: "top", params: []string{"text"}},
	"defconst":               {context: "top", params: []string{"symbol", "constant"}},
	"up-find-remote":         {context: "fact action", params: []string{"type-op", "unit-id", "type-op", "result-count"}, bounds: map[int][2]int64{3: {0, 40}}, notes: "Remote search uses the focus player; this is a signature check, not proof of focus or visibility state."},
	"up-filter-distance":     {context: "action", params: []string{"type-op", "min-distance", "type-op", "max-distance"}, bounds: map[int][2]int64{1: {-1, 32767}, 3: {-1, 32767}}},
	"up-filter-status":       {context: "action", params: []string{"type-op", "object-status", "type-op", "object-list"}, allowed: map[int][]int64{1: {0, 2, 3, 4, 5}, 3: {0, 1}}, notes: "The reference says normal find-local/find-remote ignore this filter; status-specific searches are separate commands."},
	"up-clean-search":        {context: "action", params: []string{"search-source", "object-data", "search-order"}, allowed: map[int][]int64{0: {1, 2}, 2: {0, 1, 2}}, notes: "ObjectData -1 requests duplicate removal; other object-data selectors are only partially catalogued."},
	"up-get-object-data":     {context: "fact action", params: []string{"object-data", "goal-id"}, bounds: map[int][2]int64{1: {1, 16000}}, notes: "Relevant micro selectors have individual meaning, units and context metadata. Broader ObjectData coverage remains partial; no selector's runtime availability is established."},
	"up-set-target-object":   {context: "fact action", params: []string{"search-source", "type-op", "search-index"}, allowed: map[int][]int64{0: {1, 2}}, bounds: map[int][2]int64{2: {0, 240}}, notes: "Upper boundary is uncertain: command page says 0..240, generic Index page says local 0..239 / remote 0..39. No engine-verified boundary."},
	"up-set-target-point":    {context: "action", params: []string{"point-goal-id"}, notes: "Upper boundary is uncertain: command page says 15598, general point-pair documentation says 15998. No fixed bound is enforced here."},
	"up-target-objects":      {context: "action", params: []string{"target-option", "duc-action", "formation", "attack-stance"}, allowed: map[int][]int64{0: {0, 1}, 1: {0, 1, 2, 3, 4, 5, 6, 7, 8, 11, 18}, 2: {-1, 2, 4, 7, 8}, 3: {-1, 0, 1, 2, 3}}, notes: "Only the action subset documented for target-objects is checked; target-point supports additional actions."},
	"up-target-point":        {context: "action", params: []string{"point-or-current", "duc-action", "formation", "attack-stance"}, bounds: map[int][2]int64{0: {41, 15998}}, exceptions: map[int][]int64{0: {0}}, allowed: map[int][]int64{1: {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, 2: {-1, 2, 4, 7, 8}, 3: {-1, 0, 1, 2, 3}}, notes: "DE actions 19..21 require explicit definitions. Train/research have distinct non-combat overloads; escrow behavior is unverified."},
	"shared-goal":            {context: "fact", params: []string{"shared-goal-id", "shared-value"}, bounds: map[int][2]int64{0: {1, 256}, 1: {-2147483648, 2147483647}}, mechanism: "global_shared_goals", notes: "ID/value equality shape, no comparison operator. Guide describes shared state across all AIs including enemies, not private team storage. Concurrent changes and runtime synchronization are unverified."},
	"set-shared-goal":        {context: "action", params: []string{"shared-goal-id", "shared-value"}, bounds: map[int][2]int64{0: {1, 256}, 1: {-2147483648, 2147483647}}, mechanism: "global_shared_goals", notes: "Guide describes state shared by all AIs including enemies; use only when the all-AI scope is intended. No delivery/ordering/atomicity guarantee is established."},
	"up-get-shared-goal":     {context: "action", params: []string{"type-op", "shared-goal-id", "goal-id"}, bounds: map[int][2]int64{1: {1, 256}, 2: {1, 16000}}, mechanism: "global_shared_goals", notes: "Reads a global shared goal into a local goal. Parameter range says 1..256 but the example uses 0; boundary uncertainty remains explicit and range violations only warn."},
	"up-set-shared-goal":     {context: "action", params: []string{"type-op", "shared-goal-id", "type-op", "shared-value"}, bounds: map[int][2]int64{1: {1, 256}, 3: {-2147483648, 2147483647}}, mechanism: "global_shared_goals", notes: "Global all-AI shared state, not ally-private state. Static operand checks do not establish ordering, atomicity or protocol correctness."},
	"set-signal":             {context: "action", params: []string{"signal-id"}, bounds: map[int][2]int64{0: {0, 255}}, mechanism: "legacy_scenario_signal", notes: "Guide sets a legacy AI Signal flag, documented for single-player scenario conditions. One single-player P2 canary on DE 101.103.57257.0 observed set-signal 1 -> type-12 condition -> scenario Script Call -> attached XS -> scenario-named xsdat at game time 4. See docs/SCENARIO_WRITER.md for evidence hashes and scope. No Multiplayer AI Signal, fe-set-signal, resource synchronization or exact resource-delta claim; general runtime semantics remain unverified."},
	"cc-add-resource":        {context: "action", params: []string{"stockpile-resource", "resource-amount"}, allowed: map[int][]int64{0: {0, 1, 2, 3}}, bounds: map[int][2]int64{1: {-32768, 32767}}, notes: "Reference adds a constant amount to the computer player's food/wood/stone/gold stockpile, including negative amounts; documented to work with cheats disabled. Signed 16-bit amount range is reference-backed, not an engine clamp or overflow claim. MCB canary did not measure the starting balance and does not verify exact resource-delta semantics or multiplayer synchronization."},
	"up-get-signal":          {context: "action", params: []string{"type-op", "signal-id", "goal-id"}, bounds: map[int][2]int64{1: {0, 255}, 2: {1, 16000}}, mechanism: "legacy_scenario_signal", notes: "Reads a scenario signal into a local goal. Reading fe-set-signal's multiplayer channel through this command is not established by this catalogue."},
	"up-set-signal":          {context: "action", params: []string{"type-op", "signal-id", "type-op", "signal-value"}, bounds: map[int][2]int64{1: {0, 255}, 3: {-2147483648, 2147483647}}, mechanism: "legacy_scenario_signal", notes: "Guide documents single-player AI Signal conditions; DE Multiplayer AI Signal uses fe-set-signal instead. No implicit cross-channel synchronization."},
	"fe-set-signal":          {context: "action", params: []string{"type-op", "multiplayer-signal-id", "type-op", "signal-value"}, bounds: map[int][2]int64{1: {0, 255}, 3: {-2147483648, 2147483647}}, mechanism: "multiplayer_scenario_signal", notes: "DE-only command for Multiplayer AI Signal conditions in single/multiplayer scenarios. Distinct from legacy AI Signal; delivery/ordering and a corresponding read path remain unverified."},
	"chat-to-allies":         {context: "action", params: []string{"text"}, mechanism: "ally_chat", notes: "Ally text chat, not verified shared state. Guide reports a leading numeral can invoke a taunt. No reliable delivery, ordering or AI consumption is established."},
	"up-chat-data-to-player": {context: "action", params: []string{"player-selector", "text", "type-op", "integer"}, mechanism: "recipient_chat", notes: "Formatted chat with an explicit recipient selector; every-ally is scoped in the catalogue. Recipient/wildcard resolution, delivery, ordering and synchronization remain unverified."},
}

var parameterConstants = map[string]map[string]ReferenceConstant{
	"stockpile-resource": {"food": {Value: 0}, "wood": {Value: 1}, "stone": {Value: 2}, "gold": {Value: 3}},
	"player-selector":    {"any-ally": {Value: -101}, "every-ally": {Value: -201}},
	"search-source":      {"search-local": {Value: 1}, "search-remote": {Value: 2}},
	"search-order":       {"search-order-none": {Value: 0}, "search-order-asc": {Value: 1}, "search-order-desc": {Value: 2}},
	"object-status":      {"status-pending": {Value: 0}, "status-ready": {Value: 2}, "status-resource": {Value: 3}, "status-down": {Value: 4}, "status-gather": {Value: 5}},
	"object-list":        {"list-active": {Value: 0}, "list-inactive": {Value: 1}},
	"formation":          {"formation-line": {Value: 2}, "formation-box": {Value: 4}, "formation-stagger": {Value: 7}, "formation-flank": {Value: 8}},
	"attack-stance":      {"stance-aggressive": {Value: 0}, "stance-defensive": {Value: 1}, "stance-stand-ground": {Value: 2}, "stance-no-attack": {Value: 3}},
	"object-data":        objectDataConstants(),
	"duc-action":         {"action-default": {Value: 0}, "action-move": {Value: 1}, "action-patrol": {Value: 2}, "action-guard": {Value: 3}, "action-follow": {Value: 4}, "action-stop": {Value: 5}, "action-ground": {Value: 6}, "action-garrison": {Value: 7}, "action-delete": {Value: 8}, "action-unload": {Value: 9}, "action-train": {Value: 10}, "action-gather": {Value: 11}, "action-lock": {Value: 12}, "action-work": {Value: 13}, "action-ungarrison": {Value: 14}, "action-drop-relic": {Value: 15}, "action-pack": {Value: 16}, "action-unpack": {Value: 17}, "action-none": {Value: 18}, "action-attack-move": {Value: 19, DEOnly: true, NeedsDefinition: true}, "action-transform": {Value: 20, DEOnly: true, NeedsDefinition: true}, "action-research": {Value: 21, DEOnly: true, NeedsDefinition: true}},
}

func hasValue(values []int64, value int64) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func referenceFamily(name string) string {
	if name == "include" || name == "xs-script-call" || strings.HasPrefix(name, "fe-") {
		return "DE"
	}
	if strings.HasPrefix(name, "up-") {
		return "UP"
	}
	return "AoC"
}

func (s commandSpec) signature(name string) string {
	if len(s.params) == 0 {
		return "(" + name + ")"
	}
	return "(" + name + " <" + strings.Join(s.params, "> <") + ">)"
}

func commandSource(name string) string {
	if name == "" {
		return catalogueSource
	}
	return "https://airef.github.io/commands/commands-details.html#" + name
}

func typoCandidates(name string) []string {
	var candidates []string
	for known := range perCatalogue {
		if oneEditApart(name, known) {
			candidates = append(candidates, known)
		}
	}
	sort.Strings(candidates)
	return candidates
}

func oneEditApart(a, b string) bool {
	if len(a) > len(b) {
		a, b = b, a
	}
	if len(b)-len(a) > 1 {
		return false
	}
	i, j, edits := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
			j++
			continue
		}
		edits++
		if edits > 1 {
			return false
		}
		if len(a) == len(b) {
			i++
		}
		j++
	}
	if i < len(a) || j < len(b) {
		edits++
	}
	return edits == 1
}
