## Correctness & Compatibility — B
Scope: Code-area review of packet-01/candidate/process.go and its author tests against original/README.md. The paired original/process.go establishes the requested evolution; this is a local library, module example.com/process, go 1.22, unchanged exported signature. Unrelated legacy behavior is excluded.
Coverage: Traced active/empty input, already-canceled nonempty/empty input, ordered success/failure, accepted-prefix counts, between-job cancellation, callback-error/cancellation composition, non-comparable causes, deadlines, last-callback success, later cancellation, and cooperative external cancellation. Independently exercised the relevant cases in disposable copies on darwin/arm64, Go 1.26.5 and Go 1.22.12; see evidence.md and checks.json.
Rationale: One moderate supported contract mismatch selects B. Its reach is the empty-input admission boundary: the caller receives success instead of cancellation; no callbacks or accepted jobs are lost. No major, systemic, or critical reach is established.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [CC-G1] candidate/process.go:14-20 retains independent callback errors and increments accepted only after a nil result. Author TestSequentialProgress and reviewer TestReviewerDeadlineAndIndependentFailure verify the accepted prefix and retained identities.
- [CC-G2] candidate/process.go:16,25-26 uses errors.Join without comparing arbitrary error values. The unchanged candidate passes TestReviewerCauseReturnedByCallback with the same non-comparable sliceCause as both callback error and cancellation cause on Go 1.22.12.
- [CC-G3] candidate/process.go:20-22 returns success after the last accepted callback without a final cancellation check. The author coincident-cancellation test and reviewer success-then-cancellation probe verify this explicit completion rule.

Bad

- [F1][moderate][existing-in-scope] Call-level cancellation is checked only inside the job loop (candidate/process.go:10-12), so nil or nonnil empty jobs bypass it and reach the nil error at line 22. original/README.md requires already-canceled calls to return classification/cause and grants empty-input success only otherwise. TestReviewerCanceledEmpty independently returns accepted=0, calls=0, err=<nil> for both empty forms on Go 1.26.5 and Go 1.22.12. A caller inspecting errors.Is/As cannot distinguish canceled admission from successful empty completion. The original also lacked this behavior; this is an incomplete requested evolution, not a regression claim.

Suggested changes

- [F1] Check ctx.Err at call admission before entering the loop, retaining the between-job check and the absence of a final check. Return cancellationError(ctx) for already-canceled empty calls. Verify nil and nonnil empty jobs with context.Canceled/custom cause, active empty success, and last-callback completion success. Primary remediation owner: Correctness & Compatibility.

Limits: Exact executed commands/results are in checks.json; reviewer diagnostics are preserved in reviewer_contract_test.go. Supplied held checks corroborate F1 but their source is unavailable and was not treated as inspected code. No unsupported nil context/callback behavior, external consumer release matrix, or unpromised platform was graded. The race/shuffle check exercises the listed reviewer/author paths, not all possible callbacks.

