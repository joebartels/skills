from pathlib import Path
import subprocess
import sys

path = Path("main.go")
original = path.read_text()
needle = "func writeRecord(dir, id, content string) error {\n"
assert original.count(needle) == 1
try:
    path.write_text(original.replace(needle, needle + '\treturn os.WriteFile(filepath.Join(dir, id), []byte(content), 0o644)\n'))
    command = ["rtk", "proxy", "go", "test", "-timeout=30s", "-run", "^TestPartialWriteRetainsAcceptedReplacement$", "./..."]
    print("Temporary direct-write mutation; inner command: " + repr(command), flush=True)
    result = subprocess.run(command, timeout=45)
finally:
    path.write_text(original)
sys.exit(result.returncode)
