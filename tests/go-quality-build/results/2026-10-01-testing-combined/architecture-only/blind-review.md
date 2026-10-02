# Independent Go changeset review

Reviewed the supplied `original/` → `candidate/` directory diff only, against the supplied README and three supplied review skills. No author report, external expectations, other candidate, or repository workspace was inspected. All production/test line references below are relative to `candidate/` in `/private/tmp/go-independent-review-wcpvc8sf`. No candidate repairs were made. Verification and mutation copies are disposable sibling directories.

The README's “makes one GET” is interpreted as one outbound HTTP request, not merely one call to `http.Client.Do`. This is the literal action-count contract at README.md:10. Under that interpretation, F1 is a verified production violation; F2 is the independently actionable regression-test gap. F1's primary remediation owner is Correctness & Compatibility; F2's is Testing. These are two corrections, not duplicated topic counts.

## Testing — B
Scope: Supplied original/candidate Go1.22 standard-library library and CLI changeset, including the requested regression-test scope. Existing Put and helper tests are retained; added ApplyCSV, Refresh, Serve, and actual-process tests are assessed.
Coverage: Inspected every supplied source and test file; traced success, rejection, error identity, accepted-prefix effects, old/new snapshot ownership, per-invocation lifecycle state, cancellation cleanup, process streams/exits, and dependency/resource cleanup. Executed the full suite with race detection, shuffle, and three repeats. Verified two assertions with controlled production mutations. No fuzz targets or benchmarks were supplied; neither is necessary to establish these finite contract checks. Actual Go1.22 runtime execution and Windows replacement/signal behavior were not assessed.
Rationale: One moderate issue leaves a plausible HTTP rejection/request-count path unchecked. It is contained to redirected refreshes rather than making the whole suite misleading. The strong exercised state/lifecycle tests do not cancel that gap. Counts select B.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [T-G1] apply_test.go:75–105 checks retained accepted replacement, absent later rows, ErrInvalidRecord identity, and CSV error type; apply_test.go:114–150 checks read/write rejection and temporary cleanup. The corresponding package tests passed. These assertions exercise ordered partial commitment rather than an invented rollback contract.
- [T-G2] refresh_test.go:50–79 asserts exact compact file bytes, preserved record order/text, empty-array representation, a pre-opened reader's old snapshot, and one response-body close. Replacing `publish` with direct `os.WriteFile` in a disposable copy failed both snapshot cases at refresh_test.go:79. This is a verified safeguard against in-place truncation, beyond simply checking successful output.
- [T-G3] serve_test.go:87–159 holds the first callback beyond the interval and measures recurrence from its completion, while checking non-overlap and the supplied context. A disposable mutation measuring the interval from callback start failed at serve_test.go:139 (observed about 45µs versus the required 40ms). serve_test.go:185–258 gates cooperative cleanup and checks release ordering plus independent joined error identities. These tests control lifecycle timing through observable events with finite failure bounds.
- [T-G4] cmd/indexer/process_test.go:34–55 and :67–120 build/run the actual binary and assert exit code, empty stdout, appropriate stderr, exact Put/Apply state, and accepted-prefix rejection. :203–312 exercises recurring watch publication, interrupt and SIGTERM, in-flight request cancellation, failed first refresh, and failure after a successful refresh against loopback HTTP. These passed under the escalated finite race/shuffle/repeat run, rather than relying solely on command helpers.
- [T-G5] Temp directories own file state; serve cleanup cancels, unblocks held callbacks, and waits (serve_test.go:45–57); command cleanup kills/joins on failure (process_test.go:142–159); local servers are registered for cleanup (:123–131). Race detection found no exercised races. Tests use the Go1.22-compatible API set; `go vet -stdversion ./...` passed.

Bad

- [F2][moderate][introduced] The added HTTP regression scope never returns a redirect with a Location header. `responseClient` (refresh_test.go:29–39) returns headerless responses, and the rejection table (:85–105) tests 503/201 but no redirect; the real command servers likewise return final success/failure responses only (process_test.go:203–312). The full supplied suite passes while F1's finite probe demonstrates two GETs and a changed destination after a 302. Thus a normal HTTP-client path can break the explicit one-GET/status-rejection contract without a test failing. This is a regression-detection gap with contained impact, independently corrected by adding an assertion; it is not a claim that fake transports are inherently poor.

Suggested changes

- [F2] Add a redirecting transport or loopback server test with an existing destination. Assert exactly one outbound GET, a non-nil status rejection, exact prior bytes, and response-body closure. A header-bearing 302 to a valid 200 array must fail the current implementation and pass the F1 correction. One representative redirect exercises the shared policy root cause; exhaustive status enumeration is optional.

Limits: The first sandboxed full-suite commands failed solely because `listen tcp 127.0.0.1:0` was denied. The same finite suite outside the sandbox passed with `-race -shuffle=on -count=3 -timeout=60s ./...`. This does not establish arbitrary scheduler behavior or unexecuted races. Go1.26.5/darwin/arm64 was installed; Go1.22 was not installed or downloaded. Standard-version vet passed but is not a Go1.22 execution claim. Publication-error tests force rename rejection via a directory and inspect cleanup; they do not simulate every disk write/close failure. No dependencies or tools were installed. Exact commands/stdout/stderr/exit details are in checks.json.

Ungraded related finding: [F1] is the production redirect-policy violation described in the Correctness & Compatibility card.

## Correctness & Compatibility — B
Scope: Supplied original/candidate changeset implementing ApplyCSV, Refresh, Serve, and apply/watch CLI commands while retaining exported function/method types and Put/process contracts in the Go1.22 module.
Coverage: Compared exported declarations and prior implementations, inspected all supplied callers/tests and go.mod, and traced key/text validation, duplicate replacement, accepted-prefix failures, JSON shape/trailing data, atomic file publication, body ownership, callback recurrence/error/cancellation/release flow, and CLI startup/signal/error-stream behavior. Executed full race/shuffle/repeat tests and an isolated redirect reproduction. Effective Go1.22 standard-library symbol compatibility was checked with stdversion vet; actual Go1.22 and Windows execution remain limits.
Rationale: One moderate, verified contained contract mismatch occurs for an ordinary default-client redirect response. It causes an extra request and accepts/publishes a response after the specified initial operation should reject. No supported evidence establishes severe broad damage or a major lifecycle/state failure. Counts select B.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [C-G1] index.go:63–83 commits rows in order and immediately returns on parse, validation, or publication failure. It wraps errors without losing the validation sentinel; the field-count branch retains both ErrInvalidRecord and CSV parse identity. Existing arbitrary Put text still bypasses import-only nonblank validation (index.go:25–32, :86–90). Passed external-consumer tests verify independent roots, exact empty/binary text, rejected replacement retention, accepted prefix, and absent later keys.
- [C-G2] refresh.go:33–57 accepts an actual array, explicitly rejects null, checks for a second value/trailing read failure before publication, validates every record, then marshals compact JSON plus newline. `publish` writes/closes a sibling temporary before rename and defers cleanup (index.go:36–49). Passed tests verify exact destination retention on fetch/status/read/decode/validation rejection, body closure, old-reader/new-reader snapshots, and rename failure cleanup. Direct-publication mutation failure substantiates snapshot protection on this POSIX filesystem.
- [C-G3] Serve runs refresh synchronously (serve.go:21–28), so its deferred release (:20) cannot overtake callback cancellation cleanup. Interval validation precedes the release defer (:17–20); interval wait starts after callback completion; cancellation-only error trees are separated from joined independent failures (:41–63). Passed lifecycle tests verify these behaviors and both callback/release error identities, including parent deadline/cause handling.
- [C-G4] Public declaration types remain consumer-compatible, with external assignment checks at apply_test.go:15–21, refresh_test.go:16, and serve_test.go:15. Existing Put replacement tests and real binary Put/Apply stream/status checks pass. main.go:75–82 owns signal cancellation and exit 2 diagnostics; the actual watch process tests confirm recurring publication and joined shutdown for both requested POSIX signals.

Bad

- [F1][moderate][introduced] `client.Do(req)` at refresh.go:25 uses the supplied client's automatic redirect policy before the status check at :30. With an ordinary `http.Client` and a transport returning `302 Location: https://records.invalid/target` followed by a valid 200 array, Refresh issues two GETs, returns nil, and overwrites the prior destination with the redirected array. README.md:10 promises one GET and acceptance of status 200 only. The isolated `redirect-probe/independent_redirect_test.go` reproduced `calls=2 closes=2 error=<nil>` and destination `[{"key":"alpha","text":"redirect target"}]\n`; its explicit contract assertion exited 1. The command creates a default redirect-following client at main.go:58–61, so the same policy is reachable in watch. This is a contained HTTP-operation mismatch; acquired bodies do close, and there is no claim of a systemic or irreversible failure.

Suggested changes

- [F1] Enforce a no-follow policy for this one-GET operation without mutating the borrowed caller client's configuration—for example, make a shallow operation-local client copy and set CheckRedirect to return `http.ErrUseLastResponse`. Inspect/reject the resulting initial non-200 response and retain the destination. Preserve the configured transport and caller context. Verify with the F2 regression, including request count, prior bytes, status error, and body closure.

Limits: Full tests passed with `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go test -race -shuffle=on -count=3 -timeout=60s ./...` outside the sandbox after sandbox loopback failures. `go vet -stdversion ./...` and `go list -m all` exited 0; only the main module was listed. No actual Go1.22 toolchain run or Windows replacement claim is made. Other caller-custom HTTP policies beyond the supplied contract were not exhaustively modeled. The report interprets “one GET” as one actual outbound request; if redirects are intentionally allowed, that contract must be explicitly revised rather than inferred from `http.Client` defaults. Exact check evidence is preserved in checks.json.

## Architecture & Design — A
Scope: Production structure changes in the supplied Go1.22 library/CLI changeset: shared atomic publication/record validation, callback lifecycle orchestration, and command-owned HTTP client/signal composition.
Coverage: Assessed package/API boundaries, actual consumers, dependency injection, resource ownership, context/error propagation, and proportionality across all supplied production files and their tests. No additional package/interface architecture is required by this small index library plus command. Remote production deployment, broader consumers outside the supplied module, and Windows execution were unavailable; they are not assumed.
Rationale: No actionable architectural issue is established. The important ownership and API choices are clear and directly exercised by the tests. A is supported by the simple synchronous lifecycle and explicit borrowed dependency boundaries. A+ is not claimed merely for routine appropriate composition. F1 is a local HTTP-operation policy correctness defect; it can be corrected inside Refresh without changing package/API/ownership design, and is recorded below as an ungraded related issue.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `Serve` accepts the caller context and two narrow function dependencies, keeps per-invocation state local, calls refresh synchronously, and owns exactly-once release after work returns (serve.go:16–36). This satisfies the present protocol with no worker pool, background coordinator, or unnecessary interface hierarchy. The passed cancellation-cleanup and independent-invocation tests verify the ownership consequences.
- [A-G2] `Refresh` borrows the concrete caller HTTP client and closes the response body it acquires (refresh.go:20–29). The command constructs its own transport/client and supplies release to close idle connections after Serve stops (main.go:58–65). Process signals and exits stay in `main` (:75–82), leaving the library usable by independent host components. Passed real HTTP/process cancellation and library body-close tests exercise those boundaries.
- [A-G3] `publish` is a small unexported helper shared by Put and Refresh, and `validateRecord` shares the imported record rule without changing Put's accepted text (index.go:24–49, :86–90; refresh.go:48–57). No new DTO/interface/package duplicates the same model. The old-reader snapshot mutation check and arbitrary-text Put tests verify the two relevant contracts.

Bad

- None found within Architecture & Design.

Suggested changes

- None needed for package, API, dependency, or lifetime structure.

Limits: The grade covers the supplied small library and CLI rather than hypothetical future consumers. Direct idle-connection close timing was traced through synchronous Serve and the command release closure, not instrumented with an alternate command transport; in-flight request cancellation and callback-before-release ownership were exercised independently. Go1.26.5/darwin/arm64 was the execution environment. No persistent background subsystem, external service, new dependency, or platform abstraction was introduced.

Ungraded related finding: [F1] automatic redirect following violates the concrete Refresh action-count/status contract. Primary owner: Correctness & Compatibility. The targeted local request policy correction does not require architectural redesign.

# Supporting verification facts

- All checks used disposable copies. `candidate/` matches the initial untouched `verification/` copy byte-for-byte for every supplied candidate file. A SHA-256 artifact manifest is saved as source-manifest.json.
- Go: go1.26.5 darwin/arm64, module directive `go 1.22`. No toolchain/dependency downloads were requested. All verification Go commands used GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local.
- Sandboxed `go test -timeout=60s ./...`: exit 1, library passed; watch HTTP integration failed because loopback bind was denied.
- Sandboxed `go test -race -shuffle=on -count=3 -timeout=60s ./...`: exit 1 for the same environmental loopback limit; library passed.
- Finite outside-sandbox `go test -race -shuffle=on -count=3 -timeout=60s ./...`: exit 0; library and command packages passed (reported 2.502s and 5.129s respectively), stderr empty.
- `go vet -stdversion ./...`: exit 0, stdout/stderr empty. This checks supported standard-library symbol versions against the directive, not a complete Go1.22 runtime matrix.
- `go list -m all`: exit 0; stdout `example.com/indexer` only.
- Single-GET redirect contract probe: exit 1 by its explicit assertion; two requests, both bodies closed, nil Refresh error, and redirected destination bytes reproduced.
- Direct-publication mutation: targeted snapshot test exit 1, both cases show the old opener reading new bytes.
- Callback-start interval mutation: targeted lifecycle test exit 1, premature recurrence assertion at serve_test.go:139.
- Full exact command/cwd/stdout/stderr/exit/elapsed records are in output/checks.json; intentional probe/mutation failures are labeled and are distinct from baseline-suite failures.
