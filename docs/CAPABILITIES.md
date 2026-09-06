*Exact invocations, flags, and returned JSON keys for every command are in `API_REFERENCE.md`
(machine-readable twin: `api_reference.json`), generated from the binary itself.*

Here's the full map, domain by domain. Everything below I verified against the tree and manual tonight — where a capability has a hard boundary, I say so, because the Kit itself says so (every report carries a `structure_verified` vs `engine_verified` label, and that distinction is the spine of the whole design).

---

## 1. Scenario files (`kit scen`, `pkg/scenario`)

**Read:** Full parse of DE 1.55–1.58 `.aoe2scenario` — every section from DataHeader through Files. It proves its own correctness by rebuilding the decompressed body byte-identically. On top of the raw parse sit graded views:
- `info`/`check`/`describe` — orientation: players, map, units, trigger fingerprint, embedded AI files, full field tree on request.
- `triggers` (now with `--conditions` — full condition decode including operation add/subtract, from tonight), `effects` (per-type summaries, `--census` for count-only passes over huge files, `--where` to stream one effect family), `units --named`, `strings`, `map`.
- `glossary` — the "first command an AI runs on an unfamiliar scenario": effect families, name prefixes, variable ids, XS calls, unit/tech/attribute vocabulary. `idioms` — names detected *techniques* (colored text, HUD loops, caption objects, XS bridges, relay graphs) with confidence, without claiming author intent.
- `diff` — two scenarios compared by stable parsed facts.

**Write:** Recipe-driven patching to a *separate output file*, always re-verified by reopen + rebuild + invariant checks. It can append full trigger structures (20 effect types, 8 condition types), edit existing triggers, add/edit/remove units, set terrain rectangles, attach embedded XS, flip player slots, edit GlobalVictory/diplomacy/starting resources. `smoke` writes a canonical editor-visible test pattern. Deliberately not exposed yet: structural map resize, arbitrary existing-unit field edits.

**Safety:** `lint` catches engine hazards (looping display spam, corrupt strings, invariant violations) and **cites engine-fact IDs** from the evidence ledger. `deploycheck` statically preflights XS deployment — script_name/path literal match, on-disk resolution, include chains, cross-file `extern` discipline — the exact failure classes we hit live (Error 0354) turned into pre-flight checks.

## 2. Replays (`kit replay`, `pkg/replay` — the largest domain, ~24k lines)

**Identity & metadata:** `summary` is the mgz-style rollup — record hash, versions, scenario name, map hash, active data-mod identity, lobby settings (game type, speed, treaty, population...), duration, players with civ/color/team/profile-id/rating. `identify` fingerprints against growable registries (trigger-graph hash first, terrain hash fallback).

**Events & actions:** `events` (chat/taunts/flares/resigns/telemetry with result inference), `actions` (decoded MOVE/ORDER/BUILD/DELETE/RESIGN/DE_QUEUE/RESEARCH... plus raw-preserved unknowns), `player-events` (filterable canonical event table), `chat` (merges body chat + lobby-header chat, dedupes echoes, tags prior-session `backlog` injection — the hallucination fix), `feedback`, `camera` (the POV player's own viewlock stream — attention data).

**The sync/checksum layer — our decode campaign's instrument:** `sync` surfaces the op=2 heartbeat with decoded word semantics (word_1 = stockpile, word_2 = unit-type sum, word_6 = object count, word_8 = player, word_10 = instance-id sum; 3/4/7 partial; 9 still caged). `player-series` turns that into per-player state curves with typed add/remove estimates. `checksum-probe` matches expected fixture pulses. `combat` joins net object-loss windows to nearby targeted commands — pressure stories without kill claims.

**Objects:** `objects`/`object-state`/`object-shapes` — the initial-object index: owner, type, HP, position, production queues, decoded down to three 50-byte opaque suffixes on the CBA anchor. `--objects` enrichment lets event rows say "P6's gate," not "id 58914."

**The frontier system (rare anywhere):** `coverage` is a byte-accounting ledger — decoded + opaque must sum to file size, so parser gaps *can't hide*. `opaque-spans`/`frontier`/`opaque-clusters` profile and bucket what's still dark across a corpus. `diff-state` + `scan-value` are the controlled-experiment bridge: plant a known value in a fixture, find exactly which region it lands in. This is how words 1–10 got named.

**Deep game-data decode:** `effective-data`/`effective-units`/`datamod-check` — reads the per-player effective game-data table embedded in replays, decodes the 864-slot unit-availability array, cross-references to DAT slots, and can verdict "was this game modded" against a baseline. `ai-manifest` extracts AI module references. `postgame`/`postgame-corpus` — proved DE does *not* serialize scoreboard stats (metadata only), which is why the sidecar road matters.

**Acquisition:** `fetch` pulls replays from Microsoft's public endpoint by gameId/profileId, with `--all-povs` roster-walking — every POV of a game, cached politely.

**Batch:** `corpus`, `inbox` (playtest intake: dedupe, issue cards, per-replay story lines), `issues`, `telemetry` (schema-validated markers), `unknowns` (corpus-wide undecoded-action mining), `story` (designer-facing brief with a `claims` ledger *and a `missing` list* — the tool declares what it hasn't earned).

## 3. XS & the sidecar loop (`kit xs`, `xsdat`, `pkg/xs`)

Authoring generators: `bridge` (named wrappers over anonymous trigger variables, with read/write lint), `shims` (scan XS modules → emit script_call trigger recipes), `datagen` (JSON → GoKu-style xsArray init modules). Then the crown jewel: `xsdat decode --ledger` — typed decode of `.xsdat` sidecar files with assertion-fixture semantics: declare expected rows, run the game unattended, get pass/fail. `replay sidecar-sync` closes the loop by correlating sidecar ground truth against the replay's checksum matrix. `replay carrier` generalizes authored self-reporting probes. This is the "assert engine state without playing" machine.

## 4. DAT engine (`kit dat`, `pkg/datfile` + `pkg/datcodec`)

A span/index patch engine, deliberately not a genieutils port — unknown bytes preserved, everything verified by re-index + readback + neighbor canaries. **Read:** units, techs, effects (with semantic overlays and `effect-explain`/`tech-explain` authoring prose for named command types, packed attack/armor deltas, local-building effects, and rename-unit string IDs), ability joins (Tech+Effect as one authoring concept), civs, graphics (incl. particles), sounds, terrains, tech-tree, unit-headers, `refs` (who points at this ID, with command-shape candidates split from arbitrary possible operands), `availability` (per-civ trainability cards), `diff` (decoded-structure comparison of vanilla vs mod), `roundtrip` (codec safety gate). **Write:** patch units (HP, graphics, attacks, armor, costs, tasks, train buttons...), patch/create graphics, create units/effects/techs/sounds, create paired ability records (`create_ability` = new effect + new tech wired together), tech-tree connection edits, semantic deletes with reference ledgers, and `delete-plan` — which answers "what does delete safely *mean*" before anything mutates, including paired ability delete plans that refuse non-tail or still-referenced tech/effect pairs. `semantics-pack`/`readback` generate and verify effect-command semantic documentation against live files, with a focused local-building feature pack for the 200/201/202/204 command family.

## 5. Graphics & FX (`kit gfx`, `kit fx`)

`gfx info/export` — SLD sprite container parse and PNG frame export. `fx new/bind/lint` — the proven custom-particle pipeline: descriptor JSON + DDS atlas + metadata triangle, DAT binding, with the engine-validated ImageMagick invocation pinned. This is the fireball pipeline, productized.

## 6. AI files, mods, packaging (`kit ai`, `kit mod`, kit core)

`ai fingerprint/lint/diff` — content-integrity hashing and sanity checks of `.ai/.per` files, with a registry keyed to replay-side AI signatures. `mod check` is an *activation* doctor — catches the silent "local mod exists but the Definitive Set is actually running" failure via mod-status.json. `inventory/manifest/verify/pack/portable-check` — the spore pattern: self-verifying, self-packaging, local-path-leak detection, ships its own SHA.

## 7. Verification harnesses — the loop-closers

`verify-run --contract` (replay vs pre-registered expectations: identity, data-set, telemetry, chat markers, result — render claims stay honest "unknown" until a human oracle exists). `ci init/check` (per-project post-playtest gate). `release-check` (identity + lint + diff + modcheck + pack in one daily gate). `campaign generate` (engine-verified cross-scenario `.xsdat` persistence modules — writer/reader/manifest). `facts list/check` — the evidence ledger itself: every engine fact with tier, date, and fixture citation.

## 8. Network & registries

`player stats` (Microsoft career-stats API, honestly labeled "career aggregate, not per-match truth"), `replay fetch` (above), `identify --register` + `collect` (folder-scale fingerprint-and-register), `known_scenarios.json`/`known_ais.json` as shipping wisdom banks where the description field is deliberately load-bearing: a hash match becomes *orientation text for a fresh AI*.

## 9. The CBA layer (`kit cba`, `pkg/cba`) — quarantined interpretation

Everything CBA-specific lives here and consumes generic facts, never redefines them: `sidechannels` (V292 accounting decode + canonical register map), `phase-facts` (per-civ Castle/Imperial thresholds + raze requirements, DAT-joined), `replay phases` (the new phase detector: exact anchors + checksum shadows), `raze-pressure`, `trigger-razes`, `trigger-spawns` (per-civ spawn wave counts/seconds), `replay progression/razes/perf/doctrine` (chrae's matchup doctrine as an overlay), `balance` (changelog-mined parameter table that refuses to guess missing civs). Underneath: the full **ladder system** — corpus (JSONL store), multi-axis contribution-share scoring with chrae's anti-stacking tenet in the comments, the API-race archiver, RAM-bounded worker pools, SQL export for the ladder site, and the V292 identity rule (filename token + trigger band, because graph hashes split across host copies).

## 10. Odds and ends

`swatch patterns` (map-tile geometry generators for Sandbox work), `scen regions` (starter region-context files), replay context JSON (teach the Kit *your* map's region names without polluting generic truth), resource guardrails (1GiB heap soft limit, 32MiB scenario inflate refusal — the huge-community-scenario protection).

---

**The one-sentence version:** we have a dependency-free Go instrument that can open, verify, diff, patch, and *generate* every major AoE2DE file format; prove which bytes of a replay it does and doesn't understand; run controlled experiments against the engine and bank the results as cited facts; assert live engine state through the sidecar loop without a human watching; and interpret CBA on top of all that without ever contaminating the generic layer — with honesty labels load-bearing at every seam.

The known frontiers, equally honestly: full object-body field decode (three 50-byte suffixes left on the anchor), deterministic v68 trigger-section walking (some records still identify by terrain fallback), word_9, per-kill attribution (structurally absent from replays — hence anchors + sidecars), and render truth (no screenshot oracle yet).
