# lineexport
Go 1.22 library and CLI. Export copies newline-separated records from a borrowed
reader to its owned writer, returns the accepted record count and closes the
writer exactly once on every path. Each record is written before reading the next.
The task makes finalization errors observable and adds ExportLimit(r,w,limit).
Limit zero means unlimited; positive limit stops after that many records; negative
limit is rejected before reading/writing but still closes the owned writer.
Export remains unlimited and source compatible. Preserve direct error equality
when only one operation fails, and make both independent causes inspectable when
operation and Close fail. A final unterminated record counts. Empty records count.
CLI: lineexport INPUT OUTPUT [LIMIT]. Success is exit 0 without stdout; failure is
exit 1 with stderr and no success message. Invalid usage exits 2. Accepted prefix
may already exist on failure; no atomic-publication promise. Paths are positional,
including names beginning with -. No external dependency is needed.
