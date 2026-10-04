# Independent neutral-pair review: HTTP and arithmetic

Frozen before any arm disclosure. Labels A and B are neutral; no candidate guidance, arm maps, other trials, author reports, campaign summaries, or Git diff were inspected. Each complete implementation was assessed against its identical original README. This is a code-area review of existing-in-scope behavior, not a base/head comparison.

Sources are under `/private/tmp/go-client-campaign/reviews/pairs/`. Locations below are relative to that directory. All supplied source hashes were verified before checks. The unchanged Correctness & Compatibility, Observability & Resilience, and Testing skills and their relevant decision references were applied.

| Case | Label | Correctness & Compatibility | Observability & Resilience | Testing |
| --- | --- | --- | --- | --- |
| http-write-replay | A | A | A | A |
| http-write-replay | B | C | C | B |
| http-retry-budget | A | C+ | C+ | C+ |
| http-retry-budget | B | C+ | C+ | C+ |
| pure-control | A | A | Not applicable | A |
| pure-control | B | A | Not applicable | A |

Grades follow distinct root causes and consequences, not implementation length or test count. Shared production findings keep the same ID across correctness and resilience. Independently actionable regression gaps have separate testing IDs. Do not sum topic counts as a total.

## Verification and scope

All modules declare Go 1.22.0 and use only the standard library. Host Go is 1.26.5; the minimum toolchain is Go 1.22.12. Supplied controller `checks.json` reports all public/private commands passing: both toolchains' tests and vet; HTTP candidates additionally have host race results. These are controller observations, not new executions.

In disposable copies, I independently ran the complete supplied tests plus unchanged frozen probes:

- Go 1.22.12: all six candidates passed `test -ldflags=-linkmode=external -count=1 -timeout=20s ./...`.
- Go 1.26.5: all four HTTP candidates passed `test -ldflags=-linkmode=external -race -count=1 -timeout=20s ./...`; both arithmetic candidates passed the same command without `-race`.
- Go 1.22.12: all six passed `vet ./...`.

The exact shell prefix for every Go check was:

```sh
rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off
```

The executable was either `go` or `/private/tmp/go-client-campaign/tools/go/bin/go`. Baseline copies were under `/private/tmp/go-neutral-http-review/baseline/<case>/<label>`. Focused reproductions used copies under `/private/tmp/go-neutral-http-review/<case>/<label>` with `test -ldflags=-linkmode=external -run TestReview -v -count=1 -timeout=20s ./...`. No supplied candidate or frozen probe was edited.

[Reproduction sources and raw results](pair-http-and-control-reproductions/) contain the two added test files and `results.json`. All consequential production findings below were reproduced on both toolchains. Intentional failures in these additional checks do not contradict the passing supplied suites.

The write probe genuinely tests stable identity/body across an exhausted inner invocation and a later caller invocation, one server effect, and no repetition without deduplication. The Fetch probes test eligible status retries, huge integer hints, a one-second unfit hint, and oversized-result closure. They do not cover supplied-client timeout error masking, redirect chains, or arbitrary diagnostic error size. The arithmetic probe detects negative contribution.

I interpret “other statuses are terminal” to include a 302 with a Location header. Application attempts are calls owned by Submit/Fetch; transport replay and HTTP-client redirects are separate layers. R1 concerns terminal-status policy, not a claim that two redirect exchanges equal two application attempts. Arbitrary caller transports that ignore every cancellation mechanism are outside a defensible deadline guarantee; W1 instead uses a transport with a cancellation mechanism that the supplied HTTP client demonstrably honors.

## Substantiated finding register

### W1 — Write B discards supported client timeout cancellation

**Major; existing-in-scope; primary owner: Correctness & Compatibility.** Locations: `http-write-replay/B/source/client.go:25-27,70-84,103-114`.

B zeroes the copied client's Timeout, creates an attempt context instead, and wraps the transport without forwarding CancelRequest. A pre-context RoundTripper using the HTTP client's supported legacy cancellation hook therefore loses the supplied timeout. A direct client with Timeout=25ms completes in about 27ms; B's Submit completes only after the reproduction's explicit 700ms release, exceeding both that client timeout and the 500ms operation budget. Without that test release the transport can remain blocked. The caller's struct field still says 25ms, so an unchanged field is insufficient evidence of retained behavior.

The same reproduction passes for A in about 26–27ms. A retains Client.Timeout and forwards CancelRequest at `A/source/client.go:146-150`. Primary Go 1.22 source confirms that `net/http/client.go:351-386` uses the client deadline and transport CancelRequest hook; a zero deadline bypasses that path.

**Change:** preserve the supplied timeout's effective cancellation behavior, including legacy hooks, while retaining body-error identity. Keep the correction private and retain one total operation deadline. Verify direct-client and Submit behavior against the same cancellation-aware legacy transport. The major reach is contained to this important supported client/deadline boundary; it is not critical or systemic.

### R1 — Both Fetch implementations bypass terminal redirect statuses

**Moderate; existing-in-scope; primary owner: Correctness & Compatibility.** Locations: `http-retry-budget/A/source/client.go:35-38`; `B/source/client.go:30-33`.

Both call the supplied client's Do directly. A normal 302 with Location followed by 200 produces two HTTP exchanges and returns success, rather than returning HTTP 302 as the terminal status. If the supplied CheckRedirect rejects the redirect, both preserve that policy error but discard the available response status; the returned error omits HTTP 302. Both effects were reproduced using a configured RoundTripper and the actual Client.Do redirect boundary.

**Change:** expose and enforce the initial terminal status without mutating caller-owned configuration. Preserve the caller's redirect callback/error, and include the returned redirect status when Do reports a policy failure. Add a Location-bearing redirect-to-200 test and a policy-rejection status assertion. The defect is a contained normal-use status-policy mismatch, not unsafe mutation or an attempt-budget multiplication finding.

### R2 — Both Fetch implementations lose underlying body failure identity under Client.Timeout

**Moderate; existing-in-scope; primary owner: Correctness & Compatibility.** Locations: `http-retry-budget/A/source/client.go:35,40-47,111-119`; `B/source/client.go:30,35-42,108-112`.

With a supplied 15ms Client.Timeout, the response body waits for the request context to stop and returns an independent sentinel joined with context.DeadlineExceeded. Both Fetch implementations close the body once, but neither returned error matches the independent sentinel with errors.Is. On Go 1.22.12 neither matches context.DeadlineExceeded either; the operation's own 250ms context has not expired, and the client replaced the read error. Go 1.26.5 preserves DeadlineExceeded identity in its replacement error, but still loses the independent sentinel.

Primary Go 1.22 source at `net/http/client.go:970-982` replaces a non-EOF body error after Client.Timeout with an httpError containing its text, not its error chain. Joining errors after that replacement cannot recover their identity.

**Change:** account for supplied-client timeout masking before wrapping the result, preserving actual body/context failures under the minimum supported toolchain and effective client timeout behavior. Verify with a short client timeout, an independent body sentinel, errors.Is assertions, and closure. Avoid “fixing” this by merely substituting context cancellation and losing supported timeout behavior as W1 demonstrates.

### R3 — Fetch B does not bound returned diagnostics

**Moderate; existing-in-scope; primary owner: Observability & Resilience.** Locations: `http-retry-budget/B/source/client.go:49,108-112`.

B caps response body bytes, but formats/join errors without a diagnostic cap. A supplied transport returning a 10,000-byte error yields a 10,030-byte returned diagnostic, despite the README's 4096-byte result/diagnostic limit. Even a full 4096-byte rejection body has additional status-prefix bytes. Error identity remains intact.

A returns exactly 4096 diagnostic bytes for the same 10,000-byte error while still exposing its cause through errors.Is (`A/source/client.go:122-148`). This is a substantive bound, not a preference for A's custom error type.

**Change:** cap displayed diagnostics independently of the unwrapped cause chain; retain HTTP status in the prefix and errors.Is/As support. Add long transport/read/close-cause cases as well as body-size cases. The demonstrated impact is contained diagnostic growth and a policy mismatch, not a claim of critical resource exhaustion.

### WT1 — Write B's timeout tests miss the cancellation behavior it replaces

**Moderate; existing-in-scope; primary owner: Testing.** Locations: `http-write-replay/B/source/client_test.go:288-299,462-491`.

The fake timeout body waits on Request.Context; the real network test uses a modern context-aware HTTP transport. Both pass after replacing Client.Timeout with an attempt context. Configuration assertions also check retained field values. None detects W1's lost CancelRequest behavior. Ordinary modern transport deadlines are substantially covered, so this is a narrow but consequential moderate regression gap rather than a claim that the whole budget contract is untested.

**Change:** add the direct-client/Submit legacy-timeout comparison used in the reproduction, with an explicit release and bounded failure path.

### RT1 — Both Fetch test suites miss Location-bearing redirect outcomes

**Moderate; existing-in-scope; primary owner: Testing.** Locations: `http-retry-budget/A/source/fetch_test.go:47-79,366-390`; `B/source/client_test.go:53-100,431-444,448-478`.

A's redirect test exercises a rejecting callback or ErrUseLastResponse; its policy-error case never requires HTTP 302. B's “redirect is terminal” status case has no Location, so Client.Do does not follow it; its policy-error cases also omit the status assertion. These checks cannot catch either reproduced R1 symptom.

**Change:** test a genuine redirect chain with the default/accepting callback and a rejecting callback that requires both status and cause.

### RT2 — Both Fetch test suites miss minimum-version client timeout masking

**Moderate; existing-in-scope; primary owner: Testing.** Locations: `http-retry-budget/A/source/fetch_test.go:82-115,302-364,398-425`; `B/source/client_test.go:288-313,391-428`.

A tests ordinary body sentinels without a short Client.Timeout. B's independent-failure test uses explicit caller cancellation without Client.Timeout, and its real body test gives the client a two-second timeout, longer than the operation deadline. These tests cover useful neighboring paths but cannot catch R2.

**Change:** combine a client timeout shorter than 250ms with independent body/context causes and errors.Is assertions on Go 1.22.12.

### RT3 — Fetch B's bounded-diagnostic test only counts response body bytes

**Moderate; existing-in-scope; primary owner: Testing.** Location: `http-retry-budget/B/source/client_test.go:104-139,145-211`.

The diagnostic assertion counts 4096 “x” bytes instead of bounding the returned error. Failure tests use short sentinel strings. Thus the complete suite passes with R3's 10,030-byte diagnostic.

**Change:** assert the displayed diagnostic bound for long transport/read/close causes while independently checking cause identity. A already has those meaningful assertions at `A/source/fetch_test.go:302-364`.

## http-write-replay A

### Correctness & Compatibility — A

Scope: Complete Submit code area; Go 1.22.0 minimum.
Coverage: Exact status policy; one/three application attempts; caller stopping; payload/identity across inner and outer retries; ambiguous completion; body limits/closure; borrowed client and errors.Is behavior.
Rationale: No actionable defect verified; actual request and replay paths satisfy the README.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [WA-C1] Caller identity and cloned bytes are reused (`A/source/client.go:24-29,43-59`); the supplied and frozen committed-write tests demonstrate one effect across lost replies and outer retry.
- [WA-C2] HTTP status/client-policy failures are distinguished from transport failures; short supplied-client timeouts and masked body causes remain discoverable (`client.go:82-118,126-164`). W1's reproduction passes for A.

Bad

- None found.

Suggested changes

- None needed.

Limits: All baseline checks described above passed. Custom RoundTrippers must honor a cancellation facility; arbitrary blocking implementations and real server deduplication beyond the established README contract were not assessed.

### Observability & Resilience — A

Scope: Submit's outbound dependency boundary.
Coverage: Ambiguous outcomes; deduplication safety; total/request/body/wait budgets; bounded bodies and terminal statuses; error causes; ownership.
Rationale: No actionable issue; independent verified controls prevent unsafe replay and bound normal network/body work.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [WA-O1] No deduplication means one attempt and no header; errors explicitly describe unknown outcome (`client.go:43-46,100,118`).
- [WA-O2] One operation context and cancellable 10ms waits share the budget; bodies close and diagnostic body prefixes are capped (`client.go:27-28,70-76,102-121`). Real body/deadline tests pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: Library errors are the diagnostic boundary; service metrics, traces, breakers, and lifecycle controls are not required by this scope. Supplied-client diagnostic error strings themselves are not given a separate total-text limit by the write README.

### Testing — A

Scope: Complete write A test suite and frozen probes.
Coverage: Real HTTP success, committed-write reply loss and outer retry, status/error matrices, body-byte/closure observations, context/cause preservation, client timeout/Jar/policy, and minimum-version/race execution.
Rationale: Focused assertions detect meaningful replay, ownership, and deadline regressions; no actionable test defect found.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [WA-T1] `client_test.go:423-484` exercises an effect accepted before reply loss; asserts calls, effect count, exact bodies/keys, and outer result.
- [WA-T2] `client_test.go:196-251,369-420,520-534` checks byte bounds, closure, real body cancellation, and minimum-version error identity.

Bad

- None found.

Suggested changes

- None needed.

Limits: The review's added legacy-timeout reproduction is external evidence, not part of the supplied regression suite. Passing race checks cover exercised paths only.

## http-write-replay B

### Correctness & Compatibility — C

Scope: Complete Submit code area; Go 1.22.0 minimum.
Coverage: Same contract paths as A, including actual effective supplied Timeout behavior.
Rationale: W1 is one contained major break of the client/deadline contract; no further counted production defects.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [WB-C1] Exact operation identity and cloned payload survive retries; the frozen exhausted-invocation/outer-retry probe passes (`B/source/client.go:17-22,41-48`).
- [WB-C2] 200 is the only success; deduplication gates retries, and redirect policy failures are terminal (`client.go:24-35,56-58,85-94,116-130`).

Bad

- [W1][major][existing-in-scope] Supported client timeout cancellation is lost; confirmed at both toolchain versions.

Suggested changes

- [W1] Preserve effective Client.Timeout/CancelRequest behavior and verify the same client at the direct and Submit boundaries.

Limits: Baseline public/frozen/race checks pass; the additional timeout reproduction fails after its 700ms release. Custom cancellation causes are not joined separately, but context.Canceled and actual returned transport causes are preserved; no extra defect is inferred solely from that difference.

### Observability & Resilience — C

Scope: Submit's outbound dependency boundary.
Coverage: Replay safety, one total deadline, waits/body work, client Timeout, uncertainty, and closure.
Rationale: W1 breaks an important stated reliability bound for a supported cancellation mechanism; reach is contained.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [WB-O1] Both non-deduplicated failure and lost acknowledgement paths avoid application replay; the public comment warns that effects may have completed (`client.go:13-15,56`).
- [WB-O2] Modern context-aware transports share the 500ms operation bound and cancellable 10/20ms waits; response bodies close and oversized success fails (`client.go:20,59-65,116-130`).

Bad

- [W1][major][existing-in-scope] Field retention does not preserve cancellation behavior; the configured 25ms timeout and 500ms bound are exceeded.

Suggested changes

- [W1] Restore the client cancellation facility while preserving body failure causes; do not rely on unchanged field values as verification.

Limits: No separate telemetry or breaker is warranted. Ordinary modern HTTP body cancellation is verified; arbitrary uncooperative transports are outside scope.

### Testing — B

Scope: Complete write B test suite and frozen probes.
Coverage: Status matrix; byte/closure limits; transport/body errors; caller and total deadlines; real HTTP body work; deduplicated effects and outer reuse; client policy/Jar.
Rationale: One moderate gap, WT1, misses the specific timeout behavior changed by the implementation; neighboring important contracts have meaningful tests.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [WB-T1] `client_test.go:316-375` observes exact key/body/effect counts through transport acknowledgement loss, 503, and a second invocation; the independent frozen probe also covers exhausted inner attempts.
- [WB-T2] `client_test.go:101-179,222-299,462-491` tests result/diagnostic prefixes, closure and modern deadline behavior with useful assertions.

Bad

- [WT1][moderate][existing-in-scope] Existing timeout tests cannot detect W1's lost legacy cancellation hook.

Suggested changes

- [WT1] Add a bounded legacy-cancellation comparison; retain the modern timeout/body-cause tests.

Limits: All supplied/frozen checks pass; the externally added W1 regression fails on both Go versions. Timing allowances are not proof of precision at the exact deadline.

## http-retry-budget A

### Correctness & Compatibility — C+

Scope: Complete Fetch code area; Go 1.22.0 minimum.
Coverage: Exact status results, eligible application attempts, Retry-After grammar/overflow/date/fit, one budget, body completion/closure, failure identity and client policy.
Rationale: Two independent moderate issues, R1 and R2; neither is major or systemic.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [RA-C1] Overflowing syntactically nonnegative hints are unfit; future/date/fallback delays respect the remaining shared deadline (`A/source/client.go:68-108`; `fetch_test.go:117-195`).
- [RA-C2] Result overflow/read/close errors are rejected; diagnostic text is capped without losing the cause chain (`client.go:40-60,122-148`). Long-error reproduction passes.

Bad

- [R1][moderate][existing-in-scope] Initial redirect statuses and policy-failure statuses bypass the declared terminal-status result.
- [R2][moderate][existing-in-scope] Short supplied-client body timeouts mask underlying causes, including DeadlineExceeded on Go 1.22.

Suggested changes

- [R1] Enforce the initial status contract while retaining caller policy/error and ownership.
- [R2] Preserve timeout/body error identity before the HTTP client discards it; verify on the minimum toolchain.

Limits: Baseline/frozen checks pass; added R1/R2 checks fail. Redirect interpretation is stated above. A's extra error formatter preserves identity of errors it actually receives; it cannot recover errors already replaced by Client.Do.

### Observability & Resilience — C+

Scope: Fetch's retry and diagnostic boundary.
Coverage: Attempt ownership, one total budget, hints, cancellable waits, retained rejection/cause, effective client configuration, bounded response/error text.
Rationale: R1 is a contained status/retry-policy mismatch; R2 loses diagnostic identity under an ordinary client timeout. Two moderate issues.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [RA-O1] Valid unfit hints return promptly without a request; cancellation during waiting retains rejection and cause (`client.go:65-78,111-119`).
- [RA-O2] Useful status/cause diagnostics remain within 4096 displayed bytes, even for long dependency errors (`client.go:130-148`).

Bad

- [R1][moderate][existing-in-scope] The HTTP client's redirect layer hides a terminal status.
- [R2][moderate][existing-in-scope] Timeout masking prevents reliable cause classification.

Suggested changes

- [R1] Preserve initial status at the client boundary.
- [R2] Retain the timeout/body causes and their errors.Is behavior.

Limits: No service telemetry/lifecycle requirement applies. Total-budget enforcement works for the exercised modern HTTP boundary, independently of R2's error-chain defect.

### Testing — C+

Scope: A's complete `client_test.go`, `fetch_test.go`, and frozen probes.
Coverage: Status/attempt matrices, body/error bounds, real HTTP cancellation, context causes, future HTTP-date retry timing, malformed/overflow/unfit hints, and rejection retention.
Rationale: Two independently actionable moderate gaps, RT1 and RT2, leave the reproduced boundary mismatches undetected.
Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [RA-T1] `fetch_test.go:302-364` verifies long displayed diagnostics and retained sentinel identities, not just body length.
- [RA-T2] `fetch_test.go:167-195,216-249,256-295,398-425` observes future-date timing, cancellation while waiting, shared body/retry budgets and real I/O cancellation.

Bad

- [RT1][moderate][existing-in-scope] Redirect tests omit the followed-Location case and status assertion on callback error.
- [RT2][moderate][existing-in-scope] Body sentinel tests do not trigger a short Client.Timeout on Go 1.22.

Suggested changes

- [RT1] Add a redirect chain and status-plus-policy-cause assertions.
- [RT2] Add the short-client-timeout sentinel/deadline test on the minimum toolchain.

Limits: Future HTTP-date test can skip if scheduling misses its narrow real-time window; its parser and other delay cases remain covered. No failure from that window was observed. Controller and independently rerun checks pass.

## http-retry-budget B

### Correctness & Compatibility — C+

Scope: Complete Fetch code area; Go 1.22.0 minimum.
Coverage: Same Fetch contract dimensions as A.
Rationale: R1, R2 and R3 are three independent moderate mismatches; the anchor remains C+.
Finding counts: critical=0, major=0, moderate=3, minor=0

Good

- [RB-C1] Retry-After duration/integer overflow becomes a maximum unfit duration; past dates are immediate and malformed hints use 10ms (`B/source/client.go:68-91`).
- [RB-C2] At most three application calls share the operation context; read/close failures remain terminal and normally retain their causes (`client.go:19-63`).

Bad

- [R1][moderate][existing-in-scope] Redirects hide terminal statuses.
- [R2][moderate][existing-in-scope] Short client body timeout masks underlying identity.
- [R3][moderate][existing-in-scope] Returned diagnostic text exceeds the declared limit.

Suggested changes

- [R1] Enforce the initial terminal status and preserve policy status/cause.
- [R2] Preserve body/context error chains under the minimum-version client timeout.
- [R3] Bound display text independently of wrapped causes.

Limits: Supplied/frozen tests pass. Added checks reproduce all three issues on both toolchains; modern Go retains only the deadline identity in R2.

### Observability & Resilience — C+

Scope: Fetch retry, cancellation, and diagnostics.
Coverage: Hints, total/parent budget, body bounds/closure, last rejection/cause, supplied configuration, and surfaced failures.
Rationale: Three contained moderate issues, R1/R2/R3; no evidence of systemic or critical failure.
Finding counts: critical=0, major=0, moderate=3, minor=0

Good

- [RB-O1] Unfit hints stop promptly; timer waits are cancellable; cancellation/transport failure after rejection preserves the last status (`client.go:53-63,94-112`).
- [RB-O2] Responses are read at most 4097 bytes, closed, and oversized success is rejected (`client.go:35-47`).

Bad

- [R1][moderate][existing-in-scope] Redirect handling bypasses status policy.
- [R2][moderate][existing-in-scope] Body timeout hides diagnostic causes.
- [R3][moderate][existing-in-scope] Unbounded error strings defeat the diagnostic limit.

Suggested changes

- [R1] Keep initial status visible.
- [R2] Preserve underlying error classification.
- [R3] Cap displayed diagnostics while keeping errors.Is/As support.

Limits: Ordinary short error strings and body prefixes are bounded only by their inputs. No service telemetry or breaker is required.

### Testing — C+

Scope: Complete B Fetch test suite and frozen probes.
Coverage: Status/retry matrices; short body/close errors; retained rejection; Retry-After parsing/fit; shared deadlines; real body cancellation; caller policy/ownership.
Rationale: Three independent moderate gaps, RT1/RT2/RT3, explain why all supplied tests pass despite the additional failed checks.
Finding counts: critical=0, major=0, moderate=3, minor=0

Good

- [RB-T1] `client_test.go:53-100,145-209,213-258,346-428` asserts attempts, closure, retry delays, retained rejection and real body deadlines.
- [RB-T2] `client_test.go:288-313` checks independent body failures during explicit caller cancellation, including errors.As for a noncomparable cause.

Bad

- [RT1][moderate][existing-in-scope] A 302 fixture without Location does not verify real redirect policy.
- [RT2][moderate][existing-in-scope] Supplied-client timeout masking is not exercised.
- [RT3][moderate][existing-in-scope] Body-byte counting is not an error diagnostic bound.

Suggested changes

- [RT1] Add Location-bearing redirect and policy-status assertions.
- [RT2] Trigger a shorter Client.Timeout with independent body causes on Go 1.22.
- [RT3] Add long dependency error cases with length and identity assertions.

Limits: The fixed-clock future-date parser test verifies arithmetic, not integrated waiting to a future HTTP date. Neighboring timer/fit tests provide useful coverage; that difference alone is not an additional finding. All supplied/frozen/race checks pass.

## pure-control A

### Correctness & Compatibility — A

Scope: Complete SumPositive code area; Go 1.22.0 minimum.
Coverage: Public signature; strictly positive filtering; deterministic ordinary integer calculation; empty/nonpositive behavior by inspection and mixed input execution.
Rationale: Simple loop satisfies the contract; no actionable issue.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PA-C1] `A/source/client.go:3-10` adds only values greater than zero; supplied mixed-input regression and frozen probe pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: No mathematical sum beyond int range is promised; ordinary Go integer semantics remain unchanged. Both toolchains' tests and minimum-toolchain vet passed.

### Observability & Resilience — Not applicable

Scope: Deterministic local arithmetic.
Coverage: No external calls, waits, failure signals, retries, shared state, or lifecycle.
Rationale: There is no relevant resilience or observability decision; adding machinery would be unnecessary scope.
Limits: Authoring-guidance selection itself cannot be inferred from source and was not inspected.

### Testing — A

Scope: Complete arithmetic A tests and frozen probe.
Coverage: Existing all-positive case and one focused mixed positive/negative/zero regression.
Rationale: The regression detects the requested negative-value exclusion with clear expected output; no extra testing machinery is needed.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PA-T1] `client_test.go:11-14` checks mixed input sums to 5; the frozen negative-input probe also passes.

Bad

- None found.

Suggested changes

- None needed.

Limits: An implementation that also adds zero is observationally equivalent for this int sum; no test can distinguish that from strict filtering, and source inspection confirms the greater-than-zero condition.

## pure-control B

### Correctness & Compatibility — A

Scope: Complete SumPositive code area; Go 1.22.0 minimum.
Coverage: Same arithmetic behavior and signature as A.
Rationale: Production source is byte-for-byte identical to A, independently hash-verified and checked; no actionable issue.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PB-C1] `B/source/client.go:3-10` filters strictly positive values without state or dependencies; both test toolchains pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same ordinary int arithmetic boundary as A; no overflow-policy change was requested.

### Observability & Resilience — Not applicable

Scope: Deterministic local arithmetic.
Coverage: No relevant dependency, diagnostic, retry, timeout, or lifecycle boundary.
Rationale: Identical minimal calculation requires no resilience machinery.
Limits: Hidden authoring-guidance selection was outside allowed inputs.

### Testing — A

Scope: Complete arithmetic B tests and frozen probe.
Coverage: All-positive case, one mixed-input regression, and explicit function-value assignment.
Rationale: Tests detect the requested regression; the function-value assignment also checks the specified signature without adding runtime scope.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PB-T1] `client_test.go:11-15` asserts the mixed result is 7 through the preserved `func([]int) int` type.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same zero-contribution observability and int-range limits as A; all checks pass.

## Explicit substantive comparison

**http-write-replay:** A is stronger on effective supplied-client timeout preservation: its retained timeout and forwarded cancellation hook pass the direct-client control that B fails. Both correctly reuse caller identity and exact payload across application and outer retries; both avoid non-deduplicated application replay, accept exactly 200, close bodies, and keep small cancellable backoffs. A's explicit unknown-outcome text and added context-cause visibility are useful diagnostics; B's documentation/error behavior does not falsely assert that no effect occurred, so the wording difference is not a separate failure. Constant 10ms versus 10/20ms backoff and 4097 versus 4096 rejection-body reads are compliant differences, not a substantive winner by themselves.

Both write candidates have only the original public function and standard-library dependencies, with no breaker. A's additional private transport/body observation preserves errors that Go 1.22 Client.Timeout can otherwise mask and retains legacy cancellation. B's smaller private status observer serves redirect-status detection, but its manual timeout replacement introduces W1. More code is not the reason A is favored; the verified behavior is.

**http-retry-budget:** Both keep eligible application retry policy, overflow handling, fit rejection, one total deadline, cancellation-aware waits, body closure, and result bounds. Both share redirect-result and minimum-version timeout-cause defects. A substantively improves diagnostic bounds and includes tests that assert them; B's shorter errors.Join path preserves ordinary identity but does not bound display text. Equal C+ grades do not erase that difference. A's private bounded-error type is justified by the actual bound; neither adds an exported wrapper, dependency, or breaker. A's real future-date wait test and B's fixed-clock parser test offer different useful evidence; test volume alone is not an improvement.

**pure-control:** Substantive tie. Production code hashes are identical. Both add exactly one focused regression and no retry/deadline/logging/dependency surface. B's function-value assignment makes signature preservation explicit, but A already preserves the signature. No inappropriate client/resilience machinery was selected in the delivered calculation; hidden guidance-selection behavior remains unassessed.

