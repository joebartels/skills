# Configuration bundle loader

This Go 1.22 package loads service configuration with standard `encoding/json`.
Extend `Load` for configuration received from an untrusted reader. Preserve its
public API and invoke `replace` exactly once after accepting a complete bundle;
return the callback's error if it fails.

Accept exactly one JSON object with the case-sensitive member `services`, a
non-null array of at most `MaxServices` service objects. Each service must have
exactly `name` (a nonempty string) and `port` (an integer from 1 through 65535).
An empty services array is valid. Reject unknown, missing, null, wrongly typed,
or duplicate members, comparing member names after JSON escape decoding.

The complete document, including surrounding whitespace, must fit within
`MaxDocumentBytes`. Read at most `MaxDocumentBytes+1` bytes from the supplied
reader. Require the real end of the input after optional JSON whitespace before
calling `replace`. Syntax errors, invalid values, extra input, oversized input,
and reader errors must return an error without calling `replace`.
