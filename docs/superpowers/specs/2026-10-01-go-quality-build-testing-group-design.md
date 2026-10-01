# Go quality build: testing group

## Status and intent

Approved by the user on 2026-10-01. The implementation plan was subsequently accepted for native execution with independent behavioral authors/reviewers. This spec does not claim that either skill exists or is effective.

Continue the multi-harness Go writing collection with `go-behavior-tests` and `go-test-isolation`. The intended user is an agent changing a Go library, CLI, service or worker. Success means its tests detect meaningful regressions and observe the promised behavior reliably, with less unnecessary setup, abstraction and dependency cost.

The [canonical record](../../go-quality-build/README.md) identifies this group as the next priority. The first group's [blind combined run](../../../tests/go-quality-build/results/2026-10-01-combined/trial-summary.md) earned Architecture A, Correctness B and Testing C-. Four test mutations survived: removing the host's worker join, changing CLI failure to exit successfully, removing blank-text validation and publishing through direct truncation. Composition trials also missed discarded host cancellation and incorrect recurrence timing. These are observed test-signal gaps; they do not make every corresponding production behavior defective. The separate assisted A/A/A repair remains historical assisted evidence.

## Approach and alternatives

Use two focused skills with one owner for each decision. `go-behavior-tests` selects the behavior, cases, observation boundary and assertions; `go-test-isolation` selects dependencies, fixtures, scheduling controls and cleanup that make the observation trustworthy. A task may need either or both.

A single broad testing skill would be simpler to select, but would load fixture and concurrency advice into small pure-function tasks and obscure which guidance improved an outcome. A larger group including concurrency, context and error-contract writing would cover more production decisions, but would weaken attribution and increase overlap. Start with the two testing skills; merge them if realistic trials cannot establish distinct utility. Keep future production-concurrency and error skills proposed.

## Packaging and deliverables

Extend the existing `plugins/go-quality-build/` peer package using its canonical runtime layout and established Claude, Codex and OpenCode manifests/catalogs. Preserve the separate review collection. No new coordinator, global harness installation or test-framework dependency is required.

Proposed runtime artifacts are:

- `skills/go-behavior-tests/SKILL.md`, with a focused `references/behavior-observations.md` only for examples that need more explanation than the main skill.
- `skills/go-test-isolation/SKILL.md`, with `references/isolation-patterns.md` for fixture lifetime, dependency fidelity and async observation examples.

Author draft runtime files outside the installable `skills/` tree, under `docs/go-quality-build/drafts/<skill-name>/`, until the promotion gate passes. Run draft validation and trials against those exact bytes. Promote the evaluated snapshot by copying it into the runtime tree and verifying its hashes. Only then update package descriptions, discovery metadata and user documentation. Do not expose a merely proposed skill as installed.

Keep suites under `tests/go-quality-build/<skill-name>/evals/` and dated evidence under `tests/go-quality-build/results/`. Document installed versus draft versus evaluated status separately. The existing validator already accepts candidate suites without installed runtime skills; extend it only if a concrete new schema or gate needs it.

## Skill responsibilities

### `go-behavior-tests`

Trigger when a Go task adds or changes meaningful behavior, repairs a defect, or adds/strengthens tests for a supported contract. Skip comment-only, formatting-only or mechanical edits with no behavioral change and no requested test improvement. Do not turn every trivial edit into a test expansion.

Start from the task, project contracts, callers and existing tests. Trace the changed path to its observable result: returned values/errors, state and side effects, process status, wire/file representation or lifecycle completion. Identify consequential success, rejection, boundary and partial-failure cases. Choose an oracle from the promised behavior rather than calculating the expected value by duplicating the implementation under test.

Own these decisions:

- Check complete meaningful results, supported error identities/types when promised, relevant side effects and their absence on failure. A non-nil result or absence of panic rarely proves a richer contract.
- Exercise the boundary that makes the claim true. A `run` helper result cannot establish actual CLI exit status; a worker unit test cannot establish the host's stop/join/release sequence. Use a process or host-level check when that is the contract being changed.
- Include accepted-prefix and retained-state checks where failure may follow successful work. Include actual success/recurrence paths for workers; cancellation-only tests do not exercise recurrence.
- Choose table-driven subtests when cases share setup; use distinct tests when setup or behavior differs. Choose external or same-package tests according to access and the contract, without a universal package/file naming rule.
- Use executable examples for requested runnable usage. Use property/fuzz tests when a useful invariant and broad input risk justify them. Round trips alone can miss matching bugs on both sides; include independent contract observations where needed. Do not require fuzzing for every parser or every edit.
- Ask what plausible regression each important test would catch. For a bug fix, check that the relevant pre-fix behavior fails when practical. Use a small targeted mutation or controlled counterexample for an ambiguous assertion, with a time budget; do not demand a mutation campaign for every task.

Coverage locates unexecuted paths; it does not replace assertion review. Report the behavior verified, meaningful checks actually run, and material untested boundaries. Isolation mechanics belong to the neighboring skill.

### `go-test-isolation`

Trigger when Go test work involves shared mutable/process state, files, external dependencies, test doubles, goroutines, timers, fixture lifetime or parallel execution. Also apply when repairing flaky or order-dependent test setup. Skip pure deterministic calculations with local inputs and ordinary assertions when none of those choices changes. This is an authoring skill; it does not replace a systematic investigation of an unexplained failure.

Inspect what the test owns, borrows and shares, including subtests and work that may outlive the test body. Pick the smallest controlled fixture that preserves the behavior being asserted. Keep repeatability and dependency fidelity together: replacing a real boundary with an easier double must not erase the claim being checked.

Own these decisions:

- Use temporary directories and explicit restoration/cleanup for owned state. Account for parent/subtest lifetimes. Cancel and join started work before deleting or closing resources it may still use, including setup and assertion failure paths. Bound waits so a regression fails with useful diagnostics.
- Keep process-global environment, current directory and global-default mutation serial or isolate it in a child process. `t.Setenv` restores values but does not make concurrent process-state mutation safe. Parallelize only when resources and dependency behavior are independent.
- Choose real dependencies, standard-library local fixtures, handwritten fakes, function seams or existing mocks from the boundary at risk and project conventions. No mandatory mocking library, goleak dependency, Docker service or integration build tag. A transaction rollback only cleans up work performed through that transaction; it does not isolate independently committed writes.
- Preserve relevant dependency semantics: cancellation, failure, response/body lifetime, ordering and side effects. Use real protocol/file/process checks where an in-memory double would hide the contract. A `ResponseRecorder` can establish handler behavior, but does not establish network behavior. Disclose reduced scope if an environment restriction forces a substitute.
- Observe async startup, blocking, cancellation, completion and resource release through explicit events. Avoid sleeps that guess readiness and immediate negative checks that race an unstarted observer. When elapsed time itself is contractual, measure from the correct event and justify tolerances. Use `testing/synctest` only when the supported Go version and in-process behavior fit; real processes and external I/O need other controls.
- Report worker outcomes to the test goroutine before failing the test. Use applicable race, individual-test, shuffle and repeated checks to probe isolation. Passing these checks is limited to exercised paths and schedules; it does not prove the absence of races or flakes.

A test seam may be a small internal runner or an existing dependency boundary. This skill must not require production interfaces, exported hooks, constructors or clock frameworks solely to satisfy a preferred test pattern.

## Ownership and compatibility

| Decision | Owner |
| --- | --- |
| Supported inputs, results, process/wire behavior and release compatibility | Existing `go-api-contracts`; behavior tests verify that contract |
| Package placement and import direction | Existing `go-package-boundaries` |
| Production dependency/abstraction and host lifecycle design | Existing `go-interfaces-and-composition` |
| Which behavior and regression the test must observe | `go-behavior-tests` |
| How the test controls dependencies, state, timing and cleanup | `go-test-isolation` |
| Production cancellation, synchronization, error exposure and performance policy | Future topic owners; these skills cover verification choices only |
| Independent findings and grades | Unchanged Go quality review skills |

The main skills must be useful independently. Brief adjacent-owner notes must not require loading every skill or duplicate their full guidance. Runtime references remain local and portable; repository-only audit/evaluation material is not a runtime dependency.

Respect the module/file's effective Go version and supported toolchain/platforms. Baseline examples should remain usable in Go 1.22 projects. Optional `t.Context` examples require Go 1.24; stable `synctest.Test` requires Go 1.25. A newer host compiler alone does not establish those APIs are supported by the project. Show a compatible cancellation/synchronization alternative rather than raising a project's minimum to use a preferred helper. Older-module parallel loop captures must be assessed against the applicable loop semantics.

## Source policy

Keep upstream `samber/cc-skills-golang` pinned to `19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`. A new `docs/go-quality-build/testing-source-audit.md` must inventory `golang-testing/SKILL.md` and all seven references: benchmarks, coverage, examples, helpers, HTTP testing, integration testing and mocking. Inventory relevant testing-layout and naming/testing material from the previously audited families as well. Record section-level copy/adapt/omit decisions with a local owner and rationale before runtime authoring.

Preserve useful behavior examples when accurate. Adapt or omit blanket named-table, integration-tag, parallelism, source-file naming, sub-millisecond speed, consumer-interface, tool-installation and third-party-library rules. Benchmark validity and methodology remain with future performance guidance unless necessary to distinguish scope. Check version-sensitive examples and tooling claims against primary Go documentation and actual project support. Do not change the review rubric to make imported advice pass.

Copied substantial content must carry the upstream MIT notice, revision and affected paths in the plugin. New wording/examples must be identified as original; a source audit link is not a substitute for a distributed license notice.

## Behavioral evaluation design

Create realistic fresh tasks before authoring either runtime draft. Contracts belong in ordinary task/project inputs; expected judgments and mutation recipes remain outside author input. The prior combined fixture and assisted repair may be used as historical diagnostic checks, but cannot serve as unseen benefit evidence for this group.

Initial suites should cover these proposed cases; fixture design may refine them without substituting a style assertion for behavior:

| Suite/case | Task and observation risk |
| --- | --- |
| Behavior: library parser/codec | Extend an input contract; assert accepted/rejected values and independent expected representations, including a tempting insufficient round-trip property |
| Behavior: CLI partial failure | Add batch behavior; verify process status, accepted prefix and absence of later side effects |
| Behavior: service validation/publication | Refresh a stored view; verify complete values, independent invalid fields and retained prior state on failure |
| Behavior: worker host | Change lifecycle behavior; verify actual host completion/release and a successful recurrence path |
| Behavior: non-selection | Documentation/format-only change with no requested testing work; avoid irrelevant tests or new dependencies |
| Isolation: library file fixtures | Add independent instances/subtests; detect shared paths and cleanup that precedes child completion |
| Isolation: CLI environment | Test configuration in isolation; preserve process environment and actual child-process behavior without blanket parallelism |
| Isolation: service dependency | Select a controlled HTTP boundary that preserves relevant status, body and cancellation behavior; distinguish handler-only from protocol observations |
| Isolation: worker timing | Repair synchronization/cleanup while preserving behavior; exercise supplied-context cancellation and completion-relative recurrence on a supported Go version |
| Isolation: non-selection | Pure deterministic library change with local inputs; avoid fixture, concurrency or dependency machinery |

Freeze input revisions and controller-held semantic probes before trials. Use fresh independent author contexts for baseline and skill-on, with the same model, reasoning setting, task, fixture and relevant existing architecture skills in both arms. Individual suites evaluate their candidate alone: withhold the other new testing skill from both arms, even if it has already been promoted. Change only availability of the candidate testing skill. Record exact wrappers, routing descriptions, opened skill/reference bytes, prompts, patches, resulting file hashes, command outputs and model/harness identity. Description-based selection is distinct from automatic harness routing.

Independent reviewers receive anonymized original/candidate code and task contracts, without the skill, arm label, expected assertions, author report or other review. Apply `go-testing` and relevant Correctness review; use Architecture review when testability changes production structure. Reproduce plausible surviving mutations in disposable copies. A compile/vet failure or unrelated test crash is not proof that a behavioral assertion catches the intended mutation. Unsupported fixture expectations must be corrected and disclosed.

Compare unique confirmed gaps, new regressions, meaningful mutation detection, fixture/dependency/production-code cost and effort. Do not average grades, count tests or reward rule repetition. If baselines are already strong, record a no-regression result rather than claiming uplift. Repeat the improvement-bearing task after a revision or an ambiguous result. Evaluate each promoted revision across its applicable cases and non-selection control; results from older drafts do not automatically cover final bytes.

After individual promotion, use a fresh combined task with the existing architecture group. Compare four matched arms: no new testing skill, behavior only, isolation only, and both. Withhold expected outcomes in each. This checks complementary benefit, contradictions, overbuilding and whether the split remains justified. Confirmed weaknesses require revision at their decision owner and affected reruns; review-guided repair stays separate from blind evidence.

## Verification and promotion

For every fixture/candidate, run meaningful scoped Go tests and applicable vet/build/process checks. Exercise relevant concurrent paths with race detection when feasible; use targeted repetitions/shuffle for isolation claims. Check both clean and mutated code. Include the declared minimum toolchain and relevant platform when available. Record unavailable listeners, services, toolchains or platform behavior as limits, and do not equate a fake with successful real integration.

Promote one skill before expanding runtime authoring to the next. Require a clear selection boundary, distinct useful guidance, demonstrated improvement on fresh matched work, useful non-selection behavior, no unexplained harmful regression, independently reviewed outcome evidence and portable packaging. A failed or inconclusive candidate remains a draft/reference, is revised, or is merged; it must not be advertised as effective.

Use skill-creator for runtime authoring, writing-skills for baseline/skill-on evaluation, and verification-before-completion before promotion. Run the repository validator, root/build-eval tests, relevant layout checks, quick skill validation and available harness validators. Unit-test validator code only when that code changes. Structural success is separate from automatic routing or runtime harness loading. Preserve raw evidence, even when archived diagnostic whitespace needs a narrowly disclosed exception.

## Sequence and continuation

After written-spec review, write a separate implementation plan covering source audit, fixture/probe preparation, blind baselines, behavior-skill draft/evaluation/promotion, isolation-skill draft/evaluation/promotion, combined evaluation and delivery review. Implementation execution method is selected at the plan handoff. This spec is not that plan.

Update the canonical design record after each meaningful stage with exact artifacts, verified results, decisions, remaining limitations and the next action. Keep the earlier blocked zero-value composition evaluation separate; this group does not retry that rejected action or treat the missing result as passed.

## Primary documentation checked for this design

Checked 2026-10-01; recheck details that become runtime advice during authoring:

- [Go testing package](https://pkg.go.dev/testing): test/subtest lifetime, cleanup, process-global environment, supported APIs and test-goroutine failure reporting.
- [testing/synctest](https://pkg.go.dev/testing/synctest): stable API version, in-process time and goroutine scope, and external-I/O limits.
- [Go race detector](https://go.dev/doc/articles/race_detector): executed-path coverage and runtime tradeoffs.
- [Loop variables in Go 1.22](https://go.dev/blog/loopvar-preview): version-dependent capture semantics.
- [Go fuzzing](https://go.dev/doc/security/fuzz/): meaningful properties, seed inputs and deterministic targets.
- Local [Testing decisions](../../../plugins/go-quality-review/skills/go-testing/references/testing-decisions.md) and [review guide](../../go-quality-review/README.md): independent outcome criteria and ownership.
