# Independent changeset review

Reviewed 2026-10-01. The supplied boundary is `original/` → `candidate/` under `/private/tmp/go-independent-review-0hqvmxet`; these directories have no supplied commit identifiers. I inspected the README, every changed production file, all supplied tests, the Go module declaration, and the three dispatched review skills and their local decision references. No author report, other candidate, repository design record, or evaluation expectations were inspected. The original and candidate source were not changed.

The implementation is a standard-library file-backed library plus a CLI, declaring Go 1.22. The executed platform was Go 1.26.5, darwin/arm64, CGO enabled. All Go checks used `GOCACHE=/private/tmp/go-quality-testing-cache` and `GOTOOLCHAIN=local`. Verification and mutations used disposable copies. No dependencies or toolchains were installed.

## Testing — A+

Scope: The supplied changeset and the README's requested new regression scope, including preservation of existing consumers and actual command process contracts.

Coverage: Assessed assertions for successful replacement, ordered CSV application, rejected replacements and accepted prefixes, parser/read/write/publication failures, response-body closure, exact JSON representation, redirect rejection, real HTTP cancellation, POSIX snapshots, callback sequencing and error identity, instance independence, command status/streams/stdin, recurrence, and both interrupt and termination during an in-flight request. Assessed temp-directory ownership, helper-child limits, cleanup/join paths, and parallel test state. No benchmarks or fuzz targets are supplied; neither is necessary for a substantiated gap in this inspected scope.

Rationale: No actionable introduced or worsened testing issue was found. Two independent safeguards earn A+: (1) isolated child-process RLIMIT fault injection and an already-open reader verify preservation of actual filesystem data, and reject a direct-write mutation; (2) channel-gated callback cleanup verifies that release/return cannot overtake started work, and rejects an early-release mutation. These control distinct data-loss and lifecycle-ordering risks; the grade is based on assertion signal, not test count or coverage percentage.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [T-G1] `candidate/write_failure_unix_test.go:22-90` induces a real partial write in an owned, bounded child for Put, CSV and Refresh. It checks exact prior bytes, the accepted CSV prefix, absent later rows, and removed staging files. `candidate/refresh_test.go:164-187` also reads the old file descriptor after replacement. Replacing the publication helper with direct `os.WriteFile` failed all three retained-data cases and the old-reader assertion.
- [T-G2] `candidate/serve_test.go:145-195` holds cooperative cleanup behind an explicit gate and checks release/return before unblocking, callback error identity, and a single release. A mutation that returned on cancellation without joining the callback failed both subcases promptly. `launchServe` owns cancellation/unblocking and a bounded join (`serve_test.go:238-267`).
- [T-G3] `candidate/serve_test.go:77-142` holds the callback past the interval, then measures recurrence from completion. Starting the timer before the callback produced a roughly 1.459 µs delay and failed the required 40 ms assertion. This is testing elapsed-time behavior, not guessing when unrelated work completes.
- [T-G4] CSV tests assert retained values and absent later rows (`candidate/apply_test.go:75-135`); skipping CSV validation failed empty-text and Unicode-whitespace cases. The real binary assertions check status and both streams (`candidate/cmd/indexer/process_test.go:237-261`); changing exit 2 to exit 1 failed the usage subcases. Watch tests prove two published cycles and cancellation of an in-flight third request for both signals (`process_test.go:112-209`).
- [T-G5] The unchanged-source verification copy passed the full suite and three shuffled race runs, including the loopback HTTP and process tests. Parallel cases own their roots and state; Go 1.22's declared loop-variable semantics apply.

Bad

- None found.

Suggested changes

- None needed.

Limits: The initial sandboxed full/race runs failed at `httptest.NewServer` because localhost binding was denied; those environmental failures are not candidate defects. Both checks passed with sandbox escalation. Go 1.22 itself was not executed; the module remains at 1.22 and `go vet -stdversion ./...` passed. The RLIMIT test executes only on darwin/linux; Windows replacement execution is explicitly outside the supplied contract. These checks do not prove every possible scheduler interleaving or input.

## Correctness & Compatibility — A+

Scope: Introduced or worsened behavior in the supplied library/CLI changeset against the README contracts, with the original exported signatures and legacy Put behavior as the compatibility baseline.

Coverage: Traced ordinary, empty, malformed and rejected inputs; ordered partial completion; exact text/order and empty-array encoding; failures before publication; existing and later filesystem readers; errors.Is identity; cancellation and release ordering; consumer function-value compatibility; and actual CLI success/failure/signal behavior. Static Go 1.22 API/language compatibility was checked; the runtime execution claim is darwin/arm64.

Rationale: No actionable introduced or worsened correctness issue was found. Two independent verified safeguards earn A+: (1) publication stages in the destination directory and renames only after successful write/close, preserving prior data across real partial-write failures and old-reader snapshots; (2) Serve executes callbacks synchronously and joins the callback's cooperative cleanup before its sole deferred release, while joining callback/release error identities. These protect state and lifecycle contracts independently.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [C-G1] `candidate/index.go:27-34` still permits arbitrary Put text. New import/refresh validation is separate (`index.go:71-75`), so the tighter nonblank rule does not tighten existing Put behavior. Empty, whitespace, NUL-containing text and positional text beginning with '-' pass; external function-type assignments in `candidate/apply_test.go:17-25` compile unchanged.
- [C-G2] `candidate/index.go:48-68` validates and writes one row before reading the next, preserving accepted prefixes and duplicate-key order. Parse, validation and write failures stop later effects; `errors.Join` retains ErrInvalidRecord for invalid field counts. The supplied tests verify exact retained values and error identities.
- [C-G3] `candidate/refresh.go:21-63` makes a GET with the caller context, rejects redirects/non-200 responses, requires a non-null array with EOF after whitespace, validates every record before publication, preserves order/text, and closes acquired bodies on success/rejection. Empty arrays publish `[]\n`. Read failures after the array and later invalid records retain the exact prior bytes (`candidate/refresh_test.go:61-110`).
- [C-G4] `candidate/index.go:80-93` stages beside the destination and publishes by rename. The real RLIMIT and POSIX snapshot tests passed, while the direct-write mutation failed them. This demonstrates complete prior/new snapshots on the evaluated filesystem without inventing fsync durability or CSV all-or-nothing semantics.
- [C-G5] `candidate/serve.go:16-33` rejects invalid intervals before lifecycle acquisition, skips refresh for an already-canceled context, measures the interval after callback completion, and preserves both callback and release causes. Gated cleanup, independent invocations, recurrence, and three shuffled race runs passed.
- [C-G6] `candidate/cmd/indexer/main.go:27-43,64-83` preserves Put's positional arguments, imports stdin through ApplyCSV, scopes signals to main, and treats ordinary watch cancellation as successful only when the refresh error matches the canceled context. Actual processes preserved status 0/2 and stream behavior, accepted-prefix effects, repeated publication, and graceful SIGTERM/interrupt shutdown.

Bad

- None found.

Suggested changes

- None needed.

Limits: Passing tests establish the exercised behavior, not arbitrary inputs or all interleavings. No Go 1.22 binary was executed; no newer API was identified and the version-aware vet check passed. Windows replacement and other operating-system runtime behavior were not claimed. Nil callbacks/client and power-loss durability are not supplied supported contracts and were not invented as requirements. The original unimplemented operations were intentionally implemented and are not regressions; ErrNotImplemented remains exported.

## Architecture & Design — A+

Scope: The production structure added for CSV import, HTTP refresh, callback lifecycle management and command composition in this small Go 1.22 library/CLI.

Coverage: Assessed public signatures, package dependencies, protocol/persistence boundaries, shared validation/publication responsibilities, visible dependencies, ownership, cancellation/error propagation, and proportionality. No additional service framework, interface hierarchy or package split is warranted by the supplied use case.

Rationale: No actionable introduced or worsened architecture issue was found. Two independent verified safeguards earn A+: (1) a private shared staged-publication boundary enforces one complete-snapshot invariant across three workflows without exposing infrastructure abstractions to consumers; (2) borrowed and owned HTTP/lifecycle resources have explicit scopes, so Refresh preserves caller client configuration/lifetime and the command releases its owned connections after synchronous callback completion. The design controls data-boundary and resource-ownership risks separately while remaining small.

Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [A-G1] `candidate/index.go:71-93` uses a small validation helper and one staged-publication helper. Put does not inherit import-only validation, and Refresh does not inherit directory creation or CSV prefix semantics. Real partial-write and old-reader tests verified the shared publication invariant; their direct-write mutation failures demonstrate that this boundary matters.
- [A-G2] `candidate/refresh.go:27-33` copies the borrowed client's configuration to apply a one-request redirect rule without mutating the caller's policy or closing its transport. The supplied redirect test verifies the original policy remains intact (`candidate/refresh_test.go:113-137`). An additional reviewer-only probe verified repeated use of the supplied cookie jar/transport, two body closes, zero idle-connection closes, and retention of the supplied client timeout.
- [A-G3] `candidate/cmd/indexer/main.go:64-75` creates a private transport/client and supplies release to Serve; `candidate/serve.go:19-33` runs work synchronously before release. The gated-cleanup and real signal-process checks verify shutdown ordering. Process signals and os.Exit stay in the command entry point (`main.go:78-83`), rather than the reusable library.
- [A-G4] The existing concrete Index, io.Reader import boundary, injected *http.Client, and callback functions remain sufficient. Record is the actual JSON protocol representation; there is no demonstrated divergent domain model requiring duplicate DTOs. No added module dependency, interface, background scheduler or package layer obscures these contracts.

Bad

- None found.

Suggested changes

- None needed.

Limits: Architecture is assessed for the supplied small program and its stated evolution. No unsupported future-scale layering or durability requirement was assumed. The command-owned idle-close ordering was traced through synchronous Serve and its verified release ordering; it was not separately instrumented inside the child binary. The borrowed-client probe is reviewer evidence in a disposable copy, not an added candidate test.

## Verification facts

`output/checks.json` preserves exact underlying verification command arguments, stdout, stderr, exit status and finite outer deadline; the inspection transcript is included with the tool's combined-output limitation stated. Commands below ran through `rtk proxy` with the required Go environment.

| Check | Result |
| --- | --- |
| `go test -count=1 -timeout=45s ./...` | Exit 0 after localhost escalation; both packages passed. |
| `go test -race -shuffle=917 -count=3 -timeout=45s ./...` | Exit 0 after localhost escalation; both packages passed. |
| `go vet -stdversion ./...` | Exit 0; no stdout/stderr. |
| Direct-write publication mutation, snapshot and partial-write tests | Exit 1; old-reader snapshot and all three prior-data assertions failed. |
| Timer started before callback, completion-interval test | Exit 1; delay after completion was below 40 ms. |
| Cancellation return before callback join, gated-cleanup test | Exit 1; both subcases detected early release/return. |
| Skipped CSV validation, rejected-replacement test | Exit 1; blank-text cases detected acceptance. |
| Wrong os.Exit status, actual usage processes | Exit 1; expected status 2 was observed as 1. |
| Reviewer borrowed-client/configuration probes | Exit 0; repeated use, jar/transport ownership and client timeout verified. |

The mutation failures are intentional evidence of regression detection, not candidate failures. No confirmed bad findings are counted across the three report cards.

