## Architecture & Design — C-

Scope: Code-area review of candidate D: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `2ce34acd60ee9a1e1c3c18df4756a0aa62fe82ab48c994edaa443e132b694cd3`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: the host-supplied Lease protocol, injected Open dependency, cancellation/error propagation and cohort ownership are design boundaries; reviewed API, ownership and composition without demanding another framework. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: Unique topic causes: major=2, moderate=0, minor=0, critical=0. The first matching unchanged rubric row selects C-. All majors are contained at the helper/fixture boundary; no systemic or critical reach is claimed. Strengths do not cancel findings.

Finding counts: critical=0, major=2, moderate=0, minor=0

Good

- [D/go-architecture-and-design/G1] The exact Job/Lease/Serve API remains compatible and construction stays host-owned through the Open function, with no unrelated framework/public API; both actual toolchains compile every source package.
- Candidate-specific verified context: The prefilled two-job batch is concurrent, the cohort WaitGroup join precedes any release, and every successful Open result is collected/closed even when another Open fails. Actual minimum/host build and error identity checks pass; no sparse-input concurrency success is inferred.

Bad

- [D/P1][major][existing-in-scope] Nonblocking job snapshot fixes a potentially one-job cohort and prevents admission of later available jobs while Run/Close proceeds. `candidates/D/host.go:43`, `candidates/D/host.go:73`, `candidates/D/host.go:139` — After job 1 starts on open input, job 2 becomes available while capacity 2 has one owned lease. Its sender and Run cannot progress until job 1 is released. This is serial execution under sparse input, forbidden by the explicit fixture contract. The accepted batch policy does not permit leaving an available second job blocked behind a one-job batch. One contained important admission/concurrency cause; setup timeouts are not four separate defects. Evidence: `evidence/D-host-independent.log: TestIndependentLaterAvailableInputStarts misses second Run while first holds 1 of 2 slots`; `evidence/D-minimum-independent.log: same positive first-start and bounded second-start test`; `evidence/D-host-contract.log: supplied partial-start setup cannot reach blocked third admission`.
- [D/P2][major][existing-in-scope] Close results are accumulated in result but never mark the cohort failed. `candidates/D/host.go:156`, `candidates/D/host.go:161` — The next cohort is admitted after an observed Close error; retained diagnostic identity does not enforce the explicit stop-admission contract. Failure is contained in the helper. Evidence: `evidence/D-host-independent.log: limit 1 produces three opens/closes after first Close failure`; `evidence/D-minimum-independent.log: same intended count assertion`.

Suggested changes

- [D/P1] Keep admission responsive while capacity remains and successful acquired jobs can run; retain the cohort join/Close barrier at release. Why: After job 1 starts on open input, job 2 becomes available while capacity 2 has one owned lease. Its sender and Run cannot progress until job 1 is released. This is serial execution under sparse input, forbidden by the explicit fixture contract. The accepted batch policy does not permit leaving an available second job blocked behind a one-job batch. One contained important admission/concurrency cause; setup timeouts are not four separate defects. Verification: First job arrives and starts; while its Run is gated, make job 2 available. Require job 2 to start before first release, then release/join both and close owned leases. Primary owner: host.go implementation.
- [D/P2] Set failed/cancel after any Close error while finishing every acquired release, then stop admission. Why: The next cohort is admitted after an observed Close error; retained diagnostic identity does not enforce the explicit stop-admission contract. Failure is contained in the helper. Verification: Healthy Run, failing Close, later queued jobs, limit 1: expect one Open/Close and errors.Is sentinel. Primary owner: host.go implementation.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-architecture-and-design](../../../review-guidance/go-architecture-and-design/SKILL.md); [topic decisions](../../../review-guidance/go-architecture-and-design/references/design-decisions.md). Arithmetic: [topic ledger](go-architecture-and-design.ledger.json), [calculator result](go-architecture-and-design.ledger.grade.json).
