package lineexport

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// Export writes all records from r to w.
// It follows the ownership and error behavior of ExportLimit.
func Export(r io.Reader, w io.WriteCloser) (int, error) {
	return ExportLimit(r, w, 0)
}

// ExportLimit writes at most limit records from r to w and returns the accepted count.
// Zero means unlimited. A negative limit is rejected before reading or writing.
// Empty and final unterminated records count; output records end with a newline.
// It closes w exactly once on every path and leaves r open. A single failure is
// returned unchanged; independent operation and Close failures are joined.
// Reading uses bufio.Scanner's default token size limit and may read ahead.
func ExportLimit(r io.Reader, w io.WriteCloser, limit int) (n int, err error) {
	defer func() {
		closeErr := w.Close()
		if closeErr == nil {
			return
		}
		if err == nil {
			err = closeErr
			return
		}
		err = errors.Join(err, closeErr)
	}()
	if limit < 0 {
		return 0, fmt.Errorf("negative record limit: %d", limit)
	}
	scanner := bufio.NewScanner(r)
	for limit == 0 || n < limit {
		if !scanner.Scan() {
			return n, scanner.Err()
		}
		record := scanner.Text() + "\n"
		written, err := io.WriteString(w, record)
		if err != nil {
			return n, err
		}
		if written < len(record) {
			return n, io.ErrShortWrite
		}
		n++
	}
	return n, nil
}
