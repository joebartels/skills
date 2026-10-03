#!/usr/bin/env python3
"""Write already reconciled independent review artifacts; does not grade evidence."""
import hashlib
import json
from pathlib import Path
from run_checks import PACKET, OUT, GO122, hashes

TOPICS = [
    ('architecture', 'Architecture & Design', 'go-architecture-and-design', 'design-decisions.md'),
    ('code-quality', 'Code Quality & Go Idioms', 'go-code-quality-and-idioms', 'idiom-decisions.md'),
    ('correctness', 'Correctness & Compatibility', 'go-correctness-and-compatibility', 'correctness-compatibility-decisions.md'),
    ('testing', 'Testing', 'go-testing', 'testing-decisions.md'),
    ('security', 'Security', 'go-security', 'security-decisions.md'),
    ('observability', 'Observability & Resilience', 'go-observability-and-resilience', 'observability-resilience-decisions.md'),
    ('performance', 'Performance & Resource Management', 'go-performance-and-resource-management', 'performance-decisions.md'),
    ('dependencies', 'Dependencies & Reproducibility', 'go-dependencies-and-reproducibility', 'dependency-reproducibility-decisions.md'),
    ('deployment', 'Deployment & Operations', 'go-deployment-and-operations', 'deployment-operations-decisions.md'),
]

REASONS = {
    'architecture': 'Applicable: per-instance ownership, zero-value construction, public API and coherent value publication are library design decisions. The complete supplied API and implementation are assessed.',
    'code-quality': 'Applicable and mandatory: complete Go implementation and tests contain receiver, mutex ownership, documentation and Go-version idiom decisions.',
    'correctness': 'Applicable and mandatory: exact update count/sum, zero value, simultaneous Add/Snapshot, coherent independent values, isolated instances and retained public signatures are explicit README contracts.',
    'testing': 'Applicable and mandatory: candidate tests must meaningfully detect concurrent coherence, accepted-count, arithmetic and instance-state regressions. Assertions and termination are checked with compiled behavioral mutations.',
    'security': 'Not applicable in this bounded code area: fixed-width integer arithmetic and a caller-owned local mutex introduce no attacker-controlled I/O, privilege boundary, credentials, sensitive sink or input-sized resource amplification. No external dependency is selected. This is a scope judgment, not a claim that any embedding application is secure.',
    'observability': 'Not applicable in this bounded code area: synchronous local arithmetic has no fallible external dependency, retry, queue, telemetry export or service lifecycle contract. Mutex ownership/liveness is assessed under Correctness and Performance. No context, instrumentation or operational machinery is required by the supplied contract.',
    'performance': 'Applicable: adding synchronization changes contention and resource ownership. Fixed work/state, finite critical sections, no internal goroutines, per-instance locks and allocations are assessed. No throughput/latency budget or speed claim is supplied.',
    'dependencies': 'Applicable: standard-library-only and minimum Go 1.22 are explicit build contracts. Standalone module resolution and compilation/tests on Go 1.22.12 and host 1.26.5 are assessed with network disabled.',
    'deployment': 'Not applicable in this bounded code area: it supplies a local library and no change to release tags, packaging, CI enforcement, process configuration, deployment or operator contract. Source builds are assessed under Dependencies and Correctness. Excluded surrounding delivery infrastructure is not alleged to be missing.',
}

GOOD = {
    'architecture': [
        'G1: `totals.go:12-15` owns one mutex and two counters per Totals. `Add` commits synchronously, and Snapshot publishes a two-scalar value. There is no global state, background worker, constructor or new public abstraction. This is the smallest direct ownership model needed by the README.',
        'G2: The external-consumer signature assignments and zero-value/value-ownership assertions in `contract_probe_test.go` compile and pass on both toolchains; no caller lifecycle setup is introduced. Global-state mutations are rejected by the original instance tests on both toolchains.',
    ],
    'code-quality': [
        'G1: `totals.go:19-30` keeps each critical section visible and short. Pointer receivers preserve mutation and mutex identity; returned Snapshot has only scalar values. The documented no-copy-after-use constraint is retained at `totals.go:11` and in README.',
        'G2: Independent `gofmt -d` is empty and `go vet -mod=readonly ./...` succeeds on Go 1.22.12 and 1.26.5. Integer-range loops compile under the declared `go 1.22` semantics. There is no optional modernization requirement or ignored-error path in this arithmetic-only implementation.',
    ],
    'correctness': [
        'G1: `Add` holds `t.mu` across both `count++` and `sum += delta` (`totals.go:20-23`); Snapshot evaluates both fields while holding that same mutex (`totals.go:28-30`). Return-value evaluation precedes deferred unlock. This establishes one observable update and excludes a mixed publication for supported uncopied instances.',
        'G2: Independent race/shuffle runs and the event-based checkpoint probe pass on both toolchains. The latter waits for reader readiness, parks every writer after its first accepted update, asserts intermediate `{4 4}`, releases the remaining updates, joins work and asserts final `{4000 4000}`. It does not infer completion from a guessed sleep or the numeric result.',
        'G3: The external-consumer probe passes zero snapshot, zero/positive/negative deltas, exact per-call count, retained old snapshots, caller mutation isolation and independent instance state. Supplied held tests, independently rerun, also verify mixed concurrent deltas `{1000 0}` and separate concurrent instances. The public method types and Snapshot fields are unchanged.',
    ],
    'testing': [
        'G1: Both original candidate test suites reject the split-update mutation using their actual `incoherent snapshot` assertion on host and Go 1.22.12. The mutant protects each field access with the same mutex but unlocks between the two changes; these non-race runs demonstrate multi-field invariant sensitivity rather than relying on compilation failure or race output.',
        'G2: Both suites reject process-global state using their instance assertions and reject sign normalization using the sequential exact `{Count:2, Sum:1}` assertion on both toolchains. The compiler accepts all four behavior-changing mutants; no build error is counted as detection.',
    ],
    'performance': [
        'G1: `totals.go:12-30` stores only a mutex and two integers, performs a fixed amount of arithmetic, returns a fixed-size value, and creates no internal goroutines, queues, buffers, I/O or deferred lifetime beyond one method call. Neither method runs arbitrary caller work while locked. State and owned work do not grow with update count.',
        'G2: `TestReviewConstantResourceUse` verifies zero allocations per Add/Snapshot pair using `testing.AllocsPerRun(100, ...)` on both toolchains in the independent contract/race runs. No global lock couples separate Totals. The normal direct mutex implementation fits the requested design without speculative lock-free optimization.',
    ],
    'dependencies': [
        'G1: The complete `go.mod` declares Go 1.22 and has no require/replace/exclude/toolchain directives. Source and tests import only standard-library packages. `go list -mod=readonly -m all` returns only `example.com/library-shared-state`; no absent go.sum defect is inferred for a standard-library-only module.',
        'G2: Builds, ordinary tests, vet and race/shuffle runs pass on the actual Go 1.22.12 binary and host Go 1.26.5 with `GOTOOLCHAIN=local`, `GOWORK=off`, `GOPROXY=off`, `GOSUMDB=off`, a review-owned module cache and initially empty review-owned build cache. This verifies standalone network-disabled builds from the frozen source copies.',
    ],
}

COVERAGE = {
    'architecture': 'Complete supplied module/API: Snapshot, Totals, Add, Snapshot; per-instance state, construction, copy constraint and synchronous ownership. No persistence, transport, remote error or background-lifecycle boundary exists here.',
    'code-quality': 'Complete `totals.go` and `totals_test.go`, README and go.mod: names, comments, value/pointer receivers, no-copy semantics, lock flow, tests as local Go code, formatting/vet and effective Go version. Assertion strategy belongs to Testing.',
    'correctness': 'Complete ordinary/zero/signed delta behavior, count semantics, coherent publication, simultaneous readers/writers, instance isolation, returned-value ownership, public method types and supported minimum Go version. Illegal copying of a used Totals and nil-pointer use are outside the explicit contract.',
    'testing': 'Complete candidate assertions and test goroutine lifecycles; actual baseline/race/repeat/shuffle execution; four independent compiled behavioral mutations per candidate on two toolchains; added reviewer-owned event-based/held checks for the production contract are explicitly separate from candidate tests.',
    'performance': 'Complete runtime resource ownership and clearly derived fixed state/work; independent allocation assertion and multi-instance checks. Mutex contention under a production workload has no supplied budget or optimization claim; no unmeasured speed conclusion is made.',
    'dependencies': 'Complete standalone module/import/build context; Go 1.22.12 and 1.26.5 darwin/arm64 builds/test/vet, disabled network resolution, selected module list and preservation of all inputs. No generated/native/vendor/private-module input exists in the complete supplied module.',
}

FINDING = {
    'id': 'A-F1', 'candidate': 'A', 'topic': 'Testing', 'severity': 'moderate', 'systemic': False,
    'classification': 'existing-in-scope', 'primary_owner': 'candidate A concurrency test lifecycle',
    'cause': 'Reader completion is conditional on observing the expected final Count rather than on writer completion.',
    'location': {'path': str(PACKET / 'candidates/A/totals_test.go'), 'start': 48, 'end': 58},
    'trigger': 'A coherent count-loss regression leaves Count==Sum but below 8000 after every writer returns.',
    'consequence': 'The reader continues polling forever and the test goroutine waits at snapshotsDone.Wait(); the final exact-count assertion is unreachable. A test-runner deadline rejects the run, but reports a timeout instead of the wrong final value and delays feedback.',
    'impact_rationale': 'Moderate test fragility: a normal important regression prevents orderly completion and result assertions. The suite still rejects through an external go test timeout, and incoherence/instance/arithmetic assertions are effective. This is not an unverified core contract, a production race, a systemic major or a critical failure.',
    'evidence': [
        'candidates/A/totals_test.go:42-51 has no stop event other than incoherence or exact expected Count; :55-58 joins the reader before asserting final state.',
        'raw/A-mutation-drop-unit-updates-host.json and raw/A-mutation-drop-unit-updates-go122.json: all calls returning without unit updates leave coherent {0 0}; commands exit 1 only when panic: test timed out after 2s. Stack shows test waiting on snapshotsDone and reader polling Snapshot.',
        'raw/A-mutation-drop-unit-updates-full-suite-go122.json reproduces the timeout in the whole candidate suite, not only a filtered test.',
        'raw/A-mutation-drop-unit-updates-event-probe-go122.json: independent checkpoint/worker-completion probe terminates and reports checkpoint/final wrong state via explicit assertions.',
    ],
    'correction': 'Make reader shutdown depend on the writer-completion event (and give waits a bounded failure path), join the reader, then assert exact final Count/Sum on the test goroutine.',
    'verification': 'Keep the coherent dropped-unit-update mutant and require a prompt final-state assertion with got {0 0}, not a timeout; rerun ordinary and -race/-shuffle tests on Go 1.22.12 and host. Preserve the split-update incoherence assertion.',
    'sources': ['A/testing/F1'],
    'related_topics': [],
    'reconciliation': 'One independently actionable test-lifecycle cause. Not duplicated into Architecture, Code Quality, Correctness, Performance or Observability; the production methods are unchanged and correct in the reviewed scope.',
}

def rawlinks(c, topic):
    ids = [c + '-test-go122', c + '-race-go122', c + '-race-host']
    if topic == 'code-quality':
        ids = [c + '-gofmt', c + '-vet-go122', c + '-vet-host']
    elif topic == 'dependencies':
        ids = [c + '-version-go122', c + '-modules', c + '-build-go122', c + '-build-host']
    elif topic in ('correctness', 'architecture', 'performance'):
        ids += [c + '-contract-go122', c + '-contract-host']
    elif topic == 'testing':
        ids += [c + '-mutation-' + m + '-go122' for m in ('split-update', 'global-state', 'wrong-signed-sum', 'drop-unit-updates')]
    return '; '.join('[' + i + '](../../raw/' + i + '.json)' for i in ids)

snapshots = hashes()
before = json.loads((OUT / 'source-hashes-before.json').read_text())
expected = json.loads((PACKET / 'source-manifest.json').read_text())
assert snapshots == before
for c in ['A', 'B']:
    assert snapshots['candidates/' + c] == expected[c]

(OUT / 'findings.json').write_text(json.dumps({
    'scope_mode': 'bounded code-area review; original supplied as contract/baseline context',
    'confirmed': [FINDING],
    'candidate_counts': {'A': {'critical': 0, 'major': 0, 'moderate': 1, 'minor': 0},
                         'B': {'critical': 0, 'major': 0, 'moderate': 0, 'minor': 0}},
    'ungraded_or_rejected': [
        {'id': 'Q1', 'candidates': ['A', 'B'], 'concern': 'No explicit guarantee that the original stress-test reader samples intermediate states before all writers finish.',
         'decision': 'Optional deterministic checkpoint improvement, not a confirmed defect. Existing candidate assertions actually reject a race-free split-update mutation on both toolchains; supplied/rerun -race checks exercise shared paths. No demonstrated false-pass failure for a required normal-use contract is claimed. The independent event-based probe closes the production outcome check without being credited as candidate-authored coverage.'},
        {'id': 'Q2', 'candidates': ['A'], 'concern': 'No separate saved-snapshot test in A.',
         'decision': 'Not graded: the required public Snapshot contains only two int64 fields and returns by value. Independence follows from the unchanged public shape and is additionally checked by the external consumer probe. A pointer-return API mutation would be a build error and would not count as behavioral assertion sensitivity.'},
        {'id': 'Q3', 'candidates': ['A', 'B'], 'concern': 'Mutex contention versus atomics/RWMutex.',
         'decision': 'Not graded: no performance budget, contention measurement or speed claim. Multi-field coherence requires joint publication, and a direct implementation is explicitly required. Fixed resource use is verified.'},
    ],
    'mutations_scope': 'Supplementary independent probes only; no inference about unspecified original frozen mutation targets or author-specific outcomes.',
}, indent=2) + '\n')

manifest = {
    'scope': 'Independent bounded code-area outcome review of candidate A and candidate B; complete 4-file modules; original README/source/tests and supplied checks/held tests are context.',
    'packet': str(PACKET), 'date': '2026-10-02', 'reviewer': 'independent-library-review',
    'source_snapshots_sha256': snapshots,
    'candidate_manifest_matches': True, 'before_after_hashes_match': True,
    'guidance': [{'path': str(p), 'sha256': hashlib.sha256(p.read_bytes()).hexdigest()}
                 for p in sorted((PACKET / 'review-guidance').rglob('*')) if p.is_file()],
    'exclusions': ['repository planning/results and profiles', 'author reports', 'exposure identity', 'other agents',
                   'embedding application/runtime/release workflows outside supplied code area', 'unsupported nil receivers and copying a used Totals'],
    'review_method': 'Single independent reviewer, candidates assessed separately; no delegation; unchanged packet skills/references; only disposable copies mutated.',
    'toolchains': {'minimum': GO122, 'minimum_observed': 'go1.22.12 darwin/arm64', 'host_observed': 'go1.26.5 darwin/arm64'},
    'supplied_checks': ['checks-A.json', 'checks-B.json'],
    'executed_checks': 'raw/*.json and corresponding raw/*.txt; run_checks.py, run_probes.py and run_followup.py retain exact construction/commands.',
    'boundaries': [
        {'id': 'public-api-and-zero-value', 'owner': 'reviewer', 'status': 'complete', 'evidence': 'external-consumer signatures and zero snapshot probe'},
        {'id': 'one-observable-update', 'owner': 'reviewer', 'status': 'complete', 'evidence': 'common mutex trace, checkpoint race tests, split-update assertion sensitivity'},
        {'id': 'count-and-signed-sum', 'owner': 'reviewer', 'status': 'complete', 'evidence': 'sequential/mixed-delta/zero-delta assertions, lost-count and wrong-signed-sum mutations'},
        {'id': 'instance-and-return-value-ownership', 'owner': 'reviewer', 'status': 'complete', 'evidence': 'scalar return shape, held/reviewer tests and global-state mutation'},
        {'id': 'test-completion-signal', 'owner': 'reviewer', 'status': 'complete', 'evidence': 'test lifecycle trace, lost-count whole-suite reproduction and event-probe assertions'},
        {'id': 'minimum-go-and-stdlib', 'owner': 'reviewer', 'status': 'complete', 'evidence': 'actual Go1.22.12 network-disabled standalone builds/module list'},
    ],
    'material_unavailable_coverage': [],
    'limits': ['No exhaustive proof of every schedule is claimed from finite race/repeat runs.',
               'No release/CI/platform matrix or embedding application is assessed; none is promised in bounded packet.',
               'No vulnerability scan, profiler or byte-identical binary claim; no threat/speed/artifact-identity contract implicates those checks.',
               'Reviewer probes do not become candidate-authored tests or replace unspecified frozen evaluation targets.'],
    'outputs': {'report': 'review.md', 'findings': 'findings.json', 'cards': 'cards/<A|B>/<topic>.md',
                'applicability': 'applicability-<A|B>.json', 'ledgers': 'ledger-<A|B>.json'},
}
(OUT / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')

for c in ['A', 'B']:
    folder = OUT / 'cards' / c
    folder.mkdir(parents=True, exist_ok=True)
    applicability = []
    coverage_rows = []
    for topic, title, skill, ref in TOPICS:
        applicable = topic in COVERAGE
        grade = ('B' if c == 'A' and topic == 'testing' else 'A') if applicable else 'Not applicable'
        state = 'complete' if applicable else 'not_applicable'
        applicability.append({'id': topic, 'topic': title, 'applicable': applicable, 'coverage_status': state,
                              'grade': grade, 'reason': REASONS[topic],
                              'skill': str(PACKET / 'review-guidance' / skill / 'SKILL.md'),
                              'reference': str(PACKET / 'review-guidance' / skill / 'references' / ref),
                              'card': 'cards/' + c + '/' + topic + '.md',
                              'essential_unavailable_evidence': []})
        coverage_rows.append({'id': topic, 'status': state, 'reason': REASONS[topic]})
        scope = ('Candidate ' + c + ', complete bounded code area `candidates/' + c +
                 '/{README.md,go.mod,totals.go,totals_test.go}`; original README/source are the behavioral baseline. '
                 'Go 1.22 minimum, standard library only. Exact file SHA-256 values are in [manifest](../../manifest.json).')
        prefix = ('## ' + title + ' — ' + grade + '\n\nScope: ' + scope + '\n\n'
                  'Skill used: [' + skill + '](' + str(PACKET / 'review-guidance' / skill / 'SKILL.md') +
                  '), including [' + ref + '](' + str(PACKET / 'review-guidance' / skill / 'references' / ref) + ').\n\n')
        if not applicable:
            content = prefix + 'Coverage: The complete supplied library, input/ownership model and module metadata were inspected to decide applicability.\n\nRationale: ' + REASONS[topic] + '\n\nLimits: This is justified irrelevance within the bounded local code area, not unavailable material evidence or an approval of an external application/release. Surrounding application, deployment and policy evidence is outside the requested scope.\n'
        else:
            issue = c == 'A' and topic == 'testing'
            rationale = ('One confirmed moderate test-lifecycle defect selects B. Other important assertions have demonstrated sensitivity; the timeout still rejects lost updates, so this is contained fragility rather than effective absence of core verification.' if issue else
                         'No actionable in-topic issue is confirmed. The relevant material decisions are assessed and the specific strengths below are verified; this supports A. The direct mutex/ordinary test/build mechanisms are routine correct setup, not two independently verified nonroutine safeguards for A+.')
            content = prefix + 'Coverage: ' + COVERAGE[topic] + '\n\nRationale: ' + rationale + '\n\nFinding counts: critical=0, major=0, moderate=' + ('1' if issue else '0') + ', minor=0\n\nGood\n\n'
            good = list(GOOD[topic])
            if topic == 'testing' and c == 'B':
                good.append('G3: `totals_test.go:54-59` stops the reader after writers finish and then asserts exact final state. The coherent dropped-unit-update mutant reaches `final snapshot = {Count:0 Sum:0}` promptly on both toolchains and in the whole-suite Go 1.22 check. The saved Snapshot test also checks value retention explicitly at `totals_test.go:66-71`.')
            content += '\n'.join('- [' + item.replace(': ', '] ', 1) for item in good) + '\n\nBad\n\n'
            if issue:
                content += '- [F1][moderate][existing-in-scope] `totals_test.go:48-58` makes reader termination depend on expected Count==8000. If coherent updates are lost, writer completion cannot stop the reader; `snapshotsDone.Wait()` blocks before the final count assertion. The compiled dropped-unit-update mutation times out after 2s on both toolchains and in the whole Go 1.22 suite. This is [canonical A-F1](../../findings.json), owned by the test lifecycle.\n\nSuggested changes\n\n- [F1] Stop the reader on a writer-completion event, bound lifecycle waits, join the reader, then assert final Count/Sum. Retain the coherent dropped-update mutation and require a prompt wrong-value assertion instead of a timeout. The independent checkpoint probe demonstrates this signal in a disposable copy; it is not counted as an existing candidate test.\n'
            else:
                content += '- None found.\n\nSuggested changes\n\n- None needed.\n'
            limits = ('Checks are actual independent executions in disposable copies; supplied logs are corroboration only. '
                      'No unrelated full-repository or shipping-environment claim is made. Race/repeat runs sample schedules and do not prove all interleavings. '
                      'No compilation failure is credited as behavioral detection; all four mutants compile. '
                      'The external review/held tests check outcomes and do not become candidate-authored regression coverage. ')
            if topic == 'performance':
                limits += 'No latency/throughput budget or comparison is supplied; no speed or lock-contention conclusion is made. '
            if topic == 'dependencies':
                limits += 'The exercised target is darwin/arm64 on the two stated toolchains; no unspecified target matrix or byte-identical artifact claim. '
            if topic == 'testing' and c == 'A':
                limits += 'The 2s Go runner deadline is an independent guard; the candidate reader itself has no stop/deadline for wrong-but-coherent totals. Timeout rejection is distinguished from final-state assertion detection. '
            content += '\nLimits: ' + limits + '\n\nExecuted evidence: ' + rawlinks(c, topic) + '. Complete command/stdout/stderr/status pairs are under [raw](../../raw); [mutation sensitivity](../../mutation-sensitivity.json) retains intended assertions.\n'
        (folder / (topic + '.md')).write_text(content)
    (OUT / ('applicability-' + c + '.json')).write_text(json.dumps(applicability, indent=2) + '\n')
    ledger_findings = []
    if c == 'A':
        ledger_findings.append({'id': 'A-F1', 'severity': 'moderate', 'systemic': False,
            'cause': FINDING['cause'], 'evidence': 'candidates/A/totals_test.go:48-58; raw dropped-unit-update host/go122 and full-suite Go1.22 logs show 2s timeout with test waiting for reader. Event-based reviewer probe terminates with wrong-state assertions.',
            'correction': FINDING['correction'], 'sources': FINDING['sources']})
    ledger = {'scope': 'Candidate ' + c + ': complete bounded 4-file local library code area; frozen SHA-256 source snapshots in manifest.json.',
              'coverage': coverage_rows, 'findings': ledger_findings,
              'strengths': [
                  {'id': c + '-G1', 'evidence': 'totals.go:20-30 uses one per-instance mutex around both update fields and snapshot publication; independent race/checkpoint/mixed-delta/value-ownership tests pass on Go1.22.12 and1.26.5.'},
                  {'id': c + '-G2', 'evidence': 'Actual Go1.22.12 and host standalone network-disabled readonly builds/tests/vet pass; standard-library-only selected module graph; gofmt diff empty.'},
                  {'id': c + '-G3', 'evidence': 'Candidate-authored assertions reject compiled split-update, global-state and wrong-signed-sum behavioral mutations on both toolchains; raw logs preserve exact failing assertions.'},
              ], 'safeguards': []}
    (OUT / ('ledger-' + c + '.json')).write_text(json.dumps(ledger, indent=2) + '\n')

mutation_rows = json.loads((OUT / 'mutation-sensitivity.json').read_text())
mutation_doc = '# Independent supplementary mutation sensitivity\n\nEvery mutation was applied only to a disposable copy. All commands use RTK and disable automatic toolchain/workspace/network resolution. The tested candidate test source is unchanged; added reviewer probes are explicitly identified below. These four independently selected mutations do not replace or claim credit for unspecified frozen evaluation targets.\n\n| Candidate | Mutation | Intended assertion | Host / Go 1.22.12 | Evidence |\n| --- | --- | --- | --- | --- |\n'
for c in ['A', 'B']:
    for name in ['drop-unit-updates', 'split-update', 'global-state', 'wrong-signed-sum']:
        rows = [r for r in mutation_rows if r['candidate'] == c and r['mutation'] == name]
        outcomes = [('behavioral assertion' if r['behavioral_assertion_detected'] else 'runner deadline only' if r['deadline_rejection'] else 'not detected') for r in rows]
        mutation_doc += '| ' + c + ' | ' + name + ' | ' + rows[0]['intended_assertion'] + ' | ' + ' / '.join(outcomes) + ' | ' + ', '.join('[' + r['toolchain'] + '](' + r['raw_record'] + ')' for r in rows) + ' |\n'
mutation_doc += '\nNo build error occurred; no build error is awarded behavioral detection credit. The split-update mutation synchronizes each field separately and yields between changes; the raw non-race assertions validate coherent-pair detection. The dropped-unit-update mutant returns normally and retains a coherent {0 0}; A reaches only the 2s test-runner timeout while B reports the incorrect final value. Whole-suite Go1.22 checks confirm that distinction.\n\nThe reviewer-owned `TestReviewConcurrentCheckpoint` starts a reader before writers, parks every writer after one completed update, records an intermediate snapshot, releases work, waits on writer completion, stops/joins the reader, and asserts intermediate/final values. In both dropped-update copies it completes and reports wrong-state assertions; this verifies the proposed independent event signal without giving candidate-authored coverage credit. Exact commands/results: [A](raw/A-mutation-drop-unit-updates-event-probe-go122.json), [B](raw/B-mutation-drop-unit-updates-event-probe-go122.json).\n'
(OUT / 'mutation-sensitivity.md').write_text(mutation_doc)

report = '# Independent Go Quality Reports — candidate A: B; candidate B: A\n\nScope: Separate bounded code-area reviews of the complete candidate A and B four-file library modules. The original README/source establish the same contract for each: Go 1.22, standard library only, zero value, no copying after use, exact Count/Sum, coherent independent snapshots and isolated instances. Exact candidate and original SHA-256 snapshots, skill/reference hashes and coverage are in [manifest](manifest.json). No author report, profile, exposure identity, repository planning/results or other reviewer material was used.\n\nCoverage: All nine topic skills and relevant references were read unchanged. Architecture, Code Quality, Correctness, Testing, Performance and Dependencies are assessed completely within the supplied area. Security, Observability and Deployment have explicit scope-based Not applicable reasons; none is substituted for missing material evidence. No material requested obligation remains unavailable. Each candidate was independently exercised on host Go1.26.5 and actual Go1.22.12; mutations exist only in disposable copies.\n'
for c in ['A', 'B']:
    grade = 'B' if c == 'A' else 'A'
    report += '\n## Candidate ' + c + ' — ' + grade + '\n\n'
    report += ('Rationale: One unique moderate test-lifecycle defect selects B; no production defect is confirmed. Lost updates remain coherent, so A waits forever for expected Count before it can run its final-state assertion. The independent runner deadline rejects the mutant, but that is weaker and slower diagnostic evidence than a result assertion. Other important assertions have verified sensitivity.\n\nUnique finding counts: critical=0, major=0, moderate=1, minor=0.\n' if c == 'A' else
               'Rationale: No actionable defect is confirmed in the complete bounded scope. The mutex establishes joint publication, signature/zero-value/value-ownership checks pass, and the original test assertions detect all four compiled supplementary mutations. Reader shutdown follows writer completion, so a wrong final count produces an immediate assertion. Verified ordinary correct setup supports A; no two nonroutine safeguards support A+.\n\nUnique finding counts: critical=0, major=0, moderate=0, minor=0.\n')
    report += '\n### Report cards\n\n| Topic | Grade/state | Coverage and applicability | Card |\n| --- | --- | --- | --- |\n'
    for topic, title, _, _ in TOPICS:
        topic_grade = ('B' if c == 'A' and topic == 'testing' else 'A') if topic in COVERAGE else 'Not applicable'
        report += '| ' + title + ' | ' + topic_grade + ' | ' + REASONS[topic] + ' | [card](cards/' + c + '/' + topic + '.md) |\n'
    report += '\n### Good\n\n- [' + c + '-G1] The same per-instance mutex protects both arithmetic updates and snapshot evaluation; scalar returns preserve independent value ownership. External API/zero/mixed/checkpoint/held race checks pass on both toolchains — preserves the explicitly required coherent observation contract.\n- [' + c + '-G2] Standalone network-disabled readonly builds/test/vet succeed on the minimum and host toolchains; gofmt diff is empty and module graph is standard-library-only — preserves the declared build floor without hidden resolution inputs.\n- [' + c + '-G3] Candidate assertions actually reject incoherent-pair, process-global-state and signed-sum regressions after successful compilation — provides observable regression signal beyond test names or clean baseline output.\n'
    if c == 'B':
        report += '- [B-G4] The dropped-update mutation reports `final snapshot = {Count:0 Sum:0}` immediately on both toolchains; completion does not depend on the expected numeric result — keeps exact-count verification reachable when that behavior regresses.\n'
    report += '\n### Bad\n\n'
    if c == 'A':
        report += '- [A-F1][moderate] [Reader completion depends on expected Count](cards/A/testing.md), at `candidates/A/totals_test.go:48-58`. A coherent dropped-update mutant finishes its writers but leaves the reader polling and the test waiting until the 2s runner deadline; the final-state assertion never runs. Confirmed on host and Go1.22.12, including the whole candidate suite. This is a test-lifecycle defect, counted once.\n\n### Suggested changes\n\n- [A-F1][priority 2] The concurrency-test owner should stop/join the reader from a writer-completion event, bound failure waits and then assert exact totals. Verify that the retained dropped-update mutant produces a prompt wrong-value assertion while the split-update mutant still fails coherence. The reviewer-owned [event-based probe](contract_probe_test.go) demonstrates the intended signal; it is separate from candidate tests.\n'
    else:
        report += '- None found.\n\n### Suggested changes\n\n- None needed. Optional deterministic intermediate checkpoints may strengthen scheduling coverage; they are not graded defects.\n'
    report += '\n### Limits\n\nThe grade is for the supplied library area, not an embedding application or release. Finite race/shuffle runs do not prove every schedule. All actual baseline commands and checks are preserved in [raw](raw); supplied [checks-' + c + '](../checks-' + c + '.json) are distinguished from independent reruns. The [mutation matrix](mutation-sensitivity.md) records each intended assertion and separates assertion failures, runner deadlines and build errors. None of the four supplementary mutants failed compilation; A timeout is explicitly not final-state assertion credit. No unspecified frozen evaluation result is inferred. No speed/latency, vulnerability-scan, unspecified platform-matrix or byte-identical artifact conclusion is claimed. Sources match their initial hashes and supplied candidate manifest.\n\nDetails: [applicability](applicability-' + c + '.json), [ledger](ledger-' + c + '.json), [calculator result](grade-' + c + '.json), [deduplicated findings](findings.json).\n'
report += '\nIndependent evidence: [source hashes before](source-hashes-before.json), [after probes](source-hashes-after-probes.json), [probe construction](run_probes.py), [exact baseline construction](run_checks.py), [follow-up construction](run_followup.py). Review performed by one independent reviewer without delegation; identical production mechanisms are described separately for each snapshot and are never counted as multiple defects.\n'
(OUT / 'review.md').write_text(report)
(OUT / 'source-hashes-final.json').write_text(json.dumps(snapshots, indent=2) + '\n')
print(json.dumps({'source_hashes_match': True, 'candidate_manifest_matches': True, 'cards_written': 18,
                  'confirmed_findings': 1, 'report': str(OUT / 'review.md')}, indent=2))
