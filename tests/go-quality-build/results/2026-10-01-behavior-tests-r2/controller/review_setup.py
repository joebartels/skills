from pathlib import Path
import hashlib,json,shutil,tempfile,sys

root=Path.cwd()
run_id,name,tag=sys.argv[1:4]
run=root/"tests/go-quality-build/results"/run_id
arc=run/name
scratch=Path(tempfile.mkdtemp(prefix="go-independent-review-",dir="/private/tmp"))
shutil.copytree(arc/"original",scratch/"original")
shutil.copytree(arc/"candidate",scratch/"candidate")
skills=scratch/"review-skills"; skills.mkdir()
for skill in ["go-testing","go-correctness-and-compatibility","go-architecture-and-design"]:
 shutil.copytree(root/"plugins/go-quality-review/skills"/skill,skills/skill)
out=scratch/"output"; out.mkdir()
prompt=(arc/"prompt.txt").read_text()
wrapper=f"""Independently review the changeset between {scratch}/original and {scratch}/candidate. This is an anonymized Go changeset. You have no author report or prior review. Do not inspect any other workspace, skills repository, evaluation expectations, writing skill, controller probe or another candidate. Do not delegate or alter original/candidate source. Disposable verification/mutation copies inside {scratch} are allowed.

The actual user request was:
{prompt}

Read the supplied project README contracts and source, then use {skills}/go-testing/SKILL.md and {skills}/go-correctness-and-compatibility/SKILL.md for separate graded report cards. Read their local references where needed. Use {skills}/go-architecture-and-design/SKILL.md for a separate graded Architecture report as well when this integrated task changes production structure; explain any narrowly unsupported topic.

Assess introduced/worsened changes and the requested new test scope, separating legacy limitations. Verify plausible findings, useful test assertions and dependency/lifecycle control with meaningful finite checks. Do not infer quality from test count, style compliance or coverage percentage. Do not add code repairs to the candidate. Prefix shell commands with rtk, using rtk proxy for raw checks; Go commands use GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local. Toolchain/platform/network limits must be explicit.

Write unedited report cards and supporting verification facts to {out}/review.md, with substantiated file/line findings, severity/grade rationale and meaningful verified safeguards. Preserve exact command/stdout/stderr/exit information in {out}/checks.json when possible. Return report path and concise findings. Do not guess expected grades."""
(arc/(tag+"-dispatch.txt")).write_text(wrapper)
(scratch/"dispatch.txt").write_text(wrapper)
launch=f"Read and execute only the independent review dispatch at {scratch}/dispatch.txt. That single file is authorized input; it identifies anonymized source and review guidance. Do not inspect any repository files or another candidate. No delegation."
(arc/(tag+"-launch.txt")).write_text(launch)
(arc/(tag+"-context.json")).write_text(json.dumps(dict(scratch=str(scratch),report=str(out/"review.md"),checks=str(out/"checks.json"),label_disclosure=False,neutral_launch_path=True,author_report_disclosure=False),indent=2)+"\n")
print(launch)
