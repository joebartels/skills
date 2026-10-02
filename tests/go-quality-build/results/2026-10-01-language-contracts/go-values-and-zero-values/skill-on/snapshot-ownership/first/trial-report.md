# Trial report

Date: 2026-10-01

## Selected and opened skills

- `catalog/go-api-contracts/SKILL.md` — selected because `Graph.Fork() Graph` adds an exported method. The API stays additive, and an external-package test checks its usable method set and behavior.
- `catalog/go-values-and-zero-values/SKILL.md` — selected because the change defines borrowed versus owned data and deep-copy behavior.
- `go-interfaces-and-composition` was not selected: this change introduces no interface, dependency, constructor, options API, or lifecycle decision.
- `go-package-boundaries` was not selected: the existing package keeps its responsibility and import direction.

## Decisions and artifacts

- Added `Graph.Fork`, which recursively clones the reachable nodes and memoizes original node pointers so shared targets and cycles retain their shape.
- Cloned maps, byte payloads, and link slices; preserved nil versus non-nil empty maps, slices, and payloads. A nil root and the zero `Graph` remain usable.
- Kept `RootView` borrowing the original root and documented borrowed view and owned fork behavior on the exported methods.
- Added package tests for bidirectional mutation isolation, map edits, byte edits and append paths, link append paths, shared nodes, cycles, nil links, nil/empty collections, and zero-value use. Added an external-package API test.
- Changed files: `source/graph.go`, `source/graph_test.go`, `source/graph_api_test.go`.

## Checks

- `GOWORK=off GOCACHE=/private/tmp/go-language-trials-3b23/gocache-values go test ./...` — passed (`ok example.com/editgraph`).
- `GOWORK=off GOCACHE=/private/tmp/go-language-trials-3b23/gocache-values go vet ./...` — passed with no diagnostics.
- `git diff --check` — passed.

All shell commands used the `rtk` prefix. The Go cache is outside the source directory.

## Limits

The package tests and an external-package compile/run exercise the local API; no downstream repository consumer was available. Concurrent mutation is outside the documented contract and was not tested.
