Selected and opened:

- **`go-context-and-deadlines`** — directly applies to carrying caller cancellation and deadline through successive requests, response body read and close, one operation-wide budget, and narrower per-request budgets.
- **`go-interfaces-and-composition`** — relevant if the implementation changes who supplies or owns the HTTP client or other dependencies. It may be unnecessary for a local context and timeout change.

I explicitly decline **`go-behavior-tests`** and **`go-test-isolation`** because this is a selection-only probe, with no test work authorized. I also decline **`go-api-contracts`** and **`go-package-boundaries`** because no exported contract or package-boundary change was requested.

I read the workspace `RTK.md` and used `rtk`-prefixed commands to open the two selected skills. No project task files were read; I made no edits, ran no tests, browsed externally, or delegated.