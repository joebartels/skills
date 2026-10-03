## Testing — C-

Scope: Code-area review of candidate D: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `2ce34acd60ee9a1e1c3c18df4756a0aa62fe82ab48c994edaa443e132b694cd3`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: reviewed candidate host_test.go assertions, doubles, event gates, cleanup and whether meaningful contract violations evade its authored suite. External held/supplementary/independent probes verify behavior; they are not additions to the candidate regression suite. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: Unique topic causes: major=2, moderate=0, minor=0, critical=0. The first matching unchanged rubric row selects C-. All majors are contained at the helper/fixture boundary; no systemic or critical reach is claimed. Strengths do not cancel findings.

Finding counts: critical=0, major=2, moderate=0, minor=0

Good

- [D/go-testing/G1] Authored tests include meaningful finite work/error identity assertions and bounded channel waits rather than only test names; the independently executed authored suites pass once on both actual versions. Candidate-specific additional positives/limits are below.
- Candidate-specific verified context: The prefilled two-job batch is concurrent, the cohort WaitGroup join precedes any release, and every successful Open result is collected/closed even when another Open fails. Actual minimum/host build and error identity checks pass; no sparse-input concurrency success is inferred.

Bad

- [D/T1][major][existing-in-scope] Concurrency tests preload closed full batches and the partial-open fixture forbids any Run, leaving sparse-input concurrent admission unverified. `candidates/D/host_test.go:29`, `candidates/D/host_test.go:109` — The important capacity/concurrency contract is verified only for an already full buffered batch; the implementation can serialize ordinary unbuffered input while these tests pass. Evidence: `evidence/D-host-authored.log: suite passes P1`; `evidence/D-host-independent.log: later-available-input assertion fails`.
- [D/T2][major][existing-in-scope] The error-retention fixture has one job and combined Run/Close failure, so Close-only stop-admission is unverified. `candidates/D/host_test.go:74` — The important Close-stage admission boundary can regress while existing independent error-identity checks remain green. Evidence: `evidence/D-host-authored.log: suite passes P2`; `evidence/D-host-independent.log: Close-only admission continues`.

Suggested changes

- [D/T1] Add separately arriving jobs on caller-owned open input; gate first Run, require second Run with available capacity, and check partial startup cleanup separately. Why: The important capacity/concurrency contract is verified only for an already full buffered batch; the implementation can serialize ordinary unbuffered input while these tests pass. Verification: Use positive first-start and actual second availability; no claim that every Open must complete in a particular relative order. Primary owner: host_test.go regression tests.
- [D/T2] Add healthy Run plus independent Close failure with remaining jobs and exact acquired/released counts. Why: The important Close-stage admission boundary can regress while existing independent error-identity checks remain green. Verification: Intended count assertion rejects unchanged D and accepts corrected stop transition. Primary owner: host_test.go regression tests.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-testing](../../../review-guidance/go-testing/SKILL.md); [topic decisions](../../../review-guidance/go-testing/references/testing-decisions.md). Arithmetic: [topic ledger](go-testing.ledger.json), [calculator result](go-testing.ledger.grade.json).

Ungraded related production findings: D/P1, D/P2 are graded by relevant production topics, not added to Testing counts. The test fixes above are independently necessary regression corrections.
