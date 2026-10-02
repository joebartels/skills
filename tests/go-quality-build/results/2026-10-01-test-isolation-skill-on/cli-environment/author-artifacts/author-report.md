# Configuration and command tests

Changed only `/private/tmp/go-testing-author-iu16ahig/cli-environment/config_test.go`. Production implementation, public API, dependencies and the declared Go 1.22 minimum remain unchanged.

Added 14 named configuration cases, shared across the loader and actual command boundary: unset defaults; explicit HTTP/read and HTTPS/write; each variable supplied alone; empty endpoint and mode; missing or unsupported scheme; missing host; malformed URL; unknown, capitalized and whitespace-padded modes. Successful loader values are compared with explicit expected Config values; failures require zero Config and diagnostics identifying the invalid variable and nonempty invalid value.

Environment mutation remains serial. `t.Setenv` registers restoration before any unsetting. Each loader case checks that loading leaves its environment unchanged and that subtest cleanup restores the parent value and presence. Additional nested cases verify restoration under originally unset, present-empty and explicit-value parents; focused selection does not rely on excluded siblings.

The CLI tests build `./cmd/showcfg` once into an owned `t.TempDir`, then execute the actual binary. Child environments explicitly remove both configuration keys before applying each case, while the parent carries deliberately invalid configuration. Success requires exit 0, empty stderr, the two lowercase JSON fields with expected values, and exactly one newline after the JSON value. Failure requires the real exit code 2, empty stdout and a useful diagnostic. Parent configuration is checked after each child case. Build deadline: 60 seconds. Command deadline: 10 seconds. Both use CommandContext and a five-second WaitDelay.

## Verified checks

All checks completed with exit code 0; stdout/stderr and exact command arguments are preserved in `checks.json`. Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local` through `rtk proxy`.

- `gofmt -w config_test.go`; final `gofmt -l config_test.go` emitted no paths.
- Local toolchain: `go version go1.26.5 darwin/arm64`.
- `go test -v -timeout=90s -run '^TestLoadFromEnv$' ./...`.
- `go test -v -timeout=90s -run '^TestShowcfgCommand$/^defaults$' ./...`.
- `go test -v -timeout=90s -run '^TestEnvironmentRestoration$/^present_empty$/^empty_mode$' ./...`.
- `go test -race -shuffle=on -count=5 -timeout=120s ./...`.
- `go vet ./...`.

The verification driver also imposed finite process deadlines of 20–150 seconds. `staticcheck` was not installed and was not run; no tools or dependencies were installed.

## Limitations and remaining risks

The Go 1.22 minimum is preserved and test helpers use APIs available by that version, but checks ran on the available Go 1.26.5 toolchain, not an actual Go 1.22 installation. Execution was on macOS arm64; Windows-specific environment behavior and the executable suffix branch were not exercised. Race detection instrumented the test process; the independently built CLI binary used ordinary `go build`. The CLI requires a local Go toolchain for its build fixture. Repetition covers the exercised serial environment and process paths and does not establish all schedules or platforms are defect-free.

The tests do not contact endpoint hosts: the documented boundary only parses configuration and prints it. Exotic URL forms beyond the selected malformed examples and output-device write failures are not exhaustively covered. No production behavior change was needed.
