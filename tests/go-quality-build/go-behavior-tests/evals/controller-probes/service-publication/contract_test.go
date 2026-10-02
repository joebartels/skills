package mirror_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"example.com/mirror"
)

type probeTransport func(*http.Request) (*http.Response, error)

func (f probeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func probeClient(body string) *http.Client {
	return &http.Client{Transport: probeTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
}

func TestFieldValidationAndRetention(t *testing.T) {
	for _, body := range []string{`[{"code":" ","label":"good"}]`, `[{"code":"c","label":"\t"}]`, `[{"code":"c","label":"good"}] {}`, "not JSON"} {
		path := filepath.Join(t.TempDir(), "items.json")
		old := []byte(`[{"code":"old","label":"Retained"}]`)
		if err := os.WriteFile(path, old, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := mirror.Refresh(context.Background(), probeClient(body), "http://example.test/items", path); err == nil {
			t.Errorf("Refresh accepted %q", body)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != string(old) {
			t.Errorf("retained bytes = %q, %v", got, err)
		}
	}
}

func TestCompletePublication(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("pre-opened handle replacement probe specifies Unix semantics")
	}
	path := filepath.Join(t.TempDir(), "items.json")
	old := `[{"code":"old","label":"Retained"}]`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	handle, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()
	body := `[{"code":"new","label":"New label"},{"code":"two","label":"Second"}]`
	if err := mirror.Refresh(context.Background(), probeClient(body), "http://example.test/items", path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got []mirror.Item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	want := []mirror.Item{{Code: "new", Label: "New label"}, {Code: "two", Label: "Second"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("published = %#v; want %#v", got, want)
	}
	retained, err := io.ReadAll(handle)
	if err != nil || string(retained) != old {
		t.Errorf("old handle = %q, %v; want unchanged old inode", retained, err)
	}
}
