from pathlib import Path
import difflib, hashlib, json, os, re, resource, shutil, signal, subprocess, time
ROOT=Path('/Users/jb/.codex/worktrees/b633/skills')
OUT=Path('/private/tmp/go-behavior-promotion-review')
UPSTREAM=Path('/private/tmp/go-quality-build-upstream')
PIN='19a0626ae8565d27a7b7bdf59d8d99d94d7e284c'
records=[]
env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local',GOPROXY='off')
def run(args,cwd=ROOT,input=None,preexec_fn=None):
 start=time.monotonic()
 try:
  p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env=env,text=True,capture_output=True,timeout=55,input=input,preexec_fn=preexec_fn)
  r=dict(command=['rtk','proxy',*args],cwd=str(cwd),env={k:env[k] for k in ['GOCACHE','GOTOOLCHAIN','GOPROXY']},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3))
  if input: r['stdin']=input
 except subprocess.TimeoutExpired as e:
  r=dict(command=['rtk','proxy',*args],cwd=str(cwd),exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
 records.append(r)
 (OUT/'additional-checks.json').write_text(json.dumps(records,indent=2)+'\n')
 return r
run(['git','status','--short'],UPSTREAM)
run(['git','rev-parse','HEAD'],UPSTREAM)
source_names=['skills/golang-testing/SKILL.md', *[f'skills/golang-testing/references/{n}.md' for n in ['benchmarks','coverage','examples','helpers','http-testing','integration-testing','mocking']], 'skills/golang-naming/references/testing.md','skills/golang-project-layout/references/testing-layout.md','LICENSE']
sources={}
for name in source_names:
 committed=run(['git','show',f'{PIN}:{name}'],UPSTREAM)
 assert committed['exit_code']==0 and committed['stdout'].encode()==(UPSTREAM/name).read_bytes()
 sources[name]=hashlib.sha256((UPSTREAM/name).read_bytes()).hexdigest()
audit=(ROOT/'docs/go-quality-build/testing-source-audit.md').read_text()
draft_paths=[ROOT/'docs/go-quality-build/drafts/go-behavior-tests/SKILL.md',ROOT/'docs/go-quality-build/drafts/go-behavior-tests/references/behavior-observations.md']
copy_comparison=[]
for local in draft_paths:
 local_words=re.findall(r'[\w]+',local.read_text().lower())
 matches=[]
 for name in source_names[:-1]:
  remote_words=re.findall(r'[\w]+',(UPSTREAM/name).read_text().lower())
  match=difflib.SequenceMatcher(None,local_words,remote_words,autojunk=False).find_longest_match()
  matches.append(dict(source=name,consecutive_word_count=match.size,local_start_word=match.a,upstream_start_word=match.b))
 copy_comparison.append(dict(local=str(local),longest_matches=sorted(matches,key=lambda x:x['consecutive_word_count'],reverse=True)))
records.append(dict(kind='pinned-copy-scope-check',pin=PIN,source_sha256=sources,longest_word_matches=copy_comparison,note='Consecutive-word comparison supplements manual comparison; shared generic Go/test terms are not by themselves copied substantial content.'))

for arm in ['baseline','skill-on']:
 module=OUT/'reconstructed'/arm/'cli-partial-failure'
 dst=OUT/'partial-write'/arm
 dst.mkdir(parents=True)
 binary=dst/'ledgerload'
 assert run(['go','build','-o',str(binary),'.'],module)['exit_code']==0
 def child_limit():
  signal.signal(signal.SIGXFSZ,signal.SIG_IGN)
  resource.setrlimit(resource.RLIMIT_FSIZE,(3,3))
 limited=run([str(binary),'--dir',str(dst/'records')],module,input='web,1\nweb,12345\nlater,9\n',preexec_fn=child_limit)
 limited['setup']='child-only RLIMIT_FSIZE soft=hard=3; inherited SIGXFSZ ignored'
 limited['contents']={p.name:p.read_text() for p in (dst/'records').iterdir() if p.is_file()}
 limited['parent_limit_after']=resource.getrlimit(resource.RLIMIT_FSIZE)
 print(arm,'partial-write',limited['exit_code'],limited['contents'],flush=True)

ref=draft_paths[1].read_text()
snippet=re.search(r'```go\n(.*?)\n```',ref,re.S).group(1)+'\n'
example=OUT/'example'
example.mkdir()
(example/'go.mod').write_text('module example.com/finish\n\ngo 1.22\n')
(example/'finish_test.go').write_text(snippet)
example_hash=hashlib.sha256(snippet.encode()).hexdigest()
records.append(dict(kind='exact-example',source_sha256=example_hash,archive_sha256=json.loads((ROOT/'tests/go-quality-build/results/2026-10-01-behavior-tests-skill-on/example-verification.json').read_text())['source_sha256']))
run(['go','test','-count=1','-timeout=30s','./...'],example)
run(['go','vet','-stdversion','./...'],example)
mutation=snippet.replace('return errors.Join(workErr, closeResource())','if closeErr := closeResource(); closeErr != nil {\n\t\treturn errors.Join(workErr, closeErr)\n\t}\n\treturn nil')
assert mutation!=snippet
(example/'finish_test.go').write_text(mutation)
run(['go','test','-run','^$','-timeout=30s','./...'],example)
run(['go','test','-count=1','-timeout=30s','./...'],example)
(example/'finish_test.go').write_text(snippet)
run(['python3','-B','/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py','docs/go-quality-build/drafts/go-behavior-tests'])
run(['python3','-B','scripts/validate.py'])
run(['git','status','--short'])
(OUT/'additional-checks.json').write_text(json.dumps(records,indent=2)+'\n')
print('comparison max lengths',[(p['local'],p['longest_matches'][0]['consecutive_word_count']) for p in copy_comparison],flush=True)
