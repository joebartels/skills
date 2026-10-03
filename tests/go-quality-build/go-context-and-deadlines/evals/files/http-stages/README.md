# Sequential outbound stages

Update FetchAll and tests without changing its signature, dependencies or Go 1.22 minimum. It uses the supplied nonnil client and context, finite HTTP/HTTPS endpoints, and positive total/stage durations. Existing tests cover basic sequential success; add the requested budget/lifetime behavior.

Create one total operation budget derived from the caller, respecting any earlier parent deadline. Each stage gets at most its configured stage duration within that remaining total scope; never restart the total budget for another stage. An already-canceled call starts no request. Propagate cancellation to active request and body consumption.

GET in endpoint order; only status 200 is accepted. Append a body only after complete reading and successful owned-body closure. Retain the ordered completed prefix if a later request/status/read/close fails. Preserve inspectable independent failures and cancellation classification/custom cause when observed at a failed stage's decision. Fully completed success is not invalidated by later cancellation.

Own and close each response body returned by a successful `client.Do`, including later status/read/close failure exits; borrow the client/transport. When `Do` returns an error, do not close a returned response body: redirect-policy error responses have already been closed by the client. Keep each stage scope live through reading and closure, then release its cancel resources. Empty endpoints succeed without calls. Context-aware interfaces alone do not prove remote work has stopped; scope verification claims to the exercised transport/body boundary.
