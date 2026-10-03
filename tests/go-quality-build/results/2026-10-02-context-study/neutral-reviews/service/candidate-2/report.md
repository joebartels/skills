# Go Quality Report — C-

Scope: Changeset review: neutral original snapshot to candidate-2, the complete supplied README.md, go.mod, stages.go and stages_test.go. This is a sequential outbound HTTP library function, Go minimum 1.22. Snapshot hashes: snapshot.json. No repository or planning context inspected.

Coverage: Complete supplied implementation and test changeset; all nine topics assessed for applicability. Seven applicable topics have cards; Security and Deployment are justified Not applicable only within this bounded changeset. No material coverage gap is hidden as N/A.

Rationale: Three contained independent major causes (one implementation error-contract defect and two independently necessary test corrections) select C-. Shared implementation causes are counted once across topic cards; independently necessary regression-test changes remain distinct. The unchanged absolute calculator validates counts only. No averages or clean-topic dilution.

Unique finding counts: critical=0, major=3, moderate=0, minor=0

## Report cards

| Topic | Grade/state | Assessed chunks and coverage limits | Cards |
| --- | --- | --- | --- |
| Architecture & Design | C | API/dependency shape, explicit lifecycle ownership, success boundary and inspectable cancellation/error contract. All production paths inspected. | [architecture](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/architecture.md) |
| Code Quality & Go Idioms | C | All production/test Go, error flow, manual cancel paths, nil empty success, documentation and Go 1.22 support; gofmt and vet executed. Repeated cancel calls are clear here and are not penalized as an optional style preference. | [code-quality](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/code-quality.md) |
| Correctness & Compatibility | C | Ordered GET/200-only acceptance; prefix; one total budget/earlier parent; stage ceiling; canceled admission; read/close context lifetime; independent failures; custom cause classification; empty result; completed success. Supplied and independent targeted checks inspect the full function. | [correctness](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/correctness.md) |
| Testing | C- | All author tests/assertions and doubles, timing, goroutine result synchronization, cleanup, supplied probes/results and frozen lost-total sensitivity. Post-exposure error combination is diagnostic evidence only. | [testing](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/testing.md) |
| Security | Not applicable | No changed security-relevant trust or authorization boundary is established in this bounded changeset. These edits derive budgets and aggregate errors using the supplied client, finite HTTP/HTTPS endpoints and positive durations; they do not alter destinations, redirects, TLS policy, credentials or privilege. | [security](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/security.md) |
| Observability & Resilience | C | Total/stage deadline ownership, active cancellation, failure decisions, prefix/error containment and synchronous stage completion. Telemetry/retries/process shutdown are not owned by this function. | [resilience](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/resilience.md) |
| Performance & Resource Management | A | Every body acquisition/close exit, read buffering, per-stage timer cancellation, total cancellation, client reuse and sequential in-flight work. Workload byte/concurrency profiles are unknown; no performance improvement or fixed response-size bound is claimed. | [resources](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/resources.md) |
| Dependencies & Reproducibility | A | Compared original/candidate go.mod and all imports; standalone GOWORK=off/GOTOOLCHAIN=local/GOPROXY=off build/test/vet checks; independently executed Go 1.22.12 tests with CGO_ENABLED=0 and supplied original minimum-version check records. No external sums, vendor, local replace, generation or native inputs are required by these edits. | [reproducibility](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/reproducibility.md) |
| Deployment & Operations | Not applicable | No release workflow, process startup/shutdown, packaging, CI enforcement or deployed runtime configuration is changed by this library function implementation. The explicit Go minimum build obligation is assessed under Dependencies & Reproducibility. | [deployment](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/deployment.md) |

## Good

- [G1] Canceled empty lists succeed and canceled nonempty lists make no request.
- [G2] Author tests assert simultaneous read/close and status/close errors, exact owned-body closure and prefix retention.
- [G3] Independent source/runtime checks verify inherited total/stage deadlines, close lifetime and completed success after later cancellation. Evidence: [executed checks](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/evidence) and topic cards.

## Bad

- [P1][major] failedStageError preserves a custom cause but omits the independent standard ctx.Err cancellation classification. [stages.go:58](/private/tmp/go-context-outcome-20261002/service/candidate-2/stages.go:58). candidate-2/stages.go:58-62; supplementary.log: independent transport failure and custom stop are retained, but errors.Is(err, context.Canceled) is false.
- [T1][major] The author shared-total regression test fails the first stage and never observes a successful transition to stage two. [stages_test.go:206](/private/tmp/go-context-outcome-20261002/service/candidate-2/stages_test.go:206). candidate-2/stages_test.go:206-219; frozen lost-total-budget mutation compiles, author suite exits 0, and held TestTotalAndEarlierParentDeadline fails; independently repeated same statuses.
- [T2][major] The custom-cause test reader returns ctx.Err itself, masking omission of cancellation classification from the error-aggregation helper on independent failures. [stages_test.go:150](/private/tmp/go-context-outcome-20261002/service/candidate-2/stages_test.go:150). candidate-2/stages_test.go:150-175,198-202; author suite exits 0 although supplementary independent transport failure + custom-cause test fails.

## Suggested changes

- [P1] Join the observed ctx.Err alongside the operation failure and context.Cause before releasing the stage scope. Owner: failedStageError implementation. Verification: the original contract trigger or negative control recorded in [findings](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/findings.json).
- [T1] Exercise at least two completed stages and assert one total deadline across them, including earlier-parent clipping. Owner: Author regression tests. Verification: the original contract trigger or negative control recorded in [findings](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/findings.json).
- [T2] Have a failing transport/read/close return an independent sentinel after custom cancellation and assert all three identities: independent failure, custom cause, and context.Canceled. Owner: Author regression tests. Verification: the original contract trigger or negative control recorded in [findings](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/findings.json).

## Limits

This is an independent packet-only reviewer pass over the specified small changeset, reviewed sequentially without delegated workers. Candidate identity/arm mapping was not inspected or inferred. Candidate source and supplied probes remain unchanged, verified against snapshot hashes. Own checks use disposable scratch copies.

Executed baseline: Go 1.26.5 build, author test, vet and gofmt all exit 0 for both candidates. Contract/lifetime/real local HTTP probes and repeated race/shuffle runs exit 0. Go 1.22.12 author and contract tests exit 0 with GOTOOLCHAIN=local, GOWORK=off, GOPROXY=off and CGO_ENABLED=0. Default cgo-enabled Go 1.22 host linking and other targets were not reverified. Supplied earlier checks remain labeled supplied, not claimed as executed here.

Post-exposure supplementary tests fail on one explicit contract edge per candidate; their defects are supported by the original README but these tests cannot count as frozen efficacy evidence. Only lost-total-budget is a supplied frozen mutation. The removed-stage and lost-combined-close experiments are reviewer-created diagnostic negative controls. The original held stage watchdog also passes the removed-stage mutation: this limit is preserved, and the supplied study inputs were not repaired. See [probe validity](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/contract-validity.md).

Cancellation/stoppage conclusions apply only to the exercised cooperative transport/body/local-handler boundaries. Host application deployment/security policy and non-supplied callers are unknown and outside this changeset; N/A does not certify those broader areas. No benchmark or all-platform claim is made. Failure-path test-worker cleanup was inspected but not broadly stress-tested; optional follow-ups stay ungraded in Testing.

Details: [manifest](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/manifest.json), [deduplicated finding index](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/findings.json), [ledger](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/ledger.json), [calculator output](/private/tmp/go-context-outcome-20261002/service/reviews/candidate-2/grade-output.json).
