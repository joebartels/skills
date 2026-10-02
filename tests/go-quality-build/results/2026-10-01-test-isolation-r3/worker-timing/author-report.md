# Poll test implementation

Changed only `module/poll_test.go`. `poll.go`, `go.mod`, the public API, Go 1.22 minimum, and standard-library dependency policy remain unchanged.

Replaced the cancellation test's 5 ms readiness sleep with a callback-start event, checked caller-context identity, and retained normal cancellation and callback-error checks. Callback errors now require exact identity, including when the callback cancels its context before returning its error.

Added already-canceled and zero/negative interval cases, plus three-cycle recurrence coverage. The recurrence test holds a known active callback past an interval, checks sequential execution, and measures later starts from the preceding callback's final cleanup action. It enforces only the documented lower timing bound; late scheduling is permitted within diagnostic deadlines.

Added a real temporary-file fixture. The callback performs work, observes cancellation, enters blocked cleanup, writes during cleanup, and finishes before Poll returns. The caller can still use the file after Poll completes. Registered cleanup releases gates, cancels, joins both callback and Poll, then closes the file; a later cleanup observation verifies closure and exact persisted contents. Join failures report a stall and retain the open resource instead of claiming a timed-out borrower was stopped.

Added independently selectable parallel `TestIndependentInvocations/callback_error` and `/cancellation` cases. They start concurrent Poll calls with separate contexts and events, verify recurrence can progress while another callback is blocked, and verify one invocation's error or cancellation does not complete the other invocation. Each case owns its data and joins its work, so selecting one named child does not depend on another child running.

## Actual checks

Every recorded check ran in `/private/tmp/go-fresh-author-8hjrllxp/task-4/module` with `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Commands use `rtk proxy`; `report/run-check.py` records command arrays, cwd, the explicit non-secret environment, separate stdout/stderr, exit codes, and durations in `checks.json`. Its process deadline is 120 seconds; Go test commands also have 20- or 30-second deadlines.

- Baseline `go test -timeout=20s ./...`: passed before edits.
- `gofmt -w poll_test.go` and final `gofmt -l poll_test.go`: passed; no formatting differences remain.
- `go test -v -timeout=20s ./...`: all eight top-level tests and all named children passed.
- `go vet ./...`: passed.
- `go test -race -shuffle=on -count=50 -timeout=30s ./...`: passed (6.517 seconds reported by Go).
- Focused race runs, each repeated 20 times, passed for `TestRecurrence`, `TestFixtureTeardown`, both independent-invocation children, and `TestInvalidInterval/negative`.
- Toolchain: `go1.26.5 darwin/arm64`.
- `command -v staticcheck`: exited 1, so staticcheck was not available and was not run. No tools or dependencies were installed.

The check-log inspection found all 13 check records present. Its initial blanket assertion that every command exited 0 failed because it included the expected unavailable-staticcheck probe; the corrected integrity check distinguishes that probe from successful verification. No Go test or vet check failed.

## Limits and remaining risks

The module minimum remains Go 1.22, and the added helpers use APIs available by that version. Execution used the available Go 1.26.5 binary, so an actual Go 1.22 runtime was not exercised. No listeners, external services, or network substitutes were used; the exercised boundary is synchronous Poll callbacks, real timers/contexts, goroutines, and local temporary-file ownership.

Real timer checks allow scheduler lateness and use three-second per-event/join safety bounds. The completion timestamp is recorded immediately before the callback's final buffered send/return, a conservative lower bound on actual return. The 25 ms blocked-cleanup observation and race/shuffle repetitions do not prove correctness under every possible schedule. Cleanup paths are registered before workers start and release gates on early assertions, but deliberate failing-child/process fault injection was not performed. Timeouts diagnose non-cooperation; they cannot forcibly terminate a goroutine.

Selection decisions and exact opened guidance/reference paths are in `selection.json`. No delegation, commits, external changes, repository inspection beyond the authorized dispatch, or other-author inspection occurred.
