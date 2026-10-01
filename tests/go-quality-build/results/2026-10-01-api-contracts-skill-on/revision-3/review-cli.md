# Candidate C revision 3 reassessment

Target: `/private/tmp/go-quality-build-api-eval/skill-on-luna-revision-3/wire-contract`, compared to `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-api-contracts/evals/files/wire-contract`. The supplied directory diff is the changeset; no commit identifiers were supplied. This is a follow-up aware of the earlier leading-dash filename defect. The original report remains unchanged.

Review-integrity limitation: a recursive file read inadvertently included the candidate's `trial-report.md`, contrary to the requested exclusion. This reassessment therefore cannot claim blindness to the author report. No build skill or eval definitions were read. Conclusions below use inspected implementation and independently executed probes, not author claims.

## Architecture & Design — A
Scope: Revised private Go binary, public joblist 1.x CLI and JSON output, module declares Go 1.22.
Coverage: Private/public data naming, CLI argument interpretation, filtering ownership, error flow, and the new two-function parser. No new concurrency, external dependencies or lifecycle decisions.
Rationale: No substantiated architectural defect. The explicit parser has a concrete role: extending the CLI grammar while retaining its legacy one-file interpretation. Its size and locality fit this command; no redesign is needed.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [R3-G1] `main.go:10–15` keeps private `Key` independent of the documented wire field through `json:"id"`. Independent executions preserve `id`, `state`, empty string fields and the response envelope.
- [R3-G2] `main.go:48–84` owns argument interpretation in private helpers; `main.go:50–51` explicitly preserves the original one-argument grammar. The returned presence boolean permits an empty filter value without conflating it with omission. Independent probes verify these behaviors.

Bad

- None found.

Suggested changes

- None needed.

Limits: Review covers the supplied command fixture, not external dashboard or script source. No general API redesign or style audit was undertaken. Author-report exposure is disclosed above.

## Correctness & Compatibility — A
Scope: Same revised directory versus original fixture. Reassesses exact optional filtering, unchanged 1.x wire/exit contract and private rename, including prior finding C-F1.
Coverage: Omitted, exact, case-sensitive and explicit-empty filters; original order; unknown input fields; empty results, empty arrays and null; malformed, trailing and wrong-type JSON; missing files and usage errors; real process streams/status; leading-dash filenames; flag placement, equals syntax, delimiter, duplicate flags and missing operands.
Rationale: No actionable issue found. Prior C-F1 is resolved by verified legacy handling. Core behavior and material argument-grammar paths pass. Correct handling is sufficient for A; no A+ claim is made.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [R3-G3] `main.go:50–51` fixes C-F1: with an actual `-jobs.json` file containing `[]`, both original and revised binaries invoked with the single argument `-jobs.json` exit 0, emit exactly `{"jobs":null}\n` and emit no stderr. Additional actual files named `--state`, `--state=ready` and `--` also work as sole arguments in both versions. This preserves existing filename grammar even when a filename resembles new syntax.
- [R3-G4] `main.go:34–40,60–72` distinguishes flag presence from value. With records in the order empty-state, ready, Ready, ready, omission returns all four, exact `ready` returns the second and fourth in order, and an explicit empty string returns the first. No-match produces `{"jobs":null}\n`.
- [R3-G5] `main.go:24–43` preserves errors and representation: empty array and null produce null jobs; empty id/state fields remain present; unknown input fields are ignored. Malformed JSON, multiple documents, numeric id, missing file, no arguments and excess file arguments yield exit 2, no stdout and nonempty stderr. Successful probes yield exit 0 with no stderr.
- [R3-G6] `main.go:57–84` handles expanded grammar consistently in probes: options before and after FILE, `--state=ready`, `--state=`, and `--state ready -- FILE` succeed; `-- FILE` disables option interpretation; `--state -- FILE` treats `--` as the state value. Duplicate state options, missing FILE, missing state after FILE, unknown option plus FILE, and extra positionals after `--` yield usage errors without stdout.

Bad

- None found.

Suggested changes

- None needed.

Limits: `go test ./...` passed in disposable copies of revised and original code. Both executables built with `go build -o <temporary-binary> .`; the independent Python harness invoked actual binaries and checked outputs and statuses. Runtime was Go 1.26.5 darwin/arm64, using a disposable GOCACHE. Go 1.22 itself, other OS targets, external consumers and writer failures were not executed. Checks preserve candidate source/configuration. Author-report exposure is disclosed above.

## Argument-grammar assessment notes

A sole `--state` is intentionally interpreted as a filename for compatibility; it is a read error when absent and succeeds when such a file exists. This is the same behavior as the original binary and is not a newly introduced missing-value defect. The candidate's `TestInvalidStateFlagUsage` case for sole `--state` consequently verifies the observable error in a directory without that file, rather than proving parser rejection. That distinction does not invalidate the implementation's contract.

An unrecognized dash-prefixed token can be a filename; adding another positional filename makes it a usage error. Existing single-file compatibility justifies this interpretation. Option-shaped filenames in multi-argument filtered invocations can use the `--` delimiter or an explicit `./` path. No new supported-contract regression was demonstrated.

## Reproduction artifacts

- Independent harness: `/private/tmp/review-cli-revision-3-independent.py`, run as `rtk proxy python3 /private/tmp/review-cli-revision-3-independent.py`, exit 0.
- Disposable source copies and binaries: `/var/folders/jt/s80drf_d19n1z3bdj2fzqyb40000gn/T/cli-independent-review-h_7z5z4s`.
- All main CLI/JSON and extended grammar probes passed. The reused harness also ran candidate D's existing tests incidentally; this addendum does not reassess D.
