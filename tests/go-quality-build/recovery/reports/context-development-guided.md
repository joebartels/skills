# Guided context implementation trial

Completed 2026-10-03 in `/private/tmp/go-skill-recovery-20261003/context-development-guided`.

## Chosen guidance

- `go-context-and-deadlines`: one total budget, stage budgets within it, admission checks, body-lifetime ownership, and cancellation classification/cause at failed decisions.
- `go-package-boundaries`: retained the existing package and responsibility; no new dependency edge or package was warranted.
- `go-api-contracts`: preserved the exact `FetchAll` function type, ordered complete prefix, inspectable errors, empty success, and Go 1.22 module minimum.
- `go-interfaces-and-composition`: continued borrowing the supplied concrete HTTP client/transport; owned response bodies and context cancel functions. Added no interface or configuration layer.
- `go-behavior-tests`: independent expected bodies, status 206 rejection, separate and combined read/close failures, cancellation-shaped independent errors, and a noncomparable custom cause.
- `go-test-isolation`: independent fixtures, explicit request/body-start events, bounded waits and cleanup, real local HTTP cancellation, focused subtest runs, and shuffled race checks.

All six supplied `SKILL.md` files were read. The behavior-observation and test-isolation reference files were also read. No repository/history, other trial workspace, handoff, or controller-probe material was inspected.

## Implementation and decisions

`stages.go` now derives one operation context from the caller. A private stage helper derives each stage scope from that operation, checks cancellation before admitting work, and keeps the scope alive through reading and owned-body closure. It cancels each stage before the next starts and releases the operation scope on exit. Earlier parent deadlines remain effective.

Only a complete body with successful closure enters the result. Request, status, read, and close failures retain the earlier completed prefix. Independent failures are joined; a failed-stage decision samples `ctx.Err()` once and includes its standard classification and `context.Cause`. It performs no arbitrary error-interface equality. Successful completion has no late cancellation check, while cancellation observed before remaining work stops that work.

`stages_test.go` retains sequential success and adds focused contract tests. A function-type assignment checks the supported signature. Test fixtures check exact deadlines, release of prior stage resources, live contexts during read/close, rejected 206 status, prefix/error retention, custom cancellation causes, completion policy, local transport cancellation during request and body consumption, and stage timeout classification.

`go.mod` remains `module example.com/stages` with `go 1.22`; dependencies are unchanged. No commit or PR was created.

## Exact verification

Commands ran from the trial workspace. All Go test/vet commands used the required `GOWORK=off` and `/private/tmp/go-skill-recovery-cache` cache.

1. Before implementation:

   ```sh
   rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run 'TestEmptyAndCanceled|TestBudgetsAndStageLifetime|TestFailedStageRetainsPrefixAndErrors|TestCancellationRetainsIndependentFailures|TestCompletionAndNextStageCancellation' -timeout=10s ./...
   ```

   Exit 1, with meaningful assertions for unexpected canceled-call requests, absent deadlines and missing stage cleanup, lost close/custom-cause errors, and unwanted later work. This demonstrated regression signals against the original implementation.

2. Formatting:

   ```sh
   rtk proxy gofmt -w stages.go stages_test.go
   rtk proxy gofmt -w stages_test.go
   rtk proxy gofmt -l stages.go stages_test.go
   ```

   All exited 0; the final formatting check returned no files.

3. Full suite:

   ```sh
   rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -timeout=15s ./...
   ```

   Initial sandbox run exited 1 because `httptest` could not bind a loopback socket (`bind: operation not permitted`). The same command, authorized through sandbox escalation, exited 0: `ok example.com/stages 0.172s`.

4. Final suite with race detection, ten repetitions, and randomized ordering:

   ```sh
   rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -count=10 -shuffle=on -timeout=30s ./...
   ```

   Ran with loopback permission; exit 0: `ok example.com/stages 1.372s`.

5. Vet:

   ```sh
   rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...
   ```

   Exit 0, no diagnostics.

6. Independently selected body-cancellation subtest:

   ```sh
   rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestLocalTransportCancellation/body$' -count=5 -timeout=15s ./...
   ```

   Ran with loopback permission; exit 0: `ok example.com/stages 0.182s`.

7. Independently selected simultaneous read/close failure:

   ```sh
   rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestFailedStageRetainsPrefixAndErrors/read_and_close$' -timeout=15s ./...
   ```

   Exit 0: `ok example.com/stages 0.276s`.

8. Toolchain:

   ```sh
   rtk proxy go version
   ```

   Exit 0: `go version go1.26.5 darwin/arm64`.

## Verification limits

The runtime checks used Go 1.26.5. Production and test APIs are available in Go 1.22, and the module minimum was preserved, but a Go 1.22 toolchain was not executed.

The real local HTTP tests establish cancellation propagation through the exercised transport and body-consumption boundary, including the local handler observing request cancellation. They do not establish completion of arbitrary remote work or forcibly interrupt an uncooperative transport/body. Error/closure instrumentation verifies the constructed body boundary. The supplied client's HTTP policies remain under its owner's control.
