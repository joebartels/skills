# Store test extension

Changed only `store_test.go` in the isolated module. The original serial replacement checks remain, with their fixture now owned by `t.TempDir`.

Added `TestStoreIndependentInstances`: three independent Store roots use the same key while their instance subtests run in parallel. Each instance has grouped, serial replacement child cases, repeats an earlier value, and verifies exact round trips. The cases include empty values and text containing a newline and NUL. Missing-key reads must fail before writes. The enclosing `instances` group waits for its parallel children; parent-owned temporary directories remain available for final per-instance reads after that group completes. Those final reads also verify cross-root independence.

No production code, public API, dependency, or `go.mod` changes. The module remains `go 1.22`, and the added code uses APIs available at that minimum.

## Verification actually run

All executed checks exited 0. Full command arguments, process deadlines, stdout, stderr, and exit codes are preserved in `checks.json`.

- `go version`: installed local compiler is Go 1.26.5 on darwin/arm64.
- Baseline: `go test -race -count=1 -timeout=30s ./...`.
- `gofmt -w store_test.go`, then `gofmt -l store_test.go` (no output).
- `go vet ./...`.
- `go test -v -count=1 -parallel=8 -timeout=30s ./...`: output shows the alpha, beta, and empty instance subtests continuing concurrently, with all nested cases passing.
- `go test -race -count=50 -shuffle=on -parallel=8 -timeout=30s ./...`.

Every Go invocation used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Go test commands had a 30-second test deadline and 60-second process deadline. No tools or dependencies were installed.

## Limitations and remaining behavior risks

`staticcheck` is unavailable and was skipped. The exact Go 1.22 compiler was not available or run; the unchanged Go minimum and added APIs were checked from the source. Parallel subtest scheduling allows concurrency but does not guarantee overlap of each filesystem operation. Same-key concurrent replacement within one root is outside the README contract and remains untested. These tests focus on independent-root concurrency and grouped fixture lifetime; they do not add a full invalid-key validation matrix or filesystem-failure coverage.

## Guidance selection

Read only the offered `go-core-style/SKILL.md`; the other offered guidance was not relevant because this change modifies tests without changing API contracts, production composition, or package boundaries. `selection.json` records the actual opened task/source/reference paths and each selection decision.
