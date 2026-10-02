"""Reconstruct and capture checks once before sealing; source outcomes remain unchanged."""
from pathlib import Path
import concurrent.futures,hashlib,json,os,shutil,subprocess,sys,tempfile,time
ROOT=Path(__file__).resolve().parents[4];OUT=Path(__file__).resolve().parent;EVALS=ROOT/'tests/go-quality-build/go-concurrency-and-ownership/evals'
MIN='/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
m=json.loads((OUT/'manifest.json').read_text());rows={a['number']:a for a in m['attempts']};numbers=[int(x) for x in sys.argv[1:]];assert numbers and all(rows[n]['status']=='completed' for n in numbers)
ENV={'GOCACHE':'/private/tmp/go-quality-concurrency-cache','GOTOOLCHAIN':'local'}
def run(argv,cwd,extra=None):
 env=os.environ.copy();env.update(ENV);env.update(extra or {});start=time.monotonic()
 try:
  p=subprocess.run(['rtk','proxy']+argv,cwd=cwd,env=env,capture_output=True,text=True,timeout=90);status=p.returncode;stdout=p.stdout;stderr=p.stderr
 except subprocess.TimeoutExpired as e:status=124;stdout=(e.stdout or b'').decode() if isinstance(e.stdout,bytes) else e.stdout or '';stderr=(e.stderr or b'').decode() if isinstance(e.stderr,bytes) else e.stderr or '';stderr+='\nController subprocess watchdog expired.'
 return {'argv':['rtk','proxy']+argv,'cwd':str(cwd),'environment_overrides':ENV|dict(extra or {}),'status':status,'stdout':stdout,'stderr':stderr,'elapsed_seconds':round(time.monotonic()-start,3)}
def verify(n):
 row=rows[n];case=row['case'];archive=OUT/row['directory'];work=Path(tempfile.mkdtemp(prefix='go-check-',dir='/private/tmp'));shutil.copytree(EVALS/'files'/case,work,dirs_exist_ok=True)
 reconstruction=run(['git','apply',str(archive/'source.patch')],work);assert reconstruction['status']==0,reconstruction
 hashes={str(p.relative_to(work)):hashlib.sha256(p.read_bytes()).hexdigest() for p in work.rglob('*') if p.is_file()};assert hashes==json.loads((archive/'source-hashes.json').read_text()),n
 ordinary=[reconstruction]
 for argv in [['go','version'],[MIN,'version'],['go','build','-o','/dev/null','./...'],['go','test','-count=1','-timeout=30s','./...'],['go','vet','./...'],['gofmt','-l','.'],[MIN,'test','-count=1','-timeout=30s','./...']]:ordinary.append(run(argv,work,{'GOQUALITY_GO':MIN} if argv[0]==MIN else {}))
 heldpath=work/('cmd/pipeline' if case=='cli-pipeline-stop' else '')/'controller_contract_test.go';shutil.copyfile(EVALS/'controller-probes'/case/'contract_test.go',heldpath)
 held=[run(['go','test','-count=1','-timeout=30s','./...'],work),run([MIN,'test','-count=1','-timeout=30s','./...'],work,{'GOQUALITY_GO':MIN})]
 if case!='not-concurrency-work':held.append(run(['go','test','-race','-shuffle=on','-count=3','-timeout=60s','./...'],work))
 supplementary=[]
 if case=='service-owned-workers':
  shutil.copyfile(OUT/'preparation/supplementary-service-probes.go.txt',work/'supplementary_contract_test.go')
  supplementary=[run(['go','test','-run','^TestReview(AvailableJobWithOpenInput|RunFailureCancelsPendingOpen)$','-count=1','-timeout=30s','./...'],work),run([MIN,'test','-run','^TestReview(AvailableJobWithOpenInput|RunFailureCancelsPendingOpen)$','-count=1','-timeout=30s','./...'],work)]
 result={'supplementary_review_discovered_checks':supplementary,'supplementary_limit':'Not a replacement for four original frozen mutation targets or primary frozen held observation.','reconstruction_hashes_match':True,'ordinary_checks':ordinary,'held_contract_checks':held,'mutations':'Concrete author-specific mutation sensitivity recorded separately; no controller repair credit.'}
 (archive/'verification.json').write_text(json.dumps(result,indent=2)+'\n')
 print(n,case,'ordinary',[c['status'] for c in ordinary],'held',[c['status'] for c in held],flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:list(pool.map(verify,numbers))
