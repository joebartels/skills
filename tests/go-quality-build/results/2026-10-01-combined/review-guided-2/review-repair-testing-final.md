# Independent Testing addendum: second test-only repair

Reviewed 2026-10-01. Candidate: `/private/tmp/go-quality-build-combined-eval/skill-on/combined-evolution`. This addendum supersedes the remaining T4 finding in `review-repair-testing.md`. The earlier independent results for the other three mutations remain applicable. I compared the final candidate with my saved first-repair baseline: only `fielddesk_test.go` and the author's `review-guided-report.md` changed. Production code and every other test were unchanged. No candidate or repository files were edited; all execution and mutation used fresh disposable copies under `/private/tmp/go-quality-build-combined-eval/independent-testing-final`.

## Testing — A
Scope: Original-to-final repaired Fielddesk changeset, with this recheck focused on the replacement publication test and prior review findings. Go 1.22 module; actual execution used Go 1.26.5 darwin/arm64.
Coverage: Independently inspected the complete second-repair diff, traced the new assertions and resource ownership, recreated the direct-write mutation, ran it ten times with default scheduling and ten times with `GOMAXPROCS=1`, and ran the unmodified race-enabled suite ten times with shuffled order. Previous review evidence covers CLI ingress, host joining, field validation, compatibility, cancellation and worker-cycle tests; those portions are unchanged. Other-platform publication behavior and supported-toolchain execution remain untested.
Rationale: No actionable Testing defect remains in the inspected scope. The previous moderate publication finding is resolved: observation no longer depends on catching an in-flight write. The test checks that a handle opened before publication retains the original bytes and that reopening the path yields exact replacement bytes. All twenty independent direct-write mutation runs now fail. The checks support A; no exceptional-strength A+ claim is needed.
Finding counts: critical=0, major=0, moderate=0, minor=0

Good

- [G1] `fielddesk_test.go:93` opens the snapshot before sync (`:103`), synchronously completes publication (`:113`), then checks the old handle's exact bytes (`:116`) and the new path's exact bytes (`:120`). Replacing temporary-file/rename publication with direct `os.WriteFile` fails 10/10 default and 10/10 single-processor runs at line 118: the old handle reads the new bytes. This directly resolves the demonstrated false-negative scheduling window.
- [G2] The new test uses a small payload and no observer goroutine, gated response body, timer or scheduler yield. `defer oldFile.Close()` (`fielddesk_test.go:107`) closes its handle on success and assertion failure, before the temporary directory's registered cleanup. The previous publication test's failure-path observer cleanup concern is removed with the observer.
- [G3] The unchanged CLI process, host join and text-validation tests retain their previously demonstrated detection of the other three mutations (each failed 10/10 independent first-review runs). The final complete unmodified suite also passes ten shuffled race-enabled runs across all three packages.

Bad

None found.

Suggested changes

None needed for the reviewed repair.

Limits: Results establish behavior on the available macOS filesystem only. `fielddesk_test.go:94` explicitly skips the publication safeguard on Windows; this is a visible coverage limit, not evidence of Windows correctness. No Linux, Windows or Go 1.22 execution was performed, and no cross-platform support promise beyond the supplied contract is inferred. The open-handle test detects in-place mutation; it is not an exhaustive proof of all possible atomic-publication regressions or a simulated disk-full/short-write test. Live listener/signal integration was not rerun. The prior review's optional failure-path cleanup observations in unchanged cancellation tests remain observations, with no reproduced persistent leak or baseline flake. Race detection covers executed paths only.

### Exact checks and evidence

From the final unmodified disposable copy:

`rtk proxy env GOCACHE=/private/tmp/go-quality-build-combined-eval/cache go test -race -shuffle=on -count=10 ./...`

Result: exit 0; `example.com/fielddesk`, `cmd/import-checkins`, and `cmd/server` all passed.

In a separate fresh copy, replaced only the temporary-file/write/close/rename block of `SyncBulletins` with `return os.WriteFile(filepath.Join(s.root, "bulletins.json"), data, 0644)`; no tests were changed. Executed via Python subprocess, with the explicit writable cache:

`go test -json . -run '^TestSyncBulletinsPublicationPreservesOpenSnapshot$' -count=10 -timeout=60s`

- Default scheduling: exit 1, 10 test failures, 0 passes.
- Same command with `GOMAXPROCS=1`: exit 1, 10 test failures, 0 passes.

Every failure was the expected old-handle byte assertion, not compilation or setup failure. Raw JSON logs are `/private/tmp/go-quality-build-combined-eval/independent-testing-final/direct-write-default.jsonl` and `/private/tmp/go-quality-build-combined-eval/independent-testing-final/direct-write-1.jsonl`.
