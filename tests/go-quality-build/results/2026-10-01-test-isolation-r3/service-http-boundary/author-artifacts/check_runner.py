import json
import os
from pathlib import Path
import subprocess
import sys

report = Path('/private/tmp/go-fresh-author-8hjrllxp/task-3/report/checks.json')
cwd = '/private/tmp/go-fresh-author-8hjrllxp/task-3/module'
command = sys.argv[1:]
overrides = {
    'GOCACHE': '/private/tmp/go-quality-testing-cache',
    'GOTOOLCHAIN': 'local',
}
environment = os.environ.copy()
environment.update(overrides)
entry = {'command': command, 'cwd': cwd, 'environment': overrides, 'process_timeout_seconds': 90,
         'execution_profile': os.environ.get('CHECK_EXECUTION_PROFILE', 'sandbox_default')}
try:
    result = subprocess.run(command, cwd=cwd, env=environment, capture_output=True, text=True, timeout=90)
    entry.update(stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode)
except subprocess.TimeoutExpired as error:
    entry.update(stdout=error.stdout.decode() if isinstance(error.stdout, bytes) else error.stdout or '',
                 stderr=error.stderr.decode() if isinstance(error.stderr, bytes) else error.stderr or '',
                 exit_code=None, timed_out=True)
checks = json.loads(report.read_text()) if report.exists() else []
checks.append(entry)
report.write_text(json.dumps(checks, indent=2) + '\n')
print(json.dumps(entry, indent=2))
sys.exit(entry['exit_code'] if entry['exit_code'] is not None else 124)
