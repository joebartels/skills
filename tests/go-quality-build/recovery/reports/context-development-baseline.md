# Sequential outbound stages implementation report

Date: 2026-10-03
Workspace: `/private/tmp/go-skill-recovery-20261003/context-development-baseline`

## Chosen supplied skills

- `go-api-contracts`: preserve the exported function type, ordered completed-prefix results, inspectable errors and Go 1.22 module contract.
- `go-interfaces-and-composition`: use the supplied concrete HTTP client, own response bodies and context cancellation resources, and keep each stage alive through read and close.
- `go-behavior-tests`: choose discriminating cases for exactly status 200, separate and combined read/close failures, completed-prefix retention, independent cancellation causes and completed success.
- `go-test-isolation`: use local transport/body fixtures and real local HTTP cancellation observations, synchronize readiness/completion with channels, and bound joins and waits.

Read each chosen `SKILL.md` from the read-only guidance directory and the behavior-observations and isolation-patterns references. No package-boundary change was needed.

## Decisions and artifacts

Changed `stages.go` and `stages_test.go`; added this report. The approved `README.md` and `go.mod` remain unchanged. No dependency, exported signature, constructor, interface, commit or PR was added.

`FetchAll` creates one caller-derived total timeout for the nonempty operation. Each private `fetchStage` gets a timeout within that operation, checks cancellation before making a request, and releases its own cancellation resources after body reading and closure. The next endpoint starts only after that release. The client and transport remain borrowed.

Only a returned status 200 is accepted. Read and close must both succeed before appending a body; all failure exits retain the completed prefix. Returned bodies are closed, including a nonnil response returned alongside a client error. The supplied client's HTTP policy is preserved.

`stageFailure` joins independently observed failures with context error and cancellation cause when cancellation is observed at that failed-stage decision. It never compares arbitrary errors directly. A successful final read and close remains success even when the caller is canceled during closure.

Tests preserve the existing sequential-success observation and add: stage release before the next request; empty and already-canceled calls; request/status/read/close failures and absence of later endpoints; separate and combined read/close errors; one shared total deadline; shorter stage and earlier parent deadlines; budgets covering reads and closes; noncomparable custom cancellation causes; success surviving cancellation; and active request/body cancellation through a real local HTTP transport and handler. A compile-time function assignment checks the exported function type.

## Exact verification commands and results

All Go tests and vet commands used `GOWORK=off` and `GOCACHE=/private/tmp/go-skill-recovery-cache`.

- `rtk proxy go version` — exit 0; `go version go1.26.5 darwin/arm64`.
- `rtk proxy gofmt -w stages.go stages_test.go` — exit 0.
- `rtk proxy gofmt -w stages_test.go` — exit 0 after the final test additions.
- `rtk proxy gofmt -l stages.go stages_test.go` — exit 0, no output.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...` — initial sandboxed run exited 1 because loopback listening was denied (`bind: operation not permitted`), before the local HTTP fixture could run. Final run with local socket access exited 0; `ok example.com/stages 0.312s`.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...` — exit 0, no diagnostics, both before and after final test additions.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -count=10 -shuffle=on ./...` — exit 0 with local socket access; initial result `ok example.com/stages 1.997s`, final result `ok example.com/stages 2.626s`.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestFailureRetainsOnlyCompletedPrefix/read_and_close$' -count=1 ./...` — exit 0 on restored final implementation; `ok example.com/stages 0.213s`.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -run '^TestCancellationAtLocalHTTPBoundary/body_read$' -count=3 ./...` — exit 0 with local socket access; `ok example.com/stages 1.192s`.

Signal check: temporarily changed the read/close failure return to omit `closeErr`, then ran `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestFailureRetainsOnlyCompletedPrefix/read_and_close$' -count=1 ./...`. It compiled and exited 1 with the meaningful assertion `error=read failed, missing close failed`. A `finally` block restored the source before the successful checks above.

## Verification limits

The module still declares Go 1.22, and implementation/test APIs are available at that minimum; execution used Go 1.26.5 rather than a Go 1.22 toolchain. Real cancellation was observed at the local HTTP transport/body and handler boundary. This does not establish termination for arbitrary remote services or noncooperative custom body implementations. Tests do not exercise external DNS, TLS or remote endpoints. Race/shuffle/repeated passing runs establish only the schedules exercised.

Status: implementation and stated verification complete.
