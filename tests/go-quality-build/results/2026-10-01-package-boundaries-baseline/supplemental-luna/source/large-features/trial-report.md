# Blind baseline trial report

Implemented the requested settlement command and boundary changes in this disposable fixture.

## Package decision

- `internal/invoice` owns the shared `ID,CENTS` row model and codec. It validates required IDs, positive integer amounts, and rejects commas or CR/LF in IDs. Its streaming reader stops at the first malformed row or callback failure. Settlement input and ledger reads share that reader; billing uses the same encoder for persisted rows.
- `internal/billing` owns charge validation/orchestration and ledger persistence. Its `Gateway` interface is defined on the consumer side, keeping billing rules independent from HTTP.
- `internal/gateway` owns the HTTP gateway protocol and implements the billing contract. Command packages assemble the HTTP adapter and billing service.
- `internal/catalog` remains independent and unchanged. It continues to own its unrelated JSON HTTP behavior.

`cmd/settle FILE` processes rows sequentially and exits nonzero at the first parse or charge error. Both commands use `LEDGER` and `GATEWAY_URL`. Ledger and gateway wire bytes remain `ID,CENTS\n` and the POST body `ID,CENTS` respectively.

## Verification

- `rtk go test ./...` — passed (8 tests across 7 packages).
- `rtk git status --short` — unavailable because this disposable fixture is not a Git repository.

Tests cover codec round-trip and invalid records, stop-on-first-invalid-row behavior, shared ledger decoding, billing persistence only after gateway success, validation before gateway calls, and exact HTTP request bytes/method.
