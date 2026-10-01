# Independent combined-case review

Reviewed 2026-10-01. Original: `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/combined/evals/files/combined-evolution`. Candidate: `/private/tmp/go-quality-build-combined-eval/skill-on/combined-evolution`. Locations below are relative to the candidate. Read the exact task, project README, both implementations and tests, and the two topic skills/references. Did not read expected assertions, trial reports, build skills, design records, or other reviews. All executable checks and mutations used disposable copies under `/private/tmp/fielddesk-review-correctness`; candidate and repository were not edited.

## Correctness & Compatibility — B
Scope: Supplied original/candidate changeset; Fielddesk library, HTTP service, new CSV CLI, feed sync and server lifecycle. Module declares Go 1.22; checks used Go 1.26.5 darwin/arm64.
Coverage: Existing exact Go function/method types, duplicate check-ins, HTTP methods and shapes, ordered CLI rejection and malformed input, feed status/JSON/field validation and preservation, caller client/context propagation, worker serialization/cancellation, host startup and shutdown code. Network listener integration was blocked by sandbox permissions; no other platform run.
Rationale: One contained moderate regression in server process failure signaling selects B. No major or critical issue was substantiated. Feed publication validates before touching the destination and uses a temporary file in the same directory followed by rename; shutdown code does contain a worker join. Those strengths do not offset the verified failure status regression.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [CG1] Existing public signatures, check-in implementation and HTTP handlers remain unchanged (`fielddesk.go:25`, `:29`, `:55`, `:172`). Existing tests and reviewer probes pass, including exact method-value assignments, duplicate replacement and 405 responses. The new CLI calls the same `SubmitCheckin` operation.
- [CG2] Feed failures return before publication, including non-success status, malformed JSON, trailing values, JSON null, and blank fields (`fielddesk.go:94–137`). Reviewer probes confirmed old snapshot bytes survive null, object, blank-text and trailing-input failures, and successful text reaches the HTTP response. Same-directory temporary publication avoids exposing partially written snapshots to service readers by inspection.
- [CG3] `http.NewRequestWithContext` uses the supplied client (`fielddesk.go:85–89`); synchronous cycles and a context-aware interval wait (`:150–168`) prevent overlapping cycles within one worker invocation. The supplied cancellation tests pass under race detection and establish that cancellation reaches a blocked transport before worker return.
- [CG4] Actual CLI subprocesses return 1 for malformed and rejected row 2, preserve row 1, and do not process row 3; duplicate successful rows return 0 and leave the final value. This matches the ordered ingress contract.

Bad

- [C1][moderate][introduced] `cmd/server/main.go:37–46` logs a non-`ErrServerClosed` serving error, completes cleanup and returns from main successfully. The original used `log.Fatal`. A built candidate binary under a denied bind logged `http server: listen tcp :8080: bind: operation not permitted` and exited 0. The same branch handles address-in-use failures. A launcher or operator using process status receives success despite no service having started. Primary owner: server host implementation.

Suggested changes

- [C1] Preserve the serving error through cleanup and return a nonzero process status after joining the worker. Verify failed startup exits nonzero and ordinary signal shutdown still completes cleanup.

Limits: Baseline `rtk proxy env GOCACHE=/private/tmp/fielddesk-review-gocache go test -race ./...` passed all existing tests; the same command with an added reviewer contract probe passed. Actual binary CLI cases described above passed. The server bind-denial reproduction verified the failure branch without a working listener. Initial commands using the default Go cache failed with sandbox cache-write errors; reruns used the writable temporary cache. No actual supported Go 1.22 toolchain run, disk-full injection, Windows filesystem test, or live socket shutdown integration was performed. A clean race run covers only executed interleavings. Late cancellation after response validation and unusual custom transports were not assigned stronger guarantees than the task requires.

## Testing — C-
Scope: New and retained tests in the same supplied changeset, assessed for meaningful regression detection of the requested evolution.
Coverage: Read all tests and implementation; ran baseline race tests, additional reviewer probes, binary ingress checks, and four independent mutations against copies containing only the candidate tests. Existing compatibility tests were treated as retained coverage; pre-existing gaps in unchanged behavior were not charged to this change.
Rationale: One major gap leaves the explicitly required host shutdown contract effectively unverified; three independent moderate gaps weaken CLI process semantics, feed validation and snapshot publication coverage. One major plus at least two moderate findings selects C-. These are independent test corrections, not counts of hypothetical production defects. No critical or systemic severity is claimed.
Finding counts: critical=0, major=1, moderate=3, minor=0

Good

- [TG1] `cmd/import-checkins/main_test.go:10–26` verifies an accepted prefix, row-specific rejection, and absence of a later file through real filesystem effects. It meaningfully checks service reuse and stopping on rejected input.
- [TG2] `fielddesk_test.go:50–86` exercises success and several failure classes against a real persisted snapshot. The cancellation tests at `:89–147` block on the actual request context, use bounded completion waits, and check cancellation before completion. The race-enabled suite passes.
- [TG3] Existing external-package caller/HTTP tests remain present and pass. They exercise the public construction and operations plus exact successful/empty response encodings.

Bad

- [T1][major][introduced] The requested host shutdown behavior has no test at the host boundary (`cmd/server/main.go:23–46`; `go test` reports `cmd/server [no test files]`). `TestRunBulletinSyncCancellationJoinsRequest` exercises the library method, not the host's cancellation/wait/resource sequence. Removing `<-workerDone` from main leaves the entire race-enabled candidate suite green. This makes the important explicit requirement that the host wait for its worker before returning effectively unverified. Primary owner: host lifecycle tests.
- [T2][moderate][introduced] CLI tests invoke `run` only (`cmd/import-checkins/main_test.go:17`, `:35`), so none checks the explicitly required nonzero process exit. Changing `os.Exit(1)` to `os.Exit(0)` in `cmd/import-checkins/main.go` leaves the entire race-enabled suite green. Malformed CSV coverage also has no later valid row or prefix-side-effect assertion. Primary owner: CLI ingress tests.
- [T3][moderate][introduced] Feed field validation checks only blank ID (`fielddesk_test.go:69–75`). Removing `strings.TrimSpace(b.Text) == ""` from `fielddesk.go:113` leaves every test green, allowing a valid-ID/blank-text response to replace the last good snapshot undetected. This is a normal malformed-feed case explicitly called out by the task. Primary owner: feed validation tests.
- [T4][moderate][introduced] Snapshot preservation tests fail before publication and inspect only completed writes (`fielddesk_test.go:50–86`). Replacing the entire temp-write/close/rename block with `os.WriteFile` to `bulletins.json` leaves the full race-enabled suite green. That regression can expose empty/partial JSON to concurrent HTTP readers and destroy the last good file on a failed write. Current tests provide useful validation-failure coverage, but no signal for publication atomicity or publication failure preservation. Primary owner: snapshot persistence tests.

Suggested changes

- [T1] Exercise the host with a blocked feed operation and observable shutdown stages, or factor a narrowly testable host runner. Verify cancellation, worker completion, and subsequent release/return ordering; the no-join mutation must fail. Include failed startup status to guard [C1].
- [T2] Add a process-level CLI failure test checking exit status and exact persisted prefix. Include a malformed row followed by valid input; the exit-zero mutation must fail.
- [T3] Add valid-ID/blank-text, empty-text, and successful full-value assertions with preserved snapshot checks on rejection. The removed-text-validation mutation must fail.
- [T4] Add an appropriate publication-boundary check: concurrent readers must only observe complete old/new snapshots, and a failed publication must preserve the old snapshot. Use a controlled failure/observation seam if needed to make the test deterministic; direct truncating publication must be detected.

Limits: All four mutations independently passed `rtk proxy env GOCACHE=/private/tmp/fielddesk-review-gocache go test -race ./...` in `no-text-validation`, `no-worker-join`, `cli-success-exit`, and `no-atomic-write` disposable directories. They were not combined, and reviewer-added probes were not included in those mutation runs. The publication mutation demonstrates absent regression signal, not a publication defect in the current candidate. Host no-join mutation likewise establishes a test gap rather than claiming the current host omits its join. No live network integration could be run under the sandbox. No fuzzing, exhaustive mutation campaign or timing stress campaign was needed for these findings.

Optional ungraded follow-up: `fielddesk_test.go:161` closes `second` on every call after the first, so a sufficiently delayed test goroutine could allow a third cycle and panic. The no-overlap check also measures only immediate transport calls, not the full decode/publication cycle. Consider a blocked cycle and an explicit release handshake when strengthening lifecycle tests; no reproduced flake is counted here.
