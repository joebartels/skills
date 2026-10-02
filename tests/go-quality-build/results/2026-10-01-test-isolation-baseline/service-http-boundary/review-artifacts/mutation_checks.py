from pathlib import Path
import subprocess,json,shutil,hashlib
root=Path('/private/tmp/go-independent-review-bq9orwps');out=root/'output'
checks=json.loads((out/'checks.json').read_text())
original_hash={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (root/'candidate').iterdir() if p.is_file()}
mutations=[
 ('omit-body-close','defer response.Body.Close()','// mutation: omit close','TestFetchSuccessClosesBody|TestFetchInvalidJSONOrFieldsClosesBody|TestFetchNon200StatusClosesBody|TestFetchReadFailureClosesBody'),
 ('accept-trailing-json','json.Unmarshal(data, &result)','json.NewDecoder(strings.NewReader(string(data))).Decode(&result)','TestFetchInvalidJSONOrFieldsClosesBody'),
 ('ignore-read-error','data, err := io.ReadAll(response.Body)\n\tif err != nil {','data, err := io.ReadAll(response.Body)\n\tif err != nil && len(data) == 0 {','TestFetchReadFailureClosesBody'),
 ('return-partial-metadata','return Metadata{}, fmt.Errorf("invalid metadata")','return result, fmt.Errorf("invalid metadata")','TestFetchInvalidJSONOrFieldsClosesBody'),
 ('accept-any-2xx','response.StatusCode != http.StatusOK','response.StatusCode < 200 || response.StatusCode >= 300','TestFetchNon200StatusClosesBody'),
 ('trim-name','return result, nil','result.Name = strings.TrimSpace(result.Name)\n\treturn result, nil','TestFetchSuccessClosesBody'),
 ('drop-caller-context','http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)','http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)','TestFetchCancellation'),
 ('drop-query','response, err := client.Do(req)','req.URL.RawQuery = ""\n\tresponse, err := client.Do(req)','TestFetchHTTPRequest'),
 ('wrong-method','http.MethodGet','http.MethodPost','TestFetchHTTPRequest'),
]
for label,before,after,pattern in mutations:
 cwd=root/'mutations'/label
 cwd.parent.mkdir(exist_ok=True)
 shutil.copytree(root/'candidate',cwd,dirs_exist_ok=True)
 path=cwd/'fetch.go';source=path.read_text()
 assert source.count(before)==1,(label,source.count(before))
 path.write_text(source.replace(before,after,1))
 command=['rtk','proxy','env','GOCACHE=/private/tmp/go-quality-testing-cache','GOTOOLCHAIN=local','go','test','-count=1','-timeout=20s','-run',pattern,'./...']
 try:
  p=subprocess.run(command,cwd=cwd,capture_output=True,text=True,timeout=25)
  entry={'label':'mutation-'+label,'sandbox_permissions':'require_escalated','mutation':{'file':'fetch.go','before':before,'after':after},'command':command,'cwd':str(cwd),'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
 except subprocess.TimeoutExpired as e:
  entry={'label':'mutation-'+label,'sandbox_permissions':'require_escalated','mutation':{'file':'fetch.go','before':before,'after':after},'command':command,'cwd':str(cwd),'exit':None,'timeout_seconds':25,'stdout':(e.stdout or b'').decode() if isinstance(e.stdout,bytes) else e.stdout or '', 'stderr':(e.stderr or b'').decode() if isinstance(e.stderr,bytes) else e.stderr or ''}
 checks.append(entry);(out/'checks.json').write_text(json.dumps(checks,indent=2)+'\n')
 print(json.dumps(entry),flush=True)
final_hash={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (root/'candidate').iterdir() if p.is_file()}
assert original_hash==final_hash,'candidate source changed'
checks.append({'label':'candidate-preserved','before_sha256':original_hash,'after_sha256':final_hash,'identical':True})
(out/'checks.json').write_text(json.dumps(checks,indent=2)+'\n')
print('Candidate source unchanged.',flush=True)
