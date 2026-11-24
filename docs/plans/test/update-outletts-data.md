# Test plan — `actions/link/update-outletts-data_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, test-data spec).
> **Package:** `actions`. **Test file:** `actions/link/update-outletts-data_test.go`.

All of these write to a local file under `MODEL_DIR` (or the action's own
`ModelDir` field). Use a `t.TempDir()` as the model dir. For actions with a
`ModelDir string` field, set it directly. For the others (`ColumnToBc`,
`ReflineToBc`, `HdftoHdfDatasetAction`), set the `MODEL_DIR` env var (Change A)
to the temp dir via `t.Setenv("MODEL_DIR", tmp)` before calling `Run()`.

#### `update-outletts-data_test.go` — **replace** (exists, stale absolute path)
- **Tag:** `fixture`
- **Covers:** `UpdateOutletTSAction.Run`.
- **Fixtures:** `ElkRiver_at_Sutton.b01` (has `Outlet TS`) + the results HDF5
  that supplies the observed flow series (`lewisville.p01.hdf` or
  `Duwamish_17110013.p01.hdf` — whichever the payload references).
- **Setup:** `t.TempDir()`; set `ModelDir`; attributes `bFile` + `outletTS`.
- **Assertions:** after `Run()`, the `Outlet TS` block's flow column is
  replaced with the values read from the source HDF5; the index column is
  unchanged; row count is unchanged.
