import datetime
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

module = Path('/private/tmp/go-fresh-author-txf1lez2/task-3/module')
report = Path('/private/tmp/go-fresh-author-txf1lez2/task-3/report')
label = sys.argv[1]
command = ['rtk', 'proxy', 'env', 'GOCACHE=/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN=local', *sys.argv[2:]]
started = time.monotonic()
entry = {
    'label': label,
    'command': command,
    'cwd': str(module),
    'environment': {'GOCACHE': '/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN': 'local'},
    'started_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'process_timeout_seconds': 150,
    'source_sha256': {name: hashlib.sha256((module / name).read_bytes()).hexdigest() for name in ['fetch.go', 'fetch_test.go', 'go.mod']},
}
process = subprocess.Popen(command, cwd=module, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, start_new_session=True)
try:
    stdout, stderr = process.communicate(timeout=150)
    entry['timed_out'] = False
except subprocess.TimeoutExpired:
    entry['timed_out'] = True
    os.killpg(process.pid, signal.SIGTERM)
    try:
        stdout, stderr = process.communicate(timeout=5)
    except subprocess.TimeoutExpired:
        os.killpg(process.pid, signal.SIGKILL)
        stdout, stderr = process.communicate()
entry.update(stdout=stdout, stderr=stderr, exit_code=process.returncode, duration_seconds=round(time.monotonic() - started, 3))
checks_path = report / 'checks.json'
checks = json.loads(checks_path.read_text()) if checks_path.exists() else []
checks.append(entry)
checks_path.write_text(json.dumps(checks, indent=2) + '\n')
print(json.dumps({'label': label, 'exit_code': process.returncode, 'timed_out': entry['timed_out'], 'duration_seconds': entry['duration_seconds'], 'stdout': stdout, 'stderr': stderr}))
sys.exit(process.returncode if process.returncode >= 0 else 1)
