"""Prepare and archive isolated, reproducible Go skill trials (no model runner)."""

import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess

REPO = Path(__file__).resolve().parents[2]
SCRATCH = Path('/private/tmp/go-language-trials-3b23')
RESULTS = REPO / 'tests/go-quality-build/results/2026-10-01-language-contracts'
ARCHITECTURE = ('go-package-boundaries', 'go-api-contracts', 'go-interfaces-and-composition')


def run(args, cwd):
    result = subprocess.run(['rtk', 'proxy', *args], cwd=cwd, text=True, capture_output=True)
    if result.returncode:
        raise RuntimeError(f'{args}: {result.stdout}\n{result.stderr}')
    return result.stdout


def hashes(root):
    return {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in sorted(root.rglob('*')) if p.is_file() and '.git' not in p.parts}


def locations(skill, arm, case, repeat):
    for value in (skill, arm, case, repeat):
        if not value or any(c not in 'abcdefghijklmnopqrstuvwxyz0123456789-' for c in value):
            raise ValueError(f'invalid trial component {value!r}')
    return SCRATCH / skill / arm / case / repeat, RESULTS / skill / arm / case / repeat


def prepare(skill, arm, case, repeat, model, effort):
    trial, archive = locations(skill, arm, case, repeat)
    if trial.exists():
        raise FileExistsError(trial)
    suite = json.loads((REPO / f'tests/go-quality-build/{skill}/evals/evals.json').read_text())
    selected = next(c for c in suite['evals'] if c['id'] == case)
    source = trial / 'source'
    shutil.copytree(REPO / f'tests/go-quality-build/{skill}/evals/files/{case}', source)
    catalog = trial / 'catalog'
    for name in ARCHITECTURE:
        shutil.copytree(REPO / f'plugins/go-quality-build/skills/{name}', catalog / name)
    if arm == 'skill-on':
        shutil.copytree(REPO / f'tests/go-quality-build/{skill}/draft', catalog / skill)
    elif arm != 'baseline':
        raise ValueError('arm must be baseline or skill-on')
    descriptions = []
    for entry in sorted(catalog.glob('*/SKILL.md')):
        text = entry.read_text()
        description = next(line for line in text.splitlines() if line.startswith('description:'))
        descriptions.append(f'{entry.parent.name}: {description[len("description:"):].strip()}\nPath: {entry}')
    prompt = (f'Complete this Go change in {source}.\n\n{selected["prompt"]}\n\n'
              'Read README and existing code/tests. Select only relevant skills from this catalog; '
              'read their SKILL.md if selected. Do not read any other authoring/review skill, '
              'other trial, repository fixture expectation, or result.\n\n' + '\n\n'.join(descriptions) +
              f'\n\nImplement and verify in source only. Do not commit or stage; keep binaries outside source. '
              f'Write {trial}/trial-report.md with selected/opened skill files, decisions, actual '
              'checks, and limits. Prefix shell commands with rtk. No external actions. '
              'This is a disposable evaluation task; do not update repository progress files.')
    (trial / 'prompt.txt').write_text(prompt)
    run(['git', 'init', '-q'], source)
    run(['git', 'add', '.'], source)
    run(['git', '-c', 'commit.gpgsign=false', '-c', 'user.name=Evaluation', '-c', 'user.email=evaluation@example.invalid',
         'commit', '-qm', 'trial input'], source)
    metadata = {'skill': skill, 'arm': arm, 'case': case, 'repeat': repeat,
                'harness': 'Codex desktop multi-agent', 'author_model': model,
                'author_effort': effort, 'input_hashes': hashes(source),
                'catalog_hashes': hashes(catalog), 'base_commit': run(['git', 'rev-parse', 'HEAD'], source).strip()}
    (trial / 'metadata.json').write_text(json.dumps(metadata, indent=2) + '\n')
    print(prompt)


def archive_trial(skill, arm, case, repeat):
    trial, archive = locations(skill, arm, case, repeat)
    source = trial / 'source'
    metadata = json.loads((trial / 'metadata.json').read_text())
    if hashes(trial / 'catalog') != metadata['catalog_hashes']:
        raise RuntimeError('trial mutated fixed catalog')
    if archive.exists():
        raise FileExistsError(archive)
    archive.mkdir(parents=True)
    run(['git', 'add', '.'], source)
    patch = run(['git', 'diff', '--binary', '--full-index', metadata['base_commit']], source)
    (archive / 'source.patch').write_text(patch)
    for name in ('prompt.txt', 'trial-report.md'):
        shutil.copyfile(trial / name, archive / name)
    metadata['output_hashes'] = hashes(source)
    metadata['patch_sha256'] = hashlib.sha256(patch.encode()).hexdigest()
    (archive / 'manifest.json').write_text(json.dumps(metadata, indent=2) + '\n')
    check = trial / 'reconstructed'
    shutil.copytree(REPO / f'tests/go-quality-build/{skill}/evals/files/{case}', check)
    if patch:
        run(['git', 'apply', str(archive / 'source.patch')], check)
    if hashes(check) != metadata['output_hashes']:
        raise RuntimeError('patch reconstruction hash mismatch')
    print(json.dumps({'archive': str(archive), 'files': len(metadata['output_hashes']), 'reconstruction': 'PASS'}))


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('action', choices=('prepare', 'archive'))
    parser.add_argument('skill')
    parser.add_argument('arm', choices=('baseline', 'skill-on'))
    parser.add_argument('case')
    parser.add_argument('--repeat', default='first')
    parser.add_argument('--model', default='gpt-6-luna')
    parser.add_argument('--effort', default='medium')
    args = parser.parse_args()
    if args.action == 'prepare':
        prepare(args.skill, args.arm, args.case, args.repeat, args.model, args.effort)
    else:
        archive_trial(args.skill, args.arm, args.case, args.repeat)
