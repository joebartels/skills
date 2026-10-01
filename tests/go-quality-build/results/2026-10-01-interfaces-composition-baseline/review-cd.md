# Independent blind review: candidates C and D

Candidate paths below are relative to `/private/tmp/go-quality-build-composition-eval/review-baseline/`. Original fixtures are `zero-value` and `library-lifecycle` under `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-interfaces-and-composition/evals/files/`. This reviews introduced changes against those fixtures, using only candidate source/tests/documentation, original fixture source/tests/documentation, the supplied tasks, and the requested review skills/references. No expected assertions, evaluation configuration, build skills, author reports, or other trials were read. No candidate or repository file was edited.

## Candidate C: Architecture & Design — A
Scope: Candidate C versus the original zero-value fixture; internal single-worker Bag, Go 1.22 module.
Coverage: Configuration/API size, zero-value construction, private state, distinct-key enforcement, rejection atomicity, existing-key allowance, Clear retention, and original caller compatibility. Concurrent use and changing configuration on a populated bag are explicitly outside the fixture contract.
Rationale: No actionable architecture issue. A small concrete type and one setter express the optional configuration, and mutation is guarded before map allocation/count updates. Relevant safeguards were inspected and exercised; these are routine correct implementation of this small contract, so A rather than A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-A-G1] `candidate-c/bag.go:17-22,27-35` keeps configuration optional and leaves a zero-valued Bag usable. The existing map remains private; there is no added interface, constructor requirement, or ownership layer.
- [C-A-G2] `candidate-c/bag.go:28-34` checks existence and capacity before mutation, preserving counts on rejection and permitting existing keys when full. The independent full-capacity probe passed, including the empty-string key.
- [C-A-G3] `candidate-c/bag.go:41` clears only counts, retaining configuration; the candidate test fills the bag again and checks rejection after Clear (`bag_test.go:55-67`).

Bad

- None found.

Suggested changes

- None needed.

Limits: Reviewed the complete supplied fixture-sized module. The supplied suite passed race-enabled repeated/shuffled runs on Go 1.26.5 with a Go 1.22 module directive; the Go 1.22 compiler itself was not run. Exact commands and independent checks appear below.

## Candidate C: Testing — B
Scope: Candidate C test changes against the original fixture, including retained zero-value coverage.
Coverage: Assertions for default operation, missing and empty keys, per-instance isolation, limit rejection, rejection preserving counts, Clear/configuration retention, and invalid configuration. Independently probed the actual full-capacity existing-key case and tested whether the supplied suite detects a regression there.
Rationale: One moderate boundary-coverage defect. The suite meaningfully checks the main capacity and Clear behavior, but misses existing-key updates at capacity. This is a localized regression-detection hole in otherwise exercised limit behavior, rather than absence of verification for the whole feature.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [C-T-G1] `candidate-c/bag_test.go:45-53` checks the rejection sentinel, absence of the rejected key, and preservation of a prior count; it would detect ordinary mutation-on-rejection regressions.
- [C-T-G2] `candidate-c/bag_test.go:55-67` adds new keys after Clear and verifies the old capacity remains enforced.

Bad

- [C-T1][moderate][introduced] The purported existing-key-at-limit assertion runs before capacity is reached: `candidate-c/bag_test.go:33` sets limit 2, while lines 36-40 add only `jobs`; `errors` is first added at line 42. A disposable mutation that rejects every Add when `len(counts) >= limit`, including existing keys, passed the entire supplied suite ten times under the race detector. Thus an ordinary update after filling the bag can regress undetected.

Suggested changes

- [C-T1] Add an existing-key increment after both distinct keys have been inserted; check success and the new count. This must fail the demonstrated mutation while preserving the rejection assertions.

Limits: Reviewer-added probes establish production behavior, not candidate test quality; they are not credited as authored safeguards. Invalid configuration before any positive limit is tested, but more general reconfiguration is outside the supplied task.

## Candidate D: Architecture & Design — B
Scope: Candidate D versus original library-lifecycle fixture; embedded library preserving New/Refresh and adding host-controlled periodic runs, Go 1.22 module.
Coverage: Construction and dependency ownership; immediate/successive refreshes; completion-relative intervals; callback error propagation; cancellation; joining active callbacks; independent instances; single-active-run guard; timer/channel synchronization; and shutdown examples. The fixture excludes concurrent starts/restarts as a requirement, though the implementation also serializes its running flag.
Rationale: One moderate issue in the documented host shutdown composition. The implementation gives callers suitable cancellation and join primitives, but its general README example can return on the first error before joining the second run. This undermines the example's resource-ownership guidance under an ordinary callback failure; the library API itself does not need redesign.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [D-A-G1] `candidate-d/refresh.go:46-48,90-100` gives each run its own child context, cancellation function, completion channel, and returned error. Stop requests cancellation and Wait observes callback completion. The independent gated callback probe verified Wait remains blocked until callback exit.
- [D-A-G2] `candidate-d/refresh.go:57-76` executes callbacks serially and creates the timer after successful completion. The independent probe held the first callback longer than the interval, then verified a full interval before call two and exact propagation of call two's failure.
- [D-A-G3] `candidate-d/refresh.go:49-55,98-100` publishes the result before closing done, establishing synchronization for Wait reads; running state is mutex-protected. Race-enabled repeated runs and probes found no race. No process signal handlers, process exit, or shared cancellation ownership were introduced.

Bad

- [D-A1][moderate][introduced] `candidate-d/README.md:34-43` stops both runs but returns immediately when the first Wait reports a callback failure (lines 37-38), skipping the second Wait at line 40. A host following this orderly-shutdown recipe can return into deferred resource teardown while the second canceled callback is still unwinding. Cancellation alone does not join it; the fixture explicitly requires waiting before resource release. The runnable example covers only callbacks that return cancellation and therefore does not expose this error branch.

Suggested changes

- [D-A1] Always collect both Wait results before returning either error, then handle/aggregate non-cancellation failures. Verify with one run failing and the second callback held in cleanup until explicitly released; shutdown must not finish until that second callback returns.

Limits: No production race or deadlock was observed. Timer draining is safe for the inspected single-reader loop: the canceled select branch has not consumed timer.C elsewhere. Tests ran using Go 1.26.5; the module declares Go 1.22. No claim is made about arbitrary callbacks ignoring cancellation, which the fixture excludes. Parent deadline during an active callback can return the callback's DeadlineExceeded unchanged; documentation gives callback errors unchanged precedence in Wait's API comment, while README cancellation wording could be clearer, but this ambiguity is not counted as an independent architecture defect.

## Candidate D: Testing — C
Scope: Candidate D's added periodic tests and runnable two-instance example, with original synchronous tests retained.
Coverage: Reviewed actual assertions and synchronization for cancellation/joining, independent instances, callback failures, argument validation, duplicate-run rejection, scheduling, and resource lifecycle. Ran race-enabled repetition and targeted probes; used a disposable scheduling mutation to establish lost test signal.
Rationale: One contained major gap leaves successful periodic scheduling effectively unverified; one moderate gap weakens independent-instance verification. The major issue is contained to the central scheduling contract, not systemic across unrelated workflows. The grade follows the one-major/at-most-one-moderate anchor.
Finding counts: critical=0, major=1, moderate=1, minor=0

Good

- [D-T-G1] `candidate-d/refresh_test.go:42-67` synchronizes on callback start, cancels the active callback, checks cancellation, and checks its completion marker after Wait.
- [D-T-G2] `candidate-d/refresh_test.go:118-137` checks exact callback error identity and one invocation; `example_test.go:11-45` is an executable two-instance API example.

Bad

- [D-T1][major][introduced] No supplied periodic callback ever returns success and continues into a second scheduled invocation (`candidate-d/refresh_test.go:45-49,73-85,121-123,141-143`; `example_test.go:13-15`). Consequently no assertion checks completion-relative intervals or successful recurrence. A disposable mutation returning immediately after the first successful callback, deleting the actual periodic behavior, passed the entire supplied suite ten times under the race detector. The central requested scheduling contract is therefore effectively unverified.
- [D-T2][moderate][introduced] `candidate-d/refresh_test.go:106-110` checks the newly created waitResult channel with an immediate default; the helper only starts its Wait goroutine at lines 166-169. Even if the second run has already terminated incorrectly, the select can choose default before that goroutine sends. The subsequent Stop/Wait expects cancellation anyway, so it does not repair this false-negative path. The test can miss exactly the cross-instance cancellation it claims to detect.

Suggested changes

- [D-T1] Exercise at least two invocations with a successful first callback, hold that callback past the requested interval, and assert that the next invocation starts only after a full interval from completion. Assert no overlap and propagate a deliberate later callback failure. Use synchronization and bounded waits suitable for the Go 1.22 target.
- [D-T2] Observe the second callback's own cancellation/continued-work state after the first run has been joined, with a bounded observation or explicit handshake that cannot race a freshly started result-forwarding goroutine. Demonstrate that shared-cancellation behavior fails the revised test.

Limits: Several existing channel receives and Wait calls have no local timeout (`refresh_test.go:58,99-100,132`); review executions used a 30-second process test timeout. This limits failure diagnostics and was not counted as an additional independent finding. Reviewer-added timing/join probes assess implementation behavior and do not repair the submitted test suite. The README shutdown finding D-A1 has Architecture as its primary owner; it is not counted again as a separate testing finding.

## Executed checks and reproducibility

- Runtime: `rtk proxy go version` returned `go version go1.26.5 darwin/arm64`.
- All checks used disposable copies under `/private/tmp/composition-cd-independent-check/`; original candidates were preserved.
- Initial `rtk proxy go test -race -count=20 -shuffle=on -timeout=30s ./...` commands failed at setup because the default user Go build cache was not writable in the sandbox. This was an environment failure, not a candidate failure.
- With a writable cache, each candidate passed `rtk proxy env GOCACHE=/private/tmp/composition-cd-go-cache go test -race -count=20 -shuffle=on -timeout=30s ./...` (C: 1.226s; D: 1.346s).
- C plus `reviewer_test.go` passed `rtk proxy env GOCACHE=/private/tmp/composition-cd-go-cache go test -race -count=10 -timeout=30s ./...` (1.373s). The probe fills a limit-one bag using the empty key, then increments that existing key and checks count two.
- D plus `reviewer_test.go` passed the same race/count/timeout command (2.236s). Probes verify a full 30ms completion-relative interval after holding the first callback 60ms, subsequent error propagation, active cancellation, and Wait remaining blocked while the canceled callback is deliberately held before returning.
- `mutant-c` contains only supplied candidate C tests and changes Add to reject any key when at capacity. The same race/count/timeout command passed (1.270s), demonstrating C-T1.
- `mutant-d` contains only supplied candidate D tests and inserts an immediate return after any successful callback before timer creation. The same command passed (1.453s), demonstrating D-T1.

No candidate source/configuration was changed and no external services were involved. Passing race detection covers only executed paths. No broader quality, security, or performance grade is implied.
