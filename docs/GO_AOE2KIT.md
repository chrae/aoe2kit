# Go AoE2Kit

`aoe2kit` is the durable Go tool spine for general AoE2 scenario, replay,
data-mod, mod-packaging, and sandbox work. It lives under `SDS/AoE2Kit`, but
the module is not SDS-specific.

The rule is:

```text
If a workflow is real, repeated, or release-adjacent, it belongs in Go.
Python is not part of the durable toolchain.
```

## Shape

The tools are split by AoE2 object domain, with shared packages underneath:

```text
cmd/kit        main CLI: doctor/version/manifest/verify/inventory/pack plus scen/dat/mod/swatch/cba
cmd/scen       optional developer entrypoint for .aoe2scenario inspection
cmd/dat        optional developer entrypoint for empires*.dat inspection
cmd/mod        optional developer entrypoint for local mod folder checks
cmd/swatch     optional developer entrypoint for Scenario Sandbox geometry/pattern work
cmd/cba        optional developer entrypoint for Castle Blood Automatic analysis

pkg/aoe2       generic AoE2 file metadata and shared vocabulary
pkg/cba        CBA-specific replay interpretation built from generic replay facts
pkg/datfile    current-AoE2DE span/index .dat inspection and patch engine
pkg/geom       map-tile geometry and moving pattern generators
pkg/kit        portable bundle inventory, verification, and packaging
pkg/modpack    local mod-folder inspection/checking
pkg/scenario   current-AoE2DE scenario section parser and trigger verifier
pkg/xs         XS authoring generators: named variable bridge, script_call shims, xsArray data
```

`kit pack` writes a profile-aware portable zip containing Go source, docs,
manifests, registries, recipes, and other non-ignored artifacts. The public
profile is source-only and excludes internal diagnostics/task notes. The sandbox
profile may include the one allowed ignored artifact: root `kit`, a measured
static linux-amd64 binary for no-toolchain AI sandboxes. It verifies the kit
before packing and reports the archive SHA-256. Ship that SHA as a sidecar file
next to the zip.

## Current Commands

Build:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -o kit ./cmd/kit
```

Portable bundle:

```sh
./kit version
./kit doctor
./kit docs lint docs --text
./kit summary .
./kit inventory
./kit manifest
./kit credits --text
./kit credits notice > NOTICE.md
./kit verify .
./kit pack AoE2Kit-public.zip --profile public --sha-sidecar --update-manifest
./kit portable-check . --profile public
./kit project inspect . --limit 50 --text
./kit project snapshot . --out project-snapshot.json --text
./kit project diff before-project-snapshot.json after-project-snapshot.json --text
./kit project lineage --input source.aoe2scenario --output built.aoe2scenario --tool "kit scen patch" --manifest built.lineage.json --text
./kit recipe list --text
./kit recipe show scen.xs-carrier --recipe-only
./kit release-check path/to/file.aoe2scenario --mod path/to/moddir --status path/to/mod-status.json
./kit release-check path/to/file.aoe2scenario --previous old.aoe2scenario --pack /tmp/AoE2Kit.zip
./kit verify-run path/to/replay.zip --contract docs/VERIFY_RUN_CONTRACT_EXAMPLE.json --text
./kit ci init path/to/file.aoe2scenario --root path/to/project-ci --text
./kit ci check path/to/project-ci/aoe2kit-ci.json --text
./kit campaign generate docs/CAMPAIGN_SPEC_EXAMPLE.json --out-dir /tmp/campaign-xs --text
./kit identify path/to/file.aoe2scenario
./kit identify path/to/file.aoe2record --registry known_scenarios.json
./kit identify path/to/replay.zip --registry known_scenarios.json
./kit identify path/to/file.aoe2scenario --register "Scenario Name" --family "Scenario Family" --description "What a fresh AI should understand on sight"
./kit identify path/to/file.aoe2record --ai-registry known_ais.json
./kit collect path/to/replay-folder --registry known_scenarios.json
./kit collect path/to/replay-folder --registry known_scenarios.json --dry-run
./kit player stats 5725572 --text
./kit player stats 5725572 --match-type 3 --cache-dir .aoe2kit-cache
./kit ai fingerprint path/to/ai-or-folder
./kit ai fingerprint path/to/ai-or-folder --registry known_ais.json --register "AI Name" --signature AI_FILES_SIGNATURE
./kit ai lint path/to/ai-or-folder
./kit ai diff before-ai-folder after-ai-folder
./kit xs bridge variables.json --out-dir build/xs_bridge
./kit xs shims resources/_common/xs --out build/xs_shims.recipe.json
./kit xs datagen arrays.json --out resources/_common/xs/generated_arrays.xs
./kit xs inspect run.xsdat --types string,int,string,int --text
./kit xsdat decode run.xsdat --ledger expected_ledger.json --text
./kit dat semantics-readback DAT_COMMAND_SEMANTICS_EXPECTED.json --dat empires2_x2_p1.dat --scenario file.aoe2scenario --replay run.aoe2record --xsdat run.xsdat --out DAT_COMMAND_SEMANTICS_READBACK.md --text
```

`kit verify` checks source buildability with `go build ./...`, required docs,
manifest drift, inventory, local mod folders, and bundled `.dat` files by
parsing them through AoE2Kit's current-DE DAT inspector. `kit portable-check`
adds profile-specific checks: public source archives must exclude stale binaries,
while sandbox/full profiles require a bundled linux-amd64 binary presence/hash
check.

All green kit checks are explicit structural/package checks unless the report
says otherwise. `verification.engine_verified=false` means the game/editor has
not been used as the oracle.

Engine facts:

```sh
./kit facts check --text
./kit facts list --domain xs --text
./kit facts list --include-provisional --text
```

`kit facts` exposes the Kit's evidence ledger. Each fact has a tier, date, and
fixture/readback citation. `kit scen lint` and `kit scen deploycheck` cite this
ledger in findings where a deterministic rule is backed by an engine fact.
Engine-verified facts are active by default; lower-tier facts stay informational
unless a command supports and receives `--include-provisional`.

Scenario:

```sh
./kit scen info  path/to/file.aoe2scenario
./kit scen check path/to/file.aoe2scenario
./kit scen blank build/BlankRoot.aoe2scenario --players 4 --dummy-starters --text
./kit scen triggers path/to/file.aoe2scenario
./kit scen units path/to/file.aoe2scenario
./kit scen units path/to/file.aoe2scenario --named
./kit scen strings path/to/file.aoe2scenario
./kit scen refs path/to/file.aoe2scenario --kind unit --id 910101 --text
./kit scen delete-plan path/to/file.aoe2scenario unit 910101 --text
./kit scen effects path/to/file.aoe2scenario
./kit scen effects path/to/file.aoe2scenario --raw
./kit scen effects path/to/huge.aoe2scenario --census --max-mb 192
./kit scen effects path/to/huge.aoe2scenario --where 105 --limit 25 --max-mb 192 --text
./kit scen triggers path/to/huge.aoe2scenario --grep mana --limit 25 --max-mb 192 --text
./kit scen triggers path/to/file.aoe2scenario --effect create_object --player 3 --area 10,10,20,20 --text
./kit scen trigger-neighborhood path/to/file.aoe2scenario --unit-ref 75382 --depth 2 --text
./kit scen audit-player-coverage path/to/file.aoe2scenario --players 1,2,3,4 --text
./kit scen mechanic path/to/file.aoe2scenario --kind garrison-token --grep heal --text
./kit scen trigger-flow path/to/file.aoe2scenario --grep heal --text
./kit scen glossary path/to/huge.aoe2scenario --limit 25 --max-mb 192 --text
./kit scen idioms path/to/file.aoe2scenario --text
./kit scen analyze path/to/file.aoe2scenario
./kit scen map path/to/file.aoe2scenario
./kit scen terrain path/to/file.aoe2scenario --id 50 --text
./kit scen palette-usage path/to/file.aoe2scenario --dat path/to/empires2_x2_p1.dat --text
./kit scen regions path/to/file.aoe2scenario
./kit scen describe path/to/file.aoe2scenario
./kit scen describe path/to/file.aoe2scenario --section triggers --full --json
./kit scen verify path/to/file.aoe2scenario
./kit scen lint path/to/file.aoe2scenario --text
./kit scen xs path/to/file.aoe2scenario --deploy-tree path/to/mod-or-profile-root --text
./kit scen xs attach path/to/file.aoe2scenario out.aoe2scenario --xs path/to/script.xs --name Entry.xs --text
./kit scen xs deploy out.aoe2scenario path/to/mod-or-profile-root --force --check --text
./kit scen xs embed path/to/file.aoe2scenario out.aoe2scenario --xs path/to/script.xs --replace-trigger 0 --text
./kit scen xs compare out.aoe2scenario path/to/script.xs --text
./kit scen xs extract out.aoe2scenario --out /tmp/extracted.xs --force --text
./kit scen deploycheck path/to/file.aoe2scenario path/to/mod-or-profile-root --text
./kit scen diff before.aoe2scenario after.aoe2scenario
./kit scen write-check before.aoe2scenario after.aoe2scenario --text
./kit scen plan path/to/file.aoe2scenario --recipe docs/SCEN_RECIPE_EXAMPLE.json
./kit scen patch path/to/file.aoe2scenario out.aoe2scenario --recipe docs/SCEN_RECIPE_EXAMPLE.json
./kit scen smoke-recipe --x 145 --y 175
./kit scen smoke path/to/file.aoe2scenario /tmp/AoE2Kit_WriteSmoke.aoe2scenario --x 145 --y 175
./kit scen patch path/to/file.aoe2scenario /tmp/AoE2Kit_WriteSmoke.aoe2scenario --recipe docs/SCEN_WRITE_SMOKE_RECIPE.json
```

Scenario recipe trigger conditions currently include `timer`, `object_selected`,
`object_in_area` / `objects_in_area`, `own_objects`, `own_fewer_objects`,
`accumulate_attribute`, `object_visible`, and `variable_value`.

XS authoring:

```sh
./kit xs bridge variables.json --out-dir build/xs_bridge
./kit xs shims resources/_common/xs --out build/xs_shims.recipe.json --prefix "Quest Shim"
./kit xs datagen arrays.json --out resources/_common/xs/zone_arrays.xs
./kit xs inspect run.xsdat --types string,int,string,int --text
./kit xsdat decode run.xsdat --ledger expected_ledger.json --text
```

`kit xs bridge` generates named wrappers around anonymous trigger variables:
`extern int xsVariableN`, `extern const int` named constants, and `get_name()`
functions that call `xsTriggerVariable(N)`. Its lint report flags variables
read but never written, written but never read, or referenced without
declaration.

`kit xs shims` scans runtime modules for XS functions and emits scenario recipe
triggers using effect type 55 `script_call` with the call stored in the effect
`message` field. Parameterized functions are skipped by default.

`kit xs datagen` converts structured JSON arrays into GoKu-style `xsArrayCreate`
and `xsArraySet` initialization modules. The command verifies the generated
shape against the JSON input, but engine execution remains a separate run.

`kit xs inspect` decodes `.xsdat` sidecar files using the documented XS File I/O
byte encodings. Use `--types` when the writer order is known. Auto mode is
heuristic because `.xsdat` files store raw values without embedded type tags.

`kit xsdat decode` layers assertion-fixture behavior on top of the same sidecar
decoder. With `--ledger`, it infers the expected write order, prints pass/fail
rows, and exits non-zero for missing, truncated, or mismatched values.
Use `--schema rtv12` for the v12 Executioner Ledger sidecar, `--schema rtv13`
for the v13 Directed Attribution sidecar, `--schema rtv14` for the v14 Castle
Kill Calibration sidecar, or `--schema rtv141`/`rtv142` for the v14.1/v14.2
reduced-slot Castle Kill Calibration sidecar; these name the wide phase rows
directly and accept partial runs that did not write the close footer. `rtv142`
is an alias for the v14.1 row layout because v14.2 intentionally reused the
same XS sidecar format.

Replay:

```sh
./kit replay info path/to/file.aoe2record
./kit replay info path/to/file.aoe2record --text
./kit replay info path/to/file.aoe2record --full
./kit replay graph path/to/file.aoe2record
./kit replay graph path/to/replay.zip
./kit replay triggers path/to/file.aoe2record --effect create_object --player 2 --unit-const 83 --text
./kit replay triggers path/to/file.aoe2record --condition accumulate_attribute --player 5 --text
./kit replay trigger-neighborhood path/to/file.aoe2record --trigger 120 --depth 2 --text
./kit replay diff-triggers before.aoe2record after.aoe2record --text
./kit replay chat path/to/file.aoe2record --text
./kit replay fetch 493687875 --profile 5725572 --out save-analysis/fetched --text
./kit replay fetch 493687875 --all-povs --profile 5725572 --out save-analysis/fetched --text
./kit replay fetch 493687875 --all-povs --from-roster path/to/replay.zip --out save-analysis/fetched --text
./kit replay coverage path/to/replay.zip --text
./kit replay opaque-spans path/to/replay.zip --min-bytes 64 --limit 20 --text
./kit replay frontier path/to/replay.zip --text
./kit replay corpus path/to/replay-folder --text
./kit replay opaque-clusters path/to/replay-folder --min-bytes 256 --limit 40 --text
./kit replay effective-data path/to/replay.zip --dat path/to/empires2_x2_p1.dat --player 2 --limit 40 --text
./kit replay effective-data path/to/replay.zip --dat path/to/empires2_x2_p1.dat --player 2 --tree
./kit replay effective-data path/to/replay.zip --known-template SHA256 --text
./kit replay effective-units path/to/replay.zip --dat path/to/empires2_x2_p1.dat --text
./kit replay effective-units path/to/replay.zip --dat path/to/empires2_x2_p1.dat --include-duplicates --limit 20 --text
./kit replay datamod-check path/to/replay.zip --baseline-replay path/to/vanilla_anchor.aoe2record --dat path/to/empires2_x2_p1.dat --text
./kit replay diff-state before.aoe2record after.aoe2record --focus opaque --limit 20 --text
./kit replay scan-value path/to/replay.zip --value 123456789 --width 32 --string SDSDBG --limit 50 --text
./kit replay header-anchors path/to/replay.zip --text
./kit replay events path/to/file.aoe2record
./kit replay events path/to/replay.zip --text
./kit replay events path/to/replay.zip --all
./kit replay events path/to/replay.zip --context docs/REPLAY_CONTEXT_EXAMPLE.json
./kit replay actions path/to/replay.zip --text
./kit replay actions path/to/replay.zip --action-id 3 --raw
./kit replay player-events path/to/replay.zip --text
./kit replay player-events path/to/replay.zip --player 1 --action-id 0 --from 02:00 --to 04:00 --limit 50
./kit replay player-events path/to/replay.zip --type chat --phase pre_game
./kit replay objects path/to/replay.zip --referenced --text
./kit replay object-state path/to/replay.zip --referenced --limit 20 --text
./kit replay object-state path/to/replay.zip --referenced --class gate --text
./kit replay object-shapes path/to/replay.zip --text
./kit replay spawns path/to/replay.zip --text
./kit replay lifecycle path/to/replay.zip --text
./kit replay postgame path/to/replay.zip --text
./kit replay camera path/to/replay.zip --text
./kit replay camera path/to/replay.zip --limit 0 --tail 4
./kit replay sync path/to/replay.zip --text --checksums
./kit replay sync path/to/replay.zip --raw-words --json
./kit replay checksum-phase path/to/replay.zip --word 9 --player 1 --text
./kit replay checksum-phase path/to/replay.zip --all-words --text
./kit replay checksum-probe path/to/replay.zip --player 1 --state-sum-delta 3 --carry-sum-delta 5 --text
./kit replay combat path/to/replay.zip --text
./kit replay combat path/to/replay.zip --window-ms 10000 --min-net-loss 10
./kit replay combat path/to/replay.zip --sort time --limit 0 --text
./kit replay deaths path/to/replay.zip --text
./kit replay xs-telemetry path/to/v3-probe.aoe2record --text
./kit replay carrier path/to/probe.aoe2record --ledger expected_ledger.json --text
./kit replay sidecar-sync path/to/probe.aoe2record --xsdat run.xsdat --ledger expected_ledger.json --fail-on-mismatch --text
./kit replay sidecar-sync path/to/probe.aoe2record --xsdat run.xsdat --schema rtv14 --text
./kit replay actions path/to/replay.zip --unknown-only --samples 3 --text
./kit replay postgame-corpus path/to/folder --text
./kit replay opaque-target path/to/replay.zip --text
./kit replay feedback path/to/file.aoe2record
./kit replay feedback path/to/replay.zip --text
./kit replay playtest path/to/replay.zip --text
./kit replay playtest path/to/replay.zip --window 3 --include-computers --json
./kit replay story path/to/replay.zip
./kit replay story path/to/replay.zip --brief
./kit replay story path/to/replay.zip --context docs/REPLAY_CONTEXT_EXAMPLE.json --brief
./kit replay player-profile path/to/replay.zip
./kit replay player-profile path/to/replay.zip --text
./kit replay player-profile path/to/replay.zip --context docs/REPLAY_CONTEXT_EXAMPLE.json --window-sec 60 --dead-gap-sec 10 --include-events
./kit replay player-series path/to/replay.zip --player 1 --changes-only
./kit replay player-series path/to/replay.zip --text
./kit replay inbox path/to/replay-folder --text
./kit replay telemetry path/to/replay.zip --schema docs/TELEMETRY_SCHEMA_EXAMPLE.json --text
./kit replay telemetry path/to/replay-folder --schema docs/TELEMETRY_SCHEMA_EXAMPLE.json
./kit replay issues path/to/replay.zip --context docs/REPLAY_CONTEXT_EXAMPLE.json --text
./kit replay unknowns path/to/replay-folder --text
```

`kit replay info` is intentionally compact by default. It reports the same
routine metadata surface as `kit replay summary`: roster, scenario identity,
data-set identity, lobby settings, map hash, duration, and known result fields.
Use `--full` only for forensic work that needs the raw parsed replay object; on
large custom scenarios that can include multi-megabyte trigger graphs and other
heavy structures.

CBA:

```sh
./kit cba replay progression path/to/replay.zip --text
./kit cba replay razes path/to/replay.zip --text
./kit cba replay perf path/to/replay.zip
./kit cba replay perf path/to/replay.zip --text
./kit cba balance --text
./kit cba trigger-razes path/to/replay.zip --dat path/to/empires2_x2_p1.dat --text
./kit cba trigger-spawns path/to/replay.zip --dat path/to/empires2_x2_p1.dat --text
```

Release gate:

```sh
./kit release-check path/to/file.aoe2scenario
./kit release-check path/to/file.aoe2scenario --previous previous.aoe2scenario --text
./kit release-check path/to/file.aoe2scenario --mod path/to/moddir --status path/to/mod-status.json
./kit release-check path/to/file.aoe2scenario --mod path/to/moddir --pack /tmp/AoE2Kit.zip
```

Data file:

```sh
./kit dat info  path/to/empires2_x2_p1.dat
./kit dat check path/to/empires2_x2_p1.dat
./kit dat diff vanilla.dat modded.dat --text --limit 20
./kit dat diff vanilla.dat modded.dat --section units --text --limit 20
./kit dat civs path/to/empires2_x2_p1.dat
./kit dat spans path/to/empires2_x2_p1.dat
./kit dat graphics path/to/empires2_x2_p1.dat
./kit dat graphics path/to/empires2_x2_p1.dat --limit 20
./kit dat graphics path/to/empires2_x2_p1.dat --particle
./kit dat graphics path/to/empires2_x2_p1.dat --name-contains smoke
./kit dat palette path/to/empires2_x2_p1.dat --min-variants 16 --text
./kit dat graphic path/to/empires2_x2_p1.dat 3396
./kit dat graphic path/to/empires2_x2_p1.dat 3396 --spans
./kit dat sprite path/to/unit.sld --out /tmp/unit_frames --text
./kit gfx info path/to/unit.sld --text
./kit gfx export path/to/unit.sld --out /tmp/unit_frames --limit 8 --text
./kit dat effects path/to/empires2_x2_p1.dat --limit 20
./kit dat effects path/to/empires2_x2_p1.dat --command-type 102 --operand 244 --limit 20
./kit dat effect path/to/empires2_x2_p1.dat 1
./kit dat effect-explain path/to/empires2_x2_p1.dat 1 --text
./kit dat command-matrix path/to/empires2_x2_p1.dat --text
./kit dat techs path/to/empires2_x2_p1.dat --limit 20
./kit dat techs path/to/empires2_x2_p1.dat --name-contains fire
./kit dat tech path/to/empires2_x2_p1.dat 2
./kit dat tech-explain path/to/empires2_x2_p1.dat 2 --text
./kit dat tech-tree path/to/empires2_x2_p1.dat
./kit dat tech-tree path/to/empires2_x2_p1.dat --full
./kit dat refs path/to/empires2_x2_p1.dat tech 244
./kit dat units path/to/empires2_x2_p1.dat --civ 0 --limit 20
./kit dat units path/to/empires2_x2_p1.dat --name-contains castle --class building
./kit dat unit path/to/empires2_x2_p1.dat 0 83
./kit dat patch-unit in.dat out.dat 0 83 --hit-points 26
./kit dat patch-unit in.dat out.dat 0 83 --standing-graphic 1284 --dying-graphic 1281
./kit dat patch-unit in.dat out.dat 0 83 --type50-attack-graphic 1279 --type50-blast-damage 2
./kit dat patch-unit in.dat out.dat 0 83 --train-time 26 --train-button 2 --train-hotkey 16122
./kit dat patch-unit in.dat out.dat 0 83 --creatable-button-icon 123 --creatable-hotkey-action 456
./kit dat patch-unit in.dat out.dat 0 83 --type50-attack 2,value=9 --type50-armour 2,value=4
./kit dat patch-unit in.dat out.dat 0 83 --attribute 1,amount=3 --cost 0,amount=55
./kit dat patch-unit in.dat out.dat 0 83 --task 0,action_type=3,work_range=2.25,combat_level=1
./kit dat patch-graphic in.dat out.dat 3396 --particle-effect-name sds_fireball_test
./kit dat patch-graphic in.dat out.dat 3396 --file-name p_mangonel_x1_sds
./kit dat patch-graphic in.dat out.dat 3396 --slp 4501
./kit dat patch-graphic in.dat out.dat 3396 --frame-count 31
./kit dat semantics-pack path/to/empires2_x2_p1.dat docs
./kit dat semantics-pack path/to/empires2_x2_p1.dat docs --feature local-building-effects
./kit dat semantics-readback DAT_COMMAND_SEMANTICS_EXPECTED.json --dat empires2_x2_p1.dat --scenario file.aoe2scenario --replay run.aoe2record --xsdat run.xsdat --out DAT_COMMAND_SEMANTICS_READBACK.md --text
./kit dat plan in.dat --recipe docs/DAT_RECIPE_EXAMPLE.json
./kit dat patch in.dat out.dat --recipe docs/DAT_RECIPE_EXAMPLE.json
```

Mod folder:

```sh
./kit mod check path/to/mod
./kit mod check path/to/data-mod --status path/to/mod-status.json
```

Swatch geometry:

```sh
./kit swatch patterns list
./kit swatch patterns validate
./kit swatch patterns json radar_sweep
```

## Growth Rule

Add shared package capability first, then expose it through one of the domain
CLIs. Avoid one-off binaries and avoid project-specific assumptions in shared
packages.

Examples:

- Replay parsing belongs in `pkg/replay`, then `cmd/decima-rec` or `cmd/rec`.
- Scenario trigger decoding belongs in `pkg/scenario` / `pkg/trigger`, then `cmd/scen`.
- Data-mod string and `.dat` rules belong in `pkg/datfile` / `pkg/modpack`, then `cmd/dat` and `cmd/mod`.
- Scenario Sandbox geometry belongs in `pkg/geom`, then `cmd/swatch`.

Existing project-specific tools, such as `SDS/cmd/decima-rec`, are source
material for extraction. They should not be moved into AoE2Kit until their
reusable format knowledge is separated from project-specific behavior.

## DAT Engine

AoE2Kit will not port `genieutils` as an eager object graph. The Go `.dat`
engine is a current-AoE2DE span/index patch engine: preserve unknown bytes,
patch numeric fields in place, splice owned records for variable-length fields,
and verify by re-indexing the decompressed payload to EOF. It is not trying to
support old Genie `.dat` versions that are not in active AoE2DE circulation.
See `docs/DAT_ENGINE.md`.

Scenario writing follows the typed-islands model too, but trigger and unit
islands are rebuilt as complete recursive sections rather than patched field by
field. See `docs/SCENARIO_WRITER.md`.

Resource guardrails:

- `kit` sets a default soft Go heap limit of 1 GiB unless `GOMEMLIMIT` or
  `AOE2KIT_GOMEMLIMIT_MB` is set.
- Scenario parsing refuses inflated bodies above 32 MiB by default. This avoids
  accidental full-tree materialization of huge community scenarios. For an
  intentional large teardown, run one command at a time and set
  `AOE2KIT_MAX_SCENARIO_MB` higher or `AOE2KIT_ALLOW_HUGE_SCENARIO=1`.

Current scenario support:

- Current AoE2DE `.aoe2scenario` read support for DE `1.55` through `1.58`.
  Kit uses the embedded APar `1.58` structure spec as the main field-order
  source and applies known older-version layout differences in memory:
  `1.55` DataHeader player data plus pre-`1.58` trigger-effect field removals.
  Scenario writing supports DE `1.57` and `1.58`; older readable versions fail
  fast as read-only inputs.
- Uncompressed file-header detection, raw DEFLATE inflate/recompress, and
  decompressed-payload roundtrip validation.
- Sequential body parsing through `DataHeader`, `Messages`, `Cinematics`,
  `BackgroundImage`, `PlayerDataTwo`, `GlobalVictory`, `Diplomacy`, `Options`,
  `Map`, `Units`, `Triggers`, and `Files`.
- Byte-identical decompressed body rebuild for read-only verification.
- Trigger summary extraction: trigger version, trigger count, display order,
  variable count, trigger names, enabled/looping flags, effect counts,
  condition counts, raw record spans, and Python-compatible `graph_sha256`.
- Trigger invariant checks for top-level trigger display order, effect and
  condition counts, optional per-trigger effect/condition display-order arrays,
  and selected-object tails.

Current replay player-profile support:

- `kit replay player-events` exposes the canonical replay-visible event table
  with filters for player, type, action id, phase, time range, and row limit.
  Action id `0` is a valid filter and is preserved explicitly in JSON output.
- `kit replay story` includes factual chapters (`pre_game`, `opening`,
  `midgame`, `endgame`) and important moments such as first action, first chat,
  first flare, first resign, and peak action window.
- `kit replay inbox` is the folder-level playtest report: scenario groups,
  duplicate replay detection, parsed/error counts, data-set/result summaries,
  profile coverage, top issue cards, and per-replay story moments.
- `kit replay issues` ranks replay-visible feedback into issue cards grouped by
  chapter, player/region, and normalized text, with evidence rows behind each
  card.
- `kit replay playtest <replay> --text` is the author-facing single-replay
  playtest report. It combines session identity, scenario build fingerprint,
  live feedback moments, duplicate-feedback collapse, and nearby checksum state
  windows. The default window is two checksum samples before/after each
  feedback moment and defaults to human players only; add `--include-computers`
  to include AI/computer rows. Its summary split is the single checksum interval
  immediately before the feedback timestamp, not a union across the whole
  window, because the split is meant to answer "who changed right before this
  report?" rather than "did anything happen during the surrounding two minutes?"
  Text output carries the honesty boundary: replay state can show correlation
  near feedback, but it cannot name the trigger, screen rendering, or author
  intent.
- Per-player replay action counts, decoded-vs-raw coverage, average APM, APM
  windows, dead gaps, top command vocabulary, coordinate-command footprint, grid
  cells, optional named-region summaries, and chat/taunt/flare/telemetry counts.
- JSON-first event-table drilldown via `--include-events`.
- Chat phase tags split `pre_game` vs `in_game` with an explicit heuristic
  claim: chat at or before 1.000s is treated as pre-game.
- `kit replay chat` reports player/human replay chat. Trigger-effect
  `send_chat` messages are simulation-generated and are not retained in
  recordings. DE can also prepend prior session chat into a new recording as an
  early same-timestamp burst; AoE2Kit tags conservative matches as
  `source=backlog` instead of treating them as live in-game commentary.
  Backlog rows remain visible in `kit replay events` and `kit replay chat`, but
  are excluded from feedback/story/profile/inbox/issue-card live-feedback
  counts.
- Global `viewlock` camera coordinates are exposed as camera observations, but
  are not attributed to any player.
- Typed action decoders include core movement/attack/build/gather/flare/resign
  actions plus high-value command vocabulary such as `DE_QUEUE` and `RESEARCH`.
- Confidence ledger labels profile facts as replay-action verified, phase
  heuristic, raw-preserved, or not-claimed. V1 does not claim skill,
  temperament, intent, combat lifecycle, or economy state.
- Blank-root scenario creation with `kit scen blank <out.aoe2scenario>`.
  By default it uses the kit's embedded DE 1.58 blank editor seed, clears the
  seed's editor toy triggers, stamps the current save timestamp, and writes a
  structurally verified output. `--players N --dummy-starters` creates active
  slots 1..N, makes P1 human and later slots computer, and places one Outpost
  dummy per active slot so DE should not inject starter Town Centers into
  diagnostic fixtures. The output is structure-verified, not engine-verified.
- Append-only scenario recipe patching for raw helper triggers. Supported
  effects include `display_instructions`, `display_timer`, `send_chat`,
  `create_object`, `kill_object`, `remove_object`, `task_object`,
  `change_ownership`, `change_object_name`, `change_object_hp`,
  `teleport_object`, `change_object_stance`, `set_player_visibility` /
  `reveal_map`, `research_technology`, `modify_attribute`, `modify_resource`,
  `script_call`, `activate_trigger`, `deactivate_trigger`, and
  `declare_victory`. Supported conditions include `timer`, `object_selected`,
  `object_in_area` / `objects_in_area`, `own_objects`, `own_fewer_objects`,
  `accumulate_attribute`, `object_visible`, and `variable_value`. Patch writes
  to a separate output file, reopens the output, verifies decompressed body
  rebuild, and checks trigger invariants.
- Narrow existing-trigger edits with `edit_trigger` and `copy_trigger`: select
  by `target_index` or `target_name`, set `enabled`, `looping`, and/or
  `set_name`, append supported effects/conditions, and remove/clear/replace
  trigger child rows while repairing local counts and display-order arrays.
- Unit inspection with `scen units` and unit recipes with `add_unit`,
  `edit_unit`, `copy_units_in_area`, `move_units_in_area`,
  `edit_units_in_area`, `remove_unit`, and guarded `remove_units_in_area`,
  maintaining the affected player section `unit_count` fields. Area copy/move
  accepts either explicit offsets or a destination top-left `target_x,target_y`.
- Map inspection with `scen map` and terrain recipes with `set_terrain_rect`,
  `set_terrain_circle`, `set_terrain_line`, `set_terrain_border`, and
  `copy_terrain_area`, changing or copying terrain tiles, elevation, and layer.
- Position-aware terrain inspection with `scen terrain`, including per-terrain
  counts, bounds, centroids, connected components, and optional per-tile rows.
- Scenario setting recipes for controlled diagnostics: embedded XS script
  attachment (`xs.name`, inline `xs.content`, or `xs.content_file`), player slot
  active/human/AI-type edits, GlobalVictory edits (`victory.conquest_required`,
  `victory.ruins`, `victory.artifacts_required`, `victory.discovery`,
  `victory.explored_percent_of_map_required`, `victory.gold_required`,
  `victory.all_custom_conditions_required`, `victory.mode`,
  `victory.required_score_for_score_victory`,
  `victory.time_for_timed_game_in_10ths_of_a_year`), pairwise diplomacy stance
  edits, and starting food/wood/stone/gold/trade-goods resource edits.
  Single-player scenario lobbies do not expose the random-map Victory dropdown,
  so replay diagnostics that need trigger-only endings should author
  `victory.conquest_required=0` into the scenario file instead of relying on a
  run instruction.
- `kit scen smoke-recipe` prints a canonical project-neutral editor-smoke
  recipe. `kit scen smoke <in> <out>` applies it directly. The smoke writes a
  visible terrain marker and unit plus disabled trigger primitives; it is still
  `structure_verified_not_engine_verified` until opened in the DE editor/game.
- Current XS source of truth: the AoE2DE UGC Guide; see `docs/XS_AUTHORING.md`.
  Scenario XS is not limited to one monolithic file: `.xs` files can include
  sibling modules with `include "file.xs";`, and trigger `script_call` effects
  invoke parameterless XS functions via the effect `message` field.
- Preferred generated XS path: use `xs.mode:"carrier"` in a scenario recipe.
  This mirrors AoE2ScenarioParser by storing full XS source in a disabled
  trigger's `script_call` message, replacing the first existing carrier and
  clearing old attachment fields by default. `carrier_trigger_index` can replace
  a known placeholder trigger in place. Front Towers provides live-author
  evidence for this carrier pattern; Kit's own carrier writer is
  structure-verified until a Kit-authored carrier fixture is engine-smoked.
- External XS caveat: `xs.mode:"attachment"` writes `Map.script_name` and
  `Files.script_file_*`. The v3 telemetry probe proved DE resolves that path
  against `resources/_common/xs/`, and v8 proved `Map.script_name` is opened
  literally. Use attachment mode only when deliberately shipping a matching
  external `.xs` module tree.
- `scen describe` composes the current scenario readers into one report:
  players, map summary, units, trigger fingerprint/counts, embedded scenario AI
  file summaries, and section index by default; `--full --json` emits the full
  parsed field tree for selected sections or all sections. Use `--no-path` (or
  `--quiet-header`) for stable textconv output where git passes temporary file
  paths.

Existing-unit edits and structural map resizing are intentionally not exposed
yet.

For editor/game validation of the write path, see
`docs/AOE2KIT_WRITE_SMOKE.md`.

Current replay and fingerprint support:

- `.aoe2record` and replay `.zip` compressed-header inflation with
  game/save/log version fields.
- Best-available replay identity with explicit tier labels:
  `trigger_graph` first, then `fallback:terrain`.
- `trigger_graph` location uses the validated dense-default/string-cluster
  heuristic, followed by explicit `trigger_count` graph parsing. If that primary
  path fails, AoE2Kit falls back to a labeled binary-structure scan rather than
  silently returning no graph.
- Replay trigger graph rows preserve the original identity fields for
  `graph_sha256`, but `kit replay graph` also enriches effect and condition
  children with `type_name` and `known_fields` where the DE field positions are
  known. This makes replay-embedded triggers easier to inspect without changing
  the canonical map fingerprint.
- `kit replay triggers` searches that embedded trigger graph by trigger text,
  effect/condition id or name, player, variable, and unit constant. It is a
  structural authoring readback, not proof that those triggers fired during the
  recorded game.
- `kit replay trigger-neighborhood` follows embedded trigger-control edges
  around one trigger, using activate/deactivate effect targets and decoded
  trigger-id condition targets. It answers structural "what is connected to
  this?" questions from a replay without needing the source scenario file.
- `fallback:terrain` parses current-DE map info and hashes map dimensions plus
  terrain/elevation grid. It is stable and useful for recognition, but it does
  not prove trigger integrity.
- `kit identify` registry lookup for scenario and replay fingerprints.
- `known_scenarios.json` is growable with `kit identify --register`.
- `kit collect <folder>` recursively fingerprints `.aoe2scenario`,
  `.aoe2record`, and replay `.zip` files, reports known/new keys, and registers
  new placeholders unless `--dry-run` is supplied.
- `kit player stats <profileId>` wraps Microsoft's public
  `GameStats/AgeII/GetFullStats` endpoint with the required
  `Origin`/`Referer` headers and a 120-second default JSON cache under the
  platform user cache directory or `AOE2KIT_CACHE_DIR`. The report is labeled
  `api_reported_career_aggregate_not_per_match_truth`: elo/standing/lifetime
  aggregates are API facts, not per-match replay facts.
- Replay AI identity is reported from the structural v68 DE header `ai_files`
  list when present: grouped AI name, file names, source mod, and a stable
  `ai_files` signature. Empty `ai_files` is reported explicitly as
  default/no-mod AI. The replay does not contain custom AI script content from
  modern DE mods; script content must still be fingerprinted from `.ai`/`.per`
  files.
- `kit replay ai-manifest <file.aoe2record>` extracts default DE AI module
  references from bounded v68 header regions. It labels exact
  `(load "Promisory/...")` and `(load "Promisory\\...")` directives as decoded
  text islands inside an otherwise opaque AI/default-loadout wrapper, reports
  `PromiDE` mentions, and resolves module basenames against a local Promisory
  tree when available. The coverage ledger also decodes adjacent ASCII AI-script
  fragments such as CRLF separators and `#load-if-not-defined` guard lines, plus
  XS module include references such as `ailib/Geometry.xs`, while leaving
  binary padding opaque. This is a manifest/reference decoder, not proof that
  the replay embeds the full PromiDE logic or `.per2` wrapper format.
- Replay active data-set identity is reported from the embedded DE mod block.
  `status=vanilla` means no active data mod was named in the replay header;
  `status=modded` lists the selected data-set name(s); `status=unknown` means
  the field was not parsed. This proves load identity, not render behavior: a
  modded replay can still be structurally loaded while an effect renders wrong.
- `known_ais.json` is a parallel AI registry keyed by the replay-side
  `ai_files` signature. `kit identify --register-ai` can name an unknown
  replay AI signature.
- `kit ai fingerprint` hashes actual `.ai`/`.per` files or folders with line
  endings normalized. This is the content-integrity side of AI verification and
  must be supplied from mod files, not from the replay.
- `kit replay events` is the shared replay witness core. By default it emits
  story-relevant parsed events: chat, taunts, flares, resigns, telemetry
  markers, and a raw-preserved unknown-operation stop if parsing reaches an
  unsupported replay operation. It also reports parsed action-stream duration,
  active data-set identity, save-chapter skips, postgame sightings,
  spatial-action counts, and result inference when resign actions identify
  losers. `--all` includes system
  sync/viewlock rows and decoded low-level action rows. `--raw` includes raw
  payload hex for action work.
- `kit replay info` exposes `scenario_name`/`scenario_metadata` when the name is
  available from the structured DE lobby tail or a bounded early-header scenario
  path string. It also includes `lobby_settings`, a structured emission of the
  DE settings prefix used by the player parser: game type, difficulty,
  population, victory, resources, ages, speed, treaty, diplomacy/ranked flags,
  map dimension/RMS id, and labeled raw adjacent bytes where enum semantics are
  not yet fully proven. It includes `record_sha256` for the normalized
  `.aoe2record` payload and `header_metadata.tail_guid_hex`, the exact 16 raw
  bytes before the DE lobby-tail name. The metadata includes source/confidence;
  it does not infer a scenario title from trigger names or canonicalize the raw
  tail GUID into stronger semantics.
- `kit replay summary <file.aoe2record|zip>` is the mgz-style one-object rollup:
  record hash, raw tail GUID metadata, header versions, scenario, map hash,
  active data-set identity, lobby settings, duration/result, and players joined
  with civ name, color, team, profile id, winner/resign, and rating when a DE
  postgame leaderboard block is present. Add `--tiles` to include the full map
  terrain/elevation grid; default output keeps only dimensions, tile count, and
  terrain hash.
- `kit replay coverage` is the full-parse ledger. It reports raw-file,
  inflated-header, and body byte spans as either `decoded` or
  `bounded_opaque`, plus body operation counts and samples. Header/body decoded
  plus opaque bytes must add back to the source region sizes; this is the guard
  against invisible parser gaps. For DE save_version 68 records it walks the
  map-info and initial-player spine, including per-player headers, the v68
  player stats/view/start-metadata island, and the bounded object/save-state
  tail after those attributes. The large non-Gaia pre-object tail is now
  recognized as each player's complete effective game-data table: P1/reference
  is emitted once as `effective_gamedata_template_reference`, other players are
  partitioned into byte-identical `effective_gamedata_template_common` runs and
  measured `effective_gamedata_u16_diff_run_measured_vs_reference_semantics_partial`
  runs. The known 864-entry/u16 effective
  unit-slot availability array inside that table is carved out as
  `effective_gamedata_unit_availability_array_mapped_to_dat_unit_slots`, matching
  the `kit replay effective-data --dat` cross-reference. Whole bounded-opaque
  spans that are byte-identical to a subrange of an already framed
  effective-game-data template/common region are promoted to decoded
  `duplicate_of_effective_gamedata_template_subrange` regions. This does not
  claim new field semantics, but it removes duplicated table bytes from the
  remaining dark frontier. The measured diff runs likewise account for bytes as
  aligned u16 deltas against the reference table while keeping field/index
  semantics partial. Object tails after the effective game-data table are
  segmented into the v68 rotating marker,
  replay-supported decoded object prefixes, decoded v68 tail islands, and
  bounded opaque gaps between/after those prefixes. Full object records are not
  field-decoded yet.
  `--fail-on-opaque` turns the ledger into a hard gate for the full-transparency
  target.
- `kit replay opaque-spans` is the unknown-byte workbench. It lists bounded
  opaque spans with byte profiles: entropy, distinct byte count, zero/0xff
  density, printable density, neighboring decoded regions, hex samples, and
  simple shape hints such as map-tile or trigger-count multiples. This is
  structural evidence only; it frames the frontier without claiming semantic
  decode.
- `kit replay frontier` is the summary map for that unknown-byte frontier. It
  buckets remaining opaque spans into hypotheses such as v68 pre-trigger tail,
  referenced object-body remainder, player object-tail bands, and AI/XS
  script-container gaps. It also reports exact duplicate
  opaque payloads across buckets, which is the current best way to find one
  repeated structure that can be decoded once and claimed in multiple places.
  Exact effective-template duplicates are already promoted by `kit replay
  coverage`; remaining `template_hits` should normally be zero unless the
  coverage promotion threshold/options change. Bucket rows still include
  `template_hit_bytes` and `residual_bytes`; after promotion, the bucket bytes
  are the residual target list for the next semantic parser.
- `kit replay corpus <folder>` scans `.aoe2record` files and replay `.zip`
  files recursively by default and composes the generic replay facts into a
  corpus report: scenario vs non-scenario counts, trigger/fallback identities,
  player counts/names, data-set identity status, duration/actions/feedback,
  body-op and action-id histograms, coverage totals/min/averages, postgame tail
  shapes, and unknown action ids grouped across the folder. Its verification
  label is `structure_verified_not_engine_verified`.
- `kit replay opaque-clusters <folder>` runs the opaque-span profiler across a
  replay corpus and groups still-opaque spans by replay space, normalized span
  name, normalized neighboring decoded regions, byte-size bucket, and shape
  hints. It reports count, file coverage, total bytes, size range, entropy,
  zero/0xff/printable density, and sample hex. Its verification label is
  `structure_verified_opaque_bytes_profiled_not_semantically_decoded`.
- `kit replay effective-data <replay> --dat <empires2_x2_p1.dat>` reads the
  effective game-data table from a replay. It reports the reference template
  SHA-256 for data-mod/template comparison, checks it against the built-in
  observed DE v68 vanilla template baseline, and decodes the measured
  864-entry/u16 availability array at tail offset `879618`. With `--dat`, each
  array index is cross-referenced to the same DAT unit slot, unit id, unit name,
  and DAT enabled flag. `--tree` prints the civ's effective unit-slot
  availability as a readable tech-tree-style list. The verification label is
  `structure_verified_array_index_to_dat_unit_slot_crossref_template_hash_build_sensitive`:
  the u16 array is replay-structure verified, unit-slot mapping is the current
  DAT cross-reference, and the vanilla detector is a build-sensitive template
  hash comparison. DAT tech/research records are exposed by `kit dat techs`, but
  replay-side effective technology/research arrays are not mapped to those DAT
  tech ids yet.
- `kit replay effective-units <replay> --dat <empires2_x2_p1.dat>` decodes the
  first field-level unit-stat island from those effective game-data tails: hit
  points via a unit-debug-name anchor and a measured `name_offset-120` i16
  field. Duplicate debug-name anchors are skipped by default; pass
  `--include-duplicates` to inspect them as lower-confidence probes. It compares
  the replay's civ-effective HP to the supplied base DAT and reports
  per-player/per-civ unit rows. Changed HP rows include `civ_span` and a
  per-row classification: `mod_confirmed_multi_civ` when the exact same unit
  name and HP change appears across multiple civs, and
  `unresolved_single_civ` when the row could still be a vanilla civ bonus,
  vanilla technology, or civ-specific scenario/data-mod edit. Current coverage
  is deliberately narrow: attack, armor, resource cost, train time, movement
  speed, and CBA spawn wave count/seconds are listed as not decoded in this
  effective-unit slice.
- `kit replay datamod-check <replay>` wraps the effective game-data decode into a
  direct verdict. Without `--baseline-replay`, it can only compare the replay's
  reference effective-data tail to the built-in observed v68 anchor hash or a
  caller-supplied `--known-template`; this is useful but not a universal proof,
  because effective tables are build/civ/scenario sensitive. With
  `--baseline-replay`, it compares effective-data tails player-by-player and
  diffs the measured unit-slot availability array, optionally mapping changed
  slots to DAT unit ids/names with `--dat`. The verification label is
  `structure_verified_effective_gamedata_tail_diff_not_engine_verified`.
- `kit replay fetch <gameId> --profile <profileId>` wraps Microsoft's public
  `GameStats/AgeII/GetMatchReplay`, storing replays as
  `AgeIIDE_Replay_<gameId>_p<profileId>.zip` so same-game POVs do not overwrite
  one another. Existing files are cache hits unless `--force` is passed. The
  `--all-povs` mode can use `--from-roster <replay>` or a seed `--profile`; it
  only fetches additional POVs when AoE2Kit can parse positive profile IDs from
  the replay roster, otherwise it reports a warning instead of guessing.
- `kit replay diff-state` compares two inflated replay headers and replay
  bodies, then attaches each changed byte run to the current coverage regions.
  It also compares the final DE checksum/sync matrix and reports per-player
  deltas for decoded/provisional words such as object count, unit-type sum,
  object-id sum, position sum, and score-adjacent word 9. Word deltas use
  signed 32-bit wrap math so wrapped checksum fields stay readable. `--focus
  opaque` filters to changes touching still-opaque spans. This is the
  controlled-experiment bridge for suspected runtime state such as trigger
  counters, object deltas, visibility/UI state, or postgame state: it can show
  exactly where bytes changed before we know what those bytes mean.
- `kit replay scan-value` searches inflated replay headers and replay bodies
  for exact byte patterns. `--value` expands a number into little-endian int8,
  int16, int32, and int64 forms where representable; `--width` narrows numeric
  searches to one or more widths; `--hex` searches exact raw bytes; `--string`
  searches ASCII and UTF-16LE text. Hits are labeled with the current coverage
  region, so controlled fixtures with weird known values can quickly expose
  whether a value lands in the static trigger graph, runtime state, replay body
  commands, or an opaque save-state island.
- `kit replay header-anchors` decodes high-confidence header probes that have
  graduated past raw scan hits. Current surfaces are object-caption records
  shaped as `ff/0/ff/u32-length/ascii`, and the v4 engine-verified player
  resource sentinel block shaped as seven u32 fields: unknown0, unknown1,
  player_index, gold, wood, food, and stone.
- `kit replay objects` builds the first object-ID index for DE save_version 68
  replays. It scans the framed initial-player payloads for conservative object
  record prefixes, joins those candidates to command-stream object references,
  and reports owner, object id, unit id, replay-native record class, x/y, and
  target/selection reference counts. It also reports byte accounting for each
  candidate: `prefix` is the 80-byte decoded fixed static-base island shared by
  all existing object record classes (type/owner/unit/object id, HP/state/facet,
  x/y/z, resource/amount/damage/group/object-props marker), while `body` is the
  bounded byte span up to the next candidate prefix or player-span end. When the
  object-props branch is absent, the parser also decodes the immediately
  following sprite-list flag/list if it terminates inside that bounded body, and
  then the mgz-defined DE extension prefix: particle declarations, versioned
  fixed bytes, and optional DE strings when their markers validate. It also
  accounts for the next fixed static-tail island and only advances into the
  animated/base-moving header through `num_path_data` when conservative
  plausibility checks pass. For matching combat/building records, it can also
  follow the counted base-moving continuation, the idle action-list marker, and
  the fixed DE combat prefix through `has_ai`, the no-AI DE combat/town-bell
  tail, and fixed building metadata through `captured_unit_count` plus an empty
  `extra_actions` sentinel when present. It then decodes the bounded
  production-queue header, endpoint/update/DE building tail, compact v68 building
  trailer, repeated v68 combat trailer, v68 action-state blocks, post-action
  state blocks, and position/order tails when those guards pass. These tail
  islands are structure-verified; their exact gameplay semantics remain partial.
  On the latest CBA anchor, referenced object-body opacity is down to three
  exact 50-byte final-span suffixes.
  This is a candidate index, not a full object parse; use the confidence fields
  and prefer replay-referenced rows.
- `kit replay object-state` promotes the decoded object islands into readable
  initial state cards. It reports object id, owner, unit/name/class, HP, object
  state, position, command/target refs, action prefix fields, town-bell/combat
  tail fields, plausible building fields, production queue capacity, v68 trailer
  ids, v68 linked-object ids, and per-object decoded-vs-opaque body bytes.
  Unset/non-finite float sentinels and implausible raw scalar values are
  suppressed at this card layer instead of exported as facts. Its verification
  label is
  `structure_verified_initial_object_state_not_runtime_delta_stream`: these are
  saved initial/object-state fields from the replay header, not a claim that the
  command stream now contains every damage/death/combat transition.
- `kit replay object-shapes` groups those bounded object candidate spans by
  record class, unit id, record byte length, prefix/body length, owner, and
  referenced-vs-unreferenced counts. This is the scouting report for the next
  object-body parser: it identifies stable length families without upgrading
  unknown variable bodies into decoded bytes. Text output includes a short
  body hex sample and 96-byte first-opaque hex sample for each shape, plus explicit
  per-record and total still-opaque body bytes after the decoded body prefix.
- `kit replay events --objects`, `kit replay actions --objects`, and
  `kit replay player-events --objects` enrich matched target/object IDs with the
  candidate index so rows can say a target was a P6 gate or a P2 castle instead
  of only showing a numeric object id. Unmatched dynamic moving-unit IDs remain
  raw IDs.
- `kit replay spawns` mines create-object trigger effects from the embedded
  trigger graph. It reports the scenario's spawn recipe catalog: unit id/name,
  target player, trigger/effect, and x/y. This is a trigger recipe catalog, not
  runtime object-id assignment for spawned units.
- `kit replay lifecycle` starts joining runtime command-stream object IDs back
  to trigger spawn recipes. Today it records every first-seen selected object,
  then promotes small-selection first sightings near trigger spawn points into
  lifecycle candidates. Villager-family candidates use unique villager spawn
  points and report the first candidate timestamp per player. This is still an
  honest candidate layer (`structure_verified_lifecycle_candidate_not_runtime_unit_type_proof`):
  large near-spawn selections are counted as overmatched and are not typed.
- `kit replay postgame` decodes the DE op=6 reverse-block tail when present.
- `kit replay sync` surfaces the op=2 heartbeat: per-sync time deltas,
  delta-shape histogram, periodic checksum payloads, and conservative
  per-player state deltas between checksum samples. DECODED SEMANTICS: the DE
  checksum trailer u32 equals current world time in ms; matrix rows are P1..P8;
  word 8 is player number; word 6 is current living object count; word 2 is the
  sum of current living object unit-type ids; word 10 is the sum of current
  living object instance ids. The v2 diagnostic fixture proves single-object
  deltas exactly: +83/+448 unit-type sums and +4..+9 instance-id sums match
  spawned Villagers/Scouts. Word 7 is decoded as the current object position
  sum `(x+y)*100`. Word 1 is now decoded as resource stockpile after the v3 XS
  probe wrote food values `900000`, `900100`, and `900200` and op=2 word 1
  carried them exactly. Word 3 tracks the sum of each checksum-counted object's
  engine `state` field: normal living objects mostly contribute `2`, flares
  contribute `3`, and observed corpses/rubble contribute `5`. Word 4 tracks the
  sum of each object's engine `carry` field. `carry` is generic: villagers use
  it for carried resources, monks for faith, flares/markers for lifetime, and
  corpses/rubble for decay timer. Word 9 remains provisional: visible-score,
  object-HP, viewport,
  mouse, time, and exploration-only mappings are disproven; the current best
  hypothesis is a per-player state digest/check word. A frozen multiplayer
  fixture shows a five-value successor ring with forward skips, but raw
  sample-index modulo is not a stable phase oracle.
  `--raw-words` emits one row per checksum sample and player slot with all 11
  matrix words, intended for attribution research and sentinel probes; it is
  opt-in so normal reports stay compact. Use `kit replay deaths`,
  `kit replay sync`, `kit replay sync-log`, and game-specific consumers such as
  `kit cba` to inspect the current kill-attribution boundary.
- `kit replay sync-log <p0-sync.txt>` parses AoE2DE's desync-side
  `*-p0-sync.txt` engine trace. This is not replay-byte parsing: it is a
  structured read of the text log the engine writes when a game goes out of
  sync. The report extracts turn/world-time headers, `Player N(...)` object
  dumps, truncated player-attribute dumps, per-player checksum-word candidates
  (resource-stockpile sum from printed attributes 0..3, object count, unit-type
  sum, instance-id sum, position-sum candidates, retarget/action/state sums),
  object census rows, and the named RNG streams (`Action`, `World`, `AI`,
  `Object`, `Player`, `Editor`, `Comm`, `UI`). Current desync evidence shows the
  engine declares a 601-wide attribute array but prints only indices 0..3 in the
  sampled artifact; attr20 kills and attr154 deaths are not available from that
  log. With `--replay`, it checks whether the log's world-time window overlaps
  replay checksum samples and compares exact world-time matches when they exist;
  no-overlap is reported explicitly.
- `kit replay checksum-phase` summarizes one checksum word's per-player value
  alphabet and successor relation, with `--all-words` to scan the whole 0..10
  matrix. It is intended for dark-word probes like word 9, where a
  rotating/check-lane pattern can survive even when sample-index modulo
  bucketing fails. Ring-shaped successor maps are reported explicitly. Text
  output renders negative sentinel values as unsigned/signed pairs, for example
  `4294967291/-5`.
- `kit replay checksum-probe` searches the decoded op=2 checksum state deltas
  for expected fixture pulses. Custom probes can match a player row plus exact
  `--object-count-delta`, `--unit-type-sum-delta`, and/or
  `--object-id-sum-delta`, plus semantic checksum-word deltas with
  `--state-sum-delta` for word 3, `--carry-sum-delta` for word 4, and
  `--digest-delta` for unresolved word 9. Legacy aliases
  `--word3-delta`, `--word4-delta`, and `--score-candidate-delta` remain
  accepted for older notes/scripts. The built-in
  `--preset v6-playground-helper` checks the automated v6 diagnostic bands:
  P3 hostile dummies (`-3/-526`), Gaia neutral dummy leakage into any P1-P8 row
  (`-3/-501`), all six dummies coalesced in one row (`-6/-1027`), and the weak
  P2 actor-side `+3` candidate. The built-in
  `--preset v7-checksum-calibration` checks the trigger-only v7 create/kill
  anchors for seven P1 unit/building constants, including the observed
  selected-ref `Kill Object` replacement signatures. In v7, those kill effects
  did not reduce object count; they kept `object_count_delta=0` and replaced
  high refs `970201..970207` with low runtime objects, producing a constant
  `object_id_sum_delta=-970192`.
  Verification boundary:
  `structure_verified_checksum_delta_match_no_death_or_kill_attribution`.
  A match means the checksum state sums moved as expected, not that the engine
  reported death, damage, killer, or kill credit.
- `kit replay combat` builds the first generic combat-story surface from those
  sync facts. It reports timestamped windows where a player's living object
  count has a net drop, then joins nearby targeted commands against that
  player's indexed objects as candidate pressure. It also groups targeted
  objects into contested-object summaries, so gates/castles/towers that absorb
  repeated player pressure are visible without reading the raw event table.
  Its verification label is
  `structure_verified_net_lifecycle_loss_with_candidate_pressure_not_kill_attribution`:
  it is not direct damage, death, killer, or kill-credit decode, and production
  can mask losses inside a checksum interval.
- `kit replay deaths` narrows that surface to checksum intervals where the
  living object count stays flat but the unit-type sum matches a known
  live-unit -> corpse/rubble replacement from `data/corpse_units.json`. This is
  a replacement/removal signal, not killer attribution: engine death, trigger
  kill, delete-object, and other selected-object effects can converge on the
  same checksum shape unless a fixture or sidecar proves the cause.
- `kit replay actions --unknown-only [--samples N]` aggregates untyped action ids (count, players, span,
  payload shapes, hex samples) so the next decoder is chosen from corpus evidence. Named from the mgz v68
  reference: 19 GUARD, 20 FOLLOW, 35 DE_RETREAT, 38 DE_AUTOSCOUT, 41 DE_TRANSFORM, 43 RATHA_ABILITY,
  44 DE_107_A, 53 AI_COMMAND, 140 DE_107_B, 196 DE_TRIBUTE. RATHA_ABILITY (43) is FULLY decoded
  ([count u32][mode u32][ids]) and golden-tested (10 events, all P1/Bengalis, modes 153/154 = the two
  Ratha forms; which mode is melee vs ranged is not claimed). DE_AUTOSCOUT (38)
  is engine-validated by a v2 diagnostic run and is preserved as a named
  raw action even though its payload semantics are not fully decoded. Action id
  42 remains unknown (also unknown to mgz): 7-byte payload
  [player u8][04 00][u32].
  The v3 AI-player fixture surfaced two new AI-owned raw targets: action id
  `130`, 372 samples, 12-byte payload, about 0.3s cadence from P2; and action
  id `37`, 13 samples, 8-byte payload from P2. They remain corpus targets, not
  decoded commands.
- `kit replay xs-telemetry` is a focused readback helper for the v3 XS
  diagnostic probe documented in `docs/XS_TELEMETRY_PROBE_V3.md`. It decodes
  only that probe's sentinel packing in op=2 word 1:
  `900000 + attr20*10000 + attr154*100 + attr43` for food attr 0 and
  `700000 + ...` for the unused-220 candidate carrier. The v3 engine run proved
  food/word-1 works, `xsChatData` does not serialize, unused attr 220 did not
  move word 1, and attr 154 stepped for both kill events while attr 20 stayed
  zero. Its verification label is still probe-specific: it is not a native
  replay outcome-stat decoder.
- `kit replay carrier <record> --ledger expected.json` generalizes authored
  self-reporting probes. It reads `reader.sync_carrier` and timed
  `reader.attempts` from the ledger, extracts the matching sync checksum carrier
  word, maps observed values to expected attempt outcomes, and marks attempts
  `unsampled` when no checksum sample landed inside the attempt window. This is
  structural replay readback of a probe contract, not a claim that arbitrary
  runtime state is decoded without an authored carrier.
- `kit replay sidecar-sync <record> --xsdat run.xsdat (--ledger expected.json|--schema rtv12|rtv13|rtv14|rtv141|rtv142)`
  correlates an XS sidecar fixture with the replay sync matrix. It decodes typed
  `.xsdat` rows via either an explicit ledger or a named diagnostic schema, finds
  the first nearby DE checksum sample for the selected player, and reports
  whether already-earned checksum words match the sidecar facts (`word_1`
  resource stockpile, `word_2` unit-type sum, `word_6` object count). Unnamed
  words and attribute mirrors are reported only as candidates; this is a
  calibrated diagnostic readback, not a general hidden-state decoder. Add
  `--fail-on-mismatch` when a fixture run should exit nonzero for sidecar decode
  failures, missing sync samples, or known-word mismatches. The default
  sync-search window is 20 seconds; use `--window-ms` only when a fixture
  deliberately holds phases longer.
- `kit replay postgame-corpus <folder>` scans recordings for postgame variants. Local corpus result
  (2026-07-21): 14 replays incl. two public aoe2insights downloads -- ALL op=6 tails are metadata-only,
  zero achievement blocks, zero action-255. The v3 trigger-victory fixture also
  ended on a real scenario victory screen and still produced metadata only.
  A true ranked RM Team 4v4 fixture,
  `save-analysis/insights_rm_team_493687875.zip`, has no trigger graph and
  still produced metadata only: a 252-byte op=6 tail with world time plus two
  leaderboard blocks. Current scoped conclusion: scenario-game and ranked-RM
  postgame display stats are recomputed for display and are not serialized as
  replay outcome facts. Campaign mode remains unsampled.
- `kit replay opaque-target` preserves the historical target analysis that led
  to `effective-data`: on latest_cba the 7 non-Gaia per-player effective
  game-data tails are ~99% byte-identical at aligned offsets, with sparse
  civ-specific u16 differences.
- `kit replay camera` (alias `viewlock`) surfaces the body op=3 viewlock stream: per-event camera x/y over
  game time plus per-stream summaries (span, bbox, distance traveled, step stats, large jumps). Honesty:
  x/y decode is structure-verified; the 4-byte tail is UNVERIFIED — when its u32 coincides with a roster
  player number the player fields are labeled inferred, never proven. Cross-replay evidence (3 fixtures,
  2026-07-21): each replay carries exactly ONE tail value equal to the RECORDING player's number — i.e. the
  stream is the POV/recorder's own camera (attention/gaze), not all players' cameras.
  On the current regression anchor it proves world-time and leaderboard metadata, and
  it explicitly reports that no per-player achievement/kills block or legacy
  action-255 achievements payload is present. Its confidence label is
  `structure_verified_scenario_and_ranked_rm_postgame_stats_not_serialized`.

Corpus acquisition note: public aoe2insights downloads resolve to Microsoft's
plain replay endpoint, `https://aoe.ms/replay/?gameId=ID&profileId=PID`, which
redirects to `api.ageofempires.com` and needs no auth. `kit replay fetch` uses
the direct Microsoft endpoint with the same `(gameId, profileId)` identity and
per-POV cache naming.
- `kit replay actions` is the low-level action workbench. It includes decoded
  MOVE, ORDER, PATROL, DE_ATTACK_MOVE, ATTACK_GROUND, BUILD, GATHER_POINT,
  AI_ORDER, DELETE, GATE, GAME, FLARE, RESIGN, and POSTGAME action surfaces
  where the current parser knows the layout, plus raw preserved events for
  still-unknown action ids. Use `--action-id N` and `--raw` to build the next
  decoder from real samples.
- `kit replay feedback` is now a filtered view over `kit replay events`, not a
  separate parser. It reports human feedback surfaces: chat, taunts, and flares,
  with AoE2DE taunt 1-105 labels carried in JSON and text output.
- `kit replay chat` is the broader transcript view. It merges replay body chat
  with header lobby chat when DE records AGP-style lobby lines, dedupes early
  lobby/body echoes, and marks likely prior-session injected chat as
  `source=backlog`.
- `kit replay story` composes replay identity, active data-set identity,
  players, parsed feedback, telemetry markers, optional named regions, event
  counts, result status, warnings, and an explicit structural-not-engine
  verification label into a designer-facing brief or machine JSON. It also
  reports player summaries, feedback clusters, a machine-readable `claims`
  ledger, and a `missing` list for capabilities AoE2Kit has not earned yet, such
  as combat/unit lifecycle and trigger-firing truth without telemetry.
- Optional replay context JSON keeps AoE2Kit project-neutral while letting a
  project bundle teach it map regions and telemetry prefixes:

  ```json
  {
    "name": "Example Scenario Context",
    "telemetry_prefixes": ["SDS EVT", "SDSDBG"],
    "regions": [
      {"name": "Hub", "box": [120, 120, 180, 180]},
      {"name": "North Trial", "box": [20, 30, 80, 90]}
    ]
  }
  ```

  Region names are enrichment from the supplied context, not replay-native
  facts. The report's `method` field is authoritative: `action_stream` can
  include timestamps and flare coordinates; `body_json_scan_after_action_stream_stop`
  means AoE2Kit preserved the action-stream stop reason but recovered chat/taunt
  JSON later in the replay body, without flare claims.
- `kit replay inbox <folder>` batch-builds replay story envelopes, hashes the
  normalized `.aoe2record` bytes for duplicate detection, and reports identity,
  active data-set identity, players, event counts, result status, warnings, and
  one brief line per replay. This is the playtest/tournament intake primitive.
- `kit replay telemetry` validates replay-visible telemetry markers against a
  neutral JSON schema. Missing required fields are errors; unknown event names
  are warnings when a schema is supplied.
- `kit replay issues` groups replay-visible feedback into draft work items by
  named region when context is supplied, otherwise by player.
- `kit replay unknowns <folder>` mines untyped action ids and unknown operation
  stops across a replay corpus, preserving counts, payload lengths, sample raw
  hex, and source replay paths for disciplined future decoder work.
- `kit verify-run <replay> --contract contract.json` is the first loop-closure
  gate. It compares a pre-registered contract against replay witnesses:
  scenario identity, active data-set identity, required/forbidden telemetry,
  required/forbidden chat markers, optional result, and render expectations.
  Render expectations are first-class but always report `unknown` until a human
  screenshot/clip oracle exists; they never pass from replay data. Overall
  `ok` requires zero failed and zero unknown claims. When a known failure mode is
  recognized, such as expected modded data set but observed vanilla, the claim
  cites the relevant gotcha and remediation.
- `kit ci init` scaffolds a project-local `aoe2kit-ci.json`, an example
  verify-run contract, and a tiny project-neutral `xs/a2k_debug.xs` include.
  `kit ci check` then runs the declared replay contract and sidecar-sync checks
  after a playtest. This is a composition runner over existing Kit witnesses;
  it is `structure_verified_not_engine_verified` and exits nonzero if any
  declared check fails or remains unknown.
- `kit campaign generate <campaign.json>` emits same-party/same-profile XS
  persistence modules using the engine-verified `.xsdat` bare-name convention.
  The reader calls `xsOpenFile` with the writer scenario stem, not `.xsdat`.
  This first slice generates writer/reader/default functions plus a manifest and
  wiring notes. It does not generate mixed-party re-declare choreography.
- `kit scen regions` emits a starter replay-context region file: whole-map
  bounds plus per-player unit-cluster boxes. These are labeled heuristic except
  for parsed map bounds.
- `kit portable-check .` gates the spore pattern: required docs, source
  buildability, manifest/verify health, local path leak detection, and
  handoff-boundary identity leak detection. It scans shipped text/source plus
  `data/*.json` for bare personal-name mentions and reports
  `file:line: phrase`.
  Local path references in non-shipped working files are reported as an
  informational count rather than a failure, so internal readbacks can keep
  exact fixture paths without making the release gate permanently red.
  Python or shell files are surfaced as warnings so durable workflow gaps are
  visible.
- `kit project inspect <folder>` is the pre-edit orientation pass. It gives a
  project-neutral role map for source, reference, generated output, evidence,
  documentation, packaged runtime files, and unknowns; flags small text files
  with local-path references; and surfaces durable script pressure. The role
  labels are path heuristics, not authorial truth.
- `kit project snapshot <folder> --out project-snapshot.json` writes that role
  map to disk as a reviewable project-state manifest for later diffs or handoff.
- `kit project diff <before-snapshot.json> <after-snapshot.json>` compares two
  project snapshots by path, role, kind, size, hash, and warnings, so a build or
  handoff can say exactly what changed.
- `kit project lineage --input PATH --output PATH --manifest lineage.json`
  writes an explicit artifact lineage report. It hashes the named inputs and
  outputs, records the tool/notes you provide, and labels the result structural
  rather than engine-verified.
- `kit recipe list` and `kit recipe show <name> --recipe-only` expose
  project-neutral scenario/DAT recipe templates using the Kit's native patch
  JSON. These are starting shapes for AI-assisted authoring, not scenario
  doctrine. The scenario templates cover common diagnostics such as XS
  attachment/carrier patterns, timer display/chat probes, timer declare-victory
  closeouts, no-conquest victory setup, marker units, terrain borders, and
  basic object spawns. The DAT templates cover common designer intents such as resource
  effects, unit attribute buffs, tech cost/time edits, ability creation,
  semantic disable/delete patterns, unit availability, production-button train
  rows, sound-item removal, sound mutes, player colours, graphics, and
  all-civ unit clones. `kit recipe export <out-dir> --domain dat` materializes
  those starter JSON files directly for editing.
- `kit docs lint [folder]` scans markdown command examples against the live
  command catalog. It catches stale command names and command-shape drift, but
  does not prove prose semantics.
- `kit scen effects` emits compact per-type summaries by default and keeps the
  large raw integer field dump behind `--raw`. For very large community
  scenarios, use `--census --max-mb N` to run a count-only forward parse through
  the trigger section. This reports effect-type counts without trigger samples
  or full scenario post-processing. Use `--where EFFECT` to stream matching
  effect sites with trigger names, indexes, display text, object ids, attributes,
  variables, XS calls, and other decoded summary fields without building the
  whole scenario tree. Add `--text` for a compact designer-facing listing.
- `kit scen triggers --grep TEXT` searches trigger names,
  objective/description/message text, effect summaries, and condition
  summaries. Add `--effect ID_OR_NAME`, `--condition ID_OR_NAME`, `--message
  TEXT`, `--unit-ref ID`, `--unit-type ID`, `--player N`, `--variable N`, or
  `--area x1,y1,x2,y2` to narrow the search, and `--text` for a compact
  listing. With no search filter, `kit scen triggers FILE` prints the full
  trigger summary, including per-trigger `effect_data` and `condition_data`
  with decoded known fields.
- `kit scen trigger-neighborhood` starts from a trigger, unit reference, or
  variable and expands nearby activate/deactivate and trigger-reference edges.
  Use this for "what else is attached to this mechanic?" investigations.
- `kit scen audit-player-coverage` groups player-named trigger families such
  as `P1 Heal Refresh` / `P2 Heal Refresh` and flags missing active-player
  coverage. It is a heuristic asymmetry finder for slot-specific bugs, not a
  claim that every mechanic should be symmetrical.
- `kit scen mechanic --kind garrison-token|transform-toggle|teleport-transition|refresh-cycle`
  classifies common scenario machinery by structural trigger evidence.
- `kit scen trigger-flow` produces a factual staged view of matching triggers:
  initial grant, use detection, cooldown/state, cleanup, refresh/regrant,
  transition, and other.
- `kit replay diff-triggers` and `kit scen diff-triggers` compare canonical
  trigger graphs: trigger/effect/condition counts, trigger adds/removes, and
  per-trigger shape changes. They are structure-verified, not engine-runtime
  verified.
- `kit scen glossary` builds a compact, streaming vocabulary of a scenario:
  effect families, trigger-name prefixes, variable ids, XS calls, user-facing
  text samples, unit ids, tech ids, and attribute ids. It is meant as the first
  command an AI runs before editing an unfamiliar large scenario, with `--text`
  for a fast human read.
- `kit scen idioms` names detected scenario techniques without claiming author
  intent: colored text, variable HUD text, create/kill display loops, caption
  label objects, XS script-call bridges, trigger relay graphs, timer
  choreography, and economy signals. Findings cite structural or heuristic
  confidence.
- `kit scen units --named` adds `unit_name` beside `unit_const` using the
  bundled unit-name table.
- `kit scen lint` reports structural and engine-safety scenario risks from the
  parsed scenario model: trigger invariants, duplicate trigger names, empty
  enabled triggers, looping display-instruction effects that can spam text every
  tick after a timer condition elapses, map tile-count mismatches, unit
  count/reference issues, player-slot warnings, embedded AI file issues, and
  corrupt user-facing strings with invalid UTF-8, control bytes, or embedded
  NUL bytes before a length-prefixed string's declared end.
- `kit scen deploycheck <scenario> <deploy-tree>` statically preflights
  scenario XS runtime deployment before an engine run. It checks that
  `Map.script_name` and `Files.script_file_path` match literally, resolve to an
  on-disk `.xs` file under the supplied deploy tree, that reachable
  `include "file.xs";` directives resolve, and that cross-file shared
  constants/variables are declared with `extern`.
- `kit scen xs <scenario> [--deploy-tree dir]` inventories the scenario's XS
  surface without modifying it: attached XS name/path/content bytes,
  inline runtime carriers, parser-style escrow carriers, ordinary trigger
  `script_call` effects and called function names, embedded/deployed XS function
  definitions, include resolution, declarations, and cross-file non-`extern`
  hazards. Carrier source is analyzed as source and excluded from ordinary
  runtime calls. It reports both full carrier-message hashes and extracted XS
  payload hashes. Without
  `--deploy-tree`, includes that live outside the embedded scenario payload are
  reported as informational unresolved references.
- `kit scen xs attach <in> <out> --xs file.xs` sets the executable XS scenario
  fields: `Map.script_name`, `Files.script_file_path`, and
  `Files.script_file_content`. Use `--also-carrier` when the scenario should also
  carry a disabled source-escrow trigger for parser-style inspection.
- `kit scen xs deploy <scenario> <deploy-tree>` writes the scenario-carried XS
  source into `<deploy-tree>/resources/_common/xs/<entry>.xs`, which is the
  runtime copy DE actually opens. `--check` immediately runs the XS deploycheck
  after writing. The claim remains structure-verified until the editor/game loads
  it.
- Recipes can set `xs.mode:"inline_runtime"` to make a self-contained single-file
  XS scenario: Kit clears the external scenario XS filename/content fields and
  inserts or replaces an enabled trigger-0 `script_call` carrier titled
  `XS string` with the full source payload.
- `kit scen xs embed|extract|compare` makes parser-style embedded XS carriers a
  first-class workflow. `embed` writes a source `.xs` into a disabled
  `script_call` escrow carrier trigger, optionally replacing a known trigger
  index; `extract` pulls the payload back out; `compare` checks scenario-carried
  XS against a source file and exits non-zero if normalized content differs. When
  a scenario has executable attachment XS, `extract` and `compare` use that by
  default; pass `--carrier` or `--trigger` to inspect carrier escrow instead.
- `kit scen blank <out> --size N --players N --human-slots N --dummy-starters` creates a
  clean current-DE seed from Kit's embedded editor-authored blank scenario,
  keeps it trigger-free unless later recipes add triggers, stamps a current save timestamp unless overridden,
  resizes the terrain grid when requested, and syncs the hidden Units-section
  player count to `N+1` including Gaia. A bare blank follows the editor-authored
  baseline: Gaia and P1 are active/human, inactive slots remain human metadata,
  and scenario `player_count` stays at the editor minimum of 2 unless playable
  slots are explicitly changed. Map sizes validate against the canonical DE
  editor presets from the shipped `MAPSIZE_*` string table: 80, 120, 144, 168,
  200, 220, 240, 252, 276, 300, 320, 360, 400, and 480. `--width` and
  `--height` can be used instead of square `--size`, but `kit scen blank`
  rejects non-square sizes because the DE editor preset table is square.
  A generated 480x480 Ludicrous scenario is engine-accepted, but large maps and
  dense swatches can be system-RAM heavy through total tile/placement volume.
  Do not infer a graphics/VRAM ceiling from swatch crashes without a smaller
  control: unresolved or illegal per-unit references, such as unit/graphic ids
  that do not resolve in the active data set or units placed on illegal terrain,
  can null-deref the loader during scenario load even on tiny maps.
  Dummy starters default to
  unit `598` (Outpost) so each active player has an authored non-production
  object and DE does not inject starter TC state into diagnostics.
- Gaia remains editor-style active by default, because that matches normal
  scenario authoring expectations and the current blank writer's verified
  setting surface. Do not assume Gaia-inactive authoring is supported until the
  separate Gaia active-state field is decoded and write-verified.
- `--no-conquest` sets `conquest_required=0` for diagnostics where empty or
  intentionally inert slots should not end the game by normal conquest rules.
- Controlled replay probes should normally start from a fresh `kit scen blank`
  file, add inert starters, disable conquest, then patch in explicit
  timer-gated probes such as `scen.timer-declare-victory`.
- `kit scen patch` refreshes the internal save timestamp unless the recipe
  explicitly sets `scenario.timestamp_of_last_save`. This keeps generated forks
  from inheriting old browser dates while preserving deterministic timestamps
  when a fixture asks for one.
- `kit scen diff` compares two readable current-DE scenarios by stable parsed facts:
  version/body size, trigger counts and graph hash, trigger summary deltas, map
  terrain counts, per-player unit counts, player-slot facts, and embedded AI
  content hashes.
- `kit scen write-check <before> <after>` composes the standard structural
  post-write gate: both files rebuild, the after-file lints cleanly, scenario
  diff succeeds, and the after-file XS surface inventories cleanly. It exits
  non-zero on failures and still labels the result
  `structure_verified_not_engine_verified`.
- `kit ai lint` checks AI files and folders for empty files, line endings,
  parenthesis balance outside semicolon comments, and rule/const presence in
  `.per/.per2` files. Empty `.ai` loader stubs are warnings, not errors.
- `kit ai diff` compares normalized AI content fingerprints per file and for the
  whole folder.
- `kit mod check` is an activation doctor as well as a packaging doctor. When a
  `mod-status.json` is found, or supplied with `--status`, it checks whether a
  data mod is present with `Enabled:false`, whether publish/private-public state
  is visible, and whether multiple enabled data mods may fight over the same
  `empires2_x2_p1.dat`. This catches the live AoE2DE failure mode where a local
  data mod exists in `mods/local` but the game silently runs the vanilla
  "Definitive Set" because the data set was never activated/published.
- `kit release-check` composes scenario identity, `scen lint`, optional
  `scen diff`, optional `mod check`, and optional `pack` into one daily gate. It
  is intentionally still `structure_verified_not_engine_verified`; a live game
  or editor check remains the only engine oracle.

Current CBA-specific support:

- `cmd/cba` and `pkg/cba` are the home for Castle Blood Automatic assumptions.
  They consume generic replay facts from `pkg/replay`; they do not define
  generic replay truth.
- `cba replay progression` derives a CBA progression proxy from `DE_QUEUE` and
  `RESEARCH` actions. It can identify first production, high-tier or Imperial
  proxy production, and early/Feudal unit production as a newbie-tell candidate.
  It does not claim direct age state unless a direct age source is decoded.
- `cba replay razes` correlates enemy building/gate target pressure with the
  later villager lifecycle payoff used by CBA-style maps. It reports first
  pressure order, first payoff order, set-play candidates where multiple
  players pressure the same target before the finisher gets a villager, and CBA
  civ-archetype fit labels such as Ethiopians opening with early raze pressure.
  Its confidence label is `behavior_inferred_raze_candidate_not_engine_raze_event`,
  because no direct raze packet has been decoded yet.

The seeded registry hashes are Python-canonical values from the mgz trigger-graph
validation. Direct registry hits prove parity for that file; misses still return
the best available `fingerprint` and `fingerprint_tier` for registration.

Registry philosophy: portability strips local paths, machine assumptions, and
corpus locations, but keeps domain wisdom. The default `known_scenarios.json`
and `known_ais.json` are the core/community wisdom bank and ship with the kit.
Project bundles may ship their own registries for project wisdom such as Decima
canonical hashes and expected AI signatures. Registry `description` fields are
load-bearing: they turn a hash match into orientation text for a fresh AI.

Full deterministic v68 trigger-section layout walking is still open. Until that
lands, non-CBA/non-ladder records can legitimately identify at `fallback:terrain`
rather than `trigger_graph`.

Current DAT support:

- Raw DEFLATE inflate/recompress with decompressed-payload roundtrip validation.
- `kit dat roundtrip <empires*.dat>` is the AGE-parity codec safety gate. It
  decodes the current-DE DAT into ordered sections, reports the currently typed
  target coverage for Graphics, Effects, UnitHeaders, Techs, Civs, Units,
  TerrainRestrictions, Terrains, Sounds, PlayerColours, and RandomMaps, re-emits
  the inflated payload from the codec document, and reports byte identity plus
  first-diff offset if it ever drifts. It compares decompressed payload bytes
  because compressed DEFLATE bytes are not a stable encoder contract. Default
  output is compact; add `--full` for the complete typed record span inventory.
- `kit dat diff <base.dat> <mod.dat>` compares decoded current-DE DAT sections
  by stable table keys after stripping byte-span bookkeeping. It reports
  added/removed/changed records for graphics, effects, unit headers, techs,
  civs, terrain restrictions, terrains, sounds, player colours, and random-map
  count records by default. Per-civ units are intentionally opt-in with
  `--section units` or `--section all` because full unit-table comparison is
  large. This is a decoded-structure diff, not a claim that every raw byte in
  still-opaque codec gaps has been semantically named.
- `kit gfx info <file.sld>` parses AoE2DE SLD sprite containers and reports
  frame canvas/hotspot/layer metadata.
- `kit dat sprite <file.sld> [--out dir] [--limit N]` and `kit gfx export`
  export the SLD main graphics layer to PNG frames plus `manifest.json` and
  `contact_sheet.png`. This first slice intentionally does not compose shadow,
  damage, or player-color layers into an in-engine final render; it is a
  structure-verified visual teardown surface for custom graphics. The SLD
  layout follows openage's public `doc/media/sld-files.md` notes.
- `kit dat terrain-restrictions`, `kit dat terrains`, `kit dat terrain`,
  `kit dat unit-headers`, `kit dat unit-header`, `kit dat sounds`,
  `kit dat sound`, `kit dat player-colours`, and
  `kit dat random-maps` expose the typed/framed DAT sections directly, so an AI
  can discover IDs and fields before authoring a codec recipe.
- `kit dat civ-patch <in.dat> <out.dat> <civ_id>` patches fixed civ fields:
  `--name`, `--tech-tree-id`, `--team-bonus-id`, `--icon-set`, and repeated
  `--resource INDEX,VALUE` entries.
- `kit dat terrain-patch <in.dat> <out.dat> <terrain_id>` patches fixed terrain
  string fields with `--name`, `--name-2`, and `--overlay-mask-name`.
- `kit dat terrain-restriction-patch <in.dat> <out.dat> <restriction_id> --terrain TERRAIN_ID,PASSABILITY`
  patches terrain passability rows inside one terrain-restriction table.
- `kit dat availability-set <in.dat> <out.dat> <unit_id> (--civ N|--all-civs) --enabled true|false|1|0`
  is the direct unit-card availability toggle. It accepts repeated or
  comma-separated `--civ` scopes and keeps the same readback contract as the
  JSON `unit_availability` recipe.
- `kit dat tech-tree-connection-create <in.dat> <out.dat> <building|unit|research> <from_index> [connection flags]`
  clones a building/unit/research tech-tree connection row and appends it to
  that family. `tech-tree-connection-patch` patches an existing row by family
  and index, and `tech-tree-connection-delete` physically removes a row from
  that connection table. Supported flags mirror the codec fields: `--id`,
  `--status`, `--buildings`, `--units`, `--techs`, `--unit-research`,
  `--mode`, `--location-in-age`, `--line-mode`, plus family-specific
  `--upper-building`, `--vertical-line`, `--required-research`,
  `--enabling-research`, `--units-techs-total`, and `--units-techs-first`.
- `kit dat graphic-create <in.dat> <out.dat> --from N [graphic scalar flags]`
  clones an existing graphic row to the tail and accepts the same scalar fields
  as `patch-graphic`: names, SLP, particle binding name, layer/player-colour
  fields, sound ids, frame/timing fields, sequence/mirroring/editor flags, and
  explicit coordinate rows.
- `kit dat graphic-patch <in.dat> <out.dat> <graphic_id> [graphic scalar flags]`
  is an alias for `patch-graphic`, provided so the CRUD command family has a
  consistent noun-verb spelling.
- `kit dat effect-create <in.dat> <out.dat> --name TEXT [--from-effect N]`
  creates a new effect row, either empty or cloned from an existing effect, and
  accepts `--command`, `--append-command`, and `--remove-command` for counted
  command-row authoring.
- `kit dat effect-patch <in.dat> <out.dat> <effect_id>` patches an existing
  effect name and command list with `--clear-commands`, `--command`,
  `--append-command`, and `--remove-command`.
- `kit dat effect-explain <empires*.dat> <effect_id> [--text|--json]`
  translates each decoded command row into authoring language, including typed
  references, packed attack/armor deltas, and warnings for mechanics whose
  structure is known but whose runtime behavior still needs engine proof.
- `kit dat effect-disable <in.dat> <out.dat> <effect_id>` is the direct
  stable-ID semantic delete for an effect: it preserves the effect row and
  clears every command.
- `kit dat effect-delete <in.dat> <out.dat> <effect_id>` runs the reference-aware
  delete planner. Unreferenced effects can be physically removed with tech
  effect-id rewrites; referenced effects are converted to the same semantic
  clear-command strategy used by `effect-disable`.
- `kit dat tech-create <in.dat> <out.dat> --from N [tech-flags]` clones one
  existing research/tech row to the tail and optionally patches the common
  scalar fields: name, prerequisite tech ids/count, civ/full-tech-mode, DLL
  ids, effect id, type, icon id, and repeatable.
- `kit dat tech-patch <in.dat> <out.dat> <tech_id> [tech-flags]` applies those
  same scalar field edits to an existing tech row.
- `kit dat tech-explain <empires*.dat> <tech_id> [--text|--json]` reports the
  tech row, linked effect row, command explanations, and the same authoring
  caveats in one view.
- `kit dat tech-delete <in.dat> <out.dat> <tech_id>` is the direct guarded
  delete alias for the current tech delete planner. Tail techs can be removed
  physically when unreferenced; referenced or unsafe cases fail with the same
  cleanup plan as `kit dat delete ... tech`.
- Resource-cost and research-location array edits are intentionally left to
  JSON `codec-patch` recipes for now; direct flags cover the scalar fields that
  are common in scenario-authoring work.
- `kit dat unit-create <in.dat> <out.dat> --from-civ N --from-unit N (--civ N|--all-civs) [unit patch flags]`
  clones a unit row into the selected civ tables. It accepts the same scalar and
  counted-row flags as `kit dat patch-unit`, so a designer can clone a unit and
  immediately tweak HP, graphics, icon, enabled state, armour/attack rows,
  costs, train locations, and tasks.
- `kit dat unit-delete <in.dat> <out.dat> <unit_id> (--civ N|--all-civs)` is
  the direct semantic unit delete wrapper. It requires an explicit civ scope,
  sets the selected unit record(s) inert/disabled through the planner, and
  includes known DAT reference cleanup where the codec has verified rewrite
  coverage.
- `kit dat sound-create <in.dat> <out.dat> --from N [sound-flags] [--item file,resource,probability,civ,icon_set]`
  clones one existing sound row to the tail, optionally patches
  `--sound-id`, `--play-delay`, `--cache-time`, and
  `--total-probability`, and can replace the cloned item table with explicit
  `--item` rows.
- `kit dat sound-patch <in.dat> <out.dat> <sound_id> [sound-flags] [--item INDEX,field=value...] [--remove-item N]`
  patches sound scalar fields, edits counted item rows by index, and removes
  counted sound-item rows with readback.
- `kit dat sound-delete <in.dat> <out.dat> <sound_id>` is the direct semantic
  delete for sound rows: it preserves the stable sound ID and sets total/item
  probabilities to `0`.
- `kit dat player-colour-create <in.dat> <out.dat> --from N [colour-flags]`
  clones one existing PlayerColour row to the tail and optionally patches
  `--colour-id`/`--color-id`, `--base`, `--outline`,
  `--selection-1`/`--selection-2`, `--minimap-1`/`--minimap-2`/`--minimap-3`,
  and `--stats`.
- `kit dat player-colour-patch <in.dat> <out.dat> <colour_id> [colour-flags]`
  patches an existing fixed-width PlayerColour row through the same field flags.
- `kit dat player-colour-delete <in.dat> <out.dat> <colour_id>` is the direct
  guarded delete alias for unreferenced tail palette rows; it refuses the same
  unsafe cases as `kit dat delete ... player_colour`.
- `kit dat palette` joins unit records to their standing graphics in one DAT
  pass and classifies the addressable art surface from `angle_count` and
  `frame_count`. `kit scen palette-usage --dat` joins placed scenario units
  back to that catalogue and reports which integer rotation/art indices are
  used or unused. Rows include a `rotation_encoding` hint: type-10 eyecandy with
  single-frame multi-angle graphics uses integer artwork indices in
  editor-authored references, ordinary combat-facing units can use radians, and
  animated units can use rotation as a phase/frame selector. These
  classifications are structure-derived hypotheses except for anchored cases
  such as IndianStatues multi-variant artwork, type-10 editor references, and
  author-confirmed animated fish; unknown mixed frame/facing surfaces are
  labeled ambiguous rather than guessed.
- `kit dat refs <in.dat> <section> <id>` is the neutral reference explorer. It
  returns the same structural reference rows used by delete planning, classified
  as `rewrite_supported`, `known_readonly`, `possible_operand`, or
  `unsupported`, with summary counts. Use `--text` for compact inspection and
  `--class`, `--confidence`, `--source-section`, and `--limit` to isolate
  specific blockers such as typed effect-command references before designing a
  mutation.
- `kit scen refs <file.aoe2scenario>` is the scenario-side reference explorer.
  It reports direct trigger-control, placed-unit, variable, and string-table
  reference fields, with `--kind` and `--id` filters for delete/edit planning.
  It is intentionally narrower than a full semantic solver: it reports fields
  the parser can name today and leaves broader multi-operation planning as a
  roadmap item.
- `kit scen delete-plan <file.aoe2scenario> <unit|trigger|variable|string>
  <id>` answers what scenario-side delete safely means. `system-prefix <prefix>`
  plans an ordered generated-bundle delete across trigger names, variable names,
  string-table text, and inline unit captions that share the prefix. It permits
  internal references inside the matched bundle, but blocks if any outside
  trigger or unit still points into it. `unit-caption
  <caption> [--player N]` resolves a unique placed-unit caption to a concrete
  reference id for blocker analysis, then emits a caption-selector delete recipe
  so AI-authored recipes can stay readable. `unit-caption-prefix <prefix>
  [--player N]` and `unit-caption-contains <text> [--player N]` resolve every
  placed unit whose inline caption starts with or contains the marker, report the
  resolved reference ids, and emit physical removal by reference id only when
  none of the matched units are referenced. `unit-type
  <unit_const> [--player N]` resolves every placed unit with that unit id,
  optionally scoped to one player, and applies the same reference-guarded batch
  delete/cleanup flow. `units-player <player>` plans a full placed-unit clear
  for one player slot, blocking if any of that player's placed units are still
  referenced by scenario logic. `trigger-name <name>` resolves a
  unique trigger name to an index for blocker analysis, then emits a name-selector
  remove/tombstone recipe when safe. Duplicate trigger names are rejected for
  delete planning. `trigger-prefix <prefix>` resolves every trigger whose name
  starts with that prefix and plans a batch delete for generated trigger blocks:
  physical removal is allowed when no trigger outside the matched batch points
  into it, and internal trigger-control references among matched triggers are
  allowed because `remove_triggers` deletes the batch atomically and rewrites
  survivor indexes. If any external direct reference exists, the plan offers
  batch tombstones plus a cleanup recipe. `trigger-contains <text>` is the
  literal grep-style counterpart for generated blocks that share a marker in
  trigger names or raw effect messages but were not consistently prefixed; it
  uses the same batch delete/tombstone/reference guards as `trigger-prefix`.
  Trigger deletes that touch effect type 55 `script_call` remain structurally
  allowed when references are clear, but `delete-plan` emits a semantic warning
  because AoE2Kit cannot prove external XS/module expectations still make sense
  after that behavior is removed.
  `variable-name <name>` resolves a unique
  trigger variable name to its stable variable id before planning removal or
  cleanup. `variable-prefix <prefix>` and `variable-contains <text>` resolve
  every variable whose name starts with or contains the marker and plan a batch
  generated-state delete by stable variable id. Referenced matched variables remain blocked until their direct trigger
  rows are disconnected. `string-text <exact_text>` resolves unique exact
  string-table text to a stable string id before planning
  clear/tombstone/cleanup; duplicate text is rejected because string slots are
  fixed IDs, not semantic labels. `string-prefix <prefix>` and
  `string-contains <text>` resolve every string-table entry whose text starts
  with or contains the marker and plan a fixed-slot batch
  clear/tombstone/cleanup without compacting or renumbering string IDs.
  `units-area <x1,y1,x2,y2>
  [--player N] [--unit N]` plans a guarded area delete using the
  same selector as the writer's `remove_units_in_area` recipe. `effect
  <trigger_index>:<child_index>` and `condition <trigger_index>:<child_index>`
  plan trigger-local child-row deletes. `effect-type <type_id_or_name>` and
  `condition-type <type_id_or_name>` plan bulk child-row deletes by decoded row
  type, optionally scoped with `--trigger N` or `--trigger-prefix PREFIX`; this
  is the preferred cleanup tool for generated systems such as "remove all
  display instructions from this generated trigger block." `effect-text-prefix
  <prefix>` and `effect-text-contains <text>` remove effect rows whose raw
  `message` field starts with or contains a marker, with the same optional
  trigger scopes; use them for narrower cleanup of generated
  display/chat/script-call probes without deleting every effect of that type. For unreferenced
  placed units, triggers, variables, and area-selected placed units it emits a
  physical-removal recipe fragment; for referenced targets it reports blockers.
  Referenced trigger, placed-unit, variable, and string plans also include
  `cleanup_command` and `cleanup_recipe` when `kit scen disconnect` can remove
  direct trigger-control/object/variable/string rows, clear direct garrison
  links, or clear direct string-id fields before a follow-up physical delete.
  String-table ids are cleared in place, not compacted, because compacting them
  would renumber later ids.
- `kit scen delete <in.aoe2scenario> <out.aoe2scenario> <kind> <target>`
  runs `delete-plan` first, refuses blocked plans, then applies the returned
  physical or semantic tombstone recipe to a separate output scenario. It
  supports the same target forms as `delete-plan`, including `system-prefix`,
  `unit-caption`, `unit-caption-prefix`, `unit-caption-contains`, `unit-type`,
  `units-player`, `trigger-name`, `trigger-prefix`, `trigger-contains`,
  `variable-name`, `variable-prefix`, `variable-contains`, `string-text`, `string-prefix`,
  `string-contains`, `units-area`, `effect-type`, `condition-type`,
  `effect-text-prefix`, and `effect-text-contains`.
- `kit scen disconnect <in.aoe2scenario> <out.aoe2scenario>
  <trigger|unit|variable|string> <id>` removes direct structural references to a
  target trigger index, placed unit `reference_id`, trigger variable id, or
  string-table id. `unit-caption <caption> [--player N]` first resolves the
  unique caption to a placed-unit `reference_id`, then applies the same unit
  disconnect. `unit-caption-prefix <prefix> [--player N]` and
  `unit-caption-contains <text> [--player N]` resolve every matched captioned
  unit to stable reference ids, then apply unit disconnect across that batch.
  `unit-type <unit_const> [--player N]` applies unit disconnect
  across every placed unit of that type. `units-player <player>` applies unit
  disconnect across every placed unit owned by one player. `trigger-name <name>` first resolves a unique trigger name to an
  index, then applies the same trigger disconnect. `trigger-prefix <prefix>`
  applies trigger disconnect to every matched trigger index and leaves the
  matched trigger rows in place. `trigger-contains <text>` applies trigger
  disconnect to every trigger whose name or raw effect message contains the
  text marker. `variable-name <name>` first resolves a unique
  variable name to its variable id, then applies the same variable disconnect.
  `variable-prefix <prefix>` and `variable-contains <text>` apply variable
  disconnect to every matched stable variable id. `string-text <exact_text>` first resolves unique exact
  string-table text to its string id, then applies the same string disconnect.
  `string-prefix <prefix>` and `string-contains <text>` apply string disconnect
  to every matched fixed string-table id.
  Trigger disconnect deletes named `trigger_id`
  effect/condition rows. Unit disconnect deletes trigger effect/condition rows
  that point at the placed unit and clears direct `garrisoned_in_id` links from
  other placed units. Variable disconnect deletes trigger effect/condition rows
  that read or write the variable. String disconnect deletes trigger child rows
  that use `string_id`, clears trigger description string IDs, and clears unit
  caption string IDs. `units-area <x1,y1,x2,y2> [--player N] [--unit N]`
  composes the placed-unit cleanup over every matched unit. All routes repair
  counts/display order through the scenario writer and leave the target row
  itself in place. Use it when `delete-plan` reports cleanup blockers:
  disconnect, re-run `delete-plan`, then physically delete only if the follow-up
  plan is clean.
- `kit dat codec-plan <in.dat> --recipe recipe.json` and
  `kit dat codec-patch <in.dat> <out.dat> --recipe recipe.json` rebuild typed
  Effects, Techs, and Sounds sections as needed, patch requested UnitHeader, Civ,
  TerrainRestriction, Terrain, PlayerColour, and semantic Unit tombstone records,
  then deflate/inflate roundtrip and reparse the output. Supported recipe
  operations: create/patch/delete Effects, semantic `disable_effect` clears,
  create/patch Techs, tail-delete
  Techs, guarded non-tail Tech tail-swap delete, paired ability create/delete,
  semantic `disable_ability` effect clears, fixed-width UnitHeader task row fields, patch Civ `name`,
  `tech_tree_id`, `team_bonus_id`, `icon_set`, and indexed resource values,
  terrain restriction passability rows, terrain debug strings (`name`, `name_2`,
  `overlay_mask_name`), create/patch Sound records, semantic Sound mutes,
  semantic Unit disables, high-level Unit availability toggles, fixed-width
  player-color palette fields (`base`, outline, selection, minimap, and
  statistics text colors), tech-tree connection patches,
  tech-tree connection-family clone-appends/removals, top-level
  `reference_rewrites`, and semantic `disconnect_techs` intents. Sound creation
  appends a full typed sound record cloned from an existing sound and rewrites
  the sound table count; direct `sound-create`, `sound-patch`, and
  `sound-delete` commands expose those same recipe paths for single-sound work.
  Broad renumber-all Tech deletion is intentionally
  blocked; non-tail Tech deletion uses a guarded tail-swap strategy only when
  the deleted tech is unreferenced and the current tail tech's references are
  all inside verified rewrite surfaces. Civ unit bodies are typed for the common current-DE unit
  header, storage attributes,
  damage-graphic lists, Type50 attack/armour lists, Creatable cost/train
  locations, action task leading fields, and raw action task tails.
  Count-changing replacements are supported for damage graphics, Type50
  attack/armour rows, train locations, and action task rows by rebuilding
  exactly one unit record and reparsing the DAT. `set_tasks` requires every row
  to carry the exact `raw_tail` bytes from a decoded source row. Random-map
  records are currently framed and counted, but not writable without a non-empty
  current-DE fixture.
- `kit dat delete-plan <in.dat> <section|effect-command|sound-item|graphic-row-kind|unit-header-task|unit-row-kind> <target> [--civ N|--all-civs] [--text]` answers
  what "delete" safely means before an AI mutates a DAT. Supported today:
  individual Effect command rows can be removed by `effect-command 12:3`,
  individual Sound item rows can be removed by `sound-item 22:4`,
  graphic child rows can be removed with `graphic-delta 339:0` or
  `graphic-angle-sound 339:0` when the angle-sound result preserves the
  DAT angle-count invariant,
  unit header task rows can be removed by `unit-header-task 74:0`,
  unit child rows can be removed with `unit-damage-graphic`, `unit-attack`,
  `unit-armour`/`unit-armor`, `unit-train-location`, `unit-drop-site`, or
  `unit-task` targets like `unit-attack 1:74:2`,
  unreferenced Effects can be physically removed with effect-id rewrites, while
  referenced Effects can be semantically cleared to an inert stable ID,
  unreferenced tail Techs can be physically removed, unreferenced non-tail
  Techs can be physically removed by guarded tail-swap when the tail tech can
  be migrated through verified references, referenced Techs can suggest
  `disconnect_techs` as a pre-delete cleanup, and those staged plans now expose
  explicit `cleanup_command` / `cleanup_recipe` fields so an AI does not have to
  infer the next mutation from the delete recipe hint. Paired abilities can be deleted as
  guarded Tech delete plus now-unreferenced Effect delete, referenced-but-unshared
  abilities can be semantically disabled with `disable_ability`, Units can be
  semantically deleted by setting `enabled=0`, and unit delete plans now suggest
  pairing that with `disconnect_units` so verified DAT references are removed
  too. When unit references are in the verified rewrite surface, unit plans also
  expose `cleanup_command` / `cleanup_recipe` as an optional staged
  `disconnect_units` pass; unlike blocked Tech cleanup, the semantic unit delete
  recipe itself remains directly actionable. Sounds can be semantically deleted by setting total/item probabilities
  to `0`. Unit, Tech, Civ, Terrain, Graphic,
  PlayerColour, and Sound plans include a structural reference ledger where the
  codec knows DAT rows that point at the requested ID. Unit and Tech plans also
  include an effect-command ledger. Known command types are reported as
  `typed_effect_command_reference` using the upstream genieutils field layout
  (`target_unit`, `unit_class_id`, `attribute_id`, `amount`) plus decoded command
  meaning and names for known unit-class/resource/operation/attribute/string
  operands, such as `disable_tech`, `spawn_unit`, `modify_tech`,
  `set_tech_cost`/`tech_cost_modifier`, `tech_time_modifier`, and the observed
  DE rename-unit special case of `gaia_set_attribute` on `name_id`. Remaining command-shape unit
  candidates are reported as
  `candidate_effect_command_operand`, and still-ambiguous numeric coincidences
  remain `possible_effect_command_operand` warning rows until every command type
  is semantically decoded. Candidate and possible rows are evidence for
  inspection, not rewrite-supported mutation surfaces. Reference summaries keep
  `possible_operand` as the broad warning-grade class and expose
  `candidate_operand` as the command-shape subset. AGE-oracle command labels
  now cover the DE effect-command families `0..8`, `10..18`, `20..28`,
  `30..38`, `40..48`, and `101..103`; `200` and `201` are named from official
  DE update notes as local-building set/add attribute commands. Types `202`
  and `204` are promoted from the local-building command-family shape plus
  current-DE DAT evidence: `202` is local-building multiply attribute, and
  `204` is a local-building advanced additive attribute form used by current
  emplacement-style rows, with packed attack/armor amounts decoded when
  present. Tech plans report
  required-tech and tech-tree references before allowing mutation.
  Non-tail tech plans refuse referenced delete targets, then inspect the current
  tail tech as the migration source. If the tail's references are all
  rewrite-supported, `delete-plan` emits a `delete_techs` recipe that moves the
  tail tech into the deleted slot, rewrites those verified tail references to
  the slot, and truncates the tail. Tail and non-tail deletion both stay blocked
  when decoded references would be orphaned. Tech-tree building/unit/research
  connection rows can be physically removed by current table index; that deletes
  a progression relationship row only, not the Unit/Tech/Civ records it names,
  and later connection rows shift down. Unreferenced tail PlayerColour rows can
  be physically removed; non-tail palette-slot compaction stays blocked because
  it would renumber stable colour IDs. Graphics, Civs, Terrains,
  TerrainRestrictions, non-tail/referenced PlayerColours, and RandomMaps return
  explicit unsupported reasons instead of pretending physical ID removal is
  safe.
- `kit dat delete <in.dat> <out.dat> <section|effect-command|sound-item|graphic-row-kind|unit-header-task|unit-row-kind> <target> [--civ N|--all-civs]`
  runs `delete-plan` first, refuses unsupported or cleanup-only plans, then
  applies the planner's own recipe hint to a separate output DAT. This is the
  direct D surface for supported DAT deletes; the JSON result includes both the
  delete decision and the codec patch readback.
- Direct child-row delete aliases call that same guarded path:
  `kit dat effect-command-delete <in.dat> <out.dat> <effect_id> <command_index>`,
  `kit dat sound-item-delete <in.dat> <out.dat> <sound_id> <item_index>`,
  `kit dat graphic-delta-delete <in.dat> <out.dat> <graphic_id> <row_index>`,
  `kit dat graphic-angle-sound-delete <in.dat> <out.dat> <graphic_id> <row_index>`,
  `kit dat unit-header-task-delete <in.dat> <out.dat> <unit_header_id> <task_index>`,
  and `kit dat unit-child-delete <in.dat> <out.dat> <damage-graphic|attack|armour|train-location|drop-site|task> <civ_id> <unit_id> <row_index>`.
- `kit dat disconnect <in.dat> <out.dat> <tech|unit> <id>` removes verified DAT
  references to a Tech or Unit without claiming that the target row itself was
  deleted. Use this when `delete-plan` reports a cleanup-only
  `disconnect_techs`/`disconnect_units` recipe: disconnect first, inspect the
  JSON readback for residual references, then re-run `delete-plan` before any
  physical row delete or semantic tombstone. The command refuses unsupported
  sections and no-op requests with no rewrite-supported references. The direct
  mutation report includes a reference summary, not the full reference list; use
  `kit dat refs` when you need the full evidence table.

Minimal codec recipe shape:

```json
{
  "create_effect": {
    "name": "Example Effect",
    "commands": [
      { "type": 0, "a": 9, "b": 21, "c": 3, "d": 7 },
      { "kind": "disable_tech", "tech_id": 244 },
      { "kind": "add_attribute", "unit_id": 9, "attribute_id": 0, "amount": 50 },
      { "kind": "upgrade_unit", "unit_id": 74, "to_unit_id": 75 }
    ]
  },
  "create_effects": [
    {
      "name": "Copied Effect",
      "from_effect": 77,
      "remove_commands": [0],
      "append_commands": [
        { "kind": "enable_unit", "unit_id": 83 }
      ]
    }
  ],
  "effects": [
    {
      "id": 1409,
      "remove_commands": [1],
      "append_commands": [
        { "kind": "enable_unit", "unit_id": 83 }
      ]
    }
  ],
  "create_tech": {
    "from": 22,
    "name": "Example Tech",
    "effect_id": 1409
  },
  "create_ability": {
    "from_tech": 22,
    "name": "Copied Ability",
    "append_commands": [
      { "kind": "disable_tech", "tech_id": 244 }
    ]
  },
  "unit_headers": [
    {
      "id": 83,
      "tasks": [{ "index": 0, "action_type": 3, "work_range": 2.25 }]
    }
  ],
  "unit_availability": [
    { "unit_id": 83, "civ_ids": [1, 2], "enabled": false },
    { "unit_id": 448, "all_civs": true, "enabled": true }
  ],
  "civs": [
    {
      "id": 1,
      "name": "Example Civ",
      "team_bonus_id": 23,
      "resources": [{ "index": 0, "value": 1234.5 }]
    }
  ],
  "reference_rewrites": [
    { "target": "unit", "from": 83, "to": 9999 }
  ],
  "disconnect_techs": [244],
  "disconnect_units": [83],
  "tech_tree": {
    "rewrites": [
      { "target": "tech", "from": 244, "remove": true },
      { "target": "unit", "from": 83, "to": 9999 }
    ],
    "unit_connections": [
      { "index": 0, "required_research": -1, "units": [83, 9999] }
    ]
  }
}
```

`create_ability` is a paired Tech+Effect helper. It can take explicit
`commands`, or it can clone commands from `from_tech`'s current `effect_id` or
an explicit `from_effect`, then append `append_commands` for the actual tweak.
The readback verifies the new Tech+Effect structure and command rows; it does
not prove in-engine trainability, UI placement, or command behavior.

Effect command recipes accept raw `{ "type", "a", "b", "c", "d" }` rows for
unknown/advanced cases, plus named helper rows where the operand semantics are
typed: `disable_tech` with `tech_id`, `enable_unit` with `unit_id`, and
`upgrade_unit` with `unit_id` and `to_unit_id`, and
`set_attribute`/`add_attribute`/`multiply_attribute` with `unit_id`,
`attribute_id`, `amount`, and optional `unit_class_id` or `unit_class_name`.
UGC `xsEffectAmount`
helpers are also available: `resource_modifier` with `resource_id`,
`operation_id`, and `amount`; `resource_multiplier` with `resource_id` and
`amount`; `spawn_unit` with `unit_id`, `building_id`, and `amount`;
`modify_tech` with `tech_id`, `tech_attribute_id`, and `amount`;
`set_unit_attribute`/`add_unit_attribute`/`multiply_unit_attribute` with
`unit_id`, `attribute_id`, and `amount`; these exact-unit attribute helpers
emit the base AGE-compatible attribute commands (`0`, `4`, `5`) with
`unit_class_id=-1`, not the team-scoped command family. `set_tech_cost`/`add_tech_cost` with
`tech_id`, `resource_id`, and `amount`; and `tech_time_modifier` with
`tech_id`, `operation_id`, and `amount`. Resource, operation, and modify-tech
attribute operands may also be supplied by XS-constant-backed names such as
`resource: "cAttributeKills"`, `operation: "cAttributeAdd"`, or
`tech_attribute_name: "cAttrSetGoldCost"`. Unit-attribute helper recipes also
accept `attribute` / `attribute_name` such as `cAttack` or `hit_points`. If a
numeric ID and a name are both present they must agree. Class-filtered
attribute helpers accept XS-style class constants such as `cCavalryClass` and
plain names such as `cavalry` or `all_units`.

`upgrade_unit` commands expose both `source_unit` and `target_unit` typed
references. AGE-oracle command families expose their scoped operands:
attribute commands expose `target_unit`, resource commands expose
`resource_id`/`operation_id`, spawn commands expose `spawn_unit` and
`spawn_building`, and modify-tech commands expose `target_tech`. Attribute
commands with `attribute_id` 8 or 9 also decode AGE's packed armor/attack
amount into `packed_type_id` and `packed_amount`. With those promotions,
`kit dat refs`, `delete-plan`, and
top-level `reference_rewrites` can see and rewrite those operands instead of
treating them as ambiguous numeric sightings.
The semantic overlay also surfaces read-only operand fields such as
`resource_id`, `operation_id`, and `tech_attribute_id` for resource,
tech-cost/time, and modify-tech command families, with XS-constant-backed names
where known (`resource_name`, `operation_name`, `tech_attribute_name`), so an AI
does not have to re-derive those meanings from raw `a/b/c/d` slots.

Effect create and patch recipes support the same copy-and-tweak command flow:
set an entire command list with `commands`, optionally clone from `from_effect`
when creating a new effect, delete specific command rows with
`remove_commands`, then add rows with `append_commands`. Command indices are
validated, duplicates are rejected, and the rebuilt effect command list is
reindexed during readback. When Kit clones existing commands for these flows it
re-expresses known command families as helper recipes with named operands, while
unknown command families stay raw `{type,a,b,c,d}` rows.
Top-level `disable_effect` / `disable_effects` is the semantic delete wrapper
for an effect row. It preserves the effect ID and clears the command list; every
tech that points at that effect will now point at inert behavior. Use
`delete_effects` only when physical row removal and effect-id renumbering are
really intended.

Top-level `reference_rewrites` own the broader currently-verified reference
surfaces: tech-tree ledgers, `tech.required_techs`, and typed effect-command
references such as `disable_tech`, `upgrade_unit`, `spawn_unit`, and
`modify_tech`. They report counts for
required-tech fields, effect commands rewritten/removed, and tech-tree values touched. Use
`tech_tree.rewrites` when you only want to mutate the tech-tree section.
Top-level `disconnect_techs` is the designer-intent wrapper around verified
tech reference removal. It leaves the physical tech row in place, removes the
currently verified references, reparses/readbacks the result, and reports a
`disconnected_techs` residual summary. If ambiguous effect-command operands or
other references remain, `complete` is false and the report tells the AI to
inspect with `kit dat refs` before claiming the tech is fully inert in-engine.
Top-level `disconnect_units` is the parallel stable-ID wrapper for unit
references. Today it removes verified tech-tree references, typed
effect-command references, unit-header task `object_id` references, and
known unit-record references (`blood_unit_id`, projectile, task, drop-site,
train-location, and linked-building rows), then reports a `disconnected_units`
residual summary. Ambiguous effect-command operand sightings remain visible
until those command semantics are promoted.
Top-level `unit_availability` is the designer-intent wrapper for setting a
unit record's DAT `enabled` flag by `civ_id`, `civ_ids`, or `all_civs`. It
preserves stable unit IDs and reports before/after enabled values. This is not
a full trainability claim: train locations, tech-tree links, triggers, civ
bonuses, and in-engine behavior may still decide whether the unit can actually
be produced or appears in a particular scenario.
Top-level `create_player_colour` / `create_player_colours` append
count-prefixed PlayerColour rows by cloning an existing row and applying
optional colour field overrides. This is structure-verified by count/header
readback; it does not claim the DE engine will expose or use newly appended
player-colour slots.
Top-level `delete_player_colour` / `delete_player_colours` removes only
unreferenced tail PlayerColour rows. This is intended as the inverse of
Kit-created appended palette rows: it preserves every earlier stable colour ID,
refuses non-tail compaction, refuses rows still referenced by graphics, rebuilds
the counted PlayerColour section, and reparses/readbacks the final count.
Top-level `delete_ability` / `delete_abilities` is the designer-intent wrapper
for the existing paired ability delete plan. It asks the Tech delete planner
whether the ability's tech row is removable, so tail abilities use physical
tail delete and eligible non-tail abilities use the guarded tail-swap strategy.
It still refuses shared effects or any ability referenced outside its own
`tech.effect_id` edge, then expands to the existing `delete_techs` plus
`delete_effects` engines and reports `deleted_abilities` with the paired
tech/effect IDs.
Top-level `disable_ability` / `disable_abilities` is the semantic delete
wrapper for the common "remove this behavior, preserve the IDs" case. It
requires the ability's `tech.effect_id` to point at a private effect, then
clears that effect's command list and reports `disabled_abilities`. The tech
row, effect row, UI placement, prerequisites, and tech-tree references remain
structurally present; this is an inert-behavior operation, not a physical row
delete.
Designer-facing ability shortcuts expose the same recipe engine without asking
an AI to hand-author a full JSON file for common cases:
`kit dat ability-create <in.dat> <out.dat> --from-tech N --name TEXT` clones a
research row, creates a paired effect, and accepts raw command tuples
`type,a,b,c,d` or full JSON command objects. `kit dat ability-patch` edits the
paired tech/effect names and command rows, `kit dat ability-disable` performs
the semantic delete above, and `kit dat ability-delete` performs the guarded
paired physical delete when the planner allows it. Use full `codec-patch`
recipes for unusual fields, shared-effect migrations, or multi-object edits.
- Early-section scan through `Graphic`, including terrain restrictions, player
  colours, sounds, graphic presence flags, and every present graphic record.
- Exact DebugString decoding: marker `0x0A60`, uint16 byte length, UTF-8 bytes.
- Graphics inspection for `name`, `file_name`, `particle_effect_name`, `slp`,
  layer/player-colour fields, base sound/Wwise sound ids, frame/angle/delta
  counts, animation timing, sequence flags, internal id, `GraphicDelta` rows,
  current-DE `GraphicAngleSound` rows, and byte spans.
  `kit dat graphics` supports `--particle` and `--name-contains TEXT`; singular
  `kit dat graphic` is flat by default and exposes rich marker/span internals
  only with `--spans`/`--raw`.
- Graphic patching for `name`, `file_name`, `particle_effect_name`, `slp`, and
  safe fixed-width scalar fields: `is_loaded`, `old_color_flag`, `layer`,
  `player_color`, `rainbow`, `transparent_selection`, 4-value `coordinates`,
  `sound_id`, `wwise_sound_id`, `frame_count`, `speed_multiplier`,
  `frame_duration`, `replay_delay`, `sequence_type`, `mirroring_mode`, and
  `editor_flag`. String edits rebuild the owned graphic record; numeric edits
  are fixed-width writes. Recipes can also replace complete `set_deltas` and
  `set_angle_sounds` child-row lists, including count-changing delta edits and
  angle-sound enable/disable. Non-empty `set_angle_sounds` must provide exactly
  one row per existing `angle_count`.
- Civ and unit indexing for current `VER 8.8+` / `VER 8.9` AoE2DE records,
  including civ type, name, resource count, tech-tree id, team-bonus id,
  icon-set id, per-civ unit-slot counts, present unit count, unit type, unit id,
  unit name, object class name where known, core graphic/icon/enabled/hit-point
  fields, and record span.
  `kit dat units` supports exact id lookup, DAT debug-name substring lookup via
  `--name` or `--name-contains`, civ filtering, and honest class buckets
  (`building` = type byte 80/90, `unit` excludes 80/90, `creatable` means a
  parsed train-location block exists).
- Tech table indexing for debug name, civ, effect id, type, icon id, required
  tech count, research-location count, and record span. `kit dat techs` supports
  `--name-contains TEXT` for teardown of large custom research tables.
- Effects table indexing, including effect debug name, command count, command
  rows (`type`, `a`, `b`, `c`, `d`), semantic overlays, and byte spans. The
  semantic overlay preserves the upstream genieutils field names (`target_unit`,
  `unit_class_id`, `attribute_id`, `amount`), names known command types and unit
  classes/attributes/resources/operations, and promotes proven IDs such as `disable_tech` amount values into
  typed references. `kit dat effects` can filter by `--command-type` and/or
  `--operand`; operand matching checks A/B/C and exact-integer D values and
  returns compact `matching_commands` snippets. `kit dat command-matrix`
  summarizes every effect-command type, operand distribution, promoted typed
  reference family, attribute/resource/operation/tech-attribute sighting, and
  bounded examples; this is the
  dark-frontier map for adding future helpers without guessing.
  `kit dat effect-explain` and `kit dat tech-explain` give the same decoded
  rows back as authoring prose. Known local-building rows are surfaced as
  `200` set, `201` add/subtract, `202` multiply, and `204` advanced packed
  add/subtract. Packed attack/armor values are decoded as class plus delta, so
  values such as `771` read as `attack[pierce +3]` instead of an opaque float.
  The writer accepts matching named helpers, including
  `multiply_local_building_attribute`, `add_local_building_armor`, and
  `add_local_building_attack`, and write reports include authoring notes when a
  recipe relies on a structure-known but not yet engine-verified path.
- Unit-header indexing via `kit dat unit-headers`, including present/absent
  header records and decoded default task rows using the same current-DE task
  layout as per-civ unit action tasks. `kit dat unit-header <dat> <id>` returns
  one header, and codec recipes can patch fixed-width unit-header task row
  fields or replace the whole unit-header task list with `set_tasks`. Task rows
  expose named leading fields plus `raw_tail` bytes for the still-unnamed
  current-DE suffix. Count-changing task list edits are supported through unit
  and unit-header `set_tasks` recipes only when each row preserves the exact raw
  tail length.
- `kit dat availability <empires*.dat> <unit_id> [--civ N|--all-civs]` emits
  per-civ availability cards for a unit. Each card combines unit record
  presence, the DAT enabled flag, creatable/train-location rows, and decoded
  tech-tree references into a structural status such as disabled record,
  enabled but not creatable, or plausibly trainable structurally. It remains
  `structure_verified_not_engine_verified`: triggers, civ bonuses, and actual
  editor/game behavior can still override the structural diagnosis. Disabled
  but present unit records include a narrow `recipe_hint` for the verified
  `unit_availability` enabled-flag patch; the hint deliberately does not claim
  to repair train locations, tech-tree links, civ bonuses, triggers, or engine
  behavior.
- Tech/research table indexing, including required techs, resource costs,
  required count, civ id, effect id, icon/type/language DLL ids, debug name,
  repeatable flag, research locations, and byte spans.
- Game metrics plus tech-tree header/connection parsing through EOF. `kit dat
  tech-tree` is compact by default; use `--full` to emit the complete age,
  building, unit, and research connection arrays. Codec recipes can patch
  building/unit/research connection rows by rebuilding the whole tech-tree
  section and verifying readback. Supported fields include `id`, `status`,
  line/location fields, `upper_building`, `required_research`,
  `enabling_research`, counted child lists (`buildings`, `units`, `techs`), and
  fixed common rows (`unit_research`, `mode`, plus building
  `units_techs_total`/`units_techs_first`). `tech_tree.create_building_connections`,
  `tech_tree.create_unit_connections`, and `tech_tree.create_research_connections`
  append counted connection rows by cloning an existing row and applying the
  same field overrides for that connection family; this is structure-verified
  only, not an in-engine trainability or UI-placement claim.
  `tech_tree.delete_building_connections`, `tech_tree.delete_unit_connections`,
  and `tech_tree.delete_research_connections` remove counted connection rows by
  current table index and verify the rebuilt section parses. This removes
  relationship rows only; it does not delete Unit/Tech/Civ records. `tech_tree.rewrites` can replace or
  remove a referenced `tech` or `unit` ID across the decoded tech-tree ledgers;
  `remove` deletes counted-list entries and clears scalar/fixed-slot references
  to `-1`. Age rows are included in rewrite helpers, but direct age-row patching
  remains future work until a concrete authoring need appears.
- Fixed-width Unit patching for `class`, `hit_points`, `line_of_sight`,
  `movement_type`, `standing_graphic_1`, `standing_graphic_2`,
  `dying_graphic`, `blood_unit_id`, `icon_id`, `enabled`, the three unit
  `attributes` rows, and damage-graphic rows.
- Type50/Creatable patching for `type50_projectile_unit_id`,
  `type50_max_range`, `type50_blast_width`, `type50_attack_graphic`,
  `type50_blast_damage`, attack/armour rows, `train_time_0`,
  `train_unit_id_0`, `train_button_id_0`, `train_hotkey_id_0`,
  `creatable_button_icon_id`, `creatable_button_hotkey_action`, resource cost
  rows, arbitrary train-location rows, and fixed-width action task rows
  (`action_type`, object/terrain targets, work values/range, targeting flags,
  and combat level). Recipes can also replace the full decoded
  `set_damage_graphics`, `set_type50_attacks`, `set_type50_armours`, and
  `set_train_locations` lists, plus action `set_drop_sites`, including
  count-changing shorter/longer lists.
- Unit patches always write to a separate output file, with re-index
  verification, exact readback, unchanged civ/unit counts, and neighbor
  canaries. Fixed-width patches require unchanged inflated length; decoded list
  replacement rebuilds exactly one unit record and allows the measured unit
  record length delta.
  Direct `patch-unit` row flags use `INDEX,field=value[,field=value...]`.
  Supported row flags are `--attribute`, `--damage-graphic`, `--type50-attack`,
  `--type50-armour`/`--type50-armor`, `--cost`, `--train-location`, and
  `--task`. Use `attribute_types=-1:-1:-1:-1` for the four task attribute-type
  slots.
- Graphic record patching for `name`, `file_name`, `particle_effect_name`,
  `slp`, frame/scalar fields, `set_deltas`, and `set_angle_sounds`, always to a
  separate output file, with re-index verification, exact readback, unchanged
  graphics count, exact inflated-length delta, and neighbor canaries.
- Graphic record creation via recipe `create_graphic` / `create_graphics`.
  Creation copies a present template graphic, assigns the next graphic id,
  rewrites the internal id field, applies supported graphic fields, bumps the
  graphics count, appends the pointer/record, and verifies count/readback,
  exact length delta, and adjacent record canaries.
- Unit record creation via recipe `create_unit` / `create_units`. Creation
  copies a template unit, assigns the next unit id consistently across selected
  civs, appends per-civ unit pointers/records, applies the same supported unit,
  Type50, and Creatable fields/list replacements as unit patching, and verifies per-civ
  counts/readback and exact length delta.
- Recipe patching with `kit dat patch <in.dat> <out.dat> --recipe recipe.json`.
  Recipes apply graphic patches, graphic creation, unit creation, and unit
  patches first, then codec-backed Effects/Techs/Civ/Terrain/Sound/
  PlayerColour operations, verifying each operation before continuing. See
  `docs/DAT_RECIPE_EXAMPLE.json`.
- Recipe planning with `kit dat plan <in.dat> --recipe recipe.json`, which runs
  the same in-memory patch and verification path but writes no output file.

`kit dat graphics`, `kit dat effects`, `kit dat techs`, and `kit dat units` return the first 200
rows by default to avoid accidental multi-megabyte terminal dumps. Graphics
truncation is loud on stderr because silent truncation can manufacture false
conclusions; use `--limit 0`/`--all` when a full graphics table is intentional.

Current FX support:

- `kit fx new <name> --preset trail|explosion|aura|projectile-fire` prints a
  flat AoE2DE particle descriptor JSON using the verified `AtlasFile`,
  `ImageFirst`, `ImageCount`, `Type`, alpha, scale, fire, and fog keys.
- `kit fx bind --dds atlas.dds --grid RxC --into resources/_common --dat
  empires2_x2_p1.dat --unit N --slot flying|standing|attack|dying --from GID
  --name fx_name` writes the particle three-file triangle and patches the DAT.
  `--frames N --frame-size WxH` can be used instead of `--grid` when the atlas
  is sized by known frame cells.
  The generated convention is descriptor `particles/<name>.json`, atlas image
  `particles/textures/atlases/<name>_atlas.dds`, and atlas metadata
  `particles/<name>_atlas.json`.
- `kit fx bind --atlas atlas.png ...` is a convenience path that shells out to
  ImageMagick with the pinned engine-validated invocation:
  `convert <png> -define dds:compression=dxt5 -define dds:mipmaps=0 <dds>`.
  If ImageMagick is unavailable, use `--dds`.
- `kit fx lint <mod-or-resources/_common>` validates descriptor JSON,
  backslash-style `AtlasFile`, atlas metadata frame coverage, DXT5/BC3 DDS
  headers, and DAT `particle_effect_name` references when a DAT is available.
- FX reports carry `structure_verified_not_engine_verified`: green output means
  the asset triangle and DAT readback are internally consistent, not that the
  particle has rendered in-engine.

Current CBA support is intentionally in `pkg/cba` / `cmd/cba` and the
`kit cba ...` shim, not folded into generic replay interpretation:

- `kit cba replay progression` extracts CBA production/research progression
  proxies from decoded queue/research commands.
- `kit cba replay razes` extracts building-pressure and raze/villager
  candidate events with explicit behavior-inferred confidence labels.
- `kit cba replay perf` emits one Python-schema-compatible row per player plus
  timestamped metric events. Row fields include `own_first_vill_s`,
  `own_vills`, `own_razes`, `own_first_raze_s`, and the corresponding
  `team_*` fields so setup/enabler players can be credited through team tempo.
  The `events` array includes `raze_candidate`, `villager_gain_candidate`,
  `unit_queued`, and `production_building_started` events. Kill attribution is
  still deliberately not claimed; `units_lost` is a compatibility field for
  own-side checksum object-removals, not engine-confirmed unit deaths, and
  `units_produced` is own-side checksum additions. Small removal samples can be
  noisy; the Combat axis ignores rows below its production floor.
- `kit cba replay doctrine` overlays CBA matchup doctrine on the existing
  replay primitives. It reports civ duty, phase-boundary evidence, pressure/raze
  candidates, production signals, and combat proxies with explicit confidence
  labels. This is a CBA-domain interpretation layer, not a generic replay decode
  claim.
- `kit cba balance` exposes a source-labeled, intentionally partial CBA Requiem
  balance table for version/civ parameters such as razes-to-villager,
  castle/imperial kill thresholds, unit counts, and spawn seconds. It includes
  recent V286/V289/V292/V293 changelog rows from `announcements_recent.txt`,
  older V3.1/V8b/V9/V10 rows from `announcements_full.txt`, and V292
  trigger-graph global attr-33 threshold evidence. It does not silently default
  missing civs; absent rows are unknown until changelog or binary evidence maps
  them. The known capture gap is 2023-04-02 to 2024-11-19.
- `kit cba trigger-razes` reads a CBA Requiem replay trigger graph directly and
  maps the civ-start technology activators to the per-player raze reward rungs.
  It cross-references technology IDs through a supplied DE DAT, yielding the
  baseline razes-to-villager buckets that changelog mining cannot recover. On
  the current V292 anchor it finds 40 reward rung triggers, 40 activator
  triggers, 5 buckets, and Mapuche in the 1-raze bucket. Its confidence label is
  `structure_verified_trigger_graph_plus_dat_tech_name_cross_reference_not_engine_runtime_verified`.
- `kit cba trigger-spawns` reads the same CBA Requiem trigger graph and maps
  civ-start technology conditions to two authored spawn-wave parameters:
  `spawn_seconds` from `Spawner N sec` display-instruction messages, and
  `spawn_unit_count` from the repeated civ-gated spawn triggers' condition type
  `4` field `2`. This is intentionally CBA-specific domain logic. On the 078
  V292/V293 random-position Requiem anchor it currently finds 72 spawner-second
  triggers, 976 spawn-count triggers, 59 complete civ/tech rows, and the key
  examples Mapuche `80 units / 10 sec`, Khmer `30 / 15`, Persians `60 / 13`,
  Malay `120 / 6`. Rows whose technology IDs map to multiple values are
  withheld and reported as conflicts rather than guessed. Its confidence label
  is `structure_verified_trigger_graph_plus_dat_tech_name_cross_reference_not_engine_runtime_verified`.
- Current CBA V292 note: scenario files written as DE `1.57` are readable and
  writable in Kit for section-preserving recipe edits. For ladder integrity,
  `kit replay info <replay>` remains the preferred path because it parses the
  embedded `VER 9.4` scenario graph from the actual submitted game.
