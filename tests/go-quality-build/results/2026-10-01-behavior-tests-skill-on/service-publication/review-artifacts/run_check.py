import json, os, pathlib, shlex, subprocess, sys, time
base = pathlib.Path(__file__).parent
label, cwd, seconds = sys.argv[1:4]
argv = sys.argv[4:]
started = time.monotonic()
try:
    result = subprocess.run(argv, cwd=cwd, capture_output=True, text=True, timeout=float(seconds))
    record = dict(label=label, cwd=cwd, argv=argv, command=shlex.join(argv), stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, timed_out=False)
except subprocess.TimeoutExpired as error:
    record = dict(label=label, cwd=cwd, argv=argv, command=shlex.join(argv), stdout=(error.stdout or b"").decode() if isinstance(error.stdout, bytes) else error.stdout or "", stderr=(error.stderr or b"").decode() if isinstance(error.stderr, bytes) else error.stderr or "", exit_code=None, timed_out=True)
record["wall_seconds"] = time.monotonic() - started
path = base / "checks.json"
records = json.loads(path.read_text()) if path.exists() else []
records.append(record)
path.write_text(json.dumps(records, indent=2) + "\n")
print(json.dumps(record))
sys.exit(0 if record["exit_code"] is None else record["exit_code"])

