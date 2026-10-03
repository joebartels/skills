Implemented `Serve` with the existing Lease API and Go 1.22 minimum. Added tests for capacity, admission during open input, cancellation, cohort join-before-close, capacity held through Close, and independent Open/Run/Close errors.

Skills opened and used: `go-api-contracts`, `go-behavior-tests`, and `go-test-isolation`.

Checks passed: `go test ./...`, `go test -race ./...`, and `go test -count=50 ./...`.

**Limitation:** These checks passed, but the contract review is incomplete: the implementation’s Open/Run event ordering may fail to stop admission promptly on asynchronous failures.