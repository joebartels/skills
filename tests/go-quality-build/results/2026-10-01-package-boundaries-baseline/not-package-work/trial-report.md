# Trial report

## Package decision

Kept the existing `mathutil` package and `buckets(n, size int) int` signature. Integer division gives complete buckets; a nonzero remainder requires one more. Checking the remainder avoids the overflow possible with `(n + size - 1) / size` at the largest `int`.

## Changed files

- `round.go`: count a final partial bucket.
- `round_test.go`: add partial-bucket cases, including the largest representable `int`.

## Commands and results

- `rtk go test ./...` before the implementation change: failed `TestPartialBucket` for `(1, 4)`, `(13, 4)`, and `(maxInt, 2)` as expected.
- `rtk go test ./...` after the implementation change: passed, 2 tests in 1 package.
- `rtk run gofmt -w round.go round_test.go`: passed.
- `rtk go test ./...` after formatting: passed, 2 tests in 1 package.
