# Personal agent skills

Reusable skills for Claude Code, Codex, and OpenCode v2. Each collection lives in one `plugins/<name>/` directory; its `skills/<skill-id>/SKILL.md` files are the canonical runtime content. Harness manifests and catalogs only point to those files.

## Collections

| Package | Skills | Purpose |
| --- | ---: | --- |
| [go-quality-review](plugins/go-quality-review/README.md) | 10 | Graded Go code reviews: one overall report skill and nine focused topic skills. |
| [go-quality-build](plugins/go-quality-build/README.md) | 7 authoring skills | Package boundaries, APIs/composition, testing, context/deadlines, and concurrency/ownership. |

The detailed [Go review guide](docs/go-quality-review/README.md) covers grading, evaluations, and authoring. The [architecture and implementation plan](docs/superpowers/specs/2026-09-29-multiharness-skills-design.md) record the harness choices.

The [Go quality build design record](docs/go-quality-build/README.md) tracks decision owners, revision-specific evidence and open limits. Version `0.3.0` adds [context/deadlines](plugins/go-quality-build/skills/go-context-and-deadlines/SKILL.md) and [concurrency/ownership](plugins/go-quality-build/skills/go-concurrency-and-ownership/SKILL.md) to the five existing authoring skills. Their [small recovery loop](docs/go-quality-build/context-concurrency-recovery.md) preserves strong baselines and checks practical budget, cancellation, supervision and resource-lifetime guard rails on Go 1.22.12 and 1.26.5. Targeted ownership/error-origin corrections are documented separately from frozen comparisons. Broader effectiveness, automatic routing and Codex/OpenCode runtime loading remain unverified. Use the installation commands below with `go-quality-build` to load the package.

## Use the skills

### Claude Code

For local development from this repository root, run `claude --plugin-dir ./plugins/go-quality-review`. To install the collection through its marketplace:

```sh
claude plugin marketplace add /absolute/path/to/this/repo
claude plugin install go-quality-review@joebartels-skills
```

Claude names an installed skill `/go-quality-review:<skill-id>`. Its [plugin and marketplace docs](https://code.claude.com/docs/en/plugins/create) describe local development and installation.

### Codex

This repository's `.agents/plugins/marketplace.json` exposes the same package. In the Codex desktop app, open the Plugins Directory and choose the `joebartels-skills` source after opening this repository and restarting the app. You can also add this repository as a marketplace with the CLI, then install the plugin:

```sh
codex plugin marketplace add /absolute/path/to/this/repo
codex plugin add go-quality-review@joebartels-skills
```

The package's root `plugin.json` is the [portable Agent Plugins manifest](https://developers.openai.com/plugins/build/plugins); `skills/` is discovered from the package root. The ten skills remain individually selectable after installation.

### OpenCode v2

From this repository root, `opencode.json` adds `./plugins/go-quality-review/skills` as a skills source. To use the collection while working in other projects, add its **absolute** path to your user-level `opencode.json` or `opencode.jsonc`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "skills": ["/absolute/path/to/this/repo/plugins/go-quality-review/skills"]
}
```

OpenCode v2 resolves a relative `skills` source from the **active working directory**, even if a config was found elsewhere. Use an absolute path for cross-project use, or copy selected skill directories into one of its [discovered skills directories](https://opencode.ai/v2/docs/skills/). A skills-only collection does not need executable OpenCode plugin code. The local OpenCode binary used during this migration is v1.18.7; use a v2 session to verify loading by its exact skill ID.

## Develop and validate

Use Python 3.10+ and install the development dependency before running the repository validator. From the repository root:

```sh
python3 -m pip install -r requirements-dev.txt
python3 scripts/validate.py
python3 -m unittest discover -s tests -p 'test*.py'
python3 -m unittest discover -s tests/go-quality-build -p 'test_*.py'
python3 tests/go-quality-review/test_layout.py
python3 tests/go-quality-review/go-quality-report/evals/test_grade.py
claude plugin validate ./plugins/go-quality-review
claude plugin validate ./plugins/go-quality-build
claude plugin validate .
```

The Python validator checks skill frontmatter and local references, shared Go review contracts and evaluation fixtures, build evaluation suites, plus harness manifests and catalog paths. The calculator tests check the grade arithmetic. These are structural checks; behavior prompts under `tests/go-quality-review/` and `tests/go-quality-build/` require separate model runs and independent review. `claude plugin validate` checks Claude's package and marketplace formats. OpenCode v2 loading and Codex installation require a session in those harnesses.

To add a skill to an existing package, create `skills/<unique-id>/SKILL.md` with `name` and a trigger-focused `description`; keep supporting files in that skill's own directory. For a new, independent collection, create `plugins/<package-name>/` with a portable `plugin.json`, `.claude-plugin/plugin.json`, and `skills/` tree; add one entry to each marketplace and one skills source to `opencode.json`. Keep the package name and version equal in both manifests, and increment both versions when releasing changed skills so Claude marketplace updates deliver the new content. Add meaningful evaluation cases and run the validator before installation. Go topic skills also participate in the fixed shared review contract under `tests/go-quality-review/shared/`, so adding one requires updating that contract and the umbrella routing guidance.

## References

- [Claude Code plugin components](https://code.claude.com/docs/en/plugins/components), [marketplaces](https://code.claude.com/docs/en/plugins/publish), and [validation](https://code.claude.com/docs/en/plugins/create)
- [Codex portable plugin packaging](https://developers.openai.com/plugins/build/plugins)
- [OpenCode v2 skills](https://opencode.ai/v2/docs/skills/) and [plugins](https://opencode.ai/v2/docs/plugins/)
