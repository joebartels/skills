---
name: go-observability-and-resilience
description: Review and grade observability and resilience in a defined Go changeset or code area across libraries, CLIs, and services. Use for logs, metrics, traces, deadlines, retries, overload, failure isolation, probes, or graceful shutdown; not for a general Go style, security, performance, or deployment audit.
metadata:
  review-contract: "3"
---

# Go Observability and Resilience Review

Review whether operators and callers can understand failures and whether the changed code behaves predictably when dependencies slow, fail, or overload. Give one grade for this topic, grounded in the actual runtime and contract. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Assess the relevant questions and record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact diff or code area. Inspect relevant callers, transport configuration, tests, `go.mod`, telemetry setup, and deployment configuration when available. Distinguish introduced defects from existing constraints. A local library, CLI, and network service need different signals and controls.
2. Identify which work can block or fail, who owns its context and lifetime, what the caller may safely retry, and what operational signals exist at the affected boundary. Trace the actual call path before declaring a missing timeout, correlation field, or shutdown step.
3. Report distinct, substantiated consequences. Group repeated symptoms of one root cause. Treat absent service-level objectives, deployment behavior, or telemetry plumbing as limits when they affect the judgment; do not invent them. Avoid demanding instrumentation or resilience machinery for a pure local operation.
4. Check version-sensitive API advice against the module's effective Go version and the project's chosen instrumentation libraries. Existing `zap`, `zerolog`, Prometheus, OpenTelemetry, or platform facilities can be sound; migration to `slog` is not a finding by itself.

Read [Observability and resilience decisions](references/observability-resilience-decisions.md) when the review encounters logging correlation, metric or trace design, retries, circuit breakers, panic recovery, probes, or shutdown. It supplies context-sensitive distinctions and links to primary documentation.

## Review questions

| Area | What to establish |
| --- | --- |
| Logs and correlation | Do important failures produce useful, appropriately leveled events at an owned boundary without duplicate noise? Are request or trace identifiers attached where needed by the configured logger or handler? Could logged values disclose secrets or sensitive data? Structured text can be adequate; JSON and dynamic log levels are choices, not universal requirements. |
| Metrics and alerts | Can operators detect user-visible failure, latency, saturation, or backlog for this workload? Are metric names, labels, and units stable and useful? Could unbounded label values multiply time series? Are histogram boundaries suitable for decisions such as an SLO, rather than merely customized? Can telemetry export failure unexpectedly block primary work? |
| Traces | Where distributed calls exist, does configured propagation carry the active context across relevant boundaries? Are spans meaningful, ended, and annotated without leaking sensitive values? Check whether auto-instrumentation already supplies spans; do not require a child span for every function. Account for sampling when interpreting absent traces. |
| Deadlines and retries | Do external operations respect caller cancellation and an appropriate total budget? Are retries limited to safe, transient failures, with bounded attempts or elapsed time, backoff and jitter when contention warrants it, and no multiplication across layers? Consider existing SDK retries and server-provided retry guidance. |
| Failure containment | Can a slow or failed dependency exhaust workers, queues, connections, or memory? Are concurrency limits, backpressure, fallback, or a circuit breaker justified by observed failure modes? Is retry safety or deduplication defined for side-effecting operations that are actually retried? |
| Service lifecycle | For applicable services, do liveness and readiness reflect their distinct purposes without causing restart loops or needless traffic loss? Is startup probing needed? On shutdown, does the process stop accepting work, allow owned work to finish within its budget, and flush telemetry as needed? Check long-lived or hijacked connections separately. |

This topic assesses diagnostic signals, failure handling, retries, and lifecycle code. Deployment & Operations primarily assesses how manifests and rollout configuration use those controls; Performance assesses resource cost; Security assesses attacker-driven impact or sensitive disclosure; Testing assesses verification. A shared timeout, queue, or shutdown defect stays assessable here when it affects resilience; use the shared overlap rule instead of deferring it away.

## Grade observability and resilience

Classify each **distinct, substantiated** issue by its effect in this codebase. A minor issue reduces local diagnostic clarity or adds low-impact fragility; a moderate issue creates a plausible missed signal, failed request, or substantial operational burden under normal use; a major issue makes an important failure hard to detect or contain, or breaks a stated reliability contract; a critical issue creates severe systemic failure or defeats the stated purpose. A merge recommendation, if requested, is separate from the grade. Count root causes, not occurrences.

Apply the first matching row from the top. These anchors match the other Go review skills; explain context-dependent severity.

### Apply the anchors consistently

Admit a finding only when evidence connects an in-scope cause to a concrete consequence under supported use or ordinary maintenance. Trace relevant callers, guards and rule owners. Source tracing can suffice without an executed reproduction. If the consequence depends on an unverified assumption, keep it ungraded and name the missing evidence; uncertainty is not minor severity.

Grade maintenance costs through a specific reading/change task and a correction that removes the cost while preserving required responsibilities. Trace a suspected redundant input/wrapper's purpose or the decision requiring synchronized edits in repeated code; similar calls, values or construction patterns alone are insufficient. Without that evidence, the simplification remains ungraded. Established local friction or a small repeated edit is minor; moderate requires a concrete failure mechanism or substantial coordination/reimplementation cost beyond site counts. Different representations and hypothetical capabilities alone are ungraded.

Count independently actionable causes: merge symptoms resolved by one causal correction; retain unrelated corrections. Strengths do not cancel defects. Judge each topic's consequence independently; another topic's severity or a second label for the same maintenance cost does not establish an additional topic consequence.

An important supported-contract failure is normally major even if the fix is small. Systemic major requires demonstrated reach across distinct important workflows; critical requires supported severe, broad or irreversible impact. Severity reflects consequence, not confidence, remediation size, or words such as race or unsafe.

For A+, establish two independent, verified nonroutine safeguards controlling distinct meaningful risks. Routine correct setup supports A; complexity/tool counts or repeated descriptions do not support A+.

Use Not applicable only for irrelevant topics; Insufficient evidence for essential missing evidence. Confirmed findings may receive an inspected-scope grade with explicit coverage limits. Do not award A/A+ with unassessed material risk or without a verified relevant strength.

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

If the target has no relevant failure, operational signal, or lifecycle decision, return **Not applicable**. If the topic is relevant but essential code, contract, or runtime configuration is unavailable after reasonable inspection and could change the judgment, return **Insufficient evidence**. Explain what is missing and what would permit grading. Do not use either state to avoid a supportable judgment.

## Report

Create a review record using the shared [record contract](../go-quality-report/references/review-record.md), then run its renderer. Reconcile candidate evidence and causes before assigning topic severity; record missing material evidence as unresolved. Use the rendered report for the saved artifact and final response unless the user requests another format. The shared helper is bundled with this plugin; other topic audits are not prerequisites.
