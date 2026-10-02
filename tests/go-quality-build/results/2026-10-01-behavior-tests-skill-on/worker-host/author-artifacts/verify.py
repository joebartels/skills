from pathlib import Path
import json
import shutil
import subprocess
import time

module = Path("/private/tmp/go-testing-author-z3h6ca5c/worker-host")
report = Path("/private/tmp/go-testing-author-z3h6ca5c/worker-host-report")
checks = []
go = ["rtk", "proxy", "env", "GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local", "go"]

def run(name, command, expected=0):
    began = time.monotonic()
    record = {"name": name, "command": command, "process_timeout_seconds": 60, "expected_exit_code": expected}
    try:
        result = subprocess.run(command, cwd=module, text=True, capture_output=True, timeout=60)
        record.update(stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode)
    except subprocess.TimeoutExpired as exc:
        record.update(stdout=str(exc.stdout or ""), stderr=str(exc.stderr or ""), exit_code=None, timed_out=True)
    record["elapsed_seconds"] = time.monotonic() - began
    checks.append(record)
    report.joinpath("checks.json").write_text(json.dumps(checks, indent=2) + "\n")
    print(json.dumps({"name": name, "exit_code": record["exit_code"], "stdout": record["stdout"], "stderr": record["stderr"]}), flush=True)
    if record["exit_code"] != expected:
        raise RuntimeError(name + " did not have the expected exit status")

run("format", ["rtk", "proxy", "gofmt", "-w", "host.go", "worker.go", "host_test.go"])
host = module.joinpath("host.go")
worker = module.joinpath("worker.go")
new_host, new_worker = host.read_text(), worker.read_text()
try:
    host.write_text('''package sweeper
import ("context"; "errors"; "time")
func Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error {
    return errors.Join(once(ctx, sweep), release())
}
''')
    worker.write_text('''package sweeper
import "context"
func once(ctx context.Context, sweep func(context.Context) error) error { return sweep(ctx) }
''')
    run("old implementation regression signal", go + ["test", "-v", "-count=1", "-timeout=20s", "-run", "TestInvalidIntervalHasNoEffects|TestAlreadyCanceledSkipsSweepAndReleases|TestWorkerAndReleaseErrorCombinations", "./..."], expected=1)
finally:
    host.write_text(new_host)
    worker.write_text(new_worker)
try:
    worker.write_text(new_worker.replace("\t\tif err := sweep(ctx); err != nil {", "\t\ttimer := time.NewTimer(interval)\n\t\tif err := sweep(ctx); err != nil {").replace("\t\t\treturn err", "\t\t\ttimer.Stop()\n\t\t\treturn err").replace("\n\t\ttimer := time.NewTimer(interval)\n\t\tselect", "\n\t\tselect"))
    run("interval-before-completion counterexample", go + ["test", "-v", "-count=1", "-timeout=20s", "-run", "TestSuccessfulCyclesWaitAfterCompletionAndNeverOverlap", "./..."], expected=1)
finally:
    worker.write_text(new_worker)
run("local toolchain", go + ["version"])
run("final race suite", go + ["test", "-v", "-race", "-count=1", "-timeout=45s", "./..."])
run("repeat and shuffle with race detector", go + ["test", "-race", "-count=10", "-shuffle=on", "-timeout=45s", "./..."])
run("vet", go + ["vet", "./..."])
run("format verification", ["rtk", "proxy", "gofmt", "-l", "host.go", "worker.go", "host_test.go"])
staticcheck = shutil.which("staticcheck")
if staticcheck:
    run("staticcheck", ["rtk", "proxy", "env", "GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local", staticcheck, "./..."])
else:
    checks.append({"name": "staticcheck", "status": "skipped", "reason": "not installed; dispatch forbids installing tools"})
    report.joinpath("checks.json").write_text(json.dumps(checks, indent=2) + "\n")
