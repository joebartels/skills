# Interfaces and composition baseline — 2026-10-01

Stage: five blind model implementations and three independent blind Architecture/Testing reviews are archived. This baseline does not establish skill benefit.

## Inputs and separation

The [candidate suite](../../go-interfaces-and-composition/evals/evals.json) contains five cases and 20 listed fixture files. Each fixture is an independent Go module declaring Go 1.22, with a project README, implementation and starting tests. The six existing tests verify prior behavior, not the requested changes. The arithmetic case intentionally lacks partial-batch behavior; other cases require new features or composition changes.

| ID | Work requested | Decision exercised |
| --- | --- | --- |
| protocol-vs-mock | Add CSV and host encoder example; assess teammate's mock facade suggestion | Real package-owned protocol and fake-only interface proposal in one cohesive task |
| visible-dependency | Support concurrent tenant senders with distinct clients and URLs | Hidden environment/default-client dependency versus host-owned configuration |
| zero-value | Add optional distinct-key capacity and Clear | Useful zero value alongside optional configuration |
| library-lifecycle | Add periodic refresh and two-instance shutdown example | Host-controlled start/cancel/wait and resource ownership |
| not-abstraction-work | Fix private ceiling division, including largest int | Non-selection control without dependency or lifecycle work |

For each blind author, copy only that case's listed files into a disposable work directory and supply the exact `prompt` value. Do not expose evals.json, expected_output, assertions, this results directory, source audit, runtime build skills or prior model outputs. Ordinary project requirements live in fixture README files; no API solution, directory layout, constructor pattern, grading outcome or required test-double shape is prescribed there. The protocol case intentionally presents a teammate proposal for judgment rather than requiring its acceptance. The lifecycle contract states host needs without dictating Start/Run/Close method shapes.

Save the dispatch wrapper, model/harness identity, exact prompt, patch, raw author report, verification and output hashes per case. Use the same model for baseline and skill-on comparisons. Independent Architecture/Testing reviewers receive the contract and evidence after implementation. Requirements still need code review and targeted probes against the eventual API; the initial green tests alone do not establish fulfillment.

## Verification

[fixture-checks.json](fixture-checks.json) preserves command arrays, working directories, stdout, stderr and exit codes. All five modules pass `rtk go test ./...`, uncached verbose Go tests and `rtk proxy go vet ./...`: 15 successful test/vet commands, six starting tests. All Go files were formatted. Build-eval unittest discovery passes 9 tests. Repository validation reports 10 review skills/120 review cases and 14 build cases, including this candidate suite without a runtime skill. Whitespace checking passes.

[fixture-manifest.json](fixture-manifest.json) records the preparation base commit and SHA-256 of evals.json and every input file. Preparation used Go 1.26.5 on darwin/arm64. Go 1.22 execution, other platforms, race behavior of future implementations, automatic selection, behavioral baselines and skill-on outcomes are unverified. Starting fixtures have no third-party modules or live network requirements. The existing receipt test deliberately reflects the old global configuration design; model authors must replace that usage for concurrent tenant behavior.

All fixture code and prose are newly authored. Upstream source decisions remain pinned at `19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`; no upstream code/prose was copied. The approved context-sensitive protocol/construction rules take precedence over conflicting universal rules in local Go writing guidance.

The preparation stage above preceded the baseline runs described below.

## Archived blind baselines

[manifest.json](manifest.json) records five fresh `gpt-6-luna` collaboration authors (`fork_turns=none`, no reasoning override) and three fresh `gpt-6-astra` blind reviewers. Each case directory contains the exact eval task prompt, unedited author report, a source-only patch and fresh verification. Full dispatch-wrapper bytes were not saved; known isolation instructions and model settings are controller-attested and explicitly distinguished from exact task prompt bytes. The control author byte-compared its starting copy to the fixture before editing.

All five patches reconstruct the complete author source tree byte-for-byte, including project docs and examples. Those source hashes also equal the blinded reviewer candidates. Fresh ordinary tests, uncached verbose tests with a 30-second timeout, and vet passed for all five reconstructed modules (15 commands). Verification uses a writable temporary GOCACHE. Per-case verification.json files preserve command arrays, outputs and exit codes. Source and artifact hashes appear in the manifest; the three review files are byte-identical copies of the supplied reports.

| Case | Architecture | Testing | Confirmed review result |
| --- | --- | --- | --- |
| protocol-vs-mock | A | B | Protocol retained without fake-only facade; tests omit multiple-record order (A-T1, moderate) |
| visible-dependency | A | B | Independent explicit clients; unbounded test waits can hang on early request failure (B-T1, moderate) |
| zero-value | A | B | Zero-value/configuration behavior sound; existing-key test runs before capacity (C-T1, moderate) |
| library-lifecycle | B | C | README returns on first Wait error before joining second run (D-A1, moderate); no successful recurrence test (D-T1, major), weak independence assertion (D-T2, moderate) |
| not-abstraction-work | Not applicable | A | Private arithmetic repair with meaningful normal and overflow regression tests |

Unedited reviews: [A/B](review-ab.md), [C/D](review-cd.md), [control](review-control.md). Their disposable probes establish implementation behavior and demonstrate surviving test mutations; those reviewer probes do not improve the baseline authors' regression suites. Fresh archive verification does not rerun every reviewer probe. The reviews also disclose initial sandbox cache failures followed by successful writable-cache runs; those setup failures are not candidate defects.

The lifecycle finding concerns documented host composition: cancellation must be followed by joining every owned run even if one has failed. The library's primitives themselves allow correct shutdown. This is a bounded candidate target for composition guidance, along with verification that actually executes successful recurrence and completion. Other test findings remain recorded independently; they do not justify turning this skill into general testing guidance.

No automatic routing, minimum Go 1.22 execution, other-platform behavior, broad correctness audit or skill uplift is established. No source fix, runtime skill or commit was made in this archive stage. Next: controller uses the observed ownership failure to scope runtime authoring and subsequent same-model trials. Canonical design-record updates are deferred to the controller at its explicit request to leave docs unchanged during packaging.
