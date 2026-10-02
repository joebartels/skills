# Trial report

Implemented `CollectMatching` in `source/frames.go`, added tests in
`source/frames_test.go`, and documented the borrowed-view and ownership rules
in `source/README.md`.

## Skills consulted

- Opened `catalog/go-api-contracts/SKILL.md` because this adds an exported API.
- Did not open `go-interfaces-and-composition` or `go-package-boundaries`; no
  interface, dependency wiring, lifecycle, or package responsibility changed.

## Decisions

- Added the specified function signature and routed `Collect` through it with a
  nil predicate, preserving Collect's exact function type and retain-all
  behavior.
- Invoke a non-nil predicate on the borrowed source view, then copy only frames
  accepted by the predicate. Returned copies preserve nil and non-nil empty
  slices and remain stable if the source storage is reused.
- On non-EOF errors, return the accepted prefix and the original error.
- Predicate callers must inspect but not mutate or retain the borrowed view.
  Source ownership remains borrowed and no Close lifecycle was added.

## Verification

- `GOCACHE=/private/tmp/go-language-trials-3b23/go-values-and-zero-values/baseline/borrowed-transfer/first/gocache GOWORK=off go test ./...` passed.
- `GOCACHE=/private/tmp/go-language-trials-3b23/go-values-and-zero-values/baseline/borrowed-transfer/first/gocache GOWORK=off go vet ./...` passed with no diagnostics.
- Compile-time function assignments in package tests check the existing
  `Collect` signature and the new `CollectMatching` signature.

## Limits

Verification covers the package module and compile-time API shapes. No
downstream consumer module or release policy was provided, so external adoption
and compatibility beyond the preserved `Collect` type were not assessed.
