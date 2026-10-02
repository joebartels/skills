import difflib
import hashlib
import json
import re
import subprocess
from pathlib import Path

ROOT = Path('/Users/jb/.codex/worktrees/b633/skills')
OUT = Path('/private/tmp/go-isolation-promotion-review-20261001')
UP = Path('/private/tmp/go-quality-build-upstream')
PIN = '19a0626ae8565d27a7b7bdf59d8d99d94d7e284c'
result = {'commands': [], 'assertions': [], 'copy_scope_evidence': []}
def sha(b): return hashlib.sha256(b).hexdigest()
def command(cwd, args):
    cmd=['rtk','proxy','git']+args
    r=subprocess.run(cmd,cwd=cwd,capture_output=True,timeout=30)
    result['commands'].append({'argv':cmd,'cwd':str(cwd),'exit_code':r.returncode,'stdout_sha256':sha(r.stdout),'stdout_bytes':len(r.stdout),'stderr':r.stderr.decode()})
    assert r.returncode==0
    return r.stdout
def assertion(name,ok,details):
    result['assertions'].append({'name':name,'ok':ok,'details':details})
    if not ok:
        (OUT/'provenance-failed-attempt.json').write_text(json.dumps(result,indent=2)+'\n')
        raise RuntimeError(name)
assertion('upstream-head',command(UP,['rev-parse','HEAD']).decode().strip()==PIN,PIN)
sources=[UP/'skills/golang-testing/SKILL.md']+sorted((UP/'skills/golang-testing/references').glob('*.md'))+[UP/'skills/golang-naming/references/testing.md',UP/'skills/golang-project-layout/references/testing-layout.md',UP/'skills/golang-project-layout/SKILL.md',UP/'LICENSE']
for p in sources:
    rel=str(p.relative_to(UP))
    blob=command(UP,['show',PIN+':'+rel])
    assertion('upstream-blob-equal:'+rel,blob==p.read_bytes(),{'sha256':sha(blob)})
audit=(ROOT/'docs/go-quality-build/testing-source-audit.md').read_text()
rows=re.findall(r'^\| \[([^\]]+)\]\(https://github.com/samber/cc-skills-golang/blob/'+PIN+r'/skills/([^#]+)#L(\d+)\) \| (copy|adapt|omit) \| ([^|]+) \| (.+) \|$',audit,re.M)
assertion('audit-section-count',len(rows)==79,len(rows))
row_lines={}
for label,rel,line,decision,owner,explanation in rows:
    n=int(line)
    p=UP/'skills'/rel
    actual=p.read_text().splitlines()[n-1]
    header=actual.lstrip('#').strip()
    if 'Preamble' not in label:
        named=label.split(' — ',1)[-1] if ' — ' in label else label.split(': ',1)[-1]
        assertion('audit-heading:'+rel+':'+line,header==named,{'actual':actual,'decision':decision,'owner':owner.strip()})
    row_lines.setdefault(rel,set()).add(n)
for p in sources:
    rel=str(p.relative_to(UP))
    if rel=='LICENSE' or rel=='skills/golang-project-layout/SKILL.md': continue
    in_code=False
    headings=[]
    for i,line in enumerate(p.read_text().splitlines(),1):
        if line.startswith('```'): in_code=not in_code
        if not in_code and re.match(r'^#{1,6} ',line): headings.append(i)
    assertion('all-headings-audited:'+rel, set(headings)<=row_lines[rel.removeprefix('skills/')], {'heading_lines':headings})
draft=ROOT/'docs/go-quality-build/drafts/go-test-isolation'
for local in sorted(p for p in draft.rglob('*') if p.is_file()):
    local_tokens=re.findall(r'\w+|[^\w\s]',local.read_text())
    best=None
    for upstream in sources:
        if upstream.name=='LICENSE': continue
        tokens=re.findall(r'\w+|[^\w\s]',upstream.read_text())
        match=difflib.SequenceMatcher(None,local_tokens,tokens,autojunk=False).find_longest_match()
        if best is None or match.size>best['token_count']:
            best={'upstream':str(upstream),'token_count':match.size,'matching_tokens':' '.join(local_tokens[match.a:match.a+match.size])}
    result['copy_scope_evidence'].append({'local':str(local),'longest_exact_token_run':best,'interpretation':'Supporting comparison only; manual comparison assessed wording, example structure and generic Go boilerplate.'})
    blocks=re.findall(r'```go\n(.*?)```',local.read_text(),re.S)
    all_upstream='\n'.join(p.read_text() for p in sources if p.name!='LICENSE')
    assertion('no-whole-go-block-copy:'+str(local.relative_to(draft)),all(block not in all_upstream for block in blocks),{'go_block_count':len(blocks)})
base=ROOT/'tests/go-quality-build/results/2026-10-01-test-isolation-baseline'
m=json.loads((base/'manifest.json').read_text())
for rel,h in m['skill_sha256'].items():
    if rel.startswith('go-core-style/'):
        p=Path('/Users/jb/.agents/skills/golang')/rel.removeprefix('go-core-style/')
        assertion('external-guidance-current-equal:'+rel,sha(p.read_bytes())==h,{'sha256':h})
    else:
        blob=command(ROOT,['show',m['fixture_commit']+':plugins/go-quality-build/skills/'+rel])
        assertion('existing-guidance-source:'+rel,sha(blob)==h,{'sha256':h})
(OUT/'provenance.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({'assertions':len(result['assertions']),'all_ok':all(r['ok'] for r in result['assertions']),'copy_scope_evidence':result['copy_scope_evidence']}))
