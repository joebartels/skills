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
run('candidate full suite', ['rtk','proxy','go','test','./...','-count=1','-timeout=40s'], base/'candidate')
run('candidate race shuffled repeated', ['rtk','proxy','go','test','-race','./...','-count=10','-shuffle=on','-timeout=60s'], base/'candidate')
for label, old, new, selection in [
    ('missing response close', 'defer response.Body.Close()', '// mutation: do not close response.Body', '^TestFetchResponse$/^success$'),
    ('ignored supplied context', 'http.NewRequestWithContext(ctx,', 'http.NewRequestWithContext(context.Background(),', '^TestFetchCancellation$'),
    ('trailing JSON accepted', 'json.Unmarshal(data, &result)', 'json.NewDecoder(strings.NewReader(string(data))).Decode(&result)', '^TestFetchResponse$/^trailing_'),
]:
    target = base/('verification-'+label.replace(' ', '-'))
    shutil.copytree(base/'candidate', target, dirs_exist_ok=True)
    source = target/'fetch.go'
    original = source.read_text()
    assert original.count(old) == 1, label
    source.write_text(original.replace(old, new))
    run('mutation: '+label, ['rtk','proxy','go','test','./...','-run',selection,'-count=1','-timeout=35s'], target, timeout=50)
