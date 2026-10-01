# Lifecycle implementation trial

## Lifecycle decision

Added `Refresher.StartPeriodic(ctx, interval) (*PeriodicRun, error)` with
`PeriodicRun.Stop()` and `PeriodicRun.Wait() error`. The first callback runs
immediately, and the next callback starts one full interval after the previous
successful callback completes. A callback error ends the run and is returned
unchanged. `Stop` cancels the callback context and is idempotent; `Wait` blocks
until callback work has returned and reports `context.Canceled` for explicit or
parent-context cancellation. Hosts should stop and wait for each run before
releasing resources captured by its callback. Different `Refresher` instances
have independent runs; one instance rejects a second concurrent periodic run.
The existing synchronous `Refresh` API remains unchanged.

## Files changed

- `refresh.go`: periodic run API, per-instance active-run guard, cancellation,
  completion wait, and error propagation.
- `refresh_test.go`: active-callback stop/wait, instance independence, refresh
  error propagation, invalid interval, duplicate run, and existing one-shot
  coverage.
- `README.md`: lifecycle contract and two-refresher shutdown snippet.
- `example_test.go`: runnable orderly-shutdown example for two refreshers.

## Checks and results

- `gofmt -w refresh.go refresh_test.go example_test.go` — completed.
- `go test ./...` — passed (`ok example.com/library-lifecycle`).
- `go vet ./...` — passed (no diagnostics).

## Limitations

The callback must honor context cancellation for `Stop` and `Wait` to complete.
The implementation does not recover callback panics. Periodic runs are serial
within one `Refresher`; starting another while one is active returns an error.
