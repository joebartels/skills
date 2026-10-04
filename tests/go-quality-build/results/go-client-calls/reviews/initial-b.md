# Independent HTTP write replay review

Reviewed the complete final code area at `/private/tmp/go-client-campaign/reviews/initial-b/http-write-replay/source` against its README. This is a library code-area review, so findings would be existing-in-scope; no base/head comparison was supplied. Inspected the complete implementation, tests, `go.mod`, controller `checks.json`, and `controller_probe_test.go.txt`. No other trial, author report, campaign outcome, or proposed skill was inspected.

No actionable defect was confirmed. Grades: Correctness & Compatibility **A+**; Observability & Resilience **A+**; Testing **A+**. Each grade rests on two independent verified safeguards described below, not on the number of tests or passing controller probes.

## Correctness & Compatibility — A+

Scope: The final `Submit` library implementation and its documented contract; `go.mod` declares Go 1.22.0. Executed with Go 1.22.12 on darwin/arm64.

Coverage: Traced ordinary and empty success, pre-work identity/cancellation rejection, exact POST body/header replay, attempt limits, transient transport and 503 failures, terminal other statuses, redirects and policy errors, total/earlier deadlines, response read failures and size boundaries, body ownership, underlying error identity, supplied timeout/cookie jar/transport ownership, and per-invocation state. No release-history or unrelated API review.

Rationale: Zero counted issues. Two independent safeguards qualify for A+: (1) replay is gated by an established deduplication contract, uses one caller identity and immutable payload, and was verified with committed writes whose replies are lost, including an outer retry; (2) bounded response reading rejects oversized successful results and caps diagnostic bodies while closing bodies, verified at the exact limit, beyond it, and on read failures. These control duplicate effects and unbounded/invalid response acceptance respectively. The shared deadline provides an additional verified safeguard.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `client.go:24`, `:29`, `:43`, `:52`, `:56` preserve caller identity/body and permit only one application attempt without deduplication. `client_test.go:423` verifies three lost replies followed by an outer retry, exactly one simulated logical effect, identical bodies/keys, and one request with ambiguous failure without deduplication. The separate controller probes corroborate this behavior.
- [C-G2] `client.go:102`–`:123` owns response closure and uses a 4097-byte sentinel to reject successful bodies larger than 4096; failure diagnostics retain at most 4096 body bytes. `client_test.go:196` checks exact success bounds, over-limit rejection, diagnostic size, read causes, and closure before the next attempt. Reviewer checks also pass for partial bytes returned together with an error and empty success.
- [C-G3] `client.go:27`, `:49`, `:64`, `:70` share a single 500ms-or-earlier context across attempts and cancel the backoff. `client_test.go:116` checks that deadlines do not restart; `:369` exercises a real server that returns 503 then stalls the successful response body, at both total and earlier caller deadlines.
- [C-G4] `client.go:33`, `:82`, `:92`, `:104`, `:134` retain borrowed transport/jar/timeout ownership and preserve errors that Go's client timeout may replace. Redirect handling stops at the initial non-200 status, calls supplied policy, and preserves policy errors. The tests cover accepted/default/rejected redirect policy, malformed locations, cookie updates, and body/transport timeout causes. The supplied function-value assertion at `client_test.go:19` verifies the requested public signature.

Bad

- None found.

Suggested changes

- None needed.

Limits: No real remote server was supplied; deduplication is an explicit server-contract assumption, exercised with a realistic local implementation. Context bounds rely on the supplied transport and callbacks honoring their normal HTTP/context contracts; arbitrary blocking implementations cannot be preempted by this library. No simultaneous success/cancellation ordering was promised beyond ordinary context semantics. Go 1.22.12, rather than the literal 1.22.0 patch, was executed; inspected APIs are available at the declared minimum. Other platforms were not executed.

## Observability & Resilience — A+

Scope: The same outbound library operation, including retry safety, cancellation, resource containment, and caller-visible failures.

Coverage: Assessed attempt ownership, safe replay after ambiguous writes, deadline propagation through response reads, cancellable waits, bounded failure diagnostics, preservation of transport/context causes, terminal failures, and borrowed client lifecycle. Service logs, metric exporters, distributed telemetry, process probes, and shutdown are not owned by this code area.

Rationale: Zero counted issues. Two independent verified safeguards qualify for A+: (1) server-contract-gated replay contains duplicate-write risk after reply loss; (2) one invocation deadline, propagated through actual network body reads and backoff, contains slow-dependency occupancy without resetting the budget. The body-size bound independently limits response/diagnostic growth. Returning useful errors is appropriate at this library boundary; a logger, breaker, or service lifecycle machinery is unnecessary here.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [R-G1] The replay guard and ambiguous-outcome error at `client.go:43`, `:67`, `:100` respect the established idempotency contract and do not claim that a lost unprotected reply means no effect occurred. The committed-write test verifies the material consequence, not merely a retry count.
- [R-G2] The shared context and cancellable timer at `client.go:27`, `:49`, `:64`, `:70` bound all application attempts and body I/O. Real stalled-body tests and the backoff cancellation test pass; the default retries remain capped at three.
- [R-G3] `client.go:97`, `:111`, `:118` return status-specific or explicitly ambiguous failures. Cause joins preserve `errors.Is` for transport, cancellation, caller causes, and timeout-masked body errors. The reviewer test confirms an independent transport cause also survives the client's header-timeout replacement.
- [R-G4] Response limits and closure contain dependency response growth; every normal retry starts after the prior response body has closed. No transport, cookie jar, or connection pool is closed or replaced on the borrowed client itself.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same dependency/context assumptions and platform limits as Correctness. No load or contention evidence justified jitter, admission controls, retry guidance handling, or a breaker beyond the explicitly requested small cancellable backoff. Controller and reviewer checks establish exercised failure handling, not guarantees about an external server's implementation.

## Testing — A+

Scope: All tests supplied with the final module and the controller's two preserved probes. Reviewer-added tests are supplementary evidence, not credited as authored regression coverage.

Coverage: Mapped the complete README contract to assertions. Inspected doubles, local HTTP boundaries, test cleanup, handler synchronization, body/cancellation ownership, error identity, deadline assertions, and source compatibility. Reviewed supplied ordinary/race/vet/Go 1.22 results and independently executed the suite with additional boundary checks, Go 1.22 vet, and race detection.

Rationale: Zero counted issues. Two independent verified safeguards qualify for A+: (1) the real HTTP lost-reply test checks committed effects and outer retry identity, exposing unsafe mutation replay that a simple mock retry count would miss; (2) the bounded-reader/closure tests verify actual bytes consumed and body closure, detecting both oversized response acceptance and resource leaks. The real stalled-body deadline test adds a third distinct safeguard against budgets that cover headers but omit response bodies. A disposable negative control verifies that the original suite detects removal of the transport-cause join.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `client_test.go:423` and the controller probes use actual socket reply loss after the server receives the POST, assert the effect count, and exercise the second caller invocation. The author test additionally asserts ambiguous unprotected outcomes and exact body/key preservation.
- [T-G2] `client_test.go:196` checks exact and oversized body boundaries, consumed bytes, error identity, attempt policy, and exactly one body close. `client_test.go:90` checks closure before another attempt; these assertions do not merely mirror a helper's return values.
- [T-G3] `client_test.go:369` exercises a response-body stall after a preceding failed attempt, at both caller and invocation budgets. Server cancellation is observable; mutexes protect handler state and cleanup closes the servers. The timing allowance accommodates host scheduling while error/deadline/count assertions provide the principal signal.
- [T-G4] Existing tests detected a negative control that removed only the join at original `client.go:92`–`:94`: `TestSubmitRetainsClientTimeout` failed on Go 1.22.12 because `errors.Is(context.DeadlineExceeded)` was lost. The supplementary reviewer test failed as expected for the independent transport cause as well. The unmodified source passes both.

Bad

- None found.

Suggested changes

- None needed. Optional: retain the supplementary partial-reader and independent transport-cause timeout cases to make those boundary expectations explicit; the current tests already detect removal of the relevant transport preservation guard.

Limits: No exhaustive scheduling, repeated-load, HTTP/2-specific, legacy-transport, or platform matrix was executed. The clean race result covers exercised paths; separate concurrent Submit calls were traced to local invocation state rather than claimed proven by the detector. The actual remote server's idempotency implementation is outside this module's tests. No fuzz target is necessary for the supplied request/retry contract.

## Verification evidence

Supplied `checks.json` records eight successful ordinary/vet/race/minimum-toolchain and probe runs; these are controller-executed facts. The preserved probes were read, not modified. Reviewer commands ran in an isolated copy with `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`:

- `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -count=1 -timeout=20s -v ./...`: exit 0, including all original and supplementary boundary tests.
- `/private/tmp/go-client-campaign/tools/go/bin/go vet ./...`: exit 0, no output.
- `/private/tmp/go-client-campaign/tools/go/bin/go test -race -ldflags=-linkmode=external -count=1 -timeout=20s ./...`: exit 0; the external linker emitted an LC_DYSYMTAB warning, and the test/race run passed.
- Negative-control runs exited 1 as expected: the original suite caught the removed transport-error join, and the reviewer timeout-cause test also caught it.

All shell commands were prefixed with `rtk`. The first sandboxed full test run could not bind the local `httptest` port; its raw failure is preserved, and the checks were rerun with approved local loopback access. This is an environment limitation, not a module defect. Source, mutation description, command/environment/exit records, and raw output are preserved in `initial-b-reproductions/`. `source-preservation.json` confirms all four original implementation/test/module/contract files remain byte-identical to the reviewed copy.

## Unnecessary machinery and ungraded follow-ups

None requires removal. The transport observer distinguishes received-status/client-policy errors from transport failures and preserves a cause that `http.Client.Timeout` can replace. The body observer serves the same explicit error-identity requirement for streamed reads. Go 1.22's `net/http/client.go` confirms the timeout replacement, redirect-policy body closure, and deprecated cancellation fallback; forwarding `CancelRequest` preserves the optional legacy interface that wrapping could otherwise hide. Two local client copies are narrowly used for policy and per-attempt observation while retaining borrowed fields. No breaker, telemetry framework, or shared retry state was added.
