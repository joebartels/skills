# Changeset 1 review

Reviewed the supplied `original` → `candidate` changeset for the request to replace guessed-readiness setup and add reliable recurrence, cancellation, teardown, and independent-invocation tests for `Poll`. Paths below are relative to `/private/tmp/go-independent-review-ho4jzoxu/changeset-1`.

## Testing — A+
Scope: Complete supplied changeset in the `example.com/poller` library. Only `poll_test.go` changes; production implementation, README, exported API, dependencies, and `go 1.22` directive are unchanged. Verification used Go 1.26.5 on darwin/arm64 with `GOTOOLCHAIN=local` and the designated cache.
Coverage: Inspected the complete implementation, old and new tests, README, and module. Assessed observable startup, sequential completion-relative recurrence, ordinary and already-requested cancellation, invalid intervals, callback error identity during cancellation, caller-owned fixture lifecycle, concurrent invocation independence, failure cleanup, parallel execution, and minimum-version API usage. Executed the full suite, race/repeated/shuffled suite, version vet check, and five isolated production mutations against the candidate tests.
Rationale: No actionable issue found; critical=0, major=0, moderate=0, minor=0. Two independent safeguards support A+: (1) the held-callback recurrence scenario and completion/start timestamps detect fixed-rate scheduling, verified by the ticker mutation failing at `candidate/poll_test.go:219`; (2) the blocked-cleanup scenario and caller-owned file checks detect returning before callback cleanup, verified by the asynchronous-return mutation failing at `candidate/poll_test.go:298`. They control distinct scheduling and lifecycle risks and go beyond routine setup. Invocation independence and cancellation/error assertions provide additional verified signal.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/poll_test.go:74` waits for an actual callback event before canceling, replacing the original guessed sleep. `candidate/poll_test.go:41` bounds event waits; `candidate/poll_test.go:23` registers cancellation and completion cleanup before launching the worker. The full suite passed, and ignoring cancellation during the interval wait caused the expected bounded failures at lines 88 and 32.
- [G2] `candidate/poll_test.go:159` holds a known active first callback for two intervals, checks absence of overlap, and compares two subsequent starts with their preceding callback completions. A fixed-rate ticker mutation failed with callback 2 starting only 4.25µs after completion rather than at least 25ms. The observation timer constructs elapsed-time behavior after readiness is established.
- [G3] `candidate/poll_test.go:230` uses a private temporary file, observes callback cleanup entering a deliberately blocked phase, then verifies Poll has not returned. Cleanup releases and cancels before waiting for callback and Poll completion; it closes the caller fixture only after both finish (`candidate/poll_test.go:256`). The test verifies caller use after Poll and actual final file closure/content. The early-return mutation failed at the intended blocked-cleanup assertion.
- [G4] `candidate/poll_test.go:314` exercises a blocked invocation alongside a recurring invocation and separately exercises cancellation isolation. Context identity and exact sentinel errors distinguish invocations. A package-wide serialization mutation failed while waiting for the independent fast invocation (`candidate/poll_test.go:340`). The candidate suite passed ten repeated, shuffled, race-enabled runs.
- [G5] The old callback error check is retained and strengthened at `candidate/poll_test.go:96`; `candidate/poll_test.go:105` checks that a callback error survives cancellation. A cancellation-hides-error mutation failed at line 118. Already-canceled and nonpositive-interval cases assert both result and zero callbacks (`candidate/poll_test.go:122`, `candidate/poll_test.go:136`). Parallel loop capture is consistent with the module's Go 1.22 language version.

Bad

- None found.

Suggested changes

- None needed.

Limits: These finite checks cannot prove freedom from every scheduler interleaving. The three-second deadlines require a functioning test host; late recurrence is allowed, while the contract's lower interval bound is asserted. Go 1.22 itself and other platforms were not executed; source/API inspection and `go vet -stdversion ./...` found no minimum-version violation. No network listener or dependency installation was needed. Deliberately broken mutations that ignore cancellation can leave a worker alive until the failed test process exits; the failures remain bounded and are not candidate defects. Exact commands, cwd, non-secret environment overrides, separate stdout/stderr, exits, mutation sources, and source-preservation hashes are in `output/checks.json`. Original and candidate source hashes were unchanged during probe verification.

## Correctness & Compatibility — A
Scope: The same complete supplied changeset, assessing introduced test-harness behavior and compatibility. The unchanged runtime implementation is contract context, not an unrelated debt review. Library API and Go 1.22 minimum are explicit requirements.
Coverage: Compared every supplied file, traced new goroutine/channel/resource ownership, checked loop capture under the declared language version and standard-library API availability, and executed the candidate suite and race/repeated/shuffled suite. Production cancellation, sequential callbacks, and direct error returns were inspected against README lines 3–12.
Rationale: No introduced or worsened correctness or consumer-compatibility issue was established. The candidate preserves production source and module configuration; new shared callback state uses atomics or event synchronization, and caller fixture closure follows completion. Those relevant strengths are verified. A applies; this test-only change does not establish two new independent runtime safeguards warranting Correctness & Compatibility A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G6] `candidate/poll.go:10` retains the exact public function signature and implementation; `candidate/go.mod:3` retains Go 1.22. The complete diff contains only test changes, and the candidate package builds/tests successfully with version vet clean.
- [G7] `candidate/poll_test.go:170` uses atomics for state accessed by callback and test goroutines, and event channels order fixture observations (`candidate/poll_test.go:269`). The full exercised concurrency paths passed `go test -race ./... -count=10 -shuffle=on -timeout=45s`.
- [G8] Cleanup joining at `candidate/poll_test.go:259` and `candidate/poll_test.go:260` keeps file closure after both borrowers finish. The candidate fixture test passed and the early-return reproduction was rejected by the intended assertion, without changing candidate source.

Bad

- None found.

Suggested changes

- None needed.

Limits: Executed toolchain/platform coverage is Go 1.26.5 darwin/arm64, not the entire Go 1.22/platform matrix. The supplied library contains no other callers, persistent state, or wire formats to assess. Race checks cover executed paths only. Full evidence is preserved in `output/checks.json`; no environmental test failure occurred.

## Architecture & Design — Not applicable
Scope: Complete test-only changeset for the Poll library.
Coverage: Checked the production diff and the new external-package test helpers for a consequential production/test seam.
Rationale: Production API, package boundaries, composition, and lifecycle design are unchanged. The tests exercise the public API with local callbacks, gates, and caller-owned fixtures; they add no clock abstraction, exported hook, production injection seam, or cross-package ownership decision. No architecture decision is implicated by this changeset.
Limits: This is a changeset assessment, not an architecture audit of the unchanged implementation.
