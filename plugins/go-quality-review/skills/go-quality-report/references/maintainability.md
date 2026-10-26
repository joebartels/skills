# Maintainability profile

Assess how safely people can understand and change the reviewed code. This profile accompanies the nine-topic defect grade; it is not a tenth grade, a weighted score, or another route to A+. For a maintainability-only request, produce the profile without silently repeating the nine-topic audit. State any prior risk grade's revision and date instead of treating it as freshly verified.

## Scope and evidence

Freeze the revision and any working patch. Inventory production, generated, test and test-support code separately. State included dependencies, build constraints, material exclusions and sampling. Select representative workflows and likely changes from actual responsibilities, including a difficult failure path; do not pick only the cleanest examples. For broad trees, delegate bounded domains and assign someone the cross-domain synthesis. Do not award a whole-codebase rating from a few clean functions.

Use three evidence labels:

- **Measured:** an executed command, repeatable source metric or observed change/test outcome; record tool/version, denominator and exclusions.
- **Source-traced:** inspected implementations, callers, contracts and assertions, with paths and symbols/lines. A proposed edit set is an estimate, not a completed change or measured effort.
- **Reported:** supplied/historical results (including claims of a source trace without inspectable paths/contracts); record their snapshot and whether still applicable. Do not claim fresh execution or infer human comprehension time from an agent's wall-clock duration.

Trace at least two distinct representative changes for a codebase comparison when feasible. Include entry points, rule owners, production/config/schema edits, affected consumers and regression checks. Compare equivalent requirements; generic OIDC configuration is not equivalent to supporting an arbitrary proprietary provider. A missing required capability is a scope mismatch, not automatically poor extensibility. Actual change experiments belong in disposable copies and require no publication. Mark unexecuted work explicitly.

Assess operation/state fit through the [architecture decision guide](../../go-architecture-and-design/references/design-decisions.md#fit-to-supported-behavior), including procedural designs and the cost of proposed abstractions. Use the existing rows below; one cause is not a separate finding for each affected dimension.

## Dimensions and evidence anchors

Use all seven rows, or explain irrelevance. Ratings are **Weak**, **Adequate**, **Strong**, **Not demonstrated**. No numeric conversion or average.

| Dimension | Evidence to inspect | Adequate | Strong requires more than absence of defects |
| --- | --- | --- | --- |
| Clarity | Trace normal and failure flows; names, state mutation, error ownership, rationale and navigation needed | Important flows can be reconstructed, with documented local friction | Independent reconstruction of representative flows/invariants agrees with source and tests, including a difficult path; no hidden cross-file assumptions in those samples |
| Change locality | Edit sets for realistic extensions/policy changes; independent decisions and affected consumers | Source traces show bounded edits and explicit compatibility obligations | At least two distinct implemented or historical changes demonstrate bounded edits and preserved contracts; record regressions/validation, not just file counts |
| Cohesion and boundaries | Responsibilities, dependency directions, API leakage, reasons components change | Responsibilities and ownership are identifiable, with limited supported coupling costs | Representative cross-boundary changes and consumers demonstrate isolation; boundaries constrain dependencies without coordination across unrelated concerns |
| Abstraction cost | Fit of operations/state to supported features; caller knowledge, wrappers, configuration and transformations | Features use identifiable rule/operation owners; caller and implementation costs are understandable | Representative features and supported changes show less caller reconstruction or localized variation without hiding invariants/lifecycle; compare a simpler alternative including its costs |
| Rule ownership | Map representative business rules to their authoritative implementations and propagation | Owners can be named; necessary cross-layer checks have distinct responsibilities | Independent traces/tests demonstrate that changes to multiple important rules do not require synchronized reimplementation; distinguish defense in depth from duplicate policy |
| Testability | Assertion quality, setup burden, dependency isolation, flakiness, diagnostics and change sensitivity | Relevant behavior can be verified with understood fixtures and dependencies | Executable failure/negative controls or meaningful mutations prove detection in representative workflows; tests remain focused through exercised changes |
| Complexity hotspots | Function-size/branch/nesting distributions; concrete outliers, callers/tests, change history if available | Significant outliers are inspected and their costs explained | High-risk/high-change outliers have evidence of controlled complexity and effective tests or successful localized evolution; small averages alone do not qualify |

**Weak** means a demonstrated, material obstacle: name the concrete maintenance consequence and evidence, not merely a preferred style violation. **Not demonstrated** means evidence is too thin or an outcome-changing uncertainty remains; it is not a failing score. Where observations are mixed, say which areas are adequate/weak and why the selected scope-level rating is justified. Keep local findings visible even if the broader rating is adequate. A rating's scope must not exceed its evidence.

Each row also gets **High / Medium / Low confidence**, independently of the rating. High needs representative corroborated evidence and no material unresolved assumption; Medium commonly means source tracing with limited dynamic/history evidence; Low signals narrow or unverified evidence, normally requiring Not demonstrated. An entire unsupported row cannot become Adequate by assigning Low confidence.

## Measurement guardrails

Use metrics to find questions, not answer them automatically. Define physical versus logical lines, named functions versus closures, complexity formula, percentile method and analyzed file set. Separate service-owned and reachable shared code where relevant. List concrete outliers; do not hide them behind an average. Generated code, tests, stubs and build variants must not silently distort production comparisons.

Do not use universal function-length/complexity cutoffs, interface counts, package counts, lint counts, comment density or test coverage percentages as cleanliness grades. Go error guards can increase branch counts while improving clarity; extracting tiny helpers can reduce a function metric while increasing navigation. Go already forbids import cycles, so zero package-import cycles is not evidence of architectural excellence. Text duplication is a lead, not proof of duplicate policy. Distinguish intended defensive checks from competing authorities.

Change-history evidence needs a time window and filtering of merges, formatting, generated files and bulk moves. No history means unknown, not stable. Do not divide defects by lines to erase critical risk, rank unequal feature sets by raw totals, reward speculative extensibility, or infer performance from source complexity. Prefer within-codebase trends and matched workflows.

## Synthesis and output

Retain a profile artifact with snapshot, scope, exclusions, representative workflows, change scenarios, measurements, row evidence, counterevidence, rating, confidence and the next validation that would change the judgment. Reconcile cross-domain disagreements against evidence; neither average ratings nor select the most alarming worker label automatically.

| Dimension | Rating | Confidence | Evidence / counterevidence | Scope / next validation |
| --- | --- | --- | --- | --- |
| [all seven dimensions] | [anchor above] | [High/Medium/Low] | [labeled evidence and references] | [limits and concrete next step] |

Conclude separately on **cleanliness**, **structure** and **extensibility**, tying each statement to the rows and matched scenarios. Name improvement priorities with a concrete benefit; do not prescribe speculative abstractions. For comparisons, show both profiles and explain where scope prevents a winner. Do not declare broad excellence from two safeguards: a claim of exceptional maintainability needs representative positive evidence across applicable dimensions and exercised change scenarios, with no material weak or unassessed dimension hidden.

Maintainability observations enter the defect ledger only when they independently satisfy an existing topic's actionable-consequence and severity rules. Deduplicate those causes normally; profile ratings themselves never change grade.py arithmetic. An unassessed profile row does not retroactively invalidate a complete risk review unless it reveals a material risk obligation was unassessed. Keep that distinction explicit.
