# ledgerload ordered batches — 2026-10-01

Implemented in `/private/tmp/go-fresh-author-mm5bi5gc/task-6/module`.

## Changes and contract decisions

- `main.go` reads, validates and publishes one CSV record at a time. Empty input remains valid. Every successful row precedes the next row's processing; invalid input, malformed CSV, input-read errors and file failures stop the batch immediately.
- IDs retain the documented simple-filename checks. Quantities retain the existing `strconv.Atoi` accepted set and machine-integer range, including positive values with leading zeroes or a plus sign, and produce normalized decimal bytes followed by newline.
- Each record is written and closed in a temporary file in the destination directory, then renamed to the target. Failed duplicate replacements therefore cannot truncate an accepted value. The helper owns the temporary file and removes it on ordinary failure. Errors identify the record and the failed operation/ID.
- The actual executable retains exit 0 with empty stdout/stderr on success, and exit 2 with an actionable stderr diagnostic and empty stdout on failure. Missing/empty `--dir`, missing flag values, unknown flags and positional arguments are checked.
- `go.mod` remains `go 1.22`; only the standard library is used. No dependencies/tools were installed, no delegation occurred, and no commits or external changes were made.

## Tests and evidence

`batch_test.go` covers normalized multiple records, ordered duplicate replacement, quoted CSV IDs, empty input, invalid IDs and quantities, wrong field counts, malformed CSV, rejected replacements, and forbidden later effects. A staged reader checks that the first file is already published when input requests its next chunk, distinguishing streaming from whole-batch prevalidation. An independent read error checks prefix retention. Filesystem obstructions check failure before mutation and after accepted writes, preserved obstruction contents and absence of leaked temporary files.

The real CLI is built under a 30-second deadline and invoked under 5-second deadlines. Process tests assert exact exit status, stdout/stderr routing, diagnostic relevance and persisted file contents, including usage errors, directory creation, directory failure and mid-batch file failure. Expected file bytes are fixed independently of production formatting helpers. The original single-record regression remains.

`write_failure_unix_test.go` executes the real CLI from a child with a local 2-byte `RLIMIT_FSIZE`. A first `web,7` succeeds, its replacement `web,12345` partially writes then fails, and `later,4` must not be written. The accepted `web` bytes must remain `7\n`, and the directory must contain only that accepted file. The limit applies only to this child; the parent limits and disk capacity are untouched. This boundary ran successfully on Darwin/arm64, without skips or approval reruns.

Before implementation, the focused batch, streaming and process-failure tests compiled and failed against the one-record implementation. A temporary direct-write mutation of the completed implementation compiled and failed the during-write test on the specific damaged bytes (`web = "12"`, expected `"7\n"`). The mutation was restored before final verification. These expected failures are preserved separately in `checks.json`.

Actual final checks, all exit 0:

- `rtk proxy go test -race -timeout=60s ./...` — full final suite passed, including actual CLI and partial-write tests.
- `rtk proxy go vet ./...` — passed.
- `rtk proxy gofmt -l main.go main_test.go batch_test.go write_failure_unix_test.go` — no unformatted files.

The normal verbose suite also passed before the final diagnostic assertions were added; the final race suite exercises those additions. All recorded checks use `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`, with an outer 90-second deadline. `checks.json` preserves command arrays, cwd, non-secret environment, stdout, stderr, elapsed times and exit codes. Its entries labeled `baseline` and `direct-write-mutation` intentionally exit 1; they are signal checks, not unresolved failures.

## Guidance and remaining limits

Used `go-api-contracts`, `go-behavior-tests` (and its behavior-observations reference), `go-core-style`, and `go-interfaces-and-composition`. `go-package-boundaries` was not opened because package responsibility and import direction are unchanged. `selection.json` records offered names, exact opened paths and selection reasons. Only the authorized dispatch, module and relevant offered guidance were inspected.

The available local toolchain is Go 1.26.5 on Darwin/arm64. The module's Go 1.22 declaration and API compatibility were preserved, but no Go 1.22 compiler execution is claimed. `staticcheck` is not available, so it was not run or installed. The partial-write test is build-tagged for Darwin/Linux; Linux and other platforms were not executed.

Replacement uses rename, so target inode identity is not preserved; final file mode is 0644 and a symlink at the target path is replaced. The README specifies contents and process outcomes, not inode, symlink or permission preservation. Crash durability, concurrent filesystem mutation, and independent cleanup failures are not tested or promised; ordinary tested failures clean up temporary files and retain the accepted prefix.

## Artifacts

- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/module/main.go`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/module/batch_test.go`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/module/write_failure_unix_test.go`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/report/selection.json`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/report/checks.json`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/report/run_check.py`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/report/check_direct_write_mutation.py`
- `/private/tmp/go-fresh-author-mm5bi5gc/task-6/report/author-report.md`
