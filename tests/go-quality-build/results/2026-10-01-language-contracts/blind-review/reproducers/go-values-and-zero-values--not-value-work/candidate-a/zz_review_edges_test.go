package paging
import "testing"
func TestReviewIntegerBoundary(t *testing.T){m:=int(^uint(0)>>1);if pages(m,1)!=m||pages(m,2)!=m/2+1||pages(m,m)!=1 {t.Fatal("integer boundary mismatch")}}
