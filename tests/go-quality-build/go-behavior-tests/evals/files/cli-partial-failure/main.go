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
	row, err := reader.Read()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	if len(row) != 2 || row[0] == "" || strings.ContainsAny(row[0], "/\\") || row[0] == "." || row[0] == ".." {
		return fmt.Errorf("invalid record")
	}
	qty, err := strconv.Atoi(row[1])
	if err != nil || qty <= 0 {
		return fmt.Errorf("invalid quantity")
	}
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*dir, row[0]), []byte(strconv.Itoa(qty)+"\n"), 0o644)
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
