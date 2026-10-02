# Trial report

## Selected and opened skills

- Selected and read `catalog/go-api-contracts/SKILL.md` because `CollectMatching` adds an exported function and its result/error behavior is part of the consumer contract.
- Selected and read `catalog/go-error-contracts/SKILL.md` because collection returns partial rows alongside Scan or iteration failures and owns cursor cleanup.
- Did not select or open `go-interfaces-and-composition` or `go-package-boundaries`: the change reuses the existing `Rows` interface and package responsibility.

## Artifacts

- Updated `source/collect.go`: added documented `CollectMatching`; kept `Collect` and delegated it with a nil predicate. Both return a valid prefix on Scan or terminal iteration errors and close once. Close errors are ignored as best-effort cleanup.
- Updated `source/collect_test.go`: covered existing collection, filtering/order, nil predicate, partial prefix on Scan and iteration errors, and one close call even when close returns an error.
- Added `source/api_external_test.go`: external-package compile checks for both exported function types and a consumer implementation of `Rows`.
- Updated `source/README.md` with filtering and failure/cleanup behavior.

## Decisions and evidence

The README defines the owned cursor, valid-prefix failure behavior, and best-effort close contract. Filtering occurs after successful Scan, preserves order, and cannot fail. `Collect` remains available with its original signature and now follows the documented terminal `Err` behavior through the shared implementation. Close errors do not replace successful results or primary Scan/iteration errors.

## Checks

- `rtk gofmt -w collect.go collect_test.go`
- `rtk env GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache go test ./...` — passed (`ok example.com/cursorload`).
- `rtk gofmt -w api_external_test.go`
- `rtk env GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache go test ./...` — passed (`ok example.com/cursorload`), including external API compilation.
- `rtk git diff --check` — passed.

## Limits

No downstream module or release policy was supplied, so compatibility evidence is limited to the unchanged `Collect` signature and an external-package compile check. No database integration was attempted; the package has no database dependency.
