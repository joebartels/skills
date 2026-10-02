from pathlib import Path
import json,hashlib,shutil
root=Path.cwd(); ws=root/'.superpowers/sdd/2026-10-01-go-quality-build-testing-group'
run=root/'tests/go-quality-build/results/2026-10-01-test-isolation-skill-on'
review=Path('/private/tmp/go-isolation-promotion-review-20261001')
assert (review/'review.md').exists() and (review/'checks.json').exists()
shutil.copy2(review/'review.md',run/'task-review.md');shutil.copy2(review/'checks.json',run/'task-review-checks.json')
for p in review.iterdir():
 if p.is_file() and p.suffix in {'.json','.md','.py','.patch'} and p.name not in {'review.md','checks.json'}:
  d=run/'task-review-artifacts';d.mkdir(exist_ok=True);shutil.copy2(p,d/p.name)
for name in ['dispatch.md','launch.txt']:
 shutil.copy2(ws/('isolation-promotion-review-'+name),run/('task-review-'+name))
draft=root/'docs/go-quality-build/drafts/go-test-isolation'; dest=root/'plugins/go-quality-build/skills/go-test-isolation'; snap=run/'skills/go-test-isolation'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
files={str(p.relative_to(draft)):sha(p) for p in draft.rglob('*') if p.is_file()}
assert all(sha(snap/n)==h for n,h in files.items())
assert not dest.exists();shutil.move(str(draft),str(dest))
assert all(sha(dest/n)==h for n,h in files.items())
(run/'promotion-byte-verification.json').write_text(json.dumps(dict(source_commit='904511b076a98b3d434186389ed2de2a57c09380',snapshot_runtime_sha256=files,equality='exact evaluated bytes; live draft removed'),indent=2)+'\n')
for f in ['plugins/go-quality-build/plugin.json','plugins/go-quality-build/.claude-plugin/plugin.json','.claude-plugin/marketplace.json']:
 p=root/f;s=p.read_text().replace('package boundaries, API contracts, composition, and behavior testing.','package boundaries, API contracts, composition, behavior testing, and test isolation.');p.write_text(s)
p=root/'plugins/go-quality-build/README.md';s=p.read_text().replace(' Fixture/timing isolation guidance is still proposed.','')
s=s.replace('## Local use and installation','`go-test-isolation` is implemented and evaluated on [five matched cases](../../tests/go-quality-build/results/2026-10-01-test-isolation-skill-on/comparison.md). It controls dependency fidelity, fixture ownership, process state, asynchronous events and cleanup. Revision 1 removes a focused-subtest independence defect (Testing B to A+) while retaining deterministic persistence checks. Strong environment/HTTP/worker baselines remain A+ and the pure control skips isolation. The [independent promotion review](../../tests/go-quality-build/results/2026-10-01-test-isolation-skill-on/task-review.md) supports unchanged-byte promotion. These bounded results do not establish broader schedule/platform effectiveness or automatic routing.\n\n## Local use and installation')
s=s.replace('`, and `/go-quality-build:go-behavior-tests`','`, `/go-quality-build:go-behavior-tests`, and `/go-quality-build:go-test-isolation`');p.write_text(s)
p=root/'README.md';s=p.read_text().replace('| 4 evaluated skills | Package responsibilities, API contracts, interfaces/composition, and behavior testing. |','| 5 evaluated skills | Package responsibilities, API contracts, interfaces/composition, behavior testing, and test isolation. |')
s=s.replace('Automatic routing and actual Go 1.22 execution remain unverified.','The evaluated [isolation skill](plugins/go-quality-build/skills/go-test-isolation/SKILL.md) removes one confirmed focused-subtest defect while retaining fixture contamination checks; see its [bounded comparison](tests/go-quality-build/results/2026-10-01-test-isolation-skill-on/comparison.md). Automatic routing and actual Go 1.22 execution remain unverified.',1);p.write_text(s)
p=run/'README.md';s=p.read_text().replace('Draft/snapshot equality is verified; runtime promotion is pending independent task review.','Independent task review accepts promotion. Snapshot/runtime equality is verified and the live draft is removed; final delivery checks are archived.');p.write_text(s)
p=run/'comparison.md';s=p.read_text().replace('Promotion is pending independent task review.','Independent task review accepts unchanged-byte promotion.');p.write_text(s)
p=root/'docs/go-quality-build/testing-source-audit.md';s=p.read_text();s+='\n2026-10-01: Isolation revision 1 promoted unchanged after five matched outcomes and independent task review. No substantial upstream copy; exact evaluated snapshot is now canonical runtime. Four-arm combined evaluation follows.\n';p.write_text(s)
p=run/'manifest.json';data=json.loads(p.read_text());data['seal_stage']='completed matched isolation revision 1 and independent promotion review; exact evaluated bytes promoted unchanged; bounded focused-child benefit';data['promotion']=dict(decision='ACCEPT',source_commit='904511b076a98b3d434186389ed2de2a57c09380',runtime_sha256=files,version_bump='Task 6');p.write_text(json.dumps(data,indent=2)+'\n')
print('Promoted exact bytes',files)
