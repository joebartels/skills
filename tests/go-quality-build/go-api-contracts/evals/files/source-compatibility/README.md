# recordfmt

A public formatting library, currently v1.4.2. The next release is v1.5.0.
Minor releases preserve supported Go consumers. Go 1.22 is supported.
The package is used by independently released reporting tools; their source
is not in this repository. Integrators register New in their formatter
factory tables as well as calling it directly.

New(prefix) creates a formatter. Format(key, value) returns
prefix + key + ":" + value. Empty strings are valid inputs.
A formatter may be reused; construction and formatting perform no I/O.
