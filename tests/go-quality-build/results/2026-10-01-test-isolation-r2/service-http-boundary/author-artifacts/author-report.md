# Fetch tests

Completed the requested test work in `/private/tmp/go-fresh-author-txf1lez2/task-3/module/fetch_test.go`. `fetch.go`, the exported API, standard-library-only dependencies and `go.mod` (Go 1.22 minimum) are unchanged. No dependencies/tools were installed, no delegation occurred, and no commit or external service changes were made.

## Tests and exercised boundaries

- `TestFetchResponse` has 25 independently selectable parallel cases: valid lowercase JSON fields, preserved nonblank name bytes, surrounding JSON whitespace, exact-200 status acceptance, malformed/empty/null/array data, trailing objects/garbage, absent/invalid fields, nonpositive and noninteger revisions, and read failures before data and after complete-looking JSON. Each fixture owns its client and instrumented body. Every acquired body must close once; failures must return zero metadata and a useful error, with status details and underlying injected read errors checked where applicable. These handwritten transports verify response interpretation and body ownership; they do not claim to exercise sockets, HTTP framing or real transport cancellation.
- `TestFetchTransportError` checks supplied-client use, caller-context propagation, zero metadata and the underlying transport failure. `TestFetchInvalidURL` checks construction failures do not reach the transport.
- `TestFetchHTTPRequest` uses a real local `httptest` server and HTTP transport to verify GET, escaped path/query preservation and complete successful response decoding. It preserves tabs/newlines/non-ASCII name bytes.
- `TestFetchRejectedRedirect` uses the actual supplied client's `http.ErrUseLastResponse` redirect policy. It checks the original 302 is rejected and the target receives no request. The policy is set on this test's owned client; no shared default configuration is mutated.
- `TestFetchCancellation/before_headers` and `/while_reading_response` use real local sockets. The server explicitly signals request arrival; the body case flushes an incomplete response and instruments the real body's first Read. Cancellation occurs after these events, with `errors.Is(err, context.Canceled)`, zero metadata, server request cancellation/completion and canceled-body closure checked. All underlying body reads/closes delegate to the real transport.

Cancellation cleanup is registered before the worker starts. It releases the server gate, cancels the caller, and joins Fetch and entered handlers before closing transport/server resources. Worker outcomes use buffered channels and synchronized close counters; assertions run in the test goroutine. Cases own independent state and fixtures, including when a single named child is selected. No sleeps guess readiness. Ten-second event bounds diagnose stalled cooperation; Go test and process deadlines provide finite execution bounds rather than claiming timers can terminate a goroutine.

## Verification

Exact command arrays, working directory, explicit non-secret environment, source fingerprints, stdout, stderr, exit codes, timeouts and durations are saved in `checks.json`. Verification commands use `rtk proxy`, `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`.

- Initial sandbox `go test -count=1 -timeout=60s ./...`: failed because the sandbox denied the `httptest` loopback bind (`operation not permitted`). The failure and stack trace are preserved.
- Approved unchanged-source rerun of that full test command with loopback listener access: passed. Recorded source fingerprints match the initial restricted run.
- Removed an unnecessary assertion about reading rejected-status bodies, so tests do not prescribe an undocumented drain policy; subsequent checks exercise the final source.
- `go vet ./...`: passed.
- `gofmt -l fetch_test.go`: passed with no output after formatting.
- Sandbox non-network suite: `go test -race -run '^TestFetch(Response|TransportError|InvalidURL)$' -count=20 -shuffle=5031 -timeout=60s ./...`: passed.
- Approved final full suite with local sockets: `go test -race -count=20 -shuffle=73423 -timeout=90s ./...`: passed.
- Both cancellation named children independently: `go test -race -run '^TestFetchCancellation/while_reading_response$' -count=20 -timeout=60s ./...` and the corresponding `before_headers` selection: passed with approved local sockets.
- Independently selected `TestFetchResponse/read_failure_after_complete_json`: 20 race-enabled repetitions passed in the sandbox.
- Toolchain check: Go 1.26.5, darwin/arm64. `staticcheck` is unavailable; it was not installed or run.

## Limitations and remaining risks

Real local HTTP request/redirect/cancellation behavior was verified only in the approved listener-enabled runs. The restricted sandbox cannot execute these server tests; no mocked substitute or skip hides that restriction. The tests do not contact external services or verify external DNS, TLS, HTTP/2, other platforms or arbitrary noncooperative implementations. Repetition and race detection cover the exercised schedules without proving all schedules race-free or flake-free. The module minimum remains Go 1.22 and the tests use compatible APIs, but an actual Go 1.22 toolchain was not available for this run.

## Guidance and artifacts

`selection.json` lists every offered guidance name, actual opened skill/reference paths and relevance decisions. Only the supplied Go core-style and test-isolation guidance (including its isolation-patterns reference) was opened. Task inputs were restricted to the authorized dispatch and this module's README/source/module files. The dispatch bootstrap was read before its shell-prefix requirement was known; subsequent shell commands and all verification commands used `rtk`.

Artifacts: `selection.json`, `checks.json`, `author-report.md`, and the report-local `run-check.py` command recorder. There is no outstanding implementation work.
