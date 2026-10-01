# Package-boundary baseline run: 2026-10-01

Status: fixtures validated; **baseline model trials and independent reviews pending**.

Fixture authoring harness: Codex desktop, GPT-6 (exact configured model identifier not exposed to this worker). No baseline implementation model is claimed. `fixture-validation.txt` contains starting-fixture test/vet evidence, not model-task results. `fixture-manifest.json` pins the 19 original inputs by SHA-256. All five modules use Go 1.22 language semantics; the actual runner version is recorded in the log.

The controller will save blind model outputs for `small-cli`, `small-library`, `medium-service`, `large-features`, and `not-package-work` under this run ID. Each case needs exact prompt, model/harness, changed-source diff, `rtk go test ./...` evidence, and independent Architecture/Correctness findings. Copy fixtures into disposable workdirs and withhold expected judgments. Preserve failed attempts and identify unrun cases. No skill uplift can be claimed from this preparatory run.
