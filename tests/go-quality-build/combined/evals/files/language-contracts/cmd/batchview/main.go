package main

import (
	"encoding/json"
	"example.com/batchview"
	"fmt"
	"os"
)

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: batchview INPUT OUTPUT")
		return 2
	}
	input, e := os.Open(args[0])
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	defer input.Close()
	items, e := batchview.Load(input)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	data, e := json.Marshal(items)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	if e = os.WriteFile(args[1], data, 0600); e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	return 0
}
