# Breaker recovery review — initial-c

This is a review of the completed code area, not an introduced-change review. The authoritative contract is `/private/tmp/go-client-campaign/reviews/initial-c/breaker-recovery/source/README.md`. References beginning `source/` are relative to that breaker-recovery directory. I inspected the complete implementation, all three test files, module files, `checks.json`, and `controller_probe_test.go.txt`, plus the relevant pinned gobreaker and Go HTTP source. I did not inspect other trials, author reports, campaign outcomes, or proposed skills.

## Correctness & Compatibility — C

Scope: The full `source/` library, including `New` and `(*Client).Fetch`; module `example.invalid/breaker-recovery`, Go 1.22.0, gobreaker/v2 v2.4.0. Existing-in-scope behavior is graded against the supplied README; no prior revision or broader compatibility policy was supplied.

Coverage: Request construction, successful and non-200 responses, response-body errors/limits/ownership, one application request, borrowed client configuration, context budget, dependency keys, consecutive health accounting, recovery admission/release, reopening, generation isolation, concurrency, and public function signatures. All source files supplied for this code area were assessed.

Rationale: One contained major issue breaks the explicitly promised cancellation exclusion in both failure history and recovery. One independent moderate issue violates the diagnostic-size bound. The major is contained to this client's health accounting rather than demonstrated systemic or irreversible failure, selecting C.

Finding counts: critical=0, major=1, moderate=1, minor=0

Good

- [G1] `source/client.go:49` derives a 250ms context from the parent, preserves context values, and cancels owned scope on every return. `TestFetchTotalBudget` checks headers, body reads, and an earlier parent using real HTTP; controller runs passed. `TestFetchReadFailureOwnsBodyAndContext` checks partial data/error propagation and scope closure after body closure.
- [G2] `source/client.go:96` shallow-copies the supplied client, retains its transport/jar/timeout, and applies `ErrUseLastResponse` only to the copy. `TestFetchDoesNotFollowRedirectOrMutateClient` verifies one redirecting request and subsequently exercises the original redirect callback successfully. External compile-time function-value checks preserve both public signatures.
- [G3] `source/client.go:61` scopes state by scheme and authority, with a synchronized per-client map. `TestBreakerDependencyScope` proves path/query sharing and independent authority, scheme, port, and Client state. The pinned breaker's generation check ignores old admissions; tests exercise late 200 and 503 completions.
- [G4] `source/client.go:102` closes owned bodies, returns exact `HTTP <status>` failures for non-200 responses, and limits successful data to 4096 bytes. Body/status tests assert closure and the byte limit rather than just success.

Bad

- [F1][major][existing-in-scope] Cancellation accompanying a successful 200 is recorded as healthy. `source/client.go:64` only applies parent-cancellation eligibility to `statusError`; `source/client.go:83` unconditionally makes nil errors non-excluded. Thus a body read that returns available bytes successfully while the parent becomes canceled still reaches gobreaker's ordinary success path. A deterministic response-body reproduction cancels the parent before body consumption finishes and returns `ok` successfully. After two eligible 503s, this completion resets history: following the third eligible 503, another request contacts the dependency (`calls=4->5`, `HTTP 503`) instead of receiving `ErrOpenState`. In half-open state, the same canceled completion establishes recovery; an excluded 400 and a subsequent eligible 503 then leave traffic admitted (`calls=6->7`) instead of reopening. This defeats the explicit exclusion/recovery contract. Primary remediation owner: Correctness & Compatibility; also assessed under Resilience as the same root cause.
- [F2][moderate][existing-in-scope] The 4096-byte diagnostic bound only holds for the new status error, not all Fetch diagnostics. `source/client.go:56`, `source/client.go:100`, and `source/client.go:106` return parser, HTTP-client, and body-read errors unchanged. A long malformed endpoint produces an 8245-byte diagnostic; a valid long endpoint with an ordinary short transport failure produces 8233 bytes through `url.Error`; a body reader error produces 8192 bytes. Both enabled and disabled modes reproduce this. The caller-visible bound is broken and callers cannot rely on bounded error logging/storage. These paths share one missing output-boundary policy. Primary remediation owner: Correctness & Compatibility; also assessed under Resilience as the same root cause.

Suggested changes

- [F1] Classify parent cancellation for every completion before health accounting, including otherwise successful 200s. Use the pinned library's exclusion mechanism with a distinct internal excluded outcome when necessary, preserving the chosen caller-visible result policy. Verify both closed-history preservation and half-open admission release/reopening for canceled successful reads; retain the existing 400, transport, canceled-error, and stale-generation coverage.
- [F2] Apply a bounded diagnostic presentation policy to parse, transport, and read errors at Fetch's return boundary, retaining useful error identity/cause access. Verify long endpoints and long body-reader errors in both modes while retaining exact status-error text and cancellation/read error checks.

Limits: Results apply to the supplied complete code area. No broader consumers, release matrix, or supported behavior for caller mutation of HTTP configuration or non-cooperative transports was supplied. Full controller checks and the targeted reviewer checks below passed; the new reproduction fails against unchanged copied implementation as expected. No claim is made that passing checks establish unexercised behavior.

## Observability & Resilience — C

Scope: The same bounded HTTP library and per-dependency breaker; existing-in-scope README reliability and diagnostic contracts.

Coverage: Caller budgets/cancellation, failure classification, dependency isolation, open rejection, single recovery admission, excluded admission release, failed recovery reopening, stale completions, disabled mode, and returned errors. As a library with no supplied telemetry/lifecycle contract, service metrics, distributed traces, readiness, and shutdown are not applicable requirements here.

Rationale: [F1] is one contained major reliability-contract violation: canceled successful completions can clear the history or remove recovery containment. [F2] is one moderate diagnostic-bound violation. They select C; the same IDs are shared with Correctness and must not be added together as separate production defects.

Finding counts: critical=0, major=1, moderate=1, minor=0

Good

- [G5] The pinned breaker provides synchronized admission and generation tracking; `MaxRequests: 1`, `Timeout: 100ms`, and a three-consecutive-failure trip rule are configured directly at `source/client.go:77`. Channel-coordinated tests verify a held recovery request rejects concurrent calls without transport contact, an excluded error releases admission, a replacement 503 reopens, and healthy recovery starts fresh failure accounting.
- [G6] Parent/total budgets reach the HTTP request and body read, with real HTTP timeout tests in supplied passing checks. Fetch has no application retry loop and explicitly prevents redirect following. Disabled mode uses the same single bounded fetch and does not trip or reject traffic.
- [G7] Non-503 status/transport/read errors are excluded through the pinned library, and a canceled 503 is separately marked ineligible. Tests confirm exclusions preserve existing failure history rather than counting success. Breaker rejections and HTTP errors remain directly observable to callers.

Bad

- [F1][major][existing-in-scope] Cross-reference the confirmed canceled-200 health-accounting defect above (`source/client.go:64`, `source/client.go:83`). An excluded canceled recovery completion instead closes containment and makes a subsequent single failing probe insufficient to reopen. This breaks the stated reliability contract; one important client's health boundary is affected, with no evidence of systemic reach.
- [F2][moderate][existing-in-scope] Cross-reference the confirmed unbounded diagnostics above (`source/client.go:56`, `source/client.go:100`, `source/client.go:106`). Long endpoints or reader errors escape the library's diagnostic bound in both modes, creating predictable excess diagnostic output for its consumers.

Suggested changes

- [F1] Apply cancellation exclusion to successful completions as well as status/error completions, and verify that an excluded 200 releases the recovery slot without closing the breaker.
- [F2] Bound returned diagnostic rendering uniformly while preserving actionable causes and exact HTTP status reporting.

Limits: No deployment, operator SLO, instrumentation configuration, or enclosing service was supplied. The review does not demand service telemetry or failure controls beyond the owned library boundary. Supplied controller checks, independently rerun focused tests/race check, and preserved reproductions are detailed below.

## Testing — C+

Scope: All supplied implementation tests and public API compile checks, plus controller probes/raw checks for the same complete library code area.

Coverage: Success/status/body limits and closure; caller configuration/redirect behavior; real HTTP header/body/parent deadlines; dependency scope; threshold/reset/exclusion behavior; canceled 503 classification; recovery exclusivity, release, reopening and healthy reset; stale 200/503 generation behavior; relevant exercised concurrency under race detection. Tests were read for assertions and lifetime ownership, not graded by count or coverage percentage.

Rationale: Two independently actionable moderate coverage gaps allow the confirmed narrow successful-cancellation branch and diagnostic-output limit violations to pass. Cancellation errors/503s and the rest of recovery are meaningfully verified, so the omission is not graded as leaving the entire important recovery contract effectively unverified. No major or critical testing issue is demonstrated; two moderate gaps select C+.

Finding counts: critical=0, major=0, moderate=2, minor=0

Good

- [G8] Tests assert exact data/status/error results, byte limits, transport counts, and body-close counts (`source/contracts_test.go:56`, `source/contracts_test.go:91`). The body-read error test checks partial data and `errors.Is`, protecting useful failure behavior.
- [G9] Real `httptest` requests exercise the redirect and timeout boundaries rather than faking HTTP context enforcement (`source/contracts_test.go:115`, `source/contracts_test.go:153`). The independent provided controller checks report all package tests and vet passing on both toolchains, plus race detection.
- [G10] Recovery and generation tests use explicit start/release channels, bounded waits, cleanup cancellation, and atomic transport counts (`source/contracts_test.go:332`, `source/contracts_test.go:412`). The 110ms wait is for the specified elapsed open period, not a guessed asynchronous completion. Reviewer Go 1.22 targeted normal/race runs also pass.
- [G11] `source/api_test.go` checks exported function values as an external consumer, including the full method expression, rather than relying only on ordinary call-site compilation.

Bad

- [T1][moderate][existing-in-scope] Successful cancellation accounting is absent. `source/contracts_test.go:296` covers cancellation with 503; the recovery cancellation case at `source/contracts_test.go:344` forces a context error. Controller canceled-probe coverage also returns a context error. None tests a successful buffered body read while caller cancellation is visible before accounting, so they all pass with [F1]. Both history preservation and excluded successful recovery need independent assertions, as the preserved reproduction demonstrates.
- [T2][moderate][existing-in-scope] Diagnostic bounds are not asserted on failure paths. `source/contracts_test.go:91` checks result size and short status text; the read failure at `source/contracts_test.go:56` uses a short sentinel, and there are no long malformed/valid endpoint cases. The suite consequently passes while ordinary transport/parser errors and reader diagnostics exceed 4096 bytes ([F2]).

Suggested changes

- [T1] Add deterministic successful-body cancellation cases for closed history and half-open admission. Assert no reset, no recovery, release of the slot, and one eligible failed replacement probe reopening. Keep observable transport-count checks so a false success cannot hide behind internal-state assertions.
- [T2] Test the diagnostic bound using long malformed and valid endpoints and a long body-read error in both enabled/disabled modes. Assert maximum rendered length alongside preserved status text and required error identity/cause behavior.

Limits: No CI configuration or repository-wide testing policy was part of this code area. Listener-based tests were inspected and supported by controller raw results; the reviewer independently reran the fake-transport/body-focused subset, including race detection. Tests/reproductions do not model adversarial transports that ignore context, or unsupported concurrent mutation of the borrowed client.

## Unnecessary machinery and optional refinements

No substantial unnecessary framework or process-global machinery was found. The implementation delegates synchronization, recovery admission, exclusion accounting, and generations to the pinned dependency, keeping its own state to a dependency map and status eligibility. `New` does allocate an empty breaker map when disabled (`source/client.go:42`); lazy allocation would be an optional small simplification, not a graded issue. Error-size correction can be made at the output boundary without adding another breaker state machine.

## Verification evidence

Controller-executed evidence: `checks.json` contains passing public package tests, vet, race tests, Go 1.22 package tests/vet, and private controller-probe package/race/Go 1.22 tests. I inspected each recorded command, exit code, and raw output. The probes focus on excluded canceled recovery, generation isolation, preserved failure history, dependency independence, and disabled behavior; their success does not cover [F1]/[F2].

Reviewer check environment for every Go command:

```text
GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off
```

Executed against the unchanged supplied source:

```text
rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off /private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -run 'TestBreaker|TestFetch(ReadFailureOwnsBodyAndContext|StatusAndBodyLimit)' -count=1 -timeout=20s ./...
```

Exit 0, `ok example.invalid/breaker-recovery 1.380s`. Raw output: `initial-c-reproductions/raw-go122-source-tests.txt`.

```text
rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off /private/tmp/go-client-campaign/tools/go/bin/go test -race -ldflags=-linkmode=external -run 'TestBreaker|TestFetch(ReadFailureOwnsBodyAndContext|StatusAndBodyLimit)' -count=1 -timeout=20s ./...
```

Exit 0, `ok example.invalid/breaker-recovery 2.510s`; the linker also emitted a platform `LC_DYSYMTAB` warning, preserved verbatim. Raw output: `initial-c-reproductions/raw-go122-source-race.txt`.

Disposable reproductions: `initial-c-reproductions/breaker-recovery/` preserves copied `client.go`, `go.mod`, `go.sum`, and `review_reproduction_test.go`. Original implementation/tests/modules/probes and Git state were not changed.

```text
rtk proxy env GOWORK=off GOTOOLCHAIN=local GOPATH=/private/tmp/go-client-campaign/gopath GOMODCACHE=/private/tmp/go-client-campaign/modcache GOCACHE=/private/tmp/go-client-campaign/cache GOPROXY=off /private/tmp/go-client-campaign/tools/go/bin/go test -ldflags=-linkmode=external -run TestReview -count=1 -timeout=10s -v ./...
```

Exit 1 as expected: both successful-cancellation tests fail their behavioral assertions; all six diagnostic-bound subtests fail, showing 8245/8233/8192 bytes for parse/transport/body errors in both modes. Full output: `initial-c-reproductions/breaker-recovery/raw-go122-test.txt`. This is a reviewer-added failing reproduction, not a failure of the supplied tests.
