import sys, subprocess, json, time, shlex, os
from pathlib import Path
base = Path("/private/tmp/go-independent-review-0vvp8kdb")
label, cwd, timeout, *cmd = sys.argv[1:]
started = time.time()
record = {"label": label, "cwd": cwd, "argv": cmd, "command": shlex.join(cmd), "timeout_seconds": float(timeout)}
try:
    p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=float(timeout))
    record.update(stdout=p.stdout, stderr=p.stderr, exit_code=p.returncode, timed_out=False)
except subprocess.TimeoutExpired as e:
    def text(x):
        return x.decode(errors="replace") if isinstance(x, bytes) else x or ""
    record.update(stdout=text(e.stdout), stderr=text(e.stderr), exit_code=None, timed_out=True)
record["elapsed_seconds"] = time.time()-started
checks = base/"output"/"checks.json"
entries = json.loads(checks.read_text()) if checks.exists() else []
entries.append(record)
checks.write_text(json.dumps(entries, indent=2)+"\n")
print(json.dumps(record))
