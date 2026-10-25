---
name: go-dependencies-and-reproducibility
description: Review and grade dependency selection, Go module integrity, toolchain and generation inputs, and repeatable builds for a defined Go changeset or code area. Use across libraries, CLIs, and services when modules, imports, workspaces, build tools, or build inputs change; not for a general vulnerability, runtime deployment, or test-quality audit.
metadata:
  review-contract: "2"
---

# Go Dependencies and Reproducibility Review

Review whether the specified Go source resolves the intended dependencies and can be built again from the stated source and build inputs on its supported targets. Give one Dependencies & Reproducibility grade based on the actual module and build context. Review only; do not edit files unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Assess the relevant questions and record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact changeset or code area and whether the deliverable is a reusable module, CLI, service, or several artifacts. Compare the changed `go.mod`, `go.sum`, `go.work`, imports, generation commands, build scripts, and relevant CI or release settings with the prior revision. Distinguish introduced issues from existing constraints. Do not infer that a file omitted from a partial excerpt is missing from the repository.
2. Establish the **resolution context**: main module or workspace; intended standalone consumer or release checkout; selected Go toolchain; supported Go versions and `GOOS`/`GOARCH`; build tags, `CGO_ENABLED`, native libraries, and relevant module environment. A passing workspace or warm-cache build does not establish that a clean standalone build works. Verify both modes when both matter.
3. Trace the **selected** graph and inputs, not just changed lines. For a changed module, inspect the effective version under minimal version selection, transitive requirements, `replace`/`exclude`/`retract`, and whether the import is used by production code, tests, or tooling. Establish whether new checksums, vendored content, local modules, generated files, or private-module access are needed by the documented build. Do not call every indirect change, old version, or absent `go.sum` an issue.
4. Use focused checks in a disposable copy when they may write sums, download inputs, or produce artifacts. Diagnostic choices include `go list -m all`, `go mod graph`, `go mod why -m`, `go mod verify`, and a build or test under the actual vendor/workspace/toolchain mode. For standalone non-vendored builds, `-mod=readonly` prevents required `go.mod` edits; it is not a blanket filesystem write guard. Do not override a shipped vendor build with readonly mode and then claim the vendor path passed. Use `go mod tidy -diff` only when the selected toolchain supports it. Prefer a clean checkout or isolated cache to test fresh resolution when the environment permits. Report exact commands, results, and which context they covered; do not claim a clean build, offline build, checksum verification, or byte-identical binary without actually checking it. Avoid `go get`, `go mod tidy` without `-diff`, or commands that rewrite module files during a review.
5. Report distinct root causes with file or command evidence and the affected target or consumer. Treat an unverified concern as a limit or question, not a grade-lowering defect. For each defect, propose the smallest correction that restores the stated dependency or rebuild contract; separate a required fix from optional hardening.

Read [Dependency and reproducibility decisions](references/dependency-reproducibility-decisions.md) when a judgment turns on `go.sum`, module selection, workspace or replace behavior, toolchain selection, private modules, vendoring, generators, or artifact identity. Check current primary Go documentation and the project's supported versions for version-sensitive claims.

## Review questions

| Area | What to establish |
| --- | --- |
| Selected modules | Does the committed module configuration select the intended versions on supported targets? Do direct, indirect, test, and tool dependencies resolve without an unexplained upgrade, accidental downgrade, or path mismatch? Check actual selection and use; the `// indirect` marker does not mean a module is absent from every build. |
| Integrity and availability | Are needed public module hashes available through `go.sum` or the configured checksum database for the build mode? Are private modules reachable through the stated authenticated path? Does a vendored tree, if used, match `go.mod`? Are local `replace` targets present in the distributable checkout? Judge the configured network and offline requirements, not a universal vendor or checksum policy. |
| Go and target versions | Does each supported build use a toolchain that accepts the module's `go` requirement and its dependencies? Does a `toolchain` suggestion or automatic download actually apply in that environment? Do build constraints, cgo, and native inputs resolve for the targets the project promises? |
| Generation and build inputs | If generated code or assets matter, can the stated source, generator version, and command reproduce them when needed? Are build-time downloads or floating versions controlled where they affect the shipped output? For a bit-for-bit reproducibility claim, are relevant flags, VCS state, paths, timestamps, cgo toolchain, and packaging inputs specified or controlled? |
| Change discipline | Can a fresh checkout perform the documented build without uncommitted module edits or hidden local state? Are graph changes reviewed for their actual effect, including support-policy changes, rather than accepted solely because `go mod tidy` or a warm local build passes? |

Dependencies & Reproducibility assesses selected graphs, toolchains, generation, and build-input resolution. Correctness & Compatibility primarily assesses behavior after those inputs are selected; Security assesses vulnerability; Architecture assesses runtime dependency boundaries; Deployment & Operations assesses artifact delivery and runtime wiring; Testing assesses regression signal. A supported build failure remains assessable here in solo use. Do not require newest versions, vendoring, a particular proxy, or byte-identical binaries without a demonstrated contract and consequence.

## Grade dependencies and reproducibility

Classify each **distinct, substantiated** issue by the reach of its dependency or rebuild failure. A minor issue is localized friction with a low-impact, repeatable workaround; a moderate issue makes a plausible development or regeneration path drift or fail while current shipped builds remain repeatable; a major issue breaks an important supported build, module consumer, or declared artifact path, or leaves a material build input uncontrolled; a critical issue causes severe, broad loss of ability to reproduce the core deliverable or makes essential production dependency contents untraceable across the release path. An important build failure can still be **major** when its demonstrated reach is one target or consumer. Reserve F for evidence of critical reach, not a missing preferred file or a single failing command. Explain the target, prerequisites, and impact. A merge recommendation, if requested, is separate from the grade. Count root causes, not every affected package or checksum line.

Apply the first matching row from the top. These anchors match the other Go review skills; explain context-dependent severity.

### Apply the anchors consistently

Admit a finding only when evidence connects an in-scope cause to a concrete consequence under supported use or ordinary maintenance. Trace relevant callers, guards and rule owners. Source tracing can suffice without an executed reproduction. If the consequence depends on an unverified assumption, keep it ungraded and name the missing evidence; uncertainty is not minor severity.

For maintenance findings, name the ordinary change and the avoidable work or mistake mechanism imposed by the current design. File/line counts, missing abstractions and hypothetical new capabilities do not establish that cost. Local reading friction or a small repeated edit is minor. Moderate needs a concrete failure mechanism or substantial coordination/reimplementation cost; explain what makes it substantial beyond the number of sites. Check whether apparent duplication implements the same rule or distinct responsibilities.

Count independently actionable causes: merge symptoms resolved by one causal correction; retain unrelated corrections. Strengths do not cancel defects. Judge each topic's consequence independently; another topic's severity is not evidence.

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

If the target changes no dependency or build input and presents no dependency or rebuild decision to assess, return **Not applicable**. If the topic is relevant but essential module, build, target, or release context is unavailable after reasonable inspection and could change the judgment, return **Insufficient evidence**. Explain what is missing and what would permit grading. A confirmed failure can still be graded when other paths are unknown; do not use either state to avoid a supportable judgment.

## Report

Use the fixed report structure below in both the saved report and the final user-facing response, unless the user explicitly requests another format. Keep the section names and order; do not replace the findings table with prose or only a link to an artifact. Include every counted finding in the table, ordered by severity (critical to minor), then numeric finding ID. Use the same IDs, severities and topics in the table and details. Use the topic display names from the skills; separate multiple topics with `; ` in overall reports. A standalone topic report lists only its own topic. Keep table cells concise, escape literal pipes as `\|` (including inside code), and put multiline evidence in the details.

Every finding has one detail entry with the fields shown, including a correction and proposed verification. Distinguish proposed verification from executed checks. For zero findings retain the empty table header and separator, followed by `None found.`; do not invent an F0 row. Use `None found.` for empty details and `None verified.` for unsupported strengths. For Not applicable or Insufficient evidence retain the same sections and explain the state; omit counts when no findings are confirmed. Preserve any confirmed findings under Insufficient evidence without implying the remaining scope passed. Counts, when present, describe confirmed findings only. Keep optional suggestions and ungraded related observations outside the counted table, after the required sections, labeled explicitly.

A topic title identifies that topic's grade; it does not imply an overall review. Full reviews use `Go Quality Report` as the title and append the report-card table and maintainability profile in that order. Grade/state cells contain only a letter grade, `Not applicable`, or `Insufficient evidence`; put qualifications in the coverage column. Apply the existing coverage, severity and grading rules unchanged.

```markdown
# Dependencies & Reproducibility — [grade | Not applicable | Insufficient evidence]
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

Prefer `path:line` references when files are available. Give each finding a corresponding change and rationale. The grade assesses Dependencies & Reproducibility in the requested scope, not overall Go quality.
