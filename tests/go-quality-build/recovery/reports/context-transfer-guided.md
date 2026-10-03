# Guided implementation report

Implemented the approved README contract in `receipt.go`, with focused regression tests in `receipt_test.go`. `go.mod`, the exact ApplyBatch function type, package responsibility, and callback dependencies remain unchanged. No production goroutines or new dependencies were added.

## Selected supplied skills

- `go-context-and-deadlines`: choose cancellation admission boundaries, preserve independent failures and caller causes, detach only required finalization, and release its bounded scope after cooperative completion.
- `go-package-boundaries`: keep the existing receipt package responsible for sequencing and receipt ownership; no split or import-direction change is justified.
- `go-api-contracts`: preserve the exact public function type and Go 1.22 minimum; the compile-time function assignment covers a representative consumer of that shape.
- `go-interfaces-and-composition`: retain the existing supplied function dependencies and caller-owned data; invoke finish synchronously and wait through its cleanup.
- `go-behavior-tests`: observe accepted counts, forbidden later callbacks, independent wrapped/joined errors, cancellation classification and causes, and callback completion.
- `go-test-isolation`: use local contexts and explicit callback gates, bounded waits, and registered gate-release/join cleanup; retain Go 1.22-compatible test APIs.

## Decisions

Empty input returns success without checking cancellation or calling either callback. Each nonempty next item checks caller cancellation before admission. A failed apply retains its original error, and samples caller cancellation once to include the standard classification and custom cause when present. No late check changes a successfully completed final item into failure.

Any accepted prefix triggers exactly one synchronous finish. Its context combines `context.WithoutCancel(ctx)` with `context.WithTimeout(..., finalizationBudget)`, retaining values while discarding caller cancellation and deadlines. The owner defers cancellation until finish returns. `errors.Join` exposes independent apply and finish failures, with no arbitrary error-interface equality checks.

Tests cover ordered success and caller-context identity; canceled empty and nonempty entry; a custom caller deadline cause; retained accepted work after cancellation; metadata and an independent finalization deadline; scope release; successful final-item cancellation; first/partial apply failure; finish-only failure; joined independent failures during cancellation; a noncomparable custom cause; and waiting for cooperative finish after its deadline until it releases borrowed data.

## Exact checks and results

All commands ran from `/private/tmp/go-skill-recovery-20261003/context-transfer-guided` and were prefixed with `rtk`.

1. `rtk proxy gofmt -w receipt_test.go` — exit 0.
2. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -timeout 10s ./...` against the original implementation and new tests — exit 1, expected meaningful assertion failures: absent finalization deadline/scope release; callbacks after caller cancellation/deadline; missing caller causes; wrong accepted prefix; missing finish after partial failure; final callback receiving cancellation; and finish without a working budget. Existing successful ordering and independent finish-only/first-apply failure cases still passed.
3. `rtk proxy gofmt -w receipt.go receipt_test.go` — exit 0 after implementation, and again after comment/assertion clarity edits.
4. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -timeout 10s ./...` — exit 0, `ok example.test/receipt 0.206s`.
5. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...` — exit 0, no diagnostics, after final source edits.
6. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -shuffle=on -count=10 -timeout 20s ./...` — exit 0, `ok example.test/receipt 1.555s`, after final source edits.
7. `rtk proxy go version` — exit 0, `go version go1.26.5 darwin/arm64`.
8. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestReceiptIndependentFailures/both_during_cancellation$' -count=1 -timeout 10s ./...` — exit 0, `ok example.test/receipt 0.234s`.
9. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestReceiptWaitsForCooperativeFinish$' -count=1 -timeout 10s ./...` — exit 0, `ok example.test/receipt 0.373s`.

## Verification limits

The installed toolchain is Go 1.26.5; an exact Go 1.22 toolchain was not run. The module minimum remains 1.22, all selected context/error APIs exist by Go 1.21, and no newer test helper was used. The bounded repeat/race run checks exercised schedules, not every possible schedule. Callbacks are tested at the specified cooperative function boundary; no unprovided downstream integration is claimed. Inputs outside README assumptions were not given new validation or behavior.

No commit or pull request was created.
