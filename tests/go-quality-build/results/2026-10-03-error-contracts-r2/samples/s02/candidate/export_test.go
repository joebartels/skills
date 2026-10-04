package lineexport_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"example.com/lineexport"
)

var _ func(io.Reader, io.WriteCloser) (int, error) = lineexport.Export
var _ func(io.Reader, io.WriteCloser, int) (int, error) = lineexport.ExportLimit

type testWriter struct {
	output   bytes.Buffer
	closes   int
	closeErr error
	writes   int
	failAt   int
	writeErr error
	partial  int
}

func (w *testWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		_, _ = w.output.Write(p[:w.partial])
		return w.partial, w.writeErr
	}
	return w.output.Write(p)
}

func (w *testWriter) Close() error {
	w.closes++
	return w.closeErr
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

type readFailure []string

func (readFailure) Error() string { return "read failed" }

type closeFailure struct{}

func (*closeFailure) Error() string { return "close failed" }

type recordReader struct {
	records []string
	reads   int
	closes  int
	check   func() error
}

func (r *recordReader) Read(p []byte) (int, error) {
	if r.check != nil {
		if err := r.check(); err != nil {
			return 0, err
		}
	}
	r.reads++
	if len(r.records) == 0 {
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

func TestExportRecords(t *testing.T) {
	for _, tc := range []struct {
		name, input, output string
		count               int
	}{
		{"empty input", "", "", 0},
		{"terminated records", "a\nb\n", "a\nb\n", 2},
		{"unterminated record", "a\nb", "a\nb\n", 2},
		{"empty records", "\na\n\n", "\na\n\n", 3},
		{"CRLF records", "a\r\nb\r\n", "a\nb\n", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := new(testWriter)
			n, err := lineexport.Export(strings.NewReader(tc.input), w)
			if n != tc.count || err != nil || w.output.String() != tc.output {
				t.Fatalf("Export = (%d, %v), output %q; want (%d, nil), output %q", n, err, w.output.String(), tc.count, tc.output)
			}
			if w.closes != 1 {
				t.Errorf("Close calls = %d, want 1", w.closes)
			}
		})
	}
}

func TestExportErrors(t *testing.T) {
	readErr := errors.New("read failed")
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")
	for _, tc := range []struct {
		name              string
		readErr, writeErr error
		closeErr          error
		count             int
		output            string
	}{
		{"success", nil, nil, nil, 2, "a\nb\n"},
		{"close only", nil, nil, closeErr, 2, "a\nb\n"},
		{"read only", readErr, nil, nil, 1, "a\n"},
		{"read and close", readErr, nil, closeErr, 1, "a\n"},
		{"write only", nil, writeErr, nil, 1, "a\n"},
		{"write and close", nil, writeErr, closeErr, 1, "a\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var r io.Reader = strings.NewReader("a\nb\n")
			if tc.readErr != nil {
				r = io.MultiReader(strings.NewReader("a\n"), failedReader{tc.readErr})
			}
			w := &testWriter{closeErr: tc.closeErr, failAt: 2, writeErr: tc.writeErr}
			if tc.writeErr == nil {
				w.failAt = 0
			}
			n, err := lineexport.Export(r, w)
			if n != tc.count || w.output.String() != tc.output {
				t.Errorf("Export count/output = %d/%q, want %d/%q", n, w.output.String(), tc.count, tc.output)
			}
			workErr := tc.readErr
			if tc.writeErr != nil {
				workErr = tc.writeErr
			}
			switch {
			case workErr == nil:
				if err != tc.closeErr {
					t.Errorf("Export error = %v, want identical error %v", err, tc.closeErr)
				}
			case tc.closeErr == nil:
				if err != workErr {
					t.Errorf("Export error = %v, want identical error %v", err, workErr)
				}
			default:
				if !errors.Is(err, workErr) || !errors.Is(err, tc.closeErr) {
					t.Errorf("Export error = %v, want both causes %v and %v", err, workErr, tc.closeErr)
				}
			}
			if w.closes != 1 {
				t.Errorf("Close calls = %d, want 1", w.closes)
			}
		})
	}
}

func TestExportErrorTypes(t *testing.T) {
	readErr := readFailure{"source"}
	closeErr := new(closeFailure)
	w := &testWriter{closeErr: closeErr}
	n, err := lineexport.Export(failedReader{readErr}, w)
	var gotRead readFailure
	var gotClose *closeFailure
	if !errors.As(err, &gotRead) || len(gotRead) != 1 || gotRead[0] != "source" {
		t.Errorf("Export error = %v, want read failure type and data", err)
	}
	if !errors.As(err, &gotClose) || gotClose != closeErr {
		t.Errorf("Export error = %v, want original close failure", err)
	}
	if n != 0 || w.writes != 0 || w.closes != 1 {
		t.Errorf("count/writes/closes = %d/%d/%d, want 0/0/1", n, w.writes, w.closes)
	}
}

func TestExportLimit(t *testing.T) {
	for _, tc := range []struct {
		name   string
		limit  int
		count  int
		output string
	}{
		{"zero is unlimited", 0, 3, "a\n\nlast\n"},
		{"one record", 1, 1, "a\n"},
		{"empty record counts", 2, 2, "a\n\n"},
		{"unterminated record counts", 3, 3, "a\n\nlast\n"},
		{"limit exceeds input", 4, 3, "a\n\nlast\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := new(testWriter)
			n, err := lineexport.ExportLimit(strings.NewReader("a\n\nlast"), w, tc.limit)
			if n != tc.count || err != nil || w.output.String() != tc.output {
				t.Errorf("ExportLimit = (%d, %v), output %q; want (%d, nil), output %q", n, err, w.output.String(), tc.count, tc.output)
			}
			if w.closes != 1 {
				t.Errorf("Close calls = %d, want 1", w.closes)
			}
		})
	}
}

func TestExportLimitRejectsNegativeBeforeIO(t *testing.T) {
	closeErr := errors.New("close failed")
	for _, finalErr := range []error{nil, closeErr} {
		r := &recordReader{records: []string{"a\n"}}
		w := &testWriter{closeErr: finalErr}
		n, err := lineexport.ExportLimit(r, w, -1)
		if n != 0 || err == nil {
			t.Errorf("ExportLimit(-1) = (%d, %v), want (0, error)", n, err)
		}
		if finalErr != nil && (!errors.Is(err, finalErr) || err == finalErr) {
			t.Errorf("ExportLimit error = %v, want validation and close causes", err)
		}
		if r.reads != 0 || w.writes != 0 || w.closes != 1 || r.closes != 0 {
			t.Errorf("reads/writes/writer closes/reader closes = %d/%d/%d/%d, want 0/0/1/0", r.reads, w.writes, w.closes, r.closes)
		}
	}
}

func TestExportLimitStopsBeforeNextRead(t *testing.T) {
	r := &recordReader{records: []string{"a\n", "b\n"}}
	w := new(testWriter)
	n, err := lineexport.ExportLimit(r, w, 1)
	if n != 1 || err != nil || w.output.String() != "a\n" || r.reads != 1 {
		t.Errorf("ExportLimit = (%d, %v), output %q, reads %d; want (1, nil), output a\\n, one read", n, err, w.output.String(), r.reads)
	}
	if w.closes != 1 || r.closes != 0 {
		t.Errorf("writer/reader closes = %d/%d, want 1/0", w.closes, r.closes)
	}
}

func TestExportLimitReturnsCloseErrorAtLimit(t *testing.T) {
	closeErr := errors.New("close failed")
	w := &testWriter{closeErr: closeErr}
	n, err := lineexport.ExportLimit(strings.NewReader("a\nb\n"), w, 1)
	if n != 1 || err != closeErr || w.output.String() != "a\n" || w.closes != 1 {
		t.Errorf("ExportLimit = (%d, %v), output %q, closes %d; want (1, close error), output a\\n, one close", n, err, w.output.String(), w.closes)
	}
}

func TestExportWritesBeforeNextRead(t *testing.T) {
	w := new(testWriter)
	r := &recordReader{records: []string{"a\n", "b\n"}}
	r.check = func() error {
		if w.writes < r.reads {
			return errors.New("next record read before previous record written")
		}
		return nil
	}
	n, err := lineexport.Export(r, w)
	if n != 2 || err != nil || w.output.String() != "a\nb\n" {
		t.Errorf("Export = (%d, %v), output %q; want (2, nil), output a\\nb\\n", n, err, w.output.String())
	}
	if w.closes != 1 || r.closes != 0 {
		t.Errorf("writer/reader closes = %d/%d, want 1/0", w.closes, r.closes)
	}
}

func TestExportPartialWrite(t *testing.T) {
	writeErr := errors.New("write failed")
	for _, tc := range []struct {
		name     string
		writeErr error
		wantErr  error
	}{
		{"partial write with error", writeErr, writeErr},
		{"short write without error", nil, io.ErrShortWrite},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &recordReader{records: []string{"a\n", "b\n", "c\n"}}
			w := &testWriter{failAt: 2, writeErr: tc.writeErr, partial: 1}
			n, err := lineexport.Export(r, w)
			if n != 1 || err != tc.wantErr || w.output.String() != "a\nb" || r.reads != 2 {
				t.Errorf("Export = (%d, %v), output %q, reads %d; want (1, %v), output a\\nb, two reads", n, err, w.output.String(), r.reads, tc.wantErr)
			}
			if w.closes != 1 || r.closes != 0 {
				t.Errorf("writer/reader closes = %d/%d, want 1/0", w.closes, r.closes)
			}
		})
	}
}
