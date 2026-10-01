# Trial report

## API decision

Implemented the accepted breaking v2 contract: module/import path
`example.com/portnum/v2`, with `Parse(string) (int, error)`. Parsing accepts
one or more ASCII digits in the range 0..65535, including leading zeros;
invalid input returns `(0, non-nil error)`. No v1 shim was retained. The
example CLI reports parse failures to stderr and exits 2.

## Files changed

- `go.mod`: switched the module path to `/v2`.
- `port.go`: implemented strict ASCII decimal parsing and documented its
  behavior.
- `port_test.go`: covered zero, leading zeros, range boundaries, signs,
  whitespace, non-ASCII digits, empty input, and overflow/range failures; a
  compile-time function assignment pins the new signature.
- `cmd/portcheck/main.go`: migrated the consumer import and error handling.
- `README.md`: corrected the release and API summary.
- `docs/migration-v2.md`: documented import, error-handling, and input changes.

## Checks and results

- `gofmt` completed for Go source files.
- `go test ./...` passed (`example.com/portnum/v2` and the CLI package).
- `go vet ./...` passed.
- Built and exercised the CLI binary: valid `00080` printed `80` with exit 0;
  invalid `+80` printed only a diagnostic to stderr with exit 2; wrong argument
  count printed usage to stderr with exit 2.
- The initial Go test invocation could not write the default user cache in
  this sandbox; reran tests and vet with `GOCACHE` inside this disposable case
  directory, where both passed.

## Limits

Verification covers the fixture module and its shipped example consumer; no
downstream repository consumers were available. CLI behavior was checked on
the current host/toolchain, with the module's declared Go 1.22 compatibility
not separately exercised using a Go 1.22 toolchain.
