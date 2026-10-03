Updated [Serve](/private/tmp/go-task-vdyrcwyz/host.go) to admit up to `limit` leases concurrently, stop admission on failure or cancellation, wait for every started Run, and then close every acquired lease. It now retains independent Open, Run, Close, and caller-cancellation errors through `errors.Is`.

Added [tests](/private/tmp/go-task-vdyrcwyz/host_test.go) for capacity, admission, completion and release order, acquired leases whose Run never starts, retained successful effects, and failure combinations. The existing Lease API, standard-library-only dependency boundary, and Go 1.22 module version remain intact.

Skills opened and used: `go-api-contracts`, `go-behavior-tests`, `go-interfaces-and-composition`, and `go-test-isolation`. Checks passed: focused tests, `go test -race -shuffle=on -count=30 ./...`, and `go vet ./...`.

Limitation: checks ran with Go 1.26.5; a Go 1.22 toolchain was not available for a direct version check.