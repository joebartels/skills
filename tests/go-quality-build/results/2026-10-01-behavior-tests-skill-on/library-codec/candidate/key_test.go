package keycodec_test

import (
	"errors"
	"fmt"
	"testing"

	"example.com/keycodec"
)

// Assignments protect the 1.x function signatures used by consumers.
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
		{"simple", keycodec.Key{Region: "east", Name: "web"}, "east/web"},
		{"slashes", keycodec.Key{Region: "us/west", Name: "web/blue"}, "us%2Fwest/web%2Fblue"},
		{"spaces", keycodec.Key{Region: "us west", Name: "web blue"}, "us%20west/web%20blue"},
		{"percent", keycodec.Key{Region: "us%west", Name: "web%blue"}, "us%25west/web%25blue"},
		{"escaped-looking values", keycodec.Key{Region: "%2F", Name: "%20"}, "%252F/%2520"},
		{"Unicode", keycodec.Key{Region: "日本", Name: "café"}, "%E6%97%A5%E6%9C%AC/caf%C3%A9"},
		{"literal plus", keycodec.Key{Region: "us+west", Name: "web+blue"}, "us+west/web+blue"},
		{"URL punctuation", keycodec.Key{Region: "us?west", Name: "web#blue"}, "us%3Fwest/web%23blue"},
		{"escaped fields only", keycodec.Key{Region: "/", Name: "%"}, "%2F/%25"},
		{"whitespace-only fields", keycodec.Key{Region: " ", Name: " "}, "%20/%20"},
		{"README example", keycodec.Key{Region: "us/west", Name: "web blue"}, "us%2Fwest/web%20blue"},
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

func TestEncodeRejectsEmptyFields(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  keycodec.Key
	}{
		{"zero value", keycodec.Key{}},
		{"empty region", keycodec.Key{Name: "web"}},
		{"empty name", keycodec.Key{Region: "east"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keycodec.Encode(tc.key)
			if !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Errorf("Encode(%#v) error = %v, want ErrInvalidKey", tc.key, err)
			}
			if got != "" {
				t.Errorf("Encode(%#v) result = %q, want empty result", tc.key, got)
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
		{"simple", "east/web", keycodec.Key{Region: "east", Name: "web"}},
		{"slashes", "us%2Fwest/web%2Fblue", keycodec.Key{Region: "us/west", Name: "web/blue"}},
		{"spaces", "us%20west/web%20blue", keycodec.Key{Region: "us west", Name: "web blue"}},
		{"percent", "us%25west/web%25blue", keycodec.Key{Region: "us%west", Name: "web%blue"}},
		{"one unescape", "%252F/%2520", keycodec.Key{Region: "%2F", Name: "%20"}},
		{"Unicode", "%E6%97%A5%E6%9C%AC/caf%C3%A9", keycodec.Key{Region: "日本", Name: "café"}},
		{"literal plus", "us+west/web+blue", keycodec.Key{Region: "us+west", Name: "web+blue"}},
		{"escaped plus", "us%2Bwest/web%2Bblue", keycodec.Key{Region: "us+west", Name: "web+blue"}},
		{"URL punctuation", "us%3Fwest/web%23blue", keycodec.Key{Region: "us?west", Name: "web#blue"}},
		{"lowercase escapes", "us%2fwest/caf%c3%a9", keycodec.Key{Region: "us/west", Name: "café"}},
		{"escaped fields only", "%2F/%25", keycodec.Key{Region: "/", Name: "%"}},
		{"whitespace-only fields", "%20/%20", keycodec.Key{Region: " ", Name: " "}},
		{"legacy unescaped values", "us west/café", keycodec.Key{Region: "us west", Name: "café"}},
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

func TestDecodeRejectsInvalidWire(t *testing.T) {
	for _, tc := range []struct {
		name string
		wire string
	}{
		{"empty", ""},
		{"no separator", "east"},
		{"only escaped separator", "east%2Fweb"},
		{"extra segment", "east/web/blue"},
		{"empty region", "/web"},
		{"empty name", "east/"},
		{"both empty", "/"},
		{"extra leading separator", "/east/web"},
		{"extra trailing separator", "east/web/"},
		{"incomplete region escape", "east%/web"},
		{"short region escape", "east%2/web"},
		{"nonhex region escape", "east%GG/web"},
		{"incomplete name escape", "east/web%"},
		{"short name escape", "east/web%2"},
		{"nonhex name escape", "east/web%G0"},
		{"nonhex second digit", "east/web%0G"},
		{"invalid name after valid region escape", "us%2Fwest/web%Q1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keycodec.Decode(tc.wire)
			if !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Errorf("Decode(%q) error = %v, want ErrInvalidKey", tc.wire, err)
			}
			if got != (keycodec.Key{}) {
				t.Errorf("Decode(%q) result = %#v, want zero Key", tc.wire, got)
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
	fmt.Printf("Region: %q, Name: %q\n", key.Region, key.Name)
	// Output: Region: "us/west", Name: "web blue"
}

func FuzzRoundTrip(f *testing.F) {
	f.Add("east", "web")
	f.Add("us/west", "web blue")
	f.Add("%2F", "%20")
	f.Add("日本", "café")
	f.Add("", "web")
	f.Add("east", "")
	f.Add("\x00\xff", "\x01+/%%")
	f.Fuzz(func(t *testing.T, region, name string) {
		want := keycodec.Key{Region: region, Name: name}
		wire, err := keycodec.Encode(want)
		if region == "" || name == "" {
			if !errors.Is(err, keycodec.ErrInvalidKey) || wire != "" {
				t.Fatalf("Encode(%#v) = %q, %v, want empty result and ErrInvalidKey", want, wire, err)
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
