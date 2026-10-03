## Dependencies & Reproducibility — A

Scope: Code-area review of candidate A: host.go, host_test.go, go.mod, README.md; exact supplied hashes in [source manifest](../../../source-manifest.json). host.go SHA256 `c45a48c184493ebe70f1f52cbb831194f2a25604e49776687930cd951fd34ab7`. Standard-library owned worker host, unchanged Job/Lease/Serve protocol; Go 1.22 minimum.

Coverage: Applicable: go.mod declares Go 1.22 and standard-library-only production/tests. Verified standalone readonly resolution, host and actual minimum compilation/test execution, imports, module graph and exact input hashes. No generated/native/release inputs or bit-identical artifact contract. All bounded applicable obligations complete; outside-scope unavailable host evidence is excluded, not passed.

Rationale: No actionable in-topic cause is confirmed. Verified strengths and complete bounded material coverage support A. No A+ safeguards are claimed: these are competent required lifecycle/build controls, without a separate justified pair beyond routine correct setup.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A/go-dependencies-and-reproducibility/G1] Actual go1.22.12 and go1.26.5 compile the standalone go 1.22 standard-library-only module with GOWORK=off, GOTOOLCHAIN=local, GOFLAGS=-mod=readonly; go list -m all returns only example.com/service-owned-workers. No external go.sum is needed.
- Candidate-specific verified context: The supplied held and supplementary fixtures pass on both versions. Async Open is joined and acquired leases are closed after the actual Run results; pending cooperative Open is interrupted on an observed Run failure. The bounded stress evidence adds the uncommon post-Open ready-result case that one-shot checks missed.

Bad

- None found.

Suggested changes

- None needed.

Limits: Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated.

Skill and reference inspected: [go-dependencies-and-reproducibility](../../../review-guidance/go-dependencies-and-reproducibility/SKILL.md); [topic decisions](../../../review-guidance/go-dependencies-and-reproducibility/references/dependency-reproducibility-decisions.md). Arithmetic: [topic ledger](go-dependencies-and-reproducibility.ledger.json), [calculator result](go-dependencies-and-reproducibility.ledger.grade.json).
