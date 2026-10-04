# Independent review: grpc-transfer, A and B

Raw findings are frozen before any arm disclosure. A/B are neutral packet labels; no arm map, candidate guidance, other trial, author report, campaign summary, or Git diff was consulted.

The review treats each supplied source directory as a complete library code area, independently against its original README. Paths below are relative to `/private/tmp/go-client-campaign/reviews/pairs/grpc-transfer/`. Both READMEs, module files and existing smoke tests have identical recorded hashes. Every source hash was independently verified with no mismatches. Both declare Go 1.22.0, grpc v1.67.3 and protobuf v1.34.2. The unchanged correctness, observability/resilience and testing skills and their relevant decision references were applied.

## A: Correctness & Compatibility — A

Scope: `A/source/client.go`, complete supplied tests/module and README; borrowed-connection unary client library, Go 1.22.0 and pinned grpc v1.67.3.

Coverage: Generated method and request/response bodies; operation identity, unrelated and appended metadata, caller metadata ownership; native retry policies and application invocation count; header commitment and caller retry after an effect; total and earlier deadlines, in-flight cancellation, status codes and borrowed connection reuse. Inspected the pinned generated stub and metadata/retry implementation. No baseline revision was supplied, so this is a code-area review, not a claim about introduced changes.

Rationale: No substantiated contract or compatibility defect. The generated service boundary and failure paths are exercised, not bypassed by a fake client. The implementation retains the exact API, module requirements and supported Go version. These are verified sound choices; no additional safeguards beyond the assessed contract are claimed for A+.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-C1] `A/source/client.go:19` rejects empty identity before dereferencing the connection, demonstrated by the nil-connection test at `A/source/commit_behavior_test.go:119`.
- [A-C2] `A/source/client.go:24` reads a copied, flattened outgoing metadata map, sets exactly the caller identity, and passes exact payload bytes through `SimpleRequest.Payload.Body` at line 30. `metadata.FromOutgoingContext` in pinned grpc source copies slices and merges appended values (`metadata/metadata.go:260`). Tests at `A/source/commit_behavior_test.go:91` verify stale identity replacement, appended metadata, unchanged caller metadata, binary payload and distinct response body. This avoids mutating shared caller context metadata.
- [A-C3] `A/source/client.go:30` is one generated `UnaryCall`, whose pinned generated stub invokes `/grpc.testing.TestService/UnaryCall` once (`interop/grpc_testing/test_grpc.pb.go:102`). Tests at `A/source/commit_behavior_test.go:126` count both the application interceptor and server attempts under configured, absent, disabled and nonretryable policy. No loop, reconfiguration or close is present.
- [A-C4] Tests at `A/source/commit_behavior_test.go:178` verify a header-committed lost acknowledgement surfaces as Unavailable after one application invocation, then a caller repeat retains identity/payload and returns the effect result. The frozen probes independently verify three native attempts cause one deduplicated effect, and two outer calls after header commitment cause one effect.
- [A-C5] `A/source/client.go:22` derives one 500ms context and returns transport errors unchanged at line 32; decoded response payload is returned at line 34. The deadline/cancellation tests at `A/source/commit_behavior_test.go:215` and line 260 verify both server and caller behavior. Status codes are checked across transient, permanent, committed, canceled and deadline outcomes. There is no wrapping or substitution that would discard the borrowed interceptor/transport error.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied tests plus the frozen probes passed in a disposable copy under Go 1.22.12; host race testing also passed. Transport faults are modeled by actual generated-service handlers returning failure after an effect; no external server or TCP fault proxy was used. The server's deduplication guarantee is the stated contract, not a production server implementation being audited. Concurrent mutation of a caller-owned payload during an in-flight RPC is not a promised usage pattern. No wider callers or release history were supplied.

## A: Observability & Resilience — A

Scope: A's outgoing RPC budget, retry ownership, commitment, ambiguous acknowledgement and borrowed connection lifecycle.

Coverage: Safe native replay and stable identity, retry eligibility and disablement, response-header commitment, cancellation during execution, total deadline across retry delay/attempts, status diagnostics and resource ownership. Service deployment, metrics, probes and shutdown are outside this small client library's supplied scope.

Rationale: No actionable resilience defect. Native grpc owns retry eligibility, attempts, backoff, jitter and commitment; the client supplies stable identity and a total context budget. Returning the original status error is appropriate diagnostic behavior at this library boundary. No independent application telemetry or circuit breaker is required by the contract.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-R1] `A/source/client.go:22` bounds the whole generated invocation rather than resetting time per attempt. `A/source/commit_behavior_test.go:215` uses 150ms server pushback then a blocking second attempt, checks both server deadlines remain near the original 500ms end, and waits for handler cancellation.
- [A-R2] [A-C3/A-C4] A leaves effective retry policy untouched. The pinned grpc `stream.go:633` checks commitment and disabled retry, requires trailers-only status failures, uses the effective method retry policy, enforces attempt limits, and interrupts retry waits with `cs.ctx.Done()` at line 724. Header commitment and stable replay are confirmed by the supplied tests and frozen probes.
- [A-R3] `A/source/client.go:14` documents that a failed acknowledgement may follow an effect. The failed return at line 32 preserves status diagnostics and makes no false no-effect claim. Same-connection successive calls in the metadata and committed-failure tests verify borrowed ownership in practice.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same executed checks and effect-model limits as A correctness. Native retries may commit early because of grpc replay-buffer limits or other effective connection settings; delegating that decision to grpc is required behavior, not a client defect. Telemetry supplied by the caller's connection/interceptors was not separately audited.

## A: Testing — A

Scope: `A/source/client_test.go`, `A/source/commit_behavior_test.go`, controller `A/checks.json` and the packet frozen probes.

Coverage: Assertions, actual generated service and serialization, application/native attempt distinction, replay request stability, commitment, retry eligibility, metadata ownership, connection reuse, deadline/cancellation synchronization, resource cleanup and minimum-version compilation. Controller checks are supplied evidence; fresh independent checks are separately listed below.

Rationale: No substantiated test gap that defeats an important contract. The suite exercises meaningful negative paths and would reject a client loop, changed identity/body, ignored retry policy, or lost context budget. Passing checks are supported by inspected assertions rather than test count or coverage percentage.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-T1] `A/source/commit_behavior_test.go:35` uses registered generated grpc service over bufconn, bounded blocking dial, context-aware dialer, connection cleanup and an observed server shutdown. This exercises the real grpc retry/commitment boundary.
- [A-T2] `A/source/commit_behavior_test.go:143` counts application invocations independently from server attempts. Line 178 exercises headers before failure and explicit outer caller retry; each observed request is checked for exact identity and payload. The frozen probe adds a real identity-to-payload deduplication map and checks unrelated metadata during replay.
- [A-T3] `A/source/commit_behavior_test.go:215` checks deadline continuity across two actual native attempts. Line 292 uses start/finish channels for in-flight cancellation and bounds its waits and goroutine cleanup. Server cancellation is observed, rather than inferred from the client return alone.
- [A-T4] The compile-time function assignment at `A/source/commit_behavior_test.go:22` checks the API shape; fresh Go 1.22.12 tests/vet and host race tests pass with the frozen probes included.

Bad

- None found.

Suggested changes

- None needed.

Limits: Some request-observation receives use the suite's timeout as their ultimate bound. They still make a missing request fail verification; this is not evidence of an undetected contract regression. The original controller results report success for eight public/private test/vet commands but do not expose private test names. No coverage percentage, mutation score, CI enforcement or external infrastructure was inferred.

## B: Correctness & Compatibility — A

Scope: `B/source/client.go`, complete supplied tests/module and README; borrowed-connection unary client library, Go 1.22.0 and pinned grpc v1.67.3.

Coverage: Generated method and bodies, identity and metadata, native policy and one application invocation, header commitment and repeated caller identity/payload, deadline/cancellation, status preservation and borrowed ownership. Inspected the same pinned generated stub and metadata/retry implementation. This is an independent code-area assessment against B's original README.

Rationale: No substantiated contract or compatibility defect. B's executable production behavior matches the README in the ordinary and failure paths assessed. No A+ claim is made for ordinary correct contract implementation.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-C1] `B/source/client.go:19` rejects empty identity before the generated invocation; `B/source/commit_test.go:140` verifies zero interceptor invocations and zero server attempts.
- [B-C2] `B/source/client.go:24` uses the pinned grpc metadata copy/merge operation, replaces old base and appended identities, and sends exact request bytes at line 30. `B/source/commit_test.go:101` checks binary input, different returned data, multivalue tenant metadata, appended trace metadata and unchanged original metadata.
- [B-C3] `B/source/client.go:30` makes exactly one generated unary invocation and leaves the borrowed connection untouched. `B/source/commit_test.go:155` distinguishes absent, disabled, successful, exhausted and permanent-status native policies with application and server counters and checks identity, payload and unrelated metadata on every observed attempt.
- [B-C4] `B/source/commit_test.go:207` validates the identity and exact accepted payload, records one effect, sends headers before Unavailable, then permits an explicit caller repeat to retrieve that effect's result. The frozen probes independently confirm native and outer replay identity, payload and one effect.
- [B-C5] `B/source/client.go:22` supplies the total context; error at line 32 is returned unchanged and result comes from the generated response body at line 34. Tests at `B/source/commit_test.go:260`, line 304 and line 326 verify operation/caller deadlines, native waiting cancellation and in-flight cancellation. A subsequent caller Commit after committed failure succeeds on the same borrowed connection.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied tests plus frozen probes passed in a disposable copy under Go 1.22.12; host race testing passed. Same realistic-handler fault-model, server-guarantee and unsupported concurrent-payload-mutation limits as A. No additional wrapper is introduced, so status code/error preservation follows the direct return and exercised transport statuses; no separately injected wrapper-error test was needed to establish that path.

## B: Observability & Resilience — A

Scope: B's outgoing RPC budget, native retry ownership, commitment, ambiguous acknowledgement and borrowed connection lifecycle.

Coverage: Native retry eligibility/disablement/exhaustion, safe replay identity, committed failure, cancellation, total budget during native retry waiting, status diagnostics and ownership. Service telemetry/deployment/lifecycle are not part of this supplied library area.

Rationale: No actionable resilience defect. B preserves effective grpc policy and commitment, bounds the invocation under the caller context, returns informative transport status and documents ambiguity. No extra resilience machinery is justified by the supplied contract.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-R1] `B/source/client.go:22` creates one total context; `B/source/commit_test.go:304` supplies 1000ms pushback, verifies return near 500ms, DeadlineExceeded, one application invocation and only one server attempt. The earlier deadline and cancellation tests verify propagation to the handler.
- [B-R2] [B-C3/B-C4] Native retry configuration is honored without an application loop or call-option override. The pinned grpc retry source implements bounded attempts, jitter, context-interruptible waits and non-replay after response headers. B's exhausted-policy test and frozen probes confirm the behavior through actual handlers.
- [B-R3] `B/source/client.go:14` explicitly assigns retry policy and connection ownership to the caller and explains that failure may follow a completed effect. Returning the original error at line 32 preserves failure diagnostics without asserting no effect.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same executed checks and fault-model limits as B correctness; no audit of caller telemetry, production server retention or infrastructure. The operation identity is supplied by the caller, as required, so its uniqueness and use across distinct operations remain caller responsibilities.

## B: Testing — A

Scope: `B/source/client_test.go`, `B/source/commit_test.go`, controller `B/checks.json` and frozen probes.

Coverage: Generated-service boundary and serialization, assertion quality, one invocation versus native attempts, retry disablement/exhaustion, stable request metadata/payload, commitment and caller repeat, context budgets/cancellation, cleanup, API/minimum Go compatibility and exercised races.

Rationale: No substantiated consequential testing defect. The suite directly verifies the critical gRPC semantics and observes cancellation on both client and server. Additional cases are useful because of their assertions, not merely because B has more lines or cases.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-T1] `B/source/commit_test.go:39` serves the generated service over bufconn with bounded context-aware dial and cleanup. `grpc.WaitForHandlers(true)` plus bounded stop observation ensures owned handlers finish during fixture teardown.
- [B-T2] `B/source/commit_test.go:155` checks retry exhaustion at three attempts, native success at the configured final attempt, absent/disabled policy and permanent failure, while an interceptor counts exactly one application invocation. The committed-effect model at line 207 validates repeated request content before returning the original effect result.
- [B-T3] `B/source/commit_test.go:304` distinguishes waiting for native retry from running another attempt, verifying deadline cancellation during a long pushback. The cancellation test at line 326 synchronizes on server start, records server finish and bounds goroutine cleanup.
- [B-T4] The API assignment at `B/source/commit_test.go:22`, fresh Go 1.22.12 test/vet, host race checks and independent frozen probes all pass. Metadata assertions include caller-owned original values and unrelated values across replay.

Bad

- None found.

Suggested changes

- None needed.

Limits: No CI enforcement or quantitative coverage claim is made. Controller private checks are reported successes without disclosed individual test names. Tests exercise bufconn grpc transport, not production networking; the scoped client contract does not require the latter.

## Verification evidence

All commands ran in disposable scratch copies at `/private/tmp/grpc-transfer-independent-xs70nkrf/{A,B}`. Supplied candidates and probes were not edited. The following environment was used for every fresh Go check:

```text
GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath
GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off
```

Verification shell commands were prefixed with `rtk`; the Go commands below were invoked using `rtk env` followed by those variables.

| Fresh command, in each label's copy with frozen probe installed | A | B |
| --- | --- | --- |
| `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -count=1 -timeout=20s -v ./...` | exit 0, 1.039s | exit 0, 1.448s |
| `/private/tmp/go-client-campaign/tools/go/bin/go vet ./...` | exit 0, empty output | exit 0, empty output |
| `go test -ldflags=-linkmode=external -race -count=1 -timeout=20s ./...` | exit 0, 2.095s | exit 0, 2.509s |

The host toolchain is the supplied Go 1.26.5; the minimum toolchain is Go 1.22.12. Each packet's `checks.json` additionally reports exit 0 for all eight original public/private test, vet and race/minimum-version commands. These raw controller outcomes corroborate the fresh checks; they do not alone establish assertion coverage. Separate raw frozen-probe runs and the unchanged probe/hash inputs are retained in `pair-grpc-transfer-reproductions/`. No defect required an additional custom reproduction.

## Substantive comparison

**Production correctness, policy and dependency/surface tie.** Both production bodies are identical except for the empty-identity error text; that wording is not constrained by the README. Comments differ slightly, with B more explicit about connection/retry ownership, but both correctly warn about ambiguous effects. Both use the same generated unary call, timeout, copied metadata update, exact payload, direct error return and decoded result. Neither adds public API, production helper machinery, retry loops, transport reconfiguration, connection ownership, code generation or dependencies. Module and sum hashes match exactly.

**Tests have complementary substantive strengths, without an overall winner.** A explicitly checks that two server attempts see the same original total deadline after pushback (`A/source/commit_behavior_test.go:215`), and rejects empty identity with a nil connection. B explicitly checks native-policy exhaustion at the configured third attempt (`B/source/commit_test.go:155`), long retry waiting canceled before the next attempt (line 304), appended stale identity replacement and multivalue metadata (line 101), and handler-aware fixture shutdown (line 39). A covers deadline continuity during a subsequent attempt; B covers the long wait before that attempt. Both test application/native invocation distinction, committed lost acknowledgement, outer caller repeat, exact binary payload, status codes, cancellation and borrowed connection reuse. The frozen probes close the same native and outer deduplication loop for both. Neither extra comments nor additional test cases establish a production improvement.

There are no substantiated Bad findings or grade differences; all applicable topics tie at A. No ungraded consequential issue was found in the requested unnecessary-surface/dependency comparison.
