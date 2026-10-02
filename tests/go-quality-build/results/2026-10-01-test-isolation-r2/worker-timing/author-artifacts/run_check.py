"""Run an authorized local verification command and preserve its outcome."""
import json
import subprocess
import sys
import time
from pathlib import Path

REPORT = Path(__file__).parent
MODULE = REPORT.parent / "module"
argv = ["rtk", "proxy", "env", "GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local", *sys.argv[1:]]
record = {
    "argv": argv,
    "cwd": str(MODULE),
    "environment": {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local"},
    "process_timeout_seconds": 60,
}
start = time.monotonic()
try:
    result = subprocess.run(argv, cwd=MODULE, capture_output=True, text=True, timeout=60)
    record.update(stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode)
except subprocess.TimeoutExpired as exc:
    record.update(stdout=(exc.stdout or b"").decode() if isinstance(exc.stdout, bytes) else (exc.stdout or ""),
                  stderr=(exc.stderr or b"").decode() if isinstance(exc.stderr, bytes) else (exc.stderr or ""),
                  exit_code=None, timed_out=True)
record["duration_seconds"] = round(time.monotonic() - start, 3)
checks_path = REPORT / "checks.json"
checks = json.loads(checks_path.read_text()) if checks_path.exists() else []
checks.append(record)
checks_path.write_text(json.dumps(checks, indent=2) + "\n")
print(json.dumps(record))
sys.exit(record.get("exit_code") or (124 if record.get("timed_out") else 0))
