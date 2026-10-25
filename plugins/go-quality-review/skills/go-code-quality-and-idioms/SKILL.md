---
name: go-code-quality-and-idioms
description: Review and grade code quality and Go idioms in a defined Go changeset or code area across libraries, CLIs, and services. Use for readability, naming, error handling, documentation, value semantics, or version-aware language and standard-library choices; not for a general architecture, testing, security, or performance audit.
metadata:
  review-contract: "2"
---

# Go Code Quality and Idioms Review

Review the specified Go code for clarity, appropriate local idioms, and ease of maintenance. Give one code-quality grade based on observed consequences, not a count of style rules. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Assess the relevant questions and record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact diff or code area and inspect enough callers, types, tests, and nearby conventions to understand it. Distinguish newly introduced issues from existing constraints. Do not grade unrelated debt.
2. Check `go.mod`, relevant `go.work` context, and file build constraints before judging a language feature or old idiom. The installed toolchain does not by itself determine a file's effective language version. If version-sensitive advice matters and that context is unavailable, qualify the finding rather than assume the newest semantics.
3. Determine whether code is generated. Review its generator or source of truth when available; do not request hand edits to generated output. Apply the same quality standard to libraries, CLIs, and services, scaled to their actual users and contracts.
4. Trace suspected clarity or idiom problems at their use sites. Explain the existing reading or change burden and how the correction reduces it while preserving behavior. Identify verified strengths as specifically as defects.

Read [Idiom decisions](references/idiom-decisions.md) when a finding depends on error exposure, nil versus empty values, receivers, comments, testing idioms, concurrency, or newer Go features. It supplies context-sensitive distinctions and links to primary Go guidance.

## Review questions

| Area | What to establish |
| --- | --- |
| Error flow | Are returned errors checked or deliberately handled? Is context useful without redundant wrapping or duplicate logs? Does `%w` expose a cause callers should be allowed to depend on? Does an ignored cleanup or deferred error affect the result? |
| Readability | Can a reader follow the normal path and state changes? Do guard clauses reduce real nesting? Are helpers, abstractions, generics, and comments earning their complexity? Are shadowed variables or implicit side effects hiding behavior? |
| Names and documentation | Are names clear at their use sites, with scope-appropriate length and consistent initialisms? Does exported documentation state the contract where users need it? Is apparent package stutter actually redundant in client code? |
| Values and methods | Do nil, empty, and zero values preserve the intended observable contract? Do receiver choices preserve mutation and method-set behavior? Are copy-sensitive values such as mutexes copied? |
| Version-aware idioms | Would a supported standard-library or language feature make this code materially clearer? Is that feature available under the file's effective Go version? Avoid modernization churn that changes semantics or merely swaps syntax. |
| Mechanical evidence | If files are available, use `gofmt -d` and relevant `go vet` or build output when useful. A tool diagnostic is evidence to investigate, not an automatic grade. Do not run mutating formatters or `go fix` during a review; `go fix -diff` can preview suggestions on supported toolchains. |

Code Quality grades local clarity, repeated implementation, error flow, value semantics and version-appropriate idioms, including language-semantics bugs in standalone use. Unnecessary API or dependency plumbing belongs to Architecture; restating that boundary cost as reader friction does not establish a local finding. Grade additional local effects only with use-site evidence beyond the boundary itself. Correctness also assesses observable breakage; Testing owns assertion strategy, Security exploitability, and Performance measured cost. Optional stylistic preferences are ungraded.

## Grade code quality

Classify by the supported local consequence: minor for localized readability friction or avoidable inconsistency; moderate for a concrete bug mechanism in the data/control flow or substantial ordinary maintenance cost; major for broken important behavior; critical for severe systemic failure. Optional style preferences are ungraded. A merge recommendation, if requested, is separate from the grade.

Apply the first matching row from the top. These anchors match the other Go review skills; explain any severity judgment that depends on context.

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

If the target contains no assessable Go code or idiom decision, return **Not applicable**. If it is relevant but essential code or version context is unavailable after reasonable inspection and would change the judgment, return **Insufficient evidence**. Explain what is missing and what would permit grading. Do not use either state to avoid a supportable judgment from the available code.

## Report

Use the fixed report structure below in both the saved report and the final user-facing response, unless the user explicitly requests another format. Keep the section names and order; do not replace the findings table with prose or only a link to an artifact. Include every counted finding in the table, ordered by severity (critical to minor), then numeric finding ID. Use the same IDs, severities and topics in the table and details. Use the topic display names from the skills; separate multiple topics with `; ` in overall reports. A standalone topic report lists only its own topic. Keep table cells concise, escape literal pipes as `\|` (including inside code), and put multiline evidence in the details.

Every finding has one detail entry with the fields shown, including a correction and proposed verification. Distinguish proposed verification from executed checks. For zero findings retain the empty table header and separator, followed by `None found.`; do not invent an F0 row. Use `None found.` for empty details and `None verified.` for unsupported strengths. For Not applicable or Insufficient evidence retain the same sections and explain the state; omit counts when no findings are confirmed. Preserve any confirmed findings under Insufficient evidence without implying the remaining scope passed. Counts, when present, describe confirmed findings only. Keep optional suggestions and ungraded related observations outside the counted table, after the required sections, labeled explicitly.

A topic title identifies that topic's grade; it does not imply an overall review. Full reviews use `Go Quality Report` as the title and append the report-card table and maintainability profile in that order. Grade/state cells contain only a letter grade, `Not applicable`, or `Insufficient evidence`; put qualifications in the coverage column. Apply the existing coverage, severity and grading rules unchanged.

```markdown
# Code Quality & Go Idioms — [grade | Not applicable | Insufficient evidence]
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

Prefer `path:line` references when files are available. Give each finding a corresponding change; distinguish required corrections from optional refinements. The grade assesses code quality and idioms in the requested scope, not overall Go quality.
