# Editor-authored reference scenarios (ground truth)

A scenario author hand-built these in the AoE2DE **editor** (all load in-game — they are known-good),
specifically so the kit can learn from them. The **filenames are lab notes** and the
file creation times give the chronology. Pulled from a local profile scenario
folder under `resources/_common/scenario/`.

These exist because **kit-authored scenarios (kit scen blank + kit scen patch add_unit)
crash AoE2DE on load** while the editor's equivalents do not. Fault (parsed from the
minidumps): `ACCESS_VIOLATION` — a null-pointer + field-offset deref inside `AoE2DE_s`
during `ScreenLoadGameTransition` (READ @0x3B5 for a placed statue; WRITE @0x0 for a
placed archer). i.e. the loader looks something up for a placed unit, gets NULL, derefs.
So the divergence between a kit blank and an editor blank is load-bearing.

## The scenarios (chronological)

| file (lab note) | what it establishes |
|---|---|
| Create scenario, do nothing, click save | **Minimal valid editor blank.** 763 B, 0 units, player_count 2, 120x120, terrain 0. THE structural target for `kit scen blank`. |
| Clicked regenerate terrain, it defaults to beach terrain | Editor "regenerate terrain" defaults to **beach**, not grass. |
| Regenerate with Grass 1 as terrain seed | grass-seeded regen |
| Generate seed, grass 1, medium 4 player size | **medium = 4 players** |
| Generate seed, grass 1, normal 6 player size | **normal = 6 players** |
| same, large 8 player | **large = 8 players** |
| Same, size name is huge, doesnt say number of players its for | huge size; UI does not state player count |
| Ludicrous | ludicrous size |
| I added the same object, 2 rotations | **Object-placement ground truth:** unit_const 2411 (Rock Limestone Hover, type 10, class 14) placed twice at rotation **0** and **3**, on tile centers (x.5,y.5). Plus 13763 auto-grass (1358). |

## Ground-truth findings the kit should absorb

1. **Rotation encoding is TYPE-DEPENDENT.** For **type-10 eyecandy**, the scenario `rotation`
   field is a **small INTEGER** (grass 1358 uses 0-7; rock 2411 uses 0 and 3), NOT radians.
   (Contrast: combat/type-70 units in Mandala-style placements store `rotation` in RADIANS, 0..2pi.)
   The kit palette/placement semantics should encode this per-type. My swatchgen writing
   integer 0..38 (statue angle_count) AND my "radian fix" both crashed — the real convention
   is small integers, and likely bounded (<=7 for the 8-orientation eyecandy).

2. **`kit scen blank` player init DIVERGES from the editor blank** (structural diff of the
   do-nothing editor blank vs kit blank --size 120 --gaia-active):
   - editor: player_0.human=true, player_2.active=false, players 2-4 human=true
   - kit:    player_0.human=false, player_2.active=true, players 2-4 human=false
   This is a prime suspect for the load crash. **Make `kit scen blank` produce a
   byte-faithful editor-equivalent blank** (match player active/human flags exactly).

3. **Map-size -> default player count** (editor UI): medium=4, normal=6, large=8
   (huge/ludicrous unspecified). Complements the discrete size-preset table.

4. **Editor `reference_id` is 0-based**; kit recipe/patch starts at 1. Verify this is benign.

5. **Eyecandy placed at tile centers (x.5, y.5).**

## Suggested actions
- Add these as **golden fixtures**; assert kit-authored blank == editor blank structurally.
- Fix `kit scen blank` player init to match the editor.
- Encode type-dependent rotation semantics in the palette/placement layer.
- A `kit scen patch` "editor-parity" check: a kit-authored scenario should be as loadable as
  an editor one. (I'm isolating the exact trigger in-game now and will send the confirmed
  root cause + the minidumps.)
