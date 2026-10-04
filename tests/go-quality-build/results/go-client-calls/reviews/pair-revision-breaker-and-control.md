# Neutral pair review: revision-breaker-recovery and revision-pure-control

The breaker pair favors **A on three verified contract differences**: completion-time health neutrality, supplied redirect policy, and bounded diagnostics. The local arithmetic pair is a **substantive tie**. These conclusions use neutral labels only. No arm disclosure, candidate guidance, other trial, author report, arm map, campaign summary, or Git diff was inspected.

## Boundary and evidence

This is a code-area review of each supplied implementation against its original README, not a historical changeset review. Locations beginning `BR/` resolve under `/private/tmp/go-client-campaign/reviews/pairs/revision-breaker-recovery/`; `PC/` resolves under `/private/tmp/go-client-campaign/reviews/pairs/revision-pure-control/`. Findings are existing-in-scope. Complete source, tests, modules, both READMEs, raw `checks.json`, hashes, and all supplied probes were inspected. Every file listed in each `source-hashes.json` matched its SHA-256 hash.

Applied the unchanged Correctness & Compatibility, Observability & Resilience, and Testing review skills and their decision references. Consulted primary Go 1.22 HTTP client source and pinned gobreaker/v2 v2.4.0 source. The HTTP source explicitly defines `Do` as following client policy, including redirects. One application `Do` invocation may contain multiple redirect exchanges. Gobreaker's exclusion counters preserve consecutive history and release half-open admission; its completion generation check rejects stale results. Neither property requires a custom breaker framework.

The Testing grades assess candidate-owned regression suites. Frozen probes and independent reproductions establish their gaps and production consequences; they are not credited as tests authored by a candidate. Production roots BR-B1/B2/B3 are counted once per applicable production topic. BT IDs are independently actionable test corrections. Counts across topics must not be added together.

| Pair / label | Correctness & Compatibility | Observability & Resilience | Testing |
| --- | --- | --- | --- |
| revision-breaker-recovery / A | A | A | A |
| revision-breaker-recovery / B | C- | C- | C- |
| revision-pure-control / A | A | Not applicable | A |
| revision-pure-control / B | A | Not applicable | A |

### Executed verification

All checks ran in disposable copies at `/private/tmp/pair-revision-breaker-control-scratch`, with `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`. Supplied candidates and probes were never edited. Go 1.26.5 was the host toolchain; `/private/tmp/go-client-campaign/tools/go/bin/go` was Go 1.22.12. Local listener execution was authorized.

- Candidate-only copies: `go test -ldflags=-linkmode=external -count=1 -timeout=20s ./...` and `go vet ./...` passed for all four implementations.
- All supplied probes copied unchanged into scratch: the same host test command and the Go 1.22.12 equivalent passed for breaker A and both pure versions; breaker B failed the supplied policy, operation-deadline 503, canceled-200 history/recovery, and diagnostic-bound probes. Both breaker copies were also tested with `go test -race -ldflags=-linkmode=external -count=1 -timeout=20s ./...`; A passed, B had the same assertion failures, and neither run reported a data race.
- Independent probes: `go test -ldflags=-linkmode=external -run TestIndependent -v -count=1 -timeout=20s ./...` passed for A and failed for B on allowed redirect behavior, expired completed-200 health neutrality, and diagnostic limits. Both retained error causes; A also retained transport error type access through a bounded wrapper.

Raw command arrays, exits, and outputs are saved in the adjacent `pair-revision-breaker-and-control-reproductions/` directory: `candidate-checks.json`, `frozen-probe-checks.json`, and `additional-probe-checks.json`. `additional_review_test.go.txt` contains the independent reproduction source. `raw-findings.json` freezes the findings and grades before any disclosure. Commands above are the underlying Go invocations; shell execution was prefixed with `rtk proxy`.

The supplied raw controller results agree with the fresh runs: both candidates' own suites pass, A's private breaker checks pass, B's private breaker checks fail on the same cases, and both control versions pass host/minimum checks and vet. The additional body-health probe passes on both implementations, confirming failed 200 body reads are neutral in both consecutive history and recovery. This preserves an important tie rather than attributing every error path to B's completion bug.

## revision-breaker-recovery / A

### Correctness & Compatibility — A

Scope: `BR/A/source/`, reusable HTTP client library, preserved `New` and `(*Client).Fetch` signatures, Go 1.22.0 and pinned gobreaker/v2 v2.4.0.

Coverage: Ordinary 200 and non-200 status behavior; result/error bounds; body ownership; supplied client policy; operation and earlier-parent budgets; disabled mode; dependency scope; eligible history; neutral cancellation, deadline and body-read completions; recovery admission/release; reopened cooldown; and old generations. Exact API function types are asserted in `client_test.go:19`. Oversize handling is discussed below rather than assigned an unsupported mandatory policy.

Rationale: No verified actionable defect. Relevant strengths are demonstrated by source tracing and passing frozen/independent checks. Grade A records correct behavior in the assessed scope; additional machinery or test volume does not warrant a higher grade.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [BR-AG1] `BR/A/source/client.go:69` checks the **operation context after completion**, before any health result. Canceled or expired 200 and 503 completions are neutral. Frozen canceled-200 probes and independent expired-200 probes preserve history and leave recovery pending. Returning already-read bytes with nil error during cancellation is not itself a finding: the required distinction is neutral breaker accounting.
- [BR-AG2] `BR/A/source/client.go:83` scopes breakers by lowercased scheme and host authority per client. `Allow`/completion and pinned generation checks preserve independent dependencies, single recovery admission, excluded release, and stale-generation isolation. Scope, recovery, and old-completion tests pass.
- [BR-AG3] `BR/A/source/client.go:101` invokes the supplied client directly once per admitted Fetch. The independent redirect reproduction returns `ok` through two permitted exchanges and one caller policy callback. `client.go:116` bounds long diagnostics to 4096 bytes while retaining `errors.Is` and `errors.As` access to underlying causes/types.

Bad

- None found.

Suggested changes

- None needed for the verified contract.

Limits: Passing checks are bounded evidence, not proof of all schedules. No live external dependency was exercised. A caller-supplied transport/body must cooperate with the operation context for wall-clock cancellation; custom transport probes returning a completed result after stopping are used to verify health classification. Nil-client support is extra behavior outside the supplied nonnil contract and was not graded.

### Observability & Resilience — A

Scope: The same client area, operation time budget, failure classification, breaker containment, and caller-facing diagnostics.

Coverage: Cancellation throughout headers/body, caller and client timeout policy, 503 eligibility, neutral non-503 failures, single recovery concurrency, stale generations, and bounded cause-preserving errors. No owned service lifecycle, telemetry exporter, retry loop, or deployment control is present or required.

Rationale: No actionable issue. Both failure containment and diagnostic access work in the assessed conditions. No service instrumentation requirement is inferred for this library.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [BR-AG4] `BR/A/source/client.go:52` sets one total operation context for the request and body lifecycle. `client_test.go:173` exercises operation, earlier parent, and borrowed client timeouts at a real HTTP boundary.
- [BR-AG5] `BR/A/source/client.go:71` and `client.go:90` distinguish caller stopping from dependency failure, preserve excluded history, and use the existing library to contain failures and admit one recovery. Neutral reads and release behavior pass all supplied probes.
- [BR-AG6] `BR/A/source/client.go:124` keeps the emitted diagnostic bounded and the cause inspectable; independent long-body and long-transport errors return exactly 4096 diagnostic bytes with retained causes.

Bad

- None found.

Suggested changes

- None needed.

Limits: No workload-level SLO or deployment context was supplied; no logs, metrics, traces, jitter, or service shutdown machinery is demanded. The library does not retry and uses caller-owned HTTP policy.

### Testing — A

Scope: `BR/A/source/client_test.go` and its candidate-owned verification, with frozen probes/reproductions used as independent review evidence.

Coverage: API types, realistic HTTP success/redirect/deadline boundaries, exact status rejection, 4096-byte boundary, long cause-preserving diagnostics, body closure, borrowed identity, enabled/disabled behavior, neutral canceled 200/503, dependency history, recovery release including read failures, and late 200/503 generations. Concurrent tests use channels and bounded joins; cooldown waits measure the specified elapsed interval.

Rationale: No verified actionable regression-test issue. Assertions check downstream state and request counts rather than merely successful calls or internal implementation shape. The candidate suite, minimum-version evidence, independent probes, and exercised race run agree.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [BR-ATG1] `BR/A/source/client_test.go:52`, `:119`, and `:142` assert status/results, policy callback, cause identity, diagnostic length, and owned-body closure. These tests expose meaningful API regressions.
- [BR-ATG2] `BR/A/source/client_test.go:229`, `:348`, and `:457` observe history/admission after neutral and stale completions. Cleanup releases and joins blocked goroutines, and stale-generation testing checks that the old context has not already expired (`:518`), reducing a masking risk.

Bad

- None found.

Suggested changes

- None needed for the assessed candidate suite. Porting the independent local-expiry completion and allowed-redirect cases would be a useful optional extension, without changing this grade.

Limits: The authored suite does not enumerate every operation-expiry/outcome combination. Frozen probes assess those combinations directly. Timing tests have broad bounded tolerances and cannot prove scheduler-independent timing; fresh host/minimum/race executions passed.

## revision-breaker-recovery / B

### Correctness & Compatibility — C-

Scope: `BR/B/source/`, reusable HTTP client library, the same original README, Go 1.22.0 and pinned dependency. Public signatures remain source-compatible (`api_test.go:10`).

Coverage: The same material behavior as A, including every health completion class, enabled/disabled policy, result and diagnostic bounds, and concurrency generation transitions.

Rationale: One contained major issue and two independent moderate issues select C-. BR-B1 breaks the explicit health/recovery contract; BR-B2 changes ordinary caller HTTP policy; BR-B3 breaches the explicit diagnostic size bound. The major reach is this client breaker workflow, not a claimed systemic or critical failure.

Finding counts: critical=0, major=1, moderate=2, minor=0

Good

- [BR-BG1] `BR/B/source/client.go:61` and `:77` preserve per-dependency/client scope, 3-failure trip, one recovery admission, 100ms cooldown, exclusions of ordinary non-503 errors, and library generation isolation. Scope, failed-read health, concurrent recovery/release, and stale-completion probes pass.
- [BR-BG2] `BR/B/source/client.go:49` propagates the total operation context into HTTP and body reading. `client.go:102` closes owned bodies; disabled mode has one application `Do` and no retry. B retains underlying cause identity and transport error type access, despite unbounded display text.

Bad

- [BR-B1][major][existing-in-scope] **Health accounting misses canceled/expired successful completions and local expiry on 503.** `BR/B/source/client.go:62`–`:67` only decorates `statusError` and uses the parent `ctx.Err()`, while `:82`–`:85` accepts every nil error as healthy. After two eligible 503s, a locally expired 503 incorrectly opens the breaker; a completed canceled/expired 200 clears the two-failure history. In recovery, that same 200 closes the breaker, so the next eligible 503 does not reopen immediately. Frozen `TestPairOperationDeadline503IsNeutral`, both canceled-200 probes, and independent `TestIndependentDeadline200IsNeutral` reproduce the consequences. These symptoms share the missing completion-level classifier; they are counted once.
- [BR-B2][moderate][existing-in-scope] **Fetch bypasses borrowed redirect policy.** `BR/B/source/client.go:95`–`:98` copies the supplied client and replaces `CheckRedirect`. A supplied no-follow callback is never invoked; a supplied follow policy returns HTTP 302 instead of the final 200. Both disabled and enabled reproductions show zero policy callbacks. Leaving the caller's original struct unchanged preserves its storage but does not preserve its behavior during Fetch. One application call does not authorize suppressing policy-driven redirect exchanges.
- [BR-B3][moderate][existing-in-scope] **Error diagnostics exceed 4096 bytes.** Parse errors at `BR/B/source/client.go:56`, transport errors at `:100`, and body errors at `:106` return arbitrary underlying text. Supplied probes produced 8245, 8233, and 8192 diagnostic bytes in both modes. Independent long-cause tests reproduced 7528/7500 bytes. Results remain bounded; diagnostics do not.

Suggested changes

- [BR-B1] Classify health after every request completion using the **derived operation context**, independently of the returned caller error. If stopping has occurred, report an excluded breaker outcome even when already-read data and nil error are returned. Otherwise only an eligible 503 fails and a successful 200 establishes health. Reuse the pinned two-step admission API or an equivalent existing-library mechanism; verify history and recovery with the supplied expiry/canceled-success probes.
- [BR-B2] Use the supplied client's policy for the single application `Do` call. Remove the forced redirect callback and verify both `ErrUseLastResponse` and permitted-follow policies, including caller callback counts and final response.
- [BR-B3] Bound displayed errors on every return path while preserving unwrapping. Verify `len(err.Error()) <= 4096`, `errors.Is`, and `errors.As` for long parsing, transport, and body failures in both modes.

Limits: A canceled completed 200 returning data/nil is not counted as a caller-result bug; its health effect is the defect. Failed 200 body reads already return an error and remain neutral. No unsupported nil-client panic is counted. The oversize result choice is discussed separately.

### Observability & Resilience — C-

Scope: The same client failure handling, budgets, borrowed policy, breaker containment, and diagnostic boundary.

Coverage: Headers/body cancellation, health eligibility and recovery, dependency separation, disabled mode, stale generations, and diagnostic/cause access. No service-owned telemetry or shutdown boundary applies.

Rationale: BR-B1 is a contained major failure of the stated containment/recovery policy. BR-B2 is moderate because caller-configured external-call behavior is bypassed; BR-B3 is moderate because error output can exceed the explicitly bounded diagnostic surface. One major plus two moderates selects C-.

Finding counts: critical=0, major=1, moderate=2, minor=0

Good

- [BR-BG3] The derived request budget and real header/body tests contain cooperative HTTP slowness; there is no application retry multiplication.
- [BR-BG4] Existing gobreaker exclusions correctly release neutral recovery admission and retain consecutive history for 400/500/transport/read errors; both body-health contexts and stale-generation checks pass.

Bad

- [BR-B1][major][existing-in-scope] Completion classification can open on local expiry or recover while the caller is stopping, defeating the reliability contract; evidence and locations are in the correctness section.
- [BR-B2][moderate][existing-in-scope] Borrowed redirect policy is bypassed during Fetch, changing external-call outcomes in both modes; same root and evidence as above.
- [BR-B3][moderate][existing-in-scope] Unbounded error text breaks the diagnostic containment requirement. Cause access works, so the finding is about the emitted size rather than lost identity.

Suggested changes

- [BR-B1] Apply completion-level operation-context neutrality before success/failure accounting.
- [BR-B2] Honor the borrowed HTTP policy on the single application call.
- [BR-B3] Add a bounded diagnostic representation with cause access on all error paths.

Limits: These overlap the same production roots as Correctness & Compatibility. No requirement for application metrics, logging, tracing, process globals, retries, or a custom breaker is inferred.

### Testing — C-

Scope: `BR/B/source/client_test.go`, `api_test.go`, and `contracts_test.go`; external frozen probes and scratch reproductions demonstrate what the candidate-owned passing suite misses.

Coverage: Candidate tests cover success, public API types, short read-error identity/body ownership, request context lifecycle, result limits, operation headers/body budgets, dependency scope, ordinary exclusions, canceled 503, concurrency admission/release, recovery, and stale generations. The failed 200 body-read health path is independently verified as correct by supplied probes.

Rationale: Two major candidate-suite issues and one moderate diagnostic gap select C-. One test actively encodes the opposite of the supplied-client contract; the completion-time neutrality sub-contract is effectively unverified for successful stopping and local expiry. These require independent test changes even after fixing production behavior.

Finding counts: critical=0, major=2, moderate=1, minor=0

Good

- [BR-BTG1] `BR/B/source/api_test.go:10` asserts public function types externally. `contracts_test.go:56` verifies partial bytes, cause identity, owned body closure, context values, and cancellation after Fetch.
- [BR-BTG2] `BR/B/source/contracts_test.go:332` and `:412` use channel coordination, bounded waits/cleanup, and observable request counts for recovery and stale generations. These cover real state-machine risks and pass the exercised race check apart from unrelated failing assertions.

Bad

- [BR-BT1][major][existing-in-scope] **The redirect test enforces the wrong contract.** `BR/B/source/contracts_test.go:115` configures a supplied follow callback, then `:134`–`:136` requires HTTP 302 and zero callback invocations. It detects a correct implementation as a regression and allows BR-B2 to pass. The real borrowed-client contract is important and explicitly supplied; checking the original client later (`:141`) does not validate Fetch's policy.
- [BR-BT2][major][existing-in-scope] **Completion-time neutrality is missing at the important health boundary.** `contracts_test.go:153` tests time budgets without subsequent breaker state assertions; `:230` uses already-canceled entries; `:296` covers canceled 503 only; `:332` cancellation completes with an error. None catches nil-error 200 completion during stopping or the distinction between parent and local operation expiry. The candidate suite passes while frozen/independent tests demonstrate false history reset, false recovery, and early trip.
- [BR-BT3][moderate][existing-in-scope] **Bound tests check returned bytes but omit diagnostic bounds.** `BR/B/source/contracts_test.go:91` verifies truncated result/status strings, and `:56` uses a short read error. There is no long parse/transport/body error length assertion, so all six supplied diagnostic failures pass the candidate suite.

Suggested changes

- [BR-BT1] Assert supplied redirect callbacks and final policy-selected response in both modes; distinguish one application call from HTTP exchanges. Remove the suppression expectation.
- [BR-BT2] Add observable history and recovery tests for completed canceled 200, local-expiry 503, and local-expiry completed 200. Keep valid returned bytes/nil error permitted where the body completed; assert neutral accounting and exclusive replacement admission.
- [BR-BT3] Add long-error cases on every diagnostic path with byte-bound and cause/type assertions; these fail the current implementation and independently guard the correction.

Limits: The external probes already detect these defects; the findings concern the candidate-owned suite and its misleading pass. No race was reported by the fresh race run. This grade does not claim all B tests are weak or award A merely for having more tests.

## Breaker pair: substantive comparison and surface

A is substantively better on BR-B1, BR-B2, and BR-B3. The distinction is demonstrated behavior, not comments, test counts, or grades. Both correctly retain per-scheme/authority dependency scope, per-client state, ordinary excluded history, failed-200-body neutrality, single recovery admission, excluded release, 100ms reopen behavior for eligible failure, stale-generation isolation, owned body closure, bounded operation context, and no application retry. Both keep the same Go directive, pinned module requirement, and module sums; neither introduces a dependency or custom framework. A's two-step library use addresses the need to separate caller result from health result; its diagnostic wrapper addresses a stated byte bound. Those are justified surface.

Disabled A does not allocate a breaker map (`client.go:43`) and follows a direct bounded request path. Disabled B also bypasses breaker execution, but allocates an unused map (`client.go:42`) and copies/replaces HTTP policy on every request. The unused map and `enabled`/map duplication are small optional simplifications, not independently consequential defects; the policy override is BR-B2. A's private sentinel errors are package-level identities, not shared mutable breaker/process state.

For an oversized 200, A reads at most 4097 bytes and returns a bounded overflow error, whereas B returns the first 4096 bytes successfully. The frozen observation deliberately asserts only the result bound. The README states a byte bound without defining truncation versus rejection; this review treats overflow as a local bounded-result error, analogous to a body-read error, and does not invent a required truncation contract. **Neither oversized policy is credited as a pair improvement.** If the intended meaning of “200 succeeds” is specifically “always return a truncated successful 200 despite overflow,” that additional policy would require A to change; no such clarification was supplied. Both preserve exact status acceptance at the HTTP-status layer.

## revision-pure-control / A

### Correctness & Compatibility — A

Scope: `PC/A/source/client.go`, deterministic local `SumPositive([]int) int`, Go 1.22.0.

Coverage: Strict positive selection, mixed negative/zero/positive input, existing all-positive behavior, zero initialization/empty iteration by inspection, and exact public function type.

Rationale: No actionable issue; the minimal loop implements the original README and fresh host/minimum tests pass.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PC-AG1] `PC/A/source/client.go:6` uses `v > 0`; `client_test.go:12` retains the public function type and the mixed-input regression yields 6.

Bad

- None found.

Suggested changes

- None needed.

Limits: No overflow semantics beyond normal Go `int` arithmetic were specified; none are invented. No external operation or concurrency is present.

### Observability & Resilience — Not applicable

Scope: The same deterministic local arithmetic function.

Coverage: Complete implementation and module inspected for external operations, fallible boundaries, operational signals, and lifecycle ownership; none are present.

Rationale: No relevant failure-containment or operational-signal decision. A timeout, breaker, metric, retry, logger, or dependency would add unrequested surface.

Limits: This applies to the supplied local function only, without assuming a surrounding service.

### Testing — A

Scope: `PC/A/source/client_test.go` and the supplied frozen arithmetic probe.

Coverage: Existing all-positive behavior and one focused mixed-input regression, public function-value type, host/minimum toolchains, and vet.

Rationale: No actionable test issue. The focused regression detects negative contributions and preserves the requested simple test scope.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PC-ATG1] `PC/A/source/client_test.go:11` includes negatives, zero, and multiple positives with an exact expected sum; fresh tests and the frozen probe pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: No fuzz, race, or timing test is justified by this deterministic loop. Zero addition is numerically indistinguishable from exclusion, but the inspected predicate is strictly positive.

## revision-pure-control / B

### Correctness & Compatibility — A

Scope: `PC/B/source/client.go`, deterministic local `SumPositive([]int) int`, Go 1.22.0.

Coverage: Strict positive selection, mixed negative/zero/positive input, existing all-positive behavior, zero initialization/empty iteration by inspection, and public signature.

Rationale: No actionable issue. The production file is byte-identical to A; its focused mixed-input regression and frozen probe pass on host/minimum Go.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PC-BG1] `PC/B/source/client.go:6` selects only strictly positive values; `client_test.go:12` verifies the mixed-input result 5.

Bad

- None found.

Suggested changes

- None needed.

Limits: Native `int` overflow policy is unspecified and unchanged. No I/O, concurrency, or context is implicated.

### Observability & Resilience — Not applicable

Scope: The same deterministic local arithmetic function.

Coverage: Complete function and module inspected; no external failure, operational signal, resilience mechanism, or service lifecycle exists.

Rationale: No applicable operational or failure-handling decision. The implementation correctly keeps this calculation simple.

Limits: No surrounding service was supplied or assumed.

### Testing — A

Scope: `PC/B/source/client_test.go` and the supplied frozen arithmetic probe.

Coverage: Existing all-positive case, one focused mixed-input regression, host/minimum execution, and vet.

Rationale: No actionable issue. The expected sum fails if negative values contribute. The absence of A's extra function-value assertion does not produce a substantive regression-detection loss for this unchanged signature.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [PC-BTG1] `PC/B/source/client_test.go:11` directly checks mixed non-positive/positive input against 5; all executed arithmetic checks pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: No broad-input framework or concurrency test is necessary for the stated repair. Strict exclusion of zero is established by implementation inspection; adding zero cannot change the numerical sum.

## Control pair comparison

**Tie.** Both production files and Go modules are identical; both retain a small local loop, add one focused regression, and add no dependency, timeout, breaker, context, or operational surface. A's function-value assertion supplies a small additional compile-time check, while B calls the unchanged signature directly. That difference does not establish a substantive improvement. Both satisfy the independent frozen probe and the same supported minimum version.
