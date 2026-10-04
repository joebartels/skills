package lineexport

import (
	"bufio"
	"errors"
	"io"
)

// Export copies all records from r to w and returns the accepted record count.
// It borrows r and closes w exactly once, reporting read, write and close errors.
func Export(r io.Reader, w io.WriteCloser) (int, error) {
	return ExportLimit(r, w, 0)
}

// ExportLimit copies records from r to w and returns the accepted record count.
// Zero means unlimited; a positive limit stops after that many records.
// A negative limit fails before reading or writing. It borrows r and closes w
// exactly once on every path. A sole error is returned unchanged; independent
// operation and close errors are joined for errors.Is and errors.As.
func ExportLimit(r io.Reader, w io.WriteCloser, limit int) (n int, err error) {
	defer func() {
		closeErr := w.Close()
		if err == nil {
			err = closeErr
			return
		}
		if closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	if limit < 0 {
		return 0, errors.New("limit must be non-negative")
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		record := scanner.Text() + "\n"
		written, err := io.WriteString(w, record)
		if err != nil {
			return n, err
		}
		if written != len(record) {
			return n, io.ErrShortWrite
		}
		n++
		if limit > 0 && n == limit {
			return n, nil
		}
	}
	return n, scanner.Err()
}
