# Independent blind review

Reviewed candidate-a and candidate-b only against the supplied `large-features` original fixture and user contract. Candidate provenance was not inferred. No eval specifications, expected outputs, agent reports, build skill, or other trial directories were read. References below are relative to the named candidate directory. This is a changeset review of a Go 1.22 module containing CLIs and a catalog HTTP service; the supplied organizational context is separate feature owners with a shared release and independently evolving gateway HTTP protocol.

## Candidate A — Architecture & Design — A

Scope: Supplied candidate-a versus `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-package-boundaries/evals/files/large-features`; affected billing, settlement, invoice command, gateway protocol, codec, and catalog boundary.
Coverage: Inspected all supplied source and tests, concrete composition, import direction, ownership, reader consumers, sequencing, and protocol isolation. No missing surrounding application behavior was presumed.
Rationale: No actionable architecture issue. The extracted boundaries each have an actual consumer and a stated reason to change. Ordinary correct construction supports A; no claim of two safeguards beyond routine correct setup is made.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-AG1] `internal/billing/billing.go:14` defines the consumer's one-method Gateway contract; `Charge` at line 25 validates, invokes it, and persists. `internal/gatewayhttp/client.go:20` owns HTTP construction/status handling. Both commands wire the concrete dependency. Gateway protocol changes can occur without embedding HTTP policy in billing rules.
- [A-AG2] `internal/invoicecodec/codec.go:1` describes its focused billing/operations row responsibility. Settlement uses its Reader (`cmd/settle/main.go:38`); billing writes with Encode and offers ReadLedger through DecodeAll (`internal/billing/billing.go:33`, `:37`). This is demonstrated shared format reuse, with no unrelated catalog encoding mixed in.
- [A-AG3] Catalog and shop remain byte-for-byte identical to the original and import no billing code. The module remains shared. Direction is commands → billing and gatewayhttp; gatewayhttp → billing only for its compile-time contract assertion; billing and settlement → invoicecodec; catalog → standard JSON/HTTP. The assertion does not reverse the runtime abstraction.

Bad

- None found. The reader length mismatch [A-C1] below is a contained implementation contract defect; it does not require redesigning the package dependencies.

Suggested changes

- None needed for architecture.

Limits: All candidate code was inspected, and its fresh test suite passed. Package documentation explains responsibilities; there is no standalone architecture document in the supplied candidate. Assessment does not invent requirements for modules or layers outside this affected slice.

## Candidate A — Correctness & Compatibility — A-

Scope: Same supplied changeset; preservation of invoice CLI, catalog HTTP, ledger bytes, gateway request bytes, and new sequential settlement/ledger decoding behavior.
Coverage: Assessed parsing, invalid IDs and amounts, EOF/error paths, sequential partial completion, gateway failures, ledger append order, environment wiring, CLI main exit handling, and original behavior. Ran supplied tests and additional disposable-copy checks.
Rationale: One minor issue affects the narrow case of otherwise-valid IDs approximately 1 MiB or longer. Ordinary inputs, stop behavior, and existing byte contracts have positive evidence. The issue is recoverable and localized; broad state corruption or systemic failure is not established.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [A-CG1] Existing tests pass freshly, including settlement's first malformed row and first charge failure (`cmd/settle/main_test.go:20`, `:36`), byte-exact gateway requests (`internal/gatewayhttp/client_test.go:15`), successful append/decline exclusion, row round trips, and raw Export preservation.
- [A-CG2] `cmd/invoice/main.go:15` retains argument parsing, LEDGER/GATEWAY_URL wiring, five-second HTTP timeout and nonzero error exit. Gateway POST body and acceptance of only 204 match the original. Catalog source is unchanged; additional found/missing-item recorder checks passed, including exact JSON bytes and content type.

Bad

- [A-C1][minor][introduced] The new reader cannot decode all valid ledgers the existing billing interface can produce. `internal/invoicecodec/codec.go:34` sets a 1 MiB Scanner maximum at line 37, whereas Validate at line 99 and Encode at line 20 impose no corresponding ID length restriction. Original billing also accepted IDs of arbitrary length absent comma/CR/LF. A valid invoice with an ID of 1,048,576 ASCII letters and amount 1 encodes successfully, but reading those exact bytes through NewReader.Next fails with `read invoice row: bufio.Scanner: token too long`. This also reaches ReadLedger (`internal/billing/billing.go:43`). Primary remediation owner: correctness/codec implementation.

Suggested changes

- [A-C1] Use a row reader that supports the valid existing encoding range (for example, buffered ReadString with explicit EOF/error handling), and retain a regression test that decodes a successfully encoded row larger than 1 MiB. Adding an arbitrary validation limit would instead change the preserved invoice input contract and needs an explicit product decision.

Limits: `go version` was `go1.26.5 darwin/arm64`; declared language version remains 1.22. No full supported-platform matrix was supplied. Initial default-cache access failed due to sandbox permissions; rerunning with GOCACHE under /private/tmp succeeded. No real external gateway or full application was exercised. Main's process exit path was inspected rather than launched as a subprocess.

## Candidate B — Architecture & Design — A

Scope: Supplied candidate-b versus the same original fixture; CLI/service composition and affected feature/codec/gateway boundaries.
Coverage: Inspected every supplied source/test, runtime dependency wiring, consumer contract, codec reuse, ledger reader usability, catalog independence, and shared-module release fit.
Rationale: No actionable architecture issue. Existing feature cohesion is preserved and extraction addresses the two demonstrated independent responsibilities. A is supported by verified dependency separation and actual shared codec consumers.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-AG1] `internal/billing/billing.go:12` exposes the minimal Gateway contract used by billing. `internal/paymentgateway/client.go:17` owns the HTTP protocol and imports no billing or codec package; commands compose it at `cmd/invoice/main.go:27` and `cmd/settle/main.go:31`. The protocol can change independently of billing and the ledger format.
- [B-AG2] `internal/invoicerow/row.go:1` explicitly owns the shared settlement/ledger format. Encode is used by billing (`internal/billing/billing.go:24`), while Reader is used by settlement (`cmd/settle/main.go:32`) and is directly usable by export consumers on an opened ledger. A redundant billing-specific wrapper is not required to make this a billing ledger reader.
- [B-AG3] Catalog JSON/HTTP source and shop are unchanged and remain independently understandable. Import direction is commands → billing/paymentgateway/invoicerow; billing → invoicerow; paymentgateway and catalog → their standard-library dependencies. Shared release ownership is respected by retaining one module.

Bad

- None found.

Suggested changes

- None needed.

Limits: Responsibilities are documented in codec/gateway package comments and billing declarations; no separate prose deliverable was supplied. This review assesses the code's boundary explanation and structure rather than assuming missing conversational explanations. No outer application packages were provided.

## Candidate B — Correctness & Compatibility — A

Scope: Same candidate-b changeset and consumer contracts assessed for candidate-a.
Coverage: Valid/malformed rows, encoding/decoding and final unterminated row, amount overflow/positivity by inspection, invalid IDs, gateway rejection and sequential stopping, ledger ordering, existing CLI/environment/protocol behavior, catalog responses, and large row round trip.
Rationale: No actionable correctness issue found. Tests and code support the changed contracts. Correct behavior and ordinary safeguards support A; the evidence is not presented as extraordinary independent safeguards warranting A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [B-CG1] Fresh supplied tests passed, including settlement against a local HTTP server and real temporary input/ledger files (`cmd/settle/main_test.go:14`, `:42`). They verify only preceding successful rows persist and later rows are not charged after either malformed input or gateway rejection.
- [B-CG2] Billing validates/encodes before charging and appends only after success (`internal/billing/billing.go:23`). Gateway requests retain POST with exact ID,CENTS bytes and no newline (`internal/paymentgateway/client.go:18`); non-204 returns an error at line 27. Export still copies raw bytes. Invoice parsing/environment wiring and nonzero error exit are preserved.
- [B-CG3] Reader's EOF handling accepts final unterminated rows (`internal/invoicerow/row.go:58`), and codec validation rejects missing fields, invalid integer amounts, nonpositive amounts, commas and line breaks in IDs. Additional checks passed for the 1 MiB ID encode/decode round trip and exact catalog JSON/404 behavior.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same Go/runtime and external-system limits as candidate-a. Initial tests with local HTTP listeners were blocked by sandbox bind permissions; an escalated rerun passed, so that environment failure is not attributed to the code. Main exit handling was inspected rather than process-tested. CRLF and leading-plus syntax differ between the candidates, but neither is graded as a defect: the task does not establish a CRLF acceptance or digits-only textual amount contract, and original ledger writes use LF.

## Executed verification and artifacts

All test modifications were confined to disposable `check-candidate-a` and `check-candidate-b` copies under this review directory. Original candidate source and the repository were not edited.

- `rtk proxy env GOCACHE=/private/tmp/go-quality-review-cache go test -count=1 ./...` in each pristine copy: candidate-a passed; candidate-b passed after allowing local httptest listeners.
- Added `internal/invoicecodec/review_test.go` in check-candidate-a and `internal/invoicerow/review_test.go` in check-candidate-b: TestReviewLongLedgerRoundTrip encodes Invoice{ID: strings.Repeat("a", 1024*1024), Cents: 1}, then decodes and compares it. Added `internal/catalog/review_test.go` in both copies with found-item bytes/content-type and missing-item status assertions.
- `rtk proxy env GOCACHE=/private/tmp/go-quality-review-cache go test -count=1 ./internal/... -run TestReview`: candidate-a catalog passed but long-row round trip failed with Scanner token-too-long; candidate-b both checks passed.

No other defects are inferred from absent tests, hypothetical durability guarantees, or unsupported concurrency. The original billing service is explicitly serial; no redesign for concurrent use is required by this task.
