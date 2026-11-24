# Test plan — `column-to-bc_test.go`

> **Read `docs/plans/test/GENERAL.md` first.** It holds the shared context this
> plan assumes: the testing tiers, the `testdata/` convention, the cc-go-sdk
> `PluginManager` construction (including the exported `cc.ConnectStores` and the
> `go.mod` `replace` policy), the test-data spec, and the CI/env setup. This file
> is self-contained for everything *specific* to the column-to-bc test.

- **Test file:** `actions/link/column-to-bc_test.go` — **new**
- **Package:** `package actions` (same package as the action — required, because
  `getRowVal2` is unexported).
- **Production file under test:** `actions/link/column-to-bc.go` (the *only*
  production file the implementer needs to read, in addition to this plan +
  `GENERAL.md`).
- **Fixtures:** `actions/link/testdata/column-to-bc/` (already created; see below).

---

- **Self-contained.** All fixtures are small files checked into VCS under
  `actions/link/testdata/column-to-bc/`; no big RAS output files and no
  `RAS_RUNNER_TEST_DATA` env are needed.
- **Covers:** `ColumnToBcAction.Run` (adapter) / `MigrateColumnData` / `getRowVal2`.

**Testing strategy — split by what is actually under test.** The column-to-bc
*logic* (time matching, column selection, the `timeval<=0` shortcut, the
migration write) lives entirely in `MigrateColumnData` and `getRowVal2`, and
**neither touches `cc-go-sdk`**: `getRowVal2` is pure, and `MigrateColumnData`
takes a `*cc.DataStore` but only reads three value fields (`StoreType`,
`DsProfile`, `Parameters`) — no methods, no store session. So the logic is
tested **directly, with no `cc.PluginManager`**. The `Run()` method is a thin
*adapter* (parse attributes → resolve the input data source → delegate to
`MigrateColumnData`); it is the only piece that wires in the SDK, and it gets
**one** integration test scoped to adapter concerns (wiring, not migrated
values). **No production refactor is needed** — the logic is already isolated.

**How the action works (verified from source):**
- `Run()` reads `column_index` (string, **1-based**), `time_tolerance` (float,
  default `actions.Tolerance` = `0.000001`), `src` map (`name` + `datapath`),
  `dest` map (`name` + `datapath`). It resolves the input data source via
  `PluginManager.GetInputDataSource(src.name)` and takes the source path from
  that data source's `Paths["hdf"]` **directly** (no store-root prepending —
  see the wrinkle note below).
- `MigrateColumnData(src, srcstore, src_datapath, dest, dest_datapath, readcol,
  timeTolerance)`: if `srcstore.StoreType == "S3"` it builds an S3 URL (needs
  AWS env); **otherwise it opens the local path with `util.OpenFile`**. The
  dest path is resolved at call time as `actions.MODEL_DIR + "/" + dest`.
- `getRowVal2(srcVals, srcTimes, timeval, readcol, timeTolerance)` returns the
  source row's column `readcol-1` where `|destTime − srcTime| < timeTolerance`.
  **Special case: `timeval <= 0.0` returns source row 0 with no time matching.**
  No match → error `"unable to find corresponding input source record for time
  %f"`.
- `MigrateColumnData` rewrites the dest dataset whole: column 0 (time)
  preserved, column 1 replaced.

**Fixtures (already created — checked in at `actions/link/testdata/column-to-bc/`):**
- `src.hdf` (28 KB) — source results file:
  - `Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/Unsteady Time
    Series/2D Flow Areas/ray-roberts/Reference Lines/Flow` — float32, **{841, 6}**
    (one column per reference line).
  - `Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/Unsteady Time
    Series/Time` — float64, **{841}** (sibling `Time` dataset, derived by
    `actions.TimePath`).
  - Extracted from `testData/ray-roberts.p01.hdf` via `h5copy -p` (exact paths
    preserved). Regeneration: `runScripts/` or a throwaway `h5copy` — see `GENERAL.md` Section 3.
- `dest.hdf` (9 KB) — destination input file:
  - `Event Conditions/Unsteady/Boundary Conditions/Flow Hydrographs/2D:
    Lewisville BCLine: ray-roberts_to_lewisville_1` — float32, **{841, 2}**
    (time, value).
  - Extracted from `testData/lewisville.p01.tmp.hdf`. **Its col1 holds a real,
    pre-existing BC series that is *different* from `src.hdf`'s Flow col5**
    (e.g. row 5: dest `50.4956` vs source `0.001292`), so the migration produces
    a visibly different result and the golden assertion is non-tautological.
- `payload-column-to-bc` (1 KB) — a `cc.Payload` JSON. It is a copy of
  `testData/actions/link/column-to-bc/payload` with four field edits:
  - `stores[0].store_type`: `"S3"` → `"FS"` (forces the local open branch).
  - `stores[0].profile`: `"FFRD"` → `""` (empty `DsProfile` for `util.OpenFile`).
  - `inputs[0].paths.hdf`: `"ray-roberts.p01.hdf"` → `"src.hdf"`.
  - `actions[0].attributes.dest.name`: `"lewisville.p01.tmp.hdf"` → `"dest.hdf"`.
  - **Unchanged:** `column_index: "5"`, `time_tolerance: 0.00001`, `src.datapath`,
    `dest.datapath` (the HDF5 dataset paths are identical in the fixtures).
  - `column_index "5"` → Flow dataset column 5 (1-based) = index 4 = the
    `ray-roberts_to_lewisville_1` reference line.

**Tier / tag split:**
- `unit` — `MigrateColumnData` and `getRowVal2` direct tests. Pure logic, no
  `cc.PluginManager`, no CC env. Need the HDF5 C library (rides on
  `hdf5-base` in CI).
- `integration` — the single `Run()` adapter test. Exercises the cc-go-sdk
  wiring (`GetInputDataSource`/`GetStore`). Same HDF5 requirement.

**Setup (shared helpers):**
- Resolve fixtures relative to the package dir: `dir := "testdata/column-to-bc"`.
  Skip if `src.hdf`/`dest.hdf` absent (defensive; they are checked in).
- `unit` `MigrateColumnData` tests: `modelDir := t.TempDir()`; copy
  `dest.hdf` → `modelDir/dest.hdf`; `t.Setenv("MODEL_DIR", modelDir)`. Pass an
  **absolute** source path (the fixture path, or a copy in `t.TempDir()`).
- `getRowVal2` tests: open `src.hdf`, get the `Flow` and `Time`
  `*util.HdfDataset`s; call `getRowVal2` with a chosen `timeval`/`readcol`/
  `tolerance`. No temp dir, no env.
- The `integration` `Run()` test:
  1. Decode `payload-column-to-bc` → `cc.Payload` (load from disk — **no
     `CcStore`, no `InitPluginManager`, no `CC_*` env**).
  2. **Connect the payload's stores** with the exported SDK helper
     `cc.ConnectStores(&payload.Stores)` (added to `cc-go-sdk`; see note below).
     This populates each `DataStore.Session` the way production does. For the
     `FS`/BlockFS store it is env-free (`Connect` reads `params.root` only);
     BlockFS `GetSession` is `nil` by design, but populating `Session` keeps the
     harness faithful.
  3. **Set the absolute source path in-memory:**
     `payload.Inputs[0].Paths["hdf"] = filepath.Join(storeRoot, "src.hdf")`.
     Required because `MigrateColumnData` reads `Paths["hdf"]` **raw** (no
     store-root prepending), so a relative `"src.hdf"` would resolve against the
     test CWD and not be found.
  4. Build `pm := &cc.PluginManager{Payload: payload}`,
     `pm.IOManager.SetParent(&pm.IOManager)`, and for each action
     `pm.Actions[i].IOManager.SetParent(&pm.IOManager)`; then
     `ColumnToBcAction{ActionRunnerBase: cc.ActionRunnerBase{ActionName:
     "column-to-boundary-condition", Action: pm.Actions[0], PluginManager: pm}}`.

  **`cc.ConnectStores` dependency:** this test requires the `cc-go-sdk` change
  that exports `ConnectStores` (the store-connection half of
  `InitPluginManager`, exposed for callers that already hold a `Payload`). Until
  that version is published and the `require` in `go.mod` is bumped, ras-runner
  uses a local `replace` directive:
  `replace github.com/usace-cloud-compute/cc-go-sdk => /workspaces/cc-go-sdk`.
  **Remove the `replace` and bump the `require`** once the SDK version is cut.
  Fallback if the SDK bump is unavailable: replicate the connect loop inline
  (`cc.DataStoreTypeRegistry.New(storeType)` → assert `ConnectionDataStore` →
  `.Connect(ds)` → assign `Session`).

**Test cases:**

*`unit` — `MigrateColumnData` (direct, no `PluginManager`):*
- `TestMigrateColumnData_TimeMatched` — **primary coverage.** Call
  `MigrateColumnData(absSrcPath, &cc.DataStore{StoreType:"FS"}, srcDataPath,
  "dest.hdf", destDataPath, 5, 0.00001)`. Pre-compute golden from `src.hdf`
  (`Flow[i][4]`, all 841 rows; every dest time matches exactly) and the pristine
  dest. After the call, re-open `modelDir/dest.hdf`:
  - col0 (time) identical to original for all rows.
  - col1 == `Flow[i][4]` for all rows (float32, within `1e-3` relative).
  - col1 **differs** from the original dest col1 for at least the middle of the
    series (proves mutation).
  - col1 **is not** `Flow[i][5]` (col 6) for at least one row — off-by-one /
    0-vs-1-based guard.
- `TestMigrateColumnData_NoTimeMatch` — build a tiny dest (in-test or a
  committed `tiny-dest.hdf`) with one row whose time (e.g. `2.5`) is within no
  source time. Call with `readcol=2, tol=0.001`. Assert error contains
  `"unable to find corresponding input source record for time 2.500000"`.
  (The real fixtures' times all match, so this path is unreachable without a
  synthetic file.)
- `TestMigrateColumnData_ZeroTimeShortcut` — build a tiny `src` with
  `Time = [5.0, 6.0]` (`Time[0] != 0`) and a dest whose row 0 time is `0.0`.
  Assert row 0 col1 == `Data[0][readcol-1]` with no time comparison — pins the
  `timeval <= 0.0` shortcut (lines 247–253).

*`unit` — `getRowVal2` (pure):*
- `TestGetRowVal2_TimeMatch` — a `timeval` equal to a source `Time[i]` returns
  `Flow[i][readcol-1]`.
- `TestGetRowVal2_ToleranceBoundary` — `timeval` within `tolerance` of a source
  time matches; just outside it does not (drive both sides).
- `TestGetRowVal2_ZeroTimeShortcut` — `timeval = 0.0` (or negative) returns
  `Flow[0][readcol-1]` regardless of `Time[0]`.
- `TestGetRowVal2_NoMatch` — `timeval` matching no source time returns the
  `"unable to find corresponding input source record for time …"` error.
- `TestGetRowVal2_ColumnSelection` — `readcol=1` returns col 0, `readcol=6`
  returns col 5 (1-based guard, no off-by-one).

*`integration` — `Run()` adapter (one test, thin):*
- `TestColumnToBcRun_Adapter` — decode the payload, stage `src.hdf` (store
  root) and `dest.hdf` (`MODEL_DIR`), build `pm`, call `Run()`. Assert `Run()`
  returns `nil` and the dest file was modified (mtime/bytes changed). **Do not
  assert on the migrated values** — that is `MigrateColumnData`'s job, already
  covered in `unit`. Subtests for attribute validation (no HDF5 data needed):
  - `column_index: "abc"` → `Run()` error contains `"invalid column index"`.
  - missing `src` map → `"missing src attribute data"`.
  - missing `dest` map → `"missing dest attribute data"`.

**Synthetic HDF5 for the error paths:** build at test setup with the
`github.com/usace-cloud-compute/go-hdf5` library (a `buildTinyHdf(t)` helper
writing into `t.TempDir()`), or commit two ~1 KB files
(`tiny-src.hdf`, `tiny-dest.hdf`) if the library's write API is awkward.
Preference: build at test time (self-contained, no repo bloat).

**Verification matrix:**

| Concern | Caught by |
|---|---|
| Correct column selected (1-based) | `TestMigrateColumnData_TimeMatched` golden vs `Flow[i][4]` + `TestGetRowVal2_ColumnSelection` |
| Time matching works | `TestMigrateColumnData_TimeMatched` + `TestGetRowVal2_TimeMatch` / `ToleranceBoundary` |
| Time column untouched | `TestMigrateColumnData_TimeMatched` asserts col0 identical |
| No-match error path | `TestMigrateColumnData_NoTimeMatch` + `TestGetRowVal2_NoMatch` |
| `timeval<=0` shortcut preserved | `TestMigrateColumnData_ZeroTimeShortcut` + `TestGetRowVal2_ZeroTimeShortcut` |
| Dest resolved via `MODEL_DIR` | `TestMigrateColumnData_TimeMatched` (dest staged in temp dir, found via env override) |
| Non-S3 local open branch | every test (store forced to FS) |
| Attribute parsing + wiring | `TestColumnToBcRun_Adapter` subtests |

**cc-go-sdk construction (verified — no open spike):** a
`&cc.PluginManager{Payload: payload}` + `SetParent` is **sufficient** for
`GetInputDataSource`/`GetStore` — both are pure in-memory slice lookups over
the `IOManager`'s `Stores`/`Inputs` (cc-go-sdk `payload.go:176` `GetStore`,
`payload.go:215` `GetInputDataSource`). No `InitPluginManager` and no `CC_*` env
vars are needed. **Stores:** connect the payload's own stores with the exported
`cc.ConnectStores(&payload.Stores)` (see Setup above) so `Session` is populated
the way production does it; this is env-free for the `FS`/BlockFS store.
**Wrinkle:** `MigrateColumnData` reads `Paths["hdf"]` **raw**, without store-root
prepending (it does not call `GetAbsolutePath`). So the `Run()` adapter test must
set `payload.Inputs[0].Paths["hdf"]` to an **absolute** staged path in-memory
before building `pm`; otherwise `util.OpenFile` resolves the relative `"src.hdf"`
against the test CWD (`actions/link/`) and won't find it.

**Optional hardening (separate task, not part of this test plan):** make
`Run()` resolve the source path via `GetAbsolutePath` (store-root + path) so a
relative `Paths["hdf"]` doesn't depend on CWD. This would also remove the
absolute-path injection from the adapter test. File as a bug-hardening note; do
not fold into the test implementation.

**Appendix A — self-contained implementation notes (no need to read other code).**
Everything below is verified against the dependency sources at the pinned
versions in `go.mod`. The implementer needs only `actions/link/column-to-bc.go`
plus this section.

*Functions under test (from `column-to-bc.go`):*
- `MigrateColumnData(src string, srcstore *cc.DataStore, src_datapath string,
  dest string, dest_datapath string, readcol int, timeTolerance float64) error`
  — **exported**; call directly. `dest` is a bare file name, resolved internally
  as `actions.MODEL_DIR + "/" + dest`. `src` must be a path the local open
  branch can resolve (use an **absolute** path).
- `getRowVal2(srcVals, srcTimes *util.HdfDataset, timeval float32, readcol int,
  timeTolerance float64) (float32, error)` — **unexported**, but the test file is
  in `package actions` (same package), so it is callable.
- `ColumnToBcAction.Run() error` — adapter; needs a `cc.PluginManager` (see
  construction note above).

*`util.HdfDataset` read API (`go-hdf5` `util/dataset.go`):*
- Open a dataset from an open `*hdf5.File`:
  ```go
  ds, err := util.NewHdfDataset(datapath, util.HdfReadOptions{
      Dtype:        reflect.Float32, // or reflect.Float64 for Time
      File:         f,
      ReadOnCreate: true,
  })
  ```
- `ds.Dims() []uint` (dims[0]=rows, dims[1]=cols), `ds.Rows() int`,
  `ds.Cols() int`.
- `ds.ReadRow(r int, dest interface{}) error` — `dest` must be a
  `*[]float32` (or `*[]float64`) slice **pre-allocated to the full column count**
  for a 2-D dataset: `row := make([]float32, ds.Cols()); ds.ReadRow(i, &row)`.
  (Internally `readVector` does `value.Set(dv.Slice(offset, offset+colnum))`,
  which panics — recovered to an error — if the dest slice is not exactly
  `colnum` long.) For the 1-D `Time` dataset use `make([]float64, 1)`.
- `ds.Close()`.
- To read an entire column, loop `ReadRow` and collect `row[c]`, or use
  `ds.ReadColumn(c int, dest *[]float32)`.

*Building the synthetic error-path files (`go-hdf5`):*
- `f, err := hdf5.CreateFile(path, hdf5.F_ACC_TRUNC)` (or `hdf5.OpenFile(path,
  hdf5.F_ACC_TRUNC)`).
- Nested groups: create the **full path** in one call —
  `f.CreateGroup("/a/b/c")` works (it creates parents). Use the same group
  paths as the real fixtures so the datapath strings match.
- `ds, err := f.CreateDataset("name", hdf5.T_NATIVE_FLOAT, dspace)` for
  float32, `hdf5.T_NATIVE_DOUBLE` for float64.
- `dspace, err := hdf5.CreateSimpleDataspace([]uint{rows, cols}, []uint{rows,
  cols})` (2-D) or `[]uint{n}` (1-D).
- `ds.Write(&slice)` where `slice` is a flat `[]float32` of length `rows*cols`
  (row-major).
- Close `ds`, `dspace`, `f`.
- The synthetic `src` for `getRowVal2_ZeroTimeShortcut` must have
  `Time = [5.0, 6.0]` (so `Time[0] != 0`) and a `Flow` with ≥1 row; the dest
  row 0 time must be `0.0`. The synthetic dest for `NoTimeMatch` needs one row
  whose time (e.g. `2.5`) is within no source `Time`.
- If the `go-hdf5` write API fights you, fall back to committing two ~1 KB
  `tiny-src.hdf` / `tiny-dest.hdf` fixtures built once via a throwaway Go
  program (or `h5py` if available).

*`cc.DataStore` literal (for direct `MigrateColumnData` calls):*
  `&cc.DataStore{StoreType: "FS"}` — `StoreType` is the only field that matters
  for the local open branch (`DsProfile`/`Parameters` stay zero-valued).

*File copy helper (staging):*
  ```go
  func copyFile(t *testing.T, src, dest string) {
      t.Helper()
      b, err := os.ReadFile(src)
      if err != nil { t.Fatalf("read %s: %v", src, err) }
      if err := os.WriteFile(dest, b, 0644); err != nil { t.Fatalf("write %s: %v", dest, err) }
  }
  ```
  (Mirror the existing `copyFile` in `ras-extract-action_test.go`.)

*Build tags & file header:*
  - The `unit` and `integration` tests can live in the same file. If the team
    wants `go test -tags unit` to skip the SDK-dependent adapter test, split
    into `column-to-bc_unit_test.go` (`//go:build unit`) and
    `column-to-bc_integration_test.go` (`//go:build integration`). Both are in
    `package actions`.
  - The `getRowVal2` and `MigrateColumnData` tests are the `unit` tier; the
    `Run()` adapter test is `integration`.

*Minimal skeleton (the shape, not the full implementation):*
  ```go
  //go:build unit
  package actions

  import (
      "os"
      "path/filepath"
      "reflect"
      "testing"

      cc "github.com/usace-cloud-compute/cc-go-sdk"
      "github.com/usace-cloud-compute/go-hdf5"
      "github.com/usace-cloud-compute/go-hdf5/util"
  )

  const (
      c2bDir        = "testdata/column-to-bc"
      c2bSrcFlowPath = "Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/" +
          "Unsteady Time Series/2D Flow Areas/ray-roberts/Reference Lines/Flow"
      c2bSrcTimePath = "Results/Unsteady/Output/Output Blocks/DSS Hydrograph Output/" +
          "Unsteady Time Series/Time"
      c2bDestPath    = "Event Conditions/Unsteady/Boundary Conditions/Flow Hydrographs/" +
          "2D: Lewisville BCLine: ray-roberts_to_lewisville_1"
  )

  // openSrc opens src.hdf and returns (Flow *util.HdfDataset f32 841x6,
  // Time *util.HdfDataset f64 841, file *hdf5.File). Caller closes all.
  func openSrc(t *testing.T) (*util.HdfDataset, *util.HdfDataset, *hdf5.File) { /* ... */ }

  func TestMigrateColumnData_TimeMatched(t *testing.T) {
      src := filepath.Join(c2bDir, "src.hdf")
      if _, err := os.Stat(src); err != nil { t.Skipf("no fixture: %v", err) }
      modelDir := t.TempDir()
      copyFile(t, filepath.Join(c2bDir, "dest.hdf"), filepath.Join(modelDir, "dest.hdf"))
      t.Setenv("MODEL_DIR", modelDir)
      // golden: read Flow[i][4] for i in 0..840 from src
      // before:  read dest.hdf col0/col1 (snapshot col1)
      if err := MigrateColumnData(src, &cc.DataStore{StoreType: "FS"},
          c2bSrcFlowPath, "dest.hdf", c2bDestPath, 5, 0.00001); err != nil {
          t.Fatalf("MigrateColumnData: %v", err)
      }
      // after: re-open modelDir/dest.hdf; assert col0 unchanged,
      // col1 == golden (1e-3 rel), col1 != before (middle rows), col1 != Flow[i][5]
  }

  func TestGetRowVal2_TimeMatch(t *testing.T) {
      flow, time, f := openSrc(t)
      defer flow.Close(); defer time.Close(); defer f.Close()
      // pick a known source time, e.g. t=Time[10]; assert
      // getRowVal2(flow, time, float32(Time[10]), 5, 1e-5) == Flow[10][4]
  }
  // ... NoTimeMatch, ZeroTimeShortcut, ToleranceBoundary, ColumnSelection ...
  ```

*`integration` skeleton (`column-to-bc_integration_test.go`, `//go:build
integration`, `package actions`):*
  ```go
  func TestColumnToBcRun_Adapter(t *testing.T) {
      payloadBytes, _ := os.ReadFile(filepath.Join(c2bDir, "payload-column-to-bc"))
      var payload cc.Payload
      if err := json.Unmarshal(payloadBytes, &payload); err != nil { t.Fatalf("payload: %v", err) }

      // stage files
      storeRoot := t.TempDir()
      copyFile(t, filepath.Join(c2bDir, "src.hdf"),  filepath.Join(storeRoot, "src.hdf"))
      modelDir  := t.TempDir()
      copyFile(t, filepath.Join(c2bDir, "dest.hdf"), filepath.Join(modelDir, "dest.hdf"))
      t.Setenv("MODEL_DIR", modelDir)

      // connect the payload's stores (exported SDK helper; env-free for FS/BlockFS)
      if err := cc.ConnectStores(&payload.Stores); err != nil { t.Fatalf("ConnectStores: %v", err) }

      // wrinkle: MigrateColumnData reads Paths["hdf"] raw -> make it absolute
      payload.Inputs[0].Paths["hdf"] = filepath.Join(storeRoot, "src.hdf")

      pm := &cc.PluginManager{Payload: payload}
      pm.IOManager.SetParent(&pm.IOManager)
      for i := range pm.Actions { pm.Actions[i].IOManager.SetParent(&pm.IOManager) }

      runner := ColumnToBcAction{ActionRunnerBase: cc.ActionRunnerBase{
          ActionName:    "column-to-boundary-condition",
          Action:        pm.Actions[0],
          PluginManager: pm,
      }}
      if err := runner.Run(); err != nil { t.Fatalf("Run: %v", err) }
      // assert dest file was modified (mtime/bytes); do NOT assert migrated values
      // subtests: column_index "abc" -> "invalid column index";
      //           missing src -> "missing src attribute data";
      //           missing dest -> "missing dest attribute data"
  }
  ```

*What the implementer does NOT need to read:* the rest of `actions/`,
`ras/`, `fragilitycurve/`, other test files (except to copy the `copyFile`
pattern), or the cc-go-sdk internals (the construction + wrinkle are stated
above). The only production file to read is `actions/link/column-to-bc.go`.

