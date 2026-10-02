# Clamp lower-bound repair

Changed `clamp.go` to return `low` instead of `low + 1` when `value <= low`. This restores the documented inclusive range behavior. The public signature and `go 1.22` module minimum are unchanged.

Added an ordinary `TestLowerBound` regression test in `clamp_test.go`, covering a value below the lower bound, the lower endpoint, and a single-value interval. The existing interior test remains.

## Verification

- Before the fix, `go test -run '^TestLowerBound$' -count=1 -timeout=30s ./...` failed in all three new cases, with exit code 1.
- After the fix, the same focused command passed with exit code 0.
- `gofmt -s -w clamp.go clamp_test.go` completed, and `gofmt -l clamp.go clamp_test.go` returned no files; both exited 0.
- `go vet ./...` exited 0.
- `go test -v -race -count=1 -timeout=30s ./...` passed all tests with exit code 0.
- `go version` reported `go1.26.5 darwin/arm64`.

Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Verification subprocesses had 60-second deadlines; the version check had a 15-second deadline. Shell commands after reading the dispatch were prefixed with `rtk proxy`. Full verification stdout, stderr, and exit codes are preserved in `checks.json`.

## Guidance

Read only the offered `go-core-style` skill. The other three offered skills were not selected because this is a local repair restoring an existing documented contract without API, interface, lifecycle, or package-boundary changes. Selection reasons and opened paths are recorded in `selection.json`.

## Limitations and remaining behavior risks

Staticcheck was unavailable on PATH, and no tool was installed. The installed Go 1.26.5 toolchain performed verification; Go 1.22 itself was not installed or run. The changes use ordinary Go features supported by the declared minimum.

The README requires callers to supply `low <= high`; invalid intervals remain outside the documented contract. The focused regression does not add coverage for the unchanged upper-bound branch. No I/O, process-state, dependency, or concurrency behavior is involved in this repair.
