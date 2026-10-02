package keycodec_test

import (
	"errors"
	"fmt"
	"testing"

	"example.com/keycodec"
)

// Consumers may use function values as well as ordinary calls.
var (
	_ func(keycodec.Key) (string, error) = keycodec.Encode
	_ func(string) (keycodec.Key, error) = keycodec.Decode
)

func TestEncode(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  keycodec.Key
		wire string
	}{
		{"existing simple key", keycodec.Key{Region: "east", Name: "web"}, "east/web"},
		{"slash in region", keycodec.Key{Region: "us/west", Name: "web"}, "us%2Fwest/web"},
		{"slash in name", keycodec.Key{Region: "east", Name: "web/api"}, "east/web%2Fapi"},
		{"spaces in both fields", keycodec.Key{Region: "us west", Name: "web blue"}, "us%20west/web%20blue"},
		{"percent in both fields", keycodec.Key{Region: "%2F", Name: "100%"}, "%252F/100%25"},
		{"unicode in both fields", keycodec.Key{Region: "東京", Name: "café"}, "%E6%9D%B1%E4%BA%AC/caf%C3%A9"},
		{"escaped delimiter in both fields", keycodec.Key{Region: "/", Name: "/"}, "%2F/%2F"},
		{"literal plus", keycodec.Key{Region: "us+west", Name: "web+blue"}, "us+west/web+blue"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keycodec.Encode(tc.key)
			if err != nil {
				t.Fatalf("Encode(%#v): %v", tc.key, err)
			}
			if got != tc.wire {
				t.Errorf("Encode(%#v) = %q, want %q", tc.key, got, tc.wire)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tc := range []struct {
		name string
		wire string
		key  keycodec.Key
	}{
		{"existing simple key", "east/web", keycodec.Key{Region: "east", Name: "web"}},
		{"slash in region", "us%2Fwest/web", keycodec.Key{Region: "us/west", Name: "web"}},
		{"slash in name", "east/web%2Fapi", keycodec.Key{Region: "east", Name: "web/api"}},
		{"spaces in both fields", "us%20west/web%20blue", keycodec.Key{Region: "us west", Name: "web blue"}},
		{"percent in both fields", "%252F/100%25", keycodec.Key{Region: "%2F", Name: "100%"}},
		{"unicode in both fields", "%E6%9D%B1%E4%BA%AC/caf%C3%A9", keycodec.Key{Region: "東京", Name: "café"}},
		{"escaped delimiter in both fields", "%2F/%2F", keycodec.Key{Region: "/", Name: "/"}},
		{"lowercase escapes", "us%2fwest/caf%c3%a9", keycodec.Key{Region: "us/west", Name: "café"}},
		{"literal plus", "us+west/web+blue", keycodec.Key{Region: "us+west", Name: "web+blue"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keycodec.Decode(tc.wire)
			if err != nil {
				t.Fatalf("Decode(%q): %v", tc.wire, err)
			}
			if got != tc.key {
				t.Errorf("Decode(%q) = %#v, want %#v", tc.wire, got, tc.key)
			}
		})
	}
}

func TestEncodeRejectsEmptyFields(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  keycodec.Key
	}{
		{"both empty", keycodec.Key{}},
		{"empty region", keycodec.Key{Name: "web"}},
		{"empty name", keycodec.Key{Region: "east"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keycodec.Encode(tc.key)
			if !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Errorf("Encode(%#v) error = %v, want ErrInvalidKey", tc.key, err)
			}
			if got != "" {
				t.Errorf("Encode(%#v) = %q on failure, want empty string", tc.key, got)
			}
		})
	}
}

func TestDecodeRejectsInvalidWire(t *testing.T) {
	for _, tc := range []struct {
		name string
		wire string
	}{
		{"empty wire", ""},
		{"no separator", "east"},
		{"escaped separator only", "east%2Fweb"},
		{"three segments", "east/web/api"},
		{"extra trailing separator", "east/web/"},
		{"extra leading separator", "/east/web"},
		{"both empty", "/"},
		{"empty region", "/web"},
		{"empty name", "east/"},
		{"bare percent in region", "%/web"},
		{"bare percent in name", "east/%"},
		{"short escape in region", "east%2/web"},
		{"short escape in name", "east/web%2"},
		{"invalid hex in region", "east%GG/web"},
		{"invalid hex in name", "east/web%GG"},
		{"invalid second hex digit", "east/web%2G"},
		{"valid escape followed by malformed escape", "east/web%20%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keycodec.Decode(tc.wire)
			if !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Errorf("Decode(%q) error = %v, want ErrInvalidKey", tc.wire, err)
			}
			if got != (keycodec.Key{}) {
				t.Errorf("Decode(%q) = %#v on failure, want zero Key", tc.wire, got)
			}
		})
	}
}

func TestSimpleRoundTrip(t *testing.T) {
	want := keycodec.Key{Region: "east", Name: "web"}
	wire, err := keycodec.Encode(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := keycodec.Decode(wire)
	if err != nil || got != want {
		t.Fatalf("round trip = %#v, %v, want %#v", got, err, want)
	}
}

func FuzzRoundTrip(f *testing.F) {
	for _, key := range []keycodec.Key{
		{Region: "east", Name: "web"},
		{Region: "us/west", Name: "web blue"},
		{Region: "%2F", Name: "東京+café"},
		{},
	} {
		f.Add(key.Region, key.Name)
	}
	f.Fuzz(func(t *testing.T, region, name string) {
		want := keycodec.Key{Region: region, Name: name}
		wire, err := keycodec.Encode(want)
		if region == "" || name == "" {
			if !errors.Is(err, keycodec.ErrInvalidKey) || wire != "" {
				t.Fatalf("Encode(%#v) = %q, %v, want empty string and ErrInvalidKey", want, wire, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("Encode(%#v): %v", want, err)
		}
		got, err := keycodec.Decode(wire)
		if err != nil || got != want {
			t.Fatalf("Decode(Encode(%#v)) = %#v, %v", want, got, err)
		}
	})
}

func ExampleEncode() {
	wire, err := keycodec.Encode(keycodec.Key{Region: "us/west", Name: "web blue"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(wire)
	// Output: us%2Fwest/web%20blue
}

func ExampleDecode() {
	key, err := keycodec.Decode("us%2Fwest/web%20blue")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("region=%q name=%q\n", key.Region, key.Name)
	// Output: region="us/west" name="web blue"
}
