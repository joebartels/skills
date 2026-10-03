## Testing — C
Scope: Bounded code-area review of candidate B source snapshot (see source-hashes-before.json), CLI/private runPipeline and delivered tests, judged against original README/source; Go1.22 minimum, actual Go1.22.12 and host Go1.26.5 darwin/arm64. Findings are existing-in-scope in this delivered snapshot.
Coverage: The delivered tests claim to verify concurrency and error preservation. Their assertions, event gates, isolation and mutation sensitivity require assessment. Assessed entire supplied code area relevant to this topic.
Rationale: One contained major error-contract failure supports C; this is the same canonical production cause cross-referenced across relevant topics.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [B/testing/G1] Finite/FIFO expectations use real values and error identity checks. The independent four-mutation probes each have an original-green target, compiling mutant and intended bounded assertion failure. These review probes assess current behavior but are not credited as candidate-authored regression coverage. B author early-stop/join and wrapped-both-error tests reject the corresponding compiled mutations through intended assertions; the original targets were green.

Bad

- [B/F2][major][existing-in-scope] The suite conflates cancellation-class independent failures with wrapped errors: it never asserts preservation of a bare sentinel returned before any coordinated stop, or an independent exact deadline after a Canceled stop. candidates/B/cmd/pipeline/pipeline_test.go:93-131; the candidate ordinary suite is green on host and actual Go1.22.12, while the supplied held test and independent exact-error probes deterministically fail. The both-error test meaningfully rejects broad errors.Is-based suppression, but that assertion does not cover this current provenance defect. — An explicit important error-preservation contract is unverified for the exact sentinel paths that currently fail. Production repair and test additions are independent changes; the two timing/class examples belong to one assertion-matrix gap.

Suggested changes

- [B/F2] Add bounded tests for each callback returning bare context.Canceled while the child and parent are active, and for an independent context.DeadlineExceeded while cleanup observes Canceled. Assert error identity after joins, and retain the existing wrapped/joined cases. Verify using the referenced bounded probes and target-specific mutation checks; preserve accepted effects and source signature.

Limits: Review is confined to the supplied bounded CLI/pipeline snapshots, original contract, candidate tests and supplied held checks. Actual tested platform is darwin/arm64; no cross-platform packaging, service deployment, signals, containers, rollout, remote telemetry or vulnerability-advisory audit is claimed. Arbitrary blocking Reader cancellation is explicitly excluded by README. Race runs cover the exercised paths only. Original and review-mutant checks are distinguished in raw logs. Package-alarm timeouts are not counted as mutation detection. Review is one independent reviewer performing bounded sequential reviews, without author reports/profiles or external repository plans/results. Actual checks: host/minimum ordinary build, test and vet; empty gofmt diff; held tests (A all green, B independent-cancellation-class failure); independent FIFO, cleanup gates, writer failure and cancellation-class probes; actual executable cases; host race/shuffle count=3. See raw/independent-commands.json and raw/followup-commands.json for exact RTK-prefixed argv/status/stdout/stderr.

Skill invoked and read: /private/tmp/go-neutral-pipeline-study/review-guidance/go-testing/SKILL.md. References read: /private/tmp/go-neutral-pipeline-study/review-guidance/go-testing/references/testing-decisions.md.

Ungraded related production finding: B/F1 (cancellation-class suppression). The Testing grade counts only the independently necessary assertion corrections listed above.
