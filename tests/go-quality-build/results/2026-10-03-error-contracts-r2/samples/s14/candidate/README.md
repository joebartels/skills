# Private calculation

Go 1.22 package. The private `pages(items, size)` helper accepts nonnegative items
and a positive page size. Zero items need zero pages; a partial page rounds up.

For example, `pages(0, 2)` returns `0`, `pages(4, 2)` returns `2`, and
`pages(3, 2)` returns `2`.

Run tests with `go test ./...`.
