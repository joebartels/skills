# Blinded language-contract code outcomes

Initial blind review frozen: 2026-10-01 (America/Chicago). Reviewer: independent outcome-review agent. Candidates A/B had no exposed arm names during this phase. This is a review of supplied fictional Go changesets against their supplied originals and task contracts, not an effectiveness verdict.

## Boundary and method

Reviewed all files in `original`, `candidate-a`, and `candidate-b`, plus each `task.txt` and `probe_test.go.txt`, under `/private/tmp/go-language-blind-outcomes-3b23`. Used the unchanged repository review skills `plugins/go-quality-review/skills/go-code-quality-and-idioms/SKILL.md` and `plugins/go-quality-review/skills/go-correctness-and-compatibility/SKILL.md`, including their decision references. Did not read arm map, author trial reports, original arm archives, new draft skills, current design/progress/outcome documents, or excluded-run output. The canonical progress update is left to the parent to preserve this blind boundary.

Each candidate is assessed as a changeset relative to its paired `original`, including the new behavior expressly required by `task.txt`/README. Existing limitations and unsupported inputs are not invented as new defects. Shared findings are counted once as root causes across the two review topics; topic grades are not overall Go-quality scores. No A+ is awarded: working implementations of these supplied requirements support A, without asserting extra independent safeguards beyond the task’s routine contract implementation.

All modules retain `go 1.22`; no candidate adds a dependency or build tag. Checks used installed `go1.26.5 darwin/arm64`, `GOWORK=off`, and `GOCACHE=/private/tmp/go-language-contract-check-cache`. Native Go 1.22 and other OS/architecture builds were not run. Version-compatible source/API usage was inspected. No new concurrency guarantee was assumed except the counter’s expressly shared usage.

Candidates were preserved. Executable checks and inserted tests were in `/private/tmp/go-language-review-copies-3b23/<pair>/<candidate>`. Across 24 candidates, `rtk proxy go test -count=1 ./...` with the supplied private probe, `rtk proxy go vet ./...`, and non-mutating `rtk proxy gofmt -l <candidate Go files>` all passed with no vet/formatting diagnostics. Both counter copies additionally passed `rtk proxy go test -race -count=1 ./...`. Reviewer edge tests were then run with `rtk proxy go test -count=1 -run TestReview -v .`: only reader A’s simultaneous parse/read assertion failed. Raw local evidence is retained in `verification.json`, `review-edges.json`, and `cli-verification.json` under the reviewer-copy root. These are temporary evidence files, not portable repository artifacts; decisive reproducer/output is included below.

## Grade matrix

| Pair | A: Code Quality | A: Correctness | B: Code Quality | B: Correctness | Confirmed root causes |
| --- | --- | --- | --- | --- | --- |
| `go-error-contracts--reader-errors` | B | B | A | A | A: F1, one moderate |
| `go-error-contracts--backend-errors` | A | A | A | A | None found |
| `go-error-contracts--cli-completion` | A | A | A | A | None found |
| `go-error-contracts--csv-transfer` | A | A | A | A | None found |
| `go-error-contracts--cursor-iteration` | A | A | A | A | None found |
| `go-error-contracts--not-error-work` | A | A | A | A | None found |
| `go-values-and-zero-values--default-wire` | A | A | A | A | None found |
| `go-values-and-zero-values--snapshot-ownership` | A | A | A | A | None found |
| `go-values-and-zero-values--required-construction` | A | A | A | A | None found |
| `go-values-and-zero-values--not-value-work` | A | A | A | A | None found |
| `go-values-and-zero-values--borrowed-transfer` | A | A | A | A | None found |
| `combined--language-contracts` | A | A | A | A | None found |

## Confirmed finding and reproduction

[F1][moderate][introduced in the new implementation] Reader candidate A drops the caller’s read cause when parsing the same terminal chunk fails. Primary remediation owner: Code Quality/error flow; Correctness cross-references the same root cause.

Location: `go-error-contracts--reader-errors/candidate-a/load.go:37`–38 under the blind source root. `ReadString` can return both final bytes and a non-EOF error. The immediate `return records, err` discards `readErr`, even though the original task explicitly says caller reader errors remain inspectable. A caller receives its valid record prefix and a `*LineError`, but cannot discover the read failure through `errors.Is`/`As`. This is a contained diagnostic/classification contract loss, not complete data loss or systemic failure; one moderate issue selects B in both applicable topics. Candidate B preserves both causes with `errors.Join`.

Contract interpretation is explicit here: the original README says, “The caller's reader errors remain inspectable” and immediately requires processing bytes returned alongside a read error. It gives no exception for malformed bytes in that same read and no instruction that the parse error suppresses the read cause. The new structured line-error requirement applies concurrently. F1 is therefore assessed against those two stated obligations, not against B merely exposing more errors. The original implementation also returns the read cause for the identical input (verified separately: nil prefix, `readCause=true`); A newly hides it while adding the requested prefix/line behavior. If the fixture owner instead intends parse-error precedence to suppress the read cause, that would change this finding, but that exception is absent from the supplied contract. This differs from the limit boundary, where the task does not promise inspection of every buffered read error after the requested record count is met.

The existing supplied `onceReader` returns its bytes and configured error on the same call. Added only to the disposable copy:

```go
func TestReviewSimultaneousMalformedReadFailure(t *testing.T) {
    cause := errors.New("underlying read failure")
    got, err := Load(&onceReader{data: "a=1\nbroken", err: cause})
    var line *LineError
    if len(got) != 1 || !errors.As(err, &line) || !errors.Is(err, cause) {
        t.Fatalf("prefix=%#v err=%v line=%v readCause=%v",
            got, err, line, errors.Is(err, cause))
    }
}
```

Executed independently for both reader copies using the environment above:

```text
rtk proxy go test -count=1 -run TestReview -v .
A: FAIL TestReviewSimultaneousMalformedReadFailure
   prefix=[{Key:"a", Value:"1"}]
   err=line 2: malformed record "broken"
   readCause=false
B: PASS TestReviewSimultaneousMalformedReadFailure
A and B: PASS TestReviewBytesWithEOF
```

Suggested correction: preserve the non-EOF read failure alongside the `LineError` when both are available; verify both `errors.As(err, &line)` and `errors.Is(err, cause)` and keep the valid prefix. If retaining buffering, account for an underlying read error already captured with bytes whose delimiter causes `ReadString` to defer returning that error; do not overclaim co-failure retention merely from joining one local variable. The immediate no-final-newline reproduction is sufficient to establish F1.

## Actual CLI checks

Built each candidate’s real `./cmd/lineexport` or `./cmd/batchview` binary in its disposable copy with `rtk proxy go build -o <copy>/review-cli ./cmd/<name>`, then ran it as a subprocess with separate captured stdout/stderr. All 32 cases matched the specified statuses and streams; output contents were inspected as well.

| Exercise | lineexport A/B | batchview A/B |
| --- | --- | --- |
| ordinary success | exit 0, both streams empty, `a\nb\nc\n` | exit 0, both streams empty, both Item records with original JSON tags |
| optional selector, `-input`/`-output` paths | limit 2 publishes `a\nb\n`, exit 0, quiet | prefix `a` publishes only key `a`, exit 0, quiet |
| invalid argument count | exit 2, usage on stderr, empty stdout | same |
| output path is a directory | exit 1, stderr error, empty stdout | same |
| missing input | exit 1, stderr error, empty stdout | same |
| late input failure | long second Scanner token: exit 1, `a\n` remains published | malformed second record: exit 1, valid first Item remains published |
| read failure from directory input | exit 1, read error on stderr, empty output file | exit 1, read error on stderr, published `null` prefix |
| noninteger limit | exit 2, usage, no output file | not applicable |
| negative limit | exit 1, rejection on stderr, empty output file | not applicable |

No real file close failure was injected into CLI binaries; library-owned writer probes cover Close and dual-cause behavior. Nonzero-prefix input I/O failure is tested through library readers; the OS directory-input CLI check has an empty valid prefix. No atomic publication, retry, or rollback guarantee was required or inferred.

## Semantic differences that do not lower grades

- Export limit boundary: with one `Read` returning `"a\n"` plus a failure and limit 1, A returns `(1, readFailure)`, B `(1, nil)`; both write `"a\n"` and close once. A consults `scanner.Err` after breaking (`export.go:38`), while B returns upon reaching the limit (`export.go:43`). The task defines maximum accepted records but does not explicitly settle precedence for an already-observed buffered read failure at that exact boundary. This remains a contract decision to clarify, not a confidently graded defect or measured advantage.
- Counter zero Count: reviewer observation confirms A panics with `signedcounter: use New`; B returns 0. Zero Counter is expressly unsupported, so neither supported consumer behavior nor constructor validation is broken.
- Reader A wraps ordinary read failures with context and streams by line; B retains direct read failures and reads all input before parsing. Inspectability is required; direct reader-error equality and exact reader consumption are not. Only the confirmed simultaneous-cause loss is graded.
- Combined CLI B prints an input error as well as output failure when both occur; A prioritizes output failure. Both preserve failing status and emit stderr; dual-failure text is not specified.

## Candidate report cards

Paths below are relative to `/private/tmp/go-language-blind-outcomes-3b23`. Each topic inherits the exact executed check outcomes and platform limits above. Unnecessary-cost notes are ungraded unless a concrete consequence is established.

### Pair `go-error-contracts--reader-errors`

Change costs and semantic/document differences: A changes to bufio.Reader/ReadString and wraps reader errors with operation context; B retains io.ReadAll and explicitly joins simultaneous parse/read errors. A stops reading after malformed input; B consumes the input first, matching the original strategy. This is an observable consumption/resource difference, but no exact consumption amount is promised. A adds a private parseLine helper with pointer-to-slice mutation; no grade deduction for that choice. Both add only the requested LineError API. B documents simultaneous error preservation; A makes the broad inspectability claim that F1 contradicts.

#### candidate-a

##### Code Quality & Go Idioms — B

Scope: supplied `original` → `candidate-a`, recordload library, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, blank/empty fields, physical line errors, partial data with read errors, data plus EOF, 100,000-byte records, borrowed-reader behavior, exported LineError, and unchanged Load function shape.

Rationale: F1 is one moderate error-flow/observable-inspectability issue; the accepted prefix survives, so major/systemic severity is unsupported.

Finding counts: critical=0, major=0, moderate=1, minor=0.

Good

- [G1] `go-error-contracts--reader-errors/candidate-a/load.go`: The exported pointer LineError retains physical line and original text; successful results are literal nil error interfaces. Both preserve long records and ordinary read-failure prefixes.

Bad

- [F1][moderate][introduced] `load.go:37` returns the LineError while discarding the simultaneous read cause; see the frozen reproduction above.

Suggested changes

- [F1] Retain both parse and non-EOF read causes with the valid prefix, and cover co-failure explicitly.

Limits: Common execution/platform limits above apply. Supplied probes pass; the added co-failure reproduction fails only for this candidate.

##### Correctness & Compatibility — B

Scope: supplied `original` → `candidate-a`, recordload library, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, blank/empty fields, physical line errors, partial data with read errors, data plus EOF, 100,000-byte records, borrowed-reader behavior, exported LineError, and unchanged Load function shape.

Rationale: F1 is one moderate error-flow/observable-inspectability issue; the accepted prefix survives, so major/systemic severity is unsupported.

Finding counts: critical=0, major=0, moderate=1, minor=0.

Good

- [G1] `go-error-contracts--reader-errors/candidate-a/load.go`: The exported pointer LineError retains physical line and original text; successful results are literal nil error interfaces. Both preserve long records and ordinary read-failure prefixes.

Bad

- [F1][moderate][introduced] `load.go:37` returns the LineError while discarding the simultaneous read cause; see the frozen reproduction above.

Suggested changes

- [F1] Retain both parse and non-EOF read causes with the valid prefix, and cover co-failure explicitly.

Limits: Common execution/platform limits above apply. Supplied probes pass; the added co-failure reproduction fails only for this candidate.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, recordload library, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, blank/empty fields, physical line errors, partial data with read errors, data plus EOF, 100,000-byte records, borrowed-reader behavior, exported LineError, and unchanged Load function shape.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--reader-errors/candidate-b/load.go`: The exported pointer LineError retains physical line and original text; successful results are literal nil error interfaces. Both preserve long records and ordinary read-failure prefixes.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, recordload library, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, blank/empty fields, physical line errors, partial data with read errors, data plus EOF, 100,000-byte records, borrowed-reader behavior, exported LineError, and unchanged Load function shape.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--reader-errors/candidate-b/load.go`: The exported pointer LineError retains physical line and original text; successful results are literal nil error interfaces. Both preserve long records and ordinary read-failure prefixes.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-error-contracts--backend-errors`

Change costs and semantic/document differences: Production lookup implementations are identical. Both add only the required ErrInvalidKey and standard-library imports. A tests successful original-key forwarding; B has an explicit call flag for rejected input. README wording remains the original task-oriented wording in both, including the stale sentence “Empty keys currently reach the backend.” This was pre-existing and the task did not require rewriting docs, so it is recorded rather than counted as an introduced defect.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, accountlookup library, Go 1.22; source, tests, README and task inspected.

Coverage: Blank-key rejection before backend access, original nonblank key forwarding, direct missing sentinel equality, public unavailable classification, discarded unusable values, and hidden backend identity/type/diagnostic text.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--backend-errors/candidate-a/lookup.go`: Both wrap only ErrUnavailable with quoted key/operation context, preserve direct ErrMissing equality, and reject whitespace-only keys before invoking Backend. The private probe checks Is/As exposure and sensitive text.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, accountlookup library, Go 1.22; source, tests, README and task inspected.

Coverage: Blank-key rejection before backend access, original nonblank key forwarding, direct missing sentinel equality, public unavailable classification, discarded unusable values, and hidden backend identity/type/diagnostic text.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--backend-errors/candidate-a/lookup.go`: Both wrap only ErrUnavailable with quoted key/operation context, preserve direct ErrMissing equality, and reject whitespace-only keys before invoking Backend. The private probe checks Is/As exposure and sensitive text.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, accountlookup library, Go 1.22; source, tests, README and task inspected.

Coverage: Blank-key rejection before backend access, original nonblank key forwarding, direct missing sentinel equality, public unavailable classification, discarded unusable values, and hidden backend identity/type/diagnostic text.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--backend-errors/candidate-b/lookup.go`: Both wrap only ErrUnavailable with quoted key/operation context, preserve direct ErrMissing equality, and reject whitespace-only keys before invoking Backend. The private probe checks Is/As exposure and sensitive text.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, accountlookup library, Go 1.22; source, tests, README and task inspected.

Coverage: Blank-key rejection before backend access, original nonblank key forwarding, direct missing sentinel equality, public unavailable classification, discarded unusable values, and hidden backend identity/type/diagnostic text.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--backend-errors/candidate-b/lookup.go`: Both wrap only ErrUnavailable with quoted key/operation context, preserve direct ErrMissing equality, and reject whitespace-only keys before invoking Backend. The private probe checks Is/As exposure and sensitive text.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-error-contracts--cli-completion`

Change costs and semantic/document differences: A exposes an additional ErrInvalidLimit sentinel although only rejection was required; this is an extra public compatibility commitment, with no demonstrated consumer breakage. B uses an unexported constructed error and checks short writes; A retains the original unchecked short-write count. A non-nil-error-short-write is correctly handled by both; a writer returning a short count with nil violates io.Writer, so that robustness difference is ungraded. A uses finish; B uses deferred finalization. Both retain Scanner’s pre-existing size limit and read-ahead behavior, for which this fixture promises no uncapped lines. A adds CLI tests; B has no candidate-authored CLI tests but independently exercised binaries pass. Limit-boundary read-error precedence differs as detailed below; it is not counted.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, lineexport library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Owned-writer single close; lone error identity; operation/close combined causes; empty and final records; positive/zero/negative limits; exact Export function type; actual CLI status, streams, dash-prefixed paths, read/output errors and partial publication.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cli-completion/candidate-a/export.go`: Both preserve direct equality for single operation or Close errors, expose both independent causes on dual failure, retain accepted prefixes, and close exactly once in normal/error returns. Actual binaries publish the expected limited prefix.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, lineexport library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Owned-writer single close; lone error identity; operation/close combined causes; empty and final records; positive/zero/negative limits; exact Export function type; actual CLI status, streams, dash-prefixed paths, read/output errors and partial publication.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cli-completion/candidate-a/export.go`: Both preserve direct equality for single operation or Close errors, expose both independent causes on dual failure, retain accepted prefixes, and close exactly once in normal/error returns. Actual binaries publish the expected limited prefix.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, lineexport library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Owned-writer single close; lone error identity; operation/close combined causes; empty and final records; positive/zero/negative limits; exact Export function type; actual CLI status, streams, dash-prefixed paths, read/output errors and partial publication.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cli-completion/candidate-b/export.go`: Both preserve direct equality for single operation or Close errors, expose both independent causes on dual failure, retain accepted prefixes, and close exactly once in normal/error returns. Actual binaries publish the expected limited prefix.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, lineexport library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Owned-writer single close; lone error identity; operation/close combined causes; empty and final records; positive/zero/negative limits; exact Export function type; actual CLI status, streams, dash-prefixed paths, read/output errors and partial publication.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cli-completion/candidate-b/export.go`: Both preserve direct equality for single operation or Close errors, expose both independent causes on dual failure, retain accepted prefixes, and close exactly once in normal/error returns. Actual binaries publish the expected limited prefix.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-error-contracts--csv-transfer`

Change costs and semantic/document differences: Both leave WriteRows unchanged and add only WriteSelected. B directly returns writer.Error; A uses an explicit error branch. B states explicitly that failure count includes the failing submitted row; A’s text also describes submitted rows. No API, rename, copy, or dependency cost beyond the requested function.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, csvselect library, Go 1.22; source, tests, README and task inspected.

Coverage: Selection order, nil predicate, valid quoting, count including the submitted failing row, immediate buffer-triggered write failure, final flush failure, borrowed writer ownership, and unchanged WriteRows.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--csv-transfer/candidate-a/csv.go`: Both increment the submitted-row count before encoder Write and check writer.Error after Flush. The private delayed-failure probe and reviewer immediate-failure probe preserve the underlying cause without closing the writer.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, csvselect library, Go 1.22; source, tests, README and task inspected.

Coverage: Selection order, nil predicate, valid quoting, count including the submitted failing row, immediate buffer-triggered write failure, final flush failure, borrowed writer ownership, and unchanged WriteRows.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--csv-transfer/candidate-a/csv.go`: Both increment the submitted-row count before encoder Write and check writer.Error after Flush. The private delayed-failure probe and reviewer immediate-failure probe preserve the underlying cause without closing the writer.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, csvselect library, Go 1.22; source, tests, README and task inspected.

Coverage: Selection order, nil predicate, valid quoting, count including the submitted failing row, immediate buffer-triggered write failure, final flush failure, borrowed writer ownership, and unchanged WriteRows.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--csv-transfer/candidate-b/csv.go`: Both increment the submitted-row count before encoder Write and check writer.Error after Flush. The private delayed-failure probe and reviewer immediate-failure probe preserve the underlying cause without closing the writer.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, csvselect library, Go 1.22; source, tests, README and task inspected.

Coverage: Selection order, nil predicate, valid quoting, count including the submitted failing row, immediate buffer-triggered write failure, final flush failure, borrowed writer ownership, and unchanged WriteRows.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--csv-transfer/candidate-b/csv.go`: Both increment the submitted-row count before encoder Write and check writer.Error after Flush. The private delayed-failure probe and reviewer immediate-failure probe preserve the underlying cause without closing the writer.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-error-contracts--cursor-iteration`

Change costs and semantic/document differences: Runtime logic is identical. B expands Rows/Collect docs and adds external function-type/interface assertions. A keeps the implementation small and covers the same core failure behavior with internal tests. No new dependency, interface method, unnecessary copy, or unrelated rename.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, cursorload library, Go 1.22; source, tests, README and task inspected.

Coverage: Filtered order, nil predicate, scan and terminal iteration failures, valid prefix, exactly-once close, best-effort close error precedence, and unchanged Rows/Collect compatibility.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cursor-iteration/candidate-a/collect.go`: Both route Collect through CollectMatching and return rows.Err after iteration stops. Scan failure returns the existing prefix, while deferred Close remains best-effort. Supplied tests and probes verify filtering, late failure and close count.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, cursorload library, Go 1.22; source, tests, README and task inspected.

Coverage: Filtered order, nil predicate, scan and terminal iteration failures, valid prefix, exactly-once close, best-effort close error precedence, and unchanged Rows/Collect compatibility.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cursor-iteration/candidate-a/collect.go`: Both route Collect through CollectMatching and return rows.Err after iteration stops. Scan failure returns the existing prefix, while deferred Close remains best-effort. Supplied tests and probes verify filtering, late failure and close count.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, cursorload library, Go 1.22; source, tests, README and task inspected.

Coverage: Filtered order, nil predicate, scan and terminal iteration failures, valid prefix, exactly-once close, best-effort close error precedence, and unchanged Rows/Collect compatibility.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cursor-iteration/candidate-b/collect.go`: Both route Collect through CollectMatching and return rows.Err after iteration stops. Scan failure returns the existing prefix, while deferred Close remains best-effort. Supplied tests and probes verify filtering, late failure and close count.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, cursorload library, Go 1.22; source, tests, README and task inspected.

Coverage: Filtered order, nil predicate, scan and terminal iteration failures, valid prefix, exactly-once close, best-effort close error precedence, and unchanged Rows/Collect compatibility.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--cursor-iteration/candidate-b/collect.go`: Both route Collect through CollectMatching and return rows.Err after iteration stops. Scan failure returns the existing prefix, while deferred Close remains best-effort. Supplied tests and probes verify filtering, late failure and close count.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-error-contracts--not-error-work`

Change costs and semantic/document differences: Production code is identical. A uses a table and B separate tests. No new errors, API, copies, normalization, renaming or dependencies; unchanged README remains accurate. Invalid negative inputs or zero sizes are outside the stated contract.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--not-error-work/candidate-a/pages.go`: Both use quotient plus a remainder test, correcting zero and exact multiples without overflow from adding size to items. The helper remains private and maximum-int checks pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--not-error-work/candidate-a/pages.go`: Both use quotient plus a remainder test, correcting zero and exact multiples without overflow from adding size to items. The helper remains private and maximum-int checks pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--not-error-work/candidate-b/pages.go`: Both use quotient plus a remainder test, correcting zero and exact multiples without overflow from adding size to items. The helper remains private and maximum-int checks pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-error-contracts--not-error-work/candidate-b/pages.go`: Both use quotient plus a remainder test, correcting zero and exact multiples without overflow from adding size to items. The helper remains private and maximum-int checks pass.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-values-and-zero-values--default-wire`

Change costs and semantic/document differences: A uses an unlimited boolean, B configured; both are valid local representations. A uses fmt.Errorf without substitutions; B errors.New. This is an optional spelling choice, not a defect. No unnecessary input clone or API change beyond NewWithLimit; no new dependency. A adds fuller zero-value type/New docs; B states this in README.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, reportview library, Go 1.22; source, tests, README and task inspected.

Coverage: Zero/New default ten, explicit unlimited zero, positive and negative configuration, exact New function type, nil versus empty JSON, unchanged tag, and borrowed input.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--default-wire/candidate-a/view.go`: Both represent explicit configuration separately from the zero value, preserve New’s exact function type and default, and reslice only a local slice header before JSON encoding. Probes distinguish null from [].

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, reportview library, Go 1.22; source, tests, README and task inspected.

Coverage: Zero/New default ten, explicit unlimited zero, positive and negative configuration, exact New function type, nil versus empty JSON, unchanged tag, and borrowed input.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--default-wire/candidate-a/view.go`: Both represent explicit configuration separately from the zero value, preserve New’s exact function type and default, and reslice only a local slice header before JSON encoding. Probes distinguish null from [].

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, reportview library, Go 1.22; source, tests, README and task inspected.

Coverage: Zero/New default ten, explicit unlimited zero, positive and negative configuration, exact New function type, nil versus empty JSON, unchanged tag, and borrowed input.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--default-wire/candidate-b/view.go`: Both represent explicit configuration separately from the zero value, preserve New’s exact function type and default, and reslice only a local slice header before JSON encoding. Probes distinguish null from [].

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, reportview library, Go 1.22; source, tests, README and task inspected.

Coverage: Zero/New default ten, explicit unlimited zero, positive and negative configuration, exact New function type, nil versus empty JSON, unchanged tag, and borrowed input.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--default-wire/candidate-b/view.go`: Both represent explicit configuration separately from the zero value, preserve New’s exact function type and default, and reslice only a local slice header before JSON encoding. Probes distinguish null from [].

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-values-and-zero-values--snapshot-ownership`

Change costs and semantic/document differences: Implementations are substantively the same. B expands view/Name docs and adds external consumer coverage; A has concise ownership docs. The copies are required by the fork contract; RootView keeps borrowing and existing receivers stay values. No forced constructor, clone API beyond Fork, new dependency, or unrelated rename. The contract preserves node sharing, not arbitrary backing-array alias relationships among payload fields.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, editgraph library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed RootView, value method/interface use, zero graph, independent nodes/maps/bytes/link slices, shared node targets, cycles, nil links, nil/empty collections, and mutation in both directions.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--snapshot-ownership/candidate-a/graph.go`: Both memoize copied nodes before traversing links, preserving shared targets/cycles, and copy maps, byte values and link slices. Candidate tests and private probes verify graph topology and independent edits, including nested mutable storage.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, editgraph library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed RootView, value method/interface use, zero graph, independent nodes/maps/bytes/link slices, shared node targets, cycles, nil links, nil/empty collections, and mutation in both directions.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--snapshot-ownership/candidate-a/graph.go`: Both memoize copied nodes before traversing links, preserving shared targets/cycles, and copy maps, byte values and link slices. Candidate tests and private probes verify graph topology and independent edits, including nested mutable storage.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, editgraph library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed RootView, value method/interface use, zero graph, independent nodes/maps/bytes/link slices, shared node targets, cycles, nil links, nil/empty collections, and mutation in both directions.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--snapshot-ownership/candidate-b/graph.go`: Both memoize copied nodes before traversing links, preserving shared targets/cycles, and copy maps, byte values and link slices. Candidate tests and private probes verify graph topology and independent edits, including nested mutable storage.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, editgraph library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed RootView, value method/interface use, zero graph, independent nodes/maps/bytes/link slices, shared node targets, cycles, nil links, nil/empty collections, and mutation in both directions.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--snapshot-ownership/candidate-b/graph.go`: Both memoize copied nodes before traversing links, preserving shared targets/cycles, and copy maps, byte values and link slices. Candidate tests and private probes verify graph topology and independent edits, including nested mutable storage.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-values-and-zero-values--required-construction`

Change costs and semantic/document differences: A adds a zero-key panic check in Count, matching Sign’s guard; B returns zero for zero Counter.Count. The zero Counter is explicitly unsupported, so this is an observed unsupported-case difference rather than a weakened constructor or required defect. Both continue to document New as required, introduce only Count, and avoid copying mutexes or adding clone/alternate constructors.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, signedcounter package, Go 1.22; source, tests, README and task inspected.

Coverage: Constructor validation, original Sign behavior, copied key ownership, Count before/after Sign, shared mutex receiver semantics, and concurrent Sign/Count under race detection.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--required-construction/candidate-a/counter.go`: Both use a pointer receiver and the existing mutex for Count, retain constructor validation/key copying, and preserve Sign. Race tests pass; independent key mutation does not change signatures and counts track calls.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, signedcounter package, Go 1.22; source, tests, README and task inspected.

Coverage: Constructor validation, original Sign behavior, copied key ownership, Count before/after Sign, shared mutex receiver semantics, and concurrent Sign/Count under race detection.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--required-construction/candidate-a/counter.go`: Both use a pointer receiver and the existing mutex for Count, retain constructor validation/key copying, and preserve Sign. Race tests pass; independent key mutation does not change signatures and counts track calls.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, signedcounter package, Go 1.22; source, tests, README and task inspected.

Coverage: Constructor validation, original Sign behavior, copied key ownership, Count before/after Sign, shared mutex receiver semantics, and concurrent Sign/Count under race detection.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--required-construction/candidate-b/counter.go`: Both use a pointer receiver and the existing mutex for Count, retain constructor validation/key copying, and preserve Sign. Race tests pass; independent key mutation does not change signatures and counts track calls.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, signedcounter package, Go 1.22; source, tests, README and task inspected.

Coverage: Constructor validation, original Sign behavior, copied key ownership, Count before/after Sign, shared mutex receiver semantics, and concurrent Sign/Count under race detection.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--required-construction/candidate-b/counter.go`: Both use a pointer receiver and the existing mutex for Count, retain constructor validation/key copying, and preserve Sign. Race tests pass; independent key mutation does not change signatures and counts track calls.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-values-and-zero-values--not-value-work`

Change costs and semantic/document differences: Production code is identical. B adds a single-exact-page case to a table; A uses separate tests. No unrelated value API, clone, normalization, constructor, rename, or dependency is added. Invalid inputs remain outside the stated nonnegative-items/positive-size contract.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--not-value-work/candidate-a/pages.go`: Both use quotient plus a remainder test, preserving integer range without the ceiling-addition overflow hazard. Regression probes pass and the helper remains private.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--not-value-work/candidate-a/pages.go`: Both use quotient plus a remainder test, preserving integer range without the ceiling-addition overflow hazard. Regression probes pass and the helper remains private.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--not-value-work/candidate-b/pages.go`: Both use quotient plus a remainder test, preserving integer range without the ceiling-addition overflow hazard. Regression probes pass and the helper remains private.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, paging private helper, Go 1.22; source, tests, README and task inspected.

Coverage: Zero items, exact and partial pages, positive-size contract, and maximum-int boundaries.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--not-value-work/candidate-b/pages.go`: Both use quotient plus a remainder test, preserving integer range without the ceiling-addition overflow hazard. Regression probes pass and the helper remains private.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `go-values-and-zero-values--borrowed-transfer`

Change costs and semantic/document differences: A delegates Collect to CollectMatching(nil); B duplicates the acquisition/error/retention loop and extracts cloneFrame. B therefore has two loops to maintain, but no demonstrated inconsistent behavior or substantial cost warrants a grade deduction. Neither copies rejected frames before the predicate, and accepted-frame copies are necessary. No ownership/lifecycle expansion, rename, or dependency.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, framecollect library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed predicate view identity, filtering, source buffer reuse, independent retained frames, nil/empty frame distinction, source error prefix/cause, exact Collect function type and borrowed lifecycle.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--borrowed-transfer/candidate-a/frames.go`: Both invoke keep on the source buffer before copying accepted frames and preserve nil versus non-nil empty slices. The private probe reuses the buffer across Next calls, then mutates it after return; retained frames remain stable.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, framecollect library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed predicate view identity, filtering, source buffer reuse, independent retained frames, nil/empty frame distinction, source error prefix/cause, exact Collect function type and borrowed lifecycle.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--borrowed-transfer/candidate-a/frames.go`: Both invoke keep on the source buffer before copying accepted frames and preserve nil versus non-nil empty slices. The private probe reuses the buffer across Next calls, then mutates it after return; retained frames remain stable.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, framecollect library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed predicate view identity, filtering, source buffer reuse, independent retained frames, nil/empty frame distinction, source error prefix/cause, exact Collect function type and borrowed lifecycle.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--borrowed-transfer/candidate-b/frames.go`: Both invoke keep on the source buffer before copying accepted frames and preserve nil versus non-nil empty slices. The private probe reuses the buffer across Next calls, then mutates it after return; retained frames remain stable.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, framecollect library, Go 1.22; source, tests, README and task inspected.

Coverage: Borrowed predicate view identity, filtering, source buffer reuse, independent retained frames, nil/empty frame distinction, source error prefix/cause, exact Collect function type and borrowed lifecycle.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `go-values-and-zero-values--borrowed-transfer/candidate-b/frames.go`: Both invoke keep on the source buffer before copying accepted frames and preserve nil versus non-nil empty slices. The private probe reuses the buffer across Next calls, then mutates it after return; retained frames remain stable.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

### Pair `combined--language-contracts`

Change costs and semantic/document differences: A joins malformed errors with readErr even when readErr is nil; B joins only dual failures via a private parse helper. No stable malformed-error identity was promised. B reports both input and output errors if publication fails, while A reports only output failure. The fixture does not specify dual-failure message contents, so this diagnostic difference is ungraded. B has two adjacent loadErr checks; simplification is optional. Both preserve tags/API, add only Select, avoid concurrency claims, and use only the standard library. Selected nested-state copying is required by the API contract; no benchmark establishes waste or benefit.

#### candidate-a

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-a`, batchview library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, partial malformed/read data and combined causes, long records, exact Load signature, Item JSON fields, Select order/deep ownership/nilness, and actual CLI prefix publication, streams/status, positional dash paths and output failure.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `combined--language-contracts/candidate-a/load.go`: Both preserve partial input plus inspectable read causes, deep-copy selected nested maps/bytes with nilness intact, and write selected valid prefixes before reporting incomplete input. Private probes and actual binaries confirm the library and CLI behaviors compose.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-a`, batchview library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, partial malformed/read data and combined causes, long records, exact Load signature, Item JSON fields, Select order/deep ownership/nilness, and actual CLI prefix publication, streams/status, positional dash paths and output failure.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `combined--language-contracts/candidate-a/load.go`: Both preserve partial input plus inspectable read causes, deep-copy selected nested maps/bytes with nilness intact, and write selected valid prefixes before reporting incomplete input. Private probes and actual binaries confirm the library and CLI behaviors compose.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

#### candidate-b

##### Code Quality & Go Idioms — A

Scope: supplied `original` → `candidate-b`, batchview library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, partial malformed/read data and combined causes, long records, exact Load signature, Item JSON fields, Select order/deep ownership/nilness, and actual CLI prefix publication, streams/status, positional dash paths and output failure.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `combined--language-contracts/candidate-b/load.go`: Both preserve partial input plus inspectable read causes, deep-copy selected nested maps/bytes with nilness intact, and write selected valid prefixes before reporting incomplete input. Private probes and actual binaries confirm the library and CLI behaviors compose.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

##### Correctness & Compatibility — A

Scope: supplied `original` → `candidate-b`, batchview library and CLI, Go 1.22; source, tests, README and task inspected.

Coverage: Comments, partial malformed/read data and combined causes, long records, exact Load signature, Item JSON fields, Select order/deep ownership/nilness, and actual CLI prefix publication, streams/status, positional dash paths and output failure.

Rationale: No actionable introduced issue found in this assessed contract scope. The implementation and executed checks establish the strengths below; optional implementation preferences are ungraded.

Finding counts: critical=0, major=0, moderate=0, minor=0.

Good

- [G1] `combined--language-contracts/candidate-b/load.go`: Both preserve partial input plus inspectable read causes, deep-copy selected nested maps/bytes with nilness intact, and write selected valid prefixes before reporting incomplete input. Private probes and actual binaries confirm the library and CLI behaviors compose.

Bad

- None found.

Suggested changes

- None needed.

Limits: Common execution/platform limits above apply. Supplied candidate tests and private contract probes pass; no claim of exhaustive reliability or performance benefit.

## Frozen interpretation boundary

There are 12 paired tasks and 24 independently graded candidates. At this blind stage, 23 candidates receive A in both topics; reader candidate A receives B in both topics for the same single moderate root cause. This is not 48 independent outcome observations and is not evidence of broad reliability. All private supplied probes passing did not eliminate the need for contract-edge inspection.

The matrix, F1, and ungraded semantic/cost observations above are frozen before provenance disclosure. No conclusion about which skill caused improvement, automatic triggering/routing, extra-task generalization, measured speed, token/copy cost, or interaction benefit follows from this review alone. One clean pair per task cannot separate skill effects from author/run variation. The parent has described the first primed value-control run as excluded; its contents and provenance were not inspected for these grades. A later arm interpretation should be recorded separately without retroactively relabeling a blind ambiguity as a defect based on arm identity.
