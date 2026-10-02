# Independent Poll review — 2026-10-01

Boundary: the supplied diff from `/private/tmp/go-independent-review-4qorzxhe/original` to `/private/tmp/go-independent-review-4qorzxhe/candidate`. This is a standard-library-only Go library declaring Go 1.22. The actual request was to replace guessed readiness and add reliable recurrence, cancellation, fixture teardown, and independent-invocation tests while preserving the public API and minimum Go version. The complete supplied changeset changes only `poll_test.go`; `poll.go`, `README.md`, and `go.mod` are byte-identical. No author report, evaluation expectations, other candidate, or unrelated repository source was consulted. Original and candidate source were not edited; probes are retained under this review directory's `mutations/`.

Paths below are relative to `candidate/` unless explicitly qualified. Grades assess introduced/worsened changes and the requested test scope, not unrelated legacy debt.

## Testing — A+

Scope: supplied original/candidate diff, public `Poll` library contract in `README.md:3`, Go 1.22 module semantics; execution with Go 1.26.5 on darwin/arm64.
Coverage: immediate first callback, pre-cancellation, invalid intervals, exact callback-error identity, recurring sequential completion-relative callbacks, cancellation during callback cleanup, simultaneous callback error and cancellation, caller-owned fixture lifetime, concurrent invocations, individually selectable parallel subtests, failure-path cancellation/join, and race/order checks were assessed. Fuzzing and benchmarks are not relevant to this finite coordination change. No external dependencies or omitted project files are needed for the supplied scope.
Rationale: no actionable issue was established. Two independent verified safeguards support A+: (1) the event/gate and completion-timestamp recurrence test detects start-relative scheduling after a long callback, controlling recurrence ordering/timing risk; (2) the blocked callback-cleanup test and helper stop/join protocol detect premature `Poll` completion and preserve worker teardown after an early test failure, controlling callback/resource lifetime risk. These safeguards address different documented contracts; test count and coverage percentage were not used to grade.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `poll_test.go:94`, `poll_test.go:160`, and `poll_test.go:167` acknowledge startup, hold the first callback beyond an interval, measure the next start from a preceding callback's final deferred work, check active counts, and bound event waits. In the disposable start-relative mutant, `TestRecurrence` failed because callback 2 started 1.625µs after completion instead of at least 40ms. Correct-source repeated/race/shuffled runs passed. This verifies useful recurrence signal without a guessed startup delay.
- [T-G2] `poll_test.go:179` observes cancellation and explicitly blocks deferred callback cleanup before checking that `Poll` remains running; it also checks cleanup completion after joining. The asynchronous-return-on-cancellation mutant failed at `poll_test.go:215` with “Poll returned while callback cleanup was blocked.” This tests the documented synchronous lifetime contract directly.
- [T-G3] `poll_test.go:398` registers cancellation/gate release and a bounded join before launch. A disposable forced `t.Fatal` while callback cleanup was blocked reported the intentional failure and then “early Fatal released callback cleanup gate and joined Poll,” with no additional error or race. Removing that gate release in a separate disposable copy produced the five-second join error and both worker/cleanup-still-running diagnostics. This verifies the early-failure teardown path rather than inferring it from passing happy paths.
- [T-G4] `poll_test.go:255` registers the file close before the helper's later cleanup, so the helper joins first; it verifies the callback's deferred write and the file's continued availability before caller teardown. Each independent child at `poll_test.go:351` owns its context, error, counters, and temporary path. Both `alpha` and `beta` passed ten race-enabled runs selected alone; the complete suite passed ten race-enabled shuffled runs. No sibling-run prerequisite or shared mutable fixture was found.
- [T-G5] `poll_test.go:39` retains callback-error verification and strengthens it to the README's exact identity contract; `poll_test.go:232` checks error priority over simultaneous cancellation. Separate wrapping and cancellation-masks-error mutants failed their targeted assertions. The immediate/pre-canceled/invalid-interval tests also assert callback counts instead of relying solely on return values.

Bad

- None found.

Suggested changes

- None needed.

Limits: the tests use real monotonic time and bounded observation windows after acknowledged events. Repeated passing runs and killed mutations provide finite evidence, not proof of every scheduler interleaving or arbitrarily slow host. Five-second event/join bounds explicitly fail stalled runs. The actual Go 1.22 toolchain and other platforms were not run; `go vet -stdversion` passed under Go 1.26.5, and the new APIs/language constructs were inspected against the retained 1.22 minimum. No test clock or exported hook is required by this project's contract.

## Correctness & Compatibility — A

Scope: supplied original/candidate diff and preserved public `Poll` contract, Go 1.22 module semantics, darwin/arm64 verification with Go 1.26.5.
Coverage: production input validation, cancellation, error priority/identity, callback ordering and context, synchronous completion, timer lifecycle, public signature/configuration preservation, and concurrent test-helper publication/cleanup were inspected. The complete supplied source and contract were available. No persistence, serialization, CLI, or third-party consumer migration boundary is implicated.
Rationale: no introduced or worsened functional or consumer-compatibility issue was established. The production function and module are unchanged; the new tests exercise the documented behavior and use test-local state with synchronized error publication. These verified strengths support A. The grade does not claim a new production safeguard or upgrade-path change deserving additional credit.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `poll.go:10` and `go.mod:3` are byte-identical to original. The public signature, Go 1.22 directive, standard-library dependencies, and production control flow are preserved. Byte comparisons and hashes are recorded in `checks.json`.
- [C-G2] `poll.go:15` checks prior cancellation, `poll.go:18` calls the callback synchronously and returns its exact error before considering cancellation, and `poll.go:21` starts the timer only after callback completion. The new normal/race tests passed these paths; the recurrence, early-return, and cancellation/error mutations demonstrate that the corresponding assertions distinguish contract violations.
- [C-G3] `poll_test.go:393` publishes `run.err` by closing `done`, and `poll_test.go:420` reads it after receiving that completion. Atomic callback counters and per-invocation contexts/paths avoid shared test state. Ten race-enabled shuffled full-suite runs and ten separately selected runs per independent child reported no races. The forced early-failure probe also passed its cleanup assertions under the race detector.

Bad

- None found.

Suggested changes

- None needed.

Limits: no production behavior was introduced by this diff. The review does not grade unsupported inputs, unrelated hypothetical legacy behavior, or platforms/toolchains outside the executed darwin/arm64 Go 1.26.5 environment. Actual execution on Go 1.22 remains unverified; the retained module directive, inspected API usage, and successful `stdversion` check support the assessed minimum-version compatibility without substituting for that execution. Cancellation/readiness interleavings were exercised finitely rather than exhaustively.

## Architecture & Design — Not applicable

Scope: the supplied test-only diff.
Coverage: production public API, package boundary, dependency list, callback ownership, and timer lifetime were compared with original; test helper ownership was assessed under Testing.
Rationale: no production seam, abstraction, lifetime, dependency, or package decision changes consequentially. `poll.go`, `go.mod`, and the ownership contract in `README.md` are unchanged. New private test helpers do not add consumer-visible hooks or alter production composition.
Limits: this is non-applicability for the changeset, not a general architecture grade for existing `Poll`.

## Supporting verification facts

`output/checks.json` preserves each verification's exact argv, working directory, environment overrides, stdout, stderr, exit code, duration, and timeout. All Go subprocesses use `rtk proxy go` with `GOCACHE=/private/tmp/go-quality-testing-cache`, `GOTOOLCHAIN=local`, and `GOPROXY=off`. There was no network dependency. Toolchain: `go version go1.26.5 darwin/arm64`; `CGO_ENABLED=1`.

| Check | Result |
| --- | --- |
| Original `go test -count=1 -timeout=20s ./...` | Exit 0; baseline passes. |
| Candidate `go test -count=1 -timeout=20s ./...` | Exit 0. |
| Candidate `go test -race -count=10 -shuffle=20261001 -timeout=40s ./...` | Exit 0; no race diagnostics. |
| Each candidate child selected alone: `go test -race -count=10 -run=^TestIndependentInvocations$/^alpha$ -timeout=20s ./...`, likewise `beta` | Both exit 0; no race diagnostics. |
| Candidate `go vet -stdversion ./...` | Exit 0; stdout/stderr empty. |
| Start-relative recurrence mutant, targeted `TestRecurrence` | Exit 1; lower-bound assertion failed at 1.625µs versus 40ms. |
| Early asynchronous return mutant, targeted callback-cleanup test | Exit 1; returned with cleanup blocked. |
| Cancellation masks callback error mutant | Exit 1; nil returned instead of original callback error. |
| Callback-error wrapping mutant | Exit 1; wrapped error rejected by exact identity assertion. |
| Forced early `t.Fatal` with original stop/join helper, race-enabled | Exit 1 as intentionally induced; post-helper probe confirmed gate release and joined worker, without additional errors or race reports. |
| Same forced failure with cleanup gate release removed | Exit 1 after finite five-second helper bound; join and post-helper checks diagnosed the blocked worker. |
| Final production/configuration comparisons and source hashes | Exit 0; original/candidate production, README, and module match. Probe edits are confined to disposable copies. |

No actionable finding was verified, and no legacy limitation was charged against the changeset. Mutation failures above are successful regression-detection evidence, not defects in the candidate.
