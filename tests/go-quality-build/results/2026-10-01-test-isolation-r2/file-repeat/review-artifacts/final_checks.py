import hashlib
import json
import os
import pathlib
import subprocess
import time

base = pathlib.Path(__file__).resolve().parents[1]
log = base/'output'/'checks.json'
records = json.loads(log.read_text())
env = dict(os.environ, GOCACHE='/private/tmp/go-quality-testing-cache', GOTOOLCHAIN='local', GOPROXY='off')
for label, command, cwd in [
    ('module-version stdlib vet', ['rtk','proxy','go','vet','-stdversion','./...'], base/'candidate'),
    ('supplied changeset diff', ['rtk','proxy','diff','-ru',str(base/'original'),str(base/'candidate')], base),
]:
    started=time.time()
    result=subprocess.run(command, cwd=cwd, env=env, capture_output=True, text=True, timeout=40)
    records.append(dict(label=label, argv=command, cwd=str(cwd), env={k:env[k] for k in ['GOCACHE','GOTOOLCHAIN','GOPROXY']}, stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, elapsed_seconds=round(time.time()-started,3)))
    print(json.dumps(records[-1]),flush=True)
records.append(dict(label='unchanged production/module/documentation bytes', files={name:dict(original_sha256=hashlib.sha256((base/'original'/name).read_bytes()).hexdigest(),candidate_sha256=hashlib.sha256((base/'candidate'/name).read_bytes()).hexdigest(),equal=(base/'original'/name).read_bytes()==(base/'candidate'/name).read_bytes()) for name in ['store.go','go.mod','README.md']}))
log.write_text(json.dumps(records, indent=2)+'\n')
