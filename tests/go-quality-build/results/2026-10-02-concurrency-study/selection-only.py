"""Twelve bounded native selection probes, separate from implementation authors."""
from pathlib import Path
import hashlib,json,os,shutil,subprocess,tempfile,time,sys
OUT=Path(__file__).resolve().parent
DISC=OUT.parent/"2026-10-02-concurrency-discovery"
override=(DISC/"preparation/host-skill-disable-followed-override.txt").read_text().strip()
manifest=json.loads((OUT/"manifest.json").read_text());assert hashlib.sha256((override+"\n").encode()).hexdigest()==manifest["host_disable_sha256"]
profiles={"reference":("gpt-6-luna","medium"),"alternate-model":("gpt-6-sol","medium"),"alternate-reasoning":("gpt-6-luna","low")}
requests=[("applicable","Implement a Go host that starts available jobs up to a lease capacity, observes worker failure while acquiring more work, cooperatively stops and joins owned workers, then releases their leases."),("applicable-paraphrase","For a bounded Go worker host, dispatch available work without waiting for a full queue, supervise failures during input/acquisition, and wait for users before closing acquired resources."),("control","Fix a pure Go sum helper to include its final positive value and add one focused regression test. Keep it private and serial."),("control-paraphrase","Correct a small deterministic integer helper's last-element loop bound with a local assertion; no concurrency, I/O or resource lifetime changes.")]
assert len(sys.argv)==2 and sys.argv[1] in profiles
model,reasoning=profiles[sys.argv[1]]
for label,request in requests:
 out=OUT/"selection"/sys.argv[1]/label;assert not out.exists(),"No duplicate/retry"
 out.mkdir(parents=True);work=Path(tempfile.mkdtemp(prefix="go-select-",dir="/private/tmp"));shutil.copytree(OUT/"skills",work/".agents/skills")
 prompt=request+"\n\nSelection-only probe: do not implement, edit, test, browse externally, discover outside this workspace or delegate. Select applicable skills from the actual native catalog and read their SKILL.md/local references using rtk-prefixed commands if useful. Report which you selected/opened and why; explicitly decline irrelevant guidance. Do not read any other task."
 (out/"prompt.txt").write_text(prompt+"\n")
 argv=["codex","exec","--ignore-user-config","--ephemeral","--model",model,"-c",'model_reasoning_effort="'+reasoning+'"',"-c",override,"--enable","skip_host_skill_discovery","--disable","multi_agent","--approve-for-me","--skip-git-repo-check","-C",str(work),"--json","-o",str(out/"report.md"),"-"]
 (out/"dispatch.json").write_text(json.dumps({"argv":argv,"cwd":str(work),"stdin":"prompt.txt","kind":"native selection only","host_disable_sha256":manifest["host_disable_sha256"]},indent=2)+"\n")
 start=time.monotonic()
 with (out/"events.jsonl").open("w") as stream,(out/"stderr.txt").open("w") as err:
  result=subprocess.run(argv,input=prompt,text=True,cwd=work,stdout=stream,stderr=err,timeout=600)
 events=[json.loads(line) for line in (out/"events.jsonl").read_text().splitlines()]
 commands=[e["item"]["command"] for e in events if e.get("type")=="item.completed" and e.get("item",{}).get("type")=="command_execution"]
 summary={"status":result.returncode,"elapsed_seconds":round(time.monotonic()-start,3),"observed_commands":commands,"usage":[e.get("usage") for e in events if e.get("type")=="turn.completed"],"declarations":"report.md","mode":"native catalog, no forced explicit skill metadata/body"}
 (out/"completion.json").write_text(json.dumps(summary,indent=2)+"\n");print(sys.argv[1],label,summary["status"],summary["elapsed_seconds"],flush=True)
