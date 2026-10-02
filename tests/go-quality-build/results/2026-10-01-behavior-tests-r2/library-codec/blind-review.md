# Independent review: changeset 1

Review boundary: supplied `original/` versus `candidate/` for the keycodec library; the actual request is escaped field support, high-quality contract tests, and usage documentation. The entire supplied module was inspected. No other candidate or external evaluation material informed this card. File references below are relative to this changeset.

## Testing — A+
Scope: Supplied diff in `candidate/key.go`, `candidate/key_test.go`, and `candidate/README.md`; module minimum Go 1.22, executed with Go 1.26.5 on darwin/arm64.
Coverage: Independent consumer-facing Encode and Decode results; slash, space, percent, Unicode, lowercase escapes, literal plus and exactly-once unescaping; malformed escapes, field emptiness, segment count, sentinel identity and failure values; source function types; executable examples; deterministic fuzz property and finite fuzz execution. No concurrency or external resources are involved.
Rationale: No actionable issue was found. Two independent verified safeguards justify A+: fixed expected wire/field values exercise each public direction independently and detect wrong URL-escaping semantics, while invalid-input tests separately enforce error identity and zero failure results. The first safeguard rejected QueryEscape/QueryUnescape mutations; the second rejected a fresh-error mutation despite its identical error text. These checks control value/wire compatibility and error-contract compatibility, respectively. The broad round-trip fuzz property is additional evidence rather than the sole oracle.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/key_test.go:17` and `candidate/key_test.go:44` assert exact independent results, rather than accepting a matching encoder/decoder pair as sufficient. Disposable query-escaping mutations failed these tests for space and plus semantics.
- [G2] `candidate/key_test.go:72` and `candidate/key_test.go:93` assert `errors.Is` and empty/zero outputs across invalid boundaries. Replacing the returned Decode sentinel with a fresh identically worded error made all invalid-wire cases fail.
- [G3] `candidate/key_test.go:140` asserts arbitrary-byte round trips and empty-field failures; a three-second, one-worker fuzz run passed. `candidate/key_test.go:168` and `candidate/key_test.go:178` verify documented examples as executable output.
- [G4] `candidate/key_test.go:12` compiles the existing function types from an external consumer package.

Bad

- None found.

Suggested changes

- None needed.

Limits: `rtk proxy go test ./...` passed (exit 0) in an unchanged disposable copy. `rtk proxy go test -run=^$ -fuzz=FuzzRoundTrip -fuzztime=3s -parallel=1` passed (exit 0; 333,153 executions, with an existing local fuzz cache contributing baseline inputs). Each described assertion mutation failed as expected (exit 1) in its own disposable copy. Finite fuzzing is not exhaustive. The Go 1.22 toolchain itself and other platforms were not available or executed. Exact command, cwd, non-secret environment, stdout, stderr and exit details are in `output/checks.json`.

## Correctness & Compatibility — A
Scope: Supplied original/candidate diff; the documented 1.x Key, Encode, Decode and ErrInvalidKey contract, with escaped-field support added for the next minor release; Go 1.22 remains declared.
Coverage: Changed encoding, decoding, input validation and error paths; public signatures and function-value use; stored wire representation for simple keys; exactly-once escaping/unescaping, lowercase hex and literal plus; documented usage. No shared mutable state, cancellation, persistence operation or platform-specific implementation is introduced.
Rationale: No actionable introduced or worsened issue was found. Relevant successful and invalid paths are verified, and the implementation uses the intended URL path-segment semantics while retaining the public types and sentinel. This is a complete, small implementation of the requested contract; the evidence supports A without treating routine parser setup as an extra safeguard.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/key.go:24` escapes each field independently before joining. Exact expected tests confirm delimiter escaping, `%20`, Unicode bytes and percent escaping while preserving `east/web`.
- [G2] `candidate/key.go:31` checks the unescaped delimiter structure before unescaping once per field (`candidate/key.go:35` and `candidate/key.go:39`). Verified `%252F` and escaped-delimiter cases preserve field contents instead of changing the wire segment structure.
- [G3] `candidate/key.go:32`, `candidate/key.go:36` and `candidate/key.go:40` reject invalid segment structure or escapes with the existing ErrInvalidKey and zero Key; consumer tests verify that observable failure contract.
- [G4] `candidate/README.md:26` provides usage matching the tested examples, and `candidate/README.md:44` explains percent, plus and exactly-once decoding semantics. The module minimum and exported function types are unchanged.

Bad

- None found.

Suggested changes

- None needed.

Limits: Verification is local Go 1.26.5/darwin/arm64, with the module declaring Go 1.22. Minimum-toolchain execution and cross-platform execution were not performed. The documented rejection rules were assessed; no additional requirement to reject every noncanonical but otherwise parseable legacy spelling was inferred. Passing tests and finite fuzzing are evidence for the inspected contract, not an exhaustive proof. Details are in `output/checks.json`.

## Architecture & Design — Not applicable
Scope: The supplied keycodec diff.
Coverage: Reviewed public signatures, imports, pure codec implementation and test boundaries to establish applicability.
Rationale: The change uses standard-library escaping within the existing pure-function API. It introduces no consequential production/test seam, dependency ownership, lifecycle or package boundary decision; the external-package tests call the existing public API directly.
Limits: Architecture is not graded for local parser implementation choices. Actual consumer compatibility is covered above.
