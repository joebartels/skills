# ledgerload streamed ordered batches

Changed `main.go` to read and validate one two-field CSV record at a time, write its normalized positive decimal quantity immediately, and continue until EOF. Duplicate IDs replace earlier contents in input order. The first input, CSV, directory or write failure returns an error containing record/operation context; already accepted writes remain and later records are not processed. Directory creation remains lazy, so empty input and invalid first records create no files. `main` preserves exit 0/2 and silent stdout. `go.mod` remains Go 1.22 with no added dependencies.

Expanded `main_test.go` with success and single-record regression cases, duplicate ordering, leading zeros/plus signs, quoting, CRLF and final rows without newline; invalid IDs, field counts, malformed CSV, invalid quantities and overflow; empty input and directory creation; portable directory/target obstructions; stdin read failures; and usage errors that cannot consume stdin. A controlled reader checks that writes are visible before the next row is supplied and that an invalid row stops further reads. Filesystem assertions check the full resulting record set, including unchanged earlier writes and absence of later writes. Eight actual compiled-binary cases verify exit status, empty stdout, success stderr, useful error stderr and retained filesystem effects.

Verification actually run:

- The new tests failed against the original single-record implementation before the change. This was run twice, with the second failure transcript preserved in `checks.json`.
- `rtk proxy gofmt -w main.go main_test.go` completed successfully. A later bounded `gofmt -l` check returned no filenames.
- `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go test -v -race -timeout=90s ./...` passed. The CLI tests build with a 30-second deadline and give each binary invocation a 5-second deadline.
- `go vet ./...` passed through `rtk proxy` with the required cache/toolchain environment and a 60-second subprocess deadline.
- `go version` reports Go 1.26.5 on darwin/arm64. The module version and APIs remain compatible with Go 1.22, but Go 1.22 itself was not installed or executed.
- Staticcheck is unavailable; no tools or dependencies were installed.

`checks.json` retains commands, outputs and exit codes. The test runner exposed combined stdout/stderr for the initial and race runs; separate streams were preserved for formatting, toolchain and vet checks.

Known limits and remaining behavior risks: quantity range remains the native `int` range accepted by the old `strconv.Atoi` implementation. CSV buffering may read bytes ahead, while records are validated and written strictly in order. File writes retain the original `os.WriteFile` behavior and are not transactional: tests exercise deterministic target-directory and directory-path failures, not real disk exhaustion or mid-write device failure. No maximum row size, symlink policy, concurrent-writer coordination or atomic replacement contract was added. The requested accepted-prefix behavior and supported process contracts have executable coverage; no further implementation work is pending.
