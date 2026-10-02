# Poll test implementation

Changed only `poll_test.go` in `/private/tmp/go-testing-author-iu16ahig/worker-timing`. Exact comparisons confirm `poll.go` and `go.mod` retain their initial contents, including the Go 1.22 minimum. No dependencies, production testing hooks, interfaces, or public APIs were added.

The former readiness sleep is replaced by a callback-start event. Existing cancellation and callback-error checks remain, with direct error identity and call counts. New tests cover already-canceled contexts, invalid intervals, three sequential callback starts with two completion-relative interval assertions, callback error identity during cancellation, propagation of the exact caller context, and Poll waiting for active callback cleanup. Recurrence holds the first acknowledged callback for two intervals and samples completion in its final defer; a release gate is not treated as completion. Only the documented timing lower bound is asserted; scheduling delays are permitted.

A real temporary-file fixture is borrowed by the callback and written in deferred cleanup. The test joins Poll before reading the marker, verifies the caller's file remains open, and registers cancellation/join cleanup after the fixture closer so it executes first. If a join stalls, the closer reports that ownership is unresolved instead of claiming a safe close. Independent parallel alpha/beta children own their paths, contexts, counts, and results and do not rely on excluded siblings. A separate concurrent-invocation test acknowledges both active callbacks, cancels one, verifies the other remains active, and then cancels and joins the second.

All launched Poll runs register cleanup before launch. Cleanup cancels, releases owned callback gates where present, and waits for actual Poll completion. Outcomes are published through a closed completion channel, counters use atomics, and assertions stay in the test goroutine. Owned timers are stopped. Event and join waits have five-second diagnostic bounds; Go commands additionally have finite suite and process deadlines.

## Checks actually run

Every Go command used `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local`. `checks.json` preserves command arguments, stdout, stderr and exit codes.

- `go version`: installed `go1.26.5 darwin/arm64`.
- Baseline `go test -v -race -timeout=20s ./...`: passed the original two tests.
- `gofmt -s -w poll_test.go`: passed after writing and after the final concurrent-invocation addition.
- Initial expanded `go test -v -race -timeout=30s ./...` and `go vet ./...`: passed.
- Final `go test -race -shuffle=on -count=25 -timeout=30s ./...`: passed, including every final test.
- Focused `go test -v -race -count=5 -timeout=20s ./...` runs with `-run=^TestIndependentInvocations$/^alpha$`, `-run=^TestIndependentInvocations$/^beta$`, `-run=^TestRecurrence$`, and `-run=^TestConcurrentInvocations$`: each passed.
- Final `go vet ./...`: passed.
- `gofmt -l poll.go poll_test.go`: exit 0 with empty output.
- Exact-text comparisons of `poll.go` and `go.mod` against their initial read: passed.
- `staticcheck` availability: not installed; not run or installed.

## Known limits and remaining behavior risks

Verification used the installed Go 1.26.5 runtime, not an actual Go 1.22 toolchain. The module retains Go 1.22, uses only standard-library APIs available by that version, and passes vet with the module's effective version.

Timing uses real monotonic time and acknowledged callback events rather than a virtual clock. Completion is sampled at the end of callback cleanup immediately before the final defer returns; no public clock seam was introduced. The five-second event bound tolerates late schedules but a severely stalled host can still fail diagnostically. Twenty-millisecond negative-observation windows begin only after callback startup/cleanup acknowledgement; finite observations and shuffled/race repetitions do not prove every possible schedule is covered.

A timed-out join cannot forcibly terminate a noncooperative goroutine. Authored callbacks cooperate with cancellation or have release gates included in cleanup, and all actual runs joined successfully. Arbitrary future regressions that ignore both cancellation and callback completion may require the outer test/process deadline; such a failure must not be described as safe fixture teardown.
