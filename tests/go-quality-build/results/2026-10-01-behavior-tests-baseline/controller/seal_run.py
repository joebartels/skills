from pathlib import Path
import hashlib,json,sys

root=Path.cwd()
run=root/"tests/go-quality-build/results"/sys.argv[1]
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def hashes(folder): return {str(p.relative_to(folder)):sha(p) for p in sorted(folder.rglob("*")) if p.is_file()}
manifest=json.loads((run/"manifest.json").read_text())
for case in manifest["cases"]:
 arc=run/case["id"]
 assert hashes(arc/"original")==case["input_sha256"],case["id"]
 assert hashes(arc/"candidate")==case["source_sha256"],case["id"]
 assert sha(arc/"dispatch.txt")==case["dispatch_sha256"],case["id"]
 for filename in ["prompt.txt","dispatch.txt","launch.txt","selection.json","author-report.md","source.patch","verification.json","blind-review.md"]:
  assert (arc/filename).exists(),(case["id"],filename)
 case["artifact_sha256"]=hashes(arc)
manifest["stage"]="completed blind baselines, controller verification and independent outcome reviews; no writing-skill uplift claimed"
manifest["review_settings"]={"harness":"fresh Codex collaboration contexts, fork_turns=none","model":"inherited session model; exact ID not exposed","reasoning":"inherited setting; no override","blinding":"anonymized original/candidate code and task; no arm label, writing skills, expected judgments, author reports, other reviews or controller probes"}
manifest["artifact_sha256"]={str(p.relative_to(run)):sha(p) for p in sorted(run.rglob("*")) if p.is_file() and p.name not in {"manifest.json","checksums.sha256"}}
manifest["checksum_note"]="checksums.sha256 covers manifest.json and every other archived file except itself; its digest is anchored in the canonical design record."
(run/"manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
files=[p for p in sorted(run.rglob("*")) if p.is_file() and p.name!="checksums.sha256"]
index=run/"checksums.sha256"
index.write_text("".join(sha(p)+"  "+str(p.relative_to(run))+"\n" for p in files))
for line in index.read_text().splitlines():
 digest,name=line.split("  ",1);assert sha(run/name)==digest,name
print(f"PASS {len(files)} artifact SHA-256 hashes; index {sha(index)}")
