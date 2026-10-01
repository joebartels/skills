# Independent Go API review

Reviewed 2026-10-01. Changeset interpretation: candidate-a relative to the supplied source-compatibility fixture, and candidate-b relative to the supplied intentional-break fixture. No Git revisions were supplied. Paths below are relative to `/private/tmp/go-quality-build-api-eval/review-skill-on/` unless explicitly identified as original fixtures. Read both topic skills and their decision references. Did not consult the build skill, evaluation assertions, author reports, or other trial outputs. The explicitly assigned factory probe was the sole file read from the baseline results directory.

# Candidate A — recordfmt minor release

## Architecture & Design — A
Scope: candidate-a versus original source-compatibility; public library v1.4.2 to v1.5.0; module declares Go 1.22.
Coverage: exported constructor/function-value contract, construction and formatting state, zero values, separator semantics, existing consumers described in README, and tests. No I/O, resource lifecycle, asynchronous work, interfaces, or cross-package data mapping is introduced.
Rationale: No actionable design issue. The additive constructor fits the small, fixed configuration need, and preserves the established constructor type. This is a direct, proportionate API design rather than an additional abstraction framework.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate-a/format.go:12` retains `func(string) *Formatter`; `NewWithSeparator` at line 16 is additive. The external factory table probe compiles and passes, supporting the independently released consumers described in `README.md:5`.
- [A-G2] `candidate-a/format.go:8` explicitly distinguishes an unset separator from an explicitly empty separator. The state remains private, constructor configuration is immutable through the exported API, and `Format` performs no I/O or mutation.

Bad

- None found.

Suggested changes

- None needed.

Limits: Full supplied implementation, README, and tests were inspected. Tests ran on Go 1.26.5 darwin/arm64 with module language version 1.22; no Go 1.22 binary or other platform matrix was exercised. Real external reporting-tool source is unavailable; the supplied external factory test exercised the stated contract.

## Correctness & Compatibility — A+
Scope: same candidate-a changeset and supported minor-release consumers.
Coverage: direct construction, exact constructor function type, external factory registration, explicit empty and nonempty separators, zero-value formatting, empty key/value inputs, and comparison with original behavior. Nil receiver behavior is unchanged and not newly promised.
Rationale: No actionable defect. Two independent verified safeguards control distinct compatibility risks: the exact-type compile-time assertion plus external factory probe protect source compatibility, while the separator-presence representation plus zero-value and empty-separator tests protect default behavior without erasing a valid empty value. These explicitly protect upgrade cases beyond the ordinary happy-path formatting test.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G3] `candidate-a/format_test.go:8` asserts the old constructor function type; the assigned external `factory_test.go` passed from a separate consumer module with a local replacement.
- [A-G4] `candidate-a/format.go:23` restores colon for unset separator state, while `NewWithSeparator` sets presence even for an empty string. Candidate tests pass for both custom and empty separators. An independent probe of zero-value formatting for `("k", "v")` and `("", "")` passes against both original and candidate, producing `k:v` and `:` respectively.

Bad

- None found.

Suggested changes

- None needed.

Limits: Checks and runtime limits are recorded below. Tests and code inspection cover the supplied contract, not unavailable external consumers' every possible use.

# Candidate B — portnum major release

## Architecture & Design — A
Scope: candidate-b versus original intentional-break; public library and example CLI, accepted v2.0.0 decision; module declares Go 1.22.
Coverage: accepted parsing API, major-version import boundary, error contract, CLI ownership of process exit/output, migration documentation, and absence of a legacy shim. Resource lifecycle and background work are not implicated.
Rationale: No actionable architecture issue. The accepted major release explicitly authorizes the signature break; retaining v1 compatibility inside v2 would oppose that decision. The small pure parsing function and CLI error translation fit the actual consumers.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-G1] `candidate-b/go.mod:1` uses `example.com/portnum/v2`; `candidate-b/cmd/portcheck/main.go:4` imports that path. `candidate-b/port.go:11` exposes exactly the accepted `(int, error)` API without a second legacy parsing API.
- [B-G2] Parsing returns errors to its caller; process exits and diagnostics are owned by `candidate-b/cmd/portcheck/main.go:15`. Built CLI probes confirm error cases cannot fall through to printing an apparently valid zero.

Bad

- None found within architecture. The migration-text defect B-F1 is graded under correctness below; it does not change the API or ownership design.

Suggested changes

- None needed for architecture.

Limits: Assessed supplied source and release decision, not publication to a module registry or availability of historical release tags. The accepted decision states v1 remains available; publishing was not part of this review.

## Correctness & Compatibility — A-
Scope: same candidate-b changeset, including requested migration guide and CLI behavior.
Coverage: decimal input grammar, zero/max boundaries, every integer in 0..65535, leading zeros, signs, whitespace, non-ASCII digits, overflow, empty input, returned value on failure, updated consumer, module import path, and old-to-new behavioral documentation.
Rationale: The intended breaking API is implemented correctly and is authorized by the major-release decision. One minor introduced documentation defect misstates a narrow old behavior; no parser or CLI defect was demonstrated.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [B-G3] `candidate-b/port.go:15` rejects non-ASCII decimal bytes before conversion, and line 21 restricts range to 16 unsigned bits. Tests cover every valid numeric value plus independent syntax/overflow cases. One thousand leading zeros followed by `80` correctly returns 80.
- [B-G4] `candidate-b/cmd/portcheck/main.go:16` handles invalid input before output. The compiled executable returns exit 2 with empty stdout and a stderr diagnostic for invalid values and wrong argument counts; valid inputs produce the number plus newline, no stderr, and exit 0.

Bad

- [B-F1][minor][introduced] `candidate-b/docs/migration-v2.md:20` says “In v1 these failures returned zero without an error” after a list including signs. Original `port.go:8` uses `strconv.Atoi`, so `+80` previously returned 80. An independent original-fixture probe confirms that result, while the candidate rejects `+80`. This understates an actual valid-to-invalid input change for migrating callers. The new grammar is documented correctly, so the consequence is limited migration guidance rather than an implementation failure. Primary remediation owner: Correctness & Compatibility.

Suggested changes

- [B-F1] State explicitly that v1 accepted a leading plus sign for an otherwise valid port, whereas v2 rejects all signs. Limit the old-zero statement to inputs v1 actually rejected. Verify the guide against original `Parse("+80") == 80` and candidate `Parse("+80") == (0, non-nil error)`.

Limits: No guarantee of stable error identity or error text was requested, so fresh error values are not a defect. No platform-specific behavior was found in this range-limited parser, but only darwin/arm64 was run. Candidate CLI has no committed tests; its behavior was verified independently here, and a separate testing-quality audit was not requested.

# Executed verification

All additional files and build products were created in disposable copies at `review-library-probes/`; candidate and repository files were not edited.

- Toolchain: `go version` reported `go1.26.5 darwin/arm64`.
- Environment for final runs: `GOWORK=off`, `GOTOOLCHAIN=local`, and `GOCACHE=/private/tmp/go-quality-build-api-eval/review-skill-on/review-library-probes/cache`.
- `go test ./...` in disposable `a`, `b`, `original-a`, `original-b`, and external `consumer` modules: all exit 0. The first attempts for b/original-b encountered a sandbox-denied default build cache path; rerunning with the writable cache resolved the environmental failure.
- Assigned factory test copied unchanged into `consumer/factory_test.go`, with `require example.com/recordfmt v1.4.2` and local `replace => ../a`: passed.
- Independent `review_zero_test.go` in a and original-a: passed on zero-value formatting with normal and empty inputs.
- Independent `review_contract_test.go` in b: all 65,536 valid port values passed; rejected `0x50`, `1_000`, embedded NUL, trailing newline, `-0`, `+0`, 65536, and 1,000 nines with zero plus non-nil error; accepted 1,000 leading zeros followed by 80.
- Independent original-b sign probe: passed, confirming `Parse("+80") == 80`.
- `go build -o ../portcheck ./cmd/portcheck` (actual output path was absolute in the disposable directory): exit 0.
- Built CLI inputs `00080`, `0`, and `65535`: exit 0, respective stdout `80\n`, `0\n`, `65535\n`, empty stderr.
- Built CLI inputs `+80`, `65536`, empty string, no arguments, and two arguments: exit 2, empty stdout, diagnostic on stderr.

No live network or package download was needed. No issue counts are combined across topics.
