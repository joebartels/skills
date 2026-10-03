#!/usr/bin/env python3
import json
import shutil
from run_checks import OUT, GO122, run

for c in ['A', 'B']:
    work = OUT / 'work' / ('drop-unit-updates-' + c)
    run(c + '-mutation-drop-unit-updates-full-suite-go122', work,
        [GO122, 'test', '-mod=readonly', '-count=1', '-timeout=2s', './...'])
    shutil.copy2(OUT / 'contract_probe_test.go', work / 'review_contract_test.go')
    run(c + '-mutation-drop-unit-updates-event-probe-go122', work,
        [GO122, 'test', '-mod=readonly', '-count=1', '-timeout=5s', '-run', '^TestReviewConcurrentCheckpoint$', './...'])
