# Test plan — `ras/outletTS_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec).
> **Package:** `ras`. **Test file:** `ras/outletTS_test.go`.

#### `ras/outletTS_test.go` — **new**
- **Tag:** `fixture` (plus one `unit` case for the sentinel)
- **Covers:** the `Outlet TS - <name>` block parser.
- **Fixtures:** `ElkRiver_at_Sutton.b01`.
- **Assertions:**
  - Header row 0 is `Outlet TS - <name>`; row 1 is the count.
  - Index/flow pairs are read 5 per line.
  - The block ends at the ` 3.4E+38` sentinel; any lines after the sentinel
    (belonging to the next block) are NOT consumed as data.
  - `unit` case: feed a 6-line in-memory block (count 2, then a sentinel) and
    assert exactly 2 pairs are returned and the next line is untouched.
