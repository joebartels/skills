#!/usr/bin/env python3
"""Finite independent checks in disposable reconstructed modules."""
import argparse
import json
import os
from pathlib import Path
import re
import shutil
import sys
import tempfile

sys.path.insert(0, '/private/tmp/go-isolation-r3-promotion')
from verify import ROOT, OUT, RESULTS, RUNS, call, digest, inventory

def main(mode):
    env = dict(os.environ, GOCACHE=str(OUT / 'cache'), GOTOOLCHAIN='local')
    records = []
    def run(args, cwd, chosen=None, label=None, limit=180):
        r = call(args, cwd, chosen or env, limit)
        e = chosen or env
        r['env'] = {'GOCACHE': e.get('GOCACHE', '<unset>'), 'GOTOOLCHAIN': e.get('GOTOOLCHAIN'), 'HOME': e.get('HOME', '<unset>')}
        if label:
            r['label'] = label
        records.append(r)
        (OUT / ('executed-' + mode + '.json')).write_text(json.dumps(records, indent=2) + '\n')
        print(label or args, r['exit_code'], flush=True)
        return r
    if mode in ('basic-restricted', 'listener-rerun'):
        for name in RUNS:
            m = json.loads((RESULTS / name / 'manifest.json').read_text())
            for c in m['cases']:
                work = OUT / 'reconstructed' / name / c['id']
                prefix = name + '/' + c['id']
                if mode == 'listener-rerun' and c['id'] not in ('service-http-boundary', 'isolation-only', 'both'):
                    continue
                run(['go', 'test', '-count=1', '-timeout=60s', './...'], work, label=prefix + ' clean')
                if mode == 'basic-restricted':
                    run(['go', 'vet', './...'], work, label=prefix + ' vet')
                    run(['gofmt', '-l', '.'], work, label=prefix + ' format')
                if c['id'] != 'not-isolation-work':
                    run(['go', 'test', '-race', '-shuffle=20261001', '-count=3', '-timeout=90s', './...'], work, label=prefix + ' race/shuffle/three')
    if mode == 'focused':
        for case, leaf in [('library-file-fixtures', 'first'), ('file-repeat', 'initial')]:
            work = OUT / 'reconstructed' / RUNS[0] / case
            prefix = '^TestStoreIndependentInstances$/^instances$/^alpha$/^' + leaf + '$'
            run(['go', 'test', '-race', '-run', prefix, '-count=1', '-timeout=30s', './...'], work, label=case + ' focused leaf')
            with tempfile.TemporaryDirectory(prefix='shared-root-', dir=OUT) as tmp:
                mutant = Path(tmp) / 'module'
                shutil.copytree(work, mutant)
                patch = RESULTS / RUNS[0] / case / 'mutations/store-shared-root/source.patch'
                run(['git', 'apply', str(patch)], mutant, label=case + ' shared-root apply')
                run(['go', 'test', '-run', '^$', '-timeout=30s', './...'], mutant, label=case + ' shared-root compile')
                run(['go', 'test', '-parallel=1', '-count=1', '-timeout=30s', './...'], mutant, label=case + ' shared-root parallel1 assertion')
        reference = ROOT / 'docs/go-quality-build/drafts/go-test-isolation/references/isolation-patterns.md'
        illustration = re.findall(r'```go\n(.*?)```', reference.read_text(), re.S)[0]
        example = OUT / 'illustration'
        example.mkdir(exist_ok=True)
        (example / 'go.mod').write_text('module example.com/illustration\n\ngo 1.22\n')
        (example / 'illustration_test.go').write_text(illustration)
        run(['go', 'test', '-race', '-count=3', '-timeout=30s', './...'], example, label='unchanged Go1.22 illustration on host')
        run(['go', 'vet', './...'], example, label='illustration vet')
    if mode == 'defaults':
        home = OUT / 'valid-default-home'
        home.mkdir(exist_ok=True)
        runtime_env = dict(env, HOME=str(home))
        runtime_env.pop('GOCACHE', None)
        for name, case in [(RUNS[0], 'cli-environment'), (RUNS[1], 'isolation-only'), (RUNS[1], 'both')]:
            work = OUT / 'reconstructed' / name / case
            listed = run(['go', 'list', '-f', '{{if .TestGoFiles}}{{.ImportPath}}{{else if .XTestGoFiles}}{{.ImportPath}}{{end}}', './...'], work, label=case + ' package list')
            for i, package in enumerate(listed['stdout'].split()):
                exe = OUT / ('default-tests-' + case + '-' + str(i))
                run(['go', 'test', '-c', '-o', str(exe), package], work, label=case + ' compile unchanged tests')
                location = run(['go', 'list', '-f', '{{.Dir}}', package], work, label=case + ' package cwd')
                run([str(exe), '-test.count=1', '-test.timeout=60s'], Path(location['stdout'].strip()), runtime_env, case + ' unset GOCACHE valid HOME full unchanged binary')
    if mode == 'selected-mutants':
        for case in ('isolation-only', 'both'):
            work = OUT / 'reconstructed' / RUNS[1] / case
            for semantic in ('automatic-redirect-following', 'accept-all-2xx', 'unsafe-cancellation-cause-comparison'):
                with tempfile.TemporaryDirectory(prefix='mutant-', dir=OUT) as tmp:
                    mutant = Path(tmp) / 'module'
                    shutil.copytree(work, mutant)
                    patch = RESULTS / RUNS[1] / case / 'mutations' / semantic / 'source.patch'
                    run(['git', 'apply', str(patch)], mutant, label=case + '/' + semantic + ' apply')
                    run(['go', 'test', '-run', '^$', '-timeout=30s', './...'], mutant, label=case + '/' + semantic + ' compile')
                    run(['go', 'test', '-count=1', '-timeout=60s', './...'], mutant, label=case + '/' + semantic + ' candidate tests only')
    run(['go', 'version'], ROOT, label='actual toolchain')

if __name__ == '__main__':
    p = argparse.ArgumentParser()
    p.add_argument('mode', choices=['basic-restricted', 'listener-rerun', 'focused', 'defaults', 'selected-mutants'])
    main(p.parse_args().mode)
