# Reconciliation and overall grading

## Establish the evidence before counting

1. Validate each card's target snapshot, assigned scope, topic skill use and evidence. Preserve the topic's report structure. A letter grade without findings/counts or verified strengths is not enough to synthesize. Report errors and missing cards are coverage gaps until corrected.
2. Namespace each local ID as `chunk/topic/local-id`; local F1 is not globally unique. Build one canonical finding per independently actionable cause, retaining all source-card IDs and evidence. Merge repeated symptoms when the same causal correction resolves them; shared wording, location or severity alone is insufficient. A production fix and a test change remain distinct only when independently necessary.
3. For each unique cause, use its greatest **supported** consequence across topics for the overall severity. A critical label requires supported severe/broad/irreversible reach. A systemic major requires demonstrated multiple distinct important boundaries/workflows, not repeated filenames or many agent mentions. Ask the responsible worker to substantiate a disputed classification; do not average disagreement or blindly accept the largest label. If an unresolved dispute could change the grade, keep that material uncertainty in coverage/limits rather than presenting a settled overall grade.
4. Preserve the original cards. Record reconciliation decisions, accepted evidence, canonical ID, primary remediation owner and topic-specific consequences. Correct a card's unsupported finding or grade with an explicit addendum. Topic severities may legitimately differ; do not force the overall severity onto every topic. Recompute any topic rollup from its unique supported in-topic causes using its own rubric, not the mean or worst chunk grade.
5. Deduplicate positive safeguards too. Two descriptions of one cancellation mechanism or one controlled risk do not establish two independent safeguards. A+ needs two verified nonroutine controls of distinct meaningful risks and complete material coverage, not two A+ cards or a minimum number of tools.

## Scope and grade

For a full requested scope, all material obligations must be complete or justified Not applicable. Partial, pending, unavailable or outcome-changing uncertainty means overall **Insufficient evidence**. Still report confirmed defects; an optional **Assessed-scope defect grade** describes their deduplicated counts only and must not be presented as the full result. No confirmed defects plus incomplete coverage does not warrant a partial A. If no topic is applicable, use **Not applicable**.

With complete coverage, apply the first matching row below to the unique confirmed causes. A contained important-contract failure is normally major. Moderate means a concrete localized quality/verification gap; minor means a local clarity/diagnostic/maintainability issue with limited consequence. Use each topic's specific definitions and impact evidence. Strengths do not cancel defects. Independent defects accumulate across the whole requested scope regardless of packet boundaries, number of reviewers or codebase size. No normalization by lines or averaging letter grades.

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

## Reconciled ledger and calculator

Use `python3 scripts/grade.py /path/to/ledger.json` from this skill directory, or invoke the script through its resolved absolute path. Python 3.10+; no extra packages. Its input is the orchestrator's **already reconciled** ledger. The calculator does not merge raw cards, discover missing obligations, verify evidence or assess whether two risk descriptions are semantically independent.

```json
{
  "scope": "base abc123 to head def456; complete requested diff",
  "coverage": [
    {"id": "accounts/authorization", "status": "complete", "reason": "handler and domain cards plus boundary trace"},
    {"id": "deployment", "status": "not_applicable", "reason": "no release/runtime decision implicated"}
  ],
  "findings": [
    {"id": "F1", "severity": "major", "systemic": false,
     "cause": "update occurs before tenant authorization",
     "evidence": "handler.go:42; denied caller reaches Store.Update",
     "correction": "authorize before persistence",
     "sources": ["accounts/security/F1", "accounts/correctness/F2"]}
  ],
  "strengths": [{"id": "G1", "evidence": "request deadline forwarded to storage; cancellation test passes as recorded"}],
  "safeguards": []
}
```

Coverage status is `complete`, `not_applicable`, `pending`, `partial` or `unavailable`; each needs a reason. IDs must be unique within each collection. Findings require cause, evidence, correction and severity; `systemic` is boolean and only applies to major. A critical or systemic-major finding additionally needs a `reach` explanation. Calculator evidence fields are nonempty strings: summarize evidence arrays from the detailed finding index into a string when preparing this ledger. Strengths require evidence. Each safeguard requires `id`, canonical `risk`, `evidence` and boolean `nonroutine`; only verified controls belong in this ledger. Normalize equivalent risks before calculation. Zero findings is not itself positive evidence.

Maintain card source links, locations, verification plans and ownership in the working finding index; extra ledger fields are allowed and preserved by the artifact, though the calculator only reads the grading fields. Rejected/unconfirmed concerns live separately and do not enter counts. Calculator output names the overall state, optional assessed-scope defect grade, unique counts and coverage gaps. Explain its result in prose and audit the source-to-canonical mapping.
