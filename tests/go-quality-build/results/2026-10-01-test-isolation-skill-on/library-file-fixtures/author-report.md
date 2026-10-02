# Store test extension

Changed only `store_test.go`. The existing serial replacement test now uses `t.TempDir`. Added three parallel independent Store instances, each with its own parent-owned temporary root and the identical `same-key`. Each instance runs serial named child cases for an initial long value, a shorter replacement, an empty value, another replacement, and an identical repeated value. Instance-specific values expose root crossover.

A serial `instances` grouping subtest waits for all parallel children. The parent then opens fresh Store instances and verifies the last expected value remains persisted in each exercised root. Expectations come from selected input values, never production return values. Each child writes only its own instance's bookkeeping, and the parent reads it after the group joins. Tracking whether a root was written also makes parent assertions valid when `-run` selects one instance or only a nested value case. `t.TempDir` cleanup keeps roots alive through descendants and the parent's persistence checks.

Production code, public API, dependencies, and the `go 1.22` module declaration were preserved. All newly used APIs are available in Go 1.22. No tools or dependencies were installed, no delegation was used, and no commit or external mutation was performed.

## Verification actually run

Every shell command used `rtk proxy`. Go commands set `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Process deadlines were 15 seconds for formatting/version checks, 90 seconds for vet and ordinary test runs, and 105 seconds for the shuffled repeated suite. Tests also used finite Go test deadlines.

All recorded commands exited 0:

- `gofmt -w store_test.go`, followed by `gofmt -d store_test.go` (no diff).
- `go version`: Go 1.26.5 on darwin/arm64.
- `go vet ./...`.
- `go test -race -v -timeout=30s ./...`.
- `go test -race -timeout=30s -run=^TestStoreIndependentInstances$ ./...`.
- Three race-enabled, 10-repetition focused leaves, each with `-timeout=30s`: `TestStoreIndependentInstances/instances/beta/replacement`, `TestStoreIndependentInstances/instances/alpha/empty`, and `TestStoreIndependentInstances/instances/gamma/repeat` (exact anchored selection patterns are preserved in checks.json).
- `go test -race -timeout=30s -run=^TestStoreSerialChildren$ ./...`.
- `go test -race -shuffle=on -count=25 -timeout=45s ./...`.

The module declaration was reread after verification and confirmed unchanged. `checks.json` preserves exact command arguments, stdout, stderr, exit codes, and process deadlines.

## Limitations and remaining behavior risks

Staticcheck was not installed and was not run. Verification ran on the available Go 1.26.5 toolchain rather than an actual Go 1.22 installation, and only on this local macOS filesystem. Parallel subtests exercise concurrent independent roots; passing race and repeated runs do not prove every possible schedule free of races or flakes. Concurrent same-key replacement in a shared root remains outside the requested contract and is not exercised. Invalid/missing-key behavior was not part of this extension.

## Guidance selection

Opened go-core-style, go-test-isolation, and the latter's isolation-patterns reference. The other offered guidance was irrelevant because this test-only work changes no public contract, production composition, or package boundary. Exact paths and decisions are recorded in selection.json.
