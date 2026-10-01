# Blind baseline library review

Reviewed 2026-10-01. Candidates were compared with the supplied original source-compatibility and intentional-break fixture directories. No build skill, expected assertions, evals.json, author reports, or other trials were read. Architecture & Design and Correctness & Compatibility review skills and their decision references supplied the grading contract. Findings concern introduced changes. Paths below are relative to `/private/tmp/go-quality-build-api-eval/review-baseline/` unless stated otherwise.

## Candidate A — Architecture & Design — B
Scope: Original source-compatibility fixture versus candidate-a; public recordfmt library, v1.4.2 to v1.5.0, Go 1.22 declared.
Coverage: Full implementation, public constructors and formatter state, README, tests, external factory registration, default and explicit-empty behavior. No dependency, lifecycle, I/O, or concurrency machinery is introduced.
Rationale: One moderate state-contract issue. The additive constructor is otherwise a small design that fits the request and preserves factory registration. A1 is primarily owned by Correctness & Compatibility; its architectural consequence is the loss of a representation distinguishing default zero state from explicitly configured empty separation.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [A-G1] candidate-a/format.go:12 retains the exact `func(string) *Formatter` constructor type and delegates to the new constructor. The supplied external factory-table probe compiles and passes.
- [A-G2] candidate-a/format.go:16 exposes the requested separator through a direct constructor with immutable private strings. This is proportionate to one setting and preserves explicitly empty separator values.

Bad

- [A1][moderate][introduced] candidate-a/format.go:5 and :22 use the zero string separator both for unconfigured formatter state and intentional empty separation. An exported `Formatter` could previously be zero-initialized and format `key:value`; it now formats `keyvalue`. This produces a contained output change for clients embedding or allocating the public type without New. The original public method had no constructor-only precondition, and README:4 promises minor releases preserve supported consumers. Actual independently released clients were not supplied; the demonstrated client is a reviewer probe.

Suggested changes

- [A1] Preserve a distinguishable default state, for example a private boolean recording that a separator was explicitly supplied, and use colon for the zero state. Keep `NewWithSeparator(prefix, "")` producing no separator. Verify both zero initialization and explicit-empty construction, plus the factory-table probe.

Limits: Full small source area inspected. No actual external consumer repository was available. The zero-value compatibility judgment relies on previously usable exported state and the broad minor-release promise; README illustrates construction through New without explicitly documenting or forbidding zero-value use. Fresh package tests and registered-factory probe passed using Go 1.26.5 darwin/arm64. Go 1.22 itself was not installed or exercised. Check details follow below.

## Candidate A — Correctness & Compatibility — B
Scope: Same supplied before/after library snapshots, including the minor-release preservation policy.
Coverage: Old constructor signature and function-value assignment; colon default; custom and empty separators; exported zero value; string concatenation and absence of mutable shared state. Module path and Go directive remain unchanged.
Rationale: One moderate compatibility issue, A1, yields incorrect formatting for a narrow class of existing users. No evidence supports systemic, major, or critical reach. Ordinary constructor consumers and the explicitly identified factory-registration consumer work.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [A-G3] candidate-a/format_test.go:8, :15, and :22 passed fresh, checking old default output, a custom separator, and explicit empty separator.
- [A-G4] Supplied evaluator-only factory_test.go passed in a disposable external module with a local replacement to the unmodified candidate. New remains assignable in `map[string]func(string) *recordfmt.Formatter`.

Bad

- [A1][moderate][introduced] candidate-a/format.go:22 changes zero-value formatting from `key:value` to `keyvalue`. The identical external `TestZeroValue` passed against the original and failed against the candidate with `zero Format = "keyvalue"`. This is shared with the architecture finding, not a second defect.

Suggested changes

- [A1] Preserve the colon behavior for an unconfigured Formatter while tracking explicitly empty separators separately, and retain a regression probe covering this distinction.

Limits: Zero-value support has the qualification recorded in the architecture section. No claim that a supplied real integrator used this path. No Go 1.22 runtime or other platform execution. No nil-receiver contract was inferred. All exercised checks used disposable copies or external modules and preserved candidate files.

## Candidate B — Architecture & Design — A
Scope: Original intentional-break fixture versus candidate-b; portnum v1.9.0 to accepted v2.0.0 branch, Go 1.22 declared.
Coverage: Complete parsing API and implementation, module path, example command, accepted decision, tests, and migration guide. Assessed error ownership, process exit ownership, package boundaries, and the intentional breaking release.
Rationale: No actionable architectural defect. The explicit `(int, error)` API fits the accepted contract; the library stays pure and the executable owns diagnostics and process exit. These are verified strengths, without claiming two safeguards beyond routine correct setup for A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-G1] candidate-b/port.go:11 exposes the accepted Parse signature, and candidate-b/go.mod:1 adopts `example.com/portnum/v2`. There is no legacy shim competing with the new contract.
- [B-G2] candidate-b/cmd/portcheck/main.go:14 handles errors at the application boundary. The library contains no process exits or I/O; command probes verified stderr-only failures and exit 2.
- [B-G3] candidate-b/MIGRATING.md:3 explains the import change, error handling, strict input grammar, and continued availability of v1.

Bad

- None found.

Suggested changes

- None needed.

Limits: Scope is the supplied small library and example executable; publication/tagging of an actual v2 release was not tested. The migration guide's assertion that v1 remains available is the accepted release plan, not independently checked registry availability. Go 1.22 itself was not run.

## Candidate B — Correctness & Compatibility — A
Scope: Same before/after snapshots and the explicitly accepted v2 decision; preserving the v1 signature is intentionally not required.
Coverage: ASCII grammar, signs/whitespace/non-digits, empty input, all valid integer values, long leading zeros, bounds and overflow protection, error results, module/import consistency, CLI success/failure streams and exit codes, migration instructions.
Rationale: No actionable correctness defect. The approved breaking signature and path change are implemented consistently. Fresh tests and independent probes verify the significant behavior; routine correct validation does not alone warrant A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-G4] candidate-b/port.go:19 rejects every non-ASCII decimal byte, and :23 checks the range before multiplication. Independent tests passed for all 65,536 valid values and each with leading zeros, as well as long zero padding, overflow, Unicode digits, signs, whitespace, NUL, and other invalid text.
- [B-G5] candidate-b/port.go:13, :20, and :24 return zero with a non-nil error on every failure path. Candidate tests check error identity with errors.Is; fresh tests passed.
- [B-G6] candidate-b/cmd/portcheck/main.go:10 and :14 meet the accepted command behavior. Built-binary probes verified successful normalized output and invalid-input diagnostics only on stderr with exact exit 2.

Bad

- None found.

Suggested changes

- None needed.

Limits: Go 1.26.5 darwin/arm64 executed with module Go directive 1.22; no actual Go 1.22 toolchain or alternate platform run. No fuzz campaign was performed. The parser is deterministic and all value bounds were exercised; arbitrary byte strings were sampled, with the per-byte validation inspected. No unsupported v1 compatibility requirement was imposed.

## Executed verification

All check artifacts reside in `/private/tmp/go-quality-library-review-checks`. Candidate source and repository were not edited. Copies excluded `.gocache`. Commands below were invoked through `rtk proxy`; Go commands used `GOCACHE=/private/tmp/go-quality-library-review-checks/cache GOWORK=off`.

- `go version`: go1.26.5 darwin/arm64.
- In candidate-a copy: `go test -count=1 ./...` — passed.
- In candidate-b copy before reviewer additions: `go test -count=1 ./...` — passed; command package has no tests.
- In consumer-a external module, requiring example.com/recordfmt v1.4.2 and replacing it with candidate-a: `go test -count=1 -run TestRegisteredFactory ./...` — passed. factory_test.go was copied unchanged from the supplied evaluator-only probe.
- Identical external tests using original and candidate replacements: `go test -count=1 ./...` — original passed; candidate failed only TestZeroValue (`keyvalue` versus `key:value`). Reviewer zero-value test is stored in consumer-a/zero_test.go and consumer-original/zero_test.go.
- In candidate-b disposable copy after adding reviewer_test.go: `go test -count=1 ./...` — passed. Probe checked 0..65535 each plain and prefixed with 000; empty/sign/space/newline/Unicode/decimal-point/hex/NUL text; 65536; 1,000 nines; and 1,000 leading zeros followed by 65535.
- In candidate-b copy: `go build -o /private/tmp/go-quality-library-review-checks/portcheck ./cmd/portcheck` — passed.
- Python subprocess assertions against built portcheck: 00080 and 65535 produced `80\n` and `65535\n`, empty stderr, exit 0. +80, 65536, and empty argument produced empty stdout, diagnostic stderr, exit 2. Zero or two arguments produced empty stdout, usage stderr, exit 2. All assertions passed.

No grades from other reviewers or trials informed this report. A1 is one shared root cause across the two topic grades.
