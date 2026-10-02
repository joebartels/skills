import json
import os
from pathlib import Path
import subprocess
import sys
import time

report = Path(__file__).parent
label = sys.argv[1]
command = sys.argv[3:]
env = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
started = time.time()
try:
    result = subprocess.run(command, cwd="/private/tmp/go-fresh-author-mm5bi5gc/task-6/module", env=env, capture_output=True, text=True, timeout=90)
    code, stdout, stderr = result.returncode, result.stdout, result.stderr
except subprocess.TimeoutExpired as exc:
    code = 124
    stdout = (exc.stdout or b"").decode() if isinstance(exc.stdout, bytes) else (exc.stdout or "")
    stderr = (exc.stderr or b"").decode() if isinstance(exc.stderr, bytes) else (exc.stderr or "")
    stderr += "\ncheck exceeded 90-second process deadline"
entry = {"label": label, "command": command, "cwd": "/private/tmp/go-fresh-author-mm5bi5gc/task-6/module", "environment": {"GOCACHE": env["GOCACHE"], "GOTOOLCHAIN": env["GOTOOLCHAIN"]}, "process_timeout_seconds": 90, "elapsed_seconds": round(time.time() - started, 3), "stdout": stdout, "stderr": stderr, "exit_code": code}
path = report / "checks.json"
checks = json.loads(path.read_text()) if path.exists() else []
checks.append(entry)
path.write_text(json.dumps(checks, indent=2) + "\n")
print(json.dumps(entry, indent=2))
sys.exit(code)
