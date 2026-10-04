# entryzip

`entryzip` writes ordered ZIP entries to a caller-owned `io.Writer` using the
standard library. It requires Go 1.22 or later.

```go
var output bytes.Buffer
err := entryzip.WriteSelected(&output, []entryzip.Entry{
    {Name: "notes.txt", Body: strings.NewReader("notes")},
    {Name: "debug.log", Body: strings.NewReader("debug")},
}, func(name string) bool {
    return strings.HasSuffix(name, ".txt")
})
// Check err before using output as a complete ZIP archive.
```

`WriteSelected` filters names before creating entries or reading their bodies,
preserving selected names, contents, and order. A nil predicate includes all
entries. `WriteArchive(w, entries)` retains its original signature and writes
all entries.

Success includes writing the central directory. Creation, body copy, and
completion errors reach the caller; independent processing and completion
failures are joined and can be inspected with `errors.Is` or `errors.As`.
A lone caller-provided error is returned directly. The owned ZIP writer is
completed once on every return path. Body readers and the output remain open
and caller-owned. On failure, output may contain an incomplete archive; no
rollback or durability is guaranteed.
