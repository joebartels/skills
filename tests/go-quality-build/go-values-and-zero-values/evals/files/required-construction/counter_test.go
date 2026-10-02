package signedcounter

import "testing"

func TestNew(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("accepted missing key")
	}
	c, err := New([]byte("sixteen-byte-key!"))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Sign(nil)) != 32 {
		t.Fatal("invalid signature")
	}
}
