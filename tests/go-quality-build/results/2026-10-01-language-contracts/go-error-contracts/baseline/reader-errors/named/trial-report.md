# Trial report

Selected and opened skill: `catalog/go-api-contracts/SKILL.md`.

Other supplied skills were not opened: the change does not add an interface,
change dependency wiring or resource ownership, or move package responsibility.

## Decisions

- Added exported `*LineError` with one-based physical `Line` and original `Text`.
- `Load` ignores empty lines and lines whose first character is `#`; it accepts
  empty keys and values and splits each record at its first `=`.
- Parsing returns the valid prefix on malformed input or reader failure.
- Reader failures are wrapped with `%w`; data returned with a non-EOF error is
  parsed before returning that error. EOF is normal completion.
- Used `bufio.Reader.ReadString` to avoid imposing a line-size limit.
- Documented the behavior and exercised the exported error through an external
  test package.

## Checks

- `rtk gofmt -w load.go load_test.go api_contract_test.go` completed.
- Initial `rtk go test ./...` could not access the default Go build cache under
  `/Users/jb/Library/Caches/go-build` in this sandbox.
- `rtk env GOCACHE=/private/tmp/go-error-contracts-reader-named-cache go test ./...`
  passed: `ok example.com/recordload`.

## Limits

Verification is limited to the disposable module and supplied tests. No external
downstream repository or supported release policy was provided. CRLF-specific
normalization is not documented or added; the existing parser also retained a
carriage return as line text.
