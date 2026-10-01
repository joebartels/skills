# ledgerload streamed batches

Implemented ordered streaming CSV processing in `main.go`, with explicit two-field records, per-row validation and useful row/ID diagnostics. Each successful record is written before the next record is parsed. Empty input remains valid, duplicates replace in order, quantities normalize to decimal plus newline, and the CLI retains its exit 0/2 and stdout/stderr contracts. `go.mod` still declares Go 1.22 and uses only the standard library.

A temporary file in the target directory is written and closed before rename publishes the record. Failed writes or failed publication remove the temporary file and preserve any previously accepted value, including repeated IDs. The implementation owns and closes temporary files; it borrows stdin/stderr without closing them. README now describes the implemented batch behavior.

## Behavior checks

- Original single-record regression retained.
- Known byte oracles cover empty input, multiple records, duplicates, quoted CSV, CRLF, decimal normalization and missing final newline.
- Invalid field counts, IDs, quantities and malformed CSV after accepted records assert rejection, complete retained prefix, and absence of later records.
- Directory creation and record publication obstructions assert errors and retained state; input-reader failure preserves its accepted prefix.
- A staged reader observes that the first record already exists before the next read, rejecting full-batch prereading/prevalidation.
- Built executable tests assert actual exit status, empty stdout, appropriate stderr, and filesystem effects for success, validation/CSV failure, I/O failure and usage errors.
- On Darwin/Linux, a bounded child sets a two-byte RLIMIT_FSIZE, then execs the real command. Its first `keep,1` write succeeds, the later `keep,123` replacement partially writes and fails, and `later,2` remains absent. The original `keep` bytes survive and no temporary files remain.

## Verification actually run

1. Before implementation, `go test -count=1 -timeout=60s ./...` compiled and failed meaningful batch, later-error, streaming, executable-status and partial-write assertions against the original one-record implementation (exit 1).
2. Initial implementation passed the same suite (exit 0).
3. Temporarily replacing atomic publication with direct `os.WriteFile` made `TestCLIPartialWriteRetainsAcceptedDuplicate` fail with `keep = "12", want "1\\n"` (exit 1). The source was restored immediately. This proves the test detects destructive partial replacement rather than only an opening obstruction.
4. `gofmt -s -w main.go main_test.go write_failure_unix_test.go` succeeded; subsequent `gofmt -l` returned no files (exit 0).
5. `go vet ./...` passed (exit 0).
6. `go test -v -race -count=1 -timeout=90s ./...` passed, including all executable and partial-write checks (exit 0).
7. Toolchain evidence: `go version go1.26.5 darwin/arm64`. All Go commands used `GOTOOLCHAIN=local` and `GOCACHE=/private/tmp/go-quality-testing-cache`. Tests bound builds to 30 seconds and executable children to 5 seconds; the verification wrapper bounded processes to 110 seconds.
8. Staticcheck availability was checked; it is not installed. It was not run and nothing was installed.

Exact captured outputs and exit codes are in `checks.json`; early checks have merged stdout/stderr, final checks capture the streams separately.

## Limits and remaining behavior risks

The suite ran on Darwin/arm64 with local Go 1.26.5; the Go 1.22 module target and compatible standard-library APIs were preserved, but an actual Go 1.22 toolchain was not used for this run. The partial-write regression is compiled only on Darwin/Linux; Linux was not executed here. Non-Unix rename replacement behavior is not verified. Close/cleanup failures and concurrent filesystem mutation are not fault-injected. Publication provides per-record failure retention, without an fsync/crash-durability guarantee. Replacement publishes a new regular file with mode 0644; inode metadata, hard-link identity and existing symlink following are outside the README's byte/process contract.
