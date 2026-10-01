# Small library trial report

Implemented `Range.Intersect(other Range) (Range, bool)` in `span.go` next to the existing `Range.Contains` method. This keeps the operation on the public `Range` type in the existing `example.com/span` package, with no import path, symbol, signature, or zero-value changes. It returns the maximum start and minimum end only when the resulting half-open interval is nonempty. Empty and reversed struct-literal ranges return `Range{}, false`.

Changed files: `span.go`, `span_test.go`. The added external-package table test covers overlap, containment, touching, disjoint, empty, reversed operands, and zero value.

Commands and results:

- `rtk go test ./...` before implementation: expected build failure because `Intersect` was undefined.
- `rtk gofmt -w span.go span_test.go`: passed.
- `rtk go test -race ./...`: passed, 10 tests in 1 package.
- `rtk proxy go vet ./...`: passed (exit 0, no output). The filtered `rtk go vet ./...` printed “No issues found” but returned exit 1, so the raw proxy was used to confirm the Go command's result.
- `rtk git diff --check`: unavailable because this disposable copy is not a Git repository.
