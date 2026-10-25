---
name: go-quality-report
description: Orchestrate a comprehensive graded Go quality review of a changeset, codebase, package, or supplied code. Use when the user wants an overall report across relevant Go quality topics, delegated report cards, a maintainability comparison, or a pre-merge quality assessment; use a single topic skill for a specifically limited topic review.
metadata:
  review-contract: "3"
---

# Go Quality Report

Coordinate the nine Go review skills and synthesize one evidence-backed overall result plus a separate maintainability profile. Review only; fixes, installation, publication and merge actions require their own user request. This skill needs the applicable topic skills, discovered by name through the host or a local skill catalog. It does not contain substitutes for them.

## 1. Fix the target

Use the user's explicit target. For a changeset, record exact base/head or the supplied patch; for staged, unstaged or untracked work, record the actual snapshot and included paths. For a codebase/code-area review, record the revision/snapshot and named area. Ask a focused question only if plausible interpretations materially change what gets reviewed. Offer observed choices; do not guess a default branch or silently substitute the current checkout for a requested revision.

For maintainability-only requests, read [maintainability](references/maintainability.md), assess that profile, and do not invent or rerun unrelated risk grades. Full codebase/code-area reports include the profile; changeset reports assess affected dimensions and explain the remainder.

Inspect metadata and file inventory first: relevant instructions, changed paths/statistics, modules, project type, support versions and tests/build configuration. Delegate deeper discovery for a large tree. Track later source changes and invalidate affected cards. Non-Go files such as migrations, CI, schemas and manifests can affect Go quality. If nothing implicates the Go topics, return Not applicable with scope and reason.

## 2. Map coverage and bounded work

Read [orchestration](references/orchestration.md) to discover topic skills, route risks, size work and dispatch workers. Create a small manifest of the target, exclusions, chunks, important boundary contracts, topic applicability and owners. Every requested area must be accounted for. Assign cross-boundary and repository-wide obligations explicitly; local approvals alone do not verify integration.

Start with coherent domains or behavior boundaries; combine a small change into one packet and split large ones further. Select topic combinations per packet, not one agent per skill by default. Every topic is considered, but invoke only those with a relevant review question. Record why topics are not applicable; unavailable evidence or missing skills are coverage gaps, not irrelevance. Related files can be read to trace effects. Changeset grades include only introduced, worsened or newly exposed defects; code-area grades include existing in-scope defects.

## 3. Dispatch and collect

Use the host's available delegation API. Give each worker the precise target, bounded paths/contracts, resolved topic-skill locations, relevant context, read-only constraints and artifact destination. Workers must read/invoke each assigned skill and its relevant references, returning a structured record and rendered card for each topic, plus inspected coverage and a short handoff. Do not send expected findings or preselect grades.

Choose permitted model capabilities and reasoning effort for the task; respect user pins and actual host support. Use bounded queues within concurrency limits. If delegation is unavailable, review bounded packets sequentially with the same skills and disclose that limitation. Missing applicable skills stay explicitly unassessed.

Keep the orchestrator context to the manifest, compact card summaries, finding index and unresolved questions. Store full cards and evidence in artifacts; retrieve relevant portions for reconciliation. Follow [artifact and output contract](references/reporting.md). Do not paste entire files, logs or repeated cards into synthesis context.

## 4. Reconcile and close coverage

Read [grading and reconciliation](references/grading.md). Check record scope, skill invocation, evidence and limits before considering grades. Namespaced local finding IDs prevent collisions. Deduplicate the same causal defect across chunks/topics, retaining separate independently actionable defects. Resolve contradictions against evidence, not votes or the most alarming label. Return focused questions to workers or inspect the necessary code. Keep original records and cards; reconcile into one canonical record with source IDs and accepted corrections.

Perform targeted boundary reviews and close remaining material obligations. Retry/reassign an incomplete packet only with a concrete recovery plan; repeated identical blocked attempts add no evidence. Respect user time/cost constraints and report residual gaps. Sampling or a completed subset must not become a full-codebase claim.

## 5. Grade and report

Grade unique confirmed root causes using the shared rubric, never averages of card grades, clean-chunk dilution or automatic worst-card selection. Recompute topic rollups from unique in-topic causes when useful. A+ safeguards must also be independent and verified. Changing packet boundaries alone must not change the result.

Use the [review record](references/review-record.md) for canonical findings, topic assessments, coverage and ungraded decisions. Run `python3 /absolute/path/to/go-quality-report/scripts/render.py /output/review.json --output /output/report.md`, resolving the script relative to this skill. It validates the record and calls the existing calculator for unique overall and per-topic grades; do not author counts or duplicate tables manually. Python 3.10+ needs no additional packages. If unavailable, disclose that limit and follow the [report structure](references/reporting.md) manually.

Deliver the rendered report with its full findings table and details unless the user requests another format. Material coverage gaps or unresolved evidence yield Insufficient evidence while retaining confirmed defects. Topic cards and provenance remain linked for inspection.
Apply [maintainability](references/maintainability.md) alongside the defect report: seven evidence-backed ratings with confidence, representative change scenarios, and separate conclusions on cleanliness, structure and extensibility. Keep profile judgments separate from grade.py arithmetic.
