# Independent Go changeset review

Reviewed on 2026-10-01. Supplied directory diff: `/private/tmp/go-independent-review-otiblxwu/original` → `/private/tmp/go-independent-review-otiblxwu/candidate`. No author report, evaluation expectations, other candidate, or unrelated repository material was consulted. Only `store_test.go` changes; `README.md`, `go.mod`, and `store.go` are identical. Locations below refer to candidate files.

The supplied request is to extend Store tests for concurrent independent roots and grouped child cases, preserve useful serial checks and resource lifetime, and keep the public API, standard-library dependencies, and Go 1.22 minimum. README lines 3–15 establish these contracts. Same-key concurrent replacement within one root is expressly outside scope.

## Testing — B

Scope: Supplied original/candidate diff for the `example.com/filestore` library, principally `candidate/store_test.go:10–73`; module language Go 1.22, executed with local Go 1.26.5 on darwin/arm64.

Coverage: Assessed the existing serial replacement checks, real filesystem boundary, new parallel independent roots, grouped serial replacement children, repeated same-key values, empty/newline/NUL strings, missing-key checks, final persistence reads, fixture cleanup ownership, Go-version-sensitive captures, focused child selection, repeated/shuffled runs, race detection, and whether realistic wrong behavior is rejected. No fuzzing, benchmarks, new dependencies, or production seams were added. Existing invalid-key test coverage is outside this requested extension and was not treated as introduced debt.

Rationale: One introduced moderate issue causes misleading failures in ordinary focused subtest runs. The complete suite meaningfully verifies the requested contracts and passes race/shuffle repetition, but strengths do not cancel that actionable test-isolation defect. Counted severities select B; the issue does not invalidate full-suite regression detection or establish major contract loss.

Finding counts: critical=0, major=0, moderate=1, minor=0

Good

- [T-G1] `candidate/store_test.go:32–34,49–60` preserves exact values through repeated replacements, including empty strings and newline/NUL content, using the exported API and actual filesystem. In disposable `verification/lossy-values`, changing Get to remove NUL bytes caused `empty/replacements/value-1` to fail at the value comparison (check `mutation-lossy-values`, exit 1). This confirms useful assertions for unchanged complete strings rather than relying on test count or coverage percentage.
- [T-G2] `candidate/store_test.go:36–38,44–48,66–70` assigns distinct roots and uses the same key across distinct payloads, with initial missing-key and post-group final-state checks. Mutating New to use `filepath.Dir(root)` shared the roots and failed deterministically with `-parallel=1`: later missing-key checks succeeded unexpectedly, and beta/empty final values were alpha (check `mutation-shared-root`, exit 1). This verifies detection of cross-root contamination independently of a lucky concurrent interleaving.
- [T-G3] `candidate/store_test.go:38,42–64,66–70` makes the outer test own every temporary root and waits for the parallel instance group before reading final values. Moving root ownership to the instance child in a disposable copy preserved the child round trips but removed the files before outer final reads; all three final reads failed (check `mutation-early-cleanup`, exit 1). This verifies lifetime-sensitive assertions. The cleanup behavior is consistent with the [testing package contract](https://pkg.go.dev/testing#T.TempDir).
- [T-G4] `candidate/store_test.go:26,45` schedules independent roots concurrently, while each root's replacement children remain serial. Thirty complete race/shuffled runs with `-parallel=8` passed. A separate disposable observation copy added a mutex-protected first-Put barrier requiring three distinct roots to reach Put before any could proceed; the independent-instance test passed under `-race` with a five-second test timeout (check `candidate-overlap-observation`, exit 0). This establishes that the test can exercise simultaneous independent instance work, without introducing sleeps or altering the candidate.
- [T-G5] `candidate/store_test.go:10–23` retains the original useful serial replacement checks and replaces manual temporary-directory cleanup with test-owned cleanup. Both original and candidate complete suites pass.

Bad

- [F1][moderate][introduced] The final reads at `candidate/store_test.go:66–70` assume every instance and every replacement child ran. Selecting only `instances/alpha` with normal `go test -run` succeeds inside that selected child, then falsely fails the outer test because beta and empty never wrote their keys. Selecting only `alpha/replacements/value-1` additionally falsely compares the valid current value `"alpha replacement"` with the excluded last row's value `"alpha"`. Checks `candidate-focused-instance` and `candidate-focused-leaf` both exit 1 on the unchanged, otherwise passing candidate. This couples named cases to excluded siblings and makes focused diagnosis misleading. Both symptoms have one cause: final expectations are based on the full table rather than executed work. Primary remediation owner: Testing.

Suggested changes

- [F1] Preserve post-group persistence checks, but record per-instance whether a successful Put ran and the last value actually written by an executed serial child. After the group completes, inspect only exercised instances against that recorded value. Another structure is acceptable if it keeps the lifetime and cross-root checks while making selected children independently runnable. Verify the two focused commands below and the complete race/shuffled suite; retain root-sharing and early-cleanup regression detection.

Limits: The exercised platform/toolchain is Go 1.26.5 darwin/arm64; the actual Go 1.22 toolchain and other platforms were not run. Go commands used the mandated cache and `GOTOOLCHAIN=local`, with module/network resolution disabled. Race detection establishes no race in the executed paths, not all possible consumer uses. Mutation checks demonstrate these specific regression signals, not exhaustive correctness. Exact commands, separate stdout/stderr, exits, and working directories are preserved in `output/checks.json`.

## Correctness & Compatibility — A

Scope: Supplied test-only library changeset, README public Store contracts, supported Go 1.22 minimum, and changed test build/lifetime behavior. Production source, exported signatures, module directive, and dependency declarations are unchanged.

Coverage: Compared all supplied files; traced serial and different-root concurrent calls through real Put/Get; assessed cleanup and group completion before final reads; inspected language/API compatibility of added fmt/testing usage; ran module-language Go 1.22 compilation, the standard-version vet analyzer, complete tests, and exercised race checks. Same-root concurrent replacement remains outside the documented task. No production repair or broader legacy audit was performed.

Rationale: No introduced or worsened production behavior, public API incompatibility, dependency change, or supported-language violation was found. The unchanged source and verified Go 1.22 language/API checks provide a relevant compatibility strength, selecting A. F1 is a testing-selection defect and is not counted again as consumer-visible Store breakage. A+ is not warranted by this test-only production diff.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/store.go:9–42` and `candidate/go.mod:1–3` are byte-identical to original. Store continues to derive paths from its supplied root, and no exported declaration or module/dependency setting changes. Complete tests and race/shuffle repetition pass; final preservation comparisons show both supplied source directories remain unchanged relative to their pristine verification copies.
- [C-G2] `candidate/store_test.go:43–45` captures the per-iteration instance value safely under the declared Go 1.22 module language. This is supported by the [Go 1.22 language changes](https://go.dev/doc/go1.22). Compiling this module's packages with `-lang=go1.22` passes, and `go vet -stdversion ./...` reports no too-new standard-library usage. The added `t.TempDir` is available since Go 1.15 according to the [testing API](https://pkg.go.dev/testing#T.TempDir).
- [C-G3] `candidate/store_test.go:38,42–64` preserves roots and ensures child completion before outer reads in the complete suite; the lifecycle mutation verifies why this ownership matters. No API lifetime change is introduced: `candidate/store.go:12` continues to place root ownership on callers.

Bad

- None found.

Suggested changes

- None needed in production/API compatibility scope.

Limits: Actual Go 1.22 runtime/toolchain and other platforms were not executed. Compatibility evidence consists of unchanged production/module files, compatible inspected APIs, module-specific Go 1.22 language compilation on Go 1.26.5, and the standard-version vet check. An initially overbroad reviewer command using `-gcflags=all=-lang=go1.22` failed in the installed Go 1.26 standard library's Go 1.23 iterator syntax; that is not a candidate defect. The corrected module-specific command passes. No toolchain download was attempted. Network/module resolution was disabled for Go checks. Exact outputs are retained rather than hidden.

Ungraded related finding: Testing [F1] causes false failures when selecting child tests; its correction and grading are owned by the Testing report card.

## Architecture & Design — Not applicable

Scope: Supplied test-only diff; unchanged production Store type, New/Put/Get API, root ownership documentation, package layout, module language and dependencies.

Coverage: Checked the complete diff for consequential changes to production seams, package boundaries, composition, and lifetime decisions. All fixture changes remain within external-package tests.

Rationale: No production seam, package, API-design, or production lifetime decision changes. The dispatch's conditional architecture review does not apply; the new test fixture lifetime and grouping are assessed in Testing.

Limits: No general architecture audit was requested or performed.

## Verification facts

All verification and mutations ran in disposable copies under `/private/tmp/go-independent-review-otiblxwu/verification`. The candidate and original were never repaired or changed. Go checks were prefixed exactly with `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off`. `output/checks.json` stores full command text/argv, cwd, exact stdout, exact stderr, exit code, and duration for every check below.

| Check ID | Check and observed result |
| --- | --- |
| `toolchain` | `go version`: Go 1.26.5 darwin/arm64, exit 0. |
| `original-serial` | Original `go test ./... -count=1 -timeout=30s`: passes, exit 0. |
| `candidate-full` | Candidate complete suite, same flags: passes, exit 0. |
| `candidate-race-shuffle` | `go test ./... -race -count=30 -shuffle=on -parallel=8 -timeout=60s`: passes, exit 0. |
| `candidate-focused-instance` | Select only `TestStoreIndependentInstances/instances/alpha`: outer beta/empty final reads fail, exit 1. |
| `candidate-focused-leaf` | Select only alpha replacement value-1: outer final alpha comparison plus beta/empty reads fail, exit 1. |
| `candidate-vet` | `go vet ./...`: clean, exit 0. |
| `mutation-shared-root` | New uses `filepath.Dir(root)`; missing-key and final isolation checks reject it, exit 1. |
| `mutation-lossy-values` | Get strips NUL; exact value-1 assertion rejects it, exit 1. |
| `mutation-early-cleanup` | Roots owned by instance children, outer pointers retained; final reads reject deletion before parent work, exit 1. |
| `candidate-go122-language` | Reviewer overbroad `all=-lang=go1.22` fails installed standard library iterator syntax, exit 1; not attributed to candidate. |
| `candidate-go122-module-language` | `go test ./... -gcflags=example.com/filestore...=-lang=go1.22 -count=1 -timeout=30s`: passes, exit 0. |
| `candidate-stdversion` | `go vet -stdversion ./...`: clean, exit 0. |
| `candidate-overlap-observation` | Disposable first-Put barrier requires three distinct roots active; test passes under race with 5s timeout, exit 0. |
| `original-preserved` | `rtk proxy diff -ru` original vs pristine verification/original: no differences, exit 0. |
| `candidate-preserved` | `rtk proxy diff -ru` candidate vs pristine verification/candidate: no differences, exit 0. |

Exact focused reproductions from the unmodified candidate verification copy:

```sh
rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off go test ./... -run '^TestStoreIndependentInstances$/^instances$/^alpha$' -count=1 -timeout=30s
rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off go test ./... -run '^TestStoreIndependentInstances$/^instances$/^alpha$/^replacements$/^value-1$' -count=1 -timeout=30s
```
