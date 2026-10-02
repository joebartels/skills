#!/usr/bin/env python3
"""Independent read-only evidence audit; all writable work lives beside this file."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time

ROOT = Path('/Users/jb/.codex/worktrees/b633/skills')
OUT = Path('/private/tmp/go-isolation-r3-promotion')
RESULTS = ROOT / 'tests/go-quality-build/results'
TARGET = '17e38eb69c00302c23eb64ef147924cf7196b0dc'
RUNS = ['2026-10-01-test-isolation-r3', '2026-10-01-testing-combined-r3']
HISTORY = ['2026-10-01-test-isolation-baseline', '2026-10-01-test-isolation-skill-on',
           '2026-10-01-test-isolation-r2', '2026-10-01-testing-combined',
           '2026-10-01-testing-combined-r2', '2026-10-01-testing-combined-final']

def digest(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

def inventory(p):
    return {x.relative_to(p).as_posix(): digest(x) for x in p.rglob('*') if x.is_file()}

def call(args, cwd=ROOT, env=None, limit=180):
    start = time.monotonic()
    try:
        p = subprocess.run(['rtk', 'proxy'] + args, cwd=cwd, env=env,
                           text=True, capture_output=True, timeout=limit)
        return {'command': ['rtk', 'proxy'] + args, 'cwd': str(cwd),
                'exit_code': p.returncode, 'stdout': p.stdout, 'stderr': p.stderr,
                'elapsed_seconds': round(time.monotonic() - start, 3)}
    except subprocess.TimeoutExpired as e:
        return {'command': ['rtk', 'proxy'] + args, 'cwd': str(cwd),
                'exit_code': None, 'timeout_seconds': limit,
                'stdout': str(e.stdout or ''), 'stderr': str(e.stderr or '')}

def audit(sealed=False):
    data = {'target': TARGET, 'phase': 'sealed' if sealed else 'preparation',
            'runs': [], 'history_seals': [], 'failures': []}
    def check(ok, label):
        if not ok:
            data['failures'].append(label)
    resolved = call(['git', 'rev-parse', TARGET[:7]])
    check(resolved['exit_code'] == 0 and resolved['stdout'].strip() == TARGET, 'exact target resolution')
    target_paths = ['SKILL.md', 'references/isolation-patterns.md']
    data['target_sha256'] = {}
    for rel in target_paths:
        local = ROOT / 'docs/go-quality-build/drafts/go-test-isolation' / rel
        committed = call(['git', 'show', TARGET + ':docs/go-quality-build/drafts/go-test-isolation/' + rel])
        check(committed['exit_code'] == 0 and committed['stdout'].encode() == local.read_bytes(), 'target bytes ' + rel)
        data['target_sha256'][rel] = digest(local)
    for name in RUNS:
        d = RESULTS / name
        m = json.loads((d / 'manifest.json').read_text())
        summary = {'run_id': name, 'cases': [], 'mapped_hashes': 0}
        check(m['skill_commit'] == TARGET, name + ' skill commit')
        check(m['revision'] == 3, name + ' revision')
        for rel, sha in m['skill_sha256'].items():
            check(digest(d / 'skills' / rel) == sha, name + ' guidance ' + rel)
            summary['mapped_hashes'] += 1
        for rel in target_paths:
            check(digest(d / 'skills/go-test-isolation' / rel) == data['target_sha256'][rel], name + ' isolation snapshot ' + rel)
        parent = RESULTS / m['control']['parent_run']
        pm = json.loads((parent / 'manifest.json').read_text())
        check(m['author_settings'] == pm['author_settings'], name + ' matched inherited settings')
        for c in m['cases']:
            case = d / c['id']
            guidance = Path(c['workdir']).parent / 'guidance'
            check(guidance.is_dir() and inventory(guidance) == c['guidance_sha256'],
                  name + '/' + c['id'] + ' actual frozen author guidance snapshot')
            check(inventory(case / 'original') == c['input_sha256'], name + '/' + c['id'] + ' original completeness')
            check(inventory(case / 'candidate') == c['source_sha256'], name + '/' + c['id'] + ' candidate completeness')
            fixture_suite = 'testing-combined' if 'combined' in name else 'go-test-isolation'
            for rel, sha in c['input_sha256'].items():
                obj = call(['git', 'show', m['fixture_commit'] + ':tests/go-quality-build/' + fixture_suite + '/evals/files/' + c['fixture_id'] + '/' + rel])
                check(obj['exit_code'] == 0 and hashlib.sha256(obj['stdout'].encode()).hexdigest() == sha,
                      name + '/' + c['id'] + ' frozen commit input ' + rel)
            for rel, sha in c['guidance_sha256'].items():
                check(digest(d / 'skills' / rel) == sha, name + '/' + c['id'] + ' case guidance ' + rel)
            for rel, sha in c['artifact_sha256'].items():
                check(digest(case / rel) == sha, name + '/' + c['id'] + ' mapped artifact ' + rel)
            check(digest(case / 'dispatch.txt') == c['dispatch_sha256'], name + '/' + c['id'] + ' exact dispatch')
            summary['mapped_hashes'] += len(c['input_sha256']) + len(c['source_sha256']) + len(c['guidance_sha256']) + len(c['artifact_sha256']) + 1
            original_parent = parent / (c['fixture_id'] if 'combined' not in name else c['id'])
            if original_parent.exists():
                check((original_parent / 'prompt.txt').read_bytes() == (case / 'prompt.txt').read_bytes(), name + '/' + c['id'] + ' matched prompt')
                check(inventory(original_parent / 'original') == c['input_sha256'], name + '/' + c['id'] + ' matched original')
            for rel, sha in c['guidance_sha256'].items():
                if rel.startswith(('go-test-isolation/', 'go-behavior-tests/')):
                    continue
                check(digest(parent / 'skills' / rel) == sha, name + '/' + c['id'] + ' matched common guidance ' + rel)
            available = c['available_skills']
            check(('go-behavior-tests' in available) == (c['id'] == 'both'), name + '/' + c['id'] + ' other testing availability')
            copy = OUT / 'reconstructed' / name / c['id']
            if copy.exists():
                shutil.rmtree(copy)
            copy.parent.mkdir(parents=True, exist_ok=True)
            shutil.copytree(case / 'original', copy)
            patched = call(['git', 'apply', str(case / 'source.patch')], copy)
            check(patched['exit_code'] == 0, name + '/' + c['id'] + ' patch apply')
            check(inventory(copy) == c['source_sha256'], name + '/' + c['id'] + ' exact reconstruction')
            summary['cases'].append({'id': c['id'], 'reconstruction': patched,
                                     'source_files': len(c['source_sha256']),
                                     'input_files': len(c['input_sha256']),
                                     'author_settings': m['author_settings']})
        data['runs'].append(summary)
    for name in HISTORY + (['2026-10-01-test-isolation-r3'] if sealed else []):
        d = RESULTS / name
        index = d / 'checksums.sha256'
        expected = {}
        for line in index.read_text().splitlines():
            sha, rel = line.split('  ', 1)
            expected[rel] = sha
            check((d / rel).is_file() and digest(d / rel) == sha, name + ' sealed artifact ' + rel)
        actual = inventory(d)
        actual.pop('checksums.sha256', None)
        check(set(actual) == set(expected), name + ' full sealed inventory')
        anchor = digest(index)
        text = (ROOT / 'docs/go-quality-build/README.md').read_text()
        check(anchor in text, name + ' external index anchor')
        data['history_seals'].append({'run_id': name, 'index_sha256': anchor, 'artifacts': len(expected), 'external_anchor': anchor in text})
    data['integrated_seal'] = 'unsealed by design; delivery/whole-branch artifacts pending'
    runtime = ROOT / 'plugins/go-quality-build/skills/go-test-isolation'
    r1 = RESULTS / '2026-10-01-test-isolation-skill-on/skills/go-test-isolation'
    data['runtime_remains_revision1'] = inventory(runtime) == inventory(r1)
    check(data['runtime_remains_revision1'], 'runtime isolation remains revision1')
    behavior = ROOT / 'plugins/go-quality-build/skills/go-behavior-tests'
    current_behavior = RESULTS / '2026-10-01-testing-combined-r3/skills/go-behavior-tests'
    check(inventory(behavior) == inventory(current_behavior), 'both exact promoted behavior2')
    data['unchanged_go_illustration'] = re.findall(r'```go\n(.*?)```', (runtime / 'references/isolation-patterns.md').read_text(), re.S) == re.findall(r'```go\n(.*?)```', (ROOT / 'docs/go-quality-build/drafts/go-test-isolation/references/isolation-patterns.md').read_text(), re.S)
    check(data['unchanged_go_illustration'], 'unchanged Go illustration')
    main = ROOT / 'docs/go-quality-build/drafts/go-test-isolation/SKILL.md'
    data['local_links'] = {rel: (main.parent / rel).is_file() for rel in re.findall(r'\]\(([^)]+)\)', main.read_text())}
    check(all(data['local_links'].values()), 'portable local links')
    data['passed'] = not data['failures']
    (OUT / ('integrity-final.json' if sealed else 'integrity-preparation.json')).write_text(json.dumps(data, indent=2) + '\n')
    print(json.dumps({'passed': data['passed'], 'failures': data['failures'], 'runs': [{k:v for k,v in x.items() if k != 'cases'} for x in data['runs']], 'history_seals': data['history_seals']}, indent=2))

if __name__ == '__main__':
    p = argparse.ArgumentParser()
    p.add_argument('--sealed', action='store_true')
    a = p.parse_args()
    OUT.mkdir(parents=True, exist_ok=True)
    audit(a.sealed)
