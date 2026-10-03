---
name: go-testing
description: Review and grade testing and verification for a defined Go changeset or code area across libraries, CLIs, and services. Use for test coverage of behavior, assertions, isolation, integration boundaries, concurrency tests, fuzzing, or benchmark validity; not for a general Go style, security, or performance audit.
metadata:
  review-contract: "1"
---

# Go Testing Review

Review whether the available tests and verification would detect meaningful regressions in the specified Go code. Give one Testing grade based on observable evidence and the risk of the changed behavior. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Assess the relevant questions and record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
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

Count independently actionable root causes after grouping repeated symptoms. Two symptoms are one issue when the same causal correction resolves both; unrelated corrections remain separate even in one function. State the counted severities in the report. Verified strengths do not cancel defects, and optional preferences or unconfirmed concerns do not enter the counts.

A contained failure of one important function, target, or contract is normally major. Use systemic major only when one shared cause demonstrably affects multiple distinct important boundaries or workflows; repetition across files alone is insufficient. Critical requires supported severe, broad, or irreversible impact in the topic being reviewed, not merely the words race, unsafe, missing test, or build failure. Explain any systemic or critical classification. These reach rules refine the topic-specific severity definitions above.

For A+, identify two independent safeguards, the distinct meaningful risks they control, and evidence that each works. Ordinary correct setup supports A. Complexity, tool count, or splitting one safeguard into two descriptions does not earn A+; simple designs can qualify when their risk control is demonstrated. If no issue is found but the evidence cannot support even one relevant strength, use Insufficient evidence.

Use Not applicable only when no relevant decision is implicated. Use Insufficient evidence when the topic applies but essential evidence prevents a judgment. A confirmed defect can still receive a letter grade for the inspected scope, with the unassessed parts in Coverage and Limits; do not imply those parts passed. Do not award A or A+ for an incompletely assessed material risk.

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

Use the template below. For a letter grade retain every field and section; write `None verified`, `None found`, or `None needed` when appropriate. For Not applicable or Insufficient evidence, retain Scope, Coverage, Rationale, and Limits; omit counts and unsupported Good/Bad/changes. Reference each Bad finding by ID in Suggested changes. Put optional follow-ups or ungraded related findings after the required sections and label them explicitly.

```markdown
## Testing — [grade | Not applicable | Insufficient evidence]
Scope: [base/head, supplied diff, or code area; library/CLI/service and relevant version/runtime context]
Coverage: [applicable areas assessed; material areas not assessed and why]
Rationale: [why the evidence and counted severities select this grade; explain systemic/critical reach or both A+ safeguards]
Finding counts: critical=[n], major=[n], moderate=[n], minor=[n]

Good

- [G1] [specific choice and evidence/location] — [why it benefits the relevant contract or risk]

Bad

- [F1][severity][introduced | worsened | existing-in-scope] [issue and evidence/location] — [trigger and supported consequence]

Suggested changes

- [F1] [targeted correction] — [why it resolves the consequence; how to verify it]

Limits: [material missing evidence; exact checks and outcomes, or checks not run]
```

Prefer `path:line` references when files are available. Give each bad finding a corresponding change and rationale. Positive findings need evidence too. The grade assesses Testing in the requested scope, not overall Go quality.
