import json, os, pathlib, subprocess, sys, time
out = pathlib.Path(__file__).parent
label, cwd, timeout = sys.argv[1:4]
argv = sys.argv[4:]
env = os.environ.copy()
env["GOCACHE"] = "/private/tmp/go-quality-testing-cache"
env["GOTOOLCHAIN"] = "local"
start = time.monotonic()
try:
    result = subprocess.run(argv, cwd=cwd, env=env, capture_output=True, text=True, timeout=float(timeout))
    item = dict(label=label, argv=argv, cwd=cwd, environment={"GOCACHE":env["GOCACHE"], "GOTOOLCHAIN":env["GOTOOLCHAIN"]}, stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, elapsed_seconds=round(time.monotonic()-start, 3), timeout_seconds=float(timeout))
except subprocess.TimeoutExpired as exc:
    item = dict(label=label, argv=argv, cwd=cwd, environment={"GOCACHE":env["GOCACHE"], "GOTOOLCHAIN":env["GOTOOLCHAIN"]}, stdout=exc.stdout.decode() if isinstance(exc.stdout,bytes) else exc.stdout, stderr=exc.stderr.decode() if isinstance(exc.stderr,bytes) else exc.stderr, exit_code=None, timed_out=True, elapsed_seconds=round(time.monotonic()-start, 3), timeout_seconds=float(timeout))
checks = json.loads((out / "checks.json").read_text())
checks.append(item)
(out / "checks.json").write_text(json.dumps(checks, indent=2)+"\n")
print(json.dumps(item), flush=True)
