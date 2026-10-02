# Periodic worker implementation

Implemented the requested lifecycle in the isolated `example.com/sweeper` module. `Run` keeps its exact exported signature; `go.mod` remains Go 1.22 with no added dependencies.

## Changes

- `host.go`: reject nonpositive intervals before effects; create a worker-owned child context; start and wait for the periodic worker; cancel that context before invoking release exactly once; join callback and release errors.
- `worker.go`: replace the one-shot helper with serial periodic sweeps. Check cancellation before each callback, invoke the first immediately, and create a fresh interval timer only after successful callback completion. Cancellation stops the timer and ends normally.
- `README.md` and the exported `Run` comment: clarify that callback errors remain observable when they coincide with cancellation. Cancellation by itself returns nil; a callback that returns an independent error during cancellation retains that error. `Run` waits for release itself before returning.
- `host_test.go`: exercise the exported API from an external test package and compile its original function type. Tests cover zero/negative intervals without effects; already-canceled startup; immediate startup and prompt cancellation during a long interval; three serial callbacks with a long first callback and full completion-relative delays; failure after two successful cycles; held cancellation cleanup before release/return; exactly-once release; all four worker/release error combinations; owned context cancellation without canceling the caller; held release before return; and independent simultaneous runs.

Callbacks and release remain function dependencies. No interface, constructor, process signal handler, external resource, or timer abstraction was needed.

## Checks actually run

Final checks were executed through `rtk proxy`, with finite 60-second subprocess deadlines. Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`.

- `gofmt -w host.go worker.go host_test.go`: exit 0.
- `go test -v -race -count=1 -timeout=45s ./...`: exit 0.
- `go test -race -count=10 -shuffle=on -timeout=45s ./...`: exit 0; all ten repetitions passed with the race detector.
- `go vet ./...`: exit 0.
- `gofmt -l host.go worker.go host_test.go`: exit 0, no output.
- Local toolchain: `go1.26.5 darwin/arm64`. The supported module directive remains `go 1.22` and the implementation/tests use APIs available by Go 1.22.

Before the final checks, temporary source counterexamples demonstrated the tests' signal, and were restored in `finally` blocks:

- Original one-shot implementation, focused regression tests: expected exit 1 with concrete wrong-effect assertions (invalid intervals swept/released; already-canceled context swept; callback-failure release saw an uncanceled context).
- Timer started before callback completion, recurrence test: expected exit 1 because cycle 2 began approximately 6 microseconds after completion instead of waiting at least 60 milliseconds. The case compiled and failed the intended timing assertion.

An initial race suite and vet check also passed before the last tests were added; their original combined tool outputs are preserved in `initial-checks.json`. Final separate stdout/stderr, exit codes, exact argv, and process deadlines are in `checks.json`. `verify.py` records the final sequence and counterexamples. Staticcheck was unavailable and was not installed, as required by the dispatch.

## Limits and remaining behavior risks

The available local Go toolchain was newer than 1.22; no Go 1.22 binary was installed or downloaded. Module compatibility was checked through the unchanged directive, Go 1.22-compatible APIs, compilation, and vet rather than an actual 1.22 runtime.

Timers use the real clock. Tests require the promised minimum wait and use controlled callback gates with generous finite observation deadlines, but do not promise exact upper scheduling latency. Repeated shuffled race runs passed on this host.

Cancellation is cooperative. A callback that ignores its context or never finishes its cleanup, or a release callback that never returns, keeps `Run` waiting. This follows the requested ownership rule; releasing resources while such work remains active would be unsafe. Panic handling and nil callback/context behavior are not specified or added.

No repository, downstream integration, or external resource outside the authorized isolated task was inspected or modified. The offered package-boundary skill was not opened because this extends the package's existing worker responsibility.
