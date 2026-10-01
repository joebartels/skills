import hashlib
import json
import os
import pathlib
import resource
import shutil
import signal
import subprocess

ROOT = pathlib.Path('/private/tmp/go-independent-review-0rd44ar6')
OUT = ROOT / 'output'
CANDIDATE = ROOT / 'candidate'
checks = []
GO_ENV = ['GOCACHE=/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN=local']

def check(name, args, cwd=CANDIDATE, stdin=None, timeout=60, preexec_fn=None, setup=None):
    command = ['rtk', 'proxy', *args]
    result = subprocess.run(command, cwd=cwd, input=stdin, text=True,
                            capture_output=True, timeout=timeout, preexec_fn=preexec_fn)
    facts = dict(name=name, command=command, cwd=str(cwd), stdin=stdin,
                 stdout=result.stdout, stderr=result.stderr, exit_code=result.returncode,
                 timeout_seconds=timeout)
    if setup is not None:
        facts['setup'] = setup
    checks.append(facts)
    (OUT / 'checks.json').write_text(json.dumps(checks, indent=2) + '\n')
    print(json.dumps(facts), flush=True)
    return result

OUT.mkdir(exist_ok=True)
check('toolchain', ['env', *GO_ENV, 'go', 'version'])
check('platform', ['uname', '-a'])
check('candidate-tests', ['env', *GO_ENV, 'go', 'test', '-count=1', '-timeout=45s', './...'])
check('candidate-order-isolation', ['env', *GO_ENV, 'go', 'test', '-count=3', '-shuffle=on', '-timeout=45s', './...'])
binary = OUT / 'ledgerload'
check('candidate-build', ['env', *GO_ENV, 'go', 'build', '-o', str(binary), '.'])

# A finite local filesystem write failure after a shorter duplicate succeeded.
# Keep the OS file-size limit isolated to the child; the reviewer process is unchanged.
def limited_child():
    signal.signal(signal.SIGXFSZ, signal.SIG_IGN)
    resource.setrlimit(resource.RLIMIT_FSIZE, (3, 3))

limited_dir = OUT / 'limited-records'
limited_dir.mkdir(exist_ok=True)
check('partial-duplicate-write', [str(binary), '--dir', str(limited_dir)],
      stdin='web,1\nweb,12345\nlater,9\n', preexec_fn=limited_child,
      setup='Child ignores SIGXFSZ and has RLIMIT_FSIZE soft=hard=3 bytes; output directory initially empty.')
check('partial-duplicate-result', ['python3', '-c',
      'import json,pathlib; p=pathlib.Path(' + repr(str(limited_dir)) + '); print(json.dumps({f.name:f.read_text() for f in p.iterdir()},sort_keys=True))'])

# Two plausible mutations test independent assertion safeguards.
for name in ['buffered-input', 'discard-write-error']:
    copy = ROOT / ('mutation-' + name)
    shutil.copytree(CANDIDATE, copy, dirs_exist_ok=True)
    path = copy / 'main.go'
    original = path.read_text()
    if name == 'buffered-input':
        mutation = original.replace('"encoding/csv"', '"encoding/csv"\n\t"bytes"')
        mutation = mutation.replace('reader := csv.NewReader(input)',
                                    'all, err := io.ReadAll(input)\n\tif err != nil { return err }\n\treader := csv.NewReader(bytes.NewReader(all))')
        selected = '^TestRunWritesBeforeReadingNextRow$'
    else:
        mutation = original.replace('return fmt.Errorf("record %d: writing %q: %w", record, id, err)', 'continue')
        selected = '^TestRunFileFailureRetainsPrefix$'
    assert mutation != original
    path.write_text(mutation)
    checks.append(dict(name='mutation-setup-' + name, path=str(path),
                       original_sha256=hashlib.sha256(original.encode()).hexdigest(),
                       mutation_sha256=hashlib.sha256(mutation.encode()).hexdigest(),
                       description='Read all stdin before writing' if name == 'buffered-input' else 'Continue after os.WriteFile failure'))
    check('mutation-' + name, ['env', *GO_ENV, 'go', 'test', '-count=1', '-timeout=15s', '-run', selected, '.'], cwd=copy)

check('candidate-source-hashes', ['shasum', '-a', '256', 'README.md', 'go.mod', 'main.go', 'main_test.go'])
