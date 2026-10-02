package keycodec_test

import (
	"errors"
	"fmt"
	"testing"

	"example.com/keycodec"
)

// These assignments also check the existing API from a consumer package.
var (
	_ func(keycodec.Key) (string, error) = keycodec.Encode
	_ func(string) (keycodec.Key, error) = keycodec.Decode
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		key  keycodec.Key
		wire string
	}{
		{"legacy", keycodec.Key{Region: "east", Name: "web"}, "east/web"},
		{"region slash", keycodec.Key{Region: "us/west", Name: "web"}, "us%2Fwest/web"},
		{"name slash", keycodec.Key{Region: "east", Name: "web/blue"}, "east/web%2Fblue"},
		{"both slashes", keycodec.Key{Region: "/", Name: "/"}, "%2F/%2F"},
		{"region space", keycodec.Key{Region: "us west", Name: "web"}, "us%20west/web"},
		{"name space", keycodec.Key{Region: "east", Name: "web blue"}, "east/web%20blue"},
		{"percent", keycodec.Key{Region: "100%", Name: "%2F"}, "100%25/%252F"},
		{"Unicode", keycodec.Key{Region: "日本", Name: "café"}, "%E6%97%A5%E6%9C%AC/caf%C3%A9"},
		{"path plus", keycodec.Key{Region: "a+b", Name: "c+d"}, "a+b/c+d"},
		{"reserved", keycodec.Key{Region: "a?b", Name: "c#d"}, "a%3Fb/c%23d"},
		{"combined", keycodec.Key{Region: "us/west", Name: "web blue"}, "us%2Fwest/web%20blue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keycodec.Encode(tt.key)
			if err != nil || got != tt.wire {
				t.Fatalf("Encode(%#v) = %q, %v; want %q, nil", tt.key, got, err, tt.wire)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name string
		wire string
		key  keycodec.Key
	}{
		{"legacy", "east/web", keycodec.Key{Region: "east", Name: "web"}},
		{"region slash", "us%2Fwest/web", keycodec.Key{Region: "us/west", Name: "web"}},
		{"name slash", "east/web%2Fblue", keycodec.Key{Region: "east", Name: "web/blue"}},
		{"both slashes", "%2F/%2F", keycodec.Key{Region: "/", Name: "/"}},
		{"spaces", "%20west/web%20", keycodec.Key{Region: " west", Name: "web "}},
		{"percent", "100%25/%252F", keycodec.Key{Region: "100%", Name: "%2F"}},
		{"Unicode", "%E6%97%A5%E6%9C%AC/caf%C3%A9", keycodec.Key{Region: "日本", Name: "café"}},
		{"literal plus", "a+b/c+d", keycodec.Key{Region: "a+b", Name: "c+d"}},
		{"escaped plus", "a%2Bb/c%2Bd", keycodec.Key{Region: "a+b", Name: "c+d"}},
		{"lowercase hex", "us%2fwest/caf%c3%a9", keycodec.Key{Region: "us/west", Name: "café"}},
		{"escaped unreserved", "%65ast/%77eb", keycodec.Key{Region: "east", Name: "web"}},
		{"reserved", "a%3Fb/c%23d", keycodec.Key{Region: "a?b", Name: "c#d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keycodec.Decode(tt.wire)
			if err != nil || got != tt.key {
				t.Fatalf("Decode(%q) = %#v, %v; want %#v, nil", tt.wire, got, err, tt.key)
			}
		})
	}
}

func TestEncodeInvalidKey(t *testing.T) {
	for _, key := range []keycodec.Key{{}, {Region: "east"}, {Name: "web"}} {
		t.Run(fmt.Sprintf("%q/%q", key.Region, key.Name), func(t *testing.T) {
			got, err := keycodec.Encode(key)
			if got != "" || !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Fatalf("Encode(%#v) = %q, %v; want empty result and ErrInvalidKey", key, got, err)
			}
		})
	}
}

func TestDecodeInvalidKey(t *testing.T) {
	wires := []string{
		"", "east", "/", "/web", "east/", "east/web/blue", "/east/web", "east/web/",
		"east//web", "%/web", "east/%", "%2/web", "east/%2", "%GG/web", "east/%GG",
		"%2X/web", "east/%X2", "valid%20region/bad%", "bad%/valid%20name",
	}
	for _, wire := range wires {
		t.Run(wire, func(t *testing.T) {
			got, err := keycodec.Decode(wire)
			if got != (keycodec.Key{}) || !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Fatalf("Decode(%q) = %#v, %v; want zero Key and ErrInvalidKey", wire, got, err)
			}
		})
	}
}

func Example() {
	key := keycodec.Key{Region: "us/west", Name: "web blue"}
	wire, err := keycodec.Encode(key)
	if err != nil {
		panic(err)
	}
	fmt.Println(wire)
	decoded, err := keycodec.Decode(wire)
	if err != nil {
		panic(err)
	}
	fmt.Printf("region=%q name=%q\n", decoded.Region, decoded.Name)
	// Output:
	// us%2Fwest/web%20blue
	// region="us/west" name="web blue"
}

func FuzzRoundTrip(f *testing.F) {
	for _, key := range []keycodec.Key{
		{Region: "east", Name: "web"},
		{Region: "us/west", Name: "web blue"},
		{Region: "日本%", Name: "café+/%2F"},
		{},
	} {
		f.Add(key.Region, key.Name)
	}
	f.Fuzz(func(t *testing.T, region, name string) {
		key := keycodec.Key{Region: region, Name: name}
		wire, err := keycodec.Encode(key)
		if region == "" || name == "" {
			if wire != "" || !errors.Is(err, keycodec.ErrInvalidKey) {
				t.Fatalf("Encode(%#v) = %q, %v; want empty result and ErrInvalidKey", key, wire, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("Encode(%#v): %v", key, err)
		}
		got, err := keycodec.Decode(wire)
		if err != nil || got != key {
			t.Fatalf("Decode(Encode(%#v)) = %#v, %v", key, got, err)
		}
	})
}
