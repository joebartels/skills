## Correctness & Compatibility — C
Scope: Bounded code-area review of candidate B source snapshot (see source-hashes-before.json), CLI/private runPipeline and delivered tests, judged against original README/source; Go1.22 minimum, actual Go1.22.12 and host Go1.26.5 darwin/arm64. Findings are existing-in-scope in this delivered snapshot.
Coverage: FIFO, early normal completion, failure cancellation, callback joining, error identity and CLI output/status are explicit important contracts. Assessed entire supplied code area relevant to this topic.
Rationale: One contained major error-contract failure supports C; this is the same canonical production cause cross-referenced across relevant topics.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [B/correctness-and-compatibility/G1] Independent FIFO probes at capacity 1/3, held early/failure cancellation checks, accepted-output writer-failure probe and actual host/minimum executable output/status cases pass. Public-facing parser, consumer, host and private signature are preserved.

Bad

- [B/F1][major][existing-in-scope] suppressStopError uses only an active parent and membership in the two exact cancellation sentinels; it has no evidence that a callback error was caused by coordinated stopping. candidates/B/cmd/pipeline/pipeline.go:43,46-49. TestControllerIndependentCancellationClassBeforeStop fails on host, actual Go1.22.12 and race/shuffle. Independent TestReviewIndependentExactBeforeStop reproduces both producer and consumer first exact-Canceled errors being lost before child cancellation. TestReviewIndependentDeadlineAfterStop also loses DeadlineExceeded when the actual child stop is Canceled. — Contained important error-preservation failure in the private pipeline. Both callback roles can report an independent cancellation failure and callers receive success; multiple symptoms share the single classification correction, so this is one major rather than a systemic major.

Suggested changes

- [B/F1] Attach actual child stop/error state to each callback result, retain the first independent failure observed before stopping, and suppress only the exact matching context error generated for coordinated stopping. Preserve other exact cancellation-class errors and every wrapped/joined independent error. Verify using the referenced bounded probes and target-specific mutation checks; preserve accepted effects and source signature.

Limits: Review is confined to the supplied bounded CLI/pipeline snapshots, original contract, candidate tests and supplied held checks. Actual tested platform is darwin/arm64; no cross-platform packaging, service deployment, signals, containers, rollout, remote telemetry or vulnerability-advisory audit is claimed. Arbitrary blocking Reader cancellation is explicitly excluded by README. Race runs cover the exercised paths only. Original and review-mutant checks are distinguished in raw logs. Package-alarm timeouts are not counted as mutation detection. Review is one independent reviewer performing bounded sequential reviews, without author reports/profiles or external repository plans/results. Actual checks: host/minimum ordinary build, test and vet; empty gofmt diff; held tests (A all green, B independent-cancellation-class failure); independent FIFO, cleanup gates, writer failure and cancellation-class probes; actual executable cases; host race/shuffle count=3. See raw/independent-commands.json and raw/followup-commands.json for exact RTK-prefixed argv/status/stdout/stderr.

Skill invoked and read: /private/tmp/go-neutral-pipeline-study/review-guidance/go-correctness-and-compatibility/SKILL.md. References read: /private/tmp/go-neutral-pipeline-study/review-guidance/go-correctness-and-compatibility/references/correctness-compatibility-decisions.md.
