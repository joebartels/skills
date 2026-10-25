# Review artifacts and delivery

Use one artifact directory outside the reviewed checkout. Keep the snapshot/coverage manifest, worker records and rendered cards, reconciled `review.json`, rendered `report.md`, and supporting evidence. A separate maintainability artifact is useful for long scenario traces. Share paths only when workers can access them; otherwise use the host's artifact transfer. Preserve enough to resume synthesis after compaction without copying source files or secrets into handoffs.

## Worker handoff

Return the chunk/snapshot, topic skills read, record/card paths, canonical candidate IDs and evidence pointers, inspected boundaries, unreviewed obligations and unresolved questions. Keep command logs in artifacts. Distinguish supplied results, executed checks and proposed verification. The orchestrator reads compact handoffs first, then the records and relevant source for contradictions, shared corrections and boundary questions. Preserve independent minor findings when shortening prose.

## Record and rendering

Follow [review-record.md](review-record.md). Worker records and overall synthesis use the same format. Preserve raw worker records; map namespaced source IDs into canonical findings during reconciliation. Report counts and grades are projections of that record, not copied from worker summaries or averaged.

The standard report contains, in order:

1. Topic or `Go Quality Report` title with computed grade/state; Scope, Coverage, Rationale and computed Finding counts. Incomplete reports omit zero counts and may include an assessed-scope defect grade when defects are confirmed.
2. **Findings summary:** ID, Severity, Topic, Finding / consequence, Evidence, Correction. Every counted cause appears once, ordered by severity then numeric ID.
3. **Finding details:** one matching ID/title with Severity, Topic, Status, Trigger, Evidence, Consequence, Correction and Verification. Verification is proposed; actual execution is recorded separately. Include analysis and source provenance.
4. **Strengths:** verified choices and evidence, plus independently supported safeguards.
5. **Checks and limits:** exact commands/outcomes, supplied results, unexecuted checks and material limits.
6. Overall only: **Report cards**, with all nine topics, calculated grade/state, coverage and source-card references; then the seven-row **Maintainability profile**, with separate cleanliness, structure and extensibility conclusions.
7. **Ungraded decisions**, when present: rejected candidates, unresolved evidence and merges into canonical findings.

Empty findings retain the table header and `None found.` in summary and details. Unsupported strengths say `None verified.`. Not applicable and Insufficient evidence retain the structure without implying a clean assessment. Topic reports grade only their topic; distinct supported topic severities remain visible in overall finding details. Profile ratings never alter defect arithmetic.

Deliver the rendered report as both saved artifact and final response unless the user requests another presentation. Link long evidence rather than expanding table cells or dropping findings. With no Python, use this same structure manually and disclose that the report was not deterministically rendered. Maintainability-only requests continue to use the [profile contract](maintainability.md), without fabricating topic grades.
