# Go Quality Build Context and Concurrency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Evaluate two focused Go authoring candidates and deliver only guidance with demonstrated individual benefit and compatible package behavior.

**Architecture:** Complete the context study before the concurrency study, with the exact five-skill 0.2.0 catalog fixed in both individual comparators. Freeze original fixtures and held observations, compare fresh independent authors, decide each candidate separately, then run a four-arm interaction study only if both qualify. Keep drafts/evidence outside the installable package and promote exact accepted bytes.

**Tech Stack:** Markdown, existing JSON eval schema, standard-library Go 1.22 fixture modules, Python 3 repository checks, available native harnesses/toolchains, independent author and reviewer contexts.

**Spec:** [Context/concurrency design](../specs/2026-10-02-go-quality-build-context-concurrency-design.md), accepted for planning after the [independent alignment review](../../go-quality-build/reviews/context-concurrency-spec/review.md) resolved F1. Plan accepted 2026-10-02 (“looks good. onwards”); the recommended native method proceeds with independent behavioral/promotion reviewers.

## Global Constraints

- “Deliver two focused candidates, `go-context-and-deadlines` and `go-concurrency-and-ownership`, completing each individual study before proceeding to the next.”
- “Existing package/API/composition/testing skills remain the frozen comparator.” Both individual arms use the exact five-skill 0.2.0 catalog at `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96`; context is excluded from concurrency's individual comparator even if accepted.
- “Error/value drafts remain non-installable, names/docs remains a reference, and the optional error-promotion study is separate work.”
- “No coordinator, retry framework, telemetry package, general Go handbook, global installation or new evaluation framework belongs in this group.” No push/PR/merge is part of this plan.
- “Drafts live under `tests/go-quality-build/<candidate>/draft/`; author evaluations consume immutable snapshots of those exact bytes.” Runtime destinations remain `plugins/go-quality-build/skills/<candidate>/`.
- Samber pin `8e899e20ff0cd4dc524af3993e4c62d8ee8c5717`; spf13 `9ac6eca43161163bb21621a520a57df50d5ad464`; Uber `1d60a91aa5e87d443002e23c21903c49489dbde5`; Google `fc981500047ef2cc2e2509b00dcd23c9badaf8e4`. Retain required notices for substantial copying; historical audits keep their pins.
- “Preserve supported signatures rather than imposing context on every function.” “Keep serial work serial when it satisfies the contract.” No mandatory channels, locks, pools, errgroup, mocks, integration tags or third-party test dependencies.
- “Test actual Go 1.22 and an available newer supported toolchain when possible; a 1.22 directive under a newer compiler is insufficient.” Record unavailable toolchains/platforms and effective module/file versions; do not raise minimums for convenient APIs.
- “Each attempted author launch counts toward the 36-author ceiling and the applicable allocation, including an infrastructure failure or retry.” Per-candidate maximum 16; integrated maximum four. At most one retry per failed launch, using an unspent slot; revisions consume remaining slots.
- “This is at most 40 probes across two candidates and five profiles”; selection-only probes are counted separately from authors, reviews and checks. Do not describe forced file reads as native discovery.
- “At least two independent matched primary-case pairs correcting the same confirmed relevant defect, plus successful final-byte transfer behavior.” Strong ties preserve behavior but do not supply corrections.
- “If discovery finds no relevant weakness, stop before drafting and record the strong baseline.” Insufficient rerun budget leaves a reviewed draft and declared extension proposal.
- “Independent outcome review follows the unchanged `go-quality-report` coverage contract: consider all nine topics across the requested scope and record each topic's applicability, assigned boundary and assessed coverage.” Do not alter grade rules or average grades.
- “Do not retry the earlier auto-review-blocked composition trial as part of this group.” Prefix shell commands with `rtk`; update canonical status/log after every meaningful stage and before stopping.

## Review Focus

1. Successful work can coincide with cancellation without defining the return decision; accepted progress, independent failures and late success must follow explicit observation order (Task 1, `TestResultDecisionOrder`).
2. An HTTP response can arrive before its body completes; the operation budget and ownership must cover consumption and closure (Task 1, `TestResponseScopeAndClose`).
3. Required finalization can inherit an already-canceled signal or be detached without a bound; use its stated separate scope and observe actual completion/process failure (Task 1, `TestFinalizationAfterCancel`).
4. Partial startup or blocked admission can survive cancellation while another worker reports failure; stop admission, observe all started completions and release resources afterwards (Task 3, `TestPartialStartAndBlockedAdmission`).
5. A consumer can finish normally while a producer is blocked; stop/join both stages without changing normal completion into failure or hiding an independent error (Task 3, `TestEarlyConsumerAndProducerFailure`).

---

## File map and evidence interface

- Create `docs/go-quality-build/context-concurrency-source-audit.md`: section-level copy/adapt/omit decisions, local owners, primary/version evidence and actual notice obligations. Start from the scoped source review, not a new repository survey.
- Create each candidate's `evals/evals.json`, `evals/files/<case>/` inputs and `evals/controller-probes/<case>/` held tests. `evals/controller-probes/mutations.md` records at most four frozen semantic mutation meanings per candidate. Held tests/mutations are never listed in author-visible `files`.
- Create a draft `SKILL.md` only after a relevant discovery defect. Add conditional references only for demonstrated explanatory needs; do not precreate reference/script/asset directories.
- Use `tests/go-quality-build/context-concurrency-combined/evals/` for the fresh interaction suite; leave historical `combined` suites unchanged.
- Create result archives named `<actual-date>-context-discovery`, `-context-study`, `-concurrency-discovery`, `-concurrency-study`, `-context-concurrency-combined` and `-context-concurrency-delivery` under `tests/go-quality-build/results/`, with suffixes only when a distinct revision/run requires them. Link prior originals instead of duplicating them.
- Modify root/plugin READMEs, both plugin manifests and `.claude-plugin/marketplace.json` only for accepted runtime additions. Codex/OpenCode discovery paths already cover the package; preserve them. Add distributed notices only when required by actual retained content.
- Maintain `docs/go-quality-build/README.md` and this plan's checkboxes. Existing runtime/review skills, drafts, historical fixtures and sealed results remain unchanged unless a reproduced owner-specific contradiction is separately evaluated.

Each archive contains `README.md`, `manifest.json`, `comparison.md` and exact `skills/` snapshots. Each attempted author directory contains `dispatch.txt`, `prompt.txt`, unedited `author-report.md`, `selection.json`, `source.patch`, `source-hashes.json`, `verification.json`, and a linked neutral `reviews/` packet; failed attempts retain whatever evidence exists plus the failure reason. Capture exact wrappers, not reconstructed recollections.

The manifest records full commits/hashes for input inventories and guidance, actual profile settings/capabilities, author slot ledger, neutral review source mappings, all-nine topic applicability/coverage, mutation meanings, toolchains/platform, attempts/costs and limits. `verification.json` records command argument arrays, cwd, non-secret overrides, stdout/stderr, status and elapsed time. Every candidate reconstructs byte-for-byte from frozen inputs plus its patch. Preserve raw cards and reconciled finding/correction IDs; freeze reviews before exposing arm identities.

Authors see only their task, listed inputs and allowed catalog, in fresh isolated contexts without the planning tree, held probes, expected judgments, other outcomes or reviewer repairs. Outcome reviewers receive neutral launch names/paths, original task contracts and original/candidate sources, without writing guidance or arm labels. Their mode is a code-area review of the completed requested task paths/contracts; original sources establish the requested evolution and exclude unrelated legacy behavior. Guidance/promotion reviewers receive the full disclosed evidence after outcome cards are frozen. Record accidental priming and assisted repairs honestly.

Freeze these profiles before dispatch: reference model/reasoning/native harness; alternate model with the same reference reasoning/harness where supported; alternate reasoning with the reference model/harness; each other native target harness among Claude Code, Codex and OpenCode. Actual supported identifiers and versions belong in the manifest. Native profiles with different models disclose confounding. Missing capability is a gap, not permission to substitute a simulated harness or invent settings.

## Fixture interfaces and observations

All case directories contain `README.md`, `go.mod` (`go 1.22`), the named source files and starting tests. The README freezes supported inputs, previous/requested behavior, public signatures, progress/error/ownership policy and version/platform assumptions. Supplied host/parser/client code keeps the requested author change on the candidate's decision; it must not hide the boundary under review.

### Context fixtures (Task 1)

| Case / source files | Exact interface and held observation |
| --- | --- |
| `library-call-cause`: `process.go`, `process_test.go` | `Process(ctx context.Context, jobs []int, apply func(context.Context, int) error) (accepted int, err error)`. Jobs are processed sequentially; each nil callback result increments accepted. Already-canceled calls return cancellation classification/cause with no work; otherwise empty jobs succeed with zero accepted. Check cancellation before starting each callback. A callback error ends work, preserves accepted count and the independent error; cancellation already observed at that decision also preserves classification and cause. Last callback success completes the operation; cancellation after that completion does not invalidate success. `TestCanceledBeforeStart` asserts zero callbacks/accepted; `TestCustomCauseAndIndependentFailure` controls cancellation before releasing a failing callback, checks `errors.Is` classifications and `errors.As` into a slice-backed legal error with independently checked elements. `TestResultDecisionOrder` separates empty input, between-job cancellation, coincident failure and completed success with gates. |
| `service-total-budget`: `stages.go`, `stages_test.go` | `FetchAll(ctx context.Context, client *http.Client, endpoints []string, total, stage time.Duration) (bodies [][]byte, err error)`. Positive budgets and fixture GET endpoints returning status 200 are supported. One operation scope derives from the caller; each stage narrows that scope. Only completely read successful bodies enter the ordered accepted prefix. Own response bodies, borrow the client. `TestTotalAndEarlierParentDeadline` observes client-side request deadlines across gated stages and an earlier parent; `TestResponseScopeAndClose` holds a body read beyond header arrival and verifies cancellation reaches it, accepted prefix survives and each owned body closes. Use a real client plus controlled transport for deadline/close observations and an actual local HTTP boundary where permitted; label each claim precisely. |
| `cli-finalization-budget`: `run.go`, `run_test.go`, `cmd/finalize/main.go` | `Run(ctx context.Context, jobs []int, apply func(context.Context, int) error, finalize func(context.Context, int) error, finalizationBudget time.Duration) (accepted int, err error)`. Positive finalization budget. Finalize accepted work once when accepted is nonzero, under a separate bounded scope, and wait for its cooperative completion. Processing cancellation/failure and independent finalization failure remain inspectable; no success claim for either failure. Fixture-supplied command host reads integer jobs, reports accepted events and owns signal/context setup; writes a final accepted-count receipt. Exit 0 succeeds, exit 2 reports processing/finalization failure. `TestFinalizationAfterCancel` asserts a live bounded finalization context after parent cancellation, accepted receipt and status 2 through a real child. Exercise finalization failure separately. Platforms lacking the supplied signal mechanism retain an explicit gap. Transfer only after draft freeze. |
| `not-context-work`: `range.go`, `range_test.go` | Private `inRange(value, low, high int) bool`; repair inclusive upper-bound behavior for `low <= high`. `TestInclusiveUpperBound` asserts lower, upper and adjacent outside values. Preserve the function and ordinary local test structure; no context/deadline/configuration machinery. |

Context mutation meanings: unsafe equality of arbitrary causes; loss of inspectable custom cause; fresh full budget at each stage; finalization using the canceled parent or lacking its declared bound (choose one concrete substitution before exposure). Compile each mutation and verify its intended assertion on a controller conformance implementation; these demonstrations never count as author benefit.

### Concurrency fixtures (Task 3)

| Case / source files | Exact interface and held observation |
| --- | --- |
| `library-shared-state`: `totals.go`, `totals_test.go` | `type Snapshot struct { Count int64; Sum int64 }`; `type Totals struct` with private state; `(*Totals).Add(delta int64)` and `(*Totals).Snapshot() Snapshot`. Zero value usable, no copying after use. Add increments count once and adds delta atomically with respect to snapshots; returned value is an independent observation. `TestCoherentSnapshots` uses concurrent delta=1 updates and checks Count==Sum on every snapshot, plus exact final totals from an independent schedule/count; a separate mixed-delta run checks final sum/count. `TestIndependentTotals` starts two instances with distinct workloads and observes both independently. No extra public API or framework. |
| `service-owned-workers`: `host.go`, `host_test.go` | `type Job int`; `type Lease interface { Run(context.Context, Job) error; Close() error }`; `Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error`. Supplied cooperative lifecycle/acquisition protocol; positive limit (primary case 2). Open returns a nonnil lease on success or nil plus error on failure. Limit bounds acquired in-use leases, not just running callbacks. Caller owns the input channel; host owns successfully acquired leases, including any acquired before stopping prevents a run from starting. Stop further admission on failure/cancellation, request all affected work to stop, join all started runs before closing their resources, and retain independent operation/close failures under the stated error policy. Successful job effects remain accepted. `TestPartialStartAndBlockedAdmission` holds cleanup of a started run, fails another open/run, keeps a further admission blocked, and checks no later open/close/host return occurs before the required completion events. Run variants for startup failure, worker failure and caller cancellation; `TestCapacityAndAcceptedWork` checks cap 2 and retained successful effects. |
| `cli-pipeline-stop`: `pipeline.go`, `pipeline_test.go`, `cmd/pipeline/main.go` | Private `runPipeline(ctx context.Context, capacity int, produce func(context.Context, chan<- int) error, consume func(context.Context, <-chan int) error) error`; capacity 1 in primary observations. Fixture-supplied command parser/consumer emits integer lines in input order, accepts a positive `--take` limit and owns process setup. Normal early consumer completion stops/joins the producer and exits 0. An independent producer/consumer error observed before coordinated stopping remains a failure, exit 2, with accepted output preserved. Parser supports finite ordinary integer-line inputs; malformed first input is a deterministic process failure. `TestEarlyConsumerAndProducerFailure` gates blocked send/receive and independent failure ordering, observes both completions, then verifies real-child ordered prefix/status on normal early completion and failure. No claim that arbitrary blocking `io.Reader` calls can be killed by context. Transfer only after draft freeze. |
| `not-concurrency-work`: `sum.go`, `sum_test.go` | Private `sumPositive(values []int) int`; repair inclusion of the final positive element. `TestFinalPositiveValue` independently checks empty, negative-only and mixed input sums. Preserve serial code and local values; no lock/channel/goroutine/pool. |

Concurrency mutation meanings: split compound update/publication; close/release before every user completes; admission that cannot observe stopping; stage send/receive that cannot observe coordinated early completion. Freeze the specific compiling substitutions and assertions before exposure, at most four; supplementary discoveries stay separately labeled.

### Interaction fixture (Task 5)

`request-scoped-fanout`: `fanout.go`, `fanout_test.go`, README and module. Fixed supplied API: `type Session interface { Do(context.Context, int) (int, error); Close() error }`; `type Result struct { Job, Value int }`; `Fanout(ctx context.Context, jobs []int, limit int, total time.Duration, session Session) ([]Result, error)`. Positive limit (2) and total budget; ownership of the supplied session transfers to this call. Results are successful jobs in input order, retaining their original Job values even when another job fails. One derived total scope, bounded admission, cancellation reaches cooperative session work, independent errors remain inspectable, all started work completes before Close/return, Close occurs once. Supplied construction/host behavior avoids a new architecture task. `TestFanoutBudgetAdmissionAndRelease` gates active work, blocked admission, independent failure and held cleanup; `TestFanoutOrderedAcceptedResults` checks retained results and error identity. No old integrated source is an unseen holdout.

## Task 1: Source audit, context fixtures and discovery

**Files:** Create the source audit, context eval/probe files above and `results/<date>-context-discovery/`; modify canonical record.

**Interfaces:** Consume the accepted spec, source review, existing five-skill runtime bytes and unchanged review decisions. Produce suite `skill_name="go-context-and-deadlines"` with its four case IDs, exact input/probe hashes, actual profile/slot manifest and three fresh discovery/control baselines. Transfer has no author outcome yet.

- [x] **Step 1: Complete the section audit.** Inventory the selected files/sections at the four pins, map copy/adapt/omit to decision owners and primary/version evidence, and record actual licenses/copy scope. Recheck any retained version-sensitive example; preserve existing audits. Use original wording by default, not imported manuals.
- [x] **Step 2: Write the context contracts and observations.** Create the exact fixtures above, ordinary pre-change tests and held requested-behavior tests. Run held tests on the starting source and record the relevant missing behavior; distinguish intentional pre-change behavior from broken fixture infrastructure. Give every gate cleanup and a bounded diagnostic wait, with worker errors delivered to the test goroutine.
- [x] **Step 3: Verify fixture feasibility and mutation assertions.** In disposable copies, implement the minimum controller conformance behavior and show held tests pass; apply each frozen mutation and show it compiles but fails the intended assertion. Preserve these as controller demonstrations, never baselines. Use events for ordering; real-time bounds are failure guards, not guessed readiness.
- [x] **Step 4: Validate and freeze inputs/settings.** Run the fixture checks below; expect clean old-contract tests, vet/format/build, complete listed inventories and controller-probe exclusion. Run repository validation; expect 42 build cases, unchanged 10 review skills/120 cases. Freeze five-skill hashes, native capabilities, profiles, task/probe inventories, primary case and all slot reservations. Commit as `test: prepare Go context discovery fixtures`.
- [x] **Step 5: Run three blind baselines.** Reference profile, `library-call-cause`, `service-total-budget`, `not-context-work`; no candidate draft. Archive every launch and actual settings. Reconstruct each source, run checks/held observations and test candidate assertions against mutations separately from controller probes.
- [x] **Step 6: Independently review and decide whether to draft.** Account for all nine topics; require Correctness/Code Quality for changed Go, Testing for assertion sensitivity and applicable budget/ownership/resource/process questions. Freeze cards before revealing arms. Identify the confirmed primary defect or declare an alternate discovery primary/repeat design before consuming remaining slots. If no relevant discovery weakness, stop this candidate before drafting and record the strong baseline. Verify hashes/provenance, update status/log and commit `test: record Go context discovery outcomes`.

## Task 2: Context draft, matched study and disposition

**Files:** Conditionally create `tests/go-quality-build/go-context-and-deadlines/draft/SKILL.md` and justified local references; `results/<date>-context-study/`; modify audit use notes/canonical record. No runtime destination yet.

**Interfaces:** Consume Task 1's frozen catalog/contracts/settings and confirmed relevant defect. Produce exact-byte draft snapshots, matched outcomes/loading probes, reconciled comparison and independent accepted/not-accepted disposition. If Task 1 stopped, record that disposition and skip authoring.

- [x] **Step 1: Author the smallest useful draft.** Use skill-creator and writing-skills. Implement the discriminating trigger, parent/child/total budget decisions, cancel ownership, contract-sensitive result/cause policy and deliberately bounded detachment. Distinguish requesting stop from observing completion; avoid duplicating composition, error, testing or telemetry guidance.
- [x] **Step 2: Validate and freeze exact guidance.** Run skill-creator quick validation, local reference checks and source/license/example checks. Review clarity, single decision ownership, standalone use and absence of universal dependencies. Snapshot all draft bytes and hash them before exposure; do not tune the draft on transfer results.
- [x] **Step 3: Run matched reference exposures.** Same three Task 1 cases/settings/catalog, toggling only this candidate. Record actual opening/exposure and control non-selection, raw outputs, reconstructions and checks. Keep failed launches; allocate retries only from unspent slots.
- [x] **Step 4: Run final-byte transfer and primary profile pairs.** After freeze, run baseline/exposure for `cli-finalization-budget`; then primary pairs at alternate model, alternate reasoning and each available other native harness. Allocation including Task 1 is 6 reference authors + 2 transfer + 2 alternate model + 2 alternate reasoning + up to 4 native = 16. Within each pair all other inputs/settings match.
- [x] **Step 5: Measure native selection separately.** At each available profile, run applicable and control requests plus a meaning-preserving paraphrase of each, without code implementation. Capture genuine native catalog/loading/access behavior through temporary local packaging; no global installation. Maximum 20 probes for this candidate. Record unavailable runtime support and unknown/combined axes explicitly.
- [x] **Step 6: Review, compare and resolve guidance findings.** Obtain neutral outcome cards and independent exact-guidance/evidence review. Count confirmed causes, preserved behavior and added complexity/costs. Candidate changes invalidate affected final-byte outcomes/loading evidence; rerun inside the same ceiling or leave a reviewed draft with an extension proposal. Assisted repairs remain separately labeled.
- [x] **Step 7: Record disposition and commit.** Require two independent matched primary corrections of the same confirmed defect, successful final-byte transfer, preserved applicable/control behavior, no material conflict/regression and accepting guidance/outcome review. Missing native profiles limit reuse claims, not invented passes. Record exact accepted hashes or reasons for nonpromotion, reconstruct/seal the study and commit `docs: decide Go context candidate readiness`. Finish its disposition before Task 3.

Completed disposition: reviewed draft, not promoted. The documented checked-catalog reallocation consumed16 attempts; alternate-model15/16 failed pre-model, and alternate-reasoning/other-native implementation pairs remain gaps. Native selection14/20 complete. No promotion gate was relaxed.

## Task 3: Concurrency fixtures and discovery

**Files:** Create concurrency eval/probe files above and `results/<date>-concurrency-discovery/`; modify audit/canonical record.

**Interfaces:** Consume the accepted design/audit and Task 2's completed disposition. Produce suite `skill_name="go-concurrency-and-ownership"`, frozen contracts/probes and three fresh reference baselines; comparator remains the original five skills, independent of context acceptance.

- [x] **Step 1: Create the concurrency inputs and held tests.** Use the exact interfaces/contracts above. Supply host-facing lifecycle boundaries, keep shared-state invariants concrete and freeze error/accepted-work policy. Starting ordinary tests must pass; held requested-behavior checks record relevant missing behavior.
- [x] **Step 2: Verify conformance and mutation meaning.** In disposable controller copies, make all requested checks pass; validate the four compiling semantic mutations fail intended assertions. Hold worker cleanup, admission, sends and receives with events; ensure regression cleanup cannot turn a meaningful assertion into an uninterpretable hung test.
- [x] **Step 3: Freeze and commit preparation.** Run fixture checks, JSON inventories/hash/probe exclusions and repository validation; expect 46 build cases and unchanged review cases. Freeze the original comparator, profile settings, primary case and at most 16 author slots. Commit `test: prepare Go concurrency discovery fixtures`.
- [x] **Step 4: Run three fresh baselines.** `library-shared-state`, `service-owned-workers`, `not-concurrency-work`; archive/reconstruct/check as Task 1. Context draft/runtime bytes are unavailable in author catalogs and working copies.
- [x] **Step 5: Obtain independent discovery review and close the stage.** All-nine applicability, Correctness/Code Quality/Testing plus actual ownership/resource/budget questions. Distinguish a concurrency-mechanics weakness from composition's already-covered lifecycle design. If no relevant confirmed weakness, stop before drafting; do not manufacture a new skill by repeating composition. Verify evidence, update status/log and commit `test: record Go concurrency discovery outcomes`.

## Task 4: Concurrency draft, matched study and disposition

**Files:** Conditionally create `tests/go-quality-build/go-concurrency-and-ownership/draft/SKILL.md` and justified references; `results/<date>-concurrency-study/`; modify audit/canonical record.

**Interfaces:** Consume Task 3's confirmed primary defect/frozen inputs. Produce its own final-byte snapshots, matched benefit/transfer/reuse evidence and independent disposition, without credit borrowed from context.

- [x] **Step 1: Author and validate the focused draft.** Use skill-creator/writing-skills for invariant ownership/publication, fitting synchronization, capacity/admission, channel closure and observable stop/join/release mechanics. Keep minimal serial controls; no universal lock/channel/errgroup/panic-recovery choices. Quick-validate, verify local references/examples/notices and freeze snapshot hashes.
- [x] **Step 2: Run matched reference exposures.** The same three Task 3 cases and original five-skill comparator; only concurrency exposure changes. Preserve exact input/source/launch/review evidence.
- [x] **Step 3: Run transfer/profile pairs and selection probes.** After freeze, baseline/exposure for `cli-pipeline-stop`, primary alternate-model and alternate-reasoning pairs and up to two other native-harness pairs. Total at most 16 authors including discovery; at most 20 selection-only probes, following Task 2's same matching/counting/loading rules.
- [x] **Step 4: Reconcile outcome and guidance reviews.** Apply all-nine applicability and unchanged report contracts. Confirm same-defect repeated correction, successful final-byte transfer and preservation. Resolve conflicts at their actual owner; changed guidance requires affected reruns within the cap. If the evidence fails, keep a reviewed draft/reference disposition rather than promoting for symmetry.
- [x] **Step 5: Verify, seal and commit the decision.** Reconstruct all outcomes, verify hashes/cost/slot records, retain original contrary/failed/assisted results, update canonical status and commit `docs: decide Go concurrency candidate readiness`.

Completed disposition: reviewed draft, not promoted. Fifteen counted authors, twelve selection probes, zero matched primary corrections; exact guidance accepted with transfer/preservation/routing limitations. Original discovery excluded from matched benefit under corrected toolchain environment.

## Task 5: Fresh interaction gate

**Files:** Conditionally create `tests/go-quality-build/context-concurrency-combined/evals/` and `results/<date>-context-concurrency-combined/`; modify canonical record.

**Interfaces:** Consume both individual dispositions. Produce either a documented skipped four-arm study or suite `skill_name="context-concurrency-combined"` with `request-scoped-fanout`, four matched outcomes and resolved interaction findings.

- [x] **Step 1: Apply the eligibility gate.** If both qualify, prepare the supplied interaction API/contract/tests above. If one/none qualify, record the skip; Task 6 reviews any accepted skill with the existing package without inventing a rejected neighbor or substituting old tasks.
- [x] **Step 2: Check and freeze the new interaction fixture.** Show controller conformance and meaningful held failure observations, then freeze task/catalog/profile/skill hashes. Repository validation expects 47 build cases when this suite exists, otherwise 46. No candidate guidance changes from preparation observations without reentering individual final-byte gates.
- [x] **Step 3: Run the four matched arms within four launches.** Same reference profile and task, original five-skill catalog plus neither/context only/concurrency only/both accepted snapshots. Reserve one launch per arm; keep labels outside authors and neutral reviewers. Count every attempt within four integrated slots and 36 total; any retry consumes an unspent arm slot and leaves an explicit coverage gap rather than a complete four-arm claim.
- [x] **Step 4: Independently review and resolve interaction.** Record all-nine coverage, outcomes, conflicts, complexity and costs; integrated success cannot replace individual benefit. A guidance change reopens affected individual and integrated evidence within remaining budget; otherwise retain an explicit incomplete/reviewed-draft disposition. Verify/seal the archive, update status/log and commit `test: record Go context and concurrency interaction`.

Resolved checkboxes2–4 mean mandatory skipped branches, not performed checks. Completed: eligibility failed for both candidates. Steps2–4 behavioral preparation/authors/review are explicitly skipped; no fixture or integrated launch. The decision record is verified and committed.

## Task 6: Exact-byte package delivery and independent closeout

**Files:** Copy only accepted final snapshots to runtime destinations; conditionally modify `plugins/go-quality-build/{plugin.json,.claude-plugin/plugin.json,README.md}`, `.claude-plugin/marketplace.json`, root `README.md`, distributed notices; create `results/<date>-context-concurrency-delivery/`; update canonical record/plan.

**Interfaces:** Consume individual/interaction dispositions, exact accepted hashes and preserved raw archives. Produce a truthful package version/scope, independent whole-branch review, portable reconstruction/seal records and explicit measured/unverified handoff.

- [x] **Step 1: Promote only accepted exact bytes.** Verify draft/snapshot/runtime equality for all retained files. Accepted additions make both manifest versions 0.3.0 with matching scope/descriptions; neither accepted leaves runtime 0.2.0. Preserve five existing runtime skills and review contracts; references/drafts stay uninstalled unless independently accepted.
- [x] **Step 2: Verify package, evidence and claims.** Run shared checks below, runtime quick validation/local links and all new reconstructions/hashes. Confirm actual native loading separately from manifests. Review READMEs for bounded benefit, strong ties, failed attempts, assisted repairs and missing reuse/toolchain/platform coverage; do not describe intended checks as completed.
- [ ] **Step 3: Obtain independent whole-branch delivery review.** Review exact guidance, source/notice obligations, triggers/ownership, individual/combined evidence, slot ledger, metadata and continuation state. Account for all nine Go topics on fixture outputs; package/spec review remains a distinct ungraded decision. Adjudicate every declined judgment with evidence/cost/limit. Fix supported findings; guidance changes require affected behavioral gates, while documentation-only fixes get scoped structural rechecks.
- [ ] **Step 4: Preserve portable seals and external anchors.** Verify original+patch reproduces every final source hash; hash dispatched guidance/inputs/raw artifacts. Seal each completed archive once immutable, excluding only its own checksum index; record that index hash outside the archive in the maintained canonical record. Preserve approved raw whitespace exceptions by exact path; never rewrite captured evidence to pass formatting.
- [ ] **Step 5: Verify final state, commit and hand over.** Repeat only affected checks after fixes/sealing, confirm runtime equality and clean intended Git scope, complete plan checkboxes/decisions and canonical status/log. Commit `docs: deliver Go context and concurrency study`. Retain managed worktree/local commits; remove only this plan's disposable scratch copies after durable evidence verification. No publication/global installation.

Prepared for independent review: exact five-skill0.2.0 comparator/package unchanged;31 portable reconstructions, archived failures/counterevidence, shared checks and discovery/skip seals verified. Step3 review, Step4 final delivery seal and Step5 closure remain.

## Verification commands and expected results

Run within each disposable fixture module, with a task-specific writable cache variable/path rather than repurposing system variables:

```sh
rtk proxy go version
rtk proxy go build ./...
rtk proxy go test -count=1 -timeout=30s ./...
rtk proxy go vet ./...
rtk proxy gofmt -l .
rtk proxy go test -race -shuffle=on -count=3 -timeout=60s ./...
```

Expect exit 0/no unformatted files for clean conforming candidates; meaningful failing requested-behavior/mutation assertions are separately recorded. Apply race/repetition to reachable concurrent paths, not unrelated pure controls. Execute actual child binaries for CLI claims and faithful HTTP/client/body boundaries as described. Run actual Go 1.22 and an available supported newer compiler when feasible, recording executable/version/effective module settings and unavailable capability. Keep waits at most 60 seconds and poll only running operations.

Run from the repository root at fixture freeze, promotion and delivery as applicable:

```sh
rtk proxy python3 -B -m unittest discover -s tests -p 'test*.py'
rtk proxy python3 -B -m unittest discover -s tests/go-quality-build -p 'test_*.py'
rtk proxy python3 -B tests/go-quality-review/test_layout.py
rtk proxy python3 -B tests/go-quality-review/go-quality-report/evals/test_grade.py
rtk proxy python3 -B scripts/validate.py
rtk claude plugin validate ./plugins/go-quality-build
rtk claude plugin validate .
rtk proxy git diff --check
rtk proxy git diff --cached --check
```

Current unit counts: root 6, build-eval 10, review layout 1, grade 12. Validator review scope stays 10 skills/120 cases; build counts progress 38 → 42 → 46 → 47 if interaction exists. No validator change is expected: it already accepts candidate suites without runtime skills. Modify validation only for a reproduced gap, with a failing meaningful test first. Resolve the skill-creator `scripts/quick_validate.py` from its actual installed skill location and run it against each draft/promoted runtime folder; expect `Skill is valid!`. Structural success does not establish behavioral utility or native loading.

## Execution handoff

The plan is approved for native execution. Both execution methods require fresh independent behavioral authors and reviewers, separately counted from fixture/draft implementation work. Current task/step state is recorded above and in the canonical progress record.

Recommend **native implementation**: the six tasks share a frozen catalog, evidence protocol and cumulative slot ledger, so retaining that context reduces repeated setup. I implement the fixtures/drafts/bookkeeping in this session and an independent reviewer checks final delivery; behavioral authors, outcome reviewers and promotion reviewers remain independent at the stated gates. **Subagent-driven implementation** instead uses a fresh implementer and reviewers at each task boundary, with the same evaluation gates and budget. Obtain plan review and execution-method selection before Task 1.
