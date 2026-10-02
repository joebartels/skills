# Independent Go review

Reviewed the complete supplied directory diff from `/private/tmp/go-independent-review-4kte569z/original` to `/private/tmp/go-independent-review-4kte569z/candidate`. This is a snapshot library, despite the HTTP dependency. The production change implements the previously reserved `Refresh` and adds private `replaceSnapshot`; the existing `Write`, exported `Item`, `Refresh` signature, README, and `go.mod` are unchanged. The requested new test scope is assessed as part of Testing. No author report, evaluation expectations, other candidate, or unrelated repository input was inspected. Original and candidate source were not edited; mutation and additional verification files exist only in disposable copies below the review root.

Source references below are relative to `/private/tmp/go-independent-review-4kte569z/`. The contract is `candidate/README.md:3-16`: use the supplied HTTP client and context for GET; accept exactly a complete 200 JSON array; reject invalid items and preserve nonblank values; leave exact prior bytes intact on the specified failures; replace the inode so supported Unix open readers retain the old snapshot; close bodies; preserve API, Go 1.22, and standard-library dependencies. It does not promise file mode preservation, crash durability, Windows equivalence, or a storage abstraction.

## Testing — A+

Scope: Supplied original/candidate diff; regression tests for the newly implemented library `Refresh`, including requested remote and publication behavior. Module declares Go 1.22; checks ran with Go 1.26.5 on darwin/arm64.

Coverage: Assessed success output and ordering, empty arrays, request method/URI and supplied-client use, caller cancellation, non-200 and transport errors, malformed/null/non-array input, invalid field types and blank/missing values, complete-document checking and read failures, exact previous bytes, successful/failed response-body lifetime, public function type, actual filesystem replacement, temporary-file cleanup, missing parent, and isolated fixtures. Assessed the handwritten RoundTripper, request recorder, tracked reader, context bounds, and `t.TempDir` lifecycle. No fuzzing/benchmarks are supplied or required for this scoped change. Real socket assertions could not run because loopback binding is prohibited; no external network was used. There is no new background work or shared production state; a concurrent-writer stress test was not run and is not part of the stated contract.

Rationale: No actionable issue found, counts all zero. Two independent verified safeguards meet A+: (1) byte-exact failure preservation detects destructive effects even when an error is returned, controlling rejected refreshes corrupting the stored snapshot; (2) an old open handle plus inode identity detects in-place publication, controlling readers losing their coherent old snapshot. These are observable behavior assertions beyond routine fixture setup, and each rejected its targeted compiling mutation. Additional mutations independently demonstrated signal for trailing input, body ownership, value normalization, and request context.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/refresh_test.go:68-125`, `143-156`, `158-172`, and `174-192` assert failure and exact prior bytes using a deliberately noncanonical prior snapshot (`298-315`). In the `overwrite_on_failure` disposable mutation, the function still returned errors but the tests failed at line 119 on `"corrupt"` bytes. This validates failure preservation independently of return-value checking.
- [T-G2] `candidate/refresh_test.go:194-229` exercises real file handles and `os.SameFile`, scopes the open-handle claim to Unix, and closes the old handle through cleanup. The `in_place_publication` mutation failed at line 220 because the old handle saw `[]`. This directly catches the core replacement regression instead of mocking the publication operation.
- [T-G3] `candidate/refresh_test.go:23-65` checks exact serialized values, lowercase keys, ordering, duplicates, empty output, and method/URI through the supplied `http.Client`. `normalize_values` failed at line 55. The transport double does not replace the implementation's request creation, JSON decoding, validation, or actual file writes.
- [T-G4] `candidate/refresh_test.go:92-96`, `112-123`, and `128-140` check trailing JSON, trailing malformed data, underlying reader failure, and explicit response-body close counts. Removing the second decode caused the selected trailing/read-failure cases to succeed unexpectedly; removing the body close caused failure-path and success close-count assertions to fail. These doubles supply observable dependency behavior rather than mirroring calls to private helpers.
- [T-G5] `candidate/refresh_test.go:174-192` makes cancellation observable through the received request context and has a one-second fallback. The `ignore_context` mutation failed at line 189 after the fallback rather than hanging. External Go command timeouts also bound verification. Temp directories are independent and tracked bodies are local to individual tests; shuffled repeated race runs passed.
- [T-G6] `candidate/refresh_test.go:232-271` exercises actual creation/rename failures with a missing parent and a nonempty destination directory; it checks cleanup and body closure. The external function-type assertion at line 21 verifies the exact documented signature for a consumer package.

Bad

- None found.

Suggested changes

- None needed.

Limits: All original and candidate supplied tests passed; candidate also passed `-race -shuffle=on -count=5`. Mutation checks used selected tests and are evidence of those assertions, not exhaustive mutation coverage. The added real-HTTP verification copy failed before its assertions with `listen tcp6 [::1]:0: bind: operation not permitted`; this is an environment restriction, not a candidate defect. The candidate's standard `http.Client` call is exercised through controlled transports and its filesystem publication through actual OS files. Native Go 1.22, other Unix filesystems, Windows behavior, disk-full/write-close faults, and crash durability were not executed. The latter two are not promised contracts; pre-rename write/close failure paths were inspected. Exact commands, stdout, stderr, exit codes, and working directories are in `output/checks.json`.

## Correctness & Compatibility — A+

Scope: Introduced `Refresh` implementation and private publication helper in the complete supplied diff; library API and file/HTTP contracts from README. `Item`, `Write`, `Refresh` function type, `go 1.22`, and the standard-library-only module remain unchanged.

Coverage: Traced request creation, supplied-client execution, transport failure, non-200 return, response ownership, complete JSON parsing, array-versus-null distinction, each-item validation, unchanged nonblank values, empty-array encoding, temporary file creation/write/close/rename/removal, and Unix open-handle publication. Inspected supported language/library calls and public signatures. No Windows equivalence claim is made. The unchanged `Write` truncation behavior is legacy local-only context; this change does not route `Refresh` through it.

Rationale: No actionable issue found, counts all zero. Two independently verified safeguards control distinct important risks: the complete-response validation gate prevents rejected remote data from changing prior file bytes; sibling-temp-file publication and rename preserve open-reader snapshots and present a complete new file. Failure-preservation cases and the open-handle/inode test verify the respective effects, and mutations demonstrate that neither assertion is vacuous. The implementation uses these safeguards directly without a repair to the candidate. This supports A+ for the inspected contract, without making untested platform or durability claims.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/mirror.go:35-45` builds a GET with the supplied context, calls the supplied client, returns wrapped request/transport failures, defers body closure, and rejects all non-200 status. The request, cancellation, transport, non-200, and close-count tests passed. The callback cancellation test verifies actual request-context propagation, not merely presence of a context argument in the public signature.
- [C-G2] `candidate/mirror.go:48-67` rejects null/non-array, malformed or incomplete documents, additional JSON, underlying post-array reader errors, and invalid items before touching the destination. `strings.TrimSpace` is used only to validate; original values are marshaled. The detailed failure matrix and preserved-value success case passed. Empty input arrays remain non-nil and publish `[]`.
- [C-G3] `candidate/mirror.go:75-92` creates a sibling temporary file, writes and closes it before renaming, and removes it on error. It never truncates the old path. The Unix old-handle/new-inode test and actual rename-failure cleanup test passed on the current Darwin filesystem. Write and close failures return before the rename and cannot modify the old destination through this code path.
- [C-G4] Exported `Item` JSON tags remain `code` and `label`; `Refresh` retains `func(context.Context, *http.Client, string, string) error`. The external function-type check compiled; `go.mod` is unchanged at Go 1.22 with no dependencies. `go vet -stdversion ./...` passed under the available toolchain. Nil client and invalid request are checked before execution rather than introducing panics from those inputs.

Bad

- None found.

Suggested changes

- None needed.

Limits: Executed Go 1.26.5 darwin/arm64, not a native Go 1.22 toolchain. Module language semantics, API inspection, and the passing standard-library version vet check support the minimum but do not constitute an executed Go 1.22 run. `-race` only covers exercised paths. No additional supported filesystem/platform behavior is inferred from the Darwin check. Additional socket-level verification was blocked by the environment as recorded above. No permission-mode, crash-durability, process-cancellation of disk operations, or writer-ordering promise is inferred. All check details are preserved in `output/checks.json`.

## Architecture & Design — A

Scope: New synchronous orchestration and private publication helper. Architecture applies because implementation of the reserved function introduces consequential HTTP response ownership and temporary-file lifetime decisions, even though no exported seam or package is added.

Coverage: Assessed package/API boundary, explicit client/context/path composition, transport/persistence representation, error propagation, response/file ownership, and absence of background work, mutable global hooks, or unnecessary abstraction. Read the supplied architecture skill and its design-decisions reference in addition to testing/correctness guidance.

Rationale: No actionable issue found, counts all zero. The existing API fits the small library, while dependencies and per-call lifetimes remain explicit. Routine coherent ownership and a private helper support A; no separate architectural safeguards beyond those already evaluated as behavior protections are claimed for A+.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate/mirror.go:31-43` keeps the caller-configured `*http.Client` and context visible and scoped to this operation; it owns only the acquired response body. Supplied-client and cancellation assertions passed. There is no global client, hidden network construction, or independently running task to shut down.
- [A-G2] `candidate/mirror.go:72-92` keeps same-filesystem temporary-file ownership in a small private publication helper. Success is reported only after close and rename, and failure cleanup is locally owned. The helper changes no public method sets, injects no process-wide hooks, and creates no abstraction requiring downstream migration. The actual filesystem and body lifecycle tests passed.
- [A-G3] Existing `Item` serves the matching remote and local JSON contract without redundant mapping types. The design does not need a bespoke storage interface to verify its observable contract: the supplied tests use consumer-facing `Refresh` with controlled HTTP behavior and real temporary files.

Bad

- None found.

Suggested changes

- None needed.

Limits: Assessment is limited to this complete small supplied module and its stated public contract. It makes no claim about unseen application deployment needs. Socket-level execution was environmentally blocked; normal dependency composition and request/response behavior remain inspected and tested through the supplied client. No source repair or production abstraction was introduced by this review.

## Supporting verification facts

Every Go command ran through `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local`; there were no toolchain downloads or external network calls. Exact structured execution records are in `output/checks.json`.

| Check | Verified result |
| --- | --- |
| Toolchain/platform | `go version go1.26.5 darwin/arm64`; GOOS=darwin, GOARCH=arm64, CGO_ENABLED=1 |
| Original `go test -timeout=20s ./...` | exit 0; original local-write test passed |
| Candidate `go test -timeout=20s ./...` | exit 0; supplied suite passed |
| Candidate `go test -race -shuffle=on -count=5 -timeout=30s ./...` | exit 0; repeated shuffled suite passed |
| Candidate `go vet -stdversion ./...` | exit 0; no diagnostics |
| Disposable `in_place_publication` mutation | exit 1; old-handle byte assertion at test line 220 failed |
| Disposable `missing_body_close` mutation | exit 1; close-count assertions at lines 121 and 138 failed |
| Disposable `missing_trailing_check` mutation | exit 1; selected trailing/read-error cases failed at line 114 because Refresh succeeded |
| Disposable `normalize_values` mutation | exit 1; exact output assertion at line 55 failed |
| Disposable `ignore_context` mutation | exit 1; cancellation identity assertion at line 189 failed after bounded fallback |
| Disposable `overwrite_on_failure` mutation | exit 1; exact-prior-byte assertions at line 119 failed despite errors being returned |
| Reviewer-added real HTTP tests in a separate copy | exit 1 before assertions; sandbox prohibited loopback bind, `operation not permitted` |

All mutation failures were test assertion failures rather than build errors. Each disposable mutation source is preserved under `mutations/<name>/mirror.go`. Additional HTTP verification source is preserved under `real-http-verification/review_boundary_test.go`; its assertions were not exercised. Source-file SHA-256 values for the supplied original/candidate snapshots are recorded in `output/source-sha256.json` for artifact identity.
