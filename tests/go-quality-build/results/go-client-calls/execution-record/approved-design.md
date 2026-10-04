# Design: Go client calls

Review artifact for a new `go-client-calls` authoring skill. Kept outside the repository because maintained documentation excludes future proposals. This design specifies the skill and its evaluation; it makes no implementation or effectiveness claim.

## Purpose and selection

Help an agent build or modify Go HTTP and unary gRPC clients whose configuration and failure behavior fit the remote operation's contract. Success means preserved consumer behavior, safe repetition, bounded failure handling and appropriate client reuse, without unnecessary resilience machinery.

Proposed trigger: use when Go work creates or changes an outbound HTTP/gRPC client, remote-call configuration, replay/idempotency policy, retries, circuit breaking or remote-failure interpretation. Skip unrelated calculations, server-only handlers and mechanical edits. A simple remote call can need client guidance while needing no retry or breaker.

Use one shared main skill with conditional HTTP and unary gRPC references. Separate transport skills would make protocol selection straightforward but duplicate replay and failure-policy rules. The shared approach keeps those decisions under one owner; references supply transport-specific behavior. Streaming recovery, hedging and application-wide resilience orchestration are outside the first version.

## Decision ownership

| Owner | Responsibility |
| --- | --- |
| Client calls | Remote-operation classification, replay safety, client-specific configuration, retry-layer selection, attempt accounting, breaker policy and protocol failure interpretation. |
| API contracts | Supported signatures, caller-visible outcomes and permitted compatibility changes. |
| Interfaces and composition | Dependency construction, abstraction choices, borrowed versus acquired resource lifetime. |
| Context and deadlines | Propagation, total/stage scope mechanics, cancellation/result precedence and cancel ownership. |
| Concurrency and ownership | Shared-state synchronization, admission bounds and stop/join/release. |
| Behavior tests and test isolation | Independent assertions, faithful fixtures, event control and cleanup. |

General error representation, safe exposure and telemetry remain with their existing guidance and contracts. Client calls adds HTTP/gRPC classification and ambiguous-outcome decisions without prescribing a new public error hierarchy. References explain relevant client mechanics without changing neighboring owners.

## Main procedure

1. Inspect the actual caller, remote API, supplied client, effective configuration, supported Go/library versions and existing middleware. Establish what one logical operation means, whether it changes state, and what a failure leaves known or uncertain.
2. Separate operation policy from request attempts. Identify which layer may repeat the operation, which layers already retry, and what the attempt limit counts. Prefer an existing suitable client capability over adding a wrapper. Account explicitly for unavoidable transparent transport behavior instead of promising an unsupported wire-attempt maximum.
3. Establish replay safety before classifying transient failures. Inspect naturally repeatable semantics or a documented server deduplication contract. Keep identity and semantically identical payload stable for that operation, including across outer workflow retries when required. A newly generated key per attempt, or a header unsupported by the server, cannot establish deduplication. Treat a lost response after a mutation as uncertain completion.
4. Choose retry eligibility from the operation and specific failure. Separate permanent rejection, authentication/configuration problems, caller stopping and plausible dependency transients. A status category alone cannot prove repeat safety; a transaction-level retry may require restarting the enclosing operation. Preserve independent errors and promised classifications.
5. Bound attempts and waiting under one total operation budget. Stage/attempt budgets narrow what remains. Make backoff cancellable, apply jitter where synchronized retries would add load, and handle valid server retry guidance within the remaining budget. Define exhausted outcomes without inventing success or erasing uncertainty. Avoid fixed universal timeout/attempt values.
6. Decide whether a breaker serves a demonstrated persistent-failure/resource problem. Define its dependency scope, health-eligible failures, whether counts describe calls or attempts, recovery admission and observable rejection. Distinguish caller cancellation and ordinary business rejection from dependency health. Use existing project mechanisms where suitable; timeouts, retry limits and capacity control may suffice.
7. Verify through the relevant client boundary. Observe actual requests/RPC executions, payload/metadata identity, retained outcomes, elapsed budget, cancellation and resource closure. Report verified behavior and material limits. A generic helper test cannot establish an effective transport retry policy.

Client configuration guidance will cover reusable clients/transports/connections, instance-specific endpoints and credentials, secure transport, workload-appropriate response/message/connection limits, and preserving borrowed configuration. Avoid process-global mutation, per-call connection creation, insecure production recipes, arbitrary pool tuning and mandatory instrumentation/frameworks. Where instrumentation exists, distinguish logical-call outcomes from attempts.

## HTTP reference

Explain reusable `http.Client`/`Transport` configuration and copying configuration safely; request context and body lifetime; semantic status handling; redirect effects on targets/methods/credentials; replayable bodies; response ownership on each outcome; and bounded response/error consumption. Draining for reuse is conditional and bounded rather than an unlimited cleanup requirement.

Account for `net/http`'s existing network-error retries. Its transport's replay classification is a particular implementation contract, distinct from an application's semantic idempotency decision. Rebuild each permitted attempt from stable operation data and preserve its required idempotency identity. Define supported `Retry-After` handling and final errors rather than retrying every 5xx/429 by category.

Primary sources: [Go HTTP client](https://pkg.go.dev/net/http#Client), [transport](https://pkg.go.dev/net/http#Transport), [request replay](https://pkg.go.dev/net/http#Request.GetBody), and [HTTP semantics](https://www.rfc-editor.org/rfc/rfc9110.html).

## Unary gRPC reference

Explain reusable `ClientConn` and generated clients, credentials/metadata, per-method policy, resolver/default service-config precedence, deadline-bound wait-for-ready, and status interpretation at the requested operation boundary. Read the repository's selected grpc-go version before using an API or interpreting a dial option.

Inspect native retry/service configuration before adding retry interceptors or application loops. Cover configured versus transparent retries, retry throttling/pushback and call commitment. `ABORTED` can require a higher-level operation restart; `UNAVAILABLE` still requires repeat safety. Do not translate a deadline result into proof that the server performed no mutation.

One source-sensitive safeguard: current Go library documentation says `WithDisableRetry` leaves transparent retries active. Its `WithDefaultServiceConfig` is a fallback when the resolver supplies no valid config, unless resolver service configs are disabled. The reference must use the selected Go implementation's documented behavior rather than assuming all language implementations match a general guide.

Primary sources: [grpc-go dial options](https://pkg.go.dev/google.golang.org/grpc#WithDisableRetry), [default service configuration](https://pkg.go.dev/google.golang.org/grpc#WithDefaultServiceConfig), [gRPC retries](https://grpc.io/docs/guides/retry/), [status codes](https://grpc.io/docs/guides/status-codes/) and [wait-for-ready](https://grpc.io/docs/guides/wait-for-ready/).

## Evaluation and promotion

Prepare four realistic paired tasks and one non-selection control before drafting the guidance. Hold task inputs, neighboring-skill catalog, model/effort settings and allowed information equivalent across baseline and skill-on arms. Freeze source/configuration hashes and exact prompts; disclose settings the harness cannot expose. Authors receive requirements and applicable skill descriptions, while expected assertions, private probes and previous outcomes stay withheld. Use independent reviewers with neutral arm labels.

| Case | Observation and plausible regression |
| --- | --- |
| HTTP mutation with lost acknowledgement | Exactly one promised effect under the server's deduplication contract; stable identity/payload; unsafe repeat prevented when deduplication is unavailable. Reject a key regenerated per attempt or truncated replay body. |
| HTTP overload and attempt budget | Contract-specific retryable/permanent responses, valid server delay, cancellation while waiting and one total deadline; successful and terminal bodies released. Reject blanket retry, budget restart or borrowed-client mutation. |
| Unary gRPC policy | Effective per-method native retry policy without multiplying it in an interceptor, meaningful status distinctions and caller-observable outcomes. Reject nested retry amplification or assuming disabling policy also removes transparent retries. |
| Breaker isolation and recovery | Failures affect the declared dependency scope, eligible failures drive opening, recovery admits the promised probes, and simple calls retain minimal configuration. Reject per-request breaker state, cross-dependency failure sharing or caller/business failures incorrectly changing health. |
| Non-selection control | A local deterministic edit remains local and skips this skill. Reject added client/resilience dependencies or abstractions. |

After freezing a candidate, run a fresh paired transfer task with a different client shape. Assess benefit, regressions and added complexity; include actual minimal-client preservation within the applicable cases. Starting evaluation budget is six paired tasks (12 author contexts) plus independent assessment. Allow one evidence-driven guidance revision and rerun its affected cases; further expansion requires a stated reason and agreement. Preserve strong baselines and mixed results instead of searching indefinitely for a higher grade.

Use real local HTTP and unary gRPC client/server execution where transport behavior is asserted; controlled doubles support narrower failures. Pin fixture grpc-go/protobuf dependencies and compatible module/toolchain versions before dispatch. Exercise actual supported toolchains; a module declaration is not minimum-version execution. Avoid a new universal minimum or protoc installation requirement by using a small existing generated service where suitable.

Run meaningful tests/vet and race checks for concurrent state. Use compiling, assertion-failing mutations for consequential claims, then restore source. Independent correctness, resilience, testing and relevant architecture reviews assess completed candidates. Packaging checks are separate from behavioral evidence.

Author a useful candidate outside runtime while measuring its benefit. Promote only if guidance demonstrates distinct value and preserves controls without unjustified machinery; no grade target substitutes for that evidence. Keep trial artifacts under `tests/go-quality-build/`, with no outcome/status narrative in maintained guides.

## Package shape and review criteria

On promotion, add `skills/go-client-calls/SKILL.md` with `references/http-clients.md` and `references/grpc-clients.md` to `plugins/go-quality-build/`. Keep the main concise and self-contained; references should resolve concrete protocol decisions with small verified examples. Package manifests advance together only when runtime content changes. Add a short decision-owner row to the package guide and validate canonical paths, links and evaluation suites.

Retain existing error-contract evaluation changes in the current workspace. Implementation planning must choose isolation appropriate to those concurrent edits. Do not modify completed evaluation archives, review grading rules or neighboring skills to improve this skill's apparent results.

Design review should confirm the shared-skill boundary, HTTP plus unary gRPC scope, conditional breaker guidance and the bounded evaluation campaign. Implementation planning follows approval of this concrete design.
