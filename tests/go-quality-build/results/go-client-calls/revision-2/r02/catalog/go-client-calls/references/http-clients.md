# HTTP clients

Read the supplied `http.Client`, transport and call contract. Reuse the client's
connection pool; configure owned clients at construction. Select timeouts and
pool/response limits from the workload. Preserve TLS verification, proxy,
authentication, cookie jar and redirect policies. Clone a suitable owned
`http.Transport` when independent transport configuration is needed; arbitrary
RoundTrippers and replaced global defaults need not be `*http.Transport`.
Avoid changing a borrowed client's fields during concurrent use.

## Requests, redirects and retry ownership

Construct requests with the operation context and keep its scope alive through
owned body consumption/closure. For each permitted application attempt, rebuild
the request/body from stable operation data. `bytes.Reader`, `strings.Reader`
and `bytes.Buffer` request constructors can supply `GetBody`; arbitrary streams
need an explicit replay strategy. Replayability is a mechanism, not semantic
permission to repeat a mutation.

Inspect the selected Go transport's network-error retry conditions. Its notion
of replay includes certain methods and idempotency header presence, subject to
body and connection conditions. Setting an idempotency marker can therefore
affect native retries even before an application loop is added. Use it only
under an appropriate operation contract. Count application `Do` calls separately
from transport or server executions; do not promise an unsupported wire ceiling.

`Client.Do` applies redirects. A final 200 can hide an initially rejected status,
different target or changed method/body. Decide whether the API accepts final
responses under the supplied redirect policy or requires inspecting the first
response. For a stricter per-operation policy, use an appropriate local client
configuration while preserving required host behavior and borrowed fields.
Check credentials and idempotency identity across redirects; extra targets may
lie outside the server's deduplication scope.

Treat non-success statuses according to the remote API, not every 2xx or 5xx
by habit. A `Do` error can reflect client policy or protocol/configuration
failure, not merely a transient network fault. With a redirect-policy error,
the returned response body is already closed; do not close it again. When `Do`
succeeds, you own body closure on every status/read outcome. `ErrUseLastResponse`
returns a response with an owned, unclosed body. `Client.Timeout` can mask
independent transport/body failures; verify required error identity through that
configured path instead of replacing failures with `ctx.Err()`.

If cause preservation requires private transport/body observation, preserve the
client's effective timeout and cancellation facilities. Zeroing `Client.Timeout`
and substituting an attempt context can lose supported behavior: custom
RoundTrippers may use the legacy `CancelRequest` hook rather than request context.
A transport decorator must forward the cancellation facilities its underlying
transport needs. Retained field values alone do not establish retained behavior.
Observe failures before timeout rewriting when their identities are promised;
verify the configured client boundary with an independent body cause and a
shorter client timeout on the minimum supported Go version. Check effective
cancellation as well as cause identity; neither justifies losing the other.
When `Do` returns a policy error and response, retain its known status if the
caller promises status-bearing errors, while respecting the already-closed body.
Choose initial versus final response acceptance from the operation contract;
do not override a required host redirect policy based on an unstated assumption.

When dependency health requires a completed response, classify after body
consumption. A success status with a read failure need not prove health. Check
that outcome against the defined exclusion/history/recovery policy, alongside
successful consumption during caller stopping; headers alone do not establish
the completed result.

## Retry hints and response bounds

When the operation supports Retry-After, parse a nonnegative integer number of
seconds or an HTTP-date. Distinguish malformed input from a syntactically valid
delay too large to represent or fit the remaining budget. Use checked arithmetic;
do not cast overflowing seconds into a negative/zero duration. A future valid
hint forbids an earlier retry. If it cannot fit, return the contract's last
rejection without restarting the budget. Define malformed and past-date handling.
Wait with caller cancellation, and release the previous owned body before retrying.

Bound consumption of success and diagnostic bodies. Detect oversized complete
results instead of silently returning a truncated representation; prefix/stream
results require an explicit caller contract. A cap plus one sentinel byte can
detect overflow, with checked cap arithmetic. Truncation can suit diagnostics.
An upstream parser, transport or reader error can render arbitrarily large text
even when response reads are bounded: if diagnostic text is capped, bound its
presentation separately from cause access. Avoid exposing sensitive endpoint,
credential or response details. Do not drain unlimited bodies merely for pool
reuse; choose bounded draining or closure according to the workload.

This Go 1.22-compatible example assumes exact final-200 success under the supplied
client policy, a complete bounded result and one application invocation. Its
limit applies to response content. The caller owns the context budget and client:

```go
package remote

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net/http"
)

func FetchOnce(ctx context.Context, client *http.Client, endpoint string, maxBytes int64) ([]byte, error) {
    if maxBytes < 0 || maxBytes == 1<<63-1 {
        return nil, errors.New("invalid response limit")
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
    if err != nil {
        return nil, err
    }
    resp, err := client.Do(req)
    if err != nil {
        return nil, err // An error response's redirect body is already closed.
    }
    data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
    closeErr := resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, errors.Join(fmt.Errorf("HTTP %d", resp.StatusCode), readErr, closeErr)
    }
    if int64(len(data)) > maxBytes {
        return nil, errors.Join(errors.New("response exceeds limit"), readErr, closeErr)
    }
    if err := errors.Join(readErr, closeErr); err != nil {
        return nil, err
    }
    return data, nil
}
```

Sources: [Go HTTP client](https://pkg.go.dev/net/http#Client),
[Transport](https://pkg.go.dev/net/http#Transport),
[request replay](https://pkg.go.dev/net/http#Request.GetBody), and
[HTTP retry guidance](https://www.rfc-editor.org/rfc/rfc9110.html#section-10.2.3).
