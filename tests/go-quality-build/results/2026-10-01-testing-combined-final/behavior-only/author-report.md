# Standalone indexer implementation

Implemented the README contracts in `/private/tmp/go-fresh-author-s21flpf_/task-1/module` without dependencies, delegation, commits or external changes. Kept `go 1.22`, exported signatures, the `Record` JSON tags and existing positional `put` grammar.

## Changes

- `index.go`: CSV records are parsed and applied sequentially with exactly two fields, shared key/nonblank validation, accepted-prefix effects and wrapped validation identity. Existing `Put` still permits arbitrary text, including empty text. A private publication helper writes a sibling temporary file and renames only after writing and closing successfully.
- `refresh.go`: one context-aware GET through the supplied client configuration; redirects are rejected to prevent extra GETs. Only 200 and one non-null JSON array are accepted. Read, decode, validation and publication failures leave prior bytes intact. Successful output preserves record order/text and emits compact JSON plus newline, including `[]\n`. Every acquired response body is closed.
- `serve.go`: immediate synchronous callbacks, interval measured after completion, cancellation checks before startup, and one release after callback completion. Error joins preserve independent work/release causes. Cancellation-only wrapped/joined chains are suppressed during ordinary parent shutdown; independent and non-comparable callback errors survive.
- `cmd/indexer/main.go`: `apply` consumes stdin; `watch` validates arguments, owns signal cancellation and closes client idle connections only after `Serve` ends. Existing exit 0/2 and stdout/stderr conventions remain.

Kept the existing library and command packages. Concrete `http.Client` and callback dependencies already express the necessary variation; no interface, options framework or package split was needed.

## Regression evidence

- `index_test.go`: compatible consumer function assignments, old put behavior, exact key rules, independent roots, CSV quoting/text/order/duplicates, parse/validation/read/rename failures, accepted prefixes and forbidden later rows.
- `refresh_test.go`: independent known JSON bytes, empty arrays, complete old-reader/new-opener snapshots, exactly one GET and context propagation, redirect/206/503 rejection, read/decode/trailing-data/validation/publication failures, error identity, body closure and temporary-file cleanup.
- `serve_test.go`: invalid/already-canceled startup, each work/release error combination, cancellation with independent or non-comparable errors, callback cleanup before release/return, successful recurrence after a long first callback, completion-based spacing, no overlap and independent invocations.
- `write_failure_unix_test.go`: bounded child processes on Darwin/Linux set a local 64-byte file-size limit and ignore SIGXFSZ. Actual writes fail after progress; prior Put/Refresh bytes and the last accepted CSV replacement survive, earlier rows remain, later rows are absent and temporary files are removed.
- `cmd/indexer/process_test.go`: builds and runs the actual executable, checks exit status and both streams, persisted Put/CSV outcomes and usage failures, rejects a success-shaped 206 first refresh, observes two successful watch publications plus a third in-flight GET, and exercises both interrupt and SIGTERM cancellation through process exit.

A temporary direct `os.WriteFile` publication mutation compiled and produced concrete retained-byte assertion failures in all three partial-write cases (Put, CSV and Refresh). It was restored in a `finally` block, confirmed byte-identical by SHA-256, and the affected test passed afterward. The mutation failure is intentional test-signal evidence, not a remaining implementation failure.

## Actual checks and execution boundary

All recorded build/test/static-check commands used `GOCACHE=/private/tmp/go-quality-testing-cache`, `GOTOOLCHAIN=local`, `rtk proxy`, and finite deadlines. `checks.json` preserves command arrays, cwd, environment, stdout, stderr, exit codes, elapsed times and source hashes for the later checks.

1. Formatting completed; final `gofmt -l .` produced no output.
2. Sandboxed `go test -timeout=45s ./...` passed the library and failed the command suite when `httptest` could not bind a local listener (`operation not permitted`). Both restricted attempts are preserved.
3. Automatic approval granted an unsandboxed rerun of the unchanged source. `go test -timeout=45s ./...` passed both packages. Source hashes match the immediately preceding restricted attempt.
4. `go vet ./...` passed.
5. Approved `go test -race -shuffle=on -count=3 -timeout=90s ./...` passed both packages on the same source. The library took 11.096 seconds and the command package 3.016 seconds.
6. The publication mutation failed on meaningful retained-state assertions; the restored-source targeted test passed with `-count=1 -timeout=30s`.

## Limits and remaining risks

- Execution used the installed Go 1.26.5 Darwin/arm64 toolchain. The module minimum remains Go 1.22 and implementation/test APIs are compatible with that minimum, but an actual Go 1.22 toolchain was unavailable and was not installed.
- POSIX replacement/snapshot semantics were exercised on Darwin. Windows replacement behavior is outside the execution claim. The file-size failure regression runs only on Darwin/Linux.
- Local HTTP process tests required the approved unsandboxed boundary; the original restricted failure remains in the report. Library HTTP tests use the supplied transport boundary; real local HTTP is exercised by command tests.
- Staticcheck was unavailable and was not installed. Vet and race detection passed.
- No claim is made about crash durability (`fsync` is not part of the requested atomic publication contract), hostile unbounded remote payloads, external downstream integration, or forced termination of a non-cooperative callback. `Serve` deliberately waits for started cooperative work before release.

Reports: `selection.json`, `checks.json`, `mutation.json`, and this file in `/private/tmp/go-fresh-author-s21flpf_/task-1/report`.
