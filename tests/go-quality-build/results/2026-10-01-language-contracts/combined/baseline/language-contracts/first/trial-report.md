# Trial report

## Skills

- Selected and opened `/private/tmp/go-language-trials-3b23/combined/baseline/language-contracts/first/catalog/go-api-contracts/SKILL.md`. The change extends exported `Load` behavior, adds exported `Select`, and changes the supported CLI grammar, partial-result behavior, and error reporting.
- Did not select/open `go-interfaces-and-composition`: no interfaces, constructors, dependency wiring, or lifecycle design changed.
- Did not select/open `go-package-boundaries`: all behavior remains in the existing package and package responsibilities/import direction are unchanged.

## Decisions and implementation

- Preserved `Load(io.Reader) ([]Item, error)`, `Item` fields, and JSON tags. `Load` skips blank and `#` comment lines, retains complete items before malformed input or a read failure, and exposes a simultaneous read cause through `errors.Is`. If malformed input and a read failure both occur, the returned joined error exposes both.
- Added `Select([]Item, string) []Item`. It filters keys by prefix in input order and copies each matching item, tag map, and non-nil tag byte slice. It preserves nil input as nil and returns a non-nil slice for non-nil input with no matches. It preserves nil versus empty tag-byte slices.
- Extended the CLI to accept `INPUT OUTPUT [PREFIX]`, retaining positional paths that begin with `-`. It writes the selected parsed prefix before reporting malformed input on stderr with exit status 1. Successful invocations emit no stderr and return 0; usage errors return 2; output errors return 1. Updated the README with these contracts and ownership limits.
- Added library tests for comments, malformed prefixes, bytes-plus-error reads, inspectable causes, EOF completion, deep snapshot independence, nil/empty semantics, and CLI tests for prefix selection, partial output, dash-leading paths, usage status, and output failures.

## Verification

- `GOWORK=off GOCACHE=/private/tmp/go-language-trials-3b23/cache/combined-baseline-language-contracts-first go test -count=1 ./...` — passed for `example.com/batchview` and `example.com/batchview/cmd/batchview`.
- `GOWORK=off GOCACHE=/private/tmp/go-language-trials-3b23/cache/combined-baseline-language-contracts-first go vet ./...` — passed.
- Toolchain: Go 1.26.5 darwin/arm64. The module declares Go 1.22; execution on Go 1.22 was not available or verified.

## Limits

- No downstream consumer repository was available; compatibility evidence covers the preserved signature and JSON tags plus representative package/CLI tests.
- Reader failure with partial bytes is tested at the library boundary using a reader returning `n > 0` and a non-EOF error in one call. The CLI's disk-file input cannot reproduce that reader behavior, so CLI partial-file-error reporting was not exercised end to end.
- Documentation promises no concurrency safety, atomic publication, retry behavior, or line-size cap.
