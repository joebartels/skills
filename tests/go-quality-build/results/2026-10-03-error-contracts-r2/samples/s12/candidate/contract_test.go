package entryzip_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"math/rand"
	"strings"
	"testing"

	"example.com/entryzip"
)

func TestWriteArchiveErrors(t *testing.T) {
	testArchiveErrors(t, entryzip.WriteArchive)
}

func TestWriteSelectedErrors(t *testing.T) {
	testArchiveErrors(t, func(w io.Writer, entries []entryzip.Entry) error {
		return entryzip.WriteSelected(w, entries, nil)
	})
}

func TestWriteSelected(t *testing.T) {
	var output observedWriter
	skipped := &observedReader{reader: strings.NewReader("unread"), err: errors.New("must not read")}
	entries := []entryzip.Entry{
		{Name: "b.txt", Body: strings.NewReader("bravo")},
		{Name: strings.Repeat("x", 1<<16), Body: skipped},
		{Name: "empty.txt", Body: strings.NewReader("")},
		{Name: "b.txt", Body: strings.NewReader("second")},
	}
	var names []string
	err := entryzip.WriteSelected(&output, entries, func(name string) bool {
		names = append(names, name)
		return name == "b.txt" || name == "empty.txt"
	})
	if err != nil {
		t.Fatal(err)
	}
	checkArchive(t, output.Bytes(), []archiveEntry{{"b.txt", "bravo"}, {"empty.txt", ""}, {"b.txt", "second"}})
	if len(names) != len(entries) {
		t.Fatalf("predicate called %d times, want %d", len(names), len(entries))
	}
	for i, name := range names {
		if name != entries[i].Name {
			t.Errorf("predicate name %d = %q, want %q", i, name, entries[i].Name)
		}
	}
	if skipped.reads != 0 || skipped.closes != 0 || output.closes != 0 {
		t.Error("filtered entry was read or a borrowed resource was closed")
	}
}

func TestArchiveDefaults(t *testing.T) {
	// These assignments also check source compatibility at a consumer boundary.
	var legacy func(io.Writer, []entryzip.Entry) error = entryzip.WriteArchive
	var selected func(io.Writer, []entryzip.Entry, func(string) bool) error = entryzip.WriteSelected
	for _, tc := range []struct {
		name  string
		write func(io.Writer, []entryzip.Entry) error
	}{
		{"legacy", legacy},
		{"nil predicate", func(w io.Writer, entries []entryzip.Entry) error { return selected(w, entries, nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := tc.write(&output, []entryzip.Entry{
				{Name: "z.txt", Body: strings.NewReader("last")},
				{Name: "a.txt", Body: strings.NewReader("first")},
			}); err != nil {
				t.Fatal(err)
			}
			checkArchive(t, output.Bytes(), []archiveEntry{{"z.txt", "last"}, {"a.txt", "first"}})
			output.Reset()
			if err := tc.write(&output, nil); err != nil {
				t.Fatal(err)
			}
			checkArchive(t, output.Bytes(), nil)
		})
	}
}

func TestWriteSelectedRejectsAll(t *testing.T) {
	var output bytes.Buffer
	if err := entryzip.WriteSelected(&output, []entryzip.Entry{{Name: "skip", Body: nil}}, func(string) bool { return false }); err != nil {
		t.Fatal(err)
	}
	checkArchive(t, output.Bytes(), nil)
}

func TestNonComparableBodyError(t *testing.T) {
	for _, failOutput := range []bool{false, true} {
		name := "body only"
		if failOutput {
			name = "body and completion"
		}
		t.Run(name, func(t *testing.T) {
			want := opaqueError{details: []string{"body failure"}}
			outputErr := errors.New("completion failure")
			var output observedWriter
			if failOutput {
				output.err = outputErr
			}
			body := &observedReader{reader: strings.NewReader("partial"), err: want}
			got := entryzip.WriteSelected(&output, []entryzip.Entry{{Name: "data", Body: body}}, func(string) bool { return true })
			var cause opaqueError
			if !errors.As(got, &cause) || &cause.details[0] != &want.details[0] {
				t.Fatalf("write = %v, want original typed body cause", got)
			}
			if failOutput {
				if !errors.Is(got, outputErr) {
					t.Errorf("write = %v, want completion cause", got)
				}
			} else if _, ok := got.(opaqueError); !ok {
				t.Errorf("write = %T, want direct body error", got)
			}
		})
	}
}

func TestOutputFailureDuringCopy(t *testing.T) {
	data := make([]byte, 256<<10)
	if _, err := rand.New(rand.NewSource(1)).Read(data); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		write func(io.Writer, []entryzip.Entry) error
	}{
		{"legacy", entryzip.WriteArchive},
		{"selected", func(w io.Writer, entries []entryzip.Entry) error { return entryzip.WriteSelected(w, entries, nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want := opaqueError{details: []string{"copy output failure"}}
			output := &observedWriter{err: want}
			body := &observedReader{reader: bytes.NewReader(data)}
			later := &observedReader{reader: strings.NewReader("later")}
			got := tc.write(output, []entryzip.Entry{{Name: "data", Body: body}, {Name: "later", Body: later}})
			if direct, ok := got.(opaqueError); !ok || &direct.details[0] != &want.details[0] {
				t.Errorf("write = %T %v, want original copy output error", got, got)
			}
			if body.bytesRead == 0 || body.bytesRead >= len(data) {
				t.Errorf("read %d bytes, want failure during body processing", body.bytesRead)
			}
			if later.reads != 0 || body.closes != 0 || later.closes != 0 || output.closes != 0 {
				t.Error("write used a later body or closed a borrowed resource")
			}
		})
	}
}

func testArchiveErrors(t *testing.T, write func(io.Writer, []entryzip.Entry) error) {
	t.Helper()
	bodyErr := errors.New("body failure")
	outputErr := errors.New("output failure")
	for _, tc := range []struct {
		name               string
		bodyErr, outputErr error
	}{
		{name: "success"},
		{name: "body", bodyErr: bodyErr},
		{name: "completion", outputErr: outputErr},
		{name: "both", bodyErr: bodyErr, outputErr: outputErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &observedReader{reader: strings.NewReader("prefix"), err: tc.bodyErr}
			later := &observedReader{reader: strings.NewReader("later")}
			output := &observedWriter{err: tc.outputErr}
			got := write(output, []entryzip.Entry{
				{Name: "first.txt", Body: body},
				{Name: "later.txt", Body: later},
			})
			switch {
			case tc.bodyErr == nil && tc.outputErr == nil:
				if got != nil {
					t.Fatalf("write = %v, want nil", got)
				}
			case tc.bodyErr == nil:
				if got != tc.outputErr {
					t.Errorf("write = %v, want original output error", got)
				}
			case tc.outputErr == nil:
				if got != tc.bodyErr {
					t.Errorf("write = %v, want original body error", got)
				}
			default:
				if !errors.Is(got, tc.bodyErr) || !errors.Is(got, tc.outputErr) {
					t.Errorf("write = %v, want both causes", got)
				}
			}
			if output.writes != 1 {
				t.Errorf("output writes = %d, want one buffered archive flush", output.writes)
			}
			if body.closes != 0 || later.closes != 0 || output.closes != 0 {
				t.Error("write closed a borrowed resource")
			}
			if tc.bodyErr != nil && later.reads != 0 {
				t.Error("later entry was read after body failure")
			}
			if tc.outputErr == nil {
				want := []archiveEntry{{"first.txt", "prefix"}, {"later.txt", "later"}}
				if tc.bodyErr != nil {
					want = want[:1]
				}
				checkArchive(t, output.Bytes(), want)
			}
		})
	}
}

func TestWriteArchiveCreationFailure(t *testing.T) {
	var output observedWriter
	body := &observedReader{reader: strings.NewReader("unread")}
	later := &observedReader{reader: strings.NewReader("later")}
	err := entryzip.WriteArchive(&output, []entryzip.Entry{
		{Name: strings.Repeat("x", 1<<16), Body: body},
		{Name: "later.txt", Body: later},
	})
	if err == nil {
		t.Fatal("write succeeded with an unrepresentable ZIP name")
	}
	if body.reads != 0 || later.reads != 0 {
		t.Error("entry bodies read after creation failure")
	}
	if output.writes == 0 {
		t.Error("archive completion did not write to the output after creation failure")
	}
	if body.closes != 0 || later.closes != 0 || output.closes != 0 {
		t.Error("write closed a borrowed resource")
	}
}

func TestWriteArchiveOutputErrorIdentity(t *testing.T) {
	// A valid long name flushes the ZIP buffer during entry creation.
	want := opaqueError{details: []string{"output failure"}}
	output := &observedWriter{err: want}
	body := &observedReader{reader: strings.NewReader("unread")}
	got := entryzip.WriteArchive(output, []entryzip.Entry{{Name: strings.Repeat("x", 5000), Body: body}})
	if direct, ok := got.(opaqueError); !ok || &direct.details[0] != &want.details[0] {
		t.Errorf("write = %T %v, want original non-comparable output error", got, got)
	}
	if body.reads != 0 || body.closes != 0 || output.closes != 0 {
		t.Error("creation failure used or closed a borrowed body/output")
	}
}

type opaqueError struct {
	details []string
}

func (e opaqueError) Error() string { return strings.Join(e.details, ": ") }

type observedReader struct {
	reader                   io.Reader
	err                      error
	reads, closes, bytesRead int
}

func (r *observedReader) Read(p []byte) (int, error) {
	r.reads++
	n, err := r.reader.Read(p)
	r.bytesRead += n
	if err == io.EOF && r.err != nil {
		return n, r.err
	}
	return n, err
}

func (r *observedReader) Close() error {
	r.closes++
	return nil
}

type observedWriter struct {
	bytes.Buffer
	err            error
	writes, closes int
}

func (w *observedWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.err != nil {
		return 0, w.err
	}
	return w.Buffer.Write(p)
}

func (w *observedWriter) Close() error {
	w.closes++
	return nil
}

type archiveEntry struct {
	name, body string
}

func checkArchive(t *testing.T, data []byte, want []archiveEntry) {
	t.Helper()
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	if len(archive.File) != len(want) {
		t.Fatalf("archive has %d entries, want %d", len(archive.File), len(want))
	}
	for i, file := range archive.File {
		if file.Name != want[i].name {
			t.Errorf("entry %d name = %q, want %q", i, file.Name, want[i].name)
		}
		body, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		got, readErr := io.ReadAll(body)
		closeErr := body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("entry %d: read = %v, close = %v", i, readErr, closeErr)
		}
		if string(got) != want[i].body {
			t.Errorf("entry %d body = %q, want %q", i, got, want[i].body)
		}
	}
}
