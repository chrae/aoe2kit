# Notices and Credits

AoE2Kit is licensed LGPL-3.0-only.

This notice is generated from `data/credits.json`, the credit source of truth. If a source, author, project, or community should be credited and is missing, that is a bug.

## Credit Principle

Attribution is abundant and automatic. Personal thanks is scarce, deliberate, and human. The ledger keeps unpaid gratitude visible instead of pretending it has been paid.

## References

### Advanced Genie Editor (AGE)

- Authors: Tapsa and contributors
- URL: https://github.com/Tapsa/AGE
- License: GPL-3.0
- Permission: GPL-3.0
- Gave: A long-standing open-source Genie DAT editor and practical oracle for DAT editing behavior, command terminology, and field naming.
- Role: DAT editor oracle and terminology reference
- Relationship: reference_implementation_and_behavior_oracle
- Rung earned: credited
- Thanks status: owed
- Give-back status: prepared_by_lgpl_release_and_notice
- Correction state: open_to_correction

AoE2Kit's DAT authoring and semantic naming work was informed by AGE's long-standing Genie DAT editor behavior, command names, and field terminology. AGE is credited deliberately as a reference and behavior oracle.

Notes:
- AoE2Kit does not vendor AGE source code.
- AGE remains the community reference editor for Genie DAT files.
- Where AoE2Kit uses AGE-shaped terminology, behavior is still verified through AoE2Kit roundtrip/readback tests and engine evidence when available.
- Personal thanks to Tapsa is a human action and remains owed until it is actually sent.

### genieutils

- Authors: Tapsa and contributors
- URL: https://github.com/Tapsa/genieutils
- License: LGPL-3.0
- Permission: LGPL-3.0
- Gave: Open-source Genie file library/tooling and a cautionary reference for codec scope, API shape, and file-corruption risk.
- Role: Genie file tooling reference
- Relationship: format_reference
- Rung earned: credited
- Thanks status: owed
- Correction state: open_to_correction

AoE2Kit's DAT codec work was designed in conversation with the shape and risks of existing Genie tooling, including genieutils, while choosing a Go-native codec and verification-first architecture.

Notes:
- AoE2Kit does not vendor genieutils source code.
- The codec architecture favors explicit spans, roundtrip gates, and narrow verified writes.

### aoc-mgz

- Authors: happyleavesaoc and contributors
- URL: https://github.com/happyleavesaoc/aoc-mgz
- License: MIT
- Permission: MIT
- Gave: A mature recorded-game parser ecosystem and reference point for AoE2 replay parsing.
- Role: Recorded-game parser reference
- Relationship: format_reference
- Rung earned: credited
- Thanks status: owed
- Correction state: open_to_correction

AoE2Kit's replay work benefits from the aoc-mgz ecosystem and keeps replay claims explicit about what is parsed, inferred, or still opaque.

### AoE2DE UGC Guide

- URL: https://ugc.aoe2.rocks/
- Permission: public documentation reference
- Gave: Public documentation for AoE2DE scenario, trigger, XS, and modding concepts.
- Role: Scenario, trigger, XS, and data-mod documentation reference
- Relationship: documentation_reference
- Rung earned: credited
- Thanks status: owed
- Correction state: open_to_correction

AoE2Kit uses the UGC Guide as a public reference point for scenario, trigger, XS, and modding concepts, then validates critical behavior against real files and engine runs.

### CB Front Towers OG Enhanced

- Authors: SpiRaL
- URL: https://steamcommunity.com/id/theogspiral/
- License: All rights reserved by the author
- Permission: studied for technique only; the author forbids modifying or redistributing the map without consent
- Gave: A masterclass in scenario-only UGC technique: the reroll shop (train-to-consume renamed button units), object and technology renaming with no data mod, the repeatable-technology purchase trick (disable-then-enable reset), and a live Objective-Panel stats display.
- Role: Scenario-design technique oracle and inspiration
- Relationship: behavior_oracle_and_inspiration
- Rung earned: credited
- Thanks status: owed
- Correction state: open_to_correction

AoE2Kit's scenario authoring — its object/technology rename operations, repeatable-technology decode, and shop idioms — was materially informed by studying SpiRaL's CB Front Towers, which has been a direct inspiration. Front Towers served as a behavior oracle; every technique was re-implemented independently and verified against real files.

Notes:
- AoE2Kit does not vendor, modify, or redistribute CB Front Towers content. The map and its XS were studied for technique only, in keeping with the author's license.
- The reroll shop, button/technology rename, repeatable-technology reset, and live-panel idioms were reverse-engineered as a behavior oracle and re-implemented from scratch.
- Personal thanks to SpiRaL is a human action and remains owed until it is actually sent.

