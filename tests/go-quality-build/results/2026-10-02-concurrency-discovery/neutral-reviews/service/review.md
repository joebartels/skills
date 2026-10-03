# Go Quality Report — C-

Scope: Candidate code-area outcome review of [host.go](/private/tmp/go-independent-concurrency-service/candidate/host.go), [host_test.go](/private/tmp/go-independent-concurrency-service/candidate/host_test.go), and [go.mod](/private/tmp/go-independent-concurrency-service/candidate/go.mod) against [original/README.md](/private/tmp/go-independent-concurrency-service/original/README.md). Original source supplies comparison context; this is an implementation outcome review, not a skill or promotion review. Candidate host.go SHA-256: `18b161e49c3ddfde132e5c0e634b75bb7b461442b4c77d88c035fe823c84ec26`.

Coverage: Seven applicable topics assessed completely; Security and Deployment & Operations are justified Not applicable. Admission/progress, acquired capacity through Close, acquisition/execution supervision, failure propagation, stop versus actual Run completion, acquired-but-never-run cleanup, error identity, successful effects, caller-owned input and the unchanged public lifecycle API were inspected or exercised. Candidate hashes before and after verification match. No further subagents were used.

Rationale: The candidate does not satisfy the requested Serve contract. Two independent, contained major source failures can stall ordinary supported calls; three separately necessary test changes protect important contracts the authored suite currently misses. The five unique major causes select C- under the unchanged rubric. Shared source causes appear in several topic cards but count once overall. No critical or systemic reach is claimed and no grades are averaged.

Unique finding counts: critical=0, major=5, moderate=0, minor=0.

## Report cards

| Topic | Grade/state | Assessed coverage and limits | Card |
| --- | --- | --- | --- |
| Architecture & Design | C- | Admission and concurrent phase supervision; all lifecycle boundaries assessed. | [architecture-and-design](/private/tmp/go-independent-concurrency-service/cards/architecture-and-design.md) |
| Code Quality & Go Idioms | C | Error-flow timing plus formatting/vet and supported idioms. | [code-quality-and-idioms](/private/tmp/go-independent-concurrency-service/cards/code-quality-and-idioms.md) |
| Correctness & Compatibility | C- | Complete Serve outcome, cancellation, joins, errors and source API. | [correctness-and-compatibility](/private/tmp/go-independent-concurrency-service/cards/correctness-and-compatibility.md) |
| Testing | C- | Candidate-authored regression signal; external controller checks separately evidenced. | [testing](/private/tmp/go-independent-concurrency-service/cards/testing.md) |
| Security | Not applicable | No security trust, authorization, secret, sink or attacker-abuse boundary implicated. | — |
| Observability & Resilience | C | Cooperative failure containment, cancellation and shutdown completion. | [observability-and-resilience](/private/tmp/go-independent-concurrency-service/cards/observability-and-resilience.md) |
| Performance & Resource Management | C- | Acquired-lease capacity, resource progress, actual joins and Close completion. | [performance-and-resource-management](/private/tmp/go-independent-concurrency-service/cards/performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | Standard library, unchanged Go 1.22 minimum; standalone build and both toolchains. | [dependencies-and-reproducibility](/private/tmp/go-independent-concurrency-service/cards/dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | No release, CI, manifest, runtime wiring or rollout decision in this library packet. | — |

## Good

- [G1] [host.go:82](/private/tmp/go-independent-concurrency-service/candidate/host.go:82) joins all started Runs before any acquired lease is closed. Held cleanup tests confirm stop acknowledgement alone does not release ownership. An acquired lease whose Run never starts is closed exactly once.
- [G2] [host.go:65](/private/tmp/go-independent-concurrency-service/candidate/host.go:65), lines 84 and 91, retain independent failures with `errors.Join`. A controlled Open failure, independent Run error, held Close error and caller cancellation during stopping remain inspectable through `errors.Is`.
- [G3] [host.go:89](/private/tmp/go-independent-concurrency-service/candidate/host.go:89) waits for Close completion before admitting another cohort. Independent and controller probes verify acquired-lease capacity 2, held Close ownership and accepted successful effects. The Job/Lease/Serve API and dependency injection parameter are preserved.
- [G4] Ordinary authored tests pass on Go 1.22.12 and 1.26.5; current-toolchain race/shuffle repetitions, go vet, gofmt and an offline standalone readonly build pass. The supplied release-before-completion mutation is meaningfully detected by both the author tests and the corresponding originally passing controller test.

## Bad

- [F1][major] [host.go:24-43](/private/tmp/go-independent-concurrency-service/candidate/host.go:24) waits for a full cohort or caller input closure before starting anything. One queued job, limit 2 and an open input channel leave both slots idle. A producer waiting for that job before sending the next can deadlock. This is also a regression from the original first-job start. Independent reproduction fails on both toolchains and repeated race/shuffle runs.
- [F2][major] [host.go:62-75](/private/tmp/go-independent-concurrency-service/candidate/host.go:62) reads only acquisition completions while started Run results wait in another channel; failure processing starts at [line 82](/private/tmp/go-independent-concurrency-service/candidate/host.go:82). A Run failure cannot cancel a cooperating pending Open that awaits stop, so Serve hangs until external caller cancellation and retains its acquired lease. Independent reproduction fails on both toolchains and repeated race/shuffle runs. The original also failed to propagate Run errors; this is an unmet candidate outcome, not a claim that all failure behavior was newly introduced.
- [T1][major] [host_test.go:39](/private/tmp/go-independent-concurrency-service/candidate/host_test.go:39), lines 100-133 and 136-167, cover preclosed batches or empty/canceled input, but never require sparse open input to start its available job. The authored suite passes with F1 present.
- [T2][major] [host_test.go:100-122](/private/tmp/go-independent-concurrency-service/candidate/host_test.go:100) lets the sibling Run finish immediately and has no pending-Open failure case. A separate mutation removing only Run-error `cancel()` passes authored tests under `-race -count=5`, while the independent waiting-sibling probe fails. Coordinated failure-to-stop propagation therefore lacks meaningful authored regression coverage.
- [T3][major] [host_test.go:48](/private/tmp/go-independent-concurrency-service/candidate/host_test.go:48), lines 53-69 and 95, count active Runs and immediate Close calls. Capacity is defined over acquired leases through completed Close. A separate mutation deferring successful-cohort Close until Serve returns passes authored tests under `-race -count=5` while the independent held-Close probe records three acquired leases at limit 2 and fails.

## Suggested changes

- [F1][P1] Serve admission/lifecycle implementation owner: start available jobs before waiting for future input, retaining bounded cohorts and the cohort join/release barrier. Verify one job on an open channel and a producer waiting for first completion.
- [F2][P1] Serve failure-controller owner: supervise acquisition and Run completion together, stop affected work on the first observed failure, retain all independent causes and drain actual completions before Close. Verify a failed Run with an Open blocked on the shared context, then repeat held-cleanup and never-started-lease probes.
- [T1][P1] Candidate test owner: add a sparse-open-input progress assertion that observes first Run before offering another job or canceling.
- [T2][P1] Candidate test owner: add controlled Run-failure cases with both a waiting sibling Run and a cooperating pending Open. Require stop propagation, actual cleanup joins, error identity and exactly-once release; reject the no-Run-stop mutation.
- [T3][P1] Candidate test owner: count leases from successful acquisition until Close returns, hold Close while another job is pending and include a never-started lease. Reject the deferred-Close mutation.

## Limits

All execution used disposable copies, with `GOWORK=off`, `GOTOOLCHAIN=local`, `GOPROXY=off` and a scratch build cache. Verified environments were Go 1.22.12 and Go 1.26.5 on darwin/arm64; no unsupported platform, profiler or bit-identical artifact claim is made. The source preconditions (nonnil context, positive limit, cooperative callbacks, successful nonnil leases) are authoritative. No unrelated permission, telemetry, process-budget or dependency framework requirement is added.

Executed checks and outcomes are captured verbatim in [executed-checks.json](/private/tmp/go-independent-concurrency-service/review-validation/executed-checks.json), with exact argv, cwd and environment overrides. `rtk proxy go test -count=1 -timeout=15s ./...`, `rtk proxy go test -race -shuffle=on -count=3 -timeout=20s ./...`, `rtk proxy go vet ./...`, `rtk proxy gofmt -d host.go host_test.go`, `rtk proxy go build -mod=readonly -o /dev/null ./...` and the ordinary Go 1.22.12 test pass on the baseline. The six independent probes fail only F1/F2 on both versions; four positive lifecycle/resource probes pass. Race/shuffle repeats the two same failures without race diagnostics. The two independent sensitivity mutations pass author tests and fail their targeted probes. The supplied release-before-completion mutation compiles and fails both authored and held join tests; its original named controller test passes.

The supplied open-input controller failure occurs at the **initial work-start event**, before its cancellation assertions execute. I inspected that source, reproduced the same initial-start failure, and independently confirmed F1 against the README. It does **not** prove that cancellation itself is unresponsive: the independent sparse-input probe also confirms cancellation returns while input stays open. The supplied unresponsive-admission mutation is already failing at baseline and receives **no sensitivity credit**. Candidate-authored testing is the graded code area; controller checks and newly written reviewer probes are external evidence, with their stronger coverage explicitly distinguished.

Details: [applicability](/private/tmp/go-independent-concurrency-service/applicability.json), [manifest](/private/tmp/go-independent-concurrency-service/manifest.json), [deduplicated findings and rejected claims](/private/tmp/go-independent-concurrency-service/findings.json), [grading ledger](/private/tmp/go-independent-concurrency-service/ledger.json), and [raw validation files](/private/tmp/go-independent-concurrency-service/review-validation). The grading calculator returns C-, five unique majors and no coverage gaps; it validates arithmetic, not evidence truth.
