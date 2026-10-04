# Trial report

Implemented `New` and `Fetch` in `source/client.go`, with focused tests in `source/contracts_test.go` and consumer signature checks in `source/api_test.go`. Public signatures, Go 1.22.0 declaration, and pinned dependencies remain unchanged. No staging, commits, or external actions were performed.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t04/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t04/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t04/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t04/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t04/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t04/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t04/catalog/go-test-isolation/SKILL.md`

Also opened the behavior-observations and isolation-patterns references from those selected skills, and the pinned gobreaker primary source files `gobreaker.go`, `twostep_gobreaker.go`, and `counter.go`. Package-boundary work was unnecessary: responsibility and package placement remain unchanged.

## Decisions

- Keep the supplied nonnil HTTP client borrowed. Copy its configuration for each request and override redirect handling on the copy, preserving the caller's client, transport, timeout, and redirect policy.
- Use one caller-derived 250ms context through request, body reading, and owned body closure. Earlier parent deadlines win. Already canceled calls do not start transport work.
- Accept exact HTTP 200 and return at most 4096 bytes. Other statuses return `HTTP <status>` without incorporating an unbounded response body into diagnostics. Preserve partial bytes and the original read error on body-read failure.
- Use the pinned gobreaker per normalized scheme/host authority, with a synchronized per-client registry, three consecutive eligible 503 outcomes, 100ms open duration, and one recovery request.
- Mark only HTTP 503 failures completed without caller cancellation as health eligible. Exclude other errors, preserving consecutive history and releasing recovery admission. Let gobreaker handle generations, excluded counters, and state transitions.
- Avoid retry loops, background work, global configuration changes, and a custom breaker framework.

## Checks actually run

Every Go check used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`, through `rtk proxy env`. Go 1.22 test commands additionally used `-ldflags=-linkmode=external` for this macOS host.

- Pre-change Go 1.22 focused regression run: `go test -run 'TestFetchStatusAndBodyLimit|TestBreakerExcludedOutcomesPreserveFailureHistory' ./...`. It failed at meaningful assertions for the 5000-byte unbounded result and missing breaker rejection after three failures.
- Post-change Go 1.22 full suite: `go test ./...` — passed.
- Final Go 1.22 full suite with concurrency reliability checks: `go test -race -count=3 -shuffle=on ./...` — passed. The host linker emitted a nonfatal LC_DYSYMTAB warning.
- Go 1.22 static analysis: `go vet ./...` — passed.
- Host toolchain full suite: `go test ./...` — passed.
- Go 1.22 focused selection: `go test -run 'TestBreakerRecoveryAdmissionAndExcludedCompletion/canceled|TestBreakerCanceled503DoesNotCountOrReset|TestFetchReadFailureOwnsBodyAndContext' ./...` — passed.
- `go version` confirmed Go 1.22.12 and host Go 1.26.5, both darwin/arm64.
- Go 1.22 `gofmt -l client.go contracts_test.go api_test.go` — no unformatted files.

Tests verify exact statuses, size/closure/read errors, preserved request context and cancellation ownership, real local HTTP redirects and header/body deadlines, dependency and client scope, success reset, disabled behavior, excluded failure history, one concurrent recovery admission, excluded recovery release, reopening, successful recovery, and late-generation completion isolation.

## Limits

Verification used local HTTP servers and controlled transports on macOS; no external service or other-platform integration was run. Deadline enforcement depends on transport/body cooperation with the request context, as with `net/http`; arbitrary noncooperating borrowed transports cannot be forcibly interrupted. Signature compatibility was checked through a representative external-package consumer, without downstream repository access. Race/repeat runs cover the exercised schedules, not every possible interleaving.
