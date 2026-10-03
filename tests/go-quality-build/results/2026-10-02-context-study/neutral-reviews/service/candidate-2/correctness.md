## Correctness & Compatibility — C
Scope: Changeset review: neutral original snapshot to candidate-2, the complete supplied README.md, go.mod, stages.go and stages_test.go. This is a sequential outbound HTTP library function, Go minimum 1.22. Snapshot hashes: snapshot.json. No repository or planning context inspected.
Coverage: Ordered GET/200-only acceptance; prefix; one total budget/earlier parent; stage ceiling; canceled admission; read/close context lifetime; independent failures; custom cause classification; empty result; completed success. Supplied and independent targeted checks inspect the full function.
Rationale: One contained major supported error-contract failure: independent stage failure observed with custom cancellation loses errors.Is(context.Canceled), although both other causes survive. No broad/irreversible reach is demonstrated.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [G1] Empty canceled calls succeed, while canceled nonempty calls start no transport; stages_test.go:47-64 and supplementary empty probe verify it.
- [G2] Total/parent, cooperative read and close, stage ceiling, ordered prefix and completed-success probes pass.

Bad

- [P1][major][newly exposed] failedStageError preserves a custom cause but omits the independent standard ctx.Err cancellation classification. Evidence: candidate-2/stages.go:58-62; supplementary.log: independent transport failure and custom stop are retained, but errors.Is(err, context.Canceled) is false.

Suggested changes

- [P1] Join the observed ctx.Err alongside the operation failure and context.Cause before releasing the stage scope. Owner: failedStageError implementation. Verify the original defect trigger and negative control documented in findings.json.

Limits: All inspected source is in this neutral packet. Checks ran on disposable copies; exact argv, cwd, environment, status, stdout and stderr are in evidence/*.json and *.log. Current Go is 1.26.5 darwin/arm64. Minimum-version runs executed the explicit Go 1.22.12 binary with CGO_ENABLED=0; default cgo-enabled Go 1.22 linking and other targets were not reverified. Supplied historical results remain distinct from these executed checks. The supplementary empty/custom-cause probes and review-created mutations are post-exposure diagnostics, not frozen-mutation efficacy evidence. No candidate was repaired. No A+ safeguards are claimed.

Skill read: ../../review-skills/go-correctness-and-compatibility/SKILL.md and its topic reference.
