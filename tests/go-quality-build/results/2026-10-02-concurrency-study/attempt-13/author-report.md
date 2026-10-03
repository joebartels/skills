Updated [totals.go](/private/tmp/go-task-b0p0xjxk/totals.go) to guard each `Add` and `Snapshot` with a mutex. The zero value and public API remain unchanged.

Added tests for concurrent updates and snapshots, independent snapshot values, and isolation between `Totals` instances.

Skills used: `go-api-contracts`, `go-behavior-tests`, and `go-test-isolation`.

Checks: `go test ./...` and `go test -race -count=20 ./...` passed.