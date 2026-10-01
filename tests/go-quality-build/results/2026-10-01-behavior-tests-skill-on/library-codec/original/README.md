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
