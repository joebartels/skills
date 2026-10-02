import json
import os
from pathlib import Path
import shutil
import subprocess
import time

ROOT = Path(__file__).resolve().parent.parent
OUTPUT = ROOT / 'output'
CHECKS = []
ENV = {'GOCACHE': '/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN': 'local'}

def run(label, cwd, args, env=None, expected=0, timeout=90):
    overrides = dict(ENV)
    overrides.update(env or {})
    command = ['rtk', 'proxy', 'env'] + [f'{k}={v}' for k, v in overrides.items()] + args
    started = time.monotonic()
    completed = subprocess.run(command, cwd=cwd, text=True, capture_output=True, timeout=timeout)
    entry = {'label': label, 'cwd': str(cwd), 'command': command, 'environment_overrides': overrides,
             'stdout': completed.stdout, 'stderr': completed.stderr, 'exit_code': completed.returncode,
             'expected_exit_code': expected, 'elapsed_seconds': round(time.monotonic()-started, 3)}
    CHECKS.append(entry)
    (OUTPUT / 'checks.json').write_text(json.dumps(CHECKS, indent=2)+'\n')
    print(json.dumps(entry), flush=True)
    if completed.returncode != expected:
        raise SystemExit(f'Unexpected result: {label}')

def copy(label):
    target = ROOT / ('verification-' + label)
    shutil.copytree(ROOT / 'candidate', target)
    return target

candidate = ROOT / 'candidate'
run('runtime', candidate, ['go', 'version'])
run('baseline race repeated shuffled', candidate, ['go', 'test', '-race', '-count=5', '-shuffle=on', '-timeout=30s', './...'])
run('focused replacement only', candidate, ['go', 'test', '-race', '-run=^TestStoreIndependentInstances$/^instances$/^beta$/^replacement$', '-count=2', '-timeout=30s', '-v', './...'])
run('focused first only', candidate, ['go', 'test', '-run=^TestStoreIndependentInstances$/^instances$/^gamma$/^first$', '-timeout=30s', '-v', './...'])
run('language stdlib compatibility vet', candidate, ['go', 'vet', '-stdversion', './...'])

shared = copy('shared-root')
path = shared / 'store.go'
source = path.read_text().replace('func New(root string) *Store { return &Store{root: root} }',
    'var sharedRoot string\nfunc New(root string) *Store { if sharedRoot == "" { sharedRoot = root }; return &Store{root: sharedRoot} }')
path.write_text(source)
run('mutation shared root', shared, ['go', 'test', '-run=^TestStoreIndependentInstances$', '-parallel=1', '-timeout=30s', '-v', './...'], expected=1)

append = copy('append-put')
path = append / 'store.go'
source = path.read_text().replace('return os.WriteFile(filepath.Join(s.root, key), []byte(value), 0o600)',
    'f, err := os.OpenFile(filepath.Join(s.root, key), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)\n\tif err != nil { return err }; defer f.Close()\n\t_, err = f.WriteString(value); return err')
path.write_text(source)
run('mutation append replacement', append, ['go', 'test', '-run=^TestStoreIndependentInstances$', '-timeout=30s', '-v', './...'], expected=1)
