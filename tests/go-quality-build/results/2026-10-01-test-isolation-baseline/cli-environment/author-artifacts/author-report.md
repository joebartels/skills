# Configuration and command tests

Implemented only tests in the isolated module. Production `config.go`, `cmd/showcfg/main.go`, the public API, standard-library dependency policy and `go 1.22` declaration were preserved.

Changed artifacts:

- `/private/tmp/go-testing-author-y6_tmx1p/cli-environment/config_test.go`: 14 configuration cases cover both defaults, individual defaults, HTTP and HTTPS endpoints, read/write, explicit empty values, unsupported or absent schemes, absent hosts, malformed URLs and invalid modes. Failures must identify the relevant variable and return the zero Config. Every case sets or unsets both variables explicitly.
- `/private/tmp/go-testing-author-y6_tmx1p/cli-environment/command_test.go`: builds the real command into a test-owned temporary directory, then applies the same 14 cases to subprocesses. Success requires exit 0, empty stderr, exact lowercase JSON fields and exactly one trailing newline; failure requires exit 2, empty stdout and a diagnostic naming the invalid variable. No `go run` wrapper obscures the actual program exit code.

Environment isolation uses `t.Setenv` cleanup even when a case subsequently calls `os.Unsetenv`. Unit cases remain serial and compare both values and presence after each child cleanup. Additional nested cases verify restoration for parent environments that are unset, explicitly empty or supplied. Command environments strip inherited configuration keys before setting case values. Deliberately invalid parent values demonstrate that defaults do not depend on the invoking shell. No environment modifications survive test cleanup.

Verification actually run:

- Baseline: `rtk proxy go test -timeout=90s ./...` — exit 0.
- Final: `rtk proxy go test -v -race -shuffle=on -count=3 -timeout=90s ./...` with deliberately invalid INDEXER_ENDPOINT and INDEXER_MODE — exit 0, all cases passed on all three repetitions.
- Formatting: `rtk proxy gofmt -w config_test.go command_test.go`, then `rtk proxy gofmt -l config_test.go command_test.go` — final check exit 0 and empty output.
- `rtk proxy go vet ./...` — exit 0, empty stdout/stderr.
- `rtk proxy go version` — Go 1.26.5 on darwin/arm64.

Every Go invocation used GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. Verification Go tests had a 90-second test deadline and a 120-second enclosing process deadline. Command builds have a 60-second context; each tested command has a 5-second context. Subprocess WaitDelay is one second. Captured stdout, stderr and exit codes are in `checks.json`.

Known limitations and remaining behavior risks: staticcheck was unavailable and was not installed, as required by the dispatch. Execution was verified on Go 1.26.5, not a separately installed Go 1.22 toolchain; the module's minimum version was unchanged and the test APIs used are available in Go 1.22. Windows-specific environment filtering and executable suffix handling were not executed on this macOS host. These tests cover configuration and the requested command boundary; they do not inject stdout write failures or assert exact error wording beyond useful variable identification. No production behavior risk was introduced by these test-only changes.

Guidance selection and actual opened files are recorded in `selection.json`. Only the authorized dispatch, this module and the selected isolated go-core-style guidance were inspected. No delegation, dependencies, tool installation, commits or external changes were performed.
