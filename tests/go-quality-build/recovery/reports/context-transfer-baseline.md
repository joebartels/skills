# Accepted import receipts implementation report

Date: 2026-10-03

## Selected guidance

- `go-api-contracts`: preserved `ApplyBatch`'s exact signature and accepted-count/error contract. Read its `SKILL.md`.
- `go-interfaces-and-composition`: kept the supplied function dependencies and synchronous lifetime; finalization owns and releases its timeout scope. Read its `SKILL.md`.
- `go-behavior-tests`: checked accepted prefixes, forbidden later effects, each independent failure combination, cancellation classifications, and completed-success policy. Read its `SKILL.md` and `references/behavior-observations.md`.
- `go-test-isolation`: used explicit callback events, a release gate, bounded waits, and cleanup that releases and joins the operation. Read its `SKILL.md` and `references/isolation-patterns.md`.

Package boundaries were unchanged, so `go-package-boundaries` was not selected.

## Implementation and decisions

Changed `receipt.go` and `receipt_test.go`; left the approved README and `go.mod` unchanged. No dependencies, production goroutines, new exported APIs, commits, or PRs were added.

`ApplyBatch` checks `ctx.Err()` before each item and counts only successful callbacks. Cancellation observed at that boundary joins the standard classification with `context.Cause(ctx)`. An apply error remains the stopping error; the function does not replace it by checking cancellation after the callback. In particular, a successful final callback can cancel the caller without retroactively failing the completed batch.

If the accepted count is positive, finalization runs exactly once under `context.WithTimeout(context.WithoutCancel(ctx), finalizationBudget)`. This retains caller values and detaches cancellation and the caller deadline. The function waits synchronously for the cooperative callback and cancels the scope on return. `errors.Join` preserves independent apply and receipt failures, including wrapped and joined context errors. Zero accepted items require no receipt; empty input succeeds even when canceled.

Tests verify ordered success; canceled empty and nonempty input; first-item failure without a receipt; retained accepted prefixes; apply-only, finish-only, and combined failures; cancellation before the next item; final-item cancellation success; caller metadata; the independent deadline; scope release; deadline classification and custom cause; and waiting while a timed-out finalizer still holds caller-owned work.

The finalization test observes caller deadline expiry inside the first successful apply callback, then observes the receipt's own deadline and holds that callback until released. Cleanup releases the gate and waits for actual operation completion. No readiness sleeps or process-global state changes are used.

## Exact commands and verified results

All shell commands used the required `rtk` prefix. Commands ran from `/private/tmp/go-skill-recovery-20261003/context-transfer-baseline`.

Discovery used `rtk read README.md`, `rtk read /Users/jb/.codex/RTK.md`, `rtk ls`, and `rtk read receipt.go receipt_test.go go.mod`; all exited 0. Each selected guidance file and the two referenced files above was read using `rtk read` with its absolute path under `/private/tmp/go-skill-recovery-20261003/guidance`; all exited 0.

| Exact command | Result |
| --- | --- |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...` before the implementation change | Exit 1: meaningful assertions failed for pre-canceled input, missing finalization after apply failures, missing independent finish failure, starting the next item after cancellation, inherited canceled receipt context, and inherited receipt deadline. |
| `rtk proxy gofmt -w receipt.go receipt_test.go` | Exit 0. |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...` after the implementation change | Exit 0: `ok example.test/receipt 0.358s`. |
| `rtk proxy gofmt -w receipt_test.go` after adding the first-item-failure case | Exit 0. |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -count=20 -shuffle=on ./...` | Exit 0: `ok example.test/receipt 3.955s`; included the first-item-failure case. |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...` | Exit 0, no diagnostics. |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestIndependentFailures$/^both$' -v ./...` | Exit 0: the independently selectable dual-failure subtest passed. |
| `rtk proxy go version` | Exit 0: `go version go1.26.5 darwin/arm64`. |

## Limits and next action

The module still declares `go 1.22`; the added APIs are available by Go 1.22. Execution on an actual Go 1.22 toolchain was not performed. These library-level tests verify the supplied cooperative-callback contract, not arbitrary non-cooperation or every possible scheduler interleaving. No further implementation action is required for the approved README.
