# Completed blind baseline comparison

| Case | Testing | Correctness | Architecture | Verified outcome |
| --- | --- | --- | --- | --- |
| library-codec | A+ | A | Not applicable | Independent wire/value/error assertions catch coordinated field swaps. |
| cli-partial-failure | B | C | Not applicable | Strong streaming/process/prefix controls; missing failed duplicate partial-write observation exposes accepted-state loss. |
| service-publication | A+ | A+ | A | Complete field/value/rejection/body/publication assertions catch missing Label validation and direct truncation. |
| worker-host | B | A | A | Strong host join/recurrence controls; rewritten tests omit callback-error/successful-release coverage. |
| not-behavior-work | Not applicable | Not applicable | Not applicable | Comment-only edit preserves behavior and existing tests; no test expansion. |

Five fresh independent authors completed, and all five candidates reconstruct byte-for-byte. Clean test/vet/format checks and applicable race/ten-shuffle checks pass. All four frozen contract probes pass. Candidate tests detect all seven pre-dispatch semantic mutations with clean and compiled gates. These are strong baseline controls, not evidence of a new skill's benefit.

Independent blind reviews identify two unique assertion gaps:

1. CLI write failure is tested only before an existing file is opened. A later duplicate write under a child-only three-byte file-size limit returns exit 2 but leaves web=123 instead of the earlier accepted 1\n. The controller independently reproduced it. This is both a new batch data-integrity defect and a test gap, not a request for whole-batch rollback or crash durability. Exact diagnostic inputs/results are preserved in cli-partial-failure/partial-write-controller-confirmation.json and the raw review/harness.
2. Worker error cases always make release fail. A post-review error-dropping mutation passes the whole candidate suite when release succeeds; the independent review's extra probe catches it. Original legacy coverage of this combination was lost during test rewriting. The controller independently confirmed the surviving mutation. Production behavior itself is correct.

The two post-review diagnostics are explicitly separate from frozen probes/mutations and are not retroactively labelled preplanned. No baseline repair or reviewer feedback reached an author. These observations motivate conditional guidance about failure stages, retained state, independent failure combinations and preserving useful existing regression cases.

Limits: author/reviewer settings are inherited consistently but exact model IDs/reasoning metadata are unavailable. This is not a historical same-model comparison. Go1.26.5 darwin/arm64 with Go1.22 module declarations was used; actual minimum-toolchain and other-platform execution remain unverified. Socket binds were denied, so mirror assertions exercise client/transport/recorder seams and Unix file replacement, without proving live network or Windows behavior. Grades remain independent; no averaging or test-count score is used.

Next: write a focused draft outside the runtime package, then compare fresh matched skill-on authors against the unchanged tasks/fixtures and these recorded controls. Nothing in this run proves that a testing-writing skill exists or is effective.
