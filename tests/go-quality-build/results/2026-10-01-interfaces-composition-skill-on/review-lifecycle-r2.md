# Independent lifecycle changeset review

Compared the original fixture `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-interfaces-and-composition/evals/files/library-lifecycle` with supplied candidate `/private/tmp/go-quality-build-composition-eval/skill-on-r2/library-lifecycle`. Paths below are relative to the candidate. Read both complete README, go.mod, source and test files, the exact supplied task prompt, and Architecture and Testing review skills and their decision references. No assertion manifest, trial report, build skill, other trial or previous review was read. Candidate directory inventory contains no executable example or Go example test.

## Architecture & Design — A
Scope: Supplied original-to-candidate library changeset; Go module declares Go 1.22. Assessed new periodic lifecycle while retaining New and Refresh.
Coverage: Public lifecycle shape, dependencies, host resource ownership, callback serialization, interval placement, error identity/precedence, cancellation, instance isolation, shutdown composition and compatibility with one-shot API. Concurrent starts/restarts are explicitly outside the supplied requirement and candidate documentation; lack of concurrency protection for that unsupported use is not counted.
Rationale: No substantiated architecture defect under the documented supported use. The concrete Run handle makes cancellation and joining explicit, avoiding hidden host lifecycle ownership. Material lifecycle paths are inspectable and exercised by candidate tests and a reviewer host-context probe. This is solid correct setup; no A+ claim.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `refresh.go:24`–`31`, `refresh.go:41`–`50`: Start creates a run-specific cancellation function and completion channel; Stop requests cancellation, and Wait observes completion before exposing the result. Channel closure follows callback return. This gives the host a sound resource-release boundary and independent handles. Active-callback joining and independent-instance tests passed under the race detector.
- [A-G2] `refresh.go:53`–`77`: A single loop invokes callbacks synchronously, waits only after successful completion, and exits with the original first callback error. No process signal hooks or global cancellation state are introduced. Original Refresh remains unchanged in behavior at `refresh.go:17`.
- [A-G3] `README.md:15`–`25` documents nonblocking Stop, joining, resource ownership, callback-error precedence, and stopping all runs before joining. The reviewer probe verifies that canceling the supplied host context reaches an active callback and permits Wait to complete.

Bad

- None found.

Suggested changes

- None needed for the graded architecture scope. See the separate task-delivery finding below.

Limits: Code inspection and local Go 1.26.5 darwin/arm64 tests, not a Go 1.22 toolchain execution. No external I/O dependencies exist in this fixture. Checks and mutation limitations are recorded below. Cancellation cannot force completion of callbacks that ignore context, which is outside the README's callback contract.

## Testing — B
Scope: Tests introduced for the same supplied library changeset, retaining original one-shot tests.
Coverage: Inspected every test and assertion; exercised periodic success, elapsed interval, active callback stop/join, error identity and precedence, independent instances, invalid intervals and one-shot compatibility. Examined missing host-context cancellation coverage through a controlled mutation. No benchmarks or fuzz targets are relevant to this small lifecycle change.
Rationale: One moderate introduced coverage gap: host cancellation through Start's supplied context is a documented normal-use control, but every added Start call uses context.Background. This is a specific uncovered entry path within an otherwise substantively tested cancellation/join contract, rather than an entirely unverified lifecycle contract, so moderate is appropriate.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [T-G1] `refresh_test.go:76`–`102` keeps callback cleanup blocked after Stop and verifies Wait does not return before cleanup; afterward it checks callback error identity. This meaningfully exercises the resource-release contract and error precedence.
- [T-G2] `refresh_test.go:43`–`73` holds the first callback beyond the configured interval, then measures the next callback relative to first completion. This checks the required completion-based delay, rather than simply counting invocations.
- [T-G3] `refresh_test.go:105`–`130` waits for both instances to make progress, joins A, then observes further progress from B while A's count stays fixed. `refresh_test.go:142`–`155` checks first error identity and exactly one call. These tests exercise observable outcomes with bounded progress checks.

Bad

- [T1][moderate][introduced] Start's host-context control can regress undetected. Relevant locations: `refresh_test.go:86`, `refresh_test.go:109`, `refresh_test.go:113`, `refresh_test.go:146` and the remaining Start calls at lines 58 and 136 all pass context.Background. The documented control at `README.md:14`–`17` depends on `refresh.go:28` deriving cancellation from the supplied context. In a disposable copy, replacing `context.WithCancel(ctx)` with `context.WithCancel(context.Background())` left the entire candidate suite passing three race-enabled runs. Such a regression disconnects host shutdown/deadlines from active periodic work unless the host separately calls Stop.

Suggested changes

- [T1] Add a test that passes a cancelable host context, waits for a callback to enter, cancels that parent without calling Stop, and verifies callback cancellation and a bounded Wait result. Also checking a propagated context value or deadline would protect the complete supplied-context relationship. The required cancellation test should fail for the demonstrated mutation.

Limits: Passing tests do not establish behavior in unexecuted schedules. The timing-based negative assertion at `refresh_test.go:94`–`98` depends on scheduling, although the test additionally checks the eventual callback result. No demonstrated flake is claimed. Direct waits in some tests rely on the overall go-test timeout if broken. No independent severity is assigned to those observations.

## Ungraded task-delivery finding

- [D1][moderate][introduced omission] The exact task explicitly asks for a runnable example of orderly shutdown for two refreshers. `README.md:27`–`40` supplies only a helper taking already-started Run values, with assumed imports and prose explaining startup failure cleanup. There is no main package or runnable Example function, no callback/resource setup, and no demonstrated startup/teardown call site. The helper correctly orders both Stop calls before both Wait calls, but it does not satisfy the runnable-example deliverable. Add a compiling runnable example creating two refreshers, handling second-start failure by joining the first, stopping both before either join, and releasing captured resources after both joins. This is recorded separately rather than inventing an architecture defect in an otherwise clear ownership API.

## Executed checks

All modifications were confined to `/private/tmp/review-lifecycle-r2-checks`; the candidate and original fixture were not modified.

1. Candidate copy: `go test -race -count=3 -timeout=20s ./...` — PASS (`ok example.com/library-lifecycle 2.025s`).
2. Host-context-discarding mutation: initial command hit sandbox denial for the default Go cache before tests ran. Retried using writable temporary cache: `GOCACHE=/private/tmp/review-lifecycle-r2-checks/go-cache go test -race -count=3 -timeout=20s ./...` — PASS (`ok example.com/library-lifecycle 1.973s`). This is mutation-survival evidence for T1, not evidence that the mutant is correct.
3. Unmutated candidate copy plus reviewer-written host-context cancellation probe: `GOCACHE=/private/tmp/review-lifecycle-r2-checks/go-cache go test -race -timeout=10s ./...` — PASS (`ok example.com/library-lifecycle 1.458s`). The probe waits for callback entry, cancels the parent, and requires Wait to return context.Canceled within one second, without using Stop to drive cancellation.
4. `go version` — `go version go1.26.5 darwin/arm64`.

No network lookup or external service was used. Findings rely only on the supplied contract and inspected/executed artifacts.
