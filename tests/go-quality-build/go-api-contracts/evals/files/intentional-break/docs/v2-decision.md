# Accepted: explicit parse errors in v2

The maintainers approved a major release to eliminate silent invalid ports.
The new API is Parse(text string) (int, error). Valid input is one or more
ASCII decimal digits representing a value in 0..65535. Leading zeros are
accepted. Empty input, signs, whitespace, non-digits and values outside that
range are errors. Failed parses return zero and a non-nil error.

Publish v2 as example.com/portnum/v2. Preserve no v1 Parse shim or second
legacy parsing API in this module: the existing v1 release remains available
at its original import path. Update our example consumer and write a concise
migration guide for external callers, including imports, error handling and
behavior changes. cmd/portcheck takes exactly one argument, prints the parsed
number plus newline on success, and on invalid input prints a diagnostic only
to stderr and exits 2. No network access is needed to parse a number.
