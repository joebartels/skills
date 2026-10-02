from pathlib import Path
import json
import os
import shutil
import subprocess
import time

ROOT=Path('/private/tmp/packet-02-review-diagnostics')
ENV=dict(os.environ,GOWORK='off',GOTOOLCHAIN='local',GOCACHE='/private/tmp/packet-02-go-cache')
records=json.loads((ROOT/'executed-checks.json').read_text())
def run(label,cwd,argv):
    started=time.monotonic()
    result=subprocess.run(argv,cwd=cwd,env=ENV,capture_output=True,text=True,timeout=45)
    records.append(dict(label=label,cwd=str(cwd),argv=argv,environment_overrides={k:ENV[k] for k in ['GOWORK','GOTOOLCHAIN','GOCACHE']},status=result.returncode,stdout=result.stdout,stderr=result.stderr,elapsed_seconds=round(time.monotonic()-started,3)))
    (ROOT/'executed-checks.json').write_text(json.dumps(records,indent=2)+'\n')
    print(label,'status='+str(result.returncode),result.stdout.strip(),result.stderr.strip(),flush=True)
for label in ['boundary','restart-total-each-stage','cancel-before-body','ignore-close-only-failure','invalidate-complete-success']:
    run(label+'-focused-no-listener',ROOT/label,['rtk','proxy','go','test','-mod=readonly','-run','^TestReview','-skip','StandardTransport','-count=1','-timeout=30s','./...'])
stress=ROOT/'snapshot-stress'
shutil.copytree(ROOT/'baseline',stress,dirs_exist_ok=True)
shutil.copy(ROOT/'stress_snapshot_test.go',stress/'stress_snapshot_test.go')
run('cancellation-snapshot-stress',stress,['rtk','proxy','go','test','-mod=readonly','-run','^TestReviewCancellationSnapshot$','-count=1','-timeout=30s','./...'])
