from datetime import datetime, timezone
from hashlib import sha256
import json
from pathlib import Path
import subprocess

ROOT = Path('/private/tmp/go-context-outcome-20261002/control')
OUT = ROOT / 'reviews'
SKILLS = ROOT / 'review-skills'
snapshot = json.loads((OUT / 'snapshot.json').read_text())
assert snapshot['original_unchanged'] and snapshot['review_contracts_unchanged']

topics = [
    ('architecture-and-design', 'Architecture & Design', 'complete',
     'The private function boundary and the explicitly required small serial design are in scope.',
     'The entire package, private signature, import/dependency boundary, and simplicity obligations were inspected; there are no runtime ownership or transport boundaries.',
     '[G1] `range.go:3` preserves `func inRange(value, low, high int) bool` and expresses the behavior with one pure Boolean expression. The only added function is a focused test at `range_test.go:11`; no exported API, options, context, lifecycle machinery, or unrelated abstraction was introduced. This fits the stated private predicate contract.',
     'Source inspection and the original comparison establish the narrow API/design obligations. No external consumer architecture is claimed.'),
    ('code-quality-and-idioms', 'Code Quality & Go Idioms', 'complete',
     'The implementation and regression tests present local clarity and Go idiom decisions.',
     'All Go source and tests, names, control flow, scalar value semantics, effective Go directive, and mechanical format/vet diagnostics were assessed.',
     '[G1] `range.go:3` reads directly as the inclusive-interval contract and has no hidden state or arithmetic workaround. `range_test.go:11-14` names the upper-bound expectation and emits a useful failure. Independent `gofmt -d range.go range_test.go` produced no diff and `go vet -mod=readonly ./...` exited 0.',
     'Mechanical checks used Go 1.26.5; no optional rewrite to a table, helper, or newer API is necessary for this tiny predicate.'),
    ('correctness-and-compatibility', 'Correctness & Compatibility', 'complete',
     'Inclusive lower and upper boundaries, private signature compatibility, and the Go 1.22 minimum are mandatory behavior contracts.',
     'The complete predicate, both tests, README, and go.mod were assessed. Ordinary, outside, lower, upper, negative, singleton, zero, and int-extreme inputs were traced and checked. Signature and support metadata were compared with the original.',
     '[G1] `range.go:3` uses `value >= low && value <= high`. For `low <= high`, this includes both bounds and excludes values outside them, without addition/subtraction that could overflow at int extremes. The independently executed contract/boundary suite passes, including singleton intervals and MinInt/MaxInt. The only production difference from the original is `<` to `<=`; the private signature and `go.mod:3` Go 1.22 directive are preserved.',
     'The promised domain is `low <= high`; no inverted-interval policy is invented. Runtime checks used Go 1.26.5 darwin/arm64, and an actual Go 1.22 runtime was not exercised. The host compiler accepted the package with `-lang=go1.22`; unchanged primitive operators and legacy testing APIs introduce no newer-language or library dependency.'),
    ('testing', 'Testing', 'complete',
     'Meaningful regression detection for the inclusive upper bound is required.',
     'All supplied candidate tests and the original test were inspected; assertions, failure diagnostics, isolation, and direct invocation of the private function were traced. Each candidate was checked separately against both its implementation and the original buggy implementation in disposable copies.',
     None,
     'The boundary probes are diagnostic scratch tests, not edits to candidate source. No race, repeated-order, fuzz, or benchmark check was needed for this deterministic scalar predicate with no shared state or resource lifecycle. No external CI enforcement is claimed.'),
    ('security', 'Security', 'not_applicable',
     'The supplied code is a private scalar comparison with no sensitive assets, privileged action, trust boundary, parsing, I/O sink, credentials, external dependency, or variable resource amplification. Nothing in the specified repair creates a security decision.',
     'The whole packet code area and imports were inspected to establish applicability; security is not inferred from a hypothetical caller.',
     None,
     'No vulnerability scanner was run. Security properties of external consumers or deployments were not supplied and are outside this packet, not evidence of clean external systems.'),
    ('observability-and-resilience', 'Observability & Resilience', 'not_applicable',
     'The pure local predicate cannot block on or fail an external operation, has no retry or lifecycle contract, and does not require logs, metrics, tracing, cancellation, or deadlines. README explicitly preserves the small serial design.',
     'The complete function, tests, and README were inspected; no operational signal or failure-containment decision is implicated.',
     None,
     'No external operational environment is supplied. Test failure diagnostics are assessed under Testing, not used to invent runtime telemetry requirements.'),
    ('performance-and-resource-management', 'Performance & Resource Management', 'complete',
     'Preserving the small serial implementation implicates the predicate cost and resource-use decision.',
     'The complete scalar expression and all source paths were assessed for bounded work, arithmetic, allocation, retained state, I/O, and concurrency. There are no acquired resources or lifecycle obligations.',
     '[G1] `range.go:3` performs at most two integer comparisons with short-circuit Boolean evaluation. Work and space are constant per invocation, independent of interval width, and the implementation creates no goroutine, timer, buffer, external resource, or heap-backed object. The upper-bound fix preserves this bound and the required serial implementation.',
     'This is a source-derived complexity/resource assessment. No measured latency, throughput, compiler allocation diagnostic, profile, or benchmark result is claimed; none is required to justify the bounded expression.'),
    ('dependencies-and-reproducibility', 'Dependencies & Reproducibility', 'complete',
     'The named code-area includes the standalone module and the explicit Go 1.22 support requirement, so unchanged build inputs are assessed.',
     'All supplied module configuration, source imports, and file inventory were inspected. The selected module graph and standalone readonly tests were checked with workspace disabled, network resolution disabled, and fresh diagnostic caches.',
     '[G1] `go.mod:1-3` keeps module `example.com/rangecheck` and `go 1.22`, with no requires, replacements, tools, or generated inputs. Production has no imports; tests use only `testing`. Independent `go list -mod=readonly -m all` selects only the main module, and the fresh-cache standalone `go test -count=1 -mod=readonly ./...` succeeds with `GOWORK=off`, `GOPROXY=off`, and `GOTOOLCHAIN=local`. No external dependency or checksum file is needed for this graph.',
     'Fresh diagnostic caches prove the executed standalone host build without network resolution, not byte-identical artifacts or every platform. An actual Go 1.22 toolchain was not used; the preserved directive, pre-existing syntax/APIs, and host `-lang=go1.22` check support the narrow minimum-version assessment. Release tooling and broader platform matrices are outside the supplied area.'),
    ('deployment-and-operations', 'Deployment & Operations', 'not_applicable',
     'This packet defines a private library predicate repair and tests, with no executable artifact, release workflow, packaging, runtime configuration, probe, shutdown, rollout, or recovery obligation. Standalone module resolution is assessed under Dependencies & Reproducibility.',
     'README and the full four-file packet inventory were inspected to establish scope. No delivery or operational decision is implicated by the requested repair.',
     None,
     'External CI, release, and deployment environments are not supplied and remain unassessed; their absence from the bounded packet is not graded as missing infrastructure.'),
]

all_manifest = {
    'review_id': 'go-context-outcome-20261002-neutral-control-code-area',
    'mode': 'code-area',
    'reviewer': '/root/context_checked_control_review',
    'requested_scope': 'Independently assess each supplied candidate code area against original README/code, without source edits or arm disclosure.',
    'snapshot_artifact': str(OUT / 'snapshot.json'),
    'source_paths': ['README.md', 'go.mod', 'range.go', 'range_test.go'],
    'exclusions': ['Repository/planning/results/writing-skill material outside packet', 'External consumers, CI, releases, deployment environments'],
    'supplied_results_policy': 'provided-checks-1.json and provided-checks-2.json were not read or used; raw independent check outputs are preserved.',
    'delegation': 'One bounded review worker; topic assessments sequential within this worker, two independent candidate check jobs executed concurrently.',
    'skills_read': [str(SKILLS / 'go-quality-report' / 'SKILL.md')] + [str(SKILLS / ('go-' + t[0]) / 'SKILL.md') for t in topics],
    'references_read': [str(SKILLS / 'go-quality-report' / 'references' / n) for n in ('orchestration.md', 'reporting.md', 'grading.md')] + [str(SKILLS / topic / 'references' / filename) for topic, filename in (
        ('go-code-quality-and-idioms', 'idiom-decisions.md'),
        ('go-testing', 'testing-decisions.md'),
        ('go-correctness-and-compatibility', 'correctness-compatibility-decisions.md'),
        ('go-dependencies-and-reproducibility', 'dependency-reproducibility-decisions.md'))],
    'candidates': []
}

for n in (1, 2):
    name = f'candidate-{n}'
    dest = OUT / name
    cards = dest / 'cards'
    cards.mkdir(exist_ok=True)
    checks = json.loads((dest / 'independent-checks.json').read_text())
    assert checks['all_expected_outcomes'] and checks['source_unchanged']
    for filename, expected in snapshot['candidate_hashes'][name].items():
        assert sha256((ROOT / name / filename).read_bytes()).hexdigest() == expected
    coverage = []
    table = []
    for slug, title, status, rationale, extent, strength, limits in topics:
        grade = 'A' if status == 'complete' else 'Not applicable'
        coverage.append({'id': slug, 'status': status, 'reason': rationale, 'owner': all_manifest['reviewer'],
                         'card': str(cards / (slug + '.md'))})
        table.append(f'| {title} | {grade} | {rationale} | [Card](cards/{slug}.md) |')
        scope = (f'Code-area review of the complete `{name}` four-file packet at the SHA-256 snapshot in '
                 f'`../snapshot.json`; private library predicate in module `example.com/rangecheck`, `go 1.22`. '
                 f'Original README/code are contract and comparison context. Candidate source is `{ROOT / name}`.')
        if slug == 'testing':
            test_name = 'TestUpperInclusive' if n == 1 else 'TestUpperBoundIncluded'
            extra = (' Candidate 1 also asserts exclusion of 7 at `range_test.go:15-17`; that behavior is already covered by the retained original test.' if n == 1 else ' The retained original test already asserts exclusion of 7 above the upper bound.')
            strength = (f'[G1] `{test_name}` at `range_test.go:11-14` calls the real private function at `(6, 2, 6)` and fails if the upper bound is excluded. '
                        f'Independent execution passes on this candidate; replacing only the implementation in a disposable copy with original `value < high` causes this exact test to fail at line 13 with exit 1. '
                        f'This establishes regression detection for the requested bug rather than relying on a test name or coverage number.{extra}\n\n'
                        '- [G2] The retained `TestInteriorAndLower` at `range_test.go:5-9` checks lower, interior, below, and above outcomes without external resources, clocks, global state, or test doubles. It passes independently.')
        if status == 'complete':
            text = (f'## {title} — A\n\nScope: {scope}\n\nCoverage: {extent}\n\n'
                    f'Rationale: No actionable issue is substantiated. The relevant strength below is verified and all material obligations of this bounded area were assessed, selecting A under the unchanged topic contract. Routine correct setup and tests do not establish two nonroutine safeguards for A+.\n\n'
                    'Finding counts: critical=0, major=0, moderate=0, minor=0\n\n'
                    f'Good\n\n- {strength}\n\nBad\n\n- None found.\n\nSuggested changes\n\n- None needed.\n\n'
                    f'Limits: {limits} Exact independently executed commands, working directories, stdout/stderr, and exit codes are in [independent-checks.json](../independent-checks.json) and [evidence](../evidence/). Supplied result JSON was not used. Reviewed source and review contracts remained unchanged.\n')
        else:
            text = (f'## {title} — Not applicable\n\nScope: {scope}\n\nCoverage: {extent}\n\n'
                    f'Rationale: {rationale}\n\nLimits: {limits}\n')
        (cards / (slug + '.md')).write_text(text)
    (dest / 'applicability.json').write_text(json.dumps(coverage, indent=2) + '\n')
    ledger = {
        'scope': f'Complete supplied {name} code area, README.md/go.mod/range.go/range_test.go; snapshot ../snapshot.json',
        'coverage': [{'id': c['id'], 'status': c['status'], 'reason': c['reason']} for c in coverage],
        'findings': [],
        'strengths': [
            {'id': 'G1', 'evidence': f'{name}/range.go:3 correct inclusive comparison; independent suite plus singleton/negative/int-extreme probes pass.'},
            {'id': 'G2', 'evidence': f'{name}/range_test.go:11 upper-bound regression test passes on candidate and fails on original buggy implementation in scratch.'},
            {'id': 'G3', 'evidence': f'{name}/go.mod preserved, only main module selected, fresh-cache standalone readonly test succeeds without network resolution.'}],
        'safeguards': []
    }
    ledger_path = dest / 'ledger.json'
    ledger_path.write_text(json.dumps(ledger, indent=2) + '\n')
    grade_cmd = ['rtk', 'proxy', 'python3', str(SKILLS / 'go-quality-report/scripts/grade.py'), str(ledger_path)]
    grade_run = subprocess.run(grade_cmd, capture_output=True, text=True)
    assert grade_run.returncode == 0, grade_run.stderr
    grade = json.loads(grade_run.stdout)
    assert grade['overall'] == 'A' and not grade['coverage_gaps']
    (dest / 'grade.json').write_text(grade_run.stdout)
    (dest / 'evidence/grading-command.json').write_text(json.dumps({'command': grade_cmd, 'exit_code': grade_run.returncode, 'stdout': grade_run.stdout, 'stderr': grade_run.stderr}, indent=2) + '\n')
    findings = {'confirmed': [], 'unconfirmed': [], 'reconciliation': 'No causal defects or contradictions were found. Topic strengths overlap and are not counted as A+ safeguards.'}
    (dest / 'findings.json').write_text(json.dumps(findings, indent=2) + '\n')
    report = f'''# Go Quality Report — A

Scope: Code-area review of complete `{name}` source (`README.md`, `go.mod`, `range.go`, `range_test.go`), identified by the per-file SHA-256 values in [snapshot.json](../snapshot.json). Private predicate library; Go 1.22 minimum. Original README/code supply the intended contract and comparison baseline. No source edits.

Coverage: All nine topics considered. Six are applicable and complete; Security, Observability & Resilience, and Deployment & Operations are genuinely irrelevant to this bounded local operation. All mandatory correctness, code quality, and testing obligations are complete.

Rationale: The minimal `<` to `<=` production fix implements the stated inclusive interval, preserves its signature and constant-time serial design, and has a meaningful focused regression test. No independently actionable defect was confirmed. Verified strengths and complete applicable coverage select A; ordinary setup and tests do not earn A+.

Unique finding counts: critical=0, major=0, moderate=0, minor=0

## Report cards

| Topic | Grade/state | Assessed coverage and rationale | Cards |
| --- | --- | --- | --- |
{chr(10).join(table)}

## Good

- [G1] `range.go:3` uses `value >= low && value <= high`; independent contract checks pass at lower, upper, interior, outside, singleton, negative, zero, and integer-extreme inputs. This satisfies the inclusive-boundary contract without an overflow-prone arithmetic workaround.
- [G2] The focused upper-bound regression test at `range_test.go:11-14` passes on this candidate and fails at line 13 when the original implementation is substituted in a scratch copy. It detects the actual defect.
- [G3] The private signature, `go 1.22`, serial Boolean expression, and dependency-free production remain preserved. Formatting/vet checks are clean and an independently executed fresh-cache standalone readonly suite passes with network resolution disabled.

## Bad

- None found.

## Suggested changes

- None needed.

## Limits

Checks were executed independently in disposable copies, using Go 1.26.5 darwin/arm64. An actual Go 1.22 runtime was not exercised; the preserved directive, unchanged old syntax/APIs, and passing host `-lang=go1.22` check support this narrow source compatibility assessment. No broader runtime/platform matrix, CI, release, external consumer, vulnerability scan, benchmark, or deployment claim is made. The supported interval domain is `low <= high`. Supplied result JSON was not used. One bounded worker assessed topics sequentially; no second reviewer was used. Raw logs preserve exact commands and outcomes, including the expected failed-original regression check. All original/candidate source and packet review contracts remained unchanged.

Details: [all-nine applicability](applicability.json), [independent checks](independent-checks.json), [findings](findings.json), [ledger](ledger.json), [calculator output](grade.json), [manifest](../manifest.json).
'''
    (dest / 'report.md').write_text(report)
    all_manifest['candidates'].append({'id': name, 'status': 'complete', 'grade': 'A', 'coverage': coverage,
                                     'boundary_owner': all_manifest['reviewer'],
                                     'boundaries': ['Inclusive interval for low <= high', 'Private signature and serial implementation', 'Meaningful upper-bound regression detection', 'Go 1.22 metadata/source compatibility and standalone module inputs'],
                                     'evidence': str(dest / 'independent-checks.json'), 'report': str(dest / 'report.md')})

(OUT / 'manifest.json').write_text(json.dumps(all_manifest, indent=2) + '\n')
for path, expected in snapshot['review_contract_hashes'].items():
    assert sha256((ROOT / path).read_bytes()).hexdigest() == expected
for area, expected_hashes in [('original', snapshot['original_hashes']), *snapshot['candidate_hashes'].items()]:
    for name, expected in expected_hashes.items():
        assert sha256((ROOT / area / name).read_bytes()).hexdigest() == expected

freeze_paths = [OUT / 'manifest.json', OUT / 'snapshot.json']
for n in (1, 2):
    dest = OUT / f'candidate-{n}'
    freeze_paths += sorted(p for p in dest.rglob('*') if p.is_file() and 'scratch' not in p.parts)
freeze = {'frozen_at_utc': datetime.now(timezone.utc).isoformat(),
          'disclosure_state': 'Cards, applicability, ledgers, reports and independent evidence finalized before arm disclosure.',
          'source_and_review_contracts_unchanged': True,
          'artifacts': {str(p.relative_to(OUT)): sha256(p.read_bytes()).hexdigest() for p in freeze_paths}}
(OUT / 'freeze.json').write_text(json.dumps(freeze, indent=2) + '\n')
print(json.dumps({'grades': {'candidate-1': 'A', 'candidate-2': 'A'},
                  'finding_counts_per_candidate': {'critical': 0, 'major': 0, 'moderate': 0, 'minor': 0},
                  'cards_per_candidate': 9, 'freeze': str(OUT / 'freeze.json'),
                  'artifact_count': len(freeze_paths)}, indent=2))
