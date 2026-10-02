# cursorload
Go 1.22 library. Collect owns the supplied cursor and returns every scanned string.
Failures may leave a valid prefix; the returned error distinguishes failed from
complete iteration. The new task adds CollectMatching(rows, keep) with a
func(string)bool filter (nil keeps all), preserves order, and keeps Collect.
The cursor is closed once on every path. Its close cannot invalidate already read
rows and is best-effort; iteration/Scan error is primary. No retry or buffering of
the complete input is needed. The callback cannot fail. No database dependency.
