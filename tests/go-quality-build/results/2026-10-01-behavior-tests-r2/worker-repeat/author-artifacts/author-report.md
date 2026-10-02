# Periodic worker implementation

Implemented in `/private/tmp/go-fresh-author-mm5bi5gc/task-7/module` only.

## Changes and contract decisions

- `host.go`: preserved the exact exported `Run` function type; reject nonpositive intervals before either callback; create a child context, wait synchronously for the periodic worker and callback cleanup, cancel that context, invoke release exactly once, and join independent worker/release errors.
- `worker.go`: replace one-shot work with an immediate, sequential callback loop and a fresh timer after each successful callback completion. Parent cancellation stops future cycles, including during the interval.
- `README.md` and the public function comment: explicitly document callback-error precedence during concurrent cancellation. Any callback-returned error, including a callback cancellation error, is preserved; cancellation itself returns nil when the callback returns nil.
- `lifecycle_test.go`: consumer-facing exact-signature compile check and host-level tests for invalid intervals without resource effects; already-canceled context; immediate startup; cancellation between cycles; successful recurrence; full interval after a deliberately slow callback; nonoverlap; failure after two successful cycles; all independent work/release error combinations; cancel-before-release; cleanup-before-release/return while canceled callbacks remain blocked; independent noncomparable errors and joined cancellation/failure; and independent simultaneous instances. The existing `host_test.go` regression remains unchanged.

The worker runs synchronously inside Run; no additional goroutine or public lifecycle API is needed. Tests invoke Run asynchronously to observe host boundaries. Channel waits and cleanup waits have finite two-second deadlines and blocking test callbacks can be unblocked during cleanup.

## Verification

Exact commands, outputs, environments, expected failures, and exit codes are in `checks.json`.

- The new lifecycle tests failed against the original one-shot implementation (exit 1), detecting invalid interval resource effects, a sweep on an already-canceled context, missing child-context cancellation before release, and missing recurrence.
- Initial implementation: `go test -v -race -timeout=45s ./...` passed.
- A temporary mutation replacing the interval timer with a one-nanosecond wait compiled and failed the intended timing assertion: the next cycle began after 33.167 microseconds instead of at least 40 milliseconds. The source was restored immediately.
- Final restored implementation, including the failure-after-success test: `go test -race -shuffle=on -count=30 -timeout=45s ./...` passed (`ok example.com/sweeper 11.140s`).
- Final `go vet ./...` passed under a 30-second subprocess deadline.
- `gofmt -l host.go worker.go host_test.go lifecycle_test.go` returned no paths.
- `staticcheck` was unavailable; it was not installed.

## Limitations and remaining risks

The installed toolchain is Go 1.26.5 on darwin/arm64, with `GOTOOLCHAIN=local`; exact Go 1.22 runtime execution was not available. The module still declares Go 1.22 and uses only standard-library APIs available by that version. Tests exercise real timers with generous bounded synchronization and verify lower timing bounds; extreme scheduler starvation can cause a bounded test timeout. Cooperative cancellation is required: a callback that ignores cancellation or never finishes cleanup will keep Run waiting, consistent with the required release ownership. Nil callbacks and callback/release panics are unspecified and are not assigned new semantics. No real downstream consumer repository was supplied; the external test package checks the preserved representative function type and public behavior. No listeners or network boundaries were needed.

## Guidance selection and isolation

`selection.json` records offered guidance, opened skill/reference paths, relevance decisions, task files read, and changed paths. Work stayed inside the assigned module/report directories. No delegation, commits, dependency/tool installation, or external changes were performed.
