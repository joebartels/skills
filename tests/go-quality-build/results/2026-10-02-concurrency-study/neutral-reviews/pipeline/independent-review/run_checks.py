import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

ROOT = Path('/private/tmp/go-neutral-pipeline-study')
OUT = ROOT / 'independent-review'
MIN = '/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
ENV = dict(os.environ, GOTOOLCHAIN='local', GOWORK='off', GOPROXY='off', GOSUMDB='off', GOCACHE=str(OUT / 'cache'), GOQUALITY_GO=MIN)
records = []

def run(label, cwd, argv, tool=None, input=None, timeout=45):
    env = dict(ENV)
    if tool:
        env['GOQUALITY_GO'] = tool
    start = time.monotonic()
    try:
        result = subprocess.run(['rtk', 'proxy', *argv], cwd=cwd, env=env, input=input, text=True, capture_output=True, timeout=timeout)
        row = dict(label=label, argv=['rtk', 'proxy', *argv], cwd=str(cwd), environment_overrides={k:env[k] for k in ['GOTOOLCHAIN','GOWORK','GOPROXY','GOSUMDB','GOCACHE','GOQUALITY_GO']}, status=result.returncode, stdout=result.stdout, stderr=result.stderr, timed_out=False)
    except subprocess.TimeoutExpired as exc:
        row = dict(label=label, argv=['rtk','proxy',*argv], cwd=str(cwd), status=None, stdout=str(exc.stdout or ''), stderr=str(exc.stderr or ''), timed_out=True)
    row['elapsed_seconds'] = round(time.monotonic()-start, 3)
    records.append(row)
    (OUT / 'raw' / 'independent-commands.json').write_text(json.dumps(records, indent=2)+'\n')
    print(label, 'status='+str(row['status']), 'timeout='+str(row['timed_out']), flush=True)
    return row

def hashes():
    paths = [ROOT/'original', ROOT/'candidates'/'A', ROOT/'candidates'/'B', ROOT/'review-guidance', ROOT/'held-checks']
    return {str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for root in paths for p in sorted(root.rglob('*')) if p.is_file()}

before = hashes()
(OUT/'source-hashes-before.json').write_text(json.dumps(before, indent=2)+'\n')
expected = json.loads((ROOT/'source-manifest.json').read_text())
assert all(before[f'candidates/{candidate}/{name}'] == value for candidate, files in expected.items() for name, value in files.items())

for candidate in ['A','B']:
    src = ROOT/'candidates'/candidate
    ordinary = OUT/'work'/f'{candidate}-ordinary'
    shutil.copytree(src, ordinary)
    for tool, version in [('go','host'),(MIN,'minimum')]:
        run(f'{candidate}/{version}/version', ordinary, [tool,'version'], tool)
        run(f'{candidate}/{version}/build', ordinary, [tool,'build','-mod=readonly','-o',str(OUT/'work'/f'{candidate}-{version}-pipeline'),'./cmd/pipeline'], tool)
        run(f'{candidate}/{version}/ordinary-tests', ordinary, [tool,'test','-mod=readonly','-count=1','-timeout=20s','./...'], tool)
        run(f'{candidate}/{version}/vet', ordinary, [tool,'vet','-mod=readonly','./...'], tool)
        for name, args, data, code, stdout in [
            ('finite', [], '5\n4\n3\n2\n1\n', 0, '5\n4\n3\n2\n1\n'),
            ('early-prefix', ['--take','2'], '5\n4\n3\n2\n1\n', 0, '5\n4\n'),
            ('malformed-first', ['--take','2'], 'bad\n1\n', 2, ''),
            ('empty', [], '', 0, ''),
            ('invalid-take', ['--take','0'], '1\n', 2, ''),
        ]:
            row = run(f'{candidate}/{version}/executable/{name}', ordinary, [str(OUT/'work'/f'{candidate}-{version}-pipeline'),*args], tool, data, timeout=5)
            row['expected_status'] = code
            row['expected_stdout'] = stdout
            row['contract_match'] = row['status']==code and row['stdout']==stdout and ((not row['stderr']) if code==0 else (('integer input' in row['stderr']) if name=='malformed-first' else True))
    run(f'{candidate}/format', ordinary, ['gofmt','-d','.'])
    held = OUT/'work'/f'{candidate}-held'
    shutil.copytree(src,held)
    shutil.copy2(ROOT/'held-checks'/'contract_test.go', held/'cmd/pipeline/controller_contract_test.go')
    for tool, version in [('go','host'),(MIN,'minimum')]:
        run(f'{candidate}/{version}/held-tests', held, [tool,'test','-mod=readonly','-count=1','-timeout=20s','./...'], tool)
    run(f'{candidate}/held-race-shuffle', held, ['go','test','-race','-shuffle=on','-count=3','-timeout=30s','./...'], 'go')
    probe = OUT/'work'/f'{candidate}-probe'
    shutil.copytree(src,probe)
    shutil.copy2(OUT/'probes/review_probe_test.go',probe/'cmd/pipeline/review_probe_test.go')
    for tool, version in [('go','host'),(MIN,'minimum')]:
        run(f'{candidate}/{version}/independent-probes', probe, [tool,'test','-mod=readonly','-run','^TestReview','-count=1','-timeout=20s','./...'], tool)
    run(f'{candidate}/probe-race-shuffle', probe, ['go','test','-race','-shuffle=on','-count=3','-timeout=30s','-run','^TestReview','./...'], 'go')

    mutations = ['no-early-stop','no-failure-stop','skip-early-join','broad-cancel-suppression']
    targets = ['TestReviewEarlyStopJoin','TestReviewFailureStopJoin','TestReviewEarlyStopJoin','TestReviewWrappedFailures']
    for mutation, target in zip(mutations, targets):
        mutant = OUT/'work'/f'{candidate}-mutant-{mutation}'
        shutil.copytree(src,mutant)
        file = mutant/'cmd/pipeline/pipeline.go'
        data = file.read_text()
        if candidate=='A':
            if mutation=='no-early-stop':
                old='consumerStopped = r.stoppedAtReturn\n\t\t\t\tcancel()'; new='consumerStopped = r.stoppedAtReturn\n\t\t\t\t// MUTANT omits early cancellation'
            elif mutation=='no-failure-stop':
                old='if r.err != nil {'; new='if false {'
            elif mutation=='skip-early-join':
                old='consumerStopped = r.stoppedAtReturn\n\t\t\t\tcancel()'; new='consumerStopped = r.stoppedAtReturn\n\t\t\t\tcancel()\n\t\t\t\tif r.err == nil && ctx.Err() == nil { return nil }'
            else:
                old='(err == context.Canceled || err == context.DeadlineExceeded)'; new='(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))'
        else:
            if mutation=='no-early-stop':
                old='!first.producer || first.err != nil || ctx.Err() != nil'; new='(first.producer && first.err != nil) || ctx.Err() != nil'
            elif mutation=='no-failure-stop':
                old='!first.producer || first.err != nil || ctx.Err() != nil'; new='!first.producer || ctx.Err() != nil'
            elif mutation=='skip-early-join':
                old='second := <-results'; new='if !first.producer && first.err == nil && ctx.Err() == nil { return nil }\n\tsecond := <-results'
            else:
                old='(err == context.Canceled || err == context.DeadlineExceeded)'; new='(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))'
        assert data.count(old)==1, (candidate,mutation)
        file.write_text(data.replace(old,new))
        (OUT/'raw'/f'{candidate}-{mutation}.patch.txt').write_text(f'old:\n{old}\nnew:\n{new}\n')
        baseline = run(f'{candidate}/mutation/{mutation}/original-target-green', probe, ['go','test','-mod=readonly','-run','^'+target+'$','-count=1','-timeout=10s','./...'], 'go')
        compiled = run(f'{candidate}/mutation/{mutation}/compile',mutant,['go','test','-mod=readonly','-run','^$','-count=1','-timeout=10s','./...'],'go')
        # Ordinary tests establish author-test sensitivity separately from review probes.
        ordinary_mutant = run(f'{candidate}/mutation/{mutation}/ordinary-suite',mutant,['go','test','-mod=readonly','-count=1','-timeout=8s','./...'],'go',timeout=12)
        shutil.copy2(OUT/'probes/review_probe_test.go', mutant/'cmd/pipeline/review_probe_test.go')
        assertion = run(f'{candidate}/mutation/{mutation}/independent-target', mutant, ['go','test','-mod=readonly','-run','^'+target+'$','-count=1','-timeout=10s','./...'], 'go')
        assertion['confirmed_detection'] = baseline['status']==0 and compiled['status']==0 and assertion['status']!=0 and 'ASSERT' in assertion['stdout'] and not assertion['timed_out'] and 'panic: test timed out' not in assertion['stdout']

after = hashes()
(OUT/'source-hashes-after.json').write_text(json.dumps(after,indent=2)+'\n')
(OUT/'raw/independent-commands.json').write_text(json.dumps(records,indent=2)+'\n')
(OUT/'raw/source-integrity.json').write_text(json.dumps(dict(unchanged=before==after,manifest_matches=True,changed=[p for p in before if before[p]!=after[p]]),indent=2)+'\n')
assert before==after
