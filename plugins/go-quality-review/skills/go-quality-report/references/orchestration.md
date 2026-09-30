# Routing and work packets

## Discover topic skills

Resolve these names through the host's installed skill catalog, or the known local collection (commonly sibling directories). Pass resolved locations to workers; do not assume a machine-specific installation path. Verify availability before dispatch. Only applicable missing skills block coverage; do not install anything as part of a review. The orchestrator can use the table without loading nine full skills; the worker reads its assigned skills.

| Skill | Relevant decisions |
| --- | --- |
| `go-architecture-and-design` | Package/API boundaries, dependencies, composition, ownership, cross-component contracts |
| `go-code-quality-and-idioms` | Go clarity, errors, value semantics, idioms, meaningful formatting/vet/lint diagnostics |
| `go-correctness-and-compatibility` | Promised results, failure paths, shared state, concurrency, caller/CLI/wire/stored-data compatibility |
| `go-testing` | Verification of changed contracts, assertions, doubles, isolation, timing, fuzzing and benchmark validity |
| `go-security` | Trust and authorization, input/sinks, credentials, crypto, abuse limits, reachable vulnerabilities |
| `go-observability-and-resilience` | Diagnosability, metrics/traces, cancellation budgets, retries, overload, queue and lifecycle failure behavior |
| `go-performance-and-resource-management` | Workload cost, complexity, memory retention, resource ownership, pooling, contention and measurement |
| `go-dependencies-and-reproducibility` | Module/workspace/vendor selection, toolchain/targets, generated/native inputs, clean-checkout builds |
| `go-deployment-and-operations` | Release artifacts/gates, runtime configuration, images, probes, shutdown, rollout and recovery |

Treat this as a routing aid, not a keyword filter. A library can have security/resource risks; a CLI can have release operations; a config-only change can alter correctness or security. Determine applicability for every topic at least once across the requested scope, then route to the relevant chunks. A full-codebase review examines existing decisions even when no dependency or CI file changed.

## Chunk by contracts, not an arbitrary slice of lines

- Small cohesive change: one worker can use several closely related skills and return separate topic cards. Avoid reloading the same context in nine workers.
- Large feature: group a domain's entry points, core behavior and relevant tests where manageable. A transport/domain/storage split is useful only when the handoffs have explicit owners.
- Large codebase: delegate an inventory first, then divide by domains/packages; split oversized packages by behavior or state/resource ownership. Keep a compact map of source areas and review obligations. Excluded generated/vendor code still needs its relevant generator/dependency boundary assessed.
- A useful sizing warning is roughly 500 changed lines or 10 substantive files, or a much smaller concurrency/security change with several interacting states. This is a trigger to reconsider size, not a limit or permission to truncate. Include enough related code and tests to establish the contract. Split before assigning a packet that cannot fit with skill instructions, reasoning and output in one worker context.
- Create explicit obligations for important handoffs: identity/authorization propagation, error translation, context/deadline ownership, transaction boundaries, API/schema compatibility, and release/build identity where implicated. Attach each to a packet that can inspect both ends or a small boundary follow-up. Avoid an automatic second full audit.
- Assign module/toolchain/release configuration once where shared; consumers receive its compact conclusions and evidence references. A shared defect can affect several domains without becoming several causes.

The manifest records each chunk's target paths/contracts, assigned topics, worker/status and evidence artifact. It also records each boundary and every topic's applicable or not-applicable reason. Granularity follows actual risks; do not create a Cartesian product of every file and all nine topics.

## Dispatch packet

Provide:

```text
Review ID and snapshot/base/head:
Chunk ID; exact target paths/symbols and behavior boundary:
Mode: changeset (introduced/worsened) or code area (existing in scope)
Assigned topic skill names and resolved locations:
Relevant instructions, module/support versions and known contract facts:
Context paths you may inspect; cross-boundary obligations you own:
Review-only; checks must preserve user files; use disposable copies for mutating diagnostics:
Output path(s) and handoff format from reporting.md:
Report unresolved evidence or oversized work before silently narrowing scope.
```

Workers can read relevant additional files, recording why; they should ask for a separate packet if that reveals substantial new work. They should not recursively spawn untracked agents. Further delegation needs an orchestrator-assigned obligation and budget.

## Capability and cost

Use available capability descriptions, not provider/model names. Routine inventory or arithmetic needs less reasoning; interacting goroutines, authorization, compatibility and conflicting evidence need deeper reasoning and a model capable of reliable code tracing. Start with the user's configured model, adjust only where the host exposes and permits that choice, and preserve explicit user pins. If no model controls exist, change packet size, instructions and verification rather than pretending to select a model. Never select an unavailable model or a fixed cheapest/strongest model by name.

Reserve independent second looks for high-impact uncertain claims or unresolved contradictions. More agents are not evidence. Bound concurrency to host limits, run waves, and track complete/partial/blocked work. A retry must change the failing condition or narrow a genuinely oversized packet. On persistent blockage, retain useful reports and state what evidence is missing. A time budget prioritizes work; it does not redefine full coverage.
