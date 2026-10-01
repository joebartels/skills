# Skill-on trial report: source compatibility

## API decision

Added `NewWithSeparator(prefix, separator string) *Formatter` while preserving the exact `New(prefix string) *Formatter` signature. This keeps existing formatter factory assignments source-compatible and lets callers explicitly select an empty separator. `New` continues to produce colon-separated output. The formatter zero value also retains colon behavior by distinguishing an unset separator from an explicitly empty one.

## Changed files

- `format.go`: added separator configuration and the additive constructor.
- `format_test.go`: covered custom and empty separators, the default constructor, zero-value behavior, and the existing `New` function type expected by factory tables.
- `README.md`: documented the new constructor and usage.

## Checks and results

- `go test ./...` — passed (4 tests in 1 package).
- `go vet ./...` — passed with no issues.

## Limits

This trial used only the supplied disposable fixture. Independently released reporting tools are not present, so their real factory tables and downstream behavior were not compiled. The test's function assignment verifies the documented existing function type at a consumer import boundary, but does not replace those integrations.
