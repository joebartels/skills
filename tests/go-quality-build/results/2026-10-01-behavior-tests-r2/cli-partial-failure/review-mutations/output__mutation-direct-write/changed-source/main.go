package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func run(args []string, input io.Reader, stderr io.Writer) error {
	fs := flag.NewFlagSet("ledgerload", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "", "record directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dir == "" || fs.NArg() != 0 {
		return fmt.Errorf("--dir required and no positional arguments allowed")
	}
	reader := csv.NewReader(input)
	reader.FieldsPerRecord = 2
	for record := 1; ; record++ {
		row, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("record %d: read CSV: %w", record, err)
		}
		if row[0] == "" || strings.ContainsAny(row[0], "/\\\x00") || row[0] == "." || row[0] == ".." {
			return fmt.Errorf("record %d: invalid ID %q", record, row[0])
		}
		qty, err := strconv.Atoi(row[1])
		if err != nil || qty <= 0 {
			return fmt.Errorf("record %d: invalid quantity %q", record, row[1])
		}
		if err := os.MkdirAll(*dir, 0o755); err != nil {
			return fmt.Errorf("record %d: create directory: %w", record, err)
		}
		if err := writeRecord(*dir, row[0], qty); err != nil {
			return fmt.Errorf("record %d: write %q: %w", record, row[0], err)
		}
	}
}

// Publish only complete records so a failed duplicate cannot damage an earlier write.
func writeRecord(dir, id string, qty int) error {
	return os.WriteFile(filepath.Join(dir, id), []byte(strconv.Itoa(qty)+"\n"), 0o644)
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
