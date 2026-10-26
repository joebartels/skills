# Review record v1

Use this record for standalone topic cards and reconciled overall reviews. Python 3.10+ and the installed `go-quality-review` package are sufficient; no third-party runtime packages or source-repository files are needed. Resolve the script relative to this reference, then invoke its absolute path from any working directory:

```sh
python3 /absolute/path/to/go-quality-report/scripts/render.py /output/review.json --output /output/report.md
python3 /absolute/path/to/go-quality-report/scripts/render.py /output/review.json --topic go-security
```

Without `--output`, Markdown goes to stdout. `--topic` projects a topic from an overall record. Keep records and generated reports in the host's scratch/artifact directory, outside the reviewed checkout. Deliver the rendered report without rewriting its tables, counts or field labels unless the user requests a different presentation. If Python is unavailable, disclose the limitation and follow the [same output structure](reporting.md) manually; do not claim deterministic rendering.

## From candidates to a record

Inspect source before assigning grades. For each candidate, trace the trigger, callers, guards, existing rule owners and meaningful test consumers. Record counterevidence and what the correction removes after preserving necessary responsibilities in `analysis`. Admit only supported consequences; a small current maintenance cost can qualify without a runtime bug. Apply the topic's existing severity anchors.

`findings` contains admitted canonical causes. Give each its own correction and primary owner; merge symptoms fixed by the same correction. `assessments` preserves independently supported topic consequences and severities. Namespace original worker IDs in `sources`; split a source ID into explicit subclaims if it contains independent causes. Use `decisions` for consequential rejected, unresolved, merged or out-of-scope related candidates, not every search lead. `unresolved` means missing evidence that could change the assessment and automatically creates a coverage gap for its topics. Optional refinements without an established cost are `rejected`, with the reason retained.

The renderer validates shape/references, derives topic and overall ledgers, and calls the existing grade calculator. It neither discovers omissions nor proves evidence, severity, independence or coverage. Reconcile those before rendering; never adjust finding counts to obtain a desired grade.

## Fields

All fields below are required unless marked optional. Unknown fields and duplicate JSON keys are errors. Text must be nonempty; text arrays may be empty unless stated otherwise. Strings can contain Markdown links and inline code; whitespace is folded for stable fields/tables. Link long logs and traces instead of embedding them.

| Object | Fields |
| --- | --- |
| Record | `version: 1`, `report`, `mode`, `scope`, `rationale`, `topics`, `findings`, `decisions`, `strengths`, `safeguards`, `checks`, `limits`; optional `maintainability` (required for overall) |
| Topic | `id` (installed topic skill name), `rationale`, nonempty `coverage`, `cards` (Markdown references; empty for a direct review) |
| Coverage | `id` (unique within topic), `status`, `reason` |
| Finding | `id` (`F1`, etc.), `title`, `cause`, `status`, `trigger`, nonempty `evidence`, `analysis`, nonempty `assessments`, `owner`, `correction`, `verification` (proposed checks), nonempty `sources` (unique provenance IDs across findings) |
| Assessment | `topic`, `severity`, `consequence` (effect and severity rationale); optional `systemic` (boolean, default false), `reach` (required for critical or systemic major) |
| Evidence | `kind`, `reference` (source location, command/log or supplied artifact), `text` (supported observation) |
| Decision | `id` (`D1`, etc.), `title`, `disposition`, nonempty `topics`, `reason`, nonempty `evidence`; `finding_id` required only for `merged` |
| Strength | `id` (`G1`, etc.), nonempty `topics`, `description`, nonempty `evidence` |
| Safeguard | `id` (`S1`, etc.), nonempty `topics`, `risk` (canonical distinct risk), nonempty `evidence`, `nonroutine` (boolean) |
| Check | `kind`, `command` (exact command or named check), `result` (outcome or why not run) |
| Maintainability | `rows` (all seven dimensions), `cleanliness`, `structure`, `extensibility` |
| Profile row | `dimension`, `rating`, `confidence`, `evidence` (including counterevidence and labels), `limits` (scope and next validation) |

Enums and relationships:

- `report`: `overall` or one of the nine topic skill names. Overall records require all nine topics, including justified irrelevance; standalone records require exactly their named topic. Topic references must resolve within the record, except `related` decisions may name any installed topic. `mode`: `changeset` or `code-area`; include exact revisions/snapshot, exclusions and relevant runtime context in `scope`.
- Coverage `status`: `complete`, `not_applicable`, `pending`, `partial`, `unavailable`. Findings, strengths, safeguards and unresolved decisions cannot target a wholly not-applicable topic.
- Finding `status`: `introduced`, `worsened`, `existing-in-scope`; changeset findings cannot be `existing-in-scope`. Severity: `critical`, `major`, `moderate`, `minor`. Only major may be systemic. Overall severity is the greatest supported topic consequence; shared causes count once overall and once in each affected topic.
- Evidence `kind`: `inspected`, `executed`, `supplied`. Check `kind`: `executed`, `supplied`, `not_run`. Proposed verification belongs in the finding, never among executed checks.
- Decision `disposition`: `rejected`, `unresolved`, `merged`, `related`. Use `related` for confirmed observations outside the graded scope; state the boundary in `reason`. Related observations stay visible without deductions or coverage gaps. A merge target must exist and cover the decision's topics. Neither rejected nor merged decisions add deductions; unresolved decisions block complete grades while preserving confirmed findings.
- IDs are unique within each collection. Omit authored grades and counts: the renderer computes them. Safeguards do not cancel defects; A+ still needs a verified strength and two independent nonroutine safeguards with complete coverage.
- Profile dimensions, ratings and confidence use [maintainability.md](maintainability.md). Its seven ratings remain independent of defect grades. Use `Not demonstrated` with a reason for unassessed dimensions.

## Complete standalone example

```json
{
  "version": 1,
  "report": "go-code-quality-and-idioms",
  "mode": "code-area",
  "scope": "transport at abc123; login and refresh handlers and their consumers",
  "rationale": "One localized repeated mapping; no broader failure established.",
  "topics": [{
    "id": "go-code-quality-and-idioms",
    "rationale": "One minor maintenance cost.",
    "coverage": [{"id": "cookies", "status": "complete", "reason": "All issuance paths and cookie helpers inspected."}],
    "cards": []
  }],
  "findings": [{
    "id": "F1",
    "title": "Repeated cookie-suite installation",
    "cause": "Three implementations of the same result-to-cookie mapping",
    "status": "existing-in-scope",
    "trigger": "Audit how login and refresh install the complete token suite.",
    "evidence": [{"kind": "inspected", "reference": "login.go:42; refresh.go:80,130", "text": "Each installs the same three token fields."}],
    "analysis": "The three handlers have distinct responses. Cookie names and attributes already have one owner; only the suite mapping is repeated.",
    "assessments": [{"topic": "go-code-quality-and-idioms", "severity": "minor", "consequence": "One mapping must be inspected at three sites; no current divergence or substantial coordination cost."}],
    "owner": "transport",
    "correction": "Extract the suite setter, preserving each handler's response and flow-cookie handling.",
    "verification": "Exercise login and both refresh routes; assert the same cookie values and attributes.",
    "sources": ["transport/code-quality/F1"]
  }],
  "decisions": [],
  "strengths": [{"id": "G1", "topics": ["go-code-quality-and-idioms"], "description": "Cookie attributes have one owner.", "evidence": [{"kind": "inspected", "reference": "cookies.go:20", "text": "All handlers use the shared cookie constructors."}]}],
  "safeguards": [],
  "checks": [],
  "limits": ["Source inspection only; no tests executed."]
}
```

This example renders A- with one minor finding. Its paths are illustrative, not evidence about a real checkout. For zero findings use an empty `findings` array; complete scope still needs a verified relevant strength for A. Incomplete coverage renders Insufficient evidence and retains any confirmed findings with an assessed-scope defect grade.
