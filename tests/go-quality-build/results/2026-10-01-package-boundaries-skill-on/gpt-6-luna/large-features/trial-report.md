# Large application implementation trial

## Package decision

- `internal/billing` owns invoice validation, charge sequencing, and ledger access. It depends on the shared row codec and calls a narrow `Gateway` contract.
- `internal/gatewayhttp` owns HTTP request and response details. The invoice and settle commands wire it to billing, so gateway protocol changes stay outside billing rules.
- `internal/invoicecodec` owns the shared `ID,CENTS` representation. Settlement reads rows incrementally through it; billing uses it for ledger writes and the future ledger reader.
- `internal/catalog` remains untouched and continues to own its independent JSON lookup and HTTP behavior.
- All packages remain in the existing module and release unit. Import direction is commands → billing, gateway HTTP, and codec; gateway HTTP → billing contract; billing → codec. No catalog dependency was introduced.

## Changed files

- `cmd/invoice/main.go`
- `cmd/settle/main.go`, `cmd/settle/main_test.go`
- `internal/billing/billing.go`, `internal/billing/billing_test.go`
- `internal/gatewayhttp/client.go`, `internal/gatewayhttp/client_test.go`
- `internal/invoicecodec/codec.go`, `internal/invoicecodec/codec_test.go`

## Verification

- `rtk gofmt -w cmd/invoice/main.go cmd/settle/main.go cmd/settle/main_test.go internal/billing/billing.go internal/billing/billing_test.go internal/gatewayhttp/client.go internal/gatewayhttp/client_test.go internal/invoicecodec/codec.go internal/invoicecodec/codec_test.go` — formatted changed Go files.
- `rtk proxy env GOCACHE=/private/tmp/go-quality-build-cache go test ./...` — passed all packages, including row validation/round-trip, exact ledger bytes, gateway request bytes/status, and settlement stop-order tests.
- `rtk proxy env GOCACHE=/private/tmp/go-quality-build-cache go vet ./...` — passed.

The gateway tests use an in-process `RoundTripper`; this environment disallows binding local test-server sockets. The first test attempt also found the default Go build cache unwritable, so verification used a temporary cache under `/private/tmp`.
