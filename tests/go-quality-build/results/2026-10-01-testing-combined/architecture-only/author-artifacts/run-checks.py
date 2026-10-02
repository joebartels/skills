import json
import os
import pathlib
import subprocess
import sys
import time

ROOT = pathlib.Path('/private/tmp/go-combined-author-vsgga4k8/task-1/module')
REPORT = pathlib.Path('/private/tmp/go-combined-author-vsgga4k8/task-1/report')
ENV = {'GOCACHE': '/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN': 'local'}
environment = os.environ.copy()
environment.update(ENV)
records_path = REPORT / 'checks.json'
records = json.loads(records_path.read_text()) if records_path.exists() else []
checks = {
    'version': (['rtk', 'proxy', 'go', 'version'], 15),
    'format': (['rtk', 'proxy', 'gofmt', '-w', 'index.go', 'refresh.go', 'serve.go', 'apply_test.go', 'refresh_test.go', 'serve_test.go', 'cmd/indexer/main.go', 'cmd/indexer/process_test.go'], 15),
    'format-check': (['rtk', 'proxy', 'gofmt', '-l', '.'], 15),
    'test': (['rtk', 'proxy', 'go', 'test', '-timeout=90s', './...'], 120),
    'test-approved': (['rtk', 'proxy', 'go', 'test', '-count=1', '-timeout=90s', './...'], 120),
    'race': (['rtk', 'proxy', 'go', 'test', '-race', '-timeout=90s', './...'], 120),
    'repeat': (['rtk', 'proxy', 'go', 'test', '-race', '-count=5', '-timeout=90s', './...'], 120),
    'vet': (['rtk', 'proxy', 'go', 'vet', './...'], 120),
    'coverage': (['rtk', 'proxy', 'go', 'test', '-coverprofile=' + str(REPORT / 'coverage.out'), '-timeout=90s', './...'], 120),
    'tools': (['rtk', 'proxy', 'python3', '-c', 'import shutil; print("staticcheck=" + str(shutil.which("staticcheck")))'], 15),
}
for name in sys.argv[1:]:
    command, deadline = checks[name]
    print('Starting ' + name, flush=True)
    started = time.monotonic()
    try:
        completed = subprocess.run(command, cwd=ROOT, env=environment, capture_output=True, text=True, timeout=deadline)
        stdout, stderr, code = completed.stdout, completed.stderr, completed.returncode
        timed_out = False
    except subprocess.TimeoutExpired as failure:
        stdout = failure.stdout or ''
        stderr = failure.stderr or ''
        stdout = stdout.decode() if isinstance(stdout, bytes) else stdout
        stderr = stderr.decode() if isinstance(stderr, bytes) else stderr
        code, timed_out = None, True
    entry = {'name': name, 'command': command, 'cwd': str(ROOT), 'environment': ENV, 'process_deadline_seconds': deadline, 'elapsed_seconds': time.monotonic() - started, 'stdout': stdout, 'stderr': stderr, 'exit_code': code, 'timed_out': timed_out}
    entry['execution_boundary'] = 'approved-localhost-listeners' if name in {'test-approved', 'race', 'repeat', 'coverage'} else 'sandbox'
    records.append(entry)
    records_path.write_text(json.dumps(records, indent=2) + '\n')
    print(json.dumps(entry), flush=True)
    if code != 0 or timed_out:
        sys.exit(1)
