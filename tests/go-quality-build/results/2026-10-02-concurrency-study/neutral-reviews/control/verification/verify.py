#!/usr/bin/env python3
"""Bounded review checks; candidate/original inputs are never edited."""
from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import time

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "verification"
WORK = OUT / "work"
INPUTS = [ROOT / "original", ROOT / "candidates" / "A", ROOT / "candidates" / "B"]

def snapshots():
    return {
        str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest()
        for d in INPUTS for p in sorted(d.iterdir()) if p.is_file()
    }

before = snapshots()
env_delta = {
    "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off",
    "GOCACHE": str(OUT / "cache"), "GOFLAGS": "",
}
env = os.environ.copy()
env.update(env_delta)
records = []

def run(label, cwd, *args, expected=0):
    argv = ["rtk", "proxy", *args]
    started = time.monotonic()
    result = subprocess.run(argv, cwd=cwd, env=env, text=True,
                            capture_output=True, timeout=60)
    record = {
        "label": label, "argv": argv, "cwd": str(cwd),
        "environment_overrides": env_delta, "status": result.returncode,
        "expected_status": expected, "stdout": result.stdout,
        "stderr": result.stderr,
        "elapsed_seconds": round(time.monotonic() - started, 3),
    }
    records.append(record)
    (OUT / "raw-verification.json").write_text(json.dumps(records, indent=2) + "\n")
    print(f"{label}: status={result.returncode}, expected={expected}", flush=True)
    if result.returncode != expected:
        raise RuntimeError(f"unexpected check outcome: {label}")

observation = '''package positive

import "testing"

func TestReviewContractCases(t *testing.T) {
    cases := []struct {
        name string
        values []int
        want int
    }{
        {"nil", nil, 0},
        {"empty", []int{}, 0},
        {"positive singleton", []int{7}, 7},
        {"zero singleton", []int{0}, 0},
        {"negative singleton", []int{-7}, 0},
        {"negative only", []int{-9, -2, -1}, 0},
        {"zeros only", []int{0, 0, 0}, 0},
        {"positive first", []int{7, 0, -2}, 7},
        {"positive middle", []int{0, 7, -2}, 7},
        {"positive final", []int{0, -2, 7}, 7},
        {"all positive", []int{1, 2, 3, 4}, 10},
        {"mixed", []int{-5, 4, 0, 2, -1, 3}, 9},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            before := append([]int(nil), tc.values...)
            if got := sumPositive(tc.values); got != tc.want {
                t.Fatalf("sumPositive(%v)=%d, want %d", tc.values, got, tc.want)
            }
            for i, value := range before {
                if tc.values[i] != value {
                    t.Fatalf("input changed at index %d", i)
                }
            }
        })
    }
}
'''
(OUT / "contract_observation_test.go").write_text(observation)
WORK.mkdir(parents=True, exist_ok=True)
run("local toolchain", ROOT, "go", "version")

for candidate in ("A", "B"):
    source = ROOT / "candidates" / candidate
    ordinary = WORK / candidate / "ordinary"
    shutil.copytree(source, ordinary, dirs_exist_ok=True)
    run(f"{candidate} ordinary tests", ordinary, "go", "test", "-mod=readonly", "-count=1", "-timeout=30s", "./...")
    run(f"{candidate} vet", ordinary, "go", "vet", "-mod=readonly", "./...")
    run(f"{candidate} formatting", ordinary, "gofmt", "-d", "sum.go", "sum_test.go")
    run(f"{candidate} selected modules", ordinary, "go", "list", "-mod=readonly", "-m", "all")
    observed = WORK / candidate / "observed"
    shutil.copytree(source, observed, dirs_exist_ok=True)
    (observed / "review_contract_test.go").write_text(observation)
    run(f"{candidate} independent contract observation", observed, "go", "test", "-mod=readonly", "-count=1", "-timeout=30s", "-run", "^TestReviewContractCases$", "-v", "./...")
    reverted = WORK / candidate / "original-implementation"
    shutil.copytree(source, reverted, dirs_exist_ok=True)
    shutil.copyfile(ROOT / "original" / "sum.go", reverted / "sum.go")
    run(f"{candidate} supplied regression sensitivity to original implementation", reverted, "go", "test", "-mod=readonly", "-count=1", "-timeout=30s", "-run", "Final", "-v", "./...", expected=1)
    ordinary_snapshot = {
        str(p.relative_to(ordinary)): hashlib.sha256(p.read_bytes()).hexdigest()
        for p in sorted(ordinary.iterdir()) if p.is_file()
    }
    expected_snapshot = {
        p.name: hashlib.sha256(p.read_bytes()).hexdigest()
        for p in sorted(source.iterdir()) if p.is_file()
    }
    if ordinary_snapshot != expected_snapshot:
        raise RuntimeError(f"ordinary checks changed source/module files for {candidate}")

after = snapshots()
manifest = {
    "scope": "original and two private serial sumPositive code areas from neutral packet",
    "input_sha256_before": before, "input_sha256_after": after,
    "inputs_unchanged": before == after,
    "ordinary_copy_inputs_unchanged": True,
    "observation_case_count_per_candidate": 12,
    "independent_go_1_22_execution": False,
    "go_1_22_evidence": "supplied checks-A.json/checks-B.json plus source syntax and unchanged go.mod",
}
(OUT / "source-manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
if before != after:
    raise RuntimeError("source input changed")
print("Inputs preserved; independent verification complete.", flush=True)
