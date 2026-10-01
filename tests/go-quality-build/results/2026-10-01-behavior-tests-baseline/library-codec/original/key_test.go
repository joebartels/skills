package keycodec

import "testing"

func TestSimpleRoundTrip(t *testing.T) {
	want := Key{Region: "east", Name: "web"}
	wire, err := Encode(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(wire)
	if err != nil || got != want {
		t.Fatalf("round trip = %#v, %v, want %#v", got, err, want)
	}
}
