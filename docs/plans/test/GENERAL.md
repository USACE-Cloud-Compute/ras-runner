# ras-runner test plans — shared context

This file is the **shared, cross-cutting context** that every per-test plan
assumes. An implementation agent for a single test reads **this file plus exactly
one** `docs/plans/test/<name>.md` — nothing else is required.

The per-test files live in this directory (`docs/plans/test/`), one file per
`_test.go`. Each is written to be self-contained: it names the fixtures, the
assertions, the setup, and any API references the agent needs, so the agent does
not have to read the rest of the codebase.

---

## 1. Testing tiers

| Tier | Tag | Scope | Gate / env | Runs in CI |
|------|-----|-------|------------|------------|
| **Unit** | `unit` | Pure logic: `ras/` parsing helpers, `actions/utils/blockfile.go`, `fragilitycurve`, action-registry smoke test, and self-contained action tests (e.g. `column-to-bc`) | none | always (push + PR) |
| **Fixture** | `fixture` | Tests that read real RAS files (b-files, geometry HDF5, results HDF5) | `RAS_RUNNER_TEST_DATA` env, default `./testData` | yes, but **skips cleanly** when data files are absent |
| **Integration (local-CC)** | `integration` | Full `cc-go-sdk` plugin-manager runs against a filesystem store | `.env-testing-local` env vars (`CC_STORE_TYPE=FS`, `CC_FSB_ROOT_PATH`, …) | **no** — manual / dev box only |
| **E2E (S3)** | `e2e` | Plugin manager against real S3/MinIO + running RAS binary | `.env-testing` env vars + `RAS_SCRIPT_PATH` | **no** — manual only |

A test belongs to exactly one tier. Use Go build tags so the `unit` tier compiles
and runs with zero external data. Fixture/integration/e2e tests must call a
`skipIfNoTestData(t)` / `skipIfNoEnv(t, ...)` helper (Section 2, Change B) at the
top so a missing fixture or missing env var produces a skip, not a failure.

> **Note on self-contained action tests:** an action test can be `unit`-tagged
> even though it exercises an action, provided its data is small and checked in
> under the package `testdata/` dir (Section 3) and it does not need real RAS
> output files or CC env vars. `column-to-bc` is the reference example.

---

## 2. Testability refactor — documented, NOT implemented

These are prerequisites the test code assumes. They are **documentation only**;
do not edit non-test source files as part of a test plan unless separately told.

### Change A — make `actions.MODEL_DIR` env-overridable

**File:** `actions/actions-root.go`

`MODEL_DIR` is currently a `const` equal to the hardcoded absolute path
`/sim/model`. Several actions write results directly to `actions.MODEL_DIR`
(`column-to-boundary-condition`, `refline-to-boundary-condition`, `copy-inputs`,
`post-outputs`) and cannot be pointed at a temp dir otherwise. Change:

```go
// Before
const (
    MODEL_DIR = "/sim/model"
    // ...
)

// After
var MODEL_DIR = func() string {
    if v := os.Getenv("MODEL_DIR"); v != "" {
        return v
    }
    return "/sim/model"
}()
```

(Keep it in the same const/var block; promote it to a package-level `var`.)
Actions that already carry a `ModelDir string` field (`UpdateBfileAction`,
`UpdateOutletTSAction`, `UpdateBfileSkipDSSAction`) need no change — they fall
back to `actions.MODEL_DIR` when their field is empty, and tests set the field
explicitly.

### Change B — `testdata()` and skip helpers

Add a small shared helper. Recommended location: a new package
`internal/testutil/testutil.go` (or, to avoid a new package, a `helpers_test.go`
per package that duplicates the ~6-line helper — either is acceptable).

```go
package testutil

import (
    "os"
    "testing"
)

// TestDataDir returns the root of the test fixtures. It honors the
// RAS_RUNNER_TEST_DATA env var and defaults to ./testData (repo-relative).
func TestDataDir() string {
    if v := os.Getenv("RAS_RUNNER_TEST_DATA"); v != "" {
        return v
    }
    return "./testData"
}

// SkipIfNoTestData skips the test if the fixture root does not exist.
func SkipIfNoTestData(t *testing.T) {
    t.Helper()
    dir := TestDataDir()
    if _, err := os.Stat(dir); err != nil {
        t.Skipf("test data not found at %s — set RAS_RUNNER_TEST_DATA to run this test", dir)
    }
}

// SkipIfNoEnv skips the test if any named env var is empty.
func SkipIfNoEnv(t *testing.T, vars ...string) {
    t.Helper()
    for _, v := range vars {
        if os.Getenv(v) == "" {
            t.Skipf("env var %s not set — skipping", v)
        }
    }
}
```

### Change C — `writerAccumulator.Reset()` between tests

`actions/extract/hdf/ras-extractor-writers.go` holds a package-level global
`writerAccumulator`. Every test in `actions/extract/hdf` that exercises the
extract path must call `writerAccumulator.Reset()` (or set `writerAccumulator`
to a fresh `writerAccumulatorImpl{Data: make(map[string][]map[string]any)}`) in
a `t.Cleanup` or at the top of the test to avoid cross-test contamination.

---

## 3. `testdata/` convention (self-contained test fixtures)

Some action tests are **self-contained**: their data is small, checked into VCS,
and lives in a `testdata/` directory next to the test file (the Go tooling ignores
any dir named `testdata` during `go build`). This keeps them out of the compiled
binary, separate from the real HEC-RAS output files in `testData/`, and runnable
with zero external data. The convention:

```
<package dir>/
├── <action>.go
├── <action>_test.go
└── testdata/
    └── <action>/        # one subfolder per test; a test only touches its own
        ├── src.hdf
        ├── dest.hdf
        └── payload-<action>
```

A test resolves its data relative to its own package dir (e.g.
`dir := "testdata/column-to-bc"`) and does **not** use `RAS_RUNNER_TEST_DATA`
(that env var is only for the big RAS files in `testData/` used by the
`fixture`-tagged tests).

> This convention was chosen over a top-level `testing/` tree that mirrors the
> source layout: `testdata/` in the package dir is the Go idiom (the tool ignores
> it on build), needs no path-walking to the repo root, and moves with the code
> if a package is renamed. Do not re-litigate the location.

---

## 4. cc-go-sdk construction for action tests

When a test drives an action's `Run()` (the "adapter" test), build the
`PluginManager` **without** a `CcStore` and **without** `InitPluginManager`:

1. **Load the payload from disk** — `json.Unmarshal` the payload fixture into a
   `cc.Payload`. No `CcStore`, no `CC_*` env.
2. **Connect the payload's own stores** with the exported SDK helper
   `cc.ConnectStores(&payload.Stores)`. This populates each `DataStore.Session`
   the way production does (`InitPluginManager` does this internally via the
   unexported `connectStores`). For the `FS`/BlockFS store it is env-free
   (`Connect` reads `params.root` only); BlockFS `GetSession` is `nil` by design,
   but populating `Session` keeps the harness faithful.
3. **Set any raw source paths to absolute** — some actions read
   `Paths["<key>"]` **raw** (no store-root prepending via `GetAbsolutePath`).
   For those, set `payload.Inputs[i].Paths["<key>"]` to an absolute staged path
   in-memory before building `pm`; otherwise the relative path resolves against
   the test CWD and is not found. (column-to-bc is the known case — see its plan.)
4. **Wire the IOManagers:** `pm := &cc.PluginManager{Payload: payload}`,
   `pm.IOManager.SetParent(&pm.IOManager)`, and for each action
   `pm.Actions[i].IOManager.SetParent(&pm.IOManager)`.

`GetInputDataSource` / `GetStore` / `GetOutputDataSource` are pure in-memory
slice lookups over the `IOManager`'s `Inputs`/`Stores`/`Outputs`
(cc-go-sdk `payload.go`), so no live store session is required for *lookups* —
but step 2 still populates `Session` for fidelity and for actions that do use
`pm.Get`/`pm.Put`/`pm.Copy`.

**`cc.ConnectStores` dependency:** this requires the `cc-go-sdk` change that
exports `ConnectStores`. Until that version is published and the `require` in
`go.mod` is bumped, ras-runner uses a local `replace` directive:
`replace github.com/usace-cloud-compute/cc-go-sdk => /workspaces/cc-go-sdk`.
**Remove the `replace` and bump the `require`** once the SDK version is cut.
Fallback if the SDK bump is unavailable: replicate the connect loop inline
(`cc.DataStoreTypeRegistry.New(storeType)` → assert `ConnectionDataStore` →
`.Connect(ds)` → assign `Session`).

---

## 5. Test data specification

All big HDF5 fixtures **must be generated by HEC RAS** (no synthesized HDF5). The
user supplies them. Small self-contained fixtures (Section 3) are the exception —
those are derived from the RAS files but checked in.

### 5.1 Already present in `testData/`

- `DamBreachOverlapDem.b01` — `ras/bFile_test.go`, `ras/breachData_test.go` (single breach)
- `multiDamBreach.b01` — `ras/bFile_test.go`, `ras/breachData_test.go`, `link/update-breach-data_test.go` (multiple breach structures)
- `ElkRiver_at_Sutton.b01` — `ras/bFile_test.go`, `ras/outletTS_test.go`, `link/update-outletts-data_test.go` (has an `Outlet TS` block)
- `lewisville.b01` — `ras/bFile_test.go`
- `Duwamish_17110013.p01.hdf` — results (~502 MB), `actions/extract/hdf/*`
- `clear-creek.p01.hdf` — results (~126 MB), secondary for `actions/extract/hdf/*`
- `lewisville.p01.hdf` — results (~348 MB), `actions/extract/hdf/*`
- `lewisville.p01.tmp.hdf` — temp HDF5 (~57 MB), dest in `link/hdf-to-hdf_test.go`
- `ray-roberts.p01.hdf` — results (~309 MB), secondary for `actions/extract/hdf/*`
- `testFragilityCurveOutput.json` — tiny; `fragilitycurve`, `link/update-breach-data_test.go`
- `failure_elevations.json` — large, real fragility output; `fragilitycurve`
- `testData/actions/run/unsteady-simulation/` — `Muncie.b04`, `.c02`, `.g04.hdf`, `.x04`, `.p04.tmp.hdf`, `payload`; `actions/run/unsteady-simulation-action_test.go`
- `testData/actions/run/unsteady-simulation-local-cc/payload` — local-CC variant payload
- `testData/actions/link/{column-to-bc,hdf_to_hdf,refline-to-bc,update-breach-data,update-outletsts-data}/payload` — reference `cc.Payload` JSON for the link actions

> `link/column-to-bc_test.go` does **not** use any of the big RAS output files.
> It uses small self-contained fixtures under `actions/link/testdata/column-to-bc/`
> (Section 3), extracted once from `ray-roberts.p01.hdf` and `lewisville.p01.tmp.hdf`.

### 5.2 To acquire (missing)

1. **`Duwamish_17110013.g01.hdf`** — a **pre-6.6** RAS geometry file. Referenced
   by the existing `structuresAttr_test.go` (asserts `actual["Highway 120"] == 2`
   and `len(actual) == 11`) but **absent** from `testData/`. Must contain:
   - root group attribute `File Version` that does **not** contain `"6.6"`;
   - a compound dataset at `Geometry/Structures/Attributes/` with the
     `structuresAttr` field layout: `Type`[16], `Mode`[18], `River`[16],
     `Reach`[16], `RS`[8], `Connection`[16], `Groupname`[45], `US Type`[16],
     `US River`[16], `US Reach`[16], … (see `ras/structuresAttr.go` for the full struct).
2. **A 6.6 geometry HDF5** (any model) — root `File Version` **does** contain
   `"6.6"`; compound dataset at `Geometry/Structures/Attributes/` with the
   `structuresAttr66` field layout. Used by `ras/structuresAttr_test.go` and
   `ras/hdf5version_test.go` to exercise the version dispatch.
3. **`testData/actions/extract/ras-extract/payload`** — a `cc.Payload` JSON with
   one action of type `ras-extract` whose attributes name the
   `Duwamish_17110013.p01.hdf` input source and the extraction parameters
   (group path, colnames, data type, writer type). Mirror the shape of the
   existing link payloads.
4. **`testData/actions/extract/ras-breach-extract/payload`** — same pattern,
   action type `ras-breach-extract`.

### 5.3 Formats the Go code expects (so files can be verified)

**B-file** (`.b01`/`.b04` etc.): fixed-width text, **8-character-wide cells**.
Blocks are delimited by a header line whose **first byte is a letter**.
- **`Breach Data` block:**
  - row 0: header `Breach Data`
  - row 1: breach (structure) count
  - then, per structure:
    - row 1: SNetID and other structure fields
    - **optional row 2** when the mass-wasting flag (column index 13) == `1`
    - row 5: progression-ordinate count
    - **rows 7 and 8** hold the 5 ordinates per row when the simplified
      physical breach method (column index 9) == `1`
- **`Outlet TS - <name>` block:**
  - row 0: header `Outlet TS - <name>`
  - row 1: count
  - then index/flow pairs, **5 per line**, terminated by the ` 3.4E+38`
    sentinel; any lines after the sentinel belong to the next block and must
    not be consumed.

**Fragility-curve JSON** (`fragilitycurve.ModelResult`):
```json
{"results": [{"location": "<string>", "failure_elevation": <float64>}, ...]}
```
JSON tags: `location`, `failure_elevation`, top-level `results`.

**`cc.Payload` JSON** (action test fixtures): the shape produced by the
`cc-go-sdk` — an object with an `actions` array; each action has a `type`
(the registered action name) and an `attributes` object (the per-action
parameters documented in the per-test plans). The existing payloads under
`testData/actions/...` are the reference examples.

**Geometry HDF5:** root group attribute `File Version` (string); compound
dataset `Geometry/Structures/Attributes/` (layout per `ras/structuresAttr.go`,
version-dependent).

**Results HDF5** (`.p01.hdf`): the `Results/Unsteady/Output/Output
Blocks/Base Output/Unsteady Time Series/...` tree with `Time`, `Boundary
Conditions`, `2D Flow Areas/<FA>/2D Hyd Conn/<SaConn>/Breaching Variables`
(datasets + `Breach at`, `Breach at Time (Days)`, `Variable_Unit` attributes),
and `/Results/Unsteady/Summary` with a `Solution` attribute.

---

## 6. CI — GitHub Actions workflow

Create `.github/workflows/test.yml` with the contents below. The `unit` and
`fixture` jobs run on every push/PR to `main`; the `integration` job is
`workflow_dispatch`-only (it needs the local-CC env vars and a populated
fixture store). All jobs build from the existing `Dockerfile.hdf5.base` image
for the HDF5 C library.

```yaml
name: test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:

jobs:
  unit:
    runs-on: ubuntu-latest
    container:
      image: ghcr.io/usace-cloud-compute/ras-runner:hdf5-base
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: vet
        run: go vet ./...
      - name: unit tests
        run: go test -tags unit -v ./...

  fixture:
    runs-on: ubuntu-latest
    needs: unit
    container:
      image: ghcr.io/usace-cloud-compute/ras-runner:hdf5-base
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: fixture tests (skip if no data)
        env:
          RAS_RUNNER_TEST_DATA: ./testData
        run: go test -tags fixture -v ./...

  integration:
    runs-on: ubuntu-latest
    if: github.event_name == 'workflow_dispatch'
    container:
      image: ghcr.io/usace-cloud-compute/ras-runner:hdf5-base
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - name: load local-CC env
        run: |
          set -a
          source ./.env-testing-local
          set +a
      - name: integration tests
        env:
          RAS_RUNNER_TEST_DATA: ./testData
        run: go test -tags integration -v ./...
```

Notes for the implementer:
- If the `hdf5-base` image tag differs in your registry, substitute it. The
  image provides the HDF5 C library and the `go` toolchain is layered in by
  `setup-go`.
- The `fixture` job relies on the `SkipIfNoTestData` helper (Change B) so CI
  stays green before the geometry HDF5 files in Section 5.2 are added.
- `e2e` tests are intentionally **not** wired into CI (they need the RAS binary
  via `RAS_SCRIPT_PATH` and S3/MinIO credentials from `.env-testing`).

---

## 7. Environment prerequisites

- **HDF5 C library** at build/test time — provided by `Dockerfile.hdf5.base`
  (or `Dockerfile.dev` / `hdf5-2.2.0.tar.gz` locally).
- **Go 1.24** (per `go.mod`).
- **Deps:** `github.com/usace-cloud-compute/cc-go-sdk`,
  `github.com/usace-cloud-compute/go-hdf5`,
  `github.com/usace-cloud-compute/filesapi`.
- **`RAS_SCRIPT_PATH`** — path to `run-model.sh` (default `/ras`); required only
  for the `e2e` unsteady-simulation test; the test skips without it.
- **`.env-testing-local`** — `CC_STORE_TYPE=FS`, `CC_FSB_ROOT_PATH`,
  `FFRD_STORE_TYPE=FS`, `FFRD_FSB_ROOT_PATH`, `CC_PAYLOAD_ID`,
  `CC_EVENT_IDENTIFIER`, `CC_EVENT_NUMBER`, `CC_ROOT`. Required for the
  `integration` tier (local-CC plugin-manager runs).
- **`.env-testing`** — S3/MinIO static-credential vars
  (`CC_AWS_*`, `FFRD_AWS_*`, `CC_PAYLOAD_ID`, `CC_MANIFEST_ID`, …). Required
  only for the manual `e2e` tier.

---

## 8. Out of scope / known issues to flag

- `getBreachRows only handles first breach block (silent data loss)` (noted in
  `zz-notes.txt`) — `ras/bFile_test.go` round-trip is written to **catch** this
  (it should fail today if the second breach block is dropped). Fixing the bug
  itself is a separate task.
- `update-boundary-condition-delete-me.go` is marked for deletion (per
  `zz-notes.txt`); it is still registered as `update-boundary-condition` and is
  included in the registry smoke test. If it is deleted, remove it from
  `TestAllActionsRegistered` and from `main.go`'s imports.
- `refline-to-bc` vs `column-to-bc` use **different attribute templates**
  (`source`/`destination` vs `src`/`dest`, and `refline`/`use-remote-reads` vs
  `column_index`/`time_tolerance`). The per-test plans pin each one's exact
  shape; do not "normalize" them.
