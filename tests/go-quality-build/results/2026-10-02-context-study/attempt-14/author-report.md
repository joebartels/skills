Implemented the bounded finalization contract in [run.go](/private/tmp/go-task-na7nrrsa/run.go) and added `Run` and child-process coverage.

`Run` now stops before starting work when cancellation is observed, preserves the accepted count and cancellation cause, and finalizes accepted work exactly once under a detached context with the supplied deadline. It waits for finalization to return and retains processing and finalization failures together. The supplied command host and `go.mod` version remain unchanged.

Tests cover cancellation and finalization, late cancellation after successful work, combined failures, command output and receipt, and a real receipt-write failure. The interrupt test skips Windows because `os.Interrupt` cannot be delivered to child processes there.

**Skills used:** `go-context-and-deadlines`, `go-behavior-tests`, and `go-test-isolation`, including their relevant references.

**Checks:** `rtk go test ./...`, `rtk go test -race ./...`, and `rtk go vet ./...` passed. The supplied module targets Go 1.22; checks ran with Go 1.26.5 on Darwin arm64.