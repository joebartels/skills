# Trial report

Completed 2026-10-01 in the specified disposable source directory.

## Selected and opened skills

- `/private/tmp/go-language-trials-3b23/go-error-contracts/baseline/reader-errors/first/catalog/go-api-contracts/SKILL.md`: selected because comments change the supported input format and `*LineError` changes the exported error contract.
- No other skill file was opened. Interface/composition and package-boundary skills were not selected because this change keeps the existing `io.Reader` parameter, dependencies, ownership, construction, and package organization.

## Changes and decisions

- `source/load.go`: added exported `LineError` with `Line` and `Text`; retained the existing malformed-error message and `Load` function type. Comments start with `#` at column one. Empty lines remain ignored, values split at the first equals sign, and whitespace remains unchanged.
- Retained `io.ReadAll` for a small implementation with no scanner line limit. Parse the returned bytes even when reading failed, then return the records and original reader error. EOF is handled as normal completion by `io.ReadAll`.
- On malformed input, return only records before the malformed physical line. When the same input also has a reader failure, join the line error and reader error so `errors.As` can find the line error and `errors.Is`/`errors.As` can inspect the original reader error.
- Successful empty input preserves a nil record slice and a nil error interface. Readers remain borrowed and are never closed.
- `source/load_test.go`: external-package tests cover the callable function type, exported error literal/method set, existing ordered records, comments, blank-line numbering, empty keys/values, whitespace, first-equals behavior, final lines, a 128 KiB line, valid prefixes, bytes plus EOF/failure, split reads, no close, simultaneous malformed/read errors, and typed reader-error identity.
- `source/README.md`: documented input grammar, physical line numbering, error inspection, partial results, simultaneous errors, borrowed ownership, and memory behavior.

## Actual checks

- `go version` via `rtk run`: Go 1.26.5, darwin/arm64.
- Initial `rtk run 'go version && go test ./...'`: test setup failed because the default Go build cache was outside the writable sandbox. Switched the build cache to `/private/tmp/recordload-reader-baseline-cache`.
- `rtk run 'GOCACHE=/private/tmp/recordload-reader-baseline-cache go test ./...'` after adding tests: failed to compile with `undefined: recordload.LineError`, confirming the requested exported type was absent before implementation.
- The same test command after implementation and formatting: passed.
- Final command: `rtk run 'gofmt -w load.go load_test.go && GOCACHE=/private/tmp/recordload-reader-baseline-cache go test -race ./... && GOCACHE=/private/tmp/recordload-reader-baseline-cache go vet ./... && test -z "$(gofmt -l load.go load_test.go)"'`: exit 0; race tests passed, vet reported no findings, and formatting check passed.
- Final source inventory contains only `load.go`, `load_test.go`, `README.md`, and `go.mod`; no source-tree binaries were generated. No commit or staging action was performed.

## Limits and execution notes

- Go 1.22 compatibility was kept in source/API usage and the module declaration; the actual available toolchain was Go 1.26.5. Go 1.22 itself was not executed.
- No downstream consumers or release policy were supplied. The external-package tests provide representative consumer checks, not a real downstream integration.
- The implementation reads the full input into memory. No fixed line cap is imposed, but available memory still limits input size.
- The first prompt-file read used a direct `cat` before the prompt's RTK-prefix requirement was known. Subsequent shell calls were prefixed with `rtk`. One attempted `rtk exec` read failed because `exec` is not an RTK command; it was corrected to `rtk run` without changing files.
- No other trial, expectation, review material, or root-repository file was inspected. No repository progress files were updated and no external action was taken.
