# AoE2Kit Felt Compendium

The Felt Compendium is a machine-readable visual vocabulary for AoE2DE scenario authorship.

It catalogs decorative Gaia objects and terrain by what they *feel like* in a scene: material, mood,
scale, footprint, biome, visual role, and placement behavior. The goal is to let a human or an AI
choose concrete AoE2 assets by design intent instead of guessing from internal unit names.

Example design request:

> Build a quiet autumn shrine beside shallow water.

The compendium gives an authoring agent enough grounded data to choose:

- terrain ids for forest floor, dry grass, beach vegetation, shallow water, and depth gradients
- unit constants for autumn oaks, reeds, lilies, sacred trees, stones, statues, and props
- rotation/frame choices for seasonal trees and visual variants
- pixel dimensions and tile footprints so the scene is not overcrowded
- rendered contact sheets so the choice can be inspected

## What Is Included

- `data/art-felt-compendium.json` - Gaia/decorative object felt index.
- `data/terrain-felt-compendium.json` - terrain felt index.
- `data/px-dims.json` - measured rendered pixel bounds plus tile-footprint measurements.
- `data/visual_variants.json` - variant and rotation notes.
- `data/gaia_palette.json` - broad Gaia artwork palette data.
- `data/manifests/` - render batch manifests and terrain scrape manifests.
- `contact-sheets/art/` - generated PNG contact sheets for decorative objects.
- `contact-sheets/sampled-animations/` - sampled frames for large animated pieces.
- `contact-sheets/terrain/` - generated terrain swatches.
- `scenes/` - higher-level scene specs authored against the compendium.
- `examples/the-turning/` - an end-to-end scene compile proof.

Raw `.sld` source files are intentionally not bundled here. The GitHub package ships generated
references and machine-readable indexes, not a copy of the game's art containers.

## Current Coverage

This package is a snapshot of the Mandala decorative base work as folded into AoE2Kit:

- 232 art entries in the object felt index.
- 210 object entries with measured pixel and tile-footprint data.
- 126 rendered terrain swatches.
- 219 art contact sheets.
- 9 sampled-animation contact sheets.
- 2 scene specs, including `the_turning.json`.

The compendium is useful today for AI-assisted scenario drafting, visual browsing, and recipe
compilation. It is not a claim that every AoE2DE art asset is cataloged.

## How To Use It

For humans:

1. Browse `contact-sheets/art/` or `contact-sheets/terrain/`.
2. Open `data/art-felt-compendium.json` or `data/terrain-felt-compendium.json`.
3. Search for mood, material, biome, function, or object name.
4. Use the returned `unit_const`, `terrain_id`, and placement notes in a Kit recipe.

For AI agents:

1. Treat `data/art-felt-compendium.json` and `data/terrain-felt-compendium.json` as the source of
   truth for selection.
2. Use contact sheets as visual evidence, not as the primary database.
3. Read `docs/FIELD_SCHEMA.md` before generating placements.
4. Read `docs/HOW_TO_USE.md` before compiling a scene.
5. Preserve confidence labels. Do not upgrade `ai_felt_from_render` to `engine_verified` without a
   real in-engine check.

## Example: The Turning

`examples/the-turning/` contains the first full compendium-to-scenario proof:

- a blank 120x120 base
- a painted terrain spine: alpine source -> river -> lake -> tropical delta
- seeded Gaia placement by region
- load-safety-checked scenario output

The generated scenario is included as a compact proof artifact, while the recipe and compile report
show how the scene was made.

## Verification Model

This package uses the same honesty model as the rest of AoE2Kit:

- `ai_felt_from_render`: an AI interpreted a rendered asset/contact sheet.
- `human_felt`: A human reviewer reacted in-game or from direct visual inspection.
- `engine_verified`: behavior was verified inside AoE2DE or by a replay/scenario oracle.
- `structure_verified_not_engine_verified`: Kit parsed, wrote, or checked the structure, but the DE
  engine remains the final oracle.

## Design Boundary

This is generic UGC tooling data. It is stored under AoE2Kit because it helps any scenario author
compose AoE2DE scenes. Mandala was the proving ground, not a hard dependency.
