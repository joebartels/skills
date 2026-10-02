from pathlib import Path
import json, shutil, tempfile, sys
root=Path.cwd();run=root/'tests/go-quality-build/results'/sys.argv[1];names=sys.argv[2:]
scratch=Path(tempfile.mkdtemp(prefix='go-independent-review-',dir='/private/tmp'));skills=scratch/'review-skills';skills.mkdir()
for name in ['go-testing','go-correctness-and-compatibility','go-architecture-and-design']:
 shutil.copytree(root/'plugins/go-quality-review/skills'/name,skills/name)
sections=[]
for i,name in enumerate(names):
 arc=run/name;home=scratch/('changeset-'+str(i+1));home.mkdir();out=home/'output';out.mkdir()
 for side in ['original','candidate']:shutil.copytree(arc/side,home/side)
 sections.append(f"Changeset {i+1}: original {home}/original; candidate {home}/candidate; separate output {out}. Actual request:\n{(arc/'prompt.txt').read_text()}")
 context=dict(scratch=str(scratch),case_scratch=str(home),report=str(out/'review.md'),checks=str(out/'checks.json'),label_disclosure=False,neutral_launch_path=True,author_report_disclosure=False,batching=f'{len(names)} independent changesets in one fresh review context; separate cards and checks; no comparative grading')
 (arc/'blind-review-context.json').write_text(json.dumps(context,indent=2)+'\n')
wrapper=f'''Independently review the following anonymized Go changesets. This explicitly batches {len(names)} separate source reviews in one fresh context. Assess each separately against its own README/task and emit separate unedited cards; do not compare their grades. You have no writing guidance, author reports or prior reviews. Do not inspect any other workspace, repository, evaluation expectations, controller probe or another candidate beyond those listed here. Do not delegate or alter original/candidate source. Disposable verification/mutation copies inside the corresponding changeset directory are allowed.

{chr(10).join(sections)}

Use {skills}/go-testing/SKILL.md and {skills}/go-correctness-and-compatibility/SKILL.md, including local decision references, for separate graded cards per changeset. Use {skills}/go-architecture-and-design/SKILL.md only for consequential production/test seams; otherwise state why architecture is not applicable. Assess introduced/worsened changes and the requested test scope, separating legacy limitations. Verify plausible findings and assertion signal with finite meaningful checks. Do not infer quality from test count, style or coverage percentage. No repairs to candidates.

Prefix shell commands with rtk, using rtk proxy for raw checks. Go uses GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. Do not install tools/dependencies. Preserve environmental failures and any approved unchanged-source listener rerun. Keep runtime/toolchain/platform limits explicit.

Write each changeset's review.md to its separate output directory with file/line findings, severity/grade rationale and verified safeguards. Preserve exact command/cwd/non-secret env/stdout/stderr/exit details in separate checks.json when possible. Return the report paths and concise results. Do not guess expected grades.'''
(scratch/'dispatch.txt').write_text(wrapper)
launch=f'Read and execute only the independent review dispatch at {scratch}/dispatch.txt. That file identifies authorized anonymized changesets and review guidance. No delegation or other repository inspection.'
for name in names:
 (run/name/'blind-review-dispatch.txt').write_text(wrapper);(run/name/'blind-review-launch.txt').write_text(launch)
print(launch)
