Expanded `/private/tmp/go-testing-author-y6_tmx1p/service-http-boundary/fetch_test.go` with external-package tests for the existing Fetch contract. Production source, exported API, standard-library-only dependencies, and `go 1.22` remain unchanged.

The tests cover valid metadata, preserved nonblank name bytes, surrounding JSON whitespace, malformed/null/non-object JSON, trailing objects and garbage, missing/blank/mistyped fields, nonpositive/fractional/overflowing revisions, five non-200 statuses, malformed request URLs, transport errors, and a response read error delivered alongside valid metadata bytes. Failure cases require zero Metadata and a nonempty error. Sentinel transport/read errors and caller cancellation use `errors.Is`; HTTP status errors must identify their status. Acquired synthetic bodies must be closed exactly once before return on success, status failure, decode/field failure, and read failure.

Dependency fixture boundaries actually verified:

- The function-backed `http.RoundTripper` runs through the supplied real `http.Client`, replacing HTTP transport I/O with controlled responses/errors. It verifies Fetch response handling and use of the supplied dependency. It does not verify DNS, sockets, HTTP framing, or standard transport cancellation.
- `trackedBody` counts Close calls on controlled readers. This verifies response ownership at the Go body interface; it does not prove socket reuse or connection closure.
- `failingReader` supplies valid JSON bytes and a sentinel read failure in the same Read call. This verifies that a read failure wins over apparently usable partial metadata and that the body closes.
- A loopback `httptest.Server` and its standard `http.Transport` verify the actual GET request and preservation of escaped path bytes, repeated query keys, escaped spaces, and escaped plus signs. This reaches real local TCP and HTTP/1.1, without external services.
- Cancellation fixtures observe the server request context before headers and during an incomplete response body. The body-phase transport decorator delegates all I/O to the real transport, then observes a successful body Read and Close. Cancellation occurs only after that Read signal, so the body-phase test verifies acquired-response cancellation, zero Metadata, `errors.Is(err, context.Canceled)`, and closure before return. The server independently observes context cancellation in both phases.

Cancellation uses explicit channels rather than sleeps. The test owns its worker goroutine, requests cancellation during cleanup, releases a held handler, and joins the Fetch worker. Result channels are buffered so workers can finish after an assertion failure. Every observation wait has a five-second timeout; cleanup joins have six seconds; HTTP clients have five-second timeouts. Go test processes use `-timeout=30s` and a separate 60-second process deadline.

Checks actually run, with exact commands/stdout/stderr/exit codes in `checks.json`:

- `go version`: Go 1.26.5 on darwin/arm64.
- Baseline `go test -v -race -timeout=30s ./...`: passed the original single test.
- `gofmt -s -w fetch_test.go`: passed; final `gofmt -l .` output was empty.
- `go vet ./...`: passed before and after the final fixture cleanup changes.
- Initial new `go test -v -race -count=1 -timeout=30s ./...`: synthetic-response cases passed, then the sandbox rejected a local TCP bind with `operation not permitted`; HTTP/cancellation execution was blocked in that run.
- The same full verbose race suite with local socket access outside the restrictive sandbox: passed, including the request and both cancellation cases.
- Final `go test -race -shuffle=on -count=50 -timeout=30s ./...` with local socket access: passed all 50 repetitions.
- `which staticcheck`: exit 1. Staticcheck is unavailable and was not installed.

Every shell command after reading the dispatch used RTK. All Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. No dependencies or tools were installed; no commits or external application changes were made.

Execution restrictions and remaining risks: the complete suite requires local TCP listener permission; no tests silently skip when listener access is unavailable. The synthetic-response tests can run without socket access. The Go 1.22 module minimum and compatible APIs are preserved, but the available Go 1.26.5 toolchain performed verification; Go 1.22 itself was not installed or run. These fixtures do not verify remote DNS, proxies, TLS, HTTP/2, redirects, connection reuse, or Close-error policy. Those behaviors are outside this request's demonstrated boundaries. Staticcheck remains unverified.
