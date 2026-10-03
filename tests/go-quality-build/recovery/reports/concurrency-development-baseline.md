# Independent concurrency trial
Date: 2026-10-03. Status: implementation and bounded local verification complete.

## Scope and guidance
Artifacts: `workers.go`, `workers_test.go`, and this report. README/module/API/dependencies/`go 1.22` unchanged; an exact Serve function-type assignment compiles.
Read only this trial's approved README, starting source/tests/module, RTK instructions, and selected neighboring guidance. No repository/history, probes, handoff, other trial artifacts, agents, commits, PRs, or global edits.
Selected from `/private/tmp/go-skill-recovery-20261003/concurrency-neighbors`:
- `go-context-and-deadlines/SKILL.md` (provided revision 2): stopping boundaries, independent failures, caller classification/custom cause.
- `go-api-contracts/SKILL.md`: exact exported API and error/ownership contract.
- `go-interfaces-and-composition/SKILL.md`: supplied dependencies, per-lease ownership, stop then join.
- `go-behavior-tests/SKILL.md` plus `references/behavior-observations.md`: independent assertions, work/cleanup combinations, noncomparable errors.
- `go-test-isolation/SKILL.md` plus `references/isolation-patterns.md`: event-controlled callbacks and bounded cancel/release/join cleanup.
Package-boundary guidance was unnecessary because package/import responsibility did not change. No other workflow bodies read.

## Decisions and observed behavior
One scheduler owns the reservation count; each admitted goroutine owns Open, optional Run, and Close. Completion frees its reservation only after acquisition failure or completed Close. One derived context stops cooperating work immediately on failure; a mutex protects errors. No cohort barrier or extra framework.
Admission rechecks stop. A lease acquired after stopping skips Run but closes once. Every owned lease waits for its own Run, and completed cleanup frees capacity despite unrelated running work.
Only exact Open/Run worker-context errors acknowledging an existing stop are suppressed while the caller remains uncanceled. Wrapped/joined independent failures and every Close failure are retained. Caller cancellation during stopping retains standard classification and custom cause. Complete uncanceled success returns nil; the caller owns jobs.
Tests observe sparse first-job progress, independent release, bounded reservations through held Close, all three failure stages, stopped admission, held Run cleanup, late acquisition ownership, independent failures during coordinated stopping, separate error combinations, noncomparable errors/causes, canceled entry, expired deadline, and accepted work before sparse-input cancellation.
Events establish readiness; 30 ms negative windows observe forbidden early completion/admission only after callbacks are known held. Cleanup cancels, releases gates, and joins; callbacks never make fatal test assertions.

## Exact checks
All shell commands used `rtk`; Go checks used `GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache`.
- `rtk gofmt -w workers.go workers_test.go`: passed twice.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...`: initial run interrupted (130) after discovering a full-buffer send in test setup; fixed the fixture before bounded verification.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -timeout=30s ./...`: three completed passes, 0.214s, 0.369s, then restored-source cached pass.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...`: three passes, no diagnostics.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go version`: go1.26.5 darwin/arm64.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -timeout=30s -run '^TestIndependentErrorCombinations$/^Close_only$' ./...`: temporary discarded-Close-error mutation compiled and failed (1), `Serve = <nil>, missing cause Close failure`; immediately restored, then same command passed (0.130s).
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -timeout=45s -run '^Test(SparseInputAndIndependentRelease|ReservationsIncludeHeldClose|FailuresStopAndJoin|LateAcquisitionIsClosedWithoutRun|IndependentErrorsDuringStop|CallerCancellationWhileStopping)$' ./...`: passed (1.381s), no race report.
- `rtk cat workers.go go.mod`: confirmed restored source and unchanged Go minimum.

## Limits
Executed on Go 1.26.5, not a separate Go 1.22 toolchain; source/test APIs stay within the declared minimum. Focused race checks and bounded negative windows do not prove every schedule. No expanded schedule/profile matrix or external integration. Callbacks must cooperate; held Close must finish before return. Concurrently admitted effects remain accepted; error order is unspecified.
