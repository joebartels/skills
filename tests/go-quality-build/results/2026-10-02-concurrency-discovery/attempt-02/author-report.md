Updated `Serve` to admit jobs in cohorts up to `limit`, run each cohort concurrently, join started Runs before releasing leases, and wait for all Closes before admitting another cohort. It now stops after failures or cancellation, closes acquired leases whose Run never starts, and retains Open, Run, Close, and cancellation errors with `errors.Is`. The `Lease` API and Go 1.22 minimum are unchanged.

Added tests for capacity and join-before-release, independent failure retention, and cancellation while waiting for jobs.

**Skills used:** `go-behavior-tests` and `go-test-isolation`; read both skills and their referenced guidance.

**Checks:** `rtk go test ./...` and `rtk go test -race ./...` passed. `git diff --check` could not run because this supplied directory has no Git repository metadata.