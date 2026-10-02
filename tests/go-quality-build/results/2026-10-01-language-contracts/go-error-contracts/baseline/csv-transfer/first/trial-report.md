# Trial report

Implemented `WriteSelected` in `source/csv.go`, added API behavior tests in
`source/csv_test.go`, and documented the contract in `source/README.md`.

## Skills

- Selected and opened `catalog/go-api-contracts/SKILL.md`, because this adds an
  exported function and specifies compatibility, output, error, and partial
  result behavior.
- Did not select interface/composition or package-boundary skills: no dependency
  abstraction, constructor, lifecycle, package, responsibility, or import
  direction changed.

## Decisions

- The API preserves `WriteRows` unchanged and adds the exact signature
  `WriteSelected(io.Writer, [][]string, func([]string) bool) (int, error)`.
- A nil predicate selects all rows. Selected rows reach the encoder in input
  order; the count advances when each row is submitted, including the row whose
  `Write` call reports an error. A later flush error returns the count of all
  rows submitted. Underlying errors are returned for inspection.
- The borrowed writer is not closed. The implementation adds no dependency and
  performs no retries.

## Checks

- `gofmt -w csv.go csv_test.go` — completed.
- `GOCACHE=/private/tmp/go-cache-csv-transfer-baseline-first GOWORK=off go test ./...` — passed.
- `GOCACHE=/private/tmp/go-cache-csv-transfer-baseline-first GOWORK=off go vet ./...` — passed.
- `git diff --check` — passed.
- Test coverage includes the exact exported function type, row filtering and
  order, nil predicate, failure count, and `errors.Is` on the sink error. The
  pre-existing `WriteRows` behavior test remains and passes.

## Limits

Checks used the installed Go toolchain and module-local package tests; no
downstream consumer or additional Go versions were available in this disposable
trial. The Go build cache was isolated under `/private/tmp`, outside `source`.
