# Go quality build

Focused Go authoring skills. Select those relevant to the change; follow the project's supported API, Go version and operation contracts.

| Skill | Decisions it owns |
| --- | --- |
| [go-package-boundaries](skills/go-package-boundaries/SKILL.md) | Responsibilities, package placement and dependency direction. |
| [go-api-contracts](skills/go-api-contracts/SKILL.md) | Consumer-visible behavior, compatibility and API evolution. |
| [go-interfaces-and-composition](skills/go-interfaces-and-composition/SKILL.md) | Abstractions, construction, dependencies and host-owned lifecycle. |
| [go-names-and-comments](skills/go-names-and-comments/SKILL.md) | Clear names and concise comments that retain necessary contracts and warnings. |
| [go-behavior-tests](skills/go-behavior-tests/SKILL.md) | Meaningful cases, independent assertions and observable contract boundaries. |
| [go-test-isolation](skills/go-test-isolation/SKILL.md) | Dependency fidelity, fixtures, process state and asynchronous test cleanup. |
| [go-context-and-deadlines](skills/go-context-and-deadlines/SKILL.md) | Propagation, time budgets, cancellation policies, scope ownership and required finalization. |
| [go-concurrency-and-ownership](skills/go-concurrency-and-ownership/SKILL.md) | Shared invariants, aliases, capacity, supervision, channels and stop/join/release. |
| [go-client-calls](skills/go-client-calls/SKILL.md) | Outbound HTTP/unary gRPC operation policy, replay safety, retry ownership, remote failures and breaker health. |

Use the separate [Go quality review package](../go-quality-review/README.md) to assess completed code.

## Use and develop

Load locally with `claude --plugin-dir ./plugins/go-quality-build` and invoke `/go-quality-build:<skill-name>`. See the [repository guide](../../README.md) for installation in Claude Code, Codex and OpenCode.

Runtime content is canonical under `skills/`. For changes to the collection, follow the [authoring guide](../../docs/go-quality-build/README.md).
