# Standalone indexer implementation

Implemented the README contracts in `/private/tmp/go-fresh-author-_5871z_7/task-2/module`. No dependencies, packages, public signatures or Go minimum were changed. No delegation, installation, commit, or external changes were performed.

## Changes and decisions

- `index.go`: ApplyCSV validates exactly two fields, lowercase ASCII alphabetic keys and nonblank text, writes rows in order, preserves duplicate replacement and accepted-prefix semantics, and wraps validation with ErrInvalidRecord identity. Put still accepts arbitrary text, including empty strings. Put and Refresh share a private same-directory temporary-file/close/rename publisher that preserves the destination when writing or renaming fails. Existing ErrNotImplemented remains exported for compatibility but is no longer returned by these operations.
- `refresh.go`: one context-bound GET through the supplied client's transport/timeout/cookie configuration, without following redirects or mutating caller redirect policy. Only status 200 and one non-null JSON array are accepted. Read/decode/validation completes before publication. Ordered records and original text are published as compact lowercase-field JSON plus newline, including `[]\n`. Every acquired successful response body is closed on success and rejection.
- `serve.go`: synchronous immediate/sequential callbacks, completion-based intervals, already-canceled-context handling, positive interval validation before callbacks/release, and exactly one release after callback completion. Callback and release errors are joined. Pure parent cancellation, including wrapped cancellation, returns nil; a joined independent failure during cancellation remains observable. Error classification does not compare arbitrary error values.
- `cmd/indexer/main.go`: existing positional put grammar remains intact, including leading-dash directory arguments. Apply reads stdin. Watch validates HTTP/HTTPS host, nonempty file and positive duration, recurs through Serve, and handles SIGINT/SIGTERM through cancellation. Its client has a finite request timeout and closes idle connections after Serve joins callbacks. All successful commands remain silent; failures/usage emit stderr and exit 2.

The existing library owns storage, record validation, refresh and lifecycle behavior; the existing command owns parsing, signals and resource wiring. The task provides no reason to split packages, introduce new exported interfaces or build a clock/injection framework.

## Regression evidence

New files: `csv_test.go`, `refresh_test.go`, `serve_test.go`, `retention_unix_test.go`, and `cmd/indexer/process_test.go`. Existing put regression tests were retained.

Tests observe exact text and independently specified JSON bytes, order/duplicates/empty arrays, open-reader old snapshots versus newly opened snapshots, invalid validation identity, malformed/trailing JSON, valid-body status 206 rejection, accepted CSV prefixes and forbidden later writes, retained duplicate values, source/body errors, body closure, fetch/publication failures and removal of temporary files. Representative function assignments protect consumer signature shapes.

Actual local HTTP fixtures verify redirect rejection with exactly one request and caller policy preservation, plus in-flight context cancellation. Actual built command processes verify status/stdout/stderr, legacy put grammar and arbitrary/empty text, CSV success and parse/validation/publication prefixes, invalid watch startup, failed first fetch/decode/validation/publication, successful recurrence, SIGTERM exit, and SIGINT cancellation of an in-flight request.

Child-local POSIX RLIMIT_FSIZE tests force Put, duplicate CSV replacement and Refresh writes beyond 1024 bytes. The tests require the resulting EFBIG/short-write identity, exact prior/accepted bytes, no later writes and no temporary files. An unrestricted large-text Put regression confirms arbitrary text remains accepted. File limits and SIGXFSZ policy are changed only in bounded child processes.

Lifecycle tests cover positive/invalid/already-canceled startup, callback failure alone, release failure alone, both failures, context identity, held cooperative cleanup before release/return, wrapped ordinary cancellation, cancellation joined with independent failure, noncomparable error values, successful later recurrence measured from first completion, absence of overlap while the first callback is held, and independently running invocations. Test fixtures cancel/unblock/join owned work with finite waits; recurrence snapshots use handler gates so scheduling cannot skip the observations.

## Actual checks

`checks.json` records command arrays, cwd, explicit non-secret Go environment, stdout, stderr, exit codes, deadlines and exercised execution boundaries.

- New focused success/error-combination tests compiled and failed against the original ErrNotImplemented stubs, establishing a meaningful pre-implementation signal.
- Initial `go test -timeout=45s ./...` failed because the sandbox rejected real httptest listeners with `bind: operation not permitted`. That result was preserved. The exact same command and unchanged source then passed with approved `require_escalated` loopback access; the rerun links back to the failed check.
- `go test -race -count=3 -shuffle=on -timeout=60s ./...` passed after fixture-control refinements.
- Focused partial-write, cancellation-cleanup and large-text tests passed. Independently selected `TestApplyCSVRejectedPrefix/blank_replacement` and actual-command `TestCLIProcessContracts/apply_prefix_after_validation` passed.
- Final `go test -race -timeout=60s ./...` passed after the additional partial-write error assertion and first-publication process case. Final `go vet ./...` passed. `gofmt -l .` returned no paths.
- The toolchain was `go1.26.5 darwin/arm64`, with `GOTOOLCHAIN=local` and the required `GOCACHE=/private/tmp/go-quality-testing-cache`. The module remains `go 1.22` and uses APIs available by Go1.22. Staticcheck executable lookup returned no path, so it was not run or installed.

The initial authorized dispatch read used raw `cat` before its RTK requirement was known. Every subsequent shell command used `rtk proxy`.

## Limits and remaining risks

The atomic replacement/snapshot and partial-write claims were exercised on the local POSIX macOS filesystem. Windows replacement semantics are outside this task's claim. Go1.22 itself was not installed or executed; compatibility evidence is the unchanged module directive, Go1.22-era APIs, consumer compilation and vet under the available local toolchain. Real HTTP evidence is loopback HTTP with approved listeners, not external DNS/TLS or deployed-service integration. Race/repetition results establish the exercised schedules, not every possible schedule. The command processes are actually executed but are built without race instrumentation; race instrumentation covers the library/test packages.

Refresh retains the full body, decoded records and compact output in memory because no payload limit is specified. Atomic rename protects publication visibility and failed writes; crash/power-loss durability is not claimed. Serve requires started callbacks to finish cooperatively before it can release resources, consistent with the requested join contract.

All requested implementation and regression work is complete. Optional guidance selection and opened paths are recorded in `selection.json`.
