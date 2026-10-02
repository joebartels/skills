# Periodic host implementation

Implemented in `/private/tmp/go-testing-author-7r8p3riu/worker-host`.

- `host.go`: rejects nonpositive intervals before starting work or releasing resources; owns a child context and worker result channel; requests cancellation and joins callback completion before releasing exactly once; preserves worker and release errors with `errors.Join`.
- `worker.go`: runs immediately unless already canceled, executes callbacks serially, and creates each interval timer after successful callback completion. Cancellation ends future cycles normally; callback errors end work with their identity preserved.
- `README.md` and Run's Go documentation: clarify that a callback's returned error remains inspectable even if cancellation happened during that callback, including a returned `context.Canceled`. Cancellation itself contributes no error when the callback returns nil.
- `host_test.go`: external-package tests check Run's exact function type, immediate startup with an hour-long interval, actual recurrence, a full interval after a callback held longer than the interval, absence of overlap, cancellation during blocked callback cleanup, release ordering and exact count, pre-canceled parents, cancellation between cycles, immediate and later callback failure, joined error identity/type, canceled callback outcomes, nonpositive intervals without either callback, and independent simultaneous hosts. Synchronization waits and host cleanup joins have two-second limits. Every test process has a finite deadline.

## Verification actually run

All commands used `rtk proxy`; Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Full stdout, stderr, command arguments, elapsed time, and exit codes are preserved in `checks.json`.

- `go version`: Go 1.26.5 darwin/arm64, exit 0.
- `gofmt -s -w host_test.go`: exit 0 before the initial test run.
- Before implementation, `go test -v -timeout=20s ./...`: exit 1. The new tests exposed missing recurrence, execution on an already-canceled parent, resource callbacks on invalid intervals, and failure-path release without canceling a child context.
- `gofmt -s -w host.go worker.go host_test.go`: exit 0.
- `go test -v -race -count=1 -timeout=20s ./...`: exit 0.
- `go test -race -count=20 -timeout=20s ./...`: exit 0.
- `go vet ./...`: exit 0.
- `gofmt -l host.go worker.go host_test.go`: exit 0 with empty stdout.
- Staticcheck availability checked through PATH: unavailable. No tool was installed.

Verification subprocesses were also bounded by a 60-second process deadline.

## Decisions, limitations, and remaining behavior risks

The supported module version remains Go 1.22 and all dependencies remain in the standard library. The available local executable was Go 1.26.5; an actual Go 1.22 toolchain run was not performed. No downstream repository was available: public compatibility was checked through an external test package, the exact function-type assignment, unchanged signature, and error identity/type assertions.

Real timers are used in the cadence test, so extreme scheduler delays can exceed its two-second observation limit. Twenty repeated race runs passed. The assertions check a minimum interval rather than an exact timing window.

Run intentionally waits for callback cleanup without a shutdown timeout. A callback that ignores its context or never finishes can keep Run blocked, and its resource cannot safely be released. Panic recovery and nil callbacks are outside the documented task contract.

No dependencies, delegated agents, commits, or external application changes were made. Only the authorized dispatch file, isolated module files, and the three relevant offered guidance files were inspected.
