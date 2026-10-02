import os, subprocess
from pathlib import Path
source = Path("index.go")
original = source.read_text()
try:
 source.write_text(original.replace("func publish(path string, data []byte) error {", "func publish(path string, data []byte) error {\n return os.WriteFile(path, data, 0600)"))
 environment = dict(os.environ, CHECK_PURPOSE="temporary mutation: directly truncate destination; expect meaningful retained-bytes assertion failure")
 result = subprocess.run(["rtk","proxy","python3","/private/tmp/go-fresh-author-qjdzm0iq/task-2/report/run_check.py","rtk","proxy","go","test","-timeout=20s","-run","^TestPublicationWriteFailure$/^put$","."],env=environment)
 if result.returncode != 1: raise SystemExit("mutation did not produce the expected test failure")
finally:
 source.write_text(original)
print("Restored exact original source after controlled mutation.")
