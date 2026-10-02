from pathlib import Path
import hashlib,shutil,json,sys
root=Path.cwd();name=sys.argv[1];run=sys.argv[2];draft=root/'docs/go-quality-build/drafts'/name;expected=root/'tests/go-quality-build/results'/run/'skills'/name;runtime=root/'plugins/go-quality-build/skills'/name
def hs(p):return {str(f.relative_to(p)):hashlib.sha256(f.read_bytes()).hexdigest() for f in p.rglob('*') if f.is_file()}
assert hs(draft)==hs(expected)
old=hs(runtime);shutil.rmtree(runtime);shutil.copytree(draft,runtime);assert hs(runtime)==hs(expected);shutil.rmtree(draft)
print(json.dumps(dict(skill=name,old_runtime=old,promoted_runtime=hs(runtime),exact_evaluated_snapshot=run),indent=2))
