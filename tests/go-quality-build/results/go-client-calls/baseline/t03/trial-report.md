# Trial report

Implemented `Dial` and `Probe` in `source/client.go`; added focused contract tests in `source/policy_test.go`. Public signatures, `go 1.22.0`, and pinned dependencies are unchanged. No staging, commits, external actions, or binaries in source.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t03/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t03/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t03/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t03/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t03/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t03/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t03/catalog/go-test-isolation/SKILL.md`

Also opened the selected behavior and isolation skills' `references/behavior-observations.md` and `references/isolation-patterns.md`. Read task README, existing implementation/tests/module, and pinned gRPC source for default service configuration, retry commitment/backoff, method configuration, and the manual resolver. No other trial, repository instructions, private probes, or outcomes were inspected.

## Decisions

- Use gRPC's `WithDefaultServiceConfig` with an exact Health/Check method match, three total attempts, only UNAVAILABLE retry eligibility, 10ms/20ms initial/maximum backoff, and multiplier 2.
- Prepend that fallback before host dial options, forwarding all supplied options and retaining host overrides. Leave resolver service configs and native retry commitment enabled. Add no retry loop or interceptor.
- Derive one 500ms child context around one generated Check invocation and defer its cancellation. Earlier parent deadlines and cancellation propagate. Wrap RPC failures using `%w`, retaining gRPC status classification; reject every health value except SERVING.
- Borrow the supplied connection without changing or closing it. Keep test listeners, servers, connections, and probe goroutines owned and cleaned up by their fixtures.

## Checks actually run

Go commands used `GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`, through `rtk proxy env`.

- Go 1.22.12 `go test -ldflags=-linkmode=external -timeout=30s ./...` before implementation: failed meaningfully for the missing fallback policy, absent native retries, unwrapped errors, and missing total budget.
- The same Go 1.22.12 full-suite command after implementation: passed.
- Go 1.22.12 `go test -race -ldflags=-linkmode=external -shuffle=on -count=5 -timeout=60s ./...`: passed.
- Go 1.26.5 `go test -shuffle=on -count=5 -timeout=60s ./...`: passed.
- Go 1.22.12 `go vet ./...`: passed.
- Go 1.22.12 `go test -ldflags=-linkmode=external -run 'TestDialResolverPolicyPrecedence/two_attempt_policy|TestProbeNativeRetries/response_headers_commit_call|TestProbePreservesCallerDeadlineAndCancellation/in-flight_cancellation' -count=1 -timeout=30s ./...`: passed.
- `rtk proxy /private/tmp/go-client-campaign/tools/go/bin/gofmt -w client.go policy_test.go`, followed by `gofmt -l` on those files: no remaining formatting differences.
- Confirmed toolchain versions and inspected final implementation/source file inventory.

Tests cover exact effective policy fields and method scope; third-attempt success and three-attempt exhaustion; nonretryable errors; response-header commitment; one application Check invocation with native retries; valid resolver two-attempt and empty config precedence; wrapped status codes; borrowed-connection reuse without reconfiguration; all shipped health values; one absolute 500ms budget across attempts; earlier caller deadline; and in-flight caller cancellation.

## Limits

All RPC tests use the real generated health client/service and gRPC transport over local bufconn. External DNS, TLS interoperability, remote services, and additional platforms were not tested. Backoff parameters are checked through the effective public method configuration; random retry delays are not asserted as exact elapsed durations. Deadline assertions allow 100ms scheduling/transport tolerance. Passing race and repeated runs do not prove every possible schedule.

The Go 1.22 race build emitted a macOS `LC_DYSYMTAB` linker warning; linking and all tests still passed with exit status zero.
