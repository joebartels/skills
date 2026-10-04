# Independent neutral pair review: breaker recovery

The review assessed A and B independently against their identical original README. Labels remain neutral; no arm mapping or other campaign material was inspected. Raw findings were frozen in [raw-findings.json](pair-breaker-recovery-reproductions/raw-findings.json) before any disclosure. The three specified, unchanged review skills and their decision references were applied. No candidate or supplied probe was edited.

B is substantively stronger on health accounting, bounded diagnostics and borrowed-client policy. A has a confirmed breaker-health defect, unbounded errors and a redirect-policy mismatch. Both preserve important scope, admission and generation safeguards. B still has a demonstrated regression-detection gap for failed HTTP 200 body reads; its current implementation handles that case correctly.

| Label | Correctness & Compatibility | Observability & Resilience | Testing |
| --- | --- | --- | --- |
| A | C-: 1 major, 2 moderate | C: 1 major, 1 moderate | C-: 1 major, 2 moderate |
| B | A+: no findings | A+: no findings | B: 1 moderate |

Counts are topic-specific. Shared production IDs are cross-referenced, not additional defects; independently necessary test corrections have separate IDs. No grade or test-count target was used.

## Boundary, contract interpretation and evidence

Scope is the complete supplied library code area, not a known base/head diff. Findings are **existing-in-scope**. Both modules declare Go 1.22.0 and pin `github.com/sony/gobreaker/v2 v2.4.0`; their module files and sums match. All supplied source hashes matched before verification and again after review. The only runtime implementation sources consulted outside the packet were the pinned gobreaker sources.

“One application request” is interpreted as one application `http.Client.Do` operation with no application retries, consistent with the simultaneous requirement to preserve the supplied client and its policy. Both make one such call. A additionally prevents HTTP redirects by replacing the callback on a copied client; B delegates redirects to the supplied client. A strict one-wire-request requirement would conflict with arbitrary supplied redirect policies and should be clarified if intended. The redirect finding below uses a caller policy that already stops redirects, so honoring that policy needs only one transport request.

The README specifies a result bound but does not explicitly choose truncation versus oversize rejection or partial bytes on read failure. A returns the first 4096 bytes successfully for an oversized HTTP 200 and can return partial bytes with a read error. B rejects oversized HTTP 200 and returns nil bytes on read failure. Those policies were verified and recorded, without assuming either is an additional violation. HTTP status checks apply to responses returned successfully by `Do`; underlying request, transport and body failures retain their causes.

The packet's raw controller results show both candidate suites passing host tests, vet, race tests and Go 1.22.12 tests/vet. The original frozen probes passed for both. Supplemental probes failed A's canceled-success accounting and diagnostic bounds; B passed. Fresh scratch runs reproduced those results, plus the independent concerns described below.

Fresh commands used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`, each prefixed by `rtk proxy env`, in disposable A/B copies. Exact commands, exits, toolchain versions and raw output are in [run-notes.md](pair-breaker-recovery-reproductions/run-notes.md) and the adjacent logs. Host Go was 1.26.5 on darwin/arm64; minimum Go was 1.22.12. Tests used `-ldflags=-linkmode=external` on this host.

| Fresh check | A | B |
| --- | --- | --- |
| Candidate tests + both frozen probe sets + initial equal reviewer probes, host | Fails A-F1/F2/F3 probes; candidate tests and original probes pass | Pass |
| Same scope, race | Same assertion failures; no race diagnostic | Pass; no race diagnostic |
| Same scope, Go 1.22.12 | Same assertion failures | Pass |
| Host vet | Pass | Pass |
| Additional equal failed-body-read health probe | Pass | Pass |

## A — Correctness & Compatibility — C-

Scope: Complete `A/source` library, supplied README contract, Go 1.22.0 module semantics and pinned gobreaker v2.4.0. No previous implementation or upgrade history is supplied.

Coverage: Traced ordinary HTTP 200, exact non-200 errors, body bounds/errors/closure, request construction errors, caller cancellation and operation deadlines, health accounting, consecutive failures, independent dependencies, disabled behavior, one recovery admission, exclusions, reopening and stale completions. Inspected all code/tests/modules; executed equal frozen and reviewer probes in scratch copies on both toolchains and with race detection. Supplied redirect policy was explicitly exercised.

Rationale: A-F1 breaks the important, contained breaker-health contract; A-F2 and A-F3 are two independent moderate contract mismatches. This selects C-. The failures affect this client's breaker and error/policy boundaries, with no evidence of critical or systemic reach.

Finding counts: critical=0, major=1, moderate=2, minor=0

Good

- [A-G1] `A/source/client.go:61`, `:71` scopes separate breakers by scheme and authority under a mutex. `TestBreakerDependencyScope` verifies paths/query share history while scheme, host, port and separate Client instances remain independent.
- [A-G2] `A/source/client.go:79` configures one half-open admission, a 100ms open interval and three consecutive eligible failures. Ordinary 400/500/206, transport errors and cancellation errors preserve failure history. Recovery tests verify excluded admission is released without closing, replacement 503 reopens, and genuine 200 restores closed behavior.
- [A-G3] Pinned gobreaker's generation check ignores old completions. `TestBreakerIgnoresLateGenerationCompletions` covers both late 200 and late 503 after recovery; the original frozen generation probe also passes.
- [A-G4] `A/source/client.go:49`, `:102`, `:106` keeps one request context through body consumption, closes owned bodies and returns at most 4096 bytes. Real HTTP tests cover header/body stalls and earlier parents; its read-error test verifies a retained cause, partial bytes, context values and closure before scope cancellation.

Bad

- [A-F1][major][existing-in-scope] `A/source/client.go:62`–`:68`, `:82`–`:87` use the caller error directly as the breaker outcome and only inspect cancellation when an HTTP status error exists. Successful body reading can cancel the parent and still return `"ok", nil`; that nil error resets eligible failure history or closes a half-open breaker, although caller stopping must be neutral. The supplied supplemental probes reproduced both: after two eligible 503s and a canceled 200, a fourth eligible request contacted the dependency; after canceled recovery 200, replacement 503 did not reopen. The separate deadline probe shows another manifestation: A checks parent `ctx.Err()` instead of the 250ms request context, so a completed 503 at its operation deadline is counted and prematurely opens the breaker. These are one incomplete health-classification root cause. Owner: Correctness & Compatibility; also assessed under Resilience.
- [A-F2][moderate][existing-in-scope] `A/source/client.go:51`–`:59`, `:98`–`:106` return construction, transport and read errors without bounding diagnostic text. Supplemental parse/transport/body probes produced 8245/8233/8192-byte messages in both enabled and disabled mode, exceeding the explicit 4096-byte contract. Existing causes remain inspectable, but message size is uncontrolled. Owner: Correctness & Compatibility; shared with Resilience.
- [A-F3][moderate][existing-in-scope] `A/source/client.go:95`–`:98` copies the borrowed client and replaces `CheckRedirect`, so Fetch bypasses caller policy despite preserving the original object's fields. The equal reviewer probe supplied `CheckRedirect` returning `http.ErrUseLastResponse`: both candidates made one transport call and returned HTTP 302, but A invoked the supplied callback zero times, B once, in both modes. Caller redirect validation or observation is therefore bypassed. This finding follows the policy-preservation interpretation stated above; preserving the original fields alone does not preserve their application to Fetch.

Suggested changes

- [A-F1] Classify health independently of the caller's returned bytes/error, after body consumption and before accounting. If the operation context is stopped, report exclusion even for a nil caller error; otherwise count only eligible 503, treat completed 200 without read failure as success, and exclude other errors. Preserve caller results. Verify both consecutive history and recovery, including successful reading during cancellation and completed 503 at the operation deadline.
- [A-F2] Bound every outward diagnostic to 4096 bytes with a wrapper that retains the original cause through `Unwrap`. Verify parse, transport and body errors in both modes, plus `errors.Is`/`As` through truncation.
- [A-F3] Execute Fetch through the borrowed client policy under the agreed application-request interpretation. Verify the provided redirect callback executes while its stop policy keeps transport calls at one, and retain original client configuration/ownership.

Limits: No prior API implementation, deployment or production traffic was supplied. The public function types compile in A's external test. Context-aware standard HTTP boundaries respect the budget; arbitrary custom transports/readers that ignore context cannot be forced to stop by this wrapper. Exact fresh commands and assertion output are retained in the reproduction directory. Redirect and oversized-body interpretation limits are stated above.

## A — Observability & Resilience — C

Scope: A's library failure boundary: time budget, breaker isolation/recovery, error diagnostics and response ownership; Go 1.22.0 and pinned gobreaker v2.4.0.

Coverage: Assessed dependency slowdown/failure, parent cancellation, total budget through headers/body, exclusions, healthy versus caller-visible outcomes, single recovery admission and stale generations. Reviewed emitted errors as the library's diagnostic interface. Service logs, metrics, traces, deployment probes and shutdown are outside this supplied library contract; their absence is not a finding.

Rationale: A-F1 is one major failure-containment contract violation; A-F2 is one moderate diagnostic-bound violation. This selects C. A-F3 remains a Correctness policy finding; the reproduction establishes callback bypass but does not separately establish a resilience consequence requiring another count.

Finding counts: critical=0, major=1, moderate=1, minor=0

Good

- [A-RG1] The shared 250ms context reaches `Do` and body reads, and earlier parents remain effective. Real header/body-stall tests pass on both versions, limiting normal slow HTTP work.
- [A-RG2] Per-dependency state, one recovery admission, exclusion release and library generation checks contain ordinary outages without blocking unrelated authorities. Original frozen recovery, scope and stale-generation probes pass.
- [A-RG3] Exact status errors are small and clear, and read errors retain their causes. Owned bodies close on both returned status and read paths.

Bad

- [A-F1][major][existing-in-scope] The caller-result/health conflation at `A/source/client.go:62`–`:68`, `:82`–`:87` allows canceled successful reads to falsely prove dependency health and admits new work before genuine recovery. Checking only the parent for completed 503 also falsely counts an operation-budget expiration. See the verified call counts in Correctness and `A-host.log`.
- [A-F2][moderate][existing-in-scope] The unbounded errors at `A/source/client.go:55`, `:59`, `:100`, `:106` violate the stated diagnostic bound at the library's failure interface. Parse and transport text can include a long endpoint; read failures can carry arbitrarily long causes. All three supplied cases exceed the bound in both modes.

Suggested changes

- [A-F1] Report neutral dependency health whenever the operation context is stopped, irrespective of completed caller bytes. Keep exclusion distinct from health success so release permits another single probe without closing or resetting history.
- [A-F2] Apply one bounded outward-error wrapper consistently, preserving cause inspection. Retest long diagnostics and cancellation identities rather than replacing errors with lossy strings.

Limits: The checked scope has no service-owned telemetry or lifecycle contract to grade. No remote production workload was run. No race diagnostic appeared in executed concurrency paths; this is not a proof of all possible interleavings. Command/result details are in the shared evidence record.

## A — Testing — C-

Scope: All supplied `A/source` tests and their detection of the README behavior. Equal frozen probes and independent reviewer probes are supporting evidence, not credited as A-authored regression coverage.

Coverage: Examined assertions, API checks, real HTTP boundaries, fakes, asynchronous synchronization, cleanup, elapsed-time waits, scope/history/recovery/generation coverage and error identity checks. Candidate-only raw checks pass across host/minimum/race/vet. The combined independent suite exposes the findings below.

Rationale: A-T1 leaves an important neutral-health contract effectively unverified. A-T2 and A-T3 are independently actionable moderate gaps: diagnostic bounds lack assertions, and a policy test asserts contrary behavior. One major plus two moderate selects C-. Test fixes require independent corrections in addition to production fixes; they are not double-counted production symptoms.

Finding counts: critical=0, major=1, moderate=2, minor=0

Good

- [A-TG1] External function-value checks preserve New/Fetch signatures. Real HTTP tests exercise header/body deadlines and redirection; tracked bodies verify closure and partial-read cause handling.
- [A-TG2] Recovery and stale-completion tests use channels to establish admission/order and bounded result waits; cleanup cancels and joins owned goroutines. The 110ms sleeps exercise the specified 100ms cooldown, rather than guessing asynchronous progress.
- [A-TG3] Tests inspect contact counts and state rejection, not only error strings. They cover independent schemes/authorities/ports, intervening exclusions, genuine success resets, and late 200/503 completions.

Bad

- [A-T1][major][existing-in-scope] `A/source/contracts_test.go:230`, `:296`, `:332` cover already-canceled calls, cancellation errors and canceled 503; they omit nil-error 200 completion while the caller is stopping and health classification at the operation deadline. Consequently all candidate checks pass while A-F1 breaks both eligible history and recovery. The supplied supplemental and equal reviewer probes detect these important cases.
- [A-T2][moderate][existing-in-scope] `A/source/contracts_test.go:56`, `:91` verify response-result limits and a short read cause, but do not assert long diagnostic bounds for request construction, transport and body errors. A-F2 survives the candidate suite; the equal long-error probes fail in every enabled/disabled case.
- [A-T3][moderate][existing-in-scope] `A/source/contracts_test.go:115`–`:149` labels a test as preserving the client but explicitly expects zero supplied redirect callbacks at line 135. Calling the original client later proves its fields were not mutated, not that Fetch used its policy. This misleading assertion entrenches A-F3 and must change independently of the production path.

Suggested changes

- [A-T1] Add table cases for successful body reads that cancel the parent and a completed 503 at operation expiration. Assert caller result and neutral history separately; cover both closed and half-open states with contact-count checks.
- [A-T2] Add long parse/transport/body error cases in both modes with message-length and retained-cause assertions.
- [A-T3] Assert that Fetch invokes a supplied stop-redirect policy once and makes one transport request; separately assert configuration remains borrowed and unchanged.

Limits: Candidate checks are controller-provided raw evidence; fresh replay used scratch copies with the supplied probes appended. No candidate test was rewritten. Timing bounds passed on this host, without a general flakiness claim. Production findings A-F1/F2/F3 are related context; this grade counts their independently demonstrated verification gaps.

## B — Correctness & Compatibility — A+

Scope: Complete `B/source` library against the identical README; Go 1.22.0 module and pinned gobreaker v2.4.0. No earlier version is supplied.

Coverage: Inspected all code, modules and tests and traced every outward completion: construction/admission rejection; transport errors; non-200 statuses; successful, oversized and failed body reads; caller stopping with completed 200/503; operation deadline; body closure. Evaluated failure history, scope, disabled behavior, one recovery admission, exclusion release and old generations. Equal scratch probes passed host/minimum/race checks and the additional body-health cases.

Rationale: No actionable correctness defect was verified in the interpreted contract. Two independent verified safeguards justify A+: (1) separate caller result and health classification prevents canceled completed successes or stopped 503s from changing dependency health; (2) generation-aware admission/completion prevents old success/failure from changing a new breaker generation or releasing its active recovery slot. These control distinct completion-classification and concurrency-generation risks.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-G1] `B/source/client.go:66`–`:79` reports dependency health independently of outward data/error and checks the operation context after body completion. Canceled 200 still returns completed `"ok", nil` while preserving eligible history and leaving recovery unproven. Supplemental body-cancellation probes and the equal deadline-503 probe pass.
- [B-G2] `B/source/client.go:61`, `:77`, `:88`–`:92` use the pinned two-step breaker with one recovery admission, three eligible failures and 100ms reopening. Its generation token is checked before completion updates. Candidate late-failure, late-success-during-active-recovery and original frozen probes verify rejection and stale isolation.
- [B-G3] `B/source/client.go:98`–`:115` closes returned bodies, handles exact 200 status and bounds results. Failed body reads are excluded from health; the equal new probe verifies both preserved history and released, unproven recovery.
- [B-G4] `B/source/client.go:126`–`:134` caps messages at 4096 bytes and retains the original cause through `Unwrap`. All long parse/transport/body supplemental cases pass in both modes; candidate tests assert `errors.Is` for long transport and body causes.
- [B-G5] `B/source/client.go:39`–`:43`, `:99` keeps the supplied client and executes its policy directly. Disabled construction leaves breaker storage nil. The equal redirect probe verifies the supplied stop callback executes once in both modes, while no retry loop or added framework exists.

Bad

- None found.

Suggested changes

- None needed for verified production correctness; B-T1 below concerns regression detection.

Limits: The oversized/partial-result and application-request interpretations above delimit this grade. Default supplied redirect handling may perform additional protocol requests; there is no application retry. Invalid nil clients are outside the explicitly nonnil contract. No historical consumer behavior beyond the supplied function APIs was available. Context-ignoring custom I/O and panics are not promised containment behavior. Race checks cover executed paths only.

## B — Observability & Resilience — A+

Scope: B's library budget, breaker failure containment and error diagnostic interface, Go 1.22.0 and pinned gobreaker v2.4.0.

Coverage: Reviewed dependency outages, scoped breaker state, health exclusions, recovery contention/release, stale outcomes, parent and operation timeouts through body reading, bounded errors with retained causes, and supplied HTTP policy. No library-owned logging, metrics, tracing or service lifecycle is required by this packet.

Rationale: No actionable resilience finding was verified. Two independent safeguards justify A+: neutral health accounting prevents caller stopping from falsely restoring a dependency, while a single propagated total budget bounds slow HTTP work through body reading under earlier-parent cancellation. The first is proven by canceled successful body reads and stopped 503 results; the second by real body stalls under both total and earlier-parent deadlines. Generation protection and bounded causal diagnostics add verified strengths.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-RG1] Explicit neutral health tokens and post-consumption context checks preserve failure history and half-open state for caller stopping, HTTP 400, other statuses and transport/read errors. A completed caller success does not become a false dependency-health signal.
- [B-RG2] The one operation context at `B/source/client.go:49` reaches `Do` and response reading. Real HTTP tests and the independently reproduced operation deadline confirm the total budget is not restarted across stages; earlier parents win.
- [B-RG3] Per-dependency one-probe admission, exclusion release and stale-generation checks contain outage traffic. The original frozen cancellation, held-admission, generation and dependency-isolation probes pass.
- [B-RG4] Bounded outward errors retain causes, so callers can inspect cancellation/transport/read failures without accepting oversized error messages. Successful bodies and rejected status bodies close; redirect-policy errors rely on `Do`'s already-closed body, verified by a single-close test.

Bad

- None found.

Suggested changes

- None needed in this verified resilience scope.

Limits: No deployed SLO, telemetry pipeline, workload measurements or service lifecycle was supplied or needed for this local library judgment. The wrapper depends on standard context-aware HTTP I/O. No race was reported in exercised paths. Exact command outcomes and timing limits appear in the evidence record.

## B — Testing — B

Scope: Complete supplied `B/source` test area; equal frozen/reviewer probes and one disposable mutation assess regression detection without changing B.

Coverage: Checked assertions for status/result/diagnostic bounds, retained causes, supplied redirect policy, body ownership, request budget, scope/counting, successful cancellation neutrality, exclusive admission/recovery, and late success/failure. Reviewed fake versus real HTTP boundaries, cleanup and channel synchronization. Candidate raw checks and fresh combined checks pass on host/minimum/race. A focused mutation verifies the remaining body-read health gap.

Rationale: B-T1 is one moderate missing normal-use boundary case. The breaker contract broadly has meaningful tests; the gap is limited to health accounting after an HTTP 200 body-read error, rather than leaving all exclusions or recovery unverified. One moderate and no other counted issue selects B. B's current production behavior passes the additional health probe.

Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [B-TG1] `B/source/policy_test.go:85`, `:101`, `:116`, `:135` test long diagnostics with cause retention, body closure, redirect-error single ownership and active borrowed policy. These assertions distinguish callback preservation from merely leaving the original object unchanged.
- [B-TG2] `B/source/policy_test.go:245`, `:366` distinguish caller success from neutral canceled health and assert both eligible-history preservation and recovery release/reopen. Concurrent probes are held with channels; contact counts confirm rejected calls do not hit the dependency.
- [B-TG3] `B/source/policy_test.go:161`, `:440`, `:484` exercise real HTTP body deadlines and earlier parents, late failure after recovery, and late success while a current recovery slot is occupied. Async helpers cancel and wait during cleanup with bounded failure paths.
- [B-TG4] Compile-time function-value checks preserve the APIs. Scope/disabled and status tests check actual transport counts, expected bytes/errors, read bounds and body closes.

Bad

- [B-T1][moderate][existing-in-scope] `B/source/policy_test.go:101`–`:114` tests body-read failure only with breakers disabled. Enabled exclusions at `:245`–`:288` and recovery at `:366`–`:438` omit a non-canceled HTTP 200 with a read error. The consequential branch is `B/source/client.go:73`: in a separate disposable B mutation, removing only `&& err == nil` makes failed 200 reads reset eligible history and establish recovery. All candidate tests, both supplied probe sets and the initial reviewer probes still pass; the equal additional `TestPairBodyReadErrorIsNeutral` fails both history and recovery on the mutation and passes unmodified A/B. This proves a detection gap, not a B production defect.

Suggested changes

- [B-T1] Add an enabled HTTP 200 read-error case after two eligible failures and as a half-open recovery admission. Assert the read cause still reaches the caller, history is preserved, exclusion releases admission, and the next eligible recovery failure immediately reopens. The saved equal probe supplies a minimal example; the one-condition mutation should fail the maintained suite afterward.

Limits: Only this targeted mutation was evaluated; no broad mutation-score claim is made. Its raw passing existing-suite and failing added-probe logs are retained. Supplied frozen tests are reviewer evidence rather than candidate-owned tests. Race detection did not report a race in executed paths, and the minimum-version run passed before the additional targeted body-health check, which was run on host Go.

## Substantive comparison and unnecessary surface

| Contract/risk | A | B | Assessment |
| --- | --- | --- | --- |
| Completed caller 200 during stopping | Returned success is also counted as healthy | Caller success retained, health neutral | B improvement: prevents failure resets and false recovery |
| Completed 503 at operation deadline | Checks only parent; may count | Checks operation context; excluded | B improvement within the same health root cause |
| Eligible history, 400/ordinary-error exclusion | Correct for covered non-success completions | Correct | Tie outside A-F1 |
| Recovery exclusivity and excluded release | Correct for ordinary exclusions | Correct, including canceled completed success | Both preserve one admission; B closes A's successful-cancellation hole |
| Old-generation results | Correct | Correct | Tie; B also tests late success against an active recovery slot |
| Scheme/authority dependency scope | Correct | Correct | Tie; paths/query share state, dependencies remain independent |
| Failed HTTP 200 body-read health | Correct | Correct | Production tie; B's supplied test gap is mutation-confirmed |
| Total HTTP budget and body ownership | Correct at context-aware HTTP boundary | Correct at context-aware HTTP boundary | Tie; A explicitly tests both header and body stalls |
| Bounded errors and retained causes | Causes retained; messages unbounded | Messages bounded; causes retained | B improvement, verified for all outward error classes |
| Borrowed HTTP policy | Original fields unchanged, callback bypassed in Fetch | Supplied policy executes | B improvement under the stated interpretation |
| Disabled state | Empty breaker map allocated but never used | No breaker map allocated | B slightly simpler; A has no disabled breaker behavior defect |
| API/toolchain/dependency surface | Original API, Go directive and pinned graph | Same | Tie; neither adds a framework or dependency |

B's two-step completion callback, internal health tokens and bounded diagnostic wrapper earn their added implementation surface by enforcing distinct required policies. The unexported error tokens do not create shared process breaker state. A's unused disabled map allocation is a small optional simplification, not a graded contract defect. Both stay within the pinned library and standard packages, preserve caller ownership and avoid retry machinery.

A's additional external API check, header-stall coverage and detailed body/context ownership test are real strengths. B's improvements arise from verified behavior and assertions; its larger or differently arranged test set is not itself an improvement. Oversize rejection versus truncation remains a recorded policy difference rather than a winner. The review stops at these frozen neutral findings and does not infer treatment identity.
