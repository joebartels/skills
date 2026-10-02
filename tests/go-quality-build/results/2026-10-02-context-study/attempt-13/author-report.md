Updated [`Run`](</private/tmp/go-task-f87ys_kc/run.go:10>) to check cancellation before each job, preserve the accepted prefix and cancellation cause, and finalize accepted work exactly once under a separate timeout that survives parent cancellation. It waits for the cooperative finalizer and combines processing and finalization errors. The exported signature, Go 1.22 requirement, and supplied command host are preserved.

Added library contract tests in [`run_test.go`](</private/tmp/go-task-f87ys_kc/run_test.go:10>) and real child-process tests in [`main_test.go`](</private/tmp/go-task-f87ys_kc/cmd/finalize/main_test.go:27>) for interrupt output, exit status, receipt contents, and receipt write failure. The interrupt test skips Windows, where child interrupt delivery is unsupported.

Skills opened and used: `go-api-contracts`, `go-behavior-tests`, and `go-test-isolation`. The interfaces/composition and package-boundary skills were not applicable.

Checks passed: `rtk go test ./...`, `rtk go test -race ./...`, the two CLI tests repeated three times, and `rtk go vet ./...`. Git diff checks were unavailable because this workspace has no Git metadata.