# Independent changeset review

Reviewed the supplied diff from `/private/tmp/go-independent-review-0rd44ar6/original` to `/private/tmp/go-independent-review-0rd44ar6/candidate`, without an author report or other candidate/evaluation material. The actual request is to extend the one-record CLI to a streamed ordered batch and add high-quality success/failure tests. README.md and go.mod are unchanged. Production changes are confined to candidate/main.go; tests are replaced and extended in candidate/main_test.go.

The governing contract is candidate/README.md: each accepted record writes its normalized quantity, duplicate IDs replace, the first invalid row/CSV/file failure stops processing, and writes in the accepted prefix remain. Go 1.22 and standard-library-only dependencies remain required. This review interprets a duplicate whose write fails as outside the accepted prefix: the earlier successful value must therefore survive. It does not require batch rollback or durability after power loss.

## Testing — B

Scope: Supplied original/candidate diff; standard-library Go CLI declaring Go 1.22. Reviewed the requested new batch verification and compared the original single-record test. Executed using Go 1.26.5 on darwin/arm64.

Coverage: Assessed success normalization, duplicate ordering, empty input, quoted CSV/CRLF, final rows without newline, invalid IDs/quantities/field counts, malformed CSV, accepted-prefix preservation, read errors, input streaming, usage-before-input behavior, real filesystem obstructions, CLI build/execution/output/exit behavior, test isolation, and finite subprocess lifetimes. No concurrent application path is introduced; race testing is not material to this synchronous implementation. No fuzz or benchmark claim was made. Active-file partial-write failure is the substantiated missing behavior described below.

Rationale: One moderate gap leaves a plausible, contained I/O failure behavior unchecked. Prefix preservation is well exercised for validation/read failures and filesystem errors that happen before an existing record is opened, so the important prefix contract is not wholly unverified. However, the suite passes while a failed duplicate write damages an accepted record. This gap needs a test correction independent of the production correction in C1. Two meaningful assertion safeguards were verified with disposable mutations; they do not cancel the gap.

Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [T-G1] candidate/main_test.go:160 and :185 use a reader that checks the real filesystem before releasing each next row. A disposable mutation that first calls io.ReadAll failed TestRunWritesBeforeReadingNextRow with `records = map[], want map[web:7\n]` (exit 1). This demonstrates a meaningful safeguard against buffering/prevalidating the entire input before performing the accepted writes.
- [T-G2] candidate/main_test.go:104 and :272 exercise actual filesystem and executable boundaries. The former checks the complete retained prefix and an untouched obstructing directory; the latter builds the real binary and checks exit 0/2, stdout silence, useful stderr, and complete persisted contents. A mutation that continues after os.WriteFile failure failed TestRunFileFailureRetainsPrefix with a nil-error assertion (exit 1), establishing detection of ignored file failures.
- [T-G3] candidate/main_test.go:134 rejects any stdin read following usage errors. candidate/main_test.go:206 checks `errors.Is` for a supplied input failure and verifies the retained prefix. `assertRecords` at :238 compares both the complete filename set and file contents, detecting extra later outputs as well as wrong quantities.
- [T-G4] Filesystem fixtures and built binaries live in t.TempDir; fake readers call testing failures from the synchronous test goroutine. The binary build has a 30-second context and each subprocess has a 5-second context, with cmd.Run waiting for completion. No sleeps, shared mutable fixtures, external services, or module mutations are required. The suite passed once and then three shuffled repetitions.

Bad

- [T1][moderate][introduced test-scope gap] candidate/main_test.go:104 and :294 model file failure by making the target a directory. That error occurs before an existing regular file is truncated/written. Neither these cases nor the directory obstruction at :120 exercise a failed replacement after an earlier occurrence of the same ID succeeded. The suite therefore passes despite C1: under a finite child file-size limit, `web,1\nweb,12345\nlater,9\n` exits 2 but leaves web containing `123`, losing the accepted prefix value `1\n`. This is a plausible disk/full-or-write-limit failure mode and independently actionable regression-detection gap, not a demand for more test count or coverage percentage.

Suggested changes

- [T1] Add a bounded test of a failed write to an already accepted duplicate ID, checking the prior exact bytes, no later file, error/exit 2, and no success stdout. A supported-platform subprocess file-size limit can reproduce partial writes without unreliable permission tests; alternatively use a narrowly controlled production write boundary only if that boundary is otherwise justified. Keep a real filesystem success/failure check. Verify the new test fails against this candidate and passes after the C1 correction.

Limits: Candidate `go test -count=1 -timeout=45s ./...` passed (0.471s); `go test -count=3 -shuffle=on -timeout=45s ./...` passed (0.828s). Both mutation checks failed as intended. All commands used `rtk proxy`; all Go commands used GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. Exact commands/stdout/stderr/exit codes are in checks.json. The installed compiler is Go 1.26.5, so execution on the minimum Go 1.22 toolchain was not verified. Inspected new APIs are available in Go 1.22; no network, external dependency, or external service was used. Partial-write verification is local darwin/arm64 evidence, not a claim about every supported platform.

## Correctness & Compatibility — C

Scope: Supplied original/candidate diff, focusing on the new streamed ordered batch and preserved CLI process contracts. Module declaration remains Go 1.22; verification used Go 1.26.5 on darwin/arm64.

Coverage: Traced ordered success, duplicate replacement, EOF/empty input, invalid records, malformed CSV, input failures, directory creation, file failures, retained prefix, usage and CLI diagnostics/exit behavior. Compared old and new validation/parser behavior, signatures and dependencies. Verified ordinary behavior through the candidate suite and reproduced an active duplicate write failure through the real built binary. No concurrency or supported external Go API changed. Go 1.22 execution and other OS/architecture execution were not available in this check.

Rationale: One contained major issue violates an explicit batch data-integrity contract by destroying the saved value of an already accepted record when a later duplicate cannot be fully written. The reproduction proves lost accepted state in one importer workflow; there is no evidence of broad/systemic or critical reach. The primitive non-atomic write existed in the original one-record importer, but the new batch workflow newly exposes it to an in-batch accepted-prefix promise. Unrelated legacy overwrite/security/durability limitations are not counted.

Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [C-G1] candidate/main.go:25 rejects unexpected field counts before accessing row fields; :28–41 stop on CSV/validation errors; :49 completes each accepted write before the loop reads the next row. Passing behavior tests and the buffer-first mutation demonstrate streamed ordering rather than whole-batch prevalidation.
- [C-G2] candidate/main.go:29 treats EOF/empty input as success, and :43 defers directory creation until a validated record exists. Tests at candidate/main_test.go:76, :86 and :94 verify both lazy side effects and successful nested-directory creation.
- [C-G3] candidate/main.go:33, :45 and :50 wrap underlying I/O/CSV errors with record/operation context; :55 retains the existing stderr and exit-2 boundary. The input-error identity test and real CLI cases verify failure propagation, no success stdout, and useful diagnostics. The controlled partial-write reproduction also correctly exited 2 and stopped before `later`.

Bad

- [C1][major][worsened/newly exposed by batching] candidate/main.go:49 uses os.WriteFile directly on the final path for every duplicate. Opening a regular file truncates its earlier contents before the replacement write can fail. With child RLIMIT_FSIZE=3 and SIGXFSZ ignored, input `web,1\nweb,12345\nlater,9\n` first accepts web as `1\n`, then returns `record 2: writing "web": ... file too large` (exit 2, empty stdout). The final directory is exactly `{"web":"123"}`. The earlier accepted value is destroyed and the replacement is not a normalized quantity plus newline. This contradicts README.md's accepted-prefix retention promise. The direct-write primitive is legacy, but this accepted in-batch state and its retention requirement arise from the new ordered batch.

Suggested changes

- [C1] Ensure a replacement is fully written and closed before committing it to the final name, for example through a temporary file in the same directory followed by rename, with cleanup on write/close/rename failure and preserved intended file behavior. On failure retain the prior accepted file and stop processing. Verify using the recorded finite duplicate partial-write reproduction plus the existing success/ordering/CLI checks; do not introduce rollback of earlier accepted rows. Testing owns the separately actionable missing regression assertion T1.

Limits: The built candidate and all existing tests pass. The additional controlled write-failure reproduction fails the prefix requirement while retaining correct exit/diagnostic/stop behavior. Its exact command, environment setup, input, output, exit, and filesystem result are recorded in checks.json; verify.py is the reproducible harness. No original or candidate source was altered. Minimum-version and non-Darwin runtime behavior were not executed; no installed toolchain download or network operation was attempted. The review does not infer all-or-nothing batches, crash durability, symlink hardening, or compatibility with arbitrary preexisting-file modes from this README.

## Architecture & Design — Not applicable

Scope: Production changes in candidate/main.go relative to original/main.go.

Coverage: Inspected package layout, run/main boundaries, dependency injection, ordering loop, directory initialization and ownership.

Rationale: There is still one main package, the same run(args, input, stderr) seam, the same synchronous reader/filesystem dependencies and main process lifetime, and no new interface, package, background worker, or acquired long-lived resource. The implementation extends the existing run operation with an ordered loop and lazy directory initialization; no consequential production seam/package/lifetime decision changes warrant a separate architecture grade. C1 concerns the correctness of the newly promised persistence behavior. The proposed repair has not been implemented or reviewed here.

Limits: Architecture reasoning is based on the complete supplied four-file original/candidate projects. No architectural redesign or code repair was performed.

## Supporting verification facts

- `checks.json` preserves exact rtk-prefixed command arrays, working directories, input, separate stdout/stderr, exit status, finite timeouts and reproduction setup. `verify.py` created disposable mutation copies only under this review directory.
- Installed toolchain: `go version go1.26.5 darwin/arm64`. Platform: Darwin 25.5.0 arm64. Candidate build succeeded without stdout/stderr (exit 0).
- Existing suite: exit 0. Three shuffled suite runs: exit 0. Buffer-before-write mutation: exit 1 on the meaningful filesystem assertion. Continue-after-write-error mutation: exit 1 on the expected-error assertion.
- The partial-write reproduction uses a child-only three-byte file-size limit and ignores SIGXFSZ so the program can report the I/O failure. It does not change the reviewer's limits or require disk exhaustion. The binary exits 2, stdout is empty, stderr names record 2/web/file-too-large, web is `123`, and later is absent.
- Candidate source SHA-256 after verification: README.md `9b9cb46434771af31b656f3e2f5c608453676942d4e221684fa3db7e6ad44ea7`; go.mod `7cc2e99e3d21cbb24055f9067300830bf4072ba1b15cb5b00e4932cfcc2ebc0f`; main.go `91004aa2c2406c454f8202ff9d665f37344105a77b9c5a9c323f64a1b3d96a4f`; main_test.go `58a8b8da6c272348f7cc756aaceaac6e1234ebcac80ee97feca8afe1f21d4bae`.
