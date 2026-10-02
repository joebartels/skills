from pathlib import Path
import json, subprocess, time, shlex, hashlib
base = Path('/private/tmp/go-independent-review-otiblxwu')
checks = []
def run(label, area, args, extra_env=None):
    command = ['rtk', 'proxy', 'env', 'GOCACHE=/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN=local', 'GOPROXY=off', 'GOSUMDB=off', 'GOWORK=off']
    command += extra_env or []
    command += ['go'] + args
    started = time.time()
    try:
        result = subprocess.run(command, cwd=base/'verification'/area, capture_output=True, text=True, timeout=90)
        record = dict(id=label, cwd=str(base/'verification'/area), argv=command, command=shlex.join(command), stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, elapsed_seconds=round(time.time()-started, 3))
    except subprocess.TimeoutExpired as e:
        record = dict(id=label, cwd=str(base/'verification'/area), argv=command, command=shlex.join(command), stdout=str(e.stdout or ''), stderr=str(e.stderr or ''), exit_code=None, timed_out=True, elapsed_seconds=round(time.time()-started, 3))
    checks.append(record)
    (base/'output'/'checks.json').write_text(json.dumps(checks, indent=2)+'\n')
    print(json.dumps(record), flush=True)
run('toolchain', 'candidate', ['version'])
run('original-serial', 'original', ['test','./...','-count=1','-timeout=30s'])
run('candidate-full', 'candidate', ['test','./...','-count=1','-timeout=30s'])
run('candidate-race-shuffle', 'candidate', ['test','./...','-race','-count=30','-shuffle=on','-parallel=8','-timeout=60s'])
run('candidate-focused-instance', 'candidate', ['test','./...','-run','^TestStoreIndependentInstances$/^instances$/^alpha$','-count=1','-timeout=30s'])
run('candidate-focused-leaf', 'candidate', ['test','./...','-run','^TestStoreIndependentInstances$/^instances$/^alpha$/^replacements$/^value-1$','-count=1','-timeout=30s'])
run('candidate-vet', 'candidate', ['vet','./...'])
run('mutation-shared-root', 'shared-root', ['test','./...','-run','^TestStoreIndependentInstances$','-parallel=1','-count=1','-timeout=30s'])
run('mutation-lossy-values', 'lossy-values', ['test','./...','-run','^TestStoreIndependentInstances$','-parallel=8','-count=1','-timeout=30s'])
run('mutation-early-cleanup', 'early-cleanup', ['test','./...','-run','^TestStoreIndependentInstances$','-race','-parallel=8','-count=1','-timeout=30s'])
run('candidate-go122-language', 'candidate', ['test','./...','-gcflags=all=-lang=go1.22','-count=1','-timeout=30s'])
for area in ['original','candidate']:
    hashes = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((base/area).iterdir()) if p.is_file()}
    print(json.dumps({'source_hashes': area, 'sha256': hashes}), flush=True)
