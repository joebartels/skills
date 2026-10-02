# Independent review: changeset 2

Review boundary: supplied `original/` versus `candidate/` directories in `/private/tmp/go-independent-review-b_q8mpxi/changeset-2`. The actual request is to extend the one-shot host to the README's periodic-worker lifecycle and add high-quality changed-lifecycle tests. No author report, other candidate, or evaluation guidance was used. File references below are relative to this changeset directory. Exact verification commands, working directories, non-secret environment overrides, stdout, stderr, exit codes, mutations, and source hashes are in `checks.json`.

## Testing — A+

Scope: Changed `candidate/host.go` and `candidate/worker.go`, added external-consumer `candidate/lifecycle_test.go`, retained failure/release test, and clarified callback-error precedence in `candidate/README.md`. The module declares Go 1.22; execution used Go 1.26.5 on darwin/arm64.

Coverage: Exact public function type, invalid-interval side effects, already-canceled input, immediate first sweep, recurrence, full post-completion intervals, nonoverlap, cancellation between cycles and inside callbacks, callback cleanup before release/return, failure after successful cycles, exact release count, independent callback/release error identities, callback-error precedence over cancellation, noncomparable errors, and simultaneous independent instances. Assessed synchronization, fixture cancellation/unblocking/join, meaningful timing assertions, and race/shuffled execution.

Rationale: No actionable testing issue found. Two independent verified safeguards support A+: the blocked callback plus timestamp assertions detect shortened post-completion scheduling, and the cleanup handshake detects release/return before callback cleanup. Targeted mutations failed at their intended assertions, rather than merely timing out. A separate mutation verified detection of lost callback cancellation errors.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/lifecycle_test.go:130` holds the first callback longer than the interval and compares later starts to observed callback completions. A zero-delay timer mutation failed at line 185 (`73.292µs`, expected at least `40ms`), demonstrating signal for the explicit interval-after-completion contract. It also tracks active callbacks and overlap.
- [T-G2] `candidate/lifecycle_test.go:274` gates cancellation cleanup and separately observes release and host completion. A host mutation that returned on cancellation without joining its callback failed at line 317: release ran before cleanup finished. The fixture's explicit unblock keeps the check finite even when the assertion fails.
- [T-G3] `candidate/lifecycle_test.go:218` checks independent callback and release causes with `errors.Is`; `candidate/lifecycle_test.go:274` includes `context.Canceled`, a noncomparable callback error, and a joined error. Suppressing a callback error when the context is canceled failed at line 329, proving the precedence assertion detects that regression.
- [T-G4] `candidate/lifecycle_test.go:26` registers cancellation, explicit callback unblocking, and bounded host joining as cleanup. Results are read after the done-channel synchronization; shared counters use atomics. The external package verifies consumer use, including the exact function type at line 16, without bypassing the host/worker boundary.
- [T-G5] `candidate/lifecycle_test.go:344` runs two hosts and verifies the second continues after the first is canceled and joined. The complete suite passed with `-race -shuffle=on -count=3`, exercising the lifecycle concurrently and in changing test order.

Bad

- None found.

Suggested changes

- None needed.

Limits: `rtk proxy go test -count=1 -timeout=45s ./...` and `rtk proxy go test -race -shuffle=on -count=3 -timeout=45s ./...` both passed (exit 0). All three mutation checks exited 1 at the expected assertions. Timing uses real clock intervals and bounded two-second condition waits; the negative waits assess explicitly forbidden early events rather than guess that work completed. Go 1.22 itself and other platforms were not executed. Finite race/repetition checks do not prove every interleaving. No environmental verification failure or listener rerun occurred. Original/candidate source hashes remained unchanged.

## Correctness & Compatibility — A+

Scope: Supplied one-shot-to-periodic library diff, preserved `Run(context.Context, time.Duration, func(context.Context) error, func() error) error` signature, and stated lifecycle/error contracts. The candidate README clarification is consistent with the original requirement that callback errors end work and retain their identity.

Coverage: Positive/nonpositive intervals, cancellation before/between/during cycles, sequential callback execution, interval start point, callback error termination and precedence, error joining, release exactly once after callback cleanup, child-context cancellation before release, independent Run invocations, and source compatibility of the public function type. Module/version/build constraints and standard-library imports were inspected.

Rationale: No introduced or worsened correctness issue found. Two independent verified safeguards support A+: synchronous ownership prevents release while callback cleanup is still active, and unconditional callback-error propagation plus `errors.Join` prevents losing callback or release causes during cancellation/failure. The gated cleanup and error-identity tests verify these distinct risks; mutations that remove the join or suppress canceled callback errors each fail directly.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/worker.go:10` skips already-canceled work, calls `sweep` synchronously, and starts its timer only after successful callback completion at line 19. `candidate/lifecycle_test.go:87`, `:103`, `:130`, and `:198` verify immediate start, subsequent cycles, complete intervals, no overlap, normal cancellation, and later failure termination.
- [C-G2] `candidate/host.go:20` creates a per-invocation child context and calls the worker synchronously; only after it returns does the host cancel and invoke release at lines 22–23. The blocked-cleanup and independent-host tests verify that release/return follows cleanup and one host's shutdown does not stop another.
- [C-G3] `candidate/worker.go:13` returns callback errors before considering later cancellation, and `candidate/host.go:23` joins that result with the release error. Tests at `candidate/lifecycle_test.go:218` and `:274` verify both causes, wrapped errors, noncomparable callback errors, and callback-returned `context.Canceled` during parent cancellation.
- [C-G4] `candidate/host.go:16` rejects invalid intervals before deriving a context or invoking either callback. `candidate/lifecycle_test.go:66` verifies both callbacks remain unused; line 16's exact function-type assignment verifies the public signature remains usable by consumers.

Bad

- None found.

Suggested changes

- None needed.

Limits: Full normal and race/shuffled tests passed on Go 1.26.5/darwin/arm64. Go 1.22 execution and other operating systems were not tested. The callback remains responsible for completing when its context is canceled, as explicitly allowed by the README; the tests model that cleanup behavior. No claim is made about unsupported nil callbacks or callbacks that never finish. Exact commands/results and preserved-source hashes are in `checks.json`.

## Architecture & Design — A

Scope: Only consequential worker/host lifecycle ownership and production/test seams in `candidate/host.go`, `candidate/worker.go`, and `candidate/lifecycle_test.go`. No unrelated package/design audit.

Coverage: Caller-owned context, per-invocation worker lifetime, cancellation/error flow, resource-release ownership, explicit callback dependencies, preserved public API, and external-consumer verification.

Rationale: No actionable architecture issue found. The public host owns its derived context and release; the private synchronous worker owns recurring callback sequencing and its timer. This makes the cleanup ordering explicit without creating an additional background owner or a test-only scheduling API. A verified relevant ownership strength supports A.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate/host.go:15` exposes the real required dependencies through its existing signature. `candidate/host.go:20`–`:23` visibly owns cancellation and release while `candidate/worker.go:8` remains private and synchronous. The cleanup handshake and independent-instance tests verify that ownership through the exported API; tests do not replace `runWorker` with a fake.

Bad

- None found.

Suggested changes

- None needed.

Limits: Assessment is restricted to the supplied library's consequential lifecycle seams. No wider system or deployment architecture was provided or assessed. Exact verification and source-preservation evidence is in `checks.json`.
