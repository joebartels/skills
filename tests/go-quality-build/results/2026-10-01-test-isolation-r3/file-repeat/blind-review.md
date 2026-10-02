# Independent review — changeset 2

Reviewed the supplied original/candidate trees against the Store task in `candidate/README.md:3–15`. Only `store_test.go` changes; production, README, and module files are byte-identical. No candidate repairs were made. Disposable mutation copies are recorded in `mutations.json`; complete raw commands, working directories, explicit non-secret environment, stdout, stderr, and exits are in `checks.json`; source hashes are in `source-hashes.json`.

## Testing — A+
Scope: Supplied original → candidate diff for the `example.com/filestore` library. Requested concurrent independent roots, grouped children, repeated same-key values, fixture lifetime, and retained serial checks; module minimum Go 1.22. Executed on Go 1.26.5, darwin/arm64, CGO enabled.
Coverage: Inspected every supplied implementation/test/module/task file. Assessed actual filesystem I/O, same-key repetitions, distinct-root replacement/isolation, exact value bytes, serial operations within each root, parallel independent children, group completion, root ownership, filtered subtests, and version compatibility. Race/shuffle and disposable mutation checks exercised the requested scope. Same-key concurrent replacement inside one root is explicitly excluded by the task.
Rationale: No actionable issue found. Two independent safeguards beyond routine setup are verified: distinct per-instance final values with reads after the group detect root aliasing even when identical earlier values would hide it; the special-byte replacement fixture and exact result assertions detect loss of caller value bytes. These protect storage isolation and byte fidelity, respectively. The candidate also keeps test-owned roots alive through descendants and post-group checks. The grade follows zero counted defects and independent verified assertion signal, not case count or formatting.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] [store_test.go:32](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:32), [51](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:51), and [76](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:76) combine repeated identical initial values with distinct replacements and final post-group reads. A disposable production mutation that reused the first root for all instances failed the final isolation assertions (exit 1), even with `-parallel=1`; detection does not rely on a fortunate simultaneous read/write interleaving.
- [T-G2] The beta value at [store_test.go:33](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:33) includes both a newline and a trailing NUL, and [67](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:67) checks the whole string. A disposable mutation that stripped the trailing NUL from `Get` failed both the child assertion and the post-group assertion (exit 1).
- [T-G3] [store_test.go:38](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:38) assigns each root to the outer test's `TempDir` lifetime. The group at [42](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:42) waits for parallel descendants before outer verification, while [56](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:56) keeps same-root cases serial. Twenty full race/shuffle repetitions with `-parallel=4` passed; no teardown-before-use or shared expectation-state race was observed.
- [T-G4] [store_test.go:9](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:9) retains useful serial replacement checks. [61](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:61) and [78](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:78) track the actual selected last successful write and skip excluded instances. Three race repetitions selecting only beta's initial/repeat children passed without requiring an unselected replacement or reading unwritten roots.

Bad

- None found.

Suggested changes

- None needed.

Limits: `rtk proxy env GOCACHE=/private/tmp/go-quality-testing-cache GOTOOLCHAIN=local go test ./... -timeout=30s` passed. The full `go test ./... -race -count=20 -shuffle=on -parallel=4 -timeout=30s` run, focused beta initial/repeat race run, and focused retained serial race run passed. `go vet -stdversion ./...` passed. Full exact invocation details and mutation outcomes are preserved in `checks.json`; every Go command uses the specified GOCACHE and GOTOOLCHAIN overrides. Verification uses actual local filesystem files, not a persistence double. Actual Go 1.22 execution and other platforms/filesystems were not performed; the declared language directive and stdversion check support the source/API assessment. Race success covers exercised concurrency, not every possible schedule. Invalid/missing-key tests were absent in the original and remain outside the requested new concurrent/grouped test scope; this is a legacy coverage limit, not an introduced defect. Same-key concurrent replacement within one root is outside the task.

## Correctness & Compatibility — A
Scope: The same test-only changeset, assessing introduced test execution/lifecycle behavior and preservation of the Store API and supported Go version. The unchanged production implementation supplies contract context.
Coverage: Compared every original/candidate file; traced parallel descendants, same-instance serial child execution, parent/child expectation-state access, TempDir cleanup timing, loop capture under Go 1.22, and focused `-run` behavior. Assessed build/API compatibility with compilation, the unchanged module/public source, and stdversion vet.
Rationale: Zero actionable defects found. The public API/module is unchanged, and introduced test concurrency accesses distinct per-instance state until the grouped descendants finish. Parent observations happen after the group completes, and filtered runs use expectations for work actually selected. Race, shuffled full runs, and targeted selection checks passed. This A reflects sound changed-source execution and compatibility, not a new audit of unrelated legacy production behavior.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `store.go` and `go.mod` are byte-identical to the originals (SHA-256 records in `source-hashes.json`), preserving exported signatures and the Go 1.22 minimum. Compilation and stdversion vet passed.
- [C-G2] [store_test.go:44](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:44) captures a distinct instance, [46](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:46) permits independent instances to run concurrently, and [74](/private/tmp/go-independent-review-q4rd7fh4/changeset-2/candidate/store_test.go:74) finishes the group before parent reads. Twenty race/shuffle repetitions found no exercised state race. The focused beta initial/repeat run confirms the parent does not presume all children or the replacement case ran.

Bad

- None found.

Suggested changes

- None needed.

Limits: Same recorded platform, toolchain, filesystem, and exercised-concurrency limits as Testing. No actual Go 1.22 binary was run; no tools/dependencies were installed. Unchanged invalid-key behavior and excluded same-root concurrent replacement are not graded as changed production behavior.

## Architecture & Design — Not applicable
Scope: The supplied Store test-only diff.
Coverage: Confirmed unchanged production/API/module files and inspected the grouped fixture and per-instance root ownership for consequential production/test seams.
Rationale: No new production seam, abstraction, package/API boundary, or dependency design is introduced. The grouping and TempDir ownership are test lifecycle decisions assessed in Testing and Correctness & Compatibility. They do not require an architecture letter grade.
Limits: This does not grade the unchanged library's general architecture.
