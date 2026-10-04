package entryzip

import (
	"archive/zip"
	"errors"
	"io"
)

// Entry is a named ZIP entry with a borrowed body reader.
type Entry struct {
	Name string
	Body io.Reader
}

// WriteArchive writes all entries in order. It has the same completion,
// ownership, and error behavior as WriteSelected with a nil predicate.
func WriteArchive(w io.Writer, entries []Entry) error {
	return WriteSelected(w, entries, nil)
}

// WriteSelected writes entries accepted by keep in their original order.
// A nil keep accepts all entries. Rejected entries are neither created nor read.
// Success includes writing the central directory. Independent processing and
// completion failures are joined; a lone caller error is returned directly.
// It does not close w or any body reader. Failure may leave an incomplete
// archive; no rollback or durability is guaranteed.
func WriteSelected(w io.Writer, entries []Entry, keep func(string) bool) (err error) {
	output := &archiveOutput{writer: w}
	archive := zip.NewWriter(output)
	defer func() {
		writeFailed := output.failed
		closeErr := archive.Close()
		if err == nil {
			err = closeErr
			return
		}
		// A previous output error is sticky in ZIP's buffer; Close repeats it.
		if closeErr != nil && !writeFailed {
			err = errors.Join(err, closeErr)
		}
	}()
	for _, entry := range entries {
		if keep != nil && !keep(entry.Name) {
			continue
		}
		destination, err := archive.Create(entry.Name)
		if err != nil {
			return err
		}
		if _, err := io.Copy(destination, entry.Body); err != nil {
			return err
		}
	}
	return nil
}

type archiveOutput struct {
	writer io.Writer
	failed bool
}

func (w *archiveOutput) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil {
		w.failed = true
	}
	return n, err
}
