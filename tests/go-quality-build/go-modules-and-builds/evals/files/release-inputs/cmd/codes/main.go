// Command codes checks document identifiers.
package main

import (
	"fmt"
	"io"
	"os"

	"example.com/codes"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: codes FORMAT VALUE")
		return 2
	}
	if !codes.Valid(args[0], args[1]) {
		fmt.Fprintln(stdout, "invalid")
		return 1
	}
	fmt.Fprintln(stdout, "valid")
	return 0
}
