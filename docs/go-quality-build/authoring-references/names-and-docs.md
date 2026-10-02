# Names and documentation in Go changes

Status: reviewed authoring reference proposed for reuse; not a standalone skill or a measured effectiveness result. Names and documentation usually express decisions owned by the API, error, value and composition skills. Use this reference for a requested documentation/naming change or a concrete ambiguity that affects correct use.

Judge a name at its use site: package-qualified symbols, related operations and the project's domain vocabulary. Prefer a concise name that communicates the actual operation. Conventional initialisms and Go casing help consistency, but do not justify changing a public identifier or protocol method. Avoid automatic boolean prefixes, mutation suffixes, forced Get removal or enum-zero changes. A local predicate name should clarify its domain meaning without changing lazy evaluation or side effects.

For changed documentation, keep a small factual ledger:

| Reader question | Evidence to check |
| --- | --- |
| What units, defaults and valid inputs apply? | Implementation, schema/configuration and supported creation paths. |
| Who may mutate or retain this value? | Actual aliasing, copy behavior and resource lifetime. |
| What is usable after failure? | Returned prefix/count, classification and completion behavior. |
| What safety or ordering is promised? | Existing contract and implementation evidence, including synchronization. |
| How is it used and migrated? | Compiled usage and verified old/new behavior, not a plausible invented example. |

Explain constraints readers cannot safely infer. Preserve modality: “may” does not become “always,” and a successful local test does not prove concurrency safety or durability. Remove unsupported guarantees rather than inventing rationale. Compile important examples; run examples with output assertions when observable output matters. Compilation alone cannot validate a claim about ownership, ordering or safety.

Match scope to the request. A docs-only correction should preserve code behavior; a mechanical formatting task needs no naming cleanup. No standard README quota, website, badge set or generated machine-doc artifact is required. Follow [Go Doc Comments](https://go.dev/doc/comment) for rendering and [Package names](https://go.dev/blog/package-names) for qualified use-site guidance.
