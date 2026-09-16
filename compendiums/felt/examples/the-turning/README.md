# The Turning

`The Turning` is the first end-to-end Felt Compendium compile proof.

It starts from a blank 120x120 grass scenario, paints a seasonal terrain journey, and places Gaia
decorations chosen from the compendium:

- alpine winter source
- boreal slope
- autumn woods
- still lake and island
- summer meadow
- tropical delta

Files:

- `the_turning.recipe.json` - Kit scenario patch recipe.
- `the_turning_compile_report.json` - compile summary and verification notes.
- `The Turning.aoe2scenario` - compact scenario proof artifact.

Verification at time of import:

- `kit scen check`: OK
- `kit scen lint --load-safety`: OK
- deployment pullback hash matched local file

The scenario remains `structure_verified_not_engine_verified` until opened in AoE2DE.
