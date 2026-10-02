import json
import pathlib
import shutil

import run_checks as runner

BASE = runner.BASE
runner.checks = json.loads((runner.OUT / 'checks.json').read_text())

def mutant(name, change):
    target = BASE / 'mutations' / name
    target.parent.mkdir(exist_ok=True)
    shutil.copytree(BASE / 'candidate', target, dirs_exist_ok=True)
    path = target / 'poll.go'
    source = path.read_text()
    revised = change(source)
    assert revised != source
    path.write_text(revised)
    return target

def replace_one(source, old, new):
    assert source.count(old) == 1
    return source.replace(old, new)

target = mutant('start-relative', lambda s: replace_one(s, '\t\tif err := callback(ctx); err != nil {\n\t\t\treturn err\n\t\t}\n\t\ttimer := time.NewTimer(interval)', '\t\ttimer := time.NewTimer(interval)\n\t\tif err := callback(ctx); err != nil {\n\t\t\ttimer.Stop()\n\t\t\treturn err\n\t\t}'))
runner.go('mutation-start-relative', target, 'test','-count=1','-run=^TestRecurrence$','-timeout=15s','./...')

target = mutant('async-return-on-cancel', lambda s: replace_one(s, '\t\tif err := callback(ctx); err != nil {\n\t\t\treturn err\n\t\t}', '\t\tresult := make(chan error, 1)\n\t\tgo func() { result <- callback(ctx) }()\n\t\tselect {\n\t\tcase <-ctx.Done():\n\t\t\treturn nil\n\t\tcase err := <-result:\n\t\t\tif err != nil {\n\t\t\t\treturn err\n\t\t\t}\n\t\t}'))
runner.go('mutation-early-return', target, 'test','-count=1','-run=^TestCancellationDuringCallbackCleanup$','-timeout=15s','./...')

target = mutant('cancellation-masks-error', lambda s: replace_one(s, '\t\tif err := callback(ctx); err != nil {\n\t\t\treturn err\n\t\t}', '\t\tif err := callback(ctx); err != nil {\n\t\t\tif ctx.Err() != nil { return nil }\n\t\t\treturn err\n\t\t}'))
runner.go('mutation-cancellation-masks-error', target, 'test','-count=1','-run=^TestCallbackErrorConcurrentWithCancellation$','-timeout=15s','./...')

target = mutant('wrapped-error', lambda s: replace_one(replace_one(s, '\t"errors"', '\t"errors"\n\t"fmt"'), '\t\t\treturn err', '\t\t\treturn fmt.Errorf("callback: %w", err)'))
runner.go('mutation-wrapped-error', target, 'test','-count=1','-run=^TestCallbackError$','-timeout=15s','./...')

target = BASE / 'mutations' / 'forced-early-fatal'
shutil.copytree(BASE / 'candidate', target, dirs_exist_ok=True)
testpath = target / 'poll_test.go'
source = testpath.read_text()
source = replace_one(source, '\trun := startPoll(t, ctx, time.Hour, func(callbackCtx context.Context) error {\n\t\tif calls.Add(1) != 1', '\tvar run *pollRun\n\tt.Cleanup(func() {\n\t\tselect {\n\t\tcase <-run.done:\n\t\tdefault:\n\t\t\tt.Error("review probe: worker still running after helper cleanup")\n\t\t}\n\t\tselect {\n\t\tcase <-cleanupFinished:\n\t\t\tt.Log("review probe: early Fatal released callback cleanup gate and joined Poll")\n\t\tdefault:\n\t\t\tt.Error("review probe: callback cleanup still blocked")\n\t\t}\n\t})\n\trun = startPoll(t, ctx, time.Hour, func(callbackCtx context.Context) error {\n\t\tif calls.Add(1) != 1')
source = replace_one(source, '\treceive(t, cleanupStarted, "callback cleanup start")', '\treceive(t, cleanupStarted, "callback cleanup start")\n\tt.Fatal("review probe: intentional early Fatal while callback cleanup is blocked")')
testpath.write_text(source)
runner.go('forced-early-fatal-cleanup', target, 'test','-race','-v','-count=1','-run=^TestCancellationDuringCallbackCleanup$','-timeout=15s','./...')

target = BASE / 'mutations' / 'forced-early-fatal-no-gate-release'
shutil.copytree(BASE / 'mutations' / 'forced-early-fatal', target, dirs_exist_ok=True)
testpath = target / 'poll_test.go'
source = testpath.read_text()
# Remove the cleanup gate release from the target test's stop operation. Its
# callback intentionally waits on that gate after acknowledging cancellation.
start = source.index('func TestCancellationDuringCallbackCleanup(')
end = source.index('func TestCallbackErrorConcurrentWithCancellation(')
section = replace_one(source[start:end], '\tstop := func() {\n\t\tcancel()\n\t\treleaseOnce.Do(func() { close(release) })\n\t}', '\tstop := func() { cancel() }')
source = source[:start] + section + source[end:]
testpath.write_text(source)
runner.go('forced-early-fatal-with-broken-stop', target, 'test','-v','-count=1','-run=^TestCancellationDuringCallbackCleanup$','-timeout=15s','./...')
