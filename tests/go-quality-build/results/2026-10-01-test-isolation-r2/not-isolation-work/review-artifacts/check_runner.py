import json, os, pathlib, shutil, subprocess, sys, hashlib, datetime
root = pathlib.Path(sys.argv[1])
out = root / "output"
out.mkdir(exist_ok=True)
journal = out / "checks.json"
if journal.exists():
    state = json.loads(journal.read_text())
else:
    state = {"changeset": str(root), "environment": {"GOCACHE": "/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN": "local", "GOPROXY": "off"}, "checks": [], "operations": []}
env = dict(os.environ)
env.update(state["environment"])
def save():
    journal.write_text(json.dumps(state, indent=2) + "\n")
def check(label, args, cwd, timeout=60):
    command = ["rtk", "proxy"] + args
    at = datetime.datetime.now(datetime.timezone.utc).isoformat()
    try:
        done = subprocess.run(command, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout)
        entry = {"label": label, "command_argv": command, "cwd": str(cwd), "env_overrides": state["environment"], "started_at": at, "stdout": done.stdout, "stderr": done.stderr, "exit_code": done.returncode, "timeout_seconds": timeout}
    except subprocess.TimeoutExpired as exc:
        def decode(s): return s.decode(errors="replace") if isinstance(s, bytes) else s or ""
        entry = {"label": label, "command_argv": command, "cwd": str(cwd), "env_overrides": state["environment"], "started_at": at, "stdout": decode(exc.stdout), "stderr": decode(exc.stderr), "exit_code": None, "timeout_seconds": timeout, "timed_out": True}
    state["checks"].append(entry); save()
    print(json.dumps(entry), flush=True)
    return entry
def snapshot(source, name):
    dest = out / name
    if dest.exists(): shutil.rmtree(dest)
    shutil.copytree(source, dest)
    state["operations"].append({"kind": "disposable_copy", "source": str(source), "destination": str(dest)})
    save()
    return dest
def hashes():
    state["source_hashes"] = {str(f): hashlib.sha256(f.read_bytes()).hexdigest() for kind in ["original", "candidate"] for f in sorted((root/kind).glob("*")) if f.is_file()}
    save()
if sys.argv[2] == "basic":
    hashes()
    work = snapshot(root/"candidate", "verification")
    check("toolchain", ["go", "version"], work)
    check("platform_and_cgo", ["go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "GOVERSION", "GOMOD"], work)
    check("supplied_diff", ["diff", "-ru", str(root/"original"), str(root/"candidate")], root)
    check("candidate_tests", ["go", "test", "-mod=readonly", "-count=1", "-timeout=30s", "./..."], work)
    if root.name == "changeset-1":
        check("race_repeated_shuffled", ["go", "test", "-mod=readonly", "-race", "-count=20", "-shuffle=on", "-timeout=45s", "./..."], work, 60)
    else:
        check("ordinary_repeated_shuffled", ["go", "test", "-mod=readonly", "-count=20", "-shuffle=on", "-timeout=20s", "./..."], work)

elif sys.argv[2] == "signal":
    if root.name == "changeset-1":
        def mutate(name, filename, old, new, run_name, timeout=20):
            work = snapshot(root/"candidate", "mutation-" + name)
            f = work / filename
            content = f.read_text()
            if content.count(old) != 1:
                raise RuntimeError(f"expected one mutation site for {name}, got {content.count(old)}")
            f.write_text(content.replace(old, new))
            state["operations"].append({"kind": "mutation", "name": name, "path": str(f), "old": old, "new": new, "purpose": run_name})
            save()
            return check("mutation_"+name, ["go", "test", "-mod=readonly", "-count=1", "-run="+run_name, "-timeout="+str(timeout)+"s", "./..."], work, timeout+10)
        mutate("start_relative_timer", "poll.go",
            "\t\tif err := callback(ctx); err != nil {\n\t\t\treturn err\n\t\t}\n\t\ttimer := time.NewTimer(interval)",
            "\t\ttimer := time.NewTimer(interval)\n\t\tif err := callback(ctx); err != nil {\n\t\t\ttimer.Stop()\n\t\t\treturn err\n\t\t}",
            "^TestCompletionRelativeRecurrence$")
        mutate("cancellation_swallows_error", "poll.go",
            "\t\tif err := callback(ctx); err != nil {\n\t\t\treturn err\n\t\t}",
            "\t\tif err := callback(ctx); err != nil {\n\t\t\tif ctx.Err() != nil { return nil }\n\t\t\treturn err\n\t\t}",
            "^TestCallbackErrorWinsCancellation$")
        mutate("return_before_callback_cleanup", "poll.go",
            "\t\tif err := callback(ctx); err != nil {\n\t\t\treturn err\n\t\t}",
            "\t\tcallbackDone := make(chan error, 1)\n\t\tgo func() { callbackDone <- callback(ctx) }()\n\t\tselect {\n\t\tcase <-ctx.Done(): return nil\n\t\tcase err := <-callbackDone:\n\t\t\tif err != nil { return err }\n\t\t}",
            "^TestFixtureTeardown$")
        work = snapshot(root/"candidate", "mutation-shared_invocation_lock")
        f = work/"poll.go"
        content = f.read_text().replace('"errors"', '"errors"\n\t"sync"')
        content = content.replace("// Poll runs", "var invocationMu sync.Mutex\n\n// Poll runs")
        content = content.replace(" error {\n", " error {\n\tinvocationMu.Lock()\n\tdefer invocationMu.Unlock()\n", 1)
        f.write_text(content)
        state["operations"].append({"kind": "mutation", "name": "shared_invocation_lock", "path": str(f), "purpose": "serialize independent Poll invocations with a process-wide mutex"})
        save()
        check("mutation_shared_invocation_lock", ["go", "test", "-mod=readonly", "-count=1", "-run=^TestIndependentInvocations$", "-timeout=25s", "./..."], work, 35)
        work = snapshot(root/"candidate", "mutation-fixture_close_before_join")
        f = work/"poll_test.go"
        content = f.read_text()
        section_start = content.index("func TestFixtureCleanupOnReturn")
        before, section = content[:section_start], content[section_start:]
        cleanup_start = section.index("\t\tt.Cleanup(func() {")
        cleanup_end = section.index("\n\t\tctx, cancel :=", cleanup_start)
        cleanup = section[cleanup_start:cleanup_end]
        section = section[:cleanup_start] + section[cleanup_end:]
        site = '\t\twaitEvent(t, started, "callback borrowing fixture")'
        assert section.count(site) == 1
        section = section.replace(site, cleanup + "\n" + site)
        f.write_text(before + section)
        state["operations"].append({"kind": "mutation", "name": "fixture_close_before_join", "path": str(f), "purpose": "reverse cleanup registration order so fixture-close runs before cancel-and-join"})
        save()
        check("mutation_fixture_close_before_join", ["go", "test", "-mod=readonly", "-count=1", "-run=^TestFixtureCleanupOnReturn$", "-timeout=20s", "./..."], work, 30)
    else:
        work = snapshot(root/"candidate", "mutation-original_lower_bound")
        target = work/"clamp.go"
        original = root/"original"/"clamp.go"
        shutil.copyfile(original, target)
        state["operations"].append({"kind": "mutation", "name": "original_lower_bound", "source": str(original), "path": str(target), "purpose": "run candidate regression against the original faulty branch"})
        save()
        check("mutation_original_lower_bound", ["go", "test", "-mod=readonly", "-count=1", "-run=^TestLowerBound$", "-timeout=20s", "./..."], work)
        work = snapshot(root/"candidate", "boundary-probe")
        probe = r'''package clamp_test

import (
    "testing"
    "example.com/clamp"
)

func TestReviewSupportedBoundaries(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct{ value, low, high, want int }{
        {-8, -5, 3, -5}, {-5, -5, 3, -5}, {3, -5, 3, 3},
        {8, -5, 3, 3}, {7, 7, 7, 7}, {6, 7, 7, 7}, {8, 7, 7, 7},
        {min, min, max, min}, {max, min, max, max},
        {min, min+1, max, min+1}, {max, min, max-1, max-1},
    }
    for _, tc := range cases {
        if got := clamp.Clamp(tc.value, tc.low, tc.high); got != tc.want {
            t.Errorf("Clamp(%d,%d,%d)=%d,want %d", tc.value,tc.low,tc.high,got,tc.want)
        }
    }
}
'''
        f = work/"review_boundary_test.go"
        f.write_text(probe)
        state["operations"].append({"kind": "review_probe", "path": str(f), "source": probe, "purpose": "verify inclusive, degenerate, negative and native-int extreme supported ranges"})
        save()
        check("review_boundary_probe", ["go", "test", "-mod=readonly", "-count=1", "-run=^TestReviewSupportedBoundaries$", "-timeout=20s", "./..."], work)

elif sys.argv[2] == "final":
    work = out / "verification"
    check("effective_minimum_standard_library_check", ["go", "vet", "-stdversion", "./..."], work)
    actual = {str(f): hashlib.sha256(f.read_bytes()).hexdigest() for kind in ["original", "candidate"] for f in sorted((root/kind).glob("*")) if f.is_file()}
    state["source_integrity"] = {"match": actual == state["source_hashes"], "actual_sha256": actual}
    save()
    if not state["source_integrity"]["match"]:
        raise RuntimeError("Original or candidate source changed")
    print(json.dumps({"source_integrity": state["source_integrity"]["match"]}), flush=True)
