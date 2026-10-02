# Escaped key codec

Implemented in `/private/tmp/go-fresh-author-mm5bi5gc/task-1/module`.

## Changes

- `key.go`: `Encode` uses `net/url.PathEscape` on Region and Name independently. `Decode` checks exactly one raw slash before unescaping each field once with `net/url.PathUnescape`. Empty fields, malformed escapes and incorrect segment counts return `ErrInvalidKey`, preserving empty/zero failure results. The existing exported struct, function signatures and sentinel are unchanged.
- `key_test.go`: external-consumer tests assert independently known encoder bytes and decoder values for ordinary keys, slash, spaces, percent, Unicode, literal plus, encoded delimiters and lowercase escapes. Failure tests check `errors.Is` and complete zero results for empty fields, malformed escapes in either field and incorrect segment counts. Exact function-value assignments preserve signature coverage. The original ordinary round-trip regression remains. A fuzz round-trip test supplements the fixed wire cases. Both usage examples execute under `go test`.
- `README.md`: added runnable usage and explained path escaping, split-before-unescape behavior, single decoding of `%252F`, lowercase escapes and sentinel handling. `go.mod` remains at Go 1.22 with no dependencies.

## Checks and evidence

All verification commands and their stdout, stderr, exit codes, cwd, non-secret environment and finite deadlines are in `checks.json`. Commands run through `rtk proxy` with `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`.

1. New tests against the original implementation compiled and failed (exit 1) on the promised escaped encoder bytes and decoded fields, plus the new examples and escaped fuzz seeds. This confirms the suite detects the missing behavior rather than only a compilation failure.
2. `gofmt -s -w key.go key_test.go` passed.
3. `go test -v -timeout=30s ./...` passed, including both independent contract tables, rejection tables, old round trip, fuzz seeds and examples.
4. `go vet ./...` passed.
5. `go test -race -timeout=30s ./...` passed.
6. `go test -run=^$ -fuzz=^FuzzRoundTrip$ -fuzztime=5s -parallel=1 -timeout=30s` passed with 572,935 executions. The shared mandated Go cache supplied additional baseline corpus inputs.
7. `staticcheck ./...` could not run (exit 1): the executable is not installed. No tool was installed.

The installed local toolchain reports `go1.26.5 darwin/arm64`. Only APIs available by Go 1.22 are used, but the exact Go 1.22 toolchain was not available or exercised.

## Contract observations and limits

The fixed encoder cases catch raw concatenation, query-style space escaping, missing field escaping and wrong field order. The independent decoder cases catch rejection of valid escapes, splitting after unescaping, double unescaping and treating plus as a space. Failure cases keep the other field valid, so errors in one field cannot mask the other's validation branch.

This module has no actual downstream consumer repositories. External-package tests are representative consumers, not evidence of a real downstream integration. The change deliberately adopts escaped output for fields such as space and Unicode; consumers constructing the wire format must use the documented per-field escaping. Decode follows standard URL path-unescaping semantics and retains handling of raw values without percent escapes.

No listeners or network boundaries are involved. No external changes, delegation, installations or commits were made. The only remaining verification limits are unavailable staticcheck, the unexercised exact minimum toolchain, bounded fuzzing and absent real downstream integrations.

`selection.json` records every offered guidance name, actual opened guidance/reference paths, relevance decisions and source reads.
