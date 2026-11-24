# Test plan — `main_test.go` (root)

> **Read `GENERAL.md` first** (tiers, skip helpers).
> **Package:** `main` (root). **Test file:** `main_test.go`.

#### `main_test.go` — **replace** (currently has a `TestMain` that calls `main()` and a large commented-out legacy block)
- **Tag:** `unit`
- **Remove:** the `TestMain(m)` that calls `main()` (it requires CC env vars and
  does not assert anything) and the entire commented-out block referencing
  `Duwamish_17110013.p01.hdf`.
- **New test — `TestAllActionsRegistered`:** assert that every action
  registered via `init()` is present in the registry. The expected set (from
  the `cc.ActionRegistry.RegisterAction` calls found in the codebase):
  - `ras-extract`
  - `ras-breach-extract`
  - `unsteady-simulation`
  - `create-ras-tmp`
  - `update-breach-bfile`
  - `update-outlet-ts-bfile`
  - `copy-inputs`
  - `post-outputs`
  - `refline-to-boundary-condition`
  - `column-to-boundary-condition`
  - `hdf-to-hdf`
  - `update-bfile-skip-dss`
  - `update-boundary-condition`
  This test also acts as a guard against an action file being dropped from the
  build (an `init()` that doesn't run would make the name missing).
