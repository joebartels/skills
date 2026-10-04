# recordload

A Go 1.22 library for reading `key=value` records in order. Empty lines and lines
whose first byte is `#` are ignored. Records split at the first `=`; empty keys,
empty values and repeated keys are accepted. Whitespace is preserved.

`Load` returns the valid prefix before the first malformed line, together with a
`*LineError` containing its one-based physical `Line` and original `Text` (without
the newline separator). Reader failures remain inspectable with `errors.Is` or
`errors.As`, including when malformed input and a reader failure occur together.
Bytes returned alongside a reader error are parsed; EOF is normal completion and
successful calls return a nil error.

```go
records, err := recordload.Load(strings.NewReader("#settings\nhost=localhost\nbroken"))
// records contains {Key: "host", Value: "localhost"}.
var lineErr *recordload.LineError
if errors.As(err, &lineErr) {
    fmt.Printf("line %d: %q\n", lineErr.Line, lineErr.Text) // line 3: "broken"
}
```

`Load` reads the entire input into memory without imposing a line-size cap. The
reader is borrowed and is never closed by `Load`.
