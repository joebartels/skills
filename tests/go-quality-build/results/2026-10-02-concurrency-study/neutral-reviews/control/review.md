# Neutral Go code-area review — A: A; B: A

Scope: Two separately reviewed complete code areas from `candidates/A` and `candidates/B`, against [the original task](original/README.md). Each area contains one private `sumPositive(values []int) int`, its ordinary tests, and its Go 1.22 module/README context. This is a code-area assessment: existing in-scope issues would count. The original bug is intent/sensitivity evidence, not an issue charged to either repaired candidate. Exact source identity and preservation checks are recorded in [source-manifest](verification/source-manifest.json).

Coverage: All nine topics were considered per candidate; five are applicable and fully assessed, and four are justified Not applicable. See [A applicability](cards/A/applicability.json) and [B applicability](cards/B/applicability.json). No material assessment gap remains inside this bounded helper scope.

Rationale: Candidate A and candidate B each satisfy the promised serial repair and each has a regression assertion that demonstrably detects the original missing-final-element bug. There are zero substantiated actionable root causes in either code area. This supports A for each candidate under the unchanged shared rubric. The review does not average grades, combine counts between candidates, or infer an exposure identity. A+ is unsupported because the verified loop and ordinary regression are routine correct setup, without two independent nonroutine safeguards controlling distinct meaningful risks.

Unique finding counts, candidate A: critical=0, major=0, moderate=0, minor=0.  
Unique finding counts, candidate B: critical=0, major=0, moderate=0, minor=0.

## Report cards

| Topic | A grade/state | B grade/state | Coverage/applicability | Cards |
| --- | --- | --- | --- | --- |
| Architecture & Design | Not applicable | Not applicable | The reviewed area is one private arithmetic helper and its ordinary unit tests. No package/API boundary, runtime dependency, composition, cross-component ownership, or lifecycle contract is implicated. | [A](cards/A/go-architecture-and-design.md), [B](cards/B/go-architecture-and-design.md) |
| Code Quality & Go Idioms | A | A | Local Go clarity, loop/value semantics, names, formatting, and the retained Go 1.22 language level are directly assessable. | [A](cards/A/go-code-quality-and-idioms.md), [B](cards/B/go-code-quality-and-idioms.md) |
| Correctness & Compatibility | A | A | The README explicitly promises complete positive-value summation, empty/negative-only zero results, the existing private signature, and a serial implementation. | [A](cards/A/go-correctness-and-compatibility.md), [B](cards/B/go-correctness-and-compatibility.md) |
| Testing | A | A | The candidate regression assertion must detect the original skipped-final-element defect; existing assertions and isolation are assessable. | [A](cards/A/go-testing.md), [B](cards/B/go-testing.md) |
| Security | Not applicable | Not applicable | No trust/authorization boundary, sensitive data, unsafe code, external input sink, cryptography, or attacker-reachable resource boundary is present in this private local helper packet. | [A](cards/A/go-security.md), [B](cards/B/go-security.md) |
| Observability & Resilience | Not applicable | Not applicable | The pure serial computation has no I/O failures, retries, cancellation budget, queue, overload boundary, diagnostics contract, or lifecycle failure behavior. | [A](cards/A/go-observability-and-resilience.md), [B](cards/B/go-observability-and-resilience.md) |
| Performance & Resource Management | A | A | The actual arithmetic traversal's complexity and additional storage are assessable from this complete helper; concurrency, I/O, pooling, and external-resource tuning are irrelevant. | [A](cards/A/go-performance-and-resource-management.md), [B](cards/B/go-performance-and-resource-management.md) |
| Dependencies & Reproducibility | A | A | This bounded code area includes its complete standard-library-only module and explicit Go 1.22 support contract, allowing standalone resolution and local rebuild assessment. | [A](cards/A/go-dependencies-and-reproducibility.md), [B](cards/B/go-dependencies-and-reproducibility.md) |
| Deployment & Operations | Not applicable | Not applicable | No CLI/service runtime, release artifact, pipeline, deployment configuration, probe, shutdown, rollout, or recovery decision is supplied or required for this private repair. | [A](cards/A/go-deployment-and-operations.md), [B](cards/B/go-deployment-and-operations.md) |

## Good

- [A/correctness/G1] A uses a value-range loop (`candidates/A/sum.go:5`); it visits the complete input and adds only values above zero. The twelve independent boundary/mixed/position cases pass and show no input mutation.
- [B/correctness/G1] B corrects the original bound to `i < len(values)` (`candidates/B/sum.go:5`); its direct indexed serial traversal also passes all twelve independent cases. No grade penalty is justified by this ordinary loop choice.
- [A/testing/G1], [B/testing/G1] Both candidate-authored regressions at `sum_test.go:11-15` observe `[]int{2, -1, 3} => 5`. Each fails at line 13 with `sum=2, want 5` when the original implementation alone is restored in a disposable copy; each passes in the repaired snapshot.
- [A/code-quality/G1], [B/code-quality/G1] Both implementations keep the complete computation local, preserve the private signature and Go 1.22 declaration, and add no concurrency or abstraction. Vet has no diagnostics and formatting has no diff.
- [A/performance/G1], [B/performance/G1] Both derive O(n) work and O(1) additional storage from a single scalar accumulation pass. Standalone module tests pass with workspace/proxy disabled and readonly module mode; selected-module output contains only the main module.

## Bad

None found for A or B. There are no canonical defect IDs or severity counts to reconcile.

## Suggested changes

None needed for A or B within this supplied task/code area.

## Limits

This review inspected only the neutral packet and used disposable copies; it opened none of the repository directories in raw-check provenance and used no writing guidance, planning/results, earlier review, or other-agent evidence. A and B were assessed sequentially by one independent reviewer. Source SHA-256 values match before/after, and ordinary-copy source/module files also remain unchanged.

Independent execution used Go 1.26.5 darwin/arm64. Supplied `checks-A.json` and `checks-B.json` record passing Go 1.22.12 ordinary/held checks; the retained `go 1.22` module and long-supported source constructs corroborate compatibility. Go 1.22 was not independently executed. No broader platform, release, or byte-identical-artifact claim is made.

Exact commands/environment/results are in [raw verification](verification/raw-verification.json). For each candidate, `rtk proxy go test -mod=readonly -count=1 -timeout=30s ./...`, `rtk proxy go vet -mod=readonly ./...`, `rtk proxy gofmt -d sum.go sum_test.go`, and `rtk proxy go list -mod=readonly -m all` pass. With the [independent twelve-case observer](verification/contract_observation_test.go) added only to a copied module, `rtk proxy go test -mod=readonly -count=1 -timeout=30s -run '^TestReviewContractCases$' -v ./...` passes. In the restored-original copies, `rtk proxy go test -mod=readonly -count=1 -timeout=30s -run Final -v ./...` fails as expected on the new regression assertion, rather than failing to build.

The observer is review evidence and is not candidate-authored test coverage. Candidate tests focus on the specifically requested final-element repair and retained nonpositive-tail behavior. No race/repetition/fuzz/integration/profile run was needed for the supplied serial stateless function. Integer totals beyond the representable `int` range have no specified error or exact-sum policy under the fixed API; no wider arithmetic guarantee is inferred.

Details: [manifest](manifest.json), [finding index](findings.json), [A ledger](ledger-A.json), [B ledger](ledger-B.json), and [verification runner](verification/verify.py). The packet's unchanged grade calculator is used separately for each ledger; outputs are stored under `verification/grade-A.json` and `verification/grade-B.json`.
