# How To Use The Felt Compendium

The compendium is meant to support scenario design, especially AI-assisted design.

The correct workflow is:

1. Choose the emotional scene target.
2. Select terrain first.
3. Select objects that belong on that terrain.
4. Check footprint and scale.
5. Choose rotations/variants intentionally.
6. Compile a recipe.
7. Verify in AoE2DE.

## Selection Pattern

Given a prompt like:

> a quiet ruined shrine in early autumn, beside shallow water

Search the data for:

- terrain: `autumn`, `dry-grass`, `forest floor`, `shallow water`, `beach vegetation`
- object material: `stonery`, `greenery`, `water`
- object function: `feature`, `marker`, `border`, `eyecandy`
- vibe tags: `ruin`, `shrine`, `autumn`, `still`, `sacred`, `lived-in`

Then assemble a candidate set:

- one or two landmarks
- a small number of trees
- ground decals or underbrush
- shoreline details only on water/shallow terrain
- negative space

Do not fill every empty tile. The compendium helps restraint as much as abundance.

## Terrain Comes First

Terrain is the substrate of the scene. If a scene has water, snow, road, beach, mangrove, or ruins,
paint the terrain before placing objects.

Object placement can depend on terrain:

- reeds and lilies need shallow water or wet shoreline
- mangroves need mangrove/wet terrain
- fish traps and bridges need water context
- snow trees read best on snow or snow-forest floor
- ruins read differently on grass, desert, jungle, or snow

## Rotation And Variant Semantics

Rotation is not always facing.

For many decorative objects, the "rotation" field is really a variant selector:

- seasonal trees can use rotations as life/season states
- statues and modular sets can use rotations as distinct pieces
- animated or multi-frame sprites may have complex semantics

Always check:

- `variant_semantics`
- `visual_variants.json`
- contact sheets

## AI Authoring Rules

When an AI uses this compendium, it should:

- cite the object ids and terrain ids it selected
- cite the intended mood or function for each group
- preserve confidence labels
- say when a placement is a guess
- avoid scenario-specific assumptions unless the project bundle supplies them
- generate a recipe that can be inspected and patched with Kit

## Example Output Shape

A scene compiler should produce something like:

```json
{
  "map": [
    {"op": "set_terrain_circle", "terrain_id": 1, "x1": 64, "y1": 64, "radius": 16}
  ],
  "units": [
    {"op": "add_unit", "player": 0, "unit_const": 2294, "x": 63.5, "y": 62.5}
  ]
}
```

The exact schema is Kit's scenario recipe schema. This document describes how to choose the content.

## Engine Verification

Generated scenes should be treated as:

`structure_verified_not_engine_verified`

until they are opened in the DE editor or run in-game. Contact sheets are strong evidence for visual
selection, but AoE2DE is still the final oracle for load behavior, terrain blending, captions,
pathing, and rendering context.
