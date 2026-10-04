package entryzip

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type contractBody struct {
	cause error
	done bool
	closed int
}

func (r *contractBody) Read(p []byte) (int, error) {
	if r.done { return 0, io.EOF }
	r.done = true
	return copy(p, "payload"), r.cause
}
func (r *contractBody) Close() error { r.closed++; return nil }

type contractOutput struct {
	bytes.Buffer
	cause error
	closed int
}

func (w *contractOutput) Write(p []byte) (int, error) {
	if w.cause != nil { return 0, w.cause }
	return w.Buffer.Write(p)
}
func (w *contractOutput) Close() error { w.closed++; return nil }

func TestContractArchiveCompletion(t *testing.T) {
	readErr := errors.New("source failed")
	finishErr := errors.New("destination failed")
	for _, tc := range []struct{name string; read, finish error}{
		{"success", nil, nil}, {"read", readErr, nil},
		{"completion", nil, finishErr}, {"both", readErr, finishErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &contractBody{cause: tc.read}
			output := &contractOutput{cause: tc.finish}
			err := WriteSelected(output, []Entry{{Name: "a.txt", Body: body}}, nil)
			if body.closed != 0 || output.closed != 0 {
				t.Fatalf("closed borrowed resources: body=%d output=%d", body.closed, output.closed)
			}
			switch {
			case tc.read != nil && tc.finish != nil:
				if !errors.Is(err, readErr) || !errors.Is(err, finishErr) { t.Fatalf("lost cause: %v", err) }
			case tc.read != nil:
				if err != readErr { t.Fatalf("lone read identity: %v", err) }
			case tc.finish != nil:
				if err != finishErr { t.Fatalf("lone completion identity: %v", err) }
			default:
				if err != nil { t.Fatal(err) }
				if _, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len())); err != nil { t.Fatalf("incomplete successful archive: %v", err) }
			}
		})
	}
}

func TestContractArchiveSelection(t *testing.T) {
	var output bytes.Buffer
	err := WriteSelected(&output, []Entry{
		{Name: "first.txt", Body: strings.NewReader("one")},
		{Name: "skip.txt", Body: nil},
		{Name: "last.txt", Body: strings.NewReader("three")},
	}, func(name string) bool { return name != "skip.txt" })
	if err != nil { t.Fatal(err) }
	reader, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil { t.Fatal(err) }
	if len(reader.File) != 2 || reader.File[0].Name != "first.txt" || reader.File[1].Name != "last.txt" { t.Fatalf("selected entries: %v", reader.File) }
	for i, want := range []string{"one", "three"} {
		body, err := reader.File[i].Open()
		if err != nil { t.Fatal(err) }
		got, err := io.ReadAll(body)
		closeErr := body.Close()
		if err != nil || closeErr != nil || string(got) != want { t.Fatalf("body=%q read=%v close=%v", got, err, closeErr) }
	}
}
