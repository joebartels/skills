from pathlib import Path
from datetime import datetime, timezone
import hashlib
import json
import os
import subprocess

ROOT = Path('/private/tmp/go-context-outcome-20261002/library')
OUT = ROOT/'reviews'
checks = json.loads((OUT/'checks.json').read_text())
source_hashes = json.loads((OUT/'source-hashes.json').read_text())
assert all(hashlib.sha256((ROOT/p).read_bytes()).hexdigest()==h for p,h in source_hashes.items())

topics = [
    ('architecture','Architecture & Design','go-architecture-and-design'),
    ('code-quality','Code Quality & Go Idioms','go-code-quality-and-idioms'),
    ('correctness','Correctness & Compatibility','go-correctness-and-compatibility'),
    ('testing','Testing','go-testing'),
    ('security','Security','go-security'),
    ('resilience','Observability & Resilience','go-observability-and-resilience'),
    ('resources','Performance & Resource Management','go-performance-and-resource-management'),
    ('reproducibility','Dependencies & Reproducibility','go-dependencies-and-reproducibility'),
    ('deployment','Deployment & Operations','go-deployment-and-operations'),
]
applicability = {
    'architecture':'Applicable: exported callback/context API, cancellation/error boundary, and caller ownership are central to this library area.',
    'code-quality':'Applicable: error flow, helpers, names, exported documentation, and Go 1.22-compatible idioms in both source and tests.',
    'correctness':'Applicable: sequential ordering, exact accepted count, cancellation observation order, independent errors, legal non-comparable causes, and preserved signature.',
    'testing':'Applicable: author regression detection for the new cancellation and completion contract; assertions and scratch mutation sensitivity assessed.',
    'security':'Not applicable to this code area: ctx, integer jobs, and an executable cooperative callback are supplied by the same caller; Process introduces no authorization, privilege, parser, credential, network, filesystem, or attacker-controlled sensitive sink. Error retention is within that caller boundary. No security claim is made about unseen hosts or callbacks.',
    'resilience':'Applicable: cooperative stop behavior and failure/cancellation classification at the caller boundary. Service telemetry, retries, probes and process shutdown are outside this synchronous library contract.',
    'resources':'Applicable: sequential work bound, callback lifetime and caller-owned resources; no internal goroutine, timer, queue, I/O resource, or retained input graph.',
    'reproducibility':'Applicable: standalone standard-library-only Go module, newly used context.Cause/errors.Join APIs and explicit Go 1.22 support.',
    'deployment':'Not applicable to the bounded source implementation: no release, CI enforcement, packaging, runtime configuration or process lifecycle decision is introduced or part of the supplied target. Module build/support is assessed under Reproducibility. External release/platform facts are unknown and excluded, rather than assumed absent or passed.',
}

def common(n):
    return f'Code-area review of completed neutral candidate-{n}: process.go, process_test.go, README.md and go.mod; original/README.md is authoritative contract, original source/test are behavior context. Snapshot SHA-256 identities are in [manifest.json](manifest.json). Library module example.com/process, go 1.22; independent checks used Go 1.26.5 and Go 1.22.12 darwin/arm64.'

def card(n, topic, title, grade, coverage, rationale, good, bad=None, changes=None, limits=None):
    result = f'## {title} — {grade}\n\nScope: {common(n)}\n\nCoverage: {coverage}\n\nRationale: {rationale}\n\n'
    if grade not in ('Not applicable','Insufficient evidence'):
        major = 1 if topic=='testing' else 0
        moderate = 1 if topic=='testing' else 0
        result += f'Finding counts: critical=0, major={major}, moderate={moderate}, minor=0\n\nGood\n\n'
        result += '\n'.join('- '+x for x in good)+'\n\nBad\n\n'
        result += '\n'.join('- '+x for x in (bad or ['None found.']))+'\n\nSuggested changes\n\n'
        result += '\n'.join('- '+x for x in (changes or ['None needed.']))+'\n\n'
    result += 'Limits: '+(limits or 'Limited to the supplied code area and explicit nonnil/cooperative-input contract. Exact independent argv, cwd, environment overrides, status, stdout and stderr are preserved in [checks.json](checks.json). Supplied observations are separately retained in the packet provided-checks files. No candidate/original source was changed.')+'\n'
    result += f'\nSkill contract: [SKILL.md](../../review-skills/{dict((x[0],x[2]) for x in topics)[topic]}/SKILL.md); topic decision reference inspected where relevant.\n'
    (OUT/f'candidate-{n}'/f'{topic}.md').write_text(result)

for n in (1,2):
    dest = OUT/f'candidate-{n}'
    dest.mkdir(exist_ok=True)
    source = ROOT/f'candidate-{n}'/'process.go'
    ours = [r for r in checks['checks'] if r['label'].startswith(f'candidate-{n}/') or 'version' in r['label']]
    (dest/'checks.json').write_text(json.dumps({'scratch':checks['scratch'],'checks':ours},indent=2)+'\n')
    manifest = {
        'review_id':f'library/candidate-{n}', 'mode':'code_area', 'neutral_candidate':n,
        'scope':[f'candidate-{n}/'+p for p in ('process.go','process_test.go','README.md','go.mod')],
        'context':['original/README.md','original/process.go','original/process_test.go','original/go.mod','provided-contract-probes/contract_test.go',f'provided-checks-{n}.json'],
        'snapshot_sha256':{p:h for p,h in source_hashes.items() if p.startswith(f'candidate-{n}/') or p.startswith('original/')},
        'owner':'context_checked_library_review', 'worker_model':'inherited parent model; no overrides',
        'review_mode':'one bounded worker sequentially applied topic contracts to each candidate before any arm disclosure',
        'exclusions':['Repository history, task/planning/evaluation/writing-guidance evidence outside neutral packet','Unseen application/callback implementations and release/deployment configuration','Performance throughput or binary-identity claims not in contract'],
        'applicability':[{'id':key,'topic':title,'skill':f'review-skills/{skill}/SKILL.md','status':'not_applicable' if key in ('security','deployment') else 'complete','reason':applicability[key],'card':f'{key}.md'} for key,title,skill in topics],
        'boundaries':[
            {'id':'accepted-prefix','status':'complete','owner':'correctness','evidence':'Unchanged TestSequentialProgress, acceptance increment only after nil callback, supplied contract probes and independent runs.'},
            {'id':'cancellation-observation-and-completion','status':'complete','owner':'correctness/resilience','evidence':'Entry and pre-next-job checks; error snapshot before result return; no final-success post-check. Both canceled empty and final-success behavior independently exercised.'},
            {'id':'cause-and-independent-error','status':'complete','owner':'correctness/testing','evidence':'errors.Join without interface equality; supplied same-non-comparable-cause probe passes baseline; frozen unsafe equality mutation survives author suite and fails held probe.'},
            {'id':'caller-resource-and-api-ownership','status':'complete','owner':'architecture/resources','evidence':'Context forwarded directly to synchronous callback; no new owned resources, goroutines, process effects, dependencies or signature change.'},
            {'id':'minimum-version-and-standalone-build','status':'complete','owner':'reproducibility','evidence':'Independent Go 1.22.12 build/author/held pass; Go 1.26.5 cold standalone build with fresh caches, GOWORK=off, GOPROXY=off, GOTOOLCHAIN=local, -mod=readonly passes.'},
        ],
        'evidence_types':{'supplied':'provided-checks-N.json, provided-contract-probes/contract_test.go','executed':'checks.json; frozen substitutions independently replayed; two additional mutations labeled supplementary','inspected':'all candidate/original sources and README/module; packet-only unchanged review contracts'},
        'provided_checks_handling':'Provided runs are supplied observations, not claimed as independent execution. Frozen unsafe-cause-equality and lost-custom-cause substitutions were independently replayed. Entry-check removal and spurious-empty-callback insertion are supplementary reviewer mutation/assertion checks, not frozen efficacy targets.',
    }
    (dest/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')

    # Ordinary correct contract controls support A; no nonroutine A+ safeguard claim.
    card(n,'architecture','Architecture & Design','A',
        'Entire exported Process contract, callback injection, context/error propagation, ownership and helper fit; no package, transport or persistence boundaries exist inside this area.',
        'No actionable architectural issue. A direct synchronous function expresses the actual operation without additional interfaces, options, layers, or owned background lifetimes. This is appropriate routine design, supporting A.',
        [f'[G1] candidate-{n}/process.go:9 retains the exact original function signature and accepts the caller context/callback directly; each job runs synchronously, so ownership and completion stay visible.',
         f'[G2] candidate-{n}/process.go:{20 if n==1 else 19} exposes the independent callback error plus the documented cancellation classification/cause using standard Go error traversal; independent supplied-contract runs pass.'],
        limits='No consumer implementation beyond original tests and provided probes is supplied. That does not prevent checking this fully specified callback library boundary. External hosting, release and callback behavior are unknown/outside scope. Exact checks are in [checks.json](checks.json). No architecture redesign is proposed.')

    card(n,'code-quality','Code Quality & Go Idioms','A',
        'All production/test Go source, helper contracts, error/value semantics, names, documentation via exported comment plus README, supported APIs and mechanical diagnostics.',
        'No actionable local idiom or maintainability defect. Error branches stay local and the accepted-count success path is straightforward; Join is appropriate for the required multi-cause inspection. Helpers add little complexity. Candidate variations in helper guarding/shadowing are optional preferences, not findings.',
        [f'[G1] candidate-{n}/process.go:{18 if n==1 else 17}-{26 if n==1 else 25} increments accepted only after a nil callback and returns errors immediately without hiding the independent cause.',
         '[G2] Independent gofmt -d . emits no diff; go vet ./... succeeds. context.Cause and errors.Join compile on the independently tested minimum Go 1.22.12.'],
        limits='No mutating formatter or fixer was used. The unchanged brief exported comment is supplemented by the complete supplied README contract; expanding it is optional. Test assertion strategy belongs to Testing and is not double-counted here. Exact diagnostics are in [checks.json](checks.json).')

    card(n,'correctness','Correctness & Compatibility','A',
        'Normal sequential progress, first failure after accepted prefix, active empty, already-canceled empty/nonempty, between-job stop, coincident accepted final completion, cooperative blocked callback, independent callback error with custom cause, same non-comparable callback/cancellation cause, signature and minimum-version preservation.',
        'No production contract defect confirmed. Entry/pre-job checks stop unstarted work, post-error observation retains each promised cause, and accepted final success returns without a later cancellation check. No interface equality is used. Ordinary correct implementation and focused tests justify A; no claim that the author suite detects every regression.',
        [f'[G1] candidate-{n}/process.go:10-{24 if n==1 else 23} preserves exact accepted prefix and stops before another callback; unchanged TestSequentialProgress and provided probes pass independently.',
         f'[G2] candidate-{n}/process.go:{29 if n==1 else 28} joins ctx.Err and context.Cause without comparing legal dynamic error values; independent TestCauseReturnedByCallback passes and both errors.Is/As contracts are retained.',
         '[G3] Independent baseline build, author tests and provided contract suite pass on Go 1.26.5 and Go 1.22.12. Provided-contract race/shuffle/count=3 run also passes for the exercised cancellation interleaving.'],
        limits='The API explicitly requires nonnil ctx and callback plus cooperative callback return; nil arguments or forced preemption are not required. No supported platform matrix beyond version is supplied; host runs are darwin/arm64. Race checks cover only executed paths. Two mutation failures show test gaps, not defects in the unmodified implementations. Full logs: [checks.json](checks.json).')

    test_location = f'candidate-{n}/process_test.go:{70 if n==1 else 73}'
    bad1 = f'[T1][major][existing-in-scope] {test_location}: the callback returns errors.New while the cancellation cause has a different non-comparable dynamic type. Go interface equality can compare differing dynamic types without inspecting the non-comparable value. The frozen unsafe-cause-equality mutant therefore compiles and passes the full author suite, but independent held TestCauseReturnedByCallback fails with runtime error: comparing uncomparable type process.listCause. The README explicitly makes arbitrary legal causes and panic-free equality an important boundary; its panic regression remains effectively unverified. Reach is this function contract, so contained major rather than systemic/critical.'
    fix1 = '[T1] Add an author regression where the callback cancels with a slice-containing legal error and returns that same cause (or another value of its same dynamic type). Assert zero/accepted-prefix result, standard cancellation classification and errors.As payload. Re-run the unchanged frozen unsafe equality mutation: it must compile and fail the author suite while baseline passes.'
    if n==1:
        bad2 = '[T2][moderate][existing-in-scope] candidate-1/process_test.go:29 tests already-canceled input only with a job; its active-empty case uses a live context. Removing only the entry cancellation check leaves the per-job guard intact, compiles and passes all author tests, yet held TestCanceledBeforeStart fails on nil jobs with err=<nil>. This is a localized normal-use boundary regression in the explicit already-canceled-empty behavior; correcting the missing empty-canceled case is independent of the equality panic test.'
        fix2 = '[T2] Exercise an already-canceled context with both empty and nonempty jobs, asserting zero callbacks/accepted plus cancellation classification and inspectable custom cause. The supplementary entry-check-removal mutation must fail the author suite.'
    else:
        bad2 = '[T2][moderate][existing-in-scope] candidate-2/process_test.go:50-54 uses a nil-returning callback for active empty input and asserts only count/error. A supplementary mutant that invokes apply(ctx, 0) for live empty jobs while ignoring its nil result compiles and passes all author tests, but held TestResultDecisionOrder fails with empty input invoked callback. The callback can have caller-owned effects, so this normal-use boundary needs a meaningful no-invocation assertion. Correcting that assertion is independent of the equality panic case.'
        fix2 = '[T2] Make the active-empty callback fail the test or count invocations and assert zero calls, alongside zero accepted and nil error. The supplementary spurious-empty-callback mutation must fail the author suite.'
    card(n,'testing','Testing','C',
        'All author tests/assertions and original regression, provided probes, baseline/race/Go-version runs, both unchanged frozen mutations, plus matched supplementary entry-check removal and spurious-empty-callback insertion in separate scratch copies.',
        'One contained major gap in the explicitly required legal non-comparable-cause panic boundary and one independent moderate empty-input assertion/coverage gap. The shared rubric selects C (one major with at most one moderate). Tests meaningfully check most new behavior; passing count/coverage does not override demonstrated mutation survivors.',
        ['[G1] Preserved TestSequentialProgress asserts accepted prefix, callback order and retained independent failure; both baseline author suites pass.',
         '[G2] The frozen lost-custom-cause mutant compiles and fails both full author suites: errors.Is/As assertions detect omitted custom causes, independently reproducing the supplied observation.',
         '[G3] Final-success coincident cancellation and between-job cancellation assertions detect distinct expected outcomes without sleeps or test goroutines. Provided blocked-callback probe additionally synchronizes the exercised cancellation interleaving and passes race/shuffled runs.'],
        [bad1,bad2],[fix1,fix2],
        limits='Production baseline meets these paths; findings assess author regression detection only. The provided held probes are separate reviewer evidence, not tests delivered in candidate source. Frozen efficacy facts are unsafe-cause-equality survivor and lost-custom-cause kill; the two empty-input mutations are explicitly supplementary after inspecting the packet and must not enter frozen efficacy-benefit counts. No statement-coverage percentage, fuzzing, benchmark, CI enforcement or exhaustive schedule claim is made. Exact commands/output and scratch locations: [checks.json](checks.json).')

    card(n,'security','Security','Not applicable',applicability['security'],
        'Conditional applicability examined against every source import/operation and the stated caller trust/ownership contract; no security control or trust crossing is implicated within this bounded code area.',[],
        limits='Unseen hosting applications, callback sinks, credential policy, and deployment toolchain vulnerability state are unknown. No vulnerability scanner was run and no clean-security claim is made. Missing external context is not used as evidence of safety or as a grade-lowering defect in this source-only target.')

    card(n,'resilience','Observability & Resilience','A',
        'Cooperative cancellation, between-job admission, failure/cancellation diagnostic cause exposure and completion order in the entire synchronous library area. Telemetry, retries, queues, distributed budgets and services are outside this contract.',
        'No actionable resilience issue. Cancellation requests stopping and the callback owns cooperation, exactly as documented. The function neither hides independent failure nor invalidates accepted final success. Useful returned errors are the appropriate owned signal for this small caller-controlled library.',
        [f'[G1] candidate-{n}/process.go:10-{21 if n==1 else 20} stops new callbacks on cancellation and preserves classification/custom cause with independent failure; independently executed TestCustomCauseAndIndependentFailure holds/releases the callback and succeeds.',
         f'[G2] candidate-{n}/process.go:{24 if n==1 else 23}-{26 if n==1 else 25} completes accepted final work without a later cancellation reclassification; TestResultDecisionOrder and the author final-success case pass.'],
        limits='No hard callback deadline or preemption is promised; a non-cooperative callback is a caller-owned contract violation, not a missing internal goroutine/timeout. No runtime telemetry/SLO or deployed service is supplied; none is assumed. Duplicate standard cancellation text when classification equals cause is unpromised formatting and optional to refine safely. Exact checks: [checks.json](checks.json).')

    card(n,'resources','Performance & Resource Management','A',
        'Source-derived work/memory bounds and resource ownership on normal, error and canceled exits; all Process/helper statements inspected.',
        'No actionable resource defect. Process visits each started job once in order with constant bookkeeping; cancellation/failure exits allocate only a bounded error tree. It adds no workers, internal I/O, timers, caches, copied job list or retained state. These observable bounds support A without asserting a measured speed benefit.',
        [f'[G1] candidate-{n}/process.go:{14 if n==1 else 13}-{25 if n==1 else 24} performs a single sequential loop: O(started jobs) local bookkeeping, one caller callback at a time, constant extra state apart from the bounded returned errors.',
         '[G2] Context/callback resources remain caller-owned; every source return exits synchronously with no package-owned file/socket/timer/goroutine to close or drain. The cooperative blocked-callback probe confirms Process waits for callback return rather than creating hidden parallel work.'],
        limits='Callback cost is arbitrary/caller-owned; no frequency, latency, memory budget or production workload is supplied. No benchmark/profile or relative speed/allocation claim is made. Rechecking ctx.Err or initial cancellation is a bounded contract control, not a substantiated performance regression. External workload costs are unknown, outside this bounded local ownership/complexity assessment.')

    card(n,'reproducibility','Dependencies & Reproducibility','A',
        'Entire module/import graph, unchanged go 1.22 directive, no requires/replaces/workspace/generation/cgo inputs, independent fresh-cache standalone build, minimum/current toolchain build/tests.',
        'No actionable resolution/support issue. go list -m all lists only example.com/process; added production import errors is standard library. Independent minimum Go 1.22.12 compilation and behavior checks pass. Cold standalone build with network/toolchain switching disabled demonstrates no hidden module/workspace input is needed for the local library build.',
        ['[G1] candidate-'+str(n)+'/go.mod:1-3 is unchanged, with no external module requirement; independent go list -m all outputs only example.com/process. No go.sum is needed for this standard-library-only module.',
         '[G2] Independent go build -mod=readonly ./... succeeds with candidate-specific fresh GOCACHE/GOMODCACHE, GOWORK=off, GOTOOLCHAIN=local and GOPROXY=off on Go 1.26.5; independent Go 1.22.12 build and author/provided contract tests all succeed.'],
        limits='Both actual toolchains were version-checked. Host target is darwin/arm64; no promised platform matrix or bit-identical artifact policy is supplied. Offline standalone resolution/build success is checked, not binary reproducibility or release-pipeline enforcement. Exact runtime binary paths/environment/outputs are in [checks.json](checks.json).')

    card(n,'deployment','Deployment & Operations','Not applicable',applicability['deployment'],
        'Conditional applicability examined: this target supplies a reusable callback function with no release artifact promotion, configured execution environment, CI gate or process-owned runtime lifecycle decision. Local version/build obligations are covered in Reproducibility, without inventing infrastructure requirements.',[],
        limits='External release workflow, tags, artifacts, hosting runtime and enforcement are unavailable/unknown and outside the requested source implementation. If those paths enter scope, this topic becomes applicable and requires their evidence; this N/A state does not claim they passed or are absent.')

    findings = [
        {'id':'T1','severity':'major','systemic':False,'cause':'Author error/cause regression uses differing dynamic types and misses legal same-non-comparable cause comparison panic.','evidence':f'{test_location}; independent unsafe-cause-equality compile=0 author=0 held=1; baseline provided contract=0.','correction':fix1,'sources':[f'candidate-{n}/testing/T1'],'primary_owner':'Testing','status':'confirmed'},
        {'id':'T2','severity':'moderate','systemic':False,'cause':'Canceled-empty result lacks author coverage.' if n==1 else 'Active-empty callback non-invocation lacks author assertion.',
         'evidence':'candidate-1/process_test.go:29,43; supplementary entry-check removal compile=0 author=0 held=1.' if n==1 else 'candidate-2/process_test.go:50-54; supplementary spurious empty callback compile=0 author=0 held=1.',
         'correction':fix2,'sources':[f'candidate-{n}/testing/T2'],'primary_owner':'Testing','status':'confirmed','efficacy_status':'supplementary, not frozen efficacy target'},
    ]
    ledger = {'scope':common(n),'coverage':manifest['applicability'], 'findings':findings,
              'strengths':[{'id':'G1','evidence':'Unmodified candidates preserve sequential accepted-prefix behavior and independently pass author/held suites on Go 1.22.12 and Go 1.26.5.'},
                           {'id':'G2','evidence':'Independent frozen lost-custom-cause mutation fails author suite with meaningful errors.Is/As assertions.'}], 'safeguards':[]}
    (dest/'ledger.json').write_text(json.dumps(ledger,indent=2)+'\n')
    (dest/'findings.json').write_text(json.dumps({'canonical_findings':findings,'reconciliation':'Two independently necessary test corrections; no production defect claimed or cross-topic duplicated count. Supplementary mutations excluded from frozen efficacy.','unconfirmed_or_optional':['Deduplicating cancellation error text, expanding exported comment, consolidating helper guards are optional preferences with no promised-behavior failure.','No demonstrated attacker, release, measured performance, or non-cooperative callback requirement in source-only scope.']},indent=2)+'\n')
    argv=['rtk','proxy','python3',str(ROOT/'review-skills/go-quality-report/scripts/grade.py'),str(dest/'ledger.json')]
    p=subprocess.run(argv,cwd=ROOT,text=True,capture_output=True,check=True)
    result=json.loads(p.stdout)
    assert result['overall']=='C'
    (dest/'grade-result.json').write_text(p.stdout)
    (dest/'grade-command.json').write_text(json.dumps({'argv':argv,'cwd':str(ROOT),'status':p.returncode,'stdout':p.stdout,'stderr':p.stderr},indent=2)+'\n')
    table='\n'.join(f'| {title} | {"C" if key=="testing" else "Not applicable" if key in ("security","deployment") else "A"} | {applicability[key]} | [{key}]({key}.md) |' for key,title,skill in topics)
    report=f'''# Go Quality Report — C

Scope: {common(n)}

Coverage: Seven applicable topics complete; conditional Security and Deployment have bounded, explicit not-applicable reasons. All nine considered. Code-area facts only; unseen host/release paths excluded rather than inferred passed.

Rationale: One unique contained major test-signal defect plus one independent moderate empty-input test gap select C under the shared rubric. Source behavior meets the inspected contract. This is a deduplicated defect grade, not an average of topic grades. The unchanged absolute-path grade.py result is preserved in [grade-result.json](grade-result.json).

Unique finding counts: critical=0, major=1, moderate=1, minor=0.

## Report cards

| Topic | Grade/state | Assessed coverage and limits | Card |
| --- | --- | --- | --- |
{table}

## Good

- [G1] Both minimum/current toolchain author and provided-contract suites pass independently, preserving exact accepted prefix, failure cause and final-success cancellation order.
- [G2] Local synchronous design keeps callback/context/resource ownership with the caller, adds no concurrency or external dependencies, and uses panic-free error joining.
- [G3] Frozen loss-of-custom-cause mutation is killed by meaningful author errors.Is/As assertions.

## Bad

- [T1][major] Legal same-dynamic-type non-comparable callback/cancellation cause is absent from author tests. Frozen unsafe equality mutant compiles and passes author suite; held probe fails with comparison panic. See [Testing card](testing.md).
- [T2][moderate] {findings[1]['cause']} Supplementary scratch mutant demonstrates the lost assertion/coverage; see [Testing card](testing.md).

## Suggested changes

- [T1][P1] {fix1}
- [T2][P2] {fix2}

## Limits

Review remained within neutral packet sources/contracts and unchanged review skills; no arm identity or writing guidance was used. One bounded worker assessed both candidates sequentially; this is independent of authorship, not a second independently delegated review of each card. Sources were preserved and all mutations/checks ran in separate scratch copies. Current/minimum toolchains are verified Go 1.26.5/1.22.12 on darwin/arm64. Unseen deployment/callback policies, other target platforms, exhaustive race schedules, workload throughput and binary identity remain unknown/outside scope. No clean security/scanner result is claimed.

Supplied checks are observations; independently repeated baseline/mutation commands are distinguished in [checks.json](checks.json). Entry-check removal and spurious empty callback are supplementary reviewer checks after packet inspection, not original frozen efficacy targets. No new efficacy benefit is inferred from them.

Details: [manifest](manifest.json), [findings and reconciliation](findings.json), [ledger](ledger.json), [exact executed checks](checks.json), [calculator invocation](grade-command.json). The frozen review artifact hashes are in [FROZEN.json](../FROZEN.json).
'''
    (dest/'report.md').write_text(report)

(OUT/'review-contract-hashes.json').write_text(json.dumps({str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((ROOT/'review-skills').rglob('*')) if p.is_file()},indent=2)+'\n')
(OUT/'summary.md').write_text('''# Frozen neutral library review

Candidate-1: overall C; Architecture A, Code Quality A, Correctness A, Testing C, Resilience A, Resources A, Reproducibility A; Security/Deployment not applicable within bounded source area. Testing has one major same-dynamic-type non-comparable cause panic regression gap and one moderate canceled-empty coverage gap.

Candidate-2: overall C; same topic grades/states. Testing has the same major panic regression gap and one moderate active-empty callback non-invocation assertion gap.

Both unmodified implementations independently pass Go 1.22.12/1.26.5 builds, author/provided contract suites, vet/format checks and current-toolchain exercised race/shuffled checks. Both author suites miss the frozen unsafe equality mutation and kill the frozen lost-custom-cause mutation. Supplementary empty-input mutations are independently distinguished and excluded from frozen efficacy.

Full cards, all-nine coverage and grade ledgers: [candidate-1/report.md](candidate-1/report.md), [candidate-2/report.md](candidate-2/report.md). Exact commands/output: [checks.json](checks.json). No candidate source was modified; no arm identity or external planning/writing evidence was inspected.
''')

assert all(hashlib.sha256((ROOT/p).read_bytes()).hexdigest()==h for p,h in source_hashes.items())
frozen = {'frozen_at_utc':datetime.now(timezone.utc).isoformat(), 'arm_disclosure_received':False,
          'sources_unchanged':True, 'source_snapshots':source_hashes,
          'artifact_sha256':{str(p.relative_to(OUT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(OUT.rglob('*')) if p.is_file() and p.name!='FROZEN.json'},
          'scope':'neutral library completed code areas; packet-only review; no writing-guidance review',
          'grade_contract':'unchanged review-skills/go-quality-report/scripts/grade.py used by absolute path'}
(OUT/'FROZEN.json').write_text(json.dumps(frozen,indent=2)+'\n')
print('Frozen:',OUT/'FROZEN.json')
print('Reports:',OUT/'candidate-1/report.md',OUT/'candidate-2/report.md')
