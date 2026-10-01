Opened the Go interface/composition skill: no. The task changes only a private pure calculation, which its routing description excludes.

Changed `batchCount` to add one when division has a remainder, preserving zero-item behavior and avoiding overflow. Added partial-batch regressions and a largest-int case. Kept the helper private and its positive-size precondition.

Checks: `rtk go test ./...` passed (2 tests); `rtk go vet ./...` reported no issues. Both ran with a writable `GOCACHE`.
