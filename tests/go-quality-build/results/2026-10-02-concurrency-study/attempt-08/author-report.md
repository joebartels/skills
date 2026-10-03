Updated [pipeline.go](/private/tmp/go-task-bl8a3kmy/cmd/pipeline/pipeline.go) so `runPipeline` coordinates both callbacks, stops production on consumer completion, stops the consumer on producer failure, and joins both callbacks before returning. It suppresses only exact context errors returned after coordinated cancellation; independent failures and caller cancellation remain errors.

Expanded [pipeline_test.go](/private/tmp/go-task-bl8a3kmy/cmd/pipeline/pipeline_test.go) to cover early completion and producer cleanup, producer failure, cancellation error handling, caller cancellation, ordered output, `--take`, and malformed input.

**Skills used:** `go-concurrency-and-ownership`, `go-behavior-tests`, and `go-test-isolation`.

**Checks:** `rtk go test ./...`, `rtk go test -race ./cmd/pipeline`, and `rtk gofmt -d` all passed. A Git diff check was unavailable because this directory is not a Git repository. Arbitrary blocking `Reader` calls remain outside the README’s cancellation guarantee.