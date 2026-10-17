# Codes

Codes validates document identifiers for the billing import tools.

Add `Schema(name string) ([]byte, bool)` to expose a format's JSON Schema and add
`codes --schema NAME` for import-form clients. Return a JSON object with
`$schema` equal to `https://json-schema.org/draft/2020-12/schema`, the catalog's
`title`, `type` equal to `string`, and a `pattern` accepting exactly the same
identifiers as `Valid`. For example, the invoice pattern is
`^INV-[0-9]{6}$`. Lookups are case-sensitive. Unknown names return `nil, false`;
callers may modify returned bytes without changing later results.

The CLI prints the schema document followed by a newline. An unknown schema
writes an explanation to stderr, exits with status 2, and writes nothing to
stdout. Preserve existing validation behavior and add focused tests.
