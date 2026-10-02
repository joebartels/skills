# Independent review of the escaped key codec

Review boundary: the complete supplied changeset from `original/` to `candidate/` in `/private/tmp/go-independent-review-b3sit8qm`. This is the `example.com/keycodec` library, with `go 1.22` unchanged. The original README supplies the requested next-minor-release contract: exactly two individually URL path-escaped fields, nonempty values, malformed-escape and wrong-segment rejection, unchanged public signatures, and retained `ErrInvalidKey`. The added documentation and tests were assessed against that contract, rather than treated as their own proof of correctness. No author report or evaluation expectations were inspected.

All file references below are relative to `/private/tmp/go-independent-review-b3sit8qm/candidate/`. Full command arguments, working directories, stdout, stderr, exit codes, and source-integrity hashes are preserved in `output/checks.json`.

## Testing — A+
Scope: supplied original-to-candidate library changeset, including the requested new supported-contract tests. Production `key.go`, all original and candidate tests, README contracts, and `go.mod` were inspected. The testing skill and its local decision reference were applied.
Coverage: independent Encode and Decode success contracts; legacy simple keys; slash in either or both fields; spaces; literal percent and encoded-looking values; Unicode; path versus query plus semantics; lowercase escapes; escaped unreserved characters; empty fields; malformed escapes in either field; wrong segment counts; sentinel identity; zero failure results; external consumer signatures; executable usage; fuzz properties; dependency and lifecycle control. Concurrency, integrations, and benchmarks are not relevant to this pure in-memory codec change.
Rationale: no actionable testing issue was found. Two independent safeguards beyond routine setup were demonstrated. First, known wire/value assertions detect a reciprocal codec error that a round trip cannot distinguish: replacing both path operations with query operations still passes the round-trip tests and fuzz seeds, but fails the independent Encode and Decode tests. Second, invalid-input assertions independently protect caller failure semantics: the tests reject both a partially decoded region on a name-unescape failure and a new error with the sentinel's exact text but different identity. These safeguards control distinct risks: interoperability on successful inputs, and trustworthy result/error behavior on failed inputs. Their effectiveness was executed, rather than inferred from test count or coverage.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `codec_contract_test.go:17` and `codec_contract_test.go:45` use literal expected wires and decoded values independently. The reciprocal query-codec mutation passes `TestSimpleRoundTrip` and the `FuzzRoundTrip` seed run, then fails Encode's space/plus cases and Decode's literal-plus case. This establishes a meaningful oracle for each consumer contract.
- [T-G2] `codec_contract_test.go:74` and `codec_contract_test.go:85` assert `errors.Is` and the entire failure result. The partial-decode mutation fails on `east/%` and `valid%20region/bad%`, among others. The lookalike-error mutation fails both invalid-input groups despite identical error text. Wrong segment counts and malformed escapes in either position also pass against the unmodified candidate.
- [T-G3] `codec_contract_test.go:11` compiles the exact exported function types from an external consumer package; `codec_contract_test.go:101` executes the documented example output. The complete test run includes and passes both safeguards for API/usage accuracy.
- [T-G4] `codec_contract_test.go:118` adds a deterministic nonempty-value round-trip invariant, with an explicit empty-field result/error branch. All four checked-in seeds pass, and a finite native fuzz run passes with 1,060,946 executions. This complements the independent literal expectations; it does not substitute for them.
- [T-G5] Tests directly execute the codec and standard-library serialization, without mocks or injected functions that bypass the change. All tests use local value state; no test-owned servers, goroutines, files, environment mutation, sleeps, or external dependencies require cleanup. This follows from inspection of the complete supplied test files.

Bad

- None found.

Suggested changes

- None needed.

Limits: checks ran on Go 1.26.5, darwin/arm64, with `GOTOOLCHAIN=local`, `GOCACHE=/private/tmp/go-quality-testing-cache`, `GOPROXY=off`, `GOSUMDB=off`, and two-process concurrency. The Go 1.22 directive governs language semantics, but an actual Go 1.22 toolchain and other platforms were not executed. The fuzz run is finite and started from 30 baseline entries including the four checked-in seeds and cached corpus; it does not prove the entire input space. No race run was needed for the inspected stateless paths. Mutation failures are intentional verification successes and apply only to disposable copies.

## Correctness & Compatibility — A
Scope: same supplied library changeset, assessed against the original README's 1.x API and requested escaped-field behavior. The correctness skill and its local decision reference were applied.
Coverage: ordinary, boundary, empty and invalid inputs; segment ordering; per-field escaping; decoding once; malformed-escape errors; absence of partial results; error sentinel compatibility; exported signatures; existing simple-key representation; documentation and dependency/build context. No stateful operations or asynchronous lifecycle were introduced.
Rationale: no introduced or worsened defect was found. The implementation's path-escaping choice, split-before-unescape ordering, and normalized error returns satisfy the supported contract. Successful independent tests verify the key consumer-visible behavior, and the executable example agrees with the added usage. These are verified relevant strengths; the implementation is routine correct codec construction, so the grade is A without asserting additional production safeguards beyond the A+ threshold.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `key.go:24` escapes each field separately before joining with the unescaped separator. Passing independent Encode cases verify literal slash, percent, space, Unicode, plus and reserved characters, including the specified `us%2Fwest/web%20blue` example. It preserves `east/web` for simple existing consumers.
- [C-G2] `key.go:30` checks the raw segment count before independently calling `PathUnescape` at `key.go:34` and `key.go:38`. The Decode cases at `codec_contract_test.go:52`, `codec_contract_test.go:56`, and `codec_contract_test.go:60` verify escaped separators remain field data, `%252F` decodes only once, and lowercase hexadecimal escapes are accepted. Invalid raw segment counts and malformed escapes are rejected by the executed invalid-input suite.
- [C-G3] `key.go:21`, `key.go:31`, `key.go:35`, and `key.go:39` return the original sentinel and empty/zero result on all error branches. Sentinel identity and zero-result assertions pass for empty fields, incorrect segment counts, and errors in either segment. A name failure cannot expose the previously decoded region.
- [C-G4] `Key`, Encode, Decode, and `ErrInvalidKey` retain their exported API; external function-type assignments compile. `go.mod` is unchanged and has no third-party dependencies. `README.md:43` explains one-time decoding and plus behavior; `README.md:52` documents newly escaped spaces/Unicode and preserves simple-key guidance. The executable example matches usage at `README.md:26`.

Bad

- None found.

Suggested changes

- None needed.

Limits: the exact checks and platform/toolchain limits are recorded below and in `checks.json`. No actual oldest-supported-toolchain run or alternate-platform execution was performed. No concrete callers beyond the complete supplied fixture were provided. The old rejection of slash/percent and old literal space/Unicode emission are explicitly changed by the original next-release contract, rather than graded as accidental compatibility regressions. Arbitrary stricter UTF-8 validation or rejection of noncanonical-but-decodable URL escapes is not promised by the supplied contract and was not invented as a requirement.

## Architecture & Design — Not applicable
Scope: supplied production diff and its dependency/lifecycle decisions.
Coverage: package boundary, public API, dependencies, seams and resource ownership were inspected to determine applicability.
Rationale: the change retains one package and all public types/signatures, adds only `net/url` standard-library calls within the existing codec functions, and introduces no interfaces, injection seams, lifecycle responsibilities or consequential package/composition decisions. The optional architecture skill is therefore not needed for a separate graded architecture review.
Limits: this determination applies only to the supplied codec fixture; no larger application architecture was provided or inspected.

## Supporting verification facts

Every Go check uses this prefix: `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOMAXPROCS=2`. The candidate was copied to `verification/` before executing checks; mutations used separate directories beneath the same authorized temporary root.

| Check after the prefix | Working copy | Exit and meaningful result |
| --- | --- | --- |
| `go version` | `verification/` | 0; `go version go1.26.5 darwin/arm64` |
| `go env GOOS GOARCH GOVERSION CGO_ENABLED` | `verification/` | 0; darwin, arm64, go1.26.5, CGO enabled |
| `go test -v -count=1 -timeout=30s ./...` | `verification/` | 0; all tests, four fuzz seeds, and example pass |
| `go test -run=^$ -fuzz=^FuzzRoundTrip$ -fuzztime=5s -parallel=2 -timeout=30s` | `verification/` | 0; 1,060,946 executions; PASS |
| `go test -run=^(TestSimpleRoundTrip\|FuzzRoundTrip)$ -count=1 -timeout=30s` | `mutation-query-pair/` | 0; reciprocal query codec survives round-trip tests/seeds |
| `go test -run=^(TestEncode\|TestDecode)$ -count=1 -timeout=30s` | `mutation-query-pair/` | 1; independent space/plus expectations kill the same mutation |
| `go test -run=^TestDecodeInvalidKey$ -count=1 -timeout=30s` | `mutation-partial-decode/` | 1; zero-result assertions kill partial region returns |
| `go test -run=^(TestEncodeInvalidKey\|TestDecodeInvalidKey)$ -count=1 -timeout=30s` | `mutation-error-identity/` | 1; sentinel assertions kill lookalike errors |

All recorded process stderr streams are empty; stdout and exact argv are retained without editing in `checks.json`. Before/after SHA-256 maps verify that every file in both supplied `original/` and `candidate/` remains unchanged. No candidate repairs or source modifications were made.
