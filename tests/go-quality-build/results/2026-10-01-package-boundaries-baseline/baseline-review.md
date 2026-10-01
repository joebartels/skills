# Independent blind baseline review

Reviewed 2026-10-01. The review was established from baseline-review-input.md, original/changed source diffs, all Go source/tests and go.mod files in the five cases, and the repository Architecture & Design and Correctness & Compatibility skills and their decision references. Installed skill texts were compared byte-for-byte with repository versions and matched. No evals.json, source-audit.md, proposed build skills, or trial-report.md contents were read.

All references below are relative to `/private/tmp/go-quality-build-eval.TijGDa/baseline/`; originals are the corresponding case directories under `/Users/jb/.codex/worktrees/go-quality-build/skills/tests/go-quality-build/go-package-boundaries/evals/files/`. These are supplied directory diffs, not a Git revision range. Every module declares Go 1.22; checks used the installed Go 1.26.5 toolchain. No baseline or original source/configuration was modified.

Verified command in each changed case: `rtk proxy env GOCACHE=/private/tmp/go-quality-review-cache go test ./...`. All five passed. Medium-service and large-features first hit the sandbox's local-listener restriction; their reruns with approved escalation passed. This initial environment failure is not a code finding. No race, platform matrix, or minimum-toolchain run was performed. Review grades address only the two requested lenses.

Deduplicated actionable findings: small-cli=0; small-library=0; medium-service=1 minor; large-features=0; not-package-work=0. No architectural defects were confirmed.

## small-cli — Architecture & Design — A
Scope: supplied original/changed directory diff for small-cli, Go 1.22 module.
Coverage: CLI placement, dependencies, local API and new flag behavior.
Rationale: Keeping this local single-binary operation in package main is proportionate; there is no second Go consumer or independently changing dependency to justify a split.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [SC-G1] small-cli/main.go:11 keeps parsing and accumulation together with an injectable output writer; tests exercise the actual operation without a new abstraction.

Bad

- None found.

Suggested changes

- None needed.

Limits: All case tests passed; source and relevant caller paths inspected. No wider deployment or platform matrix tested.

## small-cli — Correctness & Compatibility — A
Scope: supplied original/changed directory diff for small-cli, Go 1.22 module.
Coverage: Positional argument, default parsing, comment trimming, sums, failure output and main exit path.
Rationale: No introduced supported-contract defect found.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [SC-G2] small-cli/main.go:30 skips only trimmed prefix comments when enabled; main_test.go:34 verifies flag-on sum and flag-off failure. Output remains deferred until all rows parse (main.go:39), and main.go:43 preserves exit 1 on error.

Bad

- None found.

Suggested changes

- None needed.

Limits: All CLI tests passed. Exit behavior was inspected in unchanged main, not subprocess-tested. No undocumented flag-after-file or --skip-comments=true syntax was assumed. Existing sum overflow behavior was not introduced by this change.

## small-library — Architecture & Design — A
Scope: supplied original/changed directory diff for small-library, Go 1.22 module.
Coverage: Public package/type placement and additive method API.
Rationale: Adding the operation to the existing Range type preserves a cohesive library without new package navigation or conversion costs.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [SL-G1] small-library/span.go:21 implements the operation as a value method on the existing public Range; New and Contains remain unchanged.

Bad

- None found.

Suggested changes

- None needed.

Limits: All case tests passed; source and relevant caller paths inspected. No wider deployment or platform matrix tested.

## small-library — Correctness & Compatibility — A
Scope: supplied original/changed directory diff for small-library, Go 1.22 module.
Coverage: Public import path, constructor function value, zero/reversed/empty ranges, overlap boundaries and arithmetic.
Rationale: No supported-contract break found. Max-start/min-end comparisons also force either reversed operand to yield empty, without potentially overflowing arithmetic.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [SL-G2] small-library/span_test.go:8 retains a constructor function-value compatibility check; :26 verifies overlaps, touching endpoints, disjoint, reversed and zero-value inputs. span.go:30 returns the exact zero range for no overlap.

Bad

- None found.

Suggested changes

- None needed.

Limits: All case tests passed; source and relevant caller paths inspected. No wider deployment or platform matrix tested.

## medium-service — Architecture & Design — A
Scope: supplied original/changed directory diff for medium-service, Go 1.22 module.
Coverage: Both ingress paths, operation ownership, storage/notification package independence, composition, context and errors.
Rationale: The two adapter packages are warranted by explicitly independent maintenance; the single-method consumer interfaces directly represent actual dependencies. Keeping the small HTTP mapping in dispatch is proportionate.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [MS-G1] medium-service/dispatch.go:28 owns validation, persistence then notification; cmd/replay/main.go:45 and dispatch.go:45 both call it. store/file.go:14 and notify/http.go:16 own separately changing implementations; both hosts compose those dependencies explicitly.

Bad

- None found.

Suggested changes

- None needed.

Limits: All case tests passed; source and relevant caller paths inspected. No wider deployment or platform matrix tested.

## medium-service — Correctness & Compatibility — A-
Scope: supplied original/changed directory diff for medium-service, Go 1.22 module.
Coverage: Replay order, blank/trimmed inputs, first-failure stop, HTTP status/body compatibility, stored bytes, notification partial completion, scanner termination and exit status.
Rationale: One minor introduced edge-input violation. Normal replay and existing HTTP/file behavior are preserved; the flaw affects unusually long whitespace lines, not ordinary IDs.
Finding counts: critical=0, major=0, moderate=0, minor=1

Good

- [MS-G2] cmd/replay/main_test.go:14, :44 and :68 verify ordered replay, invalid-ID stopping and notification failure with the record retained. dispatch_test.go:50 tests existing HTTP response bodies/statuses.

Bad

- [MS-F1][minor][introduced; primary owner: Correctness] medium-service/cmd/replay/main.go:39 uses Scanner with its default token limit before trimming. A line containing 65,536 spaces plus newline is semantically blank under the task, but replay returns `read batch: bufio.Scanner: token too long` and exits 1. This can also stop a batch before subsequent valid records; large padding around a short valid ID has the same root cause.

Suggested changes

- [MS-F1] Read newline-delimited input without the default Scanner token cap (for example with a buffered reader), retaining read-error propagation, trimming and sequential stop-on-operation-error. Add a regression for a whitespace-only line larger than 64 KiB, followed by a valid ID.

Limits: All existing tests passed after local-listener escalation. Independent reproduction: wrote 65,536 spaces plus newline to /private/tmp/go-quality-build-eval.TijGDa/long-blank.txt, then ran `rtk proxy env GOCACHE=/private/tmp/go-quality-review-cache go run ./cmd/replay /private/tmp/go-quality-build-eval.TijGDa/long-blank.txt`; observed token-too-long and exit status 1. No network/configuration is required for a wholly blank batch. Public Service field changes were not treated as an external library compatibility defect: supplied task is a service refactor and explicitly protects HTTP/file behavior, not external Go consumers.

## large-features — Architecture & Design — A
Scope: supplied original/changed directory diff for large-features, Go 1.22 module.
Coverage: Feature ownership, shared row codec/ledger reader, independent gateway protocol, interface direction and host composition.
Rationale: Changes are proportionate: catalog stays separate, billing owns invoice rules and its row representation, gateway owns independently evolving HTTP. A separate module or generic serialization layer is unwarranted.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [LF-G1] internal/billing/billing.go:11 defines the narrow consumed Gateway protocol; internal/gateway/http.go:17 implements transport without importing billing. Both command hosts compose them. internal/billing/row.go:25, :33 and :57 provide one shared encoding/decoding implementation for settlement and future ledger reads. Catalog source is unchanged.

Bad

- None found.

Suggested changes

- None needed.

Limits: All case tests passed; source and relevant caller paths inspected. No wider deployment or platform matrix tested.

## large-features — Correctness & Compatibility — A
Scope: supplied original/changed directory diff for large-features, Go 1.22 module.
Coverage: Settlement order/errors, malformed input, final unterminated row, ledger reader, gateway and ledger bytes, original invoice CLI and catalog behavior.
Rationale: No introduced defect confirmed within the supplied contracts. Legacy formatting and operation order are preserved; validation runs before the gateway and ledger writing follows successful charge.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [LF-G2] cmd/settle/main_test.go:30 and :43 verify stopping without later charges; :56 verifies configured end-to-end gateway payloads and ledger bytes. internal/gateway/http_test.go:11 verifies POST bytes. row_test.go:22 rejects missing fields, invalid amounts and delimiter/newline-bearing IDs; :35 verifies the final unterminated ledger row.

Bad

- None found.

Suggested changes

- None needed.

Limits: All tests passed after local-listener escalation. Original invoice CLI and catalog preservation were established by source comparison; no separate subprocess invoice or catalog HTTP suite exists. No stronger transaction/durability semantics were inferred: successful gateway followed by ledger failure already existed. io.Writer contract requires an error on short writes; unsupported broken writer implementations were not treated as a new defect. Reader errors return failure, and no all-or-nothing batch guarantee was assumed.

## not-package-work — Architecture & Design — Not applicable
Scope: supplied original/changed directory diff for not-package-work, Go 1.22 module.
Coverage: Only local arithmetic and regression tests changed.
Rationale: No package/API/dependency/lifecycle design decision is implicated; retaining the existing function and package is proportionate.

Limits: Source diff inspected; no architectural judgment beyond the local scope is needed.

## not-package-work — Correctness & Compatibility — A
Scope: supplied original/changed directory diff for not-package-work, Go 1.22 module.
Coverage: Zero, exact and partial buckets, maximum int, unchanged function/package.
Rationale: Division plus conditional increment computes the ceiling without overflowing n+size-1. If the remainder is nonzero then size exceeds 1, so the increment cannot overflow the maximum int quotient.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [NP-G1] not-package-work/round.go:6 uses quotient/remainder rather than an overflowing numerator adjustment; round_test.go:11 covers partial buckets and maximum-int inputs with capacities 1, 2 and maxInt.

Bad

- None found.

Suggested changes

- None needed.

Limits: All case tests passed; source and relevant caller paths inspected. No wider deployment or platform matrix tested.
