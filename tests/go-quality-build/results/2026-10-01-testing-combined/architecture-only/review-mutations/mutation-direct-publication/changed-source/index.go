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

// Index stores each key in a separate file under its root.
type Index struct{ root string }

// Open returns an independent index rooted at root.
func Open(root string) *Index { return &Index{root: root} }

// Put atomically replaces key's text. Unlike imported records, text may be empty.
func (i *Index) Put(key, text string) error {
	if !keyPattern.MatchString(key) {
		return ErrInvalidRecord
	}
	if err := os.MkdirAll(i.root, 0700); err != nil {
		return err
	}
	return publish(filepath.Join(i.root, key), []byte(text))
}

// publish replaces a file only after its complete replacement has been closed.
func publish(path string, data []byte) error { return os.WriteFile(path, data, 0600) }

// Get returns the exact text stored for key.
func (i *Index) Get(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", ErrInvalidRecord
	}
	b, err := os.ReadFile(filepath.Join(i.root, key))
	return string(b), err
}

// ApplyCSV applies two-field key/text rows in order, retaining the accepted
// prefix if a later row is rejected. Duplicate keys replace earlier values.
func (i *Index) ApplyCSV(r io.Reader) error {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = 2
	for row := 1; ; row++ {
		fields, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if errors.Is(err, csv.ErrFieldCount) {
				err = errors.Join(ErrInvalidRecord, err)
			}
			return fmt.Errorf("CSV row %d: %w", row, err)
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
