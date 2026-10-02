Updated [stages.go](/private/tmp/go-task-96qna09t/stages.go) to apply one caller-bounded total deadline and a per-stage deadline that remains active through body reading and closure. It retains completed bodies on later failure, closes non-200 bodies, and preserves independently inspectable work, close, and observed cancellation causes. The function signature and Go 1.22 minimum are unchanged.

Added tests for empty and already-canceled calls, total and stage deadlines, cancellation causes, completed-body success, body closure, and retained prefixes after read, close, and status failures.

**Skills used:** `go-api-contracts`, `go-behavior-tests`, `go-test-isolation`.

**Checks:** `rtk go test ./...`, focused behavior tests, and `rtk go vet ./...` passed. Git diff checks were unavailable because the workspace has no Git metadata. Cancellation behavior is verified at the exercised transport and response-body boundary.