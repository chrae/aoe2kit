# Felt Compendium Provenance

The Felt Compendium was built from AoE2DE reference data and rendered with AoE2Kit tooling.

## Included

- Machine-readable felt indexes.
- Rendered contact sheets.
- Rendered terrain swatches.
- Measurement data.
- Scene specs and example compile artifacts.

## Excluded

- Raw `.sld` source art containers.
- Local machine paths from the original Mandala working directory.
- Scenario-specific assumptions that are not useful to generic AoE2 UGC authors.

## Evidence Sources

The compendium combines several evidence types:

- DAT-derived ids and dimensions.
- Kit-rendered contact sheets.
- Kit-rendered terrain swatches.
- Human review notes.
- Engine-verified findings where specific behavior was tested in AoE2DE.

## Licensing Note

The generated indexes, notes, and Kit-created metadata are part of AoE2Kit.

The rendered images are reference material for AoE2DE scenario authoring. They are included as
derived visual aids for modding and tooling workflows, not as a replacement for the game assets.
Redistribution posture should stay conservative: keep raw game art containers out of this package,
credit sources clearly, and remove or split the rendered reference bundle if a distribution channel
requires a lighter or stricter package.
