## Testing — B

Scope: Candidate A, complete bounded code area `candidates/A/{README.md,go.mod,totals.go,totals_test.go}`; original README/source are the behavioral baseline. Go 1.22 minimum, standard library only. Exact file SHA-256 values are in [manifest](../../manifest.json).

Skill used: [go-testing](/private/tmp/go-neutral-library-study/review-guidance/go-testing/SKILL.md), including [testing-decisions.md](/private/tmp/go-neutral-library-study/review-guidance/go-testing/references/testing-decisions.md).

Coverage: Complete candidate assertions and test goroutine lifecycles; actual baseline/race/repeat/shuffle execution; four independent compiled behavioral mutations per candidate on two toolchains; added reviewer-owned event-based/held checks for the production contract are explicitly separate from candidate tests.

Rationale: One confirmed moderate test-lifecycle defect selects B. Other important assertions have demonstrated sensitivity; the timeout still rejects lost updates, so this is contained fragility rather than effective absence of core verification.

Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [G1] Both original candidate test suites reject the split-update mutation using their actual `incoherent snapshot` assertion on host and Go 1.22.12. The mutant protects each field access with the same mutex but unlocks between the two changes; these non-race runs demonstrate multi-field invariant sensitivity rather than relying on compilation failure or race output.
- [G2] Both suites reject process-global state using their instance assertions and reject sign normalization using the sequential exact `{Count:2, Sum:1}` assertion on both toolchains. The compiler accepts all four behavior-changing mutants; no build error is counted as detection.

Bad

- [F1][moderate][existing-in-scope] `totals_test.go:48-58` makes reader termination depend on expected Count==8000. If coherent updates are lost, writer completion cannot stop the reader; `snapshotsDone.Wait()` blocks before the final count assertion. The compiled dropped-unit-update mutation times out after 2s on both toolchains and in the whole Go 1.22 suite. This is [canonical A-F1](../../findings.json), owned by the test lifecycle.

Suggested changes

- [F1] Stop the reader on a writer-completion event, bound lifecycle waits, join the reader, then assert final Count/Sum. Retain the coherent dropped-update mutation and require a prompt wrong-value assertion instead of a timeout. The independent checkpoint probe demonstrates this signal in a disposable copy; it is not counted as an existing candidate test.

Limits: Checks are actual independent executions in disposable copies; supplied logs are corroboration only. No unrelated full-repository or shipping-environment claim is made. Race/repeat runs sample schedules and do not prove all interleavings. No compilation failure is credited as behavioral detection; all four mutants compile. The external review/held tests check outcomes and do not become candidate-authored regression coverage. The 2s Go runner deadline is an independent guard; the candidate reader itself has no stop/deadline for wrong-but-coherent totals. Timeout rejection is distinguished from final-state assertion detection. 

Executed evidence: [A-test-go122](../../raw/A-test-go122.json); [A-race-go122](../../raw/A-race-go122.json); [A-race-host](../../raw/A-race-host.json); [A-mutation-split-update-go122](../../raw/A-mutation-split-update-go122.json); [A-mutation-global-state-go122](../../raw/A-mutation-global-state-go122.json); [A-mutation-wrong-signed-sum-go122](../../raw/A-mutation-wrong-signed-sum-go122.json); [A-mutation-drop-unit-updates-go122](../../raw/A-mutation-drop-unit-updates-go122.json). Complete command/stdout/stderr/status pairs are under [raw](../../raw); [mutation sensitivity](../../mutation-sensitivity.json) retains intended assertions.
