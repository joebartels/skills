## Observability & Resilience — C

Scope: Candidate 1. Changeset review of the provided immutable original to candidate snapshot; library plus supplied CLI host. Only run.go, run_test.go and newly added cmd/finalize/main_test.go differ. README.md, go.mod and cmd/finalize/main.go are unchanged contract/caller context. Exact SHA-256 snapshots are in ../source-snapshots.json. Declared Go 1.22; independently executed Go 1.22.12 and Go 1.26.5, darwin/arm64, GOTOOLCHAIN=local, GOWORK=off.

Coverage: Cancellation/failure diagnosis, independent finalization deadline, synchronous drain before return/exit, and required receipt failure handling. All material obligations within this topic boundary were assessed; scope does not expand to absent unrelated infrastructure.

Rationale: A contained major diagnostic-contract defect loses an independently actionable apply failure during cancellation. Required bounded drain and receipt failure signaling otherwise work on the tested host.

Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [G1] The finalization scope is separate from caller cancellation, has the supplied timeout and is invoked synchronously at [run.go:30](/private/tmp/go-context-outcome-20261002/transfer/candidate-1/run.go:30). Unchanged probes independently show a live bounded context after parent cancellation, no early Run return, accepted-prefix receipt persistence and exit 2.
- [G2] Cancellation resources are released at [run.go:32](/private/tmp/go-context-outcome-20261002/transfer/candidate-1/run.go:32). The supplementary captured-scope diagnostic observes context.Canceled after successful completion on both compilers.
- [G3] The command reports failure on stderr and exits 2 after cooperative finalization returns at [cmd/finalize/main.go:61](/private/tmp/go-context-outcome-20261002/transfer/candidate-1/cmd/finalize/main.go:61). Actual interrupted and write-failure checks verify the owned exit behavior.

Bad

- [F1][major][introduced] Run treats errors.Is(applyErr, ctxErr) as proof that applyErr contains no independent failure and discards the entire apply error tree. candidate-1/run.go:16-21. Reviewer diagnostic returns errors.Join(workFailure, context.Canceled) after custom parent cancellation, then an independent finalization failure: accepted=1, finalCalls=1, canceled=true, parentCause=true, workFailure=false, finalFailure=true. Fails independently on actual Go 1.22.12 and 1.26.5; authored/unchanged supplied probes otherwise pass.

Suggested changes

- [F1] Always preserve the full apply error while also retaining observed cancellation classification/cause; deduplicate only if it can preserve the complete original error tree. Add the joined cancellation plus independent callback-failure case. Owner: Run error aggregation.

Limits: Review was sequential within this neutral packet; no mapping, repository design/results, author skills, or other arms were inspected. Authored sources and supplied probes were not modified. Windows interruption is explicitly skipped by both authored suites and the supplied probe; execution here covers darwin/arm64 only. Cooperative nonnil callbacks, nonnil parent context and positive budget are contractual preconditions. No crash, hostile-input, durability/atomicity, packaging/CI, or arbitrary-platform promises were invented. Supplementary reviewer diagnostics are not frozen efficacy tests. 

Checks: Independent checks: host authored suite passes; minimum authored suite fails only the interrupted stderr assertion; standalone -mod=readonly build and vet pass on both compilers; gofmt -d is empty; supplied unchanged contract probes pass on both compilers, including real interrupted child status/output/receipt and real receipt-write failure; host -race -shuffle=on -count=3 suite with unchanged probes passes. Supplementary combined-error diagnostic fails on both compilers while finalization-scope release passes. Exact commands/environment and raw stdout/stderr are stored in ../candidate-1-*.json; supplied facts remain separately labeled in ../../provided-checks-1.json.

Skill read: ../../review-skills/go-observability-and-resilience/SKILL.md and its decision reference; packet report grading/reporting/orchestration references were also read.
