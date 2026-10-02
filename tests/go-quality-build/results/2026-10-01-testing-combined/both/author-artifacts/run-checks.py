import json
import os
from pathlib import Path
import subprocess
import sys
import time

module = Path("/private/tmp/go-combined-author-vsgga4k8/task-4/module")
report = Path("/private/tmp/go-combined-author-vsgga4k8/task-4/report")
log = report / "checks.json"
environment = {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local"}
stages = {
    "initial": [
        ["gofmt", "-w", "index.go", "refresh.go", "serve.go", "index_test.go", "apply_test.go", "refresh_test.go", "serve_test.go", "write_failure_unix_test.go", "cmd/indexer/main.go", "cmd/indexer/main_test.go", "cmd/indexer/process_test.go"],
        ["go", "test", "-timeout=60s", "./..."],
    ],
    "verify": [
        ["go", "vet", "./..."],
        ["go", "test", "-race", "-shuffle=on", "-count=3", "-timeout=90s", "./..."],
    ],
    "rerun": [["go", "test", "-timeout=60s", "./..."]],
    "focused": [
        ["go", "test", "-timeout=30s", "-run=^TestApplyCSVRejectedReplacementRetainsPrefix$/^empty_text$", "."],
        ["go", "test", "-timeout=30s", "-run=^TestPartialWriteFailureRetainsPriorData$/^csv$", "."],
        ["go", "test", "-timeout=30s", "-run=^TestServeCancellationJoinsCleanupBeforeRelease$/^callback_error$", "."],
        ["go", "test", "-timeout=30s", "-run=^TestWatchProcessContracts$/^recurrence_and_in_flight_termination$", "./cmd/indexer"],
        ["go", "test", "-timeout=30s", "-run=^TestCommandProcessContracts$/^apply_validation_prefix$", "./cmd/indexer"],
    ],
}
stages["final"] = [stages["initial"][0], *stages["verify"], *stages["focused"], ["gofmt", "-l", *stages["initial"][0][2:]], ["go", "version"], ["go", "list", "-m", "-json"], ["python3", "-c", "import shutil; print(shutil.which('staticcheck') or 'staticcheck unavailable')"]]
records = json.loads(log.read_text()) if log.exists() else []
for tool_command in stages[sys.argv[1]]:
    command = ["rtk", "proxy", "env", *[f"{key}={value}" for key, value in environment.items()], *tool_command]
    started = time.monotonic()
    try:
        result = subprocess.run(command, cwd=module, env=os.environ.copy(), capture_output=True, text=True, timeout=120)
        entry = {"command": command, "cwd": str(module), "environment": environment, "stdout": result.stdout, "stderr": result.stderr, "exit_code": result.returncode, "elapsed_seconds": time.monotonic() - started}
    except subprocess.TimeoutExpired as error:
        entry = {"command": command, "cwd": str(module), "environment": environment, "stdout": (error.stdout or b"").decode() if isinstance(error.stdout, bytes) else error.stdout or "", "stderr": (error.stderr or b"").decode() if isinstance(error.stderr, bytes) else error.stderr or "", "exit_code": None, "timeout_seconds": 120, "elapsed_seconds": time.monotonic() - started}
    records.append(entry)
    entry["execution_context"] = sys.argv[2] if len(sys.argv) > 2 else "workspace sandbox"
    log.write_text(json.dumps(records, indent=2) + "\n")
    print(json.dumps(entry), flush=True)
    if entry["exit_code"] != 0:
        sys.exit(1)
