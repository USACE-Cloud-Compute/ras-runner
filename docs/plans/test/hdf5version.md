# Test plan — `ras/hdf5version_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec, and the
> "To acquire (missing)" geometry-HDF5 fixtures in Section 5.2).
> **Package:** `ras`. **Test file:** `ras/hdf5version_test.go`.

#### `ras/hdf5version_test.go` — **new**
- **Tag:** `fixture`
- **Covers:** the `File Version` root-attribute read used for dispatch (the
  function in `structuresAttr.go` that inspects the root group attribute).
- **Assertions:**
  - Given the pre-6.6 geometry file, the version string does not contain "6.6".
  - Given the 6.6 geometry file, it does.
  - `Hdf5Float32` JSON: a NaN marshals to `null`; a normal value marshals to its
    float representation (pure → can be `unit`).
