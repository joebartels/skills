import json
from pathlib import Path
import shutil
import subprocess
import time

ROOT = Path(__file__).resolve().parent.parent
OUTPUT = ROOT / 'output'
CHECKS = json.loads((OUTPUT / 'checks.json').read_text())
ENV = {'GOCACHE': '/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN': 'local'}

def run(label, cwd, args, expected=0):
    command = ['rtk', 'proxy', 'env'] + [f'{k}={v}' for k, v in ENV.items()] + args
    started = time.monotonic()
    result = subprocess.run(command, cwd=cwd, text=True, capture_output=True, timeout=60)
    entry = {'label': label, 'cwd': str(cwd), 'command': command, 'environment_overrides': ENV,
             'stdout': result.stdout, 'stderr': result.stderr, 'exit_code': result.returncode,
             'expected_exit_code': expected, 'elapsed_seconds': round(time.monotonic()-started, 3)}
    CHECKS.append(entry)
    (OUTPUT / 'checks.json').write_text(json.dumps(CHECKS, indent=2)+'\n')
    print(json.dumps(entry), flush=True)
    if result.returncode != expected:
        raise SystemExit('Unexpected check outcome')

for version, expected in [('original', 0), ('candidate', 1)]:
    target = ROOT / ('verification-nontruncating-' + version)
    shutil.copytree(ROOT / version, target)
    path = target / 'store.go'
    source = path.read_text().replace('return os.WriteFile(filepath.Join(s.root, key), []byte(value), 0o600)',
        'f, err := os.OpenFile(filepath.Join(s.root, key), os.O_CREATE|os.O_WRONLY, 0o600)\n\tif err != nil { return err }; defer f.Close()\n\t_, err = f.WriteString(value); return err')
    path.write_text(source)
    run('mutation nontruncating Put ' + version, target, ['go', 'test', '-timeout=30s', '-v', './...'], expected=expected)
