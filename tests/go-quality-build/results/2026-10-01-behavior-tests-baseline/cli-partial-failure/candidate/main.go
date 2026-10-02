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
			return fmt.Errorf("record %d: reading CSV: %w", record, err)
		}
		id := row[0]
		if id == "" || strings.ContainsAny(id, "/\\") || id == "." || id == ".." {
			return fmt.Errorf("record %d: invalid id %q", record, id)
		}
		qty, err := strconv.Atoi(row[1])
		if err != nil || qty <= 0 {
			return fmt.Errorf("record %d: invalid quantity %q", record, row[1])
		}
		if !dirReady {
			if err := os.MkdirAll(*dir, 0o755); err != nil {
				return fmt.Errorf("record %d: creating record directory: %w", record, err)
			}
			dirReady = true
		}
		if err := os.WriteFile(filepath.Join(*dir, id), []byte(strconv.Itoa(qty)+"\n"), 0o644); err != nil {
			return fmt.Errorf("record %d: writing %q: %w", record, id, err)
		}
	}
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
