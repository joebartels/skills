# Independent Go changeset review

Reviewed 2026-10-01. Boundary: the supplied directory diff from `/private/tmp/go-independent-review-kolmdafa/original` to `/private/tmp/go-independent-review-kolmdafa/candidate`, with the unchanged README as the behavior contract. This is a small snapshot library, not a service process. No author report, other candidate, evaluation expectation, controller probe, or unrelated repository file was inspected. No delegation or source repairs were performed.

Only `mirror.go` changes in production; `refresh_test.go` and `refresh_unix_test.go` are added. `go.mod`, `README.md`, and the local-write test are unchanged. References below are relative to `/private/tmp/go-independent-review-kolmdafa`.

## Testing — A+

Scope: Supplied original/candidate diff and the requested regression-test scope for `Refresh`; Go 1.22 module, standard library only, external-package consumer tests. Native execution used Go 1.26.5 on darwin/arm64.

Coverage: Assessed successful GETs through the supplied client, endpoint/query preservation, lowercase output fields, nonblank value preservation, empty arrays, trailing whitespace, status/transport/read/parse/item failures, exact old-file preservation, response closure, cancellation through real HTTP, publication failure/temporary cleanup, repeated accepted/rejected transitions, open-reader replacement, and genuine partial storage writes. Inspected async waits and subprocess/environment/signal/resource ownership. Executed baseline, repeated shuffled race checks, and five single-defect mutations. No fuzzing or benchmarks were supplied or needed to resolve this bounded contract.

Rationale: No substantiated testing issue. Two independent safeguards justify A+: (1) the real filesystem tests distinguish replacement from in-place mutation and force partial write progress before failure; the in-place mutant fails both old-handle and exact-preservation assertions; (2) response-validation tests detect incomplete or inadmissible remote snapshots even when a complete JSON value precedes a late read failure; blank-validation, late-read-error, and trailing-JSON mutants fail their intended assertions. These control storage-publication integrity and remote-input completeness respectively, rather than earning credit from test count or coverage percentage.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/refresh_unix_test.go:22` checks the already-open old handle and a fresh open; `candidate/refresh_unix_test.go:48` confines RLIMIT_FSIZE and SIGXFSZ changes to a bounded child process. The child proves 64 bytes were written before an error (`:59–71`); the parent then checks exact old bytes and absence of leftover files (`:99–100`). The baseline passes, and the in-place mutant fails at `:42` and `:99`. This verifies the requested inode/publication contract and protects against corruption after partial progress.
- [T-G2] `candidate/refresh_test.go:114` checks rejected status/payload/item cases against deliberately noncanonical previous bytes. `candidate/refresh_test.go:162` supplies complete valid JSON followed by an actual reader error, preventing a decode-only test from overlooking incomplete I/O. The corresponding blank-field, late-read-error, and trailing-JSON mutants fail with the expected acceptance diagnostics.
- [T-G3] `candidate/refresh_test.go:68` exercises a real local HTTP server and a transport header proving use of the supplied client, with exact method/path/query and serialized output assertions. `:214` waits for the server to observe the request before canceling, then bounds both Refresh completion and server cancellation. Three shuffled race-enabled runs pass.
- [T-G4] `candidate/refresh_test.go:149`, `:176`, `:191`, and `:293` inspect response-body closure on rejection, read failure, success, and publication failure. Omitting the production defer makes each corresponding assertion fail. TempDir ownership, bounded client timeouts, buffered worker result delivery, deferred cancellation/server closure, and bounded subprocess execution keep the test dependencies controlled.

Bad

- None found.

Suggested changes

- None needed.

Limits: Full tests initially could not bind the sandbox's loopback port; rerunning with authorized local-listener access passed. Native filesystem and partial-write behavior were checked only on macOS. Linux and Windows tests compile but were not executed. The 1.22 toolchain itself was not run; the unchanged module directive and a passing `go vet -stdversion ./...` support the inspected minimum-API claim. A passing race run covers exercised paths only. Exact verification commands, separate stdout/stderr, exits, and timeouts are preserved in `output/checks.json`; inspection output is in `output/inspection-checks.json`.

## Correctness & Compatibility — A+

Scope: Supplied original/candidate diff for `Refresh`, evaluated against `candidate/README.md:3–16`. The old implementation was an unimplemented stub; defects of the unchanged local `Write` are not attributed to this change.

Coverage: Traced ordinary/empty/boundary input, strict complete-stream decoding, nonblank validation without mutation of field values, publication ordering, failures before/after temporary-file creation, old/new handle visibility, HTTP body ownership, request cancellation, exported signatures, and module/platform constraints. Tested the complete supplied suite, repeated race runs, and build/vet checks. Windows replacement semantics were not claimed or assessed at runtime.

Rationale: No substantiated introduced or worsened correctness issue. Two independent verified safeguards justify A+: complete response reading plus strict whole-buffer decoding/item validation occurs before filesystem mutation, controlling publication of invalid/incomplete remote data; same-directory staging plus checked write/close and rename controls partial-write damage and preserves the old inode. The rejection/read-failure tests and the real partial-write/open-handle tests verify those distinct risks.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/mirror.go:32–59` constructs the GET with the caller's context, uses the supplied client, requires exactly 200, closes the response body, checks the full read result, rejects trailing JSON through `json.Unmarshal`, rejects top-level null, and validates every item's trimmed code/label. Storage starts only at `:66`. Exact-byte retention tests pass for rejection and late read error; real HTTP cancellation passes.
- [C-G2] `candidate/mirror.go:55–60` uses trimming only for validation and marshals the original decoded values with the existing lowercase struct tags (`:15–17`). `candidate/refresh_test.go:70–101` verifies preserved surrounding spaces, correct output fields, empty-array output, and accepted trailing whitespace.
- [C-G3] `candidate/mirror.go:64–84` creates the staging file in the destination directory, checks write and close before rename, returns errors without publishing, and defers staging-file removal. The native old-handle, genuine partial-write, and obstructed-publication tests pass. No crash-durability requirement is inferred from the README.
- [C-G4] The exported Refresh signature and Item/Write definitions remain unchanged; `candidate/refresh_test.go:18–19` compiles a function-value assignment against the existing signature. The module remains Go 1.22 with no external dependencies; minimum-API vet and Linux/Windows compilation pass. The production comment explicitly confines old-handle guarantees to supported Unix filesystems.

Bad

- None found.

Suggested changes

- None needed.

Limits: Runtime checks used Go 1.26.5/darwin/arm64, not a Go 1.22 compiler or Linux/Windows execution. Cross-compilation establishes buildability only. Local loopback HTTP was allowed for the native suite; no external endpoint or dependency download was required. The README does not promise crash durability, writer ordering, file-mode preservation, or equivalent Windows inode behavior, so these were not invented as requirements. Nil clients or a transport that violates net/http's contract were not treated as supported callers. See exact evidence in `output/checks.json`.

## Architecture & Design — A

Scope: The supplied Refresh implementation. Architecture applies because replacing the stub introduces consequential HTTP-response and temporary-file ownership, a storage publication boundary, and use of the caller's client/context seam. Public signatures, package structure, and module dependencies do not change.

Coverage: Assessed dependency visibility/replacement, synchronous success meaning, cancellation ownership, resource acquisition/cleanup, publication boundary, and fit of the existing concrete API. Read the supplied architecture skill and its local design-decisions reference. No layering, storage-interface, or background-worker requirement was assumed.

Rationale: No substantiated architecture issue; at least one relevant strength is verified. The existing small API provides the dependencies needed by its consumers and tests, and the implementation owns acquired resources through a synchronous operation. These are sound, appropriately simple ownership/composition decisions. A is supported without treating routine cleanup or added code as independent A+ design safeguards.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate/mirror.go:31–40` keeps HTTP configuration and request lifetime with the caller. The real HTTP/client-identity and cancellation tests (`candidate/refresh_test.go:68`, `:214`) verify that the provided dependencies govern the operation. No hidden global client or new testing-only production abstraction is introduced.
- [A-G2] `candidate/mirror.go:40`, `:66–84` owns the response/staging file locally and returns success only after publication. Body-closure, cleanup, partial-write, and old-reader tests verify the corresponding lifetime/publication choices. Process signal/resource-limit changes live solely in the disposable child test, not in the library.

Bad

- None found.

Suggested changes

- None needed.

Limits: No downstream application composition was supplied beyond the library's documented API and tests. No cross-platform runtime ownership guarantee is inferred from compilation. Native checks and resource-control evidence have the limits recorded above.

## Supporting verification facts

All Go commands set `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`; every shell invocation used `rtk`, with `rtk proxy` for raw checks. Checks ran in a disposable unmodified copy or individually mutated copies. Original/candidate source hashes remain identical across all 10 files.

| Check | Verified result |
| --- | --- |
| Full baseline, `go test -count=1 -timeout=25s ./...` | Exit 0, `ok example.com/mirror 0.199s`, after enabling local listeners. |
| `go test -race -count=3 -shuffle=on -timeout=30s ./...` | Exit 0, `ok example.com/mirror 4.379s`. |
| `go vet -stdversion ./...` | Exit 0; empty stdout/stderr. |
| Windows/amd64 and Linux/amd64 `go test -c` | Both exit 0; empty stdout/stderr. No runtime/platform-semantic claim. |
| In-place-publication mutant | Exit 1; old-handle and partial-write-retention assertions fail. |
| Blank-field acceptance mutant | Exit 1; code, label, later-invalid-item, and null-item acceptance assertions fail. |
| Ignored late-read-error mutant | Exit 1; `Refresh accepted a response whose read failed after complete JSON`. |
| Missing response-close mutant | Exit 1; success/rejection/read/publication closure assertions report zero closes. |
| Trailing-JSON acceptance mutant | Exit 1; trailing JSON/garbage and rejected-transition assertions fail. |
| Original/candidate source preservation | Exact SHA-256 hashes unchanged for all 10 supplied files. |

Mutation failures are expected evidence of regression detection, not candidate findings. Exact mutation edits are in `output/mutations.json`; the five source copies are under `mutations/`. The initial sandbox bind failure is retained in checks.json and is an environment limitation, not a production/test defect.

Ungraded legacy context: `Write` still uses direct `os.WriteFile` as before. The new replacement contract belongs to Refresh, so that unchanged local-only behavior does not reduce these changeset grades.

