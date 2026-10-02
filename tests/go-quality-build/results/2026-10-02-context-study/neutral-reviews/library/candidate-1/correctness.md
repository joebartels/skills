## Correctness & Compatibility — A

Scope: Code-area review of completed neutral candidate-1: process.go, process_test.go, README.md and go.mod; original/README.md is authoritative contract, original source/test are behavior context. Snapshot SHA-256 identities are in [manifest.json](manifest.json). Library module example.com/process, go 1.22; independent checks used Go 1.26.5 and Go 1.22.12 darwin/arm64.

Coverage: Normal sequential progress, first failure after accepted prefix, active empty, already-canceled empty/nonempty, between-job stop, coincident accepted final completion, cooperative blocked callback, independent callback error with custom cause, same non-comparable callback/cancellation cause, signature and minimum-version preservation.

Rationale: No production contract defect confirmed. Entry/pre-job checks stop unstarted work, post-error observation retains each promised cause, and accepted final success returns without a later cancellation check. No interface equality is used. Ordinary correct implementation and focused tests justify A; no claim that the author suite detects every regression.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] candidate-1/process.go:10-24 preserves exact accepted prefix and stops before another callback; unchanged TestSequentialProgress and provided probes pass independently.
- [G2] candidate-1/process.go:29 joins ctx.Err and context.Cause without comparing legal dynamic error values; independent TestCauseReturnedByCallback passes and both errors.Is/As contracts are retained.
- [G3] Independent baseline build, author tests and provided contract suite pass on Go 1.26.5 and Go 1.22.12. Provided-contract race/shuffle/count=3 run also passes for the exercised cancellation interleaving.

Bad

- None found.

Suggested changes

- None needed.

Limits: The API explicitly requires nonnil ctx and callback plus cooperative callback return; nil arguments or forced preemption are not required. No supported platform matrix beyond version is supplied; host runs are darwin/arm64. Race checks cover only executed paths. Two mutation failures show test gaps, not defects in the unmodified implementations. Full logs: [checks.json](checks.json).

Skill contract: [SKILL.md](../../review-skills/go-correctness-and-compatibility/SKILL.md); topic decision reference inspected where relevant.
