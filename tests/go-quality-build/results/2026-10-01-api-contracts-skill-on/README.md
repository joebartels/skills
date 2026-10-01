# API-contract skill-on evidence

First-pass and targeted revision evidence. Promotion is a separate controller
decision; the archive distinguishes the exact skill revision and review scope.

## Provenance and reconstruction

Four fresh gpt-6-luna collaboration agents used `fork_turns=none`, no reasoning
override, and disposable copies of fixtures committed at `39c702a`. Expected
judgments were withheld. Three applicable cases were instructed to read the
draft; this is not automatic routing evidence. The private control received
the skill path and description and was instructed to decide before opening.
It selected and opened the skill, which is a false positive for this fixture.
Unlike its controller-completed baseline, this control is a fresh model trial.

`first-pass-SKILL.md` is the exact evaluated snapshot; `manifest.json` records
its SHA-256, input revision, output file hashes, model/harness and selection
provenance. Each case directory contains the exact evaluation task prompt,
unedited author report, source-only patch and fresh verification output. Full
dispatch wrapper bytes were not saved separately; the manifest records the
controller-supplied summary and this limitation, without inventing an exact
wrapper. Both independent blind reviews are copied verbatim.

All four patches reconstruct all 19 resulting source files byte-for-byte. All
12 fresh ordinary tests, uncached verbose tests and vet commands pass. These
checks do not erase defects independently found in migration docs or CLI
compatibility. Execution used Go 1.26.5 darwin/arm64 for Go 1.22 modules; the
minimum toolchain and other platforms remain unverified.

## Outcomes and limits

| Case | First-pass Architecture / Correctness | Comparison |
| --- | --- | --- |
| source-compatibility | A / A+ | Baseline B/B zero-value regression fixed; exact constructor function type and explicit-empty behavior also preserved. |
| intentional-break | A / A- | Baseline A/A; first-pass migration guide misstates old leading-plus behavior (B-F1). Parsing and approved v2 API work. |
| wire-contract | A / A- | C-F1 leading-dash filename defect also reproduced in baseline; original baseline A grade missed this check. Baseline reviewer addendum revises A to B; severity differs from first-pass A-. |
| not-public-contract | Not applicable / A | Correct local fix, but the agent selected and opened an irrelevant skill. Baseline was controller-completed, so no independent baseline selection comparison exists. |

Source compatibility improves on the same-model baseline for one bounded
case. The migration documentation defect is introduced in that first-pass
output. The private selection result is a routing failure despite correct
code. These findings warrant targeted revision and fresh comparisons before
promotion; they do not establish broader skill effectiveness.

`filename-comparison.json` records the identical executable probe using a real
file named `-jobs.json` containing `[]`. Original exits 0 with jobs:null;
baseline and first-pass skill-on both exit 2 with no stdout. Thus C-F1 is a
shared failure to preserve the original CLI contract, not evidence that the
skill caused an otherwise absent regression. Preserve original baseline
review grades as reported, and read them with the preserved baseline reviewer
addendum. The baseline's diagnostic also includes flag's recovered
String-on-zero-state panic message; no additional grade is inferred here.

The current runtime skill and canonical design docs were not edited during
archival. Controller owns the canonical progress update and revised trials.


## Revised trials

`revision-2/` preserves its exact skill snapshot, two fresh gpt-6-luna trials,
patches/hashes/checks and the unedited blind v2 review. Revised migration
wording correctly describes the old acceptance of `+80`; independent review
awarded A/A. The private selection agent decided from the revised description
that the skill did not apply and did not open it. This is instructed selection
evidence, not automatic harness routing. No independent code grade is inferred
for that revised private result.

`revision-3/` preserves the final supplied skill snapshot, fresh wire and v2
trials, reconstructable artifacts and unedited targeted reviews. Both receive
A/A. The wire review independently confirms preservation of real filenames
`-jobs.json`, `--state`, `--state=ready` and `--`, along with filtering, JSON and
process contracts. The final v2 review verifies parsing, CLI and accurate
old-to-new migration claims.

Revision-3 reviews are targeted and nonblind: prior defects were known. The
wire reviewer additionally disclosed accidental author-report exposure. Its
conclusions rely on inspected code and independent executable probes, but it
cannot be represented as a blind review. The final v2 review is likewise a
targeted follow-up; its provenance is recorded without implying blindness.

The four revised patches reconstruct all 23 resulting source files byte for
byte; their 12 fresh ordinary/uncached tests and vet commands pass. Combined
with first pass this archive contains eight reconstructed trials, 42 output
file hashes, and 24 passing test/vet commands.

### Comparison by confirmed behavior

- Source compatibility: first-pass skill use preserved exported zero state and
  explicit empty configuration, resolving baseline A1 while retaining the
  function-value consumer contract (B/B to A/A+ in the recorded reviews).
- Intentional break: revision 2 and revision 3 retain the approved v2 API and
  correct the first-pass migration misstatement; both reviewed A/A.
- CLI compatibility: baseline addendum B/moderate and first-pass A-/minor are
  judgments of the same leading-dash filename defect. Revision 3 removes that
  reproduced defect and passes wider argument checks; grade arithmetic alone
  is not the improvement evidence.
- Selection: revised private-control description prevents the observed false
  positive in one instructed trial; automatic routing remains untested.

The source case was not repeated under the final revision because the agent
thread limit prevented a further trial. Its A/A+ result belongs to the exact
first-pass snapshot. Later instructions add routing, migration and CLI clauses,
but final-revision source behavior is not directly evaluated. This coverage
limit, single trials, unavailable real downstream consumers, Go 1.22 execution
and platform limits constrain any promotion claim. Current runtime and design
docs were not edited by this archive task; no commit was created.
