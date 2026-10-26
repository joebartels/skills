---
name: go-performance-and-resource-management
description: Review and grade performance and resource management in a defined Go changeset or code area across libraries, CLIs, and services. Use for algorithmic cost, allocations, memory retention, GC, I/O and connection reuse, concurrency limits, contention, profiling, benchmarks, or PGO; not for a general Go style, security, or deployment audit.
metadata:
  review-contract: "3"
---

# Go Performance and Resource Management Review

Review whether the specified Go code uses time, memory, connections, goroutines, and other finite resources appropriately for its actual workload. Give one grade for this topic, grounded in code behavior and available measurements. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Follow the shared [source review sequence](../go-quality-report/references/source-review.md) within this topic; record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact diff or code area, its library, CLI, or service context, and relevant workload: input size, call frequency, concurrency, latency or memory budget, and resource limits when known. Inspect callers, benchmarks, profiles, configuration, `go.mod`, `go.work`, build tags, and deployment settings only as needed. Distinguish introduced issues from relevant existing constraints; do not grade unrelated debt.
2. Trace the work and its ownership. What is allocated or retained per call? What grows with input or concurrency? Who closes files, rows, bodies, timers, connections, and background work? Which paths block when a pool or queue is full? Establish whether a suspected cost is on a meaningful path before proposing an optimization.
3. Check the effective Go language and toolchain versions before recommending a version-specific API or interpreting runtime behavior. Verify a current claim against Go documentation when it matters; compiler thresholds, runtime defaults, and benchmark results can change by release and platform.
4. Prefer comparative evidence for speed and allocation claims: representative benchmarks, profiles, traces, resource metrics, or a clearly derived complexity bound. A deterministic leak or unbounded growth path can be reported from code without a production profile. Treat a compiler diagnostic, missing benchmark, or optional optimization as a lead, not an automatic defect. State uncertainty and checks not run rather than inventing a regression.
5. Report distinct root causes with code or measurement evidence and concrete consequences. Give verified strengths the same specificity. Connect each negative finding to a targeted change and explain the expected benefit and how to verify it.

Read [Performance decisions](references/performance-decisions.md) when a finding involves GC or `GOMEMLIMIT`, allocation or `sync.Pool`, profiling or PGO, benchmarks, database or HTTP pools, streaming, concurrency primitives, compiler tricks, or `unsafe`. It supplies context-sensitive distinctions and links to primary Go guidance.

## Review questions

| Area | What to establish |
| --- | --- |
| Work and complexity | Does work scale acceptably with realistic input, calls, and concurrency? Are repeated scans, copying, serialization, database round trips, or N+1 I/O on a consequential path? An asymptotic concern needs a plausible input range, not just a notation. |
| Memory and GC | Are large buffers or object graphs retained longer than needed? Would a measured hot path benefit from preallocation, streaming, reuse, or fewer pointers without undue copying or complexity? Is a `sync.Pool` used only for suitable temporary objects, with safe ownership and retention? Does a service memory limit account for non-Go memory and actual headroom? |
| Resource lifetime | Are `Close`, `Stop`, cancellation, and `rows.Err` handled where they affect completion? Can early returns, loops, or blocked goroutines retain files, sockets, timers, heap, or work indefinitely? Is a resource intentionally transferred to another owner? |
| I/O and pools | Are clients and transports reused when appropriate? Do HTTP bodies get closed, and is any attempted drain bounded by response size and cost? Are database pool limits justified by `DB.Stats`, database capacity, and wait behavior rather than fixed ratios? Is streaming preferable to materializing data for the expected volume? |
| Concurrency and contention | Can goroutines, queues, and in-flight work grow beyond the resource budget? Are backpressure and cancellation effective? If locks or atomics are changed for speed, do profiles show contention and does the replacement preserve correctness? |
| Measurement and compiler choices | Does the supplied benchmark or profile support the claimed performance conclusion for representative work and a comparable baseline? Are relevant allocations or tail behavior considered? Are PGO, inlining, bounds-check, layout, or `unsafe` changes supported by a measurable benefit and their correctness constraints? |

Performance assesses supported cost, resource growth/lifetime, and whether evidence justifies an optimization; preserving ownership and correctness is part of validating an optimization. Testing primarily assesses benchmark implementation; Architecture assesses lifecycle contracts; Observability & Resilience assesses failure containment; Security assesses attacker-driven exposure; Deployment & Operations assesses configured budgets. Apply the shared overlap rule to leaks, unsafe optimizations, and unbounded work. Do not demand profiling, PGO, runtime tuning, pooling, or a fixed benchmark count without a demonstrated need.

## Grade performance and resource management

Classify each **distinct, substantiated** issue by its effect in this codebase. A minor issue causes localized avoidable cost with little expected impact; a moderate issue causes a plausible material cost or resource pressure under normal use; a major issue breaks an important performance or resource budget, or makes likely exhaustion difficult to avoid; a critical issue creates severe systemic exhaustion or defeats the code's stated purpose. Include magnitude, frequency, and workload assumptions in the severity rationale. A merge recommendation, if requested, is separate from the grade. Count root causes, not occurrences. An unmeasured micro-optimization is not a finding merely because it could be faster.

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

If the target contains no assessable performance or resource decision, return **Not applicable**. If the topic is relevant but essential workload, code, or runtime context is unavailable after reasonable inspection and could change the judgment, return **Insufficient evidence**. Explain what is missing and what would permit grading. Absence of a profile alone does not make an obvious leak ungradable; absence of measured impact may make a speculative micro-optimization unreportable. Do not use either state to avoid a supportable judgment.

## Report

Create a review record using the shared [record contract](../go-quality-report/references/review-record.md), then run its renderer. Reconcile candidate evidence and causes before assigning topic severity; record missing material evidence as unresolved. Use the rendered report for the saved artifact and final response unless the user requests another format. The shared helper is bundled with this plugin; other topic audits are not prerequisites.
