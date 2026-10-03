"""Capture author-specific spellings of the frozen lifecycle mutations; never edit outcomes."""
from pathlib import Path
import concurrent.futures, json, os, shutil, subprocess, sys, tempfile

OUT = Path(__file__).resolve().parent
EVALS = OUT.parents[3] / "tests/go-quality-build/go-concurrency-and-ownership/evals"
ENV = os.environ | {"GOTOOLCHAIN": "local", "GOCACHE": "/private/tmp/go-quality-concurrency-cache"}

def run(argv, work):
    try:
        p = subprocess.run(["rtk", "proxy"] + argv, cwd=work, env=ENV, text=True,
                           capture_output=True, timeout=40)
        return dict(argv=["rtk", "proxy"] + argv, cwd=str(work), status=p.returncode,
                    stdout=p.stdout, stderr=p.stderr)
    except subprocess.TimeoutExpired as e:
        return dict(argv=argv, cwd=str(work), status=124, stdout=str(e.stdout), stderr=str(e.stderr))

def verify(number):
    archive = OUT / f"attempt-{number:02d}"
    source = (archive / "source/host.go").read_text()
    if "for i := 0; i < started; i++ {" in source:
        join = ("for i := 0; i < started; i++ {", "for i := 0; i < 0*started; i++ {")
    elif "for running > 0 {" in source:
        join = ("for running > 0 {", "for running > 0 && false {")
    elif "completed := 0" in source and "wg.Wait()" in source:
        start = source.index("\t\tcompleted := 0")
        finish = source.index("\t\twg.Wait()", start) + len("\t\twg.Wait()")
        join = (source[start:finish], "\t\tfailed := admissionErr != nil || ctx.Err() != nil")
    elif "workers.Wait()" in source:
        join = ("workers.Wait()", "_ = workers // mutation: no completion observation")
    else:
        join = None
    signal = next((s for s in ("case <-cohortCtx.Done():", "case <-runCtx.Done():", "case <-ctx.Done():") if s in source), None)
    variants = [("release-before-completion", join, "^TestControllerPartialStartAndBlockedAdmission$"),
                ("unresponsive-admission", (signal, "case <-make(chan struct{}):") if signal else None,
                 "^TestControllerCancellationWithOpenInput$")]
    records = []
    for meaning, substitution, target in variants:
        if substitution is None:
            records.append(dict(meaning=meaning, qualified=False, reason="No equivalent spelling selected; no detection credit."))
            continue
        work = Path(tempfile.mkdtemp(prefix="go-study-mutation-", dir="/private/tmp"))
        shutil.copytree(archive / "source", work, dirs_exist_ok=True)
        held = work / "controller_contract_test.go"
        shutil.copyfile(EVALS / "controller-probes/service-owned-workers/contract_test.go", held)
        before = run(["go", "test", "-run", target, "-count=1", "-timeout=30s", "./..."], work)
        held.unlink()
        old, new = substitution
        (work / "host.go").write_text(source.replace(old, new))
        compiled = run(["go", "test", "-run", "^$", "./..."], work)
        author = run(["go", "test", "-count=1", "-timeout=30s", "./..."], work)
        shutil.copyfile(EVALS / "controller-probes/service-owned-workers/contract_test.go", held)
        after = run(["go", "test", "-run", target, "-count=1", "-timeout=30s", "./..."], work)
        records.append(dict(meaning=meaning, substitution=dict(file="host.go", old=old, new=new),
                            original_held_named=before, compile=compiled, author_tests=author, held_named=after,
                            qualified=before["status"] == 0 and compiled["status"] == 0 and after["status"] == 1,
                            limit="Inspect intended assertion before detection credit; compile failures/invalid original target excluded."))
    (archive / "mutations.json").write_text(json.dumps(records, indent=2) + "\n")
    print(number, [(r["meaning"], r["qualified"], r.get("author_tests", {}).get("status")) for r in records], flush=True)

with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    list(pool.map(verify, map(int, sys.argv[1:])))
