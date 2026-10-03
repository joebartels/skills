# Focused post-fix addendum — NG1 resolved

Date: 2026-10-03. Scope: exact `c5ee3d0e337975d1de199c86c34bbe2f301ccf85` → `d4a0a57a61ee76ab543d87169048111223c66c15`, limited to the NG1 required-delivery-seal correction, its actual CLI regression, snapshot preservation/canonical anchoring, and identities needed to carry the original bounded Go cards. This is an addendum to the [one original independent closing review](report.md), not another full audit.

**NG1/P2 is resolved at d4a0a57. No new actionable issue was found in this focused scope.** The [current gate](/Users/jb/.codex/worktrees/ed95/skills/tests/go-quality-build/results/2026-10-02-context-concurrency-delivery/verify-delivery.py:141) now reads the current checksum index and invokes its seal verification unconditionally before PASS. A missing index raises `FileNotFoundError`; a present index still requires the canonical anchor and exact indexed delivery bytes. The current README documents this requirement and identifies both prior snapshots as historical data.

The parent’s [classification](../closing-fix/classification.json) and [adjudication](../closing-fix/adjudication.json) were read as supplied context. Their GREEN claims did not establish this result. All dynamic checks below ran independently in a disposable extraction of `git archive d4a0a57a61ee76ab543d87169048111223c66c15`; uncommitted parent bookkeeping was excluded. No historical capture writer, model study or Go command was executed.

## Fresh verification

[post-fix-checks.json](evidence/post-fix-checks.json) preserves all11 direct command arrays, cwd/environment, exit codes, stdout and stderr, plus exact archive/source/snapshot identities. [The runner](evidence/post-fix-checks-runner.txt) preserves the diagnostic sequence; mutations affected only the temporary checkout, which was restored and then removed.

| Independent check | Actual result |
| --- | --- |
| Normal current gate on exact relocated d4 bytes | Exit0/PASS;31 reconstructions,26 completed/5 failed,130 source identities and declared seals verified. |
| Current `tests/test_go_study_integrity.py` through real unittest discovery | Exit0; all3 tests pass, including normal execution, the new missing-delivery-index CLI regression and all optimized-mode subcases. |
| Direct `-O`, `-OO`, and inherited `PYTHONOPTIMIZE=1` executions | Each exit1; explicit optimization refusal; no PASS. |
| Delete only the current delivery checksum index | Exit1; `FileNotFoundError` identifies current `checksums.json`; no PASS. |
| Keep the index absent and change the delivery README | Exit1 for missing index; no PASS. This is the original NG1 false-PASS trigger. |
| Restore the original current index while retaining that same README corruption | Exit1; `sealed bytes changed`; no PASS. |
| Restore exact delivery README/index bytes | Exit0/PASS. |
| Remove the current index digest from the canonical record while retaining exact delivery bytes | Exit1; `canonical anchor missing`; no PASS. |
| Restore the canonical anchor and exact packet | Exit0/PASS. |

The CLI regression exercises the advertised command against a complete disposable checkout with only the current index removed. Its failure expectation matches the independently observed error above. The existing normal and optimized-mode cases continue to pass. No additional production or test correction is required for NG1 in this scope.

## Snapshot and anchor preservation

Each packet contains12 indexed payload artifacts plus its checksum index (13 files total). Independent static checks establish the exact index digests, all indexed payload bytes, and the presence of each external anchor in the exact d4 canonical record:

| Packet | Verified index SHA-256 |
| --- | --- |
| Original pre-optimization snapshot | `20387a7bb3d3ef460fe1231eb2c5a721a71075a84118a9e8086e1bd5a1913e93` |
| Revision2 pre-required-seal snapshot | `dd63309f526b323dea232650026ab34c91f30468d3c0bbcaa81d1793287a4909` |
| Current revision3 delivery | `f20d74230d19d3323c31f0697a31679786401b7cc612b53ba14a1e276790df2b` |

All13 revision2 snapshot files independently match their exact `c5ee3d0:<current-delivery-path>/<file>` bytes, including the prior index. The original snapshot is unchanged across the comparison. The current manifest declares both old packets as additional seals, and the independently executed current gate verifies them. Their old code remains inert historical evidence; neither copied helper was invoked or “repaired.”

## Original review and continued Go assessment

The [59-path comparison inventory](evidence/post-fix-inventory.numstat) contains no changed Go source. Its only Go/Python source paths are the changed current gate, changed CLI regression file and the added exact historical revision2 gate copy.

Independent `git ls-tree -r` comparison verifies byte-identical tracked inventories for2,664 protected entries covering the original example/module/test, historical reviewed helpers, all runtime build/review guidance, both candidates’ drafts/fixtures, context/concurrency discovery and study archives, the eligibility-skip archive, and the original delivery snapshot. The identical protected tree-inventory digest is `2406b73c798b741effb60a6972b7195b7d06e266b1936088c5091afd58c7ec68`. This is an identity check, not a new source-quality or efficacy audit.

All13 original report/manifest/findings/ledger/card artifacts retain their pre-addendum hashes. The original report SHA-256 remains `b3f5de47bf5d900783cf33c53320a1326dc8bf2c6394add5259ad22e16e6e3f2`; its finding/card hashes independently agree with the original validation record. [Preservation evidence](evidence/post-fix-original-review-preservation.json) records before/after hashes. The original report and findings deliberately retain the pre-fix NG1 observation; this separate addendum records its closure.

The original Go-only overall **A** and zero critical/major/moderate/minor counts carry forward without rerunning Go checks: Architecture, Idioms, Correctness, Testing, Resources and Reproducibility remain **A**; Security, Resilience and Deployment remain justified **Not applicable**. The source and contract/build inputs underlying those cards are unchanged. Their actual Go1.22.12/1.26.5 evidence and limits remain the original evidence, not new executions on d4.

## Limits and handoff

This follow-up assesses only the named correction and preservation boundaries. Full-branch skill efficacy, archived experimental implementation grading, outcome-card re-adjudication, other targets/schedules, unavailable native harnesses/model settings and future edits remain outside scope. No new full Go review, broad structural suite, external GitHub/CI/Copilot lookup, publication or merge action occurred. Copilot round3 completion was supplied by the parent and is not independent passing evidence in this addendum.

The verification-before-completion principle was applied to this focused closure; the original Go topic skills/cards were carried by verified identity rather than reinvoked for another audit. Parent owns canonical status/work-log updates and final tracking/commit/push. No reviewed source or canonical file was edited by this worker.
