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
	return writeAtomic(filepath.Join(i.root, key), []byte(text))
}
func (i *Index) Get(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", ErrInvalidRecord
	}
	b, err := os.ReadFile(filepath.Join(i.root, key))
	return string(b), err
}

// ApplyCSV accepts records in order and retains the accepted prefix on error.
func (i *Index) ApplyCSV(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	for row := 1; ; row++ {
		fields, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("CSV row %d: %w", row, err)
		}
		if len(fields) != 2 {
			return fmt.Errorf("CSV row %d: expected two fields: %w", row, ErrInvalidRecord)
		}
		if err := validateRecord(fields[0], fields[1]); err != nil {
			return fmt.Errorf("CSV row %d: %w", row, err)
		}
		if err := i.Put(fields[0], fields[1]); err != nil {
			return fmt.Errorf("CSV row %d: %w", row, err)
		}
	}
}

func validateRecord(key, text string) error {
	if !keyPattern.MatchString(key) || strings.TrimSpace(text) == "" {
		return ErrInvalidRecord
	}
	return nil
}

// writeAtomic publishes only after the complete temporary file is closed.
func writeAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".record-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
