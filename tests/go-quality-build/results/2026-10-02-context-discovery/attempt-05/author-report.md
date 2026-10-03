Implemented the `FetchAll` budget, cancellation, and response-body lifetime behavior while preserving its signature and the Go 1.22 minimum. Tests cover deadline limits, canceled calls, ordered-prefix retention, exact status acceptance, body closure, and inspectable operation and cancellation causes.

**Skills used:** `go-behavior-tests`, `go-test-isolation`, `go-interfaces-and-composition`, `go-api-contracts`.

**Checks:** `rtk go test ./...` passed; `rtk go vet ./...` reported no issues. Changes are in [stages.go](/private/tmp/go-task-s9sehndx/stages.go:13) and [stages_test.go](/private/tmp/go-task-s9sehndx/stages_test.go:40).

The tests use a controlled HTTP transport and body doubles. They verify cancellation at that boundary, not whether arbitrary remote work stops after cancellation.