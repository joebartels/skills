# Independent review of the periodic-worker changeset

The supplied boundary is `original/` versus `candidate/` under `/private/tmp/go-independent-review-wya6bd9e`; there are no supplied Git revision identifiers. This review used only the dispatch, those two trees, and the three supplied review skills and their references. It did not use an author report, evaluation expectations, or another candidate. References below are to the candidate unless otherwise stated. `source-manifest.json` records SHA-256 identifiers for the supplied files.

The original README already requires immediate first work, completion-relative intervals, serial callbacks, cancellation as a normal worker result, error identity, joined release failures, and cancel/wait/release ownership. The candidate's appended README clarifies callback-error precedence, child-context ownership, and blocking release; these are consistent refinements of that lifecycle. The original one-shot implementation and its single failure/release test are context, not introduced findings.

Architecture is applicable because the change introduces an owned goroutine and child context, changes the internal worker protocol, and makes shutdown ordering consequential.

## Testing — A+
Scope: Full supplied Go library changeset, including the rewritten public-boundary test suite in `/private/tmp/go-independent-review-wya6bd9e/candidate/host_test.go`. Module declares Go 1.22 and standard-library dependencies; executed on Go 1.26.5 darwin/arm64.
Coverage: Assessed immediate work, already-canceled input, zero/negative intervals without effects, multiple successful cycles, completion-relative delay, non-overlap, later callback failure, cancellation during work and interval waiting, callback cleanup before release/return, cancellation/error precedence, child-context cancellation, caller lifetime, joined causes, blocking release, independent Run lifetimes, test synchronization, cleanup ownership, and compile-time function-value compatibility. No parsing, external I/O, benchmark, or fuzz-input boundary is introduced.
Rationale: No actionable testing issues were substantiated. Two independent safeguards go beyond ordinary happy-path setup: the deliberately blocked first cycle exposes fixed-rate/catch-up scheduling and verifies completion-relative intervals; gated callback cleanup exposes premature release/return and checks the resource lifetime. Both passed for the candidate and failed against corresponding disposable mutations. These control distinct scheduling and teardown risks, supporting A+. Additional error-identity and context-ownership mutations also failed at meaningful assertions. Grade is based on behavior and mutation signal, not test count or coverage percentage.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `host_test.go:138` deliberately holds the first callback longer than an interval, measures later starts against completion, tracks in-flight calls, and checks exactly three sweeps and one release. A ticker-from-start mutation failed at `host_test.go:201`: cycle two began 2.708µs after completion instead of at least 60ms. This demonstrates detection of a plausible scheduler regression rather than simply observing repeated calls.
- [T-G2] `host_test.go:218` waits for cancellation to reach callback cleanup, keeps that cleanup blocked, observes premature release/return, then checks cleanup ownership, error identity, and exact call counts. An early-release mutation failed in both subcases at `host_test.go:262` and `host_test.go:259`. This protects callbacks from teardown while they still use host-owned resources.
- [T-G3] `host_test.go:283` asserts `errors.Is` for both worker and release errors and checks that the worker context is canceled before release while the caller remains active after callback failure. Discarding callback errors failed at `host_test.go:321`; omitting worker-owned cancellation failed at `host_test.go:331`.
- [T-G4] `host_test.go:337` uses an explicit release gate to verify Run waits for release and preserves both causes; `host_test.go:365` observes the second worker progressing after the first has joined. These passed in ordinary and repeated shuffled race runs, exercising release ownership and invocation isolation.
- [T-G5] Helpers at `host_test.go:24`, `host_test.go:45`, and `host_test.go:59` use completion channels and finite timers, call fatal assertions in the test goroutine, and synchronize error reads through the closed completion channel. Gate cleanups are registered after Run cleanup and therefore run first. The timed negative observations check elapsed-time/blocked-lifetime behavior rather than guessing when asynchronous startup has occurred.

Bad

- None found.

Suggested changes

- None needed.

Limits: Ten shuffled race runs passed, but finite stress runs cannot prove every possible interleaving. Wall-clock interval tests ran locally, not on an overloaded CI host. Go 1.22 itself was not executed; the declared language version, absence of newer APIs, `go vet -stdversion`, and an explicit legacy timer-channel run assess the relevant compatibility risks. There is no supplied CI configuration or broader supported-platform matrix. Exact commands, stdout, stderr, exit status, and durations are in `checks.json`.

## Correctness & Compatibility — A+
Scope: Full supplied library changeset from the one-shot host to the periodic worker; `/private/tmp/go-independent-review-wya6bd9e/candidate/host.go`, `worker.go`, README, module, and consumer-boundary tests. Public Run signature and Go 1.22 declaration are retained.
Coverage: Traced validation, initial cancellation, serial successful cycles, completion-relative waiting, cancellation during waiting and callback execution, callback errors concurrent with cancellation, release ordering and error joining, child/parent context ownership, and independent concurrent invocations. Reviewed effective language version and standard-library API availability; executed finite normal, race, timer-mode, and mutation checks.
Rationale: No introduced or worsened correctness or compatibility issue was substantiated. Two independently verified safeguards control meaningful distinct failures: synchronous callback execution followed by a fresh timer prevents overlap and stale-tick catch-up; host-owned cancellation with callback completion before release prevents resource teardown during callback cleanup. The scheduler and premature-release mutations fail tests that pass on the actual implementation. These safeguards, together with verified identity-preserving errors, support A+. No systemic or critical reach is claimed.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `worker.go:10` checks cancellation before each cycle; `worker.go:13` executes the callback serially; `worker.go:17` starts the interval only after successful completion. The successful-cycles test and ticker mutation establish completion-relative scheduling. The already-canceled test confirms no initial callback, and the hour-interval test confirms cancellation promptly exits a pending interval.
- [C-G2] `host.go:21` derives a per-invocation child context; `host.go:27` receives worker completion only after the callback returns; `host.go:28` cancels before `host.go:29` invokes release. Parent cancellation reaches a callback through the child context even while the host waits. Gated cleanup and context-ownership tests pass, and both early-release and omitted-child-cancellation mutations fail. Release itself runs synchronously before Run returns.
- [C-G3] `worker.go:13` returns a callback error before considering a subsequent cancellation result, preserving the documented callback-error precedence. `host.go:29` joins worker and release causes. Tests verify normal cancellation is nil, callback failures remain inspectable during cancellation, and combined failures preserve both identities; discarding the callback error is detected.
- [C-G4] `host.go:17` rejects invalid intervals before creating work or invoking release, confirmed for zero and negative values. The exported signature is unchanged, and the external-package function-value assertion at `host_test.go:15` compiles. No dependency, build-tag, or module-version change is introduced.

Bad

- None found.

Suggested changes

- None needed.

Limits: Execution used Go 1.26.5 darwin/arm64 with CGO enabled, not a Go 1.22 executable or another platform. Go commands explicitly set GOTOOLCHAIN=local and GOPROXY=off, preventing toolchain/dependency downloads, and use the required disposable GOCACHE. Source inspection and version-aware vet found no too-new API/language feature; explicit legacy timer-channel semantics passed. Unsupported nil callback/context inputs, panic recovery, callback-spawned work that outlives callback return, and external resource implementations are not promised in the supplied contract and were not invented as requirements. Checks do not establish general application behavior beyond the supplied library.

## Architecture & Design — A
Scope: Full supplied small library changeset; the new owned worker, internal periodic helper, per-call context, and exported Run lifecycle.
Coverage: Assessed package/API boundary, dependency injection through callback functions, ownership of work and release, synchronous error/result contract, cancellation propagation, and simplicity for the supplied consumer. No transport, persistence, process-global state, or new interface is involved.
Rationale: No substantiated architecture issue. The single package and small private worker keep host ownership clear without adding an unnecessary public abstraction. Verified lifecycle ownership and explicit callback dependencies support A. This card does not claim separate architectural safeguards beyond the ordinary correct composition needed by this small API, so A+ is not awarded here.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `host.go:16` retains explicit context, interval, sweep, and release dependencies and the existing function signature. The external consumer tests exercise the real exported boundary without requiring access to the private worker, and the function-value assertion verifies the preserved API shape.
- [A-G2] `host.go:21` through `host.go:29` keep cancellation, completion, release, and error ownership in one host invocation; `worker.go:8` owns only serial scheduling. Callback/release errors are returned to that caller, and no goroutine intentionally continues callback work after Run returns. The gated cleanup/release tests and two concurrent Run lifetimes verify this ownership boundary.
- [A-G3] The implementation keeps one package and standard-library dependencies. Tests use the supplied callback seam and real context/timer behavior; there is no hidden process-wide cancellation owner or production-only interface introduced to support tests.

Bad

- None found.

Suggested changes

- None needed.

Limits: Judged the API against the supplied library README and consumer tests, not an assumed service deployment. No broader application callers or architecture are supplied. Runtime portability limits are those recorded in the correctness card and verification facts.

## Supporting verification facts

Commands were prefixed with `rtk proxy`. Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache`, `GOTOOLCHAIN=local`, and `GOPROXY=off`. All builds and mutations ran in disposable copies within the authorized review directory; original/candidate production source was not repaired or changed.

- `go version`: exit 0; `go version go1.26.5 darwin/arm64`.
- `go env GOVERSION GOOS GOARCH CGO_ENABLED GOMOD GOWORK`: exit 0; Go 1.26.5, darwin, arm64, CGO 1, disposable verification module, no active workspace.
- `go test -count=1 -timeout=30s ./...`: exit 0; `ok example.com/sweeper 0.581s`.
- `go test -race -count=10 -shuffle=on -timeout=60s ./...`: exit 0; `ok example.com/sweeper 4.835s`; no race diagnostic.
- `go vet -stdversion ./...`: exit 0; stdout/stderr empty.
- `GODEBUG=asynctimerchan=1 go test -count=3 -timeout=30s ./...`: exit 0; `ok example.com/sweeper 1.234s`.
- Mutation `ticker-from-start`: targeted completion-relative scheduling test exited 1 with interval assertions at `host_test.go:201`; fixed-rate stale-tick behavior was detected.
- Mutation `release-before-callback-cleanup`: targeted cancellation/cleanup test exited 1 with early return/release assertions at `host_test.go:262` and `host_test.go:259`; unfinished cleanup was detected.
- Mutation `discard-callback-error`: targeted error-combination test exited 1 with callback-identity assertions at `host_test.go:321` for callback-only and combined errors.
- Mutation `omit-child-cancellation`: targeted error-combination test exited 1 with release-before-cancel assertions at `host_test.go:331` for callback-failure cases.
- `rtk proxy diff -ru candidate verification` with absolute authorized paths: exit 0, stdout/stderr empty; confirms the test verification copy matches the candidate. Mutations are separate copies. Raw command arrays and all actual streams remain in `checks.json`.

No findings were withheld as suspected defects; no confirmed production defect requires an ungraded related finding. No repair is proposed.
