Updated [process.go](/private/tmp/go-task-mvam02bo/process.go) to check cancellation before each callback, preserve the accepted count, and retain callback errors alongside cancellation classification and cause. A successful final callback remains successful even if it cancels the context.

Expanded [process_test.go](/private/tmp/go-task-mvam02bo/process_test.go) to cover pre-cancellation, empty input, between-job cancellation, coincident completion, and combined callback and cancellation errors—including a non-comparable custom cause. The `Process` signature, Go 1.22 support, and dependency set are unchanged.

Skills opened and used: `go-api-contracts` and `go-behavior-tests`.

Checks: `rtk go test ./...` passed; `rtk go vet ./...` passed. No external consumer integration was tested.