# Reconciliation and overall grading

## Establish the evidence before counting

1. Validate each card's target snapshot, scope, topic skill use and causal evidence. Trace its trigger and actual callers, guards, existing rule owners and meaningful test consumers. Record counterevidence and the work removed by the correction after preserving necessary responsibilities; a label or site count is not this analysis. Source tracing can suffice without reproduction. Unsupported concerns remain ungraded, not discounted to minor; missing material evidence remains a coverage gap. Admit supported causes into `findings`; retain consequential rejected, merged and unresolved candidates in `decisions`.
2. Namespace each local ID as `chunk/topic/local-id`; local F1 is not globally unique. Build one canonical finding per independently actionable cause, retaining all source-card IDs and evidence. Merge repeated symptoms when the same causal correction resolves them; shared wording, location or severity alone is insufficient. A production fix and a test change remain distinct only when independently necessary.
3. For each unique cause, use its greatest **supported** consequence across topics for the overall severity. A critical label requires supported severe/broad/irreversible reach. A systemic major requires demonstrated multiple distinct important boundaries/workflows, not repeated filenames or many agent mentions. Ask the responsible worker to substantiate a disputed classification; do not average disagreement or blindly accept the largest label. If an unresolved dispute could change the grade, keep that material uncertainty in coverage/limits rather than presenting a settled overall grade.
4. Preserve original cards and record accepted corrections, canonical IDs, owners and topic-specific consequences. Reconcile maintenance overlap by consequence: local repeated implementation belongs to Code Quality; unnecessary API/dependency plumbing belongs to Architecture. A second label for the same cost does not justify another topic deduction. Distinct supported effects can still affect both topics with different severities. Recompute rollups from unique in-topic causes, not the mean or worst chunk grade.
5. Deduplicate positive safeguards too. Two descriptions of one cancellation mechanism or one controlled risk do not establish two independent safeguards. A+ needs two verified nonroutine controls of distinct meaningful risks and complete material coverage, not two A+ cards or a minimum number of tools.

## Scope and grade

For a full requested scope, all material obligations must be complete or justified Not applicable. Partial, pending, unavailable or outcome-changing uncertainty means overall **Insufficient evidence**. Still report confirmed defects; an optional **Assessed-scope defect grade** describes their deduplicated counts only and must not be presented as the full result. No confirmed defects plus incomplete coverage does not warrant a partial A. If no topic is applicable, use **Not applicable**.

With complete coverage, apply the first matching row below to unique confirmed causes using each topic's severity definitions. An important supported-contract failure is normally major even if its fix is small. Maintenance findings require a specific reading/change task and a correction preserving required responsibilities. Establish the apparently redundant work's purpose or the decision requiring synchronized edits; similar calls or values alone are insufficient. Local friction or a small repeated edit is minor; moderate requires a concrete failure mechanism or substantial coordination/reimplementation cost beyond site counts. Unknown implementation roles and hypothetical capabilities remain ungraded.

Strengths do not cancel defects. Independent causes accumulate across the requested scope regardless of packet boundaries, reviewer count or codebase size. Do not normalize by lines, average grades, or use confidence or remediation size as severity.

| Grade | Anchor |
| --- | --- |
| F | At least one critical issue. |
| C- | Two or more independent major issues, or one major plus two or more moderate issues, or one systemic major issue. |
| C | One contained major issue, with at most one moderate issue. |
| C+ | No major or critical issue; two or more independent moderate issues. |
| B- | No major or critical issue; one moderate issue plus two or more minor issues. |
| B | No major or critical issue; one moderate issue with at most one minor issue. |
| B+ | No moderate or worse issue; two or more minor issues. |
| A- | Exactly one minor issue and no moderate or worse issue. |
| A+ | Meets A, with two independent, verified safeguards controlling distinct meaningful risks beyond routine correct setup. |
| A | No actionable issue; at least one relevant strength is verified and material applicable risks are assessed. |

A requires a relevant verified strength and assessment of all material risks. The two A+ safeguards can occur anywhere relevant within that completed scope; every topic need not itself earn A+. A grade is a quality assessment, not an automatic merge authorization.

## Record and calculator

The [review record](review-record.md) is the maintained authoring format. `render.py` projects its canonical causes, per-topic consequences, coverage, strengths and safeguards into ledgers and calls `grade.calculate`. Unresolved decisions create coverage gaps in every affected topic. The renderer derives overall severity from the greatest supported topic consequence and counts each canonical cause once; it cannot prove the semantic reconciliation.

The existing `grade.py /path/to/ledger.json` CLI remains available for arithmetic-only consumers. A ledger has `scope`, `coverage` (id/status/reason), `findings` (id/severity/cause/evidence/correction; optional systemic and required reach for critical/systemic major), `strengths` (id/evidence), and `safeguards` (id/risk/evidence/nonroutine). Evidence here is text, unlike record evidence arrays. Coverage statuses and severity anchors are unchanged. New reviews should author the richer record and let the renderer build these projections rather than maintain a second ledger by hand.

## Separate maintainability assessment

A/A+ describes the completed defect review under these anchors, not a universal claim of elegance or ease of change. The [maintainability profile](maintainability.md) supplies positive, scenario-based evidence separately. Its ordinal ratings and confidence are not calculator inputs and do not alter this table. Profile findings count only when they establish independently actionable causes under the relevant topic rubric.
