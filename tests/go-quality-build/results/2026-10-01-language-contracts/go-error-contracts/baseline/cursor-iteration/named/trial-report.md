# Trial report

## Skills

- Selected and opened `catalog/go-api-contracts/SKILL.md` because `CollectMatching` adds an exported function contract.
- Did not select the interfaces/composition or package-boundaries skills: this change adds no interface, dependency, constructor, package, or import responsibility.

## Decisions

- Added `CollectMatching(rows Rows, keep func(string) bool) ([]string, error)`; nil `keep` accepts every scanned string and retained values preserve cursor order.
- Kept `Collect` as a compatibility wrapper around `CollectMatching(rows, nil)`.
- Closed the cursor once using `defer`. Returned scan errors and terminal `Rows.Err()` with the collected prefix; ignored close errors as best-effort.
- Documented behavior in README and covered filtering, nil filter, scan/iteration errors, partial results, close-once behavior, and the function type.

## Checks

- `rtk gofmt -w collect.go collect_test.go` and `rtk gofmt -w collect_test.go` completed.
- `rtk env GOCACHE=/private/tmp/go-language-trials-3b23/go-error-contracts/baseline/cursor-iteration/named/gocache go test ./...` passed (`ok example.com/cursorload`).
- `rtk env GOCACHE=/private/tmp/go-language-trials-3b23/go-error-contracts/baseline/cursor-iteration/named/gocache go vet ./...` passed.
- `rtk git -C /private/tmp/go-language-trials-3b23/go-error-contracts/baseline/cursor-iteration/named/source diff --check` passed.
- The initial test attempt used Go's default user cache and could not open a cache file; rerunning with a disposable cache under the trial directory passed.

## Limits

Verification covers the supplied package and a compile-time function-type assignment. No downstream consumer or database integration was available or needed for the documented interface.
