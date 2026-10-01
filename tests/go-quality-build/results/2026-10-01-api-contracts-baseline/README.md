# API-contract baseline preparation — 2026-10-01

Three fresh blind gpt-6-luna baseline implementations are archived. The private
helper control was completed by the controller after an automatic approval
rejection blocked the blind author; it is not a model baseline implementation.
Both independent reviews are archived. No API-contract runtime skill has been authored
or promoted.

## Blind dispatch

Read cases from `../../go-api-contracts/evals/evals.json`. Give each fresh author
only its exact `prompt` and a disposable copy of its listed `files`, with the
case-directory prefix removed. Do not supply this directory, `expected_output`,
`assertions`, other cases, or proposed build guidance. Existing tests exercise
ordinary prior behavior; they do not encode the requested change or the hidden
consumer check. Project README/release docs are legitimate input contracts.
Preserve the exact dispatch wrapper, model/harness, source revision and input
hashes. Run same-model baseline/skill-on pairs; independent reviewers receive
the original task, project contracts and diff with evaluation assertions and
condition labels withheld. Capture their findings before comparing the cases.

## Cases

- `source-compatibility`: stable minor-release formatter configuration. The
  documented external factory consumer is absent from the author checkout.
- `intentional-break`: accepted v2 module/API change with a CLI consumer and
  precise grammar; preserving the old signature is not the requested outcome.
- `wire-contract`: CLI state filtering and internal field rename under a
  documented JSON/exit contract, including empty results and empty-state flags.
- `not-public-contract`: private deadline comparison; a non-selection control.

The evaluator-only `probes/source-compatibility/factory_test.go` is a real
external-module consumer. Put it in a temporary module named
`example.com/reportconsumer`, require `example.com/recordfmt v1.4.2`, and add a
local replace to the evaluated fixture. Run `go test ./...` unchanged against
baseline and skill-on outputs. Do not copy it into author workspaces. For the
other contracts, reviewers should verify behavior from task and project docs,
not from a prescribed solution or expected grade.

## Preparation evidence

- `fixture-manifest.json`: toolchain and SHA-256 of all 18 original fixture files.
- `fixture-checks.json`: exact stdout/stderr and exits for ordinary tests,
  uncached verbose tests, and vet in all four modules (12 successful commands).
- `source-consumer-check.json`: the separate consumer compiles and runs against
  the original source fixture.
- Validator TDD: new candidate-suite test failed (`([], 0) != ([], 1)`), then
  all 9 tests passed after discovering present candidate suites as well as
  required installed-skill suites. The unchanged repository first retained
  5 build cases; with API fixtures it validates 9. Candidate data does not
  install or promote a runtime skill.

Go 1.26.5 ran modules declaring Go 1.22; minimum-toolchain execution and any
behavioral benefit remain unverified. Next: controller dispatches blind
baseline implementations, archives reconstructable changes and verification,
and obtains independent Architecture/Correctness review before skill authoring.

## Archived baseline implementations

`manifest.json` records the per-case artifacts and output hashes. Each case
directory contains its exact task prompt, unedited author/controller report,
source-only patch and fresh verification output. All four patches reconstruct
all 19 resulting source files byte-for-byte; all 12 ordinary/uncached test and
vet commands pass. The withheld external function-value consumer also passes
against the source-compatibility result (`external-consumer-check.json`).

The three applicable cases used fresh gpt-6-luna collaboration agents with
`fork_turns=none`, no reasoning override, and build guidance and expected
judgments withheld. Exact case prompts are saved; complete wrapper bytes were
not preserved separately. The manifest records the controller-supplied known
prefix and summarizes additional instructions without presenting them as an
exact reconstruction.

The `not-public-contract` agent made no edit because automatic approval review
rejected its disposable temporary copy as unrelated to the authorized skills
workspace. The controller verified the four input files against the fixture,
retried the direct patch successfully, and added the equality regression. Its
unedited report and patch are archived for transparency. This control cannot
establish blind model performance or successful skill non-selection.

Library and CLI reviewers independently reviewed anonymized candidates
with expected judgments and author reports withheld. Reports and grades are
archived; no skill uplift is claimed. Docs and runtime skills are unchanged in
this evidence-packaging stage, as requested by the controller.


### CLI/control independent review

The unedited `review-cli.md` assesses the CLI as Architecture A / Correctness A
and the private helper as Architecture Not applicable / Correctness A. These
are outcome grades, not uplift evidence. The private result retains the
controller-completed provenance limit. Library/v2 findings and grades appear below.


### Library/v2 independent review

The unedited `review-library.md` assesses source compatibility as Architecture B
/ Correctness B and the intentional v2 break as A / A. Finding A1 is one shared
moderate root cause, not two defects: the new separator storage makes an
exported zero-value Formatter emit `keyvalue` instead of its previous
`key:value`. The external factory assignment still works. The reviewer ran the
same zero-value probe against original and changed libraries, observing pass
and fail respectively. Its unedited probe is retained as
`probes/source-compatibility/zero_test.go`; it was added after blind authors
finished and must remain outside future author inputs.

The reviewer qualifies the contract judgment: no actual downstream zero-value
consumer was supplied; it rests on usable public state, absence of a
constructor-only precondition, and the minor-release preservation policy. This
is a specific compatibility failure for future guidance to address, not a
claim that every module must preserve every behavior. Intentional v2 and CLI
outcomes are useful successful controls. Runtime authoring and same-model
skill-on comparisons remain next; there is no uplift evidence yet.


### CLI compatibility addendum

`review-cli-addendum.md` preserves the original baseline reviewer's targeted
reassessment; `review-cli.md` remains unedited. An identical real `-jobs.json`
invocation succeeds in original and fails in baseline. The reviewer revises
baseline Correctness from A to B (one moderate C-F1), with Architecture A
unchanged. The first skill-on reviewer graded the same failure minor/A-.
These are differing severity judgments about one root cause, not evidence of
skill-on improvement. The same executable comparison is saved in
`../2026-10-01-api-contracts-skill-on/filename-comparison.json`.
