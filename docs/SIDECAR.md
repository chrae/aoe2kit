## What problem it solves

AoE2DE is a black box at runtime. A recording gives you player *commands* — move, build, attack — plus a periodic sync heartbeat. It does **not** give you trigger state, XS variables, custom resources, or anything the scenario itself computed. We proved this repeatedly: trigger-generated chat never enters a recording, and the postgame block is metadata only, not serialized stats.

So a scenario author debugging a 3,000-trigger machine has historically had two bad options: print state as in-game chat and read it off the screen, or run an "observer AI" in a player slot that harvests state and spits it out as invisible chat (SpiRaL's actual workaround). Both mean *a human watches a game to learn anything*, and both give you strings to parse rather than values.

The sidecar removes the human from the loop. The scenario writes its own state to a file on disk, in typed binary, timestamped, at moments you choose.

## How it works, mechanically

**The write side.** XS exposes a tiny file API — `xsCreateFile(append_bool)`, `xsWriteString/xsWriteInt/xsWriteFloat/xsWriteVector`, `xsCloseFile`, `xsOpenFile`. You call these from functions that triggers invoke via `script_call`. There's no path argument: DE decides where the file goes, naming it after the scenario stem — `<DE profile dir>\<scenario name>.xsdat`.

**The file format.** This is the part worth knowing precisely, because it explains every design decision downstream. The container has **no type tags whatsoever**. It is a bare concatenation of values in write order:

- `string` → u32 little-endian byte length, then the raw bytes
- `int` / `uint` → 4 bytes little-endian
- `float` → 4 bytes IEEE-754 little-endian
- `vector` → three consecutive floats, 12 bytes

That's the entire spec. A `.xsdat` is a headerless typed stream where **the writer's call order *is* the schema**. Nothing in the file tells you whether the next four bytes are `1092616192` or `10.0` — those are the same bytes.

**The read side** (`kit xsdat decode`) therefore runs in one of two modes. Given the write order — via `--types string,int,float,...`, a `--ledger`, or a named `--schema` — it walks the stream deterministically and emits typed rows. Without one, it falls back to a labeled heuristic: try a plausible length-prefixed printable run as a string, otherwise take 4 bytes and report both the int and float interpretation as candidates, flagged `heuristic: true`. It never silently guesses.

There's one quiet invariant that does a lot of work: the decode is only `OK` if there were no errors **and** `RemainingBytes == 0`. The stream must be consumed exactly. A schema that's wrong by one field leaves a tail, and the tail is the alarm — you can't misread a sidecar and get a clean-looking result.

**The assertion layer.** With `--ledger expected.json`, you declare what a correct run must produce. The decoder compares position by position and emits per-row `passed / failed / unknown` plus a summary, exiting non-zero on mismatch. That's what converts "I dumped some state" into a test: you pre-register the expectation, run the game unattended, and read a verdict.

**The correlation layer** (`kit replay sidecar-sync`) is the piece that makes it a *calibration instrument* rather than just a log. It decodes the sidecar's phase rows, then for each phase finds the first DE checksum sample at-or-after that timestamp for the chosen player, within a window (default 20s), and compares the sidecar's known-true values against the replay's checksum words — stockpile, unit-type sum, object count. When no sample lands in the window it says `no_checksum_sample_in_phase_window` rather than inventing a match. So you get: *here is what was actually true inside the engine at time T, and here is what the replay's opaque word said at time T.* That's how you name an unknown field.

**The file semantics we had to learn by failing.** These aren't in any doc; each cost a run:

- Rows are **buffered and flush at scenario unload** — the file doesn't exist mid-game, so there's no live streaming.
- DE holds an **exclusive lock** while the game runs; harvesting needs a share-mode-tolerant read or a wait until exit.
- `xsCreateFile(false)` **truncates**, `(true)` appends — so header-once logic matters across reruns.
- A run that ends before writing its footer leaves a **valid partial file**, which is why the schemas accept partial runs.
- Reading with the *bare stem* from a different scenario works, in a later session — that's real cross-scenario persistence.
- **MP law:** file I/O is per-client. Every client runs the same triggers, so same-party writes mirror identically and are sync-safe; but any file you *read* must be byte-identical on every client or you desync.

## What it bought us

The headline is that **it named the checksum words**. The v3 probe wrote food values `900000`, `900100`, `900200` from XS; op=2 word 1 carried them exactly — that's how word_1 became "resource stockpile" instead of a hypothesis. The v7 trigger-only calibration got word_3 to "+2 per created object" and word_4 to "a Monk create adds +100, other creates add 0." Words 2, 6, 8, 10 fell to the same method. Every decoded word in that matrix traces back to a controlled fixture with known ground truth.

The **negative results were worth more than the positives**, and only the sidecar could produce them. In the v15 castle-kill ground truth run, the sidecar showed P1's credited kills moving 0→1→2→3 in real time — and *no* checksum word mirrored it. That single fixture is why we stopped hunting for a kill packet in the sync stream and pivoted to the anchors-and-inference approach that produced the phase detector. Same with `xsChatData` not serializing, and attr 220 not moving word_1: cheap, decisive eliminations.

It also **proved cross-scenario persistence** (the v9 writer/reader pair), which is the engine capability behind `kit campaign generate` — multi-map campaigns with carried state, which SpiRaL independently named as his design dream.

And it changed the *shape* of the work. Because assertions run unattended, an engine question stopped costing "chrae plays a game and watches" and started costing "generate fixture, run, decode." chrae's manual probe sessions — the Kill test, Dawn Accord, the fog and score probes — got their leverage from having a machine-checkable record on the other side.

## What it could do for others

**SpiRaL, directly.** He stated his number-one pain as asserting engine state without playing, and his feed/AFK detection want as "data and timestamp." That's the sidecar's literal signature: when your detector fires, write `player_id`, `counter`, `xsGetGameTime()`. He currently burns a player slot on an observer AI and parses invisible chat — the sidecar removes the slot, removes the parsing, gives typed values, and works in games nobody watched. It's not a better version of his workaround; it deletes the workaround.

**GoKu.** Roughly 670KB of multi-file XS running zones, difficulty scaling, and targeting logic with no runtime visibility. A sidecar journal is a drop-in `include` — no restructuring, just calls at the decision points he already has.

**Any scenario author.** It's printf-debugging plus unit tests for a game engine that offers neither. Concretely: regression-test a scenario across versions (did the V293 changes alter spawn timing?), measure balance from real runs instead of feel, verify a mod actually loaded, instrument a quest to see which branch players actually take, or carry state between maps in a campaign. `kit ci init` scaffolds exactly this — a project-local contract plus a tiny `a2k_debug.xs` include, then `kit ci check` after a playtest.

**The honest limit, which matters for how it's pitched:** it only works on scenarios *you instrumented*. It tells you nothing about someone else's uninstrumented game — you can't sidecar a random ladder replay. That's precisely why the two halves complement each other: the sidecar gives ground truth on fixtures, and that ground truth calibrates what can be *inferred* from ordinary recordings of games nobody prepared. The instrument teaches you to read the wild.
