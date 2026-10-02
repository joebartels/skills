"""Read-only frozen identities/reconstruction/outcome-review gate; never rewrites evidence."""
from pathlib import Path
import hashlib,json,shutil,subprocess,tempfile
OUT=Path(__file__).resolve().parent;ROOT=OUT.parents[3];EVALS=ROOT/'tests/go-quality-build/go-concurrency-and-ownership/evals'
m=json.loads((OUT/'manifest.json').read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
assert m['preparation_frozen'] and m['primary_case']=='service-owned-workers' and m['context_available'] is False
assert len(m['skill_hashes'])==8 and len(m['input_hashes'])==23 and len(m['author_files'])==17
for name,h in m['input_hashes'].items():assert sha(EVALS/name)==h,name
for name,h in m['skill_hashes'].items():assert sha(OUT/'preparation/skills'/name)==h,name
assert sha(OUT/'preparation/host-skill-disable-followed-override.txt')==m['host_disable_sha256']
assert m['authors_spent_total']==3 and m['group_authors_spent_total']==19 and len(m['attempts'])==3 and all(a['counted'] for a in m['attempts'])
for row in m['attempts']:
 archive=OUT/row['directory'];work=Path(tempfile.mkdtemp(prefix='go-integrity-',dir='/private/tmp'));shutil.copytree(EVALS/'files'/row['case'],work,dirs_exist_ok=True)
 result=subprocess.run(['rtk','proxy','git','apply',str(archive/'source.patch')],cwd=work,capture_output=True,text=True);assert result.returncode==0,result.stderr
 hashes={str(p.relative_to(work)):sha(p) for p in work.rglob('*') if p.is_file()};assert hashes==json.loads((archive/'source-hashes.json').read_text()),row['number'];shutil.rmtree(work)
 v=json.loads((archive/'verification.json').read_text());assert v['reconstruction_hashes_match'] and all(c['status']==0 for c in v['ordinary_checks']);assert not v['ordinary_checks'][6]['stdout'].strip()
 expected=[1,1,1] if row['case']=='service-owned-workers' else [0,0] if row['case']=='not-concurrency-work' else [0,0,0];assert [c['status'] for c in v['held_contract_checks']]==expected
for group in ('library','service','control'):
 freeze=json.loads((OUT/'neutral-reviews'/(group+'-freeze.json')).read_text());assert freeze['frozen_before_arm_disclosure']
 for name,h in freeze['hashes'].items():assert sha(OUT/'neutral-reviews'/group/name)==h,(group,name)
assert m.get('discovery_disposition') in ('draft-justified','stop-before-draft')
if (OUT/'checksums.json').exists():
 seal=json.loads((OUT/'checksums.json').read_text());actual={str(p.relative_to(OUT)):sha(p) for p in OUT.rglob('*') if p.is_file() and p!=OUT/'checksums.json'};assert actual==seal['sha256'],'Archive seal mismatch'
print('PASS: frozen 23/17/eight-file identities, three reconstructed outcomes, declared service failure and independently frozen reviews'+(' plus seal' if (OUT/'checksums.json').exists() else ' (not yet sealed)'))
