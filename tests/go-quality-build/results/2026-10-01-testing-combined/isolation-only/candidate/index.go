package indexer

import (
	"encoding/csv"
	"errors"
	"fmt"
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

// ApplyCSV applies records in order, retaining the accepted prefix on failure.
func (i *Index) ApplyCSV(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	for row := 1; ; row++ {
		fields, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read CSV row %d: %w", row, err)
		}
		if len(fields) != 2 || !validRecord(fields[0], fields[1]) {
			return fmt.Errorf("CSV row %d: %w", row, ErrInvalidRecord)
		}
		if err := i.Put(fields[0], fields[1]); err != nil {
			return fmt.Errorf("write CSV row %d: %w", row, err)
		}
	}
}

func validRecord(key, text string) bool {
	return keyPattern.MatchString(key) && strings.TrimSpace(text) != ""
}
