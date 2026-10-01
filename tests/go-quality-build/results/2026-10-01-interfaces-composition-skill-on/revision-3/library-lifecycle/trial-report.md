Trial report — lifecycle support

Task
Add periodic refresh alongside one-shot Refresh, document the lifecycle API, provide an orderly two-run shutdown example, and cover active stopping, independent instances, and callback errors.

Choices
- Kept `New` and the concrete `Refresher`; the host-provided callback remains the only dependency and keeps ownership of its captured resources.
- Added `Refresher.Start(ctx, interval) (*Run, error)`. It rejects nil contexts and nonpositive intervals, invokes the first callback immediately, then waits the full interval after each successful completion. Callback calls within a run are sequential. A refresher rejects another Start while a run is active.
- `Run.Stop` requests cancellation and returns without joining. `Run.Wait` joins callback cleanup and returns its callback error unchanged, including `context.Canceled` when a stop cancels the callback. Parent-context cancellation follows the same callback-error behavior. The run stops on its first callback error.
- The host must stop every run and wait for every run before releasing resources used by those callbacks. The README example starts two runs, stops both, waits for both, then releases the shared resource.

Artifacts changed
- `refresh.go`: lifecycle implementation and API comments.
- `refresh_test.go`: tests for stopping during callback cleanup, independent instances (both observed recurring before one is stopped, and the other keeps recurring), first callback error propagation, and waiting the interval after callback completion.
- `example_test.go`: executable two-refresher stop/join/resource-release example.
- `README.md`: lifecycle semantics and runnable host shutdown example.

Checks
- `rtk gofmt -w refresh.go refresh_test.go example_test.go`
- Initial `GOCACHE=/private/tmp/go-quality-build-composition-eval/skill-on-r3/gocache rtk go test ./...` had two assertion failures because tests incorrectly expected nil from Wait after Stop. The implementation returned the documented callback `context.Canceled` unchanged. Updated those assertions.
- Final `GOCACHE=/private/tmp/go-quality-build-composition-eval/skill-on-r3/gocache rtk go test ./...`: passed, 7 tests/examples in 1 package.
- Final `GOCACHE=/private/tmp/go-quality-build-composition-eval/skill-on-r3/gocache rtk go vet ./...`: passed, no issues.

Open questions / next action
- None for this task. The module is a non-git clean copy; no repository changes were made outside the requested trial artifacts.
