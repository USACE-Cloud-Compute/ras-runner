# ras-runner plans

Working handoffs for implementation agents and developers, split by lifetime:

- **`test/`** — durable test plans. One file per `_test.go`, each self-contained
  (fixtures, assertions, setup, API refs) so a single agent can write one test
  without ingesting the rest of the codebase. Plus `test/GENERAL.md`, the shared
  cross-cutting context every test plan assumes.
- **`impl/`** — ephemeral implementation / requirements plans. One file per
  feature, refactor, or API change. Archived (not deleted) when the work merges.
  See `impl/README.md` for the conventions.

**Dispatch rule for a test agent:** *"Read `docs/plans/test/GENERAL.md` and
`docs/plans/test/<name>.md`. Implement the test. You do not need any other file."*

## Index

### Test plans (`test/`)

`test/GENERAL.md` is the shared context every plan assumes. Each per-test file is
named for the `_test.go` it produces and holds that test's spec.

**`ras/` (core RAS file parsing):**

| File | Test file | Tag | Summary |
|------|-----------|-----|---------|
| `test/bFile.md` | `ras/bFile_test.go` | `fixture` | `InitBFile` + block parsers; round-trip; read-modify-write guard. |
| `test/breachData.md` | `ras/breachData_test.go` | `fixture` | `BreachData` row parser in detail; mass-wasting / simplified-breach edge cases. |
| `test/outletTS.md` | `ras/outletTS_test.go` | `fixture` (+`unit`) | `Outlet TS` block parser; sentinel termination. |
| `test/structuresAttr.md` | `ras/structuresAttr_test.go` | `fixture` | `ReadSNetIDToNameFromGeoHDF` + pre-6.6/6.6 version dispatch. |
| `test/hdf5version.md` | `ras/hdf5version_test.go` | `fixture` | `File Version` root-attribute read; `Hdf5Float32` NaN→null. |

**`actions/extract/hdf/` (HDF5 extraction):**

| File | Test file | Tag | Summary |
|------|-----------|-----|---------|
| `test/ras-extractor.md` | `ras-extractor_test.go` | `fixture` | Generic `DataExtract[T]`: dtypes, postprocess, exclude, writers, modes. |
| `test/ras-breach-extract.md` | `ras-breach-extract_test.go` | `fixture` | Breach data reader against a real results file. |
| `test/ras-extract-action.md` | `ras-extract-action_test.go` | `fixture` | `RasExtractAction.Run` end-to-end. |
| `test/ras-breach-action.md` | `ras-breach-action_test.go` | `fixture` | `RasBreachExtractAction.Run` end-to-end. |

**`actions/link/` (boundary-condition / b-file update):**

| File | Test file | Tag | Summary |
|------|-----------|-----|---------|
| `test/column-to-bc.md` | `column-to-bc_test.go` | `unit` + `integration` | **Worked example** — self-contained; fixtures checked in at `actions/link/testdata/column-to-bc/`. |
| `test/refline-to-bc.md` | `refline-to-bc_test.go` | `fixture` | `ReflineToBc.Run` / `MigrateRefLineData`. |
| `test/hdf-to-hdf.md` | `hdf-to-hdf_test.go` | `fixture` | `HdftoHdfDatasetAction.Run` / `CopyHdf5Dataset`. |
| `test/update-breach-data.md` | `update-breach-data_test.go` | `fixture` | `UpdateBfileAction.Run`. |
| `test/update-outletts-data.md` | `update-outletts-data_test.go` | `fixture` | `UpdateOutletTSAction.Run`. |
| `test/update-bfile-skip-dss.md` | `update-bfile-skip-dss_test.go` | `unit` | `UpdateBfileSkipDSSAction.Run` (idempotent block append). |

**`actions/run/` (simulation):**

| File | Test file | Tag | Summary |
|------|-----------|-----|---------|
| `test/unsteady-simulation.md` | `unsteady-simulation-action_test.go` | `e2e` (+`unit` sub) | Shells out to RAS binary; summary-read `unit` sub-test. |

**`actions/utils/`:**

| File | Test file | Tag | Summary |
|------|-----------|-----|---------|
| `test/blockfile.md` | `blockfile_test.go` | `unit` | `NewBlockFile` / `BlockFile` JSON parsing + filtering. |
| `test/copy-inputs-action.md` | `copy-inputs-action_test.go` | `integration` | `CopyInputsAction.Run`. |
| `test/post-outputs-action.md` | `post-outputs-action_test.go` | `integration` | `PostOutputsAction.Run` / reserved-name skipping. |
| `test/create-ras-tmp.md` | `create-ras-tmp_test.go` | `integration` | `CreateRasTmpAction.Run`. |

**`fragilitycurve/` + root:**

| File | Test file | Tag | Summary |
|------|-----------|-----|---------|
| `test/fragilitycurve.md` | `fragilitycurve/fragilitycurve_test.go` | `unit` | `ModelResult` JSON round-trip. |
| `test/main.md` | `main_test.go` | `unit` | `TestAllActionsRegistered` registry smoke test. |

### Impl plans (`impl/`)

| File | Status | Summary |
|------|--------|---------|
| — | — | None yet. See `impl/README.md`. |
