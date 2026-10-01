package keycodec_test

import (
	"errors"
	"testing"

	codec "example.com/keycodec"
)

func TestWireRepresentationIndependent(t *testing.T) {
	k := codec.Key{Region: "us/west", Name: "web blue"}
	got, err := codec.Encode(k)
	if err != nil || got != "us%2Fwest/web%20blue" {
		t.Fatalf("Encode = %q, %v; want fixed wire us%%2Fwest/web%%20blue", got, err)
	}
	decoded, err := codec.Decode("eu%2Fnorth/name%25one")
	if err != nil || decoded != (codec.Key{Region: "eu/north", Name: "name%one"}) {
		t.Fatalf("Decode = %#v, %v; want independently known fields", decoded, err)
	}
	for _, wire := range []string{"", "one", "/two", "one/", "a/b/c", "%/b", "a/%zz"} {
		if _, err := codec.Decode(wire); !errors.Is(err, codec.ErrInvalidKey) {
			t.Errorf("Decode(%q) error = %v; want ErrInvalidKey", wire, err)
		}
	}
	for _, k := range []codec.Key{{Region: "r"}, {Name: "n"}, {}} {
		if _, err := codec.Encode(k); !errors.Is(err, codec.ErrInvalidKey) {
			t.Errorf("Encode(%#v) error = %v; want ErrInvalidKey", k, err)
		}
	}
}
