Implemented `Commit` in `source/client.go` and added focused consumer-boundary tests in `source/commit_behavior_test.go`. The original tests, public signature, Go 1.22.0 module requirement, and pinned dependencies remain unchanged. No staging or commits were performed.

Selected and opened skills:

- `/private/tmp/go-client-campaign/trials/t11/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t11/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t11/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t11/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t11/catalog/go-test-isolation/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t11/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t11/catalog/go-concurrency-and-ownership/SKILL.md`

Also opened the selected skills' `go-behavior-tests/references/behavior-observations.md` and `go-test-isolation/references/isolation-patterns.md`. Inspected pinned gRPC v1.67.3 primary source for outgoing metadata copying, generated UnaryCall invocation, configured retries, disabled retries, and header commitment.

Decisions:

- Reject an empty identity with `InvalidArgument` before accessing the connection.
- Derive one 500ms child context from the caller and defer cancellation. This retains earlier deadlines and covers all native retry attempts.
- Copy outgoing metadata and replace `operation-id` with the explicit argument; retain other keys, values, and previously appended metadata without changing the caller's context.
- Keep exactly one generated UnaryCall at the application boundary. Do not add a retry loop, options that override native policy, connection configuration, or connection closure.
- Return successful response payload bytes and return gRPC failures directly. Document that a failed acknowledgement may follow an effect and that the server's deduplication contract supports repeating the identity and payload.
- Use the shipped generated service over bufconn with real gRPC client/server behavior. Test ownership uses cleanup, bounded joins, channels, and atomic counters; the borrowed connection is exercised by a subsequent call.

Checks actually run, with `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off` for all Go checks:

- `rtk proxy env ... /private/tmp/go-client-campaign/tools/go/bin/go version`: confirmed Go 1.22.12 on darwin/arm64.
- Before implementation, `go test -ldflags=-linkmode=external -run 'TestCommitPayloadMetadataAndBorrowedConnection|TestCommitTotalBudgetIncludesNativeRetries' -count=1 ./...`: failed meaningfully on the stale operation identity and missing 500ms budget.
- `rtk proxy /private/tmp/go-client-campaign/tools/go/bin/gofmt -w client.go commit_behavior_test.go`: completed.
- `go test -ldflags=-linkmode=external -count=1 ./...`: passed.
- `go test -race -ldflags=-linkmode=external -count=5 -shuffle=on ./...`: passed twice, including after the final test cleanup refinement.
- `go test -ldflags=-linkmode=external -count=1 -run '^TestCommitConnectionRetryPolicy$/^configured_retry$|^TestCommitEarlierDeadlineAndCancellation$/^in-flight_cancellation$' ./...`: passed twice, including on final source.
- `go vet ./...`: passed twice, including on final source.
- `rtk git diff --check`: passed on final source; status shows only modified `client.go` and untracked `commit_behavior_test.go`.

The Go commands above used `rtk proxy env` with the full environment listed above and the actual Go 1.22.12 binary. Tests cover payload bytes, metadata and its immutability, identity rejection, absent/configured/disabled retry policy, nonretryable status, one application invocation, committed headers followed by failure, caller repetition after an accepted effect, the total native-retry budget, earlier caller deadlines, cancellation, and exact exported function compatibility.

Limits: verification used in-memory gRPC transport, not an external production server or TCP/TLS network. The server deduplication guarantee is a supplied contract; tests model retained effects and repeated acknowledgements rather than verifying a production persistence system. Real-time deadline assertions have scheduling tolerances. Race builds emitted a macOS `LC_DYSYMTAB` linker warning but completed successfully. No Go/dependency versions or global settings were changed, and no binaries were written to source.
