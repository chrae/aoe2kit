# AI GUIDE

Read this before changing any bundled AoE2 artifact.

## Rules

1. One Go source tree, one local build command, no runtime deps, no network, no
   Python. If a task tempts you to write a Python/bash script for durable work,
   STOP. That means a Go tool is missing. Record the gap; do not script around
   it.

2. AoE2Kit is project-neutral. Kit code must not assume Decima, CBA, Scenario
   Sandbox, SDS local paths, or any other project-specific fact. Project facts
   belong in the bundle artifacts and bundle-local notes, not in shared kit code.

3. Read before you mutate. Every write path must follow:

```text
inspect -> dry-run/explain -> apply -> verify -> report evidence
```

## First Moves

```sh
go build -o kit ./cmd/kit
./kit doctor
./kit summary .
./kit inventory
./kit project inspect . --limit 50 --text
./kit recipe list --text
```

Use the output to identify what the bundle contains: scenarios, records, data
files, AI files, and local mod folders.

## Resource Guardrails

Kit sets a default soft Go heap limit of 1 GiB unless `GOMEMLIMIT` or
`AOE2KIT_GOMEMLIMIT_MB` is set. Scenario parsing refuses inflated bodies above
32 MiB by default because some community scenarios inflate to very large trigger
graphs. For an intentional large teardown, run one command at a time and set
`AOE2KIT_MAX_SCENARIO_MB` higher or use a command-specific cap such as
`kit scen effects huge.aoe2scenario --census --max-mb 192`,
`kit scen effects huge.aoe2scenario --where 105 --max-mb 192`,
`kit scen triggers huge.aoe2scenario --grep mana --max-mb 192`,
`kit scen glossary huge.aoe2scenario --max-mb 192`, or
`kit scen idioms file.aoe2scenario --text`. Do not use
`AOE2KIT_ALLOW_HUGE_SCENARIO=1` unless you genuinely need an unlimited full
parse.

## Engine Facts

Run `kit facts check --text` before relying on a handoff bundle's fact ledger.
Run `kit facts list --text` to see engine-verified rules and
`kit facts list --include-provisional --text` only when you need lower-tier
hypotheses. Scenario lint/deploycheck findings cite `fact_id`, tier, verified
date, and fixture reference where they are backed by the ledger.

## Current Write Boundary

Scenario writing is implemented as a controlled, typed recipe patcher, not a
general editor clone and not an unrestricted map generator. Scenario inspection
supports DE `1.55` through `1.58`, but writes currently target DE `1.58` only;
older readable scenario versions must be treated as read-only inputs. Current
surfaces:

- `players`: set player slot active/human state and AI name/type fields.
- `diplomacy`: set one player-to-player stance at a time.
- `resources`: set food, wood, stone, gold, and trade-goods values.
- `victory`: set authored GlobalVictory fields such as `conquest_required`.
  For lobby-proof diagnostics, set `conquest_required=0` in the scenario file;
  DE single-player scenario lobbies do not expose the random-map Victory
  dropdown.
- `xs`: embed an XS entry script payload and scenario XS name/path fields.
  Follow `docs/XS_AUTHORING.md`; runnable bundles should also ship the XS module
  tree under `resources/_common/xs/`.
- `units`: add, edit, and remove units in existing player unit sections.
- `map`: edit existing terrain tiles by rectangle, circle, or line. Tile
  terrain id, elevation, and layer are supported; map dimensions are not
  resized.
- `triggers`: add triggers, add display/create helper triggers, or edit an
  existing trigger selected by index/name. Edits can set enabled/looping/name
  and append supported effects/conditions while maintaining count and display
  order arrays.

Supported trigger effects are `display_instructions`, `display_timer`,
`send_chat`, `create_object`, `kill_object`, `remove_object`, `task_object`,
`change_ownership`, `change_object_name`, `change_object_hp`,
`teleport_object`, `change_object_stance`, `set_player_visibility` /
`reveal_map`, `research_technology`, `modify_attribute`, `modify_resource`,
`script_call`, `activate_trigger`, `deactivate_trigger`, and
`declare_victory`. Supported conditions are `timer`, `object_selected`,
`object_in_area` / `objects_in_area`, `own_objects`, `own_fewer_objects`,
`accumulate_attribute`, `object_visible`, and `variable_value`.

Data-mod writing is available for the current AoE2DE `.dat` patch surfaces
documented in `docs/DAT_ENGINE.md` and `docs/GO_AOE2KIT.md`: graphic record
strings/fixed fields; existing unit/common/Type50/Creatable fixed fields,
row patches, and action task-row patches; graphic/unit record creation;
codec-backed Effects/Techs/Civ/Terrain/Sound/PlayerColour recipe edits;
semantic unit tombstones; and semantic sound mutes. AoE2Kit does not target
obsolete `.dat` layouts unless a current workflow proves they matter.

If the needed write is outside that surface, record the Go-tool gap. Do not use
Python `genieutils` as a durable substitute.

Before authoring a new patch JSON from scratch, check `kit recipe list` and
`kit recipe show <name> --recipe-only` for a native scenario/DAT starting shape.
After generating an output artifact, use `kit project lineage --input PATH
--output PATH --manifest lineage.json` when the handoff needs explicit input and
output hashes.

For visual sprite inspection, use `kit gfx info/export` on `.sld` files. The
current exporter writes main-layer PNG frames for teardown; it does not compose
shadow, damage, or player-color layers into a final in-engine render.

## XS Authoring Helpers

Use `kit xs` for durable XS glue instead of generating anonymous bridge code by
hand:

- `kit xs bridge <variables.json> --out-dir DIR`: generate named
  `xsVariableN` accessors plus trigger-variable definition JSON and read/write
  lint.
- `kit xs shims <file-or-dir>... --out recipe.json`: inventory XS functions and
  emit trigger `script_call` recipe fragments using the effect `message` field.
- `kit xs datagen <arrays.json> --out module.xs`: generate xsArray data modules
  from structured JSON.
- `kit xs inspect <file.xsdat> [--types string,int,...]`: decode XS sidecar
  bytes written by `xsWriteString`, `xsWriteInt`, `xsWriteFloat`, and
  `xsWriteVector`. Without `--types`, this is heuristic because `.xsdat` files
  do not carry embedded type tags.
- `kit xsdat decode <file.xsdat> --ledger expected.json`: decode an XS sidecar
  and compare each value against an explicit expected payload ledger. Use this
  for assertion fixtures where missing, truncated, or wrong runtime writes must
  fail loudly.
- `kit xsdat decode <file.xsdat> --schema rtv12|rtv13|rtv14|rtv141|rtv142 --text`:
  decode the v12 Executioner Ledger, v13 Directed Attribution, v14 Castle Kill
  Calibration, or v14.1/v14.2 reduced-slot Castle Kill Calibration sidecar as
  named phase rows, including partial runs without a footer. `rtv142` is an
  alias for the v14.1 row layout because v14.2 intentionally reused the same XS
  sidecar format.
- `kit replay sidecar-sync <file.aoe2record> --xsdat run.xsdat (--ledger
  expected.json|--schema rtv12|rtv13|rtv14|rtv141|rtv142)`: correlate typed sidecar rows with
  nearby replay sync-matrix samples. Use `--schema` for named fixture sidecars
  such as v12 Executioner Ledger, v13 Directed Attribution, and v14/v14.1/v14.2
  Castle Kill Calibration. Treat unnamed-word and attribute-mirror output as
  candidates until cross-fixture evidence promotes it. Use `--fail-on-mismatch`
  for automated fixture gates; it fails on sidecar decode failures, missing sync
  samples, or known-word mismatches.

These are structural authoring tools. A real scenario run is still the oracle
for XS syntax, include resolution, and runtime behavior.

Before handing off an XS-backed scenario for an engine run, preflight the
scenario against the exact deployed mod/profile tree:

```sh
./kit scen xs path/to/file.aoe2scenario --deploy-tree path/to/deploy-root --text
./kit scen deploycheck path/to/file.aoe2scenario path/to/deploy-root --text
./kit ci init path/to/file.aoe2scenario --root path/to/project-ci --text
./kit ci check path/to/project-ci/aoe2kit-ci.json --text
./kit campaign generate docs/CAMPAIGN_SPEC_EXAMPLE.json --out-dir /tmp/campaign-xs --text
```

`scen xs` inventories attached XS, script_call effects, called function names,
function definitions, includes, declarations, and cross-file symbol hazards.
`scen deploycheck` gates handoff: it catches mismatched scenario XS filename
fields, missing deployed XS entry files, unresolved includes, and cross-file
non-`extern` symbol references before the engine opens its modal error box. It
is still
`structure_verified_not_engine_verified`; a clean deploycheck does not replace
the engine run.

For repeat playtests, use `kit ci` as the local loop gate. `ci init` creates a
portable check config, an example replay contract, and a small neutral debug XS
include. `ci check` composes `verify-run` and `replay sidecar-sync`; any failed
or unknown check exits nonzero. Keep project-specific assertions in the CI
config, not in Kit source.

For same-profile campaign persistence, use `kit campaign generate`. It produces
XS writer/reader modules that use the engine-verified bare writer-scenario
`.xsdat` read convention. This is only the same-party fast path; mixed-party
multiplayer continuity still needs explicit synchronized re-declaration logic.

For replay chat, use `kit replay chat <file.aoe2record> --text` when you need a
human/player transcript. Trigger-effect `send_chat` output is not retained in
recordings. Treat `source=backlog` rows as likely prior-session chat injected by
DE at the start of a new recording, not as live commentary from that match.

## Reports

When you finish work, report:

```text
what changed
commands run
evidence: counts, hashes, parsed artifacts, before/after checks
gaps found in AoE2Kit
```

Never report only "done."
