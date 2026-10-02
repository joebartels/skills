from pathlib import Path
import json,shutil,tempfile,hashlib,subprocess
root=Path.cwd();run=root/'tests/go-quality-build/results/2026-10-01-testing-combined';suite=root/'tests/go-quality-build/testing-combined/evals';base=root/'plugins/go-quality-build/skills';scratch=Path(tempfile.mkdtemp(prefix='go-combined-author-',dir='/private/tmp'))
def hashes(d):return {str(p.relative_to(d)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(d.rglob('*')) if p.is_file()}
source_commit=subprocess.check_output(['rtk','proxy','git','rev-parse','HEAD'],text=True).strip();prompt=json.loads((suite/'evals.json').read_text())['evals'][0]['prompt']
all_names=['go-package-boundaries','go-api-contracts','go-interfaces-and-composition','go-core-style','go-behavior-tests','go-test-isolation']
for n in all_names:
 src=Path('/Users/jb/.agents/skills/golang') if n=='go-core-style' else base/n
 shutil.copytree(src,run/'skills'/n,dirs_exist_ok=True)
manifest=dict(run_id=run.name,stage='frozen four-arm dispatches before authors',fixture_commit=source_commit,skill_commit=source_commit,external_style_commit=None,skill_sha256=hashes(run/'skills'),author_settings={'model':'inherited session model; exact model ID not exposed by collaboration harness','reasoning':'inherited session setting; no override','fresh_context':'collaboration.spawn_agent fork_turns=none','harness':'Codex collaboration tools'},host={'go':'go1.26.5','platform':'darwin/arm64','GOTOOLCHAIN':'local'},control={'difference':'only behavior/isolation writing-guidance availability','task_and_fixture':'identical frozen bytes','architecture_style':'same exact snapshots','controller_probes_reports_other_arms':'withheld'},cases=[])
for k,(arm,extra) in enumerate([('architecture-only',[]),('behavior-only',['go-behavior-tests']),('isolation-only',['go-test-isolation']),('both',['go-behavior-tests','go-test-isolation'])]):
 home=scratch/('task-'+str(k+1));home.mkdir();work=home/'module';shutil.copytree(suite/'files/indexer-evolution',work);report=home/'report';report.mkdir();guides=home/'guidance';guides.mkdir();allowed=all_names[:4]+extra
 for n in allowed:shutil.copytree(run/'skills'/n,guides/n)
 descriptions=[]
 for p in sorted(guides.glob('*/SKILL.md')):
  desc=next(l.split(':',1)[1].strip().strip('"') for l in p.read_text().splitlines() if l.startswith('description:'));descriptions.append(f'- {p.parent.name}: {desc} (path {p})')
 wrapper=f'''You are completing a standalone Go task in an isolated module at {work}. This is actual implementation work, not a review. Work only in that module; write report files only in {report}. Do not inspect the skills repository, design records, eval JSON, controller probes, other author workspaces or results. Those are outside this task. Do not delegate, install dependencies/tools, commit or make external changes.

Read this module's README and source. Complete this request:
{prompt}

The only optional writing guidance offered for this task is:
{chr(10).join(descriptions)}
Read only guidance whose description is relevant. Record actual opened files and relevance decisions. Do not open any other writing or review skills, including unavailable testing guidance. No expected evaluation judgments or controller probes are supplied.

Keep the supported Go version. Run meaningful verification with finite test/process deadlines. Prefix every shell command with rtk and use rtk proxy for raw output. Use GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. If local listeners are restricted, preserve the failure and any approved unchanged-source rerun separately; disclose the exercised boundary.

When done, write {report}/selection.json with offered names, opened skill/reference paths and selection reasons. Write {report}/author-report.md with changes, actual checks, known limitations and remaining risks. Preserve command arrays, cwd, non-secret environment, stdout/stderr/exit codes in {report}/checks.json when possible. Return concise results and report paths.'''
 arc=run/arm;arc.mkdir(exist_ok=True);(arc/'prompt.txt').write_text(prompt);(arc/'dispatch.txt').write_text(wrapper)
 launch=f'Read and execute only the standalone Go author dispatch at {arc}/dispatch.txt. That single file is authorized input and identifies your isolated module and optional writing guidance. Do not inspect any other repository files or another author workspace. No delegation.'
 (arc/'launch.txt').write_text(launch)
 manifest['cases'].append(dict(id=arm,fixture_id='indexer-evolution',isolation_relevant=True,workdir=str(work),reportdir=str(report),available_skills=allowed,input_sha256=hashes(work),guidance_sha256=hashes(guides),dispatch_sha256=hashlib.sha256(wrapper.encode()).hexdigest()))
 print(arm,launch)
assert len({json.dumps(c['input_sha256'],sort_keys=True) for c in manifest['cases']})==1
(run/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n');print('PASS four input hash maps and inherited settings identical; exact launches frozen')
