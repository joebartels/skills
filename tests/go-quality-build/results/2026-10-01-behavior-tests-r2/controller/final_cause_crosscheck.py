from pathlib import Path
import difflib, hashlib, json, os, shutil, subprocess, tempfile, time

root = Path.cwd()
results = root / 'tests/go-quality-build/results'
run = results / '2026-10-01-testing-combined-final'
probe = '''package indexer_test

import (
    "context"
    "errors"
    "fmt"
    "testing"
    "time"
    "example.com/indexer"
)

type diagnosticCauseError struct { details []string }
func (e diagnosticCauseError) Error() string { return "shutdown cause" }

func TestDiagnosticNonComparableCause(t *testing.T) {
    for _, joined := range []bool{false, true} {
        t.Run(fmt.Sprintf("joined=%v", joined), func(t *testing.T) {
            ctx, cancel := context.WithCancelCause(context.Background())
            defer cancel(nil)
            cause := diagnosticCauseError{details: []string{"shutdown"}}
            independent := errors.New("independent callback failure")
            completed, releasedAfterCompletion := false, false
            releases := 0
            defer func() {
                if p := recover(); p != nil { t.Errorf("Serve panicked for valid non-comparable cancellation cause: %v", p) }
                if releases != 1 || !releasedAfterCompletion { t.Errorf("release count/order = %d/%v, want 1/true", releases, releasedAfterCompletion) }
            }()
            err := indexer.Serve(ctx, time.Hour, func(context.Context) error {
                cancel(cause)
                completed = true
                wrapped := fmt.Errorf("fetch stopped: %w", cause)
                if joined { return errors.Join(wrapped, independent) }
                return wrapped
            }, func() error { releases++; releasedAfterCompletion = completed; return nil })
            if joined && !errors.Is(err, independent) { t.Errorf("independent failure lost: %v", err) }
        })
    }
}
'''
def hashes(d): return {str(p.relative_to(d)): hashlib.sha256(p.read_bytes()).hexdigest() for p in d.rglob('*') if p.is_file()}
for prior, case in [('2026-10-01-testing-combined-final', 'behavior-only'), ('2026-10-01-testing-combined-final', 'both')]:
    candidate = results / prior / case / 'candidate'
    out = run / 'cause-crosscheck' / (prior + '--' + case)
    out.mkdir(parents=True, exist_ok=True)
    (out / 'diagnostic_cause_test.go').write_text(probe)
    records = []
    with tempfile.TemporaryDirectory(prefix='go-cause-crosscheck-', dir='/private/tmp') as tmp:
        work = Path(tmp) / 'module'; shutil.copytree(candidate, work)
        shutil.copy2(out / 'diagnostic_cause_test.go', work / 'diagnostic_cause_test.go')
        for args in [['gofmt', '-w', 'diagnostic_cause_test.go'], ['go', 'test', '-run', '^$', '-timeout=30s', './...'], ['go', 'test', '-run', '^TestDiagnosticNonComparableCause$', '-count=1', '-timeout=30s', './...']]:
            command = ['rtk', 'proxy', *args]; start = time.monotonic()
            p = subprocess.run(command, cwd=work, env=dict(os.environ, GOCACHE='/private/tmp/go-quality-testing-cache', GOTOOLCHAIN='local'), capture_output=True, text=True, timeout=55)
            records.append(dict(command=command, cwd=str(work), env={'GOCACHE':'/private/tmp/go-quality-testing-cache','GOTOOLCHAIN':'local'}, stdout=p.stdout, stderr=p.stderr, exit_code=p.returncode, duration_seconds=round(time.monotonic()-start,3)))
        shutil.copy2(work / 'diagnostic_cause_test.go', out / 'diagnostic_cause_test.go')
        patch = ''.join(difflib.unified_diff([], (work/'diagnostic_cause_test.go').read_text().splitlines(True), fromfile='/dev/null', tofile='b/diagnostic_cause_test.go'))
        (out/'source.patch').write_text(patch)
        before=hashes(candidate); after=hashes(work)
        rebuild=Path(tmp)/'rebuild';shutil.copytree(candidate,rebuild)
        p=subprocess.run(['rtk','proxy','git','apply',str(out/'source.patch')],cwd=rebuild,capture_output=True,text=True);assert p.returncode==0,p.stderr;assert hashes(rebuild)==after
    assert records[1]['exit_code']==0
    (out/'checks.json').write_text(json.dumps(dict(kind='independent contract reproduction; no candidate repair; no prescribed nil versus returned cause',candidate_run=prior,case=case,candidate_sha256=before,probe_added_sha256=after,reconstruction='exact',checks=records),indent=2)+'\n')
    print(prior,case,'compile',records[1]['exit_code'],'contract',records[2]['exit_code'])
