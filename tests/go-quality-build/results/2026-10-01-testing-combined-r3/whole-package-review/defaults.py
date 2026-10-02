from pathlib import Path
exec(Path('/private/tmp/go-testing-final-package-review/verify.py').read_text().split("run(['git','status'")[0])
records=json.loads((out/'checks.json').read_text())
home=out/'default-home';home.mkdir(exist_ok=True);defaultenv=dict(env,HOME=str(home));defaultenv.pop('GOCACHE',None)
for arm in ['isolation-only','both']:
 dest=out/('integrated-'+arm)
 run([str(dest/'suite1'),'-test.count=1','-test.timeout=90s'],dest/'cmd/indexer',label=arm+' compiled command suite unset GOCACHE',custom=defaultenv)
