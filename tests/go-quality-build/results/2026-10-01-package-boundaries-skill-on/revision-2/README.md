# Revised medium trial

The controller added guidance about preserving legacy facades by delegating to the same concrete adapters, then ran a fresh gpt-6-luna medium-service trial. The exact uncommitted skill bytes are preserved in `skill-snapshot/`, their SHA-256 values in [manifest.json](manifest.json), and their difference from `05a4e6095bd8130647c5570f94319b739105cd18` in `skill-change.patch`. This revision is distinct from the seven first-pass trials.

The original user task is in `gpt-6-luna/medium-service/prompt.txt`; its unedited author report, source-only patch and fresh raw verification output are alongside it. The patch was applied to a disposable original fixture and every reconstructed Go/module file compared byte-for-byte with the trial. Fresh tests, uncached verbose tests and vet all pass.

The unedited [blind independent review](independent-review.md) grades Architecture A and Correctness A-. Source inspection confirms a reusable operation without concrete adapter imports and a retained facade delegating to the same implementations. This removes the confirmed ownership problem from the same-model supplemental baseline (Architecture B) and the duplicated legacy implementations in the first skill-on pass (Architecture B). This one revised case demonstrates an observed architecture improvement. It does not establish broad reliability or explain outcomes in already-correct standard cases.

The minor Scanner long-line defect remains in all three medium arms, so no correctness improvement is claimed. The reviewer independently reproduced it despite passing supplied tests. Additional HTTP compatibility checks passed; full CLI exits were inspected rather than subprocess-tested. See the review for complete limits and environment retries.

## Fresh repeat

A second independent gpt-6-luna trial used the same revised skill and original fixture in a fresh context (`fork_turns=none`, no reasoning override supplied). Its [manifest](repeat/manifest.json), exact prompt, source patch, unedited author report, and fresh verification are under `repeat/`. The [unedited independent review](repeat/independent-review.md) again grades Architecture A and Correctness A-. The operation owns validation and ordering, concrete implementations each have one owner, and the retained facade composes them. The same minor Scanner issue remains independently reproduced.

Both revised implementations therefore remove the confirmed architecture problem observed in the same-model baseline and first pass. Two successful observations on this one task offer a limited stability check; they do not establish effectiveness across Go projects. The repeat patch reconstructs byte-for-byte and passes fresh ordinary tests, uncached verbose tests and vet. Its reviewer separately verified compatibility and long-line behavior, with limits retained in the report. Candidate-d was matched byte-for-byte to the repeat output before associating the review.

Verification used Go 1.26.5 darwin/arm64 with Go 1.22 declarations, a disposable cache and approved local httptest listener access. No minimum-toolchain, race or cross-platform matrix was run. Wrapper provenance limits and exact source hashes are recorded in the manifest. No archived implementation was corrected.
