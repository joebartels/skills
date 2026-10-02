# Trial report: snapshot ownership

## Skills

- Selected and opened `catalog/go-api-contracts/SKILL.md` because `Graph.Fork` is a new exported method and existing zero-value, value-method, and borrowed-view behavior must remain supported.
- Did not select `go-interfaces-and-composition`: the change adds no interface, dependency, constructor, options, or resource lifecycle.
- Did not select `go-package-boundaries`: package responsibilities and imports do not change.

## Decisions

- Added `Graph.Fork() Graph`. A nil root returns a usable zero `Graph`.
- Cloned each reachable node once with a node-pointer memo, preserving shared links, cycles, and nil links while assigning distinct nodes to the fork.
- Copied maps, link slices, and byte payloads. Preserved nil maps/slices and non-nil empty maps/slices and byte payloads.
- Kept `RootView` borrowed and documented the distinction between its borrowed root and the fork's owned graph. Kept `Graph` as a value handle with its existing `Name` method.

## Checks run

- `rtk gofmt -w graph.go graph_test.go`
- `GOCACHE=/private/tmp/go-values-snapshot-gocache rtk go test ./...` — 3 tests passed in 1 package.
- `GOCACHE=/private/tmp/go-values-snapshot-gocache rtk go vet ./...` — no issues found.
- Test coverage exercises independent edits in both directions, shared-node identity, cycles, nil links, nil versus empty collections, and zero-Graph behavior.

## Limits

- The module declares Go 1.22; checks used the Go toolchain installed in this environment, and a separate Go 1.22 toolchain run was not performed.
- No concurrency behavior is promised or tested, consistent with the task contract.
