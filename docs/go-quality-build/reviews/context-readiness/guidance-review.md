# Independent context exact-guidance review

**Date:** 2026-10-02  
**Guidance verdict:** Accepted as a focused, reviewed draft; no remaining material guidance finding.  
**Promotion/evidence verdict:** This guidance-only phase held the verdict pending frozen transfer cards. The final [evidence disposition](readiness-review.md) is reviewed draft, not promoted. Guidance acceptance does not establish benefit or runtime readiness.

The exact reviewed snapshot is [study/skills/go-context-and-deadlines/SKILL.md](../../../../tests/go-quality-build/results/2026-10-02-context-study/skills/go-context-and-deadlines/SKILL.md), SHA-256 `6368240e5a000cd8f056afa5980099317882d8ebcefa0e1f29921975ccc4627d`. Fresh checks confirm the current draft has identical bytes. The draft directory contains only `SKILL.md`; there are no runtime references, scripts, UI files or additional assets.

This independent ungraded review applies the [accepted design](../../../superpowers/specs/2026-10-02-go-quality-build-context-concurrency-design.md), [Task 2 plan](../../../superpowers/plans/2026-10-02-go-quality-build-context-concurrency.md), canonical continuation record, source audit/inventory, original preflight and resolution. It compares the five frozen comparator skills and their relevant references, the unchanged runtime report/topic guidance and its contract-sensitive architecture, idiom, correctness, resilience, resource and testing decisions. It is separate from the neutral graded code-outcome reviews and does not revise their rubric or findings.

## Findings and alignment

No material defect in the frozen guidance requires a text change. Each selected decision has a concrete condition, boundary and owner:

| Dimension | Exact-guidance assessment |
| --- | --- |
| Trigger | Changes to cancellation propagation, operation/stage budget, cause decisions or required finalization activate the skill. Pure calculations and synchronization-only work are excluded. An incidental context parameter or mechanical edit does not meet a described decision trigger. Supported signatures are preserved. |
| Result decisions | Entry, next-work admission, independent failure and completed success are distinct observation points. The entry rule explicitly permits a canceled empty call to succeed when its contract does. Accepted progress and independent errors follow the task's promise; there is no blanket cancellation-wins or final cancellation check. |
| Classification and custom cause | Sampling `Err` once before reading `Cause` makes the observation boundary explicit. Standard classification and a different custom cause remain distinct. Arbitrary error-interface equality is prohibited; legal slice/map-containing errors have a concrete inspection approach. `errors.Join` is conditional on promised multiple-error exposure, agreeing with the API and idiom owners. |
| Parent and budget | The caller scope is propagated when it fits; a child needs a distinct local budget or cancellation owner. One operation parent supplies all stages, and a stage inherits an earlier deadline. The text identifies the concrete full-budget-reset failure. It prescribes no universal duration or timeout/client framework. |
| Cancel/body lifetime | Default ownership remains with the creator until explicit transfer. The named recipient and release boundary prevent premature creator cleanup. Per-stage cleanup avoids whole-loop retention. The live context covers governed body consumption and closure; acquired bodies and borrowed clients have distinct owners. |
| Deliberate finalization | Detachment requires accepted work that the contract says must be finalized. It removes inherited deadline, signal, error and cause, then adds a stated local bound and cancel owner. The small synchronous helper waits for actual cooperative completion and returns its error. Processing and finalization outcomes remain separately owned. |
| Limits of cancellation | The draft expressly limits stopping claims to cooperating boundaries. Neither context acceptance, expiration nor `AfterFunc` registration stopping proves arbitrary blocking work or started callbacks have completed. Uncooperative work requires a remaining-work/ownership decision. |
| Values and versions | Needed request metadata can cross API boundaries through values; ordinary dependencies/configuration stay explicit. `Cause`/`WithCancelCause` and `WithoutCancel`/`AfterFunc` floors are correct. The opening and finalization sections require preserving the project's minimum, rather than a silent upgrade. No timer/reset recipe is introduced. |
| Focus and elegance | Four compact sections, one boundary table and one synchronous helper express the decisions without another layer, public API, dependency, goroutine, coordinator, telemetry package or retry framework. Verification selects applicable observations through existing boundaries rather than requiring a universal case count or new test infrastructure. |
| Standalone use | Necessary cause, ownership, metadata, detachment and version rules are local. Error/value drafts are not prerequisites; there are no absent-skill links, absolute host paths, harness personas/tool lists, automatic configuration or delegation instructions. |

The existing composition skill owns dependencies, host-facing start/stop/completion interfaces and joining owned runs. The context draft adds operation observation order, standard/custom cancellation preservation, total-versus-stage budgets and deliberate bounded finalization. Its stop-versus-completion and borrowed-resource reminders constrain those context decisions; they do not replace composition or synchronization mechanics. API contracts continue to own supported error exposure and API evolution. Behavior tests and test isolation continue to own general assertion sensitivity and fixture/timing design. No tested or textual conflict with those owners is established by this guidance review. Whether the extra guidance adds implementation utility remains an outcome question.

## Preflight disposition

Both required original preflight findings are resolved in these exact bytes:

- **G1, cancel transfer:** The current owner calls cancel on every owned exit; an explicit transfer names the recipient and release boundary. This covers both local stages and returned live work without prescribing a wrapper or changed API.
- **G2, value boundary:** The draft explicitly permits needed request metadata and keeps dependencies/configuration in ordinary parameters or owned fields. Standalone use no longer depends on a neighboring skill for this selected rule.

The optional child-derivation and helper-floor clarifications are also present. This is the independent recheck of the changed frozen hash that the controller's preflight resolution intentionally did not claim.

## Sources and fresh verification

[guidance-verification.json](guidance-verification.json) preserves exact identities, commands, exit statuses, output and limits; [verify-guidance.py](verify-guidance.py) reproduces the checks without author or selection launches.

Fresh verification established:

- The current draft and frozen snapshot match the stated SHA-256. All nine catalog files match the study manifest. The five comparator skills and their three references match both current runtime and original `0d0a339b27cd1ff7ff3cc177f28a9a4455f91a96` bytes. All 23 review-guidance files match that original revision.
- All 20 source/license captures match the section-audit inventory's sizes/hashes at its four pinned revisions. Selected context entrypoint/references were inspected against the actual draft; the wording and small helper are consistent with the original-expression, zero upstream-copy disposition. No substantial imported passage/example requiring a new distributed notice is apparent or declared. The audit remains the source provenance record.
- Official host API-history records confirm `Cause`/`WithCancelCause` in Go 1.20 and `WithoutCancel`/`AfterFunc` in Go 1.21. Primary comments in the official host `context` and `net/http` source support detachment, metadata, stop-versus-completion and response-body lifetime; these corroborate the archived primary package documentation. The minimum downloaded distribution omits API-history files, so it is not represented as the source of those records.
- A fresh disposable module ran the archived reviewer semantic check on actual Go 1.22.12 and Go 1.26.5 darwin/arm64. Before execution, the exact frozen finalization function was checked as present byte-for-byte in that program. Both executions passed the live/bounded detached scope, preserved metadata/error, cancel-transfer counterexample, shared/earlier deadlines, legal non-comparable cause inspection and `AfterFunc` stop/completion observations. These are reviewer demonstrations, not author outcomes.
- Skill-creator quick validation returned exit 0 and `Skill is valid!` for the current exact draft.

These checks establish only the identities, named semantic/example boundaries and structural validity. They do not establish general model consistency, implementation benefit, full package installation, native routing on missing harnesses, other platforms or exhaustive stopping schedules. No author, selection probe, candidate edit, comparator edit, rubric edit or raw-outcome edit was performed by this reviewer.

## Evidence handoff from the guidance-only phase

The disclosed study has consumed 16/16 author attempts: six earlier discovery attempts excluded from the checked matched benefit, eight checked reference/transfer completions, and two pre-model alternate-model failures. The primary checked library pair is a strong implementation tie and both author suites miss the frozen unsafe-equality mutation. No repeated-primary correction can be claimed from those observations; failed launches, source-language checks or selection-only access cannot supply one.

Twelve native selection-only observations across three Codex profiles record applicable body reads, control non-reads and unchanged catalog bytes, separately from the two earlier capability probes. This is useful local discovery evidence. It cannot establish full package installation or alternate-model/reasoning implementation benefit; Claude authentication and target OpenCode v2 loading remain unavailable.

At this phase, final evidence disposition awaited the frozen neutral transfer packet and reconciled comparison, preserving contrary/supplementary findings rather than tuning the draft or choosing a new primary after outcomes. Even a clean transfer would not repair the missing repeated-primary-correction gate. The final [readiness review](readiness-review.md) now records that completed handoff and nonpromotion decision. The parent implementer maintains the canonical status/work log and links both independent artifacts.
