# Test plan — `actions/utils/blockfile_test.go`

> **Read `GENERAL.md` first** (tiers, skip helpers, test-data spec).
> **Package:** `utils`. **Test file:** `actions/utils/blockfile_test.go`.

#### `blockfile_test.go` — **new**
- **Tag:** `unit` (pure JSON parsing)
- **Covers:** `NewBlockFile` / `BlockFile` in `blockfile.go`.
- **Fixtures:** a small in-memory (or temp-file) JSON array of `RealzBlock`:
  `[{"realization_index":1,"block_index":1,"block_event_count":10,"block_event_start":0,"block_event_end":9,"important_events":[2,5]}]`
  plus an important-events file.
- **Assertions:**
  - `NewBlockFile` decodes the blocks; `BlockEventCount` == `End-Start+1`.
  - `StartRealz`/`EndRealz` filtering returns only the in-range blocks.
  - `ProcessImpEventsPerBlock` with `ImportantEvents` set yields the expected
    per-block event lists.
  - Malformed JSON returns an error (not a panic).
