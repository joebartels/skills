Updated `Serve` to enforce capacity through lease closure, stop admission on observed cancellation or failure, join started Runs before closing any lease, and retain independent Open, Run, and Close errors. It also closes acquired leases whose Runs never started and leaves the caller’s jobs channel open.

Added tests for concurrent capacity, the Run/Close lifecycle boundary, failure retention, cleanup of acquired leases, and cancellation during Open.

**Skills used:** `go-api-contracts`, `go-behavior-tests`, and `go-test-isolation`.

**Checks:** `rtk go test ./...` and `rtk go test -race ./...` passed. No external integration checks were run.