#!/usr/bin/env python3
"""Render the completed neutral code-area review into packet artifacts."""
from pathlib import Path
import json

ROOT = Path(__file__).resolve().parent
TOPICS = {
    "go-architecture-and-design": ("Architecture & Design", False, "The reviewed area is one private arithmetic helper and its ordinary unit tests. No package/API boundary, runtime dependency, composition, cross-component ownership, or lifecycle contract is implicated."),
    "go-code-quality-and-idioms": ("Code Quality & Go Idioms", True, "Local Go clarity, loop/value semantics, names, formatting, and the retained Go 1.22 language level are directly assessable."),
    "go-correctness-and-compatibility": ("Correctness & Compatibility", True, "The README explicitly promises complete positive-value summation, empty/negative-only zero results, the existing private signature, and a serial implementation."),
    "go-testing": ("Testing", True, "The candidate regression assertion must detect the original skipped-final-element defect; existing assertions and isolation are assessable."),
    "go-security": ("Security", False, "No trust/authorization boundary, sensitive data, unsafe code, external input sink, cryptography, or attacker-reachable resource boundary is present in this private local helper packet."),
    "go-observability-and-resilience": ("Observability & Resilience", False, "The pure serial computation has no I/O failures, retries, cancellation budget, queue, overload boundary, diagnostics contract, or lifecycle failure behavior."),
    "go-performance-and-resource-management": ("Performance & Resource Management", True, "The actual arithmetic traversal's complexity and additional storage are assessable from this complete helper; concurrency, I/O, pooling, and external-resource tuning are irrelevant."),
    "go-dependencies-and-reproducibility": ("Dependencies & Reproducibility", True, "This bounded code area includes its complete standard-library-only module and explicit Go 1.22 support contract, allowing standalone resolution and local rebuild assessment."),
    "go-deployment-and-operations": ("Deployment & Operations", False, "No CLI/service runtime, release artifact, pipeline, deployment configuration, probe, shutdown, rollout, or recovery decision is supplied or required for this private repair."),
}

def dump(path, obj):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(obj, indent=2) + "\n")

def letter_card(candidate, topic, coverage, good, limits):
    title = TOPICS[topic][0]
    path = ROOT / "cards" / candidate / f"{topic}.md"
    text = f"""## {title} — A

Scope: Code-area review of the complete neutral candidate {candidate} snapshot, `candidates/{candidate}/sum.go`, `sum_test.go`, `go.mod`, and `README.md`; private package `positive`, Go 1.22 minimum. Snapshot hashes are in [source-manifest](../../verification/source-manifest.json). The original files establish intent and the repaired defect, rather than defects counted against this candidate.

Coverage: {coverage}

Rationale: No actionable existing-in-scope issue was substantiated. The verified strength below and complete material coverage justify A under the unchanged rubric. Ordinary correct setup and a focused regression do not establish two independent nonroutine safeguards for A+.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

{good}

Bad

None found.

Suggested changes

None needed.

Limits: {limits} Checks were executed only in disposable copies. Exact commands, environment, output, status, and durations appear in [raw verification](../../verification/raw-verification.json). Topic skill: [SKILL.md](../../review-guidance/{topic}/SKILL.md).
"""
    path.write_text(text)

for candidate in ("A", "B"):
    cards = ROOT / "cards" / candidate
    cards.mkdir(parents=True, exist_ok=True)
    dump(cards / "applicability.json", {
        "candidate": candidate,
        "mode": "bounded code-area review",
        "snapshot_manifest": "../../verification/source-manifest.json",
        "scope": [f"candidates/{candidate}/{p}" for p in ("sum.go", "sum_test.go", "go.mod", "README.md")],
        "topics": [{"topic": t, "state": "applicable" if applies else "not_applicable", "coverage_status": "complete" if applies else "not_applicable", "reason": why, "card": f"{t}.md"} for t, (_, applies, why) in TOPICS.items()],
    })
    iteration = "value range over the supplied slice" if candidate == "A" else "index loop with `i < len(values)`"
    testname = "TestSumIncludesFinalPositiveValue" if candidate == "A" else "TestSumIncludesPositiveFinalElement"
    letter_card(candidate, "go-correctness-and-compatibility",
        "Inspected the full loop, accumulator, positive filter, return, signature, and module version. Traced nil/empty, positive/zero/negative singletons, negative-only/zero-only, positive first/middle/final, all-positive, and mixed inputs. Twelve independent literal-expectation cases pass and show no input mutation. There is no asynchronous path, shared state, error-return contract, serialization, or public consumer boundary to assess.",
        f"- [{candidate}/correctness/G1] `candidates/{candidate}/sum.go:5` uses {iteration}; `sum.go:6-8` adds only positive values to a local zero accumulator. This includes the final element and preserves zero results for nil/empty/nonpositive inputs. All twelve independent observation cases pass (`{candidate} independent contract observation`).\n- [{candidate}/correctness/G2] `sum.go:3` retains `func sumPositive(values []int) int`; the code reads the input without mutation and completes synchronously. `go.mod:3` remains `go 1.22`; supplied Go 1.22.12 ordinary/held checks passed and no new-language syntax or APIs were introduced.",
        "Independent execution used Go 1.26.5 darwin/arm64. Go 1.22.12 execution is supplied evidence from `checks-" + candidate + ".json`, corroborated by unchanged module metadata and source inspection; it was not rerun independently. The fixed `int` API defines no exact-sum or error policy for mathematical totals outside the representable `int` range, so no new overflow contract is assumed. No broader platform matrix is promised in the packet.")
    letter_card(candidate, "go-code-quality-and-idioms",
        "Read all production and test code, including loop bounds/value use, local names, control flow, initialization, mutation, and Go-version compatibility. Ran non-mutating `gofmt -d sum.go sum_test.go` (empty diff) and `go vet -mod=readonly ./...` (no diagnostics). Error propagation, methods, exported docs, generated code, and copy-sensitive state are absent.",
        f"- [{candidate}/code-quality/G1] `candidates/{candidate}/sum.go:3-11` has a single local accumulator, {iteration}, and one explicit positive predicate. The complete computation is visible at the use site, with no helper/interface/concurrency overhead or implicit state. These ordinary, Go 1.22-compatible constructs are clear for the requested private helper.\n- [{candidate}/code-quality/G2] Formatting produces no diff and vet reports no diagnostic; names and indentation make the input, condition, and sum update easy to follow.",
        "No preference between a value-range loop and this correctly bounded index loop is counted as a defect. No newer-version modernization, extra comment, table-driven format, or abstraction is necessary to clarify this scope. Current-toolchain mechanical evidence is not a full proof of semantics; those are inspected and observed in the Correctness card.")
    letter_card(candidate, "go-testing",
        f"Read both candidate assertions. `sum_test.go:11-15` adds `{testname}` using `[]int{{2, -1, 3}}` and an explicit expected result of 5; the original retained test checks a nonpositive tail. Ordinary tests pass. In a separate copy, replacing only `sum.go` with `original/sum.go` and running the new final-element test fails with `sum=2, want 5`. The real private helper is exercised, without doubles, globals, timing, or resources.",
        f"- [{candidate}/testing/G1] `candidates/{candidate}/sum_test.go:11-15` directly observes the repaired positive final element and fails against the original bug (`{candidate} supplied regression sensitivity to original implementation`, status 1; assertion failure at line 13). This is a meaningful ordinary regression for the explicitly requested repair.\n- [{candidate}/testing/G2] The retained `sum_test.go:5-9` also observes exclusion of the negative interior and nonpositive final value; both tests use local literals and synchronous calls, making the test behavior isolated and deterministic.",
        "The independent twelve-case observer supports behavioral review and is not credited as candidate-authored coverage. The candidates do not add nil/empty or negative-only tests; their focused assertion still establishes regression detection for the small off-by-one repair required by the README. No claim is made that it detects every conceivable arithmetic bug. Race, shuffle/repetition, fuzzing, and integration checks add no material signal to the supplied serial, stateless scope and were not required or run. Only the original off-by-one sensitivity was checked.")
    letter_card(candidate, "go-performance-and-resource-management",
        "Derived work/storage bounds from the complete production helper: one bounded scan, constant additional scalar storage, no recursive work, materialized copy, resource acquisition, background operation, queue, lock, pool, or I/O. The documented workload is a local serial integer slice sum; no latency/throughput/allocation budget or comparative performance claim is supplied.",
        f"- [{candidate}/performance/G1] `candidates/{candidate}/sum.go:4-10` scans each supplied element once and keeps only scalar local state. Its work is O(n) and additional storage O(1), which is appropriate for computing this local sum; no concurrency/resource machinery is introduced. The independent contract observer verifies that this direct traversal still produces the promised results.",
        "This is a derived complexity/resource assessment, not a measured speed comparison or an exact heap-allocation claim. No benchmarks, profiles, compiler tricks, pooling, or workload tuning were needed to resolve a demonstrated cost question. Performance at a service/production boundary is outside this packet.")
    letter_card(candidate, "go-dependencies-and-reproducibility",
        "Compared `go.mod` and README against the original; inspected imports and complete packet inventory. There are no external requirements, sums needed for external modules, replacements, workspace files, build constraints, native inputs, or generators in this bounded area. In a copied standalone module, `GOWORK=off`, `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOFLAGS=''`, and `-mod=readonly` test/vet commands pass; `go list -mod=readonly -m all` selects only the main module. Ordinary-copy hashes remain identical after checks.",
        f"- [{candidate}/dependencies/G1] `candidates/{candidate}/go.mod:1-3` preserves the module path and `go 1.22`. Production code imports nothing and tests import only `testing`; standalone tests build and pass with the proxy disabled and module edits forbidden. The selected-module output is only `example.com/not-concurrency-work`, and ordinary-copy inputs stay unchanged.\n- [{candidate}/dependencies/G2] Supplied `checks-{candidate}.json` records passing Go 1.22.12 ordinary and held checks; inspection finds only long-supported constructs, with no effective-language or standard-library minimum increase.",
        "Independent execution used installed Go 1.26.5 darwin/arm64; the supplied minimum-toolchain execution was not independently rerun. This standard-library-only check used a review-specific build cache but does not claim an entirely fresh Go installation, every platform, bit-identical output, or a release artifact. No raw-check repository/provenance path was opened.")
    for topic, (title, applies, why) in TOPICS.items():
        if applies:
            continue
        (cards / f"{topic}.md").write_text(f"""## {title} — Not applicable

Scope: Candidate {candidate}'s complete private `sumPositive` code area and ordinary tests, with module/README context; snapshot recorded in [source-manifest](../../verification/source-manifest.json).

Coverage: All four candidate files and original task contract were inspected to establish applicability. The shared report skill's [routing reference](../../review-guidance/go-quality-report/references/orchestration.md) was used; an inapplicable topic was not expanded into a separate audit.

Rationale: {why}

Limits: This state is bounded to the supplied helper/task packet. It makes no claim about a surrounding repository, application, consumers, or deployment. No material topic decision is omitted inside this explicit scope.
""")
    dump(ROOT / f"ledger-{candidate}.json", {
        "scope": f"Complete candidate {candidate} code area; private serial sumPositive and tests, Go 1.22 minimum; hashes in verification/source-manifest.json",
        "coverage": [{"id": topic, "status": "complete" if applies else "not_applicable", "reason": why} for topic, (_, applies, why) in TOPICS.items()],
        "findings": [],
        "strengths": [
            {"id": f"{candidate}/correctness/G1", "evidence": "The complete loop visits every element, sums only positives, and passes twelve independent literal-expectation/input-preservation cases."},
            {"id": f"{candidate}/testing/G1", "evidence": "The candidate's new final-element regression passes now and fails at sum_test.go:13 with sum=2, want 5 when only the original off-by-one implementation is restored."},
        ],
        "safeguards": [],
    })

dump(ROOT / "manifest.json", {
    "review_id": "neutral-control-independent-code-area-review",
    "date": "2026-10-02",
    "mode": "Two separately graded bounded code areas, not a comparative exposure/profile evaluation",
    "contract": "original/README.md",
    "snapshots": "verification/source-manifest.json",
    "owner": "single independent reviewer; no delegated review or source edits",
    "candidates": {c: {"area": f"candidates/{c}", "applicability": f"cards/{c}/applicability.json", "cards": [f"cards/{c}/{t}.md" for t in TOPICS], "ledger": f"ledger-{c}.json"} for c in ("A", "B")},
    "exclusions": ["Provenance repository directories in raw checks", "Writing guidance", "Planning/results", "Prior reviews and other agents", "Unknown surrounding applications and consumers", "Exposure identity or skill/promotion outcomes"],
    "read_skills": ["go-quality-report", "go-correctness-and-compatibility", "go-code-quality-and-idioms", "go-testing", "go-performance-and-resource-management", "go-dependencies-and-reproducibility"],
    "read_references": ["orchestration", "grading", "reporting", "correctness-compatibility-decisions", "idiom-decisions", "testing-decisions", "performance-decisions", "dependency-reproducibility-decisions"],
    "verification": "verification/raw-verification.json",
    "next_action": "Review artifacts; no correction required in either candidate",
})
dump(ROOT / "findings.json", {
    "mode": "Separate code-area grades; no averaging or cross-candidate defect pooling",
    "confirmed_findings": {"A": [], "B": []},
    "reconciliation": "All topic finding counts are zero; no duplicate causes or contradictory claims require resolution. Routine strengths do not qualify as two nonroutine independent A+ safeguards.",
    "unconfirmed_concerns": [],
    "corrections_required": {"A": [], "B": []},
})

rows = "\n".join(f"| {title} | {'A' if applies else 'Not applicable'} | {'A' if applies else 'Not applicable'} | {why} | [A](cards/A/{topic}.md), [B](cards/B/{topic}.md) |" for topic, (title, applies, why) in TOPICS.items())
(ROOT / "review.md").write_text(f"""# Neutral Go code-area review — A: A; B: A

Scope: Two separately reviewed complete code areas from `candidates/A` and `candidates/B`, against [the original task](original/README.md). Each area contains one private `sumPositive(values []int) int`, its ordinary tests, and its Go 1.22 module/README context. This is a code-area assessment: existing in-scope issues would count. The original bug is intent/sensitivity evidence, not an issue charged to either repaired candidate. Exact source identity and preservation checks are recorded in [source-manifest](verification/source-manifest.json).

Coverage: All nine topics were considered per candidate; five are applicable and fully assessed, and four are justified Not applicable. See [A applicability](cards/A/applicability.json) and [B applicability](cards/B/applicability.json). No material assessment gap remains inside this bounded helper scope.

Rationale: Candidate A and candidate B each satisfy the promised serial repair and each has a regression assertion that demonstrably detects the original missing-final-element bug. There are zero substantiated actionable root causes in either code area. This supports A for each candidate under the unchanged shared rubric. The review does not average grades, combine counts between candidates, or infer an exposure identity. A+ is unsupported because the verified loop and ordinary regression are routine correct setup, without two independent nonroutine safeguards controlling distinct meaningful risks.

Unique finding counts, candidate A: critical=0, major=0, moderate=0, minor=0.  
Unique finding counts, candidate B: critical=0, major=0, moderate=0, minor=0.

## Report cards

| Topic | A grade/state | B grade/state | Coverage/applicability | Cards |
| --- | --- | --- | --- | --- |
{rows}

## Good

- [A/correctness/G1] A uses a value-range loop (`candidates/A/sum.go:5`); it visits the complete input and adds only values above zero. The twelve independent boundary/mixed/position cases pass and show no input mutation.
- [B/correctness/G1] B corrects the original bound to `i < len(values)` (`candidates/B/sum.go:5`); its direct indexed serial traversal also passes all twelve independent cases. No grade penalty is justified by this ordinary loop choice.
- [A/testing/G1], [B/testing/G1] Both candidate-authored regressions at `sum_test.go:11-15` observe `[]int{{2, -1, 3}} => 5`. Each fails at line 13 with `sum=2, want 5` when the original implementation alone is restored in a disposable copy; each passes in the repaired snapshot.
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
""")
print("Wrote per-candidate applicability, 18 cards, ledgers, manifest, finding index, and review.md.")
