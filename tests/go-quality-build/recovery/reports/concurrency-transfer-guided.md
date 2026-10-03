# Pipeline implementation trial

Date: 2026-10-03. Approved contract: this workspace's README.md.
Artifacts: pipeline.go, pipeline_test.go, and this report; go.mod unchanged.

Selected guidance:
- go-concurrency-and-ownership: callback lifetime, close ownership, early completion, exact stop acknowledgements.
- go-context-and-deadlines: caller classification/cause and independent failure retention.
- go-behavior-tests plus behavior-observations: accepted effects and independent cancellation-shaped failures.
- go-test-isolation plus isolation-patterns: event gates, bounded joins, cleanup independent of assertions.
- go-interfaces-and-composition: existing function callbacks already express the dependency boundary.
- API/package guidance was read; no public shape, package, or dependency change was needed.

Decisions:
- Two goroutines invoke the callbacks concurrently under one owned child context.
- Producer failure and every consumer return request cancellation before publishing completion.
- Only the producer wrapper closes values, after the actual producer callback returns.
- Buffered outcome channels join both actual callback returns, including held cleanup.
- Each outcome samples the child stop error before its own stop request.
- Suppression uses exact equality to that standard stop error only while the caller remains uncanceled.
- Wrapped/joined errors and an independent bare DeadlineExceeded remain visible.
- Caller cancellation adds both ctx.Err() classification and context.Cause(ctx); arbitrary causes are never compared.
- The capacity bounds only the value channel. Signature, standard-library dependencies, and Go 1.22 minimum remain unchanged.

Exact checks (all commands prefixed with rtk; workspace as working directory):
- `rtk proxy go version` -> go1.26.5 darwin/arm64.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^TestAlreadyCanceledCallerInvokesNeitherCallback$' -count=1 ./...` -> exit 1 before implementation; observed callbacks invoked and nil error.
- `rtk proxy gofmt -w pipeline.go pipeline_test.go` -> exit 0.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -count=1 -timeout=20s ./...` -> exit 0, 0.309s.
- `rtk proxy gofmt -w pipeline_test.go` -> exit 0 after making a test startup wait cancellation-aware.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run '^(TestEarlyConsumerCompletionJoinsProducerCleanup|TestProducerFailureStopsAndJoinsConsumer|TestIndependentWrappedAndJoinedFailures|TestIndependentBareCancellationBeforeStop|TestIndependentDeadlineAfterCoordinatedStop|TestCallerCancellationJoinsCallbacksAndPreservesCause|TestAlreadyCanceledCallerInvokesNeitherCallback)$' -count=1 -timeout=20s ./...` -> exit 0, 0.266s.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...` -> exit 0, no diagnostics.
- `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -count=3 -timeout=20s ./...` -> exit 0, 1.450s, no race diagnostics.
- `rtk proxy gofmt -l pipeline.go pipeline_test.go` -> exit 0, no paths reported.

Verified behavior: ordered values beyond capacity; early consumer success; accepted values retained on producer failure; both cleanup lifetimes joined; independent wrapped/joined and bare context failures retained; caller cancellation with a noncomparable cause; canceled entry invokes neither callback.

Limits: race/repeated runs cover exercised schedules, not every schedule. Pending-result checks use a 20ms observation window after explicit cleanup-start events; readiness does not depend on sleeps. Test joins have 3s diagnostic bounds and release gates during cleanup. Go 1.22 compatibility was checked from the preserved module and API choices, not by running a Go 1.22 toolchain. Suppression relies on the approved callback convention for exact shared-context acknowledgements. Noncooperating callbacks remain uninterruptible and must still be joined.

No commits, PRs, external dependencies, subprocess hosting, panic recovery, or additional workflow machinery were introduced.
