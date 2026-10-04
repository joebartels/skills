# lineexport

Go 1.22 library and CLI for copying newline-separated records. No external
dependencies are required.

Build and run:

```sh
go build -o lineexport ./cmd/lineexport
./lineexport input.txt output.txt       # all records
./lineexport input.txt output.txt 100   # at most 100 records
```

Usage is `lineexport INPUT OUTPUT [LIMIT]`. Omitted or zero `LIMIT` means
unlimited; a positive integer limits the record count. Invalid arguments,
including negative, malformed or overflowing limits, exit 2 before opening files.
Paths are positional, including names beginning with `-`. Success exits 0 with
empty stdout and stderr. I/O or finalization failure exits 1 with a diagnostic on
stderr and no success message. The output file is created or truncated; an
accepted prefix or partial record may remain on failure. Publication is not atomic.

Library usage:

```go
n, err := lineexport.Export(r, w)           // unlimited
n, err = lineexport.ExportLimit(r, w, 100)  // bounded
```

Both functions borrow the `io.Reader`, own the `io.WriteCloser`, and close the
writer exactly once on every return path. A negative library limit fails before
reading or writing but still closes the writer. The returned count includes only
fully written records, even when close fails. A sole operation or close error is
returned unchanged; when both fail, `errors.Is` and `errors.As` can inspect both
causes.

Empty records and a final unterminated record count. Each accepted record gets a
trailing newline; CRLF input is normalized to LF. Records are written before the
next scan. The standard `bufio.Scanner` size limit and input buffering apply, so
the reader may advance beyond the last exported record.

Run tests with `go test ./...`.
