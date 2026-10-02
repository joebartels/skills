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
run('baseline race repeated shuffled hostile parent', candidate,
    ['go', 'test', '-race', '-count=3', '-shuffle=on', '-timeout=60s', './...'],
    env={'INDEXER_ENDPOINT': 'not a URL from parent', 'INDEXER_MODE': 'bad parent mode'})
run('focused defaults present-empty parent', candidate, ['go', 'test', '-run=^TestLoadFromEnv$/^defaults$', '-timeout=30s', '-v', './...'],
    env={'INDEXER_ENDPOINT': '', 'INDEXER_MODE': ''})
run('focused command malformed', candidate, ['go', 'test', '-run=^TestShowcfgCommand$/^malformed_endpoint$', '-timeout=60s', '-v', './...'])
run('language stdlib compatibility vet', candidate, ['go', 'vet', '-stdversion', './...'])

leak = copy('loader-env-leak')
path = leak / 'config.go'
source = path.read_text().replace('return Config{Endpoint: endpoint, Mode: mode}, nil',
    'os.Setenv("INDEXER_ENDPOINT", endpoint)\n\tos.Setenv("INDEXER_MODE", mode)\n\treturn Config{Endpoint: endpoint, Mode: mode}, nil')
path.write_text(source)
run('mutation loader environment leak', leak, ['go', 'test', '-run=^TestLoadFromEnv$/^defaults$', '-timeout=30s', '-v', './...'], expected=1)

newline = copy('command-no-newline')
path = newline / 'cmd/showcfg/main.go'
source = path.read_text().replace('if err := json.NewEncoder(os.Stdout).Encode(config); err != nil {',
    'data, marshalErr := json.Marshal(config)\n\tif marshalErr != nil { panic(marshalErr) }\n\tif _, err := os.Stdout.Write(data); err != nil {')
path.write_text(source)
run('mutation command missing newline', newline, ['go', 'test', '-run=^TestShowcfgCommand$/^defaults$', '-timeout=60s', '-v', './...'], expected=1)

exitcode = copy('command-exit-one')
path = exitcode / 'cmd/showcfg/main.go'
source = path.read_text().replace('os.Exit(2)', 'os.Exit(1)')
path.write_text(source)
run('mutation command wrong exit', exitcode, ['go', 'test', '-run=^TestShowcfgCommand$/^empty_mode$', '-timeout=60s', '-v', './...'], expected=1)
