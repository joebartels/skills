# Trial report

## Package decision

Extracted `internal/operation` because both the HTTP server and batch replay need the same validation and persistence-then-notification sequence. The existing root `dispatch` package remains the HTTP facade and preserves its `Service` fields, `Submit` method, status mapping, and response body. `cmd/replay` calls the shared operation directly. `internal/recordstore` owns the record file encoding, and `internal/notifier` owns the outbound HTTP request protocol; storage and notification have independent maintenance pressure. Commands/facade wire these concrete implementations into the operation. The operation imports only its standard-library needs and defines the narrow consumer contracts; it does not import adapters.

## Files changed

- `dispatch.go`: retained the HTTP facade and delegated submission to the shared operation.
- `internal/operation/operation.go`: added shared validation and sequencing contracts.
- `internal/recordstore/files.go`: moved file persistence and the exact `queued\n` bytes into the storage implementation.
- `internal/notifier/http.go`: moved the outbound POST and 204 response check into the notifier implementation.
- `cmd/replay/main.go`: added the file-driven replay binary; trims lines, skips blanks, stops on the first failure, and exits nonzero on argument, file, scanner, or submission errors.
- `cmd/replay/main_test.go`: covered trimmed successful replay, invalid ID stop behavior, and notification failure stop/persist behavior.

## Checks and results

- `GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/gocache go test ./cmd/replay ./internal/...` — passed.
- `GOCACHE=/private/tmp/go-quality-build-eval.TijGDa/gocache go test -run '^$' ./...` — passed compilation across all packages.
- A full `go test ./...` run was attempted. The fixture's original HTTP tests and initial replay tests use local listeners, which this sandbox prohibits (`bind: operation not permitted`). The replay tests were changed to use an in-memory `RoundTripper` and passed; original root package tests could not run here.

## Limitations

The server's original HTTP status/body and stored-byte assertions are present in the fixture's existing tests, but listener restrictions prevented executing them. Compilation passed, and the facade retains the original method and response logic. The replay tests exercise the operation and adapters through an in-memory HTTP transport; they do not exercise process-level exit status or real networking.
