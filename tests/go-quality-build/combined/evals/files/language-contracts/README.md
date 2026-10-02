# batchview
Go 1.22 library and CLI. Load reads key=payload lines, ignores blank lines, and
returns valid items before malformed input. Item.Tag data is editable caller-owned
state. Add comment lines beginning with # and handle bytes returned alongside
caller read errors: return the valid prefix and an inspectable read cause. EOF is
completion; no line-size cap is promised. Readers remain borrowed.
Add Select(items []Item, prefix string) []Item: retain matching keys in order in an
independently editable snapshot, including nested tag bytes. Empty prefix selects
all. Mutations to either input or output must not affect the other. Preserve nil
input as nil; non-nil input produces a non-nil slice even when no keys match.
Keep Load's exact signature and existing Item fields/JSON tags. Document errors,
partial results and ownership without claiming concurrency safety.
CLI: batchview INPUT OUTPUT [PREFIX]. Positional paths beginning with - remain
valid. The optional prefix uses Select. Always write the valid selected prefix,
even after malformed/read input; then report incomplete input on stderr and exit
1. Successful runs exit 0 with empty stdout/stderr. Usage exits 2; output failures
exit 1. No atomic publication or retry guarantee is required. Implement, test and
update concise usage docs; keep the package small.
