from pathlib import Path
import difflib,subprocess,sys
root=Path.cwd();ws=root/'.superpowers/sdd/2026-10-01-go-quality-build-testing-group';arm=sys.argv[1];arc=root/'tests/go-quality-build/results/2026-10-01-testing-combined'/arm;source=arc/'candidate';out=ws/'combined-mutations'/arm;out.mkdir(parents=True,exist_ok=True)
def patch(name,filename,old,new):
 p=source/filename;original=p.read_text();assert original.count(old)==1,(arm,name,original.count(old));changed=original.replace(old,new,1)
 path=out/(name+'.patch');path.write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+filename,tofile='b/'+filename)))
 return name,path
assert arm in {'behavior-only','isolation-only','architecture-only','both'},'Add explicit candidate-specific transformations after reading the candidate'
trials=[patch('process-success-on-error','cmd/indexer/main.go','os.Exit(2)','os.Exit(0)'),
patch('host-release-before-join','serve.go','refresh(ctx)','func() error { result := make(chan error, 1); go func() { result <- refresh(ctx) }(); select { case e := <-result: return e; case <-ctx.Done(): return nil } }()'),
patch('direct-snapshot-truncation', 'index.go' if arm in {'architecture-only','both'} else 'refresh.go', 'return os.Rename(f.Name(), path)' if arm in {'architecture-only','both'} else 'return os.Rename(file.Name(), path)', 'return os.WriteFile(path, data, 0600)'),
patch('discarded-caller-context','serve.go','refresh(ctx)','refresh(context.Background())'),
patch('start-relative-recurrence','serve.go','if err := refresh(ctx); err != nil {','cycleStart := time.Now()\n\t\tif err := refresh(ctx); err != nil {')]
name,path=trials[-1];text=path.read_text();original=(source/'serve.go').read_text();changed=original.replace('if err := refresh(ctx); err != nil {','cycleStart := time.Now()\n\t\tif err := refresh(ctx); err != nil {').replace('time.NewTimer(interval)','time.NewTimer(interval - time.Since(cycleStart))');path.write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/serve.go',tofile='b/serve.go')))
if arm=='behavior-only':
 trials.append(patch('omitted-text-validation','refresh.go','strings.TrimSpace(record.Text) == ""','strings.TrimSpace(record.Text) == "\\x00"'))
elif arm=='isolation-only':
 trials.append(patch('omitted-text-validation','index.go','strings.TrimSpace(text) != ""','strings.TrimSpace(text) != "\\x00"'))
elif arm=='both':
 trials.append(patch('omitted-text-validation','index.go','strings.TrimSpace(record.Text) == ""','strings.TrimSpace(record.Text) == "\\x00"'))
else:
 trials.append(patch('omitted-text-validation','index.go','strings.TrimSpace(text) == ""','strings.TrimSpace(text) == "\\x00"'))
for name,path in trials:
 r=subprocess.run(['rtk','proxy','python3','-B',str(ws/'run_mutation.py'),'2026-10-01-testing-combined',arm,name,str(path)],cwd=root);assert r.returncode==0,(name,r.returncode)
