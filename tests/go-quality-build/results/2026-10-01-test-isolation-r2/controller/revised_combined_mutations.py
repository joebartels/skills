from pathlib import Path
import difflib, re, subprocess, sys
root=Path.cwd();w=root/'.superpowers/sdd/2026-10-01-go-quality-build-testing-group';run,case=sys.argv[1:3];source=root/'tests/go-quality-build/results'/run/case/'candidate';out=w/'revision-mutations'/case;out.mkdir(parents=True,exist_ok=True)
def patch(name,file,transform):
 old=(source/file).read_text();new=transform(old);assert new!=old,(name,file)
 p=out/(name+'.patch');p.write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file)));return name,p
def replace(old,new):
 def change(s):assert s.count(old)==1,(old,s.count(old));return s.replace(old,new,1)
 return change
trials=[patch('process-success-on-error','cmd/indexer/main.go',replace('os.Exit(2)','os.Exit(0)')),
patch('host-release-before-join','serve.go',replace('refresh(ctx)','func() error { result := make(chan error, 1); go func() { result <- refresh(ctx) }(); select { case e := <-result: return e; case <-ctx.Done(): return nil } }()')),
patch('discarded-caller-context','serve.go',replace('refresh(ctx)','refresh(context.Background())'))]
for file in ['index.go','refresh.go']:
 s=(source/file).read_text();matches=re.findall(r'return os.Rename\([^\n]+, path\)',s)
 if matches:
  assert len(matches)==1;trials.append(patch('direct-snapshot-truncation',file,replace(matches[0],'return os.WriteFile(path, data, 0600)')));break
else:raise AssertionError('Read candidate publication helper before adapting')
def interval(s):
 assert s.count('if err := refresh(ctx); err != nil {')==1 and s.count('time.NewTimer(interval)')==1
 return s.replace('if err := refresh(ctx); err != nil {','cycleStart := time.Now()\n if err := refresh(ctx); err != nil {',1).replace('time.NewTimer(interval)','time.NewTimer(interval - time.Since(cycleStart))',1)
trials.append(patch('start-relative-recurrence','serve.go',interval))
for file in ['index.go','refresh.go']:
 s=(source/file).read_text();matches=re.findall(r'strings.TrimSpace\([^)]+\) [!=]= ""',s)
 if matches:
  assert len(matches)==1;trials.append(patch('omitted-text-validation',file,replace(matches[0],matches[0][:-2]+'"\\x00"')));break
else:raise AssertionError('Read validation owner before adapting')
s=(source/'refresh.go').read_text();matches=re.findall(r'(?ms)^[ \t]*\w+\.CheckRedirect = func\(.*?return http\.ErrUseLastResponse[ \t\r\n]*\}\n',s,re.M);assert len(matches)==1,matches
trials.append(patch('automatic-redirect-following','refresh.go',replace(matches[0],'')))
for name,p in trials:
 r=subprocess.run(['rtk','proxy','python3','-B',str(w/'run_mutation.py'),run,case,name,str(p)],cwd=root);assert r.returncode==0,(name,r.returncode)
