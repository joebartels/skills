# Trial report

## Package decision

- `internal/billing` owns invoice validation, charge sequencing, and the rule that only successful charges reach the ledger. Its narrow `Gateway` interface describes the capability it needs.
- `internal/invoicerow` owns the `ID,CENTS` row representation used for both settlement input and ledger writing. Its `Reader` and `Decode` provide a ledger reader for future exports. It rejects missing fields, invalid or nonpositive amounts, and IDs containing commas or line delimiters.
- `internal/paymentgateway` owns the provider's independently changing HTTP request and response protocol. It preserves the POST body bytes `ID,CENTS` and 204 success rule.
- `cmd/invoice` and `cmd/settle` compose billing with the concrete HTTP client, ledger file, and environment configuration. Settlement reads and charges rows in order, stopping at the first parse or charge error. `internal/catalog` and `cmd/shop` remain independent and unchanged. The project remains one module because the feature teams release together.

Import direction: commands → billing, invoicerow, paymentgateway; billing → invoicerow; paymentgateway → standard library. Billing does not import its concrete HTTP adapter. Catalog has no dependency on the billing packages.

## Files changed

- Changed: `internal/billing/billing.go`, `internal/billing/billing_test.go`, `cmd/invoice/main.go`.
- Added: `internal/invoicerow/row.go`, `internal/invoicerow/row_test.go`, `internal/paymentgateway/client.go`, `internal/paymentgateway/client_test.go`, `cmd/settle/main.go`, `cmd/settle/main_test.go`.

## Checks and results

- `go test ./...`: passed, 7 packages (RTK summary: 8 tests passed).
- `go vet ./...`: passed, no issues.
- `go list -deps ./cmd/invoice ./cmd/settle ./cmd/shop`: passed; confirms all three commands build their import graphs.
- Tests cover ledger bytes, malformed row rejection, reader round trip, gateway request bytes, and stop-on-first-error behavior for malformed rows and charge failure.

## Limitations

- Settlement is sequential and stops on error, but a rerun after a partial failure can charge already processed rows again. Idempotency requires a separate provider or operational contract.
- A successful gateway charge followed by a ledger write failure cannot be rolled back; this is the same ordering as the original billing behavior.
