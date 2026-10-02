# Key codec

The 1.x public API is `Key{Region, Name string}`, `Encode(Key) (string, error)`,
`Decode(string) (Key, error)`, and `ErrInvalidKey`. This module supports Go 1.22
and uses only the standard library.

The next minor release accepts slash, space, percent and Unicode within fields.
The wire grammar is exactly two individually URL path-escaped segments, Region
then Name, separated by one unescaped slash. For example, `us/west` and `web blue`
produce `us%2Fwest/web%20blue`. Escaping uses `net/url.PathEscape` for each field;
decoding splits the wire before applying `net/url.PathUnescape` once per field.
A plus sign remains a literal plus sign; it does not stand for a space.

```go
key := keycodec.Key{Region: "us/west", Name: "web blue"}
wire, err := keycodec.Encode(key)
if err != nil {
    return err
}
// wire is "us%2Fwest/web%20blue".

decoded, err := keycodec.Decode(wire)
if err != nil {
    return err
}
// decoded equals key.
```

Import the package as `"example.com/keycodec"`. Runnable examples for both
functions are included in the tests.

Empty fields, malformed escapes and wire strings with the wrong segment count
are rejected. Use `errors.Is(err, keycodec.ErrInvalidKey)` to identify these
errors. Failed encoding returns an empty string; failed decoding returns a zero
`Key`, even if its first field was valid.

Previously slash and percent were rejected inside either field. The function
signatures, exported fields, and error sentinel remain unchanged. Simple keys
such as `east/web` retain their wire representation. Spaces and Unicode now
produce escaped bytes, so consumers should use the codec instead of joining
fields or comparing old unescaped representations. Decoding retains support for
previously accepted unescaped field values and also accepts valid lowercase hex
escapes. Values that look escaped are data: `%2F` encodes as `%252F` and decodes
back to `%2F` after one unescape.

Run `go test -timeout=30s ./...` to check the contract and runnable examples.
