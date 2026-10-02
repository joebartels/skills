# Periodic sweeper author report

Completed the standalone implementation in the authorized isolated module. Only
that module, the offered relevant writing guidance, the dispatch, and this report
directory were read. No delegation, dependency/tool installation, commits,
listeners, or external changes were performed.

## Changes and contract decisions

- `host.go`: validate interval before starting work or calling release; derive an
  owned worker context; start one worker; cancel and join it on parent cancellation
  or worker failure; call release exactly once after worker completion; join errors.
- `worker.go`: replace the one-shot helper with a serial loop. Check cancellation
  before each callback; invoke the first callback immediately; start each interval
  timer after a successful callback returns; make the interval wait cancelable.
- `host_test.go`: external consumer tests cover the exact exported function type,
  zero and negative intervals with no callbacks, pre-canceled starts, an immediate
  start with an hour-long interval, cancellation during an interval wait, successful
  second and third cycles, a callback held longer than its interval, no overlap,
  cancellation during an active callback with held cleanup, worker/release errors
  independently and together, cancellation coincident with an independent or joined
  callback error, context cancellation before release on failure, held release
  completion before return, later failure after successful cycles, a noncomparable
  error carrying inspectable type/data/cause, and independently controlled instances.
- `README.md`: clarify that callback errors take precedence over cancellation when
  returned by the callback, and shutdown waits for cooperative callback cleanup.

The supported `Run` function signature, `example.com/sweeper` module path, Go 1.22
directive, and standard-library dependency policy are preserved. Ordinary parent
cancellation with a nil callback result yields a nil worker result. Every nonnil
callback error is preserved, including an error returned after cancellation. This
follows the request's callback-error identity promise and prevents cancellation from
hiding an independent failure. No exported sentinel, interface, constructor,
options API, clock injection, process control, or dependency was added.

## Verification and assertion evidence

All Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and
`GOTOOLCHAIN=local`. Subprocess deadlines were 60 seconds; test deadlines were
20 or 30 seconds. Channel observations and test cleanup joins have two-second
bounds, and callback gates are unblocked even on assertion failure.

- Local toolchain: `go version go1.26.5 darwin/arm64`.
- New tests against the unchanged one-shot production implementation compiled and
  failed with meaningful assertions: invalid intervals performed work and release,
  a pre-canceled context still swept, recurrence stopped after one cycle, failure
  released without canceling the callback context, and a third-cycle error never
  occurred. This expected failing run is retained in `checks.json`.
- Full verbose lifecycle suite passed after implementation.
- `go vet ./...` passed.
- `go test -race -shuffle=on -count=10 -timeout=30s ./...` passed, including a final
  run after all tests were added and temporary mutation restored (4.054 seconds).
- Final `gofmt -d host.go worker.go host_test.go` produced no diff.
- `go list -m all` reported only `example.com/sweeper`.
- A temporary timing mutation started the delay before the callback. The focused
  recurrence test compiled and failed on the expected assertion: cycle two began
  6.875 microseconds after completion rather than waiting 40 milliseconds. The
  original source was restored in `finally`, verified byte-for-byte, and then the
  final complete artifact passed race/repeat/shuffle and vet again.
- `staticcheck` was not installed; availability was checked with `shutil.which`
  inside the RTK-wrapped verification runner. No tool was installed.

Exact verification argument arrays, cwd, non-secret environment overrides, output,
exit codes, deadlines, baseline failures, mutation failure, and restoration evidence
are preserved in `checks.json`. The initial dispatch read used shell `cat` before
its RTK prefix requirement was known; all subsequent shell commands used `rtk proxy`.

## Limits and remaining risks

Verification exercises the public library/host boundary with real timers and controlled
callbacks; no network/process-host integration is claimed. The module remains Go 1.22
and uses APIs available there, but no Go 1.22 binary was available or installed:
actual execution used the local Go 1.26.5 runtime. Staticcheck remains unrun.

Shutdown deliberately has no forced teardown deadline: Run waits for the active
callback and release to finish. A callback that ignores cancellation or a blocked
release can hold Run indefinitely, as required by the ownership contract. Timing
checks observe a minimum delay and bounded progress on this host; they do not promise
exact scheduling latency. No additional work is pending in the requested scope.

## Final artifacts

- `/private/tmp/go-fresh-author-mm5bi5gc/task-4/module/host.go` — SHA-256 `93dce57e7a5fc314880f5089c5002d781906aac43968498b5f2254120c07cfd5`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-4/module/worker.go` — SHA-256 `bd029079e7dedff4e55ddd23fe7b2046ddf3a5de768ae8742c676a49a08c39b2`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-4/module/host_test.go` — SHA-256 `b2e3f7549b964f8b8c6f1ec15d525d633fa0ea394c4a3666899b5fff4a300bc8`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-4/module/README.md` — SHA-256 `084475df3fdcbdddc7a3bf49e7175c89714b65c1f1a4453658b29f70e28e83a1`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-4/module/go.mod` — SHA-256 `8c2eee5cd1aaa88bb71bf34122c45eed6198fc4356555ff422c767c1dd531307`
