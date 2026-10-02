# Indexer evolution author report

Completed 2026-10-01 in `/private/tmp/go-combined-author-vsgga4k8/task-4/module`.

## Changes and contract decisions

- Implemented `ApplyCSV` with two-field CSV parsing, lowercase ASCII keys, nonblank text validation, ordered duplicate replacement, accepted-prefix persistence, and stopping before later effects. Shape/key/text validation preserves `ErrInvalidRecord`; CSV parser and I/O causes remain inspectable.
- Implemented `Refresh` with the caller's context and borrowed client configuration. Redirects are disabled on a per-call client copy to enforce the single-GET/status-200 contract without mutating the supplied client. It accepts exactly one non-null JSON array, rejects trailing data and read failures, validates every record, retains order and original text, closes each acquired response body, and publishes compact JSON plus newline. Empty arrays publish `[]\n`.
- Shared a private same-directory temporary-file/close/rename helper between `Put` and `Refresh`. Failed staging or publication preserves prior destination bytes and removes temporary files. `Put` continues accepting arbitrary text, including empty, whitespace-only and NUL-containing text.
- Implemented `Serve` synchronously: no callback overlap or detached work, immediate first invocation, a fresh interval after callback completion, caller-context forwarding, and release exactly once after all work/cleanup. Invalid intervals have no lifecycle effects. Callback and release errors are joined without losing identity. Parent cancellation without a callback error succeeds; callback errors returned during cancellation remain errors.
- Preserved all exported signatures, `Record` fields/tags, sentinels, module path, Go1.22 declaration, and existing `put DIR KEY TEXT` positional grammar. Added typed `apply` and `watch` flag parsing. The watch host owns signal cancellation and an HTTP transport, calls release after refresh stops, and treats a Refresh error caused solely by its canceled context as graceful shutdown. It handles SIGINT and SIGTERM.
- Kept existing package boundaries, concrete client dependencies, and callback functions; added no dependencies, public interfaces, options framework, global mutable test hooks, or production clock seams.

Production files changed: `index.go`, `refresh.go`, `serve.go`, `cmd/indexer/main.go`. Added regression files: `apply_test.go`, `refresh_test.go`, `serve_test.go`, `write_failure_unix_test.go`, `cmd/indexer/process_test.go`. Retained existing tests and extended `index_test.go`'s arbitrary-text regression.

## Verified observations

- Consumer-side exact function-type assignments compile. Two already-open roots with the same key retain independent values through replacement.
- CSV quoted commas/multiline text, no header skipping, duplicate order, empty input, shape/key/blank-text/parse/read failures, accepted prefix, retained rejected replacement, forbidden later effects, and publication obstruction.
- Independently specified JSON bytes, empty-array representation, complete old/new POSIX snapshots across an already-open reader, status/decode/trailing/read/validation/publication/transport rejection, response-body ownership, and refusal to follow redirects while preserving the caller's redirect policy.
- Real loopback HTTP success and cancellation while reading a partially delivered response body; prior destination bytes remain exact.
- Serve's invalid/already-canceled context paths, independent work-only/release-only/joint error combinations, callback errors during cancellation, recurrence measured after a deliberately held callback, no overlap, held cancellation cleanup before release/return, and two independently controlled live invocations.
- Actual built executable status/stdout/stderr and effects for old put syntax, empty and dash-prefixed put text, usage/rejection, stdin CSV imports, and failed imports after accepted rows. Actual watch processes reject invalid startup arguments and failed first refreshes, publish two successful cycles, then cancel and join an in-flight real HTTP request on both SIGINT and SIGTERM, exiting successfully with empty streams and retaining the second published snapshot.
- Child-local `RLIMIT_FSIZE` tests on this Darwin host exercise write failure after progress for Put, duplicate CSV replacement, and Refresh publication. They check prior/accepted state, absent later rows, and temporary-file cleanup. The limit and SIGXFSZ handling remain in the owned child.

## Actual checks

All verification commands use `rtk proxy`, `GOCACHE=/private/tmp/go-quality-testing-cache`, and `GOTOOLCHAIN=local`. Full command arrays, cwd, explicitly configured non-secret environment, stdout/stderr, exit codes, timings and execution contexts are retained in `checks.json`.

The original sandboxed `go test -timeout=60s ./...` exited 1 because local httptest listeners were denied (`bind: operation not permitted`). That failure is preserved. An approved **unchanged-source** rerun of that command passed both packages. Subsequent listener-enabled runs also passed after intentional implementation/test refinements.

Final restored source checks passed:

- `go vet ./...`.
- `go test -race -shuffle=on -count=3 -timeout=90s ./...`: library 11.171s; command tests 6.009s; no race reports.
- Five separately selected checks with `-timeout=30s`: CSV blank-text rejected replacement, child-limited CSV write failure, cancellation cleanup with an independent callback error, watch recurrence/in-flight termination (both signal children), and real apply-process validation-prefix failure. Each passed when selected alone.
- `gofmt -l` across all Go source/test files produced no output.
- `go version`: `go1.26.5 darwin/arm64`; `go list -m -json`: module `example.com/indexer`, `GoVersion: 1.22`.

A controlled mutation temporarily replaced atomic staging/rename with direct `os.WriteFile` to the live destination. The bounded partial-write suite compiled and failed meaningful retained-byte assertions in **all three** Put/CSV/Refresh cases. The original source was restored in `finally`; SHA-256 before and after restoration was `2f1c57e445822651c9824c2924891555fa868cc22902af0329ef83ba4ad67fbb`. The expected failing output is retained separately in `checks.json`. All final checks ran after restoration.

## Limitations and remaining risks

- Go1.22 support is retained by the module declaration and use of APIs available by Go1.22; the installed Go1.26.5 toolchain executed the checks. An actual Go1.22 toolchain run was unavailable and no tools were installed.
- `staticcheck` was not installed; its availability probe is recorded. Formatting, vet and race verification ran.
- Real HTTP checks cover loopback transport and cancellation; external DNS/TLS/service behavior was not exercised. Synthetic transports separately cover controlled status/read/body-lifetime failures.
- Filesystem replacement and process-signal execution claims are Darwin/POSIX. Windows replacement semantics and other hosts were not evaluated. The file-size-limit fixture is built for Darwin/Linux and ran on Darwin.
- Race/shuffle/repetition checks cover exercised schedules, not every possible schedule. Serve joins cooperative callbacks synchronously; it cannot forcibly stop an arbitrary callback that ignores its context.
- No crash/power-loss durability claim is made; atomic visible publication is the requested filesystem contract. No downstream consumer repositories were available; exact-function-type checks are representative consumer evidence.

No remaining implementation work identified for the stated README contracts. No delegation, dependency installation, commits, external application writes, design-record reads, controller probes, or other-author workspace inspection occurred.

## Reports

- `selection.json`: offered guidance, actual opened skill/reference paths, relevance decisions and authorized source inputs.
- `checks.json`: retained verification evidence, sandbox failure, approved unchanged-source rerun and expected mutation failure.
- `run-checks.py` and `mutation-check.py`: local reproducible verification orchestration, confined to this module/report.
