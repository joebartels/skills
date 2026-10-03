Updated `Process` to check cancellation before starting work and between jobs. It returns the accepted count on every exit, preserves callback errors alongside cancellation classification and cause, and lets a successful final callback complete the operation even if cancellation coincides with it. The public signature and Go 1.22 support are unchanged.

Added tests for already-canceled and empty calls, cancellation between jobs, a callback failure with a non-comparable custom cause, and successful completion concurrent with cancellation.

Skills opened and used: `go-context-and-deadlines`, `go-behavior-tests`, and `go-api-contracts`.

Check run: `rtk go test ./...` passed (7 tests, 1 package). The workspace has no Git metadata, so I could not report a Git diff or status.