"""Reconstruct and check these three discovery outcomes; controller probes remain separate."""
from pathlib import Path
import concurrent.futures, hashlib, json, os, shutil, subprocess, tempfile, time
ROOT=Path(__file__).resolve().parents[4]
OUT=Path(__file__).resolve().parent
EVALS=ROOT/"tests/go-quality-build/go-context-and-deadlines/evals"
MIN="/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go"
ENV={"GOCACHE":"/private/tmp/go-quality-context-cache", "GOQUALITY_GO":MIN}
def command(argv,cwd,extra=None):
    env=os.environ.copy(); env.update(ENV); env.update(extra or {}); start=time.monotonic()
    p=subprocess.run(["rtk","proxy"]+argv,cwd=cwd,env=env,text=True,capture_output=True,timeout=90)
    return {"argv":["rtk","proxy"]+argv,"cwd":str(cwd),"environment_overrides":ENV|dict(extra or {}),"status":p.returncode,"stdout":p.stdout,"stderr":p.stderr,"elapsed_seconds":round(time.monotonic()-start,3)}
def verify(n,case):
    archive=OUT/f"attempt-{n:02}"; assert (archive/"completion.json").exists()
    original=EVALS/"files"/case; candidate=archive/"source"
    work=Path(tempfile.mkdtemp(prefix="go-check-",dir="/private/tmp")); shutil.copytree(original,work,dirs_exist_ok=True)
    check=command(["git","apply",str(archive/"source.patch")],work); assert check["status"]==0,check
    hashes={str(p.relative_to(work)):hashlib.sha256(p.read_bytes()).hexdigest() for p in work.rglob("*") if p.is_file()}
    assert hashes==json.loads((archive/"source-hashes.json").read_text()),case
    checks=[check]
    for argv in [["go","version"],["go","build","./..."],["go","test","-count=1","-timeout=30s","./..."],["go","vet","./..."],["gofmt","-l","."]]: checks.append(command(argv,work))
    shutil.copyfile(EVALS/"controller-probes"/case/"contract_test.go",work/"contract_controller_test.go")
    held=[command(["go","test","-count=1","-timeout=30s","./..."],work)]
    extra={"CGO_ENABLED":"0"} if case=="service-total-budget" else {}
    held.append(command([MIN,"test","-count=1","-timeout=30s","./..."],work,extra))
    if case!="not-context-work": held.append(command(["go","test","-race","-shuffle=on","-count=3","-timeout=60s","./..."],work))
    mutations=[]
    specifications=[]
    if case=="library-call-cause":
        specifications=[("unsafe-cause-equality","process.go","return accepted, errors.Join(callbackErr, cancellationError(ctx))", "if cause := context.Cause(ctx); cause != nil && cause != callbackErr { return accepted, errors.Join(callbackErr, cancellationError(ctx)) }; return accepted, errors.Join(callbackErr, ctx.Err())", "TestCauseReturnedByCallback"), ("lost-custom-cause","process.go","context.Cause(ctx)","nil", "TestCustomCauseAndIndependentFailure")]
    elif case=="service-total-budget":
        specifications=[("lost-total-budget","stages.go","context.WithTimeout(totalCtx, stage)","context.WithTimeout(ctx, stage)","TestTotalAndEarlierParentDeadline")]
    elif case=="not-context-work":
        specifications=[("control-upper-regression","range.go","value <= high","value < high","TestInclusiveUpperBound")]
    for name,file,old,new,test in specifications:
        mutation=Path(tempfile.mkdtemp(prefix="go-mutation-",dir="/private/tmp"));shutil.copytree(candidate,mutation,dirs_exist_ok=True)
        src=mutation/file; text=src.read_text(); assert old in text,(name,text); text=text.replace(old,new)
        if name=="lost-total-budget": text=text.replace("defer cancelTotal()", "defer cancelTotal()\n\t_ = totalCtx // Mutated stages intentionally ignore the total scope.")
        src.write_text(text);command(["gofmt","-w",file],mutation)
        row={"id":name,"file":file,"old":old,"new":new,"compile":command(["go","test","-run","^$","./..."],mutation),"author_tests":command(["go","test","-count=1","-timeout=30s","./..."],mutation)}
        shutil.copyfile(EVALS/"controller-probes"/case/"contract_test.go",mutation/"contract_controller_test.go")
        if name=="lost-total-budget": row["compilation_only_adjustment"]="_ = totalCtx; retained created/canceled scope while mutated stages derive from caller, avoiding an unrelated unused-variable error."
        row["held_tests"]=command(["go","test","-run","^"+test+"$","-count=1","-timeout=30s","./..."],mutation)
        mutations.append(row)
    assert all(row["status"]==0 for row in checks), (case,checks)
    assert all(row["status"]==(1 if case=="library-call-cause" else 0) for row in held), (case,held)
    assert all(row["compile"]["status"]==0 and row["held_tests"]["status"]==1 for row in mutations), (case,mutations)
    (archive/"verification.json").write_text(json.dumps({"reconstruction_hashes_match":True,"ordinary_checks":checks,"held_contract_checks":held,"mutation_sensitivity":mutations},indent=2)+"\n")
    events=[json.loads(x) for x in (archive/"events.jsonl").read_text().splitlines()]
    trace=[e["item"]["command"] for e in events if e.get("type")=="item.completed" and e.get("item",{}).get("type")=="command_execution" and ".allowed-skills/" in e["item"]["command"]]
    selection=json.loads((archive/"selection.json").read_text()); selection["observed_catalog_commands"]=trace;selection["limits"]=["Host scan warnings persist despite requested skip_host_skill_discovery; complete absence of host metadata not proven. Trace shows explicit catalog reads; this is not a native-routing trial."]
    (archive/"selection.json").write_text(json.dumps(selection,indent=2)+"\n")
    (archive/"cost.json").write_text(json.dumps({"usage":[e.get("usage") for e in events if e.get("type")=="turn.completed"],"elapsed_seconds":json.loads((archive/"completion.json").read_text())["elapsed_seconds"],"price_not_measured":True},indent=2)+"\n")
    print(n,case,"ordinary",[x["status"] for x in checks],"held",[x["status"] for x in held],"mutations",[(x["id"],x["compile"]["status"],x["author_tests"]["status"],x["held_tests"]["status"]) for x in mutations],flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    list(pool.map(lambda pair:verify(*pair),[(4,"library-call-cause"),(5,"service-total-budget"),(6,"not-context-work")]))
