# AI source checks

`kit ai lint <file-or-folder> [--dat-db export.sqlite] [--dat-sha256 HASH]
[--scenario map.aoe2scenario] [--text]` checks `.ai`, `.per`, and `.per2`
sources offline. Errors cause exit status 1. Warnings preserve exit status 0;
`OK` means no errors from the implemented static checks, not engine acceptance.

The `de-coordination-v4` catalogue covers census probes and the first DUC micro rules:
timers, local/remote searches, filters, cleanup, object-data reads, target setup
and object/point orders, formatted chat, and the PER-to-XS entrypoint.
It also checks `defconst`, `defrule`, `include`, `load` signatures, and boolean
operators. Facts and actions are checked in their respective rule sections.
Quoted text and semicolon comments cannot introduce parentheses or commands.
Issue locations use one-based lines and byte columns after line-ending
normalization.

JSON output includes `catalogue` version/revision/source, `coverage` counts,
per-command `checks`, source `links`, and diagnostic `source`/`confidence`.
The catalogue is pinned to the AI Scripting Encyclopedia revision
`16b6b78a6d9eb0b2d23e48ffd0b261212cf3a8bd`:
[reference data](https://github.com/airef/airef.github.io/blob/16b6b78a6d9eb0b2d23e48ffd0b261212cf3a8bd/js/commands.js).
The checker does not fetch anything at runtime.

- `reference_backed`: the signature, placement, or range comes from the pinned
  guide. This is not an engine-verified claim.
- `static_analysis`: source parsing or a declaration found in project files.
- `inferred_typo`: a command outside this catalogue differs by one edit from
  a catalogued command; the diagnostic suggests that spelling.
- `unverified`: the checker cannot establish the value, linkage, or runtime
  resolution. Unknown commands outside the small catalogue remain warnings.
- `dat_projection`: identity rows from the selected, fingerprinted SQLite
  export; this does not establish effective scenario statistics.
- `scenario_structure`: decoded trigger fields in a fingerprinted scenario;
  this does not establish that a trigger executed.

For example, `(up-find-locl c: 83 c: 10)` reports an unknown-command error
with a suggested `up-find-local`. The recognized command with missing arguments
or quoted numeric operands reports an error with its expected signature.
`(up-find-local c: 83 c: 255)` warns about the guide's documented 0–240 result
range. Ranges are checked for literals/local constant aliases, not the runtime
contents of operands addressed through `g:` or `s:`.

Constants are resolved within each file; separate AI entrypoints are not merged.
Only the scoped builtin enum constants below are resolved. Other builtin
constants, PER load expansion, and preprocessor branch selection remain outside
coverage. They produce explicit unresolved/partial-coverage warnings.
Strategic-number names and builtin unit/class/line aliases are not yet
catalogued. Numeric unit operands can use the separate SQLite lookup below.

## DUC micro catalogue

The catalogue adds these reference-backed shapes; JSON `checks[].parameters`
contains argument kinds, ranges, allowed numeric values, sentinel exceptions,
and the supported named constants. `reference_family` records the guide's
AoC/UP/DE origin, not a claim of engine testing. Text shows each used signature
and its catalogue notes once.

| Commands | Covered details |
| --- | --- |
| `up-find-local`, `up-find-remote` | Four operands; local cap 0–240, remote cap 0–40. |
| `up-filter-distance` | Two typed operands, each documented from -1 to 32767. |
| `up-filter-status` | Four operands: type/status and type/object-list. Normal local/remote finds ignore this filter according to the guide. |
| `up-clean-search` | Search source, object-data selector or -1 for duplicate removal, search order. |
| `up-get-search-state` | Four-goal output block; command/generic upper-bound discrepancy retained in notes. |
| `up-get-object-data` | Partial named selectors for HP, maximum HP, distance, action, order, target and target id; output goal 1–16000. |
| `up-set-target-object` | Search source and typed index; conflicting command/generic list bounds retained in notes. |
| `up-set-target-point` | Goal-pair shape; conflicting upper-bound documentation prevents a fixed bound check. |
| `up-target-objects` | Target option, the documented action subset, formation, stance. |
| `up-target-point` | Point pair or current-point sentinel 0, action, formation, stance; distinct noncombat train/research overloads stay explicitly uncertain. |

Named search sources/orders, object status/list, formations/stances, a partial
object-data selector set, and DUC actions are scoped to their parameter kinds.
DE-only action values 19–21 are marked in the metadata and their names require
an explicit numeric `defconst`. Undefined names stay unresolved rather than
being silently synthesized. Formation/state availability, initialized search or
point state, focus-player selection, and the runtime effects of orders are not
validated. Numeric enum/range mismatches remain reference-backed warnings.

`testdata/ai/duc-micro.per` is a source-check fixture, separate from the passive
probe. It does not establish that active orders are safe or effective in a game.

### Object-data selector context

The relevant micro set is documented below. Each entry in
`checks[].parameters[].constants` includes its numeric value, reference family,
source, meaning, units, context notes and `runtime_availability: "unverified"`.
The broader ObjectData enum remains partial; this is not a claim that all
selectors or contexts are supported. These meanings summarize the pinned
[ObjectData reference](https://airef.github.io/parameters/parameters-details.html#ObjectData).

| Selector suffix (`object-data-`) | Value / family | Meaning and reference caveat |
| --- | --- | --- |
| `id`, `type`, `class` | 0, 1, 2 / UP | Map object id, unit-type selector, object class; distinct identity spaces. Type/upgrade behavior is scenario-dependent. |
| `action`, `order` | 5, 6 / UP | ActionId and OrderId. Guide reports unavailable/-1 values while marching in formation. |
| `target` | 7 / UP | Target **class**, not map id. Guide reports absent/formation values UP -1 versus DE -2. |
| `hitpoints`, `maxhp` | 10, 11 / UP | Current damaged HP versus maximum HP; maximum includes tech/civ effects. |
| `range` | 12 / UP | Current range including tech/civ effects; no independently validated numeric scale is asserted. |
| `status`, `player` | 19, 20 / UP | Object status and player number; source metadata does not establish live state or ownership. |
| `action-time` | 22 / UP | Last movement time in milliseconds; guide reports an attacking/reset ambiguity. Not a universal last-order timestamp. |
| `target-id` | 23 / UP | Target map object id; guide reports -1 without a target or during formation marching. |
| `formation-id` | 24 / UP | Formation state id; guide says it does not match the Formation command enum. |
| `distance`, `precise-distance` | 44, 45 / UP | Distance and distance×100 to a **normal** target point; obstacles ignored. Scaled output does not imply precise input coordinates. |
| `reload-time` | 54 / UP | Configured attack interval, milliseconds, including tech/civ effects; not remaining cooldown. |
| `next-attack` | 55 / UP | Remaining attack countdown in milliseconds, measured from animation start according to the guide; not proof of damage or projectile release. |
| `min-range` | 62 / UP | Current minimum range including tech/civ effects; scale remains reference-only. |
| `target-time` | 63 / UP | Milliseconds attacking the current target; guide reports -2 outside attacks and negative villager/herding cases. |
| `frame-delay` | 78 / UP | Animation frames before projectile launch; **not milliseconds**. Frame durations vary. |
| `attack-delay` | 90 / DE | Delay in milliseconds, a DE extension. The checker requires an explicit numeric `defconst` because builtin spelling availability is unverified. |

The `attack-delay` definition requirement is the checker's conservative policy,
not proof that every DE build lacks a builtin name. Even a recognized selector
and a valid output-goal shape do not establish a selected target, populated
return value, formation availability, return sentinel, or effective behavior.
Do not substitute static DAT stats for these runtime observations.
`testdata/ai/object-data.per` exercises all 22 documented selector entries offline.

## Coordination command shapes

Coordination checks add `mechanism` and `runtime_semantics: "unverified"` to
the existing source/family/parameter metadata. They validate message/state
shapes only, not a protocol or delivery behavior.

| Commands | Family / mechanism | Parameters and boundaries |
| --- | --- | --- |
| `shared-goal` | AoC / global shared goals | Shared id 1–256 and value; an equality check with **two** operands, no comparison operator. |
| `set-shared-goal` | AoC / global shared goals | Shared id 1–256 and signed 32-bit value. |
| `up-get-shared-goal` | UP / global shared goals | Type/id and local output goal 1–16000. Parameter range says 1–256 while the example uses 0; violations warn. |
| `up-set-shared-goal` | UP / global shared goals | Type/id and type/value; shared id 1–256, signed 32-bit DE value. |
| `set-signal` | AoC / legacy scenario signal | One signal id 0–255; guide describes setting an AI Signal flag. |
| `up-get-signal` | UP / legacy scenario signal | Type/id and local output goal. Signal id 0–255, output goal 1–16000. |
| `up-set-signal` | UP / legacy scenario signal | Type/id and type/value; signal id 0–255, signed 32-bit DE value. |
| `fe-set-signal` | DE / multiplayer scenario signal | Type/id and type/value; signal id 0–255, signed 32-bit value. Separate Multiplayer AI Signal condition path. |
| `chat-to-allies` | AoC / ally chat | Quoted text/local text constant. Guide reports leading numerals can invoke taunts. |
| `up-chat-data-to-all` | UP / all-player chat | Text, type, formatted value; **not ally-only**. |
| `up-chat-data-to-player` | UP / explicit recipient chat | Recipient, text, type, formatted value. Scoped recipient constants include `any-ally` (-101) and `every-ally` (-201); other selectors remain partial. |

The guide describes shared goals as shared across **all AIs, including enemy
AIs**, not private team state. No automatic team namespace is invented.
Legacy `set-signal`/`up-set-signal` are documented for single-player **AI Signal**
conditions; `fe-set-signal` is documented for **Multiplayer AI Signal** conditions
in single/multiplayer scenarios. The catalogue does not establish that
`up-get-signal` reads the FE multiplayer channel or that signals and shared
goals interoperate. Chat is not treated as synchronized machine state.

No delivery, ordering, atomicity, synchronization, recipient resolution,
acknowledgment, or runtime state is verified. There are no active unit orders
or changes to the passive probe. `testdata/ai/coordination.per` exercises these
shapes offline; it is not an in-game coordination canary.

## Unit operands and scenario context

Pass `--dat-db` an existing export from `kit dat sql`. The linter opens it
read-only and never expands the DAT, rebuilds a cache, or writes the database.
It records the export's source-DAT fingerprint/version in `dat_catalogue`.
`--dat-sha256` optionally supplies the expected source DAT fingerprint. A
mismatch warns and leaves identities unresolved; without an expected hash the
linter makes no claim that the selected export is current for an installation.

`unit_resolutions` records each typed unit operand, including its source
location, literal/local constant value, status, source, and confidence.
`dat_matches` preserves unit indices, internal DAT names, language string ids,
and the matching civilization indices. An internal name is not a localized
display name. Multiple names/unit indices remain explicitly ambiguous; no single
canonical name is invented. Different civ-specific language string ids with the
same internal identity use `dat_match_civ_variants` and retain both groups.
Repeated ids reuse query results.

A miss in the selected DAT is a warning, not rejection: builtin class/line
selectors, another DAT version, and opaque scenario uses remain possible.
Dynamic `g:`/`s:` operands and unresolved constants remain unresolved. Without
`--dat-db`, value records say `no_dat_catalogue_selected`; shape checking alone
does not establish an id's identity.

`--scenario` separately fingerprints and parses a scenario, attaching exact
trigger indices/names, enabled/looping flags, and effect/condition positions to
matching unit operands. Only decoded unit-reference fields are indexed; absent
or unsupported fields do not establish absence of a unit type. The references
use `scenario_structure` evidence, independent of `dat_projection` identity.
A role label in a trigger name never replaces the DAT internal name or creates
a new unit id. Trigger presence also does not prove execution, unit statistics,
or any engine behavior. Full reference lists are in JSON; text output shows up
to eight references per operand.

Static unit records are separate `scenario_placements`, with player-section
index, unit index, reference id, coordinates, raw status, and structure-only
confidence. `scenario_placement_count` is explicit, including zero when units
were indexed but none matched. It is omitted when no scenario unit section was
available or the operand is dynamic/unresolved; `placements_status` in the
scenario catalogue distinguishes indexed from unavailable data. Player-section
indices preserve the parser's numbering rather than inferring lobby ownership.
JSON retains every placement; text shows the count and first eight records.
Placement records do not establish that a unit was alive, owned, or active
at runtime. Trigger references remain in their own list, so a trigger-created
type can have zero static placements and still have scenario evidence.

Each unit-matched trigger reference includes the parser's structured `effect`
or `condition` record and the complete ordered `trigger_conditions` list. Array
positions in that list are the original condition indices, so sibling timer
conditions remain available even when they have no unit operand. Decoded
player, location, quantity/limit, and timer fields retain their raw values,
including zeros where the parser exposes them in `known_fields`. Missing fields
are not inferred. These records carry `scenario_structure` confidence and the
scenario catalogue's fingerprint; configured timers and thresholds are not
observed production cadence or runtime ownership.

Empty `.ai` markers get `loader_pairs` records: `present` means a same-basename
`.per`/`.per2` sibling exists, while `runtime_status` stays `unverified`.
An unresolved pair is distinguished from unverified engine/lobby loading.

For XS linkage, a packaged sibling `xs` directory beside an ancestor `ai`
directory supplies project candidates. The checker uses Kit's XS lexer to find
top-level function declarations, ignoring comments and strings, and checks
whether an identified call target requires parameters. Missing or ambiguous
functions stay unresolved because transitive XS includes are not expanded.
Finding a packaged file/function does not prove the game's XS include-root
resolution or XS compilation. Those remain explicit warnings; game installation
and mod loading can use different roots.

The generic passive-probe fixture is under `testdata/ai/passive`. Private AI
corpora can be checked explicitly by passing their location to the CLI; their
paths and script contents are not embedded in Kit.

`cc-add-resource` is catalogued as an AoC action with two constant operands:
`(cc-add-resource <stockpile-resource> <resource-amount>)`. Resources are food,
wood, stone and gold (0..3); the pinned reference documents amounts -32768..32767,
including negative deductions. Out-of-range values warn rather than claiming
engine clamping or overflow behavior. The reference says cheats need not be enabled.

`set-signal` notes link the one-run, single-player legacy signal bridge canary
recorded in [SCENARIO_WRITER.md](SCENARIO_WRITER.md). This scoped observation does
not change static signature confidence or mark general runtime semantics verified.
It supplies no multiplayer, `fe-set-signal`, resource synchronization or exact
resource-delta evidence.

Catalogue enum names establish numeric meanings, not identifier availability.
`constant_availability_unverified` warns when a recognized symbolic operand (or
its alias initializer) lacks a visible numeric `defconst` or verified builtin
source. Numeric operands and core stockpile resource names are accepted.
Loaded/include definitions remain unresolved, so this diagnostic has unverified
confidence and does not assert that the engine will reject the symbol. Conditional,
cyclic and ambiguous definitions still cannot establish availability.

Regression evidence: MCBCombatProbe.per2 line 21 was accepted by catalogue
`de-resource-v5`, but DE reported `ERR2005: Invalid identifier` for
`(up-clean-search search-remote object-data-distance search-order-asc)`.
The exact failing token among those three operands is unknown. The source line
is preserved in `testdata/ai/undefined-duc.per`; replacing symbols with numerics
was a local workaround, not engine verification of the numeric version.

When an in-game AI/XS/scenario error is reported, preserve the exact source and
first engine error, reproduce the relevant check with current Kit, and add a
minimal regression/diagnostic or document why static checking cannot establish
it. A script workaround alone does not close a lint gap. Line-only engine errors
must retain token uncertainty until a distinguishing probe supplies evidence.

`set-strategic-number` is an action with constant SN ID (0..511) and DE signed
32-bit value; individual SN value limits remain partial. `up-compare-goal` is a
fact taking goal ID (1..16000), comparison operator, and value. Operators are
`>`, `>=`, `<`, `<=`, `==`, `!=`, optionally prefixed with `c:`; `g:` and `s:`
select goal/SN addresses and require those prefixes. Address ranges are checked
separately from signed runtime values. Symbolic SN and player selectors outside
the current catalogue remain unresolved warnings. These checks use the pinned
reference and do not establish focus-player semantics, search visibility or
runtime evaluation.

Local goals: `(goal <goal-id> <goal-value>)` is an equality fact with no
comparison operator; `(set-goal <goal-id> <goal-value>)` is an action. Both
use constant operands, DE IDs 1..16000 and signed 32-bit values, with visible
numeric `defconst` aliases supported. These are local AI goals, separate from
shared goals, SNs and scenario trigger variables. Signature validation does not
verify initialization, phase ordering or runtime behavior. Other comparisons
use `up-compare-goal`. See `testdata/ai/local-goals.per` for a static phase gate.
