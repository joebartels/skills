# streamctl

Inspect a stored service cursor with `go run ./tools/streamctl TOKEN` from the workspace root. The command prints the topic and decimal offset, one per line. Invalid cursors report an error on stderr and exit with status 1.
