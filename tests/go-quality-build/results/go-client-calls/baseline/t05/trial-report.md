# Trial report

Implemented `SumPositive` with a `v > 0` guard. The exported signature remains `func SumPositive(values []int) int`, and `go.mod` still requires Go 1.22.0 with no dependencies. Added exactly one regression test using negative, zero, and positive inputs; retained the existing positive-only test.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t05/catalog/go-api-contracts/SKILL.md`: the exported function's intended behavior changes; its function signature is preserved.
- `/private/tmp/go-client-campaign/trials/t05/catalog/go-behavior-tests/SKILL.md`: the regression checks the independently expected total and rejects the original unconditional addition.
- `/private/tmp/go-client-campaign/trials/t05/catalog/go-names-and-comments/SKILL.md`: the new test name describes the behavior without explanatory test prose.

Concurrency, context, composition, package-boundary, and test-isolation guidance did not apply to this deterministic local calculation and local-input test.

## Checks run

All Go commands used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off` through `rtk env`.

- Before the fix: Go 1.22.12 `go test -ldflags=-linkmode=external -run '^TestSumPositiveIgnoresNonPositive$' -count=1 ./...` failed at the intended assertion: got `-1`, want `5`.
- Go 1.22.12 `gofmt -w client.go client_test.go` completed.
- After the fix: Go 1.22.12 `go test -ldflags=-linkmode=external -count=1 ./...` passed.
- Go 1.26.5 `go test -count=1 ./...` passed.
- Go 1.22.12 `go vet ./...` passed.
- Both toolchain versions were confirmed with `go version`; final source and module contents were read back.

## Limits

Go 1.22 compatibility was exercised with 1.22.12, not the exact 1.22.0 release. No downstream consumers were supplied or checked. Integer-overflow behavior remains the existing Go `int` arithmetic behavior. No stage, commit, external action, or additional artifact inside source was created. The initial prompt read used `cat` before learning its `rtk` requirement; subsequent shell commands were prefixed with `rtk`.
