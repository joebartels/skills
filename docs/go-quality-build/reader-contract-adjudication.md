# Reader contract disagreement adjudication

Reviewed 2026-10-01 independently and before inspecting any arm mapping or candidate skill draft. **The supplied evidence establishes one contained, moderate defect in reader candidate A, supporting B/B for Code Quality and Correctness in the disputed scope.** The earlier ambiguity assessment identified a real omission of explicit precedence wording, but that omission does not override the stated reader-cause inspection promise. The original reports are preserved unchanged.

## Boundary and evidence

This adjudication concerns only the simultaneous malformed terminal bytes/non-EOF reader-error case in `recordload.Load`. It does not rerun the twelve-pair outcome assessment, identify exposure arms, revise author code, recommend runtime promotion, or establish a causal skill effect.

Inspected the original [task fixture](../../tests/go-quality-build/go-error-contracts/evals/files/reader-errors/README.md), its `load.go`, `load_test.go`, and Go 1.22 module; the task, original, both candidates and supplied probe under `/private/tmp/go-language-blind-outcomes-3b23/go-error-contracts--reader-errors/`; the earlier [ambiguity argument](error-contracts-baseline-review.md#reproduced-ambiguity-not-a-confirmed-defect); and [blind finding F1](language-contracts-blind-outcomes.md#confirmed-finding-and-reproduction). Read the unchanged Code Quality and Correctness review skills and decision references. Candidate skill drafts and the arm-map were not inspected. The earlier report was necessarily visible as one argument under adjudication, so independence here means a separate evidence-based judgment, not ignorance of prior conclusions.

The anonymized original's four files match the canonical fixture byte-for-byte by SHA-256. Input hashes, toolchain, exact commands, statuses, and raw output are in [verification.json](../../tests/go-quality-build/results/2026-10-01-language-contracts/reader-adjudication/verification.json). The standalone [probe](../../tests/go-quality-build/results/2026-10-01-language-contracts/reader-adjudication/probe_test.go.txt) was added only to disposable source copies.

## Precise observations

The probe reader supplies `a=1\nbroken` with a distinct typed, non-EOF error in one `Read`, then EOF. It respects the destination capacity and retains any bytes that do not fit. A separate `bufio.Reader` check confirms that the first `ReadString` returns `a=1\n` with nil error and the second returns `broken` with the exact caller error, without another underlying read. Thus the disputed read cause is already available to candidate A at its parse-failure return.

| Implementation | Valid prefix | Returned error on the disputed input | `errors.Is` caller cause | `errors.As` caller cause |
| --- | --- | --- | --- | --- |
| Original | nil | Exact caller cause | true | true |
| Candidate A | `[{a 1}]` | `*LineError` for line 2, text `broken` | false | false |
| Candidate B | `[{a 1}]` | Joined `*LineError` and caller cause | true | true |

Candidate A's `load.go:37–38` returns immediately after `parseLine` fails, discarding the non-EOF `readErr` from line 33. Candidate B's `load.go:37–39` retains the caller cause alongside the structured line error. Candidate A's own README and `Load` comment also promise inspectable reader failures without describing a malformed-input exception.

Controls establish that both candidates preserve both valid records and the caller cause for `a=1\nb=2` plus the same error. For malformed bytes plus `io.EOF`, both return the prefix and line error; for valid bytes plus EOF, both return the records and nil error. EOF therefore is not being treated as an additional failure that must be exposed.

## Contract judgment

The original README separately requires a valid prefix on malformed input or read failure, a structured malformed-record error, inspectable caller reader errors, processing of bytes accompanying a read error, and normal completion on EOF. The pivotal sentence is: “The caller's reader errors remain inspectable.” It supplies no exception for malformed bytes returned by that same call.

The earlier argument correctly observes that the README does not prescribe precedence, `errors.Join`, aggregate ordering, or the exact top-level error type for two failures. It is also correct that a requirement in another fixture must not be imported into this one. Those points leave representation choices open. They do not establish permission to remove a separately promised observable cause. The phrase “malformed input or read failure” describes when a prefix is returned; it does not say these conditions are mutually exclusive or authorize one to suppress the other. Simultaneous satisfaction is possible without changing the API, as candidate B demonstrates. B's stronger explanatory prose is not the source of the original obligation.

The standard library permits a reader to return positive bytes and a non-nil error together, and instructs consumers to process those bytes before considering the error. That ordering alone does **not** impose a universal policy for exposing every later parse/read combination; the application contract still governs exposure. Here it establishes that the reproduced input is supported, and the explicit application inspection promise determines what must survive processing. [io.Reader documentation](https://pkg.go.dev/io#Reader). `ReadString` likewise documents returning the accumulated bytes and error together when the delimiter was not found. [bufio.Reader.ReadString documentation](https://pkg.go.dev/bufio#Reader.ReadString).

The original `Load` reinforces this reading: it returns the caller cause directly on the identical input. That behavior is corroborating compatibility evidence, not the sole specification. The original discards useful bytes on read failure and lacks the requested structured error, so preserving every aspect of its implementation would itself fail the new task. Adding the missing prefix and line-error behavior does not require taking away its already-documented cause inspection. No fixture release policy authorizes that removal.

This finding needs no prospective stronger contract and no assumption that every unattempted read must occur. It is confined to a non-EOF error already delivered with bytes being processed. There is no ruling here that a streaming parser must drain the reader after detecting a complete malformed line, discover future errors, return all error causes in an exact order, preserve direct error equality, or use a particular aggregate constructor. A future explicit parse-precedence exception would change the applicable contract; none was supplied for this evaluation.

## Code Quality & Go Idioms — B

Scope: Original to candidate A, narrowly the terminal partial-read/parse-error flow and its documentation; Go 1.22 library.

Coverage: Compared original and both candidate implementations/docs; reproduced simultaneous errors, valid bytes with the same error, and EOF controls. Broader quality grading remains with the original outcome reports.

Rationale: One moderate local error-flow bug drops a cause the API promises callers can inspect. The valid prefix and parse diagnostic survive, so the evidence supports contained classification loss rather than major or systemic failure.

Finding counts: critical=0, major=0, moderate=1, minor=0.

Good

- [G1] Candidate A processes valid partial bytes, returns the accepted prefix, and preserves an ordinary reader failure through wrapping; executed controls confirm these paths.

Bad

- [F1][moderate][introduced] Candidate A `load.go:37–38` returns the parse error and discards an available non-EOF `readErr`, contradicting the supplied and resulting documentation. This is the same cause as the Correctness finding, not an additional issue.

Suggested changes

- [F1] Preserve inspectability of the already-encountered caller cause while returning the required prefix and structured malformed-record error. Verify the exact simultaneous-error input in addition to the existing separate-failure tests. No author edit is made by this adjudication.

Limits: No comprehensive new style or performance review. Grade applies to the disputed scope and verified consequences, using the unchanged rubric.

## Correctness & Compatibility — B

Scope: Same original-to-candidate-A transition and bounded reader contract.

Coverage: Caller `errors.Is` and `errors.As` inspection, accepted prefix, parse-error information, legal `io.Reader` behavior, original observable behavior, and EOF controls.

Rationale: F1 causes a caller classification/diagnostic contract mismatch on a supported input. No actual downstream consumer, retry policy, or data-loss consequence was supplied, so greater severity is unsupported. One moderate issue selects B.

Finding counts: critical=0, major=0, moderate=1, minor=0.

Good

- [G1] Candidate A retains the correct accepted prefix and structured malformed-record location; valid-byte/read-error and EOF controls preserve their expected results.

Bad

- [F1][moderate][introduced] The same early return makes both caller-cause inspection operations fail on an input for which the original preserved the cause and the task continues to promise inspection.

Suggested changes

- [F1] Keep the caller cause inspectable alongside the line error on this input; preserve normal EOF behavior. The permanent probe separates the observational reproduction from the contract assertion.

Limits: No Go 1.22 binary, other platform, real I/O integration, or broader downstream impact was tested. Candidate B satisfies this disputed case; this narrow adjudication does not independently re-award its full A/A grades.

## Verification and reconciliation

Executed in `/private/tmp/go-reader-adjudication-3b23/{original,candidate-a,candidate-b}` with `GOCACHE=/private/tmp/go-language-contract-check-cache`, `GOWORK=off`, and Go 1.26.5 on darwin/arm64. Each copy's supplied tests passed `rtk proxy go test -race -count=1 -timeout=45s ./...`, and each passed `rtk proxy go vet ./...` before adding the adjudication probe. All three observational/buffer-boundary runs passed. The separate caller-cause contract assertion passed for original and candidate B and failed for candidate A, as recorded above: 11 successful commands and one intentional defect reproduction among the 12 test/vet commands. Tests and declarations use features available at the stated Go 1.22 floor; actual minimum-toolchain execution remains unverified.

The earlier report's A/A explicitly excluded a conclusion on this combination. Preserve it as that historical, narrower review. When including this now-adjudicated case, its uncertainty should be reconciled to the single moderate F1, not erased or treated as a second independent defect. The later B/B interpretation is supported. The outcome comparison must disclose the disagreement and its resolution before revealing arms, and must not claim that one pair proves a skill caused improvement.

Root owns the canonical design-record update. Next action: link this adjudication and its evidence from that record, then interpret the already-frozen candidate comparison with provenance and the stated causal limits intact.
