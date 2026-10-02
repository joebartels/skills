package draftreview
import ("encoding/json";"errors";"maps";"slices";"testing")
func TestCompletion(t *testing.T) {
 a,b:=errors.New("primary"),errors.New("finish")
 for _,tc:=range []struct{a,b,want error}{{nil,nil,nil},{a,nil,a},{nil,b,b}} {if got:=completionError(tc.a,tc.b);got!=tc.want {t.Fatalf("identity: %v",got)}}
 both:=completionError(a,b);if !errors.Is(both,a)||!errors.Is(both,b)||errors.Unwrap(both)!=nil {t.Fatal("tree semantics")}
 if got:=errors.Join(a,nil);got==a||!errors.Is(got,a) {t.Fatal("join identity")}
}
func TestTypedNil(t *testing.T) {err:=typedNil();if err==nil {t.Fatal("interface unexpectedly nil")};var p *Problem;if !errors.As(err,&p)||p!=nil {t.Fatal("typed inspection")}}
func TestCapacity(t *testing.T) {back:=[]int{1,2,3};view:=back[:1:1];view[0]=7;if back[0]!=7 {t.Fatal("element did not alias")};grown:=append(view,8);grown[0]=9;if back[0]!=7||back[1]!=2 {t.Fatal("append reused backing array")}}
func TestShallowClones(t *testing.T) {in:=[]struct{Labels map[string]string}{{map[string]string{"x":"old"}}};cp:=slices.Clone(in);cp[0].Labels["x"]="new";if in[0].Labels["x"]!="new" {t.Fatal("nested map did not alias")};m:=map[string][]byte{"x":[]byte("old")};mc:=maps.Clone(m);mc["x"][0]='n';if string(m["x"])!="nld" {t.Fatal("nested bytes did not alias")}}
func TestEmptyCopies(t *testing.T) {in:=make([]byte,0);if append([]byte(nil),in...)!=nil {t.Fatal("append preserved empty")};if slices.Clone(in)==nil {t.Fatal("clone collapsed empty")};if slices.Clone([]byte(nil))!=nil||maps.Clone(map[string]int(nil))!=nil {t.Fatal("clone lost nil")}}
func TestJSON(t *testing.T) {type payload struct{Items []string `json:"items"`};for _,tc:=range []struct{p payload;want string}{{payload{nil},`{"items":null}`},{payload{[]string{}},`{"items":[]}`}} {b,e:=json.Marshal(tc.p);if e!=nil||string(b)!=tc.want {t.Fatalf("%s %v",b,e)}};b,_:=json.Marshal(struct{Items []string `json:"items,omitempty"`}{[]string{}});if string(b)!="{}" {t.Fatal(string(b))};b,_=json.Marshal([]byte("x"));if string(b)!=`"eA=="` {t.Fatal(string(b))}}
func TestNilMap(t *testing.T) {var m map[string]int;if len(m)!=0||m["x"]!=0 {t.Fatal("read")};delete(m,"x");defer func(){if recover()==nil {t.Fatal("nil assignment did not panic")}}();m["x"]=1}
type Item struct{}
func (*Item) Change() {}
func TestAddressableCall(t *testing.T) {var v Item;v.Change();var _ interface{Change()}=&v}

func TestCapacityBoundary(t *testing.T) {
 back:=[]int{1,2,3,4};partial:=back[:1:2];grown:=append(partial,8);if &grown[0]!=&back[0]||back[1]!=8 {t.Fatal("partial restriction should retain append capacity")}
 full:=back[:1:1];same:=append(full,[]int{}...);if &same[0]!=&back[0] {t.Fatal("empty append unexpectedly detached")}
}
