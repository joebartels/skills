# Trial report

Selected and opened skill: `catalog/go-api-contracts/SKILL.md`. The change adds exported error classification and changes the consumer-visible error message for unavailable lookups. The interface/composition and package-boundaries skills were not selected because no interface, dependency wiring, package, or ownership changes were needed.

Decision: reject keys whose `strings.TrimSpace` result is empty before calling the backend, returning the new exported `ErrInvalidKey` sentinel. Preserve the documented direct identity contract for missing accounts (`err == ErrMissing`). For all other backend failures, return `ErrUnavailable` wrapped with `lookup account %q` context; do not wrap or print the backend error, which may contain credentials or SQL.

Changed artifacts: `source/lookup.go` and `source/lookup_test.go`. Tests cover whitespace-only rejection before backend access, unavailable classification with safe operation/key context and backend diagnostic redaction, and the existing missing identity behavior.

Checks: `rtk gofmt -w lookup.go lookup_test.go` completed. `rtk go test ./...` first failed because the sandbox denied access to the default Go build cache under `~/Library/Caches/go-build`. Rerunning as `rtk env GOCACHE=/private/tmp/go-language-trials-3b23/go-error-contracts/baseline/backend-errors/first/cache go test ./...` passed (`ok example.com/accountlookup`). Build cache is outside `source`.

Limits: verification is package-local; no downstream consumer repository or supported-version matrix was provided. No separate external consumer compilation was run.
