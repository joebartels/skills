# Poll test author report

Updated `/private/tmp/go-testing-author-y6_tmx1p/worker-timing/poll_test.go`. `poll.go`, the exported API, standard-library-only dependencies, and the `go 1.22` minimum remain unchanged.

Replaced the cancellation test's guessed 5 ms readiness sleep with a callback signal. Preserved its normal-cancellation and callback-error checks, strengthening error coverage to require original error identity and a single invocation. Added pre-canceled contexts, invalid intervals, exact caller-context propagation, and callback errors returned after observing cancellation.

Recurrence testing blocks the first callback past two intervals, rejects overlapping callbacks, and checks two successive starts against preceding callback completion timestamps. An active-callback test observes cancellation before holding deferred cleanup behind an explicit gate, checks Poll stays blocked, then verifies cleanup completed before return. Its caller-owned file remains usable until fixture teardown. Cleanup registrations release test gates, cancel and join Poll, then close the fixture, including assertion-failure paths. Independent invocation coverage runs two distinct contexts together: one remains in cancellation cleanup while the other recurs and completes with its own error. Tests run in parallel with no shared mutable fixture state; background callbacks report observations through channels and atomics.

## Verification actually run

All Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`; shell commands used `rtk proxy`. Local toolchain: Go 1.26.5 on darwin/arm64.

- Baseline: `go test -timeout=15s ./...` passed before edits.
- Formatting: `gofmt -s -w .` succeeded; `gofmt -d .` produced no diff.
- `go vet ./...` passed.
- `go test -v -race -cover -timeout=15s ./...` passed, with 100.0% Poll statement coverage.
- `go test -race -count=50 -shuffle=on -cpu=1,2,4 -timeout=45s ./...` passed: 150 iterations of each top-level test across the three CPU settings, with a 90-second outer process deadline.

`checks.json` preserves commands, configured deadlines, stdout, stderr, exit codes, elapsed times, and the Go environment. Formatting and vet also had finite outer process deadlines.

## Limitations and remaining behavior risks

Staticcheck was unavailable and was not installed, as the task prohibits tool installation. The Go 1.22 declaration and compatible language/library usage are retained, but an actual Go 1.22 toolchain was not available for execution.

The interval contract is measured with real monotonic timestamps, without a clock framework or exported hooks. Timers measure the deliberately long callback and bounded cleanup observation; they do not establish readiness. Two-second event deadlines can fail on an exceptionally stalled machine. Bounded negative observations cannot prove absence for every possible scheduler interleaving. Callback cooperation with context cancellation remains an existing requirement; these fixtures cooperate and release every test-owned cleanup gate.

No production behavior risks were introduced because the implementation was unchanged. No additional work is required for this requested test scope. Guidance selection and actual opened guidance paths are recorded in `selection.json`.
