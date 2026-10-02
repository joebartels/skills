# Isolation revision 3 outcomes

| Case | Testing | Correctness | Architecture |
| --- | --- | --- | --- |
| library-file-fixtures | A+ | A | Not applicable |
| cli-environment | A+ | A | Not applicable |
| service-http-boundary | A+ | A | Not applicable |
| worker-timing | A+ | A | Not applicable |
| not-isolation-work | A | A | Not applicable |
| file-repeat | A+ | A | Not applicable |

All six patches reconstruct exactly. Ordinary/vet/format and applicable race/ten-shuffle/withheld/individual checks pass. File/repeat focused selection and persistence checks retain useful shared-root failures at parallelism1; expected values come from executed contract cases, not returned production values. Environment preserves serial parent/subtest state, including unset and present-empty, and tests the real executable with filtered application configuration. Its builder-default check passes unchanged with GOCACHE unset and valid disposable HOME. HTTP tests separately establish constructed body/error/closure observations and real local protocol/cancellation. Worker uses readiness/completion events, completion-relative recurrence and joined fixture cleanup. The pure calculation repair declines isolation and keeps its API/Go minimum.

Nine candidate-only controller diagnostics compile and fail meaningful assertions: six production regressions and three explicitly test-helper sabotages (two parent teardowns and environment restoration). Shared-root signal persists at parallelism1; cancellation context/value assertions support detections independently of additional bounded stalled-work diagnostics. Independent reviewer mutation sources/checks are separately preserved. Counts and already-strong controls alone establish no uplift.

Neutral reviews use three disclosed pairs of unrelated changesets, separate cards/checks and no comparative grading. Repeat and first file have different review contexts. Writing guidance, arm labels, reports, prior reviews, expected assertions and controller probes are withheld from these outcome contexts. Earlier labelled baseline/revision1 outcome launches limit fully blinded causal grade comparisons; reproduced focused-child failures/corrections, repeats and open promotion verification support bounded utility.

Isolation3 retains independent lifetime/fidelity/timing guidance and conditionally distinguishes builder/tool environment requirements from tested application state. [Affected integrated outcomes](../2026-10-01-testing-combined-r3/comparison.md) close the historical both setup failure when GOCACHE is unset. Exact guidance stays outside runtime until separate accepting promotion assessment; behavior2 is already promoted. No guided candidate repair counts as unseen evidence.

Go1.26.5 darwin/arm64 executed Go1.22-declared modules. Actual Go1.22, other platforms, optional newer APIs, database rollback, every schedule, automatic routing and Codex/OpenCode runtime loading remain unverified. Restricted listener failures and approved unchanged-source reruns remain raw; parent race checks do not instrument ordinary child binaries. No new probe/author name collision occurred in this revision. Older collision adapters/raw failures and mixed integrated findings remain immutable history.
