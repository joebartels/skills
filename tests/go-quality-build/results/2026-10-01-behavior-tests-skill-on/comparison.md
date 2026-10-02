# Matched behavior-testing comparison

Revision 1, committed at `3bf12e8ef0a4c9740e53e8fe5b39a1c08af94a2b`, has completed all five fresh author trials and five separate anonymized outcome reviews. Neither arm received isolation-writing guidance, expected assertions, controller probes, another implementation or review feedback. The same task, input bytes, existing architecture/style snapshots and inherited author settings were held constant. Exact model/reasoning IDs are unavailable through this harness; no comparison with historical luna trials is implied.

| Case | Baseline Testing / Correctness / Architecture | Revision 1 Testing / Correctness / Architecture | Confirmed outcome |
| --- | --- | --- | --- |
| library-codec | A+ / A / Not applicable | A+ / A / Not applicable | Independent wire/value oracles and useful properties remain strong; no measured uplift. |
| cli-partial-failure | B / C / Not applicable | A+ / A+ / A | Genuine partial failed duplicate write now preserves the accepted value; its new test detects direct truncation. |
| service-publication | A+ / A+ / A | A+ / A+ / A | Strong rejection/publication checks remain; no matched uplift claimed. |
| worker-host | B / A / A | A+ / A+ / A | Independent work-error/successful-release assertions catch the error-dropping mutation that survived the baseline. |
| not-behavior-work | Not applicable / Not applicable / Not applicable | Not applicable / A / Not applicable | Behavior skill unopened, only doc comment changed. Different applicability judgments about documentation are not uplift. |

Each raw card records its own scope, safeguards, finding counts and limits. Grades are not averaged. The CLI and worker close two unique test gaps; the CLI additionally prevents a confirmed production retention defect. The other three cases establish bounded preservation and appropriate non-selection.

## Regression sensitivity

All seven pre-dispatch semantic mutations compile and fail meaningful candidate-test assertions in both arms: coordinated codec swap, CLI success exit, CLI prefix loss, missing mirror Label validation, direct mirror truncation, missing host join, and start-relative recurrence. Their detection is preservation evidence, not improvement.

The two additional diagnostics were identified after baseline review, before drafting, and are labelled accordingly. They are not presented as unseen frozen probes:

- The baseline worker suite passes a mutation returning nil after successful release, losing callback errors. Revision 1 fails explicit callback-only error assertions (and corroborating recurrence/cleanup assertions) for the same semantic change. The blind reviewer separately verified callback-error sensitivity.
- Under child-only RLIMIT_FSIZE=3 with SIGXFSZ ignored, the same controller input `web,1\nweb,12345\nlater,9\n` leaves baseline `web=123`, violating its accepted-prefix contract. Revision 1 exits 2, keeps exact `web=1\n`, and writes no later record. Its own supported-platform child-process test rejects a direct-write mutation at the retained-value assertion. The blind reviewer independently reproduced partial-write sensitivity.

All candidates reconstruct byte-for-byte from frozen inputs and patches. Normal test/vet/gofmt and applicable race/ten-shuffle checks pass; all four controller contract probes pass. Reviewers performed additional finite mutations without repairing sources. No new third-party dependency or public testing abstraction was introduced. The CLI adds a small private resource-owning publication helper because its actual retention contract requires complete replacement.

## Effort and limits

The skill is 613 words plus a conditional 742-word reference. Authors had one fresh implementation attempt each; no guided repair or draft revision was needed. Tool-call/token accounting is not exposed, so no speed/cost advantage is claimed. More test cases or lines are not quality evidence. Clear sensitivity differences in two tasks do not require an ambiguity repeat under the approved gate; broader effectiveness still needs more tasks.

Runtime checks used Go 1.26.5 on darwin/arm64, with modules declaring Go 1.22. Actual Go 1.22 execution remains unavailable. Some reviewers cross-compiled Linux/Windows tests; that establishes compilation only. Partial-write and old-handle checks have explicit supported-platform scope. Staticcheck was unavailable. A shared writable Go cache was used; the codec reviewer disclosed supplemental cached fuzz inputs and separately verified committed-seed sensitivity. Fuzz throughput is not credited as uplift.

The baseline service author substituted an in-process transport after a sandbox bind denial. The skill-on author and controller retained bind-denial output, then ran unchanged real-loopback tests with automatically approved local listener access. This difference is not credited as matched uplift; real HTTP is verified only for this candidate on this host. See [environment note](service-publication/environment-note.md). No external integration, global installation, automatic routing, or Codex/OpenCode runtime-loading result is implied.

## Promotion decision

The evidence supports bounded promotion of unchanged revision 1: all four applicable authors opened it, the control declined it, two confirmed baseline gaps were closed, and no harmful regression was substantiated in the matched cases. The [independent task-level review](task-review.md) supports unchanged promotion after a source-audit omission is corrected. Root completed the missing 11 pinned testing-layout sections at `7403956`, bringing the audit to 79 decisions without changing evaluated guidance. The exact draft/snapshot/runtime hashes match; the live draft was removed. All shared delivery checks passed. Source wording/examples are original; no substantial upstream copying or new notice obligation was found.
