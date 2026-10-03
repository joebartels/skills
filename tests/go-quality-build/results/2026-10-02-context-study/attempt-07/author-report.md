Updated [process.go](/private/tmp/go-task-tx9nlww3/process.go) to check cancellation before callbacks and between jobs, preserve accepted counts, and join cancellation classification and cause with callback errors when cancellation is already visible. The public signature and Go 1.22 module version remain unchanged.

Added tests in [process_test.go](/private/tmp/go-task-tx9nlww3/process_test.go) for canceled and empty input, cancellation between jobs, callback failure during cancellation, and successful final completion during cancellation. The cause test uses a non-comparable error and inspects it with `errors.As`.

Skills used: `go-api-contracts`, `go-behavior-tests`, and `go-test-isolation`.

Checks: `rtk gofmt` and both `rtk go test ./...` and the focused test run passed (7 tests). `rtk git diff --check` could not run because this directory is not a Git repository. Tests ran with Go 1.26.5; Go 1.22 compatibility is maintained by using APIs available in Go 1.22.