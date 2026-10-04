# Error-contract candidate evaluation — 2026-10-03

Shelve `go-error-contracts` as a standalone runtime skill in this pass. Eight
fresh authors completed the bounded tasks without the error draft; all held
contract checks passed. Independent source review found no introduced contract
defect or unnecessary public mechanism to target. This does not demonstrate an
incremental benefit that would justify another routed skill.

The existing draft and eight runtime build skills remain unchanged. No new
error guidance was written, exposed, installed or promoted. Eight prepared
guided trials were skipped under the agreed early-stop gate; no paired effect
or added-skill trigger accuracy is claimed.

| Task | Fresh baseline authors | Observed contract behavior |
| --- | ---: | --- |
| Backend translation | 2 | Public classifications and safe context; hidden backend text, types and identity; preserved direct missing sentinel |
| Reader partial results | 2 | Valid prefix, bytes with errors, physical line metadata, and simultaneous parse/read causes |
| CLI completion | 2 | Direct lone failures, independent processing/Close causes, ownership, positional limits and actual CLI behavior |
| ZIP transfer | 1 | Selected contents/order, central-directory completion, independent causes and repeated sticky output failure identity |
| Unrelated control | 1 | Private pagination fix without an error mechanism |

The six discovery trials used historically informed fixtures with two fresh
authors per task. ZIP was reserved before dispatch as a distinct transfer case.
The malformed-data/read-error contract was clarified before any author ran.
Inputs, current build guidance and held probes were frozen; fresh authors were
given the task, existing guidance catalog and no historical outputs or probes.
They reported selected/read guidance. Global process/style skills remained
available. This measures behavior without this candidate, rather than behavior
without all Go guidance.

One reviewer inspected requirements, original/completed source and probes
under neutral IDs, without author reports or exposure labels. The first six
assessments were frozen before the final two were reviewed. Each applicable
code-quality/correctness assessment is A; backend exposure security assessments
are A and other security topics are not applicable. These scoped grades are
not averaged into an effectiveness score.

Two harness issues were preserved rather than counted as author defects. The
controller initially expected CLI status 1 for a negative limit; the original
invalid-usage contract permits the observed status 2 before file I/O. The
reviewer's first backend diagnostic counter also counted its fake's own wrapper
construction; corrected disposable harness checks passed without source changes.

Evidence: [plan](plan.json), [inputs and source hashes](manifest.json),
[sample patches/checks](samples/), [initial review](blind-review/report.md),
[supplemental review](blind-review/supplement-report.md),
[CLI adjudication](controller-adjudication.json),
[repository validation](repository-validation.json), and
[integrity check](integrity.json). Patch replay reconstructs each completed
source exactly. Candidate tests, held probes, vet and formatting passed; the
reviewer also ran independent race tests. Repository validation, 37 Python
tests and three Claude manifest validations passed.

Limits: Go 1.26.5 on darwin/arm64 was exercised, with modules declaring Go 1.22;
native Go 1.22 verification and other platforms were not exercised. Staticcheck
was unavailable. Exact inherited model/effort metadata and automatic harness
routing were not exposed. Full author tool traces were not archived, and the
reviewer's overwritten initial failing harness source could not be recovered;
its actual failures, correction and reruns remain in the check evidence. This
bounded sample cannot establish universal ineffectiveness or the effect of
guidance that was never dispatched.
