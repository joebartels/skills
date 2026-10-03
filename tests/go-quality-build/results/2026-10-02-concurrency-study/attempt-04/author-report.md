Updated [totals.go](/private/tmp/go-task-f9t_0m5b/totals.go) to protect `Count` and `Sum` with one mutex, making each `Add` atomic as an update and each `Snapshot` coherent. The zero value and public API remain intact; no dependencies were added.

Added tests for concurrent updates and snapshots, final totals, and independence between instances.

Skills used: `go-api-contracts`, `go-behavior-tests`, `go-concurrency-and-ownership`, and `go-test-isolation`.

Checks passed: `rtk gofmt -w totals.go totals_test.go`, `rtk go test ./...`, and `rtk go test -race ./...`. The race check exercises this workload; it cannot prove correctness for every possible schedule. The directory has no Git metadata, so I couldn’t report a Git diff.