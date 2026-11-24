# Test plan — `actions/link/update-breach-data_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, test-data spec).
> **Package:** `actions`. **Test file:** `actions/link/update-breach-data_test.go`.

All of these write to a local file under `MODEL_DIR` (or the action's own
`ModelDir` field). Use a `t.TempDir()` as the model dir. For actions with a
`ModelDir string` field, set it directly. For the others (`ColumnToBc`,
`ReflineToBc`, `HdftoHdfDatasetAction`), set the `MODEL_DIR` env var (Change A)
to the temp dir via `t.Setenv("MODEL_DIR", tmp)` before calling `Run()`.

#### `update-breach-data_test.go` — **replace** (exists, stale absolute path)
- **Tag:** `fixture`
- **Covers:** `UpdateBfileAction.Run`.
- **Fixtures:** a b-file with a `Breach Data` block (`multiDamBreach.b01`),
  the matching geometry HDF5 (`Duwamish_17110013.g01.hdf`), and
  `testData/testFragilityCurveOutput.json` (shape:
  `{"results":[{"location":"2","failure_elevation":100.878…}]}`).
- **Setup:** `t.TempDir()`; copy the b-file + geo HDF5 in; set
  `UpdateBfileAction{ModelDir: tmp}`; build `cc.Action` with `Attributes` map
  naming `bFile`, the geo HDF5, and the fragility JSON.
- **Assertions:** after `Run()`, re-`InitBFile` the written b-file; the breach
  elevation for the location named in the fragility JSON equals the
  `failure_elevation`; all other structures are unchanged.
