from pathlib import Path
import hashlib,json,os,re,shutil,subprocess,tempfile,time,sys
root=Path.cwd();suite,runid,case=sys.argv[1:4];run=root/'tests/go-quality-build/results'/runid;arc=run/case;m=json.loads((run/'manifest.json').read_text());entry=next(c for c in m['cases'] if c['id']==case)
before=arc/'verification-before-controller-namespace.json';assert not before.exists();shutil.copy2(arc/'verification.json',before);old=json.loads(before.read_text());assert old[-1]['kind']=='withheld-contract-probes' and old[-1]['exit_code']!=0 and 'redeclared' in old[-1]['stderr']
probe=root/'tests/go-quality-build'/suite/'evals/controller-probes'/entry.get('fixture_id',case);out=arc/'controller-namespace';out.mkdir();records=[]
with tempfile.TemporaryDirectory(prefix='go-namespaced-probe-',dir='/private/tmp') as tmp:
 work=Path(tmp)/'module';shutil.copytree(arc/'candidate',work)
 for p in probe.rglob('*'):
  if not p.is_file():continue
  raw=p.read_text();digest=hashlib.sha256(p.read_bytes()).hexdigest();mapping={name:'TestController'+digest[:8]+name[4:] for name in re.findall(r'^func (Test\w+)\(',raw,re.M)}
  changed=re.sub(r'^func (Test\w+)\(',lambda x:'func '+mapping[x.group(1)]+'(',raw,flags=re.M)
  restored=changed
  for name,new in mapping.items():restored=restored.replace('func '+new+'(', 'func '+name+'(')
  assert restored==raw
  (out/p.name).write_text(changed);(work/('controller_'+p.name)).write_text(changed)
  records.append(dict(frozen_path=str(p.relative_to(root)),frozen_sha256=digest,renamed_sha256=hashlib.sha256(changed.encode()).hexdigest(),identifier_map=mapping,semantic_equivalence='exact bytes restored by inverse top-level test function name mapping; bodies unchanged'))
 checks=[]
 for args in [['go','test','-run','^$','-timeout=30s','./...'],['go','test','-count=1','-timeout=30s','./...']]:
  cmd=['rtk','proxy',*args];start=time.monotonic();r=subprocess.run(cmd,cwd=work,env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local'),capture_output=True,text=True,timeout=55);checks.append(dict(command=cmd,cwd=str(work),env={'GOCACHE':'/private/tmp/go-quality-testing-cache','GOTOOLCHAIN':'local'},stdout=r.stdout,stderr=r.stderr,exit_code=r.returncode,duration_seconds=time.monotonic()-start,kind='withheld-contract-probes-namespaced'));assert r.returncode==0,(r.stdout,r.stderr)
 (out/'verification.json').write_text(json.dumps(dict(reason='controller/author test-name collision, not candidate failure or semantic probe change',mapping=records,checks=checks),indent=2)+'\n')
 (arc/'verification.json').write_text(json.dumps(old[:-1]+checks,indent=2)+'\n')
print('PASS unchanged-source controller probe compilation and assertions after namespace-only injection; raw collision retained')
