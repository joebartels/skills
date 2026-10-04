package entryzip

import (
	"archive/zip"
	"io"
)

type Entry struct {
	Name string
	Body io.Reader
}

func WriteArchive(w io.Writer, entries []Entry) error {
	archive := zip.NewWriter(w)
	defer archive.Close()
	for _, entry := range entries {
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
