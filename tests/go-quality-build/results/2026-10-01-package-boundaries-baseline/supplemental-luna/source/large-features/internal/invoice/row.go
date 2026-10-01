package invoice

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Row is the shared representation used by settlement input and ledger records.
type Row struct {
	ID    string
	Cents int
}

// Encode returns the canonical ID,CENTS ledger representation, including its newline.
func Encode(row Row) ([]byte, error) {
	if err := validate(row); err != nil {
		return nil, err
	}
	return []byte(row.ID + "," + strconv.Itoa(row.Cents) + "\n"), nil
}

// Decode parses one ID,CENTS row. A single trailing LF is accepted for file input.
func Decode(line string) (Row, error) {
	line = strings.TrimSuffix(line, "\n")
	if strings.ContainsAny(line, "\r\n") {
		return Row{}, fmt.Errorf("invalid invoice row")
	}
	parts := strings.Split(line, ",")
	if len(parts) != 2 {
		return Row{}, fmt.Errorf("invalid invoice row")
	}
	cents, err := strconv.Atoi(parts[1])
	if err != nil {
		return Row{}, fmt.Errorf("invalid invoice amount: %w", err)
	}
	row := Row{ID: parts[0], Cents: cents}
	if err := validate(row); err != nil {
		return Row{}, err
	}
	return row, nil
}

// Read decodes rows sequentially and stops at the first malformed row or callback error.
func Read(r io.Reader, apply func(Row) error) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		row, err := Decode(scanner.Text())
		if err != nil {
			return err
		}
		if err := apply(row); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func validate(row Row) error {
	if row.ID == "" || strings.ContainsAny(row.ID, ",\r\n") || row.Cents <= 0 {
		return fmt.Errorf("invalid invoice")
	}
	return nil
}
