# Test plan — `ras/structuresAttr_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec, and the
> "To acquire (missing)" geometry-HDF5 fixtures in Section 5.2).
> **Package:** `ras`. **Test file:** `ras/structuresAttr_test.go`.

#### `ras/structuresAttr_test.go` — **replace** (currently exists, references a missing file)
- **Tag:** `fixture`
- **Covers:** `ReadSNetIDToNameFromGeoHDF` and the version dispatch.
- **Fixtures:** TWO geometry files the user must supply (see Section 4):
  - one **pre-6.6** geometry HDF5 (e.g. `Duwamish_17110013.g01.hdf`) whose root
    group attribute `File Version` does **not** contain "6.6";
  - one **6.6** geometry HDF5 whose `File Version` **does** contain "6.6".
- **Assertions:**
  - For the pre-6.6 file: the returned map equals the expected SNetID→name
    mapping (the existing test asserts `actual["Highway 120"] == 2` and
    `len(actual) == 11` for `Duwamish_17110013.g01.hdf` — keep those two
    assertions for that specific file once it is present).
  - For the 6.6 file: the map is non-empty and uses the `structuresAttr66`
    field layout (assert a known SNetID maps to a non-empty name).
  - **Version dispatch:** the code selects between `structuresAttr` and
    `structuresAttr66` based on whether the root `File Version` attribute
    contains "6.6". Assert that feeding the pre-6.6 file does not error and
    vice-versa, i.e. the dispatch actually reads the attribute and picks the
    right struct (a regression here would silently map the wrong columns).
