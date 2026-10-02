# Store test extension

Changed only `module/store_test.go`.

- Retained the serial same-key replacement checks, replacing manual temporary-directory removal with `t.TempDir` ownership.
- Added `TestStoreIndependentInstances/instances` with parallel `alpha`, `beta` and `empty` children. Each Store receives a distinct parent-owned temporary root while using the identical `same-key` key.
- Grouped `initial` and `replacement` value cases remain serial within each root. They verify complete values, including a multiline replacement and an empty replacement.
- Parent assertions run after the parallel group completes and verify persisted final values. Expected values come from the executed input cases, so focused instance and leaf selection does not assume omitted siblings ran. Each parallel child mutates only its own case state.
- Roots belong to the outer test and survive all descendants and final parent reads. The testing runner owns parallel worker completion; no manual goroutines, shared gates, process-global changes or custom cleanup ordering are introduced.

## Actual verification

All 13 recorded commands completed with exit code 0 and without process timeout. `checks.json` preserves exact command arrays, cwd, non-secret environment, stdout, stderr, exit status and deadline. Formatting produced no outstanding files. `go vet ./...` passed.

The following test runs passed with race detection and a 30-second Go test deadline plus a 60-second process deadline:

- Verbose full suite.
- Full suite with `-shuffle=on -count=30`.
- The isolated `alpha` instance.
- The `alpha/initial` leaf, omitting its replacement and other instances.
- The `beta/replacement` leaf, omitting its initial value and other instances.
- The `empty/replacement` leaf.
- Full suite with `-parallel=1`.

Every shell command after reading the standalone dispatch used `rtk` with raw commands through `rtk proxy`. Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`.

## Limits and remaining risks

The local toolchain was Go 1.26.5 on darwin/arm64; `go list -m -json` confirmed the unchanged Go 1.22 module minimum. Tests use Go 1.22-compatible APIs and language features, but no execution under an actual Go 1.22 binary was performed. `staticcheck` was unavailable; it was not installed. Production source, public API and dependencies remain unchanged.

The checks exercise real local filesystem behavior and parallel independent roots. No listeners or external services are involved. Same-key concurrent replacement within one root remains outside the requested contract. Passing race and repeated runs cover the exercised schedules, not every possible schedule or filesystem/platform behavior.

No requested implementation work remains. Selection decisions and the actual guidance/reference paths opened are recorded in `selection.json`.
