from pathlib import Path
import subprocess, os, json, shutil, time
from concurrent.futures import ThreadPoolExecutor
root=Path('/private/tmp/go-context-outcome-20261002/service')
reviews=root/'reviews'
env=os.environ.copy()
env.update(GOCACHE=str(reviews/'go-cache'), GOWORK='off', GOTOOLCHAIN='local', GOPROXY='off')
records=[]
def run(n,label,cmd,cwd):
    start=time.monotonic()
    p=subprocess.run(['rtk','proxy']+cmd,cwd=cwd,env=env,text=True,capture_output=True,timeout=70)
    r=dict(label=label,argv=['rtk','proxy']+cmd,cwd=str(cwd),environment_overrides={k:env[k] for k in ('GOCACHE','GOWORK','GOTOOLCHAIN','GOPROXY')},status=p.returncode,stdout=p.stdout,stderr=p.stderr,elapsed_seconds=round(time.monotonic()-start,3))
    dest=reviews/f'candidate-{n}'/'evidence'
    (dest/f'{label}.json').write_text(json.dumps(r,indent=2)+'\n')
    (dest/f'{label}.log').write_text('$ '+' '.join(r['argv'])+'\ncwd: '+str(cwd)+'\nexit: '+str(p.returncode)+'\nstdout:\n'+p.stdout+'\nstderr:\n'+p.stderr)
    return n,label,p.returncode,p.stdout,p.stderr
jobs=[]
for n in (1,2):
    author=reviews/'scratch'/f'author-{n}'
    shutil.copytree(root/f'candidate-{n}',author,dirs_exist_ok=True)
    contract=reviews/'scratch'/f'candidate-{n}'
    shutil.copy(reviews/'diagnostic_test.go.txt',contract/'review_diagnostic_test.go')
    jobs.extend([(n,'version',['go','version'],author),(n,'build',['go','build','-mod=readonly','./...'],author),(n,'author-tests',['go','test','-count=1','-timeout=30s','./...'],author),(n,'vet',['go','vet','./...'],author),(n,'format',['gofmt','-l','.'],author),(n,'contract-diagnostics',['go','test','-run','^(TestReview|TestTotalAndEarlierParentDeadline|TestResponseScopeAndClose|TestStageBudgetAndCanceledAdmission|TestHTTPCancellationBoundary)','-count=1','-timeout=30s','./...'],contract),(n,'supplementary',['go','test','-run','^TestSupplementary','-count=1','-timeout=30s','./...'],contract),(n,'contract-race',['go','test','-race','-shuffle=on','-run','^(TestReview|TestTotalAndEarlierParentDeadline|TestResponseScopeAndClose|TestStageBudgetAndCanceledAdmission|TestHTTPCancellationBoundary)','-count=3','-timeout=60s','./...'],contract)])
    mutations=[('lost-total-budget', 'for _, endpoint := range endpoints {\n','for _, endpoint := range endpoints {\n\t\tvar mutCancel context.CancelFunc\n\t\ttotalCtx, mutCancel = context.WithTimeout(ctx, total)\n\t\tdefer mutCancel()\n')]
    if n==1:
        mutations.extend([('lost-stage-budget','context.WithTimeout(ctx, budget)','context.WithTimeout(ctx, time.Hour)'),('lost-close-error-on-read','return nil, failedWithContext(stageCtx, errors.Join(err, closeErr))','if err != nil { return nil, failedWithContext(stageCtx, err) }; return nil, failedWithContext(stageCtx, errors.Join(err, closeErr))')])
    for label,old,new in mutations:
        d=reviews/'scratch'/f'{label}-{n}'
        shutil.copytree(author,d,dirs_exist_ok=True)
        f=d/'stages.go'; text=f.read_text()
        assert text.count(old)==1,(label,text.count(old))
        f.write_text(text.replace(old,new,1))
        jobs.extend([(n,label+'-compile',['go','test','-run','^$','./...'],d),(n,label+'-author',['go','test','-count=1','-timeout=30s','./...'],d)])
with ThreadPoolExecutor(max_workers=4) as pool:
    futures=[pool.submit(run,*job) for job in jobs]
    for f in futures:
        n,label,status,out,err=f.result()
        print(json.dumps(dict(candidate=n,label=label,status=status,stdout=out,stderr=err)),flush=True)
