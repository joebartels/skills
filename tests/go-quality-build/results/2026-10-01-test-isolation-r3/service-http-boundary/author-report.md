# Fetch test implementation

Changed only `/private/tmp/go-fresh-author-8hjrllxp/task-3/module/fetch_test.go` in the module. Production source, the public API, standard-library dependencies, and the Go 1.22 module minimum remain unchanged.

The added cases verify ordinary success, preservation of nonblank name bytes, surrounding JSON whitespace, caller context values, GET without a request body, and exact URL path/query preservation. Error cases cover non-200 statuses, malformed/empty/null/array/scalar JSON, additional JSON or garbage after an object, wrong field types, missing/blank names, and missing/nonpositive/null revisions. Every response case checks body closure and every failure checks zero Metadata. Read and transport errors preserve sentinel identity through `errors.Is`; a complete valid object followed by a read error must still fail. An invalid URL must fail before invoking the supplied transport.

## Exercised boundaries

| Fixture | Verified behavior | Limits |
| --- | --- | --- |
| Handwritten `http.RoundTripper` and tracked in-memory bodies | Supplied-client request construction and context values; returned status/data/read errors; response body closure | Does not exercise sockets, wire framing, DNS/TLS, or redirect following. The 302 case deliberately has no Location and verifies rejection of a returned 302. No production seam or shared default client is introduced. |
| Independent local `httptest.Server` and its client | Real GET and escaped path/query on the wire, decoded metadata | Uses loopback HTTP, without external service/DNS/TLS behavior. |
| Local server plus real client/transport, before headers | Caller cancellation interrupts an in-flight HTTP request, yields `errors.Is(err, context.Canceled)` and zero Metadata, and reaches the server request context | Startup is observed by a handler event; no timing sleep guesses readiness. |
| Local server plus a forwarding wrapper around the real response body | Caller cancellation after actual body reading starts interrupts real stream I/O, propagates context cancellation to the handler, returns zero Metadata, and closes the acquired body | Wrapper instruments Read/Close while forwarding actual network I/O. It does not replace the network dependency. |

Independent parallel cases own their clients and state. HTTP fixtures register cleanup at acquisition. The cancellation test installs cleanup before starting its Fetch worker; cleanup cancels, releases blocked handlers, and bounds the worker join before server shutdown. Buffered, nonblocking handler notifications cannot strand handlers on extra requests. Worker outcomes are asserted by the test goroutine. Owned timers are stopped. Event/join and client safety bounds are five seconds; these diagnose stalls and are not performance assertions. Test processes also have explicit finite deadlines.

## Actual verification

All commands use `rtk proxy`, with `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. The checked compiler was `go1.26.5 darwin/arm64`; a Go 1.22 toolchain was not installed or executed. Tests use APIs available by Go 1.22, including channels, contexts, timers, `http.NewResponseController`, and ordinary `t.Cleanup`; they do not use `t.Context` or `testing/synctest`.

- Original baseline: `go test -timeout=30s ./...` passed.
- Formatting: `gofmt -w fetch_test.go`, followed by `gofmt -l fetch_test.go` with empty output.
- `go vet ./...` passed.
- The first expanded `go test -timeout=45s -v ./...` failed because the sandbox forbids localhost listener binds. Its complete output and exit code are retained separately from the approved unchanged-source rerun, which passed.
- After final cleanup refinements, the full suite passed `go test -race -shuffle=on -count=20 -timeout=60s ./...` with approved local listening.
- Independently selected `TestFetchSuccess/preserved_name` and `TestFetchResponseErrors/trailing_object` each passed with `-race -count=1 -timeout=30s` in the default sandbox.
- Independently selected `TestFetchCancellation/during_body` and `TestFetchCancellation/before_headers` each passed with `-race -count=1 -timeout=30s` and approved local listening.
- `which staticcheck` returned exit 1: staticcheck is unavailable and was not installed.

`checks.json` records command arrays, cwd, explicit non-secret environment overrides, captured stdout/stderr, exit codes, and a 90-second subprocess timeout. Local-listener approvals are distinguished in its execution-profile fields. Repeated passing tests exercise meaningful schedules but do not prove every schedule race-free or flake-free.

## Remaining limitations and risks

No known production defect was found or changed. Normal execution of the real HTTP tests requires permission to open loopback listeners; they intentionally fail under a blanket listener restriction instead of silently skipping or replacing that boundary. External network services, TLS, caller-specific redirect policies, cross-platform behavior, and execution under an actual Go 1.22 toolchain remain unverified. No dependencies/tools were installed, no commits were made, and no external application changes were performed.

Guidance choices and actual inspected paths are recorded in `selection.json`. The only opened guidance was `go-core-style`, `go-test-isolation`, and that skill's required isolation-patterns reference. No unavailable testing guidance, repository design/evaluation files, controller probes, or other author workspaces were inspected.
