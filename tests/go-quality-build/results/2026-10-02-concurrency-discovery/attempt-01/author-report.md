Updated [totals.go](/private/tmp/go-task-i5uwvnep/totals.go) so one mutex protects both fields during `Add` and `Snapshot`, preserving the public API and zero-value behavior. Expanded [totals_test.go](/private/tmp/go-task-i5uwvnep/totals_test.go) with concurrent snapshot coherence, final totals, and instance independence checks.

Skills used: `go-api-contracts`, `go-behavior-tests`, and `go-test-isolation`.

Checks passed: `go test -race ./...`, `go test -race -count=20 ./...`, and `go vet ./...`. `git diff --check` could not run because this directory is not a Git repository.