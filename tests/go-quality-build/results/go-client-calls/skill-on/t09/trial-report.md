# Trial report

Implemented `New` and `Fetch` in `source/client.go` and added focused tests in `source/policy_test.go`. Public function types, the Go 1.22.0 directive, and pinned module files were preserved. No staging, commits, global setting changes, agents, or external actions were used.

## Selected and opened skills

- `/private/tmp/go-client-campaign/trials/t09/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-client-calls/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t09/catalog/go-test-isolation/SKILL.md`

Also opened the selected skills' `references/http-clients.md`, `references/behavior-observations.md`, and `references/isolation-patterns.md`. Inspected the pinned gobreaker v2.4.0 primary source files `gobreaker.go`, `twostep_gobreaker.go`, and `counter.go` in the designated module cache. Inputs otherwise remained the task prompt, README, and source files.

## Decisions

- Keep the supplied nonnil HTTP client and its configuration borrowed. One application `Do` call retains its redirect policy; no application retry was introduced.
- Derive one 250ms operation context, preserving an earlier parent deadline, and retain it through response reading and owned-body closure.
- Require exact HTTP 200. Other statuses return `HTTP <status>`. Reject oversized complete results using a 4097-byte sentinel read; return no partial result. Bound diagnostic presentation to 4096 bytes while preserving underlying error matching through `Unwrap`.
- Maintain instance-local breakers keyed by case-normalized scheme and authority, excluding path and query. A mutex protects breaker lookup/creation; network work runs outside that lock.
- Use the pinned two-step gobreaker for three consecutive eligible 503 outcomes, a 100ms open interval, single recovery admission, and generation protection. Feed a separate health signal: completed 200 success is healthy; eligible 503 is unhealthy; cancellation, 400, other statuses, and non-status errors are neutral. Neutral recovery completions release admission without recovery.
- Preserve successful caller results even when cancellation accompanies completion, while excluding those completions from health accounting.
- Disabled instances bypass breaker lookup/admission and make one bounded request.

## Checks actually run

All Go checks used `GOWORK=off`, `GOTOOLCHAIN=local`, the task GOPATH/module/build caches, and `GOPROXY=off`. Go 1.22.12 tests used `-ldflags=-linkmode=external` on this macOS host. Listener-dependent checks used authorized escalation. Shell commands used `rtk`, except the initial prompt read, which used plain `cat`.

- Before implementation, Go 1.22.12 focused tests for oversized results, request deadline, and enabled dependency scope failed with meaningful assertions against the original code.
- Go 1.22.12 full tests with `-timeout=30s`: passed.
- Go 1.22.12 final full suite with `-race -count=3 -shuffle=on -timeout=30s`: passed. An earlier such run also passed.
- Go 1.22.12 focused selection of recovery/canceled-200, counting/400, scope/enabled, read failure closure, and redirect-error closure: passed.
- Go 1.22.12 `go vet ./...`: passed, including after the final test additions.
- Host Go 1.26.5 full suite with `-timeout=30s`: passed.
- `gofmt -l client.go client_test.go policy_test.go`: no output.
- Toolchain version checks confirmed Go 1.22.12 and Go 1.26.5; final module reads retained Go 1.22.0 and gobreaker v2.4.0.

Tests cover exact-status acceptance, 4096/4097 boundaries, bounded diagnostics with cause retention, body closure, borrowed redirect policy, real HTTP body cancellation, earlier-parent budgets, disabled behavior, independent scheme/authority scope, failure-history neutrality, canceled successes and 503s, success resets, held recovery admission, neutral admission release, reopening/recovery, and stale successes/failures across generations.

## Limits

Real local HTTP checks establish body-read cancellation and redirect policy. Controlled RoundTrippers establish detailed health and ownership schedules; no remote DNS/TLS/service integration was performed. The request ceiling counts application calls, not redirects or native transport replay under supplied client policy. Deadlines rely on cooperative transport/body implementations; Fetch does not launch detached work to forcibly interrupt arbitrary callbacks. Recovery tests use real 110ms cooldown waits and bounded channel joins. Race/repeat results cover exercised schedules rather than every possible schedule. The macOS external race linker emitted an `LC_DYSYMTAB` warning, but tests completed successfully.
