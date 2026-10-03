"""Freeze preparation once before authors; not a post-seal verification command."""
from pathlib import Path
import json,hashlib,shutil,subprocess
ROOT=Path(__file__).resolve().parents[4];OUT=Path(__file__).resolve().parent;PREP=OUT/'preparation';EVALS=ROOT/'tests/go-quality-build/go-concurrency-and-ownership/evals';CONTEXT=ROOT/'tests/go-quality-build/results/2026-10-02-context-study'
assert not (OUT/'manifest.json').exists(), 'Do not overwrite a frozen manifest'
cm=json.loads((CONTEXT/'manifest.json').read_text());base=cm['comparator_revision'];skills={k:v for k,v in cm['skill_hashes'].items() if not k.startswith('go-context-and-deadlines/')}
assert len(skills)==8 and sum(k.endswith('/SKILL.md') for k in skills)==5
verified=[]
for name,h in skills.items():
 dst=PREP/'skills'/name;dst.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(CONTEXT/'skills'/name,dst)
 p=subprocess.run(['rtk','proxy','git','show',base+':plugins/go-quality-build/skills/'+name],capture_output=True)
 assert p.returncode==0 and p.stdout==dst.read_bytes() and hashlib.sha256(p.stdout).hexdigest()==h,name
 verified.append({'path':name,'sha256':h,'base_equal':True})
(PREP/'snapshot-inventory-verification.json').write_text(json.dumps(verified,indent=2)+'\n')
checks=json.loads((PREP/'conformance-green.json').read_text());assert sum(map(len,checks.values()))==23 and all(c['status']==0 for rows in checks.values() for c in rows)
mutations=json.loads((PREP/'mutation-feasibility.json').read_text());expected=['incoherent Count/Sum publication','lease closed before all started users joined','blocked admission observes cancellation before input closure','blocked producer requested to stop']
for m,s in zip(mutations,expected):assert m['qualified'] and m['compile']['status']==0 and m['held_named']['status']==1 and s in m['held_named']['stdout'],m['meaning']
inputs={str(p.relative_to(EVALS)):hashlib.sha256(p.read_bytes()).hexdigest() for p in EVALS.rglob('*') if p.is_file()};cases=json.loads((EVALS/'evals.json').read_text())['evals'];author=[name for c in cases for name in c['files']];assert len(inputs)==23 and len(author)==17 and all(not name.startswith('controller-probes/') for name in author)
override=PREP/'host-skill-disable-followed-override.txt';assert hashlib.sha256(override.read_bytes()).hexdigest()==cm['host_disable_sha256']
m={'date':'2026-10-02','candidate':'go-concurrency-and-ownership','preparation_frozen':True,'execution_base':'83c3347429b4a15736a239a2f2ac47a6a17717aa','comparator_revision':base,'skill_hashes':skills,'input_hashes':inputs,'author_files':author,'host_disable_sha256':cm['host_disable_sha256'],'profiles':cm['profiles'],'primary_case':'service-owned-workers','context_available':False,'author_ceiling':16,'group_ceiling':36,'group_authors_spent_before_candidate':16,'attempts':[],'allocation':{'discovery_reference':3,'matched_reference_exposures':3,'unseen_transfer_pair':2,'alternate_model_primary_pair':2,'alternate_reasoning_primary_pair':2,'remaining_revision_or_failed_retry':4,'other_native_target_pairs':0},'selection_probe_ceiling':20,'group_selection_probe_ceiling':40,'group_selection_probes_spent_before_candidate':14,'selection_probes_spent':0,'limits':['Claude unauthenticated; intended OpenCodev2 unavailable. No simulated loading pass.','Five original runtime skills comprise eight files; context-study catalog nine includes excluded context draft.','Every failed author launch counts. A failed attempt permits at most one retry using an unspent slot.','Controller conformance is fixture validation, excluded from authors and uplift.']}
(OUT/'manifest.json').write_text(json.dumps(m,indent=2)+'\n')
print('PASS freeze: 23 input,17 author,8 comparator files;23 conformance commands;4 compiling intended mutations;0 authors')
