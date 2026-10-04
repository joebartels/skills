# Initial-a independent review

Reviewed the complete final code areas in `/private/tmp/go-client-campaign/reviews/initial-a/{http-retry-budget,grpc-policy,pure-control}/source`, with each `README.md` as the authoritative contract. These are code-area reviews, so existing defects would be in scope. Inspected every implementation, test and module file, each controller `checks.json`, and each `controller_probe_test.go.txt`. No other trial, author report, campaign outcome or proposed skill was inspected. No source, Git state or controller probe was changed.

Applied the repository Correctness & Compatibility, Observability & Resilience, and Testing skills and their decision references. Native gRPC behavior was checked against the pinned v1.67.3 source: `dialoptions.go:605–621`, `clientconn.go:742–766`, and `stream.go:634–729` establish default configuration precedence, commitment, retry limits and cancellable native waits.

| Code area | Correctness & Compatibility | Observability & Resilience | Testing |
| --- | --- | --- | --- |
| http-retry-budget | A | A | A+ |
| grpc-policy | A | A | A+ |
| pure-control | A | Not applicable | A |

No actionable finding was substantiated. Production grades recognize correct, bounded implementation; no extra machinery is needed. Network Testing A+ grades recognize independent regression safeguards described below, not test count or passing probes alone.

## http-retry-budget

### Correctness & Compatibility — A

Scope: Entire HTTP library code area, preserved `Fetch(context.Context, *http.Client, string) ([]byte, error)` API and Go 1.22.0 minimum.
Coverage: Exact-200 success, terminal statuses and transport errors, attempt count, Retry-After parsing and budget fit, cancellation/error identity, body bounds/closure, redirects, and caller client ownership.
Rationale: No actionable issue; the implementation meets the inspected contract with relevant behavior verified by focused tests and a real HTTP boundary.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [H-G1] `source/client.go:19–35` creates one context for all attempts and body work; `TestFetchSharedDeadline` verifies equal attempt deadlines and the earlier parent deadline, while `TestFetchNetworkBodyDeadline` confirms a real stalled body observes cancellation.
- [H-G2] `source/client.go:35–49` reads at most 4097 bytes, rejects oversized success, truncates rejection diagnostics to 4096 body bytes, and closes owned bodies. `TestFetchBodyLimit` checks actual bytes read and closure counts, including the exact limit.
- [H-G3] `source/client.go:53–91` restricts application retries to 429/503 and three total attempts; overlarge integer hints saturate to an unfit delay. Tests cover status eligibility, exhaustion, valid past/future dates, malformed fallback and both duration/integer overflow.
- [H-G4] `source/client.go:108–112` preserves rejection, independent I/O failures, context error and cancellation cause through joining. Tests exercise discoverability through `errors.Is`/`errors.As`; redirect tests confirm caller policy and body ownership.

Bad

- None found.

Suggested changes

- None needed.

Limits: Controller test, vet, race, Go 1.22.12 test/vet, and private probe runs all recorded exit 0. Independently reran the complete authored suite with Go 1.22.12 and readonly modules: exit 0, `ok example.invalid/http-retry-budget 0.893s`. As with the standard HTTP client, cancellation requires a supplied custom transport/body implementation to honor request context; no unsupported promise of forcibly stopping arbitrary blocking user code is inferred. No separate release history was supplied, so API preservation is judged against the contract and compile-time signature assertion.

### Observability & Resilience — A

Scope: The same HTTP library's remote call and application retry boundary.
Coverage: Caller/total deadline, retry ownership and eligibility, server retry guidance, cancellable waits, response resource bounds and caller-visible diagnostics. Service probes, shutdown and telemetry exporters do not apply to this library.
Rationale: No actionable issue. Existing controls contain relevant failure modes without a breaker, wrapper or extra retry owner; useful returned errors are adequate for this owned library boundary.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [H-G1] A shared 250ms/earlier-parent context reaches requests and body reads; real network tests verify cancellation. A new per-attempt budget cannot multiply total time.
- [H-G3] The three-attempt cap and narrow eligible status set prevent application retry multiplication. An unfit Retry-After returns promptly, and the cancellable timer prevents waiting through caller cancellation.
- [H-G2, H-G4] Bounded response consumption/diagnostics and retained rejection/failure identities make failure understandable while containing response memory use.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same executed and supplied checks as above. No workload/SLO or telemetry integration is supplied; neither is needed to assess these explicit library controls. No broad deployment or load-test claim is made.

### Testing — A+

Scope: Complete authored HTTP suite plus inspected controller probe source and raw outcomes.
Coverage: Success, rejected statuses, attempt counts, limits, read/close/transport failures, cancellation causes, Retry-After classes, shared deadlines, real body cancellation, redirect ownership and caller settings.
Rationale: No actionable test gap. Two independent verified safeguards beyond routine happy-path setup are (1) an observed-body fixture that checks the actual 4097-byte read ceiling and exactly-once closure, controlling oversized response/resource regressions, and (2) real HTTP stalled-body tests plus equal-deadline assertions across attempts, controlling loss or renewal of the total budget. The complete authored suite passes independently on the minimum-version toolchain. Disposable mutations independently confirmed the assertions fail when the read limit becomes 8192 or the budget becomes 2s; these are deliberate mutants, not defects in the reviewed source.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [H-TG1] `source/client_test.go:104–142` measures byte consumption, returned result/error and closure instead of merely checking the length of a returned slice.
- [H-TG2] `source/client_test.go:346–429` tests both shared request deadlines and a real network body that stays open until request cancellation; cleanup has an explicit release path.
- [H-TG3] Error identity, previous-body closure before the next attempt, client reuse and redirect-error ownership have concrete assertions. Timing waits that coordinate asynchronous completion use bounded observable signals.

Bad

- None found.

Suggested changes

- None needed.

Limits: Independent full-suite run excludes the controller probe file, which remains supplied evidence. The fitting HTTP-date path is covered by a deterministic parser test and traced use of the same timer as the tested fallback delay; it is not separately exercised as a real-network date wait. This does not leave the delay calculation or wait implementation unassessed. Race results are controller-supplied, not independently repeated.

## grpc-policy

### Correctness & Compatibility — A

Scope: Entire gRPC library code area; preserved Dial/Probe signatures, Go 1.22.0 and pinned gRPC v1.67.3/protobuf v1.34.2 dependencies.
Coverage: Method-specific fallback policy, forwarded dial options, valid resolver configuration precedence, native retry eligibility/commitment, one generated application Check, health value handling, wrapped status classification, total/caller deadlines and borrowed connection ownership.
Rationale: No actionable issue. Native policy configuration and a single generated call keep behavior consistent with the pinned dependency contract; dedicated tests verify materially different resolver and commitment behavior.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G-G1] `source/client.go:12–28` installs the specified policy as a default and forwards host options. `TestDialFallbackPolicy` checks every policy field and absence on Watch/other services; resolver tests demonstrate two attempts and an empty resolver policy overriding the fallback.
- [G-G2] `source/client.go:35–37` calls Check once and wraps with `%w`. `TestProbeNativeRetries` observes three native executions but one interceptor/application invocation, one execution for a permanent error and one after response headers commit the call; `status.Code` survives wrapping.
- [G-G3] `source/client.go:33–42` narrows the caller context once and accepts only SERVING. Tests exercise all shipped health values, a shared 500ms deadline across retries, an earlier parent deadline, in-flight cancellation and reuse of a borrowed connection without adding retry policy.

Bad

- None found.

Suggested changes

- None needed.

Limits: Controller test, vet, race, Go 1.22.12 test/vet and private probe runs all recorded exit 0. Independently reran the entire authored suite on Go 1.22.12 with readonly modules: exit 0, `ok example.invalid/grpc-policy 0.953s`. Primary pinned dependency source confirms the behavior traced above. A connection supplied to Probe retains its existing host policy; the contract does not authorize replacing or constraining that policy. No independent TLS handshake or release-history review was performed; credential forwarding is established by the unfiltered option path and successful bufconn credential configuration.

### Observability & Resilience — A

Scope: The same gRPC library's health RPC/retry boundary.
Coverage: Retry ownership, transient classification, native commitment, total/caller budget, failure classification, resolver policy ownership and connection lifecycle. Service lifecycle/telemetry infrastructure is outside this library's contract.
Rationale: No actionable issue. Native retries own jitter/backoff and commitment; the application supplies one total deadline and preserves diagnostically useful status errors without a second retry layer.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G-G1, G-G2] Retry eligibility is only Unavailable under the fallback, with three native attempts and the required backoff configuration; committed failures and permanent statuses are not replayed. Pinned `stream.go` confirms cancellable native retry waits and jitter within policy backoff.
- [G-G3] One context bounds the complete Probe and respects earlier caller termination. Borrowed connections remain available for later calls.
- [G-G2] Error wrapping retains `status.Code` while adding the requested service name for diagnosis.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same supplied and independently executed checks as above. No workload/SLO, service telemetry or deployment configuration was supplied; no findings requiring those unowned facilities are inferred.

### Testing — A+

Scope: Complete authored gRPC suite plus inspected controller probes and raw outcomes.
Coverage: All fallback fields, method scoping, real native execution counts, application call count, permanent status, commitment, resolver override/empty policy, borrowed connections, health states and both total/caller cancellation budgets.
Rationale: No actionable test gap. Two independent verified safeguards beyond routine setup are (1) explicitly sending response headers before Unavailable and asserting one server execution, controlling accidental replay after commitment, and (2) publishing valid resolver configurations and asserting actual two/one execution counts, controlling fallback override of host policy. These exercise native behavior rather than only inspecting JSON. Disposable mutations independently confirmed failure after adding application retries (three committed executions/application calls) and disabling resolver configs (three executions instead of two/one); these are deliberate mutants, not defects in the reviewed source.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G-TG1] `source/policy_test.go:96–158` separates application invocations from native server executions, includes explicit commitment, and checks wrapped classification and request service propagation.
- [G-TG2] `source/policy_test.go:160–205` exercises a manual resolver with a two-attempt policy and an empty policy; both require the fallback to yield.
- [G-TG3] `source/policy_test.go:248–366` observes deadlines and cancellation across the actual bufconn RPC boundary. Helpers wait for connection readiness and own server, listener and connection cleanup.

Bad

- None found.

Suggested changes

- None needed.

Limits: Independently executed authored suite excludes the controller probe file; probe and race outcomes are supplied controller evidence. No real external gRPC service or TLS boundary is needed for the stated native policy/health contract, and none was exercised.

## pure-control

### Correctness & Compatibility — A

Scope: Entire deterministic local SumPositive library code area and Go 1.22.0 module.
Coverage: Positive/nonpositive filtering, accumulation, nil/empty iteration, preserved signature and read-only input traversal.
Rationale: No actionable issue. `source/client.go:3–10` adds exactly values greater than zero; both focused tests confirm observable results, and Go 1.22.12 independently builds/runs them.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [P-G1] The single `v > 0` predicate enforces the requested selection; the mixed positive/zero/negative regression returns 5, while the existing all-positive case returns 6.
- [P-G2] The function retains its signature, traverses without mutation, and naturally returns zero for nil/empty input. No context, retry, external call or lifecycle state was introduced.

Bad

- None found.

Suggested changes

- None needed.

Limits: Controller test/vet, minimum-version test/vet and private probe results record exit 0. Independent Go 1.22.12 full-suite run: exit 0, `ok example.invalid/pure-control 0.143s`. No special integer overflow policy is specified; ordinary Go int arithmetic remains the existing API behavior.

### Observability & Resilience — Not applicable

Scope: Deterministic, local SumPositive calculation.
Coverage: Inspected complete implementation and tests for blocking work, external calls, owned background work and operational signals; none exist.
Rationale: No relevant resilience, external failure or lifecycle decision is implicated. Adding timers, cancellation, metrics or a breaker would not address this contract.
Limits: No broader caller/service context is supplied or required for this calculation review.

### Testing — A

Scope: Complete SumPositive tests plus controller probe source/results.
Coverage: Existing positive accumulation and the requested focused mixed-value regression.
Rationale: No actionable gap for the small, deterministic contract. The mixed-value test fails if negative inputs contribute, and the existing positive case protects ordinary accumulation. One focused regression is appropriate; no additional framework or concurrency verification is needed.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [P-TG1] `source/client_test.go:11–15` includes two negative values, zero and two distinct positives and asserts the exact expected sum. Setup and shared state are absent.
- [P-TG2] `source/client_test.go:5–9` retains the ordinary all-positive assertion; the independent minimum-version run passes both tests.

Bad

- None found.

Suggested changes

- None needed.

Limits: Nil/empty and nonpositive-only cases were traced in the simple loop rather than independently added as tests. Supplied private probe repeats the mixed-value property. No important unassessed test behavior was identified.

## Unnecessary machinery and review verification

None found. HTTP helpers isolate parsing, waiting and error composition required by its contract. gRPC delegates native policy behavior without wrappers/interceptors or a second application retry owner. SumPositive stays a single local loop. No service-level observability or circuit breakers are warranted by these code areas.

Independent checks used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`, with `rtk proxy` and `/private/tmp/go-client-campaign/tools/go/bin/go test -mod=readonly -ldflags=-linkmode=external -count=1 -timeout=20s ./...` in each source directory. HTTP loopback execution used escalation. Exact commands/raw results are preserved in `initial-a-reproductions/reviewer-checks.txt`. Four deliberately broken disposable copies and `mutant-checks.txt` preserve independent evidence of test regression signal; no actual defect reproduction was needed.
