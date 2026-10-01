# Independent lifecycle changeset review

The supplied original fixture and candidate directories define the base/head boundary; no Git revision comparison is implied. All candidate file references below are relative to `/private/tmp/go-quality-build-composition-eval/skill-on-r1/library-lifecycle`. Reviewed the exact supplied task prompt, original/candidate README, module, implementation, tests, and executable example. Used the Architecture & Design and Testing skills and their decision references. No trial report, expected assertions, evaluation configuration, build skill, other trial, or prior review was read.

## Architecture & Design — B
Scope: Introduced periodic API in the embedded `refreshkit` library, compared with the supplied original one-shot fixture; module declares Go 1.22.
Coverage: Public API, callback dependency, cancellation/error propagation, host-owned goroutines and joining, instance isolation, serial successful iterations, input validation, and two-refresher shutdown documentation/example. No external service boundary exists in this fixture.
Rationale: One moderate introduced inconsistency in the public cancellation result. The implementation preserves host control and resource joining; its narrower error-precedence mismatch does not prevent orderly teardown and therefore is moderate rather than major.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [AG1] `refresh.go:27-50` provides a blocking `RunPeriodic(ctx, interval)` call. Cancellation reaches the same callback context and the synchronous invocation must return before the run returns. The host can join its goroutine before releasing callback resources. `TestRunPeriodicStopsAfterActiveRefreshReturns` exercised cancellation and delayed callback cleanup successfully.
- [AG2] `refresh.go:31-49` serializes callbacks and allocates the timer after each successful callback. There is no shared instance or process cancellation state, no signal registration, and no hidden background goroutine. The supplied race runs exercised two independent runs and successful repeated callbacks.
- [AG3] `README.md:21-41` and executable `refresh_test.go:128-145` cancel both runs before consuming both results. This exposes shutdown sequencing to the host and avoids abandoning the second join when one run fails. Existing `New` and `Refresh` remain behaviorally intact.

Bad

- [F1][moderate][introduced] The cancellation result promised by `refresh.go:22-23` and `README.md:15-16` is not what the callback-error branch returns (`refresh.go:35-36`). A callback can observe cancellation and return its own error, including a wrapped cancellation error or a cleanup failure; `RunPeriodic` returns that error unchanged instead of `ctx.Err()`. A deterministic probe canceled the run during its callback, returned `errors.New("callback cleanup failed")`, and got that error rather than the documented `context.Canceled`. Hosts cannot rely on the documented result to classify shutdown. Primary remediation owner: Architecture/API contract.

Suggested changes

- [F1] Choose and document error precedence for cancellation concurrent with callback failure. Either check `ctx.Err()` after callback completion before choosing the returned error, or explicitly document that non-nil callback errors take precedence. Verify both a nil-returning canceled callback and one that returns an independent error; preserve first-error behavior when there is no cancellation.

Limits: Inspected the complete supplied fixture, not a larger host application. The synchronous structure establishes joining and no overlap for each run; concurrent starts are explicitly outside the contract. Tests ran with Go 1.26.5 on darwin/arm64, not an actual Go 1.22 binary. No unrelated architecture issue was inferred from the small package shape.

## Testing — C+
Scope: Tests and runnable example added for the same periodic library changeset; existing one-shot tests retained.
Coverage: Inspected every assertion, callback double, timeout, result channel, cleanup path, and example. Ran the supplied suite repeatedly under race detection; ran one cancellation probe and one independent interval-scheduling mutation in disposable copies.
Rationale: Three independent moderate gaps: incomplete interval semantics verification, incomplete cancellation-result verification, and fragile instance/lifecycle synchronization. Repeated successful callbacks and active callback cancellation/joining do have coverage, so the periodic lifecycle contract is not wholly unverified. None of these gaps warrants major severity in isolation.
Finding counts: critical=0, major=0, moderate=3, minor=0

Good

- [TG1] `refresh_test.go:44-82` synchronizes on callback entry and cancellation observation, delays callback cleanup through a channel, and checks the returned cancellation result after release. This exercises the central host teardown scenario without a guessed startup sleep.
- [TG2] `refresh_test.go:85-112` waits for B's second successful callback using a channel and uses atomic counters. It genuinely exercises repetition, and the supplied tests passed 20 runs under `-race`.
- [TG3] `refresh_test.go:115-125` checks exact first-error identity and exactly one invocation. `refresh_test.go:13-41` preserves the one-shot context/error and nil-callback checks. The two-refresher example has an Output directive and executes with the suite.

Bad

- [F2][moderate][introduced] The added tests do not verify that the interval begins after successful callback completion. The only repeated-success test (`refresh_test.go:85-112`) uses effectively instantaneous callbacks and merely waits for invocation two. In a disposable copy, moving `time.NewTimer(interval)` before `r.Refresh(ctx)` preserved all supplied tests under `go test -race -count=20 -timeout=30s ./...`. That mutation violates `README.md:5` for callbacks taking longer than the interval, because it permits the next refresh immediately on completion. Nonpositive interval rejection also has no assertion in the added suite, although the implementation correctly handles it. Primary remediation owner: Testing.
- [F3][moderate][introduced] The cancellation-result test's callback always returns nil after cancellation (`refresh_test.go:48-53`), so it cannot detect the independently actionable public error-precedence mismatch in F1. No added case returns a distinct callback error after cancellation. The supplied suite passes while the deterministic review probe fails. Primary remediation owner: Testing; production/API issue is F1.
- [F4][moderate][introduced] The independence test cancels A immediately after launching the goroutines (`refresh_test.go:98-101`) without waiting for A to enter its callback. Its `aCalls` counter is never asserted, so it can pass with A never active. B's second invocation also need not occur after A's completed stop if A is delayed. This weakens the intended active-instance shutdown assertion. In addition, joins at lines 101 and 110 are unbounded and cancellation/release cleanup is not registered for early failures (including lines 61 and 67 in the active-refresh test). A cancellation regression can hang until the outer test timeout or leave a callback goroutine blocked after the local assertion. These are one synchronization/ownership correction for the new async tests. Primary remediation owner: Testing.

Suggested changes

- [F2] Add a callback gate so a successful callback spans an interval, then check that the next invocation still waits a full interval after its release, with a bounded deadline and sensible scheduling tolerance. Add zero/negative-interval cases that assert rejection without callback invocation. The interval mutation should fail.
- [F3] Add cancellation cases where an active callback returns a distinct error and a wrapped cancellation error. Assert whichever precedence is chosen in F1, independently from the existing uncanceled first-error case.
- [F4] Wait for both instances to enter before stopping A, then require a B callback observed after A's join. Bound every completion wait and register cleanup that cancels contexts, releases callback gates safely, and joins work even on assertion failures. Keep successful repetition observable without relying on scheduling order.

Limits: The initial test invocation was blocked by sandbox access to the default Go build cache. Re-running with a disposable cache succeeded. Exact executed checks:

- `/private/tmp/lifecycle-review-probe`: `GOCACHE=/private/tmp/lifecycle-review-cache go test -race -count=20 -timeout=30s ./...` — PASS for the unmodified supplied implementation/tests.
- Same copy after adding only `cancellation_review_test.go`: `GOCACHE=/private/tmp/lifecycle-review-cache go test -run TestReviewCancellationResult -count=1 ./...` — FAIL, reporting `documented cancellation result = context.Canceled; got callback cleanup failed`.
- `/private/tmp/lifecycle-review-mutation`: supplied tests with timer creation moved before the callback (and timer stopped on callback error), `GOCACHE=/private/tmp/lifecycle-review-cache go test -race -count=20 -timeout=30s ./...` — PASS despite incorrect completion-relative interval semantics.
- `go version` — `go version go1.26.5 darwin/arm64`.

No candidate/repository files were changed. The probes do not establish a production timing bug in the candidate: its timer is correctly placed after callback completion. No mutation was run for F4; that finding is based on the explicit missing happens-before relationship and unbounded channel receives. Race detection only covers executed paths.
