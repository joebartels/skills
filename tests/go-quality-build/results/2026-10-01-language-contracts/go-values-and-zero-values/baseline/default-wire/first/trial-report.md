# Trial report

## Selected skill

- Opened `/private/tmp/go-language-trials-3b23/go-values-and-zero-values/baseline/default-wire/first/catalog/go-api-contracts/SKILL.md` because the change adds an exported constructor and must preserve established public and JSON behavior.
- Did not select the interface/composition or package-boundary skills: the change adds no interface or dependency wiring and keeps the existing package responsibility and imports within the standard library.

## Decisions

- Added `NewWithLimit(int) (Formatter, error)`. Positive values cap items, zero explicitly disables the cap, and negative values return an error.
- Kept `New`'s exact `func() Formatter` signature and retained the ten-item behavior for both `New()` and `Formatter{}`.
- Stored the unlimited choice separately from the numeric limit so explicit zero does not collide with the default zero value.
- Kept the `items` JSON tag and the `null` versus `[]` representation. `Format` only reslices its local input value.
- Updated the README and added checks for typed use of `New`, default behavior, configured and unlimited limits, invalid limits, input preservation, and established empty representations.

## Checks

- `GOCACHE=/private/tmp/go-language-trials-3b23/gocache-values-default go test ./...` — passed.
- `GOCACHE=/private/tmp/go-language-trials-3b23/gocache-values-default go vet ./...` — passed.
- `git diff --check` — passed.

## Limits

- Verification covers this disposable module and representative function assignment in its package tests; no downstream consumer repositories were provided.
