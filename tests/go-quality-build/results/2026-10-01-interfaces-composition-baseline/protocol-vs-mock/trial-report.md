# Trial report: protocol vs mock

## Decisions

- Added `reportkit.CSV` as a value type implementing the existing `Encoder` protocol. It writes a `name,value` header, writes record strings in order, uses `encoding/csv` quoting, flushes, and returns writer errors.
- Kept `Report` as the concrete rendering wrapper. The proposed `Exporter` interface would repeat its single `Render` method but has no production consumer here; a mock of `Report` only helps when some caller accepts that interface. The existing `Encoder` is already the useful extension seam. The tests exercise CSV's writer error directly.
- Added an external-package example with a host-defined encoder, showing how a host supplies its own format.
- CSV line endings follow `encoding/csv` defaults (LF); quoting follows its standard behavior.

## Files changed

- `report.go`: added the `CSV` encoder.
- `report_test.go`: added CSV quoting, empty-input, and encoding-failure tests, plus an external-package host encoder example; retained the JSON test.
- `trial-report.md`: recorded the trial decisions and checks.

## Checks and results

- `go test ./...` — passed (`ok example.com/protocol-vs-mock`).
- `go vet ./...` — passed with no diagnostics.

## Limitations

- The fixture has no consumers that would demonstrate an `Exporter` abstraction usefully; this trial assesses that proposal from the available API surface only.
- The failure test exercises the underlying writer through `CSV.Encode`; `Report.Render`'s existing private-buffer behavior remains as documented.
