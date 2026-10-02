# HTTP metadata test work

`Fetch(ctx context.Context, client *http.Client, endpoint string) (Metadata, error)`
uses the supplied client and context to GET the supplied absolute URL, preserving
its path/query. Success requires exactly HTTP 200 and one complete JSON object
with nonblank `name` and positive `revision`. Metadata has exported Name string
and Revision int fields with lowercase JSON keys. Nonblank name bytes are
preserved. Trailing JSON, malformed/null data, transport/read/status/field
failures produce a zero Metadata and useful error. Acquired response bodies
are closed on success or failure. Canceling the caller cancels in-flight HTTP
work and returns an error that supports errors.Is(err, context.Canceled).

The implementation already supports this behavior. Add high-quality reliable
success, error, response-lifetime and cancellation tests. State the boundary
actually verified by each dependency fixture and any execution restrictions.
Keep the public API, standard-library dependencies and Go 1.22 minimum. The
project does not prescribe a mocking framework or a new production seam.
