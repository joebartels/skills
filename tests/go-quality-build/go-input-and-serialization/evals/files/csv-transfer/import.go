// Package catalogue imports warehouse catalogue uploads.
package catalogue

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

const (
	MaxBytes      = 32768
	MaxRows       = 128
	MaxFieldBytes = 1024
)

// Item is one catalogue entry. A nil Stock means the upload did not specify it.
type Item struct {
	ID    uint64
	Name  string
	Stock *uint64
}

// Import reads an upload and replaces the warehouse catalogue.
func Import(r io.Reader, replace func([]Item) error) error {
	reader := csv.NewReader(r)
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	record, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read item: %w", err)
	}
	if len(record) != 3 {
		return fmt.Errorf("item has %d fields", len(record))
	}
	id, err := strconv.ParseUint(record[0], 10, 64)
	if err != nil {
		return fmt.Errorf("parse id: %w", err)
	}
	item := Item{ID: id, Name: record[1]}
	if record[2] != "" {
		stock, err := strconv.ParseUint(record[2], 10, 64)
		if err != nil {
			return fmt.Errorf("parse stock: %w", err)
		}
		item.Stock = &stock
	}
	return replace([]Item{item})
}
