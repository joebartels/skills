# Go quality build

Status: `go-package-boundaries` is implemented and behaviorally evaluated. Its [baseline](../../tests/go-quality-build/results/2026-10-01-package-boundaries-baseline/README.md) and [skill-on evidence](../../tests/go-quality-build/results/2026-10-01-package-boundaries-skill-on/README.md) show a bounded improvement on a medium-service ownership decision after revision, with no observed package overbuilding in the control cases. Automatic routing and Codex/OpenCode runtime loading remain unverified; see the [design record](../../docs/go-quality-build/README.md).

The standalone `go-package-boundaries` skill applies when Go work adds or moves responsibilities, changes packages, or changes import direction. It supports libraries, CLIs, services, and workers. A local arithmetic or expression fix does not need it. Package maps illustrate choices without prescribing a layout.

`go-api-contracts` is a runtime candidate under final review, not yet promoted. It applies to exported APIs and supported external CLI, wire/file and error contracts; private helper fixes without an established external contract do not need it. [Skill-on evidence](../../tests/go-quality-build/results/2026-10-01-api-contracts-skill-on/README.md) records a fixed exported zero-value regression, targeted v2 and CLI corrections, and selection limits. Interface/composition remains proposed and is not included. The separate `go-quality-review` package can independently review completed work; this package does not grade itself.

## Local use and installation

From the repository root, Claude Code can load this package with `claude --plugin-dir ./plugins/go-quality-build`; its evaluated skill is `/go-quality-build:go-package-boundaries`, and its review candidate is `/go-quality-build:go-api-contracts`. Through the repository marketplace, install `go-quality-build@joebartels-skills` using the same harness commands documented in the [repository guide](../../README.md).

Codex uses the root portable manifest and the repository's `.agents/plugins/marketplace.json`. Select `go-package-boundaries` after loading the package; `go-api-contracts` is also discoverable as an unpromoted review candidate. OpenCode v2 uses the skills source in the root `opencode.json`; for another working directory, use the absolute path to this package's `skills/` directory. Development does not require global installation.

Runtime instructions are canonical under `skills/`; fixtures and evaluation reports remain outside the package. The guidance and examples are newly written from the approved design and audited decision criteria; no upstream source text or code is copied. Upstream provenance and copy/adapt/omit decisions are in the [source audit](../../docs/go-quality-build/source-audit.md).
