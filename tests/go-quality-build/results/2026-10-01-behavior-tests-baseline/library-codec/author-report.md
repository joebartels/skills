# Key codec implementation

Implemented independent URL path escaping for Region and Name in `key.go`.
Decode verifies exactly two nonempty wire segments before unescaping each once.
All invalid inputs still return the existing `ErrInvalidKey` sentinel and the
same empty-string or zero-Key result. Exported signatures and Key fields remain
unchanged. `go.mod` still declares Go 1.22 and has no added dependencies.

Added `codec_contract_test.go` in the external consumer package. Encode and
Decode have separate literal fixtures for legacy keys, slash in either field,
spaces, percent, Unicode, path-plus semantics, and reserved characters. Decode
also exercises lowercase escapes, escaped unreserved characters, and single
unescaping. Negative tests cover empty fields, incorrect segment counts, and
malformed percent escapes in either field, including a successfully decodable
first field followed by an invalid second field. Consumer function assignments
verify the public signature shapes. An executable example documents usage, and
a seeded fuzz target supplements the independent fixtures with round trips.

Updated README.md with a full usage example, error and failure-result behavior,
escaping details, and minor-release migration guidance. The prior implementation
emitted spaces and Unicode literally; their newly escaped bytes are called out.
Simple keys retain their existing bytes.

## Verification actually run

All shell commands after reading the dispatch used `rtk` (raw output through
`rtk proxy`). Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache`
and `GOTOOLCHAIN=local`. Each verification command had a finite subprocess
deadline; tests additionally used `-timeout=30s`.

- New contract tests against the original source: expected exit 1, demonstrating
  missing escaping and decoding behavior before implementation.
- `gofmt -w key.go codec_contract_test.go`: exit 0.
- `go test -race -cover -count=1 -timeout=30s ./...`: exit 0;
  100.0% statement coverage, including the executable example and fuzz seeds.
- `go vet ./...`: exit 0.
- `gofmt -l key.go key_test.go codec_contract_test.go`: exit 0, empty output.
- `go version`: Go 1.26.5 on darwin/arm64.
- `which staticcheck`: exit 1; staticcheck was unavailable and not installed.
- `go test -run=^$ -fuzz=^FuzzRoundTrip$ -fuzztime=5s -parallel=2 -timeout=30s ./...`:
  exit 0; 1,012,843 executions, no failure.

Exact commands, environment settings, deadlines, stdout, stderr, and exit codes
are preserved in `checks.json`. `run_check.py` is the report-side capture helper.
The original `key_test.go` remains intact.

## Limitations and remaining behavior risks

Verification used the available local Go 1.26.5 toolchain, rather than an actual
Go 1.22 installation. Source and tests use APIs and language features available
in Go 1.22, and the module's version declaration was preserved. Staticcheck
could not be run. The five-second fuzz run is finite and does not establish
exhaustive behavior.

Decode uses standard `url.PathUnescape` behavior. It accepts alternate valid
escape spellings and does not require the wire to equal Encode's canonical
output. No additional UTF-8 validation or input size limit was introduced.
Consumers that previously interpreted wire fields without Decode must account
for escaped bytes; this migration is documented. There were no supplied real
downstream consumers, so consumer evidence is the external-package fixture and
signature tests in this module.

## Artifacts

- Module: `/private/tmp/go-testing-author-7r8p3riu/library-codec`
- Guidance selection: `/private/tmp/go-testing-author-7r8p3riu/library-codec-report/selection.json`
- Verification capture: `/private/tmp/go-testing-author-7r8p3riu/library-codec-report/checks.json`
- Report: `/private/tmp/go-testing-author-7r8p3riu/library-codec-report/author-report.md`

No delegation, dependency/tool installation, commits, or external changes were
performed.
