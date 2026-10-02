## Testing — C-
Scope: Changeset review: neutral original snapshot to candidate-1, the complete supplied README.md, go.mod, stages.go and stages_test.go. This is a sequential outbound HTTP library function, Go minimum 1.22. Snapshot hashes: snapshot.json. No repository or planning context inspected.
Coverage: All author assertions, doubles, channels/timers, cleanup and supplied checks/probes; frozen total-reset sensitivity independently reproduced. Additional stage/combined-error mutations are review diagnostics only.
Rationale: One major stage-budget detection gap plus two independent moderate gaps select C- under the unchanged rubric. The empty test requires a separate test correction from the production guard; discarding simultaneous close failure needs a separate case.
Finding counts: critical=0, major=1, moderate=2, minor=0

Good

- [G1] stages_test.go:109-134 records deadlines from two successful stages; frozen lost-total-budget is killed by its author suite.
- [G2] stages_test.go:150-175 asserts independent read error, standard cancellation and errors.As for a non-comparable custom cause, rather than relying on the reader returning ctx.Err.
- [G3] Failure tests assert retained prefix and one close for the failed body.

Bad

- [T1][major][newly exposed] The author suite never establishes that the stage duration is the request deadline ceiling when it is shorter than the total budget. Evidence: candidate-1/stages_test.go:109-134 uses stage longer than total; all other supplied author stages are one second with immediate work or explicit cancellation. lost-stage-budget compile and author suite exit 0, while stage-ceiling-original exits 0 and lost-stage-budget-diagnostic exits 1.
- [T2][moderate][introduced] The canceled-admission test passes nil endpoints and asserts a cancellation failure, codifying the wrong no-op result and never asking the transport to admit real work. Evidence: candidate-1/stages_test.go:136-147; supplementary.log demonstrates required empty success fails.
- [T3][moderate][newly exposed] Read and close failures are tested independently, so discarding close failure when reading also fails remains undetected. Evidence: candidate-1/stages_test.go:76-78,95-96; lost-close-error-on-read compile and author suite exit 0; review-created TestReviewDualReadCloseErrors passes original and fails mutation.

Suggested changes

- [T1] Add a stage-shorter-than-total deadline assertion plus cooperative body/close timing verification at that stage scope. Owner: Author regression tests. Verify the original defect trigger and negative control documented in findings.json.
- [T2] Separate canceled empty success from canceled nonempty no-request admission, asserting zero calls in both. Owner: Author regression tests. Verify the original defect trigger and negative control documented in findings.json.
- [T3] Add a combined read/close failure case asserting errors.Is for both and retaining only the completed prefix. Owner: Author regression tests. Verify the original defect trigger and negative control documented in findings.json.

Limits: All inspected source is in this neutral packet. Checks ran on disposable copies; exact argv, cwd, environment, status, stdout and stderr are in evidence/*.json and *.log. Current Go is 1.26.5 darwin/arm64. Minimum-version runs executed the explicit Go 1.22.12 binary with CGO_ENABLED=0; default cgo-enabled Go 1.22 linking and other targets were not reverified. Supplied historical results remain distinct from these executed checks. The supplementary empty/custom-cause probes and review-created mutations are post-exposure diagnostics, not frozen-mutation efficacy evidence. No candidate was repaired. No A+ safeguards are claimed.

Skill read: ../../review-skills/go-testing/SKILL.md and its topic reference.

Optional or ungraded related notes

- Optional, ungraded follow-up: the asynchronous read test lacks a deferred cancel and uses an unbuffered started send. Its timeout failure branches do not explicitly cancel-and-join a delayed sender. Normal and raced exercised paths complete; failure-path worker cleanup was not stress-tested.
