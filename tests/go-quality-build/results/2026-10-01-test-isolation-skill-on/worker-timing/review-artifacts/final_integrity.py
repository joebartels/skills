import hashlib
import json

import run_checks as runner

runner.checks = json.loads((runner.OUT / 'checks.json').read_text())
runner.check('final-source-integrity', ['rtk','proxy','python3','-c', 'import hashlib,pathlib; b=pathlib.Path("/private/tmp/go-independent-review-4qorzxhe"); [(print(str(p.relative_to(b)),hashlib.sha256(p.read_bytes()).hexdigest())) for d in ("original","candidate") for p in sorted((b/d).iterdir()) if p.is_file()]; assert all((b/"original"/f).read_bytes()==(b/"candidate"/f).read_bytes() for f in ("poll.go","README.md","go.mod"))'], runner.BASE)
