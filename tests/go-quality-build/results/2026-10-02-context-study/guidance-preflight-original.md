# Context draft guidance preflight — 2026-10-02

**Disposition: necessary wording fixes before behavioral snapshot freeze.** Two bounded alignment gaps remain. The core cancellation, error, budget and cooperative finalization procedure is sound in the inspected scope. This is an ungraded guidance review; no implementation author trial, benefit claim, runtime promotion or selection claim was performed.

## Exact scope

Reviewed `tests/go-quality-build/go-context-and-deadlines/draft/SKILL.md`, SHA-256 `c81936129bb3d8fc582cce191c948c967e78daeb6f0af23abb96dce51593dc48` (4,783 bytes, 659 whitespace-delimited words). Read the canonical design record's decisions/status/continuation protocol and latest context-stage log; the accepted context/concurrency spec; the section-level source audit and revision inventory; all five exact runtime writing skills; the review guide and relevant unchanged Architecture, Correctness, Code Quality, Resilience, Performance, Testing and Dependencies decision references, plus relevant Security resource-limit and Deployment shutdown/identity sections. Runtime and review skill directories have no tracked diff.

The five runtime skills remain the comparison baseline: package boundaries, API contracts, interfaces/composition, behavior tests and test isolation. Error/value drafts are not prerequisites. The canonical record currently permits a focused context draft after discovery; it does not establish this draft's behavioral effectiveness.

## Required fixes

### G1 — Name explicit cancel-ownership transfer

Draft line 27 says: “The creator owns cancellation on every path.” The accepted spec's context procedure (line 39) expressly includes early exits **or explicit transfer**. As written, the rule does not distinguish a locally owned stage from a context whose lifetime is transferred with a returned stream/run/body. Following it as a creator-local `defer cancel()` can stop the returned work before the recipient uses it, even though the current supported API intends the recipient to own that lifetime. This is a concrete scope/lifetime ambiguity; the rest of line 27 correctly requires a context to remain live through governed work.

A minimal correction is sufficient: the creator is the default owner; if the contract transfers the scope, identify the recipient and the recipient's cancellation/completion obligation. Preserve the existing rule for local stages. No new wrapper/interface or asynchronous architecture is requested.

Reviewer-owned executable evidence is in `/private/tmp/go-context-guidance-preflight-checks/main.go`: `prematureCreatorCancel` returns an already-canceled child after a creator defer, whereas `liveTransferredScope` returns a live child and its recipient cancels it. Both observations passed on actual Go 1.22.12 and Go 1.26.5. This demonstrates the consequence of the ambiguous reading, not a generated-code outcome or a claim that every reader will misinterpret it.

### G2 — Retain the accepted context-value boundary locally

The draft mentions retained values only in `WithoutCancel` at line 31. It never states the accepted spec's line 44 rule or the source audit's line 34 disposition: context values carry appropriate needed request metadata, rather than dependencies or optional configuration. This was selected context-owned guidance, not omitted material. The draft is intended to work standalone, and the existing composition skill's dependency visibility rule does not replace this precise context-value boundary.

Without the rule, a user of this skill receives no decision aid for transporting a client/configuration through values during propagation or detachment. That hides dependency and lifetime ownership and conflicts with the collection's explicit configuration/composition decisions. Add one compact local sentence restricting values to needed request metadata and keeping dependencies/optional configuration explicit. A values/tracing manual, telemetry framework or reference file is unnecessary. The [official context overview](https://pkg.go.dev/context#pkg-overview) independently supports the request-data/optional-parameter distinction.

## Other inspected decisions

| Requested dimension | Assessment |
| --- | --- |
| Discriminating trigger | Clear context/budget/cause/finalization changes; pure calculation and synchronization-only exclusions are useful. No incidental `ctx` trigger or forced contexts on every helper. |
| Cancellation/result observation | Entry/empty, next admission, accepted prefix and completed-success boundaries are visible. Sampling `Err` once before reading `Cause` avoids contradictory observation decisions; the policy does not require cancellation to win every result. |
| Errors and legal custom causes | Classification and custom cause are distinct; joining/exposure follows the promised contract. No arbitrary interface equality. `errors.As` handles a legal slice-backed cause in the check. No error/value draft dependency. |
| Total and stage budget | One operation scope, narrower stage deadlines and an earlier caller bound are explained. Stage cancel timing covers body consumption/closure. No fixed duration or automatic reader/goroutine termination claim. |
| Finalization | Detachment requires accepted work, a stated new bound and cancel owner. The snippet is synchronous and returns actual cooperative completion/error. Non-cooperation and started AfterFunc callbacks retain explicit completion obligations. [Context API documentation](https://pkg.go.dev/context#AfterFunc) confirms stop does not join; [WithoutCancel](https://pkg.go.dev/context#WithoutCancel) confirms removed inherited stop signals/cause. |
| Supported Go claims | The explicit Go 1.21 detachment claim is correct; the exact snippet executes on actual Go 1.22.12 as well as 1.26.5. No minimum bump is demanded. |
| Standalone portability | Only SKILL.md exists in the draft; no absolute host paths, local Markdown dependencies, absent-skill requirements, harness/tool/persona instructions or automatic delegation. G2 is the local knowledge gap to fix. |
| Owner overlap and scope | Context owns propagation, budget and cause choices. API/error exposure and general test/synchronization/telemetry decisions are left with their owners. Borrowed-client/body reminders are necessary local lifetime constraints and agree with composition/resource references. No new architecture, third-party dependency, coordinator or retry framework is mandated. |
| Sources/notices | Wording and the small synchronous helper are consistent with the audit's original-expression/zero-copy disposition. No substantial upstream copy is apparent or declared. Keep exact audited source pins/inventory in evidence; any later substantial copied prose/example needs its applicable distributed notices and a changed copy disposition. |

## Checks and limits

- `rtk proxy python3 /Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py tests/go-quality-build/go-context-and-deadlines/draft` — exit 0, “Skill is valid!”
- Structural inspection — description 265 characters; one draft file; zero local/absolute runtime references.
- Reviewer check program extracts the exact draft function into an ordinary Go 1.22 module. `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=... <toolchain>/bin/go run .` — exit 0 on actual Go 1.22.12 darwin/arm64 and Go 1.26.5 darwin/arm64. Checks cover canceled-parent detachment, retained metadata, a live bounded finalization scope, actual returned finalization failure, defer-cancel lifetime, non-comparable cause inspection, shared/earlier absolute deadlines, transfer counterexample, and stop-versus-completion for AfterFunc. Raw source/module and verification metadata remain in the temporary check directory. These are semantic/example checks, not author trials.
- `rtk git diff -- plugins/go-quality-build/skills plugins/go-quality-review/skills` — empty. Candidate/runtime/reviewer skills were not edited by this review.

Native routing, useful model/harness reuse, improved code outcomes and promotion gates remain for the declared exact-byte study. These checks cannot establish those claims.

## Minor/deferred improvements

1. State the spec's child-derivation predicate in a few words: an actual narrower budget or independently owned cancellation scope. Current text gives the ordinary budget case correctly; this clarification discourages redundant context layers.
2. Optional compact version markers for `context.Cause`/`errors.Join` (Go 1.20) and `context.AfterFunc` (Go 1.21) would make the existing minimum-preservation rule easier to apply to older projects. This is not a Go 1.22 compatibility defect.

Next action: make G1/G2's small local wording corrections, recheck those exact bytes, then freeze the behavioral snapshot. Keep this guidance verdict separate from the neutral graded code-outcome reviews. The root implementer should link/archive this report and record its disposition/checks in the canonical status and dated work log; no concurrent canonical-file mutation was made by this reviewer.
