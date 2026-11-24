# Test plan — `fragilitycurve_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec).
> **Package:** `fragilitycurve`. **Test file:** `fragilitycurve/fragilitycurve_test.go`.

#### `fragilitycurve_test.go` — **new**
- **Tag:** `unit`
- **Covers:** JSON round-trip of `ModelResult` / `FragilityCurveLocationResult`.
- **Fixtures:** `testData/testFragilityCurveOutput.json`
  (`{"results":[{"location":"2","failure_elevation":100.878…}]}`) and
  `testData/failure_elevations.json` (large, real).
- **Assertions:**
  - Unmarshal `testFragilityCurveOutput.json` into `ModelResult`; `len(Results)`
    is 1; `Results[0].Name == "2"`; `Results[0].FailureElevation ≈ 100.87824735121977`.
  - Unmarshal `failure_elevations.json`; `len(Results)` matches the number of
    entries in the file; every `Name` is non-empty and every
    `FailureElevation` is finite.
  - Round-trip: marshal a hand-built `ModelResult` and unmarshal it back; the
    `location`/`failure_elevation` JSON tags are preserved.
