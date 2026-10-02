import json
import os
import pathlib
import shutil
import subprocess
import time

base = pathlib.Path(__file__).resolve().parents[1]
output = base / 'output'
env = dict(os.environ, GOCACHE='/private/tmp/go-quality-testing-cache', GOTOOLCHAIN='local', GOPROXY='off')
records = []

def run(label, command, cwd, timeout=60):
    started = time.time()
    try:
        result = subprocess.run(command, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout)
        record = dict(label=label, argv=command, cwd=str(cwd), env={k:env[k] for k in ['GOCACHE', 'GOTOOLCHAIN', 'GOPROXY']}, stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode, elapsed_seconds=round(time.time()-started, 3))
    except subprocess.TimeoutExpired as exc:
        record = dict(label=label, argv=command, cwd=str(cwd), env={k:env[k] for k in ['GOCACHE', 'GOTOOLCHAIN', 'GOPROXY']}, stdout=exc.stdout.decode() if isinstance(exc.stdout, bytes) else exc.stdout or '', stderr=exc.stderr.decode() if isinstance(exc.stderr, bytes) else exc.stderr or '', exit_code=None, timeout_seconds=timeout, elapsed_seconds=round(time.time()-started, 3))
    records.append(record)
    (output/'checks.json').write_text(json.dumps(records, indent=2)+'\n')
    print(json.dumps(record), flush=True)
    return record

run('runtime', ['rtk','proxy','go','version'], base/'candidate')
run('platform', ['rtk','proxy','go','env','GOOS','GOARCH','GOVERSION','GOTOOLCHAIN'], base/'candidate')
run('candidate full suite', ['rtk','proxy','go','test','./...','-count=1','-timeout=30s'], base/'candidate')
run('candidate race shuffled repeated', ['rtk','proxy','go','test','-race','./...','-count=20','-shuffle=on','-timeout=60s'], base/'candidate')
run('candidate selected group child', ['rtk','proxy','go','test','./...','-run','^TestStoreIndependentInstances$/^instances$/^gamma$/^repeat$','-count=1','-v','-timeout=30s'], base/'candidate')
for label, old, new in [
    ('root ignored', 'func New(root string) *Store { return &Store{root: root} }', 'func New(root string) *Store { root = "/private/tmp/go-independent-review-46wnhnrc/changeset-2/verification-root-ignored/shared-root"; _ = os.MkdirAll(root, 0o700); return &Store{root: root} }'),
    ('replacement appends', 'return os.WriteFile(filepath.Join(s.root, key), []byte(value), 0o600)', 'file, err := os.OpenFile(filepath.Join(s.root, key), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)\n\tif err != nil { return err }\n\tdefer file.Close()\n\t_, err = file.WriteString(value)\n\treturn err'),
]:
    target = base/('verification-'+label.replace(' ', '-'))
    shutil.copytree(base/'candidate', target, dirs_exist_ok=True)
    source = target/'store.go'
    original = source.read_text()
    assert original.count(old) == 1, label
    source.write_text(original.replace(old, new))
    run('mutation: '+label, ['rtk','proxy','go','test','./...','-run','^TestStoreIndependentInstances$','-count=1','-timeout=30s'], target)
