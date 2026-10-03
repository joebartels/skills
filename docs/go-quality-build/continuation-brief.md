# Go skills continuation brief

Current request received 2026-10-03: recover and deliver useful context/deadline and concurrency/ownership guidance using the supplied failed-study handoff. The approved recovery is delivered locally as seven-skill package 0.3.0; new runtime revisions are context 2 and concurrency 1. Fresh packaging checks and independent review pass. See the [canonical status and recovery log](README.md) and [compact evidence](context-concurrency-recovery.md). The sections below preserve the earlier orientation checkpoint; they are not the latest task authorization or design state.

Orientation completed 2026-10-02 at detached commit `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96` in the existing managed linked worktree. The user requested familiarization before compaction and will supply the next group afterward. No new group has been selected or authorized for implementation in this orientation.

The [design record](README.md) is canonical. Read its current status and the [review guide](../go-quality-review/README.md) before continuing, inspect Git state, and update status plus a dated work-log entry after meaningful stages and before ending. Prefix shell commands with `rtk` per `/Users/jb/.codex/RTK.md`.

## Orientation checkpoint artifacts and disposition

| Artifact | Location | Disposition |
| --- | --- | --- |
| Build collection | `plugins/go-quality-build/` | Version 0.2.0; five evaluated runtime skills. |
| Review collection | `plugins/go-quality-review/` | Version 0.1.1; nine topic skills plus `go-quality-report`. |
| Error contracts | [draft](../../tests/go-quality-build/go-error-contracts/draft/SKILL.md) | Reviewed, explicitly evaluated draft; outside runtime. |
| Values and zero values | [draft](../../tests/go-quality-build/go-values-and-zero-values/draft/SKILL.md) | Reviewed, explicitly evaluated draft and mechanics reference; outside runtime. |
| Names/docs | [authoring reference](authoring-references/names-and-docs.md) | Reference, not a standalone evaluated skill. |
| Context/deadlines and concurrency/ownership | Candidate table in the design record | Proposed; no runtime skill or evaluation suite exists. |

The five build skills have distinct owners:

- `go-package-boundaries`: responsibility placement and import direction, across libraries, CLIs, services and workers; layouts are examples.
- `go-api-contracts`: consumer-visible shape, compatibility and deliberate release migration.
- `go-interfaces-and-composition`: necessary abstractions, construction, dependency wiring and resource lifetime.
- `go-behavior-tests`: cases, independent assertions and the boundary proving a promise; current revision 2.
- `go-test-isolation`: dependency fidelity, fixture lifetime, process state and controlled asynchronous observations; current revision 3.

Local history contains architecture PR #3, language-contract PR #4 (`309f4c9`) and the testing PR #5 squash commit (`0d0a339`). The previous top continuation note still described PR #5 as open; orientation reconciles that with local history. This is local Git evidence, not a fresh remote CI or review query.

The latest testing closeout proposed an error/context/concurrency group, but that proposal predates the user's next request. Reuse existing error/value work rather than treating it as missing. [Effectiveness feedback](language-contracts-effectiveness-review.md) recommends retaining both exact drafts outside runtime: one error reader pair improves under a post-draft contract adjudication; value comparisons tie at A. Content sufficiency justified draft delivery, while repeated incremental benefit and runtime promotion remain unestablished. Names/docs has no separately evaluated contribution.

## Authoring and evaluation conventions

Canonical runtime files live at `plugins/<collection>/skills/<skill-id>/SKILL.md`; trigger-focused frontmatter names match directories. Supporting runtime links must stay inside that skill directory. Fixtures, probes, audits, reports and archived guidance snapshots stay outside the installable package. Portable and Claude manifest identities/versions match; repository marketplaces and OpenCode point to the same runtime tree. See the [packaging design](../superpowers/specs/2026-09-29-multiharness-skills-design.md).

For the next requested group, read the applicable brainstorming, planning, skill-creator, writing-skills and verification skills at execution time. Settle bounded decision ownership against existing skills and review criteria. Existing specs/plans under `docs/superpowers/` describe the completed architecture, testing and language-contract efforts; they are not approval for a new group.

Upstream reuse is pinned to `samber/cc-skills-golang` revision `19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`. Audit entire relevant skills/references section by section as copy/adapt/omit, assigning local owners and checking current primary Go documentation for version-sensitive advice. Preserve the MIT notice for substantial copying. Existing runtime wording/examples are original adaptations; see [architecture](source-audit.md), [testing](testing-source-audit.md) and [language-contract](language-contracts-source-audit.md) audits.

Separate useful draft creation from evidence-based promotion. Preserve successful baselines; predeclare repeated realistic tasks, outcomes and search budget rather than searching indefinitely for wins. Use fresh independent author/reviewer contexts, fixed neighboring-skill exposure, relevance/non-selection controls, and withheld author expectations/probes. Freeze exact guidance/input/prompt/model/harness provenance; disclose unavailable metadata or compromised blinding. Inspect completed code with the independent review collection, meaningful builds/tests and compiling mutations that fail the intended assertion. Reevaluate affected single/control/repeat and combined cases after guidance changes. Keep assisted repairs distinct from blind skill outcomes.

Immutable evidence under `tests/go-quality-build/results/` retains original failures, raw reports, patches and seals. Reconstruct disposable copies for replays. Historical helper scripts have old workspace/revision assumptions: for example, `testing-combined-final/controller/verify_testing_evidence.py` compares isolation revision 2, while current runtime is revision 3. Inspect a helper before reuse; its historical report is not a fresh check. Raw-evidence whitespace exclusions are documented in [final whitespace closeout](reviews/testing-delivery/final-whitespace.json).

## Fresh orientation verification

These commands passed on 2026-10-02:

```sh
rtk proxy python3 -B scripts/validate.py
rtk proxy python3 -B -m unittest discover -s tests -p 'test*.py'
rtk proxy python3 -B -m unittest discover -s tests/go-quality-build -p 'test_*.py'
rtk proxy python3 -B tests/go-quality-review/test_layout.py
rtk proxy python3 -B tests/go-quality-review/go-quality-report/evals/test_grade.py
rtk proxy claude plugin validate ./plugins/go-quality-build
rtk proxy claude plugin validate ./plugins/go-quality-review
rtk proxy claude plugin validate .
```

Results: 29 unit tests (6 root, 10 build evaluations, 1 review layout, 12 grade calculator), 120 valid review cases, 38 valid build cases, and all three Claude validations. Build-case totals include draft suites: package 5, API 4, composition 5, behavior 5, isolation 5, errors 6, values 5, combined 2 and testing-combined 1. Structural checks do not establish behavioral effectiveness or automatic routing. CI currently runs validator/root/layout/grade checks; the local checklist additionally runs the build-evaluation unit suite.

Fresh byte comparisons confirm both current testing skills and their references match accepted final snapshots under `2026-10-01-behavior-tests-r2/skills/` and `2026-10-01-test-isolation-r3/skills/`. Error/value drafts and names/docs SHA-256 values match the exact artifacts listed in the effectiveness review. The historical 49 testing reconstructions and 2,423 sealed hashes were not replayed during orientation; their previous verified results are preserved in [final integrity](reviews/testing-delivery/final-integrity.json).

Open limits remain actual Go 1.22 execution, wider platforms/integrations/schedules, automatic routing, Codex/OpenCode runtime loading, inherited model reproduction and surviving supplementary custom-cause test mutations. Historical Go execution used 1.26.5 darwin/arm64 against Go 1.22 declarations. The older zero-value composition trial remains blocked/incomplete; do not retry it without the explicit user authorization required by its recorded approval rejection.

Both continuation documents pass 172 local-link target checks and whitespace checks; `rtk git diff --check` also passes. Only this brief and the canonical record changed, with no runtime or evidence edits.

The user subsequently asked which next group best extends targeted design coverage. Proposed scope: new `go-context-and-deadlines` and `go-concurrency-and-ownership` candidates, plus further bounded evaluation of the existing error-contract draft. Context owns propagation/budgets; concurrency owns synchronization and stop/join implementation; errors owns failure representation and partial/completion outcomes. Composition retains wiring/lifecycle API choices. Data/trust boundaries are recommended afterward. The proposal is not an approved spec, implementation or promotion.

Next action: after compaction, read this brief and current canonical status, then settle the selected group's decision ownership and evaluation design. No runtime content, evaluation evidence, installation or publication changed during orientation or recommendation research.
