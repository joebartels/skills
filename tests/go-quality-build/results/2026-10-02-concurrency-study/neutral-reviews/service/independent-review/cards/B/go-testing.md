## Testing — C-

Scope: Code-area review of candidate B: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `718c14538b977d31400614c6e32ea2d2452c88d6fbb9f24f536b89d67078dc29`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: reviewed candidate host_test.go assertions, doubles, event gates, cleanup and whether meaningful contract violations evade its authored suite. External held/supplementary/independent probes verify behavior; they are not additions to the candidate regression suite. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: Unique topic causes: major=3, moderate=0, minor=0, critical=0. The first matching unchanged rubric row selects C-. All majors are contained at the helper/fixture boundary; no systemic or critical reach is claimed. Strengths do not cancel findings.

Finding counts: critical=0, major=3, moderate=0, minor=0

Good

- [B/go-testing/G1] Authored tests include meaningful finite work/error identity assertions and bounded channel waits rather than only test names; the independently executed authored suites pass once on both actual versions. Candidate-specific additional positives/limits are below.
- Candidate-specific verified context: The full-cohort gate and Close completion ownership pass the held capacity case; already-canceled/input waiting stop and independent error identity have authored assertions. The pending-Open failure gap is expressly excluded from these positives.

Bad

- [B/T1][major][existing-in-scope] Authored Run failure fixtures only observe failure after admission can finish, leaving pending-Open stop unverified. `candidates/B/host_test.go:111`, `candidates/B/host_test.go:145` — An important failure-containment/liveness contract is not detected: the cooperating pending Open needs caller cancellation despite Run failure. Evidence: `evidence/B-host-authored.log: suite passes P1`; `evidence/B-host-contract.log: independently executed pending-Open assertion fails`.
- [B/T2][major][existing-in-scope] The Close-only regression case closes input after one job, so it cannot detect renewed admission after Close failure. `candidates/B/host_test.go:258` — The important stop-after-Close-failure contract remains unverified independently of error identity. Evidence: `evidence/B-host-authored.log: suite passes P2`; `evidence/B-host-independent.log: Close-only queued-job assertion fails`.
- [B/T3][major][existing-in-scope] The acquired-but-not-started test neither establishes a stop during Open nor checks zero Runs; it explicitly expects the second Run to start. `candidates/B/host_test.go:225`, `candidates/B/host_test.go:253` — The test name suggests a crucial ownership transition is protected, but the actual fixture allows the broken transition and checks only a normally started second Run. Evidence: `candidates/B/host_test.go:253 checks startedSecond.Load()!=1`; `evidence/B-host-authored.log: suite passes P3`; `evidence/B-host-independent.log: cancellation-before-success observes forbidden Run`.

Suggested changes

- [B/T1] Add a pending cooperative Open gated against a failing prior Run; assert it receives stop and Serve completes without caller cancel. Why: An important failure-containment/liveness contract is not detected: the cooperating pending Open needs caller cancellation despite Run failure. Verification: Intended stop-progress assertion must reject unchanged B on both supported test toolchains. Primary owner: host_test.go regression tests.
- [B/T2] Keep later jobs queued after the failing Close and assert one acquisition/release at limit 1. Why: The important stop-after-Close-failure contract remains unverified independently of error identity. Verification: Run unchanged B to verify intended Open-count failure, then corrected behavior to pass. Primary owner: host_test.go regression tests.
- [B/T3] Add an Open callback that cancels before returning success; assert zero second Runs and release of every acquired lease after joining existing Runs. Why: The test name suggests a crucial ownership transition is protected, but the actual fixture allows the broken transition and checks only a normally started second Run. Verification: Use the cancellation-before-success probe to reject P3 and keep ordinary peer failure coverage separately. Primary owner: host_test.go regression tests.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-testing](../../../review-guidance/go-testing/SKILL.md); [topic decisions](../../../review-guidance/go-testing/references/testing-decisions.md). Arithmetic: [topic ledger](go-testing.ledger.json), [calculator result](go-testing.ledger.grade.json).

Ungraded related production findings: B/P1, B/P2, B/P3 are graded by relevant production topics, not added to Testing counts. The test fixes above are independently necessary regression corrections.
