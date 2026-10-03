# Context and concurrency recovery

Implemented in `go-quality-build` 0.3.0: [context revision 2](../../plugins/go-quality-build/skills/go-context-and-deadlines/SKILL.md) and [concurrency revision 1](../../plugins/go-quality-build/skills/go-concurrency-and-ownership/SKILL.md), promoted as exact independently accepted bytes.

Approved 2026-10-03: deliver two focused Go authoring skills as practical guard rails. Preserve supported APIs and operation-specific result policies; prefer the smallest mechanism that satisfies the contract. Context owns propagation, time budgets and cancel lifetime. Concurrency owns shared invariants, admission, synchronization and observable stop/join/release. Existing API, composition and testing skills retain their decisions.

The failed study remains recoverable at Git checkpoint `350d06f0d636ca85da998ce66e6f9f97eb6d290e`. Its bulk archive and archive-dependent gate were never in this starting checkout and were not imported. Error/value drafts remain separate deferred work.

## Source decisions

The new prose adapts this repository's historical drafts and writes new conditional rules; it copies no upstream skill text or code. The checkpoint audit's Samber/Uber/Google source pins remain historical, as recorded in the canonical work log. This small revision uses primary Go contracts and existing local review decision references.

| Section | Decision and basis |
| --- | --- |
| Context result policy | Adapt historical cancellation/error decisions; add explicit empty, next-admission and completed-success boundaries. General exposure remains an API contract. |
| Total/stage scopes | Adapt historical budget/body lifetime; verify against [context](https://pkg.go.dev/context) and [Client.Do](https://pkg.go.dev/net/http#Client.Do). Revision 2 distinguishes dependency-owned redirect error closure. |
| Required finalization | Adapt narrowly: [WithoutCancel](https://pkg.go.dev/context#WithoutCancel) requires Go 1.21+, a fresh bound and cooperative completion. Ordinary work retains caller cancellation. |
| Shared invariants and aliases | Adapt historical lock/handoff guidance against [the memory model](https://go.dev/ref/mem) and [sync](https://pkg.go.dev/sync). Coherent observation and mutation ownership govern the mechanism. |
| Admission and lifetime | Adapt sparse-input progress, capacity through release, supervision before held cleanup and per-resource versus shared-cohort join rules. Check actual operation contracts rather than mandate a scheduler. |
| Channel and result policy | Adapt [pipeline stopping](https://go.dev/blog/pipelines); early consumer success still joins production, and broad cancellation classification cannot identify independent error origin. |
| Blanket rules/machinery | Omit unconditional context checks, mandatory interfaces/dependencies/frameworks, automatic panic recovery, forced concurrency, and cancellation-class-based error erasure. |

## Bounded evidence

[The replay manifest](../../tests/go-quality-build/recovery/trials.json) preserves fixed neighboring guidance, input/output identities and source/test patches. [The checker](../../tests/go-quality-build/recovery/check.py) reconstructs disposable modules and runs authored tests/vet, controller checks on Go 1.22.12 and 1.26.5, and applicable current race/shuffle checks. No duplicate source trees, transcripts or cross-model matrix are maintained.

| Case | Observation |
| --- | --- |
| Context HTTP development | Baseline and revision 1 pass frozen checks. The consumed historical input is development evidence, not an unseen transfer. |
| Accepted-finalization transfer | Both arms pass frozen checks, preserving accepted count, required finalization and independent errors. The failed-callback cause difference is ambiguous under the original contract; no baseline defect is claimed. |
| Redirect ownership diagnostic | Added after inspection. Both HTTP outputs double-close and fabricate a repeated-close error on both compilers. A fresh revision-2 author fixes it and passes all frozen/diagnostic/minimum/current/race checks. This is targeted development correction, not prespecified transfer uplift. |
| Calculation/routing control | Context declined for the simple calculation and selected for both lifetime/budget contract cards. Simple implementation passes actual minimum/current checks. |
| Concurrency development/transfer | Both baseline and guided arms pass six frozen replay checks each on minimum/current/race. These are preservation ties. |
| Acquisition-origin diagnostic | Added after baseline inspection. Independent Open failure established before peer stop is erased by baseline, retained by unchanged guided code on both compilers. This is one narrow observed difference, not prespecified/causal uplift. |
| Concurrency example/control | Exact lock example passes minimum/current race coherence checks. Serial control declines both new skills and selects them for both lifecycle cards; five current/minimum replay checks pass. |

The [focused concurrency review](../../tests/go-quality-build/recovery/concurrency-review.md) accepts distinct useful mechanisms, preserved contracts and the narrow origin correction with limits. The [focused context review](../../tests/go-quality-build/recovery/context-review.md) separates observed preservation from comparative effectiveness and identifies the narrow ownership correction. Revision-1 guidance can be recovered by reversing [the revision-2 delta](../../tests/go-quality-build/recovery/patches/context-guidance-r2.patch); its original SHA-256 is `50ac9186e3132c3771bde754cf8f27dd0135f6c23d7cc1b8a84afb9cc6578eb5`.

Initial HTTP replay failures were loopback sandbox denials; unchanged authorized runs pass. These setup failures are separate from code outcomes. Go 1.22.12 was downloaded only to temporary task storage.

The context transfer received revision 1; revision 2 changes only HTTP ownership, with finalization policy/example unchanged. Shared-state/alias authoring and the pipeline's precise late-cancellation completion boundary were not exercised by the lifecycle comparisons. The scalar lock example has a separate minimum/current race coherence check. Live-body ownership transfer and `AfterFunc` joining remain inspected guidance without direct behavior cases.

Limits: one author per comparison arm, inherited model/reasoning IDs unavailable, unblinded focused reviews, development feedback and labelled post-inspection probes, cooperative finite examples and macOS arm64. Guided implementation exposure is explicit; contract-card selection is not automatic harness routing. No universal effectiveness, every schedule, arbitrary I/O interruption, remote termination or Codex/OpenCode runtime loading claim.

[Final package checks](../../tests/go-quality-build/recovery/final-checks.json) pass 29 unit tests, 120 review/44 build fixture validation and all three Claude validators. [Independent package review](../../tests/go-quality-build/recovery/package-review.md) accepts local delivery. Runtime hashes, unchanged neighbors, reverse revision recovery, maintained links and whitespace pass; [the canonical work log](README.md) records exact boundaries. Future work should start from a concrete observed weakness; these comparisons do not justify expanding a matrix.

[PR #7 follow-up](pr-7-copilot-review.md) corrects the current HTTP fixture's ownership wording and makes replay enforce per-trial guidance/revision identities. Recorded HTTP trials restore the original historical README before checking unchanged input/output hashes. No new model comparison is claimed. All 11 replays (65 checks) and 35 repository unit tests pass after these fixes.
