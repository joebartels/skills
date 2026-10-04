# Neutral pair review: grpc-policy

A and B independently satisfy the original README contract in the inspected scope. No actionable correctness, resilience, or testing defect was substantiated. Substantive policy behavior is tied; the test suites have complementary strengths. All six topic grades are A, with zero severity counts.

Findings were frozen under neutral labels in `pair-grpc-policy-reproductions/raw-findings.json` before any arm disclosure. No arm mapping, candidate guidance, other trials, author reports, campaign summaries, or repository diff was inspected. The supplied candidates and frozen probes were not modified.

## Scope and evidence

This is a code-area review of each complete small client library against its own `source/README.md`, not a changeset attribution. References such as `A/source/client.go:27` are relative to `/private/tmp/go-client-campaign/reviews/pairs/grpc-policy`. The two README files and module files are byte-identical. Both declare Go 1.22.0, grpc-go v1.67.3, and protobuf v1.34.2. All twelve source-file hashes matched `source-hashes.json`.

Applied unchanged review skills: Correctness & Compatibility, Observability & Resilience, and Testing, including their relevant decision references. Inspected all implementation and test files, modules/checksums, raw controller results, frozen probes, and pinned grpc-go source for option application, resolver selection, native retry commitment/budget, and wrapped status classification.

The controller's `checks.json` records exit 0 for both labels on public tests, vet, race tests, Go 1.22.12 tests/vet, and private tests/race/minimum-toolchain tests. These are supplied results, not independently rerun commands.

Independent checks used disposable copies at `/private/tmp/grpc-policy-review-4a067mcc/{A,B}`. The frozen probe was copied unchanged to `frozen_review_test.go`. All commands used this environment:

```text
GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off
```

Each command below was prefixed with `rtk proxy env` and that environment. The host is the supplied Go 1.26.5; the minimum binary is `/private/tmp/go-client-campaign/tools/go/bin/go` (supplied Go 1.22.12).

| Independent command, in each scratch candidate | A | B |
| --- | --- | --- |
| `go test -ldflags=-linkmode=external -count=1 -timeout=20s ./...` (candidate tests plus frozen probes) | exit 0, 1.057s | exit 0, 0.974s |
| `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -count=1 -timeout=20s ./...` | exit 0, 1.049s | exit 0, 0.925s |
| `go test -race -ldflags=-linkmode=external -count=1 -timeout=20s ./...` | exit 0, 2.434s | exit 0, 2.285s |
| `go test -race -ldflags=-linkmode=external -count=1 -timeout=20s -v -run '^TestReview' ./...` (independent boundary checks) | exit 0, 1.551s | exit 0, 1.790s |
| `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -count=1 -timeout=20s -v -run '^TestReview' ./...` | exit 0, 0.240s | exit 0, 0.402s |

Exact commands and raw outcomes are saved in `pair-grpc-policy-reproductions/executed-checks.json`. Timings are test output, not comparative performance measurements. Independent boundary checks are saved in `pair-grpc-policy-reproductions/review_boundaries_test.go`; they verify borrowed policy/ownership, wrapped Unavailable, host `WithDisableRetry`, host empty default configuration, host interceptor forwarding, and absence of injected credentials. They passed for both candidates and did not reproduce a consequential concern.

## A: Correctness & Compatibility — A

Scope: Complete A library against `A/source/README.md`; Go 1.22.0 language/module contract and pinned grpc-go v1.67.3. Findings, if present, would be existing-in-scope.

Coverage: Signatures; exact fallback method selection and retry parameters; host dial options; resolver precedence including empty valid policy; native application-call ownership and response-header commitment; service request forwarding; health result values; earlier deadline, total deadline, cancellation; status wrapping; borrowed connection reuse and policy preservation. Both toolchains were exercised. No unsupported historical API or release contract was assumed.

Rationale: No actionable issue found. Implementation and realistic boundary tests establish the promised behavior. A is warranted by assessed material risks and verified strengths; the safeguards implement the stated contract and are not used to inflate the grade to A+.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-C-G1] `A/source/client.go:12` scopes the exact three-attempt, 10ms/20ms, multiplier-2, Unavailable-only policy to Health/Check. `A/source/client.go:27` prepends only `WithDefaultServiceConfig` and forwards host options. `A/source/policy_test.go:70` verifies every parameter and excludes Watch/other methods; `:155` executes resolver two-attempt and empty-policy precedence. Frozen probes independently pass the required three/two/one execution counts.
- [A-C-G2] `A/source/client.go:36` derives one timeout and performs one generated Check on the supplied connection. `A/source/policy_test.go:101` distinguishes one application invocation from native server attempts and verifies commitment/permanent rejection. `:206`, `:247`, `:319`, and `:362` exercise health values, deadlines, cancellation, and pre-cancellation. `%w` at `A/source/client.go:42` preserves status classification; no close or reconfiguration exists in Probe. Independent borrowed-connection checks passed.

Bad

- None found.

Suggested changes

- None needed.

Limits: Independent commands and outcomes are listed above. Local bufconn exercises gRPC execution, not production TLS handshakes, DNS availability, or network loss. Credentials are forwarded and required; no broader transport guarantee is inferred. Supported non-macOS targets were not executed.

## A: Observability & Resilience — A

Scope: A's Dial/Probe failure boundary, native retry configuration, caller context, diagnostics, and connection lifetime.

Coverage: Retry eligibility, total attempts/backoff, resolver authority, host retry disablement, absence of an application/interceptor retry layer, header commitment, total deadline/earlier caller deadline, cancellation, useful classified errors, and borrowed resource preservation. This library has no service lifecycle or owned telemetry exporter to assess.

Rationale: No actionable issue found. Failed or slow health calls stay under the caller-derived total budget, retry ownership remains native, and callers can classify wrapped RPC failures. Returning errors at this library boundary is adequate; additional logging, metrics, or circuit breakers are not implied by the contract. These verified safeguards support A.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-R-G1] `A/source/client.go:12` and `:27` let grpc-go control attempts and backoff without multiplying retries. Native grpc-go `stream.go:633` rejects committed/non-trailers-only failures, applies eligibility/attempt limits, uses jittered native backoff, and cancels waits through the RPC context. `A/source/policy_test.go:101` verifies Unavailable recovery/exhaustion, permanent rejection, sent-header commitment, and host retry disablement; frozen probes agree.
- [A-R-G2] `A/source/client.go:37` creates one total budget and defers cancellation. `A/source/policy_test.go:247` verifies the unchanged retry deadline and earlier caller deadline, `:319` verifies in-flight cancellation, and reuse assertions confirm context errors do not destroy the caller's connection. `A/source/client.go:42` returns the underlying status through wrapping.

Bad

- None found.

Suggested changes

- None needed.

Limits: No application SLO, production instrumentation, or deployment environment was supplied; these are outside the scoped library contract. Backoff values were inspected through parsed configuration rather than judged by noisy wall-clock estimates. Native retry source and bufconn tests assess the pinned grpc version; no claim is made about future dependency behavior.

## A: Testing — A

Scope: Complete A test suite and its detection of the original README behavior; controller checks and independently executed candidate/frozen/boundary checks.

Coverage: Real generated service/bufconn calls; parsed fallback policy and method exclusion; native retry recovery/exhaustion; one application invocation; commitment and permanent errors; resolver policies; signature compatibility; host options/credentials; all named health values; total/parent budgets; cancellation; cleanup and reuse. Race and minimum-toolchain checks passed.

Rationale: No actionable testing issue found. Tests assert both policy configuration and consequential runtime execution, so a service-config-shaped string or application retry loop alone would not satisfy them. Core assertions are independent of implementation mechanics. The passing checks support these strengths without implying exhaustive proof.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-T-G1] `A/source/policy_test.go:101` counts application invocations and server executions separately, sends actual headers for commitment, and covers host retry disablement. `:155` publishes parsed configuration through a manual resolver and checks actual attempt counts plus selected policy. These tests detect consequential policy regressions.
- [A-T-G2] `A/source/policy_test.go:247` inspects the application deadline and retry deadlines, while `:319` waits for observed call startup before cancellation. `:362` verifies a canceled caller produces zero server calls. Reuse at `:234` is exercised after health, deadline, and cancellation outcomes. The helper at `:36` owns connections, waits for server completion, and uses handler waiting; no race was reported on exercised paths.

Bad

- None found.

Suggested changes

- None needed.

Limits: A's own suite checks reuse on Dial-created connections rather than directly constructing an unconfigured borrowed connection. The direct pointer use, one-call assertions, and independent borrowed-connection check establish current behavior; porting B's explicit borrowed-connection case would broaden future detection, but is an optional refinement, not a counted defect. No repeated/shuffled or TLS tests were needed for a demonstrated concern.

## B: Correctness & Compatibility — A

Scope: Complete B library against `B/source/README.md`; Go 1.22.0 language/module contract and pinned grpc-go v1.67.3. Findings, if present, would be existing-in-scope.

Coverage: Signatures; exact fallback method/parameters; forwarding of host options; resolver precedence; one generated Check/native retry ownership; sent-header commitment; service forwarding; health values; total and parent deadlines; cancellation; status wrapping; borrowed connection policy and ownership. Both toolchains were exercised.

Rationale: No actionable issue found. B independently implements the README requirements and preserves observable caller contracts. Verified ordinary contract safeguards support A; the shorter implementation is not itself a quality gain.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-C-G1] `B/source/client.go:12` contains the exact required native method policy; `:26` prepends a fallback and preserves host options. `B/source/policy_test.go:75` checks parameters/method exclusion, and `:160` executes both two-attempt and empty valid resolver policies. The decimals `0.01s`/`0.02s` represent the same durations as A's `0.010s`/`0.020s`.
- [B-C-G2] `B/source/client.go:32` narrows the caller context around one generated Check and wraps failures with `%w` at `:37`. `B/source/policy_test.go:96` checks native commitment and application/server counts; `:207` supplies a plain `grpc.Dial` connection, confirms one failed execution, and successfully reuses it. Health values (`:229`), retry budgets (`:248`), and caller cancellation/deadlines (`:290`) are exercised. Independent boundary tests also confirm no reconfiguration or closing.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same local bufconn/platform limits as A. No historical public signatures beyond the supplied contract, production TLS behavior, or unsupported target matrix was assumed.

## B: Observability & Resilience — A

Scope: B's Dial/Probe failure boundary, native retry policy, contexts, diagnostics, and borrowed connection lifetime.

Coverage: Native retry ownership, eligibility/backoff/attempt bounds, resolver authority, commitment, permanent statuses, total budget, parent deadline/cancellation, classified error wrapping, host options, and resource ownership. No service or telemetry lifecycle exists in scope.

Rationale: No actionable issue found. gRPC owns the retry state, while one caller-derived timeout controls all attempts. Errors remain classifiable and include the requested service. Native retry and context safeguards are verified and support A.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-R-G1] `B/source/client.go:26` adds only the native fallback option. `B/source/policy_test.go:96` verifies three native executions versus one application Check, one permanent execution, and one execution after headers. `:160` preserves resolver authority. Pinned grpc-go source provides native jitter, pushback handling, and cancelable waits; the independent host-disable-retry check passed.
- [B-R-G2] `B/source/client.go:33` places the 500ms bound around the single RPC and releases its timer. `B/source/policy_test.go:248` observes the original budget across native attempts; `:290` observes parent cancellation at the server as well as the returned status. `B/source/client.go:37` includes service identity while retaining the wrapped status, and `B/source/policy_test.go:207` verifies failure does not consume borrowed ownership.

Bad

- None found.

Suggested changes

- None needed.

Limits: No production SLO, instrumentation, overload model, or deployment lifecycle was supplied or required. Local tests and pinned source establish this retry boundary, not guarantees about future grpc versions or production network fault distributions.

## B: Testing — A

Scope: Complete B test suite against the original README, supplied controller outcomes, and independent frozen/boundary checks.

Coverage: Real gRPC calls; exact fallback configuration; method exclusions; native attempts/application invocation distinction; recovery/exhaustion/permanent failure; sent-header commitment; valid resolver precedence; service payload; signature checks; borrowed unconfigured connection; health values; total/parent deadlines; cancellation and cleanup. Race and minimum-toolchain checks passed.

Rationale: No actionable testing issue found. The suite verifies the important native-policy behavior through actual execution and independently verifies the non-Dial borrowed-connection boundary. It also checks wrapped causes, not merely formatted strings. These meaningful assertions support A without using test count as evidence of superiority.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-T-G1] `B/source/policy_test.go:96` checks server attempts, application intercepts, requested service, status code, and `errors.Unwrap`. `:160` changes real resolver state and observes the resulting attempt count, avoiding a fake retry implementation. `:207` directly exercises a borrowed connection with no fallback policy and its subsequent reuse.
- [B-T-G2] `B/source/policy_test.go:248` collects per-attempt server deadlines; the fixture has a finite failure path if Probe omits the timeout. `:290` synchronizes on startup and server cancellation, then checks the classified return. The helper at `:37` waits for a Ready state under a bounded context and owns server/connection cleanup. Exercised concurrency is race-clean.

Bad

- None found.

Suggested changes

- None needed.

Limits: B's own suite lacks A's explicit missing-credentials, host-disable-retry, and pre-canceled-context cases. Credentials/interceptors/resolver options are otherwise exercised, current forwarding/context code is direct, and independent boundary checks cover the first two. Porting those A cases is an optional detection improvement, not a demonstrated important unverified contract. No claim of exhaustive race or transport-fault coverage is made.

## Substantive comparison

| Dimension | A | B | Assessment |
| --- | --- | --- | --- |
| Native retry ownership and commitment | Native fallback; one generated Check; sent-header test | Same | Tie. Both preserve grpc commitment and avoid retry multiplication. |
| Resolver precedence | Two attempts and empty policy; parsed-policy and execution assertions | Two attempts and empty policy; execution assertions | Tie on consequential behavior; A additionally inspects the selected resolver policy. Frozen probes pass both. |
| Deadline/cancellation/status | Total application/retry deadline, parent equality, cancellation, pre-cancellation, `%w` | Total server deadlines, parent tolerance, explicit server cancellation, `%w`/unwrap assertions | Both cover the important contract with different complementary assertions. |
| Borrowed connection preservation | Reuse after health and context outcomes; simple direct implementation | Explicit non-Dial connection failure and reuse | B has more direct supplied-suite evidence for preserving an unconfigured connection. Independent checks pass both. |
| Host options | Explicit no-credentials and retry-disable cases; forwards options | Credentials/interceptor/resolver forwarding; direct prepend implementation | A has broader supplied-suite evidence for these options. Independent checks pass both. |
| Diagnostics and documentation | Documents resolver precedence/ownership; health errors omit requested service | Errors name requested service; shorter comments | B's errors give slightly more local diagnostic context; A's comments more explicitly state ownership. No consequential defect follows from either choice. |
| Unnecessary surface/dependencies | Constant plus the required Dial/Probe; no application retry machinery | Same | Tie. Exact module/checksum equality; no additional exported APIs or dependencies. Slice construction and decimal formatting differences are inconsequential. |

Neither candidate demonstrates a substantive correctness or policy improvement over the other. Both meet the important contract, and each suite supplies useful checks absent from the other. The diagnostic wording is a small local advantage for B; the extra comments and test cases do not justify declaring A better overall. No required change or grade-based winner is supported by this evidence.
