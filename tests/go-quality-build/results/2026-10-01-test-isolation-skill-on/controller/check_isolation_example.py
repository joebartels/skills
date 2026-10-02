from pathlib import Path
import hashlib,json,os,subprocess,tempfile

root=Path.cwd();ref=root/'docs/go-quality-build/drafts/go-test-isolation/references/isolation-patterns.md'
run=root/'tests/go-quality-build/results/2026-10-01-test-isolation-skill-on';run.mkdir(exist_ok=True)
text=ref.read_text();code=text.split('```go\n',1)[1].split('```',1)[0]
formatted=subprocess.run(['rtk','proxy','gofmt'],input=code,text=True,capture_output=True,check=True).stdout
ref.write_text(text.replace(code,formatted,1));code=formatted
records=[]
with tempfile.TemporaryDirectory(prefix='go-isolation-example-',dir='/private/tmp') as tmp:
    work=Path(tmp);(work/'go.mod').write_text('module example.com/isolationexample\n\ngo 1.22\n')
    for phase,source in [('premature-close-counterexample',code.replace('\tgo func() {','\tfile.Close()\n\tgo func() {',1)),('exact-reference',code)]:
        (work/'example_test.go').write_text(source)
        cmd=['rtk','proxy','go','test','-race','-count=1','-timeout=20s','./...']
        p=subprocess.run(cmd,cwd=work,env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local'),text=True,capture_output=True,timeout=40)
        records.append(dict(phase=phase,source=source,command=cmd,cwd=str(work),stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode))
assert records[0]['exit_code']==1 and 'borrower fixture use:' in records[0]['stdout'],records[0]
assert records[1]['exit_code']==0,records[1]
(run/'example-verification.json').write_text(json.dumps(dict(declared_minimum='Go1.22-compatible APIs/source',executed_toolchain='Go1.26.5 darwin/arm64; actual minimum unavailable',optional_examples='No Go1.24/1.25 code example included; version-specific options are prose only.',reference_sha256=hashlib.sha256(ref.read_bytes()).hexdigest(),checks=records),indent=2)+'\n')
print('PASS premature fixture release assertion fails; exact original reference passes under race on host compiler.')
