## Correctness & Compatibility — C

Scope: Candidate 1. Changeset review of the provided immutable original to candidate snapshot; library plus supplied CLI host. Only run.go, run_test.go and newly added cmd/finalize/main_test.go differ. README.md, go.mod and cmd/finalize/main.go are unchanged contract/caller context. Exact SHA-256 snapshots are in ../source-snapshots.json. Declared Go 1.22; independently executed Go 1.22.12 and Go 1.26.5, darwin/arm64, GOTOOLCHAIN=local, GOWORK=off.

Coverage: Sequential acceptance, stop conditions, empty/unaccepted work, late successful apply, cancellation classification/cause, independent failures, finalization effects, and child status/output/receipt. All material obligations within this topic boundary were assessed; scope does not expand to absent unrelated infrastructure.

Rationale: One major breach of the explicit inspectable-independent-failures contract is confirmed on a contained library path. No severe broad or irreversible reach is demonstrated.

Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [G1] The finalization scope is separate from caller cancellation, has the supplied timeout and is invoked synchronously at [run.go:30](/private/tmp/go-context-outcome-20261002/transfer/candidate-1/run.go:30). Unchanged probes independently show a live bounded context after parent cancellation, no early Run return, accepted-prefix receipt persistence and exit 2.
- [G2] Cancellation resources are released at [run.go:32](/private/tmp/go-context-outcome-20261002/transfer/candidate-1/run.go:32). The supplementary captured-scope diagnostic observes context.Canceled after successful completion on both compilers.
- [G3] Both supplied unchanged process probes pass with actual Go 1.22.12 and 1.26.5; they check accepted stdout, exit 2, post-interruption receipt contents and absence of receipt after a real write failure. Successful-last-apply tests confirm no late cancellation veto.

Bad

- [F1][major][introduced] Run treats errors.Is(applyErr, ctxErr) as proof that applyErr contains no independent failure and discards the entire apply error tree. candidate-1/run.go:16-21. Reviewer diagnostic returns errors.Join(workFailure, context.Canceled) after custom parent cancellation, then an independent finalization failure: accepted=1, finalCalls=1, canceled=true, parentCause=true, workFailure=false, finalFailure=true. Fails independently on actual Go 1.22.12 and 1.26.5; authored/unchanged supplied probes otherwise pass.

Suggested changes

- [F1] Always preserve the full apply error while also retaining observed cancellation classification/cause; deduplicate only if it can preserve the complete original error tree. Add the joined cancellation plus independent callback-failure case. Owner: Run error aggregation.

Limits: Review was sequential within this neutral packet; no mapping, repository design/results, author skills, or other arms were inspected. Authored sources and supplied probes were not modified. Windows interruption is explicitly skipped by both authored suites and the supplied probe; execution here covers darwin/arm64 only. Cooperative nonnil callbacks, nonnil parent context and positive budget are contractual preconditions. No crash, hostile-input, durability/atomicity, packaging/CI, or arbitrary-platform promises were invented. Supplementary reviewer diagnostics are not frozen efficacy tests. Context value inheritance by the finalizer is not promised; candidate 2 using Background instead of WithoutCancel is not graded as a defect. The actual minimum-version candidate 1 test assertion failure is assessed separately in Testing, not mistaken for a runtime receipt/status failure.

Checks: Independent checks: host authored suite passes; minimum authored suite fails only the interrupted stderr assertion; standalone -mod=readonly build and vet pass on both compilers; gofmt -d is empty; supplied unchanged contract probes pass on both compilers, including real interrupted child status/output/receipt and real receipt-write failure; host -race -shuffle=on -count=3 suite with unchanged probes passes. Supplementary combined-error diagnostic fails on both compilers while finalization-scope release passes. Exact commands/environment and raw stdout/stderr are stored in ../candidate-1-*.json; supplied facts remain separately labeled in ../../provided-checks-1.json.

Skill read: ../../review-skills/go-correctness-and-compatibility/SKILL.md and its decision reference; packet report grading/reporting/orchestration references were also read.
