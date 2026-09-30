# Artifacts, worker handoff and final report

## Lean orchestration

Use one review artifact directory in an allowed scratch/output location; do not modify reviewed source to hold artifacts unless the user requests that location. Share paths only when workers can access them; otherwise use the host's artifact-transfer mechanism. Preserve a recoverable manifest and cards so synthesis can resume after context compaction. Do not reproduce secrets in evidence.

Keep:

- `manifest`: requested scope and snapshot, exclusions, chunk/boundary owners, applicability, coverage statuses, worker identifiers and card paths.
- `cards/<chunk>/<topic>.md`: complete standard reports from the existing skills; separate card per topic even if a worker used several skills.
- `findings`: canonical causes, source ID mapping, evidence, topic/overall severities, primary remediation owner, linked correction/verification and reconciliation notes. Keep rejected/uncertain claims separate.
- `ledger.json`: final confirmed grading inputs described in grading.md.
- `report.md`: final synthesis linked to the cards and evidence.

Filenames are an example layout, not a runtime dependency. With no filesystem, use available artifact storage or compact structured records and disclose any auditability limit.

## Worker handoff

Workers return a compact summary (typically a few hundred words): review/chunk/snapshot; assigned skill names actually read; card paths; per-topic grade/state and severity counts; local finding IDs with cause and evidence pointers; covered paths/contracts and additional context inspected; unreviewed obligations, disagreements and suggested next checks. Store lengthy commands/logs in artifacts with exact command and result summaries in cards. Distinguish supplied results, actual executed checks and proposed verification.

The orchestrator reads these handoffs first. Retrieve full cards or source excerpts for duplicate resolution, disputed claims, missing evidence, grade-changing findings and cross-boundary questions. Do not discard low-severity independent findings merely to shorten the final report; retain all in the deduplicated index and counts.

## Final report

```markdown
# Go Quality Report — [F through A+ | Not applicable | Insufficient evidence]
Scope: [mode, exact revisions/snapshot, target area, exclusions]
Coverage: [completed areas/topics and boundaries; missing material coverage]
Rationale: [why unique counts and impact support the overall result]
Unique finding counts: critical=[n], major=[n], moderate=[n], minor=[n]

## Report cards

| Topic | Grade/state | Assessed chunks and coverage limits | Cards |
| --- | --- | --- | --- |
| [all nine topics] | [rollup or Not applicable/Insufficient evidence] | [reason and extent] | [links] |

## Good

- [G1] [verified strength and evidence] — [why it matters]

## Bad

- [F1][severity] [canonical cause, scope/location and trigger] — [supported consequence; affected topics]

## Suggested changes

- [F1][priority] [specific correction; primary owner] — [why; how to verify]

## Limits

[Unreviewed areas, unresolved material assumptions, tool/agent/skill limits, exact checks and outcomes, and whether reviews were independent or sequential.]

Details: [manifest, deduplicated index and underlying cards]
```

For an incomplete review, name the overall state first. If useful, add `Assessed-scope defect grade: C (reviewed areas only)` with the confirmed counts; this is not the requested full-scope grade. For Not applicable or no assessable evidence, omit counts and unsupported sections, explain the state and what would enable review. For a complete letter grade, retain all sections and use None verified/found/needed when appropriate.

The topic table must distinguish justified irrelevance from unavailable or unreviewed work. A topic rollup has the same coverage and deduplication rules within that topic's applicable scope. Link chunk cards rather than pasting them into the synthesis. Prioritize fixes by supported impact/dependency order, retaining the rationale; optional improvements are explicitly ungraded. With many findings, summarize the most consequential ones and link the complete index so every counted finding and correction is accessible. Do not turn a review grade into an unrequested merge/deploy action.
