# Package-boundary baseline run: 2026-10-01

Status: five blind baseline implementations archived and verified; independent Architecture/Correctness review complete.

[manifest.json](manifest.json) records each case, exact task prompt, shared dispatch wrapper, model/harness, fixture revision and artifact paths. Each fresh `gpt-6-sol` Codex collaboration subagent received only its case prompt and disposable fixture, with no build skill or expected judgments. Source fixture revision: `4ab6b82106b3b04a035203956606f7be0b328229`.

Each case directory contains `prompt.txt`, unedited `trial-report.md`, a source-only `source.patch`, and `verification.txt`. The verification worker applied every patch to a fresh fixture copy and checked byte-for-byte equality against the trial's complete Go/module source. All five reconstructed modules pass `rtk go test ./...`, raw uncached verbose tests, and raw `go vet`; outputs and exit codes are preserved. Raw verbose output distinguishes tests from table subtests. The author's report remains separate from independently rerun command evidence.

`fixture-validation.txt` and `fixture-manifest.json` preserve starting-fixture tests and SHA-256 values. They are not implementation results. The actual runner is Go 1.26.5 darwin/arm64; all modules retain Go 1.22 language semantics, but minimum-toolchain execution was not performed.

The unedited [independent review](baseline-review.md) assesses all five original/changed source sets with the existing Architecture/Correctness skills, withholding eval expectations, author reports and proposed guidance.

| Case | Architecture | Correctness | Confirmed findings |
| --- | --- | --- | --- |
| small-cli | A | A | None |
| small-library | A | A | None |
| medium-service | A | A- | MS-F1: minor Scanner token-limit edge |
| large-features | A | A | None |
| not-package-work | Not applicable | A | None |

MS-F1: replay scans before trimming, so a line containing 65,536 spaces plus newline produces token-too-long and exit 1 instead of being skipped. Large padding around a valid ID has the same cause. The reviewer independently reproduced this with the replay binary path. The suggested correction is a reader without the default Scanner cap plus a long-blank-line regression. Baseline source is preserved, including the defect.

Review limits: only Architecture and Correctness were assessed; no race, platform or minimum-toolchain matrix was run by the reviewer. CLI exit and catalog/invoice compatibility were partly established by source inspection rather than new subprocess tests. Local-listener sandbox failures were rerun successfully and are not code findings. Reviewer model identity was not included in the supplied report.

All four applicable architectural outcomes were already strong. This baseline demonstrates no package-boundary skill uplift and supplies no evidence that proposed guidance is necessary. The controller will investigate additional pressure cases separately. No runtime build skill exists.

Patches use zero context so stored diff context does not trigger repository whitespace checks. Reconstruct in a disposable fixture copy with `rtk proxy git apply --unidiff-zero /absolute/path/to/source.patch`; final patch bytes were reapplied and compared to the tested source.
