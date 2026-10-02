import os, sys, json, subprocess, time
from pathlib import Path
root = Path("/private/tmp/go-independent-review-77k7qpuq")
log = root/"output"/"checks.json"
records = json.loads(log.read_text()) if log.exists() else []
name, cwd, timeout, *cmd = sys.argv[1:]
env = os.environ.copy()
env.update(GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
started = time.time()
try:
    result = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True, timeout=float(timeout))
    entry = dict(name=name, command=cmd, cwd=cwd, env={k:env[k] for k in ["GOCACHE","GOTOOLCHAIN"]}, stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, elapsed_seconds=time.time()-started)
except subprocess.TimeoutExpired as exc:
    entry = dict(name=name, command=cmd, cwd=cwd, env={k:env[k] for k in ["GOCACHE","GOTOOLCHAIN"]}, stdout=exc.stdout.decode() if isinstance(exc.stdout, bytes) else exc.stdout or "", stderr=exc.stderr.decode() if isinstance(exc.stderr, bytes) else exc.stderr or "", exit_code=None, timeout_seconds=float(timeout), elapsed_seconds=time.time()-started)
records.append(entry)
log.write_text(json.dumps(records, indent=2)+"\n")
print(json.dumps(entry, indent=2))
