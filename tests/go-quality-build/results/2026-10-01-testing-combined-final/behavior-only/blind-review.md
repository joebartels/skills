# Independent anonymized changeset review

The supplied diff is `/private/tmp/go-independent-review-igwwhqr0/original` → `/private/tmp/go-independent-review-igwwhqr0/candidate`. No revision identifiers were supplied. The supplied README defines the contracts. Only this dispatch, its anonymized original/candidate trees, and its three local review skills/references were inspected. No author report, other candidate, evaluation expectations, or repository material was used. No delegation or candidate repairs occurred.

File references below are relative to `/private/tmp/go-independent-review-igwwhqr0/`; `candidate/` identifies reviewed source. Verification and mutation copies are disposable siblings of that directory. Grades assess introduced/worsened behavior and requested regression scope, not the original stubs or unrelated legacy debt.

## Testing — A+
Scope: Supplied original/candidate diff; standard-library file-backed library and command; module declares Go 1.22, execution used Go 1.26.5 on darwin/arm64.
Coverage: Inspected all implementation and test files, CSV assertions and error identity, exact snapshot representation and failure retention, real filesystem boundaries, response-body ownership, process grammar/status/streams, immediate and recurring lifecycle behavior, cancellation cleanup, error joins, independent invocations, fixture cleanup, and finite deadlines. Executed the entire suite, race checks, focused repeated/shuffled checks, and three finite mutations. No benchmarks or fuzz targets were introduced; no performance claim depends on either. Exact Go 1.22 runtime and other platforms were not executed.
Rationale: No substantiated actionable test issue. Two independent verified safeguards go beyond routine setup: the child-local file-size limit exercises a real write failure after progress and checks exact prior/accepted bytes, while the channel-controlled lifecycle test withholds cooperative cleanup and detects release overtaking it. Their respective direct-write and early-release mutations fail for the promised consequences, not just incidental errors. The whole suite and exercised race/repeat paths pass. This satisfies A+ without using test count or coverage percentage as a score.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/write_failure_unix_test.go:25`, `:39`, `:50`, `:73` isolates RLIMIT_FSIZE and SIGXFSZ changes in deadline-controlled helper processes. It checks Put's exact old bytes, CSV's accepted replacement/prefix and absent later row, Refresh's old destination, and temporary-file cleanup. The direct-publication mutation fails all three retained-byte checks; its output shows a 64-byte partially written destination. This detects the consequential failure that open/rename obstruction alone would miss.
- [T-G2] `candidate/serve_test.go:129` gates callback cleanup with `allowCleanup`, waits for observable startup/cancellation/cleanup events, checks that neither release nor return occurs while cleanup is withheld, then unblocks and joins. Cleanup handlers cancel and unblock on a failing assertion. The early-release mutation fails at `:149` with “release overtook callback cleanup” and exits finitely. This protects lifetime ordering independently of file publication.
- [T-G3] `candidate/index_test.go:80`, `:116`, `:126` exercises malformed field counts/quotes, invalid records, reader errors, and real rename obstruction. Assertions retain accepted values, inspect promised identities with errors.Is, and reject later-row effects. Success at `:67` checks duplicate replacement and exact quoted multiline text. Existing arbitrary-text/empty-text Put behavior remains tested at `:22`.
- [T-G4] `candidate/refresh_test.go:29`, `:74`, `:123`, `:143`, `:166` checks order, duplicates, whitespace/text preservation, `[]\n`, old open readers, one GET and supplied context/transport, acquired-body closure, status/read/decode/validation/publication rejection, retained destination bytes, and no redirected second fetch. Doubles exercise caller-owned transport and body boundaries; filesystem publication is real. The direct-publication mutation additionally fails both old-reader assertions at `:64`.
- [T-G5] `candidate/cmd/indexer/process_test.go:19` builds and launches the actual command. Its helper at `:95` checks exit code and both streams; process cases inspect positional Put text, ApplyCSV stdin/prefix effects, invalid Watch startup, and failed first refresh. The HTTP fixture at `:120` observes successful recurrence and an in-flight third request, sends both interrupt and termination in separate cases, waits for child exit and request cancellation, and checks exact retained snapshot bytes. Build/child/event waits have deadlines and cleanup joins the child. These process cases passed, including three repeated shuffled runs.
- [T-G6] `candidate/serve_test.go:44`, `:88`, `:162`, `:211` checks both callback/release causes, cancellation-only suppression versus independent failure, completion-based recurrence without overlap, and independent invocations. Broad errors.Is-based cancellation suppression is caught by the third mutation at `:113`. Twenty shuffled lifecycle repetitions and the full race run pass. Function-value assignments in `candidate/index_test.go:15`, `candidate/refresh_test.go:16`, and `candidate/serve_test.go:14` also protect consumer signatures.

Bad

- None found.

Suggested changes

- None needed.

Limits: Full execution initially failed only because the sandbox prohibited httptest loopback binding; rerun with loopback permission passed. The process child built by process_test.go uses ordinary `go build`, so the outer race run does not instrument that child binary; it does instrument the library tests and process-test harness. Actual Go 1.22 execution was unavailable; declared language version, symbols, and stdversion analysis were inspected. Unix write-failure helpers are explicitly darwin/linux-only; Windows replacement and signal execution were not claimed. Race/repeat success covers exercised paths only. Check commands and raw outputs are recorded below and in checks.json.

## Correctness & Compatibility — A+
Scope: Same complete supplied diff; preserved exported library signatures, existing Put/Get behavior and process Put contract, and new ApplyCSV/Refresh/Serve/Apply/Watch contracts from README.
Coverage: Traced success, empty/boundary inputs, validation/parse/fetch/read/publication errors, accepted-prefix state, exact persisted representation, response ownership, synchronous cancellation and release order, joined error identity, URL/flag startup validation, and actual child process status/streams. Inspected Go 1.22 declaration and symbol compatibility; execution covered the supplied POSIX filesystem and local HTTP transport. Nil callbacks/clients and noncooperative callbacks have no supplied defensive-handling contract and were not invented as requirements.
Rationale: No substantiated introduced or worsened correctness defect. Two independent verified safeguards support A+: complete temporary-file write/close before rename protects meaningful prior state and open-reader snapshots even after a real partial-write failure; cancellationOnly traverses joined causes so normal cancellation succeeds without hiding an independent failure. The latter's broad-matching mutation loses the independent error and the supplied test catches it. These control distinct data-integrity and error-reporting risks; passing ordinary calls alone is not the basis for the grade.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/index.go:32` writes a sibling temporary file, closes it before rename, and removes it on failure. Put uses this helper at `:28` while retaining its original arbitrary-text key contract; Refresh publishes only after complete read/decode/validation at `candidate/refresh.go:35`, `:51`, `:60`. Real partial-write failure tests and old-reader/new-opener checks pass. The direct-write mutation demonstrates the controlled failure's consequence across Put, CSV, and Refresh. Crash durability/fsync is not a stated contract.
- [C-G2] `candidate/serve.go:39` examines joined/wrapped causes individually. Ordinary parent cancellation alone is suppressed, independent callback causes survive, and `:17` joins release error identity with the result. Supplied tests pass for wrapped/joined cancellation, mixed cancellation/failure, non-comparable errors, and callback/release failures together. Replacing this guard with broad errors.Is matching makes the cancellation-plus-failure assertion fail with a nil result.
- [C-G3] `candidate/index.go:56` fixes CSV field count at two, validates before Put, and stops on each error while wrapping the cause. Accepted rows are published sequentially; rejected replacements cannot overwrite the prior accepted value. The real parse/validation/read/write cases establish accepted-prefix rather than all-or-nothing semantics, including no later row.
- [C-G4] `candidate/refresh.go:25` preserves transport/Jar/Timeout fields by copying the supplied client while disabling redirects to enforce one fetch. `:31` closes acquired bodies on rejection and success; `:44` rejects null, `:48` rejects a second JSON value or trailing junk, `:56` emits compact JSON and `:60` adds the newline. Empty arrays, exact text/order, status 206, body read error, invalid later records, redirects, and publication failure have passing meaningful assertions.
- [C-G5] `candidate/serve.go:14` validates interval before any callback/release, `:19` handles already canceled context, `:22` calls refresh synchronously, and `:28` begins the timer after callback completion. Release therefore cannot overtake callback cleanup. Immediate execution, no overlap, completion-based interval, cancellation, independent state, and exactly-once release are exercised; race and repeated lifecycle checks pass.
- [C-G6] `candidate/cmd/indexer/main.go:23` preserves Put's positional grammar. Apply routes stdin into the library; Watch validates startup at `:55`, owns signals at `:59`, composes Serve/Refresh at `:62`, and closes idle connections in release at `:65`. `:69` retains diagnostic stderr and exit 2 with silent success. Actual process checks pass for existing/new commands, recurring publication, and cancellation during an HTTP request. Exported signatures and Record JSON tags remain compatible; ErrNotImplemented remains an exported legacy symbol without affecting implemented operations.

Bad

- None found.

Suggested changes

- None needed.

Limits: Runtime checks used Go 1.26.5/darwin-arm64 with Go 1.22 module semantics; no exact Go 1.22 binary was installed or downloaded. stdversion vet passed and inspection found no newly required post-1.22 symbol, but this is not an executed minimum-version matrix. No Windows replacement/signal execution or external endpoint was evaluated. The explicit POSIX atomic-reader claim was tested on this filesystem. Legacy lack of arbitrary nil-argument handling, crash durability, bounded HTTP body size, or uncooperative-callback termination is not newly graded debt or an invented requirement.

## Architecture & Design — A
Scope: Same supplied library/CLI changeset. Production structure changed by introducing shared publication/validation helpers and implementing refresh/lifecycle/command composition; Architecture is applicable.
Coverage: Assessed package responsibilities, preserved public APIs, standard-library dependency composition, Record's transport representation, index-root independence, state/error boundaries, client/body ownership, lifecycle completion, and process-signal ownership. No omitted architectural portion was needed to assess this small complete module.
Rationale: No actionable design issue. The two-package design keeps the file/refresh/lifecycle library usable by callers while the command owns signals, standard streams, and process exit. Dependencies remain explicit, and focused verification confirms the stated ownership choices. This is proportionate ordinary correct composition; the grade does not add credit merely for extracting helpers or having many tests.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate/index.go:32` centralizes the existing atomic-publication protocol for Put and Refresh; `:76` shares only the new nonblank-record rule while Put at `:21` retains its arbitrary-text policy. It avoids duplicated persistence failure paths and an unnecessary all-or-nothing CSV transaction. Real partial-write checks passed across all consumers; direct-publication mutation fails those contracts.
- [A-G2] `candidate/serve.go:13` accepts narrow callback functions with caller-owned context, runs synchronously, and owns release after callback completion. There is no hidden worker or separate goroutine to drain. The cleanup gate and early-release mutation confirm the ownership boundary. Joined causes give callers useful failure identity without new public error wrappers or interfaces.
- [A-G3] `candidate/refresh.go:19` accepts a concrete supplied http.Client, keeps its transport/configuration, and owns acquired response bodies without closing the caller's client. Redirect policy is local to the required one-fetch operation. The command at `candidate/cmd/indexer/main.go:59` owns process signals and supplies the client release at `:64`; library invocation has no process-global signal registration or os.Exit.
- [A-G4] Record's existing JSON fields at `candidate/refresh.go:12` express the stated transport/snapshot representation directly. No demonstrated contract divergence requires duplicate DTO/domain models, package layers, a filesystem interface, or a framework. External-package tests and process tests exercise the actual supported boundaries, and existing consumer function types still compile.

Bad

- None found.

Suggested changes

- None needed.

Limits: Design conclusions concern this supplied small library/command and its stated evolution. No future extensibility, service deployment, crash durability, or other platform requirements were inferred. Go 1.22 and non-POSIX runtime limitations are as stated above; they do not hide an inspected boundary defect.

## Supporting verification facts

Every Go command used `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go`. No tools, dependencies, or toolchains were installed. Full argv, cwd, stdout, stderr, exit code, elapsed duration, permission context where applicable, and mutation names are preserved in `output/checks.json`.

- Environment: `go version` returned `go version go1.26.5 darwin/arm64`; GOOS=darwin, GOARCH=arm64, CGO_ENABLED=1. Both commands exit 0 with empty stderr.
- `go test -count=1 -timeout=45s ./...`: initial sandbox run exits 1 because httptest cannot bind tcp6 loopback (“operation not permitted”); root library package passes. Identical rerun with loopback permission exits 0, both packages pass, stderr empty.
- `go test -race -count=1 -timeout=60s ./...`: exits 0, both packages pass, stderr empty.
- `go test -count=20 -shuffle=on -timeout=30s -run ^TestServe .`: exits 0, stderr empty.
- `go test -count=3 -shuffle=on -timeout=45s ./cmd/indexer`: exits 0, stderr empty.
- `go vet -stdversion ./...`: exits 0, stdout/stderr empty. This is a static floor check, not execution on Go 1.22.
- Direct-publication copy: replaced only publish with os.WriteFile and ran `go test -count=1 -timeout=15s -run ^(TestRefreshRepresentationsAndSnapshots|TestWriteFailureAfterProgressRetainsState)$ .`. Exits 1 with meaningful assertion failures: old open readers see replacement bytes; Put/CSV/Refresh prior destinations contain partial new bytes after write failure. No compile error or timeout accounts for failure.
- Early-release copy: moved release before work while preserving its eventual joined error and ran `go test -count=1 -timeout=10s -run ^TestServeJoinsCancellationCleanupBeforeRelease$ .`. Exits 1 with “release overtook callback cleanup”; no timeout.
- Cancellation-classification copy: replaced cancellationOnly with broad errors.Is matching and ran `go test -count=1 -timeout=10s -run ^TestServeCancellationErrorClassification$ .`. Exits 1 at cancellation_plus_failure with “independent cause lost: <nil>”; no timeout.
- `rtk proxy diff -ru candidate verification` (absolute paths in checks.json): exits 0 with empty stdout/stderr. Verification copy remains byte-identical to the candidate. Original/candidate file SHA-256 manifests are saved in `output/source-manifest.json`; mutations exist only in separate disposable copies.

No counted bad findings or ungraded related production defects were substantiated. Proposed extra fuzzing, additional packages, interfaces, or platform support are not required corrections for the supplied contracts.
