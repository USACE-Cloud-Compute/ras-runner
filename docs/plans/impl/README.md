# Implementation / requirements plans

One file per **feature, refactor, or API change** — a *working handoff*, not a
durable spec. These are distinct from the test plans in `../test/`:

| | `../test/` (test plans) | `impl/` (these) |
|---|---|---|
| **Lifetime** | Durable — re-run every CI cycle, tracked as code changes | **Ephemeral** — consumed, then done |
| **Volatility** | Stable | High (changes as implementation proceeds) |
| **Completion signal** | Tests pass in CI | Work merged → **archive the file** |
| **Audience** | An agent writing one `_test.go` | An agent/human implementing a feature or change |

## Conventions

1. **One file per unit of work.** Name it `<feature-slug>.md`, optionally
   date-prefixed (`2026-11-<slug>.md`) so stale plans sort to the top.
2. **Archive when merged.** When the work is complete and merged, move the file
   into `impl/archive/` (create it on first use). Do **not** delete it — keep it
   as a record of the decision. An archived file is not live.
3. **Link, don't merge, test plans.** If a plan includes tests, keep the test
   spec in `../test/<name>.md` (durable) and *link* to it from here. Never
   duplicate the test design into an impl plan — that's two copies that drift.
4. **State the "do-not" scope.** Every impl plan should explicitly say what is
   *out* of scope (mirroring the testability-refactor note in `../test/GENERAL.md`
   Section 2), so an agent doesn't expand scope.
5. **Record the why.** A short "context / motivation" section at the top, so the
   archive is useful later without re-deriving the rationale.

## Index

_No impl plans yet._ When the first one lands, list it here:

| File | Status | Summary |
|------|--------|---------|
| — | — | — |
