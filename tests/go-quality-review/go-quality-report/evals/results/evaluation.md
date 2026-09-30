# Go quality report evaluation

## Scope

The new umbrella coordinates the nine existing Go review skills. This evaluation checks orchestration decisions, evidence/coverage handling, root-cause reconciliation, grade arithmetic and an actual delegated code-area review. The nine topic skills and their case definitions were left unchanged.

## Checks completed

- Bundled skill-creator validation passes for the umbrella.
- The collection validator passes ten skills and 120 case definitions (99 existing topic cases, 21 umbrella cases).
- All 99 prior topic report outputs still pass the existing structural/report validator.
- Twelve calculator tests pass, covering every grade anchor, systemic/critical reach requirements, partition invariance, independent safeguards, incomplete coverage, duplicate IDs, not-applicable contradictions and malformed input. Tests initially failed before the calculator existed; a later malformed-type check reproduced TypeError and was fixed to return a useful validation error.
- Validator mutation checks confirm that umbrella rubric drift, omitted topic routing and escaping local references are rejected.
- An independent implementation review found no substantive correctness issue. Its two optional points were addressed: topic-only report validation is now explicit, and malformed ledger types receive proper errors.

## Blind behavioral evaluation

Twenty isolated planning/synthesis prompts were evaluated with the new skill and references, with expectations withheld. A separate reviewer read their outputs against the assertions: **55/55 assertions passed**, with no expected-output mismatches. Cases cover ambiguous versus explicit scope, large versus small work, no applicability, missing skills/workers, worker timeouts, local-ID collisions, independent testing gaps, duplicate causes, partition invariance, similar-but-independent causes, disputed severity, A+ safeguard deduplication, partial findings, context files, boundary handoffs, model policy, stale cards and constrained budgets.

Before authoring the umbrella, one baseline evaluator used the existing topic contract on three synthesis scenarios. It correctly handled duplicates, incomplete coverage and partition invariance; it identified that overall severity aggregation was an interpretation not explicitly defined by the existing contract. This baseline does not demonstrate that the new skill improves those already-correct decisions. The new deliverable adds explicit orchestration, applicability, artifacts, reconciliation and aggregation semantics.

## Actual delegated review

The twenty-first case reviewed the complete two-package catalog/export fixture using **two actual worker agents**, ten standard topic cards and a focused boundary check. All nine topics were considered: six applicable topics completed and three were justified Not applicable. The original tests pass while direct element mutation and export-preservation reproductions fail; a producer-only correction in a disposable copy makes both pass. Original fixture hashes remained unchanged.

The final report reconciles four production source findings into one ownership cause, and keeps one independent testing gap. Two unique major findings give **C-**, matching the calculator. Full cards, manifest, evidence, source-ID mapping, handoffs and grade ledger are saved under [live/](live/report.md). Root reviewed the six live-case assertions against those artifacts: **6/6 passed**, for **61/61 across 21 cases**.

The workflow encountered two recoverable execution problems: sandbox denial of the default Go cache, resolved by scratch GOCACHE, and array-valued evidence in a string-only calculator field. The latter was caught by validation and normalized without changing findings; root also called out the documented schema mismatch. The grading reference now explicitly instructs conversion from index evidence arrays to ledger strings. This was a real orchestration run with this small schema-format intervention, not a claim of a completely unaided run.

## Evidence files

`baseline-reports.jsonl` retains those three initial outputs. `revised-reports.jsonl` retains the twenty blind outputs. `judgments.json` records individual assertion judgments, including the live case. `summary.json` records the combined results. `calculation-artifacts.json` retains the synthesis ledgers and calculator results. `validator-checks.json` and `independent-review.md` retain structural checks and review notes.

## Limits

The scenarios are synthetic and most are supplied-fact exercises. Evaluators are from one model family; cases were batched in contexts, and the study does not measure multi-model or repeated-run defect-discovery variance. Calculator success establishes arithmetic for an already reconciled ledger, not evidence truth, omitted material coverage or semantic identity of causes/safeguards. Model selection guidance is provider-neutral and capability-based; it has not been exercised on other vendors' hosts. Real changesets with human-reviewed outcomes should extend this suite over time.
