# Fetch test implementation

Changed only `fetch_test.go`. Production code, the exported API, standard-library dependency set, and `go 1.22` remain unchanged.

## Behavior and fixture boundaries

- `TestFetchHTTP` uses a local `httptest.Server` and its real HTTP client/transport. It verifies GET, use of the supplied client, exact escaped path/query preservation, lowercase response JSON field decoding, revision decoding, acceptance of surrounding JSON whitespace, and preservation of nonblank name bytes.
- `TestFetchResponse` has 24 independent parallel cases. A handwritten RoundTripper and instrumented body verify success and close-on-return ownership, exact-200 status handling (201, 204, 304, 404, 500 rejected), a read failure after otherwise valid JSON, malformed/empty/null/non-object/trailing input, missing/blank fields, nonpositive revision, and field type errors. Every failure requires zero Metadata and a nonempty error; status/read failures require identifying detail. All acquired bodies must be closed exactly once. These fixtures verify the Fetch/client transport boundary rather than TCP framing or external services.
- `TestFetchTransportError` injects a transport failure and requires its identifying detail and zero Metadata. `TestFetchInvalidURL` verifies malformed-URL failure occurs before the transport.
- `TestFetchCancellation/before_headers` uses a real local server/transport, waits for server request startup, cancels the caller, joins Fetch, and observes server request-context cancellation. `reading_body` flushes successful headers but withholds the promised body; an instrumented real response body signals Read startup before cancellation. Both require zero Metadata and `errors.Is(err, context.Canceled)`; the acquired body is closed once in the body-read case.

Workers return results through channels. Assertions run in the test goroutine. Each cancellation case owns its server, client, cancel function, handler release gate, and worker completion event. Cleanup is registered before work starts, releases the gate and cancels/joins Fetch before the server closes, including early failures. Cases share no mutable fixtures or process-global configuration. Five-second event waits are generous stalled-work bounds, not sleep-based readiness guesses or contractual latency limits.

## Checks actually run

All command stdout, stderr, exit codes, execution scope, and finite deadlines are in `checks.json`.

- `go version`: Go 1.26.5, darwin/arm64.
- `gofmt -w fetch_test.go`, `gofmt -l fetch_test.go`, and `gofmt -s -l fetch_test.go`: passed; both listing checks returned no filenames.
- `go vet ./...`: passed.
- Initial sandbox `go test -v -race -timeout=30s ./...`: failed because sandbox policy denied httptest loopback binding; this failure is preserved.
- The same full verbose race run outside the sandbox: passed.
- Final full suite: `go test -race -shuffle=on -count=25 -timeout=30s ./...`: passed outside the sandbox.
- Selected response leaves (`read_error_after_valid_json`, `trailing_object`, `null`) with race detection, ten repetitions, and a 15-second test deadline: passed in the sandbox.
- Each cancellation child independently selected with race detection, ten repetitions, and a 20-second test deadline: passed outside the sandbox.
- HTTP success independently selected with race detection, ten repetitions, and a 20-second test deadline: passed outside the sandbox.
- `staticcheck` was not run: unavailable on PATH, and tool installation is prohibited.

Every Go command used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Verification processes had 10- or 60-second outer deadlines; timeout handling kills the process group. No command timed out. Commands were invoked through `rtk proxy`.

## Limitations and remaining behavior risks

Network tests require loopback bind permission; they currently fail in a network-restricted sandbox rather than silently skip or replace real transport behavior. Successful network verification used automatic escalation. Local plain HTTP does not verify external DNS, TLS, proxies, or remote-service behavior. Execution used Go 1.26.5; execution on Go 1.22 itself was not verified. The module minimum was preserved and the tests use Go 1.22-compatible APIs and loop semantics. Passing race/shuffle/repetition checks cover the exercised schedules and do not prove universal race freedom or absence of flakes. Body-close failures and arbitrary noncooperative dependency behavior are outside the documented Fetch contract; finite joins diagnose stalls but cannot forcibly stop a goroutine.
