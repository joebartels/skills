# Independent blind review: baselines A and B

Reviewed candidate files against the supplied original fixtures, treating this as a changeset review. No evaluation assertions, author reports, build skills, or other trials were read. Paths below are relative to `/private/tmp/go-quality-build-composition-eval/review-baseline`. Reviewed the Architecture and Testing skills and their decision references. Go runtime: go1.26.5 darwin/arm64; both modules declare Go 1.22.

# Candidate A

## Architecture & Design — A
Scope: candidate-a versus original protocol-vs-mock fixture; report-format library and exported extension protocol.
Coverage: public Record/Encoder/Report compatibility, host implementation, CSV encoding/error propagation, and buffer ownership. Inspected actual external-package example and exercised it.
Rationale: No actionable architecture defect. The existing producer-owned Encoder is a genuine host extension protocol, and the implementation adds CSV directly to that protocol without adding a mock-only Report/Exporter interface. Ordinary correct composition supports A.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-AG1] `candidate-a/report.go:15` and `candidate-a/report_test.go:57` preserve the supported Encoder signature and demonstrate a host-defined implementation in package reportkit_test. Report literals still accept host implementations.
- [A-AG2] `candidate-a/report.go:27` uses encoding/csv, flushes, and returns the buffered writer error. `report.go:43` retains private rendering and discards partial bytes when a host encoder fails. Reviewer checks verified both buffered writer failures and partial-output suppression.

Bad

- None found.

Suggested changes

- None needed.

Limits: Supplied source and tests inspected; `GOCACHE=/private/tmp/go-quality-review-ab-cache go test ./...` passed. Additional disposable-copy checks passed for multiple-record order/text preservation, a large write failure, and an external encoder that writes partial output then errors. No Go 1.22 executable was run. README's statement that JSON is the only bundled format is stale after this change; documentation housekeeping, not an architecture defect.

## Testing — B
Scope: candidate-a changes and tests versus protocol-vs-mock fixture.
Coverage: CSV quoting, empty input, writer errors, host extension use, JSON regression; inspected assertions and executed tests.
Rationale: One moderate introduced gap leaves normal multiple-record CSV behavior unchecked. Existing single-record, empty, error, and host example checks are meaningful but would all pass if CSV emitted only the first record.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [A-TG1] `candidate-a/report_test.go:23` asserts exact bytes for commas, quotes, and embedded newlines. `report_test.go:35` asserts a header on empty input.
- [A-TG2] `candidate-a/report_test.go:50` verifies an underlying writer error, including the final buffered flush path. `report_test.go:68` executes a real external-package host encoder through Report.

Bad

- [A-T1][moderate][introduced] `candidate-a/report_test.go:23` uses one CSV record, and the only other CSV cases use empty input or one record. The new loop at `report.go:32` has no regression check that all records are emitted in original order. Dropping records after the first, or reversing multiple records, would escape this suite in ordinary report usage.

Suggested changes

- [A-T1] Add a CSV test with at least two distinct records and exact expected output in order, including a text value such as `001`. This detects dropped/reordered records and accidental numeric conversion. The reviewer's disposable-copy test confirmed current implementation behaves correctly; it is not part of the submitted regression suite.

Limits: Candidate tests passed with the command above. Existing Report partial-output suppression had no regression test before this change and remains untested by candidate tests; not counted as an introduced gap. A reviewer-owned external encoder check confirmed it still works.

# Candidate B

## Architecture & Design — A
Scope: candidate-b versus original visible-dependency fixture; internal receipt library whose API may change with its callers.
Coverage: host-owned client policy, tenant endpoint isolation, concurrent calls, per-call cancellation, synchronous behavior, wire format, status handling, and response-body ownership.
Rationale: No actionable architecture defect. Endpoint and client are explicit per-sender dependencies, and the caller's context reaches the HTTP request. The implementation preserves the documented HTTP and close contract without extra interfaces or background workers.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-AG1] `candidate-b/receipt.go:13` and `receipt.go:21` store each tenant's endpoint and host-owned *http.Client privately. Concurrent tenant tests verify distinct endpoints, policy transports, and correctly escaped IDs without modifying global state.
- [B-AG2] `candidate-b/receipt.go:36` binds each request to the current caller context; `receipt.go:41` waits synchronously for that request and `receipt.go:45` closes response bodies. Independent external-package tests verified active cancellation with errors.Is, all tested status boundaries, and closure on both success and rejection.

Bad

- None found.

Suggested changes

- None needed.

Limits: `GOCACHE=/private/tmp/go-quality-review-ab-cache go test -race ./...` passed. Reviewer checks in a disposable copy also passed under race detection for statuses 199, 200, 204, 299, 300, 400, 502; response closure; transport-error identity; and cancellation after transport entry. No real remote service or Go 1.22 executable was used. Host mutation of a shared http.Client while in use is outside the configured-before-use contract.

## Testing — B
Scope: candidate-b changes and tests versus visible-dependency fixture.
Coverage: in-memory request boundary, concurrent tenant isolation, escaping, active cancellation, status rejection, nil client, and synchronization failure paths.
Rationale: One moderate introduced test-lifecycle defect. Tests meaningfully exercise the requested architecture, but the transport-entry waits lack bounded failure paths and can hang after ordinary request-construction failures.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [B-TG1] `candidate-b/receipt_test.go:18` forces overlapping transport calls and checks each endpoint and decoded ID against its client policy. This is substantive concurrent execution and passed with race detection.
- [B-TG2] `candidate-b/receipt_test.go:83` cancels after transport entry and bounds the subsequent completion wait. The transport observes the actual request context rather than a separately supplied test context.

Bad

- [B-T1][moderate][introduced] `candidate-b/receipt_test.go:68`–69 and `receipt_test.go:94` wait unconditionally for transport entry. If Send returns an error before invoking the transport, the concurrent test workers finish but the test goroutine waits forever; the cancellation test likewise never reaches its timeout select. The transport release and WaitGroup wait are also unbounded. A disposable mutation making request construction fail reproduced a package timeout, instead of prompt test completion. This makes normal regressions require the global test timeout (default ten minutes) and leaves the relevant goroutines without orderly cleanup.

Suggested changes

- [B-T1] Bound entry/completion waits, observe early Send results while waiting for entry, and guarantee release/cancel cleanup. Verify by inducing an early request error in a disposable copy and confirming the test fails promptly without a package timeout.

Limits: Candidate race tests passed. The timeout reproduction ran `go test -run TestTwoSendersCanSendConcurrently -timeout 2s` in a disposable copy with request construction deliberately changed to use an invalid URL. Missing explicit candidate assertions for response closure and the full 2xx boundary predate this change; reviewer checks confirmed preservation, so these are not counted as introduced testing findings. Initial default-cache test attempts were blocked by filesystem permissions; all reported successful runs used the temporary GOCACHE. Candidates and repository were not modified.
