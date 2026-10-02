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
    key := keycodec.Key{Region: "us/west", Name: "web blue"}
    wire, err := keycodec.Encode(key)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(wire) // us%2Fwest/web%20blue

    decoded, err := keycodec.Decode(wire)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("region=%q name=%q\n", decoded.Region, decoded.Name)
    // region="us/west" name="web blue"
}
```

Fields are values, not pre-escaped strings: a literal `%2F` is encoded as `%252F`
and decoded back to `%2F`. Decoding splits on the unescaped separator before
unescaping each field exactly once. Lowercase hexadecimal escapes are accepted.
A `+` remains a literal plus, while a space is encoded as `%20`.

Invalid inputs return `ErrInvalidKey` (also detectable with `errors.Is`). Encode
returns an empty string on failure; Decode returns the zero `Key`. No partially
decoded key is returned when either segment has a malformed percent escape.

The exported type and function signatures are unchanged for the minor release.
Existing simple keys such as `east/web` retain their wire representation. Encode
now escapes characters such as spaces and Unicode that the old implementation
emitted literally, and newly accepts slash and percent within a field. Consumers
should use Decode to recover field values instead of splitting and interpreting
the wire directly.
