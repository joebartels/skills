from pathlib import Path
import hashlib,difflib,json,tempfile,shutil,subprocess,sys
arc=Path.cwd()/'tests/go-quality-build/results'/(sys.argv[2] if len(sys.argv)>2 else '2026-10-01-testing-combined')/sys.argv[1];context=json.loads((arc/'blind-review-context.json').read_text());scratch=Path(context.get('case_scratch',context['scratch']));candidate=arc/'candidate';records=[]
def sources(d):return {str(p.relative_to(d)):p for p in sorted(d.rglob('*')) if p.is_file() and p.suffix in {'.go','.mod','.sum','.md'}}
def hashes(d):return {n:hashlib.sha256(p.read_bytes()).hexdigest() for n,p in sources(d).items()}
assert hashes(scratch/'candidate')==hashes(candidate)
folders=[f for f in scratch.rglob('go.mod') if 'candidate' not in f.relative_to(scratch).parts and 'original' not in f.relative_to(scratch).parts]
for folder in sorted({f.parent for f in folders}):
 if not folder.is_dir():continue
 old=sources(candidate);new=sources(folder);patch=''
 for n in sorted(set(old)|set(new)):
  a=old.get(n);b=new.get(n);left=a.read_text().splitlines(True) if a else [];right=b.read_text().splitlines(True) if b else []
  patch+=''.join(difflib.unified_diff(left,right,fromfile='a/'+n if a else '/dev/null',tofile='b/'+n if b else '/dev/null'))
 if not patch:continue
 out=arc/'review-mutations'/str(folder.relative_to(scratch)).replace('/','__');out.mkdir(parents=True,exist_ok=True);(out/'source.patch').write_text(patch)
 changed={n:hashlib.sha256(p.read_bytes()).hexdigest() for n,p in new.items() if n not in old or p.read_bytes()!=old[n].read_bytes()}
 for n in changed:p=out/'changed-source'/n;p.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(new[n],p)
 with tempfile.TemporaryDirectory(prefix='review-mutation-reconstruct-',dir='/private/tmp') as tmp:
  dest=Path(tmp)/'module';shutil.copytree(candidate,dest);r=subprocess.run(['rtk','proxy','git','apply',str(out/'source.patch')],cwd=dest,capture_output=True,text=True);assert r.returncode==0,r.stderr;assert hashes(dest)==hashes(folder)
 records.append(dict(reviewer_verification_cwd=str(folder),patch=str((out/'source.patch').relative_to(arc)),resulting_source_sha256=hashes(folder),changed_source_sha256=changed,reconstruction='exact source bytes; no binaries copied',kind='deliberate regression' if any(part.startswith(('mutation','verification-')) for part in folder.relative_to(scratch).parts) else 'contract reproduction; see raw checks for interpretation',checks='raw blind-review-checks.json preserves corresponding commands/results'))
(arc/'review-mutation-preservation.json').write_text(json.dumps(records,indent=2)+'\n');print('Preserved',len(records),'exact reviewer mutation reconstructions')
