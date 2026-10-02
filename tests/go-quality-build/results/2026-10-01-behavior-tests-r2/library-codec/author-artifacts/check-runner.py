import json
import os
import pathlib
import subprocess
import sys
import time

report_dir = pathlib.Path(__file__).parent
label, timeout = sys.argv[1], float(sys.argv[2])
command = sys.argv[3:]
environment = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
started = time.monotonic()
record = {
    "label": label,
    "command": command,
    "cwd": os.getcwd(),
    "environment": {"GOCACHE": environment["GOCACHE"], "GOTOOLCHAIN": environment["GOTOOLCHAIN"]},
    "deadline_seconds": timeout,
}
try:
    result = subprocess.run(command, env=environment, text=True, capture_output=True, timeout=timeout)
    record.update(stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, timed_out=False)
except subprocess.TimeoutExpired as exc:
    def as_text(value):
        return value.decode(errors="replace") if isinstance(value, bytes) else (value or "")
    record.update(stdout=as_text(exc.stdout), stderr=as_text(exc.stderr), exit_code=None, timed_out=True)
record["elapsed_seconds"] = round(time.monotonic() - started, 3)
path = report_dir / "checks.json"
records = json.loads(path.read_text()) if path.exists() else []
records.append(record)
path.write_text(json.dumps(records, indent=2) + "\n")
print(json.dumps(record, indent=2))
sys.exit(record["exit_code"] if record["exit_code"] is not None else 124)
