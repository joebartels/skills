# Go quality build

Status: candidate package for behavioral evaluation. Structural validation does not establish improvement; see the [design record](../../docs/go-quality-build/README.md) and [evaluation evidence](../../tests/go-quality-build/results/2026-10-01-package-boundaries-baseline/README.md).

The standalone `go-package-boundaries` skill applies when Go work adds or moves responsibilities, changes packages, or changes import direction. It supports libraries, CLIs, services, and workers. A local arithmetic or expression fix does not need it. Package maps illustrate choices without prescribing a layout.

API-contract and interface/composition skills remain proposed and are not included. The separate `go-quality-review` package can independently review completed work; this package does not grade itself.

## Local use and installation

From the repository root, Claude Code can load this candidate with `claude --plugin-dir ./plugins/go-quality-build`; its skill is `/go-quality-build:go-package-boundaries`. Through the repository marketplace, install `go-quality-build@joebartels-skills` using the same harness commands documented in the [repository guide](../../README.md).

Codex uses the root portable manifest and the repository's `.agents/plugins/marketplace.json`. Select `go-package-boundaries` after loading the package. OpenCode v2 uses the skills source in the root `opencode.json`; for another working directory, use the absolute path to this package's `skills/` directory. Development does not require global installation.

Runtime instructions are canonical under `skills/`; fixtures and evaluation reports remain outside the package. The guidance and examples are newly written from the approved design and audited decision criteria; no upstream source text or code is copied. Upstream provenance and copy/adapt/omit decisions are in the [source audit](../../docs/go-quality-build/source-audit.md).
