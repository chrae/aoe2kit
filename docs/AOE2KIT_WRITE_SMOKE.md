# AoE2Kit Write Smoke

Purpose: flip AoE2Kit scenario writes from structurally verified to editor/game
verified with one cheap visual test.

This recipe expects an input scenario with map coordinates through at least
`(147,177)`.

## Build The Smoke Output

From `SDS/AoE2Kit`, either apply the built-in smoke directly:

```sh
./kit scen smoke \
  path/to/input.aoe2scenario \
  /tmp/AoE2Kit_WriteSmoke.aoe2scenario \
  --x 145 --y 175
```

Or generate/apply the canonical recipe explicitly:

```sh
./kit scen smoke-recipe --x 145 --y 175 > docs/SCEN_WRITE_SMOKE_RECIPE.json

./kit scen patch \
  path/to/input.aoe2scenario \
  /tmp/AoE2Kit_WriteSmoke.aoe2scenario \
  --recipe docs/SCEN_WRITE_SMOKE_RECIPE.json

./kit scen verify /tmp/AoE2Kit_WriteSmoke.aoe2scenario
```

Open `/tmp/AoE2Kit_WriteSmoke.aoe2scenario` in the AoE2DE editor.

## Expected Visuals

- Terrain write: a 3x3 patch at tiles `(145,175)` through `(147,177)` is terrain
  id `47`, elevation `0`, layer `-1`.
- Unit write: player 1 has one added unit `83` at `(146.5,176.5)`.
- Trigger write: trigger list contains three disabled `AOE2KIT WRITE SMOKE`
  triggers covering display/chat/timer, create/task/kill/remove, and
  activate/deactivate trigger-control primitives.

## Status Rule

Until this output opens in the editor/game and the three expected visuals are
confirmed, scenario writing remains `[D]`: Go/APar-verified, not game-verified.
After confirmation, record the exact AoE2DE build/date and promote the covered
write primitives to `[V]`.
