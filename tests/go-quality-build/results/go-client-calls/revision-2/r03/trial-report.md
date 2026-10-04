# Trial report

Implemented `New` and `(*Client).Fetch` in `source/client.go` and added focused contract, ownership, deadline, counting, scope, and recovery tests in `source/client_test.go`. Public function types, Go 1.22.0, and pinned dependencies are unchanged. No staging or commits were performed.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/r03/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r03/catalog/go-test-isolation/SKILL.md`

Also opened their `references/http-clients.md`, `references/behavior-observations.md`, and `references/isolation-patterns.md`. Inspected the pinned gobreaker v2.4.0 implementation (`gobreaker.go`, `twostep_gobreaker.go`, `counter.go`) and Go 1.22.12 net/http timeout implementation as primary dependency sources. No other trial, outcome, repository instruction, or private probe was read.

## Decisions

- Retain the supplied nonnil HTTP client itself, including transport, redirect policy, and effective timeout. A nil client receives a new default client. Never mutate borrowed configuration or close its transport.
- Use one application `Do` invocation without retries. Apply a single 250ms context budget from Fetch entry through body consumption and closure, preserving earlier parent deadlines. Supplied redirect/native transport policies retain their existing behavior.
- Accept exact final HTTP 200 under the supplied redirect policy. Other final statuses return `HTTP <status>` without incorporating response-body diagnostics; close every owned response body. Preserve net/http ownership on `Do` errors, including already-closed redirect responses.
- Read at most 4097 bytes to detect a result exceeding 4096; reject oversize results. Bound error presentation to 4096 bytes while retaining an unwrap path to long request, transport, or read errors. Preserve bounded partial data on read failure.
- Allocate breakers per Client and normalized endpoint scheme/host authority, sharing paths and queries within a dependency. Protect map construction with a mutex. Use the pinned TwoStepCircuitBreaker rather than a custom state machine, with one recovery admission, a three-consecutive-failure threshold, and a 100ms open timeout.
- Report dependency health separately from the caller result. Eligible 503 outcomes fail health accounting; completed 200 results succeed; every other error and a completion during context cancellation is excluded. Exclusions preserve closed-state failure history and release half-open admission without proving recovery. Delegate generation fencing to gobreaker.

## Checks actually run

All Go commands used `rtk proxy env` with:

```text
GOWORK=off GOTOOLCHAIN=local
GOPATH=/private/tmp/go-client-campaign/gopath
GOMODCACHE=/private/tmp/go-client-campaign/modcache
GOCACHE=/private/tmp/go-client-campaign/cache
GOPROXY=off
```

The minimum-version executable was `/private/tmp/go-client-campaign/tools/go/bin/go` (confirmed Go 1.22.12, darwin/arm64). The host `go` was confirmed Go 1.26.5, darwin/arm64. Go 1.22 test commands used the host-required `-ldflags=-linkmode=external`.

- Before implementation, Go 1.22 `test -run TestFetchStatusBoundsAndOwnership ./...` failed meaningfully on absent deadlines, oversized accepted results, and unbounded diagnostics; `test -run TestBreakerScopeAndHistory ./...` failed because the tripped dependency remained admitted.
- Go 1.22 `test ./...` passed after the initial implementation.
- The first Go 1.22 `test -race -count=5 -shuffle=on ./...` failed on the added test's assumption that a native Client.Timeout body-read error matched `context.DeadlineExceeded`. Confirmed Go 1.22's `httpError` exposes `net.Error.Timeout` rather than that identity, and corrected the test without changing client behavior.
- Final Go 1.22 `test -race -count=5 -shuffle=on ./...` passed.
- Final Go 1.22 `test -count=1 ./...` passed after all source edits.
- Final Go 1.22 `test -run '^TestExcludedRecoveryReleasesAdmission/canceled_200$' ./...` passed independently.
- Go 1.22 `vet ./...` passed, including the final source.
- Host Go 1.26.5 `test ./...` passed, including the final implementation and corrected assertions.
- Ran Go 1.22 `gofmt` on changed Go files and `rtk git diff --check`, which passed. Checked status and diff statistics; only `client.go` and `client_test.go` are changed in source.

## Limits

- Local HTTP tests verify real body-read cancellation, earlier parent/client timeouts, and supplied redirect rejection. Synthetic RoundTrippers and instrumented bodies provide controlled health outcomes, ownership counts, canceled completed responses, and stale-generation schedules; they do not establish external DNS, TLS, proxy, or production service behavior.
- The request ceiling counts application `Do` calls. Borrowed redirect policies and standard-library transparent retries can produce additional wire exchanges.
- Context deadlines require cooperation from the supplied transport/body; Fetch does not launch an unjoinable goroutine to force interruption of arbitrary borrowed I/O.
- Recovery timing tests use a 110ms timer after observed trip completion, event-controlled in-flight probes, bounded joins, and explicit confirmation that held old requests had not expired. Five shuffled race runs exercise these schedules without proving all possible schedules.
- The Go 1.22 race linker emitted a macOS `LC_DYSYMTAB` warning; linking and the final test run still succeeded.
