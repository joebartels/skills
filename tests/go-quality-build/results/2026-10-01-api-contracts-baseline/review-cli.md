# Independent baseline review: candidates C and D

Reviewed candidate source against the supplied original fixtures, without consulting the build skill, evals.json, expected assertions, author reports, or other trials. Review scope is introduced changes. Paths below are relative to `/private/tmp/go-quality-build-api-eval/review-baseline/` unless stated otherwise. Applied the Architecture & Design and Correctness & Compatibility review skills and the correctness decision reference. Architecture's optional decision reference was not needed: these changes introduce no layering, context, error translation, or background-work design.

# Candidate C

## Architecture & Design — A
Scope: Supplied directory diff, `candidate-c` versus original `wire-contract` fixture; private Go implementation of joblist 1.x CLI, module Go 1.22.
Coverage: Flag parsing and optional-value representation, private record rename versus JSON boundary, response construction, dependency and abstraction fit. No service, concurrency, or resource lifecycle changes.
Rationale: No actionable architectural issue found. Explicit filter presence is represented separately from its string value, and the internal name change preserves the external schema. These are appropriate, local decisions for this binary; no additional public Go API or model layer is required. Routine correct implementation supports A.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate-c/main.go:20` and `candidate-c/main.go:56` keep filter presence separate from value; `Set` records explicit empty input. Fresh CLI checks verified omitted versus empty filters produce different intended results.
- [C-G2] `candidate-c/main.go:12` retains `json:"id"` on private `Key`; `candidate-c/main.go:15` retains the response envelope. The implementation rename does not leak through the wire boundary, confirmed by output checks.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied source, README, module declaration, and tests inspected. No external consumers were available beyond the documented scripts/dashboard contract. Checks used Go 1.26.5 darwin/arm64; Go 1.22 was not executed. No newly used feature appears to require a version above 1.22.

## Correctness & Compatibility — A
Scope: Same supplied diff and CLI release context. README lines 3–15 specify the consumer-visible input, JSON, ordering, empty-result, exit, and diagnostic contracts.
Coverage: Omitted, exact, explicit empty and equals-empty filters; no match; empty array and null input; original record order; empty fields; ignored unknown fields; malformed multiple-document input; usage, read, and flag errors. Compared original and changed JSON tags and encoder path. No concurrent state or persistence is implicated.
Rationale: No actionable introduced correctness or compatibility defect found. Both the feature and material existing wire/error contracts passed independent execution. Existing unit tests plus the new exact/empty tests pass. Verification supports A; these checks establish routine correct behavior rather than two additional independent safeguards warranting A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G3] `candidate-c/main.go:44` gates exact comparison on explicit presence, so omission lists all records and an explicit empty value selects only empty-state records. Independent checks verified case/prefix exclusions and retained order among multiple matches.
- [C-G4] `candidate-c/main.go:42` leaves an empty result nil, preserving `{"jobs":null}`; unchanged tags retain `id` and `state` even for empty values. Independently verified empty input, null input, no match, and empty values.
- [C-G5] `candidate-c/main.go:25`, `:28`, `:32`, and `:38` retain error paths before stdout encoding. Independent binary execution verified exit 2, empty stdout, and nonempty stderr for usage/read/input/flag errors; success output ends with a newline.

Bad

- None found.

Suggested changes

- None needed.

Limits: `go test -count=1 ./...` passed, and `go build -o /private/tmp/independent-cli-review-yatv8dp2/joblist .` passed in a disposable copy. Eight success cases and five error cases passed through the built binary. Supported Go 1.22 was inspected but not run; checks used Go 1.26.5 darwin/arm64. The unchanged output-writer failure path was inspected, not fault-injected. No claim is made about unspecified CLI argument conventions.

# Candidate D

## Architecture & Design — Not applicable
Scope: Supplied directory diff, `candidate-d` versus original `not-public-contract` fixture; a private worker's unexported cache helper, module Go 1.22.
Coverage: Inspected complete supplied helper, tests, README, and module declaration. Change is only `>` to `>=` and one regression assertion.
Rationale: No relevant architectural decision is implicated by this local comparison correction. The helper remains private, its signature is unchanged, and README explicitly states that it has no serialized or persisted representation. There is no basis to impose a public API migration or compatibility mechanism.
Limits: The larger worker is not supplied; its absence is not treated as a defect. Fresh package tests and independent boundary checks passed in a disposable copy.

## Correctness & Compatibility — A
Scope: Same supplied diff; requested inclusive expiry at a nonzero deadline, while zero means never expires.
Coverage: Before, at, and after the deadline; zero deadline with negative, zero, positive, and maximal now values; negative deadline; minimum/maximum int64 equality; unchanged helper visibility and signature.
Rationale: No actionable defect found. The inclusive comparison directly fixes the requested equality boundary without arithmetic overflow and preserves the zero sentinel. New regression assertion exercises the original failing case. Routine correct implementation supports A.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [D-G1] `candidate-d/internal/cache/expiry.go:5` uses `deadline != 0 && now >= deadline`; the sentinel remains decisive even when now is large, while nonzero equality expires immediately.
- [D-G2] `candidate-d/internal/cache/expiry_test.go:6` adds the deadline-equality regression alongside existing before/after/zero coverage. Fresh tests and independent boundary cases all passed.

Bad

- None found.

Suggested changes

- None needed.

Limits: Ran `go test -count=1 ./...` successfully in `/private/tmp/independent-cli-review-yatv8dp2/candidate-d`, with an additional review-only table test covering ten boundary combinations. The original candidate was not modified. Go 1.22 was not executed; validation used Go 1.26.5 darwin/arm64. No concurrency behavior is present in this pure helper.

# Verification provenance

Disposable copies: `/private/tmp/independent-cli-review-yatv8dp2/candidate-c` and `/private/tmp/independent-cli-review-yatv8dp2/candidate-d`. All Go execution used a separate `GOCACHE` under that temporary directory and fresh `-count=1` test runs. Candidate C passed `ok example.com/joblist`; candidate D passed `ok example.com/cacheworker/internal/cache`. Review checks did not modify candidate source or repository files.
