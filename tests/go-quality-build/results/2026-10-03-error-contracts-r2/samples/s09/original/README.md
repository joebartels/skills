# accountlookup
Go 1.22 library. Backend is an extension protocol; its diagnostics may contain
credentials or SQL. Service exposes ErrMissing directly for absent accounts:
existing callers use err == ErrMissing. All other backend failures are classified
as ErrUnavailable without exposing backend identity, type or sensitive text.
A nonempty value is usable only on success. Empty keys currently reach the backend.
The task adds ErrInvalidKey for blank keys and adds safe key/operation context to
unavailable failures. Preserve existing error contracts and keep the public API.
