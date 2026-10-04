# lineexport

A Go 1.22 library and CLI for copying newline-separated records, using only the
standard library.

```go
n, err := lineexport.ExportLimit(reader, writer, 100)
```

`Export(reader, writer)` exports without a record limit.
The reader is borrowed; the writer is owned and closed exactly once on every
return path. Limit zero means unlimited. A negative limit fails before reading or
writing and still closes the writer. Each record is written before advancing to
the next. Empty records and a final unterminated record count; output records end
with a newline.

The returned count includes only fully accepted records. A single read, write,
validation, or Close error is returned unchanged. When an operation and Close both
fail, `errors.Is` and `errors.As` can inspect both causes. A short write with no
error returns `io.ErrShortWrite`. Output may contain an accepted prefix and part
of a failed record; there is no atomic-publication guarantee.

Reading uses `bufio.Scanner`, with its default token size limit and possible
buffered read-ahead. Reaching a positive record limit stops scanning successfully,
without requesting another record.

Build and run:

```sh
go build -o lineexport ./cmd/lineexport
./lineexport INPUT OUTPUT [LIMIT]
./lineexport records.txt exported.txt 100
./lineexport -input -output 0
```

Paths and the optional limit are positional, including paths beginning with `-`.
The limit must be a non-negative decimal integer that fits in a Go `int`.
Omitting it or supplying zero exports all records. Success exits 0 with no stdout
or stderr. Operational failure exits 1 with a diagnostic on stderr and no success
message. Invalid arguments exit 2 before opening either path. Valid invocations
create or truncate the output, so a failed export may leave a partial file.
