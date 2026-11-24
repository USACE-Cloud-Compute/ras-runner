# Test plan — `actions/extract/hdf/ras-extractor_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, the `writerAccumulator.Reset()`
> note in Section 2 Change C, and the test-data spec).
> **Package:** `hdf`. **Test file:** `actions/extract/hdf/ras-extractor_test.go`.

#### `ras-extractor_test.go` — **replace** (exists; `//@TODO made DataExtract generic and broke all of the tests`)
- **Tag:** `fixture`
- **Covers:** the generic `DataExtract[T]` in `ras-extractor.go`.
- **Fixtures:** `Duwamish_17110013.p01.hdf` (primary); `clear-creek.p01.hdf`
  and `lewisville.p01.hdf` as secondary if available.
- **Call `writerAccumulator.Reset()` at the top of every test (Change C).**
- **Test cases (table-driven over `RasExtractInput`):**
  - **Dataset extraction, `float32`:**
    `GroupPath: "/Results/Unsteady/Output/Output Blocks/Base Output/Unsteady Time Series/Boundary Conditions"`,
    `Colnames: ["stage","flow"]`, `DataType: reflect.Float32`, `WriteSummary: true`,
    `WriterType: ConsoleWriter` — assert no error and that the console/JSON
    output contains the expected dataset names.
  - **`Postprocess: ["max"]`:** same group, assert the per-dataset max value is
    present in the writer output.
  - **`ExcludePattern`:** `ExcludePattern: "Flow per Face|Stage per Face|Flow per Cell"` —
    assert those dataset names are absent from the output.
  - **Other element types:** `float64`, `int32`, `string` via the
    `dataTypeMap` in `ras-extract-action.go` (`"float32"|"float64"|"int32"|"int64"|"string"`).
    For each, pick a real dataset in the fixture of that type and assert the
    extracted values decode correctly.
  - **JSON writer:** `WriterType: JsonWriter` — capture the JSON bytes and
    assert the shape matches `ras-extractor-json-output-file-specification.md`
    (the file at `actions/extract/hdf/ras-extractor-json-output-file-specification.md`
    is the authoritative spec; read it and assert the top-level keys).
  - **Group vs dataset vs attribute modes** (the three modes documented in
    `RasExtractAction.Run`): one test per mode.
