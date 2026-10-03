# Independent Go Quality Reports — candidate A: B; candidate B: A

Scope: Separate bounded code-area reviews of the complete candidate A and B four-file library modules. The original README/source establish the same contract for each: Go 1.22, standard library only, zero value, no copying after use, exact Count/Sum, coherent independent snapshots and isolated instances. Exact candidate and original SHA-256 snapshots, skill/reference hashes and coverage are in [manifest](manifest.json). No author report, profile, exposure identity, repository planning/results or other reviewer material was used.

Coverage: All nine topic skills and relevant references were read unchanged. Architecture, Code Quality, Correctness, Testing, Performance and Dependencies are assessed completely within the supplied area. Security, Observability and Deployment have explicit scope-based Not applicable reasons; none is substituted for missing material evidence. No material requested obligation remains unavailable. Each candidate was independently exercised on host Go1.26.5 and actual Go1.22.12; mutations exist only in disposable copies.

## Candidate A — B

Rationale: One unique moderate test-lifecycle defect selects B; no production defect is confirmed. Lost updates remain coherent, so A waits forever for expected Count before it can run its final-state assertion. The independent runner deadline rejects the mutant, but that is weaker and slower diagnostic evidence than a result assertion. Other important assertions have verified sensitivity.

Unique finding counts: critical=0, major=0, moderate=1, minor=0.

### Report cards

| Topic | Grade/state | Coverage and applicability | Card |
| --- | --- | --- | --- |
| Architecture & Design | A | Applicable: per-instance ownership, zero-value construction, public API and coherent value publication are library design decisions. The complete supplied API and implementation are assessed. | [card](cards/A/architecture.md) |
| Code Quality & Go Idioms | A | Applicable and mandatory: complete Go implementation and tests contain receiver, mutex ownership, documentation and Go-version idiom decisions. | [card](cards/A/code-quality.md) |
| Correctness & Compatibility | A | Applicable and mandatory: exact update count/sum, zero value, simultaneous Add/Snapshot, coherent independent values, isolated instances and retained public signatures are explicit README contracts. | [card](cards/A/correctness.md) |
| Testing | B | Applicable and mandatory: candidate tests must meaningfully detect concurrent coherence, accepted-count, arithmetic and instance-state regressions. Assertions and termination are checked with compiled behavioral mutations. | [card](cards/A/testing.md) |
| Security | Not applicable | Not applicable in this bounded code area: fixed-width integer arithmetic and a caller-owned local mutex introduce no attacker-controlled I/O, privilege boundary, credentials, sensitive sink or input-sized resource amplification. No external dependency is selected. This is a scope judgment, not a claim that any embedding application is secure. | [card](cards/A/security.md) |
| Observability & Resilience | Not applicable | Not applicable in this bounded code area: synchronous local arithmetic has no fallible external dependency, retry, queue, telemetry export or service lifecycle contract. Mutex ownership/liveness is assessed under Correctness and Performance. No context, instrumentation or operational machinery is required by the supplied contract. | [card](cards/A/observability.md) |
| Performance & Resource Management | A | Applicable: adding synchronization changes contention and resource ownership. Fixed work/state, finite critical sections, no internal goroutines, per-instance locks and allocations are assessed. No throughput/latency budget or speed claim is supplied. | [card](cards/A/performance.md) |
| Dependencies & Reproducibility | A | Applicable: standard-library-only and minimum Go 1.22 are explicit build contracts. Standalone module resolution and compilation/tests on Go 1.22.12 and host 1.26.5 are assessed with network disabled. | [card](cards/A/dependencies.md) |
| Deployment & Operations | Not applicable | Not applicable in this bounded code area: it supplies a local library and no change to release tags, packaging, CI enforcement, process configuration, deployment or operator contract. Source builds are assessed under Dependencies and Correctness. Excluded surrounding delivery infrastructure is not alleged to be missing. | [card](cards/A/deployment.md) |

### Good

- [A-G1] The same per-instance mutex protects both arithmetic updates and snapshot evaluation; scalar returns preserve independent value ownership. External API/zero/mixed/checkpoint/held race checks pass on both toolchains — preserves the explicitly required coherent observation contract.
- [A-G2] Standalone network-disabled readonly builds/test/vet succeed on the minimum and host toolchains; gofmt diff is empty and module graph is standard-library-only — preserves the declared build floor without hidden resolution inputs.
- [A-G3] Candidate assertions actually reject incoherent-pair, process-global-state and signed-sum regressions after successful compilation — provides observable regression signal beyond test names or clean baseline output.

### Bad

- [A-F1][moderate] [Reader completion depends on expected Count](cards/A/testing.md), at `candidates/A/totals_test.go:48-58`. A coherent dropped-update mutant finishes its writers but leaves the reader polling and the test waiting until the 2s runner deadline; the final-state assertion never runs. Confirmed on host and Go1.22.12, including the whole candidate suite. This is a test-lifecycle defect, counted once.

### Suggested changes

- [A-F1][priority 2] The concurrency-test owner should stop/join the reader from a writer-completion event, bound failure waits and then assert exact totals. Verify that the retained dropped-update mutant produces a prompt wrong-value assertion while the split-update mutant still fails coherence. The reviewer-owned [event-based probe](contract_probe_test.go) demonstrates the intended signal; it is separate from candidate tests.

### Limits

The grade is for the supplied library area, not an embedding application or release. Finite race/shuffle runs do not prove every schedule. All actual baseline commands and checks are preserved in [raw](raw); supplied [checks-A](../checks-A.json) are distinguished from independent reruns. The [mutation matrix](mutation-sensitivity.md) records each intended assertion and separates assertion failures, runner deadlines and build errors. None of the four supplementary mutants failed compilation; A timeout is explicitly not final-state assertion credit. No unspecified frozen evaluation result is inferred. No speed/latency, vulnerability-scan, unspecified platform-matrix or byte-identical artifact conclusion is claimed. Sources match their initial hashes and supplied candidate manifest.

Details: [applicability](applicability-A.json), [ledger](ledger-A.json), [calculator result](grade-A.json), [deduplicated findings](findings.json).

## Candidate B — A

Rationale: No actionable defect is confirmed in the complete bounded scope. The mutex establishes joint publication, signature/zero-value/value-ownership checks pass, and the original test assertions detect all four compiled supplementary mutations. Reader shutdown follows writer completion, so a wrong final count produces an immediate assertion. Verified ordinary correct setup supports A; no two nonroutine safeguards support A+.

Unique finding counts: critical=0, major=0, moderate=0, minor=0.

### Report cards

| Topic | Grade/state | Coverage and applicability | Card |
| --- | --- | --- | --- |
| Architecture & Design | A | Applicable: per-instance ownership, zero-value construction, public API and coherent value publication are library design decisions. The complete supplied API and implementation are assessed. | [card](cards/B/architecture.md) |
| Code Quality & Go Idioms | A | Applicable and mandatory: complete Go implementation and tests contain receiver, mutex ownership, documentation and Go-version idiom decisions. | [card](cards/B/code-quality.md) |
| Correctness & Compatibility | A | Applicable and mandatory: exact update count/sum, zero value, simultaneous Add/Snapshot, coherent independent values, isolated instances and retained public signatures are explicit README contracts. | [card](cards/B/correctness.md) |
| Testing | A | Applicable and mandatory: candidate tests must meaningfully detect concurrent coherence, accepted-count, arithmetic and instance-state regressions. Assertions and termination are checked with compiled behavioral mutations. | [card](cards/B/testing.md) |
| Security | Not applicable | Not applicable in this bounded code area: fixed-width integer arithmetic and a caller-owned local mutex introduce no attacker-controlled I/O, privilege boundary, credentials, sensitive sink or input-sized resource amplification. No external dependency is selected. This is a scope judgment, not a claim that any embedding application is secure. | [card](cards/B/security.md) |
| Observability & Resilience | Not applicable | Not applicable in this bounded code area: synchronous local arithmetic has no fallible external dependency, retry, queue, telemetry export or service lifecycle contract. Mutex ownership/liveness is assessed under Correctness and Performance. No context, instrumentation or operational machinery is required by the supplied contract. | [card](cards/B/observability.md) |
| Performance & Resource Management | A | Applicable: adding synchronization changes contention and resource ownership. Fixed work/state, finite critical sections, no internal goroutines, per-instance locks and allocations are assessed. No throughput/latency budget or speed claim is supplied. | [card](cards/B/performance.md) |
| Dependencies & Reproducibility | A | Applicable: standard-library-only and minimum Go 1.22 are explicit build contracts. Standalone module resolution and compilation/tests on Go 1.22.12 and host 1.26.5 are assessed with network disabled. | [card](cards/B/dependencies.md) |
| Deployment & Operations | Not applicable | Not applicable in this bounded code area: it supplies a local library and no change to release tags, packaging, CI enforcement, process configuration, deployment or operator contract. Source builds are assessed under Dependencies and Correctness. Excluded surrounding delivery infrastructure is not alleged to be missing. | [card](cards/B/deployment.md) |

### Good

- [B-G1] The same per-instance mutex protects both arithmetic updates and snapshot evaluation; scalar returns preserve independent value ownership. External API/zero/mixed/checkpoint/held race checks pass on both toolchains — preserves the explicitly required coherent observation contract.
- [B-G2] Standalone network-disabled readonly builds/test/vet succeed on the minimum and host toolchains; gofmt diff is empty and module graph is standard-library-only — preserves the declared build floor without hidden resolution inputs.
- [B-G3] Candidate assertions actually reject incoherent-pair, process-global-state and signed-sum regressions after successful compilation — provides observable regression signal beyond test names or clean baseline output.
- [B-G4] The dropped-update mutation reports `final snapshot = {Count:0 Sum:0}` immediately on both toolchains; completion does not depend on the expected numeric result — keeps exact-count verification reachable when that behavior regresses.

### Bad

- None found.

### Suggested changes

- None needed. Optional deterministic intermediate checkpoints may strengthen scheduling coverage; they are not graded defects.

### Limits

The grade is for the supplied library area, not an embedding application or release. Finite race/shuffle runs do not prove every schedule. All actual baseline commands and checks are preserved in [raw](raw); supplied [checks-B](../checks-B.json) are distinguished from independent reruns. The [mutation matrix](mutation-sensitivity.md) records each intended assertion and separates assertion failures, runner deadlines and build errors. None of the four supplementary mutants failed compilation; A timeout is explicitly not final-state assertion credit. No unspecified frozen evaluation result is inferred. No speed/latency, vulnerability-scan, unspecified platform-matrix or byte-identical artifact conclusion is claimed. Sources match their initial hashes and supplied candidate manifest.

Details: [applicability](applicability-B.json), [ledger](ledger-B.json), [calculator result](grade-B.json), [deduplicated findings](findings.json).

Independent evidence: [source hashes before](source-hashes-before.json), [after probes](source-hashes-after-probes.json), [probe construction](run_probes.py), [exact baseline construction](run_checks.py), [follow-up construction](run_followup.py). Review performed by one independent reviewer without delegation; identical production mechanisms are described separately for each snapshot and are never counted as multiple defects.
