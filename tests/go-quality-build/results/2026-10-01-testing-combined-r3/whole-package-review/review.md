# Independent whole-package review — Go quality build testing group

Reviewed 2026-10-02. Base `b76df28ed15a177d436672d47a5d30f6634900c7`; head `f48c68ddbdd31931cad3522aa6f9f1f17ea86d06`. One independent reviewer, no delegation. Checkout, index and HEAD remained read-only for this reviewer; all writes and executable verification copies are under `/private/tmp/go-testing-final-package-review/`. The executor's later local-delivery ruling is assessed as bookkeeping, not a changed runtime revision.

## Strengths

- The two skills have useful, distinct selection boundaries. `plugins/go-quality-build/skills/go-behavior-tests/SKILL.md:13` owns observations and independent assertions; `plugins/go-quality-build/skills/go-test-isolation/SKILL.md:12` owns fixture fidelity, process state, event control and lifetime. Pure deterministic and comment-only controls do not acquire unnecessary infrastructure. API, composition and error policy remain separate; existing architecture and review runtime files are byte-unchanged across the range.
- Guidance protects consequential behavior rather than test counts or framework conventions. Behavior `SKILL.md:34` covers exact accepted sets, valid interface values, retained state during partial effects and independent work/cleanup errors. Isolation `SKILL.md:20` covers selectable descendants, early exits and builder/application environment separation. Advice about redirect policy is conditional on the actual promise, not a general rejection of HTTP redirects.
- The shipped main/reference files match final behavior revision2 and isolation revision3 snapshots exactly. Both manifests consistently declare 0.2.0; five canonical runtime skills, catalog paths, local references and Claude validation agree. Evaluation machinery remains outside the runtime package.
- Evidence is reconstructable rather than a collection of uncheckable grades. Fresh verification reconstructed all 49 candidates from frozen inputs and patches, verified input commits/dispatch/guidance/source maps, all 2,197 artifact hashes in ten complete seals, and every seal-index digest against its external canonical-record anchor. Final integrated3 correctly remains unsealed for this review and delivery bookkeeping.
- The source audit has 79 section decisions, a pinned upstream revision, conditional local ownership and explicit original wording/example scope. Reading the shipped guidance and audit identified no substantial copied source requiring a new notice. The audit distinguishes general concepts from copied expression and retains the earlier missing-inventory correction. This is a source-review conclusion, not independent legal advice.
- Public documentation preserves the narrower basis for utility: repeated CLI/worker assertion improvements and focused-fixture selection improvement; strong controls; mixed historical arms; disclosed launch-label bias; supplementary mutation survivors; unavailable minimum toolchain, runtime-loading and platform checks. Neither the A+ cards nor mutation counts are sold as universal coverage.

## Issues

### Critical

None substantiated.

### Important

None substantiated.

### Minor

None requiring a code or guidance change. The raw-evidence whitespace exceptions below should accompany closeout; changing sealed raw bytes to satisfy whitespace tooling would be counterproductive.

## Verification and the five review targets

Exact commands, stdout/stderr, durations and exits are preserved in `checks.json`, with the approved HTTP rerun in `http-approved.json`. `integrity.json`, `anchors.json`, `package-checks.json`, and `whitespace-exclusions.json` preserve additional results. Reproduction entry points are `verify.py`, `supplement.py` and `defaults.py`; loopback-dependent steps require the same execution permission. Reviewer mutations reside only in disposable copies and do not alter archive candidates.

1. **Independent wire observations:** final codec candidate and frozen `TestWireRepresentationIndependent` probe pass. A fresh compiling coordinated encoder/decoder field swap still passes `TestSimpleRoundTrip`, while candidate-only assertions fail at `key_test.go:38` and `key_test.go:66` with concrete wrong wire/field values. This directly demonstrates the intended distinction.
2. **Real failure process status and prefix:** final CLI candidate and frozen `TestRejectedBatchProcess` probe pass. A fresh compiling `os.Exit(2)` to `os.Exit(0)` mutation fails candidate-only process assertions, including retained-prefix batch rejection and partial-write cases (`main_test.go:130`, `main_fsize_test.go:36`). The tests observe actual executable status and files rather than just helper errors.
3. **HTTP boundary/cancellation/body lifetime:** inspected final success/error/body and real server tests, then ran the unchanged final candidate plus the frozen `TestHTTPCancelAndBodyLifetime` probe with race detection. Restricted runs preserved bind-denial panics; an approved finite loopback rerun passed. Synthetic body ownership and real network cancellation remain separate claims.
4. **Parallel fixture and teardown lifetime:** final file and worker candidate suites plus `TestParallelFixtureLifetime` and `TestTeardownOrder` probes pass; both passed three shuffled race repetitions. The named nested `gamma/replacement` file case passes at parallelism1. Source inspection confirms parent-owned directories survive joined descendants and expectations reflect executed contract cases. Worker gates are released and cooperative callbacks joined before fixture closure; non-cooperating work cannot be made safe merely by timing out.
5. **Supported APIs and completion-relative recurrence:** final worker candidate and `TestCompletionRelativeRecurrence` probe pass, including three shuffled race repetitions. The candidate records the callback's final action and checks a lower interval bound to the next start, holds first work past an interval, and permits late scheduling. Runtime illustrations use Go1.22-era APIs; newer optional APIs are labelled. This review executed Go1.26.5 darwin/arm64 with `GOTOOLCHAIN=local`, not an actual Go1.22 binary. I did not mistake module declarations or a newer host pass for minimum-toolchain execution.

Fresh shared verification passes root6/build10/layout1/grade12 tests, repository validation (26 build cases and 10 review skills/120 review cases), both Claude validators, both runtime quick validations and local links. Both unchanged final integrated command suites compile and run successfully with `GOCACHE` unset and a valid disposable `HOME`; their initial loopback denials and approved successful reruns are both recorded. No fresh ordinary candidate defect was observed. Two intentional compiling mutants fail relevant assertions, as expected.

The package's cleanup/version statements were cross-checked against [testing](https://pkg.go.dev/testing#T.Cleanup), [synctest](https://pkg.go.dev/testing/synctest), and [HTTP redirect response ownership](https://pkg.go.dev/net/http#ErrUseLastResponse). Cleanup ordering, optional API version labels, external-I/O limitations and redirect-body ownership agree with that documentation.

Whole-range whitespace passes after exact preserved-raw exclusions recorded in `whitespace-exclusions.json`: `**/source.patch`; the original combined both `blind-review.md` and `review-artifacts/inspection-transcript.json`; behavior skill-on service `blind-review.md` and five raw reviewer scripts (`check_source_preserved.py`, `run_check.py`, `setup_mutations.py`, `setup_verification.py`, `validate_artifacts.py`); and final isolation3 file `author-artifacts/store_test.diff`. The first check with only the originally supplied exclusions is retained as a failure. The additional six files contain blank EOF lines; the raw diff contains normal context-prefix whitespace. Runtime, candidate source and maintained documentation are clean under this check.

## Ledger rulings and costs

1. Treating “Looks good” as plan approval and using native execution is consistent with the accepted handoff. No irreversible publication followed; changing execution method would primarily cost coordination.
2. Reusing the externally managed detached worktree provides existing isolation. A named delivery branch remains a later integration choice, not missing implementation.
3. Independent baseline/skill-on work is a sensible behavioral RED/GREEN gate for authoring guidance. Synthetic fixture implementation is not shipped production; preparation/probe quality remains a real risk, mitigated here by independent cards, frozen sources and reproduced assertions.
4. Inherited settings with unavailable exact IDs limits cross-session reproducibility. Recording that limit is honest; it does not justify a stronger same-model causal claim.
5. Earlier arm-cuing review labels reduce causal grade evidence. Preserving them, reproducing specific effects and using neutral later launches supports the bounded utility claim; it would not support a broadly unbiased grade-uplift estimate.
6. Literal one-wire-GET/initial-200 interpretation is defensible for the integrated fixture. Its unusual strictness is kept conditional in runtime advice and separate from ordinary supplied-client policy, avoiding a general behavior change.
7. Top-level controller name adapters resolve test-name collisions without changing assertion bodies or candidates. Discovery/order can change, so raw failures and reversible mappings matter; this remains disclosed rather than erased.
8. Local commits with no PR/push/global installation fit Task6's explicit local deliverable. The cost of later naming a branch/publishing is accepted future integration work, not a reason to block this package.

## Recommendations

Complete the already planned closeout: preserve these review artifacts, record the expanded exact whitespace exclusions, update continuation status and seal the final integrated3 archive after its final contents are known. Keep subsequent effectiveness claims bounded to the observations above. No runtime revision, additional behavioral campaign, rubric change or universal test convention is requested by this review.

## Declined to judge

- Actual Go1.22 runtime execution and optional Go1.24/1.25 example execution: binaries were unavailable; API/version review and host execution are narrower evidence.
- Other platforms, every scheduling interleaving, external DNS/TLS/services and live database transaction isolation: not exercised here. Local-loopback, POSIX/file/signal and bounded race runs cannot establish these claims.
- Automatic skill routing and actual Codex/OpenCode loading: preserved capability checks do not expose a supported nonmutating load validation; installation/configuration changes were outside the approved work.
- A fully arm-blinded causal grade-uplift estimate or exact cross-session model reproduction: earlier review labels and unavailable inherited IDs prevent those conclusions. Specific replicated improvements remain assessable.
- The surviving arbitrary custom-cancellation-cause comparison mutation as a current production defect, or a universal pure-custom-cause nil-versus-returned-error policy: a future sensitivity gap does not establish either; current legal-input outcomes and the actual contract are the relevant bounds.
- Unbounded non-cooperating goroutine termination by cleanup timeout: Go provides no such guarantee; the runtime reference states the limit and points to process containment where required.
- Independent legal certainty about licensing: checked actual runtime/audit copy scope and found no substantial copy, but did not perform a legal opinion or independently refetch every pinned upstream section.
- Full fresh replay of every historical author/reviewer/mutant command: the immutable archives and full hash/reconstruction checks were verified; representative fresh behavioral checks were selected rather than repeating the historical matrix without a new concern.
- The older blocked composition zero-value trial and future production error/context/concurrency skills: remain explicitly incomplete/proposed and were not silently counted as this delivery's successes.
- Final integrated3 seal and Task6 final commit before this report exists: intentionally pending closeout, not a missing implementation finding.

## Assessment

**Ready to merge? Yes.** The shipped scope follows the approved plan, exact accepted guidance matches evaluated snapshots, package checks pass, and fresh independent behavioral verification supports its carefully bounded claims. No Critical, Important or Minor code/guidance defect was substantiated. Finish the planned evidence seal and continuation bookkeeping before describing delivery as closed; that does not require changing the reviewed runtime bytes.
