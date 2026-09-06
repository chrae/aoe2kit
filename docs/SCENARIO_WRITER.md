# Scenario Writer Architecture

AoE2Kit scenario writing uses the same honesty discipline as the DAT engine, but
not the same patch granularity. Current AoE2DE `.aoe2scenario` files are
sectioned, nested, count-heavy artifacts. For the sections AoE2Kit needs to
write first, especially triggers and units, there is no safe small byte poke.

## Contract

AoE2Kit targets current AoE2DE scenario layouts. It does not carry compatibility
complexity for obsolete scenario formats unless a current workflow proves they
matter.

The writer model is:

```text
read uncompressed file header
inflate scenario body
parse sequential sections enough to locate owned islands
fully model the owned island
rebuild the whole owned island
preserve every byte outside owned islands
reassemble and deflate body
reopen output and verify
```

For trigger work, "typed island" means a complete recursive trigger block model:
triggers, display order, variables, effect arrays, condition arrays, per-trigger
effect/condition display orders, strings, selected-object id tails, and all
count fields.

## Source Of Truth

AoE2ScenarioParser is the bring-up spec and golden oracle. Do not
reverse-engineer trigger/effect/condition field order by hand when the versioned
spec already exists.

Portable source references:

```text
AoE2ScenarioParser versions/DE/v1.58/structure.json
AoE2ScenarioParser versions/DE/v1.58/effects.json
AoE2ScenarioParser versions/DE/v1.58/conditions.json
AoE2ScenarioParser trigger manager implementation
AoE2ScenarioParser trigger data-object implementation
```

Python is not part of the durable AoE2Kit implementation. APar is acceptable as
a temporary external golden oracle while the Go writer is brought up:

```text
Go write -> APar read -> compare fields -> editor/game smoke when needed
```

## Owned Islands

Implement owned islands in value order:

1. `Triggers`: highest value and highest risk. Required for gates, stores,
   relays, narrator, trial logic, FX foundry, fireball launches, and swatches.
2. `Units`: required for gates, shops, relics, markers, bosses, NPCs, and
   arena imports.
3. `Map`: terrain/tile stamps and arena imports.
4. `Players/Diplomacy`: slot state, ownership, AI references, diplomacy.
5. Metadata/messages/options: scenario name, instructions, hints, camera, and
   similar small surfaces.

Unowned islands remain byte-preserved.

## Trigger Invariants

The trigger block must enforce these invariants before writing:

```text
number_of_triggers == len(trigger_data)
number_of_triggers == len(trigger_display_order_array)
trigger_display_order_array is a valid permutation of trigger indexes
number_of_effects == len(effect_data)
effect_display_order_array is empty or a valid permutation of effect indexes
number_of_conditions == len(condition_data)
condition_display_order_array is empty or a valid permutation of condition indexes
num_selected == len(selected_object_ids)
```

Forgetting the top-level trigger display-order array is a known corruption
class. APar updates it through `TriggerManager.trigger_display_order`; AoE2Kit
must do the equivalent explicitly.

## Raw Model First

The canonical Go model should mirror the raw scenario fields first. Builder
helpers are allowed only as a thin layer that emits raw structs.

The first helper set should cover the primitives that do most real work:

```text
display_instructions effect
create_object effect
kill_object effect
task_object effect
activate_trigger effect
deactivate_trigger effect
timer condition
object_selected condition
```

No larger DSL until the raw structs, serializer, and verification are stable.

## Verification

Every write path must prove:

```text
non-owned bytes are unchanged in the decompressed payload
output reparses to EOF
trigger/effect/condition/display-order counts are consistent
added or changed fields read back exactly
decompressed payload round-trips through deflate/inflate
APar golden oracle can read the written scenario during bring-up
```

Do not compare compressed bytes. Deflate output is not stable across encoders.

## First Implementation Slice

The first useful slice is larger than the first DAT slice:

1. Parse the uncompressed file header and inflate the scenario body.
2. Parse-to-locate sections through the trigger block using the current-DE APar
   `structure.json` as the field-order source.
3. Fully model triggers/effects/conditions/display-order arrays.
4. Rebuild the trigger block byte-identical and prove full decompressed payload
   equality before mutation.
5. Add one disabled trigger with a `display_instructions` effect, write to a
   separate output file, reparse, and verify all trigger invariants.
6. Immediately add one `create_object` trigger, because create-object is the
   value atom for the FX and scenario-building work.

The command surface should mirror DAT:

```sh
./kit scen triggers in.aoe2scenario
./kit scen plan in.aoe2scenario --recipe recipe.json
./kit scen patch in.aoe2scenario out.aoe2scenario --recipe recipe.json
./kit scen verify out.aoe2scenario
```

Recipe files should express the small helper primitives declaratively, while
the writer serializes raw structs.

## Current Implementation

M0 read/rebuild support exists in Go:

```sh
./kit scen triggers in.aoe2scenario
./kit scen units in.aoe2scenario
./kit scen map in.aoe2scenario
./kit scen verify in.aoe2scenario
./kit scen plan in.aoe2scenario --recipe docs/SCEN_RECIPE_EXAMPLE.json
./kit scen patch in.aoe2scenario out.aoe2scenario --recipe docs/SCEN_RECIPE_EXAMPLE.json
./kit scen smoke-recipe --x 145 --y 175
./kit scen smoke in.aoe2scenario /tmp/AoE2Kit_WriteSmoke.aoe2scenario --x 145 --y 175
```

`plan` is read-only, but it is not just a counter: it validates current
selectors and guarded delete dependencies for `remove_trigger`, `edit_trigger`,
`remove_unit`, and `edit_unit` before reporting the operation list.

The current reader supports DE `1.55` through `1.58`, embeds APar's `1.58`
`structure.json` as the main field-order source, and applies the known `1.55`
DataHeader player-data shape difference in memory. It parses the scenario body
through the `Files` section, extracts trigger summaries, enforces the read-side
trigger invariants above, and proves byte-identical decompressed body rebuild.
Scenario writing is intentionally narrower: Kit writes DE `1.57` and `1.58`;
older readable versions fail fast as read-only inputs.

Current write primitives are append-only helper triggers, unit placement, map
tile edits, player/diplomacy/resource setup, and XS embedding. See
`docs/SCEN_RECIPE_EXAMPLE.json` for a broad recipe and
`docs/SCEN_WRITE_SMOKE_RECIPE.json` for the canonical smoke fixture.

The broad recipe intentionally includes a small scenario-setup block as well as
trigger/unit/map edits:

```json
"scenario": { "player_count": 4 },
"victory": { "conquest_required": 0, "all_custom_conditions_required": 1, "mode": 4 },
"players": [{ "player": 1, "active": true, "human": true }],
"diplomacy": [{ "from": 1, "to": 2, "stance": 3 }],
"resources": [{ "player": 1, "food": 500, "wood": 500, "gold": 250, "stone": 100 }]
```

`scenario.player_count` writes the FileHeader player count advertised to the
engine/lobby. It must be kept in sync with the intended active non-Gaia player
slots. AoE2Kit also syncs the Units-section `number_of_players` field to
`player_count + 1` including Gaia when writing the scenario count. Player-slot
indexes elsewhere remain `0` for Gaia and `1..8` for normal player slots.

`kit scen blank` writes P0/Gaia inactive by default instead of inheriting the
seed scenario's Gaia state. Use `--gaia-active` only for fixtures that
deliberately need Gaia marked active. `--dummy-starters` creates one authored
Outpost per active normal player so DE does not fill empty active slots with
starter TC state.

`kit scen patch` refreshes FileHeader `timestamp_of_last_save` to the current
Unix time by default so generated scenario forks do not keep an inherited stale
date in the DE browser. Set `scenario.timestamp_of_last_save` explicitly in a
recipe when a deterministic or intentionally pinned timestamp is needed.

Player, diplomacy, and resource indexes use the scenario's player-slot indexes:
`0` is Gaia, `1..8` are normal player slots. Diplomacy stance values are the raw
scenario/editor values until a tighter portable enum table is added.

Supported effect ops include `display_instructions`, `display_timer`,
`send_chat`, `create_object`, `kill_object`, `remove_object`, `task_object`,
`change_ownership`, `change_object_name`, `change_object_hp`,
`teleport_object`, `change_object_stance`, `set_player_visibility` /
`reveal_map`, `research_technology`, `modify_attribute`, `modify_resource`,
`script_call`, `activate_trigger`, `deactivate_trigger`, and
`declare_victory`.
Supported condition ops include `timer`, `object_selected`, `object_in_area` /
`objects_in_area`, `own_objects`, `own_fewer_objects`,
`accumulate_attribute`, `object_visible`, and `variable_value`.

`edit_trigger` selects an existing trigger by `target_index` or `target_name`.
It can set `enabled`, `looping`, and `set_name`, and can append the same
supported effect/condition ops while updating that trigger's local counts and
display-order arrays. It can also remove existing child rows with
`remove_effects` and `remove_conditions`, or clear the child lists with
`clear_effects` and `clear_conditions`; these operations repair
`number_of_effects`, `number_of_conditions`, and both local display-order arrays.
Use `replace_effects` and `replace_conditions` when the intended operation is
to replace an entire child list with a new declared list; replacement cannot be
mixed with remove/clear/append for the same child list.
`copy_trigger` selects a source trigger by `target_index` or `target_name`,
appends a raw clone, and then applies the same rename, enabled/looping, child
append/remove/clear/replace edits to the clone. Trigger-id references inside the
clone are preserved exactly; rewrite them explicitly after copying if needed.
`remove_trigger` deletes one existing trigger by `target_index` or `target_name`;
`clear_triggers` removes the whole trigger block before adding replacements. Both
repair scenario trigger counts and the trigger display-order array. `remove_trigger`
also protects trigger-control references: if another trigger references the
removed trigger id, the write fails; references above the removed index are
decremented so they keep pointing at the same surviving trigger after the index
shift.

The `variables` recipe block supports `add_variable`, `edit_variable`,
`remove_variable`, and `tombstone_variable`. `add_variable` appends a
`VariableStruct`, chooses the next free id unless `id` is supplied, and updates
`number_of_variables`. `edit_variable` targets an existing variable by
`target_id`/`id`, `target_name`, or `name` and rewrites only `variable_name`.
`remove_variable` physically removes an unreferenced variable record and updates
`number_of_variables`; surviving variable ids are explicit and are not
renumbered. `tombstone_variable` preserves the record but renames it to
`_DeletedVariable<ID>` by default. Both delete forms refuse to run while trigger
effects or conditions still reference that id. Physical variable removal is
structure-verified by write/reopen tests; treat it as not engine-verified until
a DE editor/game smoke confirms the scenario loads.

The `strings` recipe block supports `add_string`, `set_string`, `clear_string`,
and `tombstone_string` for the fixed 32-slot `PlayerDataTwo.strings` table.
`add_string` uses the first empty slot unless `id` is supplied, `set_string`
rewrites a known slot, `clear_string` empties an unreferenced slot so it can be
reused, and `tombstone_string` preserves the id while replacing the text with
`_DeletedString<ID>` by default. `old_text` can be supplied as a guard for
set/clear/tombstone operations, and both delete forms refuse to run while
scenario trigger effects still reference the slot by `string_id`. AoE2Kit does
not compact string ids because that would renumber later slots.

XS support follows the current AoE2DE UGC Guide model plus the parser-style
embedded carrier pattern; see `docs/XS_AUTHORING.md`. The preferred generated
recipe uses `xs.mode:"carrier"` to store full XS source in a disabled trigger's
`script_call` message, avoiding scenario/file drift. `xs.mode:"attachment"` keeps
the older `Map.script_name` + `Files.script_file_*` path for deliberate external
module deployment, and `xs.mode:"attachment_and_carrier"` writes both. Ordinary
trigger `script_call` effects invoke parameterless XS functions through the
effect `message` field.

`add_unit` appends one `UnitStruct` to an existing player unit section and
updates that section's `unit_count`. `edit_unit` and `remove_unit` target
existing units by `reference_id`, by `target_player` + `target_index`, or by a
unique `target_caption`. Caption targeting fails loudly if more than one unit
matches; add `target_player` to disambiguate. `copy_units_in_area` clones every
matched unit inside `target_area_x1/y1/x2/y2`, optionally narrowed by
`target_player` and `target_unit_const`, offsets positions by `offset_x` /
`offset_y` or by placing the source area's top-left at `target_x,target_y`,
assigns fresh reference ids from `reference_id_base` or the next free id, can
move the copies with `set_player`, and can append `caption_suffix`.
Garrison links are preserved only when the referenced container is also copied;
otherwise the copied unit is ungarrisoned to avoid pointing back at the source
set. `move_units_in_area` relocates matched units by `offset_x` and/or
`offset_y`, or by placing the source area's top-left at `target_x,target_y`,
without changing owner or reference ids, and refuses moves that would place any
matched unit outside the map. `edit_units_in_area` applies bulk
field edits to matched units: `unit_const`, `status`, `rotation`,
`initial_animation_frame`, `garrisoned_in_id`, `caption_string_id`,
`caption_string`, and/or `set_player`. Position edits are intentionally kept in
`edit_unit` and `move_units_in_area`. `remove_units_in_area` physically deletes
every matched unit in the same kind of area selector; the whole operation fails
before mutation if any matched unit is still referenced.
`edit_unit` can move an existing unit to another owner
section with `set_player`, and the writer repairs both source and destination
`unit_count` fields. Unit delete operations also protect direct trigger object
references before deletion: selected-object effect lists, condition
`unit_object` / `next_object`, and location object references are checked, and
the write fails with the exact trigger/child/field path if a reference would
dangle.

`set_terrain_rect` edits existing terrain tiles in an inclusive rectangle. It
can set `terrain_id`, `elevation`, and/or `layer`; it does not resize maps.
`set_terrain_circle` uses `x1,y1` as the center and requires `radius`.
`set_terrain_line` uses `x1,y1` to `x2,y2` with Bresenham tile selection.
`set_terrain_border` paints only the perimeter of an inclusive rectangle and
accepts `thickness` with a default of 1 tile.
`copy_terrain_area` copies an inclusive source rectangle to `target_x,target_y`,
preserving terrain id, elevation, layer, and reserved tile bytes.
All map operations validate source and destination bounds. Paint operations can
set `terrain_id`, `elevation`, and/or `layer`.

`settings` is a read-only audit view for player slots, AI names/types,
resources, diplomacy, allied-victory flags, and global victory fields.

The `players` recipe array edits fixed player slots. Supported fields are
`active`, `human`, `tribe_name`, `civilization`, `lock_civilization`,
`lock_personality`, `ai_name`, and `ai_type`.

The `diplomacy_options` recipe block edits section-wide diplomacy settings:
`lock_teams`, `allow_players_choose_teams`, `random_start_points`,
`max_number_of_teams`, and per-player `allied_victory` flags. Pairwise stance
edits still use the existing `diplomacy` array with `from`, `to`, and `stance`.

`delete-plan` is read-only and emits the safest available recipe for a target.
For units, triggers, and unreferenced variables it can propose physical removal
when known direct references are clear. For string ids it proposes fixed-slot
clearing rather than compaction, because the table index remains the
engine-facing contract. Trigger-local child rows are supported with `effect <trigger>:<row>`
and `condition <trigger>:<row>` targets; these emit `edit_trigger` recipes with
`remove_effects` or `remove_conditions`. Child rows are index-addressed rather
than stable-id-addressed, so inspect the trigger again before applying an old
plan. Whole-trigger deletes that include effect type 55 `script_call` now add a
semantic warning to the plan: the row may be structurally removable, but AoE2Kit
cannot prove external XS/module expectations still make sense after the call is
removed.

`smoke-recipe` prints a project-neutral editor-smoke recipe. `smoke` applies it
directly. The smoke creates one visible terrain marker, one visible unit, and
three disabled trigger primitive groups; it remains
`structure_verified_not_engine_verified` until the output opens in the DE
editor/game and the visible markers are confirmed.

When appending triggers, `patch` updates `Triggers:number_of_triggers`,
`Triggers:trigger_display_order_array`, `Options:number_of_triggers`, and
`FileHeader:trigger_count`, then reopens the output and verifies rebuild and
trigger invariants.
