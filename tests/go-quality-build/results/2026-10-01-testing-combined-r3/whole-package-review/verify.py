from pathlib import Path
import subprocess,json,os,shutil,time,re
root=Path('/Users/jb/.codex/worktrees/b633/skills');out=Path('/private/tmp/go-testing-final-package-review');records=[]
env=dict(os.environ,GOTOOLCHAIN='local',GOCACHE=str(out/'cache'));env['PYTHONDONTWRITEBYTECODE']='1'
def run(args,cwd=root,label=None,custom=None):
 start=time.monotonic()
 try:
  p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env=custom or env,text=True,capture_output=True,timeout=180)
  rec=dict(label=label or ' '.join(args),command=['rtk','proxy',*args],cwd=str(cwd),exit_code=p.returncode,stdout=p.stdout,stderr=p.stderr,seconds=time.monotonic()-start)
 except subprocess.TimeoutExpired as e: rec=dict(label=label,exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
 records.append(rec);(out/'checks.json').write_text(json.dumps(records,indent=2));print(rec['label'],rec['exit_code'],flush=True);return rec
run(['git','status','--porcelain=v1'],label='initial status')
run(['python3','-B',str(root/'tests/go-quality-build/results/2026-10-01-testing-combined-r3/controller/delivery/verify_testing_evidence.py'),str(out/'integrity.json')],label='all evidence reconstruction')
for args in [['-m','unittest','discover','-s','tests','-p','test*.py'],['-m','unittest','discover','-s','tests/go-quality-build','-p','test_*.py'],['tests/go-quality-review/test_layout.py'],['tests/go-quality-review/go-quality-report/evals/test_grade.py'],['scripts/validate.py']]:run(['python3','-B',*args])
run(['git','diff','--check','b76df28ed15a177d436672d47a5d30f6634900c7','f48c68ddbdd31931cad3522aa6f9f1f17ea86d06','--','.',':(exclude)**/source.patch',':(exclude)tests/go-quality-build/results/2026-10-01-testing-combined/both/blind-review.md',':(exclude)tests/go-quality-build/results/2026-10-01-testing-combined/both/review-artifacts/inspection-transcript.json'],label='whole range whitespace with exact raw exclusions')
run(['go','version'])
for suite,revision,case in [('go-behavior-tests','behavior-tests-r2','library-codec'),('go-behavior-tests','behavior-tests-r2','cli-partial-failure'),('go-test-isolation','test-isolation-r3','library-file-fixtures'),('go-test-isolation','test-isolation-r3','service-http-boundary'),('go-test-isolation','test-isolation-r3','worker-timing')]:
 source=root/'tests/go-quality-build/results'/('2026-10-01-'+revision)/case/'candidate';dest=out/case
 shutil.copytree(source,dest,dirs_exist_ok=True)
 run(['go','test','-count=1','-timeout=40s','./...'],dest,label=case+' candidate suite')
 probe=root/'tests/go-quality-build'/suite/'evals/controller-probes'/case/'contract_test.go'
 (dest/'review_contract_test.go').write_text(re.sub(r'func Test', 'func TestFinalReview',probe.read_text()))
 run(['go','test','-count=1','-timeout=40s','./...'],dest,label=case+' held focus probes')
 if case in ['library-file-fixtures','worker-timing']:run(['go','test','-race','-shuffle=on','-count=3','-timeout=40s','./...'],dest,label=case+' race shuffle repeat')
run(['git','status','--porcelain=v1'],label='final status')
