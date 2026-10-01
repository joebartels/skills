import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

root = Path('/private/tmp/go-independent-review-138xcq3s')
output = root / 'output'
output.mkdir(exist_ok=True)
checks = []
go_prefix = ['rtk', 'proxy', 'env', 'GOCACHE=/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN=local', 'GOPROXY=off', 'GOSUMDB=off', 'GOFLAGS=-mod=readonly']

def run(label, cwd, args, timeout=45):
    command = go_prefix + args
    started = time.monotonic()
    result = subprocess.run(command, cwd=cwd, capture_output=True, text=True, timeout=timeout)
    entry = {'label': label, 'cwd': str(cwd), 'command': command, 'stdout': result.stdout, 'stderr': result.stderr, 'exit_code': result.returncode, 'duration_seconds': time.monotonic() - started}
    checks.append(entry)
    (output / 'checks.json').write_text(json.dumps(checks, indent=2) + '\n')
    print(json.dumps(entry), flush=True)
    return result

def copy_candidate(name):
    target = root / name
    if target.exists():
        shutil.rmtree(target)
    shutil.copytree(root / 'candidate', target)
    return target

def mutate(name, before, after):
    target = copy_candidate(name)
    path = target / 'key.go'
    source = path.read_text()
    assert before in source, name
    path.write_text(source.replace(before, after))
    checks.append({'label': name + '-mutation', 'file': str(path), 'before': before, 'after': after})
    return target

def hashes(tree):
    return {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tree.rglob('*')) if p.is_file()}

before_hashes = {name: hashes(root / name) for name in ['original', 'candidate']}
run('toolchain', root, ['go', 'version'])
run('platform', root, ['go', 'env', 'GOOS', 'GOARCH', 'GOVERSION', 'CGO_ENABLED'])
run('original-suite', root / 'original', ['go', 'test', '-count=1', '-timeout=30s', './...'])
candidate = copy_candidate('verification-candidate')
run('candidate-suite', candidate, ['go', 'test', '-count=1', '-timeout=30s', './...'])
run('candidate-bounded-fuzz', candidate, ['go', 'test', '-run=^$', '-fuzz=^FuzzRoundTrip$', '-fuzztime=3s', '-parallel=1', '-timeout=20s', './...'], timeout=30)

query = mutate('mutation-query-codec', 'url.PathEscape', 'url.QueryEscape')
path = query / 'key.go'
path.write_text(path.read_text().replace('url.PathUnescape', 'url.QueryUnescape'))
checks[-1]['additional_replacement'] = ['url.PathUnescape', 'url.QueryUnescape']
run('mutation-query-roundtrip-control', query, ['go', 'test', '-run=^(FuzzRoundTrip|TestSimpleRoundTrip)$', '-count=1', '-timeout=30s', './...'])
run('mutation-query-exact-tables', query, ['go', 'test', '-run=^(TestEncode|TestDecode)$', '-count=1', '-timeout=30s', './...'])

sentinel = mutate('mutation-sentinel', 'return "", ErrInvalidKey', 'return "", errors.New("invalid key")')
path = sentinel / 'key.go'
path.write_text(path.read_text().replace('return Key{}, ErrInvalidKey', 'return Key{}, errors.New("invalid key")'))
checks[-1]['additional_replacement'] = ['return Key{}, ErrInvalidKey', 'return Key{}, errors.New("invalid key")']
run('mutation-error-identity', sentinel, ['go', 'test', '-run=^(TestEncodeRejectsEmptyFields|TestDecodeRejectsInvalidWire)$', '-count=1', '-timeout=30s', './...'])

partial = mutate('mutation-partial-key', 'name, err := url.PathUnescape(nameWire)\n\tif err != nil {\n\t\treturn Key{}, ErrInvalidKey', 'name, err := url.PathUnescape(nameWire)\n\tif err != nil {\n\t\treturn Key{Region: region}, ErrInvalidKey')
run('mutation-zero-result', partial, ['go', 'test', '-run=^TestDecodeRejectsInvalidWire/invalid_name_after_valid_region_escape$', '-count=1', '-timeout=30s', './...'])

bytes_mutation = mutate('mutation-invalid-utf8', 'url.PathEscape(k.Region)', 'url.PathEscape(strings.ToValidUTF8(k.Region, "?"))')
run('mutation-bytes-table-control', bytes_mutation, ['go', 'test', '-run=^TestEncode$', '-count=1', '-timeout=30s', './...'])
run('mutation-bytes-fuzz-seed', bytes_mutation, ['go', 'test', '-run=^FuzzRoundTrip$', '-count=1', '-timeout=30s', './...'])

after_hashes = {name: hashes(root / name) for name in ['original', 'candidate']}
checks.append({'label': 'source-preservation', 'before_sha256': before_hashes, 'after_sha256': after_hashes, 'unchanged': before_hashes == after_hashes})
assert before_hashes == after_hashes
(output / 'checks.json').write_text(json.dumps(checks, indent=2) + '\n')
print('Source and configuration preservation verified.', flush=True)
