# Context and concurrency source audit

**Date:** 2026-10-02
**Status:** Section-level authoring audit complete; original wording/examples selected. No upstream text or code imported, no candidate authored or promoted.

The [approved design](../superpowers/specs/2026-10-02-go-quality-build-context-concurrency-design.md) and [scoped source review](next-group-source-review.md) define this bounded audit. The [file inventory](context-concurrency-source-inventory.json) records the exact revisions, SHA-256 values and sizes for 16 guidance files and four root licenses retrieved during planning. All 20 local captures reverified against that inventory before this audit. Temporary captures are conveniences, not runtime or replay dependencies; retrieve each repository/path at its recorded revision to reproduce the inventory.

## Sources and copy scope

| Repository | Revision | License at pin | Retained scope |
| --- | --- | --- | --- |
| [Samber](https://github.com/samber/cc-skills-golang/tree/8e899e20ff0cd4dc524af3993e4c62d8ee8c5717/skills) | `8e899e20ff0cd4dc524af3993e4c62d8ee8c5717` | MIT, Samuel Berthe 2026 | Decision inspiration; zero copied prose/code |
| [spf13](https://github.com/spf13/go-skills/tree/9ac6eca43161163bb21621a520a57df50d5ad464) | `9ac6eca43161163bb21621a520a57df50d5ad464` | MIT, Steve Francia 2026 | Selected clarity/lifetime decisions; zero copied prose/code |
| [Uber](https://github.com/uber-go/guide/tree/1d60a91aa5e87d443002e23c21903c49489dbde5) | `1d60a91aa5e87d443002e23c21903c49489dbde5` | Apache 2.0 | Completion ownership inspiration; zero copied prose/code |
| [Google](https://github.com/google/styleguide/tree/fc981500047ef2cc2e2509b00dcd23c9badaf8e4/go) | `fc981500047ef2cc2e2509b00dcd23c9badaf8e4` | CC BY 3.0 | Selected clarity/simplicity/context/lifetime sections; zero copied prose/code |

There are no copy decisions in this audit. “Adapt” means independently express the bounded decision, after semantic/contract checks, rather than port the example. A future substantial copy changes this disposition and requires the distributed copyright/license/attribution/change notices in the same accepted revision. Do not infer that an upstream license covers all linked dependencies. Historical architecture/testing audits retain their earlier Samber pin.

## Section dispositions

Paths below are relative to the named repository. Every section of the two Samber families is accounted for; Google/spf13 are intentionally selected-section reviews rather than whole-guide adoption.

| Source section | Decision | Local owner / rationale |
| --- | --- | --- |
| Samber `golang-context/SKILL.md`: summary, Creating Contexts, Propagation | Adapt | Context: parent-derived scopes, propagation and cancel ownership. Child contexts are allowed; do not demand unchanged context identity. Existing supported APIs outrank a stylistic signature rewrite. |
| Same: Deep Dives | Adapt | Conditional local explanation only when an observed decision needs it; no whole-manual import. |
| Same: Cross-References, Enforce with Linters | Omit | No absent-skill prerequisites, harness commands or mandatory lint suite. Diagnostic inspection remains ordinary verification. |
| `cancellation.md`: Cancellation/defer cancel, Timeouts/Deadlines/nesting | Adapt | Context: cancel owner and narrower children within total caller budget. Cancel requests stopping, not completion. |
| Same: Listening/select/loop checks | Adapt | Context policy plus concurrency blocking mechanics. Replace the example's first-error return without joining with explicit ownership/completion obligations if concurrent work is actually owned. |
| Same: AfterFunc | Adapt | Context/ownership: stop does not join an already-started callback; unsafe resource-release illustration is not retained. |
| Same: WithoutCancel | Adapt | Context: require a reason, new bound and completion owner. Detached audit goroutine example is not retained as a complete pattern. |
| `http-services.md`: servers and service calls | Adapt | Context: supplied caller scope, total/stage budgets and body lifetime. Schematic signatures/results and cleanup must be executable if an example is retained. Host shutdown API design stays with composition. |
| Same: middleware enrichment | Omit | Metadata transport/tracing framework is outside this group; no new host machinery for its own sake. |
| `values-tracing.md`: values/parameters | Adapt | Context: appropriate request metadata, not dependencies/optional configuration. Preserve metadata only when the task needs it. |
| Same: cross-service tracing | Omit | Telemetry/header/SDK policy has a different owner. |
| Samber `golang-concurrency/SKILL.md`: Core Principles, Channel/Mutex/Atomic | Adapt | Concurrency: invariants and smallest fitting mechanism; neither channel handoff nor copies imply alias safety. Compound state needs one coherent synchronization policy. |
| Same: WaitGroup/errgroup, primitive reference, checklist | Adapt | Observable completion and version-aware mechanics. Do not replace every WaitGroup or require a dependency. Limited errgroup admission itself may block. |
| Same: pipelines/pools, common mistakes | Adapt | Capacity/admission, early consumer completion and join/release order, when required by workload/contract. |
| Same: parallel audits, cross-references, profile/reference catalog | Omit | No harness orchestration, personas, dependency checklists or profiling/pool work without an actual decision. |
| `channels-and-select.md`: lifetime, direction, closure, buffer, select | Adapt | Concurrency: designated closure coordinator after all sends, actual buffer bound, cancellation where a blocking contract needs it. Not every channel must close or every select include Done. |
| Same: panic recovery | Omit | No automatic recovery boundary; failure policy requires the actual task. |
| Same: repeated time.After | Adapt only if encountered | Effective Go/module timer semantics and measured cost govern advice; no universal old-version drain/reset recipe. |
| `sync-primitives.md`: Mutex/RWMutex/embedding, atomic, Map, Once | Adapt | Concurrency: protect invariants, avoid copying used locks; specialized sync.Map use, independent atomics and callbacks need their actual contracts. Zero-value/lifetime concepts are local necessities, not a dependency on value/composition drafts. |
| Same: WaitGroup and Go 1.25/fallback | Adapt | Go 1.22-compatible examples use Add before launch plus Done/Wait. Newer helpers require effective-version support; no minimum bump. |
| Same: errgroup/SetLimit | Adapt | Dependency optional; blocking admission and stop/join behavior must be explicit. |
| Same: Pool/singleflight | Omit | Optimization/coalescing is outside the evidenced need; no speculative performance recipe. |
| `pipelines.md`: pipeline, fan-out/in, worker pool, semaphore, mistakes | Adapt | Concurrency: blocked receives as well as sends, bounded admission, closure ownership and joined early completion. No framework default. |
| Same: iterators, samber/ro, leak-detection library | Omit | Optional library/modernization/testing expansion has no demonstrated need. |
| spf13 `go/SKILL.md`: clarity, early return, goroutine stop/bounds, WithoutCancel, timeout/shutdown | Adapt | Readable normal path and lifecycle policy under existing contract. Detachment needs bound/owner; host setup remains composition. Channel preference, copy preference and fixed timeout values are not universal. |
| Same: remaining architecture/configuration/testing/modernization recipes | Omit | Existing/future decision owners; no broad replacement handbook or universal libraries/layout. |
| spf13 `go-spec-reviewer/SKILL.md`: all review/calibration/orchestration sections | Omit | Inspect for conflict only. Keep the existing nine-topic rubric; no personas, tools, Cobra/Viper rules or grading imported. |
| Uber `goroutine-exit.md`, `goroutine-forget.md` | Adapt | Concurrency: observation of owned completion. Current wg.Go snippet is not Go 1.22 compatible; do not copy it. |
| Uber `channel-size.md` | Omit rule / adapt question | Buffer zero/one is not universal. Derive actual capacity and admission from the workload. |
| Uber `mutex-zero-value.md` | Adapt | Minimal usable synchronization and no copied used lock; avoid constructors/embedding rules added merely for style. |
| Google `go/guide.md`: Clarity, Simplicity | Adapt | Normal flow and concrete ownership visible; direct code rather than unnecessary abstraction. No other guide sections adopted. |
| Google `go/decisions.md`: Goroutine lifetimes, Synchronous functions, Contexts/custom contexts | Adapt | Prefer synchronous completion where sufficient; lifetime is explicit; preserve compatible context boundaries. No other decisions adopted. |

## Semantic and review checks

Primary documentation was checked during planning and rechecked for version-sensitive context/errgroup/sync points on 2026-10-02. [Context](https://pkg.go.dev/context) specifies cancellation/cause and cleanup semantics; [WithoutCancel](https://pkg.go.dev/context#WithoutCancel) drops inherited cancellation/deadline, while [AfterFunc](https://pkg.go.dev/context#AfterFunc) stop does not wait for callback completion. Both helpers were added in Go 1.21 and fit the fixture floor only when their policy is justified. The [memory model](https://go.dev/ref/mem) grounds publication; [sync](https://pkg.go.dev/sync) grounds copying/invariant choices. [WaitGroup.Go](https://pkg.go.dev/sync#WaitGroup.Go) requires Go 1.25; [errgroup.Go](https://pkg.go.dev/golang.org/x/sync/errgroup#Group.Go) can block under its active limit. No errgroup dependency is introduced by this audit.

Audit decisions match existing API/package/composition/testing ownership and the unchanged all-nine review contract. General error representation/aliasing drafts stay non-installable; necessary cancellation inspection and synchronization rules remain locally usable. Outcome reviews still judge actual correctness, idioms, ownership, budget, resource and test contracts rather than this section inventory. Structural/source checks do not prove effectiveness, native routing or actual Go 1.22 execution.

Next: prepare/freeze original fixtures and probes, then run discovery baselines before deciding whether new guidance is warranted. No useful source decision requires expanding repository discovery now.
