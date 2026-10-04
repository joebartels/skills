# Trial report

Implemented `Dial` and `Probe` in `source/client.go`; added focused policy tests in `source/policy_test.go`. Public function types, Go 1.22.0, and pinned dependencies are unchanged. Existing serving regression remains intact.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t08/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t08/catalog/go-test-isolation/SKILL.md`

Opened the selected skills' `go-client-calls/references/grpc-clients.md`, `go-behavior-tests/references/behavior-observations.md`, and `go-test-isolation/references/isolation-patterns.md`. Checked pinned grpc-go source for default-config precedence, native retry eligibility/commitment, resolver construction, method-config inspection, generated health methods, and bufconn cancellation.

## Decisions

- Prepend `WithDefaultServiceConfig` to a fresh dial-option slice and forward host options unchanged. The fallback applies only to Health/Check: UNAVAILABLE, three attempts including the first, 10ms/20ms backoff, multiplier two. Host options and valid resolver configs retain their native precedence.
- Use one generated Check invocation with a single 500ms child context. Earlier caller deadlines and cancellation remain effective; defer its cancel. Wrap RPC errors with `%w`, accept only SERVING, and leave the supplied connection open and configured by its owner.
- Exercise actual generated unary calls over bufconn. Tests inspect effective method policy and count server executions and application invocations independently. Cover retry success/exhaustion, permanent rejection, header commitment, host-disabled retry, resolver two-attempt and empty policies, health values, total budget across retries, earlier deadlines, cancellation at entry/in flight, and borrowed-connection reuse. Compile-time function assignments protect signatures.
- Fixtures use synchronized observations, bounded setup/event waits, connection cleanup, server-handler completion, and Serve completion. No external listener, protoc, new production abstraction, or additional retry layer is used.

## Checks actually run

All Go checks used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off` via `rtk proxy env`. Actual toolchains were Go 1.22.12 and Go 1.26.5, both darwin/arm64. Go 1.22 tests used `-ldflags=-linkmode=external` for the host linker.

- Go 1.22.12 baseline and implemented full `go test ./...`: passed. Final uncached full test also passed.
- Go 1.22.12 `go test -race -shuffle=on -count=5 ./...`: passed.
- Go 1.22.12 focused race tests, three repetitions: header commitment, resolver two-attempt policy, earlier caller deadline, already-canceled caller; passed with independently selected subtests.
- Go 1.26.5 `go test -count=1 ./...`: passed.
- Go 1.22.12 `go vet ./...`: passed.
- `gofmt -l client.go policy_test.go`: clean.
- Controlled regression checks: restoring the original dial/budget behavior failed the missing-policy assertion and returned UNAVAILABLE after the first attempt. Removing only the timeout failed the total-budget assertion with approximately three seconds instead of 500ms. Restored the implementation and removed the temporary backup before final checks.

## Limits

Verified local grpc-go behavior, not deployed TLS, remote DNS, or production resolver integrations. Header commitment is exercised; retry-buffer commitment, transparent transport retries, server pushback, and throttling are left to unchanged grpc-go behavior. Timing checks use real timers and tolerances, not virtual time. Race builds emitted a macOS `LC_DYSYMTAB` linker warning but completed successfully. No dependency updates, staging, commits, global settings, or external actions were performed.
