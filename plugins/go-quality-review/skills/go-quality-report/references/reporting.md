# Artifacts, worker handoff and final report

## Lean orchestration

Use one review artifact directory in an allowed scratch/output location; do not modify reviewed source to hold artifacts unless the user requests that location. Share paths only when workers can access them; otherwise use the host's artifact-transfer mechanism. Preserve a recoverable manifest and cards so synthesis can resume after context compaction. Do not reproduce secrets in evidence.

Keep:

- `manifest`: requested scope and snapshot, exclusions, chunk/boundary owners, applicability, coverage statuses, worker identifiers and card paths.
- `cards/<chunk>/<topic>.md`: complete standard reports from the existing skills; separate card per topic even if a worker used several skills.
- `findings`: canonical causes, source ID mapping, evidence, topic/overall severities, primary remediation owner, linked correction/verification and reconciliation notes. Keep rejected/uncertain claims separate.
- `ledger.json`: final confirmed grading inputs described in grading.md.
- `report.md`: final synthesis linked to the cards and evidence.
- `maintainability.md`: the separate seven-dimension profile, scenario traces/experiments, measurements, confidence and coverage limits. See [profile contract](maintainability.md).

Filenames are an example layout, not a runtime dependency. With no filesystem, use available artifact storage or compact structured records and disclose any auditability limit.

## Worker handoff

Workers return a compact summary (typically a few hundred words): review/chunk/snapshot; assigned skill names actually read; card paths; per-topic grade/state and severity counts; local finding IDs with cause and evidence pointers; covered paths/contracts and additional context inspected; unreviewed obligations, disagreements and suggested next checks. Store lengthy commands/logs in artifacts with exact command and result summaries in cards. Distinguish supplied results, actual executed checks and proposed verification.

The orchestrator reads these handoffs first. Retrieve full cards or source excerpts for duplicate resolution, disputed claims, missing evidence, grade-changing findings and cross-boundary questions. Do not discard low-severity independent findings merely to shorten the final report; retain all in the deduplicated index and counts.

## Final report

Use the fixed report structure below in both the saved report and the final user-facing response, unless the user explicitly requests another format. Keep the section names and order; do not replace the findings table with prose or only a link to an artifact. Include every counted finding in the table, ordered by severity (critical to minor), then numeric finding ID. Use the same IDs, severities and topics in the table and details. Use the topic display names from the skills; separate multiple topics with `; ` in overall reports. A standalone topic report lists only its own topic. Keep table cells concise, escape literal pipes as `\|` (including inside code), and put multiline evidence in the details.

Every finding has one detail entry with the fields shown, including a correction and proposed verification. Distinguish proposed verification from executed checks. For zero findings retain the empty table header and separator, followed by `None found.`; do not invent an F0 row. Use `None found.` for empty details and `None verified.` for unsupported strengths. For Not applicable or Insufficient evidence retain the same sections and explain the state; omit counts when no findings are confirmed. Preserve any confirmed findings under Insufficient evidence without implying the remaining scope passed. Counts, when present, describe confirmed findings only. Keep optional suggestions and ungraded related observations outside the counted table, after the required sections, labeled explicitly.

A topic title identifies that topic's grade; it does not imply an overall review. Full reviews use `Go Quality Report` as the title and append the report-card table and maintainability profile in that order. Grade/state cells contain only a letter grade, `Not applicable`, or `Insufficient evidence`; put qualifications in the coverage column. Apply the existing coverage, severity and grading rules unchanged.

```markdown
# Go Quality Report — [grade | Not applicable | Insufficient evidence]
Scope: [mode, exact revisions/snapshot, target area, exclusions, relevant version/runtime context]
Coverage: [assessed areas and boundaries; material unassessed areas and why]
Rationale: [why the evidence, coverage and counted severities select this result]
Finding counts: critical=[n], major=[n], moderate=[n], minor=[n]

## Findings summary

| ID | Severity | Topic | Finding / consequence | Evidence | Correction |
| --- | --- | --- | --- | --- | --- |
| F1 | [severity] | [topic name(s)] | [cause and present consequence] | [source reference] | [targeted correction] |

## Finding details

### F1 — [short title]

Severity: [critical | major | moderate | minor]
Topic: [same topic name(s) as the summary row]
Status: [introduced | worsened | existing-in-scope]
Trigger: [conditions or maintenance scenario exposing the issue]
Evidence: [source locations and supported facts; distinguish inspected, supplied and executed evidence]
Consequence: [supported effect and why this severity applies]
Correction: [specific change and primary owner; why it addresses the cause]
Verification: [how to verify the correction; label proposed checks]

## Strengths

- [G1] [specific verified choice and evidence/location] — [why it matters]

## Checks and limits

[Executed checks with exact commands and outcomes; supplied results labeled separately; checks not run, coverage gaps and unresolved assumptions.]

## Report cards

| Topic | Grade/state | Assessed chunks and coverage limits | Cards |
| --- | --- | --- | --- |
| [all nine topics, one row each] | [rollup or Not applicable/Insufficient evidence] | [reason and extent] | [links] |

## Maintainability profile

[Seven-dimension table with rating, confidence, evidence and scope limits; link scenarios/metrics. Summarize cleanliness, structure and extensibility separately from the defect grade.]

Details: [manifest, deduplicated index, ledger and underlying cards]
```

For an incomplete review, name the overall state first. If useful, add `Assessed-scope defect grade: C (reviewed areas only)` with confirmed findings; this is not the requested full-scope grade. No confirmed findings with incomplete material coverage does not warrant an assessed-scope A. A Not applicable result has no findings or counts.

The summary uses each canonical cause's reconciled overall severity. Preserve different supported topic severities in its detail entry and underlying cards; do not force the overall severity onto every topic. Counts are unique causes, not sums of topic rows. Every counted finding stays in the user-facing table and details, including minor findings. Link long logs and supporting evidence instead of reproducing them.

The topic table must distinguish justified irrelevance from unavailable or unreviewed work. Topic rollups retain the existing coverage and deduplication rules. Link chunk cards rather than pasting them into synthesis. Corrections should name dependencies on other finding IDs where relevant. Do not turn a review grade into an unrequested merge/deploy action.

For maintainability-only requests, use the profile output contract without fabricating nine topic grades or a defect report. Prior defect reports may be linked with their exact snapshot and execution date, clearly labeled historical.
