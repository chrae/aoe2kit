# Changelog

## Geo-Trace & Felt — 2026-09-16

AoE2Kit can now trace real Earth into playable maps, and author terrain by aesthetic *intent*
instead of terrain IDs.

### Added
- **`kit geo trace-dem` — real-map tracing.** Aim a viewport (center + span, or a bbox) at any
  latitude/longitude and render real geography into a scenario: elevation, coastlines,
  land-cover (ESA WorldCover / GLC_FCS30D forest types), lakes and rivers, interstates, and
  political borders — translated through a **felt** palette and **climate-aware** from real data
  (warm sandy shores vs. cold rock; tropical vs. boreal species).
- **Felt compendiums.** Queryable indexes of what terrains, gaia art, species, and terrain
  *layerings* feel like — so you can ask for "a cold rocky shore" or "an Everglades gloom"
  instead of memorizing IDs. Includes contact sheets, terrain sheets, and sampled animations.
- **Organic terrain primitives.** Composable recipe ops — masks, `noise_fill`,
  `erode` / `semantic_erode`, `cluster_scatter`, `layered_crossfade` (top/bottom terrain
  blending), and a bulk `terrain_grid` op for fast full-map terrain + layer writes.

### Changed
- SLD sprite stats/export improvements, including compact first-frame headers.
- Documented the geo-trace CLI and the scenario authoring primitives.
- Credits are now split into **Built On** (studied prior art) and **Personal Thanks**
  (individual people who directly helped); `NOTICE.md` is generated from `data/credits.json`.
- Public packaging pass (`portable-check`, `pack --profile public`); the terrain scrape no
  longer bakes a local path.

### Known limits
- Full-resolution (480) raster land-cover sampling isn't container-friendly yet (multi-GB RSS);
  the container path uses pre-fetched vector data + `--omit-report-cells` / `--omit-report-recipe`.
- Geo-trace output is structure-verified; the in-engine look is confirmed by eye, not
  automatically.

### Thanks
**Sekiro** taught us **terrain layering** — the per-tile `layer` field, used as a top/bottom
crossfade — and it reshaped this entire release. Every graded coastline, every water-depth
shelf, every biome seam, the layered political borders, and the whole hybrid-terrain approach
are built on it. Sekiro is also AoE2Kit's first active external user. **Thank you.**

Thanks also to **SpiRaL** and **krmyth9**, fellow designers who have directly helped along the
way. Full attribution and the standing gratitude ledger live in `NOTICE.md` / `data/credits.json`.
