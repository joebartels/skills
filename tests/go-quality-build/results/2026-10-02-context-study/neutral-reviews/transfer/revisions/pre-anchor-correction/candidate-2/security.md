## Security — Not applicable

Scope: Candidate 2. Changeset review of the provided immutable original to candidate snapshot; library plus supplied CLI host. Only run.go, run_test.go and newly added cmd/finalize/main_test.go differ. README.md, go.mod and cmd/finalize/main.go are unchanged contract/caller context. Exact SHA-256 snapshots are in ../source-snapshots.json. Declared Go 1.22; independently executed Go 1.22.12 and Go 1.26.5, darwin/arm64, GOTOOLCHAIN=local, GOWORK=off.

Coverage: Conditional applicability only: trusted cooperative callbacks, caller-owned local CLI inputs, unchanged receipt sink/permissions, no changed attacker or authorization boundary.

Rationale: No new security-sensitive actor, authorization policy, secret, dependency, input-to-sink route, or attacker-driven resource decision is implicated in this evolution. Receipt writes are still owned by the supplied local command, and detachment only changes its required post-cancellation completion. This does not claim a general security audit passed.

Limits: Review was sequential within this neutral packet; no mapping, repository design/results, author skills, or other arms were inspected. Authored sources and supplied probes were not modified. Windows interruption is explicitly skipped by both authored suites and the supplied probe; execution here covers darwin/arm64 only. Cooperative nonnil callbacks, nonnil parent context and positive budget are contractual preconditions. No crash, hostile-input, durability/atomicity, packaging/CI, or arbitrary-platform promises were invented. Supplementary reviewer diagnostics are not frozen efficacy tests.

Skill read: ../../review-skills/go-security/SKILL.md (resolved from the frozen packet); no security decision reference was needed after conditional applicability assessment.
