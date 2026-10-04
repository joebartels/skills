package main

import (
	"fmt"
	"os"
	"strconv"

	"example.com/lineexport"
)

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: lineexport INPUT OUTPUT [LIMIT]")
		os.Exit(2)
	}
	limit := 0
	if len(os.Args) == 4 {
		var err error
		limit, err = strconv.Atoi(os.Args[3])
		if err != nil || limit < 0 {
			fmt.Fprintf(os.Stderr, "invalid record limit %q: expected a non-negative integer\n", os.Args[3])
			os.Exit(2)
		}
	}
	r, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer r.Close()
	w, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_, err = lineexport.ExportLimit(r, w, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
