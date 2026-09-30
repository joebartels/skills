# Multi-harness Skills Implementation Plan

> **For agentic workers:** Use the design and this plan as the contract. Track each checkbox and report exact verification results. Worker tasks 1 and 2 own disjoint paths and may run in parallel; task 3 integrates them.

**Goal:** Package the existing ten Go review skills once for Claude Code, Codex, and OpenCode v2, with a reusable layout for later collections.

**Architecture:** `plugins/go-quality-review/skills/` is the only runtime skill source. Claude and Codex read small manifests and marketplace catalogs pointing at that package. OpenCode v2 reads its skills directory directly. Evaluations and authoring material move outside the installable package.

**Tech Stack:** Agent Skills Markdown/YAML, JSON manifests, Python 3.10+ and PyYAML for existing validation, Python standard-library calculator, Claude Code 2.1.207, Codex CLI 0.157.0. OpenCode v2 docs guide its layout; the locally installed binary is v1.18.7.

**Spec:** [Multi-harness skill collection design](../specs/2026-09-29-multiharness-skills-design.md)

## Global constraints

- Keep all ten skill IDs and their instructions and references intact; do not rewrite review behavior.
- Keep all 120 existing evaluation cases, fixtures, reports, and 12 calculator tests.
- No duplicate or symlinked runtime skill trees.
- Plugin name `go-quality-review`, version `0.1.0`, in both portable and Claude manifests.
- Do not install plugins into the user's global state, publish, or add executable hooks, tools, or MCP servers.
- Every shell command is prefixed with `rtk` per `/Users/jb/.codex/RTK.md`.

## Review focus

1. **An added Go skill:** the validator discovers it from `plugins/go-quality-review/skills/` and requires a unique ID and corresponding evaluation suite.
2. **A broken sibling reference:** each `SKILL.md` and reference Markdown link remains within its own skill and resolves after relocation.
3. **A stale catalog target:** both marketplace entries resolve to the package root and agree with its manifest name.
4. **A version or name drift:** root portable and Claude compatibility manifests must match.
5. **A deep OpenCode working directory:** documentation must explain that relative `skills` sources use the active working directory and provide an absolute cross-project example.

---

### Task 1: Migrate canonical Go content and existing validator

**Owner paths:** `golang/` (source removal), `plugins/go-quality-review/skills/`, `tests/go-quality-review/`, `docs/go-quality-review/`, `scripts/validate.py`. Do not edit manifests, marketplaces, `opencode.json`, or the root `README.md`.

**Interfaces:** Produce ten directories at `plugins/go-quality-review/skills/go-*/SKILL.md`; produce `scripts/validate.py` with CLI `python3 scripts/validate.py [--reports PATH]` and equivalent existing structural/contract/report checks; produce `tests/go-quality-review/go-quality-report/evals/test_grade.py` that runs the packaged calculator.

- [ ] Record baseline `rtk proxy python3 golang/scripts/validate.py` and `rtk proxy python3 golang/go-quality-report/evals/test_grade.py`.
- [ ] Add or adapt a failing layout check that looks for the ten packaged skill paths and separated evaluation fixtures; run it and confirm the intended missing-path failure.
- [ ] Move each `golang/go-*` runtime `SKILL.md`, `references/`, and umbrella `scripts/grade.py` into the package's `skills/` directory. Move each skill's `evals/` tree into `tests/go-quality-review/<skill>/evals/` without losing fixtures or historical results.
- [ ] Move `golang/evals/` to `tests/go-quality-review/shared/`, `.resources/` to `docs/go-quality-review/resources/`, and its large README to `docs/go-quality-review/README.md`. Remove the empty `golang/` tree. Keep archived results' historic path text as provenance.
- [ ] Move/adapt the validator to `scripts/validate.py`: resolve repository root from the script path; read the contract under `tests/go-quality-review/shared/`; discover runtime skills under the plugin; read suites and fixtures under `tests/go-quality-review/<skill>/evals/`; keep `--reports` behavior. Adapt `test_grade.py` to load the packaged calculator.
- [ ] Update paths in the moved authoring README and make it clear where runtime skills and evals live. Do not add links from runtime skill files outside their own directories.
- [ ] Run `rtk proxy python3 scripts/validate.py` and `rtk proxy python3 tests/go-quality-review/go-quality-report/evals/test_grade.py`; expect 10 skills, 120 evaluation cases, and 12 passing tests.

### Task 2: Add three harness entry points

**Owner paths:** `plugins/go-quality-review/plugin.json`, `plugins/go-quality-review/.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, `.agents/plugins/marketplace.json`, `opencode.json`. Do not edit Task 1 paths or documentation.

**Interfaces:** The package is `./plugins/go-quality-review`. Codex root manifest and Claude compatibility manifest identify `go-quality-review` at `0.1.0`. Both marketplaces expose that package. Root `opencode.json` lists `./plugins/go-quality-review/skills` as its source.

- [ ] Write a small failing JSON/path assertion or equivalent check that rejects absent manifests/catalogs/config, and run it.
- [ ] Add the portable root `plugin.json` using the Agent Plugins schema `https://agent-plugins.org/schemas/1.0.0/plugin.schema.json`; include only supported identity fields and no runtime components beyond the conventional `skills/` directory.
- [ ] Add Claude's `.claude-plugin/plugin.json` with the same name/version and a short description; the default `skills/` discovery path supplies the skills.
- [ ] Add the Claude marketplace with `name`, `owner`, and a local source `./plugins/go-quality-review`; add the Codex marketplace with a local source to the same path and the required policy/category fields.
- [ ] Add root `opencode.json` with the official schema and one `skills` entry `./plugins/go-quality-review/skills`.
- [ ] Run JSON syntax/path assertions and, if supported by installed Claude CLI, `rtk proxy claude plugin validate ./plugins/go-quality-review` and `rtk proxy claude plugin validate .`; report any version-related CLI limitation instead of changing the documented v2 format to satisfy an older runtime.

### Task 3: Integrate validation and documentation

**Owner paths:** `README.md`, `plugins/go-quality-review/README.md`, `scripts/validate.py`, optional `tests/test_packaging.py`, and this plan/spec for factual corrections. Task 1 is complete before editing `scripts/validate.py`.

**Interfaces:** `scripts/validate.py` remains the single repository validation command. Documentation gives reproducible installation/development instructions for all three harnesses and the add-a-plugin workflow.

- [ ] Add a failing validator case for a missing/mismatched manifest, an invalid marketplace target, a duplicate skill ID, and a broken OpenCode source; verify each fails for its intended reason.
- [ ] Extend `scripts/validate.py` to check the two manifest identities/version, catalog names and local paths, OpenCode source, and unique portable skill IDs. Keep the existing Go contract/eval checks.
- [ ] Write the root README with the directory map, install/use instructions for Claude, Codex, and OpenCode v2, exact validation commands, and steps to add a new collection. Document the OpenCode active-working-directory limitation and local v1/v2 test boundary.
- [ ] Write a compact package README that names the ten skills and links within the package; link authoring details from the root README to `docs/go-quality-review/README.md`.
- [ ] Re-run validator and calculator tests. Run `rtk proxy claude plugin validate` on package and marketplace, Codex CLI marketplace listing/inspection only if read-only, and JSON parser checks. Inspect `git diff` and `git status` for dropped or duplicate content.
- [ ] Report which harnesses had a live runtime validation and which had only structural checks. Do not claim OpenCode v2 loaded a skill without a v2 binary/session.
