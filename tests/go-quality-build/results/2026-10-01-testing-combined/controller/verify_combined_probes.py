from pathlib import Path
import shutil,subprocess,tempfile,os,json,time,hashlib
root=Path.cwd();suite=root/'tests/go-quality-build/testing-combined/evals';run=root/'tests/go-quality-build/results/2026-10-01-testing-combined'
work=Path(tempfile.mkdtemp(prefix='indexer-controller-green-',dir='/private/tmp'));shutil.copytree(suite/'files/indexer-evolution',work,dirs_exist_ok=True)
p=work/'index.go';s=p.read_text().replace('"errors"','"errors"\n"encoding/csv"\n"strings"');s=s.replace('return ErrNotImplemented','''rdr:=csv.NewReader(r);rdr.FieldsPerRecord=2
 for { row,err:=rdr.Read();if err==io.EOF{return nil};if err!=nil{return err};if !keyPattern.MatchString(row[0])||strings.TrimSpace(row[1])==""{return ErrInvalidRecord};if err:=i.Put(row[0],row[1]);err!=nil{return err} }''');p.write_text(s)
p=work/'refresh.go';s=p.read_text().replace('"net/http"','"net/http"\n"encoding/json"\n"fmt"\n"io"\n"os"\n"path/filepath"\n"strings"');s=s.replace('return ErrNotImplemented','''req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil);if err!=nil{return err};resp,err:=client.Do(req);if err!=nil{return err};defer resp.Body.Close();if resp.StatusCode!=200{return fmt.Errorf("status %d",resp.StatusCode)}
 var records []Record;dec:=json.NewDecoder(resp.Body);if err:=dec.Decode(&records);err!=nil{return err};if records==nil{return ErrInvalidRecord};var extra any;if err:=dec.Decode(&extra);err!=io.EOF{return fmt.Errorf("trailing data: %v",err)}
 for _,r:=range records{if !keyPattern.MatchString(r.Key)||strings.TrimSpace(r.Text)==""{return ErrInvalidRecord}}
 b,err:=json.Marshal(records);if err!=nil{return err};b=append(b,'\\n');f,err:=os.CreateTemp(filepath.Dir(path),".snapshot-*");if err!=nil{return err};defer os.Remove(f.Name());if _,err:=f.Write(b);err!=nil{f.Close();return err};if err:=f.Close();err!=nil{return err};return os.Rename(f.Name(),path)''');p.write_text(s)
p=work/'serve.go';s=p.read_text().replace(') error {',') (out error) {').replace('return ErrNotImplemented','''if interval<=0{return ErrInvalidInterval};defer func(){out=errors.Join(out,release())}()
 for { if ctx.Err()!=nil{return nil};if err:=refresh(ctx);err!=nil{return err};timer:=time.NewTimer(interval);select{case <-ctx.Done():timer.Stop();return nil;case <-timer.C:} }''');p.write_text(s)
p=work/'cmd/indexer/main.go';s=p.read_text().replace('"fmt"','"fmt"\n"context"\n"flag"\n"net/http"\n"net/url"\n"os/signal"\n"syscall"\n"time"');pos=s.index('func run(');end=s.index('\nfunc main',pos);s=s[:pos]+'''func run(args []string) error {
 if len(args)==0{return fmt.Errorf("usage")}
 switch args[0]{
 case "put":if len(args)!=4{return fmt.Errorf("usage")};return indexer.Open(args[1]).Put(args[2],args[3])
 case "apply":fs:=flag.NewFlagSet("apply",flag.ContinueOnError);dir:=fs.String("dir","","");if err:=fs.Parse(args[1:]);err!=nil{return err};if *dir==""||fs.NArg()!=0{return fmt.Errorf("dir required")};return indexer.Open(*dir).ApplyCSV(os.Stdin)
 case "watch":fs:=flag.NewFlagSet("watch",flag.ContinueOnError);endpoint:=fs.String("url","","");path:=fs.String("file","","");interval:=fs.Duration("interval",time.Second,"");if err:=fs.Parse(args[1:]);err!=nil{return err};u,err:=url.Parse(*endpoint);if err!=nil||u.Host==""||(u.Scheme!="http"&&u.Scheme!="https")||*path==""||*interval<=0||fs.NArg()!=0{return fmt.Errorf("invalid watch arguments")};ctx,cancel:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer cancel();client:=&http.Client{};return indexer.Serve(ctx,*interval,func(c context.Context)error{return indexer.Refresh(c,client,*endpoint,*path)},func()error{client.CloseIdleConnections();return nil})
 default:return fmt.Errorf("unknown command")}
}
'''+s[end:];p.write_text(s)
shutil.copy2(suite/'controller-probes/indexer-evolution/contract_test.go',work/'controller_contract_test.go')
subprocess.run(['rtk','proxy','gofmt','-w','.'],cwd=work,check=True)
checks=[]
for cmd in [['go','test','-count=1','-timeout=30s','./...'],['go','test','-race','-shuffle=42','-count=10','-timeout=30s','./...'],['go','vet','./...']]:
 start=time.monotonic();r=subprocess.run(['rtk','proxy',*cmd],cwd=work,env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local'),capture_output=True,text=True,timeout=55);checks.append(dict(command=['rtk','proxy',*cmd],cwd=str(work),env={'GOCACHE':'/private/tmp/go-quality-testing-cache','GOTOOLCHAIN':'local'},stdout=r.stdout,stderr=r.stderr,exit_code=r.returncode,duration_seconds=time.monotonic()-start));print(r.returncode,r.stdout+r.stderr)
shutil.copytree(work,run/'preparation-reference',dirs_exist_ok=True)
(run/'preparation-probe-GREEN.json').write_text(json.dumps(dict(scope='Original controller-only preparation proof; never supplied to authors; not a blind candidate, assisted repair or skill uplift',checks=checks,source_sha256={str(p.relative_to(work)):hashlib.sha256(p.read_bytes()).hexdigest() for p in work.rglob('*') if p.is_file()}),indent=2)+'\n')
assert all(c['exit_code']==0 for c in checks)
