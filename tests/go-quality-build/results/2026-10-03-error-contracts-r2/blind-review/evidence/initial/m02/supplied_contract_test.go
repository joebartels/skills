package paging
import "testing"
func TestContractPages(t *testing.T){for _,tc:=range []struct{n,size,want int}{{0,3,0},{6,3,2},{7,3,3},{1,1,1}}{if got:=pages(tc.n,tc.size);got!=tc.want{t.Fatalf("%v got %d",tc,got)}}}
