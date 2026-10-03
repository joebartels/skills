import hashlib
import json
from pathlib import Path

ROOT=Path('/private/tmp/go-neutral-pipeline-study')
OUT=ROOT/'independent-review'
before=json.loads((OUT/'source-hashes-before.json').read_text())
after={p:hashlib.sha256((ROOT/p).read_bytes()).hexdigest() for p in before}
assert before==after
assert before==json.loads((OUT/'source-hashes-after.json').read_text())
preserved={c:{f:before[f'candidates/{c}/{f}']==before[f'original/{f}'] for f in ['README.md','go.mod','cmd/pipeline/main.go']} for c in ['A','B']}
assert all(all(files.values()) for files in preserved.values())
commands=json.loads((OUT/'raw/independent-commands.json').read_text())
followups=json.loads((OUT/'raw/followup-commands.json').read_text())
assert len([r for r in commands if '/executable/' in r['label']])==20
assert all(r['contract_match'] for r in commands if '/executable/' in r['label'])
assert len([r for r in commands if r.get('confirmed_detection')])==8
assert all(r['argv'][0]=='rtk' for r in commands+followups)
for candidate in ['A','B']:
    applicable=json.loads((OUT/f'applicability-{candidate}.json').read_text())
    assert len(applicable['topics'])==9
    for row in applicable['topics']:
        assert row['state']=='applicable' and row['coverage']=='complete' and row['reason']
        card=(OUT/row['card']).read_text()
        assert all(section in card for section in ['Scope:','Coverage:','Rationale:','Finding counts:','Good\n','Bad\n','Suggested changes\n','Limits:','Skill invoked and read:'])
    ledger=json.loads((OUT/f'ledger-{candidate}.json').read_text())
    result=json.loads((OUT/f'grade-{candidate}.json').read_text())
    assert len(ledger['coverage'])==9 and not result['coverage_gaps']
    assert result['overall']=='C-'
    assert result['unique_finding_counts']['major']==(3 if candidate=='A' else 2)
    assert len({r['id'] for r in ledger['findings']})==len(ledger['findings'])
    assert all(r['magnitude'] and r['sources'] for r in ledger['findings'])
    assert len(list((OUT/'cards'/candidate).glob('*.md')))==9
for p in OUT.rglob('*.json'):
    if 'cache' not in p.parts and 'work' not in p.parts:
        json.loads(p.read_text())
validation=dict(source_unchanged=True,source_manifest_matches=True,preserved_supplied_host_parser_module=preserved,topic_cards=18,applicability_topics_per_candidate=9,actual_binary_cases=20,all_binary_contracts_match=True,confirmed_bounded_review_mutation_detections=8,narrowed_join_author_suite_count=10,primary_command_records=len(commands),followup_command_records=len(followups),calculator_unchanged_sha256=after['review-guidance/go-quality-report/scripts/grade.py'],result='ready for freeze',limits='Artifacts preserve review-created supplementary mutations; original frozen author-specific mutation artifacts are absent from this packet. No timeout-only result is credited as a detection.')
(OUT/'raw/final-validation.json').write_text(json.dumps(validation,indent=2)+'\n')
files={str(p.relative_to(OUT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(OUT.rglob('*')) if p.is_file() and 'cache' not in p.parts and 'work' not in p.parts and p.name!='freeze-manifest.json'}
(OUT/'freeze-manifest.json').write_text(json.dumps(dict(status='ready',files=files,excluded_generated_scratch=['cache/','work/'],reviewed_source_hashes=before),indent=2)+'\n')
print(json.dumps(validation,indent=2))
