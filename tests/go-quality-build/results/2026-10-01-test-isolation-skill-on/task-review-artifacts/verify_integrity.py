import hashlib
import json
import re
import shutil
import subprocess
from pathlib import Path

ROOT = Path('/Users/jb/.codex/worktrees/b633/skills')
OUT = Path('/private/tmp/go-isolation-promotion-review-20261001')
CHECKPOINT = '41373f88927cf95841b0a73b6fab1db4a2237b57'
DRAFT_COMMIT = '904511b076a98b3d434186389ed2de2a57c09380'
FIXTURE_COMMIT = '728bb6bb74ca274ea4d5774d30cd04dba93e6fbc'
ARMS = {a: ROOT / ('tests/go-quality-build/results/2026-10-01-test-isolation-' + a) for a in ['baseline', 'skill-on']}
data = {'scope': {'checkpoint': CHECKPOINT, 'draft_commit': DRAFT_COMMIT, 'fixture_commit': FIXTURE_COMMIT}, 'commands': [], 'assertions': [], 'reconstructions': []}

def digest(b):
    return hashlib.sha256(b).hexdigest()

def record(name, ok, details):
    data['assertions'].append({'name': name, 'ok': ok, 'details': details})
    if not ok:
        (OUT/'integrity-failed-attempt.json').write_text(json.dumps(data, indent=2) + '\n')
        raise RuntimeError(name)

def git_bytes(args):
    cmd = ['rtk', 'proxy', 'git'] + args
    r = subprocess.run(cmd, cwd=ROOT, capture_output=True, timeout=30)
    data['commands'].append({'argv': cmd, 'cwd': str(ROOT), 'exit_code': r.returncode, 'stdout_sha256': digest(r.stdout), 'stdout_bytes': len(r.stdout), 'stderr': r.stderr.decode()})
    if r.returncode:
        raise RuntimeError(cmd)
    return r.stdout

manifests = {a: json.loads((p / 'manifest.json').read_text()) for a, p in ARMS.items()}
for arm, archive in ARMS.items():
    m = manifests[arm]
    entries = []
    for line in (archive / 'checksums.sha256').read_text().splitlines():
        expected, rel = line.split('  ', 1)
        entries.append(rel)
        record(arm + ':checksum:' + rel, digest((archive / rel).read_bytes()) == expected, {'sha256': expected})
    files = sorted(str(p.relative_to(archive)) for p in archive.rglob('*') if p.is_file() and p.name != 'checksums.sha256')
    record(arm + ':complete-index', sorted(entries) == files, {'indexed_count': len(entries), 'actual_count': len(files), 'index_sha256': digest((archive/'checksums.sha256').read_bytes())})
    record(arm + ':manifest-hashes', all(digest((archive/r).read_bytes()) == h for r,h in m['artifact_sha256'].items()), {'count': len(m['artifact_sha256'])})
    tracked = git_bytes(['ls-tree', '-r', '--name-only', CHECKPOINT, '--', str(archive.relative_to(ROOT))]).decode().splitlines()
    record(arm + ':checkpoint-path-inventory', sorted(str(ROOT/r) for r in tracked) == sorted(str(p) for p in archive.rglob('*') if p.is_file()), {'count': len(tracked)})
    for rel in tracked:
        blob = git_bytes(['show', CHECKPOINT + ':' + rel])
        record('checkpoint-equal:' + rel, blob == (ROOT/rel).read_bytes(), {'sha256': digest(blob)})
    for c in m['cases']:
        case = c['id']
        original = archive/case/'original'
        candidate = archive/case/'candidate'
        for rel,h in c['input_sha256'].items():
            blob = git_bytes(['show', FIXTURE_COMMIT + ':tests/go-quality-build/go-test-isolation/evals/files/' + case + '/' + rel])
            record(arm + ':frozen-input:' + case + '/' + rel, blob == (original/rel).read_bytes() and digest(blob) == h, {'sha256': h})
        reconstruction = OUT/'reconstructed'/arm/case
        if reconstruction.exists():
            shutil.rmtree(reconstruction)
        shutil.copytree(original, reconstruction, dirs_exist_ok=False)
        cmd = ['rtk', 'proxy', 'git', 'apply', str(archive/case/'source.patch')]
        r = subprocess.run(cmd, cwd=reconstruction, capture_output=True, text=True, timeout=30)
        data['commands'].append({'argv': cmd, 'cwd': str(reconstruction), 'stdout': r.stdout, 'stderr': r.stderr, 'exit_code': r.returncode})
        actual = {str(p.relative_to(reconstruction)): digest(p.read_bytes()) for p in reconstruction.rglob('*') if p.is_file()}
        archived = {str(p.relative_to(candidate)): digest(p.read_bytes()) for p in candidate.rglob('*') if p.is_file()}
        record(arm + ':reconstruction:' + case, r.returncode == 0 and actual == archived == c['source_sha256'], {'files': actual})
        data['reconstructions'].append({'arm': arm, 'case': case, 'cwd': str(reconstruction), 'files': actual})

base, on = manifests['baseline'], manifests['skill-on']
record('matched-author-settings', base['author_settings'] == on['author_settings'], base['author_settings'])
record('matched-review-settings', base['review_settings'] == on['review_settings'], base['review_settings'])
record('matched-host', base['host'] == on['host'], base['host'])
record('matched-existing-guidance', all(on['skill_sha256'][r] == h and (ARMS['baseline']/'skills'/r).read_bytes() == (ARMS['skill-on']/'skills'/r).read_bytes() for r,h in base['skill_sha256'].items()), base['skill_sha256'])
record('behavior-guidance-unavailable', all(not (p/'skills/go-behavior-tests').exists() for p in ARMS.values()), 'No behavior snapshot in either archive')
for c in base['cases']:
    case = c['id']
    peer = next(x for x in on['cases'] if x['id'] == case)
    record('matched-inputs:' + case, c['input_sha256'] == peer['input_sha256'], c['input_sha256'])
    record('matched-prompt:' + case, (ARMS['baseline']/case/'prompt.txt').read_bytes() == (ARMS['skill-on']/case/'prompt.txt').read_bytes(), digest((ARMS['baseline']/case/'prompt.txt').read_bytes()))
    def normalized_dispatch(a):
        text = (ARMS[a]/case/'dispatch.txt').read_text()
        text = re.sub('/private/tmp/go-testing-author-[^/ ]+', '/private/tmp/MATCHED_AUTHOR', text)
        text = '\n'.join(line for line in text.splitlines() if not line.startswith('- go-test-isolation:'))
        text = text.replace('The two new testing-writing skills are unavailable in this task.', 'The other new testing-writing skill is unavailable in this task.')
        return text
    record('matched-wrapper:' + case, normalized_dispatch('baseline') == normalized_dispatch('skill-on'), 'Ignoring only ephemeral author root and the isolation description/path offer')
    for a in ARMS:
        launch = (ARMS[a]/case/'launch.txt').read_text()
        record(a + ':fresh-launch:' + case, 'one dispatch file is authorized input' in launch, {'launch': launch})

draft = ROOT/'docs/go-quality-build/drafts/go-test-isolation'
for p in draft.rglob('*'):
    if not p.is_file():
        continue
    rel = str(p.relative_to(draft))
    blob = git_bytes(['show', DRAFT_COMMIT + ':' + str(p.relative_to(ROOT))])
    record('draft-snapshot-source:' + rel, blob == p.read_bytes() == (ARMS['skill-on']/'skills/go-test-isolation'/rel).read_bytes(), {'sha256': digest(blob)})
record('runtime-unpromoted', not (ROOT/'plugins/go-quality-build/skills/go-test-isolation').exists(), 'Runtime destination absent at review start')
links = re.findall(r'\[[^\]]+\]\(([^)]+)\)', (draft/'SKILL.md').read_text())
record('portable-main-links', all((draft/l).is_file() for l in links), links)
(OUT/'integrity.json').write_text(json.dumps(data, indent=2) + '\n')
print(json.dumps({'assertions': len(data['assertions']), 'archives': {a:len(m['artifact_sha256'])+1 for a,m in manifests.items()}, 'reconstructions':len(data['reconstructions']), 'source_files':sum(len(r['files']) for r in data['reconstructions']), 'all_ok':all(x['ok'] for x in data['assertions'])}))
