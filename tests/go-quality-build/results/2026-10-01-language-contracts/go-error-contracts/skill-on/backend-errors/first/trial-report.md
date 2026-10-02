# Trial report

## Scope

Implemented the request in `source` only: reject whitespace-only account keys and add safe operation/key context to unavailable lookup failures. No files were staged or committed, and no external actions were taken.

## Selected and opened skills

- `/private/tmp/go-language-trials-3b23/go-error-contracts/skill-on/backend-errors/first/catalog/go-error-contracts/SKILL.md`
- `/private/tmp/go-language-trials-3b23/go-error-contracts/skill-on/backend-errors/first/catalog/go-api-contracts/SKILL.md`

`go-error-contracts` applies because lookup classifies and translates backend errors. `go-api-contracts` applies because the request adds an exported error sentinel and specifies compatibility with existing exported error behavior. The interface/composition and package-boundary skills were not selected because this change adds no dependency or package responsibility.

## Decisions and artifacts

- Added the exported `ErrInvalidKey` classification and reject keys for which `strings.TrimSpace(key) == ""` before calling the backend.
- Preserve the caller's original nonblank key when calling the backend.
- Preserve direct `ErrMissing` identity for backend missing-account results.
- Classify other backend failures with `ErrUnavailable` and add `lookup account %q` context. The backend error is translated away so its message, type, and identity are not exposed.
- Updated `source/lookup.go` and `source/lookup_test.go`.

## Checks and results

- `GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache rtk go test ./...` — passed, 4 tests in 1 package.
- `GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache rtk go vet ./...` — passed, no issues.
- `rtk git diff --check` — passed.

Tests cover blank-key rejection before backend invocation, unavailable classification and safe contextual text without backend diagnostics, direct missing sentinel identity, and successful value/key behavior.

## Limits

Verification used the disposable source module and its local tests only. No downstream consumer repository or external integration was available, so broader consumer compatibility was not independently checked.
