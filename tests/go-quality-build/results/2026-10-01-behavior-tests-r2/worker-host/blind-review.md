# Independent changeset 2 review

Reviewed the supplied `original` → `candidate` trees for the host-owned sweeper, against the original README and request to implement the periodic lifecycle with high-quality tests. The candidate README clarification that callback-returned errors remain errors during cancellation is consistent with the original callback-error identity promise. No other candidate or controller material was inspected. The legacy one-shot implementation is the requested replacement, not a separately counted defect.

## Testing — A+
Scope: Complete supplied sweeper changeset under `/private/tmp/go-independent-review-z6n_lo0k/changeset-2`; standard-library library, module `go 1.22`; execution used Go 1.26.5 on darwin/arm64.
Coverage: Reviewed every implementation/test file and README. Assessed invalid/already-canceled inputs, immediate start, recurrence and full post-completion interval, no overlap, interval cancellation, in-flight callback cleanup, release ordering/exactly-once/completion, callback/release error identity and joined causes, later-cycle failure, independent instances, consumer function type, synchronization, and bounded cleanup. Executed race/repeated/shuffled verification and two assertion-signal mutations.
Rationale: No actionable testing issue found. Two independent verified safeguards support A+: (1) a deliberately blocked callback plus recorded completion/start times detects a scheduler that consumes the interval during callback work; (2) cancellation/cleanup gates and a release observer detect resource release before the active callback finishes. These control recurrence timing and shutdown ownership, respectively. The checks exercise the public Run API rather than an injected scheduler that could bypass the production lifecycle.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/host_test.go:142`–210 holds the first callback longer than the interval, observes no overlap, then measures the full delay after two completions. `mutation-delay-before-callback` failed at line 201: cycle 2 started 3.917 microseconds after completion instead of at least 40 ms. The untouched test passed.
- [G2] `candidate/host_test.go:213`–275 observes cancellation, deliberately gates cleanup, checks that release and Run do not finish early, then inspects preserved callback error causes and exactly-once release. `mutation-early-release` failed at line 255 in all three subcases, explicitly observing release before cleanup.
- [G3] `candidate/host_test.go:278`–397 checks success/work/release/both errors with `errors.Is`, preserves a typed later-cycle failure with `errors.As`, verifies callback context cancellation at release, and gates release completion. `candidate/host_test.go:400`–444 confirms one Run instance stopping does not stop another's recurrence.
- [G4] `candidate/host_test.go:19`–71 transfers the Run result through a closed channel, owns cancellation and bounded cleanup, and uses atomic state for concurrent observations. The public function-type assignment at line 15 also checks the requested consumer signature. Ten shuffled repetitions under the race detector passed.

Bad

- None found.

Suggested changes

- None needed.

Limits: Untouched `go test -count=1 -timeout=45s ./...` passed (exit 0), and `go test -race -count=10 -shuffle=on -timeout=45s ./...` passed (exit 0; package 4.011 s). Both targeted mutation commands exited 1 for intended assertions. Timing tests use actual monotonic timestamps and bounded absence windows; passing repetitions do not prove every scheduling interleaving. No Go 1.22 executable or non-Darwin runtime was exercised. No environmental failures or listener reruns occurred. Exact commands/cwd/explicit non-secret environment/stdout/stderr/exits are preserved in `output/checks.json`.

## Correctness & Compatibility — A+
Scope: Same complete supplied library changeset; preserved public `Run(context.Context, time.Duration, func(context.Context) error, func() error) error` and original periodic/ownership contract.
Coverage: Traced rejection, already-canceled input, first/later cycles, successful callback waits, callback errors, cancellation while waiting/in a callback, callback cleanup, context cancellation before release, release completion and joined error identity, and separate concurrent instances. No new dependency or module language version was introduced.
Rationale: No introduced correctness or supported-consumer defect found. Two independent verified safeguards support A+: one worker invokes callbacks serially and creates each timer after completion, preventing overlap/catch-up cycles; the host cancels and receives the worker result before release, preventing resources from being released while callback cleanup is using them. Both mechanisms have passing public-API tests and corresponding rejected disposable regressions.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/worker.go:8`–25 checks context before calls, invokes the callback serially, and starts a cancellable interval timer after success. Already-canceled/immediate-start/post-completion/no-overlap/later-failure tests passed. Moving timer creation before callback execution was rejected by the elapsed-interval assertion.
- [G2] `candidate/host.go:21`–35 gives the host explicit ownership of worker cancellation and its completion channel. On either exit path it cancels before release; on parent cancellation it still receives the callback result before release. Gated cleanup/release tests passed, while the early-release mutation failed.
- [G3] `candidate/host.go:17`–18 rejects a nonpositive interval before worker/release effects, and line 35 joins worker and release errors without losing identity. Tests inspect all error combinations, typed callback details, exact counts, and cancellation-associated callback errors. The exported signature and module `go 1.22` are preserved.

Bad

- None found.

Suggested changes

- None needed.

Limits: Actual execution was Go 1.26.5/darwin/arm64, retaining the module's Go 1.22 language setting. Tests/race detection exercise observed paths rather than prove all interleavings. Cancellation requires cooperative callback completion as the original README permits; an uncooperative callback or release can block and is not an introduced violation. Panic recovery, nil callbacks, and nil contexts are not promised contracts. Exact evidence and source hashes are in `output/checks.json` and `output/source-manifest.json`.

## Architecture & Design — A
Scope: Consequential production/test seams introduced by turning Run into a host-owned periodic lifecycle; library API and callback/resource ownership only.
Coverage: Traced the exported Run boundary, worker result channel, context ownership, callback/release sequencing, and external-package lifecycle tests. No broader architecture audit was performed.
Rationale: No actionable architecture issue found. The consequential lifetime remains owned and visible at Run: it starts, cancels, joins, then releases, with inspectable callback/release errors. Tests drive those existing public callbacks through real goroutine/timer behavior; no global hooks, process-wide controls, or injected scheduler bypass the seam. This verified ownership strength supports A; signature preservation and ordinary callback composition are not counted as extra A+ safeguards.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/host.go:21`–35 places cancellation, join, and release ownership in one caller-visible boundary, while `candidate/worker.go:8`–25 owns only cycle execution. The cleanup/release gate tests and rejected early-release mutation verify the lifetime contract.
- [G2] `candidate/host_test.go:1`–15 and 25–71 exercise the public API from an external test package with bounded ownership of each call. The production code exposes no additional test-only lifecycle API or hidden process effects; independent-instance testing passes under race detection.

Bad

- None found.

Suggested changes

- None needed.

Limits: Graded only this introduced lifecycle and its testing seam. Callbacks remain responsible for responding to their context, and no unrelated package/module architecture was assessed.
