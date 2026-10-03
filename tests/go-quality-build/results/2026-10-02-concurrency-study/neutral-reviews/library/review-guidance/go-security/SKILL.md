---
name: go-security
description: Review and grade security in a defined Go changeset or code area across libraries, CLIs, and services. Use for trust boundaries, authorization, injection, unsafe file or network access, cryptography, secrets, sensitive data, abuse resistance, or known vulnerabilities; not for a general Go style, testing, or deployment audit.
metadata:
  review-contract: "1"
---

# Go Security Review

Review whether the specified Go code preserves its security boundaries against the inputs and actors it actually faces. Give one Security grade based on substantiated consequences. Review only; do not edit code unless asked.

## Shared review contract

- Use this skill independently; other topic names identify related expertise, not prerequisites. Grade consequences within this skill's review questions even when another topic also applies. Mention confirmed issues wholly outside this topic separately as ungraded related findings; do not silently discard them or perform an unrelated audit.
- For a changeset, record the base/head revisions or supplied diff and grade introduced or worsened issues. Unchanged code is context unless the change newly exposes its defect. For a code-area review, grade existing issues in that named area. State the interpretation when the target is ambiguous.
- Assess the relevant questions and record coverage and limits. An omitted file in a partial excerpt is not a missing implementation. Distinguish supplied facts, inspected code, executed checks, and assumptions; never invent paths, line numbers, measurements, or command results. Use symbols or quoted snippets when files are unavailable.
- In a combined review, give a shared root cause one finding ID and a primary remediation owner; other topics cross-reference it. Keep each topic's grade faithful to its own assessed consequences, as in solo use. Do not add topic counts into a total: deduplicate shared IDs first. A production defect and a test gap are separate only when they need independent corrections.
- Review checks must preserve source and configuration. Run reproductions or commands that may alter module files, generated files, or build outputs in a disposable copy when needed. Respect the project's build mode and side-effect constraints; report blocked checks as limits.

## Establish the review boundary

1. Identify the exact diff or code area, its library, CLI, or service context, the protected assets, and the trust boundaries it changes. Inspect callers, middleware, configuration, generated sources, tests, and relevant deployment controls as needed to trace an input from origin to sensitive sink. Distinguish introduced issues from relevant existing constraints; do not grade unrelated debt.
2. Check `go.mod`, `go.work`, build tags, target platforms, and the actual built Go toolchain before making version-sensitive claims. A newer installed compiler does not prove the deployed binary or every file uses its APIs. Consult current Go documentation when a security API, release, or vulnerability status matters.
3. Establish attacker control, reachable path, existing defenses, and realistic impact before reporting a vulnerability. A suspicious call or scanner diagnostic is a lead, not proof. If a protection is supplied by a proxy, framework, or caller, verify its coverage before claiming it is absent. Do not assume every CLI argument is hostile or every library function serves HTTP.
4. Report distinct root causes with concrete code or tool evidence. Give positive findings the same specificity. Where reachability, runtime configuration, or a claimed contract cannot be established after reasonable inspection, state the limit; do not turn a possibility into a high-severity finding. Handle active secrets discreetly: identify their location and type without reproducing values.

Read [Security decisions](references/security-decisions.md) when judging dependency reachability, JSON input, file paths, command execution, web origin controls, password or token handling, secret redaction, or containers. It supplies context-sensitive distinctions and links to primary guidance.

## Review questions

| Area | What to establish |
| --- | --- |
| Trust boundaries and authorization | Who controls each input and identity? Is authorization enforced before the protected action, at the right object or tenant scope, on every relevant entry path? Are defaults and error paths fail-closed where the contract requires it? |
| Parsing and resource abuse | Are untrusted bodies, decompression, archive extraction, recursion, allocation, concurrency, and work amplification bounded for the expected workload? Are decoded values validated against domain rules before sensitive use? Unknown JSON fields or duplicate names matter when they change a security decision, not merely because they exist. |
| Sensitive sinks | Can untrusted values change SQL structure, shell behavior, file targets, rendered HTML, outbound destinations, or redirects? Check parameter binding, executable and argument selection, traversal-resistant file access, contextual escaping, and destination checks at the actual sink. |
| Browser and protocol boundaries | Where browser credentials are sent automatically, are state-changing actions protected against CSRF? Check Origin or Fetch Metadata policy, tokens when needed, cookie attributes, and WebSocket handshake authorization and origin policy. For outbound requests, inspect redirects, DNS resolution, and network reachability when SSRF is plausible. |
| Cryptography, credentials, and sessions | Are keys and tokens generated with cryptographic randomness, passwords verified with a suitable password hash, and signatures or MACs verified with approved libraries? Are JWT algorithm, key, issuer, audience, time claims, and token purpose checked as the application contract requires? Are TLS and certificate checks preserved? |
| Secrets and sensitive data | Are credentials absent from source and build artifacts? Are runtime access and rotation appropriate to the threat model? Can logs, traces, errors, URLs, or responses expose secrets or personal data? Check actual emission paths; a redaction method on one type does not protect all serialization. |
| Known vulnerabilities | If dependencies or toolchain changed, or scanning evidence is supplied, inspect affected versions and advisory details. Use `govulncheck` when available and useful, with relevant build configuration. Distinguish module/package presence from reachable vulnerable symbols; account for the tool's false-positive and false-negative limits. Do not claim a clean scan when it was not run. |

Security assesses demonstrated threats and exploitable paths, including vulnerabilities in selected dependencies, sensitive telemetry, and attacker-controlled exhaustion. Dependencies & Reproducibility primarily assesses resolution; Testing assesses regression coverage; Deployment & Operations assesses delivery/runtime enforcement; Observability & Resilience assesses general failure handling. Their overlap does not exclude a relevant security consequence in solo use. Do not require a scanner, validator, secret manager, logger, framework, or container base without a demonstrated need.

## Grade security

Classify each **distinct, substantiated** issue by its effect in this codebase. A minor issue is a localized, credible hardening gap with low impact; a moderate issue permits limited abuse or materially weakens a control under plausible conditions; a major issue exposes an important trust boundary, protected operation, or sensitive data to a realistic attacker; a critical issue permits severe compromise such as broad unauthorized access, remote code execution, or major data loss. Adjust for attacker control, prerequisites, existing defenses, and blast radius. A merge recommendation, if requested, is separate from the grade. Count root causes, not repeated occurrences. Do not lower a grade for optional defense in depth alone when no relevant failure mode is supported.

Apply the first matching row from the top. These anchors match the other Go review skills; explain context-dependent severity and avoid treating a scanner label as the grade.

### Apply the anchors consistently

Count independently actionable root causes after grouping repeated symptoms. Two symptoms are one issue when the same causal correction resolves both; unrelated corrections remain separate even in one function. State the counted severities in the report. Verified strengths do not cancel defects, and optional preferences or unconfirmed concerns do not enter the counts.

A contained failure of one important function, target, or contract is normally major. Use systemic major only when one shared cause demonstrably affects multiple distinct important boundaries or workflows; repetition across files alone is insufficient. Critical requires supported severe, broad, or irreversible impact in the topic being reviewed, not merely the words race, unsafe, missing test, or build failure. Explain any systemic or critical classification. These reach rules refine the topic-specific severity definitions above.

For A+, identify two independent safeguards, the distinct meaningful risks they control, and evidence that each works. Ordinary correct setup supports A. Complexity, tool count, or splitting one safeguard into two descriptions does not earn A+; simple designs can qualify when their risk control is demonstrated. If no issue is found but the evidence cannot support even one relevant strength, use Insufficient evidence.

Use Not applicable only when no relevant decision is implicated. Use Insufficient evidence when the topic applies but essential evidence prevents a judgment. A confirmed defect can still receive a letter grade for the inspected scope, with the unassessed parts in Coverage and Limits; do not imply those parts passed. Do not award A or A+ for an incompletely assessed material risk.

| Grade | Anchor |
| --- | --- |
| F | At least one critical issue. |
| C- | Two or more independent major issues, or one major plus two or more moderate issues, or one systemic major issue. |
| C | One contained major issue, with at most one moderate issue. |
| C+ | No major or critical issue; two or more independent moderate issues. |
| B- | No major or critical issue; one moderate issue plus two or more minor issues. |
| B | No major or critical issue; one moderate issue with at most one minor issue. |
| B+ | No moderate or worse issue; two or more minor issues. |
| A- | Exactly one minor issue and no moderate or worse issue. |
| A+ | Meets A, with two independent, verified safeguards controlling distinct meaningful risks beyond routine correct setup. |
| A | No actionable issue; at least one relevant strength is verified and material applicable risks are assessed. |

If the target has no security-relevant input, privilege, sensitive data, dependency, or resource decision, return **Not applicable**. If security is relevant but essential code, policy, reachability, or runtime configuration is unavailable after reasonable inspection and could change the judgment, return **Insufficient evidence**. Explain what is missing and what would permit grading. A complete changeset with an exposed flaw is assessable even without a scanner run; a partial excerpt is not evidence that surrounding controls are absent. Do not use either state to avoid a supportable judgment.

## Report

Use the template below. For a letter grade retain every field and section; write `None verified`, `None found`, or `None needed` when appropriate. For Not applicable or Insufficient evidence, retain Scope, Coverage, Rationale, and Limits; omit counts and unsupported Good/Bad/changes. Reference each Bad finding by ID in Suggested changes. Put optional follow-ups or ungraded related findings after the required sections and label them explicitly.

```markdown
## Security — [grade | Not applicable | Insufficient evidence]
Scope: [base/head, supplied diff, or code area; library/CLI/service and relevant version/runtime context]
Coverage: [applicable areas assessed; material areas not assessed and why]
Rationale: [why the evidence and counted severities select this grade; explain systemic/critical reach or both A+ safeguards]
Finding counts: critical=[n], major=[n], moderate=[n], minor=[n]

Good

- [G1] [specific choice and evidence/location] — [why it benefits the relevant contract or risk]

Bad

- [F1][severity][introduced | worsened | existing-in-scope] [issue and evidence/location] — [trigger and supported consequence]

Suggested changes

- [F1] [targeted correction] — [why it resolves the consequence; how to verify it]

Limits: [material missing evidence; exact checks and outcomes, or checks not run]
```

Prefer `path:line` references when files are available. Give each bad finding a corresponding change and rationale. Do not include exploit instructions or live secret values in the report. The grade assesses Security in the requested scope, not overall Go quality.
