# Test plan — `actions/utils/post-outputs-action_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, the cc-go-sdk construction in Section 4, the
> `.env-testing-local` env in Section 7, test-data spec).
> **Package:** `utils`. **Test file:** `actions/utils/post-outputs-action_test.go`.

#### `post-outputs-action_test.go` — **new**
- **Tag:** `integration`
- **Covers:** `PostOutputsAction.Run` / `postOutputFiles`.
- **Setup:** `t.Setenv("MODEL_DIR", tmp)`; attributes `modelPrefix` + `plan`
  (e.g. `Muncie` + `04`); output data sources including the two **reserved**
  names that must be skipped: `rasoutput` and `<modelPrefix>.p<plan>.tmp.hdf`
  (i.e. `Muncie.p04.tmp.hdf`).
- **Assertions:**
  - The reserved `rasoutput` DS is skipped (no copy).
  - The reserved `Muncie.p04.tmp.hdf` DS is skipped.
  - All other output DS files are copied out to the output store.
