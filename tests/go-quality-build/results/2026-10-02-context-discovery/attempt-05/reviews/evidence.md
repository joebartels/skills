# Packet 02 verification evidence

All source/build/mutation diagnostics ran in disposable directories under `/private/tmp/packet-02-review-diagnostics`. Candidate SHA-256 equality was checked before and after review. `diagnostics/candidate-sha256.json` preserves the source manifest. All shell/subprocess command argv begin with `rtk`.

## Supplied evidence

`../verification.json` reports reconstruction hashes matched. Supplied ordinary build/test/vet/format checks passed on Go 1.26.5; the `GOQUALITY_GO` variable does not change the actual ordinary `go` executable. Supplied held checks passed on ordinary Go, explicitly invoked Go 1.22.12 with `CGO_ENABLED=0`, and race/shuffle/count=3. The supplied lost-total-budget mutation detached stages from the total context and was detected by author and held tests. Held-test source was not present in the packet, so these outcomes are supplied facts rather than independently inspected assertion coverage. No referenced archive paths were accessed.

## Executed baseline checks

Working directory: `/private/tmp/packet-02-review-diagnostics/baseline`.

| Exact command | Reviewer-observed outcome |
| --- | --- |
| `rtk proxy go version` | `go version go1.26.5 darwin/arm64` |
| `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache go test -mod=readonly -race -shuffle=on -count=3 -timeout=30s ./...` | Exit 0; `ok example.com/stages 1.657s` |
| `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache go vet -mod=readonly ./...` | Exit 0, no diagnostics |
| `rtk proxy gofmt -d stages.go stages_test.go` | Exit 0, no diff |
| `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache /private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -count=1 -timeout=30s ./...` | Exit 1, host loader abort: `missing LC_UUID load command`; not a candidate test assertion failure |
| `rtk proxy /private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go version` | `go version go1.22.12 darwin/arm64` (script environment includes workspace/toolchain/cache overrides) |
| `rtk proxy /private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go test -mod=readonly -count=1 -timeout=30s ./...` with `CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache` | Exit 0; `ok example.com/stages 0.369s` |
| `rtk proxy go build -mod=readonly ./...` with the script's workspace/toolchain/cache overrides | Exit 0 |
| `rtk proxy go list -m all` with the script's workspace/toolchain/cache overrides | Exit 0; selected only `example.com/stages` |

## Executed boundary and mutation checks

The exact edits, argv, working directories, environment overrides, outcomes, and raw stdout/stderr from the script are saved in `diagnostics/executed-checks.json`. `diagnostics/run_diagnostics.py` and `run_followup.py` reproduce the construction from candidate copies. Mutations compiled and passed the complete unmodified author suite; their failures came from independently added contract diagnostics, not compilation errors.

| Mutation | Author-suite outcome | Focused reviewer diagnostic |
| --- | --- | --- |
| Recreate total context per stage | Pass, exit 0 | Fails equal absolute total-deadline assertion |
| Cancel stage before body read | Pass, exit 0 | Fails live context in Read and Close, and custom cause preservation during body failure |
| Ignore close-only failure when read succeeds | Pass, exit 0 | Fails: accepted body `complete`, nil error |
| Veto completed success if total context canceled | Pass, exit 0 | Fails: completed body returned with cancellation error |

Command for non-network focused checks: `rtk proxy go test -mod=readonly -run '^TestReview' -skip StandardTransport -count=1 -timeout=30s ./...`, with script overrides, in each disposable directory. Baseline boundary diagnostic passed; each mutation failed its corresponding assertion.

The initial local server test was blocked by sandbox loopback bind permission. The authorized rerun in `/private/tmp/packet-02-review-diagnostics/boundary` used `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache go test -mod=readonly -run '^TestReview' -race -count=1 -timeout=30s -v ./...` and exited 0 (`ok example.com/stages 1.337s`). It exercised all six reviewer boundary tests, including a flushed HTTP 200 body that stalls until the request deadline: body reading stops, result classifies DeadlineExceeded, and the local handler observes its request context cancellation. This proves the exercised boundary only.

## F1 cancellation observation

Source: `candidate/stages.go:61` evaluates `ctx.Err()` before `context.Cause(ctx)`. If custom cancellation arrives between those samples, `errors.Join` receives nil classification and a nonnil custom cause. The lost classification is specific to custom causes; a default cause is itself `context.Canceled` and masks this mismatch.

Saved diagnostic: `diagnostics/stress_snapshot_test.go`.

- Real timer-context helper stress: `rtk proxy go test -mod=readonly -run '^TestReviewCancellationSnapshot$' -count=1 -timeout=30s ./...`, script overrides, `/private/tmp/packet-02-review-diagnostics/snapshot-stress`; exit 1 at iteration 21357. Returned error contained `operation failure` and `custom cancellation`, but `errors.Is(err, context.Canceled)` was false.
- Deterministic transition context: `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache go test -mod=readonly -run '^TestReviewErrorRepresentationDuringCancellationTransition$' -count=1 -timeout=30s ./...`, same directory; exit 1. The wrapper lets the first `Err()` sample precede cancellation and the subsequent `Cause` sample follow it. It asserts the semantic invariant, not a mandatory call sequence.
- Public FetchAll stress: initial 300000-call run and then `rtk proxy env GOWORK=off GOTOOLCHAIN=local GOCACHE=/private/tmp/packet-02-go-cache go test -mod=readonly -run '^TestReviewPublicFailureCancellationSnapshot$' -count=5 -timeout=30s ./...` passed; five-run output `ok example.com/stages 2.732s`. No public occurrence or production rate is claimed from that check.
- Suggested correction was applied only in `/private/tmp/packet-02-review-diagnostics/snapshot-correction`: sample `cancelErr := ctx.Err()`; return operation error if nil; otherwise join operation error, sampled classification, and cause. `rtk proxy go test -mod=readonly -run 'TestReview|TestSequential|TestAlready|TestParent|TestTotal|TestFailure|TestStatus|TestEmpty' -count=1 -timeout=30s ./...`, script overrides, exited 0 (`ok example.com/stages 1.019s`). The author suite, transition invariant, 100000 real-context iterations, and 300000 public iterations passed in this corrected disposable copy. Exact record: `diagnostics/suggested-correction-check.json`.

The deterministic wrapper is supplementary controlled evidence; the failing real timer-context stress independently establishes the race between observations in the actual context types used for stages. This is a logical observation inconsistency, not an unsynchronized-memory race, and a clean race detector does not rule it out.
