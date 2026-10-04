# Independent neutral HTTP pair review

Labels A and B are neutral. No arm disclosure was received. This report and the adjacent raw observations were frozen before any disclosure. Inputs were limited to the two supplied packets, the unchanged Correctness & Compatibility, Observability & Resilience, and Testing review skills and their decision references, and primary Go HTTP source. Candidates and supplied probes were not edited.

This is a code-area review of the complete implementations, not an inferred base/head changeset. All findings are existing-in-scope. Locations below are relative to `/private/tmp/go-client-campaign/reviews/pairs/`; `WR` abbreviates `revision-http-write-replay`, and `RB` abbreviates `revision-http-retry-budget`. Each module declares Go 1.22.0 and uses only the standard library. Shared production root causes retain the same ID across topics; testing gaps have separate IDs because they require independent regression assertions. Counts across topics must not be added.

## Contract interpretations and limits

- **Application attempts:** An attempt is one application `Client.Do` invocation. Redirect hops and safe native transport retransmissions inside that invocation are not additional application attempts. A network handler count alone cannot prove multiplication of application attempts. Both implementations in each pair limit their application loops appropriately.
- **Redirect acceptance is ambiguous in the write README.** Both WR implementations copy the supplied client, call its redirect callback, then force `ErrUseLastResponse` if the callback permits the redirect (`WR/A/source/client.go:33`, `WR/B/source/client.go:28`). They implement an initial-response reading of “other statuses are terminal”; they do not preserve permissive redirect behavior. The README also says to retain client configuration, without explicitly specifying initial versus final response acceptance. The tests establish the implementations' chosen interpretation, not an independent stronger contract. I assessed the explicitly specified retry, identity, body, budget and error requirements; the shared redirect choice earns neither a comparative benefit nor a confirmed defect. If “retain settings” requires permitted redirects to run normally, both need the same correction. If initial responses must be terminal, the restriction is appropriate. This unresolved shared choice is a limit on unconditional compatibility claims.
- **The fetch client boundary is explicit.** RB calls the supplied `client.Do` directly, preserving its native redirect policy and timeout and adding no client/transport wrapper. A redirect-policy error can return a nonnil, already-closed response; its known status can be preserved without wrapping the client. Malformed redirects may return no response and erase the status before Fetch sees it; that is a boundary limitation, not a requirement to add the prohibited wrapper.
- **Preserve available causes, not causes erased before return.** Go 1.22.12 `net/http/client.go:970` replaces a body read error with a timeout error containing its text and no unwrap chain when the client timer wins. Direct-client controls reproduced this erasure. Go 1.22 does not always expose `context.DeadlineExceeded` through that native timeout error; Go 1.26 does. Timer ordering can also leave the original error intact. The supplemental fetch timeout failure therefore is not counted as a candidate defect: restoring an erased arbitrary cause would require interception below the supplied-client boundary and conflict with “no client wrapper.” Both retain the error actually returned by the client, its timeout classification, and parent context causes available at their boundary.
- **Last fetch rejection:** Both preserve the rejection while retry waiting is stopped. A also retains a prior rejection when a later HTTP 200 body fails; B does not (`RB/A/source/client.go:42`, `RB/B/source/client.go:44`). The README specifically promises retention when waiting ends under cancellation/deadline, so I do not silently extend it to every later successful-status body failure. This difference is ungraded unless that broader retention is required.
- **Diagnostics:** WR explicitly bounds diagnostic *bodies*; neither must cap arbitrary transport-error text to 4096 under that wording. RB specifies a 4096-byte result/diagnostic limit. I interpret that as bounding the returned diagnostic text, separately from the preserved error chain. Even a body-only reading would leave A's arbitrary cause text unbounded; B meets either reading.
- Cancellation depends on the supplied transport honoring an available cancellation facility. No implementation can forcibly interrupt an arbitrary blocking RoundTripper/body that honors neither context nor native cancellation. The WR reproduction uses the native legacy `CancelRequest` facility, which B successfully supports; it is not an intentionally uncancellable transport.

## Verification evidence

All four source trees match every supplied `source-hashes.json` entry. Supplied public checks report passing tests, race checks, vet, and Go 1.22 checks for all four implementations. The private checks report all WR checks passing; RB A fails policy status, long diagnostics and timeout-cause observations; RB B fails only the timeout-cause observation. Those observations were assessed individually, not used as a target grade.

Fresh tests ran in disposable copies with unchanged candidate tests plus unchanged frozen and supplemental probes. Host execution used Go 1.26.5; minimum execution used `/private/tmp/go-client-campaign/tools/go/bin/go` (Go 1.22.12). Every scratch Go command used:

```text
rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off <go> test -ldflags=-linkmode=external -count=1 -timeout=20s -v ./...
```

The focused host reproductions added `-run TestNeutral`; the final direct-client control used `-count=3 -run TestNeutralSuppliedClientAlreadyErasesReadCause` on both toolchains. Local test listeners were used only in disposable checks. Exact commands, working directories, exit codes and raw output are archived in [raw-observations.json](pair-revision-http-reproductions/raw-observations.json), alongside the two reproduction source files.

| Pair / label | Candidate tests and frozen probes | Supplemental observations | Additional reproduction |
| --- | --- | --- | --- |
| WR A | Pass on both toolchains | Pass on both | Legacy earlier deadline and total budget fail on both |
| WR B | Pass on both toolchains | Pass on both | Legacy deadlines, total budget, causes and closure pass on both |
| RB A | Pass on both toolchains | Known status and long diagnostic fail; native timeout-cause probe fails | Policy status fails while cause and single closure pass; direct client reproduces erasure |
| RB B | Pass on both toolchains | Known status and bounded diagnostic pass; native timeout-cause probe fails | Policy status, cause and single closure pass; direct client reproduces erasure |

The initial Go 1.22 direct-client control waited on `Request.Context` and once retained its cause because context and client timers race. Its raw failure remains archived. The final control waits on the client's native `Request.Cancel` channel; it reproduced erasure in all three runs on both versions. Supplied probes remained unchanged. Fresh race/vet runs were not repeated; their supplied raw results were inspected. No external service, load or long-term behavior is inferred from these local checks.

## WR A — Correctness & Compatibility — C

Scope: Complete `WR/A/source` Submit library, Go 1.22.0 minimum; see shared redirect interpretation.
Coverage: POST method, binary payload/key across inner and outer retries, validation, status and transport policy, total/parent/native deadlines, body bounds/closure, exposed error identity, API/module compatibility. Arbitrary uncancellable transports and unconditional permissive-redirect compatibility are limited as above.
Rationale: One contained major issue breaks the stated budget/caller-stop contract through a supported legacy cancellation facility. It affects this outbound operation, not multiple distinct workflows; no systemic or critical reach is demonstrated.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [WR-AG1] `client.go:24`, `client.go:29`, `client.go:52`, `client.go:56` reject empty deduplicated identity, clone the binary payload, create a fresh POST body, and reuse the supplied key. `client_test.go:423` and frozen `TestIdentitySurvivesOuterRetry` verify three lost replies followed by an outer caller retry with one logical effect. This core replay contract is strong.
- [WR-AG2] `client.go:43`, `client.go:67`, `client.go:107` restrict application attempts to one without deduplication and three with it, classify terminal statuses, bound response reads, reject oversized success and close owned bodies. Candidate and frozen tests pass these paths on both toolchains. Known policy and malformed-redirect statuses are terminal; available transport/body causes are retained.

Bad

- [WR-A1][major][existing-in-scope] The operation context does not reach legacy transport cancellation when client timeout is absent or later than the operation deadline (`client.go:52`, `client.go:88`, `client.go:146`). Merely forwarding `CancelRequest` cannot make it run: no `Request.Cancel` is attached, and the copied client's timeout is not capped to the remaining operation budget. With a 60ms parent deadline, a legacy RoundTripper/body waits for the reproduction's external 250ms release; all four timeout/body combinations take about 250–252ms. With no client timeout, the 500ms operation waits about 651–652ms for external release. Without that safety release the call would remain blocked. Both versions return the deadline identity only afterward. Primary remediation owner: correctness.

Suggested changes

- [WR-A1] Connect caller/operation cancellation to the client's native legacy cancellation path and cap the per-call native timeout to the remaining total budget while preserving any shorter supplied timeout and leaving the supplied client unchanged. Verify both headers and body work with zero and longer client timeout, an earlier parent deadline and caller cancellation; the adjacent write reproduction isolates these cases.

Limits: Supplied short legacy-client-timeout probe passes; it does not exercise absent/longer timeout or earlier parent cancellation. Native network tests also pass. The confirmed legacy budget failure does not imply duplicate writes or failure of the independently verified identity contract. Shared verification and contract limits apply.

## WR A — Observability & Resilience — C

Scope: Submit failure handling, retry safety, cancellation and diagnostic boundary in `WR/A/source`.
Coverage: One application retry owner, deduplication and ambiguous effect, bounded attempts, cancellable backoff, budget through body reads, known status and error causes. Service logs, metrics, traces, breakers and lifecycle are not required by this small library contract.
Rationale: [WR-A1] is one contained major reliability-contract failure. Correctness owns remediation; this topic independently counts its loss of cancellation/budget containment.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [WR-ARG1] `client.go:43`, `client.go:70`, `client.go:100` provide safe replay gating, a small cancellable delay and an explicit “outcome unknown” transport diagnostic. Outer replay and no-dedup lost-reply tests confirm the effect policy.
- [WR-ARG2] `client.go:92`, `client.go:104` observe and rejoin transport/body causes that native timeout handling can replace. Owned response closure and bounded body diagnostics are verified. No speculative breaker or instrumentation was added.

Bad

- [WR-A1][major][existing-in-scope] Cross-reference the confirmed legacy budget failure above: a stopped caller can keep transport/body work alive until an unrelated longer client timer or external release. This defeats the operation's specified containment boundary.

Suggested changes

- [WR-A1] Wire native legacy cancellation and the remaining budget as described above; retain the existing single retry owner and stable operation identity.

Limits: Native context-aware HTTP transport cancellation and the shorter supplied timeout pass. No load, connection exhaustion or broader system outage is established. Shared verification and contract limits apply.

## WR A — Testing — B

Scope: Candidate-owned `WR/A/source/client_test.go`; frozen/supplemental probes and new reproductions are independent verification evidence, not maintained candidate regression tests.
Coverage: Attempt/status matrix, binary payload and key, body size/closure/read errors, cancel causes, outer failed-invocation replay and committed effects, client policy/jar, native body deadline and short client timeout. Legacy operation/parent/caller cancellation is not covered in candidate tests.
Rationale: One moderate boundary-coverage gap. Important native budget and replay behavior are meaningfully verified, so the whole cancellation contract is not effectively unverified and a major testing classification would overstate the gap.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [WR-ATG1] `client_test.go:59`, `client_test.go:196`, `client_test.go:423` assert calls, exact binary bodies, supplied key, closure before subsequent attempts, result/error identity and one server effect across an outer retry. These assertions detect consequential changes rather than mirroring implementation expressions.
- [WR-ATG2] `client_test.go:369`, `client_test.go:520` exercise real network body cancellation and a timeout-associated independent body cause. Both toolchains pass; cleanup and broad elapsed allowances contain the tests.

Bad

- [WR-AT1][moderate][existing-in-scope] `client_test.go:504` and `client_test.go:520` use transports that cooperate with `Request.Context`; no test exercises the `CancelRequest` implementation at `client.go:146` with absent/longer client timeout or parent/caller stopping. Thus all supplied checks pass while [WR-A1] remains. The shorter-timeout supplemental probe also passes and does not close this gap.

Suggested changes

- [WR-AT1] Add bounded legacy-transport/header/body tests with observable start/cancellation and cleanup release; assert the original cause, single body close and elapsed budget. The adjacent reproduction supplies a minimal failing case.

Limits: Test grading concerns the candidate regression suite. Independent probes establish behavior but do not substitute for adding regression coverage after fixing the defect. Shared verification limits apply.

## WR B — Correctness & Compatibility — A

Scope: Complete `WR/B/source` Submit library, Go 1.22.0 minimum; grade covers the resolved requirements and has the same shared redirect compatibility limit as A.
Coverage: Identity/body across inner and outer caller retries, no-dedup ambiguity, status/transport eligibility, attempts, initial-policy terminal behavior, body limits/closure, caller/total/native legacy cancellation, available causes, unchanged exported signature/module/client ownership.
Rationale: No actionable defect was verified in the assessed requirements. Stable operation identity and effective native cancellation work on both supported toolchains. These directly satisfy the specified contract; additional implementation/test size does not itself earn A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [WR-BG1] `client.go:18`, `client.go:26`, `client.go:76`, `client.go:80` preserve the same supplied key and cloned exact body; `client.go:57` forbids application replay without deduplication. Candidate lost-acknowledgement tests plus the frozen failed-outer-invocation replay verify one effect and identical key/body. A's replay strength is preserved.
- [WR-BG2] `client.go:86`, `client.go:90`, `client.go:166` link operation cancellation to the native legacy facility and cap only the copied client's timeout. Earlier-parent/header/body reproductions return around 62ms, and total-budget work around 500–502ms, while original client timeout/transport and available causes remain intact. Known policy statuses and response ownership pass.
- [WR-BG3] `client.go:112` bounds read/result work, uses one oversize byte only for success, closes bodies and preserves read/close causes. No exported API or dependency is added.

Bad

- None found in the assessed requirements.

Suggested changes

- None needed for the assessed requirements; resolve the shared redirect-acceptance ambiguity before making an unconditional compatibility guarantee.

Limits: Rejecting a redirect after calling a permissive caller callback changes its effective policy, just as A does; that unresolved shared contract choice is not a demonstrated advantage. An arbitrary transport that ignores all cancellation remains outside enforceable native behavior. Shared verification limits apply.

## WR B — Observability & Resilience — A

Scope: Submit retry, budget, cancellation, effect ambiguity and error boundary in `WR/B/source`.
Coverage: Single application retry owner; deduplication and native versus application attempts; cancellable 10/20ms backoff; total/parent/short client timeout through header/body work; legacy facility; status and available error causes. Service telemetry/lifecycle is not implicated.
Rationale: No actionable issue found in the assessed reliability requirements. The meaningful improvement is effective caller/total cancellation at the legacy boundary, not backoff shape or extra wrappers alone.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [WR-BRG1] `client.go:45`, `client.go:57`, `client.go:60` gate safe attempts and use context-cancellable delay within a single total budget. Retry/outer-replay and no-dedup lost-reply evidence passes.
- [WR-BRG2] `client.go:86`, `client.go:104`, `client.go:118`, `client.go:157` make cancellation effective through the native facility and expose transport/body/context causes despite timeout conversion. Legacy and network evidence verifies both containment and response closure.

Bad

- None found in the assessed requirements.

Suggested changes

- None needed in assessed scope.

Limits: A plain underlying EOF for a non-deduplicated lost acknowledgement leaves the effect ambiguous; it does not claim no effect occurred, and the function documentation states uncertainty. Unlike A it does not add that phrase to every transport error, which is a diagnostic choice rather than a contract violation. Shared redirect and verification limits apply.

## WR B — Testing — A

Scope: Candidate-owned `WR/B/source/client_test.go`, with independent probe/reproduction evidence assessed separately.
Coverage: Wire-level identity and binary payload, committed lost acknowledgement, repeated caller invocation, status/attempt policy, rejection/oversized bounds, cause discovery, client settings, network body deadlines, shared deadline and legacy timeout/total budget/caller cancellation with bounded cleanup.
Rationale: No substantiated regression-detection gap in the resolved requirements. Focused legacy tests verify the boundary A omits; more tests or longer runtime are not the reason for the grade.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [WR-BTG1] `client_test.go:28`, `client_test.go:71`, `client_test.go:145` assert wire identity/body, one effect after lost acknowledgement and exact application outcomes. The frozen outer retry independently verifies the three-lost-reply case covered directly by A's maintained test.
- [WR-BTG2] `client_test.go:481`, `client_test.go:507`, `client_test.go:538`, `client_test.go:556` exercise native legacy cancellation through timeout, total budget and caller stopping, preserve causes and use explicit cleanup release. New parent-deadline/header/body reproductions corroborate this coverage on both versions.

Bad

- None found in resolved scope.

Suggested changes

- None needed in resolved scope.

Limits: The direct candidate outer-replay test has a successful first invocation; the independent frozen probe supplies failed-first-invocation coverage. Both pairs' initial-response redirect tests encode the same unresolved policy choice. Shared verification limits apply.

## RB A — Correctness & Compatibility — C+

Scope: Complete `RB/A/source` Fetch library and supplied-client boundary; Go 1.22.0 minimum.
Coverage: Exact final HTTP 200 success, 429/503-only application retry, transport/read/close terminal paths, attempts and one budget, Retry-After integer/date/fallback/overflow, rejection retention while waiting, body bounds/ownership, known policy status, diagnostics and error chains. Native client erasure is limited as above.
Rationale: Two independent moderate contract mismatches: an available response status is discarded, and returned diagnostic text is not bounded. Neither changes retry eligibility or demonstrates broad/irreversible failure.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [RB-AG1] `client.go:19`, `client.go:22`, `client.go:53`, `client.go:58`, `client.go:68` correctly share one budget and limit attempts, parse integer overflow as an unfit delay, handle future/past dates and use cancellable fallback. Frozen probes and candidate native-network/body checks pass.
- [RB-AG2] `client.go:30`, `client.go:35`, `client.go:108` use the supplied client directly, close owned bodies, preserve available transport/read/close/parent causes and retain the rejection when waiting stops. No wrapper, breaker, API or dependency is added.

Bad

- [RB-A1][moderate][existing-in-scope] `client.go:31` ignores a nonnil response returned with a redirect-policy error. A caller-rejected 302 returns only `Get "/next": caller rejects redirect`, losing known `HTTP 302` despite the status-error requirement. Supplemental and ownership reproductions confirm cause preservation, one native request and one already-owned closure, isolating the missing status. Primary remediation owner: correctness.
- [RB-A2][moderate][existing-in-scope] `client.go:49`, `client.go:108` bound body bytes but join arbitrary error text without a diagnostic cap. A 10,000-byte transport cause produces a 10,030-byte returned diagnostic while retaining its identity; adding read/close/parent cause text similarly defeats the diagnostic limit. This is a contained bound/diagnostic contract mismatch. Primary remediation owner: correctness.

Suggested changes

- [RB-A1] When `Do` returns an error and nonnil response, include its status in the terminal error while retaining the native error chain and trusting Do's already-closed body ownership. Do not retry or close it a second time.
- [RB-A2] Separate bounded `Error()` text from the unmodified unwrap chain. Preserve status at the start of the bounded text and keep `errors.Is`/`As` working for all available causes; verify long transport/read/close/parent errors and full-size body diagnostics without a client wrapper.

Limits: The timeout-cause supplemental failure is not counted; the native client can already erase that cause. A's ordinary 4096-byte rejection body results in a status prefix beyond 4096 total bytes as well; body byte count alone is not the selected diagnostic-text contract. Shared verification and boundary limits apply.

## RB A — Observability & Resilience — C+

Scope: Fetch retry guidance, containment and diagnostic/error boundary in `RB/A/source`.
Coverage: Eligible statuses and terminal transport failures, native settings, total budget through waiting/body, rejection retention, bounded request work and returned signals. Service telemetry/lifecycle is not implicated.
Rationale: [RB-A1] and [RB-A2] are two moderate operational diagnostic-contract failures, shared with correctness. Budget and retry policy remain verified strengths.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [RB-ARG1] `client.go:58`, `client.go:94` promptly return an unfit hint and cancel retry waits while retaining the HTTP rejection and available parent causes. Overflow does not create an immediate storm of requests.
- [RB-ARG2] `client.go:30`, `client.go:35` preserve the supplied client's native settings and limit body reads; no second application retry owner surrounds a transport failure.

Bad

- [RB-A1][moderate][existing-in-scope] A known rejected redirect status is omitted, reducing the caller's failure signal even though the response status is available.
- [RB-A2][moderate][existing-in-scope] Arbitrary cause text bypasses the promised diagnostic bound, imposing unbounded returned/loggable text on callers. No wider memory exhaustion is claimed.

Suggested changes

- [RB-A1] Add the available status to the terminal policy error without changing native policy/ownership.
- [RB-A2] Bound displayed diagnostic text separately from error identity as described above.

Limits: No logging/metrics/tracing framework is required. Native timeout type/cause conversion remains a supplied-client constraint; adding the prohibited wrapper is not a suitable corrective change. Shared verification limits apply.

## RB A — Testing — C+

Scope: Candidate-owned `RB/A/source/client_test.go`; controller probes and new tests are independent evidence.
Coverage: Meaningful attempt/status matrix, body read count/closure, read/close/transport and parent causes, overflow/unfit hints, fallback lower bound, helper future-date arithmetic, native shared budgets and redirect policy. Diagnostic text from long causes and known redirect status assertions are missing.
Rationale: Two moderate gaps in contained public error behavior; substantial request, budget and retry-hint contracts are covered. The supplemental probes reveal these gaps without implying all testing is ineffective.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [RB-ATG1] `client_test.go:53`, `client_test.go:104`, `client_test.go:145`, `client_test.go:186` assert attempts, terminal errors, bytes read, closure and available causes, including a failure after a rejection.
- [RB-ATG2] `client_test.go:213`, `client_test.go:316`, `client_test.go:346`, `client_test.go:391` verify overflow, fallback, future-date calculation and real budget-through-body behavior on both toolchains.

Bad

- [RB-AT1][moderate][existing-in-scope] `client_test.go:104` checks only the count of response-body characters; `client_test.go:145` and `client_test.go:186` use short cause text. No maintained test requires a long returned diagnostic to remain bounded while preserving error identity, allowing [RB-A2].
- [RB-AT2][moderate][existing-in-scope] `client_test.go:431` and `client_test.go:448` verify policy cause and closure/settings but never assert the available HTTP 302 status, allowing [RB-A1]. This same assertion gap exists in B's candidate tests.

Suggested changes

- [RB-AT1] Add long transport/read/close/parent cause cases with both full diagnostic-size and `errors.Is`/`As` assertions; keep the existing body-read cap check.
- [RB-AT2] Extend the policy-error test to assert `HTTP 302`, one request and one closure along with the cause. The unchanged supplemental status probe already demonstrates the failing assertion.

Limits: A's future-date helper test proves parsing arithmetic but does not directly test a fitting future hint through Fetch's wait. B adds that integration evidence; this is a comparative strength rather than another counted defect because the separate wait helper and boundary tests already cover the composition substantially. Shared verification limits apply.

## RB B — Correctness & Compatibility — A

Scope: Complete `RB/B/source` Fetch library, supplied client boundary, Go 1.22.0 minimum.
Coverage: Success/status/transport policy, attempts, total and parent budget through wait and body, Retry-After integer/date/overflow/fallback, rejection retention during stopped waiting, ownership and body/result limits, known returned policy status and bounded diagnostic text, available error identity and unchanged public API.
Rationale: No substantiated actionable issue within the explicit client-boundary contract. Known status and bounded diagnostic behavior improve two concrete cases without violating the no-wrapper constraint. These required controls support A rather than a grade based on extra tests/comments.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [RB-BG1] `client.go:20`, `client.go:56`, `client.go:63`, `client.go:81` preserve the correct retry/budget/hint policy; future fitting-date checks confirm no request before the hint and cancellation retains the rejection. A's overflow and total-budget strengths remain intact.
- [RB-BG2] `client.go:34`, `client.go:110`, `client.go:118` retain a known policy-error status and bound `Error()` to 4096 bytes while retaining the original causes through `Unwrap`. Fresh supplemental/probe checks show `HTTP 302`, one closure, 4096 diagnostic bytes and `errors.Is` success for the long cause.
- [RB-BG3] `client.go:31`, `client.go:40`, `client.go:41` preserve native settings/ownership, enforce at most 4097 body bytes, reject oversized success and preserve available read/close causes without a prohibited client wrapper.

Bad

- None found in the explicit assessed contract.

Suggested changes

- None needed in the explicit assessed contract.

Limits: Native body timeout cause erasure is shared with A. The lost prior rejection on a later HTTP 200 read failure is outside the README's specific waiting-retention promise; a broader promise would require a change. Supported native transports must honor their contexts/cancellation; arbitrary legacy cancellation rescue is not promised by a wrapper here. Shared verification limits apply.

## RB B — Observability & Resilience — A

Scope: Fetch retry/budget containment and diagnostic boundary in `RB/B/source`.
Coverage: 429/503 eligibility, terminal transport/read/close failures, no multiplication of application retry owners, native client timeout and parent context, cancelable hints, rejection retention during waits, bounded diagnostic signals and available cause discovery. No service lifecycle or telemetry infrastructure is implicated.
Rationale: No actionable issue found. The demonstrated benefits are complete bounded diagnostic text and preserving known status while keeping native client semantics; existing retry safeguards are ties with A.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [RB-BRG1] `client.go:65`, `client.go:69`, `client.go:103` promptly reject unfit delays, cancel waits and retain last rejection/parent causes within one budget.
- [RB-BRG2] `client.go:34`, `client.go:118` make status available and cap diagnostic text independently of unwrap identity. Network timeout/body and supplemental diagnostic checks verify those signals; no breaker/client wrapper is added.

Bad

- None found in assessed scope.

Suggested changes

- None needed in assessed scope.

Limits: The supplemental timeout probe requests a cause the configured native client may already erase. Return of a discoverable `net.Error` timeout is verified even where Go 1.22 lacks a discoverable deadline sentinel. No wrapper or string-based invented cause is warranted. Shared verification limits apply.

## RB B — Testing — B

Scope: Candidate-owned `RB/B/source/client_test.go`, independent frozen/supplemental and scratch checks assessed as evidence.
Coverage: Attempt/status policy, bounds/read counts and closure, long diagnostic error identities, close errors, overflow/unfit/fallback/past/future hints, canceled waiting, shared budgets and native client timeout/body cancellation. Known policy-response status assertion remains absent.
Rationale: One moderate assertion gap, matching A's [RB-AT2]. The long-diagnostic and fitting future-date coverage are substantive improvements; neither additional test count nor slower runtime is treated as quality by itself.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [RB-BTG1] `client_test.go:90`, `client_test.go:136`, `client_test.go:170` assert complete diagnostic caps, available error identities, byte read limits and single closure for success/rejection/read/close/transport cases.
- [RB-BTG2] `client_test.go:264`, `client_test.go:335`, `client_test.go:409` verify the actual fitting future-date wait, no earlier attempt, cancel retention and native timeout/parent/total-body deadlines with observable completion. Both toolchains pass these tests.

Bad

- [RB-BT1][moderate][existing-in-scope] `client_test.go:392` checks redirect cause, policy calls, closure and settings but not `HTTP 302`. Removing status construction at `client.go:34` would still pass the candidate-owned test despite losing a stated error signal. The independent supplemental probe verifies the current implementation, but should be carried into regression coverage.

Suggested changes

- [RB-BT1] Add a known-status assertion to `TestSuppliedRedirectPolicy`, alongside the existing cause and ownership checks, or maintain the equivalent unchanged supplemental test with the suite.

Limits: The fitting HTTP-date tests depend on a whole-second wall-clock window and may skip when the scheduler misses it; that explicit bounded skip is a limit, not an observed flake. They pass fresh checks. Shared verification limits apply.

## Explicit substantive comparisons

**WR:** B improves a demonstrated reliability/correctness boundary: legacy native cancellation now enforces both an earlier caller deadline and the one total operation budget for headers/body, with shorter supplied timeout and available causes retained. Its focused maintained legacy tests are an improvement over A's context-cooperative timeout doubles. Core replay safety is a tie: both use the caller identity/exact body across inner and outer invocations, reject empty deduplicated identity, allow one no-dedup attempt and leave lost acknowledgements ambiguous; both frozen wire probes pass. A has particularly strong direct failed-outer-invocation coverage and clearer per-error “outcome unknown” wording. B's 10/20ms versus A's constant 10ms delay and suppression of `GetBody` without deduplication are not independently demonstrated improvements to the application-attempt contract. Both share the unresolved initial-response redirect choice. B adds private observation/stop machinery, but its complexity has a demonstrated native-cancellation/cause purpose; neither adds dependency or public surface. B drops rejection-body content from its returned status diagnostic, which still satisfies the stated requirement; A provides that bounded content.

**RB:** B improves two concrete public error behaviors: it retains a known redirect-policy response status and bounds the entire displayed diagnostic while preserving available causes. B also tests long error text and a fitting HTTP-date through the complete wait path. Attempt ownership, 429/503 eligibility, terminal transport failures, overflow-as-unfit policy, one budget, rejection retention while waiting, bounded body reads/closure, native settings and zero additional dependencies are ties. Both appropriately obey the no-client-wrapper constraint; neither should be penalized for an arbitrary cause erased inside the configured HTTP client. The maintained status assertion gap is shared, despite B's current implementation satisfying it. A's smaller error helper and broader preservation of an earlier rejection on a later 200-body failure are strengths, though the latter exceeds the explicit waiting-retention promise. B's private error type is justified by separating bounded text from discoverable causes and adds no exported API. No grade comparison alone, added comments, raw test count or supplemental-failure count substitutes for these behavioral differences.
