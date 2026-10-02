package main

import (
	"example.com/indexer"
	"fmt"
	"os"
)

func run(args []string) error {
	if len(args) != 4 || args[0] != "put" {
		return fmt.Errorf("usage: indexer put DIR KEY TEXT")
	}
	return indexer.Open(args[1]).Put(args[2], args[3])
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
