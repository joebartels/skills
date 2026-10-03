## Testing — C

Scope: Code-area review of completed neutral candidate-1: process.go, process_test.go, README.md and go.mod; original/README.md is authoritative contract, original source/test are behavior context. Snapshot SHA-256 identities are in [manifest.json](manifest.json). Library module example.com/process, go 1.22; independent checks used Go 1.26.5 and Go 1.22.12 darwin/arm64.

Coverage: All author tests/assertions and original regression, provided probes, baseline/race/Go-version runs, both unchanged frozen mutations, plus matched supplementary entry-check removal and spurious-empty-callback insertion in separate scratch copies.

Rationale: One contained major gap in the explicitly required legal non-comparable-cause panic boundary and one independent moderate empty-input assertion/coverage gap. The shared rubric selects C (one major with at most one moderate). Tests meaningfully check most new behavior; passing count/coverage does not override demonstrated mutation survivors.

Finding counts: critical=0, major=1, moderate=1, minor=0

Good

- [G1] Preserved TestSequentialProgress asserts accepted prefix, callback order and retained independent failure; both baseline author suites pass.
- [G2] The frozen lost-custom-cause mutant compiles and fails both full author suites: errors.Is/As assertions detect omitted custom causes, independently reproducing the supplied observation.
- [G3] Final-success coincident cancellation and between-job cancellation assertions detect distinct expected outcomes without sleeps or test goroutines. Provided blocked-callback probe additionally synchronizes the exercised cancellation interleaving and passes race/shuffled runs.

Bad

- [T1][major][existing-in-scope] candidate-1/process_test.go:70: the callback returns errors.New while the cancellation cause has a different non-comparable dynamic type. Go interface equality can compare differing dynamic types without inspecting the non-comparable value. The frozen unsafe-cause-equality mutant therefore compiles and passes the full author suite, but independent held TestCauseReturnedByCallback fails with runtime error: comparing uncomparable type process.listCause. The README explicitly makes arbitrary legal causes and panic-free equality an important boundary; its panic regression remains effectively unverified. Reach is this function contract, so contained major rather than systemic/critical.
- [T2][moderate][existing-in-scope] candidate-1/process_test.go:29 tests already-canceled input only with a job; its active-empty case uses a live context. Removing only the entry cancellation check leaves the per-job guard intact, compiles and passes all author tests, yet held TestCanceledBeforeStart fails on nil jobs with err=<nil>. This is a localized normal-use boundary regression in the explicit already-canceled-empty behavior; correcting the missing empty-canceled case is independent of the equality panic test.

Suggested changes

- [T1] Add an author regression where the callback cancels with a slice-containing legal error and returns that same cause (or another value of its same dynamic type). Assert zero/accepted-prefix result, standard cancellation classification and errors.As payload. Re-run the unchanged frozen unsafe equality mutation: it must compile and fail the author suite while baseline passes.
- [T2] Exercise an already-canceled context with both empty and nonempty jobs, asserting zero callbacks/accepted plus cancellation classification and inspectable custom cause. The supplementary entry-check-removal mutation must fail the author suite.

Limits: Production baseline meets these paths; findings assess author regression detection only. The provided held probes are separate reviewer evidence, not tests delivered in candidate source. Frozen efficacy facts are unsafe-cause-equality survivor and lost-custom-cause kill; the two empty-input mutations are explicitly supplementary after inspecting the packet and must not enter frozen efficacy-benefit counts. No statement-coverage percentage, fuzzing, benchmark, CI enforcement or exhaustive schedule claim is made. Exact commands/output and scratch locations: [checks.json](checks.json).

Skill contract: [SKILL.md](../../review-skills/go-testing/SKILL.md); topic decision reference inspected where relevant.
