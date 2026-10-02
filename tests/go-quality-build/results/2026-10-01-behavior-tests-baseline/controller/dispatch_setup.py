from pathlib import Path
import hashlib, json, shutil, subprocess, tempfile, sys

root=Path.cwd()
suite_name, run_id = sys.argv[1:3]
suite=root/"tests/go-quality-build"/suite_name/"evals"
run=root/"tests/go-quality-build/results"/run_id
run.mkdir(parents=True,exist_ok=True)
scratch=Path(tempfile.mkdtemp(prefix="go-testing-author-",dir="/private/tmp"))
skills=root/"plugins/go-quality-build/skills"
guidance=scratch/"guidance"
guidance.mkdir()
for name in ["go-package-boundaries","go-api-contracts","go-interfaces-and-composition"]:
 shutil.copytree(skills/name,guidance/name)
shutil.copytree("/Users/jb/.agents/skills/golang",guidance/"go-core-style")
shutil.copytree(guidance,run/"skills",dirs_exist_ok=True)
def sha(path):
 return {str(p.relative_to(path)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(path.rglob("*")) if p.is_file()}
commit=subprocess.check_output(["rtk","proxy","git","rev-parse","HEAD"],text=True).strip()
data=json.loads((suite/"evals.json").read_text())
manifest=dict(run_id=run_id,stage="prepared, authors not dispatched yet",fixture_commit=commit,skill_commit=commit,
 author_settings={"model":"inherited session model; exact model ID not exposed by collaboration harness","reasoning":"inherited session setting; no override","fresh_context":"collaboration.spawn_agent fork_turns=none","harness":"Codex collaboration tools"},
 host={"go":"go1.26.5","platform":"darwin/arm64","GOTOOLCHAIN":"local"},
 control={"other_testing_skill":"absent","task_and_fixture":"frozen","review_judgments":"withheld"},
 skill_sha256=sha(run/"skills"),cases=[])
for case in data["evals"]:
 name=case["id"]; work=scratch/name; shutil.copytree(suite/"files"/name,work)
 out=scratch/(name+"-report"); out.mkdir()
 arc=run/name; arc.mkdir(exist_ok=True)
 prompt=case["prompt"]
 descriptions=[]
 for p in sorted(guidance.glob("*/SKILL.md")):
  text=p.read_text(); desc=next(l.split(":",1)[1].strip().strip('"') for l in text.splitlines() if l.startswith("description:"))
  descriptions.append(f"- {p.parent.name}: {desc} (path {p})")
 wrapper=f"""You are completing a standalone Go task in an isolated module at {work}. This is actual implementation work, not a review. Work only in that module; write your report files only in {out}. Do not inspect the skills repository, its design records, eval JSON, controller probes, other author workspaces or results. Those are outside this Go task. Do not delegate, install dependencies/tools, commit, or make external changes.

Read this module's README and source. Complete this request:
{prompt}

The only optional writing guidance offered for this task is:
{chr(10).join(descriptions)}
Read guidance only if its description is relevant. Record actual opened files and your relevance decision; do not open any other writing or review skill. The two new testing-writing skills are unavailable in this task.

Keep the module's supported Go version. Run meaningful verification, with finite test/process deadlines. Prefix every shell command with rtk, using rtk proxy for raw output. Use GOCACHE=/private/tmp/go-quality-testing-cache and GOTOOLCHAIN=local when running Go. Do not include controller probes or expected evaluation judgments.

When done, write {out}/selection.json with offered guidance names, opened skill/reference paths, and reasons for selection/nonselection. Write {out}/author-report.md with the change, checks actually run, known limitations, and remaining behavior risks. Preserve verification stdout/stderr/exit codes in {out}/checks.json when possible. Return a concise final report and these paths."""
 (arc/"prompt.txt").write_text(prompt)
 (arc/"dispatch.txt").write_text(wrapper)
 manifest["cases"].append(dict(id=name,workdir=str(work),reportdir=str(out),input_sha256=sha(work),dispatch_sha256=hashlib.sha256(wrapper.encode()).hexdigest()))
 print(name,work,out)
(run/"manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
print("scratch",scratch)
