# Clamp lower-bound repair

Changed `clamp.go` to return `low` when `value < low`, allowing the inclusive boundary itself to follow the ordinary in-range return. Added `TestLowerBound` with ordinary subtests for a value below the range and a value exactly at its lower bound. The exported signature and `go 1.22` module minimum are unchanged.

## Verification actually run

Every Go invocation used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`; each verification process had a 60-second deadline. Test invocations also used `-timeout=30s`.

- Before the repair: `rtk proxy go test -run ^TestLowerBound$ -count=1 -timeout=30s ./...` exited 1, with both regression cases returning 2 instead of 1.
- `rtk proxy gofmt -s -w .` exited 0.
- `rtk proxy go version` exited 0: Go 1.26.5 on darwin/arm64.
- `rtk proxy go test -v -race -count=1 -timeout=30s ./...` exited 0; the existing interior test and both lower-bound cases passed.
- `rtk proxy go vet ./...` exited 0.
- `rtk proxy gofmt -l .` exited 0 with no unformatted files.

Complete verification stdout, stderr, and exit codes are retained in `checks.json`.

## Guidance selection

Opened only `go-core-style/SKILL.md`. The isolation guidance does not apply to pure calculations using local inputs and ordinary assertions. The other offered guidance does not apply because this change preserves the existing public contract, composition, and package responsibilities. Exact paths and selection reasons are recorded in `selection.json`.

## Limitations and remaining behavior risks

`staticcheck` was unavailable and was not installed, per task constraints. Go 1.22 itself was not executed; the source uses Go 1.22-compatible language features and the module minimum remains unchanged. Callers must continue to supply `low <= high`; inverted bounds remain outside the documented contract. This focused regression test covers the repaired lower branch and retains the existing interior test; it does not add upper-bound cases.
