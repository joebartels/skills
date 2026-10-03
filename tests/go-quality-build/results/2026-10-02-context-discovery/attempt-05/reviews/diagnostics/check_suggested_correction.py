from pathlib import Path
import json
import os
import shutil
import subprocess
import time

root=Path('/private/tmp/packet-02-review-diagnostics')
fixed=root/'snapshot-correction'
shutil.copytree(root/'snapshot-stress',fixed,dirs_exist_ok=True)
path=fixed/'stages.go'
old='\treturn errors.Join(operationErr, ctx.Err(), context.Cause(ctx))'
new='\tcancelErr := ctx.Err()\n\tif cancelErr == nil {\n\t\treturn operationErr\n\t}\n\treturn errors.Join(operationErr, cancelErr, context.Cause(ctx))'
assert path.read_text().count(old)==1
path.write_text(path.read_text().replace(old,new))
env=dict(os.environ,GOWORK='off',GOTOOLCHAIN='local',GOCACHE='/private/tmp/packet-02-go-cache')
argv=['rtk','proxy','go','test','-mod=readonly','-run','TestReview|TestSequential|TestAlready|TestParent|TestTotal|TestFailure|TestStatus|TestEmpty','-count=1','-timeout=30s','./...']
started=time.monotonic()
result=subprocess.run(argv,cwd=fixed,env=env,text=True,capture_output=True,timeout=40)
record=dict(label='suggested-cancellation-snapshot-correction',cwd=str(fixed),argv=argv,environment_overrides={k:env[k] for k in ['GOWORK','GOTOOLCHAIN','GOCACHE']},status=result.returncode,stdout=result.stdout,stderr=result.stderr,elapsed_seconds=round(time.monotonic()-started,3))
(root/'suggested-correction-check.json').write_text(json.dumps(record,indent=2)+'\n')
print(result.stdout,result.stderr,flush=True)
print('status='+str(result.returncode),flush=True)
