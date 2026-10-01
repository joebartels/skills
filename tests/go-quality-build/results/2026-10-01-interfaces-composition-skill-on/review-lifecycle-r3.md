# Independent review: library-lifecycle candidate

Review boundary: original fixture `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-interfaces-and-composition/evals/files/library-lifecycle` versus candidate `/private/tmp/go-quality-build-composition-eval/skill-on-r3/library-lifecycle`. Paths below refer to the candidate unless stated otherwise. This is an introduced-change review, not an audit of unrelated existing code. Read the exact assigned task prompt, both READMEs, both source/test sets, module declarations, candidate example, and the Architecture and Testing review skills/references. Did not read trial reports, evaluation assertions, build skills, baseline reviews, or other trials.

## Architecture & Design — A

Scope: Added asynchronous periodic lifecycle in an embedded library, alongside existing New and synchronous Refresh. Module declares Go 1.22; runtime checks used Go 1.26.5 darwin/arm64.

Coverage: Construction, explicit callback dependency, lifecycle ownership, per-instance run admission, cancellation propagation, join semantics, callback error identity, completion-relative scheduling, one-shot compatibility, and host example. No network/file implementations are present or required; the callback contract supplies cancellation cooperation.

Rationale: No actionable architecture issue confirmed. Explicit per-run controls let the host cancel and join work before releasing resources, and source inspection plus executed tests support the important ownership boundaries. These are appropriate ordinary lifecycle controls; an A+ claim is unnecessary.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [AG1] `refresh.go:44-55,59-76` gives each run its own derived context and completion channel. Stop requests cancellation; Wait observes completion after the synchronous loop returns. The active-callback test exercises propagation and deliberately delayed callback cleanup. Closing done after writing err also provides synchronization for Wait readers.
- [AG2] `refresh.go:10-14,37-53` keeps admission state on each refresher and clears it when its worker exits. There is no library-owned signal handling, process exit, or shared cancellation state. `refresh_test.go:82-112` exercises another instance continuing after the first is stopped.
- [AG3] `refresh.go:79-90` directly sequences each callback and then creates its timer; callbacks in the run cannot overlap and the implementation waits a fresh interval after success. Callback failures return unchanged, while cancellation during the interval returns ctx.Err(). The one-shot implementation remains unchanged at `refresh.go:22`.
- [AG4] `README.md:24-70` is a complete executable host example. It stops both runs before joining either and handles a second-start failure by stopping and joining the first. `example_test.go:25-32` additionally supplies an executable Go example with an output assertion and releases the example resource only after both joins.

Bad

None found.

Suggested changes

None needed.

Limits: Executed the README main program extracted verbatim into a disposable module copy; it compiled and exited successfully. The resource is a placeholder rather than a real closable file/client, but the ordering is explicit and satisfies the requested shutdown demonstration. Did not run the Go 1.22 toolchain itself. No dependency outside the standard library or version-incompatible API was observed. Context-cancel races do not imply forcible callback termination: the supplied contract explicitly requires cooperating callbacks. The architecture assessment does not infer that passing tests establish every timing property.

## Testing — B

Scope: Introduced periodic tests and runnable example, with retained original New/Refresh tests.

Coverage: Inspected all supplied assertions and synchronization. Executed the entire candidate suite with race detection three times, including its example. Ran a targeted scheduling mutation in a separate disposable copy. Reviewed invalid interval and active-run guards by inspection; no dedicated candidate tests exercise those guards. No fuzzing or benchmarks are implicated.

Rationale: One moderate issue: the test named for delay after completion cannot distinguish the promised scheduling from a fixed periodic ticker. This loses regression detection for a normal-use timing behavior, while active cancellation/join, independence, callback error identity, and one-shot compatibility have meaningful exercised assertions. The gap is contained and does not leave the whole lifecycle effectively unverified.

Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [TG1] `refresh_test.go:43-79` uses entered/cleaning/release channels to put Stop inside an active callback and to distinguish cancellation request from completion. It checks the cancellation error and that Wait does not complete before callback cleanup is released.
- [TG2] `refresh_test.go:102-112` observes the second refresher making additional progress after joining the first. Atomic counters avoid introducing races in the test observation.
- [TG3] `refresh_test.go:115-128` checks exact callback error identity and a single call, protecting first-error termination; `refresh_test.go:12-40` retains one-shot context identity/error identity and nil-constructor coverage. The example is actually executed by go test because it has an Output directive.

Bad

- [T1][moderate][introduced] `refresh_test.go:148-158` measures the second start relative to the first start, and requires only the same duration already spent deliberately blocking the first callback. The first callback is held for interval + 20ms before release. Even a scheduler whose next callback starts immediately on release satisfies `second.Sub(first) >= interval + 20ms`. Confirmed mutation: replaced the timer created after each successful completion with one fixed ticker created before the loop. All candidate tests, including `TestPeriodicWaitsIntervalAfterCompletion`, still passed under race detection three times. A fixed ticker can have a pending tick after a long callback, so this violates README.md:5 and :13 while escaping the test. Primary remediation owner: Testing.

Suggested changes

- [T1] Record the release/completion boundary and assert that the second callback starts no sooner than one complete interval after that boundary, allowing a bounded upper deadline for test completion. Retain the long-running first callback so a fixed ticker accumulates an early tick. Verify that the fixed-ticker mutation fails and the current implementation passes. This changes the assertion to observe the stated completion-relative delay.

Limits: Exact successful checks: `env GOCACHE=/private/tmp/review-lifecycle-r3-gocache go test -race -count=3 -timeout=30s ./...` in `/private/tmp/review-lifecycle-r3-check` (PASS; example executed, extracted main compiled); `env GOCACHE=/private/tmp/review-lifecycle-r3-gocache go run ./cmd/readme-example` there (exit 0); `go test -race -count=3 -timeout=20s ./...` in `/private/tmp/review-lifecycle-r3-ticker` (PASS despite deliberate contract violation). Initial candidate/example checks using the default Go build cache failed because the sandbox prevented cache writes, then succeeded with a writable temporary cache. The mutation retained callback sequencing and changed only scheduling to a ticker, isolating T1. No candidate or repository file was modified. Several positive-path channel receives and Wait calls rely on the go test process timeout on regression; no actual deadlock/flakiness was observed. Runtime testing was on Go 1.26.5, not a separate Go 1.22 installation.

## Task deliverables

The implementation supplies and documents Start/Stop/Wait, preserves New and Refresh, tests active stopping, independent instances, and refresh errors, and includes both a runnable README main and a tested two-refresher Go example. The host owns cancellation and resource teardown. The demonstrated defect is in timing-test sensitivity, not an observed failure of the current scheduling implementation.
