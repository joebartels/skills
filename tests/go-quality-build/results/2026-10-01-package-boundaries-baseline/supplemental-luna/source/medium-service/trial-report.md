# Replay implementation trial

## Changes

- Added `Service.Replay`, which reads trimmed lines, skips blank lines, submits sequentially, and returns on the first read or submission error.
- Added `cmd/replay`, accepting one file path and using `RECORD_DIR`, `NOTIFY_URL`, and the server's five-second HTTP timeout. Errors print to stderr and exit with status 1.
- Added `RecordStore` and `Notifier` interfaces with filesystem and HTTP adapters. `Submit` remains the shared validate, persist (`queued\n`), notify operation used by both ingress paths. Existing service fields and HTTP handler behavior remain in place.
- Added tests for successful sequential replay, invalid-ID stopping behavior, and notification-failure stopping behavior.

## Package decision

Storage and notification each have their own narrow interface because they are independently maintained and expected to change separately. The service owns validation and the operation order; the HTTP server and replay command both call `Submit`. No retry or queue was added.

## Commands and results

- `rtk gofmt -w dispatch.go dispatch_test.go cmd/replay/main.go` — completed.
- `rtk go test ./...` — build failed because the default Go build cache under `/Users/jb/Library/Caches/go-build` was not writable in the sandbox.
- `rtk env GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/gocache go test ./...` — passed: `example.com/dispatch`; replay and server command packages compiled.
- `rtk git diff --stat` — unavailable because this disposable copy is not a Git repository.
