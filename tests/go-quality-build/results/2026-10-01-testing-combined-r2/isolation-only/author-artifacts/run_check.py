import json, os, shutil, subprocess, sys, time
from pathlib import Path
root = Path('/private/tmp/go-fresh-author-qjdzm0iq/task-1/module')
report = Path('/private/tmp/go-fresh-author-qjdzm0iq/task-1/report')
args = sys.argv[1:]
command = ['rtk', 'proxy', 'env', 'GOCACHE=/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN=local'] + args
started = time.time()
try:
    result = subprocess.run(command, cwd=root, capture_output=True, text=True, timeout=180)
    stdout, stderr, code = result.stdout, result.stderr, result.returncode
    timed_out = False
except subprocess.TimeoutExpired as error:
    stdout = error.stdout or b''
    stderr = error.stderr or b''
    if isinstance(stdout, bytes): stdout = stdout.decode(errors='replace')
    if isinstance(stderr, bytes): stderr = stderr.decode(errors='replace')
    code, timed_out = None, True
entry = {'command': command, 'cwd': str(root), 'environment': {'GOCACHE': '/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN': 'local'}, 'stdout': stdout, 'stderr': stderr, 'exit_code': code, 'timed_out': timed_out, 'elapsed_seconds': round(time.time()-started, 3)}
path = report / 'checks.json'
checks = json.loads(path.read_text()) if path.exists() else []
checks.append(entry)
path.write_text(json.dumps(checks, indent=2) + '\n')
print(json.dumps({'command': args, 'exit_code': code, 'timed_out': timed_out, 'elapsed_seconds': entry['elapsed_seconds']}))
print(stdout, end='')
print(stderr, end='', file=sys.stderr)
sys.exit(124 if timed_out else code)
