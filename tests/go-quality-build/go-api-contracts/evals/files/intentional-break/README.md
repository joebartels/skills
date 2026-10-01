# portnum

This checkout starts at v1.9.0; it is the development branch for v2.0.0.
See docs/v2-decision.md for the accepted release decision. Go 1.22 is supported.

The v1 Parse function returns a port number. Invalid text returns zero,
which is indistinguishable from a request for an automatically assigned port.
cmd/portcheck is our example consumer and is released with the module.
