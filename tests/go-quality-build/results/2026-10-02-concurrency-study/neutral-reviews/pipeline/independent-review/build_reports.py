import collections
import hashlib
import json
from pathlib import Path
import subprocess

ROOT=Path('/private/tmp/go-neutral-pipeline-study')
OUT=ROOT/'independent-review'
GRADE=ROOT/'review-guidance/go-quality-report/scripts/grade.py'
COMMANDS=json.loads((OUT/'raw/independent-commands.json').read_text())
FOLLOWUP=json.loads((OUT/'raw/followup-commands.json').read_text())
TOPICS=[
 ('architecture-and-design','Architecture & Design','Callback/context/error ownership and cross-boundary completion are the central design decisions; the fixed private API and supplied host must remain usable.'),
 ('code-quality-and-idioms','Code Quality & Go Idioms','The local context state and error classification need clear, truthful semantics; Go 1.22, formatting and vet also apply.'),
 ('correctness-and-compatibility','Correctness & Compatibility','FIFO, early normal completion, failure cancellation, callback joining, error identity and CLI output/status are explicit important contracts.'),
 ('testing','Testing','The delivered tests claim to verify concurrency and error preservation. Their assertions, event gates, isolation and mutation sensitivity require assessment.'),
 ('security','Security','The CLI parses caller-controlled integer-line input and emits only integers/diagnostics; resource amplification and input-to-sink paths can be assessed within this local CLI boundary.'),
 ('observability-and-resilience','Observability & Resilience','Cooperative stop, callback failure propagation and caller-visible success/error signals govern failure containment and diagnostics.'),
 ('performance-and-resource-management','Performance & Resource Management','The code owns a bounded channel and two callbacks; blocked sends/receives, cleanup joins and constant per-call work are relevant resources.'),
 ('dependencies-and-reproducibility','Dependencies & Reproducibility','The documented Go 1.22 minimum and standard-library-only standalone module are a rebuild contract, checked with an actual minimum toolchain.'),
 ('deployment-and-operations','Deployment & Operations','This artifact is a CLI. Actual binary build, startup/configuration, stdout/stderr, exit codes and finite-input termination are relevant; containers, signals and rollouts are outside this scope.'),
]

def save(name,obj):
 (OUT/name).write_text(json.dumps(obj,indent=2)+'\n')

FINDINGS={
 'A':[
  dict(id='A/F1',severity='major',systemic=False,primary_owner='correctness-and-compatibility',scope='existing-in-scope',
   cause='The stop snapshot stores only a boolean; suppression then treats both Canceled and DeadlineExceeded as the coordinated stop error without matching the actual child context error.',
   evidence='candidates/A/cmd/pipeline/pipeline.go:13,19,24,57-59; TestReviewIndependentDeadlineAfterStop has an active parent, observes child ctx.Err()==context.Canceled, then independently returns context.DeadlineExceeded during cleanup. Host Go1.26.5, actual Go1.22.12 and host race/shuffle all return nil and fail the intended error-identity assertion.',
   correction='Record the actual child context error observed with each callback result and suppress only the exact matching context stop error for that coordinated stop. Preserve a DeadlineExceeded error when the work context was only Canceled, plus wrapped/joined independent causes. Keep the first independent exact error before stopping.',
   magnitude='Contained error-preservation failure in one private pipeline contract. A caller may receive success instead of an independent operation deadline failure during callback cleanup; no broad/irreversible or systemic reach was shown.',
   topics=['architecture-and-design','code-quality-and-idioms','correctness-and-compatibility','observability-and-resilience'],
   evidence_labels=['A/host/independent-probes','A/minimum/independent-probes','A/probe-race-shuffle'],
   locations=['candidates/A/cmd/pipeline/pipeline.go:57','probes/review_probe_test.go:124']),
  dict(id='A/F2',severity='major',systemic=False,primary_owner='testing',scope='existing-in-scope',
   cause='The claimed early producer join assertion tests a flag assigned true after the pipeline call and waits for eventual producer cleanup after return; it never establishes cleanup completed before host return.',
   evidence='candidates/A/cmd/pipeline/pipeline_test.go:44-69. A narrowly scoped, compiling mutation returns after normal consumer completion only while producerDone is false. Baseline author/probe targets pass; the mutant full author suite and claimed join test both pass at -count=10, while TestReviewEarlyStopJoin fails ASSERT host returned before producer cleanup release.',
   correction='Run the host asynchronously, hold producer cleanup after observing its stop, and assert host completion is still unavailable before releasing cleanup; then check both callback completion and host result. Own cancellation and gate release in cleanup and bound every event wait.',
   magnitude='An explicit important lifecycle contract is effectively unverified by the delivered join test. The regression can return to callers while callback cleanup still runs; one test-harness correction is required independently of production fixes.',
   topics=['testing'],evidence_labels=['A/narrow-join/original-author-target-green','A/narrow-join/original-probe-target-green','A/narrow-join/compile','A/narrow-join/ordinary-suite','A/narrow-join/author-target','A/narrow-join/independent-target'],
   locations=['candidates/A/cmd/pipeline/pipeline_test.go:46','candidates/A/cmd/pipeline/pipeline_test.go:62','candidates/A/cmd/pipeline/pipeline_test.go:67']),
  dict(id='A/F3',severity='major',systemic=False,primary_owner='testing',scope='existing-in-scope',
   cause='Independent cancellation-class error tests cover a pre-stop producer joined error and first consumer sentinel, but do not verify independent error preservation after stopping for both callback roles/classes.',
   evidence='candidates/A/cmd/pipeline/pipeline_test.go:88-109. The compiling broad-cancellation-suppression mutation survives the entire delivered suite. Baseline TestReviewWrappedFailures is green, and the mutant fails its consumer-cleanup independent-error assertion. The unmodified candidate also fails the independent DeadlineExceeded-during-Canceled-cleanup probe on both toolchains.',
   correction='Add independent producer/consumer error assertions before and after coordinated stopping, including an exact error different from child ctx.Err() and wrapped/joined errors. Require original-green, compiled-mutant and intended assertion failures for suppression mutations.',
   magnitude='The README explicitly requires preservation of independent failures. A regression can silently discard consumer cleanup errors without changing this suite result; a separate error-test correction is needed from the host/cleanup join assertion correction.',
   topics=['testing'],evidence_labels=['A/mutation/broad-cancel-suppression/original-target-green','A/mutation/broad-cancel-suppression/compile','A/mutation/broad-cancel-suppression/ordinary-suite','A/mutation/broad-cancel-suppression/independent-target','A/host/independent-probes'],
   locations=['candidates/A/cmd/pipeline/pipeline_test.go:88','candidates/A/cmd/pipeline/pipeline_test.go:99']),
 ],
 'B':[
  dict(id='B/F1',severity='major',systemic=False,primary_owner='correctness-and-compatibility',scope='existing-in-scope',
   cause='suppressStopError uses only an active parent and membership in the two exact cancellation sentinels; it has no evidence that a callback error was caused by coordinated stopping.',
   evidence='candidates/B/cmd/pipeline/pipeline.go:43,46-49. TestControllerIndependentCancellationClassBeforeStop fails on host, actual Go1.22.12 and race/shuffle. Independent TestReviewIndependentExactBeforeStop reproduces both producer and consumer first exact-Canceled errors being lost before child cancellation. TestReviewIndependentDeadlineAfterStop also loses DeadlineExceeded when the actual child stop is Canceled.',
   correction='Attach actual child stop/error state to each callback result, retain the first independent failure observed before stopping, and suppress only the exact matching context error generated for coordinated stopping. Preserve other exact cancellation-class errors and every wrapped/joined independent error.',
   magnitude='Contained important error-preservation failure in the private pipeline. Both callback roles can report an independent cancellation failure and callers receive success; multiple symptoms share the single classification correction, so this is one major rather than a systemic major.',
   topics=['architecture-and-design','code-quality-and-idioms','correctness-and-compatibility','observability-and-resilience'],
   evidence_labels=['B/host/held-tests','B/minimum/held-tests','B/held-race-shuffle','B/host/independent-probes','B/minimum/independent-probes','B/probe-race-shuffle'],
   locations=['candidates/B/cmd/pipeline/pipeline.go:43','candidates/B/cmd/pipeline/pipeline.go:47']),
  dict(id='B/F2',severity='major',systemic=False,primary_owner='testing',scope='existing-in-scope',
   cause='The suite conflates cancellation-class independent failures with wrapped errors: it never asserts preservation of a bare sentinel returned before any coordinated stop, or an independent exact deadline after a Canceled stop.',
   evidence='candidates/B/cmd/pipeline/pipeline_test.go:93-131; the candidate ordinary suite is green on host and actual Go1.22.12, while the supplied held test and independent exact-error probes deterministically fail. The both-error test meaningfully rejects broad errors.Is-based suppression, but that assertion does not cover this current provenance defect.',
   correction='Add bounded tests for each callback returning bare context.Canceled while the child and parent are active, and for an independent context.DeadlineExceeded while cleanup observes Canceled. Assert error identity after joins, and retain the existing wrapped/joined cases.',
   magnitude='An explicit important error-preservation contract is unverified for the exact sentinel paths that currently fail. Production repair and test additions are independent changes; the two timing/class examples belong to one assertion-matrix gap.',
   topics=['testing'],evidence_labels=['B/host/ordinary-tests','B/minimum/ordinary-tests','B/host/held-tests','B/host/independent-probes','B/mutation/broad-cancel-suppression/ordinary-suite'],
   locations=['candidates/B/cmd/pipeline/pipeline_test.go:93','candidates/B/cmd/pipeline/pipeline_test.go:110']),
 ]
}

STRENGTHS={
 'A':[
 dict(id='A/G1',evidence='TestReviewFIFO succeeds at capacities 1 and 3, and actual host/minimum binaries emit the finite integer sequence and exact --take prefix in order.'),
 dict(id='A/G2',evidence='Supplied held early cleanup gate and independent TestReviewEarlyStopJoin/TestReviewFailureStopJoin succeed on the original A snapshot. Both callbacks are joined after cooperative stopping, and the original first-sentinel-before-stop probe succeeds.'),
 dict(id='A/G3',evidence='Exact equality retains wrapped/joined independent errors on the original snapshot (TestReviewWrappedFailures passes). errors.Join preserves inspected independent error identities; caller cancellation remains an error in the delivered test.'),
 dict(id='A/G4',evidence='Host1.26.5 and actual Go1.22.12 builds/vet/ordinary suites pass in disposable standalone copies with GOWORK=off, GOTOOLCHAIN=local, GOPROXY=off and -mod=readonly; gofmt -d is empty.'),
 ],
 'B':[
 dict(id='B/G1',evidence='TestReviewFIFO succeeds at capacities 1 and 3, and actual host/minimum binaries emit the finite integer sequence and exact --take prefix in order.'),
 dict(id='B/G2',evidence='Independent early-stop and producer-failure cleanup-gate probes succeed on original B. Its author join test baseline is green, and the compiling skip-early-join mutation fails the intended joined-before-return assertion repeatedly.'),
 dict(id='B/G3',evidence='The original wrapped/joined independent-error probe succeeds. Its author both-error test rejects the compiling broad errors.Is suppression mutation through the intended missing-independent-failures assertion.'),
 dict(id='B/G4',evidence='Host1.26.5 and actual Go1.22.12 builds/vet/ordinary suites pass in disposable standalone copies with GOWORK=off, GOTOOLCHAIN=local, GOPROXY=off and -mod=readonly; gofmt -d is empty.'),
 ]
}

GOOD={
 'architecture-and-design':'The existing single-package private callback API stays intact. The pipeline owns its child context, buffered result slots and channel close, while command owns parser/output/status translation; bounded stop/join probes verify callback ownership on ordinary stop and failure paths.',
 'code-quality-and-idioms':'The code uses supported context/errors primitives, direction-restricted channel parameters and exact equality to avoid stripping wrapped causes. gofmt -d is empty and both actual toolchains pass vet; this does not validate the faulty suppression predicate.',
 'correctness-and-compatibility':'Independent FIFO probes at capacity 1/3, held early/failure cancellation checks, accepted-output writer-failure probe and actual host/minimum executable output/status cases pass. Public-facing parser, consumer, host and private signature are preserved.',
 'testing':'Finite/FIFO expectations use real values and error identity checks. The independent four-mutation probes each have an original-green target, compiling mutant and intended bounded assertion failure. These review probes assess current behavior but are not credited as candidate-authored regression coverage.',
 'security':'Production input is scanned incrementally and parsed as an integer before entering the bounded channel; output uses a fixed integer format. No input-controlled executable/path/network/authorization/secret boundary is present. Source has two fixed callback goroutines and a capacity-one command queue; there is no per-line goroutine growth.',
 'observability-and-resilience':'Cooperative cancellation reaches the context-aware callbacks and the held cleanup gates confirm normal early stopping and failed-producer containment. Actual malformed-first and writer-failure cases preserve stderr diagnostics and the checked accepted output.',
 'performance-and-resource-management':'Per pipeline call, the implementation creates one capacity-bounded integer channel and a two-result channel, with two callback goroutines. This is O(capacity) queued data and O(1) coordinator state/goroutine count. Independent gated probes confirm blocked cooperative work is stopped and callback cleanup joined.',
 'dependencies-and-reproducibility':'The standard-library-only standalone module has no external module requirements, replacement, workspace, vendor or generation inputs. Both host and the actual Go1.22.12 compiler build and test readonly offline disposable copies with GOWORK=off; source hashes remain unchanged.',
 'deployment-and-operations':'The actual independently built host/minimum binaries terminate within the five-second bound for finite input, early --take=2, malformed-first, empty input and invalid --take=0. Expected stdout and 0/2 exit status match, success stderr stays empty and malformed-first stderr contains integer input.',
}

LIMITS='Review is confined to the supplied bounded CLI/pipeline snapshots, original contract, candidate tests and supplied held checks. Actual tested platform is darwin/arm64; no cross-platform packaging, service deployment, signals, containers, rollout, remote telemetry or vulnerability-advisory audit is claimed. Arbitrary blocking Reader cancellation is explicitly excluded by README. Race runs cover the exercised paths only. Original and review-mutant checks are distinguished in raw logs. Package-alarm timeouts are not counted as mutation detection. Review is one independent reviewer performing bounded sequential reviews, without author reports/profiles or external repository plans/results.'

manifest=dict(review_id='neutral-pipeline-independent',mode='bounded-code-area',date='2026-10-02',target='Candidates A and B separately against original README/source',paths=['README.md','go.mod','cmd/pipeline/main.go','cmd/pipeline/pipeline.go','cmd/pipeline/pipeline_test.go'],baseline='original/',source_manifest='source-manifest.json',source_hashes=['source-hashes-before.json','source-hashes-after.json'],read_only_originals=True,owner='independent reviewer',excluded=['Repository plans/results and author reports','Other reviewers, profiles, exposure identity','Writing guidance/promotion','Arbitrary blocking Reader cancellation','Service/process lifecycle machinery'],boundary_obligations=['FIFO and accepted effects','normal producer drain','early consumer stop plus held producer cleanup join','producer failure stop plus held consumer cleanup join','caller cancellation','independent exact/wrapped/joined failures before/after coordinated stop','actual executable stdout/stderr/status','Go1.22 actual minimum build'],review_guidance=str(ROOT/'review-guidance'),delegation='None',applicability={},status='complete')
save('findings.json',dict(candidates=FINDINGS,strengths=STRENGTHS,reconciliation=['Within each candidate, the production cancellation-class suppression symptoms are deduplicated to one root cause. Cross-topic copies use that canonical ID.','Production correction and error-regression assertion additions are separate independently necessary corrections.','A join assertion needs a distinct event/order correction and remains separate from the error assertion matrix gap.','No systemic major or critical reach is supported. Multiple callback examples are contained in one pipeline boundary.','The first broad A skip-join mutation also erased already-observed producer errors, so only its targeted result is retained as supplementary evidence. The narrowed mutation is used for the independently graded join-test gap.'],rejected_or_unconfirmed=[dict(id='Q1',candidate='A',concern='Publication of producer result precedes close(out).',decision='Not graded: callbacks have returned, and source promises callback joining; no demonstrated persistent goroutine/channel leak or supported close-before-host-return contract was established.'),dict(id='Q2',candidate='both',concern='Two exact context.Canceled errors can be causally indistinguishable after a stop.',decision='Limit: findings rely on observable before-stop provenance and mismatched DeadlineExceeded-vs-Canceled classes, not an impossible inference about identical sentinel intent after stopping.'),dict(id='Q3',candidate='both',concern='Arbitrary blocking Scanner Reader cancellation.',decision='Excluded explicitly by original README; no defect or requested signal machinery.'),dict(id='Q4',candidate='both',concern='Package-alarm mutation failures and missing cleanup on failed author tests.',decision='Raw results retained; timeout-only runs provide no intended-assertion mutation credit. This is a verification limit and optional harness-hardening follow-up, not extra counted production defects.')]))

grades={}
for candidate in ['A','B']:
    applicable=[]
    coverage=[]
    card_paths={}
    (OUT/'cards'/candidate).mkdir(parents=True,exist_ok=True)
    scope=f'Bounded code-area review of candidate {candidate} source snapshot (see source-hashes-before.json), CLI/private runPipeline and delivered tests, judged against original README/source; Go1.22 minimum, actual Go1.22.12 and host Go1.26.5 darwin/arm64. Findings are existing-in-scope in this delivered snapshot.'
    grades[candidate]={}
    for topic,title,reason in TOPICS:
        findings=[f for f in FINDINGS[candidate] if topic in f['topics']]
        card_rel=f'cards/{candidate}/{topic}.md'
        card_paths[topic]=card_rel
        ref_dir=ROOT/'review-guidance'/f'go-{topic}'
        skill=str(ref_dir/'SKILL.md')
        refs=[str(x) for x in sorted((ref_dir/'references').glob('*.md'))]
        applicable.append(dict(topic=topic,state='applicable',reason=reason,coverage='complete',skill_read=skill,references_read=refs,card=card_rel))
        coverage.append(dict(id=f'{candidate}/{topic}',status='complete',reason=reason+' All material bounded obligations assessed; contextual exclusions are stated in card limits.'))
        relevant=[dict(id=f['id'],severity=f['severity'],systemic=f['systemic'],cause=f['cause'],evidence=f['evidence'],correction=f['correction'],sources=[f'{candidate}/{topic}/{f["id"].split("/")[-1]}']) for f in findings]
        tledger=dict(scope=scope,coverage=[dict(id=topic,status='complete',reason=reason)],findings=relevant,strengths=[dict(id=f'{candidate}/{topic}/G1',evidence=GOOD[topic])],safeguards=[])
        save(f'cards/{candidate}/{topic}.ledger.json',tledger)
        proc=subprocess.run(['rtk','proxy','python3',str(GRADE),str(OUT/f'cards/{candidate}/{topic}.ledger.json')],capture_output=True,text=True,check=True)
        grade=json.loads(proc.stdout)
        save(f'cards/{candidate}/{topic}.grade.json',grade)
        grades[candidate][topic]=grade['overall']
        counts=collections.Counter(f['severity'] for f in findings)
        rationale=('One contained major error-contract failure supports C; this is the same canonical production cause cross-referenced across relevant topics.' if len(findings)==1 else 'Two independent important assertion gaps support C-: the join order assertion and independent-error assertion matrix need separate corrections.') if findings else 'No actionable in-topic defect found; the listed bounded risk decisions and executed checks verify a relevant strength. The controls are ordinary correct implementation, so A+ is not claimed.'
        good=GOOD[topic]
        if topic=='testing':
            if candidate=='A': good+=' The existing first-consumer cancellation error test and in-process command output/status assertions exercise useful contracts.'
            else: good+=' B author early-stop/join and wrapped-both-error tests reject the corresponding compiled mutations through intended assertions; the original targets were green.'
        bad='\n'.join(f'- [{f["id"]}][{f["severity"]}][existing-in-scope] {f["cause"]} {f["evidence"]} — {f["magnitude"]}' for f in findings) or '- None found.'
        changes='\n'.join(f'- [{f["id"]}] {f["correction"]} Verify using the referenced bounded probes and target-specific mutation checks; preserve accepted effects and source signature.' for f in findings) or '- None needed.'
        checks='Actual checks: host/minimum ordinary build, test and vet; empty gofmt diff; held tests (A all green, B independent-cancellation-class failure); independent FIFO, cleanup gates, writer failure and cancellation-class probes; actual executable cases; host race/shuffle count=3. See raw/independent-commands.json and raw/followup-commands.json for exact RTK-prefixed argv/status/stdout/stderr.'
        text=f'## {title} — {grade["overall"]}\nScope: {scope}\nCoverage: {reason} Assessed entire supplied code area relevant to this topic.\nRationale: {rationale}\nFinding counts: critical={counts["critical"]}, major={counts["major"]}, moderate={counts["moderate"]}, minor={counts["minor"]}\n\nGood\n\n- [{candidate}/{topic}/G1] {good}\n\nBad\n\n{bad}\n\nSuggested changes\n\n{changes}\n\nLimits: {LIMITS} {checks}\n\nSkill invoked and read: {skill}. References read: '+', '.join(refs)+'.\n'
        if topic=='testing':
            text+='\nUngraded related production finding: '+candidate+'/F1 (cancellation-class suppression). The Testing grade counts only the independently necessary assertion corrections listed above.\n'
        (OUT/card_rel).write_text(text)
    save(f'applicability-{candidate}.json',dict(candidate=candidate,scope=scope,all_nine_considered=True,topics=applicable))
    manifest['applicability'][candidate]=applicable
    ledger=dict(scope=scope,coverage=coverage,findings=[dict(id=f['id'],severity=f['severity'],systemic=f['systemic'],cause=f['cause'],evidence=f['evidence'],correction=f['correction'],sources=[f'{candidate}/{topic}/{f["id"].split("/")[-1]}' for topic in f['topics']],magnitude=f['magnitude'],primary_owner=f['primary_owner']) for f in FINDINGS[candidate]],strengths=STRENGTHS[candidate],safeguards=[])
    save(f'ledger-{candidate}.json',ledger)
    result=subprocess.run(['rtk','proxy','python3',str(GRADE),str(OUT/f'ledger-{candidate}.json')],capture_output=True,text=True,check=True)
    grade=json.loads(result.stdout)
    save(f'grade-{candidate}.json',grade)
    manifest.setdefault('grades',{})[candidate]=grade
    table='\n'.join(f'| {title} | {grades[candidate][topic]} | {reason} | [{topic}]({card_paths[topic]}) |' for topic,title,reason in TOPICS)
    counts=grade['unique_finding_counts']
    good='\n'.join(f'- [{s["id"]}] {s["evidence"]}' for s in STRENGTHS[candidate])
    bad='\n'.join(f'- [{f["id"]}][{f["severity"]}] {f["cause"]} — {f["magnitude"]} Evidence: {f["evidence"]}' for f in FINDINGS[candidate])
    changes='\n'.join(f'- [{f["id"]}][priority 1] {f["correction"]} Primary owner: {f["primary_owner"]}. Verify: {", ".join(f["evidence_labels"])}.' for f in FINDINGS[candidate])
    (OUT/f'review-{candidate}.md').write_text(f'# Go Quality Report — {grade["overall"]}\nScope: {scope}\nCoverage: Complete bounded supplied area; all nine topics apply for the contextual reasons below. No omitted material boundary within that area.\nRationale: {len(FINDINGS[candidate])} unique contained major causes select C- under the unchanged rubric. Cross-topic copies are deduplicated, strengths do not cancel defects, and no systemic-major or critical reach is supported. A and B are graded independently.\nUnique finding counts: critical={counts["critical"]}, major={counts["major"]}, moderate={counts["moderate"]}, minor={counts["minor"]}\n\n## Report cards\n\n| Topic | Grade/state | Assessed chunks and coverage limits | Cards |\n| --- | --- | --- | --- |\n{table}\n\n## Good\n\n{good}\n\n## Bad\n\n{bad}\n\n## Suggested changes\n\n{changes}\n\n## Limits\n\n{LIMITS}\n\nActual checks are independently executed; supplied checks-A/B.json are context only. Held executable child builds use GOQUALITY_GO pointing to the actual matching host/minimum toolchain. The additional deadline-after-stop probe extends the supplied held observation and is kept distinct from it. Four review-created mutation targets are documented in raw logs/patches; no unavailable original frozen author-specific mutation target is inferred or replaced. Timeouts are preserved without intended-assertion detection credit.\n\nDetails: [manifest](manifest.json), [findings](findings.json), [ledger](ledger-{candidate}.json), [calculator result](grade-{candidate}.json), [applicability](applicability-{candidate}.json), [raw commands](raw/independent-commands.json), [narrowed join follow-up](raw/followup-commands.json), [probe source](probes/review_probe_test.go).\n')

save('manifest.json',manifest)
mutations=[]
for candidate in ['A','B']:
    for name in ['no-early-stop','no-failure-stop','skip-early-join','broad-cancel-suppression']:
        data={r['label'].split('/')[-1]:r for r in COMMANDS if r['label'].startswith(f'{candidate}/mutation/{name}/')}
        assertion=data['independent-target']
        mutations.append(dict(candidate=candidate,mutation=name,original_green=data['original-target-green']['status']==0,mutant_compiles=data['compile']['status']==0,intended_bounded_review_assertion=assertion['confirmed_detection'],review_stdout=assertion['stdout'],author_suite_status=data['ordinary-suite']['status'],author_suite_stdout=data['ordinary-suite']['stdout'],author_suite_timeout_only='panic: test timed out' in data['ordinary-suite']['stdout'],limit='Author full-suite timeouts do not count as intended-assertion detection; target-specific follow-ups and narrowed A join evidence are separate.'))
save('mutation-summary.json',dict(mutations=mutations,narrowed_A_join=dict(baseline_author_green=True,baseline_probe_green=True,mutant_compiles=True,author_suite_count10_green=True,author_target_count10_green=True,independent_probe_intended_failure=True,raw='raw/followup-commands.json'),supplied_frozen_mutations='Not available in neutral packet; review mutations are supplementary, with no credit for unknown author-specific mutation targets.'))
summary=['# Independent Go pipeline/CLI review','', 'Candidate A: **C-**, three unique major causes. Candidate B: **C-**, two unique major causes. Complete bounded packet coverage; all-nine applicability and cards are recorded separately.','', '| Candidate | Correctness | Code quality | Testing | Unique critical/major/moderate/minor |','| --- | --- | --- | --- | --- |','| A | C | C | C- | 0 / 3 / 0 / 0 |','| B | C | C | C | 0 / 2 / 0 / 0 |','', 'Both lose an independent DeadlineExceeded error during cleanup when the coordinated child stop is Canceled. B additionally loses bare independent Canceled errors before stopping, in both callback roles; those are manifestations of B’s single production classification cause. A’s stop snapshot avoids that pre-stop failure but does not preserve the actual stop class.','', 'A’s narrowed premature-join mutant compiles and survives its full delivered suite at count=10; a held cleanup gate catches host return before cleanup release. B’s existing join test detects that mutation. A also misses a broader consumer-cleanup error suppression mutation; B’s wrapped-both-error assertion detects it. Each candidate separately has a missing exact-error/timing regression assertion gap.','', 'Actual host and Go1.22.12 binaries preserve checked integer ordering, early prefix, stderr and exit codes. Both pass ordinary build/test/vet and have empty formatting diffs. A passes supplied held checks; B fails the supplied independent cancellation-class test. Independent deadline-after-stop failures are recorded separately.','', '## Artifacts','', '- [A full report](review-A.md), [B full report](review-B.md)','- [Unique findings and strengths](findings.json)','- [A ledger](ledger-A.json), [B ledger](ledger-B.json), [A calculator output](grade-A.json), [B calculator output](grade-B.json)','- [Manifest](manifest.json), [A applicability](applicability-A.json), [B applicability](applicability-B.json)','- [Raw independent results](raw/independent-commands.json), [narrowed/targeted follow-ups](raw/followup-commands.json), [mutation summary](mutation-summary.json)','- [Original source hashes](source-hashes-before.json), [after-review hashes](source-hashes-after.json), [integrity](raw/source-integrity.json)','', '## Limits','', LIMITS, '', 'No external plans, reports, profiles, other-agent output or exposure identity was used. Four reviewer-created mutation targets are supplementary; the packet contains no original frozen author-specific mutation artifacts. Mutant compilation errors and timeout-only failures receive no detection credit. No fix, promotion or writing-guidance evaluation was performed.']
(OUT/'review.md').write_text('\n'.join(summary)+'\n')
print(json.dumps({c:manifest['grades'][c] for c in ['A','B']},indent=2))
