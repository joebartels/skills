# Frozen context mutations

These four meanings are frozen before candidate exposure. Controller conformance checks establish fixture feasibility, not author uplift. Candidate-specific compiling patches are archived per trial; do not mistake a build error or unrelated timeout for sensitivity.

| ID | Changed production behavior | Intended held observation |
| --- | --- | --- |
| `unsafe-cause-equality` | Compare a callback's arbitrary error directly to the same slice-backed custom cancellation cause. | `TestCauseReturnedByCallback` reports a legal-input panic through its explicit recovery/assertion. |
| `lost-custom-cause` | Return standard cancellation and independent callback failure while dropping the inspectable custom cause. | `TestCustomCauseAndIndependentFailure` / `TestCauseReturnedByCallback` fail cause inspection. |
| `lost-total-budget` | Derive every stage from the original caller instead of the single total operation scope. | `TestTotalAndEarlierParentDeadline` fails total-deadline retention; an earlier caller may still work. |
| `unbounded-finalization` | Detach finalization but replace its time-bound child with cancellation alone. | `TestFinalizationDeadline` fails the bounded cooperative-finalization contract. |

Before author launch, record input/probe/conformance hashes, concrete substitutions and compiled/intended-failure results. Additional probes after exposure remain supplementary.
