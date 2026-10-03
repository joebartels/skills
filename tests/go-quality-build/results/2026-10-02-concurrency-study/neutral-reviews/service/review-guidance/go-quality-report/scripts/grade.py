#!/usr/bin/env python3
"""Grade an already reconciled review ledger. This does not verify evidence."""

import argparse
from collections import Counter
import json
from pathlib import Path


def required_text(row, key):
    value = row.get(key)
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"{key} must be a nonempty string")
    return value.strip()


def rows(data, key, required=False):
    result = data.get(key)
    if not isinstance(result, list) or (required and not result):
        raise ValueError(f"{key} must be {'a nonempty' if required else 'a'} list")
    identifiers = set()
    for row in result:
        if not isinstance(row, dict):
            raise ValueError(f"{key} entries must be objects")
        identifier = required_text(row, "id")
        if identifier in identifiers:
            raise ValueError(f"duplicate {key} ID: {identifier}; reconcile first")
        identifiers.add(identifier)
    return result


def defect_grade(counts, systemic):
    if counts['critical']:
        return 'F'
    if counts['major'] >= 2 or systemic or (counts['major'] and counts['moderate'] >= 2):
        return 'C-'
    if counts['major']:
        return 'C'
    if counts['moderate'] >= 2:
        return 'C+'
    if counts['moderate']:
        return 'B-' if counts['minor'] >= 2 else 'B'
    if counts['minor'] >= 2:
        return 'B+'
    if counts['minor']:
        return 'A-'
    return None


def calculate(data):
    if not isinstance(data, dict):
        raise ValueError('ledger must be an object')
    required_text(data, 'scope')
    coverage = rows(data, 'coverage', required=True)
    findings = rows(data, 'findings')
    strengths = rows(data, 'strengths')
    safeguards = rows(data, 'safeguards')
    statuses = {'complete', 'not_applicable', 'pending', 'partial', 'unavailable'}
    for item in coverage:
        if not isinstance(item.get('status'), str) or item['status'] not in statuses:
            raise ValueError('invalid coverage status')
        required_text(item, 'reason')
    counts = Counter({s: 0 for s in ('critical', 'major', 'moderate', 'minor')})
    systemic = False
    for finding in findings:
        severity = finding.get('severity')
        if not isinstance(severity, str) or severity not in counts:
            raise ValueError('invalid finding severity')
        for key in ('cause', 'evidence', 'correction'):
            required_text(finding, key)
        flag = finding.get('systemic', False)
        if not isinstance(flag, bool) or (flag and severity != 'major'):
            raise ValueError('systemic must be boolean and applies only to major')
        if flag or severity == 'critical':
            required_text(finding, 'reach')
        systemic |= flag
        counts[severity] += 1
    for strength in strengths:
        required_text(strength, 'evidence')
    risks = set()
    for safeguard in safeguards:
        risk = required_text(safeguard, 'risk')
        required_text(safeguard, 'evidence')
        if not isinstance(safeguard.get('nonroutine'), bool):
            raise ValueError('safeguard nonroutine must be boolean')
        if safeguard['nonroutine']:
            risks.add(risk.casefold())
    applicable = [c for c in coverage if c['status'] != 'not_applicable']
    gaps = [c['id'] for c in applicable if c['status'] != 'complete']
    assessed = defect_grade(counts, systemic)
    if not applicable:
        if findings:
            raise ValueError('findings conflict with all coverage being not applicable')
        overall = 'Not applicable'
    elif gaps:
        overall = 'Insufficient evidence'
    elif assessed:
        overall = assessed
    elif not strengths:
        overall = 'Insufficient evidence'
    else:
        overall = 'A+' if len(risks) >= 2 else 'A'
    return {
        'overall': overall,
        'assessed_scope_grade': assessed if gaps else None,
        'unique_finding_counts': dict(counts),
        'coverage_gaps': gaps,
        'limits': 'Arithmetic only: coverage completeness, root-cause identity, severity, evidence and safeguard independence require reviewer verification.',
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('ledger', type=Path)
    args = parser.parse_args()
    try:
        result = calculate(json.loads(args.ledger.read_text()))
    except (OSError, ValueError) as exc:
        parser.exit(1, f'Invalid ledger: {exc}\n')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()
