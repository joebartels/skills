# CSV report records

Add `Record(cells []string) string` to `lib`. It must produce one CSV record using the supplied `example.org/celltext` escaping policy, with commas between cells and a final newline. An empty or nil slice produces just a newline. Preserve cell contents, including embedded commas, quotes, and newlines. Add focused behavior tests for these cases.

`lib` is the independently published `example.org/reportcsv` library. Preserve the existing `Cell` signature and output and retain Go 1.22.0 support. Approved dependency releases are `example.org/celltext` v1.0.0 and v1.1.0, supplied under `registry`; do not add other dependencies.

The local celltext fork is intentionally used for ongoing development. Keep that workflow usable while implementing the library feature. Run existing and added tests from `lib`.
