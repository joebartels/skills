import hashlib
import json
from pathlib import Path
import shutil
import subprocess

base = Path(__file__).resolve().parents[1]
output = base / "output"
checks_path = output / "checks.json"
records = json.loads(checks_path.read_text())
env = {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local"}


def hashes():
    return {
        str(path.relative_to(base)): hashlib.sha256(path.read_bytes()).hexdigest()
        for kind in ("original", "candidate")
        for path in sorted((base / kind).iterdir())
        if path.is_file()
    }


before = hashes()
probes = base / "probes"
original_branch = probes / "original-lower-branch"
shutil.copytree(base / "candidate", original_branch)
shutil.copyfile(base / "original" / "clamp.go", original_branch / "clamp.go")
behavior = probes / "supported-boundaries"
shutil.copytree(base / "candidate", behavior)
(behavior / "review_probe_test.go").write_text('''package clamp_test

import (
    "testing"
    "example.com/clamp"
)

var compatibleFunction func(int, int, int) int = clamp.Clamp

func TestReviewSupportedBoundaries(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct{ value, low, high, want int }{
        {min, -7, 3, -7}, {-7, -7, 3, -7}, {-1, -7, 3, -1},
        {3, -7, 3, 3}, {max, -7, 3, 3},
        {min, min, max, min}, {max, min, max, max},
        {-1, 0, 0, 0}, {0, 0, 0, 0}, {1, 0, 0, 0},
    }
    for _, c := range cases {
        if got := compatibleFunction(c.value, c.low, c.high); got != c.want {
            t.Errorf("Clamp(%d,%d,%d) = %d; want %d", c.value, c.low, c.high, got, c.want)
        }
    }
}
''')

for name, probe, test, expected in [
    ("original-branch-signal", original_branch, "^TestLowerBound$", "nonzero assertion failure for values 0 and 1"),
    ("supported-boundaries-and-function-type", behavior, "^TestReviewSupportedBoundaries$", "zero exit"),
]:
    args = ["rtk", "proxy", "env"] + [f"{k}={v}" for k, v in env.items()] + [
        "go", "test", "./...", "-run", test, "-count=1", "-timeout=20s"
    ]
    run = subprocess.run(args, cwd=probe, text=True, capture_output=True, timeout=30)
    record = {
        "id": name, "command": args, "cwd": str(probe), "env_overrides": env,
        "stdout": run.stdout, "stderr": run.stderr, "exit_code": run.returncode,
        "expected_exit": expected,
    }
    records.append(record)
    checks_path.write_text(json.dumps(records, indent=2) + "\n")
    print(json.dumps(record), flush=True)

records.append({"id": "source-preservation", "before": before, "after": hashes(), "unchanged": before == hashes()})
checks_path.write_text(json.dumps(records, indent=2) + "\n")
