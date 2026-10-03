# Concurrency transfer baseline trial

Completed 2026-10-03 in `/private/tmp/go-skill-recovery-20261003/concurrency-transfer-baseline`.

## Guidance selected

Read the approved `README.md`, starting `pipeline.go`, `pipeline_test.go`, and
`go.mod`, plus the following neighboring skill bodies:

- `go-context-and-deadlines`: preserve independent wrapped/joined cancellation
  errors, retain caller classification and cause, avoid arbitrary error equality,
  and keep cancellation separate from callback completion.
- `go-behavior-tests`, with `references/behavior-observations.md`: assert actual
  accepted values, original error identity, independent failures, and callback
  ownership; demonstrate an original-code assertion failure.
- `go-test-isolation`, with `references/isolation-patterns.md`: use explicit
  callback events and owned release gates, cancel/release/join in test cleanup,
  keep worker assertions in the test goroutine, and retain Go 1.22 APIs.

Also read `go-api-contracts` to assess applicability. Its skip rule applies to
this private helper with an approved fixed signature; no API redesign or
additional compatibility machinery was needed. The package-boundary and
interface/composition skill bodies were not needed.

## Implementation and decisions

- `pipeline.go` retains the package, exact function signature, positive-capacity
  precondition, standard-library-only imports, and module minimum `go 1.22`.
- An already-canceled caller returns its standard classification and custom
  cause without invoking either callback. Arbitrary custom causes are joined,
  never compared for equality.
- Both callbacks run concurrently under one owned child context. Successful
  producer completion closes the values channel for draining; producer failure
  cancels the child. Every consumer return, including successful early
  completion, cancels the child to stop production.
- The pipeline closes the channel only after the actual producer callback
  returns. It receives both callback outcomes before returning, so cancellation
  cannot release a callback still performing held cleanup.
- Each callback's result records shared-context and caller-cancellation state
  immediately after the callback returns and before it requests peer stopping.
  This keeps the error that initiates stopping distinct from a peer's exact
  standard stop acknowledgment. Only exact equality with the shared standard
  context error is eligible for suppression, and only when neither callback
  observed caller cancellation. Wrapped/joined errors and an independent bare
  `DeadlineExceeded` remain intact. Broad `errors.Is` classification is not used
  to infer origin.
- Caller classification and cause are captured at callback return, rather than
  performing a late cancellation check after both callbacks completed.

## Tests and observations

`pipeline_test.go` retains the original one-value test and adds focused checks
for ordered values and the requested channel capacity; early consumer success;
accepted progress after producer failure; held producer/consumer cleanup;
channel closure after producer callback return; caller cancellation with a
noncomparable cause; independent wrapped/joined cancellation errors on either
side; independent bare deadline errors; preservation of both callback failures;
and exact standard errors returned before coordinated stopping.

Callbacks publish startup, accepted-value, cleanup, and completion events. The
20 ms negative-observation windows begin after those events establish that
callbacks are active and held at explicit gates. Three-second bounds diagnose
stalls; test cleanup cancels, opens all owned gates, and waits for the pipeline
call to finish. No production timer or scheduling framework was added.

The following command failed against the original production code, with actual
assertion failures rather than a timeout: callbacks were invoked and both
caller cancellation classification and custom cause were missing.

```text
rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestCanceledCallerSkipsCallbacks$' -count=1 ./...
```

Final successful checks, all with exit status 0:

```text
rtk proxy gofmt -w pipeline.go pipeline_test.go
rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -count=1 ./...
rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestIndependentErrorsAfterCoordinatedStop/consumer/bare_deadline$' -count=1 ./...
rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -run '^(TestConsumerEarlyCompletionStopsAndJoinsProducer|TestProducerFailureStopsAndJoinsConsumer|TestChannelClosesAfterProducerReturnAndConsumerIsJoined|TestCallerCancellationJoinsHeldCallbacksAndPreservesFailures)$' -count=1 ./...
rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...
```

The final full suite reported `ok example.test/pipeline 0.439s`; the focused
subtest reported `ok example.test/pipeline 0.226s`; the selected race check
reported `ok example.test/pipeline 1.450s`; vet produced no diagnostics. The
suite and vet also passed before the final test-assertion refinements.

## Limits

Executed with `go version go1.26.5 darwin/arm64`. The unchanged Go 1.22 minimum
and used API availability were inspected; a Go 1.22 toolchain was not executed.
The selected race run checks the exercised callback lifetimes and shared state,
not every possible schedule. No schedule/profile matrix, arbitrary blocking
I/O, process host, panic recovery, external dependency, or new public API was
introduced or evaluated. Callback cooperation and caller preconditions remain
the approved contract. No repository/history, controller probe, other trial
workspace/report, commit, PR, agent, or global edit was used.

Artifacts: `pipeline.go`, `pipeline_test.go`, and this `report.md`. `README.md`
and `go.mod` were not changed. No trial-local open question remains.
