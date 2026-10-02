# Language-contracts combined trial

## Selected and opened skills

- `go-api-contracts/SKILL.md`: the exported `Select` API and CLI grammar, JSON contract, positional dash-prefixed paths, and error/partial-result compatibility are in scope.
- `go-error-contracts/SKILL.md`: `Load` now consumes bytes returned with read errors, preserves the usable prefix, and exposes the caller's cause; the CLI publishes the prefix before reporting incomplete input.
- `go-values-and-zero-values/SKILL.md`: `Select` promises nilness behavior and independent ownership of nested maps and byte slices.
- Skipped `go-interfaces-and-composition` because the change adds no interface, dependency, constructor, or lifecycle design.
- Skipped `go-package-boundaries` because the package responsibilities and import direction do not change.

Only the three selected skill files, the assigned source README/code/tests, the task prompt, the assigned catalog, and the canonical build-skills design record were read. No other trial, fixture expectation, review skill, or result was consulted.

## Decisions and implementation

`Load` keeps its existing signature and `Item` JSON fields. It ignores blank and `#` comment lines, parses all returned bytes before interpreting a non-EOF read failure, and returns the parsed prefix. Reader causes remain discoverable with `errors.Is`. When malformed input and a reader error arrive together, both are preserved with `errors.Join`; parsing stops at the malformed line.

`Select` preserves order, nil input, non-nil empty output, and nested nil-versus-empty byte slices. It copies each selected item’s tag map and each non-nil tag byte slice, so mutations in either direction remain independent.

The CLI accepts two or three positional arguments without flag parsing, retaining paths that start with `-`. It applies `Select` when a prefix is present, writes the selected prefix even after a load error, reports that error to stderr, and returns the specified status. README and API comments document partial reads, ownership, and CLI behavior.

## Checks

- `GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache go test -race ./...` — passed for the library and CLI packages. CLI integration tests verify dash-prefixed paths, prefix selection, successful empty streams/status, malformed-input prefix publication/status, usage status, and output failure status.
- `GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache go vet ./...` — passed.
- `gofmt` on all changed Go files — completed.
- `git diff --check` — passed.
- Test toolchain: Go 1.26.5, darwin/arm64. The module declares Go 1.22.

## Limits

The Go 1.22 toolchain was not available for execution. Reader-error propagation is covered at the library boundary using a reader that returns bytes and an error together; the CLI integration suite verifies malformed-input publication, while a filesystem-generated CLI read error was not exercised. Windows path and publication behavior were not tested. No external downstream consumer or concurrency-safety guarantee was tested or introduced.
