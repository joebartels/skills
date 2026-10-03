## Code Quality & Go Idioms — C

Scope: Code-area review of candidate D: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `2ce34acd60ee9a1e1c3c18df4756a0aa62fe82ab48c994edaa443e132b694cd3`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: reviewed local state transitions, result collection, checked errors, context guards, docs, value/closure semantics under go 1.22; gofmt and vet independently executed. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: Unique topic causes: major=1, moderate=0, minor=0, critical=0. The first matching unchanged rubric row selects C. All majors are contained at the helper/fixture boundary; no systemic or critical reach is claimed. Strengths do not cancel findings.

Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [D/go-code-quality-and-idioms/G1] Returned Open/Run/Close error causes are collected using errors.Join or %w and checked at useful lifecycle boundaries; candidate authored error-identity tests pass. gofmt -l reports no candidate changes; go vet is clean.
- Candidate-specific verified context: The prefilled two-job batch is concurrent, the cohort WaitGroup join precedes any release, and every successful Open result is collected/closed even when another Open fails. Actual minimum/host build and error identity checks pass; no sparse-input concurrency success is inferred.

Bad

- [D/P2][major][existing-in-scope] Close results are accumulated in result but never mark the cohort failed. `candidates/D/host.go:156`, `candidates/D/host.go:161` — The next cohort is admitted after an observed Close error; retained diagnostic identity does not enforce the explicit stop-admission contract. Failure is contained in the helper. Evidence: `evidence/D-host-independent.log: limit 1 produces three opens/closes after first Close failure`; `evidence/D-minimum-independent.log: same intended count assertion`.

Suggested changes

- [D/P2] Set failed/cancel after any Close error while finishing every acquired release, then stop admission. Why: The next cohort is admitted after an observed Close error; retained diagnostic identity does not enforce the explicit stop-admission contract. Failure is contained in the helper. Verification: Healthy Run, failing Close, later queued jobs, limit 1: expect one Open/Close and errors.Is sentinel. Primary owner: host.go implementation.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-code-quality-and-idioms](../../../review-guidance/go-code-quality-and-idioms/SKILL.md); [topic decisions](../../../review-guidance/go-code-quality-and-idioms/references/idiom-decisions.md). Arithmetic: [topic ledger](go-code-quality-and-idioms.ledger.json), [calculator result](go-code-quality-and-idioms.ledger.grade.json).
