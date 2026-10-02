import json
import os
from pathlib import Path
import subprocess
import time

out=Path('/private/tmp/go-behavior-r2-promotion')
checks=json.loads((out/'checks.json').read_text())
env={'GOCACHE':str(out/'cache'),'GOTOOLCHAIN':'local'}
targets=[('2026-10-01-behavior-tests-r2','service-publication'),('2026-10-01-testing-combined-final','behavior-only'),('2026-10-01-testing-combined-final','both')]
for run,case in targets:
    cwd=out/'reconstructed'/run/case
    cmd=['rtk','proxy','go','test','-count=1','-timeout=45s','./...']
    start=time.monotonic()
    p=subprocess.run(cmd,cwd=cwd,env={**os.environ,**env},capture_output=True,text=True,timeout=60)
    result={'kind':'approved-unchanged-source-loopback-rerun','command':cmd,'cwd':str(cwd),'environment_overrides':env,'stdout':p.stdout,'stderr':p.stderr,'exit_code':p.returncode,'duration_seconds':round(time.monotonic()-start,3)}
    checks['commands'].append(result)
    (out/'checks.json').write_text(json.dumps(checks,indent=2)+'\n')
    print(json.dumps(result),flush=True)
