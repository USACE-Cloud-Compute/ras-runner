# Test plan — `actions/link/hdf-to-hdf_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, the cc-go-sdk construction in Section 4, test-data spec).
> **Package:** `actions`. **Test file:** `actions/link/hdf-to-hdf_test.go`.

All of these write to a local file under `MODEL_DIR` (or the action's own
`ModelDir` field). Use a `t.TempDir()` as the model dir. For actions with a
`ModelDir string` field, set it directly. For the others (`ColumnToBc`,
`ReflineToBc`, `HdftoHdfDatasetAction`), set the `MODEL_DIR` env var (Change A)
to the temp dir via `t.Setenv("MODEL_DIR", tmp)` before calling `Run()`.

#### `hdf-to-hdf_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** `HdftoHdfDatasetAction.Run` / `CopyHdf5Dataset`.
- **Required data sources:** `src` (input) and `dest` (output), each with
  `Paths["hdf"]` and a `DataPaths` map of equal length; for each key the
  `src.DataPaths[key]` is copied into `dest.DataPaths[key]`.
- **Fixtures:** a source HDF5 and a dest HDF5, each with at least one 2-D
  float32 dataset of identical rows×cols.
- **Assertions:** after `Run()`, the dest dataset's bytes equal the source
  dataset's bytes. Mismatched rows or cols returns the
  `source column count doesnt equal dest column count` /
  `source row count…` errors (assert both error paths with a deliberately
  mismatched fixture).
