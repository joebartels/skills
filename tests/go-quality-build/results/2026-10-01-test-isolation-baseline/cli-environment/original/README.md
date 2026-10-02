# Configuration test work

`Config` has exported string fields Endpoint and Mode, serialized as lowercase
`endpoint` and `mode`. `LoadFromEnv() (Config, error)` reads INDEXER_ENDPOINT
and INDEXER_MODE. Unset variables default to `http://127.0.0.1:9090` and `read`.
An explicitly supplied endpoint must have an HTTP/HTTPS scheme and a host;
an explicitly supplied mode must be exactly `read` or `write`. Empty supplied
values are invalid. On failure the result is the zero Config and a useful error.

`cmd/showcfg` prints exactly one Config JSON value followed by newline on
success, with exit 0 and empty stderr. Failure exits 2 with a useful stderr
diagnostic and no stdout. The implementation already supports this behavior.
Extend reliable tests for defaults, supplied and malformed values, parent
environment preservation, and the actual command boundary. Do not depend on
the invoking shell's configuration or leave test process state changed.
Keep public API, standard-library dependencies and Go 1.22 minimum unchanged.
