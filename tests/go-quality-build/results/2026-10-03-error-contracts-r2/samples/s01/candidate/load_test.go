package recordload_test

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"example.com/recordload"
)

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		want        []recordload.Record
	}{
		{"records in order", "a=1\nb=2\n", []recordload.Record{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}},
		{"comments and blanks", "# header\n\na=1\n# ignored=too\nb=2\n# last", []recordload.Record{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}},
		{"empty input", "", nil},
		{"only comments and blanks", "\n# comment\n\n", nil},
		{"empty fields", "=\nkey=\n=value", []recordload.Record{{}, {Key: "key"}, {Value: "value"}}},
		{"literal bytes", " a =1=2 # value\n #key=value\nkey=\r", []recordload.Record{{Key: " a ", Value: "1=2 # value"}, {Key: " #key", Value: "value"}, {Key: "key", Value: "\r"}}},
		{"long line", "key=" + strings.Repeat("x", 128*1024), []recordload.Record{{Key: "key", Value: strings.Repeat("x", 128*1024)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := recordload.Load(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Load error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Load records = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestLoadReadFailure(t *testing.T) {
	cause := errors.New("read failed")
	for _, tc := range []struct {
		name, input string
		readErr     error
		want        []recordload.Record
	}{
		{"before bytes", "", cause, nil},
		{"alongside bytes", "a=1\nb=2", cause, []recordload.Record{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}},
		{"after bytes", "a=1\nb=2\n", nil, []recordload.Record{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &failingReader{data: tc.input, readErr: tc.readErr, finalErr: cause}
			got, err := recordload.Load(r)
			if !errors.Is(err, cause) {
				t.Errorf("Load error = %v, want cause %v", err, cause)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Load records = %#v, want %#v", got, tc.want)
			}
			if r.closed {
				t.Error("Load closed the borrowed reader")
			}
		})
	}
}

func TestLoadBytesWithEOF(t *testing.T) {
	r := &failingReader{data: "a=1", readErr: io.EOF, finalErr: io.EOF}
	got, err := recordload.Load(r)
	if err != nil {
		t.Fatalf("Load error = %v, want nil", err)
	}
	if want := []recordload.Record{{Key: "a", Value: "1"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Load records = %#v, want %#v", got, want)
	}
	if r.closed {
		t.Error("Load closed the borrowed reader")
	}
}

func TestLoadMalformedWithReadFailure(t *testing.T) {
	cause := errors.New("read failed")
	r := &failingReader{data: "a=1\n\n# comment\nmalformed\nb=2", readErr: cause, finalErr: cause}
	got, err := recordload.Load(r)
	if !errors.Is(err, cause) {
		t.Errorf("Load error = %v, want cause %v", err, cause)
	}
	var lineErr *recordload.LineError
	if !errors.As(err, &lineErr) {
		t.Fatalf("Load error = %v, want *LineError", err)
	}
	if lineErr.Line != 4 || lineErr.Text != "malformed" {
		t.Errorf("LineError = %#v, want Line 4 and Text %q", lineErr, "malformed")
	}
	if r.closed {
		t.Error("Load closed the borrowed reader")
	}
	if want := []recordload.Record{{Key: "a", Value: "1"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Load records = %#v, want valid prefix %#v", got, want)
	}
}

func TestLoadLineError(t *testing.T) {
	for _, tc := range []struct {
		name, input, text string
		line              int
		want              []recordload.Record
	}{
		{"first line", "malformed\na=1", "malformed", 1, nil},
		{"physical line", "\n# comment\na=1\n\nbad\nb=2", "bad", 5, []recordload.Record{{Key: "a", Value: "1"}}},
		{"spaced comment is not a comment", " # comment", " # comment", 1, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := recordload.Load(strings.NewReader(tc.input))
			var lineErr *recordload.LineError
			if !errors.As(err, &lineErr) {
				t.Fatalf("Load error = %v, want *LineError", err)
			}
			if lineErr.Line != tc.line || lineErr.Text != tc.text {
				t.Errorf("LineError = %#v, want Line %d and Text %q", lineErr, tc.line, tc.text)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Load records = %#v, want valid prefix %#v", got, tc.want)
			}
		})
	}
}

func TestLoadTypedReadFailure(t *testing.T) {
	for _, input := range []string{"a=1", "a=1\nmalformed"} {
		t.Run(input, func(t *testing.T) {
			cause := readFailure{"offline"}
			r := &failingReader{data: input, readErr: fmt.Errorf("source: %w", cause)}
			got, err := recordload.Load(r)
			var readErr readFailure
			if !errors.As(err, &readErr) || !reflect.DeepEqual(readErr, cause) {
				t.Errorf("Load error = %v, want typed reader cause %#v", err, cause)
			}
			if want := []recordload.Record{{Key: "a", Value: "1"}}; !reflect.DeepEqual(got, want) {
				t.Errorf("Load records = %#v, want %#v", got, want)
			}
		})
	}
}

func TestLoadMalformedWithEOF(t *testing.T) {
	r := &failingReader{data: "a=1\nmalformed", readErr: io.EOF}
	got, err := recordload.Load(r)
	var lineErr *recordload.LineError
	if !errors.As(err, &lineErr) || lineErr.Line != 2 || lineErr.Text != "malformed" {
		t.Fatalf("Load error = %v, want LineError at line 2", err)
	}
	if errors.Is(err, io.EOF) {
		t.Error("Load exposed EOF as a failure")
	}
	if want := []recordload.Record{{Key: "a", Value: "1"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Load records = %#v, want %#v", got, want)
	}
}

func ExampleLoad() {
	records, err := recordload.Load(strings.NewReader("# settings\na=1\nmalformed\n"))
	fmt.Println(records)
	var lineErr *recordload.LineError
	if errors.As(err, &lineErr) {
		fmt.Printf("line %d: %q\n", lineErr.Line, lineErr.Text)
	}
	// Output:
	// [{a 1}]
	// line 3: "malformed"
}

type readFailure []string

func (e readFailure) Error() string { return strings.Join(e, ": ") }

type failingReader struct {
	data              string
	readErr, finalErr error
	closed            bool
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.data == "" {
		return 0, r.finalErr
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	if r.data == "" {
		return n, r.readErr
	}
	return n, nil
}

func (r *failingReader) Close() error {
	r.closed = true
	return nil
}
