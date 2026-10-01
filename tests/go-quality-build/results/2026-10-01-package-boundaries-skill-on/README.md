# Package-boundary skill-on first pass: 2026-10-01

Status: first-pass evidence archived without distinct architecture benefit; two fresh revision-2 trials demonstrate improvement on the medium case.

Update: the [separate revision-2 trials](revision-2/README.md) are archived, including a fresh repeat. Both gpt-6-luna medium results improve Architecture from baseline B and first-pass B to A; Correctness stays A- with the same Scanner issue. This is bounded evidence on one task, not a broad reliability or causal guarantee. The table below remains the unchanged first-pass comparison.

Seven fresh Codex collaboration subagents used skill revision `05a4e6095bd8130647c5570f94319b739105cd18`: five `gpt-6-sol` cases and two `gpt-6-luna` cases. Fork context was `none`; no reasoning-effort override was passed. The [manifest](manifest.json) records exact task prompt paths, model/harness provenance, reconstruction hashes, checks, candidate-to-review mapping, and comparison with each archived same-model baseline. The full wrapper's exact bytes were not retained in available context; its known invariants are recorded explicitly as such.

Each case contains an unedited `trial-report.md`, exact user task in `prompt.txt`, source-only zero-context `source.patch`, and fresh raw command output in `verification.txt`. Patches exclude binaries and author reports. Apply one to a disposable copy of the corresponding fixture with `rtk proxy git apply --unidiff-zero /absolute/path/to/source.patch`. Every reconstructed Go/module file was compared byte-for-byte to the trial output before testing. The skill snapshot preserves the exact committed instructions used in this first pass, independent of later working-tree revisions.

## Same-model comparisons

| Model | Case | Architecture baseline → skill-on | Correctness baseline → skill-on | Interpretation |
| --- | --- | --- | --- | --- |
| gpt-6-sol | small-cli | A → A | A → A | Cohesive command retained; no uplift |
| gpt-6-sol | small-library | A → A | A → A | Cohesive public library retained; no uplift |
| gpt-6-sol | medium-service | A → A | A- → A- | Sound boundary in both arms; Scanner edge remains |
| gpt-6-sol | large-features | A → A | A → A | Appropriate ownership in both arms; no uplift |
| gpt-6-sol | not-package-work | Not applicable → Not applicable | A → A | Local arithmetic remains local; skill not opened |
| gpt-6-luna | medium-service | B → B | A- → A- | New adapters coexist with duplicated legacy implementations; no ownership uplift |
| gpt-6-luna | large-features | A → A | B → A- | Architecture already appropriate; different Scanner defect thresholds do not establish package-skill benefit |

The standard baseline was already Architecture A in every applicable case. The first-pass skill does not improve those grades or reduce a confirmed architecture defect. The supplemental medium baseline's implementation ownership problem changes form: the trial extracts adapters but retains duplicated concrete fallback code. The independent reviewer confirms that the operation still shares maintenance of independently changing implementations. The large supplemental case also remains Architecture A. These observations do not pass the distinct-benefit promotion gate.

The arithmetic author report explicitly states the skill was not opened after reading its path/frontmatter description and deciding applicability. This is an instructed selection control, not evidence of automatic harness routing. Applicable trial reports record their skill use; no automatic selection success is claimed.

## Independent reviews

Unedited blind reports: [small cases](reviews/small.md), [medium cases](reviews/medium.md), [large cases](reviews/large.md). Medium and large Candidate A is `gpt-6-luna`, Candidate B is `gpt-6-sol`; small Candidates 1/2/3 are CLI/library/arithmetic respectively. Reviewers assessed source changes against original fixtures and user contracts, withholding build skill material, expected judgments, author reports and provenance. Reviewer model identity is not recorded in the supplied report artifacts.

Tests and vet pass in all seven reconstructed modules. Fresh checks use Go 1.26.5 darwin/arm64, `GOWORK=off`, a disposable cache and approved local listener access. All modules declare Go 1.22; minimum-version, race and cross-platform execution were not performed. Raw uncached verbose results distinguish actual tests from subtests. Passing supplied tests does not erase the independent long-line failures. Review-specific executable probes, additional disposable reproduction tests, and blocked attempts are described in the unedited reports; the archive worker's fresh test/vet run does not claim to repeat all those probes.

Baseline references: [standard](../2026-10-01-package-boundaries-baseline/README.md), [supplemental](../2026-10-01-package-boundaries-baseline/supplemental-luna/README.md). Single trials provide bounded observations, not a reliability estimate or a causal guarantee of skill impact.
