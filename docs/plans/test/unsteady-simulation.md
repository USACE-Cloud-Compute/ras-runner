# Test plan — `actions/run/unsteady-simulation-action_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, the `RAS_SCRIPT_PATH` env in Section 7, test-data spec).
> **Package:** `run`. **Test file:** `actions/run/unsteady-simulation-action_test.go`.

#### `unsteady-simulation-action_test.go` — **replace** (exists; writes to hardcoded `/sim/model`)
- **Tag:** `e2e` (it shells out to the RAS binary) — with a `unit` sub-test for
  the summary-read path that does **not** need the binary.
- **Setup (keep the existing pattern):**
  - `t.Setenv("MODEL_DIR", tmp)` (Change A) instead of hardcoding
    `actions.MODEL_DIR`.
  - `os.CopyFS(tmp, os.DirFS("../../testData/actions/run/unsteady-simulation"))`
    to stage the Muncie model (`Muncie.b04`, `.c02`, `.g04.hdf`, `.x04`,
    `.p04.tmp.hdf`) + `payload`.
  - Decode `cc.Payload`; build `UnsteadySimulationAction{ActionRunnerBase: …}`.
- **Gating:** `skipIfNoEnv(t, "RAS_SCRIPT_PATH")` before the real run — the
  action shells out to `run-model.sh` (env `RAS_SCRIPT_PATH`, default `/ras`).
  Without the RAS binary installed, the e2e sub-test skips.
- **`unit` sub-test (no binary):** call the internal "read the solution summary"
  path against the staged `Muncie.p04.tmp.hdf` and assert
  `Solution` attribute == `"Finished Successfully"`
  (`rasModelSummarySolutionSuccessCriteria`) is detected for a finished run, and
  that an unstable/failed summary would NOT match.
- **e2e sub-test:** `Run()` returns nil; `pm.GetOutputDataSource("output")` and
  the `log` data source are populated; the `rasModelSummaryExtractFields`
  (`Solution`, `Time Stamp Solution Went Unstable`) are present in the posted
  summary.
