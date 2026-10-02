import json
from pathlib import Path

root = Path('/private/tmp/go-fresh-author-8hjrllxp/task-3/report')
checks_path = root / 'checks.json'
checks = json.loads(checks_path.read_text())
for index, check in enumerate(checks):
    if 'execution_profile' not in check:
        check['execution_profile'] = 'approved_local_http' if index == 4 else 'sandbox_default'
checks[3]['stage'] = 'expanded tests before cleanup refinement; restricted listener failure'
checks[4]['stage'] = 'expanded tests before cleanup refinement; approved unchanged-source rerun'
checks[4]['unchanged_source_rerun_of_check_index'] = 3
checks_path.write_text(json.dumps(checks, indent=2) + '\n')
selection = json.loads((root / 'selection.json').read_text())
assert len(selection['offered_names']) == 5
assert len(selection['opened_skill_paths']) == 2
assert len(selection['opened_reference_paths']) == 1
assert (root / 'author-report.md').stat().st_size > 0
assert all(check['command'][0:2] == ['rtk', 'proxy'] for check in checks)
assert all(check['cwd'] == '/private/tmp/go-fresh-author-8hjrllxp/task-3/module' for check in checks)
assert any(check['exit_code'] == 1 and 'operation not permitted' in check['stdout'] for check in checks)
assert any(check.get('unchanged_source_rerun_of_check_index') == 3 and check['exit_code'] == 0 for check in checks)
print(f'Validated selection.json, author-report.md, and {len(checks)} preserved command checks.')
