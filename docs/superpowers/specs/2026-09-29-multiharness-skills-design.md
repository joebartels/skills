# Multi-harness skill collection design

## Intent and success criteria

This repository is a personal, extensible collection of skills. The existing ten Go review skills must retain their names, instructions, supporting references, grading behavior, and evaluation coverage. A single source copy of each skill must be usable in Claude Code, Codex, and OpenCode v2. Adding a later collection should require one self-contained package and small catalog/configuration updates, without maintaining three copies of its `SKILL.md` files.

The current `golang/` tree is untracked but is user content. Treat it as the migration source. The existing baseline is ten structurally valid skills, 120 evaluation cases, and 12 calculator tests.

## Approach

Use one plugin directory per coherent collection. The first is `plugins/go-quality-review/`, containing a portable root `plugin.json`, a Claude compatibility manifest at `.claude-plugin/plugin.json`, and canonical runtime skills at `skills/go-*/`. The Go umbrella skill and nine topic skills install together, while each skill remains independently invocable. Future independent collections can be added as peer directories under `plugins/`.

This follows the shared `skills/<name>/SKILL.md` layout supported by [Claude Code](https://code.claude.com/docs/en/plugins/components) and [portable Codex plugins](https://developers.openai.com/plugins/build/plugins). Claude Code uses the child `.claude-plugin/plugin.json`; Codex uses root `plugin.json`. Keep identity and version aligned in validation. No symlinks or generated skill copies are needed.

OpenCode v2 loads the same `plugins/go-quality-review/skills` directory through its documented [`skills` source](https://opencode.ai/v2/docs/skills/). A root `opencode.json` makes the collection available when OpenCode runs from this repository root. For use from other projects, documentation shows an absolute path in the user's OpenCode configuration or copying the skill directories into a discovered skill root. No OpenCode executable plugin is added because this collection defines no hooks or tools; [OpenCode plugins](https://opencode.ai/v2/docs/plugins/) serve those extension points.

## Repository layout

```text
plugins/
  go-quality-review/
    plugin.json
    .claude-plugin/plugin.json
    skills/go-*/SKILL.md
    skills/go-*/references/...
    skills/go-quality-report/scripts/grade.py
    README.md
.agents/plugins/marketplace.json
.claude-plugin/marketplace.json
opencode.json
tests/go-quality-review/
  shared/...
  go-*/evals/...
docs/go-quality-review/...
scripts/validate.py
README.md
```

Evaluation cases, fixtures, historical results, and authoring handbooks live outside the distributed plugin. Historical reports retain their recorded paths as provenance. The validator resolves current paths from the repository root, checks local links and fixtures, and preserves the existing contract checks and report-format option.

## Distribution and use

- Claude Code: the root `.claude-plugin/marketplace.json` points to `./plugins/go-quality-review`; use `claude plugin marketplace add` and `claude plugin install`, or `claude --plugin-dir ./plugins/go-quality-review` for local development. [Claude marketplace docs](https://code.claude.com/docs/en/plugins/publish)
- Codex: `.agents/plugins/marketplace.json` points to the same directory; its portable manifest at `plugin.json` makes the plugin installable from a local or Git marketplace. [Codex packaging docs](https://developers.openai.com/plugins/build/plugins)
- OpenCode v2: the root config reads the package's `skills/` directory when launched at the repository root. A user-level absolute `skills` source works across projects. Relative source paths resolve from OpenCode's active working directory. [OpenCode skills docs](https://opencode.ai/v2/docs/skills/)

The repository documents exact commands and the distinction between checked structural compatibility and a real model invocation. The installed local OpenCode binary is v1.18.7, so this environment cannot run a v2 loading test.

## Validation

The repository validator checks plugin identities, marketplace targets, OpenCode skills source, unique skill IDs, frontmatter, references contained within each skill, shared Go review contract, evaluation inputs and fixtures. It rejects a missing package, mismatched names or versions, broken paths, or duplicate skill IDs. Existing Go report-shape checks remain available.

Run the validator and calculator tests after migration. Run Claude's plugin validator where supported. Inspect Codex's marketplace/plugin resolution with the installed CLI without installing or changing global configuration. Record OpenCode v2 runtime validation as unavailable unless a v2 binary becomes available.

## Scope decisions

The collection is skills-only. Do not add an MCP server, hooks, commands, agents, executable OpenCode plugin, publication workflow, or marketplace installation into the user's global state. Do not rewrite the Go review guidance as part of this packaging change. Preserve all existing evaluation data. The package's initial version is `0.1.0`; future package changes should update both manifests together.
