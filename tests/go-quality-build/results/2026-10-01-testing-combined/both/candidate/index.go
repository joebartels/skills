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

// ErrInvalidRecord identifies a rejected key or import/refresh record.
var ErrInvalidRecord = errors.New("invalid record")

// ErrNotImplemented is retained for compatibility with existing consumers.
var ErrNotImplemented = errors.New("operation not implemented")
var keyPattern = regexp.MustCompile(`^[a-z]+$`)

type Index struct{ root string }

// Open returns an index rooted at root. It does not access the filesystem.
func Open(root string) *Index { return &Index{root: root} }

// Put atomically replaces key's text. Text may be empty or whitespace.
func (i *Index) Put(key, text string) error {
	if !keyPattern.MatchString(key) {
		return ErrInvalidRecord
	}
	if err := os.MkdirAll(i.root, 0700); err != nil {
		return err
	}
	return replaceFile(filepath.Join(i.root, key), []byte(text))
}

// Get returns the exact text stored for key.
func (i *Index) Get(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", ErrInvalidRecord
	}
	b, err := os.ReadFile(filepath.Join(i.root, key))
	return string(b), err
}

// ApplyCSV applies two-field key/text records in order. On failure, earlier
// accepted rows remain stored and no later rows are applied.
func (i *Index) ApplyCSV(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 2
	for row := 1; ; row++ {
		fields, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				err = errors.Join(ErrInvalidRecord, err)
			}
			return fmt.Errorf("CSV row %d: %w", row, err)
		}
		if err := validateRecord(Record{Key: fields[0], Text: fields[1]}); err != nil {
			return fmt.Errorf("CSV row %d: %w", row, err)
		}
		if err := i.Put(fields[0], fields[1]); err != nil {
			return fmt.Errorf("CSV row %d: %w", row, err)
		}
	}
}

func validateRecord(record Record) error {
	if !keyPattern.MatchString(record.Key) || strings.TrimSpace(record.Text) == "" {
		return ErrInvalidRecord
	}
	return nil
}

// replaceFile stages data beside the destination so rename publishes a complete
// snapshot. Failure before rename leaves the destination untouched.
func replaceFile(path string, data []byte) error {
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
	return os.Rename(f.Name(), path)
}
