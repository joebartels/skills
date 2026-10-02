# Format documentation maintenance

Changed only `format.go`'s doc comment to: `Format returns s enclosed in square brackets.` The implementation remains identical. Applied gofmt; no formatting changes were needed. The module still declares Go 1.22, and the existing tests remain byte-for-byte unchanged.

Checks actually run:

- `rtk proxy gofmt -w format.go`: exit 0, 15-second process deadline.
- `rtk proxy gofmt -l format.go format_test.go`: exit 0 with empty output, 15-second process deadline.
- Exact source comparison confirmed only the comment changed; SHA-256 comparisons confirmed README.md, format_test.go, and go.mod are unchanged.
- `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go test -v -race -timeout=20s ./...`: exit 0; TestFormat passed. The process deadline was 60 seconds.
- `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go vet ./...`: exit 0, 30-second process deadline.

The full command stdout, stderr, exit codes, deadlines, and unchanged-file hashes are preserved in `checks.json`. Guidance decisions and actual opened input files are recorded in `selection.json`.

Known limitation: staticcheck is unavailable, so it was skipped without installing tools, as required by the task. Existing coverage was not expanded because the request explicitly keeps tests unchanged. The function implementation is unchanged, so this patch introduces no behavior change; risks in untested preexisting inputs remain unchanged.
