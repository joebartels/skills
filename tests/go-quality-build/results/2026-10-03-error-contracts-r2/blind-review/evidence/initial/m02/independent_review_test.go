package paging
import "testing"
func TestIndependentReviewPagingDomain(t *testing.T){
 for n:=0;n<=1000;n++{for size:=1;size<=101;size++{want:=0;for left:=n;left>0;left-=size{want++};if got:=pages(n,size);got!=want{t.Fatalf("pages(%d,%d)=%d want=%d",n,size,got,want)}}}
 max:=int(^uint(0)>>1)
 for _,tc:=range []struct{n,s,want int}{{max,1,max},{max,2,max/2+1},{max,max,1},{max,max-1,2},{0,max,0}}{if got:=pages(tc.n,tc.s);got!=tc.want{t.Fatalf("boundary %#v got=%d",tc,got)}}
}
