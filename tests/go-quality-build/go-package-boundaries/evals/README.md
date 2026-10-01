# Package-boundary trials

Each case is a separate Go 1.22 module using only the standard library. The large-project fixture is an affected slice, not a synthetic source-line count benchmark. Starting tests pass; requested additions are intentionally absent.

For each baseline, give a fresh model only the case prompt and a disposable copy of its listed files. Withhold `expected_output`, `assertions`, this README, the audit, and proposed runtime guidance. Save the exact prompt, model/harness, source revision, before/after diff, test command/output, and independent Architecture and Correctness reports under `tests/go-quality-build/results/<run-id>/<case-id>/`. Do not give judgments to the implementer before saving the initial result. Skill selection can be recorded separately; no build skill exists yet.

Run `rtk go test ./...` in each copied module. Starting-test success proves fixture integrity only; it is not a baseline model trial, an independent review, or evidence of skill uplift. HTTP test servers require a runner permitted to bind loopback ports. Review the code against responsibilities and behavior, accepting alternative package names and justified layouts.
