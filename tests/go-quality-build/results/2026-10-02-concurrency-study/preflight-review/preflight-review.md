# Concurrency exact-draft preflight — Accept

Date: 2026-10-02. Target: [the concurrency draft](../../../../../tests/go-quality-build/go-concurrency-and-ownership/draft/SKILL.md), SHA-256 `0ab21aee46b9446356b455d59776263a8160fcf2a0c3b5050c2b1792ecf13002`, 5,274 bytes, 756 whitespace-separated words. This independent guidance review accepts these exact bytes for snapshot freeze and candidate exposure. **No material guidance change is required.** This is neither a behavioral utility finding nor a runtime promotion decision.

## Scope and evidence

Applied skill-creator and writing-skills principles against the [accepted design](../../../../superpowers/specs/2026-10-02-go-quality-build-context-concurrency-design.md), canonical continuation record, source audit/inventory, all five existing runtime build skills and their references, and the unchanged review collection. Read each topic's applicable review questions and decision reference; checked umbrella scope, coverage, grading and reporting rules. This is a guidance alignment review, not nine Go outcome grades.

[Input identities](raw/input-identities.json) establish that all eight files across the five runtime build skills and all 23 runtime review files are byte-equal to accepted comparator revision `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96`. The exact context draft is an uninstalled neighbor, SHA-256 `6368240e5a000cd8f056afa5980099317882d8ebcefa0e1f29921975ccc4627d`; the concurrency draft works without it. No draft, runtime skill, evaluation input, or Git index/ref was changed by this review. Review artifacts reside only in this preflight directory; Go caches reside in owned temporary locations.

## Findings and decision fit

Material findings: **None**.

The description gives concrete concurrent triggers and a serial-edit exclusion without substituting a process summary for the body. The body is a self-contained decision guide: identify the invariant and owners, choose a fitting synchronization mechanism, define what capacity counts, keep admission and completion supervision live, establish sender completion before closure, and release only after the relevant users finish. The one small example makes the compound-observation decision concrete. Its length does not justify extra references or a router: the paragraphs cover different consequential decisions, and no copied manual, framework, script dependency or placeholder appears.

The shared stop/join/release obligation agrees with composition. The useful addition is mechanical: sparse open input, observing Run failures during input/acquisition/capacity waits, the lifetime counted by a lease bound, acquired-but-never-started ownership, coordinated pipeline termination, and safe channel closure. The paragraph at line 40 is not the sole basis for a separate skill. Whole-cohort joining is explicitly conditioned on a shared lifecycle boundary; it is not exported as a universal global-close rule. Callbacks or I/O under locks, buffering, recovery, and the choice between channels and locks remain conditional on actual invariants and contracts.

The discovery primary case remains `service-owned-workers`. F1 is full-cohort prefill preventing an available first job from starting while caller input remains open. Line 36 directly addresses that progress obligation. F2 is the separate delayed failure-supervision defect: collecting every startup completion can leave a cooperative acquisition awaiting the stop that a Run failure should request. The same paragraph distinguishes that state and explains why supervision must remain observable. Neither the primary cause nor the four mutation targets was redefined. All 23 original eval/probe files still match their [frozen hashes](raw/discovery-contract-identities.json); the initial relative lookup missed the evals root and is retained separately, while the corrected check requires and verifies all 23 files.

The frozen mutation meanings remain split compound publication, skipped Run join before cohort release, unresponsive input/admission, and omitted coordinated stopping on early consumer completion. The draft covers each mechanism without turning later reviewer diagnostics into frozen efficacy targets. This review does not award mutation sensitivity credit or rerun author outcomes.

## Neighbor alignment

| Existing owner | Alignment of this draft |
| --- | --- |
| Package boundaries | Preserves placement, dependencies and APIs; adds no package hierarchy. |
| API contracts | Starts from supported effects and result/error promises; adds no new signature or error policy. |
| Interfaces and composition | Preserves the supplied host lifecycle and dependency design; supplies synchronization, admission and supervision mechanics inside that boundary. |
| Behavior tests | Its short verification paragraph identifies relevant concurrent observations; general case/oracle/assertion strategy remains with this owner. |
| Test isolation | Calls for held work/acquisition/release events and bounded cleanup, without adding a fixture framework, clock seam or leak dependency. |
| Uninstalled context neighbor | Cancellation/time-budget decisions remain with the operation contract. Stopping is distinguished from joining; no neighbor installation or helper is required. |

| Review topic | Alignment and preserved boundary |
| --- | --- |
| Architecture | Ownership remains visible; no mandatory abstraction or host redesign. |
| Code Quality & Idioms | Small pointer-receiver mutex example; version-aware mechanisms, simple flow and no automatic recovery. |
| Correctness & Compatibility | Coherent publication, reachable blocking outcomes, retained effects and independent failures follow supported contracts. |
| Testing | Race success is not treated as proof of coherent state or complete ownership; observations use the actual failure path. |
| Observability & Resilience | Admission and stop supervision remain live; budgets, retries and telemetry are not prescribed. |
| Performance & Resources | Defines what is bounded and holds capacity through release; avoids unsupported RWMutex, atomic or lock-free speed claims. |
| Dependencies & Reproducibility | Keeps supported Go versions and dependencies; groups are optional and newer helpers require effective-version support. |
| Security | Capacity can support actual abuse limits, while this draft introduces no trust-policy or scanner prescription. |
| Deployment & Operations | Does not invent signal, process, container or release machinery; host termination policy remains supplied. |

The draft neither modifies findings, severities, grades, all-nine applicability/coverage, root-cause deduplication nor review permissions. Its test paragraph cannot replace the independent outcome review contract.

## Source, semantic and notice checks

All 20 pinned guidance/license captures match the [source inventory](raw/source-inventory-verification.json). The draft follows the audit's adapt/omit decisions: original wording and a small original example; no upstream personas, forced delegation, absent-skill prerequisite, blanket channel/buffer default, automatic panic recovery, mandatory errgroup/goleak dependency, timer recipe or minimum-version upgrade. Inspection found no substantial upstream prose or example import. A [12-word normalized overlap aid](raw/source-overlap-check.json) found no runs across the 16 guidance captures; that heuristic supports inspection and is not a legal threshold or proof of authorship. No new distributed notice obligation is established by these exact draft bytes.

The [Go memory model](https://go.dev/ref/mem) supports explicit publication ordering and synchronization. The warning about separate atomic fields follows from distinct atomic operations admitting an interleaving between their updates or reads; it is a reasoned application of those guarantees. A channel send orders communication but does not remove a remaining storage alias.

[Sync documentation](https://pkg.go.dev/sync) and both actual local toolchain sources support usable zero-value mutexes, no copying after use, registration before launch, and waiting for completion. The host API history independently records `WaitGroup.Go` in Go 1.25; the draft uses no such helper. [Raw local checks](raw/local-primary-doc-verification.json) and [registration excerpts](raw/waitgroup-registration-docs.json) retain paths, hashes and locations. The minimum distribution lacks the later API-history file, so that floor is grounded in the host record and current official documentation.

The current [errgroup documentation](https://pkg.go.dev/golang.org/x/sync/errgroup) confirms that limited `Go` admission can block, and that `Wait` observes all submitted functions but returns the first nonnil error. The draft's warning correctly leaves the failure/stop controller able to run and retains the promised independent-result contract instead of assuming a group supplies every required outcome. No errgroup dependency was added.

The official [pipeline cancellation article](https://go.dev/blog/pipelines) supports stopping unnecessary upstream production after early downstream completion and closing outputs after sends finish. The draft extends that principle to whichever sends, receives or admission may actually block under the contract, without mandating Done in every select. Both local context implementations explicitly document that cancellation does not wait for work to stop; safe release therefore still needs observed user completion. Current primary-page tool responses are retained in [open evidence](raw/primary-web-open.txt) and [targeted evidence](raw/primary-web-find.txt).

## Exact example verification and limits

Extracted the Go code fence unchanged, adding only `package example` and `import "sync"` in [totals.go](example/totals.go). [Extraction identity](raw/exact-example-extraction.json) records its hash. The review harness observes zero initialization, correct positive/negative updates, stable previously returned scalars, separate instances, and concurrent snapshots satisfying `sum == 7*count` throughout updates. It does not copy a used mutex.

[Raw verification](raw/example-verification.json) records 15 successful commands: actual Go 1.22.12 and host Go 1.26.5 identity/environment, build, tests, vet, race/shuffle with five repetitions, and empty gofmt diffs on each; plus quick_validate success. `GOTOOLCHAIN=local` prevents switching and `GOPROXY=off` excludes fetching dependencies. All checks ran on darwin/arm64. Race success establishes only the exercised example paths; inspection supplies the mutual-exclusion argument. No lock-free performance, other-platform, native-discovery, author behavioral utility or promotion claim follows.

Canonical/source-audit stage descriptions were still at the prior no-concurrency-draft stage when read. The controller was notified to update those continuation records with this exact review and results before finishing the authoring stage. That record maintenance requires no draft revision and is separate from the Accept decision. Next action: freeze these exact bytes and settings, then conduct the authorized matched study within its existing budget and unchanged primary/mutation contracts.
