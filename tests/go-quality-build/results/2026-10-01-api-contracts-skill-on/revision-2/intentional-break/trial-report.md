# Revised skill implementation trial

## API decision

Implemented the approved breaking v2 contract: the module path is
`example.com/portnum/v2`, and `Parse(string) (int, error)` accepts only one or
more ASCII decimal digits with value 0 through 65535. It accepts leading
zeros. Failures return zero and an error. The old module path and one-result
API are not retained in this module. The example CLI requires one argument,
prints the parsed number and newline on success, and writes its diagnostic to
stderr with exit code 2 on invalid input.

## Files changed

- `go.mod`: moved module to the v2 path.
- `port.go`: implemented strict parsing and documented its behavior.
- `port_test.go`: added boundary, invalid-input, leading-zero, and function
  assignment coverage.
- `cmd/portcheck/main.go`: updated the example consumer and CLI behavior.
- `cmd/portcheck/main_test.go`: covered CLI success, invalid input, argument
  count, output streams, and exit codes.
- `docs/migration-v2.md`: documented import, error-handling, and input changes.

## Checks and results

- `GOCACHE=/private/tmp/go-quality-build-api-eval/gocache rtk go test ./...` —
  passed (16 checks across 2 packages).
- `GOCACHE=/private/tmp/go-quality-build-api-eval/gocache rtk go vet ./...` —
  passed with no issues.
- Initial test attempt using the default Go build cache failed because that
  cache path was unavailable in the sandbox. Re-running with a disposable
  cache under `/private/tmp` passed.

## Limits

The trial ran against the supplied isolated fixture only; it did not inspect
downstream repositories or a real external consumer. CLI contract tests call
the consumer's `run` function directly rather than launching a built process.
The fixture is not a Git repository, so `git diff --check` was unavailable.
