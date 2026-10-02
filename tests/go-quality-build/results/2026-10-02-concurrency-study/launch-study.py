"""Fixed concurrency study slots4-12 after three discovery authors; reserve before dispatch and count all attempts."""
from pathlib import Path
import concurrent.futures,difflib,hashlib,json,os,shutil,subprocess,sys,tempfile,threading,time
ROOT=Path(__file__).resolve().parents[4];OUT=Path(__file__).resolve().parent
EVALS=ROOT/"tests/go-quality-build/go-concurrency-and-ownership/evals"
DISCOVERY=ROOT/"tests/go-quality-build/results/2026-10-02-concurrency-discovery"
BATCHES={"reference":[(4,"library-shared-state",True,"gpt-6-luna"),(5,"service-owned-workers",True,"gpt-6-luna"),(6,"not-concurrency-work",True,"gpt-6-luna")],"transfer":[(7,"cli-pipeline-stop",False,"gpt-6-luna"),(8,"cli-pipeline-stop",True,"gpt-6-luna")],"alternate-model":[(9,"service-owned-workers",False,"gpt-6-sol"),(10,"service-owned-workers",True,"gpt-6-sol")],"alternate-reasoning":[(11,"service-owned-workers",False,"gpt-6-luna"),(12,"service-owned-workers",True,"gpt-6-luna")]}
assert len(sys.argv)==2 and sys.argv[1] in BATCHES
manifest=json.loads((OUT/"manifest.json").read_text());assert manifest["guidance_frozen"]
cases={x["id"]:x for x in json.loads((EVALS/"evals.json").read_text())["evals"]}
override=(DISCOVERY/"preparation/host-skill-disable-followed-override.txt").read_text().strip()
assert hashlib.sha256((override+"\n").encode()).hexdigest()==manifest["host_disable_sha256"]
for name,expected in manifest["skill_hashes"].items():assert hashlib.sha256((OUT/"skills"/name).read_bytes()).hexdigest()==expected,name
lock=threading.Lock();prepared=[]
for number,case_id,exposed,model in BATCHES[sys.argv[1]]:
 assert not any(a["number"]==number for a in manifest["attempts"]),"Already reserved/dispatched; do not relaunch."
 case=cases[case_id];output=OUT/f"attempt-{number:02}";output.mkdir()
 workspace=Path(tempfile.mkdtemp(prefix="go-task-",dir="/private/tmp"));original=EVALS/"files"/case_id
 for name in case["files"]:
  src=EVALS/name;target=workspace/src.relative_to(original);target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(src,target)
 for skill in (OUT/"skills").iterdir():
  if skill.name=="go-concurrency-and-ownership" and not exposed:continue
  shutil.copytree(skill,workspace/".agents/skills"/skill.name)
 catalog=[]
 for skill in sorted((workspace/".agents/skills").glob("*/SKILL.md")):
  catalog.append(skill.read_text().split("---",2)[1].strip()+"\nfile: "+str(skill.relative_to(workspace)))
 prompt=case["prompt"]+"\n\nWork on this task in the current directory. Read README.md and the supplied files; implement the requested behavior and meaningful tests. The allowed skill catalog below is available for this task. Select and read applicable SKILL.md files and local references as needed; decline irrelevant skills. Use only this task and allowed catalog without external discovery or delegation. Do not inspect other tasks, evaluations, planning or result directories. Prefix shell commands with rtk (rtk proxy runs the original command). Preserve specified Go version and dependency/API boundaries. In your final response report skills opened/used, changes, actual checks and limitations.\n\nAllowed skill catalog:\n"+"\n\n".join(catalog)+"\n"
 (output/"prompt.txt").write_text(prompt)
 argv=["codex","exec","--ignore-user-config","--ephemeral","--model",model,"-c",'model_reasoning_effort="'+('low' if sys.argv[1]=='alternate-reasoning' else 'medium')+'"',"-c",override,"--enable","skip_host_skill_discovery","--disable","multi_agent","--approve-for-me","--skip-git-repo-check","-C",str(workspace),"--json","-o",str(output/"author-report.md"),"-"]
 env_overrides={"GOCACHE":"/private/tmp/go-quality-author-cache","GOMODCACHE":"/private/tmp/go-quality-author-modcache","GOTOOLCHAIN":"local"}
 (output/"dispatch.txt").write_text(json.dumps({"argv":argv,"cwd":str(workspace),"environment_overrides":env_overrides,"stdin":"prompt.txt","case":case_id,"native_local_catalog":True,"host_disable_sha256":manifest["host_disable_sha256"]},indent=2)+"\n")
 manifest["attempts"].append({"number":number,"case":case_id,"arm":"exposure" if exposed else "baseline","profile":sys.argv[1],"model":model,"reasoning":"low" if sys.argv[1]=="alternate-reasoning" else "medium","status":"reserved","counted":False,"directory":output.name})
 prepared.append((number,case_id,original,workspace,output,argv,prompt,env_overrides))
(OUT/"manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
def launch(item):
 number,case_id,original,workspace,output,argv,prompt,env_overrides=item
 with lock:
  row=next(x for x in manifest["attempts"] if x["number"]==number);row.update({"status":"launching","counted":True});(OUT/"manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
 env=os.environ.copy();env.update(env_overrides);start=time.monotonic();exit_code=124
 with (output/"events.jsonl").open("w") as stdout,(output/"stderr.txt").open("w") as stderr:
  try:
   result=subprocess.run(argv,input=prompt,text=True,cwd=workspace,env=env,stdout=stdout,stderr=stderr,timeout=1800);exit_code=result.returncode
  except subprocess.TimeoutExpired:
   stderr.write("\nAuthor launch watchdog expired; attempt counted, partial source/events retained.\n")
 final=output/"source";final.mkdir();generated=[]
 for source in workspace.rglob("*"):
  relative=source.relative_to(workspace)
  if not source.is_file() or relative.parts[0] in (".agents",".git"):continue
  data=source.read_bytes()
  try:data.decode()
  except UnicodeDecodeError:generated.append({"path":str(relative),"sha256":hashlib.sha256(data).hexdigest(),"reason":"generated binary omitted from authored source"});continue
  dst=final/relative;dst.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source,dst)
 old={str(p.relative_to(original)):p.read_bytes() for p in original.rglob("*") if p.is_file()};new={str(p.relative_to(final)):p.read_bytes() for p in final.rglob("*") if p.is_file()}
 patch=""
 for name in sorted(old.keys()|new.keys()):
  if old.get(name)==new.get(name):continue
  patch+="".join(difflib.unified_diff(old.get(name,b"").decode().splitlines(keepends=True),new.get(name,b"").decode().splitlines(keepends=True),fromfile="a/"+name if name in old else "/dev/null",tofile="b/"+name if name in new else "/dev/null"))
 (output/"source.patch").write_text(patch);(output/"source-hashes.json").write_text(json.dumps({name:hashlib.sha256(data).hexdigest() for name,data in new.items()},indent=2)+"\n")
 events=[json.loads(line) for line in (output/"events.jsonl").read_text().splitlines()]
 commands=[e["item"]["command"] for e in events if e.get("type")=="item.completed" and e.get("item",{}).get("type")=="command_execution" and "skills/" in e["item"]["command"]]
 (output/"selection.json").write_text(json.dumps({"mode":"checked native local catalog plus same explicit metadata; implementation selection distinct from selection-only probes","observed_skill_commands":commands,"declarations":"author-report.md","generated_files":generated},indent=2)+"\n")
 completion={"exit_code":exit_code,"elapsed_seconds":round(time.monotonic()-start,3),"status":"completed" if exit_code==0 else "failed"}
 (output/"completion.json").write_text(json.dumps(completion,indent=2)+"\n");(output/"cost.json").write_text(json.dumps({"usage":[e.get("usage") for e in events if e.get("type")=="turn.completed"],"elapsed_seconds":completion["elapsed_seconds"],"price_not_measured":True},indent=2)+"\n")
 with lock:
  next(x for x in manifest["attempts"] if x["number"]==number).update(completion);(OUT/"manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
 print(output.name,case_id,completion,flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:list(pool.map(launch,prepared))
