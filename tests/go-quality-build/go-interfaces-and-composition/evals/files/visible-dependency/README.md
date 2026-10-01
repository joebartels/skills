# Receipts

This internal library sends one receipt per Send call. The current single-tenant host sets RECEIPT_URL and uses http.DefaultClient. The new host keeps multiple tenant accounts active simultaneously, with separate endpoints and HTTP policies (authentication transport and timeouts). It already creates and owns those clients. It cannot change process-wide settings between requests. This internal API may change together with its callers; no compatibility shim is required.

The wire contract is POST to the configured URL, Content-Type application/json, body {"id":"..."}. JSON escaping must be correct. Any 2xx response succeeds; other statuses or transport failures are errors. Calls are synchronous, close response bodies, and respect the caller's context. No retry or background delivery is requested. Go 1.22 is supported.
