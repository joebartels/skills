from pathlib import Path
import hashlib, json, shutil, subprocess, tempfile
root=Path.cwd();base=root/'plugins/go-quality-build/skills';draft=root/'docs/go-quality-build/drafts/go-behavior-tests'
commit=subprocess.check_output(['rtk','proxy','git','rev-parse','HEAD'],text=True).strip()
def hashes(d): return {str(p.relative_to(d)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(d.rglob('*')) if p.is_file()}
for integrated in [False,True]:
 oldrun=root/'tests/go-quality-build/results'/('2026-10-01-testing-combined' if integrated else '2026-10-01-behavior-tests-skill-on')
 old=json.loads((oldrun/'manifest.json').read_text())
 suite=root/'tests/go-quality-build'/('testing-combined' if integrated else 'go-behavior-tests')/'evals'
 run=root/'tests/go-quality-build/results'/('2026-10-01-testing-combined-final' if integrated else '2026-10-01-behavior-tests-r2');run.mkdir()
 scratch=Path(tempfile.mkdtemp(prefix='go-fresh-author-',dir='/private/tmp'))
 shutil.copytree(oldrun/'skills',run/'skills');shutil.rmtree(run/'skills/go-behavior-tests');shutil.copytree(draft,run/'skills/go-behavior-tests')
 if integrated:
  shutil.rmtree(run/'skills/go-test-isolation');shutil.copytree(root/'docs/go-quality-build/drafts/go-test-isolation',run/'skills/go-test-isolation')
 data=json.loads((suite/'evals.json').read_text())['evals']
 tasks=([('behavior-only',data[0]),('both',data[0])] if integrated else [(c['id'],c) for c in data]+[(alias,next(c for c in data if c['id']==fixture)) for alias,fixture in [('cli-repeat','cli-partial-failure'),('worker-repeat','worker-host')]])
 manifest=dict(run_id=run.name,stage='revision-2 dispatches frozen before authors',fixture_commit=old['fixture_commit'],skill_commit=commit,revision=2,skill_sha256=hashes(run/'skills'),author_settings=old['author_settings'],host=old['host'],control=dict(parent_run=oldrun.name,task_fixture_common_guidance='identical to frozen first-pass arm; behavior revision 2; isolation revision 2 in final both; existing architecture/style unchanged',other_testing_skill='absent in single suite; isolation revision 2 in combined both',controller_reviews_other_candidates='withheld'),cases=[])
 for n,(name,case) in enumerate(tasks):
  home=scratch/('task-'+str(n+1));home.mkdir();work=home/'module';shutil.copytree(suite/'files'/case['id'],work);out=home/'report';out.mkdir();guides=home/'guidance';guides.mkdir()
  names=['go-package-boundaries','go-api-contracts','go-interfaces-and-composition','go-core-style','go-behavior-tests']+(['go-test-isolation'] if integrated and name=='both' else [])
  for skill in names:shutil.copytree(run/'skills'/skill,guides/skill)
  descriptions=[]
  for p in sorted(guides.glob('*/SKILL.md')):
   desc=next(l.split(':',1)[1].strip().strip('"') for l in p.read_text().splitlines() if l.startswith('description:'));descriptions.append(f'- {p.parent.name}: {desc} (path {p})')
  wrapper=f'''You are completing a standalone Go task in an isolated module at {work}. This is actual implementation work, not a review. Work only in that module; write report files only in {out}. Do not inspect the skills repository, design records, eval JSON, controller probes, other author workspaces or results. Those are outside this task. Do not delegate, install dependencies/tools, commit or make external changes.

Read this module's README and source. Complete this request:
{case['prompt']}

The only optional writing guidance offered for this task is:
{chr(10).join(descriptions)}
Read only guidance whose description is relevant. Record actual opened files and relevance decisions. Do not open any other writing or review skills, including unavailable testing guidance. No expected evaluation judgments or controller probes are supplied.

Keep the supported Go version. Run meaningful verification with finite test/process deadlines. Prefix every shell command with rtk and use rtk proxy for raw output. Use GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. If local listeners are restricted, preserve the failure and any approved unchanged-source rerun separately; disclose the exercised boundary.

When done, write {out}/selection.json with offered names, opened skill/reference paths and selection reasons. Write {out}/author-report.md with changes, actual checks, known limitations and remaining risks. Preserve command arrays, cwd, non-secret environment, stdout/stderr/exit codes in {out}/checks.json when possible. Return concise results and report paths.'''
  arc=run/name;arc.mkdir();(arc/'prompt.txt').write_text(case['prompt']);(arc/'dispatch.txt').write_text(wrapper)
  launch=f'Read and execute only the standalone Go author dispatch at {arc}/dispatch.txt. That single file is authorized input and identifies your isolated module and optional writing guidance. Do not inspect any other repository files or another author workspace. No delegation.'
  (arc/'launch.txt').write_text(launch)
  manifest['cases'].append(dict(id=name,fixture_id=case['id'],isolation_relevant=integrated or case['id'] in {'cli-partial-failure','worker-host','service-publication'},workdir=str(work),reportdir=str(out),available_skills=names,input_sha256=hashes(work),guidance_sha256=hashes(guides),dispatch_sha256=hashlib.sha256(wrapper.encode()).hexdigest()))
  prior=next(c for c in old['cases'] if c['id']==(name if integrated else case['id']))
  assert prior['input_sha256']==hashes(work)
  assert (oldrun/prior['id']/'prompt.txt').read_bytes()==(arc/'prompt.txt').read_bytes()
  for skill in names:
   if skill not in {'go-test-isolation','go-behavior-tests'}: assert hashes(oldrun/'skills'/skill)==hashes(guides/skill)
  print(run.name,name,launch)
 (run/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
 (run/'README.md').write_text('# Behavior revision 2 fresh evaluations\n\nPending independent outcomes. Exact frozen task/input/common-guidance/settings match the recorded prior run. Behavior revision 2 adds conditional exact-set and interface/error counterexamples; final both also uses already-evaluated isolation revision 2. No repair or review feedback is supplied.\n')
 (run/'comparison.md').write_text('# Comparison\n\nPending author verification and neutral independent outcomes; no improvement claimed.\n')
