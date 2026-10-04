# Trial report

Implemented `Submit` and focused client-boundary tests in `source/client.go` and `source/client_test.go`. The public function type, `go.mod`, and Go 1.22.0 minimum are unchanged. No staging, commits, global configuration changes, external service calls, or agents were used.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t06/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t06/catalog/go-test-isolation/SKILL.md`

Also opened their `references/http-clients.md`, `references/behavior-observations.md`, and `references/isolation-patterns.md`. Inspected the supplied Go 1.22.12 primary `net/http` source in `client.go`, `transport.go`, and `request.go` for redirect ownership, native replay, and timeout/body error behavior. Package-boundary guidance was unnecessary because package placement and responsibility are unchanged.

## Decisions

- Derive one 500ms operation context, respecting an earlier caller deadline. Reuse it across at most three application attempts and cancellable 10ms/20ms waits.
- Require a nonempty identity before work with deduplication. Snapshot payload bytes once and recreate each request with the supplied identity. Without deduplication, omit the header, attempt once, and retain ambiguous transport errors.
- Accept exactly HTTP 200. Retry only eligible transport/body-read failures or 503 with deduplication. Other statuses, oversized success, redirect-policy failures, and observed caller stopping are terminal. A completely received successful result survives late cancellation.
- Copy the borrowed client locally. Keep its transport and cookie jar, preserve configured redirect rejection errors, and stop redirects to enforce terminal statuses. Record status at the transport boundary because malformed redirect locations otherwise hide it in `Client.Do`.
- Preserve the supplied client timeout through per-attempt child contexts. Go 1.22's native client timeout wrapper replaces body-read errors; moving the same timeout into the request scope retains underlying causes for `errors.Is`. The caller's client fields remain unchanged.
- Read at most 4096 diagnostic bytes or 4097 success bytes, using the extra success byte solely to reject overflow. Close every owned response before returning or retrying. Do-error redirect bodies already closed by `net/http` are not closed twice.
- Keep independent failures when cancellation coincides with them. No breaker, public abstraction, new dependency, or client ownership transfer was added.

## Checks actually run

All Go checks used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`, through `rtk proxy env`. Go 1.22 checks used `/private/tmp/go-client-campaign/tools/go/bin/go`; its test commands used `-ldflags=-linkmode=external`. Local-listener test runs used the authorized escalation.

- Go 1.22.12 initial `test -count=1 ./...`: passed.
- Initial Go 1.22 race/shuffle run exposed native client-timeout body error masking. Corrected it and added an independent body-failure regression.
- Final Go 1.22.12 `test -ldflags=-linkmode=external -race -shuffle=on -count=3 -timeout=20s ./...`: passed.
- Host `go version`: confirmed Go 1.26.5 on darwin/arm64.
- Final Go 1.26.5 `test -count=1 -timeout=20s ./...`: passed.
- Final Go 1.22.12 `vet ./...`: passed.
- Focused Go 1.22.12 `test -ldflags=-linkmode=external -run '^(TestSubmitMalformedRedirectIsTerminal|TestSubmitClientTimeoutPreservesBodyFailure)$' -count=1 -timeout=10s ./...`: passed.
- `rtk proxy /private/tmp/go-client-campaign/tools/go/bin/gofmt -l client.go client_test.go`: no output.
- A temporary one-byte overflow mutation made `TestSubmitBodyBounds/oversized_success` fail with a concrete 4097-byte success. Restored the correct boundary before final verification.

Tests observe exact status/header/body behavior, attempt limits, bounded reads and close counts, independent error identity, stopped entry, shared absolute deadlines, earlier caller deadlines, completed results during cancellation, supplied cookie/redirect/timeout configuration, malformed redirects, and real HTTP body-read cancellation. A real local HTTP fixture loses one acknowledgement after accepting an effect, then returns 503 and success; a second caller invocation with the same identity/body still observes exactly one accepted effect.

## Limits

The three-attempt limit counts application `Do` calls. Native transport replay can add wire requests under the established idempotency contract; native reconnection replay was inspected in Go source but not forced in a network test. No production server, external TLS/DNS/proxy, or real downstream consumer was exercised. Arbitrary custom transports/readers/closers must cooperate with cancellation; a context cannot forcibly interrupt them. Response-body close errors retain the original API's ignored-close-error behavior. Response-content limits do not cap the text of independent wrapped transport errors. Go 1.22 race linking emits a macOS `LC_DYSYMTAB` warning, but compilation and tests complete successfully.
