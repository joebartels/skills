# Store test extension

Changed only `/private/tmp/go-fresh-author-8hjrllxp/task-6/module/store_test.go`.

The original serial child replacement checks remain, with their root owned by
`t.TempDir`. Added three independent Store instances under a serial grouping
subtest. The instances run in parallel against distinct outer-test-owned roots;
each instance serially writes the same key through initial, repeated-value and
replacement child cases. Distinct replacement values include a newline and NUL
in one instance, and assertions compare the complete expected string.

After the grouping subtest joins its parallel descendants, the outer test reads
each written Store again. Expected values come from successful selected writes,
so selecting one instance or one leaf with `-run` does not require excluded
siblings to execute. Each parallel instance owns separate expectation fields;
the outer test reads them only after the group finishes. Files remain usable
through those final observations and are cleaned up by the outer test.

No production source, public API, dependencies or `go.mod` changed. The module
still declares Go 1.22, and the added tests use APIs available in Go 1.22.

## Actual verification

All 11 recorded commands exited zero, with empty stderr. Exact argument arrays,
working directory, task environment overrides, stdout/stderr, process deadlines
and exit codes are preserved in `checks.json`.

- Local toolchain: Go 1.26.5 on darwin/arm64.
- `gofmt -w store_test.go` and `gofmt -l store_test.go`; the latter printed no files.
- `go vet ./...`.
- Full verbose suite with a 30-second test timeout.
- Full suite with race detection, shuffle, 20 repetitions, parallelism 3 and a
  30-second test timeout.
- Original serial test alone with race detection and 10 repetitions.
- Named leaves `alpha/initial`, `alpha/repeat` and `beta/replacement` selected
  individually with race detection and verbose output.
- Named instance `gamma` alone with race detection, shuffle and 20 repetitions.

Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and
`GOTOOLCHAIN=local`. Commands used `rtk proxy` for raw output and finite process
deadlines. `staticcheck` was not installed, so it was not run; no tools or
dependencies were installed.

## Limits and remaining risks

Verification exercised real temporary-file I/O and concurrent independent roots
on the local filesystem. No listener or network boundary was needed. Go 1.22 was
preserved but was not itself the installed toolchain used for verification.
Repeated race/shuffle checks cannot prove every schedule free of failures.
Same-key concurrent replacement within one root remains outside the requested
contract and was not introduced into the tests.

Guidance selection and actual opened task, skill and reference paths are recorded
in `selection.json`. Work stayed inside the assigned module and report directory;
there was no delegation, commit or external application change.
