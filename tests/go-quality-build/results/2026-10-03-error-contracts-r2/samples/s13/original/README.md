# entryzip

A Go 1.22 library writes ordered ZIP entries to a caller-owned writer. Entry
bodies are borrowed readers. Keep WriteArchive source compatible and add
WriteSelected(w io.Writer, entries []Entry, keep func(string) bool) error.
A nil predicate includes all entries; otherwise filter by name before creating
an entry or reading its body. Keep selected names, contents and order.

An archive succeeds only when its central directory has been written. Report
entry creation, body copy and archive completion errors. If body processing and
archive completion fail independently, both causes must remain inspectable.
Preserve direct identity for a lone caller-provided failure. Complete the owned
archive writer once on every path; do not close the borrowed readers or output.
On failure, output can contain an incomplete archive; promise no rollback or
durability. Implement the change, add meaningful tests and concise usage docs.
Use the standard library and retain the module minimum.
