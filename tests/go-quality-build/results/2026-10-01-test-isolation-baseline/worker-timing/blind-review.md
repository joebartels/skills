# Independent review of anonymized Poll test changes

Reviewed the supplied snapshots at `/private/tmp/go-independent-review-ry7bg0wi/original` and `/private/tmp/go-independent-review-ry7bg0wi/candidate`, using only the dispatch-authorized README, source, and local review guidance. No author report, other candidate, expectations, or controller material was inspected. No delegation or source repairs were performed. All mutations were confined to disposable `verification/` copies.

The supplied diff changes only `poll_test.go`. `poll.go`, `README.md`, and `go.mod` are identical between snapshots. References below are relative to `candidate/` unless stated otherwise. The README is the contract, including immediate first invocation, completion-relative sequential recurrence, error identity and precedence over concurrent cancellation, normal nil cancellation, caller context and resource ownership, and joining active callback cleanup.

## Testing — A+

Scope: Supplied original/candidate changeset for the `example.com/poller` Go library; requested test replacement and additions. The module declares Go 1.22 and uses only the standard library. Executed with Go 1.26.5 on darwin/arm64.

Coverage: Inspected all implementation and test files and the README. Assessed assertion signal, normal and invalid inputs, recurrence timing and non-overlap, cancellation before and during execution, error identity, caller context, fixture lifetime, per-test isolation, and simultaneous invocations. Executed the whole suite normally and under repeated shuffled race detection, each test alone, production mutation checks, and forced test-failure teardown checks. Fuzzing, benchmarks, external integration, and CI enforcement are not implicated by this supplied changeset.

Rationale: No actionable introduced or worsened testing issue was found. Two independent safeguards earn A+: (1) the blocked-first-callback recurrence test rejects both fixed-tick scheduling and recurrence without the required completion-relative delay; (2) the deliberately blocked callback cleanup test rejects returning before callback cleanup finishes. They protect distinct documented timing and lifetime contracts, with executed mutation evidence rather than test count or coverage percentage. The ordinary helper ownership and synchronization support these safeguards but are not counted as separate A+ safeguards.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `poll_test.go:169` holds callback one beyond the interval, records ordered entry/completion events, checks active callback count, and asserts the lower bound on subsequent starts at `poll_test.go:233`. The fixed-ticker mutation failed with callback two only 3µs after completion; the no-delay mutation failed with 9.334µs. This detects completion-relative scheduling regressions and exercises multiple cycles rather than merely observing a first callback.
- [T-G2] `poll_test.go:245` observes callback resource use and cancellation, blocks its deferred cleanup, requires Poll to remain running, then releases cleanup and checks completion and caller resource usability. The asynchronous early-return mutation failed at `poll_test.go:294` with “Poll returned before callback cleanup completed.” The negative timing check follows explicit readiness signals; elapsed time is the property being tested.
- [T-G3] `poll_test.go:23` registers cancellation and a bounded join; the closed done channel synchronizes the returned error before `poll_test.go:54` reads it. For the fixture test, cleanup registration at lines 253, 267, and 282 yields gate release, Poll cancellation/join, then file close. Injecting a fatal assertion after observed callback cancellation finished with only the intentional failure, without a teardown timeout or race diagnostic. An analogous fatal assertion with both independent invocations active also finished promptly with only its intentional failure.
- [T-G4] `poll_test.go:311` allows invocation B to enter, recur, and return its own error while canceled A remains blocked in cleanup. A process-wide invocation mutex mutation failed on the second invocation entry. The modified test failed finitely; its additional teardown timeout belongs to that intentionally broken implementation, not the supplied candidate.
- [T-G5] `poll_test.go:57` replaces the old 5ms guessed-readiness sleep with observed callback entry and preserves the cancellation result assertion. Tests at lines 80, 97, 117, 134, and 151 add meaningful callback-count, invalid-interval, original-error, cancellation/error-precedence, and exact-caller-context assertions. Targeted mutations for delayed first execution, precanceled callback execution, invalid-input callback execution, suppressed cancellation-time errors, and substituted context each failed at the relevant assertion or bounded wait.
- [T-G6] All nine top-level tests passed when invoked alone. The full candidate passed `go test -race -shuffle=on -count=50 -timeout=30s ./...`, including its parallel tests, with no race reports. Shared mutable state is absent from the supplied tests; each invocation owns its context, channels, and atomics, and the file fixture uses `t.TempDir`.

Bad

None found.

Suggested changes

None needed.

Limits: Checks use real time and finite waits; successful repetition does not prove behavior under arbitrary scheduler starvation. Callback completion timestamps are taken immediately before return, so they are a conservative observation of completion, while the active count separately checks overlap. No clock framework or newer `testing/synctest` API is required by this Go 1.22 task. Mutation checks establish detection of the specific plausible regressions tested, not every possible faulty implementation. Forced failures exercise the candidate's teardown with the unchanged cooperative Poll implementation; no claim is made that teardown can repair an arbitrary implementation that refuses cancellation forever. Exact commands, working directories, environment, stdout, stderr, and exits are in `output/checks.json`.

## Correctness & Compatibility — A

Scope: Supplied original/candidate changeset, assessing the explicit preservation constraints and consumer-visible Poll contract while evaluating the test changes. Production implementation, exported function type, README, and module declaration are unchanged.

Coverage: Compared both complete snapshots. Traced invalid intervals, precanceled context, immediate callback execution, sequential completion-relative recurrence, callback error precedence, cancellation, caller context, active callback lifetime, and independent invocations. Inspected the new test helpers for synchronization and build compatibility. Executed normal and race-enabled tests and standard-library version vetting.

Rationale: No introduced or worsened functional or compatibility defect was found. Retaining the production implementation and public function type preserves existing consumer calls and function-value compatibility. The new tests and helpers build, synchronize their results, and preserve Go 1.22-compatible source/API choices. This is a changeset grade for verified preservation; it is not an unrelated audit of hypothetical legacy inputs.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `poll.go:10` and `go.mod:3` are byte-identical to the original: the exported signature, Go 1.22 minimum, and standard-library dependency scope are preserved. `go vet -stdversion ./...` exited 0 with empty stdout/stderr. The tests use generic helpers, atomics, contexts, timers, and existing testing APIs, without requiring a newer testing API.
- [C-G2] `poll.go:18` returns the callback error directly before considering cancellation. `TestCallbackError` and `TestCallbackErrorDuringCancellation` passed and verify original error identity, including a callback that observes caller cancellation before returning its error. The error-suppression mutation was rejected.
- [C-G3] `poll.go:18` invokes the callback synchronously and creates the next timer only after it returns at line 21. Recurrence, blocked cleanup, and independent invocation tests passed, including exercised race detection. The original and candidate implementations are identical; the changes add regression detection without changing ownership or execution semantics.

Bad

None found.

Suggested changes

None needed.

Limits: Executed toolchain/platform is Go 1.26.5 darwin/arm64 with CGO enabled. An actual Go 1.22 runtime and other platforms were not executed; minimum compatibility is supported by source inspection, the unchanged module and implementation, and `stdversion` vetting. No network access or external dependencies were needed. The README excludes non-cooperating callback cleanup from an automatic cancellation guarantee; unspecified nil callbacks/contexts or callback panics are unchanged legacy concerns and do not lower this changeset grade. A clean race run applies only to exercised paths.

## Architecture & Design — Not applicable

Scope: Supplied original/candidate changeset.

Coverage: Inspected the full diff and production implementation for changes to package boundaries, production seams, API, lifetime ownership, and invocation state.

Rationale: Only external-package tests changed. No production seam, lifetime decision, package boundary, exported hook, or dependency was introduced or changed consequentially. Test-local lifecycle helpers are assessed under Testing above. The Architecture skill was therefore not added.

Limits: This non-applicability conclusion is limited to the supplied complete changeset, not an overall architecture assessment.

## Supporting verification facts

All Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`, with `rtk proxy` for raw checks. `checks.json` preserves 27 exact command records. `snapshot-hashes.json` records the inspected snapshot files. Verification copies preserve their mutation source and tests.

- Original normal suite: exit 0, `ok example.com/poller 0.181s`.
- Candidate normal suite: exit 0, `ok example.com/poller 0.259s`.
- Candidate race/shuffle/repetition suite: exit 0, `ok example.com/poller 5.490s`.
- Each of nine candidate top-level tests invoked alone: exit 0.
- Standard-library version vetting: exit 0; empty stdout/stderr.
- Nine targeted production mutation checks: exit 1, with relevant assertion or bounded-wait failures; none required the outer 10s Go test timeout.
- Two forced assertion-failure checks under `-race`: exit 1 from their deliberate fatal assertions only, with no added teardown timeout or race diagnostic; test bodies reported 0.00s.
- The supplied diff command returns exit 1 because differences exist; its sole changed file is `poll_test.go`.
