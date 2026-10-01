package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func run(args []string, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: linesum FILE")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	total := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil {
			return err
		}
		total += n
	}
	_, err = fmt.Fprintln(out, total)
	return err
}
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
