package main

import (
	"fmt"
	"os"
	"strconv"

	"example.com/lineexport"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) != 2 && len(args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: lineexport INPUT OUTPUT [LIMIT]")
		return 2
	}
	limit := 0
	if len(args) == 3 {
		var err error
		limit, err = strconv.Atoi(args[2])
		if err != nil || limit < 0 {
			fmt.Fprintln(os.Stderr, "usage: lineexport INPUT OUTPUT [LIMIT]")
			fmt.Fprintln(os.Stderr, "LIMIT must be a non-negative integer")
			return 2
		}
	}
	r, err := os.Open(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer r.Close()
	w, err := os.Create(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	_, err = lineexport.ExportLimit(r, w, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
