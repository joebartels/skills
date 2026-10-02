import json
import os
from pathlib import Path
import subprocess
import sys

report = Path(__file__).parent / "checks.json"
label, seconds, *command = sys.argv[1:]
env = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
record = {
    "label": label,
    "command": command,
    "environment": {"GOCACHE": env["GOCACHE"], "GOTOOLCHAIN": env["GOTOOLCHAIN"]},
    "timeout_seconds": int(seconds),
}
try:
    result = subprocess.run(command, env=env, text=True, capture_output=True, timeout=int(seconds))
    record.update(stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, timed_out=False)
except subprocess.TimeoutExpired as error:
    def decode(value):
        return value.decode(errors="replace") if isinstance(value, bytes) else (value or "")
    record.update(stdout=decode(error.stdout), stderr=decode(error.stderr), exit_code=None, timed_out=True)
checks = json.loads(report.read_text()) if report.exists() else []
checks.append(record)
report.write_text(json.dumps(checks, indent=2) + "\n")
print(json.dumps(record, indent=2))
sys.exit(record["exit_code"] if record["exit_code"] is not None else 124)
