# Store test extension

Changed only `/private/tmp/go-fresh-author-txf1lez2/task-6/module/store_test.go`.
The production implementation, public API, dependencies and `go 1.22` module
directive are unchanged.

Preserved the existing serial replacement checks, using `t.TempDir` for their
fixture. Added `TestStoreIndependentInstances` with a serial `instances` group
containing three parallel independent stores. Each uses the identical key in a
distinct root, and each runs named `initial`, `replacement` and `repeat` children
serially within its own store. Distinct instance values exercise independence;
shorter and empty replacements exercise complete-value replacement; repeated
values exercise repeated same-key writes.

The outer test owns each root with `t.TempDir`, so files survive all descendants
and the observations after the group returns. Each instance sends its expected
last selected value to a buffered channel. The outer test reads those values
through a fresh Store and directly from each supplied root after all parallel
children finish. Cases excluded by `-run` contribute no expected work. All
expected values come from test inputs. The channel has room for every instance
and cannot block a child waiting for the parent to resume.

## Guidance and inspected files

Opened `go-core-style` for Go style and `go-test-isolation` plus its referenced
isolation patterns for file ownership, parallel descendant lifetime and focused
child execution. The other three offered skills were skipped because this
change does not alter exported contracts, composition or package boundaries.
Exact opened paths and individual reasons are in `selection.json`.

Task inputs inspected were the authorized dispatch, module README, go.mod,
store.go and store_test.go. No other repository files, author workspaces,
evaluation data or controller probes were inspected. No delegation, tool
installation, commits or external changes occurred. The initial dispatch read
used `cat` before its command-prefix instruction was known; all subsequent shell
commands used `rtk proxy`.

## Actual verification

All commands below passed. Command arrays, cwd, explicit non-secret environment,
stdout, stderr, exit codes and durations are preserved in `checks.json`.
Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and
`GOTOOLCHAIN=local`. Every post-change command had a 90-second process deadline;
test suites also had the finite deadlines shown below.

- Baseline: `go test -timeout=30s ./...`.
- Formatting: `gofmt -w store_test.go`, followed by `gofmt -l store_test.go`
  with empty output.
- Local compiler: `go version` reported `go1.26.5 darwin/arm64`.
- Full suite: `go test -v -timeout=30s ./...`; output shows all three independent
  instance children continuing and their nine nested cases passing.
- New test alone: `go test -race -timeout=30s -run=^TestStoreIndependentInstances$ ./...`.
- Focused leaves with race detection and 30-second suite deadlines:
  `TestStoreIndependentInstances/instances/alpha/replacement`,
  `TestStoreIndependentInstances/instances/beta/repeat`, and
  `TestStoreIndependentInstances/instances/gamma/initial`, each selected by
  anchored component regexes. These checks exclude siblings and other value rows.
- Repetition: `go test -race -shuffle=on -count=30 -timeout=60s ./...`.
- Static analysis: `go vet ./...`.

## Limits and remaining risks

Staticcheck was absent on PATH and skipped because tool installation was
forbidden. Tests were run with the available Go 1.26.5 compiler, not a Go 1.22
compiler; the implementation uses only test helpers available by Go 1.22 and
relies on the module's Go 1.22 loop-variable semantics. The exercised boundary
is real local temporary-file I/O on macOS. No listener substitute or approved
rerun was needed. Other platforms and every possible scheduling interleaving
remain unverified. Same-key concurrent writes within one root are outside the
README contract and remain serial in these tests. Passing race and repeated
runs provide evidence for exercised paths, not a proof against all flakes.
