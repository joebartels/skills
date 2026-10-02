import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time

REPO=Path('/Users/jb/.codex/worktrees/b633/skills')
OUT=Path('/private/tmp/go-behavior-r2-promotion')
ROOT=REPO/'tests/go-quality-build/results'
checks=json.loads((OUT/'checks.json').read_text())
env={'GOCACHE':str(OUT/'cache'),'GOTOOLCHAIN':'local'}
def command(args,cwd=REPO):
    start=time.monotonic()
    p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env={**os.environ,**env},capture_output=True,text=True,timeout=60)
    r={'command':['rtk','proxy',*args],'cwd':str(cwd),'environment_overrides':env,'stdout':p.stdout,'stderr':p.stderr,'exit_code':p.returncode,'duration_seconds':round(time.monotonic()-start,3)}
    checks['commands'].append(r);return r
def digest(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def maps(path):return {str(p.relative_to(path)):digest(p) for p in sorted(path.rglob('*')) if p.is_file()}

checks['fresh_structure']=[command(['python3','-B','/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py','docs/go-quality-build/drafts/go-behavior-tests']),command(['go','version'])]
draft=REPO/'docs/go-quality-build/drafts/go-behavior-tests'
checks['local_links']=[]
for p in draft.rglob('*.md'):
    for link in re.findall(r'\]\(([^)]+)\)',p.read_text()):
        if '://' not in link:
            target=(p.parent/link.split('#',1)[0]).resolve()
            ok=target.is_relative_to(draft) and target.exists()
            checks['local_links'].append({'source':str(p),'link':link,'resolved':str(target),'valid':ok});assert ok

example=OUT/'example';example.mkdir(exist_ok=True)
snippet=re.findall(r'```go\n(.*?)```',(draft/'references/behavior-observations.md').read_text(),re.S)[0]
(example/'go.mod').write_text('module example.com/finish\n\ngo 1.22\n')
(example/'finish_test.go').write_text(snippet)
checks['illustration']={'green':command(['go','test','-count=1','-timeout=30s','./...'],example),'vet':command(['go','vet','./...'],example)}
(example/'finish_test.go').write_text(snippet.replace('return errors.Join(workErr, closeResource())','if closeErr := closeResource(); closeErr != nil { return errors.Join(workErr, closeErr) }; return nil'))
checks['illustration']['red']=command(['go','test','-count=1','-timeout=30s','./...'],example)
(example/'finish_test.go').write_text(snippet)

# Independently reconstruct every raw review mutation with preserved changed-source files.
checks['review_mutation_reconstructions']=[]
for run in checks['scope']:
    rp=ROOT/run
    m=json.loads((rp/'manifest.json').read_text())
    for c in m['cases']:
        cp=rp/c['id']
        for changed in sorted(cp.glob('review-mutations/**/changed-source')):
            patch=changed.parent/'source.patch'
            if not patch.exists():continue
            path=OUT/'review-mutations'/run/c['id']/str(changed.parent.relative_to(cp/'review-mutations'))
            if path.exists():shutil.rmtree(path)
            shutil.copytree(cp/'candidate',path)
            applied=command(['git','apply',str(patch)],path)
            matched=applied['exit_code']==0 and all((path/rel).exists() and digest(path/rel)==value for rel,value in maps(changed).items())
            checks['review_mutation_reconstructions'].append({'run':run,'case':c['id'],'mutation':str(changed.parent.relative_to(cp/'review-mutations')),'changed_files':len(maps(changed)),'matched':matched,'apply':applied})

# Verify offered descriptions and scope from exact preserved wrappers, independent of labels.
checks['wrapper_guidance_matches']=[]
base=json.loads((ROOT/checks['scope'][0]/'manifest.json').read_text())
for run in checks['scope']:
    rp=ROOT/run;m=json.loads((rp/'manifest.json').read_text())
    for c in m['cases']:
        cp=rp/c['id'];dispatch=(cp/'dispatch.txt').read_text()
        descs={name:description for name,description in re.findall(r'^- ([a-z-]+): (.*?) \(path [^\n]+\)$',dispatch,re.M)}
        common={name:description for name,description in descs.items() if name not in ['go-behavior-tests','go-test-isolation']}
        baseline_dispatch=(ROOT/checks['scope'][0]/'library-codec/dispatch.txt').read_text()
        baseline_descs={name:description for name,description in re.findall(r'^- ([a-z-]+): (.*?) \(path [^\n]+\)$',baseline_dispatch,re.M)}
        assert common==baseline_descs
        if run=='2026-10-01-behavior-tests-r2':assert 'go-test-isolation' not in descs
        if run=='2026-10-01-testing-combined-final' and c['id']=='behavior-only':assert 'go-test-isolation' not in descs
        checks['wrapper_guidance_matches'].append({'run':run,'case':c['id'],'common_descriptions_matched':True,'offered_names':list(descs)})

checks['post_checks_seals']=[]
for run in checks['scope']:
    rp=ROOT/run;bad=[]
    for line in (rp/'checksums.sha256').read_text().splitlines():
        expected,rel=line.split('  ',1)
        if digest(rp/rel)!=expected:bad.append(rel)
    checks['post_checks_seals'].append({'run':run,'failed':bad});assert not bad
checks['runtime_revision1']=maps(REPO/'plugins/go-quality-build/skills/go-behavior-tests')==maps(ROOT/'2026-10-01-behavior-tests-skill-on/skills/go-behavior-tests')
checks['word_counts']={str(p.relative_to(draft)):len(p.read_text().split()) for p in draft.rglob('*.md')}
checks['primary_documentation']=[{'url':'https://go.dev/ref/spec#Comparison_operators','checked':'2026-10-01','relevance':'non-comparable dynamic interface equality can panic'},{'url':'https://pkg.go.dev/os#WriteFile','checked':'2026-10-01','relevance':'failed write can leave a partially written destination'},{'url':'https://pkg.go.dev/errors#Is','checked':'2026-10-01','relevance':'promised matching across wrapping/joining'},{'url':'https://pkg.go.dev/testing','checked':'2026-10-01','relevance':'subtests, examples and test ownership APIs'}]
(OUT/'checks.json').write_text(json.dumps(checks,indent=2)+'\n')
print(json.dumps({'structure':checks['fresh_structure'],'illustration_exits':[checks['illustration'][k]['exit_code'] for k in ['green','vet','red']],'review_mutation_reconstructions':len(checks['review_mutation_reconstructions']),'failed_review_reconstructions':[r for r in checks['review_mutation_reconstructions'] if not r['matched']],'wrapper_count':len(checks['wrapper_guidance_matches']),'runtime_revision1':checks['runtime_revision1'],'word_counts':checks['word_counts']},indent=2))
