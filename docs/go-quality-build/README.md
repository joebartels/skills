# Go quality build skills design record

This record preserves the direction for a multi-harness collection of skills that helps agents **write** strong Go code. The existing [Go quality review collection](../../plugins/go-quality-review/README.md) assesses code across nine topics. The build collection should use those topics to guide decisions while leaving the reviewer free to report defects. This is a working design record, not an implementation plan or a claim that the proposed skills have passed evaluation.

## Decisions from the discussion

- Use focused skills with distinct triggers and decisions. A skill should be valuable when its topic applies; it need not run on every Go task. A small coordinator may select the relevant skills and request an independent review of the completed changeset.
- Align writing guidance with all nine review topics. One build skill may improve several topics. Each decision rule should have one owner so overlapping skills do not give conflicting instructions.
- Judge advice against the project's actual contract, supported Go versions, consumers, workload, and deployment context. Avoid universal requirements for architectures, constructors, interfaces, nil versus empty values, test structure, third-party libraries, or tooling where the choice is contextual.
- Use [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang) as a source of content. Copy an entire skill or reference when it fits; adapt or omit conflicting instructions. Do not assume that importing the entire collection will produce A grades under the local rubric.
- Require behavioral evidence for each retained skill. Test whether it improves completed code and whether adjacent skills remain consistent, rather than only checking whether agents repeat its rules.
- Keep one canonical runtime copy usable by Claude Code, Codex, and OpenCode, following the repository's [multi-harness packaging design](../superpowers/specs/2026-09-29-multiharness-skills-design.md).

## Candidate skill boundaries

The names and count below are provisional. Split or merge candidates when real tasks show a clearer boundary. The review mappings identify expected benefits, not ownership of the review grades.

| Candidate | Decision it owns | Main review topics |
| --- | --- | --- |
| `go-package-boundaries` | Package responsibilities and dependency direction | Architecture; Code Quality |
| `go-api-contracts` | Exported APIs, CLI behavior, wire and file formats, compatibility | Architecture; Correctness |
| `go-interfaces-and-composition` | Concrete types, interfaces, constructors, dependency wiring | Architecture; Testing |
| `go-names-and-docs` | Names and documentation at consumer use sites | Code Quality; Architecture |
| `go-error-contracts` | Error handling, identity, wrapping, translation, and exposure | Code Quality; Correctness; Resilience |
| `go-values-and-zero-values` | Nil and empty values, receivers, copying, and aliasing | Code Quality; Correctness |
| `go-context-and-deadlines` | Cancellation propagation and time budgets | Correctness; Resilience |
| `go-concurrency-and-ownership` | Synchronization, goroutine lifetime, resource cleanup | Correctness; Architecture; Performance |
| `go-data-boundaries` | Domain, transport, and stored representations; transaction and format contracts | Architecture; Correctness; Security |
| `go-behavior-tests` | Cases and assertions that detect plausible regressions | Testing; Correctness |
| `go-test-isolation` | Real dependencies, fakes, fixtures, parallelism, and race checks | Testing; Reproducibility |
| `go-trust-boundaries` | Untrusted input, authorization, sensitive data, and unsafe sinks | Security; Correctness |
| `go-performance-evidence` | Benchmarks, profiles, and justified hot-path changes | Performance; Testing |
| `go-modules-and-builds` | Module resolution, Go versions, tags, generation, and build targets | Reproducibility; Correctness; Deployment |
| `go-telemetry` | Useful logs, metrics, traces, and sensitive-data limits | Observability; Security |
| `go-runtime-resilience` | Retries, overload, and downstream failure containment | Resilience; Performance |
| `go-release-operations` | Shipped artifacts, runtime configuration, rollout, and shutdown | Deployment; Reproducibility |

An optional `go-quality-build` coordinator should route a concrete task to a small set of applicable skills, track the contract and verification evidence, and invoke the separate [go-quality-report](../../plugins/go-quality-review/skills/go-quality-report/SKILL.md) for a final assessment. It should not duplicate the focused guidance or adjust the reviewer to improve a grade. Framework-specific and library-specific recipes belong in optional skills only when a project uses them. Formatting belongs primarily to tools such as `gofmt`.

## Upstream reuse policy

Before importing content, pin the upstream revision and inventory each candidate's main skill and references. For every section, record **copy**, **adapt**, or **omit**, with a reason and the local decision owner. Preserve useful examples and explanations when their behavior is correct. Remove harness-specific orchestration, automatic project configuration, universal tool or dependency requirements, and cross-references that would fail in the packaged collection unless they are intentionally supported.

The audit must compare upstream guidance with the runtime [review skills](../../plugins/go-quality-review/README.md), their decision references, and current primary Go documentation when version-sensitive. Known mismatches to resolve include unconditional empty-slice initialization, routine `%w` wrapping, interfaces always owned by consumers, mandatory integration build tags, and prescribed project layouts. Also reconcile existing local authoring material: the [code-quality handbook](../go-quality-review/resources/go-code-quality-and-idioms-handbook.md) makes an allocation claim about empty slices that the current [idiom decision reference](../../plugins/go-quality-review/skills/go-code-quality-and-idioms/references/idiom-decisions.md) correctly avoids.

Upstream's [evaluation report](https://github.com/samber/cc-skills-golang/blob/main/EVALUATIONS.md) reports a large improvement on its own assertions. Those assertions sometimes encode the prescriptive choices above, and several evaluated versions differ from current skill files. Treat that report as evidence of useful material and a benchmark design to inspect, not proof of an A grade under this collection's rubric.

The upstream [MIT license](https://github.com/samber/cc-skills-golang/blob/main/LICENSE) permits copying and modification. Retain its copyright and permission notice when copying a skill or substantial portion, and record source revisions for future updates.

## Authoring workflow

Use the `superpowers:brainstorming` skill to settle a bounded design for the first build-skill group, then `superpowers:writing-plans` for its implementation plan. Use `skill-creator` to author each runtime skill and its references. Use `superpowers:writing-skills` for behavioral baseline and skill-on evaluations, adapting its test process to the multi-harness package. Apply `superpowers:verification-before-completion` before declaring a skill ready. Use the existing Go review skills as an independent outcome check; they do not replace behavioral evaluations of the writing skills.

Complete one focused skill and its evaluation before expanding to the next. Keep a proposed skill out of the installable package until its trigger, guidance, links, and realistic behavior have been checked. Use the repository validator and the relevant harness validators for packaging. Record what was actually run and what remains unverified.

## Evaluation and promotion

For each candidate, test skill selection and non-selection, then compare completed changes on unseen Go tasks with and without the skill. Include libraries, CLIs, and services where relevant; version and contract edge cases; and tasks where an attractive blanket rule would be wrong. Run applicable builds and meaningful tests. Have the existing review skills assess the changes independently, with scope and coverage recorded. Compare confirmed findings, regressions, unnecessary code or dependencies, and effort, not only letter grades or rule compliance.

Test combinations of neighboring skills on the same task. Resolve contradictory advice at its decision owner and rerun the case. Promote a candidate only when it has a clear trigger, distinct useful guidance, and evidence of better outcomes. Merge or leave it as reference material if it adds little beyond another skill. Record the evaluated skill revision, task, model and harness, results, and limitations so later revisions can be compared honestly.

## Work status and next steps

- **Completed:** Compared the local review approach with the upstream collection; proposed candidate boundaries and the reuse and evaluation policy in this record. Added repository agent instructions and a continuation protocol so later workers maintain this record.
- **Next:** Decide package placement. A peer `plugins/go-quality-build/` package is the working recommendation; this is not yet a settled choice.
- **Next:** Pin and audit upstream source content against the review decisions, beginning with API and interface design, errors, values, concurrency, and behavioral testing.
- **Next:** Turn the agreed boundaries into a reviewed design specification and implementation plan before authoring runtime skills.
- **Next:** Build and evaluate a small first group, revise its boundaries, then expand to remaining topics that demonstrate value.

## Continuation protocol

Every agent working on this effort, including one resuming after context loss, should first read this record and the current [review guide](../go-quality-review/README.md), then inspect the repository state. The candidate table is a proposal; verify files and evaluation artifacts before reporting a candidate as implemented or effective.

After each meaningful stage and before finishing a task, update the status above and append a dated entry below. Each entry should name the work completed, exact files and upstream revision when applicable, checks run with their outcomes, decisions and reasons, remaining risks or open choices, and the next concrete action. Link detailed specs, plans, audit matrices, and evaluations rather than pasting them here. Preserve previous entries so a later worker can trace changes in direction. If work stops early, record the partial state and what would resume it.

### Work log

#### 2026-09-30 — Direction and tracking

- Recorded the proposed skill map, upstream reuse policy, and evaluation standard in this file. The first version was committed as `3fd6ee5`.
- Selected a workflow using brainstorming and planning for design, skill-creator and writing-skills for authoring and behavioral tests, and independent Go review for outcomes. This choice has not yet produced runtime build skills.
- Added `AGENTS.md` and `CLAUDE.md` so future Codex, OpenCode, and Claude sessions maintain this record; updated this file with the continuation protocol.
- Verification: `rtk python3 scripts/validate.py` passed with 10 review skills and 120 valid evaluation cases; all 10 local links in the three changed Markdown files resolved.
- Open choice: whether the writing skills ship as a peer plugin or in another package arrangement. A peer plugin remains the recommendation.
- Next action: settle the first skill group's design and package placement, then pin and audit the relevant upstream sources.
