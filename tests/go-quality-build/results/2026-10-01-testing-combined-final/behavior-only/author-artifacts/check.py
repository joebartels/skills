import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

root = Path("/private/tmp/go-fresh-author-s21flpf_/task-1/module")
report = Path("/private/tmp/go-fresh-author-s21flpf_/task-1/report/checks.json")
env = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
command = ["rtk", "proxy"] + sys.argv[2:]
timeout = int(sys.argv[1])
def sources():
    return {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob("*.go"))}
before = sources()
start = time.monotonic()
try:
    proc = subprocess.run(command, cwd=root, env=env, text=True, capture_output=True, timeout=timeout)
    result = dict(command=command, cwd=str(root), environment={"GOCACHE": env["GOCACHE"], "GOTOOLCHAIN": env["GOTOOLCHAIN"]}, process_deadline_seconds=timeout, stdout=proc.stdout, stderr=proc.stderr, exit_code=proc.returncode)
except subprocess.TimeoutExpired as exc:
    result = dict(command=command, cwd=str(root), environment={"GOCACHE": env["GOCACHE"], "GOTOOLCHAIN": env["GOTOOLCHAIN"]}, process_deadline_seconds=timeout, stdout=str(exc.stdout or ""), stderr=str(exc.stderr or ""), exit_code=None, timed_out=True)
result["elapsed_seconds"] = round(time.monotonic() - start, 3)
result["source_sha256_before"] = before
result["source_sha256_after"] = sources()
result["execution_boundary"] = os.environ.get("INDEXER_CHECK_BOUNDARY", "sandboxed")
results = json.loads(report.read_text()) if report.exists() else []
results.append(result)
report.write_text(json.dumps(results, indent=2) + "\n")
print(json.dumps(result), flush=True)
