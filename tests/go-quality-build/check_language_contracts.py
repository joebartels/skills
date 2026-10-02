"""Run author checks and withheld behavior probes on reconstructed trial copies."""

import argparse
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

SPEC = importlib.util.spec_from_file_location('trials', Path(__file__).with_name('language_contract_trials.py'))
trials = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(trials)


def check(skill, arm, case, repeat, round_name):
    trials.locations(skill, arm, case, round_name)
    trial, archive = trials.locations(skill, arm, case, repeat)
    source = trial / 'reconstructed'
    commands = []
    cache = Path('/private/tmp/go-language-contract-check-cache')
    cache.mkdir(exist_ok=True)
    environment = dict(os.environ, GOCACHE=str(cache), GOWORK='off')
    for args in (['go', 'test', '-race', '-count=1', '-timeout=45s', './...'], ['go', 'vet', './...']):
        result = subprocess.run(['rtk', 'proxy', *args], cwd=source, env=environment, text=True, capture_output=True)
        commands.append({'command': args, 'exit_code': result.returncode,
                         'stdout': result.stdout, 'stderr': result.stderr})
    probe = trials.REPO / f'tests/go-quality-build/{skill}/probes/{case}_test.go.txt'
    if probe.exists():
        checked = trial / f'checked-{round_name}'
        if checked.exists():
            raise FileExistsError(checked)
        shutil.copytree(source, checked)
        (checked / 'language_contract_probe_test.go').write_text(probe.read_text())
        args = ['go', 'test', '-race', '-count=1', '-timeout=45s', '-run', '^TestContract', './...']
        result = subprocess.run(['rtk', 'proxy', *args], cwd=checked, env=environment, text=True, capture_output=True)
        commands.append({'command': args, 'exit_code': result.returncode,
                         'stdout': result.stdout, 'stderr': result.stderr,
                         'probe_sha256': trials.hashes(probe.parent)[probe.name]})
    if skill == 'go-error-contracts' and case == 'cli-completion':
        with tempfile.TemporaryDirectory(prefix='go-language-cli-') as scratch:
            scratch = Path(scratch)
            binary = scratch / 'lineexport'
            built = subprocess.run(['rtk', 'proxy', 'go', 'build', '-o', str(binary), './cmd/lineexport'],
                                   cwd=source, env=environment, text=True, capture_output=True)
            commands.append({'command': ['go', 'build', './cmd/lineexport'], 'exit_code': built.returncode,
                             'stdout': built.stdout, 'stderr': built.stderr})
            if not built.returncode:
                (scratch / '-input').write_text('a\nb\nc\n')
                for args, code, output in ((['-input', 'out', '2'], 0, 'a\nb\n'),
                                           (['missing', 'bad'], 1, None), ([], 2, None)):
                    result = subprocess.run([str(binary), *args], cwd=scratch, text=True, capture_output=True)
                    observed = (scratch / args[1]).read_text() if output is not None and (scratch / args[1]).exists() else None
                    passed = (result.returncode == code and not result.stdout and observed == output and
                              (not result.stderr if code == 0 else bool(result.stderr)))
                    commands.append({'command': ['lineexport', *args], 'exit_code': 0 if passed else 1,
                                     'observed_status': result.returncode, 'stdout': result.stdout,
                                     'stderr': result.stderr, 'output': observed})
    if skill == 'combined' and case == 'language-contracts':
        with tempfile.TemporaryDirectory(prefix='go-language-combined-cli-') as scratch:
            scratch = Path(scratch)
            binary = scratch / 'batchview'
            built = subprocess.run(['rtk', 'proxy', 'go', 'build', '-o', str(binary), './cmd/batchview'],
                                   cwd=source, env=environment, text=True, capture_output=True)
            commands.append({'command': ['go', 'build', './cmd/batchview'], 'exit_code': built.returncode,
                             'stdout': built.stdout, 'stderr': built.stderr})
            if not built.returncode:
                (scratch / '-input').write_text('#comment\na=one\nb=two\n')
                (scratch / 'malformed').write_text('a=one\nbroken\nb=two\n')
                one = [{'key': 'a', 'tags': {'data': 'b25l'}}]
                both = one + [{'key': 'b', 'tags': {'data': 'dHdv'}}]
                for args, code, expected in ((['-input', 'out', 'a'], 0, one),
                                            (['-input', 'all'], 0, both),
                                            (['malformed', 'partial'], 1, one),
                                            (['-input', 'absent/out'], 1, None), ([], 2, None)):
                    result = subprocess.run([str(binary), *args], cwd=scratch, text=True, capture_output=True)
                    observed = None
                    if expected is not None and (scratch / args[1]).exists():
                        try:
                            observed = json.loads((scratch / args[1]).read_text())
                        except json.JSONDecodeError:
                            observed = 'invalid JSON'
                    passed = (result.returncode == code and not result.stdout and observed == expected and
                              (not result.stderr if code == 0 else bool(result.stderr)))
                    commands.append({'command': ['batchview', *args], 'exit_code': 0 if passed else 1,
                                     'observed_status': result.returncode, 'stdout': result.stdout,
                                     'stderr': result.stderr, 'output': observed})
    destination = archive / f'verification-{round_name}.json'
    if destination.exists():
        raise FileExistsError(destination)
    destination.write_text(json.dumps(commands, indent=2) + '\n')
    print(json.dumps({'skill': skill, 'arm': arm, 'case': case, 'repeat': repeat,
                      'checks': [{'command': c['command'], 'exit_code': c['exit_code']} for c in commands]}))
    return all(c['exit_code'] == 0 for c in commands)


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('skill')
    parser.add_argument('arm')
    parser.add_argument('case')
    parser.add_argument('--repeat', default='first')
    parser.add_argument('--round', default='final')
    args = parser.parse_args()
    raise SystemExit(0 if check(args.skill, args.arm, args.case, args.repeat, args.round) else 1)
