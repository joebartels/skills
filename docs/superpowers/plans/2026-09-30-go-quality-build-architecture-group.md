# Go Quality Build Architecture Group Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship and evaluate three focused Go writing skills for package boundaries, API contracts, and interfaces and composition.

**Architecture:** Add a peer `go-quality-build` plugin with one canonical runtime copy of each skill for Codex, Claude Code, and OpenCode. Promote one skill at a time only after source audit, selection checks, realistic code-change comparisons, and independent Go review. Keep evaluation fixtures and results outside the plugin.

**Tech Stack:** Markdown skills and references, JSON manifests and evaluation cases, Python 3 `unittest` and `scripts/validate.py`, Go fixtures and toolchain, available harness validators.

**Spec:** [First architecture group design](../specs/2026-09-30-go-quality-build-architecture-group-design.md)

## Global Constraints

- Runtime package: `plugins/go-quality-build/`; evaluation package: `tests/go-quality-build/`. Do not put evals in the installable plugin.
- Skill IDs and order: `go-package-boundaries`, `go-api-contracts`, `go-interfaces-and-composition`.
- Use upstream `samber/cc-skills-golang` commit `19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`; record copy/adapt/omit decisions and retain MIT notice for substantial copying.
- Maintain distinct decision owners, context-sensitive guidance, one canonical runtime copy per skill, and the separate `go-quality-review` package without grade-driven edits.
- Do not publish or install the package globally during development. Record unavailable behavioral or harness checks rather than claiming they passed.
- Update `docs/go-quality-build/README.md` status and work log after each task and before stopping.

## Review Focus

These five conditions must be exercised by the owning task's evaluation cases, not merely mentioned in prose:

1. A small CLI or library with cohesive behavior should not acquire speculative package layers (Task 2, `small-cli` and `small-library`).
2. A service adding another ingress or worker, storage, and an outbound client should split only for demonstrated coupling, with adapters pointing toward core behavior (Task 2, `medium-service`).
3. A public Go API whose changed call syntax still compiles may break function-value consumers or external implementers (Task 3, `source-compatibility`).
4. A producer-owned protocol interface may be appropriate, while a test-only interface may add needless abstraction (Task 4, `protocol-vs-mock`).
5. A multi-feature project may need feature-local ingress, jobs, and integrations plus a genuinely shared capability, without assuming that size alone requires new modules (Tasks 2 and 5, `large-features` and `combined-evolution`).

---

## File map

- `docs/go-quality-build/source-audit.md`: pinned upstream source inventory and copy/adapt/omit decisions, owned by the skill that uses each section.
- `plugins/go-quality-build/plugin.json`, `.claude-plugin/plugin.json`, `README.md`, and `skills/<id>/SKILL.md`: portable package and runtime guidance.
- `plugins/go-quality-build/skills/go-package-boundaries/references/package-maps.md`: small, medium, and large package maps across library, CLI, service, and worker contexts, with triggers and import direction.
- `.agents/plugins/marketplace.json`, `.claude-plugin/marketplace.json`, `opencode.json`, root `README.md`: make the peer package discoverable across harnesses.
- `tests/go-quality-build/<id>/evals/evals.json` and `evals/files/<case>/`: selection and code-change tasks with fixture repos; keep expected judgments outside model prompts.
- `tests/go-quality-build/results/<run-id>/`: baseline and skill-on artifacts, independent review and a concise comparison; record exact harness, model, revisions, checks and limits.
- `tests/go-quality-build/test_evals.py` and `scripts/validate.py`: structural validation of build-skill eval inputs; keep the existing review contract checks intact.
- `docs/go-quality-build/README.md`: canonical progress record and continuation protocol.

### Task 1: Pin source decisions and baseline cases

**Files:** Create `docs/go-quality-build/source-audit.md`; create `tests/go-quality-build/go-package-boundaries/evals/evals.json` and fixture files under `evals/files/`; create baseline artifacts under `tests/go-quality-build/results/<run-id>/`; modify `docs/go-quality-build/README.md`.

**Interfaces:** Produces a source audit with section, pinned URL, copy/adapt/omit choice, local owner, and reason. Produces evaluation case IDs `small-cli`, `small-library`, `medium-service`, `large-features`, `not-package-work`, each with `prompt`, `expected_output`, `assertions`, and optional `files` paths relative to its `evals/` directory.

- [ ] **Step 1: Write the source audit.** Inventory upstream project-layout, design-patterns, structs-interfaces, dependency-injection, and naming sections at the pinned commit; compare each retained decision with the local Architecture and Correctness references and primary Go docs. Record exact URLs and license handling.
- [ ] **Step 2: Create package-boundary eval fixtures and expectations.** Use a small CLI doing parsing and file work, a cohesive public library, a service gaining another ingress or worker plus storage and an outbound client, a large multi-feature project, and an unrelated local calculation. Include `go.mod` and only the code needed for each decision. Assertions must check responsibilities and dependency direction, not prescribed directory names alone.
- [ ] **Step 3: Run baseline model changes without the new skill.** With expected judgments withheld, save prompts, code diffs, `rtk go test ./...` output for runnable fixture workdirs, and independent Architecture/Correctness review findings under one dated run ID. State which harness/model ran and any case that could not run.
- [ ] **Step 4: Verify the audit and fixture changes.** Run `rtk git diff --check`; inspect every fixture referenced by the JSON cases and record baseline limits in the design record. This task does not claim skill uplift.
- [ ] **Step 5: Commit.** Commit the audit, fixtures, baseline artifacts, and record with `rtk git commit -m "test: establish Go package-boundary baselines"`.

### Task 2: Package-boundary skill and portable plugin

**Files:** Create `plugins/go-quality-build/plugin.json`, `.claude-plugin/plugin.json`, `README.md`, `skills/go-package-boundaries/SKILL.md`, and `skills/go-package-boundaries/references/package-maps.md`; modify both marketplace JSON files, `opencode.json`, root `README.md`, `scripts/validate.py`, `docs/go-quality-build/README.md`; create `tests/go-quality-build/test_evals.py` and package skill-on results under the Task 1 run ID.

**Interfaces:** Add `validate_build_evals(root: Path) -> tuple[list[str], int]` to `scripts/validate.py`; call it from `main()` without changing `validate(root)` or the existing review contract. Each build skill has a matching nonempty `tests/go-quality-build/<skill-id>/evals/evals.json`; validate unique case IDs, nonempty prompts/expectations/assertions, and fixture paths contained in that skill's `evals/` directory.

- [ ] **Step 1: Write failing validator tests.** In `tests/go-quality-build/test_evals.py`, assert a valid suite passes and missing suites, duplicate IDs, empty assertions, and a `../` fixture escape fail.
- [ ] **Step 2: Run the tests to confirm the gap.** Run `rtk python3 -m unittest discover -s tests/go-quality-build -p 'test_*.py'`; expect failure because `validate_build_evals` is absent.
- [ ] **Step 3: Implement build-eval validation.** Add the specified function and main integration; rerun the same unittest command and expect PASS without changing the review contract checks.
- [ ] **Step 4: Add first-skill packaging.** Add matching `0.1.0` manifests, marketplace/OpenCode entries, README guidance, and `go-package-boundaries` with a trigger that distinguishes package work from local expression edits.
- [ ] **Step 5: Add the package-map reference.** Show small CLI and library maps, a medium service with ingress, worker, storage, and outbound integration, and a large multi-feature map. Include the component palette from the spec, representative split triggers, import direction, composition, and a counterexample that stays in one package. Distinguish files, packages, ports, and adapters; explain that names and size do not mandate a layout.
- [ ] **Step 6: Run selection and skill-on cases.** Re-run Task 1 cases with this skill available, withholding expectations. Compare code changes with baseline, run `rtk go test ./...` in runnable fixture workdirs, and obtain independent Architecture/Correctness review. Revise harmful instructions and rerun affected cases. If no distinct benefit is shown, do not promote; record the result and revise the plan.
- [ ] **Step 7: Verify the package.** Run the build-eval unit tests, `rtk python3 scripts/validate.py`, `rtk git diff --check`, and available `rtk claude plugin validate ./plugins/go-quality-build`; check Codex/OpenCode loading when those harnesses are available. Record exact results and limits in the design record.
- [ ] **Step 8: Commit.** Commit the package, validator, cases, evidence, and record with `rtk git commit -m "feat: add evaluated Go package-boundaries skill"` only if the promotion gate passed.

### Task 3: Public API contracts skill

**Files:** Create `plugins/go-quality-build/skills/go-api-contracts/SKILL.md`, `tests/go-quality-build/go-api-contracts/evals/evals.json` and needed fixtures; update `docs/go-quality-build/source-audit.md`, results under a new run ID, and the design record.

**Interfaces:** Evaluation case IDs `source-compatibility`, `intentional-break`, `wire-contract`, `not-public-contract`. Skill owns consumer-visible shape and evolution, including release policy and migration; it defers package placement and interface necessity to the other skill owners.

- [ ] **Step 1: Write baseline cases.** Include function-value assignment or external interface implementation, a documented intentional breaking release, an observable JSON/CLI contract, and an internal-only edit.
- [ ] **Step 2: Run baseline cases without the new skill.** Save diffs, `rtk go test ./...` results in runnable fixture workdirs, and independent Correctness/Architecture review with expectations withheld.
- [ ] **Step 3: Author the minimal runtime skill.** Use audited upstream material where it helps; inspect consumers, versions, and policy before changing an API, test the actual contract, and avoid an automatic promise of backward compatibility for every module.
- [ ] **Step 4: Run selection and skill-on cases.** Compare with baseline, verify compilable consumers and behavior, independently review the changes, and revise/rerun any harmful guidance. Retain the skill only if it adds distinct value.
- [ ] **Step 5: Verify.** Run `rtk python3 -m unittest discover -s tests/go-quality-build -p 'test_*.py'`, `rtk python3 scripts/validate.py`, applicable Go checks, `rtk git diff --check`, and available harness checks; update the record with findings and limits.
- [ ] **Step 6: Commit.** Commit the skill, cases, evidence, and record with `rtk git commit -m "feat: add evaluated Go API-contracts skill"` only if the promotion gate passed.

### Task 4: Interfaces and composition skill

**Files:** Create `plugins/go-quality-build/skills/go-interfaces-and-composition/SKILL.md`, `tests/go-quality-build/go-interfaces-and-composition/evals/evals.json` and fixtures; update the source audit, results under a new run ID, and the design record.

**Interfaces:** Evaluation case IDs `protocol-vs-mock`, `visible-dependency`, `zero-value`, `library-lifecycle`, `not-abstraction-work`. Skill owns interface necessity, dependency wiring, construction, and lifecycle, while API compatibility and package placement remain with their owners.

- [ ] **Step 1: Write baseline cases.** Include a legitimate producer protocol, a fake-only interface proposal, hidden runtime dependency, useful zero value, host-owned shutdown, and an unrelated pure helper.
- [ ] **Step 2: Run baseline cases without the new skill.** Save code changes, `rtk go test ./...` results in runnable fixture workdirs, and independent Architecture/Testing review.
- [ ] **Step 3: Author the minimal runtime skill.** Adapt audited upstream guidance to real consumer/protocol needs, visible dependency ownership, optional constructors, and lifecycle control; reject universal DI or functional-options rules.
- [ ] **Step 4: Run selection and skill-on cases.** Compare with baseline, verify behavior and absence of unnecessary abstractions, independently review results, and revise/rerun harmful guidance. Retain only if it adds distinct value.
- [ ] **Step 5: Verify.** Run `rtk python3 -m unittest discover -s tests/go-quality-build -p 'test_*.py'`, `rtk python3 scripts/validate.py`, applicable Go checks, `rtk git diff --check`, and available harness checks; update the record with evidence and limits.
- [ ] **Step 6: Commit.** Commit the skill, cases, evidence, and record with `rtk git commit -m "feat: add evaluated Go interfaces and composition skill"` only if the promotion gate passed.

### Task 5: Combined consistency and delivery review

**Files:** Create `tests/go-quality-build/combined/evals/evals.json`, a `combined-evolution` fixture under `evals/files/`, and combined run results; modify `plugins/go-quality-build/README.md`, root `README.md`, and the design record as needed.

**Interfaces:** Combined task invokes all three skills on a two-feature service adding an ingress path, a background worker, and an external integration while preserving a documented caller contract. The outcome must retain acyclic dependency direction, justified interfaces, and consumer compatibility. Extend `validate_build_evals(root)` to validate `tests/go-quality-build/combined/evals/evals.json` with the same case and fixture rules, while still requiring a suite for every installed build skill.

- [ ] **Step 1: Add the combined case and validator test.** Create `combined-evolution` with a runnable fixture and a unit test proving the combined suite is validated.
- [ ] **Step 2: Run the combined task with expectations withheld.** Save selected skills, code diff, `rtk go test ./...` output, and independent Architecture/Correctness/Testing review. Check for conflicting advice, duplicated layers, and missed consumer breakage.
- [ ] **Step 3: Resolve any conflict at its owner.** Change only the responsible skill and rerun the failing case plus its adjacent single-skill case; document the before/after consequence.
- [ ] **Step 4: Verify delivery.** Run `rtk python3 -m unittest discover -s tests -p 'test*.py'`, `rtk python3 scripts/validate.py`, `rtk python3 tests/go-quality-review/go-quality-report/evals/test_grade.py`, `rtk git diff --check`, and available Claude/Codex/OpenCode package checks. Check every runtime skill link and the plugin README's activation guidance.
- [ ] **Step 5: Record conclusions and commit.** In the design record, distinguish implemented, structurally valid, behaviorally evaluated, and unverified claims. Commit the combined evidence and final corrections after an independent whole-package review.
