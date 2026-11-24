# Test plan — `ras/bFile_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec).
> **Package:** `ras`. **Test file:** `ras/bFile_test.go`.

#### `ras/bFile_test.go` — **replace** (currently exists, uses stale absolute paths)
- **Tag:** `fixture`
- **Covers:** `InitBFile` and the block parsers.
- **Fixtures:** every `*.b01` in `testData/`:
  - `DamBreachOverlapDem.b01` (single breach)
  - `multiDamBreach.b01` (multiple breach structures)
  - `ElkRiver_at_Sutton.b01` (has an `Outlet TS` block)
  - `lewisville.b01`
- **Assertions (table-driven per file):**
  - `InitBFile` returns no error and a non-nil `BfileBlocks`.
  - Every block round-trips: re-encoding the parsed blocks and re-parsing yields
    the same block count and, for each block, the same header and row count.
  - For `ElkRiver_at_Sutton.b01`: exactly one block of type `*OutletTS`; its
    `Name` matches the header text after `Outlet TS -`; `RowCount` > 0; the
    index/flow pairs parse to floats and the block is terminated by the
    ` 3.4E+38` sentinel.
  - For breach files: at least one `*BreachData` block; for each structure row,
    the SNetID cell is non-empty.
- **Read-modify-write round-trip:** pick one breach file, mutate a single
  breach elevation via the exported setter, write to a `t.TempDir()` path, re-
  `InitBFile`, and assert the mutated value is present and **all other blocks
  are byte-identical** to the original (this guards against the known bug
  `getBreachRows only handles first breach block (silent data loss)` — the test
  should FAIL if the second breach block is dropped).
