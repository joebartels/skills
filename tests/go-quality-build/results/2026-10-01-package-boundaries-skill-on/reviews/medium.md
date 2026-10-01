# Blind review: service replay candidates

Reviewed each supplied directory as a changeset against the original fixture named in review-input.md. All modules declare Go 1.22. Paths below are relative to the named candidate directory. Read all supplied Go source, tests and go.mod files, plus the Architecture & Design and Correctness & Compatibility skills and their decision references. No candidate provenance or excluded evaluation material was read.

# Candidate A

## Architecture & Design — B
Scope: Candidate A changes versus the original fixture; HTTP service and replay CLI.
Coverage: Package responsibilities, consumer interfaces, composition, validation/storage/notification sequencing, cancellation propagation, compatibility facade and both command entry points.
Rationale: One moderate issue: infrastructure implementation now exists both in adapters and in the compatibility branch of the orchestration method. This creates concrete maintenance duplication precisely across the two boundaries the task says will change independently next quarter. Other package splits are proportionate.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [A-G1] `dispatch.go:12-20` defines two single-method interfaces used by the shared operation. `cmd/server/main.go:15` and `cmd/replay/main.go:36` compose the same concrete adapters, and both ingress paths call `Service.Submit`. This centralizes validation and ordering without a generic framework.
- [A-G2] `internal/recordstore/store.go:12` owns stored bytes and `internal/notifyclient/client.go:16` owns the outbound protocol for the new command wiring. Separating these independently maintained integrations is warranted.

Bad

- [A-F1][moderate][introduced] `dispatch.go:37-57` retains full file persistence and HTTP notification implementations alongside the new adapters (`internal/recordstore/store.go:12-13`, `internal/notifyclient/client.go:16-29`). Existing field-configured callers execute one implementation while both binaries execute another. A future record-format or notification-protocol change therefore requires coordinated edits in the supposedly reusable orchestration package as well as the owning adapter package, with a concrete risk of divergent behavior between configuration paths. Primary remediation owner: Architecture & Design. This is a maintenance finding, not a claim that the two paths currently produce different results.

Suggested changes

- [A-F1] Retain legacy configuration fields if desired, but implement their fallback by constructing and delegating to the same adapter types; keep the storage bytes and HTTP protocol in one implementation each. Verify legacy field configuration and explicitly injected adapters have equivalent effects and errors.

Limits: Inspected the complete supplied candidate. `rtk go test ./...` and `rtk go test -count=1 ./...` both succeeded (RTK reports 5 passed in 5 packages). There is no demonstrated runtime protocol divergence today. No speculative future feature or mandatory architecture template was imposed.

## Correctness & Compatibility — A-
Scope: Candidate A behavior versus the fixture and explicit replay requirements.
Coverage: Trim/blank handling, sequential execution, stopping on invalid IDs or notification errors, scanner errors, argument/file failures, nonzero process exit, HTTP statuses/body, stored bytes, existing field configuration, adapter parity by inspection.
Rationale: One minor boundary-input defect was reproduced. Normal input processing and preserved HTTP/storage behavior are supported by source and tests. The defect concerns an unusually long blank line and has recoverable, contained impact.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [A-G3] `cmd/replay/main.go:52-64` processes IDs synchronously and returns the first operation or scanner error. `cmd/replay/main_test.go:23-74` verifies normal replay ordering, invalid-ID stopping, retained record on notification failure, and absence of subsequent records; fresh test execution passed.
- [A-G4] `dispatch.go:63-72` preserves the original 405 empty body, 400 `submission failed\n`, and 204 empty body behavior. Both persistence implementations retain `queued\n` with mode 0600. The normal success HTTP test passed; error response preservation was additionally inspected.

Bad

- [A-F2][minor][introduced] `cmd/replay/main.go:50` uses the Scanner default token limit before `TrimSpace` at line 54. A file containing 70,000 spaces and a newline should be skipped as a blank trimmed line, but the real replay command exits nonzero with `bufio.Scanner: token too long`. Likewise a short valid ID with sufficiently long surrounding whitespace is rejected before submission. Primary remediation owner: Correctness & Compatibility.

Suggested changes

- [A-F2] Read logical lines without the Scanner default ceiling (or implement chunked line reading) so blank/trimmed-line semantics are preserved. Add a case with a long blank line followed by an ordinary ID. Merely increasing the ceiling retains an undocumented limit unless an input bound is agreed.

Limits: The command reproduction used `go run ./cmd/replay FILE` with a temporary file containing 70,000 spaces plus newline and observed exit code 1. Supplied tests exercise replay through `process` with a fake notifier; they do not directly test the environment/argument wiring or real notification adapter. Their wiring and adapter code were inspected. An additional Python HTTP integration harness was blocked at localhost socket bind by the sandbox; no result is claimed for that harness. Existing concurrency constraints and filesystem behavior were not redefined by this task. No race or cross-platform suite was run.

# Candidate B

## Architecture & Design — A
Scope: Candidate B changes versus the original fixture; HTTP service and replay CLI.
Coverage: Shared operation, dependency direction, adapter ownership, facade composition, small interfaces, cancellation and synchronous lifetime.
Rationale: No actionable architectural issue found. The application operation depends on two narrow interfaces and both ingress paths reach it through a compatible facade. Three internal packages have concrete responsibilities justified by the second consumer and independent integration evolution. These are sound ordinary boundaries; no A+ claim is made.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-G1] `internal/submit/submit.go:9-27` centralizes validation, persistence-before-notification ordering and error propagation behind consumer-owned interfaces. The operation imports neither storage nor HTTP implementation details.
- [B-G2] `internal/recordstore/file.go:11` and `internal/notification/http.go:16` each own one external representation/protocol. `dispatch.go:20-24` preserves existing field-based composition while delegating to those single implementations. Both commands call the same facade operation, so neither duplicates policy.

Bad

- None found.

Suggested changes

- None needed.

Limits: Inspected all supplied source and tests. `rtk go test ./...` and `rtk go test -count=1 ./...` succeeded (RTK reports 6 passed in 6 packages). The extra internal operation package has a present ownership purpose and does not establish unnecessary generic infrastructure. No future implementation was presumed.

## Correctness & Compatibility — A-
Scope: Candidate B behavior versus original fixture and explicit replay requirements.
Coverage: Replay file/env wiring, trimming, blank skipping, ordering, invalid ID and notification failure stopping, error/exit propagation, HTTP response compatibility and exact stored representation.
Rationale: One minor boundary-input defect was reproduced, matching Candidate A. Otherwise the preserved facade and end-to-end replay tests provide direct evidence for the required normal/error behavior.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [B-G3] `cmd/replay/main_test.go:37-113` exercises the actual `run` function with temporary batch files, injected environment values and an HTTP notification server. It checks request order, exact record bytes, invalid ID stopping, notification error propagation and retained first record with no later record. Fresh execution passed.
- [B-G4] `dispatch_test.go:45-65` verifies 405/400/204 status and body compatibility; original success and notification-failure tests remain intact. `dispatch.go:20-35` and the adapters preserve ordering and observable format.

Bad

- [B-F1][minor][introduced] `cmd/replay/main.go:36` uses the Scanner default token limit before trimming at line 40. A 70,000-space blank line makes the real replay command fail with `bufio.Scanner: token too long`, although the contract says blank trimmed lines are skipped. Short valid IDs with sufficiently large surrounding whitespace are similarly rejected. Primary remediation owner: Correctness & Compatibility.

Suggested changes

- [B-F1] Use line reading that accommodates the specified trim/skip semantics without the default token ceiling; verify a long blank line followed by a valid ID and a heavily padded valid ID. Define an explicit input bound only if the contract is intentionally changed.

Limits: Long-blank-line reproduction via `go run ./cmd/replay FILE` returned exit code 1. Actual OS environment lookup and log.Fatal exit wiring were inspected; success and error run paths were covered by the supplied tests. The attempted extra Python HTTP harness was blocked by sandbox socket permissions before any integration operation. No cross-platform or race suite was run. Unchanged filesystem semantics and preexisting serialized-caller constraints are not introduced findings.

# Comparison

Candidate A: Architecture B (one moderate duplication issue); Correctness A- (one minor long-line defect).
Candidate B: Architecture A (no confirmed issue); Correctness A- (one minor long-line defect).

Both deliver a proportional shared-operation design and pass their supplied fresh tests. Candidate B gives the independently maintained integrations single implementation owners, while Candidate A leaves duplicate compatibility implementations. Neither scanner handles all blank trimmed lines specified by the task.
