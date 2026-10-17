# Authoring Go quality build skills

The [package guide](../../plugins/go-quality-build/README.md) lists skills and their decision owners. Runtime instructions live in `plugins/go-quality-build/skills/`; evaluation fixtures for retained skills belong in `tests/go-quality-build/`.

## Write useful guidance

- Give each decision one owner. Read the relevant existing skill before adding overlapping guidance.
- Preserve supported APIs, Go versions and operation contracts. Prefer the smallest mechanism that satisfies the task.
- State when a rule applies, what decision it changes and how to check the result. Avoid blanket requirements for layouts, interfaces, dependencies or test structure.
- Keep the main skill self-contained and concise. Add references or examples only when they clarify a decision.
- When adapting upstream content, verify it against primary Go documentation and retain required attribution and license notices.

## Check behavior

Use realistic Go tasks, including a case that should skip the skill. Compare equivalent tasks and settings with and without the guidance. Revise concrete failures, then check a fresh task for transfer and regressions.

Verify behavior on supported Go toolchains; use race checks for concurrent work. Assess preserved contracts and complexity, not just passing tests or grades. Packaging validation checks structure, not effectiveness.

Screen candidates in temporary workspaces. Follow the [repository content rule](../../AGENTS.md#repository-content): if a skill is not retained, its candidate-only evaluations, tests, tooling and documentation must not land either. Future reuse alone does not justify keeping them.

## Maintain documentation

Keep reusable guidance, usage and validation instructions. Omit progress logs, handoffs, review reports, revision histories and future proposals. Keep evaluation artifacts outside runtime instructions and use Git history for historical context.
