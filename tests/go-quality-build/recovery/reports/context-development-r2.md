# Sequential HTTP stages implementation trial

Date: 2026-10-03

## Scope and selected guidance

The approved specification was this workspace's `README.md`. I read only its
implementation/test/module files, `/Users/jb/.codex/RTK.md`, the assigned
`context-guidance-r2` skill bodies and references, and the standard initial
`using-superpowers` skill. The latter explicitly says dispatched subagents
should ignore its workflow. No repository history, other trial workspace,
handoff, controller probes, or reports were read. No agents, commits, PRs,
configuration edits, or dependency changes were made.

Applied skills from `/private/tmp/go-skill-recovery-20261003/context-guidance-r2`:

- `go-context-and-deadlines`: one total scope; stage children; cancellation
  admission, cause/classification, completion policy, and read/close lifetime.
- `go-api-contracts`: preserve the exact exported function type, Go 1.22 module
  minimum, ordered accepted prefix, and inspectable independent errors.
- `go-interfaces-and-composition`: retain the concrete supplied client and its
  policies; borrow its transport; own successful response-body closure.
- `go-behavior-tests`: derive independent assertions for status 200, failures,
  accepted progress, cleanup, and cancellation. Read its behavior-observations
  reference.
- `go-test-isolation`: controlled context events, owned loopback fixture
  lifetime, bounded waits, body instrumentation, and race/focused checks. Read
  its isolation-patterns reference.

## Artifacts and decisions

- `stages.go`: `FetchAll` creates one total timeout for nonempty work and checks
  observed cancellation before admitting each stage. A private `fetchStage`
  creates its timeout within that total scope and keeps it live through reading
  and body closure. Its deferred cancel releases the stage before another starts.
- Failed status, read, and close outcomes retain all independent errors using
  `errors.Join`. Before scope cleanup, `withObservedCancellation` samples
  `ctx.Err()` once and adds both standard classification and `context.Cause`
  when cancellation was observed. It performs no comparisons between arbitrary
  error interface values.
- Completed body acceptance requires successful reading and successful closure.
  A failed later stage returns the completed ordered prefix. A successful final
  body remains success even if the caller cancels during successful closure;
  additional work is refused when that cancellation is observed.
- Empty endpoints return the original nil successful result before cancellation
  checks. An already-canceled nonempty call starts no client request.
- A successful `http.Client.Do` transfers closure ownership. A `Do` redirect
  error may return a response whose body the client already closed; that body is
  not closed again. No borrowed client/transport fields are changed by production
  code, and neither borrowed object is shut down.
- `stages_regression_test.go`: compile-time exact function-type assertion;
  empty/canceled admission; absolute total, stage, and parent deadlines;
  context metadata and body/stage lifetime; retained prefix with request,
  206/503 status, read, close, and joined failures; cancellation-wrapping
  independent errors and a noncomparable custom cause; late completion policy;
  supplied-client redirect error and single closure; actual loopback cancellation
  before headers and after body reading begins; bounded blocking-body total and
  stage deadline expiry. Existing `stages_test.go` and `go.mod` are unchanged.

## Commands and observed outcomes

Every shell command used the required `rtk` prefix. All Go test/vet commands
used `GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache`.

1. `rtk --version`: exit 0, `rtk 0.42.4`.
2. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...`
   before new tests: exit 0, original sequential-success suite passed.
3. `rtk proxy gofmt -w stages_regression_test.go`: exit 0.
4. The same full-suite test command against the original implementation with new
   regressions: exit 1. Multiple meaningful assertions failed (missing budgets,
   admission after cancellation, missing close errors, lost lifetime ownership).
   The sandbox additionally prevented the loopback listener and caused a fixture
   panic; that listener failure is an infrastructure limit, not regression proof.
5. The same full-suite test command with approved sandbox escalation, still
   against the original implementation: exit 1, meaningful assertion failures
   also included missing cancellation classification/custom cause at the real
   loopback request boundary, and missing total/stage expiry. Runtime 4.175s.
6. `rtk proxy gofmt -w stages.go stages_regression_test.go`: exit 0.
7. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...`
   with escalation after implementation and the tightened body-read startup
   observation: exit 0, `ok example.com/stages 0.313s`.
8. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...`:
   exit 0, no diagnostics.
9. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -race -shuffle=on -count=10 ./...`
   with escalation: exit 0, `ok example.com/stages 1.764s`.
10. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go version`:
    exit 0, `go version go1.26.5 darwin/arm64`.
11. `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test -run 'TestHTTPInFlightCancellation/body=true|TestRetainedPrefixAndOwnedBodyFailures/read_and_close' ./...`
    with escalation: exit 0, `ok example.com/stages 0.225s`. Both independently
    selected named regression paths passed.

## Limits and next action

Verification used Go 1.26.5 with the unchanged `go 1.22` module declaration;
an actual Go 1.22 toolchain run was not performed by this trial. The controller
will perform that replay and independent evaluation after handoff.

The real HTTP observations cover a local plain-HTTP transport and cooperating
handler/body, including a handler observing its request context cancellation.
They do not establish arbitrary remote work termination, external DNS/TLS/service
behavior, or interruption of a noncooperating custom body. Context cancellation
requests stopping; it does not forcibly stop arbitrary blocking I/O. Passing
race/repeated checks cover only the schedules exercised. No frozen or controller
tests were available to or read by this implementation trial.
