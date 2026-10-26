# Go quality review skills

Use **[go-quality-report](../../plugins/go-quality-review/skills/go-quality-report/SKILL.md)** for an overall review of a Go changeset, codebase or code area. It scopes and chunks the target, delegates relevant combinations of the nine topic skills, reconciles shared findings, and grades unique causes using the same rubric. The ten runtime skills live under `plugins/go-quality-review/skills/` and install as one package.

Nine standalone review skills assess a defined changeset or code area across Go libraries, CLIs, and services. Each skill contains its grading rules and decision references; all use the shared structured record and renderer. Other topic names are navigation and ownership guidance. The handbooks in [resources](resources/) are authoring background and are not required at review time. Maintained evaluation inputs and tests live under `tests/go-quality-review/`, outside the installable package. Generate run artifacts outside the checkout.

## Shared reporting and grading

Every skill uses the same F, C-, C, C+, B-, B, B+, A-, A, A+ anchors. Topic and overall reports share one structure: scope, coverage, rationale and confirmed counts; a complete findings summary table; finding details with evidence, consequence, correction and verification; strengths; checks and limits. Overall reports append topic report cards and the maintainability profile. The saved report and final response use the same structure. Not applicable and Insufficient evidence retain the sections without fabricating findings or implying a clean bill of health.

Grade introduced or worsened issues in a changeset; grade existing issues when the user requests a code-area review. Count root causes rather than symptoms. Strengths do not offset defects. A+ requires two independent, verified safeguards addressing distinct meaningful risks; ordinary correct setup supports A.

In a combined report, shared causes use one finding ID and a primary remediation owner. Each topic retains its own supported assessment, including when used alone. Do not sum topic counts or average letter grades to construct an overall result; the umbrella report deduplicates causes and uses their greatest supported consequences with the same rubric. Incomplete material coverage produces an overall Insufficient evidence state, while confirmed findings remain visible.

The shared instruction contract is maintained in [review-contract.json](../../tests/go-quality-review/shared/review-contract.json) and embedded in each skill. It is not a runtime dependency. Update the contract and the embedded copies together when changing shared behavior; topic-specific severity definitions and decision references remain local.

## Validation and behavioral evaluations

From the repository root, with Python 3.10+ and PyYAML available:

```sh
python3 scripts/validate.py
python3 scripts/validate.py --reports /path/to/run
python3 scripts/validate.py --report-file /path/to/report.md
python3 -m unittest discover -s tests/go-quality-review -p 'test*.py'
```

Validation checks harness manifests and catalogs, frontmatter, matching grade/report contracts, local-reference portability, evaluation inputs, and fixture paths. `--report-file` checks a standalone topic or overall report. `--reports` checks the nine topic evaluation suites at `/path/to/run/<skill-name>/<case-id>.md`; umbrella planning scenarios remain separate from complete reports. Report checks cover section order, table/detail correspondence, severity counts, grade arithmetic, and correction/verification fields. They cannot establish that evidence, severity, systemic reach, or A+ safeguards are correct. Instruction contract version 3 routes all skills through review record version 1; the rendered Markdown retains the established report structure.

Each skill's `tests/go-quality-review/<skill>/evals/evals.json` contains behavioral tasks and expected outcomes, not Python tests. For a behavioral run:

1. Give an evaluator the skill, relevant references, one case's prompt, and any listed fixture files. Withhold expected outcomes and assertions. Use a fresh evaluator context per case where feasible; disclose batching when used.
2. Save the actual review, including command evidence and limits. Use disposable copies for reproductions; some fixtures intentionally contain bugs and their tests are expected to fail.
3. Have a separate reviewer check the saved response against the assertions and the underlying evidence. Record unsupported expectations separately from agent errors; do not equate exact wording with quality.
4. Compare to the previous skill version and rerun affected cases after changes. Include false-positive controls, missing context, non-applicability, and all grade boundaries across the suite. Repeat ambiguous cases to examine grade stability; a single run is not proof of reliability.

For deterministic reports, use the installed [record contract](../../plugins/go-quality-review/skills/go-quality-report/references/review-record.md):

```sh
python3 plugins/go-quality-review/skills/go-quality-report/scripts/render.py /path/to/review.json --output /path/to/report.md
python3 plugins/go-quality-review/skills/go-quality-report/scripts/render.py /path/to/review.json --topic go-security
python3 tests/go-quality-review/go-quality-report/evals/test_grade.py
```

The renderer validates the record and derives both topic and overall ledgers for the existing calculator. It does not discover findings, verify source evidence or decide semantic duplication. `test_render.py` exercises projections, malformed records, coverage states, deterministic presentation and installed-package operation; `test_reports.py` independently checks rendered Markdown. Keep runtime validation in `review_record.py`, the complete field reference in `references/review-record.md`, and the rendering implementation in `render.py`; do not restore copied templates to topic skills.

The fixed inventories in [grade-calibration.json](../../tests/go-quality-review/shared/grade-calibration.json) test arithmetic separately from discovery and classification. Withhold expected grades when using them with an evaluator. The shared [source review sequence](../../plugins/go-quality-review/skills/go-quality-report/references/source-review.md) traces behavior before maintenance assessment and evidence reconciliation. Exercise it with the `file-backed-body-consumption` and `file-backed-test-consumers` cases; compare discovery, admission, topic routing and severity separately. Stable letter grades alone do not establish finding agreement.

## Grading topics

### 1. Architecture and Design

- Package and API Boundaries: Do packages, public contracts, and dependencies have clear responsibilities that fit
  their actual callers? Add layers or mappings when they solve demonstrated coupling, not as a required template.
- Interfaces and Abstraction: Does each interface represent a real consumer need or exported protocol, with a coherent
  method set? Would a concrete type or function be simpler where no abstraction is needed?
- Composition and Ownership: Are significant dependencies and resource lifetimes visible? Can callers tell who starts,
  cancels, and closes work without unnecessary constructors or global state?
- Cross-Boundary Behavior: Do errors, cancellation, and asynchronous success claims preserve the intended contract?
  Use context where operations can block or outlive a caller, not on every pure function.

### 2. Code Quality and Go Idioms

- Error Handling: Are errors checked or deliberately handled? Is added context useful, and does `%w` expose only
  causes that belong in the caller's contract?
- Readability and Go Idioms: Are names, control flow, comments, value semantics, and concurrency choices clear at
  their use sites? Does a proposed modernization fit the module's effective Go version and preserve behavior?
- Idiomatic Naming: Does the code follow standard Go naming conventions? (e.g., mixedCaps, short variable names for
  local scope, clear names for exported types, avoiding stutter like user.User).
- Mechanical Evidence: Do formatting, build, vet, or configured linter diagnostics reveal a real issue? A diagnostic
  is evidence to inspect, not an automatic grade; a preferred linter suite is not mandatory.

### 3. Observability and Resilience

- Logs and Correlation: At relevant failure boundaries, are logs useful, appropriately leveled, and correlated with
  requests or traces when the runtime supports that? Are sensitive values omitted or redacted?
- Metrics and Tracing: Can operators detect meaningful failures, latency, and saturation for this workload? Are
  metric labels bounded, and are traces propagated across relevant remote calls without unnecessary spans?
- Time Budgets and Retries: Do blocking calls respect caller cancellation and an appropriate total budget? When
  retries are justified, are they safe, bounded, and coordinated with retries already performed by other layers?
- Failure Containment: Are queues, concurrency, and downstream failures contained where overload is credible? Use
  circuit breakers or idempotency mechanisms when the actual failure mode and contract warrant them.
- Service Lifecycle: Where applicable, do probes reflect restart versus traffic-readiness decisions, and does
  shutdown drain owned work within the environment's grace budget?

### 4. Testing

- Behavior & Assertions: Do tests cover meaningful changed contracts, boundaries, and failure paths? Would a plausible
  regression fail a test? Use coverage to locate gaps, not as a stand-alone quality score.
- Test Structure: Are cases and failures easy to understand? Use table-driven subtests when they reduce repetition;
  focused standalone tests can be clearer for different behavior or setup.
- Doubles & Integration: Do fakes or mocks preserve the relevant dependency contract? Are real boundaries exercised
  where a double would hide an important behavior?
- Isolation & Concurrency: Can tests run alone and in their intended order or parallelism without leaked state or
  guessed timing? Are relevant concurrent paths exercised with race detection where feasible?
- Broader Inputs & Measurement: Where risk warrants them, do fuzz targets assert useful properties and do benchmarks
  measure the intended work accurately?

### 5. Security

- Trust Boundaries and Authorization: Does the changed code identify who controls inputs and enforce access to each
  protected action and object before use? Trace callers, middleware, and alternate entry paths.
- Injection and Unsafe Destinations: Can untrusted values change SQL, command, template, file, redirect, or outbound
  network behavior? Assess the actual sink and existing defenses, including symlinks and redirects where relevant.
- Input and Abuse Resistance: Are attacker-controlled payloads and work bounded for the workload? Do decoded values
  preserve domain and security invariants rather than rely on a particular validation library or JSON setting?
- Credentials and Sensitive Data: Are authentication, token, session, cryptographic, and browser-origin controls suited
  to the protocol? Are secrets and personal data kept out of source, logs, responses, and other unintended outputs?
- Known Vulnerabilities: For changed dependencies or toolchains, is a reported vulnerability present and reachable in
  the relevant build? Use `govulncheck` evidence where useful, while recognizing its coverage limits. CI enforcement
  belongs to Deployment and Operations.

### 6. Performance and Resource Management

- Work and Complexity: Does time and I/O scale with realistic input size, call rate, and concurrency? Are repeated
  scans, copies, round trips, or N+1 operations on a consequential path?
- Memory and GC: Are large values retained unnecessarily? Would measured hot paths benefit from preallocation,
  streaming, or reuse? Are `sync.Pool`, `GOMEMLIMIT`, and layout changes justified by workload evidence?
- Resource Lifetime: Are files, response bodies, rows, timers, goroutines, and queues released or bounded by a clear
  owner, including on early returns and cancellation?
- I/O and Pooling: Are HTTP transports and database pools reused and tuned to actual demand, capacity, and wait
  behavior rather than fixed settings?
- Contention and Measurement: Do profiles, traces, and comparative benchmarks support claimed bottlenecks or
  optimizations? Are lock, atomic, compiler, and PGO changes worth their complexity on the effective Go version?

### 7. Deployment and Operations

- Build and Release Artifacts: Does the selected Go toolchain build the intended targets, and can maintainers identify
  the source, dependencies, toolchain, and flags used to rebuild the shipped artifact? For libraries and CLIs, are
  module tags, compatibility, target platforms, and
  release packages handled where applicable?
- CI and Supply Chain: Do relevant tests, analysis, and vulnerability checks cover the shipped build? Are artifact
  identity, base-image updates, provenance, or SBOM controls present where the project's policy or risk requires them?
- Containerization and Configuration: Does the runtime image meet native-library, certificate, user, and filesystem
  needs without carrying unnecessary build tools or credentials? Are required configuration and secrets delivered and
  validated appropriately? Multi-stage and minimal images are options, not fixed requirements.
- Health and Shutdown: Where the platform uses probes, do liveness, readiness, and any needed startup checks drive the
  right restart or traffic decisions? Can the process stop receiving work and drain what it owns within the grace budget?
- Runtime and Recovery: Do memory, CPU, storage, logs, and diagnostics fit the deployment environment? Can a rollout
  be observed and recovered, including compatibility with old versions or migrations when those are involved?

### 8. Correctness and Compatibility

- Behavioral Contracts: Do supported inputs, boundary cases, and failure paths produce the promised results and
  error meanings? Trace changed behavior through its callers rather than relying on tests or names alone.
- State and Concurrency: Do partial failures, shared values, and concurrent operations preserve promised invariants
  without lost work, misleading success, or unintended mutation?
- Consumer Compatibility: Do supported Go callers, CLI users, service clients, and readers of stored data retain
  their documented behavior across the change? Assess deliberate breaks against the project's release policy and
  migration path.

### 9. Dependencies and Reproducibility

- Module Resolution: Does the committed configuration select the intended direct, transitive, test, and tool modules
  in the actual main-module, workspace, and standalone consumer contexts? Check replacements and version constraints
  against the selected graph rather than treating every indirect change as bad.
- Integrity and Availability: Can a clean checkout obtain the required modules through the configured checksum,
  private proxy, vendor, or local path mechanism? `go.sum` is a hash record, not a lockfile or universal requirement.
- Toolchain and Targets: Do supported Go versions and build targets satisfy the module and dependency minimums,
  build constraints, cgo, and native inputs? Verify what toolchain the build actually selects.
- Generated and Build Inputs: Can the stated source and generator versions reproduce needed generated code? Where
  artifact identity is promised, are flags, VCS state, paths, native toolchains, and packaging inputs controlled?
- Fresh-Checkout Evidence: Does the documented build work without hidden workspace state, warm caches, or uncommitted
  module edits? Separate dependency resolution from CI enforcement and deployment of the resulting artifact.
