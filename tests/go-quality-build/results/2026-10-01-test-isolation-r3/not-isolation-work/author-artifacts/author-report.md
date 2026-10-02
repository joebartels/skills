# Clamp repair

Changed `clamp.go` to return `low` when `value < low`. Values exactly at the lower bound flow through unchanged. The exported `Clamp(value, low, high int) int` signature, package, module path, and `go 1.22` directive remain unchanged.

Added the ordinary external-package regression `TestLowerBound` in `clamp_test.go`, covering one value below the lower bound and one exactly at it. The existing interior test remains. The new regression failed before the repair for both values (returned 2 instead of 1), then passed after it.

Checks are preserved in `checks.json` with argument arrays, module cwd, explicitly set non-secret environment, stdout/stderr, exit codes, and finite process deadlines. All executed verification commands use `rtk proxy`, `GOCACHE=/private/tmp/go-quality-testing-cache`, and `GOTOOLCHAIN=local`.

- Before-fix focused regression: expected failure, exit 1.
- `gofmt -s -w .`: passed; `gofmt -l .` returned no files.
- `go test -v -race -count=1 -timeout=20s ./...`: both tests passed, exit 0.
- `go vet ./...`: passed, exit 0.
- `go version`: installed toolchain is Go 1.26.5 on darwin/arm64.
- `staticcheck ./...`: unavailable, exit 1 (`No such file or directory`). No tool was installed.

Guidance selection and actual opened paths are recorded in `selection.json`. Applied Go core style and API contract guidance. Skipped composition, package-boundary, and test-isolation guidance for the recorded relevance reasons.

Limitations and remaining risks: Go 1.22 itself was not executed; the declared minimum is preserved and the change uses ordinary syntax supported by it. Staticcheck could not run. No downstream repositories were supplied; verification exercises the exported function through this module's external test package, not a real downstream integration. Invalid intervals (`low > high`) remain outside the README's caller contract. This pure calculation requires no listeners or external runtime boundary. No dependencies, tools, commits, or external changes were made.
