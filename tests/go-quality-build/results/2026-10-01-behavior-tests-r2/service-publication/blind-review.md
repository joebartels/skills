# Independent review: changeset 2

Review boundary: supplied `original/` versus `candidate/` for the mirror library; the actual request is implementing README-defined Refresh and high-quality regression tests while retaining its public API and module minimum. The entire supplied module was inspected. No other candidate or external evaluation material informed this card. File references below are relative to this changeset.

## Testing — A+
Scope: Supplied implementation and new external consumer tests in `candidate/refresh_test.go` and `candidate/refresh_unix_test.go`; Go 1.22 minimum, executed using Go 1.26.5 on darwin/arm64.
Coverage: GET/client/context propagation; real loopback HTTP success; status, malformed JSON, trailing data, null/nonarray inputs, invalid/blank fields, nonblank value preservation and empty arrays; request, transport and read errors; body ownership; exact prior bytes; rename failure, stage cleanup, old inode handles and actual partial-write failure. Existing Write tests were also run. Unix-specific tests are explicitly limited to darwin/linux by their build constraint.
Rationale: No actionable issue was found. Two independent verified safeguards justify A+: rejection tests check the remote validation boundary and preserve exact prior bytes, while real filesystem tests independently check publication semantics and failure isolation. Removing trailing validation failed the first safeguard; replacing staging/rename with direct truncating writes failed both the open-handle and partial-write checks. These control acceptance of incomplete/invalid remote snapshots and destructive publication, respectively.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/refresh_test.go:93` checks rejection and exact previous bytes after non-200 responses, malformed/trailing JSON and invalid items, including an invalid item after a valid prefix. The omit-trailing-validation mutation failed the trailing-array, trailing-null and trailing-junk cases.
- [G2] `candidate/refresh_unix_test.go:22` asserts that a newly opened path contains the replacement while an already-open old handle retains its original bytes, across successive accepted replacements and a rejected refresh. A direct-write mutation failed the old-handle assertion.
- [G3] `candidate/refresh_unix_test.go:64` bounds a child process and imposes a real file-size limit only on that process. It verifies the partial-write error, exact old bytes and stage cleanup without a production mock seam. The direct-write mutation corrupted the old bytes and failed this assertion.
- [G4] `candidate/refresh_test.go:65` asserts exact output values and lowercase JSON fields; `candidate/refresh_test.go:138` verifies the supplied client, method, endpoint and context deadline/value. `candidate/refresh_test.go:236` additionally executes a real HTTP boundary.
- [G5] `candidate/refresh_test.go:176` exercises failed request construction, transport, cancellation and read failure including a read error after a valid JSON value, checking file retention and body-close counts. `candidate/refresh_test.go:261` checks rename failure without relying on Unix permission assumptions.

Bad

- None found.

Suggested changes

- None needed.

Limits: Initial full and race runs failed because the sandbox denied httptest's localhost listener (`bind: operation not permitted`), not because of a candidate assertion. Both failures are preserved. Automatic approval then permitted an unchanged-source listener rerun: `rtk proxy go test -count=1 -timeout=30s ./...` and `rtk proxy go test -race -count=1 -timeout=30s ./...` both passed (exit 0). Focused validation/publication mutations failed as expected (exit 1). Race detection covers executed paths only. Runtime checks were on darwin/arm64; Linux, Windows and the actual Go 1.22 toolchain were not executed. No Windows inode-equivalence claim is made. Exact command, cwd, non-secret environment, stdout, stderr, exit and escalation details are retained in `output/checks.json`.

## Correctness & Compatibility — A+
Scope: Supplied Refresh implementation replacing its prior reserved stub, the README remote-snapshot contract, and the unchanged existing Write API; Go 1.22 and standard-library-only module.
Coverage: Request construction, supplied client/context, 200-only acceptance, complete JSON-array consumption, field validation without normalization, output fields, empty arrays, failure-state preservation, response-body closure and atomic replacement on the exercised Unix filesystem. Existing public function type and module minimum were inspected; existing Write behavior remains outside the remote-refresh replacement contract.
Rationale: No actionable introduced or worsened issue was found. Two independent verified safeguards justify A+: the entire response and all items are validated before publication begins, preventing invalid/incomplete input from altering the snapshot; publication stages and closes a separate file in the destination directory before rename, preserving the existing inode through failed writes and preserving open-reader snapshots through successful replacements. Validation-removal and direct-write mutations demonstrate the distinct failures these boundaries prevent. The reach is the documented local snapshot workflow, with platform limits stated below.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/mirror.go:32` binds the request to the supplied context and `candidate/mirror.go:36` uses the supplied client; `candidate/mirror.go:40` owns response closure. Consumer tests verify forwarding and body closure, and the approved HTTP run verifies a real request.
- [G2] `candidate/mirror.go:45` through `candidate/mirror.go:64` accept only an array with successful end-of-input and valid nonblank fields before staging. Trimming is used only for validation; exact nonblank values are retained. Tests verify error paths, trailing-read failures, empty arrays and exact accepted output.
- [G3] `candidate/mirror.go:75` stages in the destination directory and `candidate/mirror.go:84`/`candidate/mirror.go:87` require successful write and close before `candidate/mirror.go:90` replaces the path. Executed Unix tests prove open old handles retain their bytes and a real partial-write failure preserves exact prior bytes.
- [G4] `candidate/mirror.go:79` removes staging files on failure; rename-failure and partial-write tests verify no stage remains. Refresh's existing signature, Item fields and Go 1.22 module directive are unchanged; `candidate/refresh_test.go:19` compiles the existing function type from a consumer package.

Bad

- None found.

Suggested changes

- None needed.

Limits: Full and race verification passed on an unchanged copy after approved loopback access; original sandbox failures remain recorded. Atomic replacement evidence is for the exercised darwin filesystem. Linux and Windows behavior, power-loss durability and the actual Go 1.22 toolchain were not executed; power-loss durability is not required by the supplied contract. No added nil-client or whole-operation cancellation contract was inferred beyond using the supplied HTTP client/context. Existing local Write still truncates its target as before; that legacy local-only API was neither changed nor newly made part of Refresh. Details are in `output/checks.json`.

## Architecture & Design — A
Scope: Refresh's new HTTP-to-snapshot boundary and private publication helper; consequential test boundaries for HTTP and partial filesystem writes.
Coverage: Public API continuity; runtime client/context ownership; parsing/validation before filesystem publication; private helper fit; response/stage lifetime; existing RoundTripper seam and real HTTP/OS tests. No package restructuring or new public storage abstraction is introduced.
Rationale: No actionable architecture issue was found. The existing client parameter exposes transport configuration to callers, while the small private replacement helper separates publication lifetime from remote-response validation. Tests can exercise consequential failures through existing standard boundaries without adding a public testing interface. These are appropriate, verified design choices for the supplied small library and support A.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `candidate/mirror.go:31` retains per-call context and client ownership at the public API. Consumer tests at `candidate/refresh_test.go:138` verify that the configured transport and context reach the request.
- [G2] `candidate/mirror.go:72` keeps temporary-file acquisition, write/close, cleanup and rename under one private synchronous owner. Success reports after publication, and errors leave cleanup to that same owner. The open-handle, partial-write and rename-failure tests exercise this boundary without exported implementation details.
- [G3] `candidate/refresh_test.go:21` uses the standard RoundTripper boundary for controlled remote inputs and `candidate/refresh_test.go:236` checks real HTTP. `candidate/refresh_unix_test.go:64` isolates process-wide resource limits in a child, avoiding production hooks or interference with the host test process.

Bad

- None found.

Suggested changes

- None needed.

Limits: Architecture was assessed only for the documented synchronous remote-snapshot workflow. No daemon, multi-writer ordering, durability or unsupported-platform policy was assumed. Behavior verification and environment details are in `output/checks.json`.
