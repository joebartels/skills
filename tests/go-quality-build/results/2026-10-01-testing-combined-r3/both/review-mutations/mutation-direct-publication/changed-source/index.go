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

// writeAtomic publishes only a completely written and closed temporary file.
// The temporary file shares the destination directory so rename is atomic on
// the supported POSIX filesystem.
func writeAtomic(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}
func (i *Index) Get(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", ErrInvalidRecord
	}
	b, err := os.ReadFile(filepath.Join(i.root, key))
	return string(b), err
}

// ApplyCSV applies headerless key/text rows in order, retaining the accepted
// prefix when a later row cannot be parsed, validated or written.
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
