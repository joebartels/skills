# Go quality build: language contracts group

## Intent and proposed scope

Build the next focused Go authoring skills for agents making concrete changes in existing libraries, CLIs, services, and workers. The user assigned testing work to the separate **Go Skills - testing** chat. This group contains candidates `go-error-contracts`, `go-values-and-zero-values`, and `go-names-and-docs`. The user approved implementation on 2026-10-01, conditioned on enough distinct content and independent effectiveness feedback. These candidates are not implemented or evaluated at this approval stage.

Success means better completed code and clearer, accurate consumer documentation, with fewer confirmed defects or unnecessary changes than matched work without the candidate skill. A skill must contribute decisions beyond the existing architecture group, select appropriately, and avoid harmful blanket rules. Passing examples, structural validation, or an A grade alone do not establish skill benefit.

Use the existing [build design record](../../go-quality-build/README.md), [first-group design](2026-09-30-go-quality-build-architecture-group-design.md), and [review guide](../../go-quality-review/README.md). The architecture group is merged in [PR #3](https://github.com/joebartels/skills/pull/3); this worktree starts at merge commit `b76df28ed15a177d436672d47a5d30f6634900c7`.

## Choice of group and packaging

Three focused skills fit the established decision owners and allow independent evaluation. A single broad idioms skill would be easier to discover but would mix error exposure, value ownership, and readability decisions, making distinct benefit harder to establish. Context/concurrency or security/data groups are useful alternatives, but the recommended group addresses common language-level changes first. These alternatives remain future candidates.

Extend `plugins/go-quality-build/` with one canonical runtime directory per promoted skill. Keep source audits, draft skills, runnable evaluation inputs, trial outputs, and reviews outside the installable package. Drafts belong under `tests/go-quality-build/<candidate>/draft/`; suites and fixtures use the existing `<candidate>/evals/` structure. Use focused references only where examples materially improve decisions. No new coordinator, runtime dependency, global installation, or prescribed linter suite is needed.

Promote one skill at a time, in the order errors, values, then names/docs. Update the existing package README and harness catalogs only after the relevant candidate meets its gate. If names/docs adds little beyond API-contract guidance and existing Go knowledge, retain useful material as a reference or defer the candidate instead of shipping a redundant skill.

## Decision ownership and triggers

| Candidate | Trigger | Owned decision | Boundary |
| --- | --- | --- | --- |
| `go-error-contracts` | Go work changes returned, handled, inspected, translated, or aggregated errors, including partial results and completion errors. | Which failures are exposed, how callers classify them, and how failure propagates through the operation. | API contracts own supported evolution; telemetry owns logging systems; resilience owns retries; concurrency owns stopping and joining work. |
| `go-values-and-zero-values` | Go work changes zero/default state, nil/empty representations, value copying, aliasing, receivers, or interface conversion. | State representation and value ownership, including the observable effects of copies and defaults. | Composition owns whether construction is required; API contracts own compatibility; concurrency owns synchronization design; performance owns optimization claims. |
| `go-names-and-docs` | Go work introduces or changes identifiers, doc comments, package documentation, or user-facing examples whose meaning affects readers. | Readability at actual use sites and accurate explanation of the implemented contract. | Package boundaries own responsibility and placement; API contracts own supported public renames and migrations; testing owns overall test strategy. |

Private edits can activate these skills when their owned decision changes. A pure arithmetic fix, mechanical formatting pass, or unrelated Go task should not activate them merely because it is Go. A name or comment change must not silently authorize behavior changes or a repository-wide cleanup.

### Errors

Guide the author to trace each changed failure through producers, callers, documentation, and the relevant consumer boundary. Choose plain errors, existing sentinels, custom types, wrapping, or translation from the information callers need. Add useful operation context without turning every layer into a wrapper. Preserve required identity/type inspection through chains; use `errors.Is` and appropriate typed inspection when causes may be wrapped. Do not invent exported error classifications for failures consumers never classify.

Treat exposing an underlying cause as an API choice. A caller-supplied reader error and a private storage implementation error can require different treatment. `%v` prevents unwrapping but retains the text; it cannot by itself make a response safe to expose. Translate sensitive diagnostics at the actual response boundary without prescribing a logging framework. Check the module's supported Go versions before using newer inspection or aggregation APIs. These decisions follow [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) and the [errors package documentation](https://pkg.go.dev/errors).

Ensure a successful `error` result is a nil interface; return literal nil when converting a typed error pointer would otherwise create a non-nil interface. Make partial-success behavior explicit, including bytes returned with a read error and accepted work before failure. Inspect iteration, flush, and close errors when they affect the promised outcome. Retain all required independent failures when aggregating; avoid losing a primary failure or converting failed completion into success. Decide handling and reporting from actual ownership rather than a universal log-or-return prohibition. Recoverable operational failures normally return errors; justified package-local panic/recover or invariant failures need their actual contract considered.

### Values and zero values

Trace supported creation paths and the difference between absent, explicitly zero, and configured values. A useful zero value is desirable when it fits the invariants; mandatory inputs may justify a constructor. Preserve existing supported zero behavior under the release policy without inventing zero usability for every type.

Choose nil or non-nil empty collections from the observable contract. Distinguish nil-map reads from writes, and initialization required for mutation from slice initialization that changes a wire result. Avoid unsupported heap-allocation claims. Explain when a returned or accepted slice/map is borrowed, shared, shallowly copied, or independently owned. Capacity restriction does not isolate existing slice elements; a shallow clone does not isolate nested reference values. Copies should satisfy the actual ownership contract rather than always deep-copying inputs.

Choose receivers from mutation, method sets, interface satisfaction, and copy-sensitive state, not a fixed size threshold. Do not copy used synchronization state. Cover typed nils at interface conversion, with error success behavior owned by the error skill and representation mechanics owned here. Use the [Go specification](https://go.dev/ref/spec), [nil-error FAQ](https://go.dev/doc/faq#nil_error), [code review comments](https://go.dev/wiki/CodeReviewComments), and relevant [sync contracts](https://pkg.go.dev/sync) to resolve disputed behavior.

### Names and documentation

Evaluate names where readers use them: qualified package symbols, related operations, local scope, and the project's domain vocabulary. Apply conventional casing, initialisms, receiver names, and error naming where they improve clarity. Avoid universal boolean prefixes, mutation suffixes, package singularization, forced `Get` removal, or changing a valid enum zero value for naming consistency. A protocol's method name or an established public identifier may need to remain.

Write comments that explain the operation and the constraints readers cannot safely infer: units, mutation/ownership, zero/default behavior, error meanings, partial results, and applicable concurrency guarantees. Preserve the implementation's actual guarantees and documented obligations. Do not invent safety, ordering, release promises, or rationale. Use Go doc comment conventions and verify important examples through real compilation/execution when feasible. Match documentation scope to the request; no automatic README rewrite, documentation website, badge set, or `llms.txt`. See [Go Doc Comments](https://go.dev/doc/comment) and [Package names](https://go.dev/blog/package-names).

## Source reuse and consistency

Retain the upstream pin `samber/cc-skills-golang@19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`. Initial exploration inspected the main skills for `golang-error-handling`, `golang-code-style`, `golang-structs-interfaces`, `golang-naming`, and `golang-documentation`; this is not a completed section/reference audit or approval to copy their content.

Before authoring each runtime draft, inventory its relevant upstream references and record copy/adapt/omit decisions with local owners in a new group audit. Revisit the earlier [architecture source audit](../../go-quality-build/source-audit.md) without rewriting its historical decisions. Adapt useful examples and conditional guidance. Omit mandatory `%w`, nil-collection bans, one-size receiver/enum rules, required third-party packages, auto-configuration, harness orchestration, and unrelated documentation requirements. Retain the complete upstream MIT notice in distributed material if copying substantial prose or examples.

Check each retained decision against the existing [idiom decisions](../../../plugins/go-quality-review/skills/go-code-quality-and-idioms/references/idiom-decisions.md) and [correctness decisions](../../../plugins/go-quality-review/skills/go-correctness-and-compatibility/references/correctness-compatibility-decisions.md). The reviewers remain independent and unchanged. New skills must work alone without requiring the testing group's unpublished skills.

## Evaluation design

Create runnable inputs and private expectations before writing a candidate skill. Authors receive only the task, contract, and input files, with no assertion ledger, expected grade, other arm's output, or controller repair. Use fresh contexts for matched baseline and skill-on work, with the same author model, effort, fixture hashes, tool access, and fixed existing-skill exposure. Record model, harness, skill bytes/revision, opened skills/references, prompt, patch, output hashes, actual checks, and limits. Explicit skill exposure establishes selection under that exposure, not automatic harness routing.

| Candidate | Required representative tasks and controls |
| --- | --- |
| Errors | Library reading caller-supplied input with inspectable errors and partial data; private backend translated to stable domain failure; CLI failure with truthful status and accepted-prefix semantics; primary operation plus completion failure; private calculation control without changed error handling. Include a successful typed-error return path and a case where exposing a cause is correct. |
| Values | Library zero value plus explicit zero configuration; wire output with intentional nil/empty distinctions; snapshot/input ownership with nested reference values; receiver/method-set compatibility and copy-sensitive state; constructor-required type that should keep its invariant; calculation control without representation changes. |
| Names/docs | New public library operation with non-obvious units/ownership/error behavior; internal rename improving a real use site; existing exported/protocol name that must remain compatible; service or CLI documentation matching observed behavior; docs-only correction that must preserve code; mechanical formatting control. |

Judge actual behavior, public consumer probes, code clarity, and documentation accuracy. Hidden checks should detect plausible relevant regressions without requiring a particular implementation. Documentation quality needs an evidence-backed manual comparison as well as compiling examples; it cannot be reduced to comment counts or preferred identifier spellings. Check each completed task with relevant Go builds/tests/vet and the appropriate existing Code Quality, Correctness, Architecture, or Security reviewer. Report unavailable minimum-toolchain, platform, or integration checks as limits.

Observe a confirmed baseline weakness before authoring guidance to fix it. If the baseline is already strong, use a materially different task rather than manufacturing a grade failure. If useful distinct benefit remains unproven, defer or merge the candidate. Preserve first-pass results separately from later review-guided repairs. Rerun affected cases after guidance changes, and run applicable positive, counterexample, and non-selection cases against the final bytes before promotion.

After individual gates, run one fresh combined library/CLI evolution task covering error exposure, snapshot/default state, and accurate public usage. Compare a fixed existing-skill arm with the same arm plus the new group. Check interactions with API contracts and composition, unnecessary code/dependencies, and contradictions at each owner. Testing skills may be included only through a separately recorded, fixed revision present identically in both arms; they are not a dependency of this group.

## Promotion, coordination, and next action

A promoted skill has a distinct trigger, checked links/examples, realistic final-revision evidence of benefit, no observed harm in the relevant controls, and independent review supporting the bounded claim. Run the repository validator, applicable existing tests when tooling changes, and available harness validators. Structural validity and runtime-loading verification remain separate claims.

The testing chat works in a different managed worktree. Reserve its two decision owners and avoid writing their runtime files or evaluations here. Shared README/catalog/design-record updates may need reconciliation when branches meet; preserve both dated work logs and test each catalog after integration. Do not message the other chat without user authorization.

This spec is approved for sequential baseline/authoring/evaluation work under the implementation plan. Promotion remains contingent on evidence. No new runtime skill, evaluation fixture, package metadata, or installation was included in the design stage.
