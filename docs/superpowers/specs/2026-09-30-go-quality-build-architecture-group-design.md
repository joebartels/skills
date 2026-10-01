# Go quality build: first architecture group

## Intent and success criteria

Add a small, multi-harness collection that helps an agent **write or change** Go code with sound architecture. The first group contains three independently useful skills: `go-package-boundaries`, `go-api-contracts`, and `go-interfaces-and-composition`. They complement the existing Go quality review skills; they do not grade their own output or promise an A rating. The intended user is an agent working in a Go library, CLI, or service with a concrete change to make.

Success means an agent selects a relevant skill at the right time, makes a better code change than it would without that skill, and avoids unnecessary layers, interfaces, constructors, or compatibility claims. Each skill must work alone and agree with the other two on a task that invokes more than one. Evidence comes from realistic before/after tasks, builds and tests where applicable, and independent review of the resulting code, including cases where a familiar blanket rule is wrong.

## Package and delivery

Ship the group as a peer `plugins/go-quality-build/` package with one canonical `skills/<name>/SKILL.md` per skill and focused references only where they improve decisions. Use the repository's established portable root `plugin.json`, Claude compatibility manifest, Codex and Claude marketplace entries, and OpenCode skills source. Keep evaluation cases, fixtures, and reports under `tests/go-quality-build/`, outside the installable plugin. Document installation and skill selection in the package README. Retain the `go-quality-review` package as an independent reviewer.

The three skills have no required coordinator. A later coordinator may route tasks, but it must not repeat their rules or adjust review grades. Every skill's description should make both its trigger and its boundary clear. An agent can invoke more than one skill when the change spans decisions; the skills should identify a primary owner for each decision and refer briefly to related topics rather than issuing competing directions.

## Skill responsibilities

### `go-package-boundaries`

Use when a Go change creates, splits, merges, or relocates packages, or changes import direction and responsibility ownership. Guide the agent to inspect existing packages and consumers, state the responsibility of each affected package, trace dependencies, and choose the smallest boundary that reduces real coupling. Assess whether code belongs together based on cohesion and likely change paths. A small service may keep HTTP, SQL, and orchestration in one package. A reusable capability may warrant a technical package. Neither a fixed directory tree nor mandatory domain/ports/adapters layers are prescribed.

This skill owns package placement and dependency direction. It can flag that a public API or interface is driving a package dependency, but `go-api-contracts` owns the consumer-visible contract and `go-interfaces-and-composition` owns whether the abstraction is justified. Its result should be a concrete package decision and the minimal code movement needed to implement the task, not a speculative whole-repository reorganization.

The skill must include a short reference with **illustrative, runnable-shaped package maps** for small, medium, and large projects. Each map names package responsibilities and shows allowed import direction; no map is a required layout or a size threshold. The examples should use one evolving service so an agent can see where HTTP transport, a consumer-side persistence port, a SQL implementation, and composition move as needs arise:

- **Small:** a `cmd/orders` entry point wires a cohesive `internal/orders` package. HTTP handling and SQL queries can live in separate files of that one package when their coupling is small and the package remains understandable. No port interface or adapter package is invented merely for future flexibility.
- **Medium:** when HTTP and database details start forcing unrelated changes into core order behavior, retain `internal/orders` for the use case and its actually needed `Store` interface, add `internal/ordershttp` for request/response translation and `internal/orderspostgres` for the database implementation, and wire them in `cmd/orders`. Both adapters may import `internal/orders`; `internal/orders` must not import either adapter. An HTTP-only service interface may instead live in `internal/ordershttp`; a producer-defined protocol may own its own interface. The example must explain those distinct choices.
- **Large:** show multiple feature packages with their own transport and persistence adapters, an application entry point that composes them, and a shared package only for a genuinely reusable capability. Explain when separate binaries or modules would be justified by independent ownership, release, or reuse needs, not source line count. Avoid a central `domain`, `ports`, or `adapters` hierarchy that forces unrelated features together.

The reference must show the change that motivates each split, the dependency direction after it, and a counterexample where keeping one package is still better. It should explicitly distinguish a new file from a new package, a port interface from an adapter implementation, and HTTP ingress from the business operation. The skill should lead agents to adapt the maps to actual consumers and project conventions rather than reproduce the names.

### `go-api-contracts`

Use when a change creates or modifies an exported Go symbol or another consumer-visible contract, including CLI behavior, wire and file formats, or documented error behavior. Guide the agent to identify actual consumers, supported versions, release policy, stated semantics, and migration needs before choosing an API shape. Check source compatibility beyond call syntax, such as function-value assignments and external implementations of exported interfaces. Preserve intentional compatibility when required; when the project deliberately changes a contract, make the change and migration explicit. Do not impose Go's standard-library compatibility promise on every third-party module.

This skill owns consumer-visible shape and evolution. It should help the agent write documentation and contract-focused tests that capture the intended behavior. Package placement is delegated to `go-package-boundaries`; interface necessity and wiring are delegated to `go-interfaces-and-composition`. For an internal-only edit with no observable contract change, this skill should not activate merely because the code is Go.

### `go-interfaces-and-composition`

Use when a change introduces or alters an interface, concrete dependency, constructor, options API, dependency wiring, or resource lifecycle. Guide the agent to start from real callers and required variation; use concrete types or functions when sufficient, small interfaces for genuine consumer needs, and producer-owned interfaces when the package deliberately defines a protocol. Make important runtime dependencies and ownership visible. Choose constructors and options only when invariants or a useful public API require them; preserve a useful zero value where feasible. Do not require an interface for every test fake, a constructor for every type, or functional options by default.

This skill owns abstraction and composition decisions, including who configures, starts, and closes acquired resources. `go-api-contracts` owns whether a public interface or constructor change preserves the supported consumer contract. `go-package-boundaries` owns where the abstraction belongs. The skill should also avoid library constructors taking process-wide control unless that is an explicit host contract.

## Shared operating pattern and conflict resolution

Each skill should ask the agent to inspect the code and task contract first, make the smallest justified design change, then verify the affected behavior. Recommendations must cite observed code, callers, or a stated requirement. When evidence is missing, the skill should identify the needed fact and avoid inventing a project policy. Guidance must respect the module's supported Go version and the project's conventions where they are compatible with the task.

For overlapping tasks, decide package responsibility first, abstraction and ownership second, and the final consumer-visible API shape with a compatibility check before implementing. This is a reasoning order, not a demand for three skill invocations. One skill may proceed alone using its own boundary checks. If advice conflicts, resolve it at the owning decision above and update the relevant skill; do not add a second universal rule.

The build skills may use the review decision references for consistency, but must express actions an author can take. They must not embed the review report template, grades, severity counts, or instructions to hide findings. Independent review of completed changes remains a separate step.

## Upstream reuse

Use `samber/cc-skills-golang` at commit `19a0626ae8565d27a7b7bdf59d8d99d94d7e284c` as an input, particularly its project-layout, design-patterns, structs-interfaces, dependency-injection, and naming material. Before copying any section into a runtime skill, record a copy/adapt/omit decision and its local owner. Keep examples that help an agent make a sound change. Adapt or omit fixed directory layouts, unconditional consumer-owned interfaces, constructors for all dependencies, and functional options as the default. Remove upstream harness-specific directions and broken cross-links. Retain the MIT copyright and permission notice when copying a skill or substantial portion.

Primary Go documentation and the local Architecture and Correctness review decision references resolve version-sensitive or conflicting advice. The upstream evaluation report is useful background but does not establish effectiveness under this collection's rubric.

## Evaluation and promotion

Implement and evaluate one skill before adding the next to the installable package, in this order: package boundaries, API contracts, interfaces and composition. For each skill, test both selection and non-selection, then run representative Go change tasks with and without the skill. Include a case where common prescriptive advice would harm the code. Assess completed code with applicable builds and meaningful tests, then use the existing review skills independently. Record harness, model, skill revision, task, scope, outcomes, review findings, and limits. A test that only checks whether the agent repeats a rule is insufficient.

Before the group is called coherent, run at least one combined task that exercises all three decision areas and check for contradictory guidance, unnecessary code, or a missed consumer break. Prefer an observed reduction in confirmed defects and unnecessary complexity over a grade alone. If a skill adds no distinct benefit, merge its useful content into another skill or leave it as reference material. Report a skill as evaluated only for the cases actually run; do not generalize one task to all Go projects.

The repository validator should accept the new package, check manifests and local links, and validate the new evaluation data without weakening the existing review checks. Run the validator and available harness validators. Record any unavailable runtime checks explicitly.

## Scope and failure handling

This first group covers architecture-oriented writing decisions. Error handling, concurrency, security, testing strategy, deployment, and other candidate topics remain future work unless needed to complete and verify a concrete evaluation task. Do not modify the review skills merely to improve build-skill scores. Do not install the package into the user's global harness state during development.

If a behavioral case reveals harmful guidance, revise the decision owner and rerun that case before promotion. If a task lacks enough contract or consumer context, state the uncertainty and seek the smallest relevant evidence rather than filling it with a generic best practice. Keep the [Go quality build design record](../../go-quality-build/README.md) current with exact artifacts, evidence, decisions, and the next action.
