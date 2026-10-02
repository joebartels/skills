# ledgerload streamed ordered batches

Implemented the README contract in the isolated module. No dependencies, toolchain directive, minimum Go version, external changes, commits or delegation were added.

## Changes

- `main.go`: read CSV records in order until EOF; require two fields; validate each ID and positive base-10 quantity before that record's effect; normalize decimal bytes; stop at the first parsing, validation, read or file error with record/stage context. Empty input succeeds without creating a directory.
- Each complete record is written and closed in a same-directory temporary file before rename. This protects an earlier accepted duplicate from a later partial write. Temporary files are removed on ordinary failure paths. No batch prevalidation or rollback discards accepted records.
- `main_test.go`: preserve the original one-record regression; build and run the actual executable to assert exact status 0/2, empty stdout, useful failure stderr and complete directory/file state. Cases cover ordered duplicates, leading zeros, plus signs, quoted IDs, a final row without newline, empty input, invalid IDs, quantity neighbors/overflow, CSV shape/malformed CSV, invalid first and later records, invalid replacement, usage, directory creation failure and publication failure. A controlled reader confirms a prior write is visible before the next input read fails.
- `main_fsize_test.go`: Darwin/Linux child-only `RLIMIT_FSIZE` fixtures exec the actual executable. They exercise failure on the first write and a write that progresses four bytes before failing while replacing an earlier accepted ID. They assert retained prior bytes, absent later writes, cleanup, exit 2 and diagnostics.

## Verified results

All verification used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`, finite subprocess/test deadlines and `rtk proxy`. Exact command arrays, cwd, non-secret environment, stdout, stderr and statuses are in `checks.json`.

- `go version`: Go 1.26.5, darwin/arm64. `go.mod` remains Go 1.22; the source uses APIs available at that minimum.
- `gofmt` completed and a formatting listing was empty.
- `go test -count=1 -timeout=45s -v ./...`: passed, including actual process status/streams/effects and both file-size failures.
- Controlled temporary mutation: replace safe publication with direct `os.WriteFile`, run the partial duplicate regression. It compiled and failed on the meaningful assertion `web = "1234", want "7\n"`, exit 1. The source was restored in a finally block.
- `go vet ./...`: passed.
- `go test -race -count=1 -timeout=45s ./...`: passed after restoring the mutation.
- Final `go test -count=1 -timeout=45s ./...`: passed after adding the no-final-newline success boundary.
- `staticcheck ./...`: unavailable; RTK reported that the executable was not found. Nothing was installed.

## Limits and remaining risks

The tested host was Darwin/arm64 with local filesystem and subprocess boundaries. No listeners were needed or restricted, and no escalation/rerun was needed. The file-size fixture is intentionally compiled only for Darwin/Linux; Linux and other platforms were not executed here. An actual Go 1.22 toolchain was not available for a minimum-version run. Staticcheck could not run.

The README does not promise crash durability, concurrent-writer coordination or preservation of inode/link metadata; these were not added or verified. Same-directory rename replaces the directory entry. File close/permission-change errors are handled but not individually forced. Ordinary error cleanup is checked; interruption or a filesystem that refuses cleanup can leave a temporary file. ID/quantity/parser cases assert behavior without treating exact diagnostic wording as a public contract.

## Guidance and selection

Read only go-api-contracts, go-behavior-tests and its behavior-observations reference, go-core-style, and go-interfaces-and-composition. Package-boundary guidance was omitted because the package and responsibility were unchanged. `selection.json` records all offered names, decisions and actual opened paths.
