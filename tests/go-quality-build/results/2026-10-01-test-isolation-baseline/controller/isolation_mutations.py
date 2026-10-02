from pathlib import Path
import difflib, json, sys

run, case = sys.argv[1:3]
arc = Path('tests/go-quality-build/results') / run / case
out = Path('.superpowers/sdd/2026-10-01-go-quality-build-testing-group/isolation-mutations') / run / case
out.mkdir(parents=True, exist_ok=True)

def patch(name, file, old, new):
    source = (arc / 'candidate' / file).read_text()
    assert old in source, (case, name)
    changed = source.replace(old, new, 1)
    path = out / (name + '.patch')
    path.write_text(''.join(difflib.unified_diff(source.splitlines(True), changed.splitlines(True), fromfile='a/'+file, tofile='b/'+file)))
    print(name, path)

if case == 'cli-environment':
    patch('environment-leak', 'config_test.go', 't.Setenv(key, value.value)', 'if err := os.Setenv(key, value.value); err != nil { t.Fatal(err) }')
elif case == 'worker-timing':
    patch('poll-discard-context', 'poll.go', 'callback(ctx)', 'callback(context.Background())')
    patch('poll-start-relative', 'poll.go', 'for {\n', 'ticker := time.NewTicker(interval)\n\tdefer ticker.Stop()\n\tfor {\n')
    p = out / 'poll-start-relative.patch'
    source = (arc/'candidate/poll.go').read_text()
    changed = source.replace('for {\n', 'ticker := time.NewTicker(interval)\n\tdefer ticker.Stop()\n\tfor {\n', 1).replace('\t\ttimer := time.NewTimer(interval)\n', '').replace('\t\t\ttimer.Stop()\n', '').replace('<-timer.C:', '<-ticker.C:')
    p.write_text(''.join(difflib.unified_diff(source.splitlines(True), changed.splitlines(True), fromfile='a/poll.go', tofile='b/poll.go')))
elif case == 'service-http-boundary':
    patch('http-ignore-cancellation', 'fetch.go', 'http.NewRequestWithContext(ctx,', 'http.NewRequestWithContext(context.Background(),')
    patch('http-body-not-closed', 'fetch.go', '\tdefer response.Body.Close()\n', '')
