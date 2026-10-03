Implemented bounded lease admission and cohort cleanup in [host.go](/private/tmp/go-task-qpozp8ug/host.go). `Serve` now starts jobs concurrently up to `limit`, stops admission on failure or cancellation, joins the cohort before closing leases, closes each acquired lease once, and retains independent errors. The public API and Go 1.22 minimum are unchanged.

Added controlled tests in [host_test.go](/private/tmp/go-task-qpozp8ug/host_test.go) for capacity, run-before-close ordering, open/run/close failures, cancellation, and lease release.

Skills used: `go-api-contracts`, `go-behavior-tests`, `go-concurrency-and-ownership`, and `go-test-isolation`.

Checks passed: `go test ./...`, `go test -race ./...`, `go test -count=25 ./...`, and `go vet ./...`. No external integration checks were performed.