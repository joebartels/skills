Added blocking `Refresher.RunPeriodic(ctx, interval) error`. It performs an immediate refresh, waits only after successful completion, rejects nonpositive intervals, and returns the callback's first error. Context cancellation reaches the callback and returns `ctx.Err()` after callback cleanup; callers own and join each run. Kept `New` and `Refresh` unchanged. No interface or new dependency was needed because the host supplies the callback and context expresses run ownership.

Updated `README.md` with API behavior and a two-run shutdown recipe that cancels both runs before collecting both results. Added an executable Go example and tests for active callback shutdown waiting, instance independence, and callback error propagation.

Checks run in the clean module copy:
- `rtk gofmt -w refresh.go refresh_test.go`
- `GOCACHE=/private/tmp/go-quality-build-composition-eval/gocache rtk go test ./...` — 6 passed in 1 package.
- `GOCACHE=/private/tmp/go-quality-build-composition-eval/gocache rtk go vet ./...` — no issues found.
