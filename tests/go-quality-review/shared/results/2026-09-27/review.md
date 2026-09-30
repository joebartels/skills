# Go review skills audit — 2026-09-27

## Scope and changes

Reviewed all nine skills for standalone use across libraries, CLIs and services, shared grading, evidence quality, topic coverage and eventual composition. No umbrella skill was created or installed.

- Embedded one shared scope, calibration, grade table and report template in every skill. Each skill remains self-contained when its directory is copied; the canonical authoring contract is not a runtime dependency.
- Added explicit Coverage and Finding counts fields, finding IDs linking defects to corrections, evidence requirements for strengths, and coherent Not applicable / Insufficient evidence outputs.
- Clarified diff versus code-area review, severity reach, root-cause grouping, A+ safeguards, solo ownership and shared finding deduplication. Topic grades remain independent; they must not be averaged or their counts summed without deduplication.
- Removed external authoring-resource dependencies from three skills. Kept contextual Go guidance rather than mandatory frameworks, architectural layers, test shapes or service-only controls.
- Expanded decision guidance for library process ownership, typed-nil errors, queue acknowledgement, worker-goroutine test failures, authenticated encryption, cache byte bounds, checked-versus-shipped commits, iteration completion errors and actual vendor/build modes.
- Corrected module diagnostic advice: readonly module mode bypasses vendoring; it does not make every Go diagnostic non-mutating. Checks that can change files use disposable copies.
- Added ten behavioral cases, including one runnable typed-nil fixture and a paired test-severity scenario. The current suite contains 99 cases.
- Added a validator for packaging, portable references, shared-contract drift, evaluation inputs and report structure/count arithmetic. Seven acceptance/rejection checks exercised the validator itself.

## Evaluation method

The original 89 cases were run using the original skills. Expected outcomes and assertions were withheld from evaluators; saved outputs were judged separately. After revision, all 98 then-current cases were run and independently judged again. Agents worked in topic batches; contexts were not fresh per scenario. Groups were rotated where practical, but this is one model family, not a multi-model or human-rated study.

These are behavioral agent evaluations, not Python unit tests. PyYAML enables packaging validation. Most cases supply a complete scenario as prose; runnable fixtures separately check returned-slice aliasing, a hidden workspace dependency and a typed-nil error. Fixture commands ran in disposable copies. The original catalog test passes while an element-mutation reproduction fails; the workspace build passes while the standalone build fails; the typed-nil valid-input test fails as intended.

| Run | Cases | Literal assertions passed |
| --- | ---: | ---: |
| Original skills | 89 | 280 / 284 |
| Revised full run | 98 | 309 / 312 |

These counts use the expectations present during each judgment. They are not directly comparable success rates because the suite and several unsupported expectations changed. Expected prose was also reviewed separately, revealing omissions not captured by individual assertions. Full reports and judgments are retained alongside this document.

## Findings from evaluation

Original-run problems included missing behavioral verification after a dependency security fix, inconsistent reporting fields, and three expectations that exceeded the supplied evidence: a contained grade for reachable arbitrary shell execution, major severity from shared-test contamination risk alone, and mandatory discussion of interface generic methods in a concrete-method-only case. Their corrections are documented by the retained original judgments rather than silently replacing them.

The revised full run found an encryption recommendation that omitted explicit stored-format compatibility, a panic-recovery review that underexplained the changed client response, and a legacy-loop grade expectation that assumed an important behavior without evidence. The first two led to targeted guidance changes. The legacy-loop expectation now requires B for the supplied localized gap; a new paired authorization scenario establishes the important contract needed for C. A fresh test review also inflated a setup failure from moderate to major without establishing importance, prompting explicit Testing severity guidance.

Separate semantic review found an omitted HTTP/1.x reuse explanation and contradictory revision-ID boilerplate. These are recorded even though the listed assertions and mechanical report checks did not catch them.

## Follow-up results

Eight targeted reports cover the four Testing boundary cases and the AEAD, HTTP panic, shipped-commit and HTTP body cases. Combining these with unchanged reports from the full revised run yields **99 cases and 316/316 listed assertions satisfied**. This is a composite result, not a fresh 99-case rerun after the final edits. Full expected prose matches 98/99: the HTTP preview report still omits an explicit explanation that EOF can aid HTTP/1.x reuse. Its A grade, closure advice and rejection of unbounded draining are correct; the decision reference contains the mechanism. This small explanation omission remains recorded rather than relaxing the expectation.

Five fresh evaluator contexts then applied 14 fixed severity inventories and four contextual Testing cases each. All 70 inventory grades and 20 contextual grades agreed. The three localized Testing defects received B, and the sole ineffective tenant-authorization verification received C. An earlier sample, retained separately, assigned C to an ordinary setup failure before the severity clarification. These bounded repetitions test calibration, not discovery of arbitrary defects.

A same-context solo/combined check found one shared buffer-ownership defect. Performance and Correctness each retained C in both modes; the combined report shared F1, named one remediation owner and counted one unique major cause. This is a composition smoke test, not a validated umbrella aggregation policy.

All nine skills pass the bundled skill-creator validator. The shared validator passes all 99 current case definitions and all 99 composite report structures, severity counts, grade arithmetic and finding-to-change links. These mechanical checks supplement the saved semantic judgments.

## Saved evidence

- `baseline-reports.jsonl` and `baseline-judgments.json`: original 89-case run.
- `revised-reports.jsonl` and `revised-judgments.json`: full 98-case revised run.
- `follow-up-reports.jsonl` and `follow-up-judgments.json`: eight targeted replacements/additions; join by skill and case ID to reconstruct the final composite.
- `stability-results.json`: six calibration samples, including the earlier drift sample and five post-clarification repetitions.
- `composition-reports.json`: solo and combined outputs.
- `final-summary.json`, `skill-manifest.json`, `validator-checks.json` and `baseline-structural-findings.json`: machine-readable summary, evaluated-file hashes and structural evidence.

## Verification limits

Passing examples do not establish exhaustive topic coverage or guarantee stable grades on arbitrary repositories. Severity still depends on supplied or traced impact, and root-cause grouping requires judgment. The validator deliberately cannot certify factual findings, critical/systemic reach, or independence of A+ safeguards. Future calibration should add anonymized real changesets and independent human judgments, especially near severity boundaries.
