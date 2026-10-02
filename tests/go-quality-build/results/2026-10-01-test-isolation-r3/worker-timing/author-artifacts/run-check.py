import json
import os
from pathlib import Path
import subprocess
import sys
import time

report = Path(__file__).parent
cwd = "/private/tmp/go-fresh-author-8hjrllxp/task-4/module"
command = sys.argv[1:]
environment = {
    "GOCACHE": "/private/tmp/go-quality-testing-cache",
    "GOTOOLCHAIN": "local",
}
started = time.monotonic()
try:
    completed = subprocess.run(
        command,
        cwd=cwd,
        env={**os.environ, **environment},
        capture_output=True,
        text=True,
        timeout=120,
    )
    record = {
        "command": command,
        "cwd": cwd,
        "environment": environment,
        "stdout": completed.stdout,
        "stderr": completed.stderr,
        "exit_code": completed.returncode,
    }
except subprocess.TimeoutExpired as error:
    record = {
        "command": command,
        "cwd": cwd,
        "environment": environment,
        "stdout": (error.stdout or b"").decode() if isinstance(error.stdout, bytes) else (error.stdout or ""),
        "stderr": (error.stderr or b"").decode() if isinstance(error.stderr, bytes) else (error.stderr or ""),
        "exit_code": None,
        "process_deadline_seconds": 120,
        "timed_out": True,
    }
record["duration_seconds"] = round(time.monotonic() - started, 3)
path = report / "checks.json"
records = json.loads(path.read_text()) if path.exists() else []
records.append(record)
path.write_text(json.dumps(records, indent=2) + "\n")
print(json.dumps(record, indent=2))
sys.exit(record["exit_code"] if record["exit_code"] is not None else 124)
