# Large features trial report

## Result

Implemented `cmd/settle`, shared billing row encoding/decoding, a reusable ledger reader, and an independent HTTP gateway adapter. Settlement charges rows in order and returns on the first malformed row or charge error; `main` exits nonzero on that error. It uses `LEDGER` and `GATEWAY_URL`.

## Package decision

- `internal/billing` owns invoice validity, the `ID,CENTS` ledger format, reading ledger rows, and the rule to record only successful charges. Its `Gateway` interface describes the capability it needs.
- `internal/gateway` owns the remote HTTP request and response protocol. It imports no billing package, so its wire format can change without changing billing rules.
- `cmd/invoice` and `cmd/settle` compose billing with the gateway adapter and environment configuration. `internal/catalog` remains separate, with its JSON HTTP behavior untouched.
- Import direction: commands → billing and gateway; billing → standard library; gateway → standard library. The gateway adapter satisfies billing's interface by method shape.

## Compatibility and verification

- Successful ledger writes remain `ID,CENTS\n`; gateway requests remain `POST` with `ID,CENTS` body, no newline; invoice command retains its arguments and configuration.
- Added tests for row round trips and rejection, ledger reading, billing side effects, gateway request bytes/status, settlement early stop, and environment-based command composition.
- `rtk go test ./...` — passed: 11 tests in 6 packages.
- `rtk go vet ./...` — passed: no issues.

## Changed files

- `internal/billing/billing.go`, `internal/billing/row.go`, and their tests
- `internal/gateway/http.go` and its tests
- `cmd/invoice/main.go`
- `cmd/settle/main.go` and its tests
