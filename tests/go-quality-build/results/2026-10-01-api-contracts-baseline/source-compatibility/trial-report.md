# Trial report: source compatibility

## API choice

Added `NewWithSeparator(prefix, separator string) *Formatter`. Kept `New(prefix string) *Formatter` with its exact signature and colon default, which preserves existing calls and assignments of `New` to one-argument factory function types. An empty separator is accepted as an explicit value.

## Files changed

- `format.go`: stored the separator on `Formatter`, routed `New` through the new constructor with `":"`, and used the configured separator in `Format`.
- `format_test.go`: added coverage for a custom separator and an explicitly empty separator; retained the original default-format test.
- `README.md`: documented the constructor and an example.

## Checks and results

- `gofmt -w format.go format_test.go` — passed.
- Initial `go test ./...` — could not start because the default Go build cache under `/Users/jb/Library/Caches/go-build` was not permitted in this environment.
- `GOCACHE=/private/tmp/go-quality-build-api-eval/baseline-luna/source-compatibility/.gocache go test ./...` — passed (`ok example.com/recordfmt`).
- `GOCACHE=/private/tmp/go-quality-build-api-eval/baseline-luna/source-compatibility/.gocache go vet ./...` — passed (no diagnostics).

## Limitations

The separator is fixed when the formatter is constructed; changing it requires constructing another formatter. No migration is needed for current `New` callers.
