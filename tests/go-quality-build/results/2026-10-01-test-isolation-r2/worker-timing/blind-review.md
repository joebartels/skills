# Changeset 1 independent review

Reviewed the supplied original/candidate directory diff for the request to replace guessed-readiness setup and add recurrence, cancellation, fixture teardown and independent-invocation tests. The only changed file is poll_test.go; poll.go, README.md and go.mod are byte-identical. No candidate or original source was edited. All execution and mutations used disposable copies under this changeset's output directory.

## Testing — A+
Scope: Supplied changeset-1 original/candidate diff; synchronous Poll library; declared Go 1.22 minimum and standard-library-only module.
Coverage: Inspected the complete supplied implementation, README contract, previous checks and new tests. Assessed immediate first-call behavior, completion-relative recurrence, sequential callbacks, cancellation before and during work, callback-error identity and precedence, caller-context propagation, fixture ownership/teardown and independent concurrent invocations. Executed ordinary tests, repeated shuffled race tests, targeted assertion mutations and effective-minimum standard-library vetting. No unrelated repository or CI assessment was performed.
Rationale: No actionable introduced or worsened testing issue was found. A+ is supported by two independent verified safeguards beyond routine setup: (1) gated recurrence measures the delay from actual callback-end events and rejects start-relative scheduling; (2) gated callback cleanup plus cancel-and-join teardown verifies that completion cannot precede resource use and rejects both early Poll return and fixture-before-join cleanup. These protect distinct documented scheduling and lifetime contracts, rather than counting multiple assertions of the same safeguard.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:155 — Callback start/completion events, release gates, concurrent-call detection and timestamps check sequential completion-relative recurrence over three cycles. A disposable mutation that starts the interval timer before the callback failed at poll_test.go:213: the second callback began 5.416µs after completion instead of at least 30ms.
- [G2] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:252 and /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:390 — The callback-cleanup gate verifies Poll remains active until cleanup can finish; the separate return-with-work-active subtest exercises cleanup itself, verifies the callback can write its still-open fixture, and checks final file close and directory removal. An asynchronous early-return mutation failed at poll_test.go:312. Reversing fixture cleanup registration to run before cancel-and-join failed with “Poll still running; fixture cannot safely close.”
- [G3] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:23 — Cancellation and a bounded worker join are registered before work starts or assertions run. Result publication through the closed finished channel synchronizes run.err, while callback counters use atomics. The cancellation test waits for an observed callback instead of a guessed sleep. Ordinary tests passed, and 20 shuffled full-suite race runs passed.
- [G4] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:82 and /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:232 — Useful callback-error coverage is preserved and strengthened, including the original error's identity during cancellation. A mutation that suppressed the callback error when ctx was canceled failed at poll_test.go:248.
- [G5] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:346 — Both independent calls must become active before one is canceled; the other must remain active and return its own callback error without its caller context being canceled. A disposable process-wide invocation mutex failed this test's bounded readiness/join checks.
- [G6] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:93 and /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:110 — Already-canceled callers and both nonpositive interval cases assert that no callback is invoked. No clock framework, exported test hooks or newer testing APIs were introduced.

Bad

- None found.

Suggested changes

- None needed.

Limits: Verification used Go 1.26.5 on darwin/arm64 with CGO_ENABLED=1, GOCACHE=/private/tmp/go-quality-testing-cache, GOTOOLCHAIN=local and GOPROXY=off. Actual Go 1.22 and other platforms were not executed; go.mod remains go 1.22, the source uses compatible facilities, and explicit go vet -stdversion ./... passed. The ordinary suite and go test -mod=readonly -race -count=20 -shuffle=on -timeout=45s ./... exited 0. Each of the five deliberate mutations exited 1 with a relevant test diagnostic, not a compiler failure; the global-lock mutation also triggered its expected bounded helper-cleanup timeout. These are finite demonstrations of assertion signal, not proof against all schedules. Exact argv, cwd, non-secret environment overrides, stdout, stderr, exits and mutation descriptions are in checks.json. There were no environmental failures or listener reruns. Source hashes were checked unchanged after verification.

## Correctness & Compatibility — Not applicable
Scope: Supplied changeset-1 original/candidate diff; test-only changes for Poll.
Coverage: Compared production source, public signature, module minimum and dependency declaration; read the documented behavior to judge the tests. Production implementation, API, README and go.mod are unchanged.
Rationale: No supported production behavior or consumer-contract decision is changed by this diff. The new test synchronization, cancellation and resource-lifetime decisions are assessed under Testing and the limited test-seam Architecture card below. This is not a correctness grade for all unchanged Poll behavior.
Limits: No broader production correctness audit is claimed. Builds and test execution used Go 1.26.5/darwin/arm64; effective Go 1.22 standard-library compatibility was vetted, but an actual Go 1.22 toolchain was not run.

## Architecture & Design — A
Scope: The consequential test-only worker/fixture lifecycle seam introduced by startPoll and the fixture subtests. No broader production architecture review.
Coverage: Assessed who starts, cancels and joins each worker, how callback resources remain caller-owned, how results become visible, and how cleanup order releases callback gates before joining and closes fixtures afterward. Production package boundaries and API are unchanged.
Rationale: No actionable seam-design issue was found. Worker lifecycle is owned by one test helper and bounded cleanup; fixture lifetimes remain owned by their test scope. Channel completion provides a clear result-publication boundary without a production clock abstraction or exported test hook. These establish a verified relevant strength supporting A.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:23 — The helper centralizes cancellation, completion publication and joining without shared process state or changes to Poll's public API; repeated shuffled race verification passed.
- [G2] /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:264, /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:299 and /private/tmp/go-independent-review-85l0o0s_/changeset-1/candidate/poll_test.go:427 — Registration order makes gate release precede cancel-and-join and makes fixture close follow worker completion. The return-with-active-callback test executes that ownership boundary, and the reversed-order mutation failed with the intended diagnostic.

Bad

- None found.

Suggested changes

- None needed.

Limits: Assessment is confined to this test lifecycle seam, not unrelated production design. A bounded join cannot force a deliberately noncooperative broken implementation to finish; timeout diagnostics preserve that failure rather than claiming a successful join. checks.json records the successful candidate runs and expected mutation failures.
