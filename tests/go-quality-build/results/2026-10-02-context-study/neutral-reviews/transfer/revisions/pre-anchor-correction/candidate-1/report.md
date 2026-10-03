# Go Quality Report — C-

Scope: Candidate 1. Changeset review of the provided immutable original to candidate snapshot; library plus supplied CLI host. Only run.go, run_test.go and newly added cmd/finalize/main_test.go differ. README.md, go.mod and cmd/finalize/main.go are unchanged contract/caller context. Exact SHA-256 snapshots are in ../source-snapshots.json. Declared Go 1.22; independently executed Go 1.22.12 and Go 1.26.5, darwin/arm64, GOTOOLCHAIN=local, GOWORK=off.

Coverage: All nine topics considered; eight applicable topic boundaries complete, Security justified Not applicable. Boundary ownership and exact paths are in [manifest](manifest.json).

Rationale: One contained major production error-retention failure plus three independently actionable moderate testing issues selects C-. The same production cause is deduplicated across Correctness, Code Quality and Resilience. Arithmetic is checked by the frozen [grade.py](../../review-skills/go-quality-report/scripts/grade.py), with [ledger](ledger.json) and [raw grade output](grade.json).

Unique finding counts: critical=0, major=1, moderate=3, minor=0

## Report cards

| Topic | Grade/state | Assessed coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | A | Callback/API composition and Run-to-command lifecycle ownership; supplied host preserved. | [architecture](architecture.md) |
| Code Quality & Go Idioms | C | Changed Run error flow/helpers, test readability, Go 1.22 API availability, formatting and vet. | [code-quality](code-quality.md) |
| Correctness & Compatibility | C | Sequential acceptance, stop conditions, empty/unaccepted work, late successful apply, cancellation classification/cause, independent failures, finalization effects, and child status/output/receipt. | [correctness](correctness.md) |
| Testing | C+ | Both authored test files, real process boundaries, portable assertions, regression detection, case lifetime, and failure-path cleanup. | [testing](testing.md) |
| Security | Not applicable | Conditional applicability only: trusted cooperative callbacks, caller-owned local CLI inputs, unchanged receipt sink/permissions, no changed attacker or authorization boundary. | [security](security.md) |
| Observability & Resilience | C | Cancellation/failure diagnosis, independent finalization deadline, synchronous drain before return/exit, and required receipt failure handling. | [resilience](resilience.md) |
| Performance & Resource Management | A | Run complexity, finalization timer/context ownership, synchronous callback completion and release; test fixture ownership is assigned to Testing. | [resources](resources.md) |
| Dependencies & Reproducibility | A | Standard-library-only module, unchanged go 1.22 directive, selected compiler and nested process-build compiler, standalone readonly builds on minimum/host. | [dependencies](dependencies.md) |
| Deployment & Operations | A | Local CLI artifact build, existing runtime configuration/signals, accepted-prefix receipt effects and exit 2 after failure/interruption. No unrelated delivery-system audit. | [deployment](deployment.md) |

## Good

- [G1] Both actual compiler supplied-probe runs independently verify cancellation-surviving bounded finalization, cooperative completion, exit 2, accepted stdout and receipt after interruption.
- [G2] Source inspection and supplementary captured-scope diagnostics verify cancellation resource release after completion. No unowned production goroutine is introduced.

## Bad

- [F1][major] Run treats errors.Is(applyErr, ctxErr) as proof that applyErr contains no independent failure and discards the entire apply error tree. Evidence: candidate-1/run.go:16-21. Reviewer diagnostic returns errors.Join(workFailure, context.Canceled) after custom parent cancellation, then an independent finalization failure: accepted=1, finalCalls=1, canceled=true, parentCause=true, workFailure=false, finalFailure=true. Fails independently on actual Go 1.22.12 and 1.26.5; authored/unchanged supplied probes otherwise pass.
- [T1][moderate] The interrupted process test requires Go 1.26-specific signal cause text that is not part of the supplied stderr contract. Evidence: candidate-1/cmd/finalize/main_test.go:154-155 expects context canceled then interrupt signal received. Actual Go 1.22.12 authored suite fails with context canceled twice, while exit/output/receipt assertions preceding this check pass and the supplied minimum-version process probes pass. Raw minimum authored and probe outcomes are separate.
- [T2][moderate] Authored error-preservation coverage exercises plain processing/finalization errors and a between-jobs cancellation, but never an apply error tree carrying both cancellation and an independent failure. Evidence: candidate-1/run_test.go:19-79. Authored host tests and unchanged supplied probes pass despite F1; the narrow supplementary diagnostic fails on both compilers and demonstrates the omitted normal cancellation/error combination.
- [T3][moderate] The finalization barrier test releases its worker only on the success path, so t.Fatal on the startup/completion checks can abandon owned work. Evidence: candidate-1/run_test.go:22-63: finalizer blocks on release at line 45; close(release) occurs only at line 59; startup timeout and early-return fatal branches precede it without cleanup. A finalizer that entered but the test misses readiness before its guard expires remains blocked after the test fails. This is source-traced failure-path ownership, not an independently observed leak in the passing immutable suite.

## Suggested changes

- [F1] Always preserve the full apply error while also retaining observed cancellation classification/cause; deduplicate only if it can preserve the complete original error tree. Add the joined cancellation plus independent callback-failure case. Owner: Run error aggregation.
- [T1] Assert the contractual cancellation diagnostic without fixing signal-context implementation wording, and run the child binary with the actual declared-minimum compiler. Owner: Command test assertions.
- [T2] Add a behavioral case that cancels during apply, returns errors.Join(independentWorkFailure, context.Canceled), and asserts accepted count plus errors.Is for classification, parent cause, independent work failure and independent finalization failure. Owner: Run regression tests.
- [T3] Give the release signal one cleanup owner on every exit path and cancel/join the test worker under a bounded controller. Keep explicit completion assertions. Owner: Run test fixture lifecycle.

## Limits

Review was sequential within this neutral packet; no mapping, repository design/results, author skills, or other arms were inspected. Authored sources and supplied probes were not modified. Windows interruption is explicitly skipped by both authored suites and the supplied probe; execution here covers darwin/arm64 only. Cooperative nonnil callbacks, nonnil parent context and positive budget are contractual preconditions. No crash, hostile-input, durability/atomicity, packaging/CI, or arbitrary-platform promises were invented. Supplementary reviewer diagnostics are not frozen efficacy tests.

Independent checks: host authored suite passes; minimum authored suite fails only the interrupted stderr assertion; standalone -mod=readonly build and vet pass on both compilers; gofmt -d is empty; supplied unchanged contract probes pass on both compilers, including real interrupted child status/output/receipt and real receipt-write failure; host -race -shuffle=on -count=3 suite with unchanged probes passes. Supplementary combined-error diagnostic fails on both compilers while finalization-scope release passes. Supplied check facts are in [provided checks](../../provided-checks-1.json); executed raw commands/results are separate ../candidate-1-*.json artifacts. The error-tree and context-release diagnostic is [supplementary reviewer source](../diagnostic_contract_test.go), and is excluded from frozen authored efficacy. Failure-path lifetime findings use explicit source tracing; candidate 2's package-timeout sensitivity is supplied evidence, not an independent rerun. No unrelated future regression search was performed.

Details: [manifest](manifest.json), [deduplicated findings](findings.json), [source snapshots](../source-snapshots.json), [source preservation](../source-preservation.json).
