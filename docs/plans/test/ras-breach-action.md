# Test plan — `actions/extract/hdf/ras-breach-action_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, the cc-go-sdk construction
> in Section 4, and the test-data spec — incl. the "To acquire" payload in 5.2).
> **Package:** `hdf`. **Test file:** `actions/extract/hdf/ras-breach-action_test.go`.

#### `ras-breach-action_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** `RasBreachExtractAction.Run` end-to-end.
- **Pattern:** same as above with `testData/actions/extract/ras-breach-extract/payload`.
- **Assertions:** `Run()` returns nil; output data source contains the breach
  records JSON; each record has the `Breach at` elevation and the `SaConn`
  location.
- **Fixture to create:** `testData/actions/extract/ras-breach-extract/payload`.
