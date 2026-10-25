---
name: go-architecture-and-design
description: Review a defined Go changeset or code area for architecture and design quality across libraries, CLIs, and services. Use when asked to grade package boundaries, APIs, dependencies, interfaces, composition, or lifecycle design; not for a general Go style or performance review.
metadata:
  review-contract: "2"
---

# Go Architecture and Design Review

Review the specified code as a design that other code and people must use and change. Give one architecture grade, grounded in observed behavior and repository context. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Assess the relevant questions and record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact changeset or code area. Read affected callers, implementations, tests, module version, and repository conventions as needed to understand it. If the target is unclear, state the interpretation used. Do not treat unrelated existing debt as introduced by the change.
2. Identify whether the code is a library, CLI, service, or a mix; distinguish public contracts from internal implementation. Judge the cost of a design choice against its actual users and expected evolution, not an assumed application template.
3. Check the applicable design questions below. Follow a value across boundaries when necessary: who owns it, which package depends on which, how errors and cancellation travel, and who shuts down acquired resources or work.
4. Trace suspected coupling or ownership problems through actual consumers and ordinary changes. Compare the proposed boundary with the existing design's responsibilities and tradeoffs; a different package or type layout is not itself an improvement. Do not invent HTTP status, durability, or transaction requirements.

For nuanced examples and handbook corrections, read [Design decisions](references/design-decisions.md) when the review encounters layering, interfaces, context, error translation, or background work. Do not load it for an unrelated, straightforward change.

## Design questions

| Area | What to establish |
| --- | --- |
| Package and API boundaries | Do packages have coherent responsibilities and dependencies? Are exported types, names, and behaviors usable without importing unrelated implementation details? Does the API shape fit its intended consumers? |
| Data and protocol boundaries | When persistence or transport models differ from business concepts, is translation owned at a clear boundary? Is there demonstrated coupling or duplicated mapping, rather than merely a struct tag or direct handler call? |
| Interfaces and abstraction | Does an interface represent a real protocol or consumer need? Is its method set no larger than users need? Would a concrete type, function, or existing standard interface be clearer? Producer-owned interfaces are valid when the protocol itself is the product. |
| Dependencies and composition | Are important runtime dependencies visible and replaceable where needed? Is construction as simple as the invariants allow? Can users tell who configures, starts, and closes a dependency? |
| Cross-boundary behavior | Can callers observe meaningful errors and cancellation without depending on accidental implementation details? Check whether raw infrastructure errors cross a public protocol boundary. Is the success contract truthful when work continues asynchronously? |
| Simplicity and fit | Does each package, layer, DTO, option, and background task solve a present problem? Could fewer boundaries express the same design more clearly? |

Architecture grades package/API boundaries, dependencies and ownership. Unnecessary API inputs belong here when their redundancy is established. Repeated local implementation under one owner belongs to Code Quality & Go Idioms; calling the same edit cost a duplicated responsibility does not make it architectural. Grade an additional architectural consequence only when a consumer or ownership boundary imposes a distinct cost. Module resolution belongs primarily to Dependencies & Reproducibility; observable breakage to Correctness & Compatibility. Assess demonstrated boundary defects in standalone use too.

## Grade the architecture

Classify by the supported design consequence: minor for localized avoidable complexity; moderate for a concrete failure mechanism or substantial ordinary maintenance cost at a dependency or ownership boundary; major for an important supported API, ownership or lifecycle contract failure; critical for severe systemic failure. A merge recommendation, if requested, is separate from the grade.

Apply the first matching row from the top. These are grade anchors, not points awarded for following a checklist. Context can raise or lower a finding's severity, but explain the evidence for that choice.

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

If no architecture decision is implicated, return **Not applicable**. If the area is relevant but essential code, contract, or runtime context is unavailable after reasonable inspection, return **Insufficient evidence**. Explain the missing evidence and what would permit grading. Do not use either state to avoid making a supportable judgment.

## Report

Use the fixed report structure below in both the saved report and the final user-facing response, unless the user explicitly requests another format. Keep the section names and order; do not replace the findings table with prose or only a link to an artifact. Include every counted finding in the table, ordered by severity (critical to minor), then numeric finding ID. Use the same IDs, severities and topics in the table and details. Use the topic display names from the skills; separate multiple topics with `; ` in overall reports. A standalone topic report lists only its own topic. Keep table cells concise, escape literal pipes as `\|` (including inside code), and put multiline evidence in the details.

Every finding has one detail entry with the fields shown, including a correction and proposed verification. Distinguish proposed verification from executed checks. For zero findings retain the empty table header and separator, followed by `None found.`; do not invent an F0 row. Use `None found.` for empty details and `None verified.` for unsupported strengths. For Not applicable or Insufficient evidence retain the same sections and explain the state; omit counts when no findings are confirmed. Preserve any confirmed findings under Insufficient evidence without implying the remaining scope passed. Counts, when present, describe confirmed findings only. Keep optional suggestions and ungraded related observations outside the counted table, after the required sections, labeled explicitly.

A topic title identifies that topic's grade; it does not imply an overall review. Full reviews use `Go Quality Report` as the title and append the report-card table and maintainability profile in that order. Grade/state cells contain only a letter grade, `Not applicable`, or `Insufficient evidence`; put qualifications in the coverage column. Apply the existing coverage, severity and grading rules unchanged.

```markdown
# Architecture & Design — [grade | Not applicable | Insufficient evidence]
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
```

Prefer `path:line` references when files are available. For each finding, provide a corresponding change; distinguish required corrections from optional refinements. Positive findings need evidence too. The grade assesses architecture in the requested scope, not overall Go quality.
