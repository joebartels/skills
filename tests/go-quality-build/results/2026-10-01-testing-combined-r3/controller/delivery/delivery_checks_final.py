from pathlib import Path
import json,subprocess,sys,time
root=Path.cwd(); out=root/'tests/go-quality-build/results'/sys.argv[1]/'delivery-checks.json'
commands=[['rtk','proxy','python3','-B','-m','unittest','discover','-s','tests','-p','test*.py'],['rtk','proxy','python3','-B','-m','unittest','discover','-s','tests/go-quality-build','-p','test_*.py'],['rtk','proxy','python3','-B','tests/go-quality-review/test_layout.py'],['rtk','proxy','python3','-B','tests/go-quality-review/go-quality-report/evals/test_grade.py'],['rtk','proxy','python3','-B','scripts/validate.py'],['rtk','claude','plugin','validate','./plugins/go-quality-build'],['rtk','claude','plugin','validate','.'],['rtk','proxy','python3','-B','/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py','plugins/go-quality-build/skills/go-test-isolation'],['rtk','proxy','python3','-B','/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py','plugins/go-quality-build/skills/go-behavior-tests'],['rtk','proxy','git','diff','--check','--','.',' :(exclude)**/source.patch'.strip()],['rtk','proxy','git','diff','--cached','--check','--','.',' :(exclude)**/source.patch'.strip()]]
records=[]
for cmd in commands:
 start=time.monotonic(); r=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=55)
 records.append(dict(command=cmd,cwd=str(root),stdout=r.stdout,stderr=r.stderr,exit_code=r.returncode,duration_seconds=time.monotonic()-start))
 print(r.returncode,' '.join(cmd)); print((r.stdout+r.stderr).strip())
out.write_text(json.dumps(dict(scope='Task 6 final package structural verification; raw source.patch evidence excluded from whitespace check',checks=records),indent=2)+'\n')
assert all(r['exit_code']==0 for r in records)
