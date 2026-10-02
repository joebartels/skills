import concurrent.futures
import hashlib
import json
import os
from pathlib import Path
import resource
import shutil
import signal
import subprocess
import time

OUT=Path('/private/tmp/go-behavior-r2-promotion')
REPO=Path('/Users/jb/.codex/worktrees/b633/skills')
ROOT=REPO/'tests/go-quality-build/results'
ENV={'GOCACHE':str(OUT/'cache'),'GOTOOLCHAIN':'local'}
CHECKS=json.loads((OUT/'checks.json').read_text())

def cmd(args,cwd,overrides=None,unset=()):
    env={**os.environ,**ENV,**(overrides or {})}
    for k in unset:env.pop(k,None)
    start=time.monotonic()
    try:
        p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env=env,capture_output=True,text=True,timeout=60)
        return {'command':['rtk','proxy',*args],'cwd':str(cwd),'environment_overrides':{**ENV,**(overrides or {})},'unset_environment':list(unset),'stdout':p.stdout,'stderr':p.stderr,'exit_code':p.returncode,'duration_seconds':round(time.monotonic()-start,3)}
    except subprocess.TimeoutExpired as e:
        return {'command':['rtk','proxy',*args],'cwd':str(cwd),'environment_overrides':{**ENV,**(overrides or {})},'unset_environment':list(unset),'stdout':str(e.stdout),'stderr':str(e.stderr),'exit_code':None,'timeout_seconds':60}

def save():
    (OUT/'checks.json').write_text(json.dumps(CHECKS,indent=2)+'\n')

def copy(run,case,suffix):
    path=OUT/'signals'/run/case/suffix
    if path.exists():shutil.rmtree(path)
    shutil.copytree(ROOT/run/case/'candidate',path)
    return path

def mutation(patch):
    case=patch.parents[2].name;run=patch.parents[3].name
    work=copy(run,case,patch.parent.name)
    steps=[cmd(['git','apply',str(patch)],work),cmd(['go','test','-run','^$','-timeout=45s','./...'],work),cmd(['go','test','-count=1','-timeout=45s','./...'],work)]
    return {'run':run,'case':case,'semantic_id':patch.parent.name,'controller_probes_absent':True,'patch_sha256':hashlib.sha256(patch.read_bytes()).hexdigest(),'steps':steps,'assertion_lines':[line for line in steps[-1]['stdout'].splitlines() if '.go:' in line], 'result':'survived' if steps[-1]['exit_code']==0 else 'inspect assertions'}

patches=[]
for run in ['2026-10-01-behavior-tests-r2','2026-10-01-testing-combined-final']:
    patches.extend(sorted((ROOT/run).glob('*/mutations/*/source.patch')))
CHECKS['fresh_mutations']=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    futures=[pool.submit(mutation,p) for p in patches]
    for future in concurrent.futures.as_completed(futures):
        item=future.result();CHECKS['fresh_mutations'].append(item);save()
        print(json.dumps({'run':item['run'],'case':item['case'],'semantic_id':item['semantic_id'],'exits':[x['exit_code'] for x in item['steps']],'assertion_lines':item['assertion_lines'][:3]}),flush=True)

# Baseline callback-error/successful-release survivor, tested using candidate tests only.
run='2026-10-01-behavior-tests-baseline';case='worker-host'
work=copy(run,case,'error-loss-survivor')
patch=ROOT/run/case/'mutations/post-review-host-drop-error-successful-release/source.patch'
steps=[cmd(['git','apply',str(patch)],work),cmd(['go','test','-run','^$','-timeout=45s','./...'],work),cmd(['go','test','-count=1','-timeout=45s','./...'],work)]
CHECKS['baseline_worker_survivor']={'candidate_tests_only':True,'steps':steps}
print('baseline worker survivor exits',[x['exit_code'] for x in steps],flush=True);save()

# The earlier integrated both suite's accepted-set survivor. Do not inject its reviewer probe.
run='2026-10-01-testing-combined-r2';case='both'
work=copy(run,case,'accepted-set-survivor')
source=work/'refresh.go';old=source.read_text();assert old.count('response.StatusCode != http.StatusOK')==1
source.write_text(old.replace('response.StatusCode != http.StatusOK','response.StatusCode < 200 || response.StatusCode >= 300'))
steps=[cmd(['go','test','-run','^$','-timeout=45s','./...'],work),cmd(['go','test','-count=1','-timeout=45s','./...'],work)]
CHECKS['preceding_accepted_set_survivor']={'candidate_tests_only':True,'exact_change':'response.StatusCode != http.StatusOK -> response.StatusCode < 200 || response.StatusCode >= 300','steps':steps}
print('preceding accepted-set survivor exits',[x['exit_code'] for x in steps],flush=True);save()

# Current implementations and the separately classified arbitrary-comparison mutants under the legal controller probe.
CHECKS['legal_cause_crosschecks']=[]
for case in ['behavior-only','both']:
    run='2026-10-01-testing-combined-final'
    for mutant in [False,True]:
        work=copy(run,case,'cause-mutant-with-probe' if mutant else 'legal-cause-current')
        steps=[]
        if mutant:steps.append(cmd(['git','apply',str(ROOT/run/case/'mutations/unsafe-cancellation-cause-comparison/source.patch')],work))
        probe=ROOT/run/'cause-crosscheck'/f'{run}--{case}'/'diagnostic_cause_test.go'
        shutil.copy2(probe,work/'diagnostic_cause_test.go')
        steps += [cmd(['go','test','-run','^$','-timeout=45s','./...'],work),cmd(['go','test','-run','^TestDiagnosticNonComparableCause$','-count=1','-timeout=45s','./...'],work)]
        item={'case':case,'mutant':mutant,'classification':'legal-input controller verification, excluded from candidate-only mutation credit','probe_sha256':hashlib.sha256(probe.read_bytes()).hexdigest(),'steps':steps}
        CHECKS['legal_cause_crosschecks'].append(item);save();print('cause crosscheck',case,mutant,[x['exit_code'] for x in steps],flush=True)

# Cache-default portability checks use unmodified compiled command test binaries.
home=OUT/'default-home';home.mkdir(exist_ok=True)
CHECKS['default_cache_checks']=[]
for case in ['behavior-only','both']:
    work=OUT/'reconstructed/2026-10-01-testing-combined-final'/case
    binary=OUT/f'{case}-command.test'
    steps=[cmd(['go','test','-c','-o',str(binary),'./cmd/indexer'],work),cmd([str(binary),'-test.count=1','-test.timeout=45s'],work/'cmd/indexer',overrides={'HOME':str(home)},unset=['GOCACHE'])]
    item={'case':case,'steps':steps};CHECKS['default_cache_checks'].append(item);save();print('default cache',case,[x['exit_code'] for x in steps],flush=True)

# Reproduce the baseline's accepted-value destruction with a resource limit confined to a child.
run='2026-10-01-behavior-tests-baseline';case='cli-partial-failure'
work=OUT/'reconstructed'/run/case;binary=OUT/'baseline-ledgerload'
steps=[cmd(['go','build','-o',str(binary),'.'],work)]
destination=OUT/'baseline-partial-write';destination.mkdir(exist_ok=True)
def child_limits():
    resource.setrlimit(resource.RLIMIT_FSIZE,(3,3))
    signal.signal(signal.SIGXFSZ,signal.SIG_IGN)
start=time.monotonic()
p=subprocess.run(['rtk','proxy',str(binary),'--dir',str(destination)],input='web,1\nweb,12345\nlater,9\n',text=True,capture_output=True,preexec_fn=child_limits,timeout=10)
steps.append({'command':['rtk','proxy',str(binary),'--dir',str(destination)],'cwd':str(REPO),'stdin':'web,1\nweb,12345\nlater,9\n','child_only_limit':{'RLIMIT_FSIZE':[3,3],'SIGXFSZ':'ignored'},'stdout':p.stdout,'stderr':p.stderr,'exit_code':p.returncode,'duration_seconds':round(time.monotonic()-start,3)})
CHECKS['baseline_cli_partial_write']={'steps':steps,'retained_web_bytes':(destination/'web').read_text(),'later_exists':(destination/'later').exists()}
print('baseline partial write',p.returncode,(destination/'web').read_text(),(destination/'later').exists(),flush=True)
save()
