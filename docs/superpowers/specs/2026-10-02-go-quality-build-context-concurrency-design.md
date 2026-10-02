# Go quality build: context and concurrency design

**Date:** 2026-10-02
**Status:** Written design for review. The user approved the scope and source/quality refinement with “looks good. onwards”; implementation, new author trials and runtime promotion have not begun.
**Base:** `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96`; package `go-quality-build` 0.2.0.

## Intent and scope

Help agents write correct, readable Go when cancellation, deadlines, shared state or asynchronous work matter. Deliver two focused candidates, `go-context-and-deadlines` and `go-concurrency-and-ownership`, completing each individual study before proceeding to the next. Conclude an unsuccessful study honestly; promotion of the first candidate is not a prerequisite for studying the second.

The user's quality requirements apply throughout: clear and precise instructions, useful consistency across models/reasoning levels/harnesses, elegant Go, minimal necessary machinery, and avoidance of vague rules and footguns. Measure preserved contracts and concrete decisions, not matching prose or identical generated code. A small, direct implementation that satisfies the task is preferable to extra abstraction.

The [approved scope/work proposal](../../go-quality-build/next-group-plan.md), [scoped source review](../../go-quality-build/next-group-source-review.md) and [canonical progress record](../../go-quality-build/README.md) provide background. This document settles the design; a separate task-level implementation plan follows its review.

Existing package/API/composition/testing skills remain the frozen comparator. Error/value drafts remain non-installable, names/docs remains a reference, and the optional error-promotion study is separate work. No coordinator, retry framework, telemetry package, general Go handbook, global installation or new evaluation framework belongs in this group.

## Responsibilities and ownership

| Decision | Owner |
| --- | --- |
| Caller cancellation, derived deadlines, total versus stage budgets, cancellation cause policy and deliberate detachment | Context candidate |
| Shared-state invariants, publication, channel closure, capacity/admission, goroutine completion and release after completion | Concurrency candidate |
| Supported API/process/wire promises and compatibility | Existing API contracts |
| Package placement and dependency direction | Existing package boundaries |
| Abstractions, dependency wiring and host-facing lifecycle interfaces | Existing interfaces/composition |
| General error representation and copying/aliasing contracts | Error/value decision owners; this group retains necessary local context/ownership rules without depending on their drafts |
| Regression observations and reliable fixtures | Existing behavior tests and test isolation |
| Independent findings, severity and grades | Unchanged Go quality review collection |

Composition already explains host lifecycle and joining started runs. The concurrency candidate must add useful decisions about synchronization, blocking, admission and completion mechanics beyond that existing advice. Do not duplicate lifecycle paragraphs to manufacture a new skill. If the candidate adds no measured utility, retain a reviewed draft/reference or propose merging the useful guidance at its actual owner.

### Context candidate

Apply when a task changes cancellable/blocking work, propagation through calls, operation budgets, custom cancellation causes or intentionally longer-lived work. A pure calculation, an incidental existing `ctx` parameter or a mechanical edit does not activate the skill by itself.

Its compact decision procedure should establish:

1. The caller's parent, cancellation signal and remaining deadline at each relevant boundary. Derive a child only for an actual narrower budget or independently owned cancellation scope.
2. Who releases each derived cancel function, including early exits or explicit transfer. Cancellation requests stopping; it does not establish completion.
3. The task's policy for successful results, partial progress, independent failures, cancellation classification and inspectable causes. Avoid interface equality on arbitrary errors and avoid a universal cancellation-wins rule.
4. Whether work must outlive the caller. If required, give it an explicit new bound and completion owner; detachment alone supplies neither. Preserve request metadata only where needed.
5. What the boundary actually proves: accepting a context is not proof that a driver or remote dependency stops promptly.

Preserve supported signatures rather than imposing context on every function. Context values carry appropriate request metadata, not hidden dependencies or optional configuration. Timer/reset recipes, HTTP timeout durations and newer helpers require version/contract justification; do not raise a project's minimum for convenience.

### Concurrency candidate

Apply when a task changes concurrent access to mutable state, goroutines, channels, admission, asynchronous completion or resources used by concurrent work. Keep serial work serial when it satisfies the contract.

Its compact decision procedure should establish:

1. Who owns each mutable invariant and what synchronization makes observations valid. A channel send does not remove aliases; independent atomics do not automatically protect a compound invariant.
2. The smallest fitting mechanism: ordinary lock, explicit handoff, bounded workers, semaphore, existing group or synchronous flow. Neither channels nor locks are universal defaults. Dependencies and optimization require concrete benefit.
3. Capacity, queue behavior and who admits work. Blocking admission, sends and receives need termination paths when the contract requires cancellation responsiveness.
4. Who can close each channel after all possible sends finish. A coordinator can own closure; not every channel needs closing.
5. How all started work is stopped and joined on success, partial startup failure, worker error and caller cancellation. Release resources only when their users are finished, or explicitly retain ownership when completion cannot be established.

A timeout does not kill a goroutine or make resource closure safe. Recovery, locking across callbacks/I/O and buffering follow the actual invariant and failure contract, rather than blanket prohibitions or automatic recovery. Keep the normal path easy to read alongside its cleanup paths.

## Runtime and source design

Keep one canonical runtime copy under `plugins/go-quality-build/skills/`. Drafts live under `tests/go-quality-build/<candidate>/draft/`; author evaluations consume immutable snapshots of those exact bytes. Main skills contain essential conditions and decisions. Add a conditional local reference only when a demonstrated decision needs additional explanation; do not create empty reference/script/asset directories or copy full manuals.

The scoped review pins Samber at `8e899e20ff0cd4dc524af3993e4c62d8ee8c5717`, spf13 at `9ac6eca43161163bb21621a520a57df50d5ad464`, Uber at `1d60a91aa5e87d443002e23c21903c49489dbde5`, and Google at `fc981500047ef2cc2e2509b00dcd23c9badaf8e4`. Complete `docs/go-quality-build/context-concurrency-source-audit.md` before authoring: inventory selected sections/references, record copy/adapt/omit, local owner, rationale, version evidence and notice obligations. Existing audits retain their historical pins.

Prefer original wording and minimal examples derived from the decisions. Any retained upstream example must have checked assumptions, valid signatures, relevant failure paths and executable behavior. Substantial copying carries the applicable distributed license/attribution/change notices. Remove upstream personas, tool lists, automatic delegation/configuration, fixed harness paths and absent-skill dependencies.

Primary semantic grounding is [context](https://pkg.go.dev/context), the [memory model](https://go.dev/ref/mem), [sync](https://pkg.go.dev/sync), [pipeline cancellation](https://go.dev/blog/pipelines) and [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup). Recheck version-sensitive advice during authoring against supported modules/toolchains. In particular, detachment removes inherited deadlines, stopping an `AfterFunc` callback does not join it, and limited `errgroup.Go` admission can block. These observations constrain examples; they do not prescribe a library or architecture.

## Fixture contracts

Prepare fresh original fixtures and ordinary task READMEs before authoring. Freeze exact APIs, input inventories, supported versions, outputs, precedence and controller observations before dispatch. Preserve the existing eval JSON shape (`skill_name`, `evals`, `id`, `prompt`, `expected_output`, `assertions`, `files`); authors receive the prompt and listed input files, with expected judgments/probes withheld.

| Candidate / case | Contract to freeze and observe |
| --- | --- |
| Context / `library-call-cause` | A cooperative sequential library operation accepts a caller context and processes work until completion, operation failure or observed cancellation. Already-canceled calls invoke no work. Successfully accepted progress remains available on failure. At the documented return decision, independently failing work remains inspectable; observed cancellation preserves its standard classification and custom cause, including a legal non-comparable cause. Cancellation arriving after completed success does not retroactively invalidate that success. Controlled gates establish coincidence and observation order. Primary benefit/repeat case. |
| Context / `service-total-budget` | A service performs several outbound stages under one configured total operation budget and any earlier caller deadline. Stage limits derive from that parent, with no fresh full budget per stage. Requests and owned response consumption stay within the operation scope; bodies are closed on all relevant exits. Inspect request propagation, already/in-flight cancellation, earlier parent limits, response lifetime and confirmed completed progress. |
| Context / `cli-finalization-budget` | A command with fixture-supplied host lifecycle stops cooperative processing on caller cancellation and finalizes already accepted work under a separate stated bounded scope. Cleanup has a completion owner. Actual child-process observations establish accepted effects and the specified failure status when processing or finalization fails. The author changes context policy rather than inventing signal/lifecycle machinery. Transfer only after draft freeze. |
| Context / `not-context-work` | Repair a private deterministic calculation with a focused ordinary regression observation. Preserve its API; add no context parameter, deadline, cancellation machinery or unrelated abstraction. |
| Concurrency / `library-shared-state` | Concurrent updates and snapshots preserve a documented multi-field invariant and publish a coherent observation. Separate instances remain independent. State exposure follows the supplied ownership contract; avoid a new concurrency framework, lock-free redesign or unrelated public API. |
| Concurrency / `service-owned-workers` | A fixture-supplied host starts workers using an existing lifecycle boundary and a declared admission/capacity limit. Stop admission and all affected work on partial startup failure, worker failure or caller cancellation; observe every started completion before releasing its in-use resource. Gates hold one worker in cleanup and another admission attempt blocked. Preserve promised accepted work and independent outcomes. Primary benefit/repeat case. |
| Concurrency / `cli-pipeline-stop` | A bounded command pipeline stops blocked sends/receives when a producer fails or the consumer finishes early. Join every owned stage; preserve the supplied accepted-output/order contract. Normal early completion and independent failure have explicitly different process outcomes. Observe the actual process and retained output, rather than a helper return alone. Transfer only after draft freeze. |
| Concurrency / `not-concurrency-work` | Repair a private serial calculation with ordinary local tests. Preserve its API; add no goroutine, channel, lock, queue or worker pool. |

Transfer baselines and exposure authors run only after guidance is frozen; their results cannot influence that draft. Historical reader/CSV/borrowed-frame/indexer fixtures are consumed evidence, not new holdouts. Controller demonstrations can validate fixture feasibility but remain excluded from author benefit evidence.

Prepare one additional fresh integrated `request-scoped-fanout` case: an existing service fans out caller-scoped work with bounded admission, collects the specified results/errors and joins all owned work before releasing resources. Its API, dependency construction and host lifecycle are supplied. A caller deadline, held cleanup and blocked admission expose interaction between context and concurrency without requiring a new package architecture. If both individual candidates pass acceptance, use four matched arms: neither candidate, context only, concurrency only, both. If only one qualifies, skip the four-arm study and assess that skill's compatibility with the existing package during delivery review; do not author a rejected neighbor to fill an arm. No old integrated fixture is substituted.

Freeze at most four meaningful semantic mutations per candidate. Targets are unsafe/lost custom-cause handling, lost total budget, skipped join/release order and blocked admission/pipeline termination as applicable. Verify compilation and failure at the intended assertion; unrelated build errors or timing crashes are not detection. Later probes are supplementary and cannot retroactively become primary outcomes.

## Comparison, reuse and budget

Both individual arms receive the exact five-skill 0.2.0 comparator; toggle only the studied candidate. Do not add the context candidate to concurrency's individual comparator. Explicit-selection trials establish exposure effects; native discovery/loading probes establish different claims.

Use fresh independent author contexts. Withhold this design, planning record, expected outcomes, held probes, reviewer repairs and other authors' results. Reviewers receive neutral source identities and original contracts, without guidance, arm labels or author conclusions. Freeze their findings before revealing exposure. Preserve accidental priming, failed attempts and assisted repairs with their actual labels.

| Allocation per candidate | Authors |
| --- | ---: |
| Reference configuration: two discovery cases and one control, baseline and exposure | 6 |
| Fresh transfer pair, after draft freeze | 2 |
| Primary repeat pair with alternate model, same reference harness/reasoning where supported | 2 |
| Primary repeat pair with alternate reasoning on the reference model/harness | 2 |
| Primary pair in each of the other two native target harnesses | Up to 4 |
| **Per candidate ceiling** | **16** |
| **Two candidates plus four integrated authors** | **36** |

Freeze actual available model/reasoning identifiers, harness versions and permissions in the implementation run manifest before execution. A different native harness may require a different model; disclose this confounding and compare exposure within its matched pair. Record toolchain, skill/reference hashes, declared and observed loading, prompts, patches, checks, tokens/time when available and every attempted launch. Unknown settings cannot support controlled-axis claims.

Each attempted author launch counts toward the 36-author ceiling and the applicable allocation, including an infrastructure failure or retry. Permit at most one retry of a failed launch using an unspent slot; record the resulting coverage gap. Do not invent spare trials or suppress failed evidence. Candidate revisions require affected final-byte reruns within the same ceiling; insufficient budget leaves a reviewed draft and a separately reviewable extension proposal.

At each available profile, four selection-only probes offer the candidate for an applicable request and a non-selection request, plus a meaning-preserving paraphrase of each. This is at most 40 probes across two candidates and five profiles; record separately from code-author runs and independent reviews. Verify loading through the actual Claude Code, Codex or OpenCode runtime, using temporary local packaging rather than global installation. A forced file read or structural validator is not native automatic routing evidence.

These profiles sample two models, two reasoning levels and three harnesses, not their full cross-product. Report unavailable profiles, missing capability and correlated settings explicitly. Broader reuse readiness remains unverified until all intended native profiles are exercised; narrower conclusions must identify the supported evidence boundary.

## Verification and acceptance

Use standard-library fixtures unless a case requires an existing dependency. Run meaningful scoped build/test/vet/format checks, race detection on reachable concurrent paths and bounded shuffle/repetition. CLI promises need actual child processes; network promises need faithful clients at the claimed boundary. Test actual Go 1.22 and an available newer supported toolchain when possible; a 1.22 directive under a newer compiler is insufficient. Missing listeners, platforms or toolchains produce scoped limits. Clean race runs do not prove unexercised schedules safe.

Independent outcome review covers Correctness for both candidates; Architecture for changed ownership/API design; Resilience for propagation/budget promises; Performance for actual capacity/resource claims; and Testing for assertion sensitivity. Review relevant topics only and deduplicate common causes. Record confirmed defects, preserved strong outcomes, added API/code/dependency costs and meaningful limits, without averaging grades or rewarding rule repetition.

Promotion requires all of the following:

- A distinct trigger and useful decision procedure that meets the collection's clarity/focus requirements.
- At least two independent matched primary-case pairs correcting the same confirmed relevant defect, plus successful final-byte transfer behavior. Strong alternate baselines may tie, but ties do not supply the repeated corrections.
- Preserved applicable strong/control behavior and no unresolved material regression or demonstrated guidance conflict in tested profiles.
- Exact-guidance and outcome reviews accepting the evidence, with reconstruction/source identities and explicit reuse/toolchain/platform limits.

If discovery finds no relevant weakness, stop before drafting and record the strong baseline. Do not mine more tasks for a win. An alternate discovery case can become the primary only through a declared repeat design before consuming its slots. A reviewed draft can remain useful reference material without a runtime benefit claim.

Integrated results check interaction and compatibility; they cannot replace individual repeated-benefit evidence. Keep assisted fixes separate. Promote only exact accepted snapshot bytes and verify runtime equality. Add accepted skills to package metadata/navigation with a minor update to 0.3.0; if neither promotes, runtime remains 0.2.0. Validate existing manifests/catalogs, local links and available native harness packaging, without claiming that structural success proves loading.

## Delivery sequence and next stage

After written-design review, prepare a task-level implementation plan covering: source audit and exact manifests; context discovery/draft/evaluation/decision; concurrency discovery/draft/evaluation/decision; reuse profiles and selection evidence; integrated trial; exact-byte packaging, independent delivery review and portable evidence sealing. Keep each study's final disposition before starting the next. Preserve one set of raw artifacts with linked comparisons instead of duplicate archives.

Implementation authoring uses skill-creator and writing-skills; verification precedes completion/promotion claims. Independent authors/reviewers must be separately authorized and launched through the execution method chosen at the implementation-plan handoff. This written-design task performs no author trials or delegation.

Update the canonical status and dated log at each meaningful stage with actual checks, artifact identities, decisions, limits and next action. Do not retry the earlier auto-review-blocked composition trial as part of this group.
