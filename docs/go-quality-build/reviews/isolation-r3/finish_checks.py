#!/usr/bin/env python3
"""Consolidate interpreted raw evidence for the exact promotion boundary."""
import json
import re
import sys
from pathlib import Path

sys.path.insert(0, '/private/tmp/go-isolation-r3-promotion')
from verify import ROOT, OUT, RESULTS, RUNS, TARGET, call, digest, inventory

integrity = json.loads((OUT / 'integrity-final.json').read_text())
assert integrity['passed']
data = {'target': TARGET, 'decision': 'accept exact isolation revision 3 unchanged for bounded promotion',
        'blocking_findings': [], 'integrity': integrity, 'executed_checks': {},
        'archived_mutation_audit': [], 'final_outcome_audit': [], 'validation_failures': []}
for mode in ['basic-restricted', 'listener-rerun', 'focused', 'defaults', 'selected-mutants', 'history-probes']:
    rs = json.loads((OUT / ('executed-' + mode + '.json')).read_text())
    for r in rs:
        label = r['label']
        expected = 0
        if mode == 'basic-restricted' and any('/' + x + ' ' in label for x in ['service-http-boundary', 'isolation-only', 'both']) and label.endswith((' clean', ' race/shuffle/three')):
            expected = 1
            assert 'bind: operation not permitted' in r['stdout'] + r['stderr']
        elif mode == 'focused' and label.endswith('shared-root parallel1 assertion'):
            expected = 1
            assert 'Get' in r['stdout']
        elif mode == 'selected-mutants' and label.endswith('candidate tests only') and 'unsafe-cancellation' not in label:
            expected = 1
            assert re.search(r'\.go:\d+:', r['stdout'])
        elif mode == 'history-probes':
            if any(label.startswith('2026-10-01-testing-combined/' + x + ' ') for x in ['architecture-only', 'behavior-only', 'isolation-only']):
                expected = 1
                assert 'requests=2, want 1' in r['stdout']
            if label.startswith('2026-10-01-testing-combined-r2/isolation-only valid '):
                expected = 1
                assert 'comparing uncomparable type' in r['stdout']
            if label == 'historical behavior2/isolation2 both unset-GOCACHE valid-HOME setup':
                expected = 1
                assert 'GOCACHE is not defined and $HOME is not defined' in r['stdout'] + r['stderr']
        if r['exit_code'] != expected:
            data['validation_failures'].append({'label': label, 'actual': r['exit_code'], 'expected': expected})
        if label.endswith(' format') and r['stdout'].strip():
            data['validation_failures'].append({'label': label, 'format_output': r['stdout']})
        r['expected_exit_code'] = expected
    data['executed_checks'][mode] = rs

def raw_records(v):
    if isinstance(v, list):
        for x in v:
            yield from raw_records(x)
    elif isinstance(v, dict):
        if any(k in v for k in ['command', 'cmd', 'argv', 'args']) and any(k in v for k in ['exit_code', 'returncode', 'return_code']):
            yield v
        for x in v.values():
            if isinstance(x, (dict, list)):
                yield from raw_records(x)

for name in RUNS:
    d = RESULTS / name
    m = json.loads((d / 'manifest.json').read_text())
    for c in m['cases']:
        case = d / c['id']
        report = case / 'blind-review.md'
        checkfile = case / 'blind-review-checks.json'
        context = json.loads((case / 'blind-review-context.json').read_text())
        assert not context['label_disclosure'] and not context['author_report_disclosure'] and context['neutral_launch_path']
        text = report.read_text()
        grades = {topic: re.search(r'## ' + re.escape(topic) + r'\s*[—& -]+\s*([^\n]+)', text).group(1).strip() if re.search(r'## ' + re.escape(topic) + r'\s*[—& -]+\s*([^\n]+)', text) else 'see exact raw report' for topic in ['Testing', 'Correctness & Compatibility', 'Architecture & Design']}
        data['final_outcome_audit'].append({'run': name, 'case': c['id'], 'report': str(report),
             'report_sha256': digest(report), 'checks_sha256': digest(checkfile), 'raw_command_records': len(list(raw_records(json.loads(checkfile.read_text())))),
             'grades': grades, 'context': context, 'author_settings': m['author_settings']})
    for p in sorted(d.glob('*/mutations/*/verification.json')):
        v = json.loads(p.read_text())
        assert v['candidate_tests_only'] and v['controller_probes_absent']
        assert digest(p.parent / 'source.patch') == v['patch_sha256']
        assert all(v[k]['exit_code'] == 0 for k in ['clean', 'applied', 'compiled'])
        expected = 0 if v['semantic_id'] == 'unsafe-cancellation-cause-comparison' else 1
        assert v['result']['exit_code'] == expected
        assertions = [x.strip() for x in v['result']['stdout'].splitlines() if re.search(r'\.go:\d+:', x)]
        assert expected == 0 or assertions
        data['archived_mutation_audit'].append({'run': name, 'case': p.parts[-4], 'semantic_id': v['semantic_id'],
            'verification_sha256': digest(p), 'compiled_exit': 0, 'result_exit': expected,
            'candidate_tests_only': True, 'controller_probes_absent': True, 'assertion_lines': assertions,
            'classification': 'test-helper sabotage' if v['semantic_id'] in ['premature-parent-teardown', 'environment-leak'] else ('supplementary survivor' if expected == 0 else 'production-regression detection')})

validation = call(['python3', '-B', '/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py', 'docs/go-quality-build/drafts/go-test-isolation'])
assert validation['exit_code'] == 0 and 'Skill is valid!' in validation['stdout']
data['quick_validation'] = validation
data['integrated_observed_artifact_sha256'] = inventory(RESULTS / RUNS[1])
data['integrated_observed_inventory_is_seal'] = False
data['limits'] = ['Go1.26.5 executed Go1.22 declarations; actual minimum compiler not run',
    'darwin/arm64; Windows and other platform execution unverified',
    'race covers exercised paths and normally built child executables are not instrumented by parent race',
    'local loopback only; external DNS/TLS/deployed services unverified',
    'all schedules, database rollback and optional newer APIs not evaluated',
    'automatic skill routing and Codex/OpenCode runtime loading unverified',
    'matching inherited settings recorded; exact model/reasoning IDs unavailable',
    'earlier arm-labelled individual review launch paths limit arm-blinded causal grade inference',
    'final single reviews disclosed three pairs; file repeat and first file use different review contexts',
    'unsafe arbitrary-comparison mutation survives both integrated candidate suites',
    'integrated revision3 remains unsealed for later delivery/whole-branch artifacts']
data['passed'] = not data['validation_failures']
assert data['passed'], data['validation_failures']
(OUT / 'checks.json').write_text(json.dumps(data, indent=2) + '\n')
print(json.dumps({'passed': data['passed'], 'fresh_command_records': sum(len(x) for x in data['executed_checks'].values()),
     'raw_outcome_records_inspected': sum(x['raw_command_records'] for x in data['final_outcome_audit']),
     'archived_candidate_only_mutants': len(data['archived_mutation_audit']),
     'sealed_artifacts_verified': sum(x['artifacts'] for x in integrity['history_seals']),
     'source_files_reconstructed': sum(x['source_files'] for r in integrity['runs'] for x in r['cases']),
     'mapped_hash_comparisons': sum(r['mapped_hashes'] for r in integrity['runs'])}, indent=2))
