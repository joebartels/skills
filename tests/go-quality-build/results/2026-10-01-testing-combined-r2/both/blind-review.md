# Independent changeset review

Reviewed the supplied `original` → `candidate` trees in `/private/tmp/go-independent-review-77k7qpuq`. No author report, other candidate, repository evaluation material, or outside workspace was inspected. The supplied README is the behavioral authority. This is a full supplied changeset for a small file-backed library and command; the original implementations of ApplyCSV, Refresh, and Serve were stubs. Findings below concern the implemented evolution and requested regression-test scope, not unrelated legacy debt. Paths below are relative to this review root.

## Testing — B
Scope: Supplied original/candidate diff; library and actual command process tests; module declares Go 1.22 and has no external dependencies.
Coverage: Assessed exact stored bytes, independent roots, exported function types, ordered CSV replacements, accepted-prefix effects on parse/validation/read/write failures, HTTP status/decode/validation/read/fetch/publication rejection, body ownership, old/new POSIX snapshot readers, callback timing and serialization, cancellation cleanup, joined error identity, process arguments/status/streams, recurring watch operation, and interrupt/termination. Exercised full tests and five shuffled race-enabled repetitions. No fuzzing or benchmark decision is introduced. Native Go 1.22 execution and Windows replacement/signal execution were not performed.
Rationale: One moderate introduced regression-detection gap selects B. The exact-200 acceptance rule lacks a success-shaped non-200 rejection assertion: a plausible all-2xx mutation passes the complete supplied suite. This leaves a contained response-status boundary insufficiently checked, rather than making the entire refresh or process contract unverified. The many verified strengths do not cancel this gap.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [TG1] `candidate/write_failure_unix_test.go:21` isolates a real partial-write failure using RLIMIT_FSIZE in a deadline-controlled child, then checks exact prior bytes or the accepted replacement, later-key absence, and temporary-file cleanup at lines 63–79. `candidate/refresh_test.go:229` holds an actual old file descriptor across replacement and asserts both old-reader and later-opener bytes. The disposable direct-write mutation failed these assertions for Put, ApplyCSV, Refresh, and the old-reader contract. This demonstrates detection of destructive publication regressions without changing limits in the test host.
- [TG2] `candidate/serve_test.go:90` gates cooperative cancellation cleanup and asserts that neither release nor Serve return overtakes it; it also checks independently returned callback errors, including context.Canceled. The early-release mutation failed all three gated cases at line 138. `candidate/serve_test.go:155` holds work longer than the interval and checks the next callback's elapsed time from completion, with overlap detection. The normal five shuffled race-enabled runs passed.
- [TG3] `candidate/cmd/indexer/main_test.go:23` builds the actual command; deadline-controlled child processes assert exit status, empty stdout, stderr diagnostics, and stored effects. Tests at lines 140 and 180 observe recurrence and real signal cancellation through local HTTP servers. Changing only `main`'s failure exit from 2 to 1 left helpers unchanged and caused the startup process cases to fail. The supplied suite therefore protects the process boundary rather than merely helper returns.
- [TG4] `candidate/index_test.go:68` and `candidate/refresh_test.go:77` assert state after rejection, not merely a non-nil error. Their errors.Is assertions preserve promised validation, read, and transport identities; success cases use literal expected bytes rather than generating expectations with the production encoder.

Bad

- [T1][moderate][introduced] Non-200 successful HTTP responses are absent from the new status regression cases. `candidate/refresh_test.go:86` exercises 503 and `candidate/refresh_test.go:168` exercises a 302 redirect; neither distinguishes status 200 from another 2xx response carrying a valid array. The README at `candidate/README.md:10` expressly accepts only 200. Replacing `candidate/refresh.go:32`'s exact-200 condition with a 200–299 range made the entire supplied suite pass. A reviewer-only 206 response containing a valid record array then showed the consequence: the mutated function returned nil and replaced the prior complete snapshot with the partial response, while the candidate correctly returned `fetch status: 206` and retained the exact prior bytes. Primary remediation owner: Testing. No corresponding production defect is present in the candidate.

Suggested changes

- [T1] Add a valid-array response with status 206 (and optionally another non-200 2xx status) to the rejection table, asserting an error, exact prior bytes, one request, and body closure. Verify that the current implementation passes and widening acceptance to all 2xx statuses fails. This tests the exact accepted-status boundary independently of malformed bodies or ordinary HTTP failures.

Limits: Execution used Go 1.26.5 darwin/arm64; the source and go.mod were inspected for Go 1.22 compatibility, and explicit stdversion vet passed. Windows replacement behavior is explicitly outside the README execution claim. The initial full run failed because the sandbox prohibited ephemeral localhost binds; the same full run and repeated race run passed with localhost access. No public internet access, tool installation, or dependency addition was needed. Exact commands, stdout, stderr, exit codes, timeouts, and relevant environment are in `output/checks.json`. The missing status assertion is demonstrated by finite mutation checks, not inferred from coverage percentage or test count.

## Correctness & Compatibility — A+
Scope: Supplied original/candidate diff; new library behavior and command process contracts; declared Go 1.22 language/API baseline and evaluated POSIX filesystem.
Coverage: Traced ordinary, empty, invalid, duplicate, partial-failure, cancellation, error-identity, snapshot-ownership, process startup, process termination, and function-type compatibility paths. Compared existing Put/Get behavior with the original implementation. No new on-disk migration or public signature change is introduced. Source/version checks and runtime verification limits are stated below.
Rationale: No substantiated introduced or worsened behavioral defect was found. Two independent verified safeguards support A+: same-directory temporary publication protects meaningful stored state and reader snapshots across write/publication failures; synchronous callback ownership protects cleanup-before-release and callback/release error identities during cancellation. Both were demonstrated by passing candidate checks and failing targeted mutations, and control distinct state and lifecycle risks.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [CG1] `candidate/index.go:68` writes and closes a same-directory temporary file before rename and removes the temporary file on failure. ApplyCSV validates before calling Put and stops on the first parse, validation, or write error at lines 47–54, preserving its accepted prefix. Refresh validates the entire array before publication at `candidate/refresh.go:39`–55. Actual partial-write tests, rename-obstruction tests, exact retained-byte assertions, and the POSIX old-reader assertion passed; substituting direct overwrite failed the relevant assertions.
- [CG2] `candidate/serve.go:15` rejects invalid intervals before installing release; lines 18–31 call refresh synchronously, wait from completion, and join callback/release errors. The already-canceled, invalid-interval, both-errors, independent-invocation, completion-delay, and gated cancellation tests passed, including under the race detector. The early-release mutation was rejected, demonstrating preservation of cooperative cleanup before resource release.
- [CG3] `candidate/refresh.go:19` passes the caller context into the request; lines 25–32 disable redirect following without altering the caller's client and reject every status other than 200. Acquired bodies are closed via the defer at line 31 on success and rejection. Real HTTP cancellation and redirect checks passed. The reviewer-only 206 probe additionally confirmed exact-status rejection and retention of prior bytes, even though the supplied regression suite omits that case (Testing finding T1).
- [CG4] Public function signatures remain source compatible, including function-value assignments in `candidate/index_test.go:14`, `candidate/refresh_test.go:20`, and `candidate/serve_test.go:14`. Put retains arbitrary text, including empty, whitespace, NUL, and positional dash text. `candidate/cmd/indexer/main.go:70`–75 retains failure status 2/stderr behavior and connects interrupt/SIGTERM cancellation to the synchronous Serve lifecycle. Actual put/apply/watch process assertions passed.

Bad

- None found.

Suggested changes

- None needed for assessed production behavior. T1 is a separately actionable regression-test improvement and is not counted as a correctness defect.

Limits: Native Go 1.22 compilation/runtime was not exercised; go.mod remains at 1.22, the changes use APIs available by 1.22, and `rtk proxy go vet -stdversion ./...` passed on Go 1.26.5. Inspected timer use does not require newer timer-draining semantics: canceled timers are stopped and discarded, not reset/reused. Windows replacement semantics are outside the supplied claim. Power-loss durability/fsync and cancellation guarantees for an uncooperative callback or custom transport are not promised by the README. These are narrow unsupported boundaries, not assessed successes. The complete checks and mutation outcomes are preserved in `output/checks.json`.

## Architecture & Design — A+
Scope: Supplied original/candidate production changes to library publication, fetch/validation, refresh lifecycle, and command composition. This topic applies because the change adds production behavior and ownership boundaries, not only tests.
Coverage: Assessed exported signatures, package fit, helper responsibilities, dependency visibility, client/body/file ownership, callback lifetime, process signal ownership, cancellation/error flow, and proportionate composition. No additional package hierarchy, interface framework, options layer, or background worker is introduced.
Rationale: No actionable introduced architecture issue was found. Two independent, verified ownership safeguards support A+: synchronous Serve makes started callback work belong to its invocation until cleanup completes; a copied redirect policy enforces the one-GET boundary without mutating caller-owned client configuration, while Refresh closes acquired bodies and the command closes its private transport after Serve finishes. These eliminate demonstrated premature-release and resource-ownership risks while keeping the solution small.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [AG1] `candidate/serve.go:14` keeps callback, context, timer, and release state local to each invocation and uses the synchronous callback at line 23 as its join mechanism. It exposes meaningful errors through errors.Join rather than hiding work in an unowned goroutine. Gated cleanup, independent-invocation, and combined-error tests passed; the early-return mutation failed. Consumers can rely on return and release occurring after callback completion without an additional lifecycle API.
- [AG2] `candidate/refresh.go:17`–31 preserves caller ownership of the client and owns each acquired body explicitly. The copied redirect policy retains the supplied transport while avoiding shared client mutation. Body-closure counters cover success and rejected responses; the real redirect test verifies one request and unchanged CheckRedirect. `candidate/cmd/indexer/main.go:54`–63 owns a private cloned transport and supplies its idle-connection release to Serve, whose joining behavior was verified. Library code owns no signal handlers or process exits; the command owns those at lines 70–75.
- [AG3] The private `publish` helper at `candidate/index.go:68` centralizes the same atomic replacement rule already needed by Put and newly required by Refresh. Validation stays where the input contracts differ: Put permits arbitrary text, while CSV and fetched Records use `validateRecord`. This avoids changing Put's public semantics or adding duplicate abstractions. Known-byte, prefix-retention, and partial-write checks verify the resulting boundaries.

Bad

- None found.

Suggested changes

- None needed.

Limits: The runtime evidence is limited to Go 1.26.5 darwin/arm64 and the supplied complete example, with no broader consumer repository or deployment contract provided. Signal handling and client cleanup placement were traced in the command and cancellation was exercised by actual processes; no custom RoundTripper with noncooperative shutdown was claimed to be supported. T1 concerns missing status regression coverage, not the architecture of the correctly implemented rejection boundary.

## Supporting verification facts

All Go checks used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Commands below are the exact tool invocations recorded by the runner; each ran in its named disposable tree under this review root.

| Check | Command | Result |
| --- | --- | --- |
| Toolchain | `rtk proxy go version` | exit 0; `go version go1.26.5 darwin/arm64` |
| Initial full suite in restricted sandbox | `rtk proxy go test -count=1 -timeout=60s ./...` | exit 1; httptest ephemeral localhost bind was denied |
| Full suite with localhost access | `rtk proxy go test -count=1 -timeout=60s ./...` | exit 0; library and command packages passed |
| Repeated race/isolation check | `rtk proxy go test -race -shuffle=on -count=5 -timeout=60s ./...` | exit 0; both packages passed all five repetitions |
| Go API version check | `rtk proxy go vet -stdversion ./...` | exit 0; empty stdout/stderr |
| Direct-write mutation | `rtk proxy go test -run '^Test(PublicationWriteFailure\|RefreshPOSIXOpenReaderKeepsOldSnapshot)$' -count=1 -timeout=10s .` | exit 1; prior-byte and old-reader assertions rejected mutation |
| Early-release mutation | `rtk proxy go test -run '^TestServeWaitsForCancellationCleanup$' -count=1 -timeout=10s .` | exit 1; all cleanup cases rejected premature return |
| Process-exit mutation | `rtk proxy go test -run '^TestStartupRejectionsAreProcessFailures$' -count=1 -timeout=35s ./cmd/indexer` | exit 1; process cases observed 1 instead of promised 2 |
| All-2xx mutation, supplied suite only | `rtk proxy go test -count=1 -timeout=60s ./...` | exit 0; demonstrated T1 gap |
| Candidate + reviewer 206 probe | `rtk proxy go test -run '^TestReviewerNon200SuccessStatusPreservesPriorBytes$' -count=1 -v -timeout=10s .` | exit 0; error `fetch status: 206`, exact prior bytes retained |
| All-2xx mutation + same reviewer probe | `rtk proxy go test -run '^TestReviewerNon200SuccessStatusPreservesPriorBytes$' -count=1 -v -timeout=10s .` | exit 1; nil error and prior snapshot replaced by the partial result |
| Candidate preservation | `rtk proxy diff -rq candidate verification` | exit 0; empty stdout/stderr |

`output/checks.json` stores exact stdout/stderr/exit information; `output/mutations.json` describes the isolated mutations; `output/review_status_probe_test.go` preserves the reviewer probe; `output/source-manifest.json` records supplied-tree SHA-256 hashes. Original and candidate source/configuration were not edited. No code repairs were applied.
