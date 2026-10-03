# Independent Go pipeline/CLI review

Candidate A: **C-**, three unique major causes. Candidate B: **C-**, two unique major causes. Complete bounded packet coverage; all-nine applicability and cards are recorded separately.

| Candidate | Correctness | Code quality | Testing | Unique critical/major/moderate/minor |
| --- | --- | --- | --- | --- |
| A | C | C | C- | 0 / 3 / 0 / 0 |
| B | C | C | C | 0 / 2 / 0 / 0 |

Both lose an independent DeadlineExceeded error during cleanup when the coordinated child stop is Canceled. B additionally loses bare independent Canceled errors before stopping, in both callback roles; those are manifestations of B’s single production classification cause. A’s stop snapshot avoids that pre-stop failure but does not preserve the actual stop class.

A’s narrowed premature-join mutant compiles and survives its full delivered suite at count=10; a held cleanup gate catches host return before cleanup release. B’s existing join test detects that mutation. A also misses a broader consumer-cleanup error suppression mutation; B’s wrapped-both-error assertion detects it. Each candidate separately has a missing exact-error/timing regression assertion gap.

Actual host and Go1.22.12 binaries preserve checked integer ordering, early prefix, stderr and exit codes. Both pass ordinary build/test/vet and have empty formatting diffs. A passes supplied held checks; B fails the supplied independent cancellation-class test. Independent deadline-after-stop failures are recorded separately.

## Artifacts

- [A full report](review-A.md), [B full report](review-B.md)
- [Unique findings and strengths](findings.json)
- [A ledger](ledger-A.json), [B ledger](ledger-B.json), [A calculator output](grade-A.json), [B calculator output](grade-B.json)
- [Manifest](manifest.json), [A applicability](applicability-A.json), [B applicability](applicability-B.json)
- [Raw independent results](raw/independent-commands.json), [narrowed/targeted follow-ups](raw/followup-commands.json), [mutation summary](mutation-summary.json)
- [Original source hashes](source-hashes-before.json), [after-review hashes](source-hashes-after.json), [integrity](raw/source-integrity.json)

## Limits

Review is confined to the supplied bounded CLI/pipeline snapshots, original contract, candidate tests and supplied held checks. Actual tested platform is darwin/arm64; no cross-platform packaging, service deployment, signals, containers, rollout, remote telemetry or vulnerability-advisory audit is claimed. Arbitrary blocking Reader cancellation is explicitly excluded by README. Race runs cover the exercised paths only. Original and review-mutant checks are distinguished in raw logs. Package-alarm timeouts are not counted as mutation detection. Review is one independent reviewer performing bounded sequential reviews, without author reports/profiles or external repository plans/results.

No external plans, reports, profiles, other-agent output or exposure identity was used. Four reviewer-created mutation targets are supplementary; the packet contains no original frozen author-specific mutation artifacts. Mutant compilation errors and timeout-only failures receive no detection credit. No fix, promotion or writing-guidance evaluation was performed.
