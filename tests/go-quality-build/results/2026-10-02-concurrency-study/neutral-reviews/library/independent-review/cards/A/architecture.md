## Architecture & Design — A

Scope: Candidate A, complete bounded code area `candidates/A/{README.md,go.mod,totals.go,totals_test.go}`; original README/source are the behavioral baseline. Go 1.22 minimum, standard library only. Exact file SHA-256 values are in [manifest](../../manifest.json).

Skill used: [go-architecture-and-design](/private/tmp/go-neutral-library-study/review-guidance/go-architecture-and-design/SKILL.md), including [design-decisions.md](/private/tmp/go-neutral-library-study/review-guidance/go-architecture-and-design/references/design-decisions.md).

Coverage: Complete supplied module/API: Snapshot, Totals, Add, Snapshot; per-instance state, construction, copy constraint and synchronous ownership. No persistence, transport, remote error or background-lifecycle boundary exists here.

Rationale: No actionable in-topic issue is confirmed. The relevant material decisions are assessed and the specific strengths below are verified; this supports A. The direct mutex/ordinary test/build mechanisms are routine correct setup, not two independently verified nonroutine safeguards for A+.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `totals.go:12-15` owns one mutex and two counters per Totals. `Add` commits synchronously, and Snapshot publishes a two-scalar value. There is no global state, background worker, constructor or new public abstraction. This is the smallest direct ownership model needed by the README.
- [G2] The external-consumer signature assignments and zero-value/value-ownership assertions in `contract_probe_test.go` compile and pass on both toolchains; no caller lifecycle setup is introduced. Global-state mutations are rejected by the original instance tests on both toolchains.

Bad

- None found.

Suggested changes

- None needed.

Limits: Checks are actual independent executions in disposable copies; supplied logs are corroboration only. No unrelated full-repository or shipping-environment claim is made. Race/repeat runs sample schedules and do not prove all interleavings. No compilation failure is credited as behavioral detection; all four mutants compile. The external review/held tests check outcomes and do not become candidate-authored regression coverage. 

Executed evidence: [A-test-go122](../../raw/A-test-go122.json); [A-race-go122](../../raw/A-race-go122.json); [A-race-host](../../raw/A-race-host.json); [A-contract-go122](../../raw/A-contract-go122.json); [A-contract-host](../../raw/A-contract-host.json). Complete command/stdout/stderr/status pairs are under [raw](../../raw); [mutation sensitivity](../../mutation-sensitivity.json) retains intended assertions.
