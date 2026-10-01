# Trial report

Status: implemented and verified.

Changed files: `main.go`, `main_test.go`.

`--skip-comments` is accepted before the positional `FILE` argument. When enabled, lines whose trimmed contents begin with `#` are skipped. Without the flag, those lines still fail integer parsing. The sum is written only after every retained line parses, and `main` retains its existing error and exit handling.

Package decision: kept the change in `package main`. This is a local single-binary CLI with no external Go consumers, and its existing `run` function already provides a testable boundary; an additional package would add no useful API boundary.

Verification:

- `rtk go test ./...` before implementation: 2 passed, 1 failed (`TestSkipComments`, expected).
- `rtk gofmt -w main.go main_test.go`: completed successfully.
- `rtk go test ./...` after implementation: 3 passed in 1 package.
