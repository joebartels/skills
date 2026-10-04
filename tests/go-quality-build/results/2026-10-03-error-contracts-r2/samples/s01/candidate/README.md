# recordload

Go 1.22 library. `Load(io.Reader)` reads `key=value` records in order, splitting
each record at its first `=`. Empty keys and values are accepted. Empty lines and
lines whose first byte is `#` are ignored. Whitespace and carriage returns are
preserved; comments are whole lines, not inline. A final newline is optional.

Malformed input returns the valid prefix and a `*LineError`. Its `Line` field is
the one-based physical line number, including blank and comment lines; `Text` is
the original line without its newline. Records after that line are omitted.

```go
records, err := recordload.Load(strings.NewReader("# settings\na=1\nmalformed\n"))
// records contains {Key: "a", Value: "1"}.
var lineErr *recordload.LineError
if errors.As(err, &lineErr) {
    fmt.Printf("line %d: %q\n", lineErr.Line, lineErr.Text)
    // line 3: "malformed"
}
```

Bytes returned alongside a read error are processed. On read failure, `Load`
returns the valid prefix and preserves reader causes for `errors.Is` and
`errors.As`. If those bytes are malformed, both the `LineError` and independent
reader failure remain inspectable. EOF is normal completion, and successful
calls return a nil error interface.

Readers are borrowed and never closed. `Load` reads the input into memory; it
imposes no line-size cap. Run `go test ./...` to check the library and usage example.
