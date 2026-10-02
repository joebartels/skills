# Language-contract trials, 2026-10-01

The three architecture skills are available identically in both arms. Authors see
only task inputs and catalog descriptions, choose relevant skills explicitly,
and record opened files. This measures explicit exposure, not automatic routing.
Each manifest records input/catalog/output SHA-256 values; `source.patch` applies
to the tracked fixture to reconstruct the completed source exactly. Two independently reviewed candidate drafts now exist outside runtime. No new
skill has been promoted and no measured incremental improvement is claimed.

## Errors: reviewed draft, promotion unproven

The independent [outcome review](../../../../docs/go-quality-build/error-contracts-baseline-review.md)
found no confirmed owned defect in four gpt-6-luna medium outcomes (A/A). The
single bounded supplementary cursor task also passes. The reserved CSV case
was consumed prospectively after drafting as a matched compatibility control.
It is not baseline failure search. Successful
baselines are evidence against claiming measured incremental benefit here.

Two early inherited-model trials are supplemental: their exact model IDs are
unavailable. The explicit gpt-6-luna medium cohort was declared before examining
completed outputs; inherited trials are not silently treated as matched samples.
The pre-draft private-calculation control opened no skills. Final-byte draft
exposure trials are now underway, with one fresh author per case/arm.

| Trial | Author model / effort | Archived verification |
| --- | --- | --- |
| backend-errors/first | gpt-6-luna / medium | verification-final.json: PASS |
| cli-completion/first | inherited root; exact id not surfaced / inherited root, no override | verification-corrected.json: PASS; verification.json: INVALID/FAIL; see correction |
| cli-completion/named | gpt-6-luna / medium | verification-final.json: PASS |
| cursor-iteration/named | gpt-6-luna / medium | verification-final.json: PASS |
| not-error-work/named | gpt-6-luna / medium | verification-final.json: PASS |
| reader-errors/first | inherited root; exact id not surfaced / inherited root, no override | verification.json: PASS |
| reader-errors/named | gpt-6-luna / medium | verification.json: PASS |

The first inherited CLI `verification.json` is invalid: its writer double exposed
a promoted `WriteString` that bypassed failure injection, and default cache access
was blocked. `verification-corrected.json` preserves the rerun with the corrected
double and writable cache. Neither failure is scored as an author defect. Other
`verification.json` files predate the cache fix but contain successful checks;
main new records use `verification-final.json`. Reader `probe-validation.json`
records a deliberately broken private copy losing a read cause; it is probe
validation, not a model failure.

Independent reviewer raw checks and extra probes are under
[expert-review](go-error-contracts/expert-review/), with SHA-256 identities.
The review is independent of authoring, but sees expected contracts/private checks.
It is not blinded to arm labels. Real Go 1.22 execution, other platforms and broad
usefulness remain unverified; checks used Go 1.26.5 darwin/arm64.

## Values

Five original cases are fixed before drafting, including a reserved borrowed-frame
transfer. Four fresh gpt-6-luna medium baselines are archived (three applicable and one private control). All pass the private checks; the control opens no skills. Independent [value outcome review](../../../../docs/go-quality-build/value-semantics-baseline-review.md) grades all applicable changes A/A with no confirmed owned defect. The user's content condition authorizes the reviewed value draft; promotion remains unproven; this is not evidence that value guidance is useless on other tasks. Reserved transfer is now a prospectively consumed matched control after drafting. Additional reviewer checks/probes and identities are under [expert-review](go-values-and-zero-values/expert-review/). No measured added benefit is claimed.

## Draft and reference review

The [exact-byte expert review](../../../../docs/go-quality-build/language-contracts-final-review.md#2026-10-01-follow-up-exact-draft-content-review)
approves the error/value drafts and names/docs reference for candidate evaluation.
Drafts remain outside runtime; content sufficiency is separate from measured
incremental effectiveness. Final hashes and reading cost are recorded. Raw
compiled-example/semantic checks and source are in [draft-review](draft-review/).
Names/docs has no standalone skill or behavior suite.

## Combined use

A fresh library/CLI case tests partial caller input errors, nested independent
snapshots, intentional nil/empty results and actual process status/prefix output.
The fixed three-skill catalog is identical; the draft-exposure arm adds both
reviewed candidates. This checks interaction and possible harm; it does not
establish automatic routing or a broad effectiveness estimate. Both arms pass the current library/private checks and five executable CLI scenarios. The draft arm reports opening API contracts plus both candidates; the baseline opens API contracts. No interaction failure is observed in these probes. Blinded expert source/outcome review is underway.

## Final-byte exposure archive

All 27 authored trials are archived and reconstructable: 14 existing-catalog
baselines and 13 draft-exposure runs. These support 12 clean matched pairs; the
extra first value control is non-blind supplemental evidence. Two inherited-model
error baselines remain supplemental because their exact model IDs are unavailable.
Every clean pair uses gpt-6-luna medium and identical original inputs/existing
catalog bytes. There are 105 successful current recorded check commands; the
invalid first CLI verification remains excluded beside its corrected result.

| Draft exposure | Case/repeat | Output files | Current checks | Provenance |
| --- | --- | ---: | --- | --- |
| combined | language-contracts/first | 6 | PASS | Fresh explicit exposure |
| go-error-contracts | backend-errors/first | 4 | PASS | Fresh explicit exposure |
| go-error-contracts | cli-completion/first | 6 | PASS | Fresh explicit exposure |
| go-error-contracts | csv-transfer/first | 4 | PASS | Fresh explicit exposure |
| go-error-contracts | cursor-iteration/first | 5 | PASS | Fresh explicit exposure |
| go-error-contracts | not-error-work/first | 4 | PASS | Fresh explicit exposure |
| go-error-contracts | reader-errors/first | 5 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | borrowed-transfer/first | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | default-wire/first | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | not-value-work/first | 4 | PASS | Non-blind supplemental control |
| go-values-and-zero-values | not-value-work/isolated | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | required-construction/first | 4 | PASS | Fresh explicit exposure |
| go-values-and-zero-values | snapshot-ownership/first | 5 | PASS | Fresh explicit exposure |

The isolated error and value controls report no opened skills. Applicable authors
report opening their candidate; the combined author opens both. Selection remains
self-reported explicit exposure, not automatic routing. Frozen reviewed hashes
match every catalog. One prepared-only error calculation directory was never
dispatched and is not counted as an author attempt; [inventory.json](inventory.json)
records this distinction. Exact task/catalog prompts and author reports are
preserved; outer dispatch messages/environment history are not full transcripts.

A fresh reviewer is assessing 12 paired candidate trees with arm identities withheld
until its initial report is frozen. Original inputs/task contracts and private probes
are visible to the reviewer, so this is arm blinding rather than expectation blinding.
No measured incremental benefit, broad reliability or runtime promotion is claimed.
