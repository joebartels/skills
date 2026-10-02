## Observability & Resilience — A

Scope: Code-area review of completed neutral candidate-2: process.go, process_test.go, README.md and go.mod; original/README.md is authoritative contract, original source/test are behavior context. Snapshot SHA-256 identities are in [manifest.json](manifest.json). Library module example.com/process, go 1.22; independent checks used Go 1.26.5 and Go 1.22.12 darwin/arm64.

Coverage: Cooperative cancellation, between-job admission, failure/cancellation diagnostic cause exposure and completion order in the entire synchronous library area. Telemetry, retries, queues, distributed budgets and services are outside this contract.

Rationale: No actionable resilience issue. Cancellation requests stopping and the callback owns cooperation, exactly as documented. The function neither hides independent failure nor invalidates accepted final success. Useful returned errors are the appropriate owned signal for this small caller-controlled library.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] candidate-2/process.go:10-20 stops new callbacks on cancellation and preserves classification/custom cause with independent failure; independently executed TestCustomCauseAndIndependentFailure holds/releases the callback and succeeds.
- [G2] candidate-2/process.go:23-25 completes accepted final work without a later cancellation reclassification; TestResultDecisionOrder and the author final-success case pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: No hard callback deadline or preemption is promised; a non-cooperative callback is a caller-owned contract violation, not a missing internal goroutine/timeout. No runtime telemetry/SLO or deployed service is supplied; none is assumed. Duplicate standard cancellation text when classification equals cause is unpromised formatting and optional to refine safely. Exact checks: [checks.json](checks.json).

Skill contract: [SKILL.md](../../review-skills/go-observability-and-resilience/SKILL.md); topic decision reference inspected where relevant.
