# Language-contract trials, 2026-10-01

The three architecture skills are available identically in both arms. Authors see
only task inputs and catalog descriptions, choose relevant skills explicitly,
and record opened files. This measures explicit exposure, not automatic routing.
Each manifest records input/catalog/output SHA-256 values; `source.patch` applies
to the tracked fixture to reconstruct the completed source exactly. No candidate
skill has been authored or promoted at this stage.

## Error decision: deferred

The independent [outcome review](../../../../../docs/go-quality-build/error-contracts-baseline-review.md)
found no confirmed owned defect in four gpt-6-luna medium outcomes (A/A). The
single bounded supplementary cursor task also passes. The reserved CSV case
remains unused. Do not consume it to continue searching for failures. Successful
baselines are evidence against claiming measured incremental benefit here.

Two early inherited-model trials are supplemental: their exact model IDs are
unavailable. The explicit gpt-6-luna medium cohort was declared before examining
completed outputs; inherited trials are not silently treated as matched samples.
The private-calculation control opened no skills. No skill-on arm exists.

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
transfer. Baseline work is underway; no value draft or benefit is claimed.
