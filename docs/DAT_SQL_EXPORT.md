# DAT SQL Export

`kit dat sql <empires*.dat> <out.db>` writes a query-oriented SQLite view of
the decoded unit table.

This export is separate from the automatic SQLite cache used by `kit dat units`
and `kit dat unit`. Those commands reuse a cache keyed by the DAT's SHA-256;
a changed DAT selects a new cache. A decoder schema upgrade also rebuilds an
older cache once. Single-unit lookups decode the requested
record stored in SQLite, including `full_fields`, without inflating the DAT.
The cache directory can be selected with `AOE2KIT_DAT_CACHE_DIR`.
The SQL export records `source_sha256` in `metadata` so its source DAT can be
checked without inflating it. Rebuild the export when that fingerprint changes
or when adopting a decoder/export layout update.

The `units` table has one row per present civ/unit record. Stable scalar fields
decoded by the current DE parser are columns. Variable-length data is normalized
into `unit_attributes`, `unit_attacks`, `unit_armours`, `unit_damage_graphics`,
`unit_costs`, `unit_train_locations`, and `unit_fields`.
`attribute_map` records the UGC attribute id, the current descriptive name when
known, a candidate DAT field path, the confidence label, and an `exported` flag.
`exported=1` means the named path is available in this database (either in
`unit_fields` or the corresponding child table); it does not mean the semantic
meaning is engine-verified. A `layout_correlated` label is structural evidence
from the DAT layout, not an engine-verified claim.

Unit-reference fields use the DAT signed-sentinel convention: `-1` means that
the reference is absent. In particular, `annex_unit_*`, `stack_unit`,
`head_unit`, `transform_unit`, `pile_unit`, and `death_spawn` are signed
references. The raw DAT encoding for an absent 16-bit reference is `0xFFFF`;
the Go decoder and SQL export normalize that value to `-1`, so queries do not
need to special-case `65535`.

The exporter also writes `attribute_map_*` coverage keys to `metadata` and emits
a warning when mapped attributes point at fields the decoder does not export.
The coverage check recognizes canonical-name aliases and list-valued fields in
their child tables. Consumers should filter on `attribute_map.exported=1` and
treat `exported=0` as an explicit decode/export gap, not as evidence that no
unit carries the attribute.

The normal DAT index does not retain full field maps for every unit. `kit dat
unit` decodes one requested record on demand; the SQL exporter does the same one
record at a time while inserting rows. This keeps the normal parser compact and
avoids multiplying the in-memory cost of the inflated DAT by retaining JSON
copies of every unit.

The current decoder includes the type-30 movement/trail block and the named
type-50 combat/projectile fields, including the type-60 missile-specials block.
Some older and DLC-specific fields remain intentionally opaque; those remain
unexported until their byte offsets are established. The charge target and
secondary projectile are decoded from the creatable tail. The export is
structure-verified against the parser and fixture corpus. It is not engine
verification of semantic meanings for fields labeled `layout_correlated` or
`unmapped`.

The unit base record, creatable tail, and building record now expose the
VER 8.8+ fields whose serialized positions are established by the parser's
fixed-size records. Related UGC paths are resolved through their normalized
tables: training fields through `unit_train_locations`, and type/amount entries
through `unit_attributes`. A mapped attribute may therefore be exported even
when that particular attribute type has no rows in the current DAT. Fields
without a reliable AoE2 DE slot remain `exported=0` (for example `unit_level`);
the older `max_attacks_in_volley` candidate was corrected to the serialized
`total_projectiles` and `max_total_projectiles` fields.

Building graphics now use their actual blocks: spawning/upgrade graphics are
creatable fields; construction/snow graphics are building fields. The former
`construction_unit`, `salvage_unit`, and `salvage_attributes` labels did not
match serialized fields and are no longer exported. The building tail exposes
the Wwise sound ids, pile unit, and six scalar `looting_table_*` entries instead.

To validate an existing export, set `AOE2KIT_TEST_SQLITE` to its database path
and run `go test ./pkg/datfile -run TestPersistentSQLiteExport -v`.
Set `AOE2KIT_TEST_DAT` to a corpus DAT to run the cache/direct decoder comparison
in `go test ./pkg/datcache -run TestCachedUnitMatchesDAT`.

The field ordering was cross-checked against the open-source GenieUtils
serializers for [Unit](https://github.com/Tapsa/genieutils/blob/master/src/dat/Unit.cpp),
[Creatable](https://github.com/Tapsa/genieutils/blob/master/src/dat/unit/Creatable.cpp),
and [Building](https://github.com/Tapsa/genieutils/blob/master/src/dat/unit/Building.cpp).
Kit does not link to or require GenieUtils at runtime.
