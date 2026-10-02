# Go Quality Build Testing Group Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Evaluate and deliver focused Go writing skills for meaningful behavior tests and reliable test isolation.

**Architecture:** Extend the existing build plugin with one canonical runtime copy per promoted skill. Prepare fresh tasks and baseline evidence before each draft, evaluate each skill alone, then compare four combinations on a new integrated task. Keep drafts and evidence outside the installable tree and preserve the independent review rubric.

**Tech Stack:** Markdown, existing JSON eval schema, Python 3 repository validation, standard-library Go fixture modules, available Go toolchains and harness validators, independent author/reviewer contexts.

**Spec:** [Approved testing-group design](../specs/2026-10-01-go-quality-build-testing-group-design.md), approved 2026-10-01. Plan accepted 2026-10-01; native implementation with independent behavioral authors/reviewers selected by the recommended handoff.

## Global Constraints

- Runtime package: `plugins/go-quality-build/`. Drafts: `docs/go-quality-build/drafts/<skill-name>/`. Suites: `tests/go-quality-build/<skill-name>/evals/`. Evidence: `tests/go-quality-build/results/<run-id>/`.
- Skill order: `go-behavior-tests`, then `go-test-isolation`. Promote one before expanding runtime authoring to the next. An inconclusive candidate remains a draft/reference, is revised, or is merged; it must not be advertised as effective.
- Upstream pin: `samber/cc-skills-golang` at `19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`. Audit copy/adapt/omit decisions before runtime authoring; distribute the MIT notice for substantial copying.
- Baseline examples should remain usable in Go 1.22 projects. Optional `t.Context` examples require Go 1.24; stable `synctest.Test` requires Go 1.25. Respect effective module/file versions; do not raise a fixture's minimum to use a preferred helper.
- Behavior tests own cases, observation boundaries and assertions. Isolation owns dependency fidelity, fixture state, timing and cleanup. Production package/API/composition decisions remain with existing owners; review rules remain unchanged.
- Individual trials withhold the other new testing skill from both arms. Existing architecture-skill availability, task, fixture, author model and reasoning setting stay identical across matched arms.
- Prior combined code and assisted repairs are diagnostic history, not unseen uplift evidence. Review-guided repairs and revised blind trials are distinct evidence stages.
- No mandatory third-party testing dependency, tool installation, integration tag, parallelism, project layout or coverage target. No global installation or publication during development.
- Prefix shell commands with `rtk`. Maintain the canonical record after each meaningful stage and before stopping. Do not retry the older rejected zero-value composition trial.

## Review Focus

1. Consistently wrong encoder/decoder implementations can pass round trips; independent wire observations must reject them (Task 1, `TestWireRepresentationIndependent`).
2. A helper can return an error while its process exits successfully; actual failure status and accepted-prefix effects must be checked (Task 1, `TestRejectedBatchProcess`).
3. A handler or convenient double can hide real HTTP/cancellation/body behavior; keep each verification claim at its exercised boundary (Task 3, `TestHTTPCancelAndBodyLifetime`).
4. Parent teardown or an assertion failure can release fixtures before child work ends; control startup and join before resource release (Task 3, `TestParallelFixtureLifetime` and `TestTeardownOrder`).
5. A newer host can conceal unsupported test APIs or weak real-time checks; preserve Go 1.22 compatibility and measure recurrence from callback completion (Task 3, `TestCompletionRelativeRecurrence` and minimum-toolchain check).

---

## File map and evidence protocol

- Create `docs/go-quality-build/testing-source-audit.md`: new section inventory for the pinned testing family plus relevant prior layout/naming material.
- Create each candidate's `evals/evals.json`, `evals/files/<case>/README.md`, `go.mod`, source and starting tests. All fixture files are listed in `files`; expected assertions are outside author inputs.
- Create `evals/controller-probes/<case>/` for semantic probes and `evals/controller-probes/mutations.md` for mutation meanings. These paths are never listed in author `files`.
- Create `docs/go-quality-build/drafts/go-behavior-tests/{SKILL.md,references/behavior-observations.md}` and the corresponding isolation draft/reference, sequentially.
- Move exact promoted draft bytes into `plugins/go-quality-build/skills/<skill-name>/`; update root/plugin READMEs and description metadata when installed scope changes. Existing Codex marketplace and OpenCode sources already discover the package; their paths need no change.
- Use `tests/go-quality-build/testing-combined/evals/` for the new integrated suite. Preserve historical `combined` inputs/results unchanged.
- Modify `scripts/validate.py` or `tests/go-quality-build/test_evals.py` only if a concrete validation gap appears; write a failing test for that gap first.
- Update `docs/go-quality-build/README.md` throughout. Use actual execution dates in run IDs; names below describe suffixes, not a promise to finish on the plan date.

Every run has `README.md`, `manifest.json`, `comparison.md`, exact `skills/` snapshots, and a directory per case/arm/revision containing `dispatch.txt`, `prompt.txt`, unedited `author-report.md`, `source.patch`, `verification.json` and `selection.json`. Preserve exact dispatched text, including wrappers; do not reconstruct unavailable wrapper bytes from memory.

The manifest records full input/skill commits and SHA-256 maps for **all** input, final source and archived artifact files; model, reasoning setting, harness, fresh-context mechanism, effective Go versions, platform, supplied/opened skills, and limits. `verification.json` contains command arrays, cwd, relevant non-secret environment overrides, stdout/stderr, exit codes and duration. Mutation records add semantic ID, candidate-specific patch, clean result, compiled result, failing behavioral assertion or surviving result. A compile error, timeout without an interpretable contract failure, or unrelated crash is not a detected behavioral regression.

Before any dispatch, select supported author/reviewer model settings from the active allowlist and freeze them in the manifest. Prefer the prior comparable author tier when available. Fresh authors receive only the exact task, fixture and allowed skill inputs. Blind reviewers receive anonymized original/candidate code, project/task contracts, and the applicable review skills; withhold arm labels, writing skills, expected judgments, author reports and other reviews. Batch only independent reviews if needed and disclose batching.

Use disposable fixture copies and a writable temporary Go cache. Each candidate must reconstruct from the frozen input plus its source patch byte-for-byte. Controller probes establish actual contract behavior; regression sensitivity is measured with candidate tests only, without injecting those probes. Preserve unsuccessful trials and raw reviews unchanged. Set explicit process/test deadlines and keep synchronous tool waits at most 60 seconds, polling only running tools as needed.

### Task 1: Source audit and behavior-testing baseline

**Files:** Create the source audit; `tests/go-quality-build/go-behavior-tests/evals/evals.json`; five fixture directories and controller probes below; `results/<date>-behavior-tests-baseline/`; modify the canonical record.

**Interfaces:** Consumes the approved spec and existing Testing/Correctness decisions. Produces suite `skill_name="go-behavior-tests"` with case IDs below, existing `prompt`, `expected_output`, `assertions`, `files` fields, frozen fixture/probe hashes and complete blind baseline archives. No new validator interface.

| Case and files inside `evals/files/<case>/` | Public/task contract and controller observation |
| --- | --- |
| `library-codec`: `key.go`, `key_test.go`, README, module | `type Key struct { Region, Name string }`; `Encode(Key) (string, error)` and `Decode(string) (Key, error)` retain `ErrInvalidKey` identity. Extend to encoded slash/space values. Format is two individually path-escaped segments in Region/Name order; `Key{"us/west", "web blue"}` encodes as `us%2Fwest/web%20blue`. Empty fields, malformed escapes and incorrect segment counts reject. `TestWireRepresentationIndependent` checks exact bytes and decoded fields; a coordinated field-swap mutation must not be accepted merely because round trips pass. |
| `cli-partial-failure`: `main.go`, `main_test.go`, README, module | `ledgerload --dir DIR` imports CSV `id,quantity`, with positive integer quantity and nonempty ID. Process exit 0 succeeds, exit 2 rejects invalid input/usage; diagnostics go to stderr. Rows apply in order, duplicates replace, row failure stops later writes while keeping the accepted prefix. `TestRejectedBatchProcess` builds/runs the binary on good/bad/good rows and checks status 2, first record bytes and absence of the last record. A malformed CSV case also has a later valid row. |
| `service-publication`: `mirror.go`, `mirror_test.go`, README, module | `Refresh(ctx context.Context, client *http.Client, endpoint, path string) error` publishes a JSON item list with nonblank Code/Label. Add refresh behavior preserving the last good snapshot on rejection/failure and exposing complete published snapshots to readers. Controller tests assert both field values, each invalid field, unchanged old bytes on rejection and publication-boundary behavior; no prescribed storage interface. |
| `worker-host`: `worker.go`, `host.go`, starting tests, README, module | Extend a periodic worker with `Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error`. Host cancels/joins before release/return; cycles wait the interval after callback completion; callback errors propagate. A controlled callback enters cancellation cleanup and waits for release permission. `TestHostJoinBeforeRelease` compares observed callback-completion/resource-release order after both complete; `TestSuccessfulRecurrence` exercises a second successful callback. |
| `not-behavior-work`: `format.go`, `format_test.go`, README, module | Correct documentation wording and Go formatting only; no behavioral change or request to strengthen tests. Preserve code behavior and tests; record whether the offered description is sufficient to decline opening the skill. |

- [x] **Step 1: Complete the pinned audit.** Inventory the main testing skill and all seven references, plus prior testing-layout/naming sections. Map every retained decision to the local owner, reviewer criterion and primary Go documentation. Mark scaffolding, mandatory tags/naming/speed/dependencies and unrelated benchmark recipes adapt/omit. Record license handling and copy scope.
- [x] **Step 2: Create the five runnable inputs and contract probes.** Use Go 1.22 modules and standard-library dependencies. Starting tests cover old behavior and pass. Controller probes for requested new behavior may initially fail; record that distinction. Freeze semantic mutation IDs for swapped codec fields, success CLI exit, omitted independent field validation, direct truncating publication and omitted host join. Publication observations must disclose platform assumptions; do not silently claim Windows behavior.
- [x] **Step 3: Check and freeze preparation.** In each fixture run `rtk proxy go test -count=1 -timeout=30s ./...`, `rtk proxy go vet ./...` and `rtk proxy gofmt -l .`; expect exit 0 and no unformatted files. Verify exact JSON file inventory, controller-probe exclusion and hashes. Run `rtk proxy python3 -B scripts/validate.py`; expect 20 build cases and the unchanged 10 review skills/120 review cases. Commit preparation before dispatch.
- [x] **Step 4: Run five fresh blind baselines.** Supply relevant existing architecture skills consistently; supply neither testing candidate nor its expected judgments. Archive exact inputs and completed changes. Reconstruct and run clean tests, semantic probes, relevant process/build checks, targeted race checks and candidate-specific mutations.
- [x] **Step 5: Obtain blind independent outcome reviews.** Assess Testing and Correctness; add Architecture when production test seams materially change structure. Confirm which gaps are in new tests rather than unchanged legacy scope. Record findings, unsupported expectations, baseline strengths, mutation outcomes and environment limits. Strong baselines are controls, not evidence of a candidate's value.
- [x] **Step 6: Verify and commit the baseline archive.** Check every reconstruction/hash and complete dispatch provenance. Update status/log with exact results and next action. Commit as `test: establish Go behavior-testing baselines`. No runtime skill or uplift claim yet.

### Task 2: Behavior-testing draft, comparison and promotion

**Files:** Create the behavior draft/reference; `results/<date>-behavior-tests-skill-on/`; modify audit use notes, canonical record and package/root README/description metadata only on promotion. Runtime destination: `plugins/go-quality-build/skills/go-behavior-tests/`.

**Interfaces:** Consumes Task 1 frozen inputs, mutation meanings and baseline findings. Produces revision-identified draft snapshots, five matched skill-on outcomes, comparison and an explicit promotion decision. Final promoted bytes must equal the fully evaluated snapshot.

- [x] **Step 1: Author the smallest draft from observed gaps.** Use skill-creator and writing-skills. Implement the approved trigger/exclusions, contract-to-observation workflow, independent oracles, complete outputs/errors/side effects, partial failures, process/host boundaries, meaningful examples/properties and bounded regression-sensitivity checks. Keep isolation mechanics with its owner. No required framework or universal test form.
- [x] **Step 2: Validate and freeze draft bytes.** Run `rtk proxy python3 -B /Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py docs/go-quality-build/drafts/go-behavior-tests`; expect `Skill is valid!`. Verify all local reference paths inside the draft, source/license compliance and version-labelled snippets. Snapshot the entire draft/reference set before dispatch.
- [x] **Step 3: Run the five matched skill-on cases.** Offer the description and allow relevance-based opening, including the control. Withhold isolation and expected assertions. Use the Task 1 author settings and architecture inputs. Archive/reconstruct/check exactly as in Task 1.
- [x] **Step 4: Blindly review and compare.** Measure confirmed unique test gaps, semantic failures, meaningful mutation detection, new production coupling/dependencies and effort. If a claimed benefit is ambiguous, run one fresh repeat of the improvement-bearing task. A no-regression result against an already-strong baseline is not uplift.
- [x] **Step 5: Revise only warranted guidance and rerun.** Preserve older evidence. For any changed draft, rerun all four applicable cases and the control against final bytes, plus an improvement-bearing repeat when needed. Require clear selection, distinct measurable benefit and no unexplained harmful regression. If benefit remains inconclusive, record the concrete limitation and keep the candidate uninstalled; revise the plan before expanding to the next runtime skill.
- [x] **Step 6: Obtain promotion review and install exact evaluated bytes.** Have an independent task reviewer inspect the guidance, raw evidence, reconstructions, claims and decision boundaries. Move unchanged draft files into the runtime destination, retain evidence snapshots outside it, verify matching hashes, and describe only measured benefit in root/plugin READMEs. Add/update `plugins/go-quality-build/THIRD_PARTY_NOTICES.md` in the same promotion if substantial upstream content was copied. Update both plugin descriptions and Claude marketplace description to include behavior testing; no new marketplace entry or OpenCode source.
- [x] **Step 7: Verify and commit.** Run the shared delivery checks below and quick validation at the runtime destination. Validator should still count 20 build cases. Record promoted versus unverified claims and commit as `feat: add evaluated Go behavior-testing skill` only after the gate passes.

### Task 3: Isolation fixture preparation and blind baseline

**Files:** Create `tests/go-quality-build/go-test-isolation/evals/evals.json`; five fixture directories and probes below; `results/<date>-test-isolation-baseline/`; update audit notes and canonical record. Do not author its runtime draft yet.

**Interfaces:** Consumes the approved isolation scope and source audit. Produces suite `skill_name="go-test-isolation"`, the same eval/evidence interfaces as Task 1, and independent baseline findings. Both arms exclude the promoted behavior-testing skill, so this candidate's independent contribution remains measurable.

| Case and files inside `evals/files/<case>/` | Contract and concrete probe |
| --- | --- |
| `library-file-fixtures`: `store.go`, `store_test.go`, README, module | `New(root string) *Store`, `(*Store).Put(key, value string) error`, `(*Store).Get(key string) (string, error)` preserve instance-local data. Add tests for concurrent independent instances and grouped child cases. Existing serial fixtures work but their lifetime is unsafe for the new child work. `TestParallelFixtureLifetime` checks each child's distinct persisted values and that resources remain usable through completion. |
| `cli-environment`: `config.go`, tests, `cmd/showcfg/main.go`, README, module | `type Config struct { Endpoint, Mode string }`; `LoadFromEnv() (Config, error)` reads `INDEXER_ENDPOINT` and `INDEXER_MODE`. Unset defaults are `http://127.0.0.1:9090` and `read`; supplied endpoints require HTTP/HTTPS and a host, modes are `read`/`write`. The command prints lowercase-key JSON plus newline or exits 2 with stderr and no success stdout. Extend tests for unset, explicit and malformed values without leaking process state. `TestEnvironmentRestored` and `TestChildConfiguration` compare parent before/after values and actual isolated child outputs; parallel process-global mutation is invalid. |
| `service-http-boundary`: `fetch.go`, `fetch_test.go`, README, module | `Fetch(ctx context.Context, client *http.Client, endpoint string) (Metadata, error)` has documented method/path/status/JSON and cancellation behavior. Add reliable success/error/cancellation tests. `TestHTTPCancelAndBodyLifetime` checks actual HTTP request/response semantics where feasible and separately verifies response body closure with an observable transport fixture. A fake that ignores the supplied context or skips protocol handling cannot support a broader claim. |
| `worker-timing`: `poll.go`, `poll_test.go`, README, module | `Poll(ctx context.Context, interval time.Duration, callback func(context.Context) error) error` already promises sequential completion-relative cycles and parent cancellation. Replace guessed-readiness tests and add recurrence/isolation checks without changing the API or Go 1.22 minimum. `TestCompletionRelativeRecurrence` holds the first callback beyond the interval, observes release/completion, and measures from completion; `TestTeardownOrder` controls cancellation cleanup before closing its resource. |
| `not-isolation-work`: `clamp.go`, `clamp_test.go`, README, module | Repair a pure deterministic boundary comparison and add an ordinary regression test using local values. Offer the isolation description; expect no opening, fixture/concurrency machinery or new dependency. |

- [x] **Step 1: Create realistic starting fixtures and withheld probes.** Keep starting serial tests passing. State actual dependency/ownership contracts in READMEs, without prescribing the isolation implementation. Freeze mutation meanings for shared roots, premature parent teardown, leaked environment, ignored HTTP cancellation, discarded host context and interval measured from callback start. Worker-failure reporting is checked by inspecting assertion ownership and bounded completion, not by accepting an unrelated crash.
- [x] **Step 2: Verify and freeze preparation.** Use the same Go/test/vet/format/hash/inventory checks as Task 1; expect validator count 25. Attempt actual Go 1.22 execution with `GOTOOLCHAIN=local` when that binary is available; a newer compiler or a go directive alone is insufficient. Compile optional reference examples against their labelled versions and record unavailable toolchains. Commit the inputs before dispatch.
- [x] **Step 3: Run five fresh blind baselines.** Supply neither new testing skill; hold other inputs/settings constant with the later isolation arm. Preserve exact text, code, checks and selected/opened skill evidence. For relevant cases run individual tests, shuffled/repeated checks and race detection with finite deadlines; start with ten repetitions for timing/cleanup claims and broaden only if results are ambiguous.
- [x] **Step 4: Review and reconcile the baseline.** Obtain independent Testing/Correctness review and Architecture review for consequential test seams. Verify whether doubles preserve claims, actual listeners were available, timing measurements use the right event, resources survive child work and failure paths join cleanup. Distinguish pre-existing fixture gaps from introduced omissions.
- [x] **Step 5: Verify and commit.** Reconstruct/hash every candidate and raw artifact, update audit provenance/status/log and commit as `test: establish Go test-isolation baselines`. No isolation-skill benefit is claimed.

### Task 4: Isolation draft, comparison and promotion

**Files:** Create `docs/go-quality-build/drafts/go-test-isolation/{SKILL.md,references/isolation-patterns.md}`; `results/<date>-test-isolation-skill-on/`; modify audit notes, canonical record and promotion metadata/docs. Runtime destination: `plugins/go-quality-build/skills/go-test-isolation/`.

**Interfaces:** Consumes Task 3 inputs/findings; produces the same revision/promotion evidence interface as Task 2. Owns reliable test control/cleanup without changing the behavior skill's assertions or existing production decision owners.

- [x] **Step 1: Author the focused draft.** Cover resource ownership, subtest lifetimes, cancellation/join before release, early-failure cleanup, bounded waits, process-global state, justified parallelism, double fidelity versus real boundaries, version-appropriate timing controls and test-goroutine assertion ownership. Explain transaction rollback scope as a contextual example, without adding a database requirement or implying it was experimentally evaluated here.
- [x] **Step 2: Validate and freeze.** Use quick skill validation on the isolation draft, verify local links and provenance, and compile version-labelled examples. Include a Go 1.22 channel/context pattern; any Go 1.25 synctest example must remain optional. An example tested only on a newer compiler is labelled accordingly.
- [x] **Step 3: Run all five matched skill-on cases.** Offer relevance-based selection and withhold behavior testing from both arms. Preserve exact final skill/reference bytes, author settings and source/check artifacts. Execute the Task 3 probes/mutations and relevant clean individual, shuffle, repeat and race checks.
- [x] **Step 4: Independently review, compare and revise.** Check observed fixture/state/timing improvement, meaningful regression detection and unnecessary production hooks/interfaces. Repeat improvement-bearing cases if needed. After a draft edit, rerun all applicable cases and the non-selection control against the final bytes. Preserve unsupported expectations, environment substitutions and unsuccessful results.
- [x] **Step 5: Review promotion and move exact bytes.** Apply the Task 2 promotion gate with a separate reviewer. Verify the final draft/snapshot/runtime hash equality, remove the live draft copy, include any required third-party notice in the same promotion, update README count to five evaluated skills only if both promotions passed, and add isolation to matching package/Claude catalog descriptions.
- [x] **Step 6: Verify and commit.** Use the shared delivery checks and runtime quick validation; expect 25 build cases. Update status/log and commit as `feat: add evaluated Go test-isolation skill` only after the gate passes. Failure to establish distinct utility leaves it uninstalled and is a recorded result.

### Task 5: Fresh four-arm combined evaluation

**Files:** Create `tests/go-quality-build/testing-combined/evals/evals.json`, `evals/files/indexer-evolution/`, controller probes and `results/<date>-testing-combined/`; update the canonical record. Preserve all historical combined files.

**Interfaces:** Suite `skill_name="testing-combined"`, one case `indexer-evolution`. Four independent arms `architecture-only`, `behavior-only`, `isolation-only`, `both` share task/input/architecture snapshots/model settings; only availability of the two testing skills differs. Produces anonymized reviews and an arm-level comparison.

- [x] **Step 1: Build and freeze a fresh integrated task.** Use an indexer library plus command that gains ordered batch ingress and a host-owned remote refresh worker publishing a local JSON index. Keep exported consumer and process/wire contracts explicit; require invalid-input retention, failed-start process status, successful recurrence and joined resource release. Inputs are newly authored, with no prior repair code or reviewer mutation instructions. Use standard-library dependencies and Go 1.22. The same eval schema requires no validator extension; expect 26 build cases.
- [x] **Step 2: Freeze independent semantic probes and prepare dispatches.** Map each promised behavior to an observation and mutation meaning, including actual process status, host join, complete publication and parent cancellation. Archive four exact wrappers with allowed skill snapshots and selection instructions. Do not force a control or reviewer to open a writing skill.
- [x] **Step 3: Run the four fresh blind arms.** Archive full source and selection evidence, reconstruct candidates, run contract/process/build checks and targeted candidate-test mutations with applicable race/isolation repetitions. No author receives expected assertions, controller probes, another arm or review feedback.
- [x] **Step 4: Obtain blind Architecture/Correctness/Testing reviews.** Compare confirmed findings and effort without averaging grades. Check complementary benefit, conflicting instructions, duplicated test machinery and unnecessary production abstraction. If splitting adds no distinct value, record that result and propose merging before changing approved ownership boundaries.
- [x] **Step 5: Resolve supported conflicts at their owner.** Any changed skill requires its final-revision single-suite cases/control and affected combined arms to rerun. Keep review-guided repairs in a separately labelled archive; they do not replace blind results. Verify all combined reconstructions/hashes and commit as `test: compare Go testing skills in combined use` with status/log updated.

### Task 6: Delivery review and package closeout

**Files:** Modify `plugins/go-quality-build/{plugin.json,.claude-plugin/plugin.json,README.md}`, `.claude-plugin/marketplace.json`, root README and canonical record; create whole-package review/check artifacts in the combined run. Add a third-party license/notice file only if the recorded copy scope requires it.

**Interfaces:** Consumes exact promoted runtime/evidence bytes and combined results. Produces matching package version `0.2.0` for the delivered testing group, bounded public claims, a current continuation record and local reviewable commits. If the evaluated outcome changes delivered scope, report the actual scope and update the plan rather than claiming the planned group shipped.

- [ ] **Step 1: Audit final scope and metadata.** Make the two manifests' versions/descriptions agree; preserve existing catalog/OpenCode package paths. Ensure only promoted skills are installable, one canonical runtime copy exists for each, descriptions/activation examples cover actual scope, and READMEs distinguish structural validity, blind behavioral evidence, assisted repairs and unverified routing/toolchains/platforms.
- [ ] **Step 2: Verify final evidence and package.** Reconstruct every new trial patch and verify all input/source/skill/artifact hashes. Run the shared delivery checks, final runtime quick validation and local reference checks. Inspect actual Codex/OpenCode verification capabilities; use available nonmutating loading checks and disclose unavailable ones. Do not infer runtime loading from JSON validity or install globally.
- [ ] **Step 3: Obtain independent whole-package review.** Review source-audit/license decisions, exact skill bytes, triggers/ownership, combinations, regression-sensitivity evidence, claims, metadata and continuation state. Correct confirmed findings with minimal changes; guidance changes re-enter behavioral gates, while documentary corrections need only relevant structural rechecks.
- [ ] **Step 4: Record and commit final delivery.** Append precise checks/results/decisions/limits and next priority to the canonical record. Preserve the older blocked composition trial as incomplete. Commit as `docs: deliver evaluated Go testing skill group`, leave the managed worktree's state explicit and hand over artifacts and measured outcomes. No PR or publication is required by this plan.

## Shared delivery checks

Run from the repository root after promotion and at final delivery:

```sh
rtk proxy python3 -B -m unittest discover -s tests -p 'test*.py'
rtk proxy python3 -B -m unittest discover -s tests/go-quality-build -p 'test_*.py'
rtk proxy python3 -B tests/go-quality-review/test_layout.py
rtk proxy python3 -B tests/go-quality-review/go-quality-report/evals/test_grade.py
rtk proxy python3 -B scripts/validate.py
rtk claude plugin validate ./plugins/go-quality-build
rtk claude plugin validate .
rtk git diff --check
```

Existing expected unit-test counts are root 6, build-eval 10, review layout 1 and grade 12, unless a justified test addition changes them. Review validation remains 10 skills/120 cases; build counts progress 15 → 20 → 25 → 26. These commands establish structural integrity, not benefit. Check newly staged files as well as unstaged changes. If preserved raw patch/review text triggers whitespace diagnostics, preserve evidence bytes and document exact exclusions; runtime/source/document files must pass.

## Execution handoff

The implementation method controls who builds the fixtures, drafts and package; **both methods still require fresh independent behavioral authors and reviewers** for the approved evaluation design.

Recommend native implementation in this session: the six tasks share fixture/archive conventions and promotion bookkeeping, so retaining that context should reduce repeated setup. Independent outcome reviews remain mandatory at the documented gates, with the most capable available reviewer for the final whole-package review. Subagent-driven implementation is also available for fresh task-level implementation/spec/quality review contexts before each next task. Obtain user plan review and method selection before Task 1.
