### Task 6: Delivery review and package closeout

**Files:** Modify `plugins/go-quality-build/{plugin.json,.claude-plugin/plugin.json,README.md}`, `.claude-plugin/marketplace.json`, root README and canonical record; create whole-package review/check artifacts in the combined run. Add a third-party license/notice file only if the recorded copy scope requires it.

**Interfaces:** Consumes exact promoted runtime/evidence bytes and combined results. Produces matching package version `0.2.0` for the delivered testing group, bounded public claims, a current continuation record and local reviewable commits. If the evaluated outcome changes delivered scope, report the actual scope and update the plan rather than claiming the planned group shipped.

- [ ] **Step 1: Audit final scope and metadata.** Make the two manifests' versions/descriptions agree; preserve existing catalog/OpenCode package paths. Ensure only promoted skills are installable, one canonical runtime copy exists for each, descriptions/activation examples cover actual scope, and READMEs distinguish structural validity, blind behavioral evidence, assisted repairs and unverified routing/toolchains/platforms.
- [ ] **Step 2: Verify final evidence and package.** Reconstruct every new trial patch and verify all input/source/skill/artifact hashes. Run the shared delivery checks, final runtime quick validation and local reference checks. Inspect actual Codex/OpenCode verification capabilities; use available nonmutating loading checks and disclose unavailable ones. Do not infer runtime loading from JSON validity or install globally.
- [ ] **Step 3: Obtain independent whole-package review.** Review source-audit/license decisions, exact skill bytes, triggers/ownership, combinations, regression-sensitivity evidence, claims, metadata and continuation state. Correct confirmed findings with minimal changes; guidance changes re-enter behavioral gates, while documentary corrections need only relevant structural rechecks.
- [ ] **Step 4: Record and commit final delivery.** Append precise checks/results/decisions/limits and next priority to the canonical record. Preserve the older blocked composition trial as incomplete. Commit as `docs: deliver evaluated Go testing skill group`, leave the managed worktree's state explicit and hand over artifacts and measured outcomes. No PR or publication is required by this plan.

## Shared delivery checks

Run from the repository root after promotion and at final delivery:

```sh
rtk proxy python3 -B -m unittest discover -s tests -p 'test*.py'
rtk proxy python3 -B -m unittest discover -s tests/go-quality-build -p 'test_*.py'
rtk proxy python3 -B tests/go-quality-review/test_layout.py
rtk proxy python3 -B tests/go-quality-review/go-quality-report/evals/test_grade.py
rtk proxy python3 -B scripts/validate.py
rtk claude plugin validate ./plugins/go-quality-build
rtk claude plugin validate .
rtk git diff --check
```

Existing expected unit-test counts are root 6, build-eval 10, review layout 1 and grade 12, unless a justified test addition changes them. Review validation remains 10 skills/120 cases; build counts progress 15 → 20 → 25 → 26. These commands establish structural integrity, not benefit. Check newly staged files as well as unstaged changes. If preserved raw patch/review text triggers whitespace diagnostics, preserve evidence bytes and document exact exclusions; runtime/source/document files must pass.

## Execution handoff

The implementation method controls who builds the fixtures, drafts and package; **both methods still require fresh independent behavioral authors and reviewers** for the approved evaluation design.

Recommend native implementation in this session: the six tasks share fixture/archive conventions and promotion bookkeeping, so retaining that context should reduce repeated setup. Independent outcome reviews remain mandatory at the documented gates, with the most capable available reviewer for the final whole-package review. Subagent-driven implementation is also available for fresh task-level implementation/spec/quality review contexts before each next task. Obtain user plan review and method selection before Task 1.
