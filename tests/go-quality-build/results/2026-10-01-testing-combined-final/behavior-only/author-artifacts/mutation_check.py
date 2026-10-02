import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path("/private/tmp/go-fresh-author-s21flpf_/task-1/module")
source = root / "index.go"
report = Path("/private/tmp/go-fresh-author-s21flpf_/task-1/report")
original = source.read_bytes()
needle = b"func publish(path string, data []byte) error {\n"
if original.count(needle) != 1:
    raise RuntimeError("unexpected publication function")
mutated = original.replace(needle, needle + b"\treturn os.WriteFile(path, data, 0600) // temporary mutation\n", 1)
env = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
try:
    source.write_bytes(mutated)
    command = ["rtk", "proxy", "python3", str(report / "check.py"), "45", "go", "test", "-count=1", "-run=^TestWriteFailureAfterProgressRetainsState$", "-timeout=30s", "."]
    proc = subprocess.run(command, cwd=root, env=env, capture_output=True, text=True, timeout=50)
    print(proc.stdout, end="")
    print(proc.stderr, end="", file=sys.stderr)
    result = json.loads((report / "checks.json").read_text())[-1]
    if result["exit_code"] != 1 or "want \"original\"" not in result["stdout"] or "want \"accepted\"" not in result["stdout"]:
        raise RuntimeError("mutation did not produce expected retained-state assertion failures")
finally:
    source.write_bytes(original)
    (report / "mutation.json").write_text(json.dumps({"change": "temporary direct os.WriteFile replacing atomic publication", "restored": source.read_bytes() == original, "restored_sha256": hashlib.sha256(source.read_bytes()).hexdigest()}, indent=2) + "\n")
