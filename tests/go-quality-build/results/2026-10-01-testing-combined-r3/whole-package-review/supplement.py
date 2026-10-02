from pathlib import Path
exec(Path('/private/tmp/go-testing-final-package-review/verify.py').read_text().split("run(['git','status'")[0])
records=json.loads((out/'checks.json').read_text())
for skill in ['go-behavior-tests','go-test-isolation']:run(['python3','-B','/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py','plugins/go-quality-build/skills/'+skill])
for loc in ['./plugins/go-quality-build','.']:run(['claude','plugin','validate',loc])
exclusions=['**/source.patch','tests/go-quality-build/results/2026-10-01-testing-combined/both/blind-review.md','tests/go-quality-build/results/2026-10-01-testing-combined/both/review-artifacts/inspection-transcript.json','tests/go-quality-build/results/2026-10-01-test-isolation-r3/library-file-fixtures/author-artifacts/store_test.diff']
prefix='tests/go-quality-build/results/2026-10-01-behavior-tests-skill-on/service-publication/'
exclusions += [prefix+'blind-review.md']+[prefix+'review-artifacts/'+f+'.py' for f in ['check_source_preserved','run_check','setup_mutations','setup_verification','validate_artifacts']]
(out/'whitespace-exclusions.json').write_text(json.dumps(exclusions,indent=2))
run(['git','diff','--check','b76df28ed15a177d436672d47a5d30f6634900c7','f48c68ddbdd31931cad3522aa6f9f1f17ea86d06','--','.',*[':(exclude)'+p for p in exclusions]],label='whole range whitespace complete raw exclusions')
for case in ['library-codec','cli-partial-failure']:
 source=root/'tests/go-quality-build/results/2026-10-01-behavior-tests-r2'/case/'candidate';dest=out/(case+'-mutant');shutil.copytree(source,dest,dirs_exist_ok=True)
 if case=='library-codec':
  p=dest/'key.go';s=p.read_text().replace('url.PathEscape(k.Region) + "/" + url.PathEscape(k.Name)','url.PathEscape(k.Name) + "/" + url.PathEscape(k.Region)').replace('Key{Region: region, Name: name}','Key{Region: name, Name: region}');p.write_text(s)
  run(['go','test','-run','^TestSimpleRoundTrip$','-count=1','./...'],dest,label='coordinated codec defect roundtrip survives')
 else:
  p=dest/'main.go';p.write_text(p.read_text().replace('os.Exit(2)','os.Exit(0)'))
 run(['go','test','-run','^$','./...'],dest,label=case+' mutant compiles')
 run(['go','test','-count=1','-timeout=40s','./...'],dest,label=case+' candidate-only mutation detection')
run(['go','test','-run','^TestStoreIndependentInstances$/^instances$/^gamma$/^replacement$','-parallel=1','-count=1','./...'],out/'library-file-fixtures',label='focused nested fixture selection')
info=json.loads((out/'integrity.json').read_text());canon=(root/'docs/go-quality-build/README.md').read_text()
assert all(r.get('checksum_index_sha256','') in canon for r in info['checks']);print('external seal anchors pass',flush=True)
(out/'anchors.json').write_text(json.dumps({'all_sealed_index_digests_in_canonical_record':True},indent=2))
for arm in ['isolation-only','both']:
 source=root/'tests/go-quality-build/results/2026-10-01-testing-combined-r3'/arm/'candidate';dest=out/('integrated-'+arm);shutil.copytree(source,dest,dirs_exist_ok=True)
 packages=run(['go','list','./...'],dest,label=arm+' packages')['stdout'].split()
 for n,pkg in enumerate(packages):run(['go','test','-c','-o',str(dest/('suite'+str(n))),pkg],dest,label=arm+' compile '+pkg)
