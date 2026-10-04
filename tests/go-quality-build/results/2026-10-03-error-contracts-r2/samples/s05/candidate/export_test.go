package lineexport

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type bufferCloser struct{ bytes.Buffer }

func (*bufferCloser) Close() error { return nil }
func TestExport(t *testing.T) {
	w := new(bufferCloser)
	n, err := Export(strings.NewReader("a\nb\n"), w)
	if n != 2 || err != nil || w.String() != "a\nb\n" {
		t.Fatalf("%d %v %q", n, err, w.String())
	}
}

type closeError struct{}

func (*closeError) Error() string { return "finalization failed" }

type recordReader struct {
	records    []string
	err        error
	reads      int
	closes     int
	beforeRead func() error
}

func (r *recordReader) Read(p []byte) (int, error) {
	if r.beforeRead != nil {
		if err := r.beforeRead(); err != nil {
			return 0, err
		}
	}
	r.reads++
	if len(r.records) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		return 0, io.EOF
	}
	n := copy(p, r.records[0])
	r.records[0] = r.records[0][n:]
	if r.records[0] == "" {
		r.records = r.records[1:]
	}
	return n, nil
}

func (r *recordReader) Close() error {
	r.closes++
	return nil
}

type recordingWriter struct {
	buffer   bytes.Buffer
	writeErr error
	closeErr error
	failAt   int
	writes   int
	closes   int
}

func (w *recordingWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		n, _ := w.buffer.Write(p[:1])
		return n, w.writeErr
	}
	return w.buffer.Write(p)
}

func (w *recordingWriter) Close() error {
	w.closes++
	return w.closeErr
}

func TestExportErrors(t *testing.T) {
	readErr := errors.New("read failed")
	writeErr := errors.New("write failed")
	closeErr := new(closeError)
	for _, tc := range []struct {
		name      string
		readErr   error
		writeErr  error
		closeErr  error
		wantCount int
		wantBytes string
		wantReads int
	}{
		{"success", nil, nil, nil, 2, "a\nb\n", 3},
		{"read only", readErr, nil, nil, 2, "a\nb\n", 3},
		{"close only", nil, nil, closeErr, 2, "a\nb\n", 3},
		{"read and close", readErr, nil, closeErr, 2, "a\nb\n", 3},
		{"write only", nil, writeErr, nil, 1, "a\nb", 2},
		{"write and close", nil, writeErr, closeErr, 1, "a\nb", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &recordReader{records: []string{"a\n", "b\n"}, err: tc.readErr}
			w := &recordingWriter{writeErr: tc.writeErr, closeErr: tc.closeErr}
			if tc.writeErr != nil {
				w.failAt = 2
			}
			n, err := Export(r, w)
			if n != tc.wantCount || w.buffer.String() != tc.wantBytes {
				t.Errorf("Export = %d, %q; want %d, %q", n, w.buffer.String(), tc.wantCount, tc.wantBytes)
			}
			opErr := tc.readErr
			if tc.writeErr != nil {
				opErr = tc.writeErr
			}
			switch {
			case opErr == nil:
				if err != tc.closeErr {
					t.Errorf("error = %v, want identical %v", err, tc.closeErr)
				}
			case tc.closeErr == nil:
				if err != opErr {
					t.Errorf("error = %v, want identical %v", err, opErr)
				}
			default:
				if !errors.Is(err, opErr) || !errors.Is(err, tc.closeErr) {
					t.Errorf("error = %v, want both %v and %v", err, opErr, tc.closeErr)
				}
			}
			if tc.closeErr != nil {
				var cause *closeError
				if !errors.As(err, &cause) || cause != closeErr {
					t.Errorf("error = %v, want typed finalization cause", err)
				}
			}
			if w.closes != 1 || r.closes != 0 || r.reads != tc.wantReads {
				t.Errorf("writer closes = %d, reader closes = %d, reads = %d; want 1, 0, %d", w.closes, r.closes, r.reads, tc.wantReads)
			}
		})
	}
}

func TestExportLimit(t *testing.T) {
	var export func(io.Reader, io.WriteCloser) (int, error) = Export
	var exportLimit func(io.Reader, io.WriteCloser, int) (int, error) = ExportLimit
	for _, tc := range []struct {
		name      string
		input     string
		limit     int
		wantCount int
		wantBytes string
	}{
		{"empty", "", 0, 0, ""},
		{"unlimited", "a\n\nlast", 0, 3, "a\n\nlast\n"},
		{"first", "a\nb\n", 1, 1, "a\n"},
		{"empty record", "\na\n", 1, 1, "\n"},
		{"final record", "last", 1, 1, "last\n"},
		{"exact", "a\nb\n", 2, 2, "a\nb\n"},
		{"above count", "a\nb\n", 3, 2, "a\nb\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := new(recordingWriter)
			n, err := exportLimit(strings.NewReader(tc.input), w, tc.limit)
			if n != tc.wantCount || err != nil || w.buffer.String() != tc.wantBytes {
				t.Errorf("ExportLimit = %d, %v, %q; want %d, nil, %q", n, err, w.buffer.String(), tc.wantCount, tc.wantBytes)
			}
			if w.closes != 1 {
				t.Errorf("writer closes = %d, want 1", w.closes)
			}
			if tc.limit == 0 {
				unlimitedWriter := new(recordingWriter)
				count, err := export(strings.NewReader(tc.input), unlimitedWriter)
				if count != tc.wantCount || err != nil || unlimitedWriter.buffer.String() != tc.wantBytes {
					t.Errorf("Export = %d, %v, %q; want %d, nil, %q", count, err, unlimitedWriter.buffer.String(), tc.wantCount, tc.wantBytes)
				}
			}
		})
	}
}

func TestExportLimitStopsBeforeNextRead(t *testing.T) {
	readErr := errors.New("reading beyond the limit")
	for _, closeErr := range []error{nil, new(closeError)} {
		r := &recordReader{records: []string{"a\n"}, err: readErr}
		w := &recordingWriter{closeErr: closeErr}
		n, err := ExportLimit(r, w, 1)
		if n != 1 || err != closeErr || w.buffer.String() != "a\n" {
			t.Errorf("ExportLimit = %d, %v, %q; want 1, %v, %q", n, err, w.buffer.String(), closeErr, "a\n")
		}
		if r.reads != 1 || w.writes != 1 || w.closes != 1 || r.closes != 0 {
			t.Errorf("reads = %d, writes = %d, writer closes = %d, reader closes = %d; want 1, 1, 1, 0", r.reads, w.writes, w.closes, r.closes)
		}
	}
}

func TestExportLimitRejectsNegative(t *testing.T) {
	for _, closeErr := range []error{nil, new(closeError)} {
		r := &recordReader{records: []string{"a\n"}}
		w := &recordingWriter{closeErr: closeErr}
		n, err := ExportLimit(r, w, -1)
		if n != 0 || err == nil || w.buffer.Len() != 0 {
			t.Errorf("ExportLimit = %d, %v, %q; want 0, error, empty output", n, err, w.buffer.String())
		}
		if closeErr != nil {
			joined, ok := err.(interface{ Unwrap() []error })
			if !ok || len(joined.Unwrap()) != 2 || !errors.Is(err, closeErr) {
				t.Errorf("error = %v, want validation and finalization causes", err)
			}
		}
		if r.reads != 0 || w.writes != 0 || w.closes != 1 || r.closes != 0 {
			t.Errorf("reads = %d, writes = %d, writer closes = %d, reader closes = %d; want 0, 0, 1, 0", r.reads, w.writes, w.closes, r.closes)
		}
	}
}

func TestExportShortWrite(t *testing.T) {
	for _, closeErr := range []error{nil, new(closeError)} {
		r := &recordReader{records: []string{"a\n", "b\n", "c\n"}}
		w := &recordingWriter{failAt: 2, closeErr: closeErr}
		n, err := Export(r, w)
		if n != 1 || w.buffer.String() != "a\nb" {
			t.Errorf("Export = %d, %q; want 1, %q", n, w.buffer.String(), "a\nb")
		}
		if closeErr == nil && err != io.ErrShortWrite {
			t.Errorf("error = %v, want identical %v", err, io.ErrShortWrite)
		}
		if closeErr != nil && (!errors.Is(err, io.ErrShortWrite) || !errors.Is(err, closeErr)) {
			t.Errorf("error = %v, want short-write and finalization causes", err)
		}
		if r.reads != 2 || w.writes != 2 || w.closes != 1 {
			t.Errorf("reads = %d, writes = %d, closes = %d; want 2, 2, 1", r.reads, w.writes, w.closes)
		}
	}
}

func TestExportWritesBeforeReadingNextRecord(t *testing.T) {
	w := new(recordingWriter)
	r := &recordReader{records: []string{"a\n", "b\n", "c\n"}}
	r.beforeRead = func() error {
		if r.reads != w.writes {
			return errors.New("read before writing the previous record")
		}
		return nil
	}
	n, err := Export(r, w)
	if n != 3 || err != nil || w.buffer.String() != "a\nb\nc\n" {
		t.Errorf("Export = %d, %v, %q; want 3, nil, %q", n, err, w.buffer.String(), "a\nb\nc\n")
	}
}

type dataErrorReader struct{ err error }

func (r dataErrorReader) Read(p []byte) (int, error) {
	return copy(p, "a\nb\n"), r.err
}

func TestExportLimitReadError(t *testing.T) {
	readErr := errors.New("read failed with data")
	for _, tc := range []struct {
		limit int
		count int
		bytes string
		err   error
	}{
		{0, 2, "a\nb\n", readErr},
		{1, 1, "a\n", nil},
		{2, 2, "a\nb\n", nil},
		{3, 2, "a\nb\n", readErr},
	} {
		w := new(recordingWriter)
		n, err := ExportLimit(dataErrorReader{readErr}, w, tc.limit)
		if n != tc.count || err != tc.err || w.buffer.String() != tc.bytes || w.closes != 1 {
			t.Errorf("limit %d: ExportLimit = %d, %v, %q, %d closes; want %d, %v, %q, 1 close", tc.limit, n, err, w.buffer.String(), w.closes, tc.count, tc.err, tc.bytes)
		}
	}
}
