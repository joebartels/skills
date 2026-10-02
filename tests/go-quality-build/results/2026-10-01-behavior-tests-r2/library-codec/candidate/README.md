# Key codec

The 1.x public API is Key{Region, Name string}, Encode(Key) (string, error),
Decode(string) (Key, error), and ErrInvalidKey. Errors retain that sentinel.
Previously slash and percent were rejected inside either field.

The next minor release accepts slash, space, percent and Unicode within fields.
The wire grammar is exactly two individually URL path-escaped segments, Region
then Name, separated by one unescaped slash. For example, us/west and web blue
produce us%2Fwest/web%20blue. Empty fields, malformed escapes and wire strings
with the wrong segment count are rejected. Both encode and decode must preserve
their independent consumer contract. No third-party dependencies are needed.

## Usage

```go
package main

import (
    "fmt"
    "log"

    "example.com/keycodec"
)

func main() {
    wire, err := keycodec.Encode(keycodec.Key{
        Region: "us/west",
        Name:   "web blue",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(wire) // us%2Fwest/web%20blue

    key, err := keycodec.Decode(wire)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%q %q\n", key.Region, key.Name) // "us/west" "web blue"
}
```

Fields are escaped independently with `net/url.PathEscape`. Spaces become
`%20`; slash and percent become `%2F` and `%25`; Unicode is escaped as UTF-8
bytes. A plus stays a literal plus rather than representing a space.
Decoding splits on the unescaped slash before calling `net/url.PathUnescape`
once per field, so `%252F` decodes to the literal field value `%2F`.
Lowercase hexadecimal escapes are also accepted.

Unescaped simple keys such as `east/web` retain their existing representation.
Consumers that construct wire keys themselves should escape each field before
joining them, especially when the field contains a slash, space or percent.
Callers can identify invalid input with `errors.Is(err, keycodec.ErrInvalidKey)`.
Encoding failures return an empty string; decoding failures return a zero `Key`.

Run the contract tests and executable examples with `go test ./...`.
