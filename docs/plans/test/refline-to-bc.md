# Test plan — `actions/link/refline-to-bc_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, the cc-go-sdk construction in Section 4, test-data spec).
> **Package:** `actions`. **Test file:** `actions/link/refline-to-bc_test.go`.

All of these write to a local file under `MODEL_DIR` (or the action's own
`ModelDir` field). Use a `t.TempDir()` as the model dir. For actions with a
`ModelDir string` field, set it directly. For the others (`ColumnToBc`,
`ReflineToBc`, `HdftoHdfDatasetAction`), set the `MODEL_DIR` env var (Change A)
to the temp dir via `t.Setenv("MODEL_DIR", tmp)` before calling `Run()`.

#### `refline-to-bc_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** `ReflineToBc.Run` / `MigrateRefLineData`.
- **Required attributes (confirmed):**
  - `refline` (string, the reference-line name to extract)
  - `use-remote-reads` (bool, optional, default `true`; set **false** for this
    test so it opens the local file)
  - `source` (via `IOManager.GetInputDataSource("source")`): DS with
    `Paths["hdf"]` and `DataPaths["refline"]`
  - `destination` (via `IOManager.GetOutputDataSource("destination")`): DS with
    `Paths["hdf"]` and `DataPaths["bcline"]`
- **Fixtures:** a results HDF5 with a reference-line group containing
  `Flow` (2-D float32), `Name` (string set), and a `Time` dataset; a dest
  input HDF5 with a 2-column float32 BC dataset of the **same row count**
  (this action assumes identical time arrays — no time matching).
- **Assertions:** after `Run()`, the dest column 2 equals the refline `Flow`
  column at the index whose `Name` == `refline`; an unknown `refline` name
  returns the `invalid reference line` error.
