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
source = (base / "candidate" / "poll.go").read_text()
wait = """\t\ttimer := time.NewTimer(interval)
\t\tselect {
\t\tcase <-ctx.Done():
\t\t\ttimer.Stop()
\t\t\treturn nil
\t\tcase <-timer.C:
\t\t}
"""
callback = """\t\tif err := callback(ctx); err != nil {
\t\t\treturn err
\t\t}
"""
async_callback = """\t\tcallbackResult := make(chan error, 1)
\t\tgo func() { callbackResult <- callback(ctx) }()
\t\tselect {
\t\tcase err := <-callbackResult:
\t\t\tif err != nil {
\t\t\t\treturn err
\t\t\t}
\t\tcase <-ctx.Done():
\t\t\treturn nil
\t\t}
"""
mutations = [
    (
        "fixed-rate",
        source.replace("\tfor {\n", "\tticker := time.NewTicker(interval)\n\tdefer ticker.Stop()\n\tfor {\n", 1).replace(wait, "\t\tselect {\n\t\tcase <-ctx.Done():\n\t\t\treturn nil\n\t\tcase <-ticker.C:\n\t\t}\n"),
        "^TestRecurrence$",
        "Replace completion-relative waiting with a ticker started before callback 1.",
    ),
    (
        "early-return",
        source.replace(callback, async_callback),
        "^TestFixtureTeardown$",
        "Return on cancellation while the active callback remains in blocked cleanup.",
    ),
    (
        "serialized-invocations",
        source.replace('"errors"', '"errors"\n\t"sync"').replace("// Poll runs", "var pollMu sync.Mutex\n\n// Poll runs").replace("\tif interval <= 0", "\tpollMu.Lock()\n\tdefer pollMu.Unlock()\n\tif interval <= 0", 1),
        "^TestIndependentInvocations$/^callback_error$",
        "Serialize all invocations behind one package-level mutex.",
    ),
    (
        "cancellation-hides-error",
        source.replace(callback, "\t\tif err := callback(ctx); err != nil {\n\t\t\tif ctx.Err() != nil { return nil }\n\t\t\treturn err\n\t\t}\n"),
        "^TestCallbackErrorDuringCancellation$",
        "Suppress the callback error if its context has also been canceled.",
    ),
    (
        "ignores-wait-cancellation",
        source.replace(wait, "\t\ttimer := time.NewTimer(interval)\n\t\t<-timer.C\n"),
        "^TestCancellation$",
        "Ignore cancellation while waiting for the next cycle.",
    ),
]

for name, mutated, test, purpose in mutations:
    assert mutated != source, name
    probe = base / "probes" / name
    shutil.copytree(base / "candidate", probe)
    (probe / "poll.go").write_text(mutated)
    args = ["rtk", "proxy", "env"] + [f"{k}={v}" for k, v in env.items()] + [
        "go", "test", "./...", "-run", test, "-count=1", "-timeout=20s"
    ]
    run = subprocess.run(args, cwd=probe, text=True, capture_output=True, timeout=30)
    record = {
        "id": "mutation-" + name,
        "purpose": purpose,
        "command": args,
        "cwd": str(probe),
        "env_overrides": env,
        "stdout": run.stdout,
        "stderr": run.stderr,
        "exit_code": run.returncode,
        "expected_exit": "nonzero test assertion failure",
        "mutated_source": str(probe / "poll.go"),
    }
    records.append(record)
    checks_path.write_text(json.dumps(records, indent=2) + "\n")
    print(json.dumps(record), flush=True)

records.append({"id": "source-preservation", "before": before, "after": hashes(), "unchanged": before == hashes()})
checks_path.write_text(json.dumps(records, indent=2) + "\n")
