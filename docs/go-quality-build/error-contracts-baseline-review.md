# Error-contract baseline outcome review

Reviewed 2026-10-01 as an independent assessment of completed code, before any error-contract candidate was authored. The recommendation is **defer a standalone runtime skill**: the inspected baselines satisfy the demonstrated contracts, and no confirmed error-contract defect establishes an incremental improvement target. Retain the source audit and candidate boundary as research, without labeling either implemented or evaluated runtime guidance.

## Scope and evidence

The author model recorded in all four manifests is `gpt-6-luna`, medium effort, using Codex desktop multi-agent trials. The baseline catalog exposed the existing API-contract, composition, and package-boundary skills. This is therefore an assessment with existing build guidance available, not a no-skill baseline. The reviewer used the existing Code Quality & Go Idioms and Correctness & Compatibility review skills and their decision references unchanged. Expected assertions and private probes were visible to the reviewer; this report is independent of authorship, not blinded to evaluation expectations.

Reviewed the complete reconstructed source, author tests, original and resulting README contracts, prompts, manifests, and available archived verification for:

| Case and trial | Reconstructed implementation | Code Quality & Go Idioms | Correctness & Compatibility |
| --- | --- | --- | --- |
| `reader-errors/named` | `load.go`, `load_test.go`, `api_contract_test.go` | A | A |
| `backend-errors/first` | `lookup.go`, `lookup_test.go` | A | A |
| `cli-completion/named` | `export.go`, `export_test.go`, `cmd/lineexport/main.go` | A | A |
| `cursor-iteration/named` | `collect.go`, `collect_test.go` | A | A |

The reconstruction root is `/private/tmp/go-language-trials-3b23/go-error-contracts/baseline/`; each row uses its `reconstructed/` directory. Permanent inputs are in [the error fixture suite](../../tests/go-quality-build/go-error-contracts/evals/evals.json), and prompts, patches and manifests are in [the baseline archive](../../tests/go-quality-build/results/2026-10-01-language-contracts/go-error-contracts/baseline/). All 18 reconstructed output-file hashes and all four archived patch hashes match their manifests. This review did not independently apply the patches or verify every input/catalog hash.

Grades concern introduced or worsened behavior in these completed changes against their supplied inputs. They do not grade unrelated fixture debt or the entire authoring model. The two report cards below cover all four changesets; each case independently has zero counted findings and relevant verified strengths, yielding the same A/A result.

## Code Quality & Go Idioms — A

Scope: Four changesets listed above; three small libraries and one library/CLI, each declaring Go 1.22. Review includes error handling, local control flow, new exported API complexity and documentation accuracy.

Coverage: Inspected every resulting Go source/test file and README. Traced caller-owned read errors, physical line tracking, backend classification and diagnostic exposure, owned writer finalization, best-effort cursor close, iteration completion and CLI error exits. New language/library calls fit Go 1.22 by source inspection; actual execution used Go 1.26.5.

Rationale: No actionable introduced quality issue is substantiated. The implementations use small, direct branches and return values suited to their stated contracts. Ordinary correct implementation supports A; the presence of several passing checks is not by itself an A+ safeguard claim.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] Reader `load.go:33–45` parses returned text before handling a read failure, preserves the valid prefix and wraps the caller-owned cause. `parseLine` returns an ordinary nil error on success and creates the requested `*LineError` only for malformed records. The external-package test exercises its public fields.
- [G2] Backend `lookup.go:18–27` rejects blank keys before calling the backend, returns `ErrMissing` directly and wraps only `ErrUnavailable`. The original backend diagnostic is absent from both the error chain and message; useful operation/key context remains.
- [G3] CLI library `export.go:20–26` preserves a lone operation or close error directly and uses `errors.Join` only when both fail. The short-write check at `export.go:39–40` avoids a false accepted-record count. `Export` delegates to one implementation.
- [G4] Cursor `collect.go:18–29` keeps scan, filtering and terminal `Err` handling together, with one deferred close. Ignoring close failure is deliberate and matches this cursor's best-effort contract.

Bad

- None found.

Suggested changes

- None needed for the graded scope.

Limits: No full Architecture, Testing, Security or Performance grade is implied. Public documentation is adequate for the requested behavior; optional preferences for additional comments, helper shape, or test organization are not findings. The reader's simultaneous malformed-data/read-error precedence is discussed below as an unresolved contract question, outside the supported grade claim for that combination.

## Correctness & Compatibility — A

Scope: The same four changesets, assessed against original README requirements, preserved signatures and representative caller behavior.

Coverage: Exercised success, partial results, caller errors, malformed input, long reader records, EOF, direct sentinel identity, unavailable-error redaction, limits, accepted counts, short writes, dual operation/close failure, best-effort close, scan/iteration failure and positional CLI paths/streams/status. No concurrent API-use contract is present; race-enabled tests are an additional check, not a concurrency guarantee.

Rationale: All specified private contract probes and additional review checks passed, and source inspection confirms the supported behavior. No confirmed introduced or worsened compatibility defect was found. A is bounded to these assessed contracts; it does not resolve unspecified combinations.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G5] Reader tests preserve two records when final valid bytes accompany a custom read error, expose the cause through `errors.Is`, preserve the structured physical line number after comments/blanks, accept a 100,000-character value, and return nil when final bytes accompany EOF.
- [G6] Backend probes verify direct `err == ErrMissing`, blank-key rejection without a backend call, empty unusable results on backend failure, and exclusion of the backend identity, concrete type and sensitive text. An additional check verifies wrapped `BackendMissing` translation and successful value return.
- [G7] Writer probes verify one close on success/failure, direct identity for a single failure, inspection of both independent failures, valid accepted counts and limits. Additional checks verify no read for negative limits even when close fails, and `io.ErrShortWrite` with count zero for an incomplete write.
- [G8] Cursor probes verify selected order/prefix and terminal iteration error, scan failure and nil filtering. An additional success-plus-close-failure check confirms close remains best-effort rather than incorrectly invalidating collected rows.
- [G9] Six executable CLI invocations verify limited/unlimited output with a leading-dash input filename, missing-input exit 1, argument-count and non-integer usage exit 2, and negative-limit export failure exit 1. Stdout is empty in every case; stderr is empty on success and nonempty on failure.

Bad

- None found.

Suggested changes

- None needed for the graded scope.

Limits: Actual Go 1.22, other platforms and real downstream integrations were not executed. CLI finalization failure was exercised at the library boundary with a writer double, not a real filesystem close failure. The inherited CLI Scanner limit/read-ahead behavior is not an introduced change, and the CLI fixture does not contain the reader fixture's explicit no-line-size-cap requirement. No rollback or atomic-publication guarantee was imposed. Test effectiveness was not mutation-tested or given a separate Testing grade.

## Reproduced ambiguity, not a confirmed defect

Reader `load.go:37–38` gives malformed input precedence over a simultaneous read failure. A reader returning `"a=1\nbroken"` with a custom error produces the valid first record and `*LineError{Line: 2, Text: "broken"}`, while `errors.Is(resultError, readCause)` is false. The reviewer reproduced this observation in `TestReviewerMalformedReadErrorObservation`.

The original README says caller reader errors remain inspectable and bytes returned with a read error must be processed. It does not specify whether malformed-data precedence or an aggregate is required when processing those bytes fails too. The author preserves the required malformed-record report and prefix; the same README also frames termination as malformed input **or** read failure. Both-cause inspection is explicitly required in the writer fixture, but no equivalent aggregation policy is explicit here. Treat the reader combination as a contract clarification opportunity, not a confirmed baseline defect or a reason to train to a newly added hidden assertion. If a future task explicitly requires both causes, evaluate that contract prospectively and disclose the change. The grades above exclude a conclusion on this ambiguous combination.

## Verification performed

All verification used disposable copies in `/private/tmp/error-outcome-review-3b23/`, preserving trial outputs and fixture sources. The environment set `GOCACHE=/private/tmp/error-outcome-review-3b23/gocache` and `GOWORK=off`.

For each of the four case directories, ran:

```sh
rtk proxy go test -race -count=1 -v ./...
rtk proxy go vet ./...
rtk proxy gofmt -l <all author Go files>
```

All four test commands and four vet commands exited 0. All four formatting checks exited 0 with empty output. Tests included the archived private probe file copied as `language_contract_probe_test.go` and six additional reviewer tests described above. The malformed-reader combination test records the observed policy; it does not assert an invented both-errors requirement.

Built the executable with `rtk proxy go build -o /private/tmp/error-outcome-review-3b23/cli-calls/lineexport ./cmd/lineexport`; it exited 0. Executed it through `rtk proxy` using the six argument combinations in G9, checking actual status and file bytes. `rtk proxy go version` reported `go1.26.5 darwin/arm64`.

Raw reviewer check output and the additional probes are local scratch artifacts: four `<case>-verification.json` files, `cli-invocations.json`, and each case's `reviewer_test.go` under `/private/tmp/error-outcome-review-3b23/`. Permanent private probes already live in the repository. This report records the independent results; temporary reviewer evidence may require preservation if later comparison needs exact extra-probe bytes.

## Useful content, costs and promotion recommendation

There remains a coherent potential error-contract subject: deciding which causes are public, separating wrapping from sanitization, preserving direct identity versus aggregate inspection, processing partial results before terminal status, checking buffered/iteration completion, and distinguishing consequential finalization from best-effort cleanup. The [source audit](language-contracts-source-audit.md) captures useful corrections in those areas. Potential content is not measured benefit. The existing [API-contract skill](../../plugins/go-quality-build/skills/go-api-contracts/SKILL.md) already covers identity, exposed causes and usable partial results; ordinary Go knowledge plus the existing catalog was sufficient on these four outcomes.

The requested new public additions are limited to `LineError` with its required fields, `ErrInvalidKey`, `ExportLimit`, and `CollectMatching`. No additional exported aggregate type, classification taxonomy, constructor, package or third-party dependency was added. Existing exported protocol types were preserved. No unnecessary exported complexity was found, so there is no cost defect that a new skill can currently claim to reduce. Command counts, token use and author wall time were not independently measured.

Under the [recorded promotion gate](README.md#evaluation-and-promotion), a candidate needs a clear trigger, distinct useful guidance and evidence of better outcomes. This review supports the first two only as plausible design properties, and supplies no evidence for the third. **Defer drafting/promotion for the current evidence set; retain the candidate and audit as research.** The supplemental cursor task also completed correctly. Do not consume the reserved CSV transfer case merely to search for a failing baseline, and do not keep adding tasks until one produces an apparent win. Any later bounded investigation should be declared explicitly, preserve these successful baselines, and use independent outcome review before claiming effectiveness.

Next action for the controller: record this deferral and its limits in the canonical status, preserve all attempted trial provenance, and continue the approved sequence with value-semantics baseline work. This reviewer changed only this assessment file; the controller owns the progress record.
