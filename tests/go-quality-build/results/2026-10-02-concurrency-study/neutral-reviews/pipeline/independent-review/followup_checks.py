import json
import os
from pathlib import Path
import shutil
import subprocess
import time

ROOT=Path('/private/tmp/go-neutral-pipeline-study')
OUT=ROOT/'independent-review'
ENV=dict(os.environ,GOTOOLCHAIN='local',GOWORK='off',GOCACHE=str(OUT/'cache'),GOPROXY='off',GOSUMDB='off')
records=[]
def run(label,cwd,args):
    start=time.monotonic()
    p=subprocess.run(['rtk','proxy','go',*args],cwd=cwd,env=ENV,capture_output=True,text=True,timeout=15)
    r=dict(label=label,argv=['rtk','proxy','go',*args],cwd=str(cwd),status=p.returncode,stdout=p.stdout,stderr=p.stderr,elapsed_seconds=round(time.monotonic()-start,3))
    records.append(r)
    (OUT/'raw/followup-commands.json').write_text(json.dumps(records,indent=2)+'\n')
    print(label,p.returncode,flush=True)
    return r

# Remove the unrelated lost-already-observed-error path from the first A skip-join
# mutant. This version only skips a producer that has not finished.
dst=OUT/'work/A-mutant-skip-unfinished-producer-join'
shutil.copytree(ROOT/'candidates/A',dst)
file=dst/'cmd/pipeline/pipeline.go'
old='consumerStopped = r.stoppedAtReturn\n\t\t\t\tcancel()'
new=old+'\n\t\t\t\tif r.err == nil && !producerDone && ctx.Err() == nil { return nil }'
assert file.read_text().count(old)==1
file.write_text(file.read_text().replace(old,new))
(OUT/'raw/A-skip-unfinished-producer-join.patch.txt').write_text('old:\n'+old+'\nnew:\n'+new+'\n')
run('A/narrow-join/original-author-target-green',OUT/'work/A-ordinary',['test','-run','^TestEarlyConsumerCompletionStopsAndJoinsProducer$','-count=10','-timeout=10s','./...'])
run('A/narrow-join/original-probe-target-green',OUT/'work/A-probe',['test','-run','^TestReviewEarlyStopJoin$','-count=1','-timeout=10s','./...'])
run('A/narrow-join/compile',dst,['test','-run','^$','-count=1','-timeout=10s','./...'])
run('A/narrow-join/ordinary-suite',dst,['test','-count=10','-timeout=10s','./...'])
run('A/narrow-join/author-target',dst,['test','-run','^TestEarlyConsumerCompletionStopsAndJoinsProducer$','-count=10','-timeout=10s','./...'])
shutil.copy2(OUT/'probes/review_probe_test.go',dst/'cmd/pipeline/review_probe_test.go')
run('A/narrow-join/independent-target',dst,['test','-run','^TestReviewEarlyStopJoin$','-count=1','-timeout=10s','./...'])

for candidate,target in [('A','TestEarlyConsumerCompletionStopsAndJoinsProducer'),('B','TestConsumerEarlyCompletionCancelsAndJoinsProducer')]:
    baseline=OUT/'work'/f'{candidate}-ordinary'
    run(f'{candidate}/author-join/original-green',baseline,['test','-run','^'+target+'$','-count=10','-timeout=10s','./...'])
    mutant=OUT/'work'/f'{candidate}-mutant-skip-early-join'
    run(f'{candidate}/author-join/mutant-target',mutant,['test','-run','^'+target+'$','-count=10','-timeout=10s','./...'])
run('B/author-early-stop/original-green',OUT/'work/B-ordinary',['test','-run','^TestConsumerEarlyCompletionCancelsAndJoinsProducer$','-count=1','-timeout=5s','./...'])
run('B/author-early-stop/mutant-target',OUT/'work/B-mutant-no-early-stop',['test','-run','^TestConsumerEarlyCompletionCancelsAndJoinsProducer$','-count=1','-timeout=5s','./...'])
