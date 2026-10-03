# Independent Go Quality Review — Six Bounded Worker Hosts

Each candidate is graded on its own exact supplied snapshot against the original README/host protocol, using the supplied unchanged Go Quality Report skill, all nine topic skills and references. These are code-area grades, including existing in-scope implementation and candidate regression-test defects. There is no combined grade, average, ranking, exposure or promotion judgment.

Coverage: all candidate files, original files, supplied held/supplementary source and checks-A..F were inspected. Seven topics are materially applicable per candidate; Security and Deployment are justified Not applicable within this trusted local library boundary. Unspecified external host/deployment/threat/profile evidence is unavailable outside this bounded scope and excluded, not labeled passing. Actual minimum and host toolchains, ordinary builds/vet/format/graph, authored suites and independent lifecycle checks were executed on disposable copies. Full raw commands/probes are linked below.

Outcome authority: exact candidate source and authored tests, original explicit fixture contract, unchanged review rubric, and independently executed evidence. Supplied checks remain supplied observations. Later guidance/evidence-disposition decisions are outside these frozen source-review outcomes. External review probes do not repair or become candidate-authored regression tests.

## Candidate A — C-

Scope: complete bounded code area at [exact hashes](../source-manifest.json); no applicable material coverage gap.

Rationale: The unchanged calculator selects C- from the deduplicated causes below. A production defect and an independently missing regression assertion are separate corrections; the same cause repeated across topics is counted once. All important failures are contained at this helper/test boundary, not critical/systemic.

Unique finding counts: critical=0, major=6, moderate=1, minor=0.

### Report cards

| Topic | Grade/state | Coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | C- | All bounded material obligations assessed | [card](cards/A/go-architecture-and-design.md) |
| Code Quality & Go Idioms | C- | All bounded material obligations assessed | [card](cards/A/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | C- | All bounded material obligations assessed | [card](cards/A/go-correctness-and-compatibility.md) |
| Testing | C- | All bounded material obligations assessed | [card](cards/A/go-testing.md) |
| Security | Not applicable | No bounded threat/sensitive sink decision | [card](cards/A/go-security.md) |
| Observability & Resilience | C- | All bounded material obligations assessed | [card](cards/A/go-observability-and-resilience.md) |
| Performance & Resource Management | C+ | All bounded material obligations assessed | [card](cards/A/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | All bounded material obligations assessed | [card](cards/A/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No configured release/process/runtime delivery decision | [card](cards/A/go-deployment-and-operations.md) |

### Good

- The supplied held and supplementary fixtures pass on both versions. Async Open is joined and acquired leases are closed after the actual Run results; pending cooperative Open is interrupted on an observed Run failure. The bounded stress evidence adds the uncommon post-Open ready-result case that one-shot checks missed.
- Actual minimum/host compile the unchanged Go 1.22 standard-only package in standalone readonly module mode; graph has no external module. Candidate gofmt/vet checks are clean.

### Bad

- [A/P1][major][production] Close failures are appended but never enter the stop state before the next cohort. `candidates/A/host.go:60`, `candidates/A/host.go:85` — At limit 1, three leases are opened/closed after the first Close fails, where the contract permits only one; unbounded input would continue admissions after observed failure. Contained important admission contract failure, not systemic/critical.
- [A/P2][major][production] Successful Open results can start Run using only stale stop state without rechecking caller/work context. `candidates/A/host.go:143`, `candidates/A/host.go:174` — Open deliberately cancels the caller context before returning its nonnil lease successfully. If its buffered result and cancellation are both ready, selecting the result starts Run after stopping; acquired lease must instead be owned and closed without Run. Reach is one host call; major contract failure.
- [A/P3][major][production] Final join/Close cleanup does not observe caller cancellation again before returning accumulated errors. `candidates/A/host.go:181`, `candidates/A/host.go:183` — After a Run failure, canceling the caller while a gated Close is held loses context.Canceled, although independent Run and Close errors remain inspectable. This breaks the explicit stopping-time diagnostic contract at the single Serve boundary.
- [A/T1][major][test] No authored Close-only failure test asserts that subsequent queued jobs are never opened. `candidates/A/host_test.go:107`, `candidates/A/host_test.go:140` — An important explicit stop-admission contract evades regression detection because the Close sentinel is paired with an already stopping Run failure. Independent test correction is required even after P1 is fixed.
- [A/T2][major][test] Cancellation-during-Open fixture asserts release but never asserts that returned leases skip Run. `candidates/A/host_test.go:151` — The explicit never-start-after-stop contract is unchecked despite a fixture already returning a lease after cancellation; a forbidden extra Run is accepted by the test.
- [A/T3][major][test] No authored assertion retains caller cancellation introduced during final held failure cleanup. `candidates/A/host_test.go:107`, `candidates/A/host_test.go:151` — Independent Run/Close error checks and initial cancellation checks do not cover cancellation observed while stopping, an explicit important diagnostic requirement.
- [A/T4][moderate][test] Failure fixture expects two acquired leases without ensuring second acquisition precedes the first Run failure. `candidates/A/host_test.go:107`, `candidates/A/host_test.go:135` — A legal early stop leaves only the first lease acquired and released; the test produces a false failure. This is a localized meaningful scheduling fragility, moderate rather than an unverified production defect.

### Suggested changes

- [A/P1] Have closeCohort set stop/cancel on any Close error, while still closing all owned leases exactly once. Why: At limit 1, three leases are opened/closed after the first Close fails, where the contract permits only one; unbounded input would continue admissions after observed failure. Contained important admission contract failure, not systemic/critical. Verify: Close-only sentinel with three queued jobs, limit 1: errors.Is cause true and Open/Close counts exactly one.
- [A/P2] After collecting/owning the Open result, observe context cancellation and set stop before starting Run; retain cancellation cause. Why: Open deliberately cancels the caller context before returning its nonnil lease successfully. If its buffered result and cancellation are both ready, selecting the result starts Run after stopping; acquired lease must instead be owned and closed without Run. Reach is one host call; major contract failure. Verify: Use callback cancellation-before-success with bounded concurrent pressure; assert zero Runs, one Close and errors.Is(context.Canceled).
- [A/P3] Append the caller cancellation observed after owned cleanup before returning, without dropping independent errors. Why: After a Run failure, canceling the caller while a gated Close is held loses context.Canceled, although independent Run and Close errors remain inspectable. This breaks the explicit stopping-time diagnostic contract at the single Serve boundary. Verify: Gate Close entry after Run failure, cancel caller, release Close, require errors.Is for Run, Close and cancellation.
- [A/T1] Add a healthy Run followed by failing Close with later queued jobs and an exact Open count. Why: An important explicit stop-admission contract evades regression detection because the Close sentinel is paired with an already stopping Run failure. Independent test correction is required even after P1 is fixed. Verify: The unchanged source must fail the intended assertion; a correct stop transition must pass without build/runner failures.
- [A/T2] Add a Run counter/forbidden event and assert zero Runs plus owned release; cover simultaneously ready Open-result/cancellation. Why: The explicit never-start-after-stop contract is unchecked despite a fixture already returning a lease after cancellation; a forbidden extra Run is accepted by the test. Verify: Bounded stress or event-gated ready result must reject P2 while checking exactly one Close.
- [A/T3] Gate final Close after an independent Run failure, cancel the caller, and assert all independent causes plus cancellation. Why: Independent Run/Close error checks and initial cancellation checks do not cover cancellation observed while stopping, an explicit important diagnostic requirement. Verify: The intended errors.Is assertion must fail unchanged A and pass corrected cleanup observation.
- [A/T4] Use an explicit second-acquisition/second-start gate before releasing the failing first Run; assert releases only for leases actually acquired. Why: A legal early stop leaves only the first lease acquired and released; the test produces a false failure. This is a localized meaningful scheduling fragility, moderate rather than an unverified production defect. Verify: Repeat the gated authored case under race/shuffle and verify one early acquired lease is allowed when no second acquisition was established.

### Limits

Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated. Candidate detail: The supplied held and supplementary fixtures pass on both versions. Async Open is joined and acquired leases are closed after the actual Run results; pending cooperative Open is interrupted on an observed Run failure. The bounded stress evidence adds the uncommon post-Open ready-result case that one-shot checks missed. Supplied D partial-start timeouts are setup-blocked events, not proof that an unstarted Run failed; D is graded on the independently confirmed late-input admission cause. All observed failing packages compiled and reached the recorded assertion; no unrelated build error is counted.

Details: [applicability](cards/A/applicability.json), [ledger](ledger-A.json), [calculator](ledger-A.grade.json), [canonical findings](findings.json).

## Candidate B — C-

Scope: complete bounded code area at [exact hashes](../source-manifest.json); no applicable material coverage gap.

Rationale: The unchanged calculator selects C- from the deduplicated causes below. A production defect and an independently missing regression assertion are separate corrections; the same cause repeated across topics is counted once. All important failures are contained at this helper/test boundary, not critical/systemic.

Unique finding counts: critical=0, major=6, moderate=0, minor=0.

### Report cards

| Topic | Grade/state | Coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | C- | All bounded material obligations assessed | [card](cards/B/go-architecture-and-design.md) |
| Code Quality & Go Idioms | C- | All bounded material obligations assessed | [card](cards/B/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | C- | All bounded material obligations assessed | [card](cards/B/go-correctness-and-compatibility.md) |
| Testing | C- | All bounded material obligations assessed | [card](cards/B/go-testing.md) |
| Security | Not applicable | No bounded threat/sensitive sink decision | [card](cards/B/go-security.md) |
| Observability & Resilience | C- | All bounded material obligations assessed | [card](cards/B/go-observability-and-resilience.md) |
| Performance & Resource Management | C- | All bounded material obligations assessed | [card](cards/B/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | All bounded material obligations assessed | [card](cards/B/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No configured release/process/runtime delivery decision | [card](cards/B/go-deployment-and-operations.md) |

### Good

- The full-cohort gate and Close completion ownership pass the held capacity case; already-canceled/input waiting stop and independent error identity have authored assertions. The pending-Open failure gap is expressly excluded from these positives.
- Actual minimum/host compile the unchanged Go 1.22 standard-only package in standalone readonly module mode; graph has no external module. Candidate gofmt/vet checks are clean.

### Bad

- [B/P1][major][production] Run errors cannot request stop while the controller is admitting input or blocked in synchronous Open. `candidates/B/host.go:39`, `candidates/B/host.go:48`, `candidates/B/host.go:61`, `candidates/B/host.go:86` — The first Run returns its independent error while Open 2 cooperatively waits for work context stop. The error remains buffered until admission completes, so Serve retains its owned lease indefinitely without a caller cancel. This is a contained important liveness/stop contract failure.
- [B/P2][major][production] Close errors never set failed, so the outer loop admits another cohort after observed Close failure. `candidates/B/host.go:99`, `candidates/B/host.go:105` — Three valid jobs are acquired instead of stopping after the first failing Close, while the error remains inspectable. Material admission contract break is confined to Serve.
- [B/P3][major][production] Open success launches Run unconditionally without checking cancellation that occurred inside Open. `candidates/B/host.go:57`, `candidates/B/host.go:59` — A lease returned successfully after callback cancellation still runs once; the fixture policy explicitly requires ownership/release with no Run. Contained important startup/stop contract break.
- [B/T1][major][test] Authored Run failure fixtures only observe failure after admission can finish, leaving pending-Open stop unverified. `candidates/B/host_test.go:111`, `candidates/B/host_test.go:145` — An important failure-containment/liveness contract is not detected: the cooperating pending Open needs caller cancellation despite Run failure.
- [B/T2][major][test] The Close-only regression case closes input after one job, so it cannot detect renewed admission after Close failure. `candidates/B/host_test.go:258` — The important stop-after-Close-failure contract remains unverified independently of error identity.
- [B/T3][major][test] The acquired-but-not-started test neither establishes a stop during Open nor checks zero Runs; it explicitly expects the second Run to start. `candidates/B/host_test.go:225`, `candidates/B/host_test.go:253` — The test name suggests a crucial ownership transition is protected, but the actual fixture allows the broken transition and checks only a normally started second Run.

### Suggested changes

- [B/P1] Let a failing Run cancel the cohort immediately, or observe Run outcomes concurrently with pending Open while collecting its returned ownership. Why: The first Run returns its independent error while Open 2 cooperatively waits for work context stop. The error remains buffered until admission completes, so Serve retains its owned lease indefinitely without a caller cancel. This is a contained important liveness/stop contract failure. Verify: Open 2 announces entry and waits for workCtx.Done; Run 1 fails after that event. Require autonomous stop, returned Run cause and exactly one Close. No caller cancel may be needed for progress.
- [B/P2] Mark failure/stop on Close error and finish all releases before returning. Why: Three valid jobs are acquired instead of stopping after the first failing Close, while the error remains inspectable. Material admission contract break is confined to Serve. Verify: Close-only sentinel, three jobs, limit 1; expect one Open, one Close and retained sentinel.
- [B/P3] Append the acquired lease, recheck cohort/caller stop, and skip Run while retaining the lease for release. Why: A lease returned successfully after callback cancellation still runs once; the fixture policy explicitly requires ownership/release with no Run. Contained important startup/stop contract break. Verify: Open cancels caller before success; assert zero Runs, exactly one Close and cancellation identity.
- [B/T1] Add a pending cooperative Open gated against a failing prior Run; assert it receives stop and Serve completes without caller cancel. Why: An important failure-containment/liveness contract is not detected: the cooperating pending Open needs caller cancellation despite Run failure. Verify: Intended stop-progress assertion must reject unchanged B on both supported test toolchains.
- [B/T2] Keep later jobs queued after the failing Close and assert one acquisition/release at limit 1. Why: The important stop-after-Close-failure contract remains unverified independently of error identity. Verify: Run unchanged B to verify intended Open-count failure, then corrected behavior to pass.
- [B/T3] Add an Open callback that cancels before returning success; assert zero second Runs and release of every acquired lease after joining existing Runs. Why: The test name suggests a crucial ownership transition is protected, but the actual fixture allows the broken transition and checks only a normally started second Run. Verify: Use the cancellation-before-success probe to reject P3 and keep ordinary peer failure coverage separately.

### Limits

Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated. Candidate detail: The full-cohort gate and Close completion ownership pass the held capacity case; already-canceled/input waiting stop and independent error identity have authored assertions. The pending-Open failure gap is expressly excluded from these positives. Supplied D partial-start timeouts are setup-blocked events, not proof that an unstarted Run failed; D is graded on the independently confirmed late-input admission cause. All observed failing packages compiled and reached the recorded assertion; no unrelated build error is counted.

Details: [applicability](cards/B/applicability.json), [ledger](ledger-B.json), [calculator](ledger-B.grade.json), [canonical findings](findings.json).

## Candidate C — A

Scope: complete bounded code area at [exact hashes](../source-manifest.json); no applicable material coverage gap.

Rationale: No confirmed actionable cause; verified API, lifecycle/test controls and actual minimum standalone build support A. No separate nonroutine safeguard pair is asserted for A+.

Unique finding counts: critical=0, major=0, moderate=0, minor=0.

### Report cards

| Topic | Grade/state | Coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | A | All bounded material obligations assessed | [card](cards/C/go-architecture-and-design.md) |
| Code Quality & Go Idioms | A | All bounded material obligations assessed | [card](cards/C/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | A | All bounded material obligations assessed | [card](cards/C/go-correctness-and-compatibility.md) |
| Testing | A | All bounded material obligations assessed | [card](cards/C/go-testing.md) |
| Security | Not applicable | No bounded threat/sensitive sink decision | [card](cards/C/go-security.md) |
| Observability & Resilience | A | All bounded material obligations assessed | [card](cards/C/go-observability-and-resilience.md) |
| Performance & Resource Management | A | All bounded material obligations assessed | [card](cards/C/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | All bounded material obligations assessed | [card](cards/C/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No configured release/process/runtime delivery decision | [card](cards/C/go-deployment-and-operations.md) |

### Good

- All authored, held, supplementary and six independent contract probes pass on both actual versions; the host race/shuffle/count=3 pass exercises held cleanup/capacity. The remove-worker-stop diagnostic mutant is detected by the authored TestServeFailureJoinsBeforeCloseAndRetainsErrors cleanupStarted gate on both versions, so that proposed test defect is rejected.
- Actual minimum/host compile the unchanged Go 1.22 standard-only package in standalone readonly module mode; graph has no external module. Candidate gofmt/vet checks are clean.

### Bad

- None found.

### Suggested changes

- None needed.

### Limits

Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated. Candidate detail: All authored, held, supplementary and six independent contract probes pass on both actual versions; the host race/shuffle/count=3 pass exercises held cleanup/capacity. The remove-worker-stop diagnostic mutant is detected by the authored TestServeFailureJoinsBeforeCloseAndRetainsErrors cleanupStarted gate on both versions, so that proposed test defect is rejected. Supplied D partial-start timeouts are setup-blocked events, not proof that an unstarted Run failed; D is graded on the independently confirmed late-input admission cause. All observed failing packages compiled and reached the recorded assertion; no unrelated build error is counted.

Details: [applicability](cards/C/applicability.json), [ledger](ledger-C.json), [calculator](ledger-C.grade.json), [canonical findings](findings.json).

## Candidate D — C-

Scope: complete bounded code area at [exact hashes](../source-manifest.json); no applicable material coverage gap.

Rationale: The unchanged calculator selects C- from the deduplicated causes below. A production defect and an independently missing regression assertion are separate corrections; the same cause repeated across topics is counted once. All important failures are contained at this helper/test boundary, not critical/systemic.

Unique finding counts: critical=0, major=4, moderate=0, minor=0.

### Report cards

| Topic | Grade/state | Coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | C- | All bounded material obligations assessed | [card](cards/D/go-architecture-and-design.md) |
| Code Quality & Go Idioms | C | All bounded material obligations assessed | [card](cards/D/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | C- | All bounded material obligations assessed | [card](cards/D/go-correctness-and-compatibility.md) |
| Testing | C- | All bounded material obligations assessed | [card](cards/D/go-testing.md) |
| Security | Not applicable | No bounded threat/sensitive sink decision | [card](cards/D/go-security.md) |
| Observability & Resilience | C- | All bounded material obligations assessed | [card](cards/D/go-observability-and-resilience.md) |
| Performance & Resource Management | C | All bounded material obligations assessed | [card](cards/D/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | All bounded material obligations assessed | [card](cards/D/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No configured release/process/runtime delivery decision | [card](cards/D/go-deployment-and-operations.md) |

### Good

- The prefilled two-job batch is concurrent, the cohort WaitGroup join precedes any release, and every successful Open result is collected/closed even when another Open fails. Actual minimum/host build and error identity checks pass; no sparse-input concurrency success is inferred.
- Actual minimum/host compile the unchanged Go 1.22 standard-only package in standalone readonly module mode; graph has no external module. Candidate gofmt/vet checks are clean.

### Bad

- [D/P1][major][production] Nonblocking job snapshot fixes a potentially one-job cohort and prevents admission of later available jobs while Run/Close proceeds. `candidates/D/host.go:43`, `candidates/D/host.go:73`, `candidates/D/host.go:139` — After job 1 starts on open input, job 2 becomes available while capacity 2 has one owned lease. Its sender and Run cannot progress until job 1 is released. This is serial execution under sparse input, forbidden by the explicit fixture contract. The accepted batch policy does not permit leaving an available second job blocked behind a one-job batch. One contained important admission/concurrency cause; setup timeouts are not four separate defects.
- [D/P2][major][production] Close results are accumulated in result but never mark the cohort failed. `candidates/D/host.go:156`, `candidates/D/host.go:161` — The next cohort is admitted after an observed Close error; retained diagnostic identity does not enforce the explicit stop-admission contract. Failure is contained in the helper.
- [D/T1][major][test] Concurrency tests preload closed full batches and the partial-open fixture forbids any Run, leaving sparse-input concurrent admission unverified. `candidates/D/host_test.go:29`, `candidates/D/host_test.go:109` — The important capacity/concurrency contract is verified only for an already full buffered batch; the implementation can serialize ordinary unbuffered input while these tests pass.
- [D/T2][major][test] The error-retention fixture has one job and combined Run/Close failure, so Close-only stop-admission is unverified. `candidates/D/host_test.go:74` — The important Close-stage admission boundary can regress while existing independent error-identity checks remain green.

### Suggested changes

- [D/P1] Keep admission responsive while capacity remains and successful acquired jobs can run; retain the cohort join/Close barrier at release. Why: After job 1 starts on open input, job 2 becomes available while capacity 2 has one owned lease. Its sender and Run cannot progress until job 1 is released. This is serial execution under sparse input, forbidden by the explicit fixture contract. The accepted batch policy does not permit leaving an available second job blocked behind a one-job batch. One contained important admission/concurrency cause; setup timeouts are not four separate defects. Verify: First job arrives and starts; while its Run is gated, make job 2 available. Require job 2 to start before first release, then release/join both and close owned leases.
- [D/P2] Set failed/cancel after any Close error while finishing every acquired release, then stop admission. Why: The next cohort is admitted after an observed Close error; retained diagnostic identity does not enforce the explicit stop-admission contract. Failure is contained in the helper. Verify: Healthy Run, failing Close, later queued jobs, limit 1: expect one Open/Close and errors.Is sentinel.
- [D/T1] Add separately arriving jobs on caller-owned open input; gate first Run, require second Run with available capacity, and check partial startup cleanup separately. Why: The important capacity/concurrency contract is verified only for an already full buffered batch; the implementation can serialize ordinary unbuffered input while these tests pass. Verify: Use positive first-start and actual second availability; no claim that every Open must complete in a particular relative order.
- [D/T2] Add healthy Run plus independent Close failure with remaining jobs and exact acquired/released counts. Why: The important Close-stage admission boundary can regress while existing independent error-identity checks remain green. Verify: Intended count assertion rejects unchanged D and accepts corrected stop transition.

### Limits

Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated. Candidate detail: The prefilled two-job batch is concurrent, the cohort WaitGroup join precedes any release, and every successful Open result is collected/closed even when another Open fails. Actual minimum/host build and error identity checks pass; no sparse-input concurrency success is inferred. Supplied D partial-start timeouts are setup-blocked events, not proof that an unstarted Run failed; D is graded on the independently confirmed late-input admission cause. All observed failing packages compiled and reached the recorded assertion; no unrelated build error is counted.

Details: [applicability](cards/D/applicability.json), [ledger](ledger-D.json), [calculator](ledger-D.grade.json), [canonical findings](findings.json).

## Candidate E — C

Scope: complete bounded code area at [exact hashes](../source-manifest.json); no applicable material coverage gap.

Rationale: The unchanged calculator selects C from the deduplicated causes below. A production defect and an independently missing regression assertion are separate corrections; the same cause repeated across topics is counted once. All important failures are contained at this helper/test boundary, not critical/systemic.

Unique finding counts: critical=0, major=1, moderate=0, minor=0.

### Report cards

| Topic | Grade/state | Coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | A | All bounded material obligations assessed | [card](cards/E/go-architecture-and-design.md) |
| Code Quality & Go Idioms | A | All bounded material obligations assessed | [card](cards/E/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | A | All bounded material obligations assessed | [card](cards/E/go-correctness-and-compatibility.md) |
| Testing | C | All bounded material obligations assessed | [card](cards/E/go-testing.md) |
| Security | Not applicable | No bounded threat/sensitive sink decision | [card](cards/E/go-security.md) |
| Observability & Resilience | A | All bounded material obligations assessed | [card](cards/E/go-observability-and-resilience.md) |
| Performance & Resource Management | A | All bounded material obligations assessed | [card](cards/E/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | All bounded material obligations assessed | [card](cards/E/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No configured release/process/runtime delivery decision | [card](cards/E/go-deployment-and-operations.md) |

### Good

- All unchanged authored, held, supplementary and independent probes pass on both versions; host race/shuffle/count=3 passes. Source correctly retains held cleanup/cancellation, skips Run after canceled acquisition, interrupts pending cooperative Open from worker failure, and stops on Close failure. The surviving isolated Close-stop mutant is a Testing finding only.
- Actual minimum/host compile the unchanged Go 1.22 standard-only package in standalone readonly module mode; graph has no external module. Candidate gofmt/vet checks are clean.

### Bad

- [E/T1][major][test] Close-only failure fixtures cannot distinguish retaining the error from stopping future admission. `candidates/E/host_test.go:310`, `candidates/E/host_test.go:345` — The source implements the important stop-after-Close contract correctly, but a compiled mutation removing only that transition evades all authored tests. The held-Close case also cancels the caller, masking the independent Close-only cause. This is a major verification gap requiring its own regression assertion, not a production defect.

### Suggested changes

- [E/T1] Add a healthy Run and failing Close at limit 1 with later jobs queued; assert no further Open while retaining Close error. Why: The source implements the important stop-after-Close contract correctly, but a compiled mutation removing only that transition evades all authored tests. The held-Close case also cancels the caller, masking the independent Close-only cause. This is a major verification gap requiring its own regression assertion, not a production defect. Verify: Current E must pass; the isolated mutation must fail the intended admission-count assertion. Unrelated compile/runner failure is not detection.

### Limits

Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated. Candidate detail: All unchanged authored, held, supplementary and independent probes pass on both versions; host race/shuffle/count=3 passes. Source correctly retains held cleanup/cancellation, skips Run after canceled acquisition, interrupts pending cooperative Open from worker failure, and stops on Close failure. The surviving isolated Close-stop mutant is a Testing finding only. Supplied D partial-start timeouts are setup-blocked events, not proof that an unstarted Run failed; D is graded on the independently confirmed late-input admission cause. All observed failing packages compiled and reached the recorded assertion; no unrelated build error is counted.

Details: [applicability](cards/E/applicability.json), [ledger](ledger-E.json), [calculator](ledger-E.grade.json), [canonical findings](findings.json).

## Candidate F — C-

Scope: complete bounded code area at [exact hashes](../source-manifest.json); no applicable material coverage gap.

Rationale: The unchanged calculator selects C- from the deduplicated causes below. A production defect and an independently missing regression assertion are separate corrections; the same cause repeated across topics is counted once. All important failures are contained at this helper/test boundary, not critical/systemic.

Unique finding counts: critical=0, major=6, moderate=0, minor=0.

### Report cards

| Topic | Grade/state | Coverage | Card |
| --- | --- | --- | --- |
| Architecture & Design | C- | All bounded material obligations assessed | [card](cards/F/go-architecture-and-design.md) |
| Code Quality & Go Idioms | C- | All bounded material obligations assessed | [card](cards/F/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | C- | All bounded material obligations assessed | [card](cards/F/go-correctness-and-compatibility.md) |
| Testing | C- | All bounded material obligations assessed | [card](cards/F/go-testing.md) |
| Security | Not applicable | No bounded threat/sensitive sink decision | [card](cards/F/go-security.md) |
| Observability & Resilience | C- | All bounded material obligations assessed | [card](cards/F/go-observability-and-resilience.md) |
| Performance & Resource Management | C- | All bounded material obligations assessed | [card](cards/F/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | All bounded material obligations assessed | [card](cards/F/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No configured release/process/runtime delivery decision | [card](cards/F/go-deployment-and-operations.md) |

### Good

- The full-initial-cohort WaitGroup join and release checks pass when Runs ignore cancellation; wrapped Open/Run/Close causes retain errors.Is identity with job context. This does not validate healthy cooperative completion or paced-input ownership capacity.
- Actual minimum/host compile the unchanged Go 1.22 standard-only package in standalone readonly module mode; graph has no external module. Candidate gofmt/vet checks are clean.

### Bad

- [F/P1][major][production] The cohort context is unconditionally canceled when admission ends, even for a healthy full cohort or normal input closure. `candidates/F/host.go:93`, `candidates/F/host.go:94` — Healthy cooperative Runs are stopped before their requested work completes, and completed uncanceled input returns a failure. This is a contained important success/lifecycle contract failure.
- [F/P2][major][production] A successful Open always calls Run without checking that stopping happened during acquisition. `candidates/F/host.go:84`, `candidates/F/host.go:85` — The explicit fixture requires a lease acquired before stop to remain owned but unstarted; F starts it after caller cancellation and later closes it. One contained important startup/stop failure.
- [F/P3][major][production] Capacity is decremented on Run completion while the acquired lease remains unreleased until later cohort Close. `candidates/F/host.go:38`, `candidates/F/host.go:61`, `candidates/F/host.go:103` — Eight paced successful jobs on open input retain eight acquired leases at capacity 2, with no Close until input ends. Peak is 4x declared bound for eight jobs; for a stream of completed Runs retention grows with jobs rather than limit. This is a contained major resource and lifecycle-contract failure; no unobserved system exhaustion is claimed.
- [F/T1][major][test] Healthy capacity fixtures ignore their Run contexts, so premature coordinated stop is invisible. `candidates/F/host_test.go:33`, `candidates/F/host_test.go:78`, `candidates/F/host_test.go:148` — The important uncanceled success contract is tested with callbacks that keep succeeding after any stop request. Their gates verify sequencing while masking whether work was allowed to complete cooperatively.
- [F/T2][major][test] Authored cancellation coverage does not exercise an acquired success returned after stopping and assert no Run. `candidates/F/host_test.go:135`, `candidates/F/host_test.go:259` — Input-wait/already-canceled tests do not protect the important partial acquisition ownership transition.
- [F/T3][major][test] Capacity tests inspect only a full initial cohort and later Run starts, without counting live acquired leases across completed Runs on open input. `candidates/F/host_test.go:33`, `candidates/F/host_test.go:78`, `candidates/F/host_test.go:148` — An explicit important capacity bound through Close is effectively unverified for a stream of successful completions; Run concurrency is insufficient as its proxy.

### Suggested changes

- [F/P1] Cancel only for an observed failure/caller stop; join normal Runs without canceling their work, then release leases. Why: Healthy cooperative Runs are stopped before their requested work completes, and completed uncanceled input returns a failure. This is a contained important success/lifecycle contract failure. Verify: Two healthy Runs announce entry and wait for an external success gate. Before release neither work context may be canceled; after release Serve must return nil.
- [F/P2] Record acquired lease ownership, observe stop before Run entry, and skip Run while joining/releasing all prior work. Why: The explicit fixture requires a lease acquired before stop to remain owned but unstarted; F starts it after caller cancellation and later closes it. One contained important startup/stop failure. Verify: Callback cancels before success; require cancellation cause, zero Runs, exactly one Close.
- [F/P3] Track acquired leases and pending reservations through completed Close, and stop admitting once those owned slots reach limit; completion alone cannot release capacity. Why: Eight paced successful jobs on open input retain eight acquired leases at capacity 2, with no Close until input ends. Peak is 4x declared bound for eight jobs; for a stream of completed Runs retention grows with jobs rather than limit. This is a contained major resource and lifecycle-contract failure; no unobserved system exhaustion is claimed. Verify: Pace valid jobs by observable Run-return events, count leases on successful Open and decrement only at Close completion; require peak<=2 and all eight successful effects/releases.
- [F/T1] Make healthy callbacks verify they are not canceled before their external success gate and return ctx.Err when improperly stopped. Why: The important uncanceled success contract is tested with callbacks that keep succeeding after any stop request. Their gates verify sequencing while masking whether work was allowed to complete cooperatively. Verify: Unchanged F must fail the intended healthy-work cancellation/success assertions.
- [F/T2] Return a successful lease after callback cancellation, assert zero Runs, and still require owned Close after cohort join. Why: Input-wait/already-canceled tests do not protect the important partial acquisition ownership transition. Verify: Reject P2 via the intended zero-Run assertion on a compiled package.
- [F/T3] Track live acquisition/release counts and use paced successful jobs on open input to assert the lease bound, plus held Close. Why: An explicit important capacity bound through Close is effectively unverified for a stream of successful completions; Run concurrency is insufficient as its proxy. Verify: An implementation that releases active capacity at Run return must fail the intended peak-owned-leases assertion.

### Limits

Review is a complete bounded code-area assessment, not a whole deployed system or proof of all interleavings. Exact commands/stdout/stderr are preserved in evidence/commands.json; original supplied checks are separately in checks-A..F.json. Host is go1.26.5 darwin/arm64; actual minimum is go1.22.12 darwin/arm64. Standalone readonly module mode, isolated Go cache, no source/config changes. No arbitrary panic/nil-context/nonpositive-limit/contract-violating Open behavior is required. No production load/profile, unspecified deployment or threat-model claim. Detection counts intended behavioral assertions only; unrelated build failures and setup-blocked fixture events are separated. Candidate detail: The full-initial-cohort WaitGroup join and release checks pass when Runs ignore cancellation; wrapped Open/Run/Close causes retain errors.Is identity with job context. This does not validate healthy cooperative completion or paced-input ownership capacity. Supplied D partial-start timeouts are setup-blocked events, not proof that an unstarted Run failed; D is graded on the independently confirmed late-input admission cause. All observed failing packages compiled and reached the recorded assertion; no unrelated build error is counted.

Details: [applicability](cards/F/applicability.json), [ledger](ledger-F.json), [calculator](ledger-F.grade.json), [canonical findings](findings.json).

## Preserved verification

- [Independent command/stdout/stderr record](evidence/commands.json): host/minimum authored, independent, held/supplementary tests; host race/shuffle/count=3; vet/build/gofmt; actual minimum module graph.
- [Additional bounded repetitions](evidence/additional-commands.json), [A race/stress record](evidence/stress-commands.json), [actual minimum A stress](evidence/minimum-stress-command.json).
- [Isolated mutation checks](evidence/mutation-commands.json): C worker-stop mutant detected at intended authored cleanup gate; E Close-stop mutant survives authored tests but fails the intended independent admission assertion. Compile/runner errors do not count as detection.
- [Portable independent probe source](probes/independent_contract_test.go), [A stress source](probes/stress-A_test.go), [probe intent/disposition](probes/README.md).
- [All calculator invocations](evidence/grade-calculator-commands.json), [all grades](grades.json), [manifest](manifest.json).

Reviews were performed sequentially by one independent reviewer with no delegation or inspection of repository planning, candidate identity/exposure reports, other agents or prior review artifacts. No implementation/guidance edit or publication action was performed.
