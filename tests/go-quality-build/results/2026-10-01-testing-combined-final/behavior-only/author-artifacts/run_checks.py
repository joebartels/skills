import json
import os
import subprocess
import time

root = "/private/tmp/go-fresh-author-s21flpf_/task-1/module"
report = "/private/tmp/go-fresh-author-s21flpf_/task-1/report/checks.json"
env = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
commands = [
    (["rtk", "proxy", "gofmt", "-w", "."], 15),
    (["rtk", "proxy", "go", "test", "-timeout=45s", "./..."], 60),
    (["rtk", "proxy", "go", "vet", "./..."], 45),
    (["rtk", "proxy", "go", "test", "-race", "-shuffle=on", "-count=3", "-timeout=90s", "./..."], 120),
]
results = []
for command, timeout in commands:
    start = time.monotonic()
    try:
        proc = subprocess.run(command, cwd=root, env=env, text=True, capture_output=True, timeout=timeout)
        result = dict(command=command, cwd=root, environment={"GOCACHE": env["GOCACHE"], "GOTOOLCHAIN": env["GOTOOLCHAIN"]}, process_deadline_seconds=timeout, stdout=proc.stdout, stderr=proc.stderr, exit_code=proc.returncode)
    except subprocess.TimeoutExpired as exc:
        result = dict(command=command, cwd=root, environment={"GOCACHE": env["GOCACHE"], "GOTOOLCHAIN": env["GOTOOLCHAIN"]}, process_deadline_seconds=timeout, stdout=str(exc.stdout or ""), stderr=str(exc.stderr or ""), exit_code=None, timed_out=True)
    result["elapsed_seconds"] = round(time.monotonic() - start, 3)
    results.append(result)
    with open(report, "w") as f:
        json.dump(results, f, indent=2)
        f.write("\n")
    print(json.dumps(result), flush=True)
    if result["exit_code"] != 0:
        break
