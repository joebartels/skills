# Independent review of the supplied key codec changeset

Boundary: the complete supplied diff from `/private/tmp/go-independent-review-138xcq3s/original` to `/private/tmp/go-independent-review-138xcq3s/candidate`; no revisions were supplied. The actual request was to implement escaped field values according to the original README, add high-quality tests for the supported contract, and document usage. Only the dispatched source, project contracts, and supplied review guidance were inspected. No author report or evaluation expectations were used. Original and candidate source/configuration were preserved; verification and mutations used disposable copies.

## Testing — A+

Scope: the supplied library changeset, particularly `candidate/key_test.go`; module `example.com/keycodec`, declared Go 1.22, executed using Go 1.26.5 on darwin/arm64.

Coverage: assessed independent Encode and Decode success contracts; malformed escapes in both fields; empty fields and wrong segment counts; sentinel identity and failure results; public function signatures and exported-field use; old simple and unescaped inputs; examples; fuzz properties, seeds, determinism, dependency boundaries, and lifecycle. The codec uses actual standard-library serialization, with no mocks. Tests create no servers, files, goroutines, environment mutations, or shared mutable fixtures. Concurrency timing, database isolation, and benchmarks are not implicated. No CI configuration was supplied or inferred to be missing.

Rationale: no actionable issue was substantiated. Two independent, verified safeguards control different meaningful risks beyond merely running tests: the independently specified wire/result tables prevent mutually consistent but incompatible encoder/decoder behavior, while the fuzz property protects against data loss across broader field combinations. In disposable copies, replacing both path functions with query functions passed the round-trip tests and fuzz seeds but failed the exact Encode/Decode tables. Separately, collapsing repeated percent characters passed those exact tables but failed the existing fuzz seed. These are distinct detected failures, not two descriptions of the same safeguard. Additional mutations confirmed the error assertions detect loss of sentinel identity and leakage of a partial result. Grade derives from demonstrated regression detection, not test count, coverage percentage, or formatting.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/key_test.go:17–45` and `:68–98` specify independent expected wires and decoded Keys for slash, space, percent, Unicode, literal plus, lowercase escapes, and a single unescape. The query-codec mutation passed `TestSimpleRoundTrip` and all `FuzzRoundTrip` seeds (exit 0), but these tables failed on spaces and plus handling (exit 1). This safeguards the independent consumer contracts explicitly required by `original/README.md:11–12`.
- [T-G2] `candidate/key_test.go:167–192` asserts exact field preservation for nonempty strings and checks error identity plus empty output for empty inputs. The committed mixed-input seed at `:174` detected a repeated-percent collapse that the fixed conformance tables did not detect: `"\x01+/%%"` became `"\x01+/%"`. A separate invalid-UTF-8 normalization mutation was also detected by that seed. The unmodified bounded fuzz run completed 350,407 executions in approximately three seconds without failure. These checks protect field preservation independently of any particular canonical wire spelling.
- [T-G3] `candidate/key_test.go:47–65` and `:100–133` assert `errors.Is(..., ErrInvalidKey)` and the complete failure result. Fresh errors with identical messages failed the identity assertions, and a malformed second field after a valid escaped first field failed the zero-Key assertion when the decoder was mutated to return the partial region.
- [T-G4] External consumer tests, typed function assignments (`candidate/key_test.go:11–15`), and executable examples (`:147–165`) compile and run in the unmodified suite. Expectations are constants rather than results obtained from the inverse codec. No fake bypasses the changed parsing or serialization.

Bad

- None found.

Suggested changes

- None needed.

Limits: the finite fuzz run is supplementary evidence, not exhaustive proof; its shared Go fuzz cache supplied 33 baseline inputs, including the seven committed seeds. Only darwin/arm64 and installed Go 1.26.5 were executed. The declared Go 1.22 minimum was inspected, but that exact compiler and other platforms were not run. No network access was needed; checks explicitly disabled module downloads and used read-only module mode. Exact command arrays, cwd, stdout, stderr, exit codes, mutation definitions, and source hashes are in `output/checks.json`. Mutation failures belong to disposable copies and are verification successes, not candidate failures.

## Correctness & Compatibility — A

Scope: the complete supplied library changeset, especially `candidate/key.go:20–42`, against the original README contract and original exported API. Module Go directive remains 1.22; executed toolchain/platform were Go 1.26.5, darwin/arm64.

Coverage: traced both functions for normal, escaped, empty, malformed, and wrong-segment-count inputs; literal percent and plus; Unicode; decode-once behavior; lowercase escapes; failure identity and zero results; signatures, field layout, and existing simple/raw decode behavior. Reviewed documentation of changed encoded representations. There is no stateful operation, resource lifetime, cancellation, aliasing container, or newly introduced concurrency path. Build-relevant imports and unchanged module configuration were inspected. The supplied scope has no additional consumer or release-policy files.

Rationale: no introduced or worsened correctness/compatibility issue was substantiated. The candidate implements the requested independently path-escaped segments and retains the 1.x exported types, signatures, and error sentinel. The original rejection of slash/percent was the limitation the task required removing; it is not charged as candidate debt. Spaces and Unicode intentionally receive new escaped encoded representations, as required by `original/README.md:7–10`, and `candidate/README.md:37–44` documents the migration consideration. Relevant behavior and failure handling were verified; an A is supported without claiming additional production guarantees beyond this small codec contract.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/key.go:24` escapes each field separately with `url.PathEscape`. `:30–38` first establishes exactly one raw separator and then unescapes each field once. Independent tests verify encoded slash remains field data, percent-looking values survive one decode, spaces use `%20`, plus remains literal, and Unicode returns unchanged.
- [C-G2] `candidate/key.go:21–22` rejects empty input fields; `:31–40` rejects malformed wires and always returns the original sentinel with a zero Key. The malformed second-field case is verified after a valid first-field decode, so successful first-field work cannot leak as a usable partial result.
- [C-G3] Exported signatures and `Key` field names/types are unchanged. Compiled external function assignments and field literals verify source compatibility. Independent tests preserve `east/web`, raw legacy spaces/Unicode, and valid lowercase escapes. Runnable examples agree with the documented `us%2Fwest/web%20blue` representation.
- [C-G4] The implementation remains stateless and uses only the standard library. The module file is unchanged, and tests/build succeeded with downloads disabled and module files read-only.

Bad

- None found.

Suggested changes

- None needed.

Limits: tests and bounded fuzzing corroborate the traced implementation but do not prove every possible string. The exact Go 1.22 compiler was unavailable to this review; no newer language construct or newly version-sensitive API was identified in the changed code, but only installed Go 1.26.5 was executed. Other platforms were not run. This grade assesses the supplied supported codec behavior, not an unspecified broader consumer/storage migration. There is no race-detector claim: no changed shared-state or asynchronous path requires one here. Full executed-check evidence is preserved in `output/checks.json`.

## Architecture & Design — Not applicable

Scope: the supplied production changes to the two pure functions in the existing package.

Coverage: inspected exported API, package layout, imports, and state/resource ownership for consequential changes.

Rationale: no consequential seam, package boundary, lifecycle, composition, or public API design decision changed. The same two functions now call standard-library path escaping and perform local parsing. Adding the architecture skill/report card is therefore unnecessary under the dispatch condition.

Limits: this does not assess the architecture of a larger application that may use this codec; none was supplied.

## Supporting verification facts

All Go commands used this exact prefix: `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly`. `checks.json` contains the full argv and working directory for every executed check; stderr was empty for every check.

| Recorded check | Go command after the prefix | Verified outcome |
| --- | --- | --- |
| original-suite | `go test -count=1 -timeout=30s ./...` | Exit 0; original simple round trip passed. |
| candidate-suite | `go test -count=1 -timeout=30s ./...` | Exit 0; all tests, fuzz seeds, and examples passed. |
| candidate-bounded-fuzz | `go test -run=^$ -fuzz=^FuzzRoundTrip$ -fuzztime=3s -parallel=1 -timeout=20s ./...` | Exit 0; 350,407 executions, no failure. |
| mutation-query-roundtrip-control | `go test -run=^(FuzzRoundTrip\|TestSimpleRoundTrip)$ -count=1 -timeout=30s ./...` | Exit 0 despite mutually incorrect query escaping. |
| mutation-query-exact-tables | `go test -run=^(TestEncode\|TestDecode)$ -count=1 -timeout=30s ./...` | Exit 1; exact tables detected wrong space/plus behavior. |
| mutation-error-identity | `go test -run=^(TestEncodeRejectsEmptyFields\|TestDecodeRejectsInvalidWire)$ -count=1 -timeout=30s ./...` | Exit 1; same-message replacement errors were detected. |
| mutation-zero-result | `go test -run=^TestDecodeRejectsInvalidWire/invalid_name_after_valid_region_escape$ -count=1 -timeout=30s ./...` | Exit 1; partial region result was detected. |
| mutation-repeated-percent-table-control | `go test -run=^(TestEncode\|TestDecode)$ -count=1 -timeout=30s ./...` | Exit 0; fixed tables alone miss this data-loss mutation. |
| mutation-repeated-percent-fuzz-seed | `go test -run=^FuzzRoundTrip$ -count=1 -timeout=30s ./...` | Exit 1; the committed seed detected repeated-percent loss. |
| mutation-bytes-table-control | `go test -run=^TestEncode$ -count=1 -timeout=30s ./...` | Exit 0; invalid-UTF-8 normalization is outside these table inputs. |
| mutation-bytes-fuzz-seed | `go test -run=^FuzzRoundTrip$ -count=1 -timeout=30s ./...` | Exit 1; seed detected changed region bytes. |

The source-preservation entry in `checks.json` confirms identical SHA-256 values before and after verification for every supplied original/candidate file. No candidate repair was made. No additional issue, optional testing preference, or ungraded related defect was found.
