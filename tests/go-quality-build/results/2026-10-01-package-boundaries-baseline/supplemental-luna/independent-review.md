# Independent blind review

Reviewed the supplied original and changed directories as a changeset. Did not read evals.json, source-audit.md, proposed build skill material, earlier reviews or baseline outputs, or trial-report.md. Applied the installed Architecture & Design and Correctness & Compatibility skills and both decision references. The source diff listed trial-report.md as an added filename; its contents were not opened. All locations below are relative to `/private/tmp/go-quality-build-eval.TijGDa/baseline-luna/`.

## medium-service: Architecture & Design — B

Scope: Original fixture versus changed medium-service; Go 1.22 module, shared service, HTTP server and replay CLI.
Coverage: Both ingress paths, operation sequencing, concrete persistence and outbound HTTP, configuration, abstraction ownership and independent evolution.
Rationale: One moderate architecture issue. Interfaces make dependencies replaceable, but the independently maintained persistence and notification implementations still reside inside the reusable operation package, and the operation still owns the storage bytes. The supplied independent maintenance/evolution requirement makes this a material boundary concern rather than a general rule that small services need layers.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [M-G1] `dispatch.go:83` and `dispatch.go:101` both invoke Submit; neither consumer duplicates validation, persistence ordering or notification.
- [M-G2] `dispatch.go:23` and `dispatch.go:28` define small consumer protocols with only the methods Submit uses. No retries, queue, or speculative framework was added.

Bad

- [M-A1][moderate][introduced/incompletely addressed] `dispatch.go:32`, `dispatch.go:38`, `dispatch.go:67`, and `dispatch.go:69`: independent file storage and notification protocol implementations remain in dispatch, with Submit constructing their defaults and supplying the literal storage encoding `queued\n`. A storage encoding/configuration or notification protocol revision therefore edits the same package that owns validation and both ingress behaviors. The new interfaces permit substitution but do not establish the independent implementation ownership requested by this task. Primary remediation owner: Architecture.

Suggested changes

- [M-A1] Move concrete file and HTTP implementations into narrowly scoped packages, compose them in the commands, and let the storage implementation own the record representation. Preserve the shared Submit operation and the existing HTTP/status/byte contracts. Verify with the existing replay and compatibility checks. This requires only concrete adapters and composition, not additional domain layers.

Limits: No future protocol specification was supplied; the finding concerns demonstrated ownership and present default coupling, not hypothetical protocol details. Architecture does not count the reader boundary defect below a second time.

Package proportionality: The one-method interfaces and reusable operation are proportionate. Keeping every implementation and file ingress behavior in the original package under-delivers on the explicit independent-maintenance signal. The added replay command is warranted.

## medium-service: Correctness & Compatibility — A-

Scope: Same changeset and Go 1.22 declared version; tested with installed Go 1.26.5.
Coverage: Existing Submit ordering and stored bytes; unchanged HTTP status/body source; replay trimming, blanks, sequential processing, invalid IDs, notification failures, command exit codes, and long trimmed input.
Rationale: One minor, narrow supported input failure; normal inputs and required failure behavior work. No evidence of broader failure.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [M-G3] Supplied tests pass and executable CLI probes confirm ordered notifications `a`, `b`, exact `queued\n` records, exit 1 at invalid ID with no later record, and exit 1 on notification failure while retaining the first record.
- [M-G4] `dispatch.go:78` retains the original HTTP handling: non-POST is 405, Submit error is 400 with `submission failed\n`, success is 204. Validation and default notification implementation preserve the original behavior by source comparison.

Bad

- [M-C1][minor][introduced] `dispatch.go:93`: default bufio.Scanner limits the raw line before trimming. A line of 70,000 spaces followed by `a` and LF represents the valid ID `a` under the requested trimmed-line contract, but Replay returns `read replay file: bufio.Scanner: token too long` without submitting it. The same limit rejects a sufficiently long whitespace-only line that should be ignored. Primary remediation owner: Correctness.

Suggested changes

- [M-C1] Use a line-reading implementation without this implicit raw-line cap, or an explicitly agreed limit consistent with the contract. Verify a padded valid ID and a long blank line followed by a valid ID. Merely increasing Scanner's finite maximum leaves an undocumented contract limit.

Limits: HTTP status/body compatibility was source-inspected rather than exercised through the actual server binary. No platform matrix or race check was run; the task requests sequential replay. Cache/socket sandbox restrictions initially blocked full tests, then approved elevated execution succeeded.

## large-features: Architecture & Design — A

Scope: Original fixture versus changed large-features; one Go 1.22 module with catalog, billing, invoice/settle CLIs and concrete gateway transport.
Coverage: All changed source and tests; catalog source and dependencies; shared codec users; gateway ownership and command composition.
Rationale: No confirmed architecture issue. Actual reuse and independent ownership justify the two added packages without an unnecessary module split.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [L-G1] `internal/billing/billing.go:13` owns the minimal Gateway protocol. `internal/gateway/http.go:17` owns request construction/status interpretation. Commands compose these at `cmd/invoice/main.go:27` and `cmd/settle/main.go:28`; billing no longer imports net/http or concrete gateway configuration. Protocol changes have a concrete separate home.
- [L-G2] `internal/invoice/row.go` is shared by settlement input, billing ledger encoding and billing.ReadLedger. Catalog remains unchanged, uses JSON directly, and imports neither invoice nor billing. There is real reuse without forcing catalog through a generic serialization package.
- [L-G3] Imports are commands → billing/gateway/invoice; billing → invoice; gateway → standard library; invoice → standard library. One module fits the supplied shared release model.

Bad

- None found.

Suggested changes

- None needed for architecture. Correct the codec behavior finding below within its existing boundary.

Limits: Future export callers are not supplied, so callback-based ReadLedger is judged against the stated reader need only. No assertion that this API serves unspecified future query requirements.

Package proportionality: The gateway transport and invoice codec extractions are proportionate to independently evolving transport and three actual codec uses. Keeping catalog separate and retaining one module fit the ownership/release facts.

## large-features: Correctness & Compatibility — B

Scope: Same changeset; installed Go 1.26.5, declared Go 1.22. Internal Service construction changes are exercised by every supplied caller; no external source compatibility promise was supplied.
Coverage: Successful and invalid codec rows, ledger encoding/reading, charging order, gateway request bytes and failure handling, settlement CLI exit/partial completion, existing catalog source and invoice CLI source compatibility.
Rationale: One moderate codec round-trip/reader contract failure for valid long IDs; ordinary settlement paths and preserved wire/ledger bytes are verified.
Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [L-G4] Supplied tests all pass, including actual HTTP gateway POST body `invoice-1,250`, validation before charging, and no ledger append after gateway failure.
- [L-G5] Executable settlement probes confirm success sends `a,1`, then `b,2` and writes `a,1\nb,2\n`; malformed second row stops after first charge with exit 1; gateway failure exits 1 without ledger bytes or later calls.
- [L-G6] `internal/invoice/row.go:61` retains the existing invoice validation (nonempty ID, no comma/CR/LF, positive amount). The gateway encoder retains the prior no-newline request body; ledger encoding retains LF. Catalog is unchanged by the diff.

Bad

- [L-C1][moderate][introduced] `internal/invoice/row.go:48`: Read uses Scanner's default token limit, while Encode and the existing invoice CLI accept IDs without a length cap. Confirmed: Encode accepts a 70,000-character ID and produces 70,003 bytes, but Read of those exact bytes invokes no callback and returns `bufio.Scanner: token too long`. Therefore the new billing.ReadLedger cannot decode all records the billing path can legitimately write, and settlement rejects those otherwise valid rows. This is a contained supported-data mismatch; no corruption or systemic failure was demonstrated. Primary remediation owner: Correctness.

Suggested changes

- [L-C1] Make row reading accept every encoding the existing billing path supports, preferably reading lines without the implicit Scanner cap. Keep existing valid invoice IDs compatible; add a long Encode → Read/ReadLedger regression case and exercise the settlement route.

Limits: No end-to-end old/new invoice binary comparison or catalog server run was needed to establish the unchanged source/protocol paths; those compatibility claims combine source comparison with gateway and ledger behavior checks. Blank rows correctly fail as missing fields. No atomic charge/ledger transaction, retries or exactly-once delivery was promised, so gateway success followed by ledger failure was not invented as a new finding. No platform matrix or race run was performed.

## Verification record

Source and configuration were unchanged. Disposable copies and reproduction tests are under `/private/tmp/go-quality-build-eval.TijGDa/luna-review-checks/`; trial-report.md was excluded from copying. In each case directory, ran `rtk proxy go test -count=1 -v ./...`; after initial sandbox cache/socket failure, elevated execution exited 0 for both. These runs contain all supplied tests plus focused reproduction tests that assert the observed limit is present, not that the desired behavior passes.

Reproductions:

- medium `TestReviewLongTrimmedLine`: valid trimmed ID `a` → `read replay file: bufio.Scanner: token too long`.
- large `TestReviewReadOwnEncoding`: `Encode bytes=70003 Read callbacks=0 error=bufio.Scanner: token too long`.

Ran `rtk proxy python3 /private/tmp/go-quality-build-eval.TijGDa/luna-review-checks/cli_check.py` with approved local socket/cache access. It builds both new commands and checks successful, malformed/invalid, and upstream failure cases against a local HTTP server; all six observations are detailed above and the harness exited 0.

Deduplicated total: 3 findings: moderate=2 (M-A1, L-C1), minor=1 (M-C1); major=0, critical=0. No trial implementation explanation was read before reaching these findings.
