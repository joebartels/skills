from pathlib import Path
import json,hashlib,subprocess,tempfile,shutil,sys
root=Path.cwd();results=root/'tests/go-quality-build/results';records=[]
run_suites={'2026-10-01-behavior-tests-baseline':'go-behavior-tests','2026-10-01-behavior-tests-skill-on':'go-behavior-tests','2026-10-01-test-isolation-baseline':'go-test-isolation','2026-10-01-test-isolation-skill-on':'go-test-isolation','2026-10-01-testing-combined':'testing-combined','2026-10-01-test-isolation-r2':'go-test-isolation','2026-10-01-testing-combined-r2':'testing-combined','2026-10-01-behavior-tests-r2':'go-behavior-tests','2026-10-01-testing-combined-final':'testing-combined','2026-10-01-test-isolation-r3':'go-test-isolation','2026-10-01-testing-combined-r3':'testing-combined'}
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def hashes(d):return {str(p.relative_to(d)):sha(p) for p in sorted(d.rglob('*')) if p.is_file()}
for name,suite in run_suites.items():
 if "--individual-only" in sys.argv and suite=="testing-combined":continue
 run=results/name;m=json.loads((run/'manifest.json').read_text());
 if '--completed-only' in sys.argv and not all(c.get('completed',False) for c in m['cases']):continue
 detail={'run':name,'cases':[]}
 if (run/'checksums.sha256').exists():
  listed={}
  for line in (run/'checksums.sha256').read_text().splitlines():digest,file=line.split('  ',1);assert sha(run/file)==digest,(name,file);listed[file]=digest
  assert set(listed)==set(hashes(run))-{'checksums.sha256'},(name,'incomplete index')
  detail.update(artifact_count=len(listed),checksum_index_sha256=sha(run/'checksums.sha256'))
 else:detail['seal']='not yet sealed; no archive-integrity completion claim'
 for f,h in m['skill_sha256'].items():assert sha(run/'skills'/f)==h,(name,'skill',f)
 for c in m['cases']:
  arc=run/c['id'];assert hashes(arc/'original')==c['input_sha256'];assert hashes(arc/'candidate')==c['source_sha256'];assert sha(arc/'dispatch.txt')==c['dispatch_sha256']
  for f,h in c['input_sha256'].items():
   path=f'tests/go-quality-build/{suite}/evals/files/{c.get("fixture_id",c["id"])}/{f}'
   b=subprocess.check_output(['rtk','proxy','git','show',m['fixture_commit']+':'+path]);assert hashlib.sha256(b).hexdigest()==h,(name,'frozen input',path)
  with tempfile.TemporaryDirectory(prefix='go-final-reconstruct-',dir='/private/tmp') as tmp:
   dest=Path(tmp)/'module';shutil.copytree(arc/'original',dest)
   if (arc/'source.patch').stat().st_size:
    r=subprocess.run(['rtk','proxy','git','apply',str(arc/'source.patch')],cwd=dest,capture_output=True,text=True);assert r.returncode==0,(name,c['id'],r.stderr)
   assert hashes(dest)==c['source_sha256'],(name,c['id'],'reconstruction')
  verification=json.loads((arc/'verification.json').read_text());assert all(v.get('exit_code',0)==0 for v in verification),(name,c['id'],'failed candidate checks')
  detail['cases'].append({'id':c['id'],'files':len(c['source_sha256']),'reconstruction':'exact','input_commit':'verified','direct_candidate_checks':'all recorded exits 0'})
 records.append(detail)
for name,run in [('go-behavior-tests','2026-10-01-behavior-tests-skill-on'),('go-test-isolation','2026-10-01-test-isolation-skill-on')]:
 if name=='go-test-isolation':
  expected=results/'2026-10-01-test-isolation-r3/skills'/name
  current=root/('docs/go-quality-build/drafts' if '--draft-r3' in sys.argv and (root/'docs/go-quality-build/drafts'/name).exists() else 'plugins/go-quality-build/skills')/name
 else:
  expected=results/'2026-10-01-behavior-tests-r2/skills'/name
  current=root/('docs/go-quality-build/drafts' if '--draft-r3' in sys.argv and (root/'docs/go-quality-build/drafts'/name).exists() else 'plugins/go-quality-build/skills')/name
 assert hashes(current)==hashes(expected),(name,'final guidance snapshot differs')
if len(sys.argv)>1:Path(sys.argv[1]).write_text(json.dumps({'scope':'all selected testing-group blind candidates; exact reconstruction, frozen inputs, guidance and available archive seals','checks':records},indent=2)+'\n')
print('PASS',sum(len(r['cases']) for r in records),'exact reconstructions;',sum(r.get('artifact_count',0) for r in records),'sealed artifact hashes; final guidance snapshot equality'+(' (draft; runtime promotion pending)' if '--draft-r3' in sys.argv else ' (runtime)'))
