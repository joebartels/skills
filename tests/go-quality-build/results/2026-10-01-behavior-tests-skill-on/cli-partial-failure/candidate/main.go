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
	dirReady := false
	for record := 1; ; record++ {
		row, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read record %d: %w", record, err)
		}
		if row[0] == "" || strings.ContainsAny(row[0], "/\\") || row[0] == "." || row[0] == ".." {
			return fmt.Errorf("record %d: invalid record ID %q", record, row[0])
		}
		qty, err := strconv.Atoi(row[1])
		if err != nil || qty <= 0 {
			return fmt.Errorf("record %d (%q): invalid quantity %q", record, row[0], row[1])
		}
		if !dirReady {
			if err := os.MkdirAll(*dir, 0o755); err != nil {
				return fmt.Errorf("create record directory: %w", err)
			}
			dirReady = true
		}
		if err := writeRecord(*dir, row[0], qty); err != nil {
			return fmt.Errorf("write record %d (%q): %w", record, row[0], err)
		}
	}
}

// Publish only a complete record, so a failed duplicate cannot truncate an
// earlier accepted value. The temporary file stays on the same filesystem.
func writeRecord(dir, id string, quantity int) error {
	file, err := os.CreateTemp(dir, ".ledgerload-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0o644); err != nil {
		file.Close()
		return err
	}
	if _, err := file.WriteString(strconv.Itoa(quantity) + "\n"); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(dir, id))
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
