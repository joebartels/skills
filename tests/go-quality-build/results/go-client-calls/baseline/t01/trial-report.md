Implemented `Submit` and focused tests in `source/client.go` and `source/client_test.go`. Public signatures, module path, dependency set, and `go 1.22.0` declaration are unchanged. Nothing was staged or committed.

Selected and opened skill paths:

- `/private/tmp/go-client-campaign/trials/t01/catalog/go-api-contracts/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t01/catalog/go-behavior-tests/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t01/catalog/go-concurrency-and-ownership/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t01/catalog/go-context-and-deadlines/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t01/catalog/go-interfaces-and-composition/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t01/catalog/go-names-and-comments/SKILL.md`
- `/private/tmp/go-client-campaign/trials/t01/catalog/go-test-isolation/SKILL.md`

Also opened the selected skills' `go-behavior-tests/references/behavior-observations.md` and `go-test-isolation/references/isolation-patterns.md`. Allowed task inputs were README, existing Go implementation/tests, go.mod, and the original client.go from task-local Git. Consulted primary Go 1.22 net/http source for response ownership, redirect processing, request cancellation, Client.Timeout error replacement, and transport replay behavior. No other skills, repository instructions, trials, private probes, metadata, input directories, or outcomes were read; no agents were used.

Decisions:

- One caller-derived 500ms context remains alive through all application attempts, bounded reads, body closure, and 10ms cancellable backoffs. Earlier caller deadlines win. Caller cancellation stops additional attempts, while completed success remains successful.
- Deduplication requires a nonempty caller identity before work. The exact supplied key is reused, with a payload snapshot for all attempts. At most three application attempts are permitted with deduplication; without it, exactly one attempt is permitted and the idempotency header is absent. Transport/read failure reports an unknown outcome and retains causes for errors.Is.
- Only HTTP 200 succeeds. HTTP 503 and transport/read failures can retry under deduplication; other statuses are terminal. Redirects remain terminal, including malformed Location responses and caller redirect-policy errors. A shallow client copy composes the supplied callback with that terminal-status rule; the caller client is not mutated. Transport, jar, timeout, legacy cancellation support, and ownership are retained. Transport and body observations preserve causes that net/http timeout handling can replace.
- Response reads use a 4097-byte probe to detect overflow. Success bodies over 4096 bytes are rejected; returned diagnostic bodies are limited to 4096 bytes. Every successfully acquired response body is closed before another attempt. Bodies already closed by http.Client on redirect errors are not closed again.
- Tests use the actual supplied http.Client boundary, handwritten transports/bodies for request bytes and ownership, and local HTTP servers for body deadlines and committed writes with lost replies. A second caller invocation uses the same key/body and observes one simulated deduplicated effect and the original result. A write without deduplication is observed once and returns ambiguity.

All Go checks used this scoped environment:

`GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off`

Checks actually run:

- Go 1.22.12 full package tests: `/private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -timeout=15s ./...` passed during implementation and after redirect classification changes.
- Restored original client.go temporarily and ran `go test -ldflags=-linkmode=external -run=TestSubmit(RejectsBeforeWork|Attempts|BodyBoundsAndFailures)$ -timeout=10s ./...`. It compiled and failed meaningful assertions for identity validation, deadlines, retry counts, ambiguity, and oversized success. Fixed source was restored in a finally block.
- Final Go 1.22.12 full suite: `go test -ldflags=-linkmode=external -race -shuffle=on -count=3 -timeout=30s ./...` passed. An additional earlier race/repetition run also passed. The macOS external linker emitted an LC_DYSYMTAB warning; test execution and race checks succeeded.
- Final host Go 1.26.5 full suite: `go test -timeout=15s ./...` passed; an earlier host run also passed.
- Go 1.22.12 `go vet ./...` passed.
- Final independent named cases: `go test -ldflags=-linkmode=external -v '-run=^TestSubmit(Attempts|BodyBoundsAndFailures|CommittedWriteReplay)$/^(503_then_success|oversized_success|deduplicates=true)$' -timeout=15s ./...` passed all three selected leaves.
- `gofmt` was applied. `git diff --check` passed. Final `git status --short` showed only unstaged changes to client.go and client_test.go.

Limits: actual minimum-version execution used Go 1.22.12, not 1.22.0. Local HTTP tests do not establish external DNS, TLS, or real production server deduplication. The three-attempt limit counts application attempts; the borrowed transport can perform its documented internal retries. Context deadlines require cooperative transports/readers and cannot forcibly terminate arbitrary custom I/O. No global settings or external actions were changed, and test binaries/cache stayed outside source. The initial prompt bootstrap read used cat before the prompt's RTK instruction was known; subsequent shell commands were prefixed with rtk.
