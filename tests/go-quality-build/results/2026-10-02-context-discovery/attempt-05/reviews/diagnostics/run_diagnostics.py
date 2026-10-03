from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import time

ROOT = Path('/private/tmp/packet-02-review-diagnostics')
SOURCE = Path('/private/tmp/go-outcome-review-20261002/packet-02/candidate')
GO122 = '/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
BASE_ENV = dict(os.environ, GOWORK='off', GOTOOLCHAIN='local', GOCACHE='/private/tmp/packet-02-go-cache')
records = []

def run(label, directory, argv, env=None):
    started = time.monotonic()
    result = subprocess.run(argv, cwd=directory, env=env or BASE_ENV, text=True, capture_output=True, timeout=45)
    record = dict(label=label, cwd=str(directory), argv=argv, environment_overrides={key:(env or BASE_ENV)[key] for key in ['GOWORK','GOTOOLCHAIN','GOCACHE','CGO_ENABLED'] if key in (env or BASE_ENV)}, status=result.returncode, stdout=result.stdout, stderr=result.stderr, elapsed_seconds=round(time.monotonic()-started,3))
    records.append(record)
    (ROOT/'executed-checks.json').write_text(json.dumps(records,indent=2)+'\n')
    print(label, 'status='+str(result.returncode), result.stdout.strip(), result.stderr.strip(), flush=True)

def copy(label):
    target = ROOT/label
    shutil.copytree(SOURCE,target,dirs_exist_ok=True)
    return target

baseline = ROOT/'baseline'
run('go122-version',baseline,['rtk','proxy',GO122,'version'])
run('go122-tests-cgo-disabled',baseline,['rtk','proxy',GO122,'test','-mod=readonly','-count=1','-timeout=30s','./...'],dict(BASE_ENV,CGO_ENABLED='0'))
run('standalone-build',baseline,['rtk','proxy','go','build','-mod=readonly','./...'])
run('selected-modules',baseline,['rtk','proxy','go','list','-m','all'])

mutations = {
    'restart-total-each-stage': (
        'totalCtx, cancelTotal := context.WithTimeout(ctx, total)',
        'totalCtx, cancelTotal := context.WithTimeout(ctx, total)\n\t_ = totalCtx',
        'stageCtx, cancelStage := context.WithTimeout(totalCtx, stage)',
        'resetCtx, resetCancel := context.WithTimeout(ctx, total)\n\t\tdefer resetCancel()\n\t\tstageCtx, cancelStage := context.WithTimeout(resetCtx, stage)',
    ),
    'cancel-before-body': (
        '\t\tbody, readErr := io.ReadAll(response.Body)',
        '\t\tcancelStage()\n\t\tbody, readErr := io.ReadAll(response.Body)',
    ),
    'ignore-close-only-failure': (
        'if readErr != nil || closeErr != nil {',
        'if readErr != nil {',
    ),
    'invalidate-complete-success': (
        '\treturn bodies, nil\n}',
        '\tif totalCtx.Err() != nil {\n\t\treturn bodies, stageError(totalCtx.Err(), totalCtx)\n\t}\n\treturn bodies, nil\n}',
    ),
}
for label, edits in mutations.items():
    directory = copy(label)
    path = directory/'stages.go'
    code = path.read_text()
    for old,new in zip(edits[::2],edits[1::2]):
        assert code.count(old)==1,(label,old)
        code = code.replace(old,new)
    path.write_text(code)
    run(label+'-author-tests',directory,['rtk','proxy','go','test','-mod=readonly','-count=1','-timeout=30s','./...'])

boundary = copy('boundary')
shutil.copy(ROOT/'review_contract_test.go',boundary/'review_contract_test.go')
run('review-contract-boundaries',boundary,['rtk','proxy','go','test','-mod=readonly','-run','^TestReview','-race','-count=1','-timeout=30s','-v','./...'])
for label in mutations:
    directory = ROOT/label
    shutil.copy(ROOT/'review_contract_test.go',directory/'review_contract_test.go')
    run(label+'-review-tests',directory,['rtk','proxy','go','test','-mod=readonly','-run','^TestReview','-count=1','-timeout=30s','./...'])

expected = json.loads((ROOT/'candidate-sha256.json').read_text())
actual = {str(p.relative_to(SOURCE)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(SOURCE.rglob('*')) if p.is_file()}
assert actual==expected,'Candidate changed during review'
print('Candidate SHA-256 manifest unchanged',flush=True)
