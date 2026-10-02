# Independent review

Reviewed only the supplied `original` → `candidate` directory changeset under `/private/tmp/go-independent-review-0vvp8kdb`, its README, and the three supplied review skills/references. There were no supplied revision identifiers. This is a standard-library file-backed library plus a CLI; `go.mod` remains `go 1.22`. Findings concern introduced/worsened changes and the newly requested regression-test scope. Existing unimplemented stubs and sparse baseline tests are historical context, not graded debt. No source repairs were made.

## Testing — A+
Scope: The supplied changeset's index, CSV, refresh, lifecycle, and actual command tests; Go 1.22 declared support, execution on Go 1.26.5 darwin/arm64.
Coverage: Inspected every supplied source/test file and the full diff. Assessed consumer signatures, exact values, accepted-prefix effects, validation/error identity, HTTP response ownership, retained state, POSIX replacement, timing/sequencing, cancellation, independent invocations, process status/streams/signals, test isolation and cleanup. Ran the full suite, race detection with three shuffled repetitions, and six independent assertion-signal mutations in disposable copies. No fuzz targets or benchmarks were added; neither is necessary to substantiate these bounded contracts. Exact Go 1.22 runtime execution and Windows process/replacement behavior were not assessed.
Rationale: No substantiated actionable testing issue. Two independent safeguards justify A+: (1) the open-reader snapshot test detects destructive in-place publication, controlling loss of the promised old snapshot; (2) the gated cancellation-cleanup test detects release before the callback joins, controlling premature resource teardown. Both pass on the candidate and fail on their respective disposable mutations. These establish meaningful test signal beyond test count or routine fixture setup.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [TG1] [index_regression_test.go:66](/private/tmp/go-independent-review-0vvp8kdb/candidate/index_regression_test.go:66) checks duplicate replacement, exact quoted/multiline text, parse/validation rejection, retained accepted values, absence of later writes, and `ErrInvalidRecord` identity. Reader/write failures are exercised at line 109. Ignoring validation errors makes three candidate cases fail in `mutant-continue-invalid-csv`.
- [TG2] [refresh_test.go:33](/private/tmp/go-independent-review-0vvp8kdb/candidate/refresh_test.go:33) asserts exact destination bytes, request method/endpoint/context, one request, and one body close across success and rejection. It rejects wrong status, null/object/malformed input, trailing JSON/garbage, invalid records, and read errors after a complete array. Removing closure fails every table case at line 93 in `mutant-no-body-close`; the fake models the HTTP/body contract rather than bypassing decoding/publication.
- [TG3] [refresh_test.go:145](/private/tmp/go-independent-review-0vvp8kdb/candidate/refresh_test.go:145) opens a real old file before refresh and reads that same descriptor afterward, then verifies a later opener's exact new bytes. `mutant-in-place-publication` fails at line 173 (`open reader lost old snapshot`, only 31 replacement bytes), showing that the test distinguishes replacement from truncation.
- [TG4] [serve_test.go:61](/private/tmp/go-independent-review-0vvp8kdb/candidate/serve_test.go:61) gates a callback past the requested interval and measures the next invocation from actual completion; it also checks context identity and absence of overlap. `mutant-no-completion-delay` fails at line 126. [serve_test.go:136](/private/tmp/go-independent-review-0vvp8kdb/candidate/serve_test.go:136) synchronizes start, cancellation cleanup, and release, with bounded waits and cleanup gates; all six variants reject `mutant-release-before-join` at line 207. Callback/release error identity assertions also reject `mutant-discard-callback-error` at line 57.
- [TG5] [process_test.go:22](/private/tmp/go-independent-review-0vvp8kdb/candidate/cmd/indexer/process_test.go:22) builds and executes the real command, verifying exit 0/2, stdout/stderr, empty and leading-dash put text, stdin CSV effects, startup rejection, failed first refresh, recurring snapshots, and both interrupt/SIGTERM during an in-flight request. Build/application waits have finite deadlines. Child processes use an explicit environment; isolated temp directories and local servers avoid external services. Worker results are delivered to the test goroutine, and cleanup cancels/joins before server closure. The full three-repeat race/shuffle check passes.

Bad

- None found.

Suggested changes

- None needed.

Limits: The first sandboxed full run failed because `httptest` could not bind loopback (`operation not permitted`), not because of candidate behavior. An approved execution with local loopback access passed; the same access was used for the race/shuffle run. Go 1.22 itself was not executed or downloaded. Static `go vet -stdversion ./...` passed against the unchanged Go 1.22 directive, and no newer standard-library API/language requirement was identified in inspected source. No exhaustive fuzzing, disk-full/close-error injection, or scheduler proof is claimed. Failure-path publication was exercised with a deterministic directory collision; crash durability/fsync is not a README contract. Windows signal/replacement tests are explicitly skipped and outside the execution claim.

## Correctness & Compatibility — A+
Scope: The supplied library/CLI changeset against the explicit README contracts, including preserved exported signatures, existing put behavior, Go 1.22 support, and POSIX snapshot semantics.
Coverage: Traced valid/invalid/empty CSV and JSON, accepted-prefix processing, duplicate keys, arbitrary existing put text, reader/fetch/status/decode/publication failures, error identity, cancellation before/during callbacks and interval waits, completed-work ownership, per-invocation state, CLI parsing/status/streams, and signal shutdown. Compared all existing public signatures and unchanged Record tags to the original. Full candidate and original tests passed in disposable copies; candidate race/shuffle and stdversion checks passed. Go 1.22 and Windows runtime execution limits remain explicit.
Rationale: No introduced/worsened supported-behavior defect substantiated. Two independent safeguards justify A+: complete same-directory temporary-file publication protects stored snapshots and retained failure state, while synchronous refresh ownership plus deferred release protects cleanup ordering and callback/release errors. Real file-descriptor and gated callback checks verify these distinct risks, rather than merely observing successful output.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [CG1] [index.go:22](/private/tmp/go-independent-review-0vvp8kdb/candidate/index.go:22) retains the existing key rule, arbitrary text behavior, root creation, and exact Get semantics. `ApplyCSV` at line 41 validates before calling Put and stops at the first rejected row/write/read, retaining earlier accepted rows. `errors.Join`/`%w` retain validation and I/O causes. Consumer function assignments at [index_regression_test.go:14](/private/tmp/go-independent-review-0vvp8kdb/candidate/index_regression_test.go:14), existing tests, and real command put/apply tests pass.
- [CG2] [refresh.go:22](/private/tmp/go-independent-review-0vvp8kdb/candidate/refresh.go:22) uses the supplied context/transport and rejects redirects without issuing a second request or mutating the caller's redirect policy. It closes the acquired body, accepts only status 200 and one non-null array, rejects trailing data/read failure, validates all records before publication, retains order/original text, and encodes an empty array as `[]\n`. Fake-boundary cases and real redirect/cancellation tests pass; [refresh_test.go:181](/private/tmp/go-independent-review-0vvp8kdb/candidate/refresh_test.go:181) independently confirms the caller can still follow its own redirect policy afterward.
- [CG3] [index.go:70](/private/tmp/go-independent-review-0vvp8kdb/candidate/index.go:70) writes/closes a temp file in the destination directory before rename and removes the temp on failure. It does not truncate the previous destination before fetch/decode/validation/publication succeeds. Exact retained bytes, deterministic rename failure/cleanup, and the POSIX old-descriptor/new-opener snapshot contract all pass. The destructive-write mutation is detected.
- [CG4] [serve.go:15](/private/tmp/go-independent-review-0vvp8kdb/candidate/serve.go:15) rejects invalid intervals before callbacks/release, checks cancellation before starting work, calls callbacks synchronously with the original context, and starts each interval after completion. Deferred release joins errors after the last callback returns. `cancellationOnly` walks joined/wrapped causes so a separate callback failure joined with cancellation is retained. The candidate passes invalid/pre-canceled, timing, cancellation cleanup, combined error, interval cancellation, and independent invocation tests; premature release and dropped callback error mutations fail.
- [CG5] [main.go:22](/private/tmp/go-independent-review-0vvp8kdb/candidate/cmd/indexer/main.go:22) preserves positional put parsing and adds explicit apply/watch parsing. URL scheme/host, nonempty file and positive interval are checked before watch setup. Signal cancellation reaches Refresh through Serve, and idle connections close only in Serve's release callback. Main retains exit 2 with stderr on failure and no success stdout. Real process tests pass through recurring refresh, cancellation of request three, clean exit for both signals, and retention of round-two bytes.

Bad

- None found.

Suggested changes

- None needed.

Limits: Executed on Go 1.26.5 darwin/arm64. The declared Go 1.22 language/API baseline was inspected and checked with stdversion, but not run with an actual Go 1.22 binary. HTTP checks use controlled transports and local real servers; no internet endpoint is involved. Supported POSIX replacement was verified on this filesystem; Windows replacement/signal behavior is outside scope. No crash durability, all-or-nothing CSV transaction, callback panic recovery, or non-cooperative callback termination is promised by the supplied contract, and none is inferred. A passing race check covers exercised concurrent paths rather than proving every interleaving.

## Architecture & Design — A
Scope: Production structure introduced by the supplied changeset: one library package and one existing command package, shared atomic publication, CSV/HTTP operations, refresh/release callback ownership, and CLI composition.
Coverage: Assessed package/API boundaries, data/JSON boundaries, dependency visibility, abstraction size, instance/caller ownership, error/cancellation propagation, process-wide effects, and proportionate design. Read all affected implementations/callers/tests and the module declaration. No additional package/interface framework was introduced. Deployment topology and long-term unprovided consumers are not inferred.
Rationale: No actionable design issue. The direct library/command composition fits the stated small-module task, preserves consumer signatures, and makes storage and refresh lifecycle ownership clear. Runtime tests verify those ownership claims. A is supported by these observed strengths without awarding an additional design grade for complexity or routine composition.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [AG1] [index.go:70](/private/tmp/go-independent-review-0vvp8kdb/candidate/index.go:70) centralizes the existing publication invariant for Put and Refresh without changing Put's public contract or requiring an extra abstraction layer. CSV remains an ordered consumer of Put; Refresh validates a complete external snapshot before calling the same publication boundary. Real retained-state and open-descriptor tests verify that this reuse preserves the required ownership semantics.
- [AG2] [serve.go:15](/private/tmp/go-independent-review-0vvp8kdb/candidate/serve.go:15) uses two small function dependencies and the caller's context, with synchronous execution making join/release order explicit. It has no hidden background worker, shared invocation state, signal handler, or library-owned client. Gated cancellation tests and independent invocation tests pass; the premature-release mutation fails, confirming that the ownership boundary is meaningful.
- [AG3] [refresh.go:18](/private/tmp/go-independent-review-0vvp8kdb/candidate/refresh.go:18) borrows the supplied concrete HTTP client. The local copy changes only the per-operation redirect behavior required for a single GET, while preserving caller transport/timeout policy. The real redirect test confirms that caller-owned policy survives. [main.go:57](/private/tmp/go-independent-review-0vvp8kdb/candidate/cmd/indexer/main.go:57) owns process signals and client release at the command boundary; real signal tests verify orderly shutdown.
- [AG4] Existing `Record` is already the JSON transport/persistence record; retaining its lowercase tags avoids an unnecessary duplicate model. Wrapped and joined errors convey the explicitly promised sentinel/callback/release identities through the existing error-returning APIs. Signature assignment and error-identity tests pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: This is a supplied complete small-module changeset, not a review of undisclosed deployment/consumer architecture. Filesystem crash durability and Windows behavior are not architectural requirements in the README. Runtime evidence and toolchain limitations are shared with the preceding report cards. Architecture is applicable because this change implements production dependency and lifecycle boundaries; there is no narrowly unsupported architecture topic requiring an additional grade.

## Supporting verification facts

Exact command, working directory, stdout, stderr, exit code, elapsed time, and finite process timeout are preserved in [checks.json](/private/tmp/go-independent-review-0vvp8kdb/output/checks.json). Each check used `rtk proxy`; every Go command set `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Disposable mutation definitions are in [mutations.json](/private/tmp/go-independent-review-0vvp8kdb/output/mutations.json).

| Check label | Result and meaning |
| --- | --- |
| `toolchain` | Exit 0: `go version go1.26.5 darwin/arm64`. |
| `baseline-tests` | Exit 1: sandbox rejected loopback listener creation. This was an execution-environment limit, subsequently resolved by approved loopback-capable execution. |
| `baseline-tests-loopback` | Exit 0: `go test -count=1 -timeout=60s ./...`; library and actual command-process suite passed. |
| `race-shuffled-repeat` | Exit 0: `go test -race -shuffle=on -count=3 -timeout=60s ./...`; both packages passed, no race diagnostics. |
| `go122-stdversion-vet` | Exit 0, no output: `go vet -stdversion ./...`; no standard-library symbol newer than the module's effective language version diagnosed. |
| `original-tests` | Exit 0: original put/command baseline tests passed. |
| `mutant-no-body-close` | Expected exit 1: response ownership assertions report one request and zero closes. |
| `mutant-in-place-publication` | Expected exit 1: old descriptor reads the 31-byte replacement rather than the original snapshot. |
| `mutant-discard-callback-error` | Expected exit 1: combined-error assertion observes only the release error. |
| `mutant-no-completion-delay` | Expected exit 1: next callback follows completion by approximately 19 microseconds rather than the 60ms interval. |
| `mutant-continue-invalid-csv` | Expected exit 1: rejected records incorrectly return success. |
| `mutant-release-before-join` | Expected exit 1: all six cancellation-cleanup variants report release overtaking cleanup. |
| `candidate-preservation` | Exit 0, no diff: the verification baseline remains byte-for-byte identical to candidate source/configuration. All mutations were applied only to distinct disposable copies. |

No executed check timed out. Mutation failures are deliberate verification evidence, not defects in the candidate. No introduced/worsened finding was substantiated; the report makes no merge or deployment recommendation beyond these three scoped topic grades.
