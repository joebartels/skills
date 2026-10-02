import json
import os
from pathlib import Path
import subprocess
import sys
import time

report = Path(__file__).resolve().parent
module = report.parent / "module"
args = sys.argv[1:]
timeout = int(args.pop(0))
command = ["rtk", "proxy"] + args
settings = {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local"}
environment = dict(os.environ, **settings)
started = time.time()
try:
    result = subprocess.run(command, cwd=module, env=environment, capture_output=True, text=True, timeout=timeout)
    stdout, stderr, code = result.stdout, result.stderr, result.returncode
except subprocess.TimeoutExpired as exc:
    stdout = exc.stdout or ""
    stderr = exc.stderr or ""
    if isinstance(stdout, bytes):
        stdout = stdout.decode(errors="replace")
    if isinstance(stderr, bytes):
        stderr = stderr.decode(errors="replace")
    stderr += "\nverification process exceeded outer deadline"
    code = 124
entry = {"command": command, "cwd": str(module), "environment": settings, "timeout_seconds": timeout,
         "stdout": stdout, "stderr": stderr, "exit_code": code, "duration_seconds": round(time.time() - started, 3)}
path = report / "checks.json"
checks = json.loads(path.read_text()) if path.exists() else []
checks.append(entry)
path.write_text(json.dumps(checks, indent=2) + "\n")
print(json.dumps(entry, indent=2))
sys.exit(code)
