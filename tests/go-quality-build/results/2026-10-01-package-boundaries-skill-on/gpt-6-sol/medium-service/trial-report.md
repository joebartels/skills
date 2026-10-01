# Replay implementation trial

## Result

Implemented `cmd/replay` for one batch file argument. It trims each line, skips blanks, submits IDs in order, and stops with a nonzero exit on the first error. It reads `RECORD_DIR` and `NOTIFY_URL` like the server. The HTTP ingress and replay ingress both call `dispatch.Service.Submit`, which delegates validation, persistence, and notification sequencing to `internal/submit.Run`.

## Package decision

- `dispatch` retains the existing public `Service`, `Submit`, and `ServeHTTP` API and wires the operation to concrete implementations. Its HTTP status and body mapping is unchanged.
- `internal/submit` owns the shared validation and save-then-notify sequence. Its narrow store and notifier contracts are defined by the consumer, and it imports neither implementation.
- `internal/recordstore` owns the file representation (`queued\n`, mode `0600`). `internal/notification` owns the outbound POST and 204 response contract. These are separate because storage and the notification API have independent maintenance and expected changes.
- `cmd/replay` owns batch file parsing, environment configuration, and process exit. The dependency direction is replay/server -> dispatch -> submit and concrete adapters; submit -> standard library only.

## Verification

- `rtk run 'GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/gocache go fmt ./...'` — passed.
- `rtk run 'GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/gocache go test ./...'` — passed with localhost socket access. Tests cover ordered replay with trimmed/blank lines, invalid ID stop, notification failure stop with retained record, existing submit behavior, and HTTP status/body compatibility.
- `rtk run 'GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/gocache go build ./...'` — passed.

The sandbox blocked localhost listeners on the first test attempt; the suite passed when rerun with socket access. No retries or queue were added.
