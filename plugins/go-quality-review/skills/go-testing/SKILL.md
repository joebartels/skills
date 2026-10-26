---
name: go-testing
description: Review and grade testing and verification for a defined Go changeset or code area across libraries, CLIs, and services. Use for test coverage of behavior, assertions, isolation, integration boundaries, concurrency tests, fuzzing, or benchmark validity; not for a general Go style, security, or performance audit.
metadata:
  review-contract: "3"
---

# Go Testing Review

Review whether the available tests and verification would detect meaningful regressions in the specified Go code. Give one Testing grade based on observable evidence and the risk of the changed behavior. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Follow the shared [source review sequence](../go-quality-report/references/source-review.md) within this topic; record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact diff or code area, its behavior contract, and the relevant library, CLI, or service context. Inspect affected implementation, tests, fixtures, CI configuration, `go.mod`, and file build constraints as needed. Distinguish issues introduced by the change from relevant existing gaps; do not grade unrelated debt.
2. Map the changed behavior to tests that exercise it. Check assertions and failure paths, not only test names, test count, or coverage percentage. Trace whether a test double bypasses the behavior it claims to verify and whether a test can pass when the contract is broken.
3. Run focused verification when the code and test environment permit it and test side effects are understood. Choose commands for the risk, such as `go test` for affected packages, `-race` for exercised concurrency, or repeated/shuffled runs for suspected order dependence. Report the exact command and outcome. Do not treat a passing command as proof that unexercised behavior works, and do not infer that absent files or CI steps are missing when the supplied scope is incomplete.
4. Report distinct findings with code or command evidence and a concrete consequence. Group symptoms of one root cause. For each finding explain why the proposed change would improve the test signal. State uncertainty as a limit, not as a defect.

Read [Testing decisions](references/testing-decisions.md) when a judgment depends on Go version, table-driven tests, test package choice, cleanup, parallelism, integration isolation, concurrent timing, fuzzing, or benchmarks. It records important exceptions and links to current Go documentation.

## Review questions

| Area | What to establish |
| --- | --- |
| Behavior and assertions | Do tests exercise the changed success, failure, boundary, and compatibility behavior that matters? Would an intentionally wrong result make them fail? Are errors checked at the promised level, including `errors.Is`/`As` when part of the contract? Is coverage used to locate gaps rather than treated as a quality score? |
| Test structure | Do cases make distinct expectations understandable and failures diagnosable? Use table-driven subtests when they reduce repetition; a focused standalone test may be clearer. Check effective language version before calling a loop-variable capture safe or unsafe. |
| Doubles and real boundaries | Do fakes or mocks model the dependency contract without mirroring implementation calls? Where correctness depends on SQL, HTTP, files, processes, or serialization, is an appropriately realistic boundary exercised? `httptest`, in-memory components, and ephemeral services are options, not universal requirements. |
| Isolation and lifecycle | Can tests run alone, in any order, and under their intended parallelism? Are temp files, environment, servers, transactions, goroutines, and other resources owned and cleaned up? Check whether rollback actually covers the connections and commits used by the code under test. |
| Concurrency and timing | Do async tests wait for an observable condition with a bounded failure path instead of relying on a guessed sleep? Are shared test state and `t.Parallel` compatible? Does race detection exercise the relevant path? Use `testing/synctest` when its semantics fit and the effective Go version supports it. |
| Fuzzing and benchmarks | For parser or other broad-input risks, would seed cases and deterministic fuzz properties add meaningful coverage? If the change includes benchmarks, do they measure the intended work without setup contamination or compiler-elided results? Treat measured performance conclusions as Performance findings. |

Testing assesses verification quality, including whether a relevant behavior has meaningful assertions and exercised concurrency. Describe any observed production defect as an ungraded related finding and assess its independently actionable test gap here. Deployment & Operations primarily assesses CI enforcement; Security assesses vulnerabilities; Performance assesses measured cost. Benchmark validity remains in scope here. Do not require a framework, table format, external test package, build tag, or CI job without a demonstrated testing consequence.

## Grade testing

Classify each **distinct, substantiated** issue by its effect on regression detection in this codebase. A minor issue reduces clarity or diagnostic value locally; a moderate issue leaves a plausible normal-use behavior insufficiently checked or makes tests meaningfully fragile; a major issue leaves an important contract effectively unverified or makes verification of that important contract misleading; a critical issue lets severe systemic failure evade the stated verification purpose. A merge recommendation, if requested, is separate from the grade. Count root causes, not occurrences. Lack of tests is graded against demonstrated behavioral risk, not a universal test-count target. For a localized broken or misleading test, use moderate when the evidence establishes lost or fragile coverage but does not establish the importance of the affected behavior. Use major when the supplied contract or traced consequences establish that importance; the mere fact that assertions never run is insufficient. Record missing impact context in Limits.

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

If the target has no behavior or verification decision to assess, return **Not applicable**. If the topic is relevant but essential implementation, tests, or contract evidence is unavailable after reasonable inspection and would change the judgment, return **Insufficient evidence**. Explain what is missing and what would permit grading. A provided full changeset with no relevant tests is assessable; a partial excerpt with no test files is not evidence that the repository lacks tests. Do not use either state to avoid a supportable judgment.

## Report

Create a review record using the shared [record contract](../go-quality-report/references/review-record.md), then run its renderer. Reconcile candidate evidence and causes before assigning topic severity; record missing material evidence as unresolved. Use the rendered report for the saved artifact and final response unless the user requests another format. The shared helper is bundled with this plugin; other topic audits are not prerequisites.
