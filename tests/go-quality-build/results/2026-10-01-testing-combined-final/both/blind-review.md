# Independent review

Review boundary: the complete supplied `original/` → `candidate/` changeset in `/private/tmp/go-independent-review-tkrcgusq`. No repository, author report, other candidate, evaluation expectations, or writing skill was inspected. The project README defines the contracts. File references below are relative to this review directory. Original and candidate source were preserved; checks and intentional mutations used separate disposable copies under `verification/`.

## Testing — B
Scope: supplied changeset; standard-library Go1.22 file-backed library and `indexer` command, including the requested new regression test scope. Executed on Go1.26.5 darwin/arm64.
Coverage: inspected every supplied test, affected implementation, README, module directive, and relevant review references. Assessed CSV success/rejection/accepted-prefix effects; HTTP status/read/decode/validation/body ownership; snapshot publication and partial-write failures; callback timing, sequentiality, cancellation cleanup, independent invocations and error identities; exact exported function types; actual CLI startup, stdin, streams, exit statuses, recurrence and both process signals. Checked fixture ownership, cleanup and finite deadlines. No fuzz target or benchmark was introduced; neither is necessary to support the assessed deterministic contract tests.
Rationale: one independently actionable moderate issue makes command verification depend on an undeclared ambient build-cache variable. It fails visibly rather than falsely passing, and the complete suite is runnable with explicit cache configuration, so this is fragile normal developer verification, not an effectively unverified or misleading important contract. That selects B. Strong verified assertions do not cancel the finding.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [TG1] `candidate/index_test.go:60` and `:75` assert exact quoted/multiline text, duplicate replacement, accepted prior rows, rejection identities and absence of later rows. `:134` exercises a real rename failure instead of making a fake report failure. These test the intentionally nontransactional CSV contract rather than inventing rollback.
- [TG2] `candidate/refresh_test.go:33` asserts exact compact JSON/newline, preserved order/text, empty-array representation, prior binary bytes on rejection, supplied request context, request count and body closure. `:98` injects fetch/read causes and exercises filesystem failure boundaries. `:166` uses an actual HTTP server and a file descriptor opened before replacement to distinguish replacement from in-place writing. An in-place-overwrite mutation failed the snapshot assertion at `:195`; the partial-write test alone correctly did not catch that successful-publication mutation.
- [TG3] `candidate/write_failure_posix_test.go:25` confines `RLIMIT_FSIZE` and signal changes to finite child processes. It verifies a write-stage error and exact retained state for Put, CSV and Refresh, including a previously accepted CSV replacement and absent later rows. A direct-destination-write mutation failed all three retained-state assertions. This is evidence of consequential fault detection, not a test-count claim.
- [TG4] `candidate/serve_test.go:82` gates cooperative callback cleanup and verifies caller-context identity, join-before-release, release count, ordinary cancellation and independently meaningful errors, including noncomparable errors. A mutation releasing before callback completion failed the gate assertions at `:149`/`:151`; a mutation discarding every callback error after cancellation failed the independent-error assertions at `:160`. The 30ms observation window is supported by the cleanup gate and final ownership assertion, rather than being the sole evidence.
- [TG5] `candidate/serve_test.go:176` deliberately holds the first callback beyond the interval and records its completion time. A mutation measuring the interval from callback start failed `:238` with a next call roughly 17µs after completion instead of at least 40ms. `:258` verifies that canceling one invocation does not release or prevent recurrence in another.
- [TG6] `candidate/cmd/indexer/main_test.go:66` runs a built executable under process deadlines and `:90` asserts exact exit status, empty stdout and meaningful stderr. The positional Put grammar, stdin application and accepted-prefix effects are exercised at `:110` and `:126`. `:205` verifies recurring published snapshots and cancellation of an in-flight HTTP request through both interrupt and SIGTERM. A stdout mutation failed real process assertions at `:114`, `:117` and `:119`.

Bad

- [F1][moderate][introduced] `candidate/cmd/indexer/main_test.go:24` drops HOME and other default build-cache discovery variables, while setting GOCACHE from `os.Getenv("GOCACHE")`; `:36` uses that environment for the `go build` in TestMain. A usual developer environment can leave GOCACHE unset and rely on Go's default cache under HOME. The new harness then aborts before any command assertions. Verified by building the unmodified command test binary with the required explicit review cache and running that binary with only ambient GOCACHE removed: exit 1, stderr `build executable: exit status 1` followed by `build cache is required, but could not be located: GOCACHE is not defined and $HOME is not defined`. The parent environment still supplies HOME; the harness strips it. Original tests had no such child-build requirement. Primary remediation owner: Testing.

Suggested changes

- [F1] Give the child build a valid cache independently of ambient GOCACHE, for example by resolving Go's effective cache path or owning a private test cache, or retain the relevant default cache-discovery variables for the build. If runtime command isolation requires a narrower environment, separate the build environment from the runtime environment. Verify that the unmodified test binary reaches and passes assertions with ambient GOCACHE both set and unset.

Limits: all declared Go review commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. The unset-cache probe deliberately removed GOCACHE only from the already-built test binary's environment to assess the candidate's own child-build setup. Initial full/race runs failed because the sandbox prohibited localhost listener creation; this was an environment limit, not a candidate finding. With loopback permission, `go test ./... -count=1 -timeout=45s` and `go test -race ./... -count=3 -shuffle=on -timeout=60s` both passed, exit 0. Race instrumentation covered the parent/library tests; TestMain builds the command executable without `-race`. File-close failure was inspected but not fault-injected. Partial-write fixtures target darwin/linux and were executed on darwin only. Windows replacement semantics, a Linux run and an actual Go1.22 runtime run are unsupported by this execution claim. Exact commands, stdout, stderr, exit codes and mutation results are in `output/checks.json`.

## Correctness & Compatibility — A+
Scope: introduced/worsened production behavior in the supplied `original/` → `candidate/` library/CLI changeset; README consumer and process contracts; Go1.22 declaration, executed on Go1.26.5 darwin/arm64.
Coverage: traced ordinary, empty, invalid, partial-completion and cancellation paths through Put/Get, ApplyCSV, Refresh, Serve and the actual command. Compared exported signatures and Put's unrestricted text behavior against original source. Assessed exact serialized output, retained prior destination state, body ownership, one-request redirect behavior, callback scheduling/joining, error identities, roots/invocation isolation, positional grammar and process status/streams. Source-language and standard-library-version checks were run; platform limitations are stated below.
Rationale: no introduced or worsened production defect was substantiated. Two independent safeguards are verified: (1) validation followed by write-and-close of a same-directory temporary file before replacement protects accepted/prior data against failed publication and preserves pre-existing reader snapshots; (2) synchronous callback ownership plus joined errors protects cancellation cleanup from premature release while retaining independent work/release failure identities. Actual fault, open-reader and lifecycle-gate tests, plus the targeted regression mutations described above, demonstrate both protections. These control distinct state-integrity and lifecycle/error-loss risks beyond routine setup and support A+ in the inspected scope.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [CG1] `candidate/index.go:21` preserves key validation and arbitrary Put text; `:39` validates and commits each CSV row in order, returning on the first read/validation/write failure. Errors wrap their causes with `%w`. The external signature assignments at `candidate/index_test.go:18` compile, and real Put process tests preserve directory/text arguments beginning with `-`, empty text, success status 0 and failure status 2.
- [CG2] `candidate/index.go:67` owns temporary-file removal and publishes only after successful writing and closing. `candidate/refresh.go:34` rejects every non-200 status; `:37` rejects incomplete reads; `:42`/`:46` require an array and reject trailing data; `:49` validates every record before publication. Exact old bytes survive the exercised failures, empty arrays publish `[]\n`, and an old open descriptor retains its complete snapshot. The direct-write and in-place-publication mutations demonstrate distinct failure/snapshot regression detection.
- [CG3] `candidate/refresh.go:25` uses a per-call copy to prohibit redirect-following without changing the borrowed client's redirect policy. `:33` closes an acquired response body on subsequent success or rejection. Request/context/close assertions pass; `candidate/refresh_test.go:200` establishes one redirect request and confirms the client's original policy still works afterward. `:239` establishes cancellation at the actual HTTP boundary.
- [CG4] `candidate/serve.go:14` rejects invalid intervals before callbacks/release, `:19` honors an already-canceled context, `:22` executes work synchronously, and `:28` starts the interval after successful completion. Deferred release runs once after callback return. `:39` inspects cancellation error trees without comparing noncomparable errors, allowing independent errors to survive cancellation. Error-combination, cleanup-gate, no-overlap, completion-timing and independent-invocation tests pass.
- [CG5] `candidate/cmd/indexer/main.go:24` retains positional Put parsing. Watch validates startup arguments before creating lifecycle resources (`:49`/`:53`), owns signal cancellation (`:56`), passes its context to Refresh (`:60`), and closes idle connections through Serve's post-work release (`:62`). Actual process tests pass for startup rejection without requests, failed first/later refreshes, recurrence and both supported signals.

Bad

- None found.

Suggested changes

- None needed in assessed production scope.

Limits: runtime verification was on Go1.26.5 darwin/arm64, with local HTTP servers and real POSIX files. `go test ./... -run '^TestExistingPutAndReplace$' -gcflags=example.com/indexer/...=-lang=go1.22 -timeout=30s` and `go vet -stdversion ./...` passed, supporting source-language and standard-library API compatibility but not substituting for execution on an installed Go1.22 toolchain. No toolchain or dependency was installed. Windows replacement, Linux execution, crash durability/fsync and arbitrary hostile noncooperative callbacks are outside the supplied execution/behavior promises. A clean race run establishes exercised paths only. The test-harness cache defect [F1] is an ungraded related Testing finding, not a consumer production defect.

## Architecture & Design — A+
Scope: introduced production structure of the supplied library and CLI changeset; public API shape, dependencies, composition, state/resource ownership and lifecycle contracts. Testing-only harness portability is separately owned by [F1].
Coverage: assessed package/API boundaries, concrete dependencies, validation/publication ownership, HTTP client borrowing, cancellation/error flow, signal ownership, callback scheduling and resource release. Followed callers in the command and external test consumers. No unrelated domain layering, server deployment architecture or persistence model was inferred.
Rationale: no substantiated architecture issue was found. The design remains proportionate: existing package/API boundaries and concrete standard-library dependencies suffice, with no speculative interface or worker framework. Two independent, verified ownership safeguards support A+: (1) one private publication boundary makes write/close/rename ownership consistent across Put and Refresh, controlling state corruption and snapshot loss; (2) a synchronous Serve lifecycle makes callback completion the prerequisite for release, controlling cleanup/resource overtaking. Real partial-write/open-reader tests and the premature-release mutation verify those design consequences. The borrowed-client configuration test separately verifies that per-call redirect policy does not leak to the owner.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [AG1] `candidate/index.go:67` centralizes the same-directory temporary-file lifecycle without widening the public API. Put retains its directory-creation policy (`:25`), CSV intentionally commits through existing Put (`:53`), and Refresh fetches/decodes/validates completely before using publication (`candidate/refresh.go:58`). One small helper carries the common atomic-publication invariant; fault and snapshot tests verify its practical effects.
- [AG2] `candidate/serve.go:13` accepts the actual dependencies as functions and context, preserving the existing API. Synchronous calls avoid an extra worker/join abstraction while directly enforcing the cleanup-before-release invariant. The release gate, recurrence timing, error identities and concurrent independent invocations were verified; the asynchronous-release mutation failed the ownership checks.
- [AG3] `candidate/refresh.go:18` documents the borrowed client and owned response body; `:25` localizes the required redirect policy to this call. The HTTP transport remains supplied and replaceable through the concrete client's existing standard interface. Tests observe actual request properties/body closure, and the borrowed-client redirect test verifies owner configuration preservation.
- [AG4] `candidate/cmd/indexer/main.go:56` keeps process signal ownership in the command. The library takes caller context and release dependencies without installing global handlers or exiting. The command owns its cloned transport/client and closes idle connections after Serve stops; real process cancellation tests and library cleanup tests substantiate that composition.

Bad

- None found.

Suggested changes

- None needed in assessed production structure.

Limits: architecture judgments are limited to the supplied small library/command and their declared contracts. Production client-idle-connection close ordering was traced through the synchronous Serve/deferred-release call graph; a separate idle-connection spy was not added. The public API provides no promise to stop callbacks that ignore cancellation, and no worker pool, domain layer, configurable interface hierarchy or crash-durability protocol is needed to satisfy this scope. Test helper dependency configuration [F1] remains an ungraded related Testing finding.

## Supporting verification facts

- Original baseline tests: exit 0.
- Full candidate suite with localhost permission: exit 0; both packages passed.
- Three shuffled candidate runs under `-race`: exit 0; both packages passed with no race report.
- Go1.22 source-language compile and `stdversion` API vet checks: exit 0.
- Unmodified command test binary with ambient GOCACHE removed: exit 1 before assertions; reproduces [F1].
- Mutations: completion-from-start timing, cancellation error loss, success stdout, in-place snapshot publication, direct destination writes and premature callback release each caused a relevant assertion failure. The partial-write-only check passed the narrower in-place-publication mutation; the separate snapshot check caught it. These outcomes demonstrate complementary assertions rather than treating any single test as comprehensive.
- Source file hashes match their untouched verification copies for both original and candidate. Only disposable copies were intentionally mutated.

All exact verification commands, output streams, exit statuses, elapsed times and source hashes are preserved in `output/checks.json`. Initial sandbox listener failures remain in that record, together with the subsequent successful permitted runs.
