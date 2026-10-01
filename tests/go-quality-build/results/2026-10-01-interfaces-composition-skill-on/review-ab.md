# Independent review of two supplied changesets

Review date: 2026-10-01. These are changeset reviews, comparing each supplied original fixture with its corresponding candidate, rather than audits of unchanged debt. Reviewed the README, go.mod, source, tests, and exact supplied task prompt for each fixture, plus the Architecture & Design and Testing review skills and their decision references. No trial reports, evaluation assertions, other trials, or build skills were consulted. No candidate or repository files were modified.

Original root: `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-interfaces-and-composition/evals/files`.
Candidate root: `/private/tmp/go-quality-build-composition-eval/skill-on-r1`.
All finding locations below are relative to the candidate root. Each candidate is a small library with `go 1.22`; checks used `go1.26.5 darwin/arm64`.

# protocol-vs-mock

## Architecture & Design — A

Scope: Original `protocol-vs-mock` versus supplied candidate; report-format library adding CSV while preserving the host extension protocol.
Coverage: Public Record/Encoder/Report API, bundled CSV implementation, writer error handling, private-buffer failure behavior, external implementation example, and absence of unnecessary abstraction. No concurrency or lifecycle mechanism was introduced.
Rationale: No substantiated introduced architectural issue. The existing interface is the public extension protocol and is used directly by the new format and a host implementation; keeping concrete Report is appropriate. These are verified, ordinary good design choices supporting A, without claiming two additional independent safeguards for A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [P-A-G1] `protocol-vs-mock/report.go:15` and `:25`, plus `report_test.go:79` and `:88`: CSV and the external host encoder use the same one-method Encoder protocol; callers still construct Report literals. The executable external-package example passed.
- [P-A-G2] `protocol-vs-mock/report.go:28` and `:37`: standard encoding/csv handles quoting and buffered writes, with Flush followed by Error preserving delayed writer errors. The focused failure test passed.
- [P-A-G3] `protocol-vs-mock/report.go:41`: Report remains a concrete orchestration type with a single genuine extension dependency. No mock-only duplicate interface or package layer was added.

Bad

- None found.

Suggested changes

- None needed.

Limits: Source and supplied consumer example cover the material architectural decisions in this small change. No external host applications were supplied. Candidate tests passed in a disposable copy with `env GOCACHE=/private/tmp/go-quality-review-ab-cache go test -race -count=1 -timeout=30s ./...`. Go 1.22 itself was not executed.

## Testing — B

Scope: Same changeset, focused on added CSV behavior and preservation of external encoder compatibility and failure handling.
Coverage: Inspected every supplied test and implementation branch; executed all tests with race detection and a focused mutation of multi-record behavior. Evaluated assertions rather than relying on test names or coverage percentages.
Rationale: One moderate introduced coverage gap leaves normal multi-record CSV output insufficiently checked. Other requested behaviors have meaningful tests. This is a contained coverage gap, not effective absence of verification for the whole feature, so moderate rather than major selects B.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [P-T-G1] `protocol-vs-mock/report_test.go:23` and `:35`: exact assertions exercise commas, quotes, embedded newlines, and header-only empty output.
- [P-T-G2] `protocol-vs-mock/report_test.go:50` and `:58`: failing writer and partial-writing failing encoder exercise delayed CSV write failure and Report returning no partial bytes. Error identity is checked with errors.Is.
- [P-T-G3] `protocol-vs-mock/report_test.go:88`: the external-package example compiles and checks host-defined output through the supported public API.

Bad

- [P-T1][moderate][introduced] `protocol-vs-mock/report_test.go:25` (coverage set also includes `:37` and `:52`): CSV tests provide either one record or no records, so they do not verify that all input records are written in order as README line 3 promises. In a disposable mutation, adding `if len(records) > 1 { records = records[:1] }` immediately before the CSV loop at `report.go:32` left the entire test suite passing. A normal report containing multiple records could silently lose later rows without regression detection. Primary remediation owner: Testing.

Suggested changes

- [P-T1] Add an exact-output CSV test with at least two distinguishable records in a deliberate order, including a text value such as `007`. Verify the first-record-only mutation fails and that reordered output would fail. This checks complete ordered output through the existing public API.

Limits: The mutation demonstrates a test gap; the submitted production loop itself correctly visits all records. Candidate suite passed under race detection; mutated suite passed with `env GOCACHE=/private/tmp/go-quality-review-ab-cache go test -count=1 -timeout=30s ./...`. No long-running fuzzing was needed for a thin standard-library encoder wrapper.

# visible-dependency

## Architecture & Design — A

Scope: Original `visible-dependency` versus supplied candidate; internal receipt library changing from environment/global configuration to independently configured senders. The original README explicitly permits coordinated API changes.
Coverage: Instance configuration, actual two-client usage, request construction, caller context propagation, synchronous delivery, client ownership, body closure, status/error behavior, and concurrent use. No compatibility shim is required by the supplied contract.
Rationale: No substantiated introduced architectural issue in the requested configured-client use. A concrete borrowed *http.Client and per-instance endpoint directly support the host's existing ownership model. Relevant behavior is verified by inspection and executed tests; this supports A rather than an exceptional A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [V-A-G1] `visible-dependency/receipt.go:13` and `:19`: each Sender holds its endpoint and supplied client, without reading environment variables during Send. The two-tenant test uses distinct clients and authorization transports concurrently and passed under race detection.
- [V-A-G2] `visible-dependency/receipt.go:11`, `:33`, and `:42`: the client is explicitly borrowed, per-call context reaches request construction, and response bodies remain scoped to the synchronous call. The cancellation test passed.
- [V-A-G3] `visible-dependency/README.md:6`: usage configures distinct transports, timeouts, endpoints, and senders, matching the host composition need without introducing an unnecessary client interface.

Bad

- None found.

Suggested changes

- None needed.

Limits: The constructor retains a nil-client fallback to http.DefaultClient (`receipt.go:20`), but the requested host supplies independently owned clients and no forbidden-nil contract was supplied; this is not counted as a demonstrated tenant-isolation defect. The README statement that no process-wide settings are used is broader than that fallback permits and could be clarified as a documentation refinement. No socket-based remote service was used or required. Tests passed in the disposable copy with `env GOCACHE=/private/tmp/go-quality-review-ab-cache go test -race -count=1 -timeout=30s ./...`. No Go 1.22 executable was tested.

## Testing — B

Scope: Same changeset, with emphasis on independent tenant configuration, cancellation, and preservation of the documented HTTP request contract.
Coverage: Inspected all receipt tests, compared their assertions with the original test, ran the whole candidate suite under race detection, and ran independent request-method and endpoint-host mutations in disposable copies. Assessed isolation, goroutine coordination, cancellation timing, status assertions, and the client/transport test boundary.
Rationale: One moderate introduced/worsened request-contract assertion gap. The tests meaningfully verify tenant-specific authorization, paths, IDs, JSON escaping, content type, cancellation, and one non-2xx failure, so the important feature is not effectively unverified. Lost method coverage and incomplete endpoint assertions are grouped as one correction to the captured request contract.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [V-T-G1] `visible-dependency/receipt_test.go:29`: two concrete clients with independent authorization transports send concurrently without changing environment variables or http.DefaultClient. Captured request data ties each tenant's credentials to its path and escaped ID; both tenants must be observed.
- [V-T-G2] `visible-dependency/receipt_test.go:101`: cancellation waits for transport entry before cancellation, then checks bounded return, avoiding a guessed sleep. It exercises the context attached to the actual request passed through http.Client.
- [V-T-G3] `visible-dependency/receipt_test.go:128`: a non-2xx response must produce an error.

Bad

- [V-T1][moderate][introduced/worsened] `visible-dependency/receipt_test.go:30`, `:40`, and `:80`: the new request snapshot drops the original test's method and full-URL assertions. It captures only URL.Path, and both tenant URLs use the same host. Changing production `http.MethodPost` to `http.MethodGet` at `receipt.go:33` leaves the entire suite passing. Independently inserting `req.URL.Host = "wrong.invalid"` before `receipt.go:37` also leaves it passing. Thus a broken POST contract or sending to an unintended authority can evade the new verification even though the prior test asserted method and full URL. Primary remediation owner: Testing.

Suggested changes

- [V-T1] Capture and assert the HTTP method and full configured URL; give the tenants distinct authorities as well as paths. Keep the real http.Client/custom transport seam. Verify both the GET mutation and wrong-host mutation fail while preserving the concurrent tenant and JSON assertions.

Limits: These are test gaps, not defects observed in the submitted request construction, which still uses POST and the configured endpoint. Both mutations passed `env GOCACHE=/private/tmp/go-quality-review-ab-cache go test -count=1 -timeout=30s ./...`. Body-close assertions and complete 2xx boundary coverage were already absent in the original fixture, and the corresponding implementation behavior was preserved; they are not counted as introduced debt. Race detection only covers executed paths. Tests have finite tool-enforced suite timeouts, though some internal channel waits rely on those suite bounds.

# Checks and preservation

Disposable checks live under `/private/tmp/go-quality-build-composition-eval/review-skill-on/checks-ab/`, with separate unmodified source copies and mutation copies. Source copies contain only the fixture README, Go sources, and go.mod; trial reports and caches were not copied or read.

Initial plain `go test -race -count=1 -timeout=30s ./...` attempts were blocked by the default Go build cache being outside the writable sandbox. Re-running with GOCACHE pointing to `/private/tmp/go-quality-review-ab-cache` succeeded for both candidates. No permission escalation or candidate modification was needed. Both race-enabled candidate suites passed, and all three intentionally wrong mutation suites passed, establishing the stated regression-detection gaps.
