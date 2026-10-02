# Poll test work

Only `module/poll_test.go` changed in the isolated module. `Poll`, its public API, `go.mod`'s Go 1.22 minimum and standard-library dependency policy are unchanged.

## Changes

- Replaced the cancellation test's readiness sleep with an observed first-callback event; retained cancellation and callback-error checks and strengthened original error identity checking.
- Added canceled-before-start and nonpositive-interval cases with invocation-count assertions.
- Added three-cycle recurrence coverage. The first callback stays blocked beyond one interval; callbacks cannot overlap, and later starts must be at least the interval after each final callback completion event. That event occurs immediately before actual return and supplies a conservative lower bound. Callback release gates are not used as completion timestamps.
- Added callback error identity during cancellation, exact caller-context propagation, and cancellation while callback cleanup is blocked. Poll must remain active until callback cleanup is released and its real file write completes.
- Added actual temporary-file fixture ownership checks: callbacks borrow an open file, workers cancel and join before close, and the owning subtest closes the file and removes its directory. A separate case returns while a callback is still active, exercising the cleanup path itself.
- Added overlapping independent Poll calls: canceling one leaves the other active with an uncanceled context, and the second returns its own callback error.
- Test goroutines send synchronized results. Test assertions run on the test goroutine. Startup, cleanup and completion observations have five-second diagnostic bounds. Cleanup is installed before work starts; blocked cleanup gates are released before cancellation/join, including early assertion exits. Independent tests and interval cases run in parallel with local state.

## Actual verification

Commands and results are preserved in `checks.json`; `run_check.py` records separate stdout/stderr and limits each verification process to 60 seconds. Go test runs additionally have a 30-second suite deadline (the initial unchanged-source baseline used 10 seconds). All Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`, routed through `rtk proxy`.

- Original two-test baseline passed before edits.
- `gofmt -w poll_test.go` succeeded; final `gofmt -l poll_test.go` produced no output.
- Full suite: `go test -count=1 -timeout=30s ./...` passed.
- Concurrency/reliability: `go test -race -count=20 -shuffle=on -timeout=30s ./...` passed.
- Focused tests each passed three runs: cancellation, completion-relative recurrence, independent invocations, zero interval alone, fixture callback cleanup alone, and fixture cleanup-on-return's owned child alone. Selection does not depend on excluded siblings.
- `go vet ./...` passed.
- Installed toolchain: `go version go1.26.5 darwin/arm64`.
- Staticcheck availability check returned exit 1; staticcheck was not run or installed.

## Limits and remaining risks

The tests exercise actual in-process contexts, goroutines, Go timers and local temporary files. No listeners, external services, process-global mutations or substitute network boundaries are involved. No sandbox rerun or dependency/tool installation was needed.

Go 1.22-compatible APIs and module language semantics were preserved, but the available Go 1.26.5 toolchain was used; an actual Go 1.22 runtime was not exercised. Real time is used only for the documented recurrence bound and post-readiness observation windows; late scheduling is permitted up to diagnostic safety bounds. Repeated shuffled race runs cover observed schedules and do not prove every possible schedule free of races or flakes. A test timeout reports a stalled worker rather than forcibly terminating an uncooperative goroutine. Fixture close is refused if worker completion was not observed; a persistently noncooperative regression ultimately requires the finite test/process deadline to contain it.

No implementation issue was found. No remaining requested work is known.
