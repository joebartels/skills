# Independent review of the anonymized indexer changeset

Reviewed only the supplied `original/` and `candidate/` trees, project README, and the three supplied review skills and their decision references. No author report, other candidate, or external repository was inspected. Original and candidate source remain unmodified. References below are relative to `/private/tmp/go-independent-review-y9l6zknd/candidate/` unless a verification copy is named.

The exact supplied diff replaces the ApplyCSV, Refresh, and Serve stubs, extracts Put's existing atomic replacement into a shared helper, adds apply/watch command handling, and expands regression tests. The README is the behavior authority. Unimplemented original operations are the starting point, not additional defects counted against this changeset.

## Testing — B
Scope: Supplied original/candidate changeset; file-backed library and real command processes; module declares Go 1.22 and uses only the standard library.
Coverage: Assessed assertions for existing Put/Get compatibility, independent roots, CSV ordering/duplicates/accepted-prefix failures, JSON/status/read/validation/publication rejection, retained bytes, acquired body closure, borrowed HTTP client behavior, POSIX old-reader snapshot behavior, Serve startup/error/cancellation/recurrence/independent invocation behavior, and actual command exit codes/streams/signals. Inspected test resource ownership, bounded waits, child environment isolation, and loop semantics. Ran the full suite, race detection, shuffled repetition, focused mutation checks, and a valid-error reproduction. Fuzzing and benchmarks are not present and are not required for the demonstrated contracts.
Rationale: One moderate gap leaves a supported callback/cancellation error representation unchecked, and the reproduced panic escapes the otherwise strong cancellation suite. This is contained: cancellation cleanup, ordinary cancellation, error joins, timing, and processes are meaningfully tested; the entire lifecycle contract is not effectively unverified. Existing tests demonstrate real regression signal rather than merely exercising statements.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [TG1] `refresh_test.go:184` opens the old snapshot before replacement and then checks both old-descriptor bytes and new-opener bytes. Replacing the rename with an in-place write in a disposable copy made this test fail at line 207. This detects the expressly required POSIX publication property.
- [TG2] `serve_test.go:108` coordinates cancellation and holds cooperative cleanup behind a gate; it checks release and return do not overtake that work, then joins cleanup. Moving release before the callback made the test fail at line 143. `serve_test.go:156` separately measures the interval from callback completion and checks nonoverlap and release counts.
- [TG3] `refresh_test.go:33` checks exact prior bytes, validation/read error identities, one GET with the same context, and exactly one body close across success/rejection cases. Removing the trailing decode caused trailing-value, trailing-junk, and read-after-array cases to fail. Dropping the callback error during release also made both callback-error subtests fail at `serve_test.go:70`.
- [TG4] `cmd/indexer/main_test.go:61`, `:100`, `:144`, `:178`, and `:223` execute the compiled binary with finite deadlines, isolated working directories and explicit environments. Assertions cover actual exit status/stdout/stderr, accepted-prefix filesystem effects, two successful watch publications, cancellation of an in-flight third request, both process signals, and process joins. These process tests passed, including under the race run and five shuffled repetitions.

Bad

- [F2][moderate][introduced] The cancellation-error tests at `serve_test.go:76` and `:274` use comparable sentinel errors and standard pointer wrappers only. They do not exercise a legal custom error value containing a slice/map. A context cancellation cause of that form, returned wrapped by the callback, reaches an unsafe comparison in `serve.go:41` and panics. The unmodified candidate suite passes while the added isolated test `reproduction-cause-error/cause_error_test.go:12` fails with `runtime error: comparing uncomparable type indexer_test.reviewCauseError`. This is an independently actionable regression-test gap; primary owner is Testing. The production correction is [F1].

Suggested changes

- [F2] Add an exported-API cancellation test using a non-comparable error value as the context cause, returning that cause through a wrapper (and, if appropriate, a joined independent error). Assert that Serve returns without panic, release occurs once after callback completion, and any independently identifiable error remains observable. Run this test against the current implementation to demonstrate the failure, then against the [F1] correction. No new framework or package structure is needed.

Limits: The initial sandbox blocked loopback listeners; those failures were environmental and are not counted as defects. Re-running with loopback access passed the complete suite and race check. Executed Go is 1.26.5 on darwin/arm64; a real Go 1.22 toolchain run was not performed. `go vet -stdversion ./...` passed against the module's Go 1.22 declaration. Permission-based retained-byte tests executed without skips in the first verbose run. Windows replacement/signals are explicitly outside the README execution claim. The child command binaries are built without race instrumentation; race detection exercises the in-process library/tests and their shared state. Five shuffled runs provide finite evidence of isolation, not proof of every possible interleaving.

## Correctness & Compatibility — B
Scope: Same supplied changeset; existing exported function/method signatures and put process behavior preserved; new README contracts assessed on the evaluated POSIX host.
Coverage: Traced and exercised normal, empty, invalid, partial-completion, failure, cancellation, and recurrence paths. Inspected supported API function values, sentinel preservation, text/order retention, JSON representation, filesystem atomicity, client/body ownership, command argument validation, exit/stream behavior, signals, and idle-connection release ordering. Verified module language declaration and stdversion; actual Go 1.22 execution and Windows behavior were not tested.
Rationale: One moderate, contained error-handling defect affects a valid but narrower cancellation-cause representation. Serve panics instead of returning normally; the deferred release still executes once in the reproducer. There is no evidence of state corruption or systemic reach. Strong verified behavior in other paths does not offset this defect.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [CG1] `index.go:39` applies records sequentially and wraps parse/validation/write errors. Success, malformed input, invalid replacement, reader failure, and publication failure tests verify accepted earlier rows remain and later rows are absent. Put still accepts arbitrary text, including blank text, and exact function-value assignments compile at `index_test.go:15`.
- [CG2] `refresh.go:40` rejects null/nonarray input, additional JSON/junk and body read failures before publication, validates all records, and preserves ordered original text. `index.go:70` writes and closes a same-directory temporary file before rename. Actual tests verify empty array output, exact retained bytes on failure, temporary-file removal, and the old/new-reader snapshot distinction.
- [CG3] `serve.go:14` rejects nonpositive intervals before release; the synchronous callback at `:22` completes before the deferred release, while the timer at `:28` starts after successful callback completion. Tests verified callback/release joined identities, cancellation cleanup, interval timing, and independent invocation state. Actual watch processes recur, cancel the in-flight HTTP request, retain the last complete snapshot, and exit successfully with no output on SIGINT/SIGTERM.

Bad

- [F1][moderate][introduced] `serve.go:41` uses `err == cause` on arbitrary error interfaces. A valid error implementation can be a value containing a slice, so comparing two interfaces holding that same dynamic type panics. Reproduction: create `context.WithCancelCause`, cancel with `reviewCauseError{details: []string{"shutdown"}}` inside refresh, and return `fmt.Errorf("fetch stopped: %w", context.Cause(ctx))`. The wrapper is unwrapped at `serve.go:56`, then the comparison panics. The finite exported-API reproduction exits 1 and reports the runtime comparison panic; release count is 1. This violates the cancellation return contract and can terminate the caller's process if unrecovered. The effect is contained to this error representation, not all cancellation or refresh workflows. Primary owner is Correctness & Compatibility.

Suggested changes

- [F1] Make cancellation classification safe for arbitrary valid error implementations. Avoid unrestricted interface equality; account for comparable identity, supported error matching and single/joined unwrapping without treating a joined independent error as cancellation alone. At minimum, an unclassifiable non-comparable error must be returned rather than triggering a panic. Add [F2]'s regression and re-run the existing cancellation/error-join and lifecycle cases to confirm independent callback and release failures remain visible.

Limits: Exact verification commands and outputs are in `output/checks.json`. The full/race/shuffled suites pass, but they do not cover [F1] until the isolated reproduction is added. Go 1.22 source/API compatibility is supported by inspection, unchanged signatures and stdversion; runtime execution used Go 1.26.5. No power-loss durability, Windows replacement, or unsupported platform guarantee is inferred from atomic rename.

## Architecture & Design — A+
Scope: Integrated library/command production changes: shared atomic publication, streaming CSV application, borrowed HTTP fetching, callback-driven lifecycle, and command-owned signal/client setup.
Coverage: Assessed package/API boundaries, production composition, callback and HTTP dependencies, validation/publication boundaries, error ownership, context flow, process effects, and cleanup ownership. The supplied small module is complete enough to assess these decisions; no unrequested application layers or interface framework is assumed necessary.
Rationale: No actionable architectural issue was found. Two independent safeguards are verified beyond ordinary wiring: (1) a shared complete-file-before-rename publication boundary prevents in-place snapshot corruption, demonstrated by the old/new-reader test and a caught in-place-write mutation; (2) synchronous callback ownership makes release wait for cooperative cleanup, demonstrated by the held-cleanup test, caught early-release mutation and real signal-driven watch shutdown. These control filesystem visibility and resource lifetime respectively. [F1] is a local comparison defect requiring no boundary redesign and is ungraded here.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [AG1] `index.go:70` is a small concrete shared publication boundary used by Put and Refresh. It preserves the existing file-backed model while centralizing write/close/rename/temporary-file ownership. The candidate uses no new dependencies, exported abstraction hierarchy or unnecessary record mapping. Snapshot tests and the in-place mutation verify the boundary's actual consequence.
- [AG2] `serve.go:13` takes explicit refresh/release functions and per-invocation context; callbacks execute synchronously, errors remain caller-visible, and release runs after the callback stops. `cmd/indexer/main.go:60` owns process signal registration and its HTTP transport, and releases client idle connections through Serve. Library callers retain control of their own process and resources. Held-cleanup, independent-run and actual SIGINT/SIGTERM tests verify that ownership.
- [AG3] `refresh.go:28` copies the borrowed client before changing redirect policy, shares the supplied transport/context, and closes acquired bodies. The real redirect test at `refresh_test.go:120` verifies one request, rejection, retained bytes, and unchanged caller redirect policy. This preserves composability while enforcing the one-GET contract.

Bad

- None found.

Suggested changes

- None needed.

Limits: Architecture was judged against the supplied small library/command requirements. Broader deployment or durability contracts were not supplied. [F1] is a verified ungraded related correctness finding, and [F2] its independently actionable test gap; neither establishes an abstraction or ownership defect.

## Supporting verification facts

All Go checks use `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go ...`. Each command's exact argv, working directory, stdout, stderr, exit code and elapsed time are preserved in `output/checks.json` (13 entries). Checks run in disposable copies; original/candidate source was not changed.

- Toolchain: `go version` -> `go version go1.26.5 darwin/arm64`, exit 0; `go env GOOS GOARCH GOVERSION` confirms darwin/arm64/go1.26.5.
- Initial `go test -count=1 -timeout=60s -v ./...` and `go test -race -count=1 -timeout=60s ./...` -> exit 1 because the sandbox denied `httptest` loopback bind (`operation not permitted`). The verbose run's nonnetwork tests and permission checks executed successfully before the network fixture panic.
- With loopback access, `go test -count=1 -timeout=60s ./...` -> exit 0, both packages pass; `go test -race -count=1 -timeout=60s ./...` -> exit 0, both pass; `go test -count=5 -shuffle=on -timeout=60s ./...` -> exit 0, both pass.
- `go vet -stdversion ./...` -> exit 0, empty stdout/stderr.
- Mutation copies remove trailing-input validation, publish in place, release before callback cleanup, or discard callback errors during release. The corresponding focused tests each exit 1 for the intended behavioral assertion (not a compilation failure).
- Unmodified production code plus `TestReviewNonComparableCancellationCause` -> exit 1 with the reproduced interface-comparison panic and one deferred release. Test/process checks have finite Go and subprocess deadlines.

No network service beyond local HTTP fixtures, extra dependency, or installed toolchain was used. No unrelated legacy limitation is counted as an introduced finding.
