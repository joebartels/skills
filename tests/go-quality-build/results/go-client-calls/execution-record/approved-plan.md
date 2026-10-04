# Go Client Calls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for native execution, or superpowers:subagent-driven-development if the user selects task delegation. Implement task by task; independent behavioral authors and reviewers are required in either mode. Steps use checkbox syntax for tracking.

**Goal:** Deliver a demonstrably useful HTTP/unary gRPC authoring skill, revising it once for concrete failures or abandoning runtime promotion if its added value is unproven.

**Architecture:** One concise main skill owns outbound-operation policy, with conditional HTTP and gRPC references. Frozen paired tasks compare the existing build catalog with the same catalog plus this candidate. Guidance enters runtime only after independently confirmed improvement, preserved controls and a fresh transfer check.

**Tech Stack:** Python 3.10+/PyYAML validation, Go standard-library HTTP fixtures, grpc-go `v1.67.3`, protobuf `v1.34.2`, gobreaker/v2 `v2.4.0`; fixture modules declare `go 1.22.0`.

**Spec:** [Approved design](/private/tmp/go-client-calls-design-2026-10-03.md). This temporary plan stays outside maintained repository documentation.

## Global constraints

- HTTP and unary gRPC only; streaming recovery, hedging and application-wide orchestration are excluded.
- Preserve supported APIs, Go/library versions, borrowed configuration and existing decision owners. Add no universal timeout, constructor, interface, dependency or breaker requirement.
- Keep operation identity/payload stable; distinguish repeat safety from transient-failure classification and ambiguous completion.
- Account for existing retry layers and transparent behavior; identify what every limit counts.
- Four applicable cases, one non-selection case and one held transfer pair: initially twelve independent author contexts, plus reviewers. One evidence-driven guidance revision may rerun affected skill-on cases and controls. No further campaign expansion without agreement.
- Fixed neighboring catalog and equivalent settings; withheld author expectations/probes; neutral review labels. Disclose unavailable metadata and selection/blinding limits.
- Preserve original failures and exact evidence; assisted repairs are separate from blind outcomes. Do not modify review rubrics to improve results.
- Runtime content remains canonical under `plugins/go-quality-build/skills/`; evidence belongs under `tests/go-quality-build/`. No status logs, proposals or outcome narrative in maintained guides.
- Prefix shell commands with `rtk`. Use scratch-owned caches/toolchains; do not change global configuration, publish, push or merge.

## Review focus

1. A caller retries the entire operation after a lost reply: preserve the same logical identity and payload across invocations.
2. A retry hint is invalid or overflows: do not turn it into an immediate retry or exceed the total budget.
3. A gRPC handler sends response headers before returning failure: respect commitment and do not add application replay.
4. A canceled half-open probe or old in-flight completion arrives: neither establishes fresh healthy recovery nor corrupts a new breaker generation.
5. A simple caller needs only one bounded request: preserve a minimal supplied-client implementation without extra policy machinery.

These conditions are assigned to Task 1's private probes and assessed in every relevant arm.

## File map

- `tests/go-quality-build/go-client-calls/evals/evals.json`: six task definitions and exact input inventories.
- `tests/go-quality-build/go-client-calls/evals/files/{http-write-replay,http-retry-budget,grpc-policy,breaker-recovery,pure-control,grpc-transfer}/`: each module has `README.md`, `go.mod`, `client.go`, `client_test.go`; dependency modules also commit `go.sum`.
- `tests/go-quality-build/go-client-calls/probes/<case>_test.go.txt`: controller-only contract checks; never author inputs.
- `tests/go-quality-build/client_call_trials.py`: isolated trial preparation, immutable catalog copying, exact patch reconstruction and sealing.
- `tests/go-quality-build/test_client_call_trials.py`: meaningful isolation/reconstruction tests for that helper.
- `tests/go-quality-build/go-client-calls/draft/{SKILL.md,references/http-clients.md,references/grpc-clients.md}`: candidate outside runtime until disposition.
- `tests/go-quality-build/results/go-client-calls/{manifest.json,research.json,baseline/,skill-on/,transfer/,reviews/,comparison.json}`: frozen settings, source identities, raw evidence and outcome decisions. Add `revision-2/` only if warranted.
- On promotion only: `plugins/go-quality-build/skills/go-client-calls/`, both package manifests and the package guide's owner table. No marketplace path changes are needed.

## Task 1: Freeze trustworthy fixtures and the evaluation harness

**Interfaces:** Produces six exact task modules, withheld probes, catalog snapshots and a harness CLI. `prepare_trial(case: str, arm: str, trial_id: str, scratch: Path, archive: Path) -> Path`; `archive_trial(trial: Path, archive: Path) -> dict`. CLI actions `prepare` and `archive` take `--case`, `--arm baseline|skill-on`, `--trial-id`, `--scratch` and `--archive`. Paths/IDs must reject escapes, archives must not overwrite and reconstructed output hashes must match.

- [ ] Read authoring/skill-creator/writing-skills/Go style and relevant runtime guidance. Reuse the linked worktree; if still clean, create `codex/go-client-calls` from its current HEAD. Preserve the existing error-evaluation branch. Stage only this effort's files.
- [ ] Run repository/root/build/layout/grade baseline checks. Existing validator passes 120 review/49 build cases at preparation; build unit suite passes 10 tests. Treat these as structure, not behavioral results.
- [ ] Write harness tests `test_prepare_freezes_exact_catalog`, `test_rejects_path_escape`, `test_refuses_archive_overwrite`, `test_detects_catalog_mutation`, and `test_patch_reconstructs_added_and_deleted_files`. Run `rtk proxy python3 -B -m unittest discover -s tests/go-quality-build -p 'test_client_call_trials.py'`; observe missing-helper failures, then implement the declared interfaces and rerun to PASS.
- [ ] Create the fixtures with these frozen public surfaces and task contracts:

| Case | Surface and requested behavior |
| --- | --- |
| `http-write-replay` | `Submit(ctx context.Context, client *http.Client, endpoint, operationID string, payload []byte, deduplicates bool) ([]byte, error)`: at most 3 application attempts under 500ms; stable caller-supplied identity/body; repeat lost acknowledgements only under the specified deduplication contract. |
| `http-retry-budget` | `Fetch(ctx context.Context, client *http.Client, endpoint string) ([]byte, error)`: exact 200 success; only 429/503 eligible; 3 application attempts and one 250ms budget; honor valid retry hints without retrying early; 4096-byte result/diagnostic limit; preserve rejection and closure. |
| `grpc-policy` | `Dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)` and `Probe(ctx context.Context, conn *grpc.ClientConn, service string) error`: use generated health service; default per-method UNAVAILABLE policy permits 3 total configured attempts; a resolver policy of 2 takes precedence; no outer replay, borrowed connection preserved. |
| `breaker-recovery` | `New(client *http.Client, enabled bool) *Client` and `(*Client).Fetch(ctx context.Context, endpoint string) ([]byte, error)`: supplied client reused; dependency scope is scheme/authority; three consecutive eligible 503 failures open for 100ms; one recovery admission; caller cancellation and ordinary 400 rejection excluded from health counting; 200 proves recovery. Disabled mode adds neither breaker nor retry behavior. |
| `pure-control` | `SumPositive(values []int) int`: local arithmetic correction; no remote client or resilience mechanism. |
| `grpc-transfer` | `Commit(ctx context.Context, conn *grpc.ClientConn, operationID string, payload []byte) ([]byte, error)`: different generated unary service with caller identity in metadata and server deduplication; preserve identity/payload after a lost acknowledgement and an outer caller retry; use the effective native retry policy. Held until Task 4. |

Fixture-specific numeric values are evaluation contracts, not universal recommendations. Preserve existing consumer signatures. Use shipped generated gRPC health/testing messages rather than installing protoc.

- [ ] Add private tests `TestIdentitySurvivesOuterRetry`, `TestLostReplyWithoutDedupIsNotRepeated`, `TestRetryAfterOverflowDoesNotRetryImmediately`, `TestRetryHintCannotRestartBudget`, `TestCommittedRPCFailureIsNotReplayed`, `TestResolverPolicyPrecedesDefault`, `TestCanceledProbeDoesNotRecover`, `TestOldCompletionCannotChangeNewGeneration`, and `TestDisabledClientStaysMinimal`. Assert actual effects/attempts/bytes, not helper implementation. Hold asynchronous work with events and useful time bounds.
- [ ] Verify old starter tests pass; new-contract probes fail named assertions against incomplete starters. In disposable controller reference implementations, show all probes pass and meaningful mutations fail assertions. Keep that reference out of all author contexts and label it as fixture validation, not skill benefit.
- [ ] Pin/download dependencies into scratch. Execute modules with host Go 1.26.5 and Go 1.22.12; obtain the latter in scratch from the official Go distribution with published checksum verification if needed. Run `rtk proxy env GOWORK=off GOTOOLCHAIN=local go test -count=1 ./...` and `go vet ./...` using each selected toolchain. Record unavailable execution explicitly; do not claim a version from its module declaration.
- [ ] Freeze manifests, prompts, allowed catalog and probes before dispatch; keep transfer inputs sealed. Commit only completed fixture/harness artifacts.

## Task 2: Obtain blind baselines

**Interfaces:** Consumes Task 1's five initial cases and fixed eight-skill catalog. Produces five reconstructed baseline candidates and neutral independently checked findings.

- [ ] Dispatch one fresh author per initial case with exact prompt/source/catalog. Neither candidate guidance nor probes, outcomes, review criteria or other trials are available as allowed inputs. Use equivalent inherited author settings for all arms and disclose inaccessible exact IDs. Preserve launch text and author selection/verification reports.
- [ ] Reconstruct every patch exactly; compare input/output/catalog hashes. Run candidate tests/vet, relevant race and focused checks, then held probes in disposable copies. Failures here are assessed outcomes, not occasions to repair the baseline.
- [ ] Have fresh outcome reviewers assess original requirements and baseline code, independently confirm consequential failures, and preserve strengths and unnecessary machinery. Grade only substantiated causes using relevant existing review skills; give no target grades.
- [ ] Record a factual baseline ledger. Strong baselines remain strong controls. Commit immutable baseline evidence before candidate authoring.

## Task 3: Author the candidate and measure its added value

**Interfaces:** Consumes approved design, frozen baseline evidence and primary-source research. Produces exact candidate guidance, five matched skill-on outcomes and a first comparison.

- [ ] Write the main and two references as original conditional guidance. Aim for a main of roughly 600–800 words; keep protocols in their references. Cover the approved procedures and confirmed risks without turning probes into an answer key. No mandatory client wrapper, breaker library or framework.
- [ ] Validate frontmatter/self-contained links and extracted runnable examples on their declared toolchains. Verify Go-specific retry/configuration semantics against the selected grpc-go source; binary success/failure classifiers must not be described as exclusion support. Save exact source URLs/versions in `research.json`; substantial copying requires attribution/license notices.
- [ ] Freeze candidate and catalog hashes. Dispatch five fresh matched authors; only candidate exposure changes. Preserve all outputs, exact reconstruction and checks as in Task 2.
- [ ] Use fresh arm-blinded reviewers on neutral baseline/skill-on pairs. Inspect same private probes, actual contracts, complexity and relevant mutation sensitivity. Freeze findings before revealing the arm map. Keep author-reported selection distinct from automatic routing.
- [ ] Compare unique policy/correctness improvements, regressions, unnecessary surface/dependencies and preserved strong controls. All-A ties or more comments/tests do not establish effectiveness.
- [ ] If a concrete guidance defect or demonstrated missed owner decision is fixable, revise once and rerun every affected skill-on case plus the selection/minimality controls with fresh authors and identical original inputs. Keep prior failed outcomes. Do not edit neighboring skills or coach authors with expected solutions. If no worthwhile correction exists, proceed to abandonment disposition rather than inventing one.
- [ ] Commit frozen candidate/evidence and the supported comparison; no runtime promotion yet.

## Task 4: Check fresh transfer and decide retain/revise/abandon

**Interfaces:** Consumes final candidate revision and sealed `grpc-transfer` task. Produces its fresh pair and an explicit disposition in `comparison.json`.

- [ ] Dispatch two fresh transfer authors with matched settings/catalog; baseline omits this candidate, guided arm receives exact final bytes. Do not adapt the task after seeing earlier outcomes.
- [ ] Reconstruct and execute both; have an independent reviewer assess neutral code/outcomes before arm disclosure. Verify operation identity, effective policy, uncertain results and preserved ownership through the actual unary boundary.
- [ ] Promote only with at least one independently confirmed meaningful marginal improvement, preserved important controls, successful final guidance transfer and no unresolved material introduced regression. State the bounded evidence; do not promise universal effectiveness.
- [ ] If added value remains unproven or the permitted revision leaves material defects, abandon runtime promotion. Preserve evaluated snapshots/evidence, remove the active draft, leave runtime/package metadata unchanged and explain the outcome to the user. Reopening the campaign requires new authorization rather than further result searching.

## Task 5: Deliver the supported disposition

**Interfaces:** Consumes Task 4's decision. Produces either an exact evaluated runtime skill or an archived abandoned experiment, with verified packaging and final independent branch review.

- [ ] For promotion, copy exact accepted bytes to runtime, remove the active draft, add the concise owner row and update both manifests to 0.5.0 from current 0.4.0. If concurrent integration changes the base version, choose the next matching minor version; do not overwrite unrelated metadata.
- [ ] For abandonment, retain only immutable evaluation evidence and reusable valid fixtures/harness checks; package guide/manifests acquire no new skill claim. Remove proposed active guidance from the working tree after its snapshot is archived.
- [ ] Run `rtk proxy python3 -B scripts/validate.py`, root/build/layout/grade suites, both Claude package validators and marketplace validation. Check local links, exact guidance snapshot equality on promotion, hashes/reconstructions and maintained-file whitespace. Exclude only explicitly inventoried immutable raw evidence where needed.
- [ ] Obtain one fresh strongest-available whole-branch reviewer; assess source/probe fidelity, claimed marginal benefit, client lifetime, breaker exclusion/recovery, native retry behavior and preservation of unrelated work. Fix substantiated findings and reverify affected evidence; guidance changes reopen applicable behavioral checks. If that would exceed the one-revision campaign limit, abandon promotion or obtain new campaign authorization instead of silently expanding it.
- [ ] Commit completed deliverables locally. Report retained or abandoned, the concrete evidence, executed toolchains and material limits. Do not add progress/history/outcome prose to maintained docs. No push/PR/merge is included.

## Execution choice

Native execution is recommended: one controller implements the sequential fixture/guidance stages while fresh independent authors/reviewers provide the behavioral comparison. Task delegation additionally introduces a new implementer and reviewer per stage; it does not replace independent evaluation authors. Review this concrete plan and choose the execution method before implementation.

## Primary dependency references

- [grpc-go v1.67.3 module](https://raw.githubusercontent.com/grpc/grpc-go/v1.67.3/go.mod): declared Go 1.21; the fixture preserves Go 1.22 compatibility.
- [gobreaker/v2 v2.4.0 module](https://raw.githubusercontent.com/sony/gobreaker/v2.4.0/v2/go.mod) and [implementation](https://raw.githubusercontent.com/sony/gobreaker/v2.4.0/v2/gobreaker.go): Go 1.22 and explicit excluded-outcome support.
- Runtime guidance remains version-sensitive and library-neutral; these pins define reproducible fixtures, not required dependencies for consumers.
