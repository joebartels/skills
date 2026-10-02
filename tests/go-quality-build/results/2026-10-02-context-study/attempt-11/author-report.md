Implemented the shared total budget and per-stage timeout in [stages.go](/private/tmp/go-task-myuxtdyr/stages.go). Each stage keeps its context through body reading and closure; failures preserve the completed body prefix and join any observed cancellation classification and cause. The `FetchAll` signature and Go 1.22 minimum are unchanged.

Added tests for sequential success, response-body lifetime and closure, non-200/read/close failures, shared deadlines bounded by the parent, already-canceled calls, and cancellation during a read failure.

**Skills used:** go-api-contracts, go-behavior-tests, go-context-and-deadlines, go-test-isolation.

**Checks:** `rtk gofmt -w stages.go stages_test.go`, `rtk go test ./...`, and `rtk go test -race ./...` passed. The HTTP transport/body doubles verify the exercised client boundary; they do not establish that arbitrary remote work stops on cancellation.