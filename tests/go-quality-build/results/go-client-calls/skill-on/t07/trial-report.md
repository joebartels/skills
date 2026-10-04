# Trial report

Implemented `Fetch` in `source/client.go` and added focused tests in `source/fetch_test.go`. Kept the exported function type, `go.mod`, Go 1.22.0 minimum, supplied client, and existing success test unchanged. No dependencies, wrapper, breaker, staged files, or commits were added.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t07/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t07/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t07/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t07/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t07/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t07/catalog/go-test-isolation/SKILL.md`

Also opened these selected-skill references:

- `go-client-calls/references/http-clients.md`
- `go-behavior-tests/references/behavior-observations.md`
- `go-test-isolation/references/isolation-patterns.md`

## Decisions

- One derived 250ms context covers all application attempts, retry waits, bounded reads, and owned-body closure; an earlier parent deadline wins. The supplied client's timeout, transport, redirect policy, and ownership stay intact.
- Accept only final HTTP 200 under the supplied client's redirect policy. Application retries are limited to 429/503 and three calls to `Do`; transparent transport retries and redirects are outside that application count. Transport, body-read, and close failures stop the loop.
- Parse digit-only delta-seconds with checked duration conversion and HTTP-date with `http.ParseTime`. Valid unfit or overflowing delays return the last rejection immediately; past dates allow immediate retry; missing/malformed hints use a cancellable 10ms wait.
- Read no more than 4097 bytes. Reject oversized successes and truncate status diagnostics. Every successful `Do` transfers body closure to `Fetch`; `Do` error responses are not closed again.
- Private errors bound displayed text to 4096 bytes while exposing original transport/read/close/cancellation causes through multi-error unwrapping. Cancellation while retrying retains the last HTTP rejection and standard/custom cancellation causes. Completed success remains successful if cancellation arrives during final cleanup.

## Checks actually run

All Go commands used `GOWORK=off`, `GOTOOLCHAIN=local`, `GOPATH=/private/tmp/go-client-campaign/gopath`, `GOMODCACHE=/private/tmp/go-client-campaign/modcache`, `GOCACHE=/private/tmp/go-client-campaign/cache`, and `GOPROXY=off`. Tests used `/private/tmp/go-client-campaign/tools/go/bin/go` and `-ldflags=-linkmode=external`.

- Confirmed the test toolchain is Go 1.22.12, darwin/arm64.
- Ran selected new tests against the original implementation: meaningful failures established missing retries, deadline behavior, size limits, and error retention/bounds.
- Ran focused client-boundary tests after implementation: passed.
- Ran the complete suite, including local `httptest` listeners with authorized escalation: passed.
- Ran the final complete suite with `-race -count=5 -shuffle=on`: passed. A macOS external-linker `LC_DYSYMTAB` warning was emitted; the build and race checks succeeded.
- Ran the duration-overflow leaf independently with `-count=10`: passed.
- Ran the future HTTP-date case independently with `-v`: passed, without skipping.
- Ran `go vet ./...` against final source: passed.
- Ran `gofmt -l client.go client_test.go fetch_test.go`: no output.

## Limits

Tests use handwritten transports/bodies for precise policy, error, byte-count, deadline, and ownership observations, plus real local HTTP for transport/body cancellation. External DNS, TLS, remote services, and pathological non-cooperative custom readers/closers were not tested; contexts cannot forcibly interrupt arbitrary implementations. The future-date test uses real time and may skip if scheduling misses its second-resolution window; its independently recorded final run executed successfully. No unsupported inputs or other trials were inspected.
