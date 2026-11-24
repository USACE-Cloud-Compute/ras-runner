# Test plan — `actions/utils/copy-inputs-action_test.go`

> **Read `GENERAL.md` first** (tiers, the `MODEL_DIR` override in Section 2
> Change A, skip helpers, the cc-go-sdk construction in Section 4, the
> `.env-testing-local` env in Section 7, test-data spec).
> **Package:** `utils`. **Test file:** `actions/utils/copy-inputs-action_test.go`.

#### `copy-inputs-action_test.go` — **new**
- **Tag:** `integration`
- **Covers:** `CopyInputsAction.Run` (copies every input DS path into `MODEL_DIR`).
- **Setup:** `t.Setenv("MODEL_DIR", tmp)`; a `cc.PluginManager` whose
  `Inputs` data sources have local paths (use `CC_STORE_TYPE=FS`,
  `.env-testing-local`).
- **Assertions:** each `ds.Paths[k]` is copied to `tmp/<basename>`; content
  matches the source; a missing input path returns an error.
