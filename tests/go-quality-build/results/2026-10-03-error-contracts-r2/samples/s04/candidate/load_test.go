package recordload

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	got, err := Load(strings.NewReader("a=1\nb=2\n"))
	if err != nil || len(got) != 2 || got[1].Key != "b" {
		t.Fatalf("%v %v", got, err)
	}
}

func TestLoadRecords(t *testing.T) {
	longValue := strings.Repeat("x", 70*1024)
	for _, tc := range []struct {
		name  string
		input string
		want  []Record
	}{
		{
			name:  "comments blanks and order",
			input: "#=ignored\n\na=1\n#ignored\nb=2=3\n=empty-key\nempty-value=\n=\na=again",
			want: []Record{
				{Key: "a", Value: "1"},
				{Key: "b", Value: "2=3"},
				{Value: "empty-key"},
				{Key: "empty-value"},
				{},
				{Key: "a", Value: "again"},
			},
		},
		{name: "empty"},
		{name: "only comments and blanks", input: "#comment\n\n#last"},
		{name: "unterminated line", input: "a=1", want: []Record{{Key: "a", Value: "1"}}},
		{name: "whitespace preserved", input: " a = b ", want: []Record{{Key: " a ", Value: " b "}}},
		{name: "long line", input: "a=" + longValue, want: []Record{{Key: "a", Value: longValue}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Load(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("Load() records = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadReadError(t *testing.T) {
	readErr := errors.New("reader failed")
	for _, tc := range []struct {
		name     string
		input    string
		terminal error
		withData bool
		want     []Record
		wantErr  error
	}{
		{name: "before data", terminal: readErr, wantErr: readErr},
		{
			name: "with data", input: "a=1\nb=2", terminal: readErr, withData: true,
			want: []Record{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}, wantErr: readErr,
		},
		{
			name: "after data", input: "a=1\nb=2\n", terminal: readErr,
			want: []Record{{Key: "a", Value: "1"}, {Key: "b", Value: "2"}}, wantErr: readErr,
		},
		{
			name: "EOF with data", input: "a=1", terminal: io.EOF, withData: true,
			want: []Record{{Key: "a", Value: "1"}},
		},
		{name: "empty EOF", terminal: io.EOF},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &terminalReader{reader: strings.NewReader(tc.input), err: tc.terminal, withData: tc.withData}
			got, err := Load(r)
			if tc.wantErr == nil {
				if err != nil {
					t.Errorf("Load() error = %v, want nil", err)
				}
			} else if !errors.Is(err, tc.wantErr) {
				t.Errorf("Load() error = %v, want cause %v", err, tc.wantErr)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("Load() records = %v, want %v", got, tc.want)
			}
			if r.closed {
				t.Error("Load() closed the borrowed reader")
			}
		})
	}
}

func TestLoadLineError(t *testing.T) {
	readErr := errors.New("reader failed")
	for _, tc := range []struct {
		name     string
		input    string
		line     int
		text     string
		want     []Record
		terminal error
		withData bool
	}{
		{name: "first line", input: "bad\nlater=2", line: 1, text: "bad"},
		{
			name: "physical line and raw text", input: "#comment\n\na=1\n bad \nlater=2",
			line: 4, text: " bad ", want: []Record{{Key: "a", Value: "1"}},
		},
		{name: "indented hash", input: " #comment", line: 1, text: " #comment"},
		{name: "whitespace line", input: " \t", line: 1, text: " \t"},
		{
			name: "EOF with malformed data", input: "a=1\nbad", line: 2, text: "bad",
			want: []Record{{Key: "a", Value: "1"}}, terminal: io.EOF, withData: true,
		},
		{
			name: "reader error with malformed data", input: "a=1\nbad\nlater=2", line: 2, text: "bad",
			want: []Record{{Key: "a", Value: "1"}}, terminal: readErr, withData: true,
		},
		{
			name: "reader error after malformed data", input: "a=1\nbad\nlater=2", line: 2, text: "bad",
			want: []Record{{Key: "a", Value: "1"}}, terminal: readErr,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			terminal := tc.terminal
			if terminal == nil {
				terminal = io.EOF
			}
			r := &terminalReader{reader: strings.NewReader(tc.input), err: terminal, withData: tc.withData}
			got, err := Load(r)
			if !slices.Equal(got, tc.want) {
				t.Errorf("Load() records = %v, want %v", got, tc.want)
			}
			var lineErr *LineError
			if !errors.As(err, &lineErr) {
				t.Fatalf("Load() error = %v, want *LineError", err)
			}
			if lineErr.Line != tc.line || lineErr.Text != tc.text {
				t.Errorf("LineError = %+v, want Line %d, Text %q", lineErr, tc.line, tc.text)
			}
			if tc.terminal != nil && tc.terminal != io.EOF && !errors.Is(err, tc.terminal) {
				t.Errorf("Load() error = %v, want reader cause %v", err, tc.terminal)
			}
			if errors.Is(err, io.EOF) {
				t.Error("Load() exposed EOF as a failure")
			}
			if r.closed {
				t.Error("Load() closed the borrowed reader")
			}
		})
	}
}

type terminalReader struct {
	reader   *strings.Reader
	err      error
	withData bool
	closed   bool
}

func (r *terminalReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if r.reader.Len() == 0 && (r.withData || err != nil) {
		return n, r.err
	}
	return n, err
}

func (r *terminalReader) Close() error {
	r.closed = true
	return nil
}
