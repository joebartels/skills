## Correctness & Compatibility — A

Scope: Candidate B, complete bounded code area `candidates/B/{README.md,go.mod,totals.go,totals_test.go}`; original README/source are the behavioral baseline. Go 1.22 minimum, standard library only. Exact file SHA-256 values are in [manifest](../../manifest.json).

Skill used: [go-correctness-and-compatibility](/private/tmp/go-neutral-library-study/review-guidance/go-correctness-and-compatibility/SKILL.md), including [correctness-compatibility-decisions.md](/private/tmp/go-neutral-library-study/review-guidance/go-correctness-and-compatibility/references/correctness-compatibility-decisions.md).

Coverage: Complete ordinary/zero/signed delta behavior, count semantics, coherent publication, simultaneous readers/writers, instance isolation, returned-value ownership, public method types and supported minimum Go version. Illegal copying of a used Totals and nil-pointer use are outside the explicit contract.

Rationale: No actionable in-topic issue is confirmed. The relevant material decisions are assessed and the specific strengths below are verified; this supports A. The direct mutex/ordinary test/build mechanisms are routine correct setup, not two independently verified nonroutine safeguards for A+.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `Add` holds `t.mu` across both `count++` and `sum += delta` (`totals.go:20-23`); Snapshot evaluates both fields while holding that same mutex (`totals.go:28-30`). Return-value evaluation precedes deferred unlock. This establishes one observable update and excludes a mixed publication for supported uncopied instances.
- [G2] Independent race/shuffle runs and the event-based checkpoint probe pass on both toolchains. The latter waits for reader readiness, parks every writer after its first accepted update, asserts intermediate `{4 4}`, releases the remaining updates, joins work and asserts final `{4000 4000}`. It does not infer completion from a guessed sleep or the numeric result.
- [G3] The external-consumer probe passes zero snapshot, zero/positive/negative deltas, exact per-call count, retained old snapshots, caller mutation isolation and independent instance state. Supplied held tests, independently rerun, also verify mixed concurrent deltas `{1000 0}` and separate concurrent instances. The public method types and Snapshot fields are unchanged.

Bad

- None found.

Suggested changes

- None needed.

Limits: Checks are actual independent executions in disposable copies; supplied logs are corroboration only. No unrelated full-repository or shipping-environment claim is made. Race/repeat runs sample schedules and do not prove all interleavings. No compilation failure is credited as behavioral detection; all four mutants compile. The external review/held tests check outcomes and do not become candidate-authored regression coverage. 

Executed evidence: [B-test-go122](../../raw/B-test-go122.json); [B-race-go122](../../raw/B-race-go122.json); [B-race-host](../../raw/B-race-host.json); [B-contract-go122](../../raw/B-contract-go122.json); [B-contract-host](../../raw/B-contract-host.json). Complete command/stdout/stderr/status pairs are under [raw](../../raw); [mutation sensitivity](../../mutation-sensitivity.json) retains intended assertions.
