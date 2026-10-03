#!/usr/bin/env python3
"""Repair only timeout parsing from preserved raw records; do not rerun evidence."""
import json
from run_checks import OUT

rows = json.loads((OUT / 'mutation-sensitivity.json').read_text())
for row in rows:
    record = json.loads((OUT / row['raw_record']).read_text())
    output = record['stdout'] + record['stderr']
    row['deadline_rejection'] = 'test timed out' in output
    row['build_failure'] = '[build failed]' in output
    row['behavioral_assertion_detected'] = ('--- FAIL:' in output and not row['deadline_rejection'] and not row['build_failure'])
(OUT / 'mutation-sensitivity.json').write_text(json.dumps(rows, indent=2) + '\n')
(OUT / 'evidence-classification-correction.json').write_text(json.dumps({
    'change': 'Recognize Go runner output panic: test timed out rather than an assumed testing: prefix.',
    'raw_execution_records_changed': False,
    'grading_changed': False,
    'behavioral_assertion_detections': sum(r['behavioral_assertion_detected'] for r in rows),
    'runner_deadline_only_rejections': sum(r['deadline_rejection'] for r in rows),
    'build_errors': sum(r['build_failure'] for r in rows),
}, indent=2) + '\n')
print('Reclassified preserved raw mutation records; no checks rerun and no source modified.')
