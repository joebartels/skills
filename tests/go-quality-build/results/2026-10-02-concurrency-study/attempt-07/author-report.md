Updated [pipeline.go](/private/tmp/go-task-m4gyhytx/cmd/pipeline/pipeline.go) so early consumer completion cancels production, producer failure stops the consumer, and both callbacks are joined. Independent errors are preserved; only an exact coordinated context error is suppressed, while caller cancellation remains an error.

Added gated concurrency and error-propagation cases in [pipeline_test.go](/private/tmp/go-task-m4gyhytx/cmd/pipeline/pipeline_test.go).

Skills used: `go-behavior-tests` and `go-test-isolation`. Checks passed: `go test ./...` and `go test -race -count=20 ./cmd/pipeline`. A Git diff check was unavailable because this directory is not a Git repository. Arbitrarily blocking Reader calls remain outside the cancellation guarantee described in the README.