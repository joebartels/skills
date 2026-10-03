# Serial control and routing trial — 2026-10-03

Workspace: `/private/tmp/go-skill-recovery-20261003/concurrency-control`.
Read all seven guidance name/description frontmatters before choosing bodies.
Pure control selection below is independent of the two routing-card decisions.

## Pure SquareSum selection

- Selected `go-behavior-tests`: fixes supported behavior and adds a regression.
- Declined `go-concurrency-and-ownership`: private serial computation; no shared state, goroutines, channels, admission, or concurrent resources.
- Declined `go-context-and-deadlines`: pure calculation; no context, cancellation, or budget decision.
- Declined `go-api-contracts`: restores the existing contract without changing the exported signature or supported API.
- Declined `go-interfaces-and-composition`: no interfaces, dependencies, constructors, wiring, or lifecycle changes.
- Declined `go-package-boundaries`: unchanged package, responsibility, and imports.
- Declined `go-test-isolation`: deterministic local inputs and ordinary assertions.
Opened the behavior-tests body; its optional failure-stage reference was unnecessary here.

## Separate routing decisions

`lease-workers/README.md`: select concurrency and context.
Concurrency owns admission through completed Close, independent lease lifetime, late acquisitions, stop/join/release, and borrowed channel ownership.
Context owns cancellation boundaries, coordinated stop acknowledgements, independent failures, and caller classification/custom cause retention.
`early-pipeline/README.md`: select concurrency and context.
Concurrency owns concurrent callbacks, early consumer completion, producer stop, actual callback joins, channel closure, and buffer-bound interpretation.
Context owns pre-canceled entry, propagation, independent failures, and caller classification/custom cause retention.
Opened both applicable bodies for these routing decisions; read only the two case README contracts.
Did not implement those cases or read their source, probes, history, reports, or other trial workspaces.

## Implementation and verification

`calc.go`: sum `value * value` synchronously; preserve API, package, dependencies, and `go 1.22`.
`calc_test.go`: one new regression, `[-3, 2, 0, 4] => 29`; retained the nil-input zero test.
All Go checks used `GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache` and shell commands used `rtk`.
Pre-fix `go test -count=1 ./...`: exit 1; regression returned 3 instead of 29, a meaningful assertion failure.
Post-fix `go test -count=1 ./...`: exit 0; `ok example.test/serialcalc 0.175s`.
`go vet ./...`: exit 0, no diagnostics. `gofmt -w calc.go calc_test.go`: exit 0.
Runtime: `go1.26.5 darwin/arm64`; a Go 1.22 toolchain run was not performed. No race check was needed for serial local work; overflow is outside the supplied contract.
