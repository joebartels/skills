# Codec implementation report

Implemented escaped field values in `/private/tmp/go-testing-author-z3h6ca5c/library-codec`.

## Change and contract

`Encode` uses `url.PathEscape` independently for Region and Name, joined by one literal slash. `Decode` validates exactly two non-empty wire segments before unescaping each once. Slashes, spaces, percent signs and Unicode are accepted as field data. Plus signs remain literal, and escaped-looking values are not decoded twice. The exported function signatures, Key fields, module path, Go 1.22 directive and ErrInvalidKey sentinel are unchanged. Both functions return their existing zero result on failure.

The README documents usage, error checks, per-field escaping and the wire-byte migration for values previously emitted without escaping. Decoding preserves previous unescaped inputs as well as valid percent escapes; canonical encoding always comes from PathEscape. No third-party dependencies were added.

## Tests and signal

External-package tests check Encode against fixed wire strings and Decode against fixed field values independently. They cover the README example, both fields containing slash/space/percent/Unicode, literal and escaped plus signs, URL punctuation, lowercase hex, escaped-looking data, entirely escaped values and whitespace-only fields. Compile-time function assignments protect supported signatures, and the original simple round-trip regression remains.

Failure tests check errors.Is with ErrInvalidKey and the complete zero result for empty fields, missing/extra separators and malformed escapes in either field, including a bad Name after a successfully decoded Region. These observations catch query escaping, unescaping before splitting, double unescaping, field reversal, silently accepted extra segments, lost sentinel identity and partial results. Runnable examples cover both functions. A fuzz target supplements fixed cases with the non-empty-field round-trip invariant and empty-field rejection; it is not used as the wire oracle.

The added contract tests were run before the implementation change. They compiled and failed meaningfully on the old behavior: slash/percent rejection, unescaped space/Unicode output and rejection of valid escaped wire inputs. Stdout and exit code 1 are preserved in checks.json.

## Verification actually run

All Go checks used GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. Every shell command after reading the task dispatch used rtk, with rtk proxy for raw output. Go subprocesses had a 60-second process deadline; test commands also had explicit test deadlines.

- `gofmt -s -w key.go key_test.go`: exit 0.
- `go version`: Go 1.26.5 on darwin/arm64, exit 0.
- `go test -v -timeout=30s ./...`: contract suite, fuzz seeds and executable examples passed, exit 0.
- `go vet ./...`: exit 0.
- `go test -race -timeout=30s ./...`: passed, exit 0.
- `go test -run=^$ -fuzz=FuzzRoundTrip -fuzztime=3s -parallel=2 -timeout=20s`: passed after 631,014 executions, exit 0.
- `gofmt -l key.go key_test.go`: no output, exit 0.

Staticcheck was unavailable and was not installed because the task prohibits tool installation. Verification stdout, stderr, exit codes, commands and deadlines are preserved in checks.json.

## Limits and remaining behavior risks

The supported Go version was kept at 1.22, and all used APIs exist at that version. Actual execution used the installed Go 1.26.5 toolchain; Go 1.22 itself was not run. External consumers were represented by the external test package and function-type assignments, not real downstream repositories. Fuzzing was bounded to three seconds and cannot exhaust the input space. No concurrency or resource state is added by the codec.

The decoder follows PathUnescape semantics rather than demanding canonical encoded spellings. It accepts legacy unescaped data and lowercase escapes. No Unicode normalization or UTF-8 validation is imposed, consistent with the existing string API. Consumers depending on the old raw wire bytes for spaces or Unicode need the documented migration. No remaining implementation defects are known from the checks run.
