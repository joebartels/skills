# Go language contracts implementation plan

> **For agentic workers:** Use `superpowers:executing-plans` for authoring in this session and `superpowers:writing-skills` for fresh-context evaluation agents. Steps use checkboxes to track verified progress.

**Goal:** Build and evaluate distinct error-contract, value-semantics, and names/documentation skills, promoting only candidates with enough useful content and supported benefit.

**Architecture:** Extend the existing portable build plugin sequentially. Keep candidate drafts and all evaluation inputs/results outside runtime; hold the architecture skills fixed in matched trials. Independent agents assess likely usefulness and actual code outcomes.

**Tech stack:** Markdown/YAML skills, Python repository validator and archive checks, standard-library Go fixtures declaring Go 1.22, available installed Go toolchain, native multi-agent evaluations.

**Spec:** [Approved language-contract design](../specs/2026-10-01-go-quality-build-language-contracts-design.md). User approved proceeding on 2026-10-01, conditioned on enough distinct content and independent expert feedback.

## Global constraints

- Retain upstream `samber/cc-skills-golang@19a0626ae8565d27a7b7bdf59d8d99d94d7e284c` and audit retained material before drafting.
- Do not edit the independent review skills or testing group's runtime/fixture files.
- Runtime promotion remains sequential. For the reconciled non-installable draft deliverable, independent error/value authoring and trials may proceed concurrently; preserve distinct statuses and claims.
- Use identical fixture hashes, model/effort, and existing-skill exposure within each comparison; disclose batching and repeats.
- No global installation, publication, new coordinator, or required third-party Go dependency.
- Keep explicit exposure separate from automatic routing and actual minimum-toolchain/platform verification.

## Review focus

- Success errors can contain typed nils; the reader fixture checks nil-interface success and bytes-with-error behavior.
- Completion can fail after a primary error; the CLI fixture checks both inspectable causes and truthful process status.
- Capacity-limited slices still alias existing elements; the ownership fixture mutates elements and nested values.
- Required construction differs from optional configuration; a constructor-required control must retain its precondition.
- Documentation can silently promise stronger safety or compatibility; executable examples and independent fact-checking cover this.

## Files and evaluation interface

- Authoring records: `docs/go-quality-build/language-contracts-source-audit.md`, `language-contracts-design-review.md`, and canonical `README.md` status/work log.
- Candidate suites: `tests/go-quality-build/<skill>/evals/evals.json`, with runnable `files/<case>/go.mod`, Go input files, and contract README.
- Draft/runtime: `tests/go-quality-build/<skill>/draft/SKILL.md`; after promotion `plugins/go-quality-build/skills/<skill>/SKILL.md`, with at most a focused reference justified by evaluations.
- Archive: `tests/go-quality-build/results/2026-10-01-language-contracts/<skill>/<arm>/<case>/`, containing `prompt.txt`, `source.patch`, `trial-report.md`, output hashes, and actual verification. Summaries preserve first-pass and revised runs separately.
- Combined suite: add `language-contracts` case to existing `tests/go-quality-build/combined/evals/evals.json`, with inputs in `files/language-contracts/`.
- Packaging: root/plugin READMEs, both plugin manifests, and Claude marketplace description. Catalog paths remain shared.

Evaluation agents consume a disposable copy of one case, the task prompt, fixed architecture-skill catalog, and (skill-on only) the candidate description/path. They select applicable skills themselves and record opened files. They may implement only inside their assigned trial directory. They do not see expected outcomes, checks, the other arm, or reviews. Root verifies artifacts; an independent reviewer assesses output without arm labels initially where practical. Root maintains the canonical record between stages.

### Task 1: Expert assessment, source audit, and error baseline

- [x] Obtain an independent, evidence-backed assessment of each candidate's distinct content and likely effectiveness; resolve important concerns and record decisions.
- [x] Audit main files/references from error-handling, code-style, structs/interfaces, safety, naming, and documentation. Inventory second-level sections, choose copy/adapt/omit, and connect retained guidance to local owners and primary docs. Use original examples; add a license notice only if substantial source content is copied.
- [x] Create four error cases: `reader-errors` (caller input, partial bytes, typed nil), `backend-errors` (domain translation and sensitive diagnostics), `cli-completion` (primary/finish failures, accepted prefix, status), `not-error-work` (private calculation control). Each starts with buildable existing behavior and a concrete requested change.
- [x] Before authoring a skill, run fresh baseline implementations and independent code-quality/correctness assessment. Run private behavioral probes after author output is archived. Record confirmed weakness or select a materially different task if none appears.
- [x] Verify fixtures with `rtk go test ./...` and `rtk go vet ./...` in each module; verify suite structure with `rtk proxy python3 -B scripts/validate.py`. Commit inputs, audit, baseline artifacts, and progress.

### Task 2: Error skill evaluation and promotion decision

- [x] Write minimal `go-error-contracts` draft addressing demonstrated error decisions, with trigger/exclusion, concise decision table, and one useful example. References contain only conditional detail.
- [x] Run matched fresh skill-on cases, including the positive exposure case and non-selection control, against final bytes. Verify probes, build/tests/vet, opened skills, unnecessary changes, and output hashes.
- [x] Have an independent reviewer assess changed code and evidence using relevant existing review skills. Frozen blind review plus separate reader adjudication confirms one observed baseline cause-loss issue; no harmful/incomplete draft guidance requires revision.
- [x] Resolve promotion: retain the reviewed draft outside runtime. One post-draft pair shows an error improvement, but repeated causal benefit remains unproven. Frontmatter, links and reproducibility pass; stage commits and canonical entries preserve the authoring/promotion distinction.

### Task 3: Value skill baseline, evaluation, and promotion decision

- [x] Create four value cases: `default-wire` (zero/configured/explicit-zero state and intentional nil/empty output), `snapshot-ownership` (nested aliasing, receivers/method sets, copy-sensitive state), `required-construction` (counterexample retaining validated inputs), `not-value-work` (private calculation).
- [x] Run baseline implementations before draft; preserve independent findings and behavioral probes.
- [x] Author `go-values-and-zero-values` from independently reviewed ownership decisions.
- [x] Run matched final-revision skill-on cases and independent Code Quality/Correctness assessment; no owned contradiction or harmful value outcome is confirmed.
- [x] Resolve promotion: retain the independently reviewed value draft outside runtime. All five pairs tie on assessed code grades; sufficient distinct content supports drafting, while measured benefit remains unproven. Checks, hashes and stage commits are preserved.

### Task 4: Names/docs disposition

- [x] Independently assess distinct content and overlap; choose a focused non-runtime reference over a standalone skill.
- [x] Author and obtain exact-byte review of `docs/go-quality-build/authoring-references/names-and-docs.md`.
- Not run: the originally proposed library/CLI/docs-only/formatting skill baselines and skill-on trials. No names/docs standalone draft, promotion or behavioral effectiveness result is claimed.

### Task 5: Combined use, package verification, and final review

- [x] Create a fresh library/CLI evolution case covering errors, defaults/snapshots and accurate usage; run fixed-existing-skill and existing-plus-draft arms. The draft author read the canonical record; preserve this context qualification rather than calling all author contexts isolated.
- [x] Independently assess code and cross-skill contradictions, preserving frozen blind output separately from provenance interpretation. Both combined candidates grade A/A; no code contradiction is confirmed. This single context-qualified pair shows compatible use, not synergy.
- [x] Resolve packaging: no new runtime promotion, so root/plugin catalogs and versions remain at the three existing architecture skills. The two drafts and names/docs reference are documented outside runtime.
- [x] Run root/build-eval/review-layout/grade tests, repository validation, Claude package/marketplace validation and both draft frontmatter validators. Verify local links, hashes and all 27 reproducible patches; preserve raw patch context whitespace as evidence.
- [x] Obtain independent whole-change/exact-byte review and final arm-blinded/provenance-aware effectiveness review. Correct combined provenance, independently adjudicate the reader disagreement, and record limits and continuation actions. No draft content changes are required after final-byte exposure.
- [ ] Commit the final verified work and deliver concise per-skill results plus independent effectiveness feedback. Do not install or publish.

## Execution decisions

Ruling: the user's “then go for it” authorizes implementation after the reviewed spec, including the implementation mechanics recorded here. Author inline and use independent agents for assessment/evaluations as explicitly requested; do not add another approval round for this plan. Matched cases may run concurrently because their trial directories are independent. Keep runtime promotion sequential; the later draft-authoring reconciliation permits independent draft exposure trials concurrently. Existing managed worktree isolation is reused.

Initial error decision (2026-10-01; draft deferral superseded below): Task 2 draft/skill-on/promotion steps are intentionally
skipped because independent baseline review found no confirmed owned defect after
the bounded alternative. Deferred is a completed candidate decision, not an
implemented/evaluated runtime skill. Preserve the unused CSV transfer. Task 3 may
proceed without error runtime changes.

Initial value decision (2026-10-01; draft deferral superseded below): four fresh baselines and independent applicable-case
review passed without a confirmed owned defect. Under the current gate, draft and
skill-on steps are deferred. The optional user preference question can revise the
deliverable standard; an unanswered question is not authorization to change it.

Draft-authoring reconciliation (2026-10-01): the human's content-sufficiency and
expert-feedback conditions are met for errors and values. The stricter pre-draft
failure gate was a workflow choice, not a human prohibition. Create and review
non-installable drafts under existing authorization; no measured benefit or
runtime promotion is implied. Names/docs receives a focused non-runtime reference
and standalone deferral; Task 4 behavioral skill steps are intentionally skipped.
Final checks concern draft quality and compatibility, not a manufactured uplift.

Final disposition (2026-10-01): two independently reviewed, explicitly exposed,
behavior-checked skill drafts; names/docs as reviewed reference; zero new runtime
promotions. The expert feedback satisfies the user's review request. There are
27 authored trials, 12 paired tasks (11 clean pairs plus one context-qualified
combined pair), and one post-draft observed reader improvement. A future promotion
study should declare repeated realistic tasks/outcomes in advance; no such repeat
study or broad benefit is claimed by this delivery.
