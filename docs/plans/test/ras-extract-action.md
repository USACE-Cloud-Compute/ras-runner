# Test plan — `actions/extract/hdf/ras-extract-action_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, the cc-go-sdk construction
> in Section 4, and the test-data spec — incl. the "To acquire" payload in 5.2).
> **Package:** `hdf`. **Test file:** `actions/extract/hdf/ras-extract-action_test.go`.

#### `ras-extract-action_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** `RasExtractAction.Run` end-to-end (the action, not just `DataExtract`).
- **Pattern:** decode a `cc.Payload` JSON from
  `testData/actions/extract/ras-extract/payload` (create this fixture — see
  Section 4), construct
  `RasExtractAction{ActionRunnerBase: cc.ActionRunnerBase{ActionName: "ras-extract", Action: payload.Actions[0], PluginManager: &cc.PluginManager{Payload: payload}}}`,
  call `Run()`.
- **Assertions:** `Run()` returns nil; the output data source (via
  `pm.GetOutputDataSource`) contains the extracted JSON.
- **Fixture to create:** `testData/actions/extract/ras-extract/payload` — a
  `cc.Payload` JSON with one action of type `ras-extract` whose attributes name
  the `Duwamish_17110013.p01.hdf` input source and the extraction parameters
  (group path, colnames, data type, writer type).
