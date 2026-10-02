# Independent review: changeset 1

Review boundary: supplied `original/` versus `candidate/` directories in `/private/tmp/go-independent-review-b_q8mpxi/changeset-1`. The actual request is to extend ledgerload to the README's streamed ordered batch contract and add high-quality success/failure tests. No author report, other candidate, or evaluation guidance was used. File references below are relative to this changeset directory. Exact verification commands, working directories, non-secret environment overrides, stdout, stderr, exit codes, mutations, and source hashes are in `checks.json`.

## Testing — A+

Scope: The supplied CLI changeset: changed `candidate/main.go`, added `candidate/batch_test.go` and `candidate/write_failure_unix_test.go`, and retained single-record test. The module declares Go 1.22; execution used Go 1.26.5 on darwin/arm64.

Coverage: Success, normalization, duplicate replacement, empty input, invalid first and later rows, malformed CSV and field counts, input failure, directory/file failure, retained prefix, absence of later writes, actual process exit/stdout/stderr, streaming order, and partial-write replacement safety. Assessed assertion signal, real filesystem/process boundaries, child-only resource limits, and bounded subprocess cleanup. No changed concurrent production behavior calls for a race run in this CLI.

Rationale: No actionable testing issue found. Two independent safeguards beyond ordinary setup are verified: an input-boundary assertion detects eager whole-input reading, and a real child-process write failure detects corruption of an already accepted duplicate. Both fail at their intended assertions under targeted production mutations, while the unmodified suite passes. These control distinct risks: deferred batch acceptance and destructive failed replacement.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/batch_test.go:90` supplies successive input chunks and checks the first file when the second chunk is requested. Replacing input handling with `io.ReadAll` made this test fail at line 99: no first file existed. This meaningfully verifies streaming rather than merely final batch contents.
- [T-G2] `candidate/write_failure_unix_test.go:28` executes the built CLI under a child-only two-byte file-size limit, then checks exit 2 and exact retained contents. Restoring direct destination writes made line 36 fail with `web = "12"`, expected `"7\n"`. The unmodified suite passed this test, establishing genuine partial-write regression detection.
- [T-G3] `candidate/batch_test.go:46` verifies exact directory contents after invalid rows; `candidate/batch_test.go:128` exercises a real rename obstruction and retained prefix. The checks assert effects as well as failure, and would detect later writes or leaked temporary files.
- [T-G4] `candidate/batch_test.go:164` and `candidate/batch_test.go:185` run the actual binary, assert exit status, empty stdout, failure diagnostics, and resulting files. CLI builds and executions have bounded contexts. Filesystem fixtures use per-test temporary directories; the file-size limit is confined to a subprocess.

Bad

- None found.

Suggested changes

- None needed.

Limits: `rtk proxy go test -count=1 -timeout=90s ./...` passed (exit 0). The two targeted mutation runs each exited 1 at the expected assertion; neither was a build failure. Go 1.22 itself and Linux execution were not available/attempted. The partial-write test is explicitly limited to Darwin/Linux by its build constraint; this review verified it on Darwin. No environmental verification failure or listener rerun occurred. Candidate/original source hashes remained unchanged.

## Correctness & Compatibility — A+

Scope: The supplied one-record-to-batch CLI diff and the README's process/data contract. The single-record API and `main` process exit contract remain in scope. Go 1.22 module semantics, standard-library-only imports, darwin/arm64 runtime.

Coverage: Ordered per-row processing, empty input, field-count/CSV errors, ID and quantity validation, normalized values, duplicate replacement, all stop/error paths, accepted-prefix state, directory creation, actual exit/diagnostic behavior, and completed-file publication. Legacy machine-sized `strconv.Atoi` quantity parsing is unchanged; no unrelated compatibility policy was inferred.

Rationale: No introduced or worsened correctness issue found. Two independent, verified safeguards support A+: committing each accepted record before requesting later input protects prefix semantics, and writing/closing a temporary file before rename protects a previous accepted value from failed replacement. The streaming and real partial-write checks verify each separately, and targeted removal of either safeguard produces its distinct failure.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/main.go:24` sets an explicit two-field CSV contract and processes records sequentially, returning on read/validation/write failure. `candidate/batch_test.go:46`, `:90`, `:111`, and `:128` verify retained prefix, streaming, and no later filesystem effects.
- [C-G2] `candidate/main.go:51` stages a replacement in the destination directory, checks write and close results, publishes with rename, and removes the staging file on failure. The real partial-write test verifies that `web,7` survives a failing later `web,12345`; the direct-write mutation proves the safeguard's consequence.
- [C-G3] `candidate/main.go:70` keeps exit 2 on errors and writes diagnostics to stderr. The process tests at `candidate/batch_test.go:185` verify success, batch/duplicate, empty input, usage failures, malformed CSV, and I/O failures through the built executable.

Bad

- None found.

Suggested changes

- None needed.

Limits: The unmodified full suite passed on Go 1.26.5/darwin/arm64. Go 1.22 runtime and other operating systems were not executed; module/version and imports were inspected. Verification is finite, and the supplied contract does not establish arbitrary integer precision or additional filesystem durability semantics. Exact evidence is in `checks.json`; original/candidate source was preserved.

## Architecture & Design — A

Scope: Only the consequential new production seam for completed-record publication in `candidate/main.go:51`, and its relationship to the existing `run`/process boundary. No unrelated architecture audit.

Coverage: File-publication ownership, error propagation, use of the existing `io.Reader` seam, and realistic filesystem/process tests. Full supplied caller context is present: `main` invokes `run`, which invokes the private writer synchronously.

Rationale: No actionable architecture issue found. The extracted writer owns the temporary file from creation through close, rename, and failure cleanup; the batch caller owns ordering and stopping. A verified, relevant strength supports A without requiring another package or a test-only filesystem interface.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate/main.go:44` synchronously delegates only complete-record publication to `candidate/main.go:51`; errors return to the row-level caller with record/ID context. The real file-obstruction and partial-write tests verify this seam without replacing production filesystem behavior with a fake.

Bad

- None found.

Suggested changes

- None needed.

Limits: Assessment is restricted to these consequential production/test seams and the supplied CLI. No broader architecture or deployment claims are made. The source-preservation and verification evidence is recorded in `checks.json`.
