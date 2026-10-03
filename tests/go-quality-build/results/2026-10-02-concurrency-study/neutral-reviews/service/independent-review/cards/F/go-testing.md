## Testing — C-

Scope: Code-area review of candidate F: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `9a07e02eb8592790dc8cbaef39f97b392403dcd6e10f7556444c2aa437980aca`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: reviewed candidate host_test.go assertions, doubles, event gates, cleanup and whether meaningful contract violations evade its authored suite. External held/supplementary/independent probes verify behavior; they are not additions to the candidate regression suite. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: Unique topic causes: major=3, moderate=0, minor=0, critical=0. The first matching unchanged rubric row selects C-. All majors are contained at the helper/fixture boundary; no systemic or critical reach is claimed. Strengths do not cancel findings.

Finding counts: critical=0, major=3, moderate=0, minor=0

Good

- [F/go-testing/G1] Authored tests include meaningful finite work/error identity assertions and bounded channel waits rather than only test names; the independently executed authored suites pass once on both actual versions. Candidate-specific additional positives/limits are below.
- Candidate-specific verified context: The full-initial-cohort WaitGroup join and release checks pass when Runs ignore cancellation; wrapped Open/Run/Close causes retain errors.Is identity with job context. This does not validate healthy cooperative completion or paced-input ownership capacity.

Bad

- [F/T1][major][existing-in-scope] Healthy capacity fixtures ignore their Run contexts, so premature coordinated stop is invisible. `candidates/F/host_test.go:33`, `candidates/F/host_test.go:78`, `candidates/F/host_test.go:148` — The important uncanceled success contract is tested with callbacks that keep succeeding after any stop request. Their gates verify sequencing while masking whether work was allowed to complete cooperatively. Evidence: `evidence/F-host-authored.log: suite passes P1`; `evidence/F-host-independent.log: cooperative healthy Run contexts detect premature cancellation`.
- [F/T2][major][existing-in-scope] Authored cancellation coverage does not exercise an acquired success returned after stopping and assert no Run. `candidates/F/host_test.go:135`, `candidates/F/host_test.go:259` — Input-wait/already-canceled tests do not protect the important partial acquisition ownership transition. Evidence: `evidence/F-host-authored.log: suite passes P2`; `evidence/F-host-independent.log: forbidden Run after callback cancellation`.
- [F/T3][major][existing-in-scope] Capacity tests inspect only a full initial cohort and later Run starts, without counting live acquired leases across completed Runs on open input. `candidates/F/host_test.go:33`, `candidates/F/host_test.go:78`, `candidates/F/host_test.go:148` — An explicit important capacity bound through Close is effectively unverified for a stream of successful completions; Run concurrency is insufficient as its proxy. Evidence: `evidence/F-host-authored.log: suite passes P3`; `evidence/F-host-independent.log: eight held leases at capacity 2`.

Suggested changes

- [F/T1] Make healthy callbacks verify they are not canceled before their external success gate and return ctx.Err when improperly stopped. Why: The important uncanceled success contract is tested with callbacks that keep succeeding after any stop request. Their gates verify sequencing while masking whether work was allowed to complete cooperatively. Verification: Unchanged F must fail the intended healthy-work cancellation/success assertions. Primary owner: host_test.go regression tests.
- [F/T2] Return a successful lease after callback cancellation, assert zero Runs, and still require owned Close after cohort join. Why: Input-wait/already-canceled tests do not protect the important partial acquisition ownership transition. Verification: Reject P2 via the intended zero-Run assertion on a compiled package. Primary owner: host_test.go regression tests.
- [F/T3] Track live acquisition/release counts and use paced successful jobs on open input to assert the lease bound, plus held Close. Why: An explicit important capacity bound through Close is effectively unverified for a stream of successful completions; Run concurrency is insufficient as its proxy. Verification: An implementation that releases active capacity at Run return must fail the intended peak-owned-leases assertion. Primary owner: host_test.go regression tests.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-testing](../../../review-guidance/go-testing/SKILL.md); [topic decisions](../../../review-guidance/go-testing/references/testing-decisions.md). Arithmetic: [topic ledger](go-testing.ledger.json), [calculator result](go-testing.ledger.grade.json).

Ungraded related production findings: F/P1, F/P2, F/P3 are graded by relevant production topics, not added to Testing counts. The test fixes above are independently necessary regression corrections.
