#!/usr/bin/env python3
import hashlib
import json
from pathlib import Path
from run_checks import PACKET, OUT, run, hashes
TOPIC_IDS = {'architecture', 'code-quality', 'correctness', 'testing', 'security', 'observability', 'performance', 'dependencies', 'deployment'}

expected = json.loads((PACKET / 'source-manifest.json').read_text())
snapshots = hashes()
assert snapshots == json.loads((OUT / 'source-hashes-before.json').read_text())
for c in ['A', 'B']:
    assert snapshots['candidates/' + c] == expected[c]
    applicability = json.loads((OUT / ('applicability-' + c + '.json')).read_text())
    assert len(applicability) == 9
    assert {x['id'] for x in applicability} == TOPIC_IDS
    assert all(not x['essential_unavailable_evidence'] for x in applicability)
    for item in applicability:
        card = OUT / item['card']
        body = card.read_text()
        for field in ('Scope:', 'Coverage:', 'Rationale:', 'Limits:'):
            assert field in body, (card, field)
        if item['applicable']:
            for field in ('Finding counts:', '\nGood\n', '\nBad\n', '\nSuggested changes\n'):
                assert field in body, (card, field)
        assert ('— ' + item['grade']) in body
    record = run('grade-' + c, OUT, ['python3', str(PACKET / 'review-guidance/go-quality-report/scripts/grade.py'),
                                    str(OUT / ('ledger-' + c + '.json'))])
    assert record['status'] == 0
    result = json.loads(record['stdout'])
    assert result['overall'] == ('B' if c == 'A' else 'A')
    assert result['coverage_gaps'] == []
    (OUT / ('grade-' + c + '.json')).write_text(json.dumps(result, indent=2) + '\n')

records = [json.loads(p.read_text()) for p in sorted((OUT / 'raw').glob('*.json'))]
ordinary = [r for r in records if '-mutation-' not in r['id']]
assert all(r['status'] == 0 and not r['timed_out'] for r in ordinary)
mutations = json.loads((OUT / 'mutation-sensitivity.json').read_text())
assert len(mutations) == 16
assert all(not r['build_failure'] for r in mutations)
assert sum(r['behavioral_assertion_detected'] for r in mutations) == 14
assert sum(r['deadline_rejection'] for r in mutations) == 2
for c in ['A', 'B']:
    r = json.loads((OUT / 'raw' / (c + '-mutation-drop-unit-updates-event-probe-go122.json')).read_text())
    assert r['status'] == 1
    assert 'checkpoint snapshot: got {Count:0 Sum:0}, want {4 4}' in r['stdout']
    assert 'final snapshot: got {Count:0 Sum:0}, want {4000 4000}' in r['stdout']
    assert 'test timed out' not in r['stdout'] + r['stderr']

artifacts = []
for p in sorted(OUT.rglob('*')):
    if p.is_file() and not any(x in {'gocache', 'modcache', 'work', '__pycache__'} for x in p.relative_to(OUT).parts):
        if p.name != 'freeze-manifest.json':
            artifacts.append({'path': str(p.relative_to(OUT)), 'bytes': p.stat().st_size,
                              'sha256': hashlib.sha256(p.read_bytes()).hexdigest()})
probe_sources = []
for p in sorted((OUT / 'work').rglob('*')):
    if p.is_file():
        probe_sources.append({'path': str(p.relative_to(OUT)), 'sha256': hashlib.sha256(p.read_bytes()).hexdigest()})
freeze = {'status': 'frozen-ready', 'date': '2026-10-02',
          'verification': {'candidate_source_hashes_preserved': True, 'original_source_hashes_preserved': True,
                           'all_nine_cards_per_candidate': True, 'grade_py_A': 'B', 'grade_py_B': 'A',
                           'material_coverage_gaps': [], 'baseline_and_contract_checks_all_pass': True,
                           'supplementary_mutation_runs': 16, 'behavioral_assertion_detections': 14,
                           'runner_deadline_only_rejections': 2, 'build_errors_credited': 0,
                           'event_probe_wrong_state_assertions_confirmed': True},
          'artifacts': artifacts, 'disposable_probe_sources': probe_sources,
          'sources': snapshots,
          'note': 'Evidence/report artifacts are ready for archival. No further source or report edits follow this manifest; build caches are excluded. All results are bounded independent supplementary checks, not unspecified original frozen-target credit.'}
(OUT / 'freeze-manifest.json').write_text(json.dumps(freeze, indent=2) + '\n')
print(json.dumps({'frozen_ready': True, 'artifact_count': len(artifacts), 'cards': 18,
                  'A': 'B', 'B': 'A', 'source_hashes_preserved': True,
                  'raw_command_records': len(records), 'freeze_manifest': str(OUT / 'freeze-manifest.json')}, indent=2))
