# Guided independent-lease implementation trial

Date: 2026-10-03.
Scope: this workspace's README contract, Serve, and focused tests only.
Artifacts: workers.go, workers_test.go, report.md. go.mod is unchanged.

## Guidance selected

Read all seven supplied build-skill bodies from concurrency-guidance.
Applied go-concurrency-and-ownership to reservation, admission, supervision,
per-lease release, stop-before-cleanup, and joining every admitted lifetime.
Applied go-context-and-deadlines to caller classification and custom causes.
Applied go-api-contracts and go-interfaces-and-composition to preserve the
existing Job, Lease, Serve, package, callback, and ownership contracts.
Applied go-behavior-tests and go-test-isolation to event-controlled callbacks,
independent errors, held cleanup, bounded waits, and cancel/release/join cleanup.
Read the supplied behavior-observations and isolation-patterns references.
Package-boundary guidance confirmed no package movement was needed.

## Decisions and implementation

Use limit fixed workers and one derived context; each worker owns at most one
reservation from admission through completed Close, then reuses its own slot.
Check stopping before admission and before starting a newly acquired lease.
Report and cancel upon independent failure before entering potentially held Close.
WaitGroup joins every worker; acquired leases close exactly once after Run returns,
including successful acquisitions whose Run is skipped because stopping won.
A mutex protects only collected errors and the once-recorded caller cancellation.
All Open and Close failures remain independent, including cancellation-shaped errors.
Only Run's exact worker-context standard error, with work stopping and caller live,
is treated as a coordinated-stop acknowledgement; wrapped/joined errors remain.
Caller cancellation retains both ctx.Err and context.Cause without comparing causes.
No new dependencies, exported hooks, framework, recovery policy, or module change.

## Verification and limits

All checks used GOWORK=off and GOCACHE=/private/tmp/go-skill-recovery-cache:
- rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -count=1 -timeout=45s ./...: PASS (0.538s).
- rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...: PASS, no diagnostics.
- rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -count=1 -timeout=45s ./...: PASS (1.655s).
- rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -count=1 -timeout=30s -run '^TestLateAcquisitionClosesWithoutRun$' ./...: PASS (0.283s).

Seven tests cover ordinary success, sparse input, reservation bounds through held
Close, independent slot reuse, stop before held Run/Close cleanup, late acquisition,
Open/Run/Close cause retention, caller cancellation during cleanup, and canceled entry.
Fixtures release gates and join on early test exits; caller channel remains usable.
Toolchain: go1.26.5 darwin/arm64. Go1.22 minimum is preserved but not run directly.
Race checks and short absence observations do not establish every possible schedule.
No broad scenario/profile/mutation matrix or external consumer integration was run.
