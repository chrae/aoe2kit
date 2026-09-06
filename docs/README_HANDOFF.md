# AoE2Kit

A dependency-free Go toolkit for Age of Empires II: Definitive Edition files —
scenarios, replays, `.dat`, XS sidecars, sprites, particles, local mods. It reads
and writes UGC **without the game running**, and it was built so an AI agent can
drive it as easily as a person can.

It owes its existence to `aoe2scenario-parser` and the community around it. This
is a Go rewrite with a different center of gravity: the scenario layer is
comparable, but the replay layer goes considerably further, and the whole thing
is built around one discipline described below.

---

## Start here

Requires Go 1.23+. No other dependencies, no network access, no CGO.

```sh
go build -o kit ./cmd/kit
./kit                      # every command, with flags
```

Four commands that show the range:

```sh
./kit scen glossary  <file.aoe2scenario> --text   # vocabulary of an unfamiliar scenario
./kit replay summary <file.aoe2record>   --text   # players, civs, teams, result, map
./kit replay chat    <file.aoe2record>   --text   # full transcript, lobby + in-game
./kit xsdat decode   <file.xsdat>        --text   # typed rows from an XS sidecar
```

Output is JSON by default; `--text` renders it for humans. Agents should parse
the JSON and ignore `--text`.

---

## The documents

| file | what it answers |
| --- | --- |
| `README.md` | why the kit is shaped this way (you are here) |
| `CAPABILITIES.md` | what it can do, domain by domain, with the boundaries stated |
| `API_REFERENCE.md` | every command: flags, required inputs, JSON keys it returns |
| `api_reference.json` | the same reference, machine-readable |
| `api_schemas.json` | the full nested shape of all 678 report types |
| `PACKAGE_API.md` | `go doc` for every package, for importing the library directly |
| `SIDECAR.md` | the XS sidecar debug loop — assert engine state without playing |

If you only read two: `CAPABILITIES.md` to see whether the kit does what you
need, then `SIDECAR.md`, which is the technique most likely to be useful in your
own scenarios regardless of whether you ever use this toolkit.

---

## The one idea that runs through everything

**Every report says how strongly it is claimed.**

The engine is a black box. A parser can be byte-perfect and still be wrong about
what the bytes *mean*, and a tool that reports both kinds of statement in the
same voice will eventually mislead you at the worst moment. So every report
carries an explicit label:

- `structure_verified` — the parser round-tripped it, the readback matched, the
  invariants held. The file really does say this.
- `engine_verified` — a live game or the editor was the oracle. The engine
  really does behave this way.
- `behavior_inferred_*`, `*_candidate`, `*_not_engine_verified` — a hypothesis
  with evidence, deliberately not dressed up as fact.

You will see this everywhere, including in places where a less careful tool
would simply assert. `kit cba replay razes` labels its output
`behavior_inferred_raze_candidate_not_engine_raze_event`, because no raze packet
has ever been decoded — what it actually found was pressure correlated with a
later payoff. `kit replay story` ships a `missing` list naming the capabilities
the kit has *not* earned. `kit facts list` exposes the whole evidence ledger:
every engine fact with its tier, the date it was verified, and the fixture that
proves it.

The practical rule when consuming this kit, especially programmatically:
**never promote a claim past its label.** If a report says candidate, it is a
candidate.

---

## Design decisions worth knowing

**Zero dependencies, standard library only.** One `go build` produces one binary
with no toolchain archaeology. It also means the kit runs in a locked-down AI
sandbox, which was a design target rather than an accident.

**The `.dat` engine patches spans, it does not model the world.** Porting
`genieutils` into an eager object graph would mean understanding every field
before touching any of them. Instead the kit indexes records, patches the fields
it genuinely knows, splices whole records when a length changes, preserves every
unknown byte untouched, and verifies by re-inflating and re-indexing to EOF.
This is why it can safely edit a modern DE `.dat` that nobody has fully mapped.

**Writes never touch the input.** Every patch path writes a new file, reopens
it, rebuilds the decompressed body, and checks invariants and neighbor records
before reporting success.

**Replay decoding is byte-accounted.** `kit replay coverage` reports every
region of the file as `decoded` or `bounded_opaque`, and the two must sum back
to the file size. A parser gap cannot hide as silence — it shows up as opaque
bytes. `opaque-spans`, `frontier`, and `opaque-clusters` then profile what is
still dark so the next decoding effort is chosen from evidence rather than
intuition.

**Generic truth and domain interpretation are separated on purpose.**
`pkg/replay` and `pkg/scenario` know about file formats. `pkg/cba` knows about
one game mode and consumes generic facts without ever redefining them. If you
fork this for your own scenario, that is the seam to copy: your interpretation
layer should be able to be wrong without contaminating the parser.

**The API reference is generated, not written**, by two methods that check each
other. A *static* pass parses the source and emits the complete nested shape of
every report type from its struct tags — that is `api_schemas.json`, it takes
seconds, and it executes nothing. A *runtime* pass then executes each read-only
command against real fixtures and records the keys it genuinely returned, which
proves the command works and names which type it emits. The static pass supplies
depth; the runtime pass supplies proof. Commands that could not be executed are
labeled with the reason rather than described from assumption.

The runtime pass is slow on purpose — it really does parse a 2 MB replay ~70
times — so it is a release-time cost, and `--reuse-probe` regenerates the
documents in about two seconds when only the docs changed.

---

## Working with an AI agent

The kit is deliberately shaped for a model in a container:

- **Structured output everywhere.** JSON first, complete objects, no scraping.
- **It describes itself.** `./kit` lists everything; `cmd/apiref` regenerates the
  full reference for *your* build.
- **Orientation is carried in data.** `known_scenarios.json` and
  `known_ais.json` map fingerprints to descriptions, so a hash match becomes
  usable context for a model that has never seen the file.
- **Honest failure.** Unknown action ids are preserved raw with sample payloads
  rather than dropped, so the agent can see what it does not understand.

A workflow that works well: run `scen glossary` before editing an unfamiliar
scenario, `replay coverage` before trusting a replay claim, and `facts list`
before asserting anything about engine behavior.

---

## What it cannot do

Stated plainly, because the frontier moves and a stale claim is worse than none:

- **No per-kill attribution from ordinary replays.** The command stream does not
  carry death or kill-credit events; controlled fixtures proved it directly. The
  kit infers combat pressure from object-count movement and says so.
- **Object bodies are only partly field-decoded.** Initial object state is
  readable; the tail bytes of some record classes remain opaque.
- **Scenario writing targets DE 1.57 and 1.58.** Older versions (1.55–1.56)
  read fine but are read-only. Map resizing and arbitrary existing-unit field edits are not
  exposed.
- **No render oracle.** The kit can prove a particle descriptor, atlas, and
  `.dat` binding are internally consistent. It cannot tell you the effect looked
  right on screen. Any report touching visuals says `engine_verified: false`.
- **Postgame stats are not in replays.** Verified across a corpus including
  ranked games: the postgame block is metadata only. Scoreboard numbers are
  recomputed for display, not serialized.

---

## Layout

```
cmd/kit        the CLI (scen, replay, dat, cba, xs, xsdat, fx, gfx, ai, mod, ci, facts...)
cmd/apiref     generates API_REFERENCE.md from the built binary
cmd/scen|dat|mod|swatch|cba   focused entry points for single domains

pkg/scenario   .aoe2scenario parser + writer + lint + XS deploy checks
pkg/replay     .aoe2record parsing: header, actions, sync matrix, objects, coverage
pkg/datfile    .dat span/index inspection and patch engine
pkg/datcodec   typed .dat section codec (effects, techs, civs, sounds, terrains)
pkg/xs         XS authoring generators and the .xsdat sidecar codec
pkg/cba        Castle Blood Automatic interpretation (domain layer, not generic truth)
pkg/kit        bundle inventory, verification, packaging, portability audit
pkg/gfx|fx|geom|aifile|modpack|campaign|ci|enginefacts   supporting domains

data/engine_facts.json   the evidence ledger behind every engine claim
```

---

## Regenerating the reference

After any change to the CLI:

```sh
go build -o kit ./cmd/kit
go run ./cmd/apiref --kit ./kit \
  --scenario <file.aoe2scenario> \
  --replay   <file.aoe2record> \
  --dat      <empires2_x2_p1.dat> \
  --xsdat    <file.xsdat> \
  --out-md API_REFERENCE.md --out-json api_reference.json
```

Fixtures are optional; commands with no fixture available are labeled
`skipped_no_fixture` rather than guessed at. That run executes the tool ~70
times and takes several minutes.

When only the documents changed, reuse the previous run's results instead —
about two seconds, and it still refreshes the schemas and package API:

```sh
go run ./cmd/apiref --kit ./kit --reuse-probe api_reference.json \
  --out-md API_REFERENCE.md --out-json api_reference.json \
  --out-schemas api_schemas.json --out-godoc PACKAGE_API.md
```

`--probe=false` skips execution entirely for a usage-only reference.

---

## Provenance

Built by **chrae** with AI assistance, as working infrastructure for real
scenario and replay analysis — not as a demo. It is shared because the people
who would find it useful are few and know each other.

Use it, fork it, take just the sidecar technique and ignore the rest. If you
find a decoding error, that is the most valuable thing you can send back: the
kit is only as honest as its labels, and a mislabeled claim is a bug of the
worst kind.
