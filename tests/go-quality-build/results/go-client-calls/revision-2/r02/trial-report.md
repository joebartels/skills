# Trial r02

Implemented `Fetch` and focused tests in `source/client.go` and `source/client_test.go`. The public function type and `go 1.22.0` module declaration remain unchanged. No dependency, client wrapper, breaker, global setting, staging, or commit was added.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/r02/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r02/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r02/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r02/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r02/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r02/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r02/catalog/go-test-isolation/SKILL.md`

Also opened their `go-client-calls/references/http-clients.md`, `go-behavior-tests/references/behavior-observations.md`, and `go-test-isolation/references/isolation-patterns.md`. Read only the task prompt, README, original code/tests/module, these catalog files, and primary Go 1.22 HTTP client/transport source.

## Decisions

- Borrow the supplied client unchanged, retaining redirect and timeout policy. Count application `Client.Do` calls; redirects and transparent transport behavior remain the client's responsibility.
- Use one 250ms child context, preserving an earlier parent deadline. Keep it alive through reading and closing each owned body. Reject observed cancellation before another request.
- Retry only completed 429/503 rejections, at most three attempts. Transport, read, and close failures are terminal. Preserve the last rejection when cancellation/deadline stops waiting or subsequent work.
- Accept only final HTTP 200 under the supplied redirect policy. Read at most 4097 bytes, reject oversized success, cap diagnostic text at 4096 bytes, and retain returned failures through error unwrapping. Include status text before body diagnostics.
- Parse digit-only Retry-After seconds with checked conversion. Numeric overflow is an unfit delay. Parse HTTP dates, allow immediate retry for past dates, use cancellable 10ms fallback for malformed hints, and return promptly when a delay cannot fit.
- Close owned bodies once before retrying, including rejection/read failure. Do not close a redirect-policy error's response again because `Client.Do` already owns that closure.

## Checks actually run

All Go build/test/vet commands used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`. Go 1.22 tests additionally used `-ldflags=-linkmode=external`. Local HTTP listeners ran with the authorized exec escalation. No test binary was written into source.

- Verified toolchains: Go 1.22.12 and Go 1.26.5, darwin/arm64.
- Pre-implementation focused Go 1.22 tests failed meaningfully: retryable statuses were not retried, oversized success returned 8000 bytes, and a total request deadline was absent.
- Initial implemented suite passed. A later supplied-client timeout assertion failed because Go 1.22's timeout body error does not support `errors.Is(context.DeadlineExceeded)`; primary standard-library source confirmed that behavior, and the test was corrected to check `net.Error.Timeout`.
- Final Go 1.22 full suite: `go test -ldflags=-linkmode=external -timeout=15s ./...` — passed.
- Final Go 1.26 full suite: `go test -timeout=15s ./...` — passed.
- Go 1.22 focused selection: `go test -ldflags=-linkmode=external -run 'TestFutureRetryDate/cancel_during_wait|TestBodyBounds/oversized_success|TestRetryAfterFallbackAndPastDate/past_date' -count=1 -timeout=10s ./...` — passed.
- Go 1.22 race/shuffle repetition: `go test -ldflags=-linkmode=external -race -shuffle=on -count=3 -timeout=30s ./...` — passed; after the additional close-failure test, the same final suite with `-count=2` also passed.
- Go 1.22 `go vet ./...` — passed twice, including final source.
- `gofmt -l client.go client_test.go` — empty output.

Tests cover exact success/rejection classification, attempt cap, closure before retry, result/read/diagnostic bounds, independent read/transport/close errors, malformed/past/future/overflowing retry hints, cancellation at entry and during a retry wait, custom cancellation causes, shared attempt/body budgets, earlier parent deadline, supplied redirect policy, supplied client timeout, and real HTTP body cancellation.

## Limits

- Cancellation depends on the supplied transport/body honoring its request cancellation; Fetch does not spawn goroutines to escape uncooperative arbitrary readers or closers.
- Go 1.22's native `http.Client.Timeout` can replace an underlying failure with an opaque timeout error before returning it. Fetch preserves errors the supplied client exposes; recovering causes that the client already discarded would require the forbidden client/transport wrapper or changing its settings. The supplied timeout behavior is tested through a real local server.
- The fitting HTTP-date test aligns with whole-second date resolution and may skip if severe scheduling delay misses the fitting window. Time checks allow late scheduling where appropriate. External DNS/TLS/services and other platforms are not tested.
- The macOS external linker emitted an LC_DYSYMTAB warning during successful race builds.
- The initial prompt read used plain `cat` before reading its shell-prefix requirement. Subsequent shell commands used `rtk`. Two attempted `rtk read` line-range commands were unsupported and failed without reading files; the primary-source reads succeeded using `rtk proxy sed`.
