# Go quality build

Version `0.3.0` contains seven focused Go authoring skills. Select the skills that apply to the change; their rules follow the project's supported API, Go version and operation contracts.

| Skill | Decisions it owns |
| --- | --- |
| [go-package-boundaries](skills/go-package-boundaries/SKILL.md) | Responsibilities, package placement and dependency direction. |
| [go-api-contracts](skills/go-api-contracts/SKILL.md) | Consumer-visible behavior, compatibility and deliberate API evolution. |
| [go-interfaces-and-composition](skills/go-interfaces-and-composition/SKILL.md) | Necessary abstractions, construction, dependencies and host-owned lifecycle. |
| [go-behavior-tests](skills/go-behavior-tests/SKILL.md) | Meaningful cases, independent assertions and observable contract boundaries. |
| [go-test-isolation](skills/go-test-isolation/SKILL.md) | Dependency fidelity, fixtures, process state and asynchronous test cleanup. |
| [go-context-and-deadlines](skills/go-context-and-deadlines/SKILL.md) | Propagation, total/stage budgets, result policies, cancel ownership and required finalization. |
| [go-concurrency-and-ownership](skills/go-concurrency-and-ownership/SKILL.md) | Shared invariants and aliases, admission/capacity, supervision, channels and stop/join/release. |

Pure calculations need neither context nor concurrency machinery. Context cancellation requests stopping; owned work must actually complete before its resources are released. The separate [Go quality review package](../go-quality-review/README.md) assesses completed code independently.

## Evidence and limits

Context revision 2 and concurrency revision 1 have [compact recovery evidence](../../docs/go-quality-build/context-concurrency-recovery.md): matched development/transfer outputs, actual Go 1.22.12 and 1.26.5 checks, applicable race tests, routing controls and focused independent reviews. Frozen comparisons preserve strong baselines. Targeted development checks support the redirect-body ownership correction and retention of an independent acquisition error during peer stopping. These bounded observations do not establish general causal effectiveness or every schedule.

The [design record](../../docs/go-quality-build/README.md) preserves prior revision-specific improvements, mixed outcomes and unverified boundaries for all skills. Existing testing mains remain behavior revision 2 and isolation revision 3. Automatic harness routing, Codex/OpenCode runtime loading, other platforms and noncooperative I/O remain unverified. Error/value drafts and names/docs reference work are separate from this package.

## Local use and installation

Claude Code loads this package from the repository root with `claude --plugin-dir ./plugins/go-quality-build`. Invoke a skill as `/go-quality-build:<skill-name>`, using the names in the table. The repository marketplace installs `go-quality-build@joebartels-skills`; see the [repository guide](../../README.md) for harness commands.

Codex uses `plugin.json` and the root `.agents/plugins/marketplace.json`. OpenCode v2 uses the skills source in root `opencode.json`; another working directory can use an absolute path to this package's `skills/`. Development does not require global installation.

Runtime instructions are canonical under `skills/`; fixtures and evaluation reports stay outside the package. Original/adapted guidance and source decisions are recorded in the [architecture audit](../../docs/go-quality-build/source-audit.md), [testing audit](../../docs/go-quality-build/testing-source-audit.md) and [recovery summary](../../docs/go-quality-build/context-concurrency-recovery.md).
