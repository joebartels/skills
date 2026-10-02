import json
import os
from pathlib import Path
import shutil
import subprocess

module = Path("/private/tmp/go-fresh-author-8hjrllxp/task-6/module")
report = Path("/private/tmp/go-fresh-author-8hjrllxp/task-6/report")
environment = {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local"}
env = dict(os.environ, **environment)
commands = [
    (["rtk", "proxy", "go", "version"], 15),
    (["rtk", "proxy", "gofmt", "-w", "store_test.go"], 15),
    (["rtk", "proxy", "gofmt", "-l", "store_test.go"], 15),
    (["rtk", "proxy", "go", "vet", "./..."], 120),
    (["rtk", "proxy", "go", "test", "-v", "-timeout=30s", "./..."], 120),
    (["rtk", "proxy", "go", "test", "-race", "-shuffle=on", "-count=20", "-parallel=3", "-timeout=30s", "./..."], 180),
    (["rtk", "proxy", "go", "test", "-race", "-count=10", "-timeout=30s", "-run=^TestStoreSerialChildren$", "./..."], 120),
    (["rtk", "proxy", "go", "test", "-race", "-v", "-timeout=30s", "-run=^TestStoreIndependentInstances$/^instances$/^alpha$/^initial$", "./..."], 120),
    (["rtk", "proxy", "go", "test", "-race", "-v", "-timeout=30s", "-run=^TestStoreIndependentInstances$/^instances$/^alpha$/^repeat$", "./..."], 120),
    (["rtk", "proxy", "go", "test", "-race", "-v", "-timeout=30s", "-run=^TestStoreIndependentInstances$/^instances$/^beta$/^replacement$", "./..."], 120),
    (["rtk", "proxy", "go", "test", "-race", "-shuffle=on", "-count=20", "-timeout=30s", "-run=^TestStoreIndependentInstances$/^instances$/^gamma$", "./..."], 120),
]
if shutil.which("staticcheck"):
    commands.append((["rtk", "proxy", "staticcheck", "./..."], 120))

results = []
for command, timeout in commands:
    try:
        completed = subprocess.run(command, cwd=module, env=env, capture_output=True, text=True, timeout=timeout)
        result = {"command": command, "cwd": str(module), "environment": environment, "process_timeout_seconds": timeout, "stdout": completed.stdout, "stderr": completed.stderr, "exit_code": completed.returncode}
    except subprocess.TimeoutExpired as error:
        def decoded(value):
            return value.decode(errors="replace") if isinstance(value, bytes) else value or ""
        result = {"command": command, "cwd": str(module), "environment": environment, "process_timeout_seconds": timeout, "stdout": decoded(error.stdout), "stderr": decoded(error.stderr), "exit_code": None, "timed_out": True}
    results.append(result)
    (report / "checks.json").write_text(json.dumps({"checks": results, "staticcheck_available": bool(shutil.which("staticcheck"))}, indent=2) + "\n")
    print(json.dumps(result), flush=True)
