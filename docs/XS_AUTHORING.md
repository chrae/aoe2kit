# XS Authoring Model

AoE2Kit treats the AoE2DE UGC Guide as the current public source of truth for
XS authoring. Source repo:

`https://github.com/Divy1211/AoE2DE_UGC_Guide`

The older Forgotten Empires-era XS notes are superseded for Kit decisions. Keep
them only as historical context when encountered.

## Runtime Placement

For custom scenarios, the UGC guide's programmer reference says the scenario's
Map tab names an XS file without the `.xs` suffix, and trigger `Script Call`
effects call parameterless functions from that XS script.

AoE2Kit supports two XS placement modes:

- preferred for generated scenarios: parser-style embedded carrier trigger;
- legacy/explicit external-file mode: scenario attachment fields plus a shipped
  `.xs` file.

The carrier mode mirrors AoE2ScenarioParser's `xs_manager.add_script(...)`
pattern. It stores the full XS source in the `message` field of a disabled
trigger `script_call` effect, usually on a disabled trigger named `XS SCRIPT`.
That source travels inside the `.aoe2scenario` itself, so the scenario and XS no
longer drift when copied between machines. `kit scen xs` reports these as
`embedded_carriers`, analyzes their functions, and excludes the carrier payload
from ordinary runtime `script_call` function invocations.

Use this recipe shape for carrier mode:

```json
{
  "xs": {
    "mode": "carrier",
    "carrier_title": "XS string",
    "carrier_trigger_index": 0,
    "content_file": "MyScript.xs"
  }
}
```

By default `mode:"carrier"` clears the older attachment fields and replaces the
first existing carrier, avoiding stale duplicate XS blobs. Set
`carrier_trigger_index` when converting an existing placeholder `script_call`
trigger into the carrier slot, which keeps the trigger list stable. Use
`"mode":"attachment"` only when you deliberately want the old external-file
deployment path, or `"mode":"attachment_and_carrier"` when you are comparing both
paths.

The same carrier workflow is available without hand-writing a JSON recipe:

```sh
./kit scen xs attach in.aoe2scenario out.aoe2scenario --xs MyScript.xs --name MyScript.xs --text
./kit scen xs deploy out.aoe2scenario path/to/mod-or-profile-root --force --check --text
./kit scen xs embed in.aoe2scenario out.aoe2scenario --xs MyScript.xs --replace-trigger 0 --text
./kit scen xs compare out.aoe2scenario MyScript.xs --text
./kit scen xs extract out.aoe2scenario --out /tmp/MyScript.xs --force --text
```

`attach` writes the executable scenario XS fields. `deploy` copies the
scenario-carried source to `resources/_common/xs/` under the supplied deploy
tree, then `--check` verifies the scenario fields, on-disk entry file, includes,
and cross-file `extern` requirements. This is the normal path when the XS should
actually run in DE.

`embed` writes the carrier and clears stale attachment fields by default.
`compare` checks scenario-carried XS against a source file and exits non-zero
when normalized content differs. `extract` pulls the script payload back out of
the scenario; it does not include the carrier title wrapper line. If executable
attachment XS exists, `compare` and `extract` use it by default; pass `--carrier`
or `--trigger` to target parser-style carrier escrow instead.

Attachment mode sets:

- `Map.script_name`
- `Files.script_file_path`
- `Files.script_file_content`

The v3 telemetry probe proved an important distribution caveat: embedded
`script_file_content` is portable source and an inspection aid, but DE still
resolved the runnable script path against `resources/_common/xs/` in that test.
The first engine run failed with XS Error 0354 until the matching `.xs` file was
present on disk. The v8 assertion fixture refined this: DE opened the literal
value stored in `Map.script_name`. A scenario where `Map.script_name` was bare
but `Files.script_file_path` included `.xs` failed by trying to open the bare
name. AoE2Kit now writes both fields consistently with the `.xs` extension.

So the safe authoring rule is:

- use carrier mode for single-file generated XS whenever possible;
- use attachment mode only when a real external module tree is intentional;
- if using attachment mode, ship the same `.xs` file under the runnable bundle's
  `resources/_common/xs/`;
- set both scenario XS name/path fields to the same extension-bearing filename;
- use ordinary trigger `script_call` effects to invoke named parameterless XS
  functions.

Front Towers provides live-author evidence for the parser-style carrier pattern:
its scenario attachment fields are empty, while a disabled trigger carries the
full XS source in a `script_call` message. Kit's carrier writer is
structure-verified by round-trip tests; keep it labeled not engine-verified until
we run a local Kit-authored carrier fixture.

## Multi-File XS

The UGC guide documents ordinary XS imports with:

```xs
include "absolute/or/relative/path/to/file.xs";
```

This is plain XS syntax, not RMS `#includeXS`. `#includeXS` belongs to random
map scripts; it is not the syntax used inside `.xs` files.

Working shipped mods prove multi-file XS is used in practice:

- `RPG_scripts.xs` includes `RNG_Function.xs`, `AI_Scripts.xs`, and
  `smart_mob_control.xs` at lines 1001-1003, after 1000 extern declarations.
- `smart_mob_control.xs` includes generated data files at the top and
  `mob_targeting.xs` at EOF.
- `zone_constants.xs` and `zone_arrays.xs` are generated data modules.

The practical Kit model is therefore a module tree:

- a named entry script selected by the scenario;
- sibling helper modules included with `include "file.xs";`;
- generated data modules kept separate from handwritten runtime logic.

Resolution against `resources/_common/xs/` is proven for the main script path by
the v3 probe. GoKu's layout strongly supports the same folder as the include
root for mod-shipped sibling XS files. Include ordering, deduplication, and
cycle behavior are not yet characterized; avoid cycles and put constants/data
before code that needs them.

## Trigger Bridge

There are two relevant bridge directions.

Scenario trigger to XS:

- set scenario XS name/path/content with the `xs` recipe block;
- add trigger effect `script_call` with `message`, for example `BootProof();`;
- call parameterless XS functions from trigger effects.

XS to scenario variables:

- GoKu declares `extern int xsVariable0 = -1;` through
  `extern int xsVariable999 = -1;` in `RPG_scripts.xs`.
- This is working-corpus evidence that scenario trigger variables can be exposed
  as XS globals under the `xsVariableN` naming convention.
- The v8 fixture proved cross-file symbols need `extern`: plain `const int`
  constants in one included module were invisible to a sibling included module.
  Shared generated constants should therefore use `extern const int NAME = N;`.

AoE2Kit should prefer generated constants/helpers around `xsVariableN` rather
than asking authors or AIs to hand-track raw variable numbers in large systems.

## File I/O

The UGC function reference documents `.xsdat` File I/O:

- `xsCreateFile`
- `xsOpenFile`
- `xsWriteString`, `xsWriteInt`, `xsWriteFloat`, `xsWriteVector`
- `xsReadString`, `xsReadInt`, `xsReadFloat`, `xsReadVector`
- file-position helpers and size helpers

Important behavioral claims from UGC:

- `xsCreateFile` creates/appends to an `.xsdat` file named after the RMS or
  scenario.
- In multiplayer, create/write is duplicated to each player.
- `xsOpenFile` requires the file to exist for all multiplayer players with the
  same data to avoid out-of-sync risk.
- String data is length-prefixed with a 32-bit length.

GoKu does not use File I/O. The v8 fixture proved scenario-context file
creation: DE created the expected `.xsdat` sidecar under the profile folder.
The first run produced a 0-byte file, strongly suggesting writes are buffered
until `xsCloseFile`. The v8.1 fixture proved that close-before-victory fires,
but also proved that reopening with `xsCreateFile(false)` truncates the
sidecar. The v8.2 fixture therefore opens once, writes all rows, closes once,
then uses `xsCreateFile(true)` for a separate append probe. Its first engine
run produced a nonzero 492-byte sidecar, with full readback still pending.

DE holds an exclusive lock on a live `.xsdat` while the game is loaded. Pull or
inspect sidecars only after the match has unloaded or DE has exited.

The v9 persistence probe engine-verified cross-scenario reads: a reader
scenario opened a writer scenario's sidecar with
`xsOpenFile("RTV9A_Persistence_Writer")`. Do not include the `.xsdat`
extension; the same name with the extension failed because DE appends it
internally. This makes `.xsdat` a viable substrate for same-profile,
multi-scenario campaign persistence.

Future carrier probes should hold each result long enough to be sampled by the
replay checksum stream. v9 used 10-second spacing and one naming attempt was
overwritten between approximately 16-second checksum samples.

## Rule Model

UGC documents top-level `rule` blocks with active/inactive state,
`minInterval`/`maxInterval`, optional `highFrequency`, group, priority, and a
body. GoKu uses inactive one-second rules for per-player mob control.

For Kit-generated XS, default to low-frequency `minInterval` rules unless a
feature explicitly requires high-frequency logic. The diagnostic v1 crash
history makes spammy text/output loops a known hazard.

## Known Hazards

UGC's bug pages and Kit probes give these practical authoring constraints:

- XS files do not reliably transfer through multiplayer lobbies; bundle the
  files in a mod or prove a parser-embedded workflow before relying on it.
- `xsResearchTechnology` twice for the same tech/player can crash.
- `%` formatting in `xsChatData` is dangerous; string concatenation is safer.
- Huge formatted ints in `xsChatData` can crash.
- `xsChatData` is not replay telemetry in the v3 probe: no `SDSTELEM` strings
  serialized, while resource writes did serialize via sync word 1.
- Avoid `highFrequency` and looping display/chat diagnostics unless there is a
  hard reason and a throttle.

## Gap 6 Impact

The source-of-truth reset expands Gap 6 from "attach one XS blob" to "author a
runnable XS module bundle plus trigger calls."

Current Kit support:

- scenario `xs` recipe fields attach one entry script by name/content;
- `script_call` trigger effects can call parameterless functions from that
  script via the effect `message` field;
- scenario writer emits consistent extension-bearing XS names in both script
  name/path fields, matching the v8 engine-run finding;
- `kit xs bridge` emits `extern int xsVariableN` and `extern const int` named
  constants for cross-file visibility;
- generated multi-file XS can be shipped in `resources/_common/xs/` by the
  surrounding mod/package workflow.

Still to build if we want first-class XS authoring:

- a bundle/package helper that copies the entry script and all included sibling
  `.xs` files into `resources/_common/xs/`;
- an XS include scanner/linter with missing-file, cycle, and path checks;
- generated `xsVariableN` naming maps from scenario variables;
- an optional File I/O probe fixture to convert UGC's File I/O documentation
  into local engine-verified knowledge.
