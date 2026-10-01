# Independent Testing review of review-guided repair

Reviewed 2026-10-01. Original fixture: `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/combined/evals/files/combined-evolution`. Repaired candidate: `/private/tmp/go-quality-build-combined-eval/skill-on/combined-evolution`. All locations below are relative to the repaired candidate. Read the exact evaluation prompt, README, implementation, all tests, original/candidate diff, Testing skill and decisions, first-pass review, and repair report. Candidate and repository were not edited. All mutations and test execution used disposable copies under `/private/tmp/go-quality-build-combined-eval/independent-testing`.

## Testing — B
Scope: Original-to-repaired supplied changeset, focused on the four first-pass surviving mutations and the repaired tests' isolation, cleanup, and determinism. Library, HTTP service, CLI and server host; module declares Go 1.22; executed with Go 1.26.5 darwin/arm64.
Coverage: Read every supplied test and traced its production path; tested host worker joining, CLI failure exit and prefix side effects, blank feed text rejection and snapshot retention, concurrent snapshot observation, cancellation, compatibility, serial cycles and failure reporting. Ran all candidate tests with race detection, repetition and shuffled order; independently recreated each of the four mutations. Live network/signal integration, injected filesystem write failures, other operating systems and Go 1.22 execution were not assessed.
Rationale: Three prior gaps now have strong demonstrated regression signal. Publication coverage improves substantially but still lets the exact direct-write mutation survive ordinary runs, and it misses that mutation consistently with one scheduler processor. This is one moderate issue because the snapshot-publication check remains meaningfully unreliable; no major or critical gap was substantiated in the reviewed scope. One moderate issue selects B.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [G1] `cmd/server/main_test.go:12` blocks worker cleanup until an explicit release and checks return ordering. Removing `<-workerDone` from the runner fails all ten targeted repetitions with `server returned before worker joined`. The fake listener still exercises the actual production lifecycle runner. `sync.Once` and cleanup release the blocked worker on assertion failure; cancellation is deferred. The listener-failure test also asserts error propagation and join completion.
- [G2] `cmd/import-checkins/main_test.go:40` runs the real `main` entry point in a child test process, requires exit 1 and a row-2 diagnostic, checks exact accepted-prefix bytes, and verifies absence of the later valid row. Changing `os.Exit(1)` to `os.Exit(0)` fails all ten repetitions. Child-only environment settings and temporary directories isolate filesystem and process state.
- [G3] `fielddesk_test.go:96` and `:97` now cover a valid ID with empty and whitespace-only text while asserting the previously accepted snapshot remains. Removing text validation fails all ten repetitions. The ordinary success path plus exact final bytes in the publication test also verifies full text persistence.
- [G4] The unmodified suite passed `rtk go test -race -shuffle=on -count=10 ./...` (120 test executions across three packages). Existing cancellation tests use observable request-start and completion channels. The worker notification now closes only on call two (`fielddesk_test.go:272`), eliminating the earlier repeated-close issue by inspection. No race report or unmodified-suite flake was observed.

Bad

- [T4][moderate][introduced; partially repaired] `fielddesk_test.go:130` through `:183` does not synchronize observation with the publication window. `ready` proves one old-snapshot read; the EOF gate stops feed decoding before marshaling/writing. After releasing EOF, the writer can truncate and complete the whole write before the observer runs again. Replacing the temporary-file/rename block with a direct `os.WriteFile` therefore still passes intermittently: first run failed 8/10 and passed 2/10; a separate JSON-recorded run failed 17/20 and passed 3/20. With `GOMAXPROCS=1`, it passed all 20/20 repetitions in each of two runs. Thus a common test configuration gives a green result for precisely the publication regression this repair intends to catch. This is a regression-detection finding, not a claim that the repaired production implementation uses direct writes. Primary owner: snapshot publication tests.

Suggested changes

- [T4] Add a deterministic publication-boundary safeguard. For example, hold an open descriptor to the old snapshot across a successful replacement and assert it still contains the old bytes, which detects in-place truncation on supported Unix filesystems; alternatively add a narrow controlled publication/failure seam that can stop a write and inspect the destination. Retain the concurrent observer as stress coverage. Verify the direct-write mutation reliably fails with both default scheduling and `GOMAXPROCS=1`; include failed-publication retention if the chosen seam supports it.

Limits: The host test's 20 ms negative wait is bounded and was effective in ten mutation runs, but cannot prove every possible scheduling interleaving. The feed publication test has no injected disk-full/short-write coverage. The child CLI test invokes main from a test binary, so it does not verify packaging/build invocation. Successful paths explicitly join helper goroutines; the publication test's early error/timeout branches do not uniformly stop and join its observer, and the cancellation tests do not defer cancellation on every failure path. No persistent leak or baseline flake was reproduced, so those observations are optional cleanup improvements rather than additional counted findings. A race-clean run covers executed paths only. No claim of complete mutation coverage or cross-platform validation is made.

### Executed mutation checks

Each mutation was applied independently to a fresh copy, without reviewer-added tests. The three deterministic checks used:

`rtk proxy env GOCACHE=/private/tmp/go-quality-build-combined-eval/cache go test <package> -run '^<test>$' -count=10 -timeout=60s`

- `no-join`: `./cmd/server`, `TestRunServerSignalShutdownJoinsWorker` — 10 failures.
- `exit-zero`: `./cmd/import-checkins`, `TestImportCheckinsProcessFailureStatusAndPrefix` — 10 failures.
- `blank-text`: `.`, `TestSyncBulletinsValidatesBeforeReplacingSnapshot` — 10 failures.
- `direct-write`: `.`, `TestSyncBulletinsPublicationOnlyExposesCompleteSnapshots` — 8 failures and 2 passes initially; a second `go test -json` run with `-count=20` recorded 17 failures and 3 passes.
- `direct-write` with `GOMAXPROCS=1`: the same targeted test with `-count=20` passed all 20; a second JSON-recorded run also passed all 20.

Machine-readable second-run evidence: `/private/tmp/go-quality-build-combined-eval/independent-testing/direct-write-default.jsonl` and `/private/tmp/go-quality-build-combined-eval/independent-testing/direct-write-1.jsonl`. The JSON commands were run through a Python subprocess with the same explicit writable `GOCACHE`; the only scheduling change in the second was `GOMAXPROCS=1`.
