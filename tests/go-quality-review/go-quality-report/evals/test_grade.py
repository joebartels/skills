"""Tests for the reconciled-ledger calculator, not for semantic review quality."""
import copy
import importlib.util
from pathlib import Path
import unittest

SOURCE = Path(__file__).resolve().parents[4] / "plugins/go-quality-review/skills/go-quality-report/scripts/grade.py"
spec = importlib.util.spec_from_file_location("grade", SOURCE)
grade = importlib.util.module_from_spec(spec)
spec.loader.exec_module(grade)


def ledger(severities=(), complete=True):
    return {"scope": "base abc to def", "coverage": [{"id": "handler/security", "status": "complete" if complete else "partial", "reason": "source inspected" if complete else "storage unavailable"}],
            "findings": [{"id": f"F{i}", "severity": s, "systemic": False, "evidence": "verified source and trigger", "cause": f"independent cause {i}", "correction": f"independent fix {i}", **({"reach": "demonstrated broad impact"} if s == "critical" else {})} for i, s in enumerate(severities, 1)],
            "strengths": [{"id": "G1", "evidence": "verified boundary assertions"}], "safeguards": []}


class GradeTests(unittest.TestCase):
    def test_grade_anchors(self):
        for severities, expected in [(('critical',), 'F'), (('major','major'),'C-'), (('major','moderate','moderate'),'C-'), (('major',),'C'), (('moderate','moderate'),'C+'), (('moderate','minor','minor'),'B-'), (('moderate',),'B'), (('minor','minor'),'B+'), (('minor',),'A-'), ((),'A')]:
            with self.subTest(expected=expected): self.assertEqual(grade.calculate(ledger(severities))["overall"], expected)
    def test_systemic_major(self):
        data=ledger(('major',));data['findings'][0]['systemic']=True
        with self.assertRaises(ValueError): grade.calculate(data)
        data['findings'][0]['reach']='checkout and rollback workflows';self.assertEqual(grade.calculate(data)['overall'],'C-')
    def test_critical_needs_reach(self):
        data=ledger(('critical',));del data['findings'][0]['reach']
        with self.assertRaises(ValueError): grade.calculate(data)
        data['findings'][0]['reach']='broad customer data access';self.assertEqual(grade.calculate(data)['overall'],'F')
    def test_missing_coverage_never_clean_grade(self):
        self.assertEqual(grade.calculate(ledger(complete=False))['overall'],'Insufficient evidence')
    def test_partial_keeps_confirmed_defect(self):
        result=grade.calculate(ledger(('major',),False));self.assertEqual(result['overall'],'Insufficient evidence');self.assertEqual(result['assessed_scope_grade'],'C')
    def test_no_strength_no_a(self):
        data=ledger();data['strengths']=[];self.assertEqual(grade.calculate(data)['overall'],'Insufficient evidence')
    def test_na_does_not_mask_incomplete(self):
        data=ledger();data['coverage'][0]['status']='not_applicable';self.assertEqual(grade.calculate(data)['overall'],'Not applicable')
        data['coverage'].append({'id':'other','status':'unavailable','reason':'missing skill'});self.assertEqual(grade.calculate(data)['overall'],'Insufficient evidence')
    def test_duplicate_ids_rejected(self):
        for key in ['findings','coverage','strengths']:
            data=ledger(('moderate',));data[key].append(copy.deepcopy(data[key][0]))
            with self.subTest(key=key),self.assertRaises(ValueError):grade.calculate(data)
    def test_partition_invariance(self):
        data=ledger(('moderate','moderate'));before=grade.calculate(data)['overall']
        data['coverage']=[{'id':f'chunk{i}','status':'complete','reason':'verified'} for i in range(20)]
        self.assertEqual(grade.calculate(data)['overall'],before)
    def test_a_plus_independence(self):
        data=ledger();data['safeguards']=[{'id':'S1','risk':'duplicate charge','evidence':'lost response test','nonroutine':True},{'id':'S2','risk':'work after cancellation','evidence':'cancellation test','nonroutine':True}]
        self.assertEqual(grade.calculate(data)['overall'],'A+')
        data['safeguards'][1]['risk']='duplicate charge';self.assertEqual(grade.calculate(data)['overall'],'A')
        data['findings']=ledger(('moderate',))['findings'];self.assertEqual(grade.calculate(data)['overall'],'B')
    def test_bad_inputs(self):
        for mutate in [lambda d:d.update(coverage=[]), lambda d:d['findings'][0].update(severity='maybe'), lambda d:d['findings'][0].update(evidence=''), lambda d:d['coverage'][0].update(status='skipped'), lambda d:d['findings'][0].update(systemic='yes'), lambda d:d['coverage'][0].update(status=[]), lambda d:d['findings'][0].update(severity={})]:
            data=ledger(('moderate',));mutate(data)
            with self.assertRaises(ValueError):grade.calculate(data)
    def test_na_with_defect_rejected(self):
        data=ledger(('major',));data['coverage'][0]['status']='not_applicable'
        with self.assertRaises(ValueError):grade.calculate(data)

if __name__ == '__main__': unittest.main()
