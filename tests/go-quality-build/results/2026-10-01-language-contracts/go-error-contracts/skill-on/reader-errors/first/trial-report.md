# Reader errors trial report

## Selected and opened skills

- `catalog/go-api-contracts/SKILL.md`
- `catalog/go-error-contracts/SKILL.md`

The change adds an exported error type and changes the documented error and
partial-result behavior. The interface/composition and package-boundary skills
were not selected because no interface, dependency, lifecycle, package, or
import boundary changes.

## Decisions

- Blank lines remain ignored; lines beginning with `#` are comments.
- Malformed lines return the parsed prefix and `*LineError` with one-based
  physical `Line` and original line `Text`.
- `io.ReadAll`'s returned bytes are parsed even when it also returns a reader
  error. Valid parsed records are returned with that reader error preserved.
- When malformed input and a reader error occur together, `errors.Join` keeps
  both inspectable. EOF completes normally.
- The README documents these behaviors, first-`=` splitting, borrowed-reader
  ownership, and the lack of a line-size cap.

## Checks

- Formatted `load.go`, `load_test.go`, and `api_test.go` with `rtk gofmt -w`.
- Ran `rtk env GOWORK=off GOCACHE=/private/tmp/go-language-trials-3b23/go-error-contracts/skill-on/reader-errors/first/gocache go test ./...` from `source`; result: `ok example.com/recordload`.
- Tests cover comments and blank lines, physical line numbering, valid-prefix
  results, ordinary and simultaneous reader/malformed failures, bytes returned
  with a reader error, and external-package use of `LineError`.

## Limits

Verification used the trial module's package tests and an external-package test;
no downstream consumer module or compatibility policy was available. No checks
other than formatting and `go test ./...` were run.
