# Independent review of candidates C and D

Reviewed supplied candidate directories against original `wire-contract` and `not-public-contract` fixtures. No revisions were supplied; the directory diffs define scope. Read both requested review skills and the correctness decision reference. No eval definitions, expected assertions, author reports, build skill, or other trial directories were read. Candidate files were not modified.

# Candidate C: joblist

## Architecture & Design — A
Scope: `/private/tmp/go-quality-build-api-eval/review-skill-on/candidate-c` versus original `wire-contract`; private Go binary with a public 1.x CLI/JSON contract, Go 1.22 module.
Coverage: CLI argument boundary, private data model, JSON mapping, filtering ownership and error propagation. No concurrency or acquired-resource lifecycle was introduced.
Rationale: No actionable architectural issue. The change fits the existing small command and preserves a deliberate mapping between private names and public fields. The filename regression below concerns parser behavior and compatibility, not a demonstrated architectural redesign problem.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate-c/main.go:11–16` retains `json:"id"` while renaming the private field to `Key`; executable checks confirm consumers still receive `id` and `state`, including empty strings. This keeps the private refactor independent of wire naming without unnecessary new layers.
- [C-G2] `candidate-c/main.go:20–35` keeps flag parsing inside the CLI boundary and models flag presence separately from its string value. `main.go:46–52` applies the filter in one existing traversal while retaining order.

Bad

- None found within architecture scope. See C-F1 under correctness for the behavioral edge case.

Suggested changes

- None needed within architecture scope.

Limits: Full small fixture inspected. No external dashboard or shell scripts were supplied. Verification used a disposable copy and Go 1.26.5 on darwin/arm64; Go 1.22 itself was not executed.

## Correctness & Compatibility — A-
Scope: Same directory diff; requested optional exact `--state STATE` filter and private `ID` to `Key` rename, assessed against README's stable 1.x command contract.
Coverage: Omitted, nonempty and explicit-empty state; exact case-sensitive matching and order; no-match, empty array and null; unknown JSON fields; malformed/type-invalid/trailing JSON; missing file; argument errors; JSON names, empty values, newline, stderr and actual process exit statuses. Existing single-file invocation compared against original for a leading-dash filename.
Rationale: One minor, introduced compatibility issue affects a narrow filename case and is recoverable with a path spelling change. Core filter and documented JSON/error behavior work in independent executions. No major or moderate issue was demonstrated.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [C-G3] `candidate-c/main.go:30–35,46–52` correctly distinguishes absent `--state` from `--state ""`; independent executable checks return all records for omission and only empty-state records for the latter. Repeated matching records retain input order and `Ready` does not match `ready`.
- [C-G4] `candidate-c/main.go:12–16,41–55` retains stable JSON tags and nil-result encoding. Independent checks return exactly `{"jobs":null}\n` for no matches, `[]`, and `null`, and preserve empty `id` and `state` fields. Invalid JSON, trailing documents, wrong field type, missing files and usage errors produce exit 2, diagnostic stderr and no stdout.

Bad

- [C-F1][minor][introduced] `candidate-c/main.go:23` now interprets a previously valid sole filename starting with a dash as a flag. With a real file named `-jobs.json` containing `[]`, the original executable invoked as `joblist -jobs.json` exits 0 with `{"jobs":null}\n`; candidate C exits 2 with `flag provided but not defined: -jobs.json` and empty stdout. The original README accepted one JSON file argument without excluding such filenames. Existing scripts using this narrow argument form regress during the 1.x upgrade. Primary remediation owner: Correctness & Compatibility.

Suggested changes

- [C-F1] Preserve supported legacy single-file handling for leading-dash filenames while introducing the optional state syntax, and add an executable or `run` regression using `-jobs.json`. Keep unknown-option error behavior deliberate. A documented `./-jobs.json` or `-- -jobs.json` spelling is a workaround, but does not make the unchanged old invocation compatible.

Limits: Ran `go test ./...` successfully in candidate C and original copies; built both binaries with `go build -o <temporary-binary> .` and exercised them via Python subprocesses. The independent harness is `/private/tmp/review-cli-independent.py`; its completed copy is `/var/folders/jt/s80drf_d19n1z3bdj2fzqyb40000gn/T/cli-independent-review-2ablnq8y`. First attempt was blocked by the default Go cache sandbox path; retry with GOCACHE inside the disposable directory passed. No external consumers or alternate OS/toolchain matrix executed; no writer failure injection performed. Existing permissive JSON decoding outside the introduced behavior was not graded as new debt.

# Candidate D: internal expiry helper

## Architecture & Design — Not applicable
Scope: `/private/tmp/go-quality-build-api-eval/review-skill-on/candidate-d` versus original `not-public-contract`; only private `internal/cache.expired` comparison, its comment, and a regression assertion change.
Coverage: Inspected all fixture files and diff. The helper remains unexported, same signature and package; README states no persisted/serialized representation and only package-local callers.
Rationale: This local boundary-condition fix implicates no new API, package, abstraction or ownership decision. There is no supported public compatibility surface to redesign.
Limits: Scope is the supplied helper fixture, not a full worker implementation. No broader architectural claims are made.

## Correctness & Compatibility — A
Scope: Same candidate D directory diff; requested inclusive deadline comparison with zero meaning never expires; Go 1.22 module.
Coverage: Before, equal to and after positive and negative nonzero deadlines; zero sentinel at negative/zero/positive current times; int64 endpoint equality; helper visibility and regression assertion.
Rationale: No actionable issue. Inclusive comparison directly implements the requested boundary and the existing zero guard preserves permanent entries. These are routine correct handling of the requested conditions, not an A+ claim.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [D-G1] `candidate-d/internal/cache/expiry.go:5` uses `deadline != 0 && now >= deadline`, expiring at equality without arithmetic that could overflow. Independent package-local checks passed at deadline minus one, exact deadline, deadline plus one, and both int64 endpoint equalities.
- [D-G2] `candidate-d/internal/cache/expiry_test.go:12–17` adds the exact-deadline regression while retaining the permanent-entry check. Independent tests confirm zero never expires for now=-1, 0 and 1. The helper remains private as requested.

Bad

- None found.

Suggested changes

- None needed.

Limits: Candidate D `go test ./...` passed before and after adding an independent test only in the disposable copy. The independent test includes 11 boundary cases and passed on Go 1.26.5 darwin/arm64. Go 1.22 was not separately executed; code uses no new language features. No broader cache lifecycle or caller context beyond this fixture was supplied or assumed.
