# Disposable trial report

## Selected and opened skills

- `catalog/go-api-contracts/SKILL.md` — selected because `CollectMatching` adds an exported function and the task requires preserving `Collect`'s exact function type.
- `catalog/go-values-and-zero-values/SKILL.md` — selected because the predicate sees borrowed bytes and accepted frames need independent retained storage while preserving nil versus non-nil empty slices.

No other skill files or trial artifacts were opened.

## Decisions

- Added `CollectMatching(src Source, keep func([]byte) bool) ([][]byte, error)` without changing `Collect`'s signature.
- Called a non-nil predicate on the `Source.Next` view before making a copy. A nil predicate accepts every frame.
- Copied accepted non-nil frames into independent storage, preserving nil and non-nil empty frames. `Collect` uses the same private copy helper and retains its prior behavior.
- Preserved the existing end/error behavior: `io.EOF` completes normally; other errors return the accepted prefix and original error cause.
- Documented the borrowing, predicate restrictions, retention, nilness, and error behavior in the package README and exported function comments.

## Checks and results

- `GOWORK=off GOCACHE=/private/tmp/go-language-contract-author-cache rtk go test ./...` — passed (1 package).
- `rtk git diff --check` — passed.
- Tests cover direct borrowed-view identity, buffer reuse after predicate calls, filtering, nil predicate, nil versus non-nil empty frames, accepted prefix and error identity, and a compile-time assignment to the original `Collect` function type.

## Limits

- This was an isolated disposable package evaluation. No downstream module or real external consumer was available.
- Predicate mutation and retaining its borrowed argument are documented caller violations, so tests do not perform either.
- No race check was run; concurrent use is not part of the stated contract.
