package indexer

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrInvalidRecord = errors.New("invalid record")
var ErrNotImplemented = errors.New("operation not implemented")
var keyPattern = regexp.MustCompile(`^[a-z]+$`)

type Index struct{ root string }

func Open(root string) *Index { return &Index{root: root} }
func (i *Index) Put(key, text string) error {
	if !keyPattern.MatchString(key) {
		return ErrInvalidRecord
	}
	if err := os.MkdirAll(i.root, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(i.root, ".record-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.WriteString(text); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(i.root, key))
}
func (i *Index) Get(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", ErrInvalidRecord
	}
	b, err := os.ReadFile(filepath.Join(i.root, key))
	return string(b), err
}
func (i *Index) ApplyCSV(r io.Reader) error {
	rdr := csv.NewReader(r)
	rdr.FieldsPerRecord = 2
	for {
		row, err := rdr.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if !keyPattern.MatchString(row[0]) || strings.TrimSpace(row[1]) == "" {
			return ErrInvalidRecord
		}
		if err := i.Put(row[0], row[1]); err != nil {
			return err
		}
	}
}
