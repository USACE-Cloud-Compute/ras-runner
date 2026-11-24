# Test plan — `ras/breachData_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec).
> **Package:** `ras`. **Test file:** `ras/breachData_test.go`.

#### `ras/breachData_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** the `BreachData` row parser in detail.
- **Fixtures:** `DamBreachOverlapDem.b01`, `multiDamBreach.b01`.
- **Assertions:**
  - Breach count (row 1 of the block) equals the number of structure rows parsed.
  - Mass-wasting flag: when the column index 13 cell parses to `1`, the parser
    consumes the extra row 2; when it is `0`, it does not. (Drive both cases by
    choosing which fixture/structure row exercises each — if neither fixture
    has a mass-wasting row, hand-construct a minimal in-memory `BreachData`
    from raw lines and assert.)
  - Simplified physical breach method: when column index 9 == `1`, rows 7/8
    hold the 5 ordinates per row; assert the ordinate count (row 5) matches the
    number of ordinate values read.
  - Edge cases (construct raw strings, no file needed → can be `unit`):
    - a `Breach Data` block with count `0` → zero structure rows, no panic.
    - a block terminated early by the next block's header letter → parser stops
      at the first line whose first byte is a letter (the "extra lines from
      next block" case).
