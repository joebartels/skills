# Independent Go review

Reviewed the supplied anonymized diff from `/private/tmp/go-independent-review-4q0updp4/original` to `/private/tmp/go-independent-review-4q0updp4/candidate`. These directories, their README contracts, and the dispatched review skills were the review inputs. There was no author report or prior review. Source digests and the exact reviewed diff are preserved in `checks.json`; original and candidate sources were not edited.

The request is to extend the CLI from one record to streamed ordered batches and add high-quality success and failure tests. The original README already requires accepting and retaining the prefix, stopping at the first failure, normalizing quantities, replacing duplicate IDs, preserving Go 1.22 and standard-library-only dependencies, and maintaining the documented process contract. The candidate README adds complete-record publication and preserving an earlier accepted duplicate on a failed replacement.

References below use paths relative to `candidate/`, unless marked `original/`.

## Testing — A+
Scope: Supplied original/candidate diff for the `example.com/ledgerload` CLI. The module still declares Go 1.22. Checks ran with Go 1.26.5 on darwin/arm64, in disposable copies.
Coverage: Assessed functional assertions, invalid-row and malformed-CSV boundaries, input and filesystem failures, accepted-prefix preservation, duplicate replacement, streaming order, real process exit/stdout/stderr behavior, child-process lifetime, filesystem cleanup, and test isolation. No production concurrency, fuzz target, benchmark, or external service is introduced. Linux/amd64 tests were compiled but not executed.
Rationale: No actionable testing issue was found. Two independent safeguards beyond routine setup are verified: (1) the staged input reader observes the already-published first record before supplying the next record and rejects a batch-prevalidation mutation; (2) a real CLI process under a child-only file-size limit exposes a partial replacement write and rejects a direct-truncating-write mutation. These protect distinct risks: batch prevalidation losing the accepted prefix, and a failed duplicate corrupting an already accepted value. Additional process assertions catch an incorrect exit-code mutation. Grades are based on behavior and assertion sensitivity, not test count or coverage percentage.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `main_test.go:125` and `main_test.go:207` control the input chunk boundary while retaining the real CSV parser and filesystem. The callback requires `first` to contain `3\n` before supplying `second`. The baseline passes; replacing the loop with `ReadAll` fails at `main_test.go:141` with "first record not published before next read". This proves the test distinguishes streaming publication from collecting the batch first.
- [T-G2] `write_failure_unix_test.go:22` runs a built CLI through a helper that applies `RLIMIT_FSIZE=2` only to its child and then `Exec`s the CLI (`:44`). Input first accepts `keep,1`, then attempts `keep,123`, then supplies a later row. Assertions require exit 2, no stdout, a relevant I/O diagnostic, exactly one record, and the intact `1\n` value (`:40`). The baseline passes; replacing `writeRecord` with direct `os.WriteFile` fails because the file contains `12`. This exercises an actual partial filesystem write rather than an error-return-only double.
- [T-G3] `main_test.go:149` builds and invokes the executable. It asserts exit status, empty stdout, useful failure stderr, successful empty stderr, exact file contents, and absence of later/unexpected entries. Changing `main` to exit 1 causes the invalid-duplicate process case to fail with the observed exit mismatch.
- [T-G4] `main_test.go:46`, `:87`, `:105`, and `:114` exercise validation, CSV, destination, directory-creation, and input-read failures after accepted work; assertions inspect the final files and prevent later writes or leaked temporary entries. Helpers use test-owned temp directories, and builds/processes have bounded contexts (`:274`, `:287`); the file-size limit never modifies the parent process. The full suite and three shuffled repetitions passed.

Bad

- None found.

Suggested changes

- None needed.

Limits: The installed toolchain is Go 1.26.5; actual Go 1.22 execution was not performed. Module language version and used APIs were inspected, and no newer API requirement was found. The partial-write test is explicitly limited to Darwin/Linux; its execution was verified on Darwin, with Linux compilation only. No platform outside these was inferred to be supported by an unspecified release matrix. No network or external service was needed; all dependencies are from the standard library. Race testing was not run because the changed production code is synchronous and the tests introduce no shared concurrent workflow. Finite mutations verify the particular assertions described, not exhaustive fault coverage. Exact commands, stdout, stderr, and exit codes are in `checks.json`.

## Correctness & Compatibility — A+
Scope: Same supplied CLI diff; introduced/worsened behavior only. Existing parsing and process behavior were compared with `original/main.go`; unchanged limitations are distinguished below.
Coverage: Traced empty and ordinary input, exact field counts, ID and quantity validation, normalization, duplicate order, first-row and midstream failures, malformed CSV, read errors, directory and rename errors, partial replacement writes, cleanup, and CLI usage/output/exit behavior. Reviewed unchanged `go.mod`, the original single-record caller test, and candidate tests. No public library API or concurrency is introduced.
Rationale: No actionable introduced correctness issue was found. Two distinct, verified safeguards justify A+: (1) the synchronous read/validate/publish loop retains the accepted prefix and stops before later writes on an input or file error; (2) writing and closing a same-directory temporary file before rename prevents a failed duplicate write from truncating a previous accepted value. Baseline success/failure assertions establish both effects, and targeted mutation failures independently confirm the corresponding regression signals. The grade applies to the supplied contracts and inspected runtime scope.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `main.go:24` fixes the CSV field count to two before indexing fields. `:27` reads one record at a time, rejects CSV/validation errors, and synchronously completes `writeRecord` before the next iteration. Tests verify exact normalized values, ordered duplicate replacement, retention of earlier accepted writes, and absence of later records; the observable-read safeguard confirms publication before further input is requested.
- [C-G2] `main.go:56` creates its temporary file in the destination directory, checks write and close errors before rename, and removes temporary state on return. The real partial-write CLI check verifies that a failed replacement retains the earlier accepted duplicate and leaves no temporary or later file. The rename-failure test verifies the accepted prefix and destination directory survive unchanged.
- [C-G3] `main.go:76` retains the original process boundary: failures print useful stderr and exit 2; successful operation emits no stdout. Real process cases verify empty input, single-record compatibility, batch success, invalid input, malformed CSV, usage errors, and I/O failures. `go.mod` still declares Go 1.22 with no third-party dependencies.

Bad

- None found.

Suggested changes

- None needed.

Limits: Actual Go 1.22 execution and Linux execution were not available in these checks; Darwin/arm64 runtime and Linux/amd64 compile verification are established. Broader OS/filesystem portability, crash durability, concurrent writers, symlink/permission preservation, and distributed atomicity are not specified contracts and were not assumed. Machine-width `strconv.Atoi` quantity limits and encoding/csv handling of blank lines are unchanged from the original implementation, not introduced findings. No public error identity contract is documented; tests appropriately verify useful diagnostics, while implementation wrapping also preserves underlying errors. The complete-publication README clarification was assessed as candidate contract, not as evidence that the original single-write implementation already guaranteed it. Exact checks are in `checks.json`.

## Architecture & Design — A
Scope: Same supplied diff; the new production `writeRecord` publication seam and its resource lifetime make this topic applicable. This remains one CLI package with internal helpers, rather than a reusable library or service.
Coverage: Assessed the read/validation/publication boundary, dependency composition, error propagation, ownership of temporary files and handles, synchronous completion, and fit of the abstraction to the requested CLI. No new exported API, package layering, background work, cancellation protocol, or external service is introduced.
Rationale: No actionable architecture issue was found. The small internal helper owns the complete temporary-file lifetime and publication operation; `run` retains stream orchestration and contextualizes failures, and `main` owns process exit. Real failure tests verify the resource-owning helper completes or cleans up synchronously. These are appropriate boundaries for this CLI and do not require an additional interface or package. A is selected for a sound, verified ownership design without claiming two independent architecture-specific safeguards beyond the behavior protections graded above.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `main.go:56` keeps acquisition, writes, close, rename, and temporary-file cleanup together. Callers receive the final publication outcome rather than a handle or incomplete operation. The partial-write and rename-failure tests pass while asserting exact destination entries, verifying ownership cleanup at realistic filesystem boundaries.
- [A-G2] `main.go:14` uses the existing standard `io.Reader`/`io.Writer` boundary for input and diagnostics and supplies record context when returning errors (`:33`, `:49`). `main.go:76` owns process exit. The observed reader tests streaming behavior without adding test-specific production hooks; subprocess tests exercise the public process boundary.

Bad

- None found.

Suggested changes

- None needed.

Limits: Assessment is limited to the supplied single-package CLI and documented present requirements. No speculative extensibility, externally managed file lifetime, or durability requirement was imposed. Runtime and platform limits match the correctness report. The straightforward helper did not require additional architecture reference material.

## Supporting verification facts

All commands were launched with `rtk`; Go commands used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Tests ran in `/private/tmp/go-independent-review-4q0updp4/verification`; mutations ran in separate sibling copies. The candidate/verification integrity diff returned 0 with empty stdout and stderr after checks. The reviewed original/candidate diff returned 1 because changes exist, not because a check failed.

| Check | Result | Meaning |
| --- | --- | --- |
| `go version` / `go env GOOS GOARCH GOVERSION GONOSUMDB GOPROXY` | Exit 0; Go 1.26.5, darwin/arm64 | Recorded actual toolchain/platform. |
| `go test -v -timeout=60s ./...` | Exit 0 | All supplied success/failure/process tests passed, including the real partial-write case. |
| `go test -count=3 -shuffle=on -timeout=60s ./...` | Exit 0 | Three shuffled suite runs passed; no order-dependent failure observed. |
| `go test -v -count=1 -run '^TestRunPublishesBeforeReadingNextRecord$' -timeout=60s .` in `mutation-prevalidate` | Exit 1 | `ReadAll` before publishing fails the first-record observation. |
| `go test -v -count=1 -run '^TestCLIPartialWriteRetainsAcceptedDuplicate$' -timeout=60s .` in `mutation-direct-write` | Exit 1 | Direct destination writes leave `keep` as `12`; expected `1\n`. |
| `go test -v -count=1 -run '^TestCLIProcessContract/invalid_duplicate$' -timeout=60s .` in `mutation-exit-code` | Exit 1 | Exit 1 is rejected; expected exit 2. |
| `GOOS=linux GOARCH=amd64 go test -c -o output/ledgerload-linux-amd64.test .` | Exit 0 | Linux/amd64 test binary compiled; not run. |

Expected failures above are sensitivity checks in disposable mutations, not defects in the candidate. Full exact command arrays, working directories, stdout, stderr, and exit codes are retained in `/private/tmp/go-independent-review-4q0updp4/output/checks.json`.
