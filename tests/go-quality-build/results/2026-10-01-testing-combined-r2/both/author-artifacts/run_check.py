import json, os, subprocess, sys, time
from pathlib import Path
root = Path(__file__).parent
command = sys.argv[1:]
env = dict(os.environ, GOCACHE="/private/tmp/go-quality-testing-cache", GOTOOLCHAIN="local")
started = time.time()
try:
 result = subprocess.run(command, cwd="/private/tmp/go-fresh-author-qjdzm0iq/task-2/module", env=env, text=True, capture_output=True, timeout=90)
 entry = {"command":command,"cwd":"/private/tmp/go-fresh-author-qjdzm0iq/task-2/module","environment":{"GOCACHE":env["GOCACHE"],"GOTOOLCHAIN":env["GOTOOLCHAIN"]},"stdout":result.stdout,"stderr":result.stderr,"exit_code":result.returncode,"elapsed_seconds":time.time()-started}
except subprocess.TimeoutExpired as e:
 entry = {"command":command,"cwd":"/private/tmp/go-fresh-author-qjdzm0iq/task-2/module","environment":{"GOCACHE":env["GOCACHE"],"GOTOOLCHAIN":env["GOTOOLCHAIN"]},"stdout":e.stdout.decode() if isinstance(e.stdout,bytes) else e.stdout or "","stderr":e.stderr.decode() if isinstance(e.stderr,bytes) else e.stderr or "","exit_code":None,"timeout_seconds":90,"elapsed_seconds":time.time()-started}
path = root / "checks.json"
entries = json.loads(path.read_text()) if path.exists() else []
entry["purpose"] = os.environ.get("CHECK_PURPOSE", "verification")
entries.append(entry)
path.write_text(json.dumps(entries,indent=2)+"\n")
print(json.dumps(entry,indent=2))
sys.exit(entry["exit_code"] if entry["exit_code"] is not None else 124)
