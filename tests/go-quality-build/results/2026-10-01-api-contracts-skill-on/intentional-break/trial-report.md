# Go API contracts skill-on trial report

## API decision

Implemented the accepted v2 breaking change. The module path is
`example.com/portnum/v2`; `Parse(string)` now returns `(int, error)` and accepts
only one or more ASCII decimal digits representing 0 through 65535. Leading
zeros remain valid. Empty input, signs, whitespace, non-digits, and out-of-range
values fail with a zero result and non-nil error. No v1 parsing shim was added,
consistent with the decision that v1 remains available at its original import
path.

## Files changed

- `go.mod`: changed module path to `example.com/portnum/v2`.
- `port.go`: implemented strict parsing and documented failure behavior.
- `port_test.go`: covered valid boundaries, leading zeros, invalid syntax,
  range overflow, and zero-on-error behavior.
- `cmd/portcheck/main.go`: updated the import and now reports invalid input to
  stderr with exit code 2.
- `docs/migration-v2.md`: documented import, return-value, accepted-input, and
  failure-behavior migration steps.

## Checks and results

- `go test ./...` — passed.
- `go vet ./...` — passed.
- `go run ./cmd/portcheck 00080` — printed `80`.
- Built the command and ran `/tmp/portcheck-v2 nope` — exit code 2, empty
  stdout, diagnostic on stderr.

## Limits

The command is the repository's example consumer, not an external downstream
integration. The invalid-input process behavior was checked by building and
running the binary; `go run` adds its own `exit status 2` wrapper text, so that
wrapper output was not used to judge the CLI contract. Tests verify non-nil
errors but do not promise a stable error identity or message.
