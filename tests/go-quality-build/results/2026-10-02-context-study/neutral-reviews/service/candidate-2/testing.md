## Testing — C-
Scope: Changeset review: neutral original snapshot to candidate-2, the complete supplied README.md, go.mod, stages.go and stages_test.go. This is a sequential outbound HTTP library function, Go minimum 1.22. Snapshot hashes: snapshot.json. No repository or planning context inspected.
Coverage: All author tests/assertions and doubles, timing, goroutine result synchronization, cleanup, supplied probes/results and frozen lost-total sensitivity. Post-exposure error combination is diagnostic evidence only.
Rationale: Two independently actionable major gaps leave important shared-total and independent-failure cancellation contracts misleadingly verified. No averaging or production/test defect duplication occurs.
Finding counts: critical=0, major=2, moderate=0, minor=0

Good

- [G1] stages_test.go:66-89 and 103-122 assert simultaneous read/close and status/close failures, exact closure count and completed prefix.
- [G2] stages_test.go:124-147 exercises stage cancellation during body reading; 178-191 tests completed success is not revoked by later cancellation.
- [G3] canceled empty/nonempty admission are separated with a shared zero-call assertion.

Bad

- [T1][major][introduced] The author shared-total regression test fails the first stage and never observes a successful transition to stage two. Evidence: candidate-2/stages_test.go:206-219; frozen lost-total-budget mutation compiles, author suite exits 0, and held TestTotalAndEarlierParentDeadline fails; independently repeated same statuses.
- [T2][major][introduced] The custom-cause test reader returns ctx.Err itself, masking omission of cancellation classification from the error-aggregation helper on independent failures. Evidence: candidate-2/stages_test.go:150-175,198-202; author suite exits 0 although supplementary independent transport failure + custom-cause test fails.

Suggested changes

- [T1] Exercise at least two completed stages and assert one total deadline across them, including earlier-parent clipping. Owner: Author regression tests. Verify the original defect trigger and negative control documented in findings.json.
- [T2] Have a failing transport/read/close return an independent sentinel after custom cancellation and assert all three identities: independent failure, custom cause, and context.Canceled. Owner: Author regression tests. Verify the original defect trigger and negative control documented in findings.json.

Limits: All inspected source is in this neutral packet. Checks ran on disposable copies; exact argv, cwd, environment, status, stdout and stderr are in evidence/*.json and *.log. Current Go is 1.26.5 darwin/arm64. Minimum-version runs executed the explicit Go 1.22.12 binary with CGO_ENABLED=0; default cgo-enabled Go 1.22 linking and other targets were not reverified. Supplied historical results remain distinct from these executed checks. The supplementary empty/custom-cause probes and review-created mutations are post-exposure diagnostics, not frozen-mutation efficacy evidence. No candidate was repaired. No A+ safeguards are claimed.

Skill read: ../../review-skills/go-testing/SKILL.md and its topic reference.

Optional or ungraded related notes

- Optional, ungraded follow-up: custom-cancellation test timeout branches do not defer cancel-and-join the worker. Its context-aware reader currently observes the one-second stage timeout; normal/raced paths complete. No failure-path stress claim is made.
