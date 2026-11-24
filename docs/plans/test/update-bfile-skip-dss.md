# Test plan — `actions/link/update-bfile-skip-dss_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, test-data spec).
> **Package:** `actions`. **Test file:** `actions/link/update-bfile-skip-dss_test.go`.

All of these write to a local file under `MODEL_DIR` (or the action's own
`ModelDir` field). Use a `t.TempDir()` as the model dir. For actions with a
`ModelDir string` field, set it directly. For the others (`ColumnToBc`,
`ReflineToBc`, `HdftoHdfDatasetAction`), set the `MODEL_DIR` env var (Change A)
to the temp dir via `t.Setenv("MODEL_DIR", tmp)` before calling `Run()`.

#### `update-bfile-skip-dss_test.go` — **new**
- **Tag:** `unit` (pure string manipulation on a b-file; no HDF5 needed)
- **Covers:** `UpdateBfileSkipDSSAction.Run`.
- **Setup:** `t.TempDir()`; write a minimal b-file; set `ModelDir`; attribute
  `bFile`.
- **Assertions:**
  - If the file does not already contain `Extra Commands\n1\nSKIP_HDF_DSS`,
    that exact block is appended.
  - If it already contains it, it is **not** appended twice (idempotency).
