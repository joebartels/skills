"""Exact native launcher for the three frozen context discovery tasks; no retries."""
from pathlib import Path
import concurrent.futures, difflib, hashlib, json, os, shutil, subprocess, tempfile, time

ROOT = Path(__file__).resolve().parents[4]
ARCHIVE = Path(__file__).resolve().parent
EVALS = ROOT / "tests/go-quality-build/go-context-and-deadlines/evals"
manifest = json.loads((ARCHIVE / "manifest.json").read_text())
assert len(manifest["attempts"]) == 3 and all(x["status"] == "failed" for x in manifest["attempts"]), "Exactly one retry per failed discovery is permitted."
cases = {case["id"]: case for case in json.loads((EVALS / "evals.json").read_text())["evals"]}
selected = ["library-call-cause", "service-total-budget", "not-context-work"]
prepared = []
for index, case_id in enumerate(selected, 4):
    case = cases[case_id]
    output = ARCHIVE / f"attempt-{index:02}"
    output.mkdir()
    workspace = Path(tempfile.mkdtemp(prefix="go-task-", dir="/private/tmp"))
    original = EVALS / "files" / case_id
    for source in case["files"]:
        src = EVALS / source
        relative = src.relative_to(original)
        dst = workspace / relative
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(src, dst)
    shutil.copytree(ARCHIVE / "skills", workspace / ".allowed-skills")
    catalog = []
    for skill in sorted((workspace / ".allowed-skills").glob("*/SKILL.md")):
        header = skill.read_text().split("---", 2)[1].strip()
        catalog.append(header + "\nfile: " + str(skill.relative_to(workspace)))
    prompt = (case["prompt"] + "\n\nWork on this task in the current directory. Read README.md and the supplied files; implement the requested behavior and meaningful tests. "
        "The allowed skill catalog below is available for this task. Select and read applicable SKILL.md files and their local references as needed; decline irrelevant skills. "
        "Use only this task and allowed catalog, without external skill discovery or delegating work. Do not inspect other task/evaluation/planning directories. "
        "Prefix shell commands with rtk (rtk proxy runs the original command). Preserve the specified Go version and dependency/API boundaries. "
        "In your final response report skills opened/used, changes, actual checks and limitations.\n\nAllowed skill catalog:\n" + "\n\n".join(catalog) + "\n")
    (output / "prompt.txt").write_text(prompt)
    argv = ["codex", "exec", "--ignore-user-config", "--ephemeral", "--model", "gpt-6-luna", "-c", 'model_reasoning_effort="medium"', "--enable", "skip_host_skill_discovery", "--disable", "multi_agent", "--approve-for-me", "--skip-git-repo-check", "-C", str(workspace), "--json", "-o", str(output / "author-report.md"), "-"]
    overrides = manifest["author_environment_overrides"]
    dispatch = {"argv": argv, "cwd": str(workspace), "environment_overrides": overrides, "stdin": "prompt.txt", "fresh_context": True, "case": case_id, "allowed_catalog": manifest["skill_hashes"]}
    (output / "dispatch.txt").write_text(json.dumps(dispatch, indent=2) + "\n")
    manifest["attempts"].append({"attempt": index, "case": case_id, "profile": "reference", "arm": "baseline", "retry_of": index-3, "status": "launch-reserved", "directory": output.name, "workspace": str(workspace)})
    prepared.append((case_id, original, workspace, output, argv, prompt, overrides))
manifest["preparation_commit"] = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip()
(ARCHIVE / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")

def launch(item):
    case_id, original, workspace, output, argv, prompt, overrides = item
    start = time.monotonic()
    env = os.environ.copy(); env.update(overrides)
    with (output / "events.jsonl").open("w") as stdout, (output / "stderr.txt").open("w") as stderr:
        result = subprocess.run(argv, input=prompt, text=True, cwd=workspace, env=env, stdout=stdout, stderr=stderr, timeout=1800)
    final = output / "source"; final.mkdir()
    for src in workspace.rglob("*"):
        rel = src.relative_to(workspace)
        if not src.is_file() or rel.parts[0] in (".allowed-skills", ".git"):
            continue
        dst = final / rel; dst.parent.mkdir(parents=True, exist_ok=True); shutil.copyfile(src, dst)
    old = {str(x.relative_to(original)): x.read_bytes() for x in original.rglob("*") if x.is_file()}
    new = {str(x.relative_to(final)): x.read_bytes() for x in final.rglob("*") if x.is_file()}
    patch = ""
    for name in sorted(old.keys() | new.keys()):
        if old.get(name) == new.get(name): continue
        patch += "".join(difflib.unified_diff(old.get(name,b"").decode().splitlines(keepends=True), new.get(name,b"").decode().splitlines(keepends=True), fromfile="a/"+name if name in old else "/dev/null", tofile="b/"+name if name in new else "/dev/null"))
    (output / "source.patch").write_text(patch)
    (output / "source-hashes.json").write_text(json.dumps({name:hashlib.sha256(data).hexdigest() for name,data in new.items()}, indent=2) + "\n")
    (output / "selection.json").write_text(json.dumps({"mode":"explicit allowed catalog; automatic native selection not tested", "evidence":"events.jsonl command trace and author-report.md declarations", "allowed_skills":list(manifest["skill_hashes"]), "host_discovery_disabled":True}, indent=2)+"\n")
    completion = {"exit_code":result.returncode,"elapsed_seconds":round(time.monotonic()-start,3),"status":"completed" if result.returncode==0 else "failed"}
    (output / "completion.json").write_text(json.dumps(completion,indent=2)+"\n")
    print(output.name,case_id,completion,flush=True)
    return completion

with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    completions = list(pool.map(launch, prepared))
for attempt, result in zip(manifest["attempts"][3:], completions): attempt.update(result)
(ARCHIVE / "manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
