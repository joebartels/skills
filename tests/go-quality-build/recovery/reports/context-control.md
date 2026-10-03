# Pure calculation selection/control trial

Date: 2026-10-03

## Scope and result

Implemented only the approved `context-control/README.md` task in this workspace. `SumPositive` now includes a value only when `value > 0`; its exported signature, `calculation` package, module path, `go 1.22` directive, and dependency-free module remain unchanged. The existing positive-input test remains. The new mixed-input regression uses `[-5, 2, 0, -1, 3]` and independently expects `5`.

The implementation remains a local synchronous calculation: no context, timers, goroutines, interfaces, constructors, package movement, or dependencies were added.

## Catalog routing for the actual calculation

Read the `name` and `description` frontmatter of all six supplied skills before choosing bodies.

| Skill | Decision | Reason |
| --- | --- | --- |
| `go-api-contracts` | Selected; body opened | `SumPositive` is exported, and its caller-visible result is being corrected to the README contract. Preserve its exact existing API and supported module/version shape. |
| `go-behavior-tests` | Selected; body opened | This is a behavioral bug fix with a requested regression assertion. Prove the assertion detects the old sum-all-values behavior. |
| `go-context-and-deadlines` | Declined for calculation | Pure calculation with local inputs; no cancellation, budget, cause, or required post-cancellation work. Its body was opened separately for the two contract-card routing decisions below. |
| `go-interfaces-and-composition` | Declined; body unopened | No interface, dependency, constructor, options, wiring, or resource lifecycle decision. |
| `go-package-boundaries` | Declined; body unopened | No package/responsibility/import-direction change. |
| `go-test-isolation` | Declined; body unopened | Deterministic ordinary assertions over local integer inputs; no shared state, files, external doubles, timers, goroutines, or parallel setup. |

The API body informed preservation checks; the behavior-test body informed the independent expected value and red-before-green check. No blanket six-skill application occurred.

## Separate context contract-card decisions

Read only the README contracts at these two paths. No source, probes, history, reports, or implementation from either task was inspected, and neither task was implemented.

| Contract card | Context skill decision | Contract evidence |
| --- | --- | --- |
| `/Users/jb/.codex/worktrees/0bc1/skills/tests/go-quality-build/go-context-and-deadlines/evals/files/http-stages/README.md` | Selected; context body opened | One caller-derived total deadline, nested stage limits, entry cancellation, active-request/body propagation, failure-time cancellation classification/custom cause, and scopes live through body closure. |
| `/Users/jb/.codex/worktrees/0bc1/skills/tests/go-quality-build/go-context-and-deadlines/evals/files/accepted-finalization/README.md` | Selected; same opened context body reused | Cancellation before admitting more work, retained accepted count, completed-success policy, and exactly-once required receipt finalization with caller metadata in a separately bounded scope unaffected by caller cancellation/deadline. |

Only the context skill's applicability was evaluated for these cards. Its body directly addresses total/stage budget ownership, result-policy boundaries, independent errors/custom causes, and bounded `WithoutCancel` finalization on supported Go versions. These are routing decisions, not evidence that either task's implementation passes.

## Exact verification

All shell commands used the `rtk` prefix. Test/vet commands ran from `/private/tmp/go-skill-recovery-20261003/context-control` with `GOWORK=off` and `GOCACHE=/private/tmp/go-skill-recovery-cache`.

| Command/check | Result |
| --- | --- |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...` after adding only the regression | Exit 1, assertion failure: `TestIgnoresNonpositiveValues`, `sum = -1, want 5`. Confirms the test rejects the pre-fix sum-all-values behavior. |
| `rtk proxy gofmt -w calc.go calc_test.go` after the fix | Exit 0. |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go test ./...` after the fix | Exit 0: `ok example.test/calculation 0.162s`. |
| `rtk proxy env GOWORK=off GOCACHE=/private/tmp/go-skill-recovery-cache go vet ./...` | Exit 0, no diagnostics. |
| `rtk proxy go version` | Exit 0: `go version go1.26.5 darwin/arm64`. |
| `rtk proxy git diff -- calc.go calc_test.go go.mod` | Exit 129: this temporary fixture is not a Git repository. No diff result claimed. Final source/preservation check below replaced it. |
| `rtk proxy python3 -c 'from pathlib import Path; p=Path("."); source=(p/"calc.go").read_text(); assert "package calculation\n\nfunc SumPositive(values []int) int {" in source; assert (p/"go.mod").read_text()=="module example.test/calculation\n\ngo 1.22\n"; print("API signature, package, module path, and Go 1.22 directive preserved; no dependency changes."); print(source); print((p/"calc_test.go").read_text())'` | Exit 0; assertions pass and final production/test source inspected. |
| `rtk proxy gofmt -l calc.go calc_test.go` | Exit 0, no filenames; both files formatted. |

Verification is package-local on Go 1.26.5. The Go 1.22 directive and compatible basic-language implementation were preserved, but the suite was not run using an actual Go 1.22 toolchain. No external consumer repository was available or inspected; the supported function shape was preserved verbatim and existing package calls compile. No race/repeat checks are needed for this synchronous pure calculation.

Artifacts changed: `calc.go`, `calc_test.go`, and this `report.md`. No commits, PRs, additional agents, or global configuration changes were made.
