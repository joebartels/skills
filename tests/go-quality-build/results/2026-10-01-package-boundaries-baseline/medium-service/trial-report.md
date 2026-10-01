# Replay implementation trial

Status: complete.

The new `cmd/replay` reads one file argument, trims each line, skips blank lines, and calls `dispatch.Service.Submit` sequentially. It stops on the first validation, storage, notification, or read error. `main` exits with status 1 on error. Both binaries use `RECORD_DIR` and `NOTIFY_URL` and the same five-second HTTP client timeout.

Package decision: `dispatch` owns validation and the save-then-notify operation, with small `RecordStore` and `Notifier` interfaces at their consumer. `store.File` owns the existing `queued\n` record format and `notify.HTTP` owns the outbound POST contract. This keeps the independently maintained dependencies behind separate packages while both ingress paths share the operation. Notification failure still leaves the record in place; no retry or queue was added.

Verification:

- `rtk go test ./cmd/replay`: initially failed because `run` was undefined, as expected before implementation.
- `rtk go test ./...`: 6 tests passed in 5 packages.
- `rtk go test -race ./...`: 6 tests passed in 5 packages.
- `rtk go vet ./...`: no issues.
- `rtk run 'gofmt -l dispatch.go dispatch_test.go cmd/server/main.go cmd/replay/main.go cmd/replay/main_test.go store/file.go notify/http.go'`: no files listed.
- `rtk run 'GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/baseline/medium-service/.gocache go build -o /private/tmp/go-quality-build-eval.TijGDa/baseline/medium-service/replay ./cmd/replay'`: passed. Running that binary with a missing file returned exit status 1. The temporary binary and cache were removed.

Changed files: `dispatch.go`, `dispatch_test.go`, `cmd/server/main.go`, `cmd/replay/main.go`, `cmd/replay/main_test.go`, `store/file.go`, `notify/http.go`, and this report.
