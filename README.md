# AoE2Kit

Portable Go tools for AoE2DE scenario, replay, data-mod, local-mod, and sandbox
work. Public releases are source-first and LGPL-3.0-only. Binary-bearing archive
profiles exist for private/no-toolchain sandboxes, but the public profile ships
source and attribution without stale built artifacts.

AoE2Kit is intentionally project-neutral. It can live inside SDS, but shared
packages must not know about Decima, CBA, Scenario Sandbox, or local machine
paths. Project-specific behavior belongs in recipes/configs outside the shared
packages.

## What's Inside

- Source for one local CLI build: `go build -o kit ./cmd/kit`.
- `kit replay`: AoE2DE v68 replay inspection, including header/player roster
  parsing, data-set identity and content checksum, trigger-graph SHA-256
  fingerprints, command/story timelines, AI loadout references, and decode
  coverage reports.
- `kit scen`: scenario string/effect/trigger analysis, trigger-intelligence reports, plus patch/recipe helpers
  for controlled scenario edits.
- `kit dat`: data-file readers and patching helpers for graphics, units, and
  particle-effect bindings.
- `kit cba`: CBA Requiem-oriented analysis commands kept separate from the
  project-neutral replay/scenario core.
- `kit mod`, `kit inventory`, `kit portable-check`, `kit project`, `kit recipe`,
  and `kit pack`: local mod checks, bundle inventory, project triage/snapshots,
  reusable recipe templates, portability audit, and handoff archive creation.
- `kit credits`: a machine-readable and human-readable attribution ledger for
  upstream tools, references, and community work that shaped the kit.

## License and Credits

AoE2Kit is licensed LGPL-3.0-only. AGE (Advanced Genie Editor), by Tapsa and
contributors, is GPL-3.0 and has materially informed AoE2Kit's DAT authoring
semantics, field terminology, and editor-oracle expectations. That debt is
credited deliberately in [NOTICE.md](NOTICE.md) and `data/credits.json`.

Check the credit ledger with:

```sh
go build -o kit ./cmd/kit
./kit credits --text
```

## Known Limitations

- Replay verification is structural unless a report explicitly says
  engine-verified. A byte-correct parser can still be wrong about engine
  behavior until checked in-game.
- AoE2DE scenario replays do not expose ordinary kill/military score tables in
  the replay stream. Use `kit replay coverage`, `kit replay deaths`, and
  game-specific consumers such as `kit cba` to inspect the current boundary.
- Lobby settings and some map/setup blocks are not fully decoded yet. Use
  `kit replay coverage` to see known, bounded, and opaque regions.
- Golden fixture tests are optional. Set `AOE2KIT_FIXTURE_ROOT` to a directory
  containing the external replay/scenario/data fixtures to run them; otherwise
  they skip cleanly.
- PromiDE/Promisory AI module resolution is optional. Set
  `AOE2KIT_PROMISORY_ROOT` to a local Promisory module tree when you want
  `kit replay ai-loadout` to resolve default DE AI module names to files.
- Kit sets a default soft Go heap limit of 1 GiB unless `GOMEMLIMIT` or
  `AOE2KIT_GOMEMLIMIT_MB` is set. Scenario parsing also refuses inflated bodies
  above 32 MiB by default; intentionally large community teardown runs must set
  `AOE2KIT_MAX_SCENARIO_MB` higher or `AOE2KIT_ALLOW_HUGE_SCENARIO=1`.
- Building from source requires Go 1.23 or newer. The kit uses only the Go
  standard library.

## Reproducibility and Release Packing

Build locally from source:

```sh
go build -o kit ./cmd/kit
GOMAXPROCS=2 go test -p 2 ./... -count=1
./kit docs lint
./kit portable-check . --profile public
```

Start with [START_HERE.md](START_HERE.md).

To produce a public source archive after changes:

```sh
./kit pack AoE2Kit-public.zip --profile public --sha-sidecar --update-manifest
```

For a private sandbox that can execute Linux binaries but cannot compile Go,
build `./kit` and use `--profile sandbox`. Do not publish stale binaries by
accident; the `public` profile excludes them.
