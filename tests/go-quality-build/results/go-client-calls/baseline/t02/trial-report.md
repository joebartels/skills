# Trial report

Implemented the README contract in `source/client.go` and focused tests in `source/client_test.go`. The exported Fetch function type and `go 1.22.0` module directive remain unchanged. No dependencies, wrapper, breaker, client mutation, staging, or commits were added.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t02/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t02/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t02/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t02/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t02/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t02/catalog/go-test-isolation/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t02/catalog/go-concurrency-and-ownership/SKILL.md`

Also opened the selected skills' `go-behavior-tests/references/behavior-observations.md` and `go-test-isolation/references/isolation-patterns.md`. Ownership guidance informed body and asynchronous HTTP-fixture lifetime; production attempts remain serial. Package-boundary work was unnecessary.

## Decisions

- Use the supplied nonnil client directly. Three application attempts mean three calls to its Do method; caller-configured redirect and transport policy remains in force.
- Derive one 250ms context, preserving an earlier caller deadline, and retain it through waiting, bounded reads, and body closure.
- Accept exactly 200. Retry only 429 and 503. Transport, body-read, and body-close failures terminate further application attempts.
- Parse digit-only nonnegative delta-seconds without duration overflow. Oversized integer/duration values become an unfit delay. Parse HTTP dates with the standard library; past dates and zero allow immediate retry. Malformed hints use a cancellable 10ms timer. An unfit delay returns the last rejection immediately.
- Read at most 4097 bytes. Success over 4096 bytes fails; rejection diagnostics include at most 4096 body bytes. Close every body acquired from successful Do, including rejected responses and read failures. A Do redirect error already owns closure of its returned body.
- Retain the last rejection alongside subsequent independent failures, cancellation classification, and caller cancellation cause using errors.Join. Do not turn a completed successful read into failure solely because cancellation arrives late.

## Checks actually run

All Go checks used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`, with rtk-prefixed shell commands. Local HTTP tests ran with the authorized sandbox escalation.

- Before implementation: `go test -run 'TestFetch(StatusAndAttempts|BodyLimit|RetryAfter|CanceledAtEntry)$' -timeout 5s ./...` failed with meaningful assertions against the original implementation for absent retries, unbounded success reads, missing diagnostics, and admission after cancellation.
- Initial implementation: `go test -timeout 10s ./...` passed.
- Final code, actual Go 1.22.12: `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -timeout 10s ./...` passed.
- Final code, Go 1.26.5: `go test -race -shuffle=on -count=3 -timeout 15s ./...` passed.
- A combined focused `-run` expression returned success; explicit single-leaf reruns below establish the intended selection without depending on that expression's slash/alternation behavior.
- Individual verbose checks passed and confirmed execution: `TestFetchRetryAfter/integer_overflow`, `TestFetchBodyFailureDuringCancellation/503`, and `TestFetchRedirectErrorBodyAlreadyClosed`.
- `go version` confirmed host Go 1.26.5 and the minimum-line toolchain Go 1.22.12, both darwin/arm64.
- `gofmt -w client.go client_test.go` ran; final `gofmt -l client.go client_test.go` produced no output.

Tests cover status neighbors, retry exhaustion, body bounds and closure counts, independent read/close failures, terminal transport failure, retained rejection and noncomparable cancellation cause, hint overflow/malformed/past values, exact future-date parsing, cancellable waits, fixed total/parent deadlines, real response-body cancellation, redirect-error closure, caller policy preservation, and client reuse.

## Limits

Verification used Go 1.22.12 rather than the exact 1.22.0 patch release and only macOS arm64. A fitting future HTTP-date delay is checked deterministically at the parser boundary; real waiting checks use the malformed-hint fallback. Real network checks are local HTTP, not external DNS/TLS/services. A supplied custom transport, body reader, or Close implementation must cooperate with cancellation for a wall-clock deadline to interrupt blocking work; no detached I/O goroutines were introduced to simulate forced cancellation. Race/repetition checks exercise the tested schedules, not every possible schedule. No downstream consumers were available beyond the compile-time public function-type assertion.
