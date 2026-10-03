"""Run exact-snippet verification only; no author behavioral trial."""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

HERE = Path(__file__).resolve().parent
EXAMPLE = HERE.parent / "example"
ROOT = HERE.parents[5]
MINIMUM = "/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go"
HOST = "/opt/homebrew/Cellar/go/1.26.5/libexec/bin/go"


def run(identifier, argv, env, cwd=EXAMPLE):
    started = time.monotonic()
    process = subprocess.run(
        ["rtk", "proxy", *argv], cwd=cwd, env=env,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=60,
    )
    (HERE / f"{identifier}.stdout.txt").write_bytes(process.stdout)
    (HERE / f"{identifier}.stderr.txt").write_bytes(process.stderr)
    record = {
        "id": identifier, "command": ["rtk", "proxy", *argv],
        "cwd": str(cwd), "exit_code": process.returncode,
        "elapsed_seconds": time.monotonic() - started,
        "environment_overrides": {key: env[key] for key in
                                  ("GOTOOLCHAIN", "GOPROXY", "GOCACHE", "GOMODCACHE")},
        "stdout": f"{identifier}.stdout.txt", "stderr": f"{identifier}.stderr.txt",
    }
    (HERE / f"{identifier}.json").write_text(json.dumps(record, indent=2) + "\n")
    return record, process.stdout, process.stderr


results = []
for label, executable in (("minimum", MINIMUM), ("host", HOST)):
    env = dict(os.environ, GOTOOLCHAIN="local", GOPROXY="off",
               GOCACHE=f"/private/tmp/concurrency-preflight-{label}-cache",
               GOMODCACHE="/private/tmp/concurrency-preflight-modcache")
    commands = (
        ("version", [executable, "version"]),
        ("environment", [executable, "env", "GOVERSION", "GOROOT", "GOOS", "GOARCH", "GOTOOLCHAIN"]),
        ("build", [executable, "build", "./..."]),
        ("test", [executable, "test", "-count=1", "-timeout=20s", "./..."]),
        ("vet", [executable, "vet", "./..."]),
        ("race-shuffle", [executable, "test", "-race", "-shuffle=on", "-count=5", "-timeout=30s", "./..."]),
        ("format", [str(Path(executable).with_name("gofmt")), "-d", "totals.go", "totals_test.go"]),
    )
    for name, argv in commands:
        record, stdout, stderr = run(f"{label}-{name}", argv, env)
        if name == "format":
            record["format_diff_empty"] = not stdout
        results.append(record)
        print(f"{record['id']}: exit={record['exit_code']}", flush=True)
        assert record["exit_code"] == 0, (record, stdout, stderr)
        assert name != "format" or not stdout, stdout

validator_env = dict(os.environ, GOTOOLCHAIN="local", GOPROXY="off",
                     GOCACHE="/private/tmp/concurrency-preflight-host-cache",
                     GOMODCACHE="/private/tmp/concurrency-preflight-modcache")
record, stdout, stderr = run(
    "skill-validator",
    ["python3", "/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py",
     str(ROOT / "tests/go-quality-build/go-concurrency-and-ownership/draft")],
    validator_env, cwd=ROOT,
)
results.append(record)
assert record["exit_code"] == 0, (record, stdout, stderr)
(HERE / "example-verification.json").write_text(json.dumps({
    "scope": "Exact inline example with package/import wrapper; preflight semantics only",
    "behavioral_utility_claim": False,
    "checks": results,
    "example_sources": {str(p.name): hashlib.sha256(p.read_bytes()).hexdigest()
                        for p in sorted(EXAMPLE.iterdir()) if p.is_file()},
}, indent=2) + "\n")
