import json
import os
import pathlib
import subprocess
import time

base = pathlib.Path(__file__).resolve().parents[1]
log = base/'output'/'listener-checks.json'
records = []
env = dict(os.environ, GOCACHE='/private/tmp/go-quality-testing-cache', GOTOOLCHAIN='local', GOPROXY='off')
for label, flags, cwd in [
    ('approved unchanged-source full suite with listeners', ['-count=1','-timeout=40s'], base/'candidate'),
    ('approved unchanged-source race shuffled repeated with listeners', ['-race','-count=10','-shuffle=on','-timeout=60s'], base/'candidate'),
]:
    command = ['rtk','proxy','go','test','./...']+flags
    started = time.time()
    result = subprocess.run(command, cwd=cwd, env=env, capture_output=True, text=True, timeout=80)
    record = dict(label=label, argv=command, cwd=str(cwd), env={k:env[k] for k in ['GOCACHE','GOTOOLCHAIN','GOPROXY']}, stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, elapsed_seconds=round(time.time()-started,3), sandbox_permissions='require_escalated')
    records.append(record)
    log.write_text(json.dumps(records, indent=2)+'\n')
    print(json.dumps(record), flush=True)
