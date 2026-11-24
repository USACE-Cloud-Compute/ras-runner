# Test plan — `actions/utils/create-ras-tmp_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, the cc-go-sdk construction in Section 4, the
> `.env-testing-local` env in Section 7, test-data spec).
> **Package:** `utils`. **Test file:** `actions/utils/create-ras-tmp_test.go`.

#### `create-ras-tmp_test.go` — **new**
- **Tag:** `integration`
- **Covers:** `CreateRasTmpAction.Run` (creates a temp HDF5 from the source,
  copying the three `RasTmpDatasets`: `Geometry`, `Plan Data`, `Event Conditions`).
- **Setup:** attributes `src` (input DS name), `local_dest` (file name),
  `save_to_remote` (default `false`). `t.Setenv("MODEL_DIR", tmp)`.
- **Fixtures:** a source input HDF5 containing all three top-level groups.
- **Assertions:** the created `<local_dest>` HDF5 contains `Geometry`,
  `Plan Data`, and `Event Conditions` groups; with `save_to_remote: true` the
  result is also written to the output store (assert the DS is registered).
