# Independent changeset 1 review

Reviewed the supplied `original` → `candidate` trees for ledgerload, against the original README and request to implement streamed ordered batches with high-quality success/failure tests. No other candidate or controller material was inspected. All graded issues concern introduced/worsened behavior or the requested new test scope; the old single-record limitation is the change request, not a separately counted defect.

## Testing — A+
Scope: Complete supplied ledgerload changeset under `/private/tmp/go-independent-review-z6n_lo0k/changeset-1`; standard-library CLI, module `go 1.22`; execution used Go 1.26.5 on darwin/arm64.
Coverage: Reviewed every implementation/test file and the README. Assessed normalization, ordered duplicate replacement, quoted IDs, empty input, usage, invalid IDs/quantities, malformed CSV/field counts, input errors, directory/rename/write failures, accepted-prefix preservation, stop-before-later-write behavior, real process exit/stdout/stderr, and resource ownership. Checked assertion signal with two disposable mutations.
Rationale: No actionable testing issue found. Two independent verified safeguards support A+: (1) the controlled reader checks that a record is committed before a subsequent input read/error, detecting loss of streaming and its accepted prefix; (2) a child-only file-size limit forces a partial duplicate write through the real CLI and checks that the prior value remains intact. These control input sequencing and destructive output failure, respectively; neither is inferred from test count or coverage.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/main_test.go:179`–208 observes filesystem effects inside the next `Read`, checks the original input error with `errors.Is`, and verifies the final prefix. The `mutation-pre-read` copy failed at lines 192 and 208 because `web` was absent, demonstrating that the test rejects whole-input pre-reading before effects.
- [G2] `candidate/main_fsize_test.go:17`–62 limits only a short-lived child and executes the built command. The partial duplicate case asserts the previous `7\n` file, complete directory contents, exit 2, empty stdout, and a record-specific diagnostic. The `mutation-direct-write` copy failed because `web` became `1234`; it did not merely assert that an error occurred.
- [G3] `candidate/main_test.go:39`–60 and 93–175 check actual process contracts and exact filesystem outcomes across accepted/failed records. Directory obstruction exercises a real publication failure without replacing production filesystem behavior with a fake. Process/build operations have bounded contexts; files and binaries live in `t.TempDir`.

Bad

- None found.

Suggested changes

- None needed.

Limits: The unchanged candidate suite passed with `go test -count=1 -timeout=45s ./...` (exit 0). Both targeted mutations exited 1 for expected behavioral assertions; no candidate source was repaired or altered. The file-size test is tagged darwin/linux and executed on Darwin only. Linux, other operating systems, and a Go 1.22 executable were not run. Chmod/close failures were traced in code rather than forced individually; no crash-durability or concurrent-writer guarantee is stated. Exact prefixed commands, working directories, non-secret explicit environment, stdout/stderr, and exits are in `output/checks.json`.

## Correctness & Compatibility — A+
Scope: Same complete supplied CLI changeset and original README contract; supported behavior reviewed through `run`, `writeRecord`, and `main`.
Coverage: Traced normal/empty/invalid CSV, quantity and filename validation, accepted-prefix writes, duplicates, read and file failures, diagnostic/exit behavior, and the preserved Go directive/dependency boundary. Source inspection and executed tests support the assessed behavior.
Rationale: No introduced correctness or promised-consumer compatibility defect found. Two independently verified safeguards support A+: record validation/publication completes before requesting the next record, so later input failure cannot erase accepted work; temporary-file publication checks write/chmod/close before rename, so a failed duplicate cannot truncate the prior accepted value. Their distinct failure consequences were demonstrated by the pre-read and direct-write mutations.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/main.go:24`–47 reads two-field CSV records sequentially, validates before effects, and stops on the first read/input/file error. Error messages identify the affected record; wrapped input/file causes remain inspectable. Batch/process tests and the stream observer passed, including invalid later replacement and no subsequent writes.
- [G2] `candidate/main.go:51`–68 writes and closes a temporary file in the destination directory before publishing it, with deferred removal on all returns. The partial duplicate write test passed unchanged and rejected direct truncating writes, verifying preservation of meaningful already accepted state under a real failed write.
- [G3] `candidate/main.go:14`–22 and 71–75 keep usage validation and exit 2/error diagnostic behavior; successful and empty batches produce no stdout. Real CLI tests validate these boundaries. `go.mod` remains `go 1.22` with no external dependencies.

Bad

- None found.

Suggested changes

- None needed.

Limits: Verification uses Go 1.26.5/darwin/arm64 with module language version 1.22, not an actual Go 1.22 executable. Platform-specific publication/file-size behavior was not verified elsewhere. Existing machine-int quantity range and standard CSV conventions remain unchanged; the README does not promise arbitrary-precision quantities, crash durability, symlink-following behavior, or preservation of an existing file's mode. No environmental verification failures occurred. See `output/checks.json` for exact evidence and `output/source-manifest.json` for supplied source hashes.

## Architecture & Design — Not applicable
Scope: Only consequential production/test seams in this supplied changeset, as authorized by the dispatch.
Coverage: Inspected the existing `run(args, io.Reader, io.Writer)` boundary, new internal `writeRecord`, real CLI tests, and child-only fault setup.
Rationale: The production injection boundary is unchanged. The new helper owns a local file operation and does not introduce a public API, replacement dependency, background lifecycle, or test-only production hook. Its publication behavior is graded under correctness, and the subprocess isolation under testing; there is no consequential new production/test seam requiring an architecture card.
Limits: This is not a general architecture audit. No claim is made about unrelated designs outside the supplied trees.
