## Testing — C

Scope: Code-area review of candidate E: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `f2f7dd6470aae322fe98a33d39e71addae18d9d749424e94723e7ca9213d208d`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: reviewed candidate host_test.go assertions, doubles, event gates, cleanup and whether meaningful contract violations evade its authored suite. External held/supplementary/independent probes verify behavior; they are not additions to the candidate regression suite. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: Unique topic causes: major=1, moderate=0, minor=0, critical=0. The first matching unchanged rubric row selects C. All majors are contained at the helper/fixture boundary; no systemic or critical reach is claimed. Strengths do not cancel findings.

Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [E/go-testing/G1] Authored tests include meaningful finite work/error identity assertions and bounded channel waits rather than only test names; the independently executed authored suites pass once on both actual versions. Candidate-specific additional positives/limits are below.
- Candidate-specific verified context: All unchanged authored, held, supplementary and independent probes pass on both versions; host race/shuffle/count=3 passes. Source correctly retains held cleanup/cancellation, skips Run after canceled acquisition, interrupts pending cooperative Open from worker failure, and stops on Close failure. The surviving isolated Close-stop mutant is a Testing finding only.

Bad

- [E/T1][major][existing-in-scope] Close-only failure fixtures cannot distinguish retaining the error from stopping future admission. `candidates/E/host_test.go:310`, `candidates/E/host_test.go:345` — The source implements the important stop-after-Close contract correctly, but a compiled mutation removing only that transition evades all authored tests. The held-Close case also cancels the caller, masking the independent Close-only cause. This is a major verification gap requiring its own regression assertion, not a production defect. Evidence: `evidence/mutation-commands.json: remove-close-failure-stop authored suite passes host and minimum`; `evidence/mutation-commands.json: intended TestIndependentCloseFailureStopsNextAdmission fails with opened=3, closed=3 on both versions`; `mutations/E/remove-close-failure-stop/mutation.patch.txt`.

Suggested changes

- [E/T1] Add a healthy Run and failing Close at limit 1 with later jobs queued; assert no further Open while retaining Close error. Why: The source implements the important stop-after-Close contract correctly, but a compiled mutation removing only that transition evades all authored tests. The held-Close case also cancels the caller, masking the independent Close-only cause. This is a major verification gap requiring its own regression assertion, not a production defect. Verification: Current E must pass; the isolated mutation must fail the intended admission-count assertion. Unrelated compile/runner failure is not detection. Primary owner: host_test.go regression tests.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-testing](../../../review-guidance/go-testing/SKILL.md); [topic decisions](../../../review-guidance/go-testing/references/testing-decisions.md). Arithmetic: [topic ledger](go-testing.ledger.json), [calculator result](go-testing.ledger.grade.json).
