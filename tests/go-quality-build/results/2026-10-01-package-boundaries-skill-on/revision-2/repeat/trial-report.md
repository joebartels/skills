# Trial report

## Package decision

Kept `dispatch.Service` as the compatibility facade and HTTP ingress. Its `Submit` delegates to `internal/operation`, which owns ID validation and the write-then-notify sequence. The concrete file format and outbound HTTP protocol now live in `internal/records` and `internal/notify`; `dispatch` composes them. The new `replay` package owns line-oriented input handling and invokes the same `Service.Submit` as HTTP. `cmd/replay` owns argument parsing, environment configuration, and process exit behavior. The dependency direction is ingress/commands → operation and concrete adapters; the operation depends only on its small consumer interfaces.

This split is warranted by the prompt’s independent change expectations for storage and notification. The HTTP status/body mapping remains in the facade. File contents remain `queued\n` with mode `0600`, and the notification remains a POST carrying the ID and expecting 204.

## Files changed

- `dispatch.go`: retained public service configuration, submission API, and HTTP compatibility behavior; composes operation and adapters.
- `internal/operation/operation.go`: common validation and sequencing.
- `internal/records/store.go`: record file persistence.
- `internal/notify/client.go`: outbound notification protocol.
- `replay/replay.go`: sequential trimmed line processing; skips blanks and stops at the first operation or scanner error.
- `cmd/replay/main.go`: accepts one file argument, reads `RECORD_DIR` and `NOTIFY_URL`, and exits nonzero on error.
- `replay/replay_test.go`: success, invalid ID stop, notification failure stop, and persistence checks.

## Checks and results

- `go test ./...` passed with `GOCACHE` redirected to `/private/tmp` and permission for loopback listeners: root package and replay tests passed; remaining packages compiled.
- `go build ./cmd/server ./cmd/replay` passed.
- The initial sandboxed test attempt was blocked by the default Go cache and loopback listener restrictions; the authorized rerun succeeded.

## Limitations

Tests exercise replay behavior through `replay.Run` and the shared service operation, but do not launch the compiled binary to assert its OS exit code. The command’s error path prints the error and calls `os.Exit(1)`. No retry or queue behavior was added.
