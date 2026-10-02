package signedcounter
import("testing";"bytes")
func TestReviewOwnedKey(t *testing.T){key:=[]byte("sixteen-byte-key!");c,e:=New(key);if e!=nil {t.Fatal(e)};before:=c.Sign([]byte("x"));key[0]='X';after:=c.Sign([]byte("x"));if !bytes.Equal(before,after)||c.Count()!=2 {t.Fatal("key alias or count mismatch")}}
func TestReviewZeroCountObservation(t *testing.T){defer func(){t.Logf("zero Count panic=%v",recover())}();var c Counter;t.Logf("zero Count=%d",c.Count())}
