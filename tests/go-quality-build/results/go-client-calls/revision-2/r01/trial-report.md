# Trial report

Implemented `Submit` and focused tests in `source/client.go` and `source/client_test.go`. The public signature, module and Go 1.22.0 minimum are unchanged. No files were staged or committed.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/r01/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/r01/catalog/go-test-isolation/SKILL.md`

Also opened their `http-clients.md`, `behavior-observations.md` and `isolation-patterns.md` references. Inspected the supplied Go 1.22 standard-library HTTP client, transport and request source for timeout rewriting, redirect ownership, cancellation and native replay behavior. No package-boundary change required that skill.

## Decisions

- One owned operation context supplies a 500ms total budget, narrowed by an earlier caller deadline. All attempts, response reading and 10ms/20ms cancellable backoffs share it.
- Deduplication requires a nonempty caller identity before work. A payload snapshot and that identity survive every application attempt; separate caller invocations use their supplied identity/body. Only 503 and transport/read failures are eligible for replay, with at most three application attempts. Without deduplication there is one application attempt and no idempotency header; ambiguous failures retain their cause.
- HTTP 200 alone succeeds. Redirects are terminal under the status contract. A local client copy invokes a configured redirect callback and preserves its rejection cause without following another target.
- The supplied transport/pool, jar and shorter client timeout are retained. Local client copies narrow longer or absent timeouts to the operation's remaining budget. Request cancellation and the transport decorator forward legacy cancellation facilities. The borrowed client is never changed or closed.
- Success reads use a 4096-byte limit plus one detection byte; oversized results reject. Status diagnostics consume at most 4096 bytes. Every successfully acquired response body closes before return or retry; redirect-policy error bodies remain owned by `Client.Do`.
- Private observation records underlying transport/body causes before `Client.Timeout` can hide them, retaining those causes and context classifications for `errors.Is`. Completed successful reads remain successful despite later caller stopping. No breaker or exported surface was added.

## Checks actually run

All build, test and vet commands used `GOWORK=off`, `GOTOOLCHAIN=local`, `GOPATH=/private/tmp/go-client-campaign/gopath`, `GOMODCACHE=/private/tmp/go-client-campaign/modcache`, `GOCACHE=/private/tmp/go-client-campaign/cache`, and `GOPROXY=off`. Shell commands were run through `rtk`. Local listener checks used the authorized disposable-task escalation.

- Go 1.22.12: `go test -ldflags=-linkmode=external ./...` — passed after fixing timeout-cause/cancellation observations. Earlier development runs exposed missing timeout classifications.
- Go 1.22.12: `go test -race -shuffle=on -count=3 -ldflags=-linkmode=external ./...` — passed. The macOS linker emitted a nonfatal object-symbol warning.
- Go 1.26.5: `go test ./...` — passed.
- Go 1.22.12: `go vet ./...` — passed.
- Go 1.22.12: focused `503_replay` and client-timeout body subtests — passed when selected independently.
- Temporarily restored the original implementation and ran `TestSubmitReplay` — compiled and failed its meaningful assertion on the first HTTP 503 as expected. Restored the implementation immediately, then verified the completed suite.
- `gofmt` and `git diff --check` — clean.

Tests observe real local HTTP identity/body stability, one committed effect across deduplicated replay and caller reuse, lost acknowledgement with and without replay guarantees, terminal redirects, exact success status, attempt limits, response limits and closure, caller cancellation, shared deadlines through body consumption, configured client policy, and original errors under short client timeouts and legacy cancellation.

## Limits

Verification uses local HTTP servers and focused handwritten transport/body fixtures; external DNS, TLS, service deployment and deduplication retention are untested. The supplied server contract establishes replay safety. The ceiling counts application attempts, not every possible native transport execution. Arbitrary transports or bodies that ignore all supported cancellation facilities cannot be forcibly interrupted. Race/repetition checks cover exercised schedules, not every possible schedule. Callers remain responsible for supplying the same identity and exact payload across their own invocations.
