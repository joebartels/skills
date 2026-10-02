# Remote snapshot implementation

Implemented `Refresh` in `mirror.go` without changing the exported function type, `Item`, `Write`, module path, Go 1.22 directive, or standard-library dependency policy.

The operation makes a GET using the supplied client and context, requires HTTP 200, reads the complete response, and rejects malformed/trailing JSON, non-array `null`, and items with blank trimmed codes or labels. It preserves nonblank values and publishes lowercase `code` and `label` fields. Empty arrays publish `[]`. A temporary file beside the destination is written and closed before rename publication, and failure paths remove staging files. Acquired response bodies are closed.

Added external-consumer regression tests in `refresh_test.go` and platform-specific tests in `refresh_unix_test.go` (Linux/macOS build constraint). Independent fixed output bytes protect representation and value preservation. Tests cover supplied-client routing, GET and endpoint/query, repeated populated/empty updates, first-time creation, response and transport failures, in-flight context cancellation, response-body closure, an error after complete-looking JSON has been read, exact prior bytes, rejected replacements after an accepted snapshot, subsequent recovery, and publication cleanup. The original local `Write` regression remains.

The Unix tests observe an already-open old handle and a fresh open after publication. A child process with an eight-second test deadline and ten-second parent context applies a 64-byte file-size limit; an independent probe confirms an actual partial write followed by an error. A larger Refresh then must fail while preserving the complete accepted snapshot and leaving no staging artifact. Limits and SIGXFSZ handling stay local to that child.

## Verification actually run

All Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`; shell commands used `rtk proxy`. Process deadlines were 120 seconds, and test deadlines were 20 seconds or 45 seconds for race/repeat/shuffle. Full stdout, stderr, commands, and exit codes are preserved in `checks.json`.

- Local toolchain: Go 1.26.5, darwin/arm64.
- Initial baseline test invocation was prevented by sandbox denial of a loopback listener; its diagnostic is retained.
- Baseline rerun with local-listener permission compiled and failed on the stub's missing behavior, rather than an unrelated crash or timeout.
- `gofmt -s -w` completed; final `gofmt -l .` returned no files.
- `go vet ./...` passed.
- `go test -v -timeout=20s ./...` passed.
- `go test -race -count=3 -shuffle=on -timeout=45s ./...` passed.
- A temporary direct `os.WriteFile` counterexample compiled and produced the intended assertion failures: the old handle contained new bytes and the previous snapshot became a truncated partial JSON write. The production source was restored in a `finally` block.
- `go test -timeout=20s ./...` passed after restoration.
- `go list -m -json` confirms `example.com/mirror` still declares Go 1.22.
- `staticcheck` was unavailable and was not installed, as instructed.

## Decisions and verification limits

Keep all implementation in the existing `mirror` package and retain `*http.Client`; no storage interface, constructor, global hook, new package, or dependency was introduced. Tests use HTTP's existing RoundTripper extension point for response lifecycle observations and a real loopback server for request routing and cancellation. No downstream repository was available; the exact exported function assignment is a representative consumer compatibility check.

Execution verified macOS filesystem behavior. The Linux/macOS-specific tests were compiled and run on macOS; Linux compilation/execution and Windows replacement semantics were not checked; no equivalent Windows guarantee is claimed. The available runtime was Go 1.26.5, so Go 1.22 execution was not performed. The source uses APIs available by Go 1.22, and the module directive was unchanged.

The contract does not specify crash durability, response-size limits, existing file permissions/ownership, symlink policy, or ordering of competing Refresh calls. The implementation does not fsync, buffers the response and encoded snapshot in memory, uses mode 0644 for publication, and lets filesystem rename determine competing publication order. Cleanup removal failures and response-body Close errors are not surfaced; tests verify Close is called and normal cleanup succeeds, without claiming fault injection for those errors. Persistent external filesystem failures can therefore leave a temporary artifact even though the prior destination is protected. No read bound or cancellation polling is added to the synchronous filesystem stage.

No remaining implementation work is known within the specified task.
