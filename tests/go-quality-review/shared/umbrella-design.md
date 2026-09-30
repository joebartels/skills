# Go quality report — design and execution notes

The user approved the orchestration design and deduplicated-findings grading on 2026-09-27, then asked to proceed.

## Intended outcome

A provider-neutral umbrella skill reviews an identified changeset, codebase or code area using the nine existing Go topic skills. It preserves their solo judgments, produces traceable report cards, and returns a consistent overall grade plus strengths, defects, rationales and prioritized corrections. It is review-only unless the user separately requests edits.

## Architecture

The orchestrator establishes a stable target and coverage manifest, divides code into coherent bounded work packets, delegates relevant combinations of topic skills, and reconciles report cards. It keeps the manifest, summaries and unresolved questions in context; detailed evidence lives in artifacts. A targeted boundary review closes gaps between chunks. Model selection uses host capabilities and permitted configuration, without vendor/model identifiers.

Coverage and findings are separate. Every relevant target/topic/boundary obligation has an owner and disposition. A skipped irrelevant topic is Not applicable with a reason; an applicable unreviewed topic is a coverage gap. Missing agents or skills cannot create positive evidence. Related files may be read to trace effects; unchanged debt remains outside a diff grade.

The overall grade applies the existing table to unique, supported root causes, with the greatest supported consequence per cause, after resolving disputed evidence. Topic grades are retained and topic totals are never averaged or summed. Independent defects accumulate regardless of chunking. Full-scope grades require complete material coverage; otherwise the overall result is Insufficient evidence, with an optional clearly scoped defect grade for reviewed work. Not applicable requires no relevant obligation. A+ requires complete coverage, zero defects and two independent verified nonroutine safeguards.

## Deliverables and checks

- `go-quality-report/SKILL.md`: lean entrypoint, workflow and reference routing.
- Local references: packet/routing rules, reconciliation/grading, output/artifact contract.
- `scripts/grade.py`: standard-library calculator for an already reconciled ledger; no inference of root causes or evidence truth.
- Evaluation cases: scope ambiguity, no-applicability, routing, bounded work, model fallback, missing evidence, duplicate IDs, independent test gaps, contradictory severity, partition invariance and A+ safeguard identity.
- One executable small Go fixture for a delegated end-to-end review.
- Update catalog and validator without changing the nine topic skills.

Execution: capture baseline behavior; write calculator tests before implementation; author the skill and local references; run blind behavioral cases and a delegated fixture review; make evidence-based corrections; save evaluation outputs and limitations. No installation, commits or publishing are part of this task.
