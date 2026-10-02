from pathlib import Path
import hashlib,json,os,shutil,subprocess,tempfile,time
root=Path.cwd();suite=root/'tests/go-quality-build/testing-combined/evals';run=root/'tests/go-quality-build/results/2026-10-01-testing-combined';run.mkdir(parents=True,exist_ok=True)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def hashes(d):return {str(p.relative_to(d)):sha(p) for p in sorted(d.rglob('*')) if p.is_file()}
checks=[]
def check(args,cwd,kind):
 env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local');start=time.monotonic();r=subprocess.run(['rtk','proxy',*args],cwd=cwd,env=env,capture_output=True,text=True,timeout=55)
 item=dict(kind=kind,command=['rtk','proxy',*args],cwd=str(cwd),env={'GOCACHE':env['GOCACHE'],'GOTOOLCHAIN':'local'},stdout=r.stdout,stderr=r.stderr,exit_code=r.returncode,duration_seconds=time.monotonic()-start);checks.append(item);print(kind,r.returncode,(r.stdout+r.stderr).strip());return r
case=suite/'files/indexer-evolution'
for cmd in [['go','test','-count=1','-timeout=30s','./...'],['go','vet','./...'],['gofmt','-l','.']]:
 r=check(cmd,case,'starting');assert r.returncode==0;assert cmd[0]!='gofmt' or not r.stdout
with tempfile.TemporaryDirectory(prefix='indexer-preparation-',dir='/private/tmp') as tmp:
 work=Path(tmp)/'module';shutil.copytree(case,work);shutil.copy2(suite/'controller-probes/indexer-evolution/contract_test.go',work/'controller_contract_test.go')
 assert check(['go','test','-run','^$','./...'],work,'probe-compilation').returncode==0
 r=check(['go','test','-count=1','-timeout=30s','./...'],work,'new-contract-RED');assert r.returncode!=0;assert 'operation not implemented' in r.stdout and 'process prefix=' in r.stdout
out=dict(input_sha256=hashes(case),controller_sha256=hashes(suite/'controller-probes'),checks=checks,distinction='old behavior passes; requested operations are compile-valid placeholders; new held probes fail meaningful assertions; no shipped production TDD claim',host='go1.26.5 darwin/arm64; actual Go1.22 unavailable')
(run/'preparation.json').write_text(json.dumps(out,indent=2)+'\n')
