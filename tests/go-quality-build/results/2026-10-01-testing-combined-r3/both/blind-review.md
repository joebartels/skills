# Independent review of the supplied indexer changeset

The review compares the complete supplied `original/` and `candidate/` snapshots under `/private/tmp/go-independent-review-xcpv8czo`; no Git revision identifiers were supplied. The original implements Put/Get and stubs ApplyCSV, Refresh and Serve. README.md is unchanged and supplies the behavior and process contracts. The review inspected every supplied production and test file, plus the three supplied review skills and their local decision references. No author report, evaluation expectations, other candidate or repository was consulted. Source files in both snapshots were preserved. Deliberately broken implementations were created only in named disposable mutation copies.

Paths below are relative to `/private/tmp/go-independent-review-xcpv8czo/candidate/`. Grades assess introduced/worsened changes and the requested regression-test scope, not unrelated legacy debt. No substantiated actionable findings were found.

## Testing — A+

Scope: Complete original-to-candidate changeset for the file-backed Go library and `cmd/indexer` process. The module remains `go 1.22`, standard-library only. Executed checks used Go 1.26.5 on Darwin/arm64.

Coverage: Assessed CSV parsing/validation/ordered duplicates/accepted-prefix retention; Put compatibility and independent roots; Refresh request method/context, body ownership, exact status, complete JSON-array decoding, ordering, preserved text, empty array, failure retention, cancellation and POSIX replacement snapshots; Serve startup, recurrence measured from completion, sequential execution, cancellation cleanup, independent invocations and error identities; actual put/apply/watch subprocess behavior, streams, status, recurrence and interrupt/termination. Assessed test isolation, cleanup, realistic HTTP and filesystem boundaries, finite deadlines, and the relevance of race detection. No fuzzing or benchmarks were added; neither is necessary to substantiate these explicit finite contracts. Exact Go 1.22 runtime execution and non-Darwin execution were not available.

Rationale: critical=0, major=0, moderate=0, minor=0 selects A; two independent verified safeguards support A+. First, the tests force partial-write failure with a child-local file-size limit and assert retained bytes, accepted-prefix effects and temporary-file cleanup. They also check an already-open reader against the replacement snapshot. Replacing atomic publication with direct WriteFile made both safeguards fail for the concrete lost-retention behavior. Second, a gated callback models cooperative cancellation cleanup; release and Serve return are checked before and after the gate opens. A nonjoining implementation failed these assertions promptly. These control independent data-integrity and lifecycle-ordering risks. Passing test count or coverage percentage was not used to determine the grade.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [TG1] `retention_unix_test.go:23`, `retention_unix_test.go:50`, `retention_unix_test.go:73` and `refresh_test.go:40` exercise genuine filesystem failure and replacement boundaries. The file-size limit and SIGXFSZ handling are confined to a deadline-bounded child, avoiding process-wide contamination of the test runner. Candidate checks pass; the direct-publication mutation loses old bytes and is rejected.
- [TG2] `serve_test.go:88`, `serve_test.go:109`, `serve_test.go:121` and `serve_test.go:135` coordinate cancellation and callback cleanup through channels, with rescue cleanup and bounded observation. The release-before-join mutation fails at `serve_test.go:129`/`:131`; it does not merely reach a global timeout.
- [TG3] `csv_test.go:45`, `csv_test.go:60`, `csv_test.go:94` and `csv_test.go:105` check exact parsed text, duplicate replacement, retained accepted state, stopped later writes, source-error identity and failed publication. These assert the requested partial-success behavior instead of inventing transaction rollback.
- [TG4] `refresh_test.go:71` checks read/decode/status/validation rejection against exact prior bytes and body-close counts. `refresh_test.go:160` uses a real HTTP redirect to verify one GET and the caller's unchanged redirect policy; `refresh_test.go:194` verifies network cancellation and fixture cleanup. Omitting body closure is rejected by success and rejection assertions.
- [TG5] `cmd/indexer/process_test.go:23`, `:92`, `:144` and `:187` build and run the real executable. They assert process status and both output streams, accepted-prefix files, recurring successful publication, SIGTERM shutdown and interruption of active HTTP work. A changed `os.Exit` status is detected by the executable tests. Helpers bound build, one-shot command and watch lifetimes at `:220`, `:244` and `:267`.
- [TG6] Public function-type assignments in `csv_test.go:14`, `refresh_test.go:19` and `serve_test.go:15` exercise consumer source compatibility. The full race run and five shuffled race repetitions of filesystem/lifecycle tests pass, providing executed evidence beyond inspection. Process helpers use explicit application environments and private temporary paths; callback state is per invocation.

Bad

None found.

Suggested changes

None needed.

Limits: Full standard and race suites pass after approved execution outside the sandbox allowed the localhost HTTP fixtures to bind. Their initial sandbox failures were `bind: operation not permitted`, not assertion or production failures. Race detection covers exercised paths; the executable built by process tests is a normal Go binary, not independently race-instrumented. Exact Go 1.22 execution, Windows replacement behavior and a broader OS matrix were not run. The Unix partial-write boundary was verified on Darwin; it is build-tagged `unix`. No claim is made about injected filesystem Close failures, crash durability, arbitrary custom malicious dependencies or exhaustive parser inputs. Those are not demonstrated gaps in the stated scope. Exact commands, stdout, stderr and exit codes appear in checks.json.

## Correctness & Compatibility — A+

Scope: Complete original-to-candidate changeset for the Go 1.22 library and existing command, including preservation of exported signatures and put behavior. README.md defines the supported POSIX replacement and process semantics.

Coverage: Traced success, empty input, validation/parse/read/fetch/status/publication failures, retained state, duplicate ordering, JSON representation, cancellation and release error identities, sequential recurrence, context forwarding, root independence, existing consumer function types, CLI grammar/status/streams and signal handling. Inspected effective module version and all newly used language/library features; `go vet -stdversion ./...` passes. Exact Go 1.22 execution and Windows replacement semantics were not assessed dynamically.

Rationale: No supported behavior regression or failed new operation was substantiated: critical=0, major=0, moderate=0, minor=0. A+ is justified by two independently verified implementation safeguards. The same-directory temporary-file boundary publishes only after successful write and close, protecting prior destination bytes and open-reader snapshots; partial writes and rejected responses were exercised, and bypassing that boundary is detected. Separately, the synchronous callback establishes a join before release and avoids overlap, protecting callback cleanup ownership; gated cancellation, elapsed completion-based timing, both error identities and independent invocations were executed, and releasing without joining is detected. These safeguards address distinct consequential contracts rather than duplicate descriptions of one mechanism.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [CG1] `index.go:34` writes and closes a same-directory temporary file before rename, with deferred removal. `index.go:28` preserves the existing Put behavior, including arbitrary and empty text. `refresh.go:68` reuses the publication boundary. Candidate partial-write checks retain exact old or earlier accepted bytes; opened old readers and later new readers observe complete separate snapshots.
- [CG2] `index.go:59` reads and validates one CSV row at a time, then calls Put before continuing. `index.go:68` retains the validation sentinel for invalid field count and `index.go:73` preserves key/text validation identity. The accepted-prefix tests verify both retained replacement text and missing later rows across validation, malformed CSV, read failure and publication failure.
- [CG3] `refresh.go:21` forwards the supplied context; `refresh.go:28` copies the client and rejects redirect follow-up; `refresh.go:36` owns the acquired body close. `refresh.go:37` accepts only 200, `refresh.go:40` completes the read, `refresh.go:49` rejects null and `refresh.go:53` rejects trailing data. Every record is validated before compact JSON plus newline is published. Executed tests verify exact ordered text, empty `[]` output, retained prior bytes and unchanged caller redirect policy.
- [CG4] `serve.go:16` rejects invalid intervals before registering release. `serve.go:19` joins callback and release errors; `serve.go:24` runs the callback synchronously; `serve.go:30` starts the interval timer after callback completion. The startup, sequential recurrence, cooperative cleanup, error-combination and independent-invocation tests pass. `serve.go:43` inspects cancellation error leaves so a joined independent failure survives; substituting a broad `errors.Is` suppression made the corresponding test fail.
- [CG5] `cmd/indexer/main.go:28` preserves positional put arguments, including leading dashes and blank text. `:34` applies CSV from stdin. `:61`/`:69` enforce watch startup requirements; `:73` composes Refresh and Serve and closes idle connections through release. `:82` confines signal ownership to the process, and `:85`/`:86` preserve diagnostics and status 2. Real process tests verify these behaviors, successful watch recurrence and graceful cancellation of active network work.
- [CG6] The module remains Go 1.22 with no dependencies. Exported function types remain unchanged and consumer type assignments compile. The standard-version vet check reports no too-new standard-library symbols; inspection found no post-1.22 language dependency.

Bad

None found.

Suggested changes

None needed.

Limits: Go 1.26.5/Darwin/arm64 was the installed execution environment. Go 1.22 compatibility has static/API-version evidence and retained module declaration, not an exact Go 1.22 run. Windows replacement semantics are explicitly outside README's execution claim. External network services were not used; HTTP checks used owned localhost fixtures. A clean race run proves only exercised concurrency paths. No crash-durability, fsync, transaction rollback, nil-dependency, or hard cancellation of noncooperating callbacks is inferred from the README. Unchanged lack of those guarantees is not graded as newly introduced debt. Exact executed results are in checks.json.

## Architecture & Design — A

Scope: Complete original-to-candidate changeset: new library operations and CLI composition, shared publication helper, callback lifecycle and dependency ownership. Architecture applies because production API implementation, dependency handling and lifecycle structure changed.

Coverage: Assessed package responsibility, exported consumer signatures, concrete HTTP client and io.Reader boundaries, function callbacks, state ownership, publication reuse, error/context propagation and the library/process boundary. Read all production callers and their new regression tests. No independent deployment system, persistence schema migration or multi-service architecture is supplied or required.

Rationale: critical=0, major=0, moderate=0, minor=0 selects A. The design fits a small standard-library library/command: existing packages remain cohesive, dependencies and shutdown are explicit, and reusable invariants are factored without extra interfaces or layers. Relevant strengths are verified by dependency observation, actual process behavior and lifecycle tests. No unsupported package/interface/DTO requirements were imposed. I do not award additional architectural credit merely because the same correctness safeguard has several tests.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [AG1] `index.go:34` owns atomic filesystem publication as one private helper used by Put and Refresh. `index.go:82` shares only the validation invariant that applies to new record-based operations, so legacy Put retains arbitrary text. The retained-state tests exercise the shared boundary through all three operations and demonstrate that the extraction did not dilute the contracts.
- [AG2] `refresh.go:20` receives the caller's context, concrete client and destination; it borrows the client and closes only its acquired body. Its client copy at `:28` enforces one-request redirect behavior without changing the caller's policy. The observed transport/body tests and real redirect test verify this ownership boundary.
- [AG3] `serve.go:15` expresses its real dependency protocol through two function parameters and per-call state. The synchronous callback at `:24` makes release ordering clear; the deferred release at `:19` preserves both errors. Gated cleanup and simultaneous independent-invocation tests verify usable lifecycle ownership without an extra background worker or coordination abstraction.
- [AG4] `cmd/indexer/main.go:73` composes the library with HTTP configuration and idle-connection release; signal installation and os.Exit live only in main at `:81`. Library consumers keep control of their own process and contexts. Executable tests verify the resulting startup and shutdown behavior, and consumer function-type assertions compile.

Bad

None found.

Suggested changes

None needed.

Limits: This grade covers the supplied small file-backed library and command, not hypothetical larger deployment or service needs. Exact Go 1.22 execution and non-Darwin platform behavior were not run; static version inspection and the standard-version vet check support the declared minimum. No narrower architecture topic was unassessable within the supplied production changes. Dependencies/reproducibility, security, performance and operational enforcement were not graded as separate topics. Verification evidence and its environment limits are recorded below and in checks.json.

## Supporting verification facts

Every Go command used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local` through `rtk proxy env`. checks.json preserves the exact argv, working directory, complete stdout/stderr, exit result and elapsed time for executed verification commands. source-hashes.json identifies the supplied snapshot artifacts by SHA-256. No tools or dependencies were installed.

| Check | Result and interpretation |
| --- | --- |
| `go version`; `go env GOOS GOARCH GOVERSION GOTOOLCHAIN` | Exit 0; Go 1.26.5, Darwin/arm64, local toolchain. |
| `go test ./... -count=1 -timeout=60s` in sandbox | Exit 1 because localhost listen was denied; recorded as an environment limit. |
| `go test -race ./... -count=1 -timeout=60s` in sandbox | Exit 1 for the same listen restriction; no candidate failure inferred. |
| Same full standard suite with approved sandbox escalation | Exit 0; library 0.718s, command package 1.017s. |
| Same full race suite with approved sandbox escalation | Exit 0; library 4.590s, command package 1.862s. |
| `go vet -stdversion ./...` | Exit 0; empty stdout/stderr. |
| Selected filesystem/lifecycle tests with `-race -shuffle=on -count=5 -timeout=30s` | Exit 0; 18.295s; exact test selector is retained in checks.json. |
| Direct-publication mutation | Expected exit 1: old-reader assertions and all put/apply/refresh partial-write retention children reject overwritten prior bytes. |
| Broad-cancellation-suppression mutation | Expected exit 1: joined independent callback failure is lost, detected at serve_test.go:143. |
| Missing-body-close mutation | Expected exit 1: success and rejection body-close counts are zero. |
| Wrong-process-status mutation | Expected exit 1: real executable status 1 is rejected where status 2 is required. |
| Release-before-join mutation | Expected exit 1: gated cancellation tests reject release/return overtaking callback cleanup at serve_test.go:129/131. |

Mutation checks intentionally change only disposable copies: `mutation-direct-publication/`, `mutation-discard-independent-error/`, `mutation-forget-body-close/`, `mutation-wrong-process-status/`, and `mutation-release-before-join/`. Their expected failures substantiate assertion signal and are not findings against the candidate. Each check has a finite test timeout and an outer subprocess deadline. The candidate's HTTP tests use owned httptest servers and real process tests use private temporary paths, controlled application environments, cancellation, and bounded waits.
