## Testing — C
Scope: Complete existing code area at repository HEAD 28c092c44f4a0709c4132dac3b3947dabfae5a84 plus untracked fixture SHA256 identities in ../../manifest.json (all six hashes verified). Library, go.mod declares Go 1.21; execution used Go 1.26.5 darwin/arm64. Paths below are relative to golang/go-quality-report/evals/files/catalog-export. Full supplied test suite and both implementations.
Coverage: All supplied test assertions, ownership/nonmutation risk, output assertions, isolation and resource/concurrency applicability. No fakes, external boundaries, timing, benchmarks or shared test resources exist.
Rationale: One major test gap leaves the explicitly important independent-mutation contract unverified. It requires an independent test correction even after production is repaired. Both test files miss the same ownership invariant; counted once, not as systemic or two issues.
Finding counts: critical=0, major=1, moderate=0, minor=0

Good

- [G1] export/export_test.go:10-12 asserts transformed content and fails on wrong output; catalog/catalog_test.go:9-13 asserts append result/catalog length. Each uses local state and synchronous fatal assertions, with no timing or cleanup risk.

Bad

- [F1][major][existing-in-scope] catalog/catalog_test.go:5-15 tests only append/length, which cannot detect writes to existing aliased elements; export/export_test.go:8-13 checks only output, omitting catalog preservation. All supplied tests pass while documented ownership is broken. Baseline and original.txt demonstrate this directly. Primary remediation owner testing. Separate from producer defect because adding assertions is independently necessary for regression detection.

Suggested changes

- [F1] Add a Snapshot element-replacement isolation assertion and an Uppercase before/after catalog-content assertion (including multiple labels). Keep existing output and append assertions. Added disposable tests demonstrate fail-before/pass-after signal; preserve these regressions in the real suite when fixing production.

Limits: Executed in /tmp/go-umbrella-eval/live/export-check/module, GOCACHE=/tmp/go-umbrella-eval/live/go-cache, GOWORK=off: `rtk proxy go version`; `rtk proxy go test ./...` baseline passes; `rtk proxy go vet ./...` passes; `rtk proxy gofmt -d catalog/catalog.go catalog/catalog_test.go export/export.go export/export_test.go` emits no diff. Added disposable contract tests fail on original implementation; producer-only replacement of Snapshot return with append([]string(nil), c.labels...) makes `rtk proxy go test ./...` pass. Exact commands/results: ../../export-check/evidence/{0-baseline,1-baseline,2-baseline,3-baseline,original,producer-copy}.txt. Added tests: ../../export-check/module/{catalog,export}/review_contract_test.go. No reviewed-source edits. Go 1.21 toolchain itself, race tests (sequential contract), benchmarks and deployment checks not run.

Ungraded related finding: export/correctness-and-compatibility/F1 is the production ownership violation; not counted in this testing card.

Invocation provenance: Read and applied golang/go-testing/SKILL.md and references/testing-decisions.md; also reporting.md and grading.md from go-quality-report. No eval expectations or other worker cards read.

### Reconciliation addendum — major retained

The go-testing SKILL.md “Grade testing” rubric defines major as “leaves an important contract effectively unverified” and further says “Use major when the supplied contract or traced consequences establish that importance.” REVIEW.md explicitly identifies independent mutation of Snapshot results and preservation of stored labels by Uppercase as public requirements of this library. Neither supplied test asserts preservation of existing stored elements, so the stated ownership contract has no regression signal. Baseline passing and added ownership assertions failing establish this verification consequence independently of the production finding's severity. Producer repair alone does not add regression detection: the test correction remains independently necessary. This is one contained important ownership contract, not a systemic major. No new checks were run during reconciliation.
