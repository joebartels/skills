# Pinned language-contracts source audit

Audited 2026-10-01 against `samber/cc-skills-golang@19a0626ae8565d27a7b7bdf59d8d99d94d7e284c`. This is a reuse decision record for the approved language-contracts effort, not runtime guidance, an implementation, or evidence of effectiveness. It supplements the [architecture audit](source-audit.md) without changing its historical decisions and follows the [group spec](../superpowers/specs/2026-10-01-go-quality-build-language-contracts-design.md) and [independent design assessment](language-contracts-design-review.md).

The controller supplied the pinned corpus under `/private/tmp/go-language-upstream-19a0626/skills/`, attesting that it fetched the files through GitHub at the stated commit. I independently read the complete bodies, including reference snippets, and calculated the SHA-256 identities below. I did not independently re-fetch the corpus from GitHub. The audit covers exactly 23 main/reference Markdown files in six directories. It inventories every second-level heading outside fenced examples plus one preamble row per file: **139 sections and 23 preambles, 162 decisions**. Third-level topics and snippets are included in their containing section's decision. Fenced README/configuration template headings are not source-document sections.

**copy** means retain source text/code substantially unchanged; **adapt** means retain a useful decision or concept with the stated corrections and original local prose/examples; **omit** means do not import it into this group. **No copy decision is selected.** No upstream source, example, runtime skill, fixture, plan or package configuration was copied/edited by this audit. Adaptation is contingent on behavioral need, not an instruction to include every retained concept.

## Local owners and comparison sources

EC = proposed `go-error-contracts`; VC = proposed `go-values-and-zero-values`; ND = deferred/reference-first `go-names-and-docs`; AC = existing API contracts; IC = existing interfaces/composition; PB = existing package boundaries. Testing is the separately assigned behavior-tests/test-isolation work. Telemetry, Performance, context/concurrency, Build and Release denote reserved neighboring decisions. An owner in an omit row identifies where any future relevant work belongs. A section with multiple topics states which decision is retained and which remains adjacent.

| Basis | Local decision reference | Primary comparison source |
| --- | --- | --- |
| I | [Idiom decisions](../../plugins/go-quality-review/skills/go-code-quality-and-idioms/references/idiom-decisions.md) | [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) |
| C | [Correctness/compatibility](../../plugins/go-quality-review/skills/go-correctness-and-compatibility/references/correctness-compatibility-decisions.md) | [Go specification](https://go.dev/ref/spec) |
| A | [Architecture decisions](../../plugins/go-quality-review/skills/go-architecture-and-design/references/design-decisions.md) | [Effective Go interfaces](https://go.dev/doc/effective_go#interfaces_and_types), [module layout](https://go.dev/doc/modules/layout) |
| T | [Testing decisions](../../plugins/go-quality-review/skills/go-testing/references/testing-decisions.md) | [testing examples](https://pkg.go.dev/testing#hdr-Examples) |
| S | [Security decisions](../../plugins/go-quality-review/skills/go-security/references/security-decisions.md) | [Go error exposure guidance](https://go.dev/blog/go1.13-errors) |
| P | [Performance decisions](../../plugins/go-quality-review/skills/go-performance-and-resource-management/references/performance-decisions.md) | [sync contracts](https://pkg.go.dev/sync), [strings.Builder](https://pkg.go.dev/strings#Builder) |
| R | [Dependency/toolchain decisions](../../plugins/go-quality-review/skills/go-dependencies-and-reproducibility/references/dependency-reproducibility-decisions.md) | [Go Toolchains](https://go.dev/doc/toolchain), [loop semantics](https://go.dev/blog/loopvar-preview) |
| E | I/C/A above | [Working with errors](https://go.dev/blog/go1.13-errors), [errors](https://pkg.go.dev/errors), [io.Reader](https://pkg.go.dev/io#Reader), [Effective Go errors](https://go.dev/doc/effective_go#errors) |
| V | I/C/P above | [Specification](https://go.dev/ref/spec), [nil-error FAQ](https://go.dev/doc/faq#nil_error), [slices.Clone](https://pkg.go.dev/slices#Clone), [maps.Clone](https://pkg.go.dev/maps#Clone), [encoding/json](https://pkg.go.dev/encoding/json#Marshal), [copylock implementation](https://go.dev/src/cmd/vendor/golang.org/x/tools/go/analysis/passes/copylock/copylock.go) |
| N | I/C above | [Go Doc Comments](https://go.dev/doc/comment), [Package names](https://go.dev/blog/package-names) |

The row basis points to the local/primary comparisons above. Source links below are immutable commit URLs; line numbers locate the audited section. General upstream metadata, persona, priority, modes, broad Go globs and tool permissions are omitted in every preamble. References' introductory upstream crosslinks are likewise omitted. This group supplies its own trigger, scope, ownership and portable reference links.

## Section inventory

### [golang-code-style/SKILL.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-code-style/SKILL.md)

Source identity: 241 lines; SHA-256 `1c53d6b8405fc2992b92524d4782c826845916d64f90a5af483b9401865b9a76`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Line Length & Breaking (L34) | adapt | ND | Keep semantic readability; omit 120-column and four-argument absolutes, prompt override and automatic options migration. | I/N |
| Variable Declarations (L54) | adapt | VC | Distinguish zero initialization, required nil-map write initialization and observable nil/empty output; keyed literals fit the actual type/contract. | I/C/V |
| Control Flow (L89) | adapt | EC | Use readable error flow; remove operand-count and early-return absolutes, preserve short-circuit behavior and partial work. | I/C |
| Function Design (L171) | omit | IC/context | General API/function design stays with composition/API/context owners; no four-parameter ceiling or mandatory newest loop syntax. | A/R |
| Value vs Pointer Arguments (L193) | adapt | VC | Choose mutation/absence/copy semantics; remove 128-byte threshold and unsupported stack/cache claims. | I/P/V |
| Code Organization Within Files (L197) | omit | PB/AC | File taxonomy/import organization and public-surface restructuring are not this group; no automatic unexport/rename tool dependency. | A/C |
| String Handling (L206) | adapt | EC | Useful quoting may clarify diagnostics; omit universal faster-conversion/Builder prescriptions without workload evidence. | I/P |
| Type Conversions (L210) | adapt | VC | Check conversion semantics when changed; generic Contains is an option, not a mandate to genericize valid any APIs. | I/C/V |
| Philosophy (L218) | omit | IC/Build | No samber/lo requirement, reflection ban or dependency change justified by a slogan. | A/R |
| Parallelizing Code Style Reviews (L226) | omit | none | Harness fan-out is omitted. | — |
| Enforce with Linters (L230) | omit | none | Honor configured tooling; no prescribed additional linter suite or project auto-configuration. | I |
| Cross-References (L234) | omit | none | Replace unsupported external skill identifiers with only required portable references. | — |

### [golang-code-style/references/details.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-code-style/references/details.md)

Source identity: 75 lines; SHA-256 `3969d18fb54e59c032de0463b44b6b5bd125dfe8b726ef0f009190da81f3df07`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Extract Complex Conditions (L3) | adapt | ND | Name predicates when that clarifies domain use; preserve laziness/side effects, without operand-count rules. | I/C |
| Value vs Pointer Arguments (L37) | adapt | VC | Retain optional nil and mutation decisions; drop 128-byte boundary and guaranteed pointer/value performance outcomes. | I/P/V |

### [golang-documentation/SKILL.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-documentation/SKILL.md)

Source identity: 245 lines; SHA-256 `a313e2a0037731074ae7e12a14727e42396ea397d84f53873400d3fc5c44fb2d`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Cross-References (L37) | omit | none | External humanizer and upstream skill identifiers are not portable runtime dependencies. | — |
| Writing Principles (L44) | adapt | ND | Retain factual concise prose and modality preservation; explain observable what as well as why/when. | I/N |
| Step 1: Detect Project Type (L60) | adapt | ND | Choose reader/use surface from actual consumers; main/cmd presence does not exclude a library or require broad docs. | A/N |
| Step 2: Documentation Checklist (L78) | omit | ND | No every-project README/license/examples/website/llms checklist or automatic scope expansion. | I/N |
| Parallelizing Documentation Work (L100) | omit | none | Harness fan-out and per-file generation are outside this authoring group. | — |
| Step 3: Function & Method Doc Comments (L108) | adapt | ND | State actual units/results/constraints; omit universal Parameters/Play/example template and fictitious URLs/guarantees. | I/C/N |
| Step 4: README Structure (L150) | omit | ND | No fixed README order, badges or asset-template requirement. | — |
| Step 5: CONTRIBUTING & Changelog (L172) | adapt | ND | Accurately describe requested setup/release behavior; omit ten-minute/tooling mandates and automatic new artifacts. | I/C |
| Step 6: Library-Specific Documentation (L178) | adapt | ND | Runnable usage where it explains a meaningful contract; omit mandatory demos/sites/registration or generous examples quota. | I/T/N |
| Step 7: Application-Specific Documentation (L191) | adapt | ND | Match supported help/config behavior; do not require a CLI framework or full docs rewrite. | I/C |
| Step 8: API Documentation (L201) | adapt | ND | Existing schema/protobuf/docs can be contract evidence; do not install generators or prescribe API-format migration. | A/C/R |
| Step 9: AI-Friendly Documentation (L213) | omit | ND | No unrequested llms.txt, machine-doc artifact or external registration. | — |
| Step 10: Delivery Documentation (L222) | omit | Release | Distribution methods/download commands and Docker/Homebrew configuration are outside this group. | R |

### [golang-documentation/references/application.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-documentation/references/application.md)

Source identity: 221 lines; SHA-256 `e9b8ddbfb0aef0090fdd29f41a020b829738a6cb504aa4276fcddff13f4767a5`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L5) | omit | none | Navigation only. | — |
| CLI Help Text (L16) | adapt | ND | Match actual flags/grammar/status/examples; no required Cobra or generic comprehensive-help rewrite. | C/N |
| Configuration Documentation (L48) | adapt | ND | Document observed precedence/defaults/required inputs; avoid copying illustrative env values and configuration order as facts. | C/N |
| Architecture & design decisions (L89) | omit | PB/IC | ADRs and architecture restructuring require separate scope; no mandated directory/template. | A |
| API Documentation (L123) | adapt | ND | Explain existing wire/RPC/event contract accurately; omit generator installs, tool dependencies and schema-format mandates. | A/C/R |

### [golang-documentation/references/code-comments.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-documentation/references/code-comments.md)

Source identity: 349 lines; SHA-256 `b30a972c758fe5f75615c91024cd2189a12e2d68f61128026c4e4809fc1679f4`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L5) | omit | none | Navigation only. | — |
| Function & Method Doc Comments (L25) | adapt | ND | Keep meaningful operation/constraint docs and truthful deprecation; omit fixed template, every-error disclosure, Play URLs and assumed safety promises. | I/C/E/N |
| File & Package Comments (L228) | adapt | ND | Keep package/comment rendering conventions; no file-count/line-count/200-line quota, invented architecture/security guarantees or automatic diagrams. | I/C/N |

### [golang-documentation/references/library.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-documentation/references/library.md)

Source identity: 210 lines; SHA-256 `3cb75b20e9965d9f0c0e3a2a1f17d37d7d2d0278ac704dc4b15f9eed1620dd1c`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L5) | omit | none | Navigation only. | — |
| Public vs Private Libraries (L18) | adapt | ND | Audience changes exposure/readers; no universal artifact quotas, especially external publication of private library material. | I/S/N |
| Go Playground Demos (L41) | omit | ND | Playground publishing and public URLs are optional requested work, not a required library deliverable. | — |
| Example Test Functions (L71) | adapt | ND | Use compilable examples; correct missing strings import in ExampleMap_strings; Output/Unordered output distinguishes execution from compile-only checks. | T/N |
| Code Examples in Doc Comments (L114) | adapt | ND | One useful verified example may suffice; no option/retry/authentication API invention to fill a generous-examples template. | I/C/N |
| godoc and pkg.go.dev (L143) | adapt | ND | Preview actual documentation and symbol links; omit installing latest pkgsite/tool directive and unsupported publication claims. | I/R/N |
| Documentation Website (L174) | omit | ND | Website frameworks, deployment, llms.txt and registries are optional projects, not names/docs requirements. | — |

### [golang-documentation/references/project-docs.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-documentation/references/project-docs.md)

Source identity: 128 lines; SHA-256 `4eb0ec8988743d95b0d6c701a6070e74c54c40a9d775c074577032764c125199`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L5) | omit | none | Navigation only. | — |
| README.md (L18) | adapt | ND | Improve requested reader facts within established structure; omit universal license creation, exact section order and badges. | I/C |
| CONTRIBUTING.md (L39) | omit | Build/Testing | No ten-minute requirement or automatic Makefile/compose/devcontainer/integration-tag configuration. | T/R |
| Changelog (L59) | adapt | ND | State verified user-visible changes and migration when requested; no every-release file/template/automation requirement. | C |
| Distribution (L102) | omit | Release | No multiple-install-path mandate, CGO disabling, pinned image prescription or automatic distribution work. | R |

### [golang-error-handling/SKILL.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-error-handling/SKILL.md)

Source identity: 91 lines; SHA-256 `536c25e624c3cf3def98962d9f89db07b258cde6c9e7fd17009a9eb204ba631f`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Best Practices Summary (L38) | adapt | EC | Replace mandatory wrapping, log-or-return and exported classifications with producer/caller/boundary decisions; preserve partial results. | I/C/E |
| Detailed Reference (L56) | omit | none | Upstream navigation is replaced by small portable local references; misleading wrapping/single-handling summaries are not retained. | I/E |
| Parallelizing Error Handling Audits (L64) | omit | none | Audit fan-out, background agents and harness commands are outside a focused authoring skill. | A |
| Cross-References (L74) | omit | none | Unpackaged upstream skill identifiers and CI orchestration are not portable dependencies. | A |
| References (L82) | omit | Telemetry | Logger/handler/vendor choices are optional telemetry recipes, not error contracts. | I/S |

### [golang-error-handling/references/error-creation.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-error-handling/references/error-creation.md)

Source identity: 157 lines; SHA-256 `8a5bda5a543ef7774ddc240cd198d34871222ed0335474df599d162f09d253c2`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only; regenerate any needed local contents. | — |
| Errors as Values (L15) | adapt | EC | Handle failures deliberately; replace always-check-before-other-results rule with the actual partial-result contract. | I/C/E |
| Error String Conventions (L42) | adapt | EC | Use concise composable messages; preserve acronyms/proper nouns and existing text contracts, rather than all-lowercase policing. | I/N |
| Creating Errors (L62) | adapt | EC | Choose existing sentinel, plain error or type from caller needs; omit oops dependency and performance-first sentinel rationale. | I/E |
| Low-Cardinality Error Messages (L94) | omit | Telemetry | Grouping and structured attributes belong at observed telemetry boundaries; dynamic diagnostic context remains valid. | I/S |
| Custom Error Types (L117) | adapt | EC | Retain structured inspection and intentional Unwrap; avoid exposing raw SQL/query diagnostics without a boundary decision. | I/S/E |

### [golang-error-handling/references/error-handling.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-error-handling/references/error-handling.md)

Source identity: 138 lines; SHA-256 `ab9c510305a1fe043502a04fe935f16dac2f95c585820ae9f0568afb2a56d357`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| The Single Handling Rule (L12) | adapt | EC | Choose responsibility for handling/reporting; avoid repeated logs without banning intentional local event plus returned signal. | I/A |
| Panic and Recover (L50) | adapt | EC | Ordinary failures return errors; preserve justified contained panic/recover; omit indiscriminate goroutine-boundary recovery recipe. | I/E |
| Why Use `samber/oops` (L99) | omit | Telemetry | No production-error vendor requirement or automatic stack/context dependency. | I/S |
| Logging Errors with `slog` (L136) | omit | Telemetry | Logger setup is adjacent ownership; local error skill needs no upstream observability skill. | I/S |

### [golang-error-handling/references/error-wrapping.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-error-handling/references/error-wrapping.md)

Source identity: 124 lines; SHA-256 `6293caf2c41eef243dffbc2dbd17d045770f9a04e8ce910c4452b3d2b6abce06`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| Error Wrapping with `%w` (L15) | adapt | EC | Choose cause exposure from the consumer, including caller-supplied readers; %v preserves text and does not sanitize responses. | I/S/E |
| Inspecting Errors: `errors.Is` and `errors.As` (L50) | adapt | EC | Inspect promised chains; preserve supported direct-sentinel equality; version-gate AsType and correct ValidationError Msg/Message mismatch. | I/C/E |
| Combining Errors with `errors.Join` (L82) | adapt | EC | Join required independent failures; preserve precedence, direct identity, formatting and old aggregate API shape; closeAll is sequential. | I/C/E |

### [golang-naming/SKILL.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-naming/SKILL.md)

Source identity: 169 lines; SHA-256 `ff6e9b3d1ba65f8e243d872501f44fe8d48f90da8e9441d682e56d3e6d9e2ac7`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Quick Reference (L34) | adapt | ND | Keep useful conventions; remove boolean/In/WithContext/constructor/enum/import rules as universal spelling requirements. | I/N |
| MixedCaps (L63) | adapt | ND | MixedCaps is convention; underscores do not break export mechanism, which depends on the initial Unicode character. | I/V/N |
| Avoid Stuttering (L79) | adapt | ND | Read qualified/local uses; contextual repetition and supported exported/protocol names can legitimately remain. | I/C/N |
| Frequently Missed Conventions (L100) | adapt | ND | Keep readability signals; enum zero semantics belong to VC, error meaning/text to EC; no acronyms-to-lowercase mandate. | I/C/N |
| Detailed Categories (L114) | omit | none | Navigation replaced with focused local references, if this deferred candidate earns one. | — |
| Common Mistakes (L128) | adapt | ND | Avoid cosmetic cleanup, mandatory mutation suffix/boolean prefixes, singular packages and enum renumbering; preserve domain vocabulary. | I/C/N |
| Enforce with Linters (L159) | omit | none | Honor configured tooling without installing lint/rename infrastructure. | I |
| Cross-References (L163) | omit | none | Unsupported upstream skill names and tool guarantees are not retained. | — |

### [golang-naming/references/functions-methods.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-naming/references/functions-methods.md)

Source identity: 124 lines; SHA-256 `859a54edf59b3f72b9176bed12d0173ee4aa8f1f8494e0c6d5e968dbce960d04`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| Functions and Methods (L11) | adapt | ND | Use domain/action semantics at call sites; preserve protocol Get and supported names; named returns may support deferred completion errors. | I/C/N |
| Functional Options Pattern (L117) | omit | IC | Three-option threshold and Options-only vocabulary are not naming necessities; existing composition owns configuration shape. | A |

### [golang-naming/references/identifiers.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-naming/references/identifiers.md)

Source identity: 176 lines; SHA-256 `857684c5e46e9ec66cf46edb03b952b21054491cef4b06a46bb6cf7cf833fdb4`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| Variables (L14) | adapt | ND | Match scope/domain ambiguity; no seven-line threshold or rule forbidding a short u for the same concept elsewhere. | I/N |
| Booleans (L112) | adapt | ND | Boolean intent should be clear; ordinary adjectives and established public methods need no is/has/can prefix. | I/N |
| Receivers (L136) | adapt | ND | Use consistent recognizable receiver identity; one/two letters are often sufficient, not mandatory. | I/N |
| Acronyms and Initialisms (L157) | adapt | ND | Conventional initialisms aid readability; preserve project/protocol/generated spelling and compatibility. | I/C/N |

### [golang-naming/references/packages-files.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-naming/references/packages-files.md)

Source identity: 96 lines; SHA-256 `e507bd5a1b2fd690638de969084c14a9cbd7ca1038e1e51a297120672b27f8c1`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Packages (L3) | adapt | ND | Purpose/use-site clarity retained; placement stays PB; no singular/hyphen mandate, cmd enforcement or module-only internal scope claim. | A/C/N |
| Files (L62) | adapt | ND | Use repository naming and real test/target suffix rules; no mandatory underscore filenames or file reorganization. | I/R |
| Import Aliasing (L77) | adapt | ND | Prefer useful qualified names; aliases can clarify generated/versioned imports as well as resolve collisions. | I/N |

### [golang-naming/references/testing.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-naming/references/testing.md)

Source identity: 37 lines; SHA-256 `be7066959d1eb1da03b38ea93a2e7d6d69d3ed82b0bb5d0600a9e4065a819471`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Test Functions (L3) | omit | Testing | Diagnostic test names stay with separate testing work; no required exact function/underscore taxonomy. | T |
| Table-Driven Tests (L13) | omit | Testing | No lowercase-acronym or input/expected field policy; parallel testing owner decides helpful structure. | T |
| Test Helpers (L35) | omit | Testing | Helper naming/assertion strategy remains with testing; no automatic helper rewrite. | T |

### [golang-naming/references/types-errors.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-naming/references/types-errors.md)

Source identity: 173 lines; SHA-256 `4402338a460d05026beeab82a92258ea0b81145f43cf6460f4363fd9a877a885`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| Interfaces (L17) | adapt | ND | Recognizable protocol names help; method name alone does not force signature/interface, and Len alone is not sort.Interface. | I/A/N |
| Structs (L67) | adapt | ND | Prefer informative type names; domain Data/Object can be meaningful rather than forbidden suffixes. | I/N |
| Constants (L81) | adapt | ND | Role-based constants and conventional case help; preserve valid enum zero/numbering, with zero meaning owned by VC. | I/C/N |
| Errors (L125) | adapt | ND | Conventional Err/Error naming only; EC owns classification/text exposure; no mandatory package-prefix or lowercase acronym rewrite. | I/C/N |

### [golang-safety/SKILL.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-safety/SKILL.md)

Source identity: 282 lines; SHA-256 `a9bce4568ba98ae804505d6b1d6fbe0376fe7afdc9f30051bfc03f898e26d50e`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Best Practices Summary (L28) | adapt | VC | Keep changed nil/aliasing facts; remove copy-all, initialize-all, sync.Once/generics/epsilon mandates and broadened safety scope. | I/C/V |
| Nil Safety (L42) | adapt | VC | Correct table: maps have len but no cap; channels use receive/send, not indexing; nil slice append remains valid. | I/C/V |
| Slice & Map Safety (L93) | adapt | VC | Capacity-limited append isolates a nonempty append result, not existing elements; concurrent read-only maps are valid. | C/V |
| Numeric Safety (L115) | omit | Correctness/context | Numeric overflow/float/division policy is outside these candidates; no epsilon or range-guard requirement without supported inputs. | C |
| Resource Safety (L158) | adapt | EC | Check acquisition/processing/completion results that matter; resource release scope belongs to IC/ownership, without banning all loop defers. | I/C/P |
| Immutability & Defensive Copying (L188) | adapt | VC | Copy only to meet retention/mutation contract; shared/borrowed exported views and writable exported configuration can be intentional. | C/V |
| Initialization Safety (L210) | adapt | VC | Separate configured/default/zero state; omit ignored sql.Open error, unconditional Once and init ban; lifecycle belongs to IC. | I/A/V |
| Enforce with Linters (L244) | omit | none | No new linter configuration or Go 1.25 reflection modernization absent a relevant changed reflection contract. | I/R |
| Common Mistakes (L259) | adapt | VC | Retain proven nil/aliasing mechanics; fix full-slice-copy claim and remove float, map-access, init and nil-channel absolutes. | I/C/V |
| Cross-References (L275) | omit | none | No upstream concurrency/security/debugging/CI skill dependencies are imported. | — |

### [golang-safety/references/nil-safety.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-safety/references/nil-safety.md)

Source identity: 213 lines; SHA-256 `2f0bcabb4184c1b88e36bab8a3f90e57887d3a95a15f66a2123445025c91dec7`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| Nil Pointer Receivers (L19) | adapt | VC | Nil receiver calls can intentionally work; preserve/document actual semantics instead of never-call/check-every-pointer rules. | I/V |
| Nil Function Values (L60) | adapt | VC | Optional callback checks, no-op default or required validated callback follow the invariant; no constructor/no-op mandate. | A/V |
| Nil and Error Comparisons (L98) | adapt | EC | Success must return a nil interface, not necessarily literal-nil syntax; Is(err,nil) cannot normalize a typed nil. | I/E/V |
| Nil in Generic Code (L127) | adapt | VC | Keep real constraint/zero semantics; generic IsZero is valid, pointer-only IsNil solves a different question; do not add reflection reflexively. | C/V |
| Patterns for Nil-Safe APIs (L162) | adapt | VC | Exported Client still permits var/zero literal despite constructor; preserve supported paths; required construction decision remains IC. | A/C/V |

### [golang-safety/references/slice-map-safety.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-safety/references/slice-map-safety.md)

Source identity: 209 lines; SHA-256 `9b048bc51d5153bd554eed7eab6685423b9ad7159462313b0fd6a1294596d86d`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Table of Contents (L3) | omit | none | Navigation only. | — |
| Range Loop Variable Capture (L18) | adapt | VC | Use effective file/module semantics; declaration versus assignment matters, and retaining loop addresses can be intentional. | I/R/V |
| Storing Pointer to Loop Variable (L45) | adapt | VC | Address of loop copy and address of collection element have different ownership/mutation semantics even after Go 1.22. | C/R/V |
| Slice Header vs Backing Array (L67) | adapt | VC | Keep backing-array sharing; cap restriction alone is not isolation, and empty append need not allocate. | C/V |
| Subslice Retains Full Backing Array (L91) | omit | Performance | Retained backing storage is a workload/lifetime concern; no mandatory clone of every small view. | P |
| Standard Library Clone Helpers (Go 1.21+) (L116) | adapt | VC | Clone only when ownership needs it; preserve nilness, shallow element/value references and Go 1.21 floor. | C/R/V |
| Map Iteration Order (L133) | adapt | VC | Map range order is unspecified; sort only for promised order; Go 1.23 iterator helpers are not Go 1.21 APIs. | C/R/V |
| Deleting During Iteration (L151) | adapt | VC | Keep map-delete semantics and slice compaction pitfalls; consider shared aliases, tail retention and supported helper version. | C/R/V |
| Comparing Slices and Maps (L189) | adapt | VC | Collections allow == nil; choose equality semantics, including nil/empty and custom elements; helpers do not define every contract. | I/C/V |

### [golang-structs-interfaces/SKILL.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-structs-interfaces/SKILL.md)

Source identity: 300 lines; SHA-256 `0bc93e1e86c3cf24571205b6fdd773a23df1b34366e7f639a9530d6984117601`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Interface Design Principles (L28) | omit | IC | Existing composition owner already decides concrete/interface/protocol needs; reject consumer-only, 1–3 method and second-implementation absolutes. | A |
| Make the Zero Value Useful (L112) | adapt | VC | Useful zero values are conditional; do not mark every constructor-dependent invariant broken or add unsynchronized lazy init. | I/A/V |
| Avoid `any` / `interface{}` When a Specific Type Will Do (L138) | adapt | VC | Preserve known type information where useful; any and reflection have valid contracts and version-sensitive alternatives. | I/R/V |
| Key Standard Library Interfaces (L150) | adapt | ND | Match established signatures when implementing that protocol; names like Read/Len do not force every type into one interface. | I/A/N |
| Compile-Time Interface Check (L165) | adapt | VC | Assertions can guard an intended method set; existing assignments may already suffice, and public changes need consumer evidence. | A/C/V |
| Type Assertions & Type Switches (L175) | adapt | VC | Use checked assertion when mismatch is supported; invariant assertions need not be rewritten; errors use chain-aware EC decisions. | I/C/V |
| Struct & Interface Embedding (L181) | adapt | VC | Track promoted method sets and zero embedded pointers; API exposure choice remains AC/IC, not an is-a template. | A/C/V |
| Dependency Injection via Interfaces (L223) | omit | IC | Concrete/function dependencies and optional construction remain valid; no interfaces or constructors for all dependencies. | A |
| Struct Field Tags (L243) | adapt | VC | Use encoder-specific tags deliberately; omission/explicit-zero mechanics here, supported wire evolution with AC; no tag-all rule. | C/V |
| Pointer vs Value Receivers (L258) | adapt | VC | Choose mutation, references, lock copying and method sets; remove mandatory uniform receivers and size-only choices. | I/C/V |
| Preventing Struct Copies with `noCopy` (L269) | adapt | VC | Do not copy state whose actual contract forbids it; a channel/pointer alone is not proof, and vet is a limited diagnostic. | I/P/V |
| Cross-References (L277) | omit | none | Upstream skill IDs and gopls guarantees are not local runtime dependencies. | — |
| Common Mistakes (L285) | adapt | VC | Retain changed value/receiver/copy traps only; omit interface-count, universal concrete returns/tags/generics and nil-slice bans. | I/A/V |

### [golang-structs-interfaces/references/struct-fields.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-structs-interfaces/references/struct-fields.md)

Source identity: 71 lines; SHA-256 `fc40bb1979f11cd8f9d20361d524a26a233447fa94b6ec3ccad71a5cebfaec2a`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Struct Field Tags (L3) | adapt | VC | Correct omitempty/omitzero distinctions and explicit-zero remedy; reject tag-all, optional vendor types and missing-backtick runtime claim. | C/V |
| Preventing Struct Copies with `noCopy` (L40) | adapt | VC | Retain no-copy semantics and optional vet marker; correct name-only/all-copies claims, Builder analogy and channel/pointer overgeneralization. | I/P/V |

### [golang-structs-interfaces/references/type-assertions.md](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/skills/golang-structs-interfaces/references/type-assertions.md)

Source identity: 70 lines; SHA-256 `02860e00115d6b1256c4accedea33b6544746bc6b1429c2a5fc8ed735692b48e`.

| Section (source line) | Choice | Local owner | Reason/correction | Basis |
| --- | --- | --- | --- | --- |
| Preamble / frontmatter (L1) | omit | none | Local trigger/scope replace upstream metadata, persona, orchestration and introductory crosslinks. | — |
| Safe Type Assertion (L3) | adapt | VC | Handle legitimate dynamic mismatch; a proven invariant may use single-result assertion. | I/C/V |
| Type Switch (L18) | adapt | VC | Explain ordered matching and nil-interface versus typed-nil cases; broad interface-first matching can be intentional. | C/V |
| Optional Behavior with Type Assertions (L39) | adapt | EC | Optional Flush is useful only when promised completion requires it; ownership/necessity stays with IC. | A/C/E |
| Asserting to an Interface, not a Concrete Type (L62) | omit | IC | Optional consumer capability belongs to existing abstraction owner; concrete assertions can be required for actual type semantics. | A |
| Errors (L66) | adapt | EC | Inspect promised wrapped types with supported-version helpers; replace unsupported crosslink. | I/E |

## Corrections required before authoring

The following are blockers to wholesale reuse, not blockers to proceeding with original local authoring and baseline work:

1. **Failure/result ordering and identity.** `error-creation.md` says to check errors before using any other results, which loses valid bytes from a reader returning data and an error together. Trace the operation's partial-result contract instead. Preserve required direct sentinels as well as chain inspection: introducing a wrapper or `Join` may change equality, formatting and `Unwrap` shape. The [errors documentation](https://pkg.go.dev/errors#Join) defines the aggregate behavior; it is not a drop-in replacement for every existing API.
2. **Exposure is not redaction.** Internal/public location does not alone decide `%w` versus `%v`. Expose the caller's supplied failure when promised, translate private backend classification when needed, and sanitize at the actual untrusted response. Vendor log/error frameworks and universal log-or-return rules are omitted. Package-local panic/recover can be justified; do not add blanket recovery that conceals failure or writes a new HTTP error after partial output.
3. **State/ownership rules must be conditional.** Nil slices and read-only nil maps can be intentional; map writes require initialization. Capacity restriction affects later append capacity, not existing elements. Shallow clone preserves nested references, so copying must match retention/mutation promises rather than exported visibility. The aliasing and typed-nil mechanics belong to VC; successful `error` interface behavior belongs to EC. Synchronization design remains adjacent.
4. **Constructor and copy claims need correction.** The exported `Client` in `nil-safety.md` still permits zero construction; a constructor with defaults does not make it inaccessible. A channel or internal pointer alone does not forbid struct copies. Used synchronization state has its actual no-copy contract. The [copylock analyzer](https://go.dev/src/cmd/vendor/golang.org/x/tools/go/analysis/passes/copylock/copylock.go) uses type/method-set and contained-field checks, not merely a field's name; no diagnostic promises every illegal copy is detected. A nonzero [strings.Builder](https://pkg.go.dev/strings#Builder) must not be copied, but that does not make it the same vet-marker example. No fixed 128-byte threshold or guaranteed allocation/cache outcome is retained.
5. **Serialization snippets are not portable guarantees.** `struct-fields.md` misstates `omitempty` as all-zero omission and offers `omitzero` as an absent/explicit-zero remedy. `omitzero` still omits a zero value; presence needs an appropriate representation. Missing a Go literal's delimiting backtick can be a compile error, while a malformed tag string can compile and be ignored. Check the actual encoder/version, nil/empty and byte-slice behavior, and existing supported wire schema. Retain no `mo.Option`/tag-all requirement.
6. **Naming/doc conventions must preserve contracts.** Keep valid enum zero values and numeric encodings; skip-zero/Unknown is a design choice owned by VC. Acronyms/proper nouns need not be lowercased in error text. Ordinary boolean adjectives, established Get/protocol names, package plurals, supported symbols and meaningful domain suffixes can remain. No `In` mutation suffix, forced constructor name or one/two-letter receiver ceiling is retained. Exportedness follows the language's initial-character rule; underscores are not a compiler/tooling incompatibility by themselves.
7. **Examples and documentation must be factual.** Correct `ValidationError.Message` versus inspection snippet `ve.Msg`, and the missing `strings` import in `ExampleMap_strings`, before any adaptation. Doc templates must not invent concurrency, idempotency, ordering, release-removal dates, installation paths or playground links. “Why, not what” cannot suppress a clear operation/result description. Retain consumer obligations and observable constraints at their owner, without documenting private implementation failures as public guarantees. No mandatory README order/badges, contributor time target, sites, registries, llms.txt, auto-configuration or dependency installation.

## Version floors and semantic checks

| Feature encountered | Minimum/support check | Authoring decision |
| --- | --- | --- |
| `%w`, `errors.Is`, `errors.As` | Go 1.13 | Use only for intended inspected contracts; direct identity can still be a supported promise. |
| `errors.Join` | Go 1.20 | Check inspection/format/precedence compatibility; do not hide primary failure or force a new aggregate API. |
| `errors.AsType` | Go 1.26 | Go 1.22 tasks need supported `As` or other existing contract-compatible inspection, unless support is intentionally changed. |
| `any`, type parameters | Go 1.18 | No generic migration merely for style; reflect/heterogeneous APIs can be valid. |
| `slices` / `maps` clone/equality helpers | Go 1.21 | Preserve shallow/nil semantics and intended equality; manual alternatives may support older versions. |
| Per-iteration range variables / integer range | Go 1.22 effective file/module semantics | A new compiler alone is insufficient; assignment to an existing loop variable differs from a declaration. Addressing a loop copy differs from addressing an element. |
| `maps.Keys` iterator and `slices.Sorted` example | Go 1.23 | Do not misclassify under the preceding Go 1.21 helper section; sorting is required only for promised order. |
| `encoding/json` `omitzero` | Go 1.24 | Check actual package/tag behavior; this is omission, not presence tracking. |
| `reflect.TypeAssert` | Go 1.25 | Optional changed-reflection technique; avoid adding reflection or upgrading solely to use it. |
| Doc `# Heading` syntax | Go 1.19 documentation tooling | Render with project tools where relevant; no heading quota or file-count split. |
| `go get -tool` / `go tool pkgsite` recipe | Go 1.24 tool-directive support | Omitted automatic tooling change; use existing preview tools when available. |

Floors were checked against the current [errors](https://pkg.go.dev/errors), [reflect.TypeAssert](https://pkg.go.dev/reflect#TypeAssert), [Go 1.24 release notes](https://go.dev/doc/go1.24), [Go Toolchains](https://go.dev/doc/toolchain), [loop change](https://go.dev/blog/loopvar-preview), [Go Doc Comments](https://go.dev/doc/comment) and standard-library documentation linked above. Their current recommendations do not override a project's declared minimum. No candidate example was compiled, and no minimum-toolchain execution is claimed by this audit.

## Uninventoried assets, license and next action

The 23-file corpus does not contain documentation asset templates, eval JSON, or vendor/tool-specific skills referenced by the main files. They are **omitted as runtime inputs**, not asserted audited: no copied templates, asset lookup, external registry submission, setup files, selected vendor package or upstream acceptance oracle is needed. The completed inventory is the six requested main skills and their 17 Markdown references, not the entire upstream repository.

The [pinned MIT license](https://github.com/samber/cc-skills-golang/blob/19a0626ae8565d27a7b7bdf59d8d99d94d7e284c/LICENSE) and prior [license decision](source-audit.md#non-markdown-inputs-and-license-handling) govern any later substantial copy. No runtime prose/code is copied now, and original local examples are expected. If later authoring changes to a substantial copy, distribute the complete copyright, permission and warranty notice and record the affected paths and source revision; a link to this audit is insufficient.

Next: run the frozen matched error baselines, author only guidance justified by the observed owned failures, then evaluate final bytes and controls before promotion. Keep VC's later baseline focused on ownership/copies with existing API/composition exposure fixed. Names/docs remains conditional on distinct benefit. This audit does not implement, evaluate or promote any proposed skill.

## Corpus identity correction

Root verified the materialized corpus against the fetched GitHub content and pinned Git tree blobs. Initial materialization added one terminal newline per file. On 2026-10-01 root restored all 23 exact fetched byte sequences, verified Git blob SHA-1 identities, and corrected the SHA-256/line identities above. Section decisions and source heading line locations are unchanged. The corrected corpus totals 3,899 lines; prior audit-stage log counts remain historical.

Exact source identity is also preserved in [the upstream corpus manifest](language-contracts-upstream-manifest.json): all 23 fetched Git blob IDs match locally computed blob hashes, with byte counts, SHA-256 and line counts. Independent follow-up verification found no discrepancies in the corrected 3,899-line corpus or the unchanged section/preamble coverage and decision totals.
