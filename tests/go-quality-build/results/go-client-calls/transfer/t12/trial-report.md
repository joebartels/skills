# Trial report

## Selected and opened skills

- /private/tmp/go-client-campaign/trials/t12/catalog/go-api-contracts/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-behavior-tests/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-client-calls/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-concurrency-and-ownership/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-context-and-deadlines/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-interfaces-and-composition/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-names-and-comments/SKILL.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-test-isolation/SKILL.md

Opened supporting references:

- /private/tmp/go-client-campaign/trials/t12/catalog/go-client-calls/references/grpc-clients.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-behavior-tests/references/behavior-observations.md
- /private/tmp/go-client-campaign/trials/t12/catalog/go-test-isolation/references/isolation-patterns.md

Read README.md, client.go, client_test.go and go.mod. Inspected the pinned grpc-go v1.67.3 metadata implementation, stream retry/commitment implementation, retry dial option documentation, generated unary client, server handler shutdown option and upstream retry tests. No other trials, repository instructions or results were consulted.

## Decisions and changes

- Preserve the Commit function type, Go 1.22.0 declaration and pinned dependencies.
- Reject empty operationID with InvalidArgument before invoking the borrowed connection.
- Derive one cancellable 500ms scope from the caller context, preserving earlier deadlines and cancellation.
- Set exactly the supplied operation-id in a copied outgoing metadata map. Preserve unrelated caller metadata and leave the caller context/map unchanged. The pinned FromOutgoingContext implementation copies the map and values.
- Keep exactly one generated UnaryCall invocation. Native policy owns any configured retries; Commit neither reconfigures nor closes the supplied connection.
- Return successful SimpleResponse.Payload.Body and return gRPC failures directly, preserving status codes. Godoc retains uncertainty about effects after failed acknowledgements and explains caller retry with the same identity and payload.
- Add commit_test.go using the shipped generated test service over bufconn, with bounded startup/completion and owned fixture cleanup. Cover binary request bytes, independent response bytes, metadata preservation, validation before invocation, absent/disabled/enabled/exhausted native retries, permanent status, header commitment after one accepted effect, caller retry with one effect, total and earlier deadlines, native retry waiting and in-flight caller cancellation. Retain the existing test.

## Checks actually run

All Go commands used GOWORK=off, GOTOOLCHAIN=local, GOPATH=/private/tmp/go-client-campaign/gopath, GOMODCACHE=/private/tmp/go-client-campaign/modcache, GOCACHE=/private/tmp/go-client-campaign/cache and GOPROXY=off. Commands were prefixed with rtk. The toolchain was /private/tmp/go-client-campaign/tools/go/bin/go, confirmed as Go 1.22.12 darwin/arm64. Test invocations included -ldflags=-linkmode=external for this host.

- Before the implementation, `go test -ldflags=-linkmode=external -run 'TestCommit(MetadataAndResult|RejectsEmptyIdentityBeforeInvocation|Deadline|BudgetIncludesNativeRetryWaiting)$' -count=1 -timeout=15s ./...` failed meaningfully for missing identity metadata/validation and missing 500ms total budget, including a second server attempt after a one-second pushback.
- After implementation, `go test -ldflags=-linkmode=external -count=1 -timeout=15s ./...` passed.
- `go test -race -ldflags=-linkmode=external -count=1 -timeout=30s ./...` passed. The macOS linker emitted a non-fatal LC_DYSYMTAB warning.
- `go test -ldflags=-linkmode=external -shuffle=on -count=3 -timeout=30s ./...` passed.
- `go test -ldflags=-linkmode=external -run '^(TestCommitNativeRetryPolicy|TestCommitDeadline)$/^(disabled|earlier_caller_deadline)$' -count=1 -timeout=10s ./...` passed, selecting important children independently.
- `go vet ./...` passed.
- Go 1.22.12 gofmt formatted the changed Go files; `gofmt -l client.go commit_test.go` returned no files.
- Source inventory confirmed no test binaries were written into source. No files were staged or committed.

## Limits

The fixtures exercise actual generated gRPC calls, native retry policy, wire metadata and header commitment over an in-memory transport. External TCP/TLS, production deduplication storage, resolver-supplied configuration precedence, transparent transport retries and retry-buffer commitment were not independently exercised. Timing checks use real time with tolerances; race and repeated runs establish only the schedules exercised. No downstream consumer repository was available; a compile-time function assignment verifies the retained exported function type.
