from pathlib import Path
import json, shutil, sys

run, case = sys.argv[1:3]
arc=Path('tests/go-quality-build/results')/run/case
context=json.loads((arc/'blind-review-context.json').read_text())
out=Path(context['report']).parent
assert (out/'review.md').exists()
shutil.copy2(out/'review.md',arc/'blind-review.md')
if (out/'checks.json').exists():shutil.copy2(out/'checks.json',arc/'blind-review-checks.json')
for p in out.iterdir():
    if p.is_file() and p.suffix in {'.json','.md','.py','.patch'} and p.name not in {'review.md','checks.json'}:
        dest=arc/'review-artifacts';dest.mkdir(exist_ok=True);shutil.copy2(p,dest/p.name)
print(case, 'raw review preserved')
