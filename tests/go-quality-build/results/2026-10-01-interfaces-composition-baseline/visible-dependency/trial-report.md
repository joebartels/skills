# Trial report

## Composition decision

`Sender` is an immutable, per-tenant value configured with `NewSender(endpoint, client)`. The host continues to own each `*http.Client`, including authentication transports and timeout policy. Each `Send` builds a request with its own caller context and submits it through that sender's client. There is no environment lookup or mutable process-wide client dependency, so separate sender values can be used concurrently.

## Files changed

- `receipt.go`: added endpoint/client configuration and nil-client error; preserved JSON POST, content type, 2xx handling, synchronous completion, and response-body close.
- `receipt_test.go`: replaced global-client mutation with network-free tests for distinct endpoint/client policy routing during overlapping calls, escaped JSON, cancellation propagation, non-2xx responses, and nil-client handling.
- `README.md`: documented one sender per tenant, concurrent use, distinct host-owned clients, and an example configuring two senders.

## Checks and results

- `gofmt -w receipt.go receipt_test.go` — completed.
- `GOCACHE=/private/tmp/go-quality-build-composition-eval/cache-baseline-luna go test ./...` — passed (`ok example.com/visible-dependency`).
- `GOCACHE=/private/tmp/go-quality-build-composition-eval/cache-baseline-luna go vet ./...` — passed with no diagnostics.
- Initial `go test ./...` with the default Go cache could not write under `/Users/jb/Library/Caches/go-build`; the temporary `/private/tmp` cache resolved it.
- An initial test using `httptest.NewServer` could not bind loopback in this sandbox (`operation not permitted`); the final tests use in-memory `RoundTripper` implementations and make no network requests.

## Limitations

The sample README snippet assumes the host has already created tenant-specific transports and defines `ctx`; it is illustrative rather than a standalone executable program. The tests verify that distinct client policies and endpoints reach the correct in-memory transport, but do not exercise live networking or real authentication transports. The Go runtime available for verification was 1.26.5; source compatibility is intended for Go 1.22 but was not separately tested with that toolchain.
