# Next Go skills: context and concurrency

**Date:** 2026-10-02
**Status:** Scope and quality/source refinement approved by the user on 2026-10-02 (“looks good. onwards”). The user-requested [independent review](reviews/context-concurrency-spec/review.md) accepts the amended [written design](../superpowers/specs/2026-10-02-go-quality-build-context-concurrency-design.md) for planning. The [task-level implementation plan](../superpowers/plans/2026-10-02-go-quality-build-context-concurrency.md) is ready for review and execution-method selection. No new skill has been authored, evaluated or promoted.
**Starting revision:** `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96` (merged PR #5, including merged PR #4).
**Goal:** Improve agents' completed Go code through focused production cancellation and concurrency decisions, with measured benefit beyond the five existing build skills.

## Collection-wide authoring requirements

The user's 2026-10-02 clarification applies to all skills, including future revisions: clear, precise and focused; consistent in useful behavior across models, reasoning levels and harnesses; elegant Go with sound practices, minimal necessary machinery and no misleading shortcuts. Apply these requirements during authoring and independent review. Existing evaluated bytes change only through a substantiated correction and affected-case evaluation.

| Requirement | Concrete authoring and outcome check |
| --- | --- |
| Clear and precise | State the condition, decision and observable consequence. Distinguish a correctness requirement from an optional preference or project convention. Replace "handle appropriately" or "use best practices" with the relevant parent, owner, invariant, bound or contract. |
| Focused | Give each decision one owner and a discriminating trigger. Keep essential guidance in the entrypoint; add a conditional reference only when it resolves a real decision. Remove repeated tutorials, lists of familiar APIs and neighboring skills' instructions. |
| Reliable reuse | Use explicit conditions and ordinary terminology rather than personas, implied steps or harness-specific commands. Package one canonical copy with working relative links. Evaluate semantic decisions and preserved contracts; equivalent correct code need not have identical wording or structure. |
| Elegant Go | Make control flow, state protection and resource lifetimes easy to follow. Prefer a direct synchronous implementation when it satisfies the task. An abstraction must enable an actual contract or reduce demonstrated complexity. Code brevity alone is not elegance. |
| Avoid overbuilding | Require a concrete need for added APIs, packages, goroutines, queues, retry logic, configuration, frameworks or dependencies. Preserve small controls and supported consumers. Do not impose a directory layout, interface, worker mechanism or library merely because an example uses it. |
| Avoid footguns | Cover relevant failure exits and ownership transitions alongside success. Verify legal inputs, cancellation versus completion, partial results, aliasing and effective Go-version support. Do not promise safety from a channel send, a context argument, a timeout or a clean race run alone. |

Use a short condition → decision → check form where it clarifies fragile behavior. For example, detached required work needs a declared lifetime, a new time bound and a completion owner; simply using `WithoutCancel` is insufficient. Keep examples small, state their assumptions and verify their behavior. Avoid rigid templates when several choices satisfy the contract.

## Inspiration and source selection

The [scoped source review](next-group-source-review.md) pins and reviews relevant material from the requested Samber and spf13 collections, plus Uber and Google guidance. It records useful ideas and proposed adapt/omit decisions, including incomplete completion examples and overly broad synchronization/dependency rules. Go's primary documentation resolves language and standard-library behavior; project contracts decide policy. Source popularity and upstream grades are not promotion evidence.

This is planning inspiration, not a completed import audit. No content is copied or installed. Before retaining an upstream passage/example, audit that exact section and its references, check versions and preserve the applicable notices. Stop source discovery when the necessary decisions have credible support; expand it for a specific missing decision rather than to increase coverage by volume.

## Coordination and verified starting point

The user requested coordination with both existing chats. Their recent histories were read, each received a read-only handoff request, and both returned completed handoffs on 2026-10-02. GitHub merge/check state was independently read; the planning checkout contains both merge commits.

| Chat | Completed work and exact publication | Handoff for this plan |
| --- | --- | --- |
| **Go Skills (1) - testing** (`01a0f950-0e55-76b2-a8b1-1a8509372e2e`) | `go-behavior-tests` revision 2 and `go-test-isolation` revision 3 in package 0.2.0; [PR #5](https://github.com/joebartels/skills/pull/5) merged as `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96`; published head `15cd8561691e144106ab1815394f8f45c61f32b6`. Both CI validations succeeded. | Reuse evaluated testing guidance. Cover legal custom cancellation causes and distinguish requesting stop from observing completion. Production synchronization belongs in the next group; fixture mechanics remain with isolation. |
| **Go Skills (2) - Behavior & Test** (`01a0fa32-5737-76e1-96fd-851616e9d9c1`) | Reviewed error/value drafts plus names/docs reference; [PR #4](https://github.com/joebartels/skills/pull/4) merged as `309f4c961b42face459a5ce109d41c68351aacc0`; published head `570b73b1c0152af80e10b89b7f8cac68b0300e8a`. Both CI validations succeeded. This chat did not implement testing skills. | Build context/concurrency next. Assess error promotion separately with neighboring guidance frozen. All old reserved transfer tasks were consumed; create fresh holdouts. |

The five implemented/evaluated runtime skills are package boundaries, API contracts, interfaces/composition, behavior tests and test isolation. The [error draft](../../tests/go-quality-build/go-error-contracts/draft/SKILL.md) and [value draft](../../tests/go-quality-build/go-values-and-zero-values/draft/SKILL.md) remain outside runtime. [Names/docs](authoring-references/names-and-docs.md) remains a reference.

The [testing package review](../../tests/go-quality-build/results/2026-10-01-testing-combined-r3/whole-package-review/review.md) and [final integrity record](reviews/testing-delivery/final-integrity.json) document 49 reconstructions and 2,423 sealed artifacts. Those counts are inherited evidence, not rerun by this planning task. The [language-contract effectiveness review](language-contracts-effectiveness-review.md) records one contract-sensitive error improvement and strong value baselines with no measured gain. Its 27 trials include 11 clean pairs and one context-qualified combined pair. Neither draft needs recreation.

## Recommendation and alternatives

**Recommend two new candidates, completed sequentially:** `go-context-and-deadlines`, then `go-concurrency-and-ownership`. Their production decisions apply to libraries, CLIs, services and workers. Keep a separately bounded `go-error-contracts` promotion study after the new group; it can also be commissioned independently without blocking the pair.

An error/context/concurrency batch would match the earlier provisional candidate list, but would mix a previously reviewed draft's promotion question with two new authoring questions. Separating them makes each result attributable. Expanding composition instead would reduce skill count, but risks mixing abstraction choices with synchronization and time-budget decisions; use that option if evaluation shows the new guidance adds little beyond composition.

Defer new value, names/docs, telemetry, retry/overload, security, performance and release skills. Keep their candidate boundaries in the [canonical record](README.md); this group should establish the context and ownership foundations first. No quality-build coordinator is needed for this group.

## Skill boundaries

| Candidate | Trigger and decisions it owns | Exclusions and neighboring owners |
| --- | --- | --- |
| `go-context-and-deadlines` | Go work changes cancellable/blocking operations, propagation across calls, request or operation budgets, cancellation causes, or deliberate work beyond the caller's lifetime. Trace the operation's parent and deadline; choose child budgets, release cancel resources, and preserve the stated cancellation/result policy. | Pure calculations need no context. API owns supported signatures/promises; error contracts owns general failure representation; concurrency owns actual termination and synchronization. Retry policy, tracing and platform shutdown budgets stay with their future owners. |
| `go-concurrency-and-ownership` | Go work changes shared mutable state, goroutines, channels, concurrent admission, asynchronous completion, or resources used by concurrent work. Establish state and channel ownership, synchronization, bounds, stopping, joining, and resource-release order on success and failure. | Do not make serial work concurrent by default. Composition owns abstraction, construction and host-facing lifecycle design; values owns copying/aliasing; context owns budgets and propagation; testing skills own assertions and fixtures. Optimization requires performance evidence. |

These candidates must work independently and must not require an uninstalled draft. Cross-references identify decision ownership rather than mandatory skill loading. Preserve existing runtime bytes unless a reproduced contradiction justifies a separately evaluated change at its owner.

### Context guidance to investigate

- Trace caller cancellation and the remaining total budget through every blocking boundary; avoid replacing them with background contexts or resetting a full budget at each stage.
- Distinguish cancellation classification, an inspectable custom cause, independent operation failure and usable partial results. Derive precedence from the contract; do not assume cancellation always wins or that an error interface is comparable.
- Give derived cancel functions a visible owner and release them on each exit path. A cancellation request does not establish that work has stopped.
- Detached cleanup or background work needs an explicit reason, a new bound and a completion owner. Request cancellation, operation deadlines and cleanup budgets are separate choices.
- Context values support request-scoped metadata; dependencies and optional configuration need their existing explicit boundaries. Avoid breaking supported APIs to enforce a stylistic context rule.

Primary grounding: [context documentation](https://pkg.go.dev/context), [contexts and structs](https://go.dev/blog/context-and-structs), and [canceling database operations](https://go.dev/doc/database/cancel-operations). Context documentation explicitly separates cancellation from waiting and documents that `WithoutCancel` removes the inherited deadline and cancellation signal. A context-aware interface alone does not prove an external dependency stops promptly.

### Concurrency guidance to investigate

- Name who mutates state, publishes results, sends/closes channels and acquires/releases each resource. Choose the smallest fitting synchronization mechanism; neither channels, locks nor atomics are universal defaults.
- Establish a happens-before relationship for observations. Account for compound invariants, callbacks, lock ordering and blocking operations while locks are held.
- Make every owned goroutine's completion observable. Stop all affected work and join all started work before releasing resources still in use, including partial startup failure and another worker's error.
- Bound in-flight work when the specified workload can exceed capacity; make blocked admission and channel operations respond to stopping. Preserve the promised order, accepted effects and partial results.
- A timeout diagnoses non-cooperation; it does not kill a goroutine or permit closing a resource it still uses. Define the remaining ownership instead of claiming completion.

Primary grounding: [Go memory model](https://go.dev/ref/mem), [sync documentation](https://pkg.go.dev/sync), and [timer changes](https://go.dev/wiki/Go123Timer). Check timer and synchronization guidance against the effective toolchain and module; avoid universal old/new timer recipes or newer helpers that silently raise the supported minimum.

## Prospective evaluation plan

Freeze the requirements, cases, mutation/probe expectations, allowed guidance, model/effort settings and run budget before dispatch. Both arms receive the same exact package 0.2.0 catalog; only the candidate exposure changes. The new context skill is not added to concurrency's individual baseline. A later factorial case checks their interaction.

Use fresh authors with no planning-record, reviewer, arm-label, held-probe or repaired-candidate access. Record exact task/wrapper text and actual settings. Give reviewers neutral candidate identities and original contracts; freeze findings before revealing arms. An author may read the exact allowed skill catalog, not this record. Explicit selection and actual supported automatic routing remain separate claims.

### Frozen cases to prepare

| Candidate / case ID | Boundary and promised observation | Timing |
| --- | --- | --- |
| Context / `library-call-cause` | A cooperative library operation handles already-canceled and in-flight calls, legal non-comparable custom causes, independent failure and promised partial results without panic or cause loss. Specify the error policy before running either arm. | Discovery; primary repeat case. |
| Context / `service-total-budget` | Several outbound stages use the supplied cancellation and remaining overall deadline; an earlier caller deadline is respected. Faithful client observations cover request propagation and owned response completion. | Discovery. |
| Context / `cli-finalization-budget` | A long-running command stops caller work and performs only its explicitly required bounded finalization, with observed command outcome and retained effects. The fixture supplies the host lifecycle so context is the principal new decision. | New transfer task; run its baseline/exposure pair only after freezing the draft. |
| Context / `not-context-work` | A private deterministic calculation stays small and does not gain a context parameter or cancellation machinery. | Non-selection control. |
| Concurrency / `library-shared-state` | Concurrent updates and observations preserve a documented compound invariant, synchronize publication and retain separate-instance behavior without an unnecessary framework. | Discovery. |
| Concurrency / `service-owned-workers` | A bounded host starts several workers, stops admission, handles partial startup and worker failures, and observes every completion before resource release. Include a worker held in cleanup and blocked admission. | Discovery; primary repeat case. |
| Concurrency / `cli-pipeline-stop` | A bounded pipeline handles producer failure and early consumer return; blocked sends stop, completion is joined, and the promised accepted output/order and process failure remain observable. | New transfer task; run its baseline/exposure pair only after freezing the draft. |
| Concurrency / `not-concurrency-work` | A serial private calculation remains serial and gains no goroutine, channel, lock or worker pool. | Non-selection control. |

These are proposed fixtures, not runnable artifacts. Prepare exact input inventories and controller probes outside author-visible files. Do not call the old reader, CSV, borrowed-frame or integrated-indexer cases unseen holdouts. They may be labeled historical regressions only.

**Budget:** Per new candidate, three initial baseline authors (two discovery cases plus control), three corresponding exposure authors, one new transfer pair, and two further fresh pairs on the named primary repeat case: **12 author runs for the core study**. Use those two repeat pairs for the alternate-model and alternate-reasoning profiles below. Add at most two native-harness primary pairs per candidate: **16 author runs maximum per candidate**, **32 for the pair**. One new integrated task gets four fresh authors with neither/context-only/concurrency-only/both exposure: **36 author runs maximum for the new group**. This proposed ceiling supersedes the earlier 28-run proposal to include the user's reuse requirement; no runs have been commissioned or performed. Independent reviews, selection probes and verification executions are counted separately. Fix actual settings before execution, record tokens/time where available, and do not infer costs from run counts.

### Reuse and portability profiles

Freeze a reference configuration and the following profiles before dispatch. Each profile's baseline/exposure pair uses identical inputs, neighboring skill bytes, tool permissions, model, reasoning setting and harness; only candidate exposure differs. Use fresh contexts and the final frozen candidate bytes.

| Profile | Controlled comparison | Author allocation per candidate |
| --- | --- | --- |
| Reference | One declared model, reasoning level and native harness; discovery, control and transfer. | Eight core authors. |
| Alternate model | Change model while holding reasoning level and harness fixed when supported. Use the primary repeat case. | One core repeat pair: two authors. |
| Alternate reasoning | Change reasoning level on the reference model/harness. Use the primary repeat case. | One core repeat pair: two authors. |
| Other native harnesses | Run the same primary contract through each of the other two target harnesses among Claude Code, Codex and OpenCode. Prefer identical model/settings where genuinely supported; otherwise record the native supported configuration and disclose that harness and model effects are confounded. | Up to two additional pairs: four authors. |

Record executable/harness version, actual model and reasoning identifiers, toolchain, loaded skill hash and reference access. Verify native discovery/loading through the harness itself. Forced file reads, packaging validation and one harness impersonating another cannot establish automatic routing or runtime support.

For each available profile, run selection-only probes for the same applicable request and non-selection control, with a meaning-preserving paraphrase of each. Preserve actual selection/loading evidence. These probes check routing without producing an implementation; count them separately and do not mistake them for behavioral author runs. Explicit selection and automatic selection remain separately reported.

Judge reuse by the promised behavior and unnecessary complexity, not identical code. Exposure must preserve the contract and avoid material regressions in every tested profile. Require the individual repeated-benefit gate where baseline defects occur; a strong alternate baseline may tie without invalidating otherwise supported benefit. Unknown settings, unavailable harnesses or inability to isolate an axis remain named gaps. Broader reuse readiness stays unverified until the intended native harness profiles are exercised; any narrower promotion must state its evidence limits.

This samples two models, two reasoning levels and three native harnesses where available. It does not test their full cross-product or guarantee invariance across every model or setting. Keep remaining sensitivity visible rather than silently adding runs. Revisions that affect reuse need affected-profile reruns within the ceiling.

If discovery shows no confirmed relevant weakness, stop before drafting and preserve the strong results. Do not spend transfer/repeat slots searching indefinitely for a win. Infrastructure failures remain excluded from benefit evidence but count toward the author-launch ceiling; retry at most once using an unspent slot and disclose the resulting coverage gap. A material candidate revision requires affected final-byte reruns within the stated budget; if the budget is insufficient, stop at a reviewed draft and propose a declared extension. Do not replace failed evidence or silently enlarge the study.

**Acceptance:** A candidate needs a distinct trigger and useful decision procedure, repeated correction of a confirmed defect on the primary case, successful new transfer behavior, preserved strong/control outcomes and no confirmed material regression. Report task-level defects and extra API/code/dependency costs; do not average grades or infer broad effectiveness from a small sample. If another discovery case provides the only benefit, a new repeat design must be declared before consuming its remaining slots. All-A ties justify retaining a draft/reference, not promotion.

Freeze up to four realistic semantic mutations per candidate before exposure. Check that they compile and fail the expected meaningful assertions. Include unsafe cause equality or lost custom-cause policy, lost caller budget, skipped join/resource-close order, and blocked pipeline/admission termination where applicable. Later discoveries are labeled supplementary; they cannot retroactively become primary outcomes. Existing survivor mutants justify prospective coverage but are not themselves proof of a new skill's benefit.

Run candidate build, test, vet and formatting checks; race detection and bounded shuffle/repetition on exercised concurrent paths; actual child-process outcomes for CLI claims; faithful HTTP checks at the claimed boundary. Execute the actual minimum Go 1.22 toolchain and an available supported newer toolchain when feasible. A `go 1.22` directive under a newer compiler is not minimum-toolchain evidence. Use standard-library fixtures by default and avoid new dependencies unless the task requires them. Restricted listeners, missing toolchains or unavailable platforms produce scoped limitations, not fabricated passes.

Independent outcome review accounts for applicability and coverage of all nine topics under the unchanged `go-quality-report` contract. Correctness and Code Quality & Idioms assess changed Go outputs; Architecture assesses exposed ownership/API changes, Resilience budget propagation, Performance bounded-resource claims and Testing assertion sensitivity. Route Dependencies & Reproducibility, Security and Deployment & Operations when the actual module/toolchain/build, trust/abuse/sensitive-data or process/shutdown/release boundaries implicate them. Review only applicable topics in bounded packets, record non-applicability reasons and evidence gaps separately, and retain topic cards and linked corrections. Duplicate symptoms share one root cause; full-scope material coverage gaps remain Insufficient evidence, not clean grades. Run the four-arm integrated case when both candidates pass individual acceptance; it checks compatibility and conflict ownership and cannot replace individual effectiveness evidence. If only one qualifies, skip the four-arm study and review that skill with the existing package; do not invent a rejected neighbor merely to populate an arm.

## Work sequence and artifacts

| Stage | Deliverable / prospective paths | Exit gate |
| --- | --- | --- |
| 1. Settle the design and freeze evaluation inputs | Approved context/concurrency spec and implementation plan under `docs/superpowers/`; `docs/go-quality-build/context-concurrency-source-audit.md`; fixture manifests in `tests/go-quality-build/<candidate>/evals/`. Start from the exact new pins in the [scoped source review](next-group-source-review.md), complete section-level copy/adapt/omit decisions for any retained content, assign decision owners and verify primary-source/version behavior. Preserve the earlier Samber pin and audits for historical work. | Reviewable scope, exact contracts, verified original fixtures/probes, frozen catalog/profiles/budget and intact prior archives. Import acceptance remains incomplete; retain applicable license/attribution notices for substantial reuse. |
| 2. Context candidate and individual evaluation | Discovery/control baselines, then only justified guidance in `tests/go-quality-build/go-context-and-deadlines/draft/`; exact snapshots and comparison evidence under `tests/go-quality-build/results/<run>/`. | Complete the bounded core/portability study within 16 authors or record the earlier stop and available-profile limits; independent exact-guidance decision. No runtime copy before acceptance. |
| 3. Concurrency candidate and individual evaluation | Equivalent baseline/draft/evidence paths for `go-concurrency-and-ownership`, using the unchanged five-skill comparator catalog. | Complete its own effectiveness and reuse gates before expanding; no credit borrowed from the context result. |
| 4. Interaction and package delivery | Four-arm new combined fixture under `tests/go-quality-build/context-concurrency-combined/evals/`; portable reconstruction/seal records; accepted exact-byte skills under `plugins/go-quality-build/skills/`; package metadata/navigation only for accepted promotions. | No unresolved guidance conflicts or material harm; independent whole-package review, repository/harness validation and final runtime-to-evaluated-snapshot equality. Version is chosen from actual accepted scope at delivery. |
| 5. Separate error-promotion study | Reuse the exact [reviewed error draft](../../tests/go-quality-build/go-error-contracts/draft/SKILL.md); new prospective cases and a frozen neighboring-skill catalog. | Individual attributable benefit and compatibility evidence; promote only after its own accepting review. Context/concurrency success does not promote errors automatically. |

The error study should explicitly settle simultaneous input/parse/finalization precedence before authors see the cases. Use three new applicable tasks (multi-cause input with valid prefix, buffered completion with borrowed/owned resources, and public classification with private diagnostics) plus one private control, with the same 12-author-run ceiling if commissioned. Fix the neighboring catalog at that study's start and toggle only error exposure. Preserve the old 614-word draft and exact-byte review; edit only for a substantiated guidance defect. Value promotion remains a later question, since its current evidence shows no incremental benefit.

Use the existing eval schema and validator rather than inventing a new orchestration framework. Store prompts, patches, source hashes, declared/observed exposure, raw results and portable reconstruction instructions once per run; link comparisons to those records. Keep immutable evidence separate from maintained status. Update the canonical status/work log after every meaningful stage.

## Planning verification and next action

The initial planning task checked both completed handoffs, exact local merge ancestry, live GitHub merge/check state, existing runtime/draft boundaries and primary Go documentation. This refinement additionally reviewed pinned source sections, local review decision references and the user's collection-wide requirements. Repository validation, local links and documentation whitespace are checked before handoff. Historical behavioral studies are cited without claiming fresh replay. No new behavioral run, accepted upstream import or portability result exists.

Next action: review the task-level implementation plan and select its execution method before authoring either new skill. The written design has passed the requested independent alignment review; scope remains two focused candidates with separate promotion gates. Error/value drafts and existing testing guidance retain their recorded dispositions.
