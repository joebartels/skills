# Independent supplemental assessment — m06 and m07

This supplemental scope was accepted and inspected after the initial six-module outputs were saved and frozen. Original contracts remain `base/README.md` plus `task.txt`. The initial report, summary and checks are unchanged.

No actionable introduced defect was substantiated. Code Quality & Go Idioms and Correctness & Compatibility each receive A for m06 and m07; Security is Not applicable in this scoped review. Grades are separate topic judgments, without averaging.

## m06

Scope: Supplied changeset m06/base -> m06/module; authoritative m06/base/README.md plus m06/task.txt; module example.com/entryzip, go 1.22, standard-library only, no build tags.

### Code Quality & Go Idioms — A

Scope: Supplied changeset m06/base -> m06/module; authoritative m06/base/README.md plus m06/task.txt; module example.com/entryzip, go 1.22, standard-library only, no build tags.
Coverage: Named error result, deferred ZIP completion, safe handling of arbitrary comparable/noncomparable caller errors, private output tracker, exact new API and source-compatible old API.
Rationale: No introduced actionable issue was substantiated. Relevant safeguards and supported behavior were verified by source inspection and independent checks; ordinary correct setup supports A without claiming A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] m06/module/archive.go:27: one deferred archive.Close handles every ordinary return; the main loop filters before entry creation and has direct error guards.
- [G2] m06/module/archive.go:28 and 57: the private archiveOutput.failed tracker earns its small amount of state. Standard ZIP buffering repeats a single output failure during Close; tracking the underlying failure keeps the original error object/type without comparing arbitrary error values or using reflection.
- [G3] m06/module/archive.go:17: WriteArchive keeps its exact signature and delegates to WriteSelected with nil predicate. Entry fields are unchanged and only the required WriteSelected API was added.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.

### Correctness & Compatibility — A

Scope: Supplied changeset m06/base -> m06/module; authoritative m06/base/README.md plus m06/task.txt; module example.com/entryzip, go 1.22, standard-library only, no build tags.
Coverage: Nil/non-nil predicate, filter-before-create/read, selected names/order/contents and duplicate names, valid empty archive, central-directory completion, creation/copy/completion failures, lone identity, independent causes, stopped later-body use and borrowed output/body ownership.
Rationale: No introduced actionable issue was substantiated. Relevant safeguards and supported behavior were verified by source inspection and independent checks; ordinary correct setup supports A without claiming A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] m06/module/archive.go:43: rejected invalid-name and nil-body entries are skipped before ZIP Create/io.Copy; independent round trips verify selected duplicate names, payloads and original order.
- [G2] m06/module/archive.go:31: a previously observed output error is treated as a repeated sticky cause; independent standard ZIP checks reproduce primary == Close error from one underlying Write. Completed-module tests preserve direct noncomparable caller error identity during both creation and body-copy output failure.
- [G3] m06/module/archive.go:38: when a body fails before the output has failed, a new completion error is joined and errors.Is/As can inspect the independent body/completion causes. Independent checks cover body-only, completion-only, both, successful central directory, empty completion failure and creation plus completion failure.
- [G4] m06/module/archive.go:30: completion is syntactically invoked exactly once in one defer; there is no Close on caller output/body readers. Independent observed resources remain open on success and error paths.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.

### Security — Not applicable

Scope: Supplied changeset m06/base -> m06/module; authoritative m06/base/README.md plus m06/task.txt; module example.com/entryzip, go 1.22, standard-library only, no build tags.
Coverage: The reviewed ZIP error/filtering change has no supplied untrusted extraction path or backend diagnostic boundary.
Rationale: No broader exposure inference was requested or needed; no security finding is substantiated within this write-only library scope.

Limits: All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.

Testing observations (ungraded): The supplied consumer tests protect function-value compatibility, filter-before-create behavior using an invalid ZIP name, central-directory readability, ownership, independent failures and noncomparable error identity. Large deterministic random input makes output failure happen during copy instead of only at final flush. Independent tests separately verify stdlib sticky-error behavior, lone identity, new completion failure after body failure, exact selected contents/order and creation plus completion. No test result is used to override the original requirements.

Considered behaviors declined

- **Always errors.Join(processing, completion), including repeated sticky output failure:** That would wrap a lone caller failure and break the explicit direct-identity contract. The small output tracker is necessary evidence-based complexity, not gratuitous infrastructure.
- **Compare arbitrary errors directly or deduplicate by diagnostic text:** Backend/body/output errors may be noncomparable and different failures can share text. The implementation records the actual output failure stage instead.
- **Close borrowed output or bodies, promise rollback/durability, sort or deduplicate ZIP entries:** The original contract explicitly keeps resources borrowed, allows incomplete output and preserves selected names/content/order. Those changes would violate or expand the task.
- **Add writer factories, compressor registry, global hooks, completion interfaces or public error taxonomies:** The supplied API requires only WriteSelected; standard ZIP Writer plus the local output-failure bit covers the specified behavior.
- **Require extraction path sanitization or arbitrary nil/panicking reader/predicate recovery:** This is a write-only ZIP library with caller-provided names/readers/predicate. No untrusted extraction or recover-from-caller-panic contract is supplied.
- **Demand direct io.ErrShortWrite identity for a writer returning n < len(p) with nil error:** That intentionally violates io.Writer and supplies no caller error identity. The entryzip contract specifically preserves lone caller-provided failures, not an added policy for errors synthesized from a nonconforming writer. Valid short writes with an actual caller error follow the tracker path.

## m07

Scope: Supplied changeset m07/base -> m07/module; authoritative m07/base/README.md plus m07/task.txt; module example.com/lineexport, go 1.22, standard-library only, no build tags.

### Code Quality & Go Idioms — A

Scope: Supplied changeset m07/base -> m07/module; authoritative m07/base/README.md plus m07/task.txt; module example.com/lineexport, go 1.22, standard-library only, no build tags.
Coverage: Small delegated Export/ExportLimit flow, named return with deferred Close, conditional joining, negative-value context, short-write handling, original CLI structure and public API size.
Rationale: No introduced actionable issue was substantiated. Relevant safeguards and supported behavior were verified by source inspection and independent checks; ordinary correct setup supports A without claiming A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] m07/module/export.go:23: named-result finalization keeps single failures unchanged and joins only when an operation and Close both fail; nil success is a literal nil error interface.
- [G2] m07/module/export.go:12: the original Export function type is unchanged; the exact required ExportLimit signature adds the limit without an options or public sentinel mechanism.
- [G3] m07/module/export.go:38: one loop condition expresses unlimited/bounded iteration; failure guards preserve the accepted count and keep the success path clear.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.

### Correctness & Compatibility — A

Scope: Supplied changeset m07/base -> m07/module; authoritative m07/base/README.md plus m07/task.txt; module example.com/lineexport, go 1.22, standard-library only, no build tags.
Coverage: Unlimited/zero/positive/negative limits, empty/unterminated records, accepted counts and partial output, bytes delivered with errors, direct single-error identity, joined completion causes/types, Close once, borrowed reader, record-scan order, positional CLI paths, status/streams and validation file effects.
Rationale: No introduced actionable issue was substantiated. Relevant safeguards and supported behavior were verified by source inspection and independent checks; ordinary correct setup supports A without claiming A+.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] m07/module/export.go:24: independent checks verify lone read/write/Close identity and original typed causes, joined operation/Close causes and retained counts; all returns close the owned writer once and leave the reader open.
- [G2] m07/module/export.go:34 and 47: negative input rejects before Read/Write, and short writes do not count as accepted records. Independent ownership/limit checks verify these paths and record-aligned read ordering.
- [G3] m07/module/cmd/lineexport/main.go:12: paths remain positional, including leading dashes; executed subprocess tests verify success 0 with no streams, operational errors 1 with stderr, invalid-argument status 2 and validation before file changes.

Bad

- None found.

Suggested changes

- None needed.

Limits: All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.

### Security — Not applicable

Scope: Supplied changeset m07/base -> m07/module; authoritative m07/base/README.md plus m07/task.txt; module example.com/lineexport, go 1.22, standard-library only, no build tags.
Coverage: No added protected-data boundary or backend diagnostics are supplied.
Rationale: No security exposure was inferred from normal locally controlled CLI I/O or retained Scanner behavior.

Limits: All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.

Testing observations (ungraded): Function-value checks verify exact API compatibility. Supplied library tests exercise counts, partial writes, read-after-limit behavior, errors with returned bytes, typed completion causes, joined validation/Close failures and reader ownership. Executed real CLI tests cover status, stdout/stderr, dash-prefixed paths and output effects. Independent tests verify typed read error delivered with final unterminated bytes, direct operation identity, completion independence and negative-before-I/O behavior.

Considered behaviors declined

- **Call CLI negative LIMIT status 2 a defect because a proposed expectation uses 1:** The original contract explicitly gives invalid usage status 2 and does not say a negative out-of-domain limit is an operational failure. The implementation rejects it before opening files. A status-by-convention argument or ambiguous probe cannot establish broken compatibility.
- **Require exact borrowed-reader byte position, no Scanner buffering or post-limit error reporting:** The original implementation uses Scanner and its existing raw-byte read-ahead behavior. The implementation writes before advancing to another Scan and stops scanning on a reached limit. The wording about reading the next record does not unambiguously promise raw-byte consumption boundaries; new docs are not used to override the original.
- **Require no line cap or CRLF byte preservation:** Those guarantees are absent from lineexport, and the original Scanner token limit/ScanLines behavior is retained.
- **Treat Write returning n > len(p) with nil error as a supported case:** That violates io.Writer. The < short-write check covers valid counts; defending every invalid return is not a task requirement.
- **Add public invalid-limit classification, flag parsing, atomic output or extra abstractions:** The required APIs and positional CLI are already met. Atomic publication is explicitly not promised, and flags would affect dash-prefixed paths.
- **Grade os.Exit skipping borrowed reader defer on failure as an introduced defect:** This CLI behavior predates the changeset; the owned writer is already closed by ExportLimit. No changed reader-cleanup/failure contract is supplied.

## Evidence and limits

Actual commands, statuses and outputs are in `supplement-checks.json`. Executed independent sources and verbatim supplied probes are preserved under `evidence/supplement/`; `evidence/manifest.json` records their hashes. The ZIP sticky-cause check directly demonstrates why its small private tracker is justified; independent body-plus-completion cases test the separate requirement to retain both causes.

The initial failing lookup harness was overwritten before file preservation. Its failing command output and the correction remain in frozen `checks.json`, and the corrected executed source is preserved under `evidence/initial/`. No reconstructed source has been labeled original.

All supplied supplemental base/module source and tests were inspected. Independent commands ran only in disposable copies with GOWORK=off, GOTOOLCHAIN=local and the requested GOCACHE, using Go 1.26.5 on darwin/arm64. Both base suites, complete author/supplied/independent race suites and vet passed; maintained Go sources have empty gofmt -d output. A native Go 1.22 matrix was not run; standard-library choices are compatible by inspected module/API use. staticcheck was unavailable in the recorded initial environment check and was not installed. No general vulnerability, extraction, dependency or performance audit was performed. The first six conclusions/files and all supplemental candidate inputs remain unchanged.
