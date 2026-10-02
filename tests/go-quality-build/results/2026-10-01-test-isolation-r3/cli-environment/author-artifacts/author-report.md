# Configuration and actual command tests

Changes are confined to `/private/tmp/go-fresh-author-8hjrllxp/task-2/module/config_test.go` and the new `/private/tmp/go-fresh-author-8hjrllxp/task-2/module/command_test.go`. Production code, public API, standard-library dependencies and the Go 1.22 module minimum are unchanged.

Fourteen named cases exercise unset defaults, explicit HTTP/read and HTTPS/write settings, independent defaulting, supplied empty values, malformed URLs, missing scheme or host, unsupported scheme, and empty/unknown/uppercase/whitespace modes. Unit failures require a zero Config and a diagnostic naming the invalid setting. Unit cases also check that LoadFromEnv does not mutate their environment.

Environment tests stay serial. The fixture uses t.Setenv cleanup before explicitly unsetting absent values, retaining the distinction between unset and present-empty parent state. Focused nested cases verify restoration to unset, empty and supplied parent values without depending on excluded siblings.

Actual-command tests build `cmd/showcfg` once in a parent-owned t.TempDir and run the executable for every case. The builder preserves its normal Go configuration/cache inputs while removing application settings. Each child gets an explicitly filtered environment with only the requested application settings. An invalid parent environment remains unchanged. Successful commands must exit 0, have empty stderr, emit lowercase endpoint/mode JSON values matching the contract, and leave exactly one newline after that JSON value. Failures must exit 2, have empty stdout, and emit a diagnostic naming the invalid setting. Builds have a 45-second deadline and two-second WaitDelay; executable runs have five-second deadlines and one-second WaitDelay.

All recorded verification commands passed:

- Full verbose suite with a 90-second Go test deadline.
- Individually selected defaults, command empty-mode failure, and present-empty restoration child cases.
- Race-enabled suite with shuffle and five repetitions, bounded by a 120-second Go test deadline.
- go vet and a clean gofmt -l check.

Commands, cwd, explicit non-secret environment overrides, stdout, stderr and exit codes are preserved in `/private/tmp/go-fresh-author-8hjrllxp/task-2/report/checks.json`. Verification deliberately supplied invalid INDEXER_ENDPOINT and INDEXER_MODE to the invoking process. Local toolchain: Go 1.26.5 on darwin/arm64, GOTOOLCHAIN=local, GOCACHE=/private/tmp/go-quality-testing-cache. Staticcheck was not installed and was not run; no tools or dependencies were installed.

Limitations and remaining risks: Go 1.22 itself and other operating systems were not exercised; tests use APIs available in Go 1.22. Builder setup was exercised with the required explicit GOCACHE override; its environment preserves normal unset-cache prerequisites but an independently unset GOCACHE run was not performed. Race/repetition results cover these serial environment and real process paths and do not prove every possible schedule. Tests require the Go executable in PATH. No listeners, remote service, tool installation, delegation, commit or external application changes were needed. No production changes remain necessary for the README request.
