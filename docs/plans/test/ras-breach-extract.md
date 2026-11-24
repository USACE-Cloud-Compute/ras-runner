# Test plan — `actions/extract/hdf/ras-breach-extract_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, the `writerAccumulator.Reset()`
> note in Section 2 Change C, and the test-data spec).
> **Package:** `hdf`. **Test file:** `actions/extract/hdf/ras-breach-extract_test.go`.

#### `ras-breach-extract_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** `ras-breach-extract.go` (breach data reader) against a real results file.
- **Fixtures:** `Duwamish_17110013.p01.hdf`.
- **Key HDF5 paths the code reads (assert these exist and parse):**
  - time steps: `/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Time`
  - 2D hyd conn: `/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/2D Flow Areas/<FA>/2D Hyd Conn`
  - breaching variables: `<2D Hyd Conn>/<SaConn>/Breaching Variables`
  - attributes read off the Breaching Variables dataset: `Breach at`,
    `Breach at Time (Days)`, `Variable_Unit`
  - columns: `Breach Flow`, `Breach Velocity`, `Bottom-Width`, `Stage HW`,
    `Stage TW` (the `BreachFlowColumnName` … `TwColumnName` consts).
- **Assertions:**
  - The reader finds at least one flow area with a `2D Hyd Conn` group.
  - For each breaching connection, the `Breach at` / `Breach at Time (Days)`
    attributes are read without error and are numeric.
  - The accumulated breach records (`writerAccumulator`) are non-empty and each
    record carries the `SaConn` location field (`breachLocationField`).
