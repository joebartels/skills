# Go Quality Report — C-
Scope: Bounded code-area review of candidate B source snapshot (see source-hashes-before.json), CLI/private runPipeline and delivered tests, judged against original README/source; Go1.22 minimum, actual Go1.22.12 and host Go1.26.5 darwin/arm64. Findings are existing-in-scope in this delivered snapshot.
Coverage: Complete bounded supplied area; all nine topics apply for the contextual reasons below. No omitted material boundary within that area.
Rationale: 2 unique contained major causes select C- under the unchanged rubric. Cross-topic copies are deduplicated, strengths do not cancel defects, and no systemic-major or critical reach is supported. A and B are graded independently.
Unique finding counts: critical=0, major=2, moderate=0, minor=0

## Report cards

| Topic | Grade/state | Assessed chunks and coverage limits | Cards |
| --- | --- | --- | --- |
| Architecture & Design | C | Callback/context/error ownership and cross-boundary completion are the central design decisions; the fixed private API and supplied host must remain usable. | [architecture-and-design](cards/B/architecture-and-design.md) |
| Code Quality & Go Idioms | C | The local context state and error classification need clear, truthful semantics; Go 1.22, formatting and vet also apply. | [code-quality-and-idioms](cards/B/code-quality-and-idioms.md) |
| Correctness & Compatibility | C | FIFO, early normal completion, failure cancellation, callback joining, error identity and CLI output/status are explicit important contracts. | [correctness-and-compatibility](cards/B/correctness-and-compatibility.md) |
| Testing | C | The delivered tests claim to verify concurrency and error preservation. Their assertions, event gates, isolation and mutation sensitivity require assessment. | [testing](cards/B/testing.md) |
| Security | A | The CLI parses caller-controlled integer-line input and emits only integers/diagnostics; resource amplification and input-to-sink paths can be assessed within this local CLI boundary. | [security](cards/B/security.md) |
| Observability & Resilience | C | Cooperative stop, callback failure propagation and caller-visible success/error signals govern failure containment and diagnostics. | [observability-and-resilience](cards/B/observability-and-resilience.md) |
| Performance & Resource Management | A | The code owns a bounded channel and two callbacks; blocked sends/receives, cleanup joins and constant per-call work are relevant resources. | [performance-and-resource-management](cards/B/performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | The documented Go 1.22 minimum and standard-library-only standalone module are a rebuild contract, checked with an actual minimum toolchain. | [dependencies-and-reproducibility](cards/B/dependencies-and-reproducibility.md) |
| Deployment & Operations | A | This artifact is a CLI. Actual binary build, startup/configuration, stdout/stderr, exit codes and finite-input termination are relevant; containers, signals and rollouts are outside this scope. | [deployment-and-operations](cards/B/deployment-and-operations.md) |

## Good

- [B/G1] TestReviewFIFO succeeds at capacities 1 and 3, and actual host/minimum binaries emit the finite integer sequence and exact --take prefix in order.
- [B/G2] Independent early-stop and producer-failure cleanup-gate probes succeed on original B. Its author join test baseline is green, and the compiling skip-early-join mutation fails the intended joined-before-return assertion repeatedly.
- [B/G3] The original wrapped/joined independent-error probe succeeds. Its author both-error test rejects the compiling broad errors.Is suppression mutation through the intended missing-independent-failures assertion.
- [B/G4] Host1.26.5 and actual Go1.22.12 builds/vet/ordinary suites pass in disposable standalone copies with GOWORK=off, GOTOOLCHAIN=local, GOPROXY=off and -mod=readonly; gofmt -d is empty.

## Bad

- [B/F1][major] suppressStopError uses only an active parent and membership in the two exact cancellation sentinels; it has no evidence that a callback error was caused by coordinated stopping. — Contained important error-preservation failure in the private pipeline. Both callback roles can report an independent cancellation failure and callers receive success; multiple symptoms share the single classification correction, so this is one major rather than a systemic major. Evidence: candidates/B/cmd/pipeline/pipeline.go:43,46-49. TestControllerIndependentCancellationClassBeforeStop fails on host, actual Go1.22.12 and race/shuffle. Independent TestReviewIndependentExactBeforeStop reproduces both producer and consumer first exact-Canceled errors being lost before child cancellation. TestReviewIndependentDeadlineAfterStop also loses DeadlineExceeded when the actual child stop is Canceled.
- [B/F2][major] The suite conflates cancellation-class independent failures with wrapped errors: it never asserts preservation of a bare sentinel returned before any coordinated stop, or an independent exact deadline after a Canceled stop. — An explicit important error-preservation contract is unverified for the exact sentinel paths that currently fail. Production repair and test additions are independent changes; the two timing/class examples belong to one assertion-matrix gap. Evidence: candidates/B/cmd/pipeline/pipeline_test.go:93-131; the candidate ordinary suite is green on host and actual Go1.22.12, while the supplied held test and independent exact-error probes deterministically fail. The both-error test meaningfully rejects broad errors.Is-based suppression, but that assertion does not cover this current provenance defect.

## Suggested changes

- [B/F1][priority 1] Attach actual child stop/error state to each callback result, retain the first independent failure observed before stopping, and suppress only the exact matching context error generated for coordinated stopping. Preserve other exact cancellation-class errors and every wrapped/joined independent error. Primary owner: correctness-and-compatibility. Verify: B/host/held-tests, B/minimum/held-tests, B/held-race-shuffle, B/host/independent-probes, B/minimum/independent-probes, B/probe-race-shuffle.
- [B/F2][priority 1] Add bounded tests for each callback returning bare context.Canceled while the child and parent are active, and for an independent context.DeadlineExceeded while cleanup observes Canceled. Assert error identity after joins, and retain the existing wrapped/joined cases. Primary owner: testing. Verify: B/host/ordinary-tests, B/minimum/ordinary-tests, B/host/held-tests, B/host/independent-probes, B/mutation/broad-cancel-suppression/ordinary-suite.

## Limits

Review is confined to the supplied bounded CLI/pipeline snapshots, original contract, candidate tests and supplied held checks. Actual tested platform is darwin/arm64; no cross-platform packaging, service deployment, signals, containers, rollout, remote telemetry or vulnerability-advisory audit is claimed. Arbitrary blocking Reader cancellation is explicitly excluded by README. Race runs cover the exercised paths only. Original and review-mutant checks are distinguished in raw logs. Package-alarm timeouts are not counted as mutation detection. Review is one independent reviewer performing bounded sequential reviews, without author reports/profiles or external repository plans/results.

Actual checks are independently executed; supplied checks-A/B.json are context only. Held executable child builds use GOQUALITY_GO pointing to the actual matching host/minimum toolchain. The additional deadline-after-stop probe extends the supplied held observation and is kept distinct from it. Four review-created mutation targets are documented in raw logs/patches; no unavailable original frozen author-specific mutation target is inferred or replaced. Timeouts are preserved without intended-assertion detection credit.

Details: [manifest](manifest.json), [findings](findings.json), [ledger](ledger-B.json), [calculator result](grade-B.json), [applicability](applicability-B.json), [raw commands](raw/independent-commands.json), [narrowed join follow-up](raw/followup-commands.json), [probe source](probes/review_probe_test.go).
